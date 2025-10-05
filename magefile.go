//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	tools = []string{
		"./internal/cmd/protoc-gen-go-flipt-sdk/...",
	}

	Default = Build
)

// Installs tools required for development
func Bootstrap() error {
	fmt.Println(" > Bootstrapping tools...")
	for _, tool := range tools {
		if tool == "./internal/cmd/protoc-gen-go-flipt-sdk/..." {
			cmd := exec.Command("go", "install", ".")

			cmd.Dir = "internal/cmd/protoc-gen-go-flipt-sdk"
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("installing tool %q: %w", tool, err)
			}
			continue
		}
		if err := sh.RunV("go", "install", "-v", tool); err != nil {
			return fmt.Errorf("installing tool %q: %w", tool, err)
		}
	}

	return sh.RunV("go", "install", "tool")
}

// Builds the project similar to a release build
func Build() error {
	mg.Deps(Prep)
	fmt.Println(" > Building...")

	if err := build(buildModeProd); err != nil {
		return err
	}

	fmt.Printf("\nRun the following to start Flipt:\n")
	fmt.Printf("\n%v\n", color.CyanString(`./bin/flipt server [--config config/local.yml]`))
	return nil
}

// Cleans up built files
func Clean() error {
	fmt.Println(" > Cleaning...")

	if err := sh.RunV("go", "mod", "tidy"); err != nil {
		return fmt.Errorf("tidying go.mod: %w", err)
	}

	clean := []string{"dist/*", "pkg/*", "bin/*", "ui/dist"}
	for _, dir := range clean {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("removing dir %q: %w", dir, err)
		}
	}

	return nil
}

// Prepares the project for building
func Prep() error {
	fmt.Println("> Preparing...")
	mg.Deps(Clean)
	mg.Deps(UI.Build)
	return nil
}

type buildMode uint8

const (
	// buildModeDev builds the project for development, without bundling assets
	buildModeDev buildMode = iota
	// BuildModeProd builds the project similar to a release build
	buildModeProd
)

func build(mode buildMode) error {
	buildDate := time.Now().UTC().Format(time.RFC3339)
	buildArgs := make([]string, 0)

	switch mode {
	case buildModeProd:
		buildArgs = append(buildArgs, "-tags", "assets")
	}

	gitCommit, err := sh.Output("git", "rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("getting git commit: %w", err)
	}

	ldFlags := []string{
		fmt.Sprintf("-X main.commit=%s", gitCommit),
		fmt.Sprintf("-X main.date=%s", buildDate),
	}

	var (
		keygenVerifyKey = os.Getenv("KEYGEN_VERIFY_KEY")
		keygenAccountID = os.Getenv("KEYGEN_ACCOUNT_ID")
		keygenProductID = os.Getenv("KEYGEN_PRODUCT_ID")
	)

	if keygenVerifyKey != "" {
		ldFlags = append(ldFlags, fmt.Sprintf("-X main.keygenVerifyKey=%s", keygenVerifyKey))
	}
	if keygenAccountID != "" {
		ldFlags = append(ldFlags, fmt.Sprintf("-X main.keygenAccountID=%s", keygenAccountID))
	}
	if keygenProductID != "" {
		ldFlags = append(ldFlags, fmt.Sprintf("-X main.keygenProductID=%s", keygenProductID))
	}

	buildArgs = append([]string{"build", "-trimpath", "-ldflags", strings.Join(ldFlags, " ")}, buildArgs...)
	buildArgs = append(buildArgs, "-o", "./bin/flipt", "./cmd/flipt/")

	return sh.RunV("go", buildArgs...)
}

type Go mg.Namespace

// Keeping these aliases for backwards compatibility for now
var Aliases = map[string]interface{}{
	"dev":    Go.Run,
	"test":   Go.Test,
	"bench":  Go.Bench,
	"lint":   Go.Lint,
	"fmt":    Go.Fmt,
	"proto":  Go.Proto,
	"ui:dev": UI.Run,
}

// Runs Go benchmarking tests
func (g Go) Bench() error {
	fmt.Println(" > Running benchmarks...")

	return sh.RunV("go", "test", "-run", "XXX", "-bench", ".", "-benchmem", "-short", "./...")
}

// Runs the Go server in development mode using the local config, without bundling assets
func (g Go) Run() error {
	config := "config/dev.yml"

	// check if config/dev.yml exists, if not use config/local.yml
	if _, err := os.Stat(config); os.IsNotExist(err) {
		config = "config/local.yml"
	}

	return sh.RunV("go", "run", "./cmd/flipt/...", "server", "--config", config)
}

// Builds the Go server for development, without bundling assets
func (g Go) Build() error {
	mg.Deps(Clean)
	fmt.Println(" > Building...")

	if err := build(buildModeDev); err != nil {
		return err
	}

	fmt.Printf("\nRun the following to start Flipt server:\n")
	fmt.Printf("\n%v\n", color.CyanString(`./bin/flipt server [--config config/local.yml]`))
	fmt.Printf("\nIn another shell, run the following to start the UI in dev mode:\n")
	fmt.Printf("\n%v\n", color.CyanString(`cd ui && npm run dev`))
	return nil
}

// Runs the Go tests and generates a coverage report
func (g Go) Cover() error {
	mg.Deps(Go.Test)
	fmt.Println(" > Running coverage...")
	return sh.RunV("go", "tool", "cover", "-html=coverage.txt")
}

