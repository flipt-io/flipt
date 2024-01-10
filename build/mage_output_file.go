// +build ignore

package main

import (
	"context"
	_flag "flag"
	_fmt "fmt"
	_ioutil "io/ioutil"
	_log "log"
	"os"
	"os/signal"
	_filepath "path/filepath"
	_sort "sort"
	"strconv"
	_strings "strings"
	"syscall"
	_tabwriter "text/tabwriter"
	"time"
	
)

func main() {
	// Use local types and functions in order to avoid name conflicts with additional magefiles.
	type arguments struct {
		Verbose       bool          // print out log statements
		List          bool          // print out a list of targets
		Help          bool          // print out help for a specific target
		Timeout       time.Duration // set a timeout to running the targets
		Args          []string      // args contain the non-flag command-line arguments
	}

	parseBool := func(env string) bool {
		val := os.Getenv(env)
		if val == "" {
			return false
		}		
		b, err := strconv.ParseBool(val)
		if err != nil {
			_log.Printf("warning: environment variable %s is not a valid bool value: %v", env, val)
			return false
		}
		return b
	}

	parseDuration := func(env string) time.Duration {
		val := os.Getenv(env)
		if val == "" {
			return 0
		}		
		d, err := time.ParseDuration(val)
		if err != nil {
			_log.Printf("warning: environment variable %s is not a valid duration value: %v", env, val)
			return 0
		}
		return d
	}
	args := arguments{}
	fs := _flag.FlagSet{}
	fs.SetOutput(os.Stdout)

	// default flag set with ExitOnError and auto generated PrintDefaults should be sufficient
	fs.BoolVar(&args.Verbose, "v", parseBool("MAGEFILE_VERBOSE"), "show verbose output when running targets")
	fs.BoolVar(&args.List, "l", parseBool("MAGEFILE_LIST"), "list targets for this binary")
	fs.BoolVar(&args.Help, "h", parseBool("MAGEFILE_HELP"), "print out help for a specific target")
	fs.DurationVar(&args.Timeout, "t", parseDuration("MAGEFILE_TIMEOUT"), "timeout in duration parsable format (e.g. 5m30s)")
	fs.Usage = func() {
		_fmt.Fprintf(os.Stdout, `
%s [options] [target]

Commands:
  -l    list targets in this binary
  -h    show this help

Options:
  -h    show description of a target
  -t <string>
        timeout in duration parsable format (e.g. 5m30s)
  -v    show verbose output when running targets
 `[1:], _filepath.Base(os.Args[0]))
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		// flag will have printed out an error already.
		return
	}
	args.Args = fs.Args()
	if args.Help && len(args.Args) == 0 {
		fs.Usage()
		return
	}
		
	// color is ANSI color type
	type color int

	// If you add/change/remove any items in this constant,
	// you will need to run "stringer -type=color" in this directory again.
	// NOTE: Please keep the list in an alphabetical order.
	const (
		black color = iota
		red
		green
		yellow
		blue
		magenta
		cyan
		white
		brightblack
		brightred
		brightgreen
		brightyellow
		brightblue
		brightmagenta
		brightcyan
		brightwhite
	)

	// AnsiColor are ANSI color codes for supported terminal colors.
	var ansiColor = map[color]string{
		black:         "\u001b[30m",
		red:           "\u001b[31m",
		green:         "\u001b[32m",
		yellow:        "\u001b[33m",
		blue:          "\u001b[34m",
		magenta:       "\u001b[35m",
		cyan:          "\u001b[36m",
		white:         "\u001b[37m",
		brightblack:   "\u001b[30;1m",
		brightred:     "\u001b[31;1m",
		brightgreen:   "\u001b[32;1m",
		brightyellow:  "\u001b[33;1m",
		brightblue:    "\u001b[34;1m",
		brightmagenta: "\u001b[35;1m",
		brightcyan:    "\u001b[36;1m",
		brightwhite:   "\u001b[37;1m",
	}
	
	const _color_name = "blackredgreenyellowbluemagentacyanwhitebrightblackbrightredbrightgreenbrightyellowbrightbluebrightmagentabrightcyanbrightwhite"

	var _color_index = [...]uint8{0, 5, 8, 13, 19, 23, 30, 34, 39, 50, 59, 70, 82, 92, 105, 115, 126}

	colorToLowerString := func (i color) string {
		if i < 0 || i >= color(len(_color_index)-1) {
			return "color(" + strconv.FormatInt(int64(i), 10) + ")"
		}
		return _color_name[_color_index[i]:_color_index[i+1]]
	}

	// ansiColorReset is an ANSI color code to reset the terminal color.
	const ansiColorReset = "\033[0m"

	// defaultTargetAnsiColor is a default ANSI color for colorizing targets.
	// It is set to Cyan as an arbitrary color, because it has a neutral meaning
	var defaultTargetAnsiColor = ansiColor[cyan]

	getAnsiColor := func(color string) (string, bool) {
		colorLower := _strings.ToLower(color)
		for k, v := range ansiColor {
			colorConstLower := colorToLowerString(k)
			if colorConstLower == colorLower {
				return v, true
			}
		}
		return "", false
	}

	// Terminals which  don't support color:
	// 	TERM=vt100
	// 	TERM=cygwin
	// 	TERM=xterm-mono
    var noColorTerms = map[string]bool{
		"vt100":      false,
		"cygwin":     false,
		"xterm-mono": false,
	}

	// terminalSupportsColor checks if the current console supports color output
	//
	// Supported:
	// 	linux, mac, or windows's ConEmu, Cmder, putty, git-bash.exe, pwsh.exe
	// Not supported:
	// 	windows cmd.exe, powerShell.exe
	terminalSupportsColor := func() bool {
		envTerm := os.Getenv("TERM")
		if _, ok := noColorTerms[envTerm]; ok {
			return false
		}
		return true
	}

	// enableColor reports whether the user has requested to enable a color output.
	enableColor := func() bool {
		b, _ := strconv.ParseBool(os.Getenv("MAGEFILE_ENABLE_COLOR"))
		return b
	}

	// targetColor returns the ANSI color which should be used to colorize targets.
	targetColor := func() string {
		s, exists := os.LookupEnv("MAGEFILE_TARGET_COLOR")
		if exists == true {
			if c, ok := getAnsiColor(s); ok == true {
				return c
			}
		}
		return defaultTargetAnsiColor
	}

	// store the color terminal variables, so that the detection isn't repeated for each target
	var enableColorValue = enableColor() && terminalSupportsColor()
	var targetColorValue = targetColor()

	printName := func(str string) string {
		if enableColorValue {
			return _fmt.Sprintf("%s%s%s", targetColorValue, str, ansiColorReset)
		} else {
			return str
		}
	}

	list := func() error {
		
		targets := map[string]string{
			"build:base": "builds Flipts base image via Dagger and buildkit.",
			"build:flipt": "builds a development version of Flipt as a Docker image and loads it into a local Docker instance.",
			"generate:screenshots": "",
			"release:changelog": "",
			"release:latest": "",
			"release:latestRC": "",
			"release:next": "",
			"release:tag": "",
			"test:cli": "runs a suite of test cases which exercise the `flipt` binary CLI.",
			"test:integration": "runs the entire integration test suite (one of [\"*\", \"list\", \"<test-case>\"] use \"list\" to see available cases).",
			"test:loadTest": "runs a load test against a running instance of Flipt using Pyroscope and vegeta.",
			"test:migration": "runs the latest Flipts migration against a database built upon the latest Flipt release.",
			"test:ui": "runs the entire integration test suite for the UI.",
			"test:unit": "runs the base suite of tests for all of Flipt.",
		}

		keys := make([]string, 0, len(targets))
		for name := range targets {
			keys = append(keys, name)
		}
		_sort.Strings(keys)

		_fmt.Println("Targets:")
		w := _tabwriter.NewWriter(os.Stdout, 0, 4, 4, ' ', 0)
		for _, name := range keys {
			_fmt.Fprintf(w, "  %v\t%v\n", printName(name), targets[name])
		}
		err := w.Flush()
		return err
	}

	var ctx context.Context
	ctxCancel := func(){}

	// by deferring in a closure, we let the cancel function get replaced
	// by the getContext function.
	defer func() {
		ctxCancel()
	}()

	getContext := func() (context.Context, func()) {
		if ctx == nil {
			if args.Timeout != 0 {
				ctx, ctxCancel = context.WithTimeout(context.Background(), args.Timeout)
			} else {
				ctx, ctxCancel = context.WithCancel(context.Background())
			}
		}

		return ctx, ctxCancel
	}

	runTarget := func(logger *_log.Logger, fn func(context.Context) error) interface{} {
		var err interface{}
		ctx, cancel := getContext()
		d := make(chan interface{})
		go func() {
			defer func() {
				err := recover()
				d <- err
			}()
			err := fn(ctx)
			d <- err
		}()
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT)
		select {
		case <-sigCh:
			logger.Println("cancelling mage targets, waiting up to 5 seconds for cleanup...")
			cancel()
			cleanupCh := time.After(5 * time.Second)

			select {
			// target exited by itself
			case err = <-d:
				return err
			// cleanup timeout exceeded
			case <-cleanupCh:
				return _fmt.Errorf("cleanup timeout exceeded")
			// second SIGINT received
			case <-sigCh:
				logger.Println("exiting mage")
				return _fmt.Errorf("exit forced")
			}
		case <-ctx.Done():
			cancel()
			e := ctx.Err()
			_fmt.Printf("ctx err: %v\n", e)
			return e
		case err = <-d:
			// we intentionally don't cancel the context here, because
			// the next target will need to run with the same context.
			return err
		}
	}
	// This is necessary in case there aren't any targets, to avoid an unused
	// variable error.
	_ = runTarget

	handleError := func(logger *_log.Logger, err interface{}) {
		if err != nil {
			logger.Printf("Error: %+v\n", err)
			type code interface {
				ExitStatus() int
			}
			if c, ok := err.(code); ok {
				os.Exit(c.ExitStatus())
			}
			os.Exit(1)
		}
	}
	_ = handleError

	// Set MAGEFILE_VERBOSE so mg.Verbose() reflects the flag value.
	if args.Verbose {
		os.Setenv("MAGEFILE_VERBOSE", "1")
	} else {
		os.Setenv("MAGEFILE_VERBOSE", "0")
	}

	_log.SetFlags(0)
	if !args.Verbose {
		_log.SetOutput(_ioutil.Discard)
	}
	logger := _log.New(os.Stderr, "", 0)
	if args.List {
		if err := list(); err != nil {
			_log.Println(err)
			os.Exit(1)
		}
		return
	}

	if args.Help {
		if len(args.Args) < 1 {
			logger.Println("no target specified")
			os.Exit(2)
		}
		switch _strings.ToLower(args.Args[0]) {
			case "build:base":
				_fmt.Println("Base builds Flipts base image via Dagger and buildkit. This can be used for debugging or cache warming. There is no resulting artefact (only buildkit cache state).")
				_fmt.Println()
				
				_fmt.Print("Usage:\n\n\tmage build:base\n\n")
				var aliases []string
				if len(aliases) > 0 {
					_fmt.Printf("Aliases: %s\n\n", _strings.Join(aliases, ", "))
				}
				return
			case "build:flipt":
				_fmt.Println("Flipt builds a development version of Flipt as a Docker image and loads it into a local Docker instance.")
				_fmt.Println()
				
				_fmt.Print("Usage:\n\n\tmage build:flipt\n\n")
				var aliases []string
				if len(aliases) > 0 {
					_fmt.Printf("Aliases: %s\n\n", _strings.Join(aliases, ", "))
				}
				return
			case "generate:screenshots":
				
				_fmt.Print("Usage:\n\n\tmage generate:screenshots\n\n")
				var aliases []string
				if len(aliases) > 0 {
					_fmt.Printf("Aliases: %s\n\n", _strings.Join(aliases, ", "))
				}
				return
			case "release:changelog":
				
				_fmt.Print("Usage:\n\n\tmage release:changelog <module> <version>\n\n")
				var aliases []string
				if len(aliases) > 0 {
					_fmt.Printf("Aliases: %s\n\n", _strings.Join(aliases, ", "))
				}
				return
			case "release:latest":
				
				_fmt.Print("Usage:\n\n\tmage release:latest <module>\n\n")
				var aliases []string
				if len(aliases) > 0 {
					_fmt.Printf("Aliases: %s\n\n", _strings.Join(aliases, ", "))
				}
				return
			case "release:latestrc":
				
				_fmt.Print("Usage:\n\n\tmage release:latestrc <module>\n\n")
				var aliases []string
				if len(aliases) > 0 {
					_fmt.Printf("Aliases: %s\n\n", _strings.Join(aliases, ", "))
				}
				return
			case "release:next":
				
				_fmt.Print("Usage:\n\n\tmage release:next <module> <versionParts>\n\n")
				var aliases []string
				if len(aliases) > 0 {
					_fmt.Printf("Aliases: %s\n\n", _strings.Join(aliases, ", "))
				}
				return
			case "release:tag":
				
				_fmt.Print("Usage:\n\n\tmage release:tag <module> <version>\n\n")
				var aliases []string
				if len(aliases) > 0 {
					_fmt.Printf("Aliases: %s\n\n", _strings.Join(aliases, ", "))
				}
				return
			case "test:cli":
				_fmt.Println("CLI runs a suite of test cases which exercise the `flipt` binary CLI.")
				_fmt.Println()
				
				_fmt.Print("Usage:\n\n\tmage test:cli\n\n")
				var aliases []string
				if len(aliases) > 0 {
					_fmt.Printf("Aliases: %s\n\n", _strings.Join(aliases, ", "))
				}
				return
			case "test:integration":
				_fmt.Println("Integration runs the entire integration test suite (one of [\"*\", \"list\", \"<test-case>\"] use \"list\" to see available cases). The suite runs a number of operations via the Go SDK against Flipt in various configurations using both HTTP and GRPC.")
				_fmt.Println()
				
				_fmt.Print("Usage:\n\n\tmage test:integration <cases>\n\n")
				var aliases []string
				if len(aliases) > 0 {
					_fmt.Printf("Aliases: %s\n\n", _strings.Join(aliases, ", "))
				}
				return
			case "test:loadtest":
				_fmt.Println("LoadTest runs a load test against a running instance of Flipt using Pyroscope and vegeta.")
				_fmt.Println()
				
				_fmt.Print("Usage:\n\n\tmage test:loadtest\n\n")
				var aliases []string
				if len(aliases) > 0 {
					_fmt.Printf("Aliases: %s\n\n", _strings.Join(aliases, ", "))
				}
				return
			case "test:migration":
				_fmt.Println("Migration runs the latest Flipts migration against a database built upon the latest Flipt release. It ensures that migrations runs successfully and Flipt continues to pass readonly tests.")
				_fmt.Println()
				
				_fmt.Print("Usage:\n\n\tmage test:migration\n\n")
				var aliases []string
				if len(aliases) > 0 {
					_fmt.Printf("Aliases: %s\n\n", _strings.Join(aliases, ", "))
				}
				return
			case "test:ui":
				_fmt.Println("UI runs the entire integration test suite for the UI.")
				_fmt.Println()
				
				_fmt.Print("Usage:\n\n\tmage test:ui\n\n")
				var aliases []string
				if len(aliases) > 0 {
					_fmt.Printf("Aliases: %s\n\n", _strings.Join(aliases, ", "))
				}
				return
			case "test:unit":
				_fmt.Println("Unit runs the base suite of tests for all of Flipt. It uses SQLite as the default database.")
				_fmt.Println()
				
				_fmt.Print("Usage:\n\n\tmage test:unit\n\n")
				var aliases []string
				if len(aliases) > 0 {
					_fmt.Printf("Aliases: %s\n\n", _strings.Join(aliases, ", "))
				}
				return
			default:
				logger.Printf("Unknown target: %q\n", args.Args[0])
				os.Exit(2)
		}
	}
	if len(args.Args) < 1 {
		if err := list(); err != nil {
			logger.Println("Error:", err)
			os.Exit(1)
		}
		return
	}
	for x := 0; x < len(args.Args); {
		target := args.Args[x]
		x++

		// resolve aliases
		switch _strings.ToLower(target) {
		
		}

		switch _strings.ToLower(target) {
		
			case "build:base":
				expected := x + 0
				if expected > len(args.Args) {
					// note that expected and args at this point include the arg for the target itself
					// so we subtract 1 here to show the number of args without the target.
					logger.Printf("not enough arguments for target \"Build:Base\", expected %v, got %v\n", expected-1, len(args.Args)-1)
					os.Exit(2)
				}
				if args.Verbose {
					logger.Println("Running target:", "Build:Base")
				}
				
				wrapFn := func(ctx context.Context) error {
					return Build{}.Base(ctx)
				}
				ret := runTarget(logger, wrapFn)
				handleError(logger, ret)
			case "build:flipt":
				expected := x + 0
				if expected > len(args.Args) {
					// note that expected and args at this point include the arg for the target itself
					// so we subtract 1 here to show the number of args without the target.
					logger.Printf("not enough arguments for target \"Build:Flipt\", expected %v, got %v\n", expected-1, len(args.Args)-1)
					os.Exit(2)
				}
				if args.Verbose {
					logger.Println("Running target:", "Build:Flipt")
				}
				
				wrapFn := func(ctx context.Context) error {
					return Build{}.Flipt(ctx)
				}
				ret := runTarget(logger, wrapFn)
				handleError(logger, ret)
			case "generate:screenshots":
				expected := x + 0
				if expected > len(args.Args) {
					// note that expected and args at this point include the arg for the target itself
					// so we subtract 1 here to show the number of args without the target.
					logger.Printf("not enough arguments for target \"Generate:Screenshots\", expected %v, got %v\n", expected-1, len(args.Args)-1)
					os.Exit(2)
				}
				if args.Verbose {
					logger.Println("Running target:", "Generate:Screenshots")
				}
				
				wrapFn := func(ctx context.Context) error {
					return Generate{}.Screenshots(ctx)
				}
				ret := runTarget(logger, wrapFn)
				handleError(logger, ret)
			case "release:changelog":
				expected := x + 2
				if expected > len(args.Args) {
					// note that expected and args at this point include the arg for the target itself
					// so we subtract 1 here to show the number of args without the target.
					logger.Printf("not enough arguments for target \"Release:Changelog\", expected %v, got %v\n", expected-1, len(args.Args)-1)
					os.Exit(2)
				}
				if args.Verbose {
					logger.Println("Running target:", "Release:Changelog")
				}
				
			arg0 := args.Args[x]
			x++
			arg1 := args.Args[x]
			x++
				wrapFn := func(ctx context.Context) error {
					return Release{}.Changelog(ctx, arg0, arg1)
				}
				ret := runTarget(logger, wrapFn)
				handleError(logger, ret)
			case "release:latest":
				expected := x + 1
				if expected > len(args.Args) {
					// note that expected and args at this point include the arg for the target itself
					// so we subtract 1 here to show the number of args without the target.
					logger.Printf("not enough arguments for target \"Release:Latest\", expected %v, got %v\n", expected-1, len(args.Args)-1)
					os.Exit(2)
				}
				if args.Verbose {
					logger.Println("Running target:", "Release:Latest")
				}
				
			arg0 := args.Args[x]
			x++
				wrapFn := func(ctx context.Context) error {
					return Release{}.Latest(ctx, arg0)
				}
				ret := runTarget(logger, wrapFn)
				handleError(logger, ret)
			case "release:latestrc":
				expected := x + 1
				if expected > len(args.Args) {
					// note that expected and args at this point include the arg for the target itself
					// so we subtract 1 here to show the number of args without the target.
					logger.Printf("not enough arguments for target \"Release:LatestRC\", expected %v, got %v\n", expected-1, len(args.Args)-1)
					os.Exit(2)
				}
				if args.Verbose {
					logger.Println("Running target:", "Release:LatestRC")
				}
				
			arg0 := args.Args[x]
			x++
				wrapFn := func(ctx context.Context) error {
					return Release{}.LatestRC(ctx, arg0)
				}
				ret := runTarget(logger, wrapFn)
				handleError(logger, ret)
			case "release:next":
				expected := x + 2
				if expected > len(args.Args) {
					// note that expected and args at this point include the arg for the target itself
					// so we subtract 1 here to show the number of args without the target.
					logger.Printf("not enough arguments for target \"Release:Next\", expected %v, got %v\n", expected-1, len(args.Args)-1)
					os.Exit(2)
				}
				if args.Verbose {
					logger.Println("Running target:", "Release:Next")
				}
				
			arg0 := args.Args[x]
			x++
			arg1 := args.Args[x]
			x++
				wrapFn := func(ctx context.Context) error {
					return Release{}.Next(ctx, arg0, arg1)
				}
				ret := runTarget(logger, wrapFn)
				handleError(logger, ret)
			case "release:tag":
				expected := x + 2
				if expected > len(args.Args) {
					// note that expected and args at this point include the arg for the target itself
					// so we subtract 1 here to show the number of args without the target.
					logger.Printf("not enough arguments for target \"Release:Tag\", expected %v, got %v\n", expected-1, len(args.Args)-1)
					os.Exit(2)
				}
				if args.Verbose {
					logger.Println("Running target:", "Release:Tag")
				}
				
			arg0 := args.Args[x]
			x++
			arg1 := args.Args[x]
			x++
				wrapFn := func(ctx context.Context) error {
					return Release{}.Tag(ctx, arg0, arg1)
				}
				ret := runTarget(logger, wrapFn)
				handleError(logger, ret)
			case "test:cli":
				expected := x + 0
				if expected > len(args.Args) {
					// note that expected and args at this point include the arg for the target itself
					// so we subtract 1 here to show the number of args without the target.
					logger.Printf("not enough arguments for target \"Test:CLI\", expected %v, got %v\n", expected-1, len(args.Args)-1)
					os.Exit(2)
				}
				if args.Verbose {
					logger.Println("Running target:", "Test:CLI")
				}
				
				wrapFn := func(ctx context.Context) error {
					return Test{}.CLI(ctx)
				}
				ret := runTarget(logger, wrapFn)
				handleError(logger, ret)
			case "test:integration":
				expected := x + 1
				if expected > len(args.Args) {
					// note that expected and args at this point include the arg for the target itself
					// so we subtract 1 here to show the number of args without the target.
					logger.Printf("not enough arguments for target \"Test:Integration\", expected %v, got %v\n", expected-1, len(args.Args)-1)
					os.Exit(2)
				}
				if args.Verbose {
					logger.Println("Running target:", "Test:Integration")
				}
				
			arg0 := args.Args[x]
			x++
				wrapFn := func(ctx context.Context) error {
					return Test{}.Integration(ctx, arg0)
				}
				ret := runTarget(logger, wrapFn)
				handleError(logger, ret)
			case "test:loadtest":
				expected := x + 0
				if expected > len(args.Args) {
					// note that expected and args at this point include the arg for the target itself
					// so we subtract 1 here to show the number of args without the target.
					logger.Printf("not enough arguments for target \"Test:LoadTest\", expected %v, got %v\n", expected-1, len(args.Args)-1)
					os.Exit(2)
				}
				if args.Verbose {
					logger.Println("Running target:", "Test:LoadTest")
				}
				
				wrapFn := func(ctx context.Context) error {
					return Test{}.LoadTest(ctx)
				}
				ret := runTarget(logger, wrapFn)
				handleError(logger, ret)
			case "test:migration":
				expected := x + 0
				if expected > len(args.Args) {
					// note that expected and args at this point include the arg for the target itself
					// so we subtract 1 here to show the number of args without the target.
					logger.Printf("not enough arguments for target \"Test:Migration\", expected %v, got %v\n", expected-1, len(args.Args)-1)
					os.Exit(2)
				}
				if args.Verbose {
					logger.Println("Running target:", "Test:Migration")
				}
				
				wrapFn := func(ctx context.Context) error {
					return Test{}.Migration(ctx)
				}
				ret := runTarget(logger, wrapFn)
				handleError(logger, ret)
			case "test:ui":
				expected := x + 0
				if expected > len(args.Args) {
					// note that expected and args at this point include the arg for the target itself
					// so we subtract 1 here to show the number of args without the target.
					logger.Printf("not enough arguments for target \"Test:UI\", expected %v, got %v\n", expected-1, len(args.Args)-1)
					os.Exit(2)
				}
				if args.Verbose {
					logger.Println("Running target:", "Test:UI")
				}
				
				wrapFn := func(ctx context.Context) error {
					return Test{}.UI(ctx)
				}
				ret := runTarget(logger, wrapFn)
				handleError(logger, ret)
			case "test:unit":
				expected := x + 0
				if expected > len(args.Args) {
					// note that expected and args at this point include the arg for the target itself
					// so we subtract 1 here to show the number of args without the target.
					logger.Printf("not enough arguments for target \"Test:Unit\", expected %v, got %v\n", expected-1, len(args.Args)-1)
					os.Exit(2)
				}
				if args.Verbose {
					logger.Println("Running target:", "Test:Unit")
				}
				
				wrapFn := func(ctx context.Context) error {
					return Test{}.Unit(ctx)
				}
				ret := runTarget(logger, wrapFn)
				handleError(logger, ret)
		
		default:
			logger.Printf("Unknown target specified: %q\n", target)
			os.Exit(2)
		}
	}
}




