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

// Bootstrap
func Bootstrap() error {
	fmt.Println("Bootstrapping deps...")
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
	return sh.RunV("buf", "generate")
}

// Clean up built files
func Clean() error {
	fmt.Println("Cleaning...")
	clean := []string{"dist/*", "pkg/*", "bin/*", "ui/dist/*"}
	for _, dir := range clean {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("failed to remove dir %q: %w", dir, err)
		}
	}
	return nil
}

func Fmt() error {
	fmt.Println("Formatting...")
	return sh.RunV("goimports", "-w", ".")
}

func Lint() error {
	fmt.Println("Linting...")
	if err := sh.Run("golangci-lint", "run"); err != nil {
		return fmt.Errorf("failed to lint: %w", err)
	}
	return sh.Run("buf", "lint")
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
	mg.SerialDeps(u.Clone)
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
	mg.SerialDeps(u.Clone)
	fmt.Println("Installing UI deps...")
	if err := os.Chdir(".build/ui"); err != nil {
		return fmt.Errorf("failed to change dir: %w", err)
	}
	return sh.RunV("npm", "ci")
}