var (
	ignoreFmtDirs  = []string{"rpc/", "sdk/"}
	ignoreFmtFiles = []string{"*.pb.go", "*.gen.go", "*_generated.go"}
)

// Formats Go code
func (g Go) Fmt() error {
	fmt.Println(" > Formatting...")
	files, err := findFilesRecursive(func(path string, _ os.FileInfo) bool {
		// only go files, ignoring generated files
		if filepath.Ext(path) != ".go" {
			return false
		}

		// Check directory exclusions
		for _, dir := range ignoreFmtDirs {
			if filepath.HasPrefix(path, dir) {
				return false
			}
		}

		// Check file pattern exclusions
		filename := filepath.Base(path)
		for _, pattern := range ignoreFmtFiles {
			if matched, _ := filepath.Match(pattern, filename); matched {
				return false
			}
		}

		return true
	})
	if err != nil {
		return fmt.Errorf("finding files: %w", err)
	}

	// Process files in batches to avoid "argument list too long" errors
	const batchSize = 50
	for i := 0; i < len(files); i += batchSize {
		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}

		batch := files[i:end]
		args := append([]string{"tool", "golang.org/x/tools/cmd/goimports", "-w"}, batch...)
		if err := sh.RunV("go", args...); err != nil {
			return err
		}
	}

	return nil
}

// Runs the Go modernize tool to fix linting errors
func (g Go) Modernize() error {
	fmt.Println(" > Running modernize...")
	return sh.RunV("go", "run", "golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest", "-category=-omitzero", "-fix", "-test", "./...")
}

// Runs the Go linters
func (g Go) Lint() error {
	fmt.Println(" > Linting...")

	if err := sh.RunV("go", "tool", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint", "run"); err != nil {
		return fmt.Errorf("linting: %w", err)
	}

	return sh.RunV("go", "tool", "github.com/bufbuild/buf/cmd/buf", "lint")
}

// Generates the Go protobuf files and gRPC stubs
func (g Go) Proto() error {
	mg.Deps(Bootstrap)
	cmdArgs := []string{"buf", "generate"}
	if _, err := exec.LookPath("buf"); err != nil {
		cmdArgs = []string{"go", "tool", "github.com/bufbuild/buf/cmd/buf", "generate"}
	}
	fmt.Println(" > Generating proto files...")
	for _, module := range []string{
		"rpc/flipt",
		"rpc/v2/environments",
		"rpc/v2/analytics",
		"rpc/v2/evaluation",
	} {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = module
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

// Generates mocks
func (g Go) Mockery() error {
	fmt.Println(" > Generating mocks...")
	return sh.RunV("go", "tool", "github.com/vektra/mockery/v3")
}

// Generates mocks and proto files
func (g Go) Generate() error {
	if err := g.Mockery(); err != nil {
		return err
	}

	return g.Proto()
}

// Runs the Go unit tests
func (g Go) Test() error {
	fmt.Println(" > Testing...")

	testArgs := []string{"tool", "github.com/rakyll/gotest", "-v", "-covermode=atomic", "-count=1", "-coverprofile=coverage.txt", "-timeout=60s", "-coverpkg=./..."}

	if os.Getenv("FLIPT_TEST_SHORT") != "" {
		testArgs = append(testArgs, "-short")
	}

	testArgs = append(testArgs, "./...")
	return sh.RunV("go", testArgs...)
}

type UI mg.Namespace

// Installs UI dependencies
func (u UI) Deps() error {
	fmt.Println(" > Installing UI deps...")

	// check if node_modules exists, if not run npm ci and return
	if _, err := os.Stat("ui/node_modules"); err != nil && os.IsNotExist(err) {
		cmd := exec.Command("npm", "ci")
		cmd.Dir = "ui"
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// only run if deps have changed
	// uses: https://github.com/thdk/package-changed
	cmd := exec.Command("npx", "--no", "package-changed", "install", "--ci")
	cmd.Dir = "ui"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Runs the UI in development mode
func (u UI) Run() error {
	mg.Deps(u.Deps)

	fmt.Println(" > Starting UI...")

	cmd := exec.Command("npm", "run", "dev")
	cmd.Dir = "ui"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Formats UI code
func (u UI) Fmt() error {
	fmt.Println(" > Formatting UI...")

	cmd := exec.Command("npm", "run", "format")
	cmd.Dir = "ui"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Runs the UI linters
func (u UI) Lint() error {
	fmt.Println(" > Linting UI...")

	cmd := exec.Command("npm", "run", "lint")
	cmd.Dir = "ui"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Builds all UI assets for release distribution
func (u UI) Build() error {
	mg.Deps(u.Deps)

	fmt.Println(" > Generating assets...")

	cmd := exec.Command("npm", "run", "build")
	cmd.Dir = "ui"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

type Dagger mg.Namespace

func (d Dagger) Call(command, args string) error {
	return sh.RunV("dagger", append([]string{"call", command, "--source=."}, strings.Split(args, " ")...)...)
}

// findFilesRecursive recursively traverses from the CWD and invokes the given
// match function on each regular file to determine if the given path should be
// returned as a match. It ignores files in .git directories.
func findFilesRecursive(match func(path string, info os.FileInfo) bool) ([]string, error) {
	var matches []string
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Don't look for files in git directories
		if info.Mode().IsDir() && filepath.Base(path) == ".git" {
			return filepath.SkipDir
		}

		if !info.Mode().IsRegular() {
			// continue
			return nil
		}

		if match(filepath.ToSlash(path), info) {
			matches = append(matches, path)
		}
		return nil
	})
	return matches, err
}
