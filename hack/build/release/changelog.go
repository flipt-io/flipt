package release

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

var (
	conventionalCommit = regexp.MustCompile(`^([^(:]*)(?:\(([^\)]*)\))?:(.*)`)
	semverHeader       = regexp.MustCompile(`^[#]+.*(v[0-9]+.[0-9]+.[0-9]+).*`)
)

type changeLogVersion struct {
	Prefix     string
	Version    string
	Date       time.Time
	Added      []string
	Changed    []string
	Deprecated []string
	Removed    []string
	Fixed      []string
	Security   []string
}

func parseChangeLogVersion(prefix, version string, date time.Time, entries ...string) changeLogVersion {
	v := changeLogVersion{
		Prefix:  prefix,
		Version: version,
		Date:    date,
	}

	for _, entry := range entries {
		matches := conventionalCommit.FindStringSubmatch(entry)
		if len(matches) == 0 {
			continue
		}

		var (
			typ     = matches[1]
			message = strings.TrimSpace(matches[3])
		)

		if scope := matches[2]; scope != "" {
			message = fmt.Sprintf("`%s`: %s", scope, message)
		}

		switch typ {
		case "feat":
			v.Added = append(v.Added, message)
		case "fix":
			v.Fixed = append(v.Fixed, message)
		case "chore":
		default:
			v.Changed = append(v.Changed, message)
		}
	}

	return v
}

func (c changeLogVersion) write(dst io.Writer) {
	for _, list := range []struct {
		name    string
		entries []string
	}{
		{"Added", c.Added},
		{"Changed", c.Changed},
		{"Deprecated", c.Deprecated},
		{"Removed", c.Removed},
		{"Fixed", c.Fixed},
		{"Security", c.Security},
	} {
		if len(list.entries) == 0 {
			continue
		}

		fmt.Fprint(dst, "### ", list.name, "\n\n")
		for _, entry := range list.entries {
			fmt.Fprintf(dst, "- %s\n", entry)
		}
		fmt.Fprintln(dst)
	}
}

func insertChangeLogEntryIntoFile(path string, log changeLogVersion) error {
	fi, err := os.Open(path)
	if err != nil {
		return err
	}

	defer fi.Close()

	dst := &bytes.Buffer{}

	if err := insertChangeLogVersion(dst, fi, log); err != nil {
		return err
	}

	_ = fi.Close()

	return os.WriteFile(path, dst.Bytes(), os.ModePerm)
}

func insertChangeLogVersion(dst io.Writer, src io.Reader, log changeLogVersion) error {
	rd := bufio.NewReader(src)
	for {
		line, err := rd.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}

		if !semverHeader.Match(line) {
			if _, err := dst.Write(line); err != nil {
				return err
			}
			if err == io.EOF {
				return nil
			}
			continue
		}

		tag := log.Version
		if log.Prefix != "" {
			tag = path.Join(log.Prefix, log.Version)
		}

		// write changelog heading
		fmt.Fprintf(dst, "\n## [%s](https://github.com/flipt-io/flipt/releases/tag/%s) - %s\n\n", log.Version, tag, log.Date.Format(time.DateOnly))

		// write out all changelog entries
		log.write(dst)

		if _, err := dst.Write(line); err != nil {
			return err
		}

		// copy the rest
		_, err = io.Copy(dst, rd)

		return err
	}
}
