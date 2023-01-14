//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	tools = []string{
		"github.com/bufbuild/buf/cmd/buf",
		"github.com/bufbuild/buf/cmd/protoc-gen-buf-breaking",
		"github.com/bufbuild/buf/cmd/protoc-gen-buf-lint",
		"github.com/golangci/golangci-lint/cmd/golangci-lint",
		"github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway",
		"github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2",
		"golang.org/x/tools/cmd/cover",
		"golang.org/x/tools/cmd/goimports",
		"google.golang.org/grpc/cmd/protoc-gen-go-grpc",
		"google.golang.org/protobuf/cmd/protoc-gen-go",
	}
)

// Bootstrap installs tools
func Bootstrap() error {
	fmt.Println("Bootstrapping tools...")
	if err := os.Chdir("_tools"); err != nil {
		return fmt.Errorf("failed to change dir: %w", err)
	}

	// create module if go.mod doesnt exist
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		cmd := exec.Command("go", "mod", "init", "tools")
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	goInstall := sh.RunCmd("go", "install", "-v")
	return goInstall(tools...)
}

// Proto generates protobuf files
func Proto() error {
	fmt.Println("Generating proto files...")
	if err := sh.RunV("buf", "generate"); err != nil {
		return fmt.Errorf("failed to generate proto files: %w", err)
	}
	return Fmt()
}

// Clean up built files
func Clean() error {
	fmt.Println("Cleaning...")
	if err := sh.Run("go", "mod", "tidy"); err != nil {
		return fmt.Errorf("failed to tidy go.mod: %w", err)
	}
	if err := sh.Run("go", "clean", "-i", "./..."); err != nil {
		return fmt.Errorf("failed to clean cache: %w", err)
	}
	clean := []string{"dist/*", "pkg/*", "bin/*", "ui/dist"}
	for _, dir := range clean {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("failed to remove dir %q: %w", dir, err)
		}
	}
	return nil
}

// Fmt formats code
func Fmt() error {
	fmt.Println("Formatting...")
	return sh.RunV("goimports", "-w", ".")
}

// Lint runs the linters
func Lint() error {
	fmt.Println("Linting...")
	if err := sh.Run("golangci-lint", "run"); err != nil {
		return fmt.Errorf("failed to lint: %w", err)
	}
	return sh.Run("buf", "lint")
}

func Prep() error {
	fmt.Println("Preparing...")
	mg.Deps(Clean)
	mg.Deps(Proto)
	mg.Deps(UI.Build)
	return nil
}

type UI mg.Namespace

// Build generates UI assets
func (u UI) Build() error {
	fmt.Println("Generating assets...")
	if err := os.Chdir(".build/ui"); err != nil {
		return fmt.Errorf("failed to change dir: %w", err)
	}
	if err := sh.RunV("npm", "run", "build"); err != nil {
		return err
	}
	return sh.RunV("mv", "dist", "../../ui")
}

// Clone clones the UI repo
func (u UI) Clone() error {
	if _, err := os.Stat(".build/ui/.git"); os.IsNotExist(err) {
		fmt.Println("Cloning UI repo...")
		if err := os.Chdir(".build"); err != nil {
			return fmt.Errorf("failed to change dir: %w", err)
		}
		if err := sh.RunV("git", "clone", "https://github.com/flipt-io/flipt-ui.git", "ui", "--depth=1"); err != nil {
			return fmt.Errorf("failed to clone UI repo: %w", err)
		}
	}
	return nil
}

// Sync syncs the UI repo
func (u UI) Sync() error {
	mg.Deps(u.Clone)
	fmt.Println("Syncing UI repo...")
	if err := os.Chdir(".build/ui"); err != nil {
		return fmt.Errorf("failed to change dir: %w", err)
	}
	if err := sh.RunV("git", "fetch", "--all"); err != nil {
		return fmt.Errorf("failed to fetch UI repo: %w", err)
	}
	if err := sh.RunV("git", "reset", "--hard", "origin/main"); err != nil {
		return fmt.Errorf("failed to reset UI repo: %w", err)
	}
	return nil
}

// Deps installs UI deps
// TODO: only install if package.json has changed
func (u UI) Deps() error {
	mg.Deps(u.Clone)
	fmt.Println("Installing UI deps...")
	if err := os.Chdir(".build/ui"); err != nil {
		return fmt.Errorf("failed to change dir: %w", err)
	}
	return sh.RunV("npm", "ci")
}
