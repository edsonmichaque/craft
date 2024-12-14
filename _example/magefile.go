//go:build mage
package main

import (
    "github.com/magefile/mage/mg"
    "github.com/magefile/mage/sh"
    "os"
    "path/filepath"
)

const ciScripts = "./scripts/ci"

// Default target to run when none is specified
var Default = Build

// Test runs the test suite
func Test() error {
    return sh.RunV(filepath.Join(ciScripts, "test"))
}

// TestWatch runs tests in watch mode
func TestWatch() error {
    return sh.RunV(filepath.Join(ciScripts, "test"), "--watch")
}

// Build builds the project
func Build() error {
    return sh.RunV(filepath.Join(ciScripts, "build"))
}

// Docker builds Docker images
func Docker() error {
    return sh.RunV(filepath.Join(ciScripts, "build"), "docker")
}

// CI runs the full CI pipeline
func CI() error {
    return sh.RunV(filepath.Join(ciScripts, "ci"))
}

// CITest tests CI configurations locally
func CITest() error {
    return sh.RunV(filepath.Join(ciScripts, "utils", "ci-tester.sh"))
}

// Clean removes build artifacts
func Clean() error {
    return sh.RunV(filepath.Join(ciScripts, "utils", "cleanup.sh"))
}

// Release creates a new release
func Release() error {
    return sh.RunV(filepath.Join(ciScripts, "tasks", "release.sh"))
}

// Namespace example
type DB mg.Namespace

// Start starts the database
func (DB) Start() error {
    return sh.RunV(filepath.Join(ciScripts, "utils", "db.sh"), "start"))
}

// Migrate runs database migrations
func (DB) Migrate() error {
    return sh.RunV(filepath.Join(ciScripts, "utils", "db.sh"), "migrate"))
}

// Seed seeds the database
func (DB) Seed() error {
    return sh.RunV(filepath.Join(ciScripts, "utils", "db.sh"), "seed"))
}
