// Package main provides the CLI tool for managing extensions and modules.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/llamagate/llamagate/internal/migration"
	"github.com/llamagate/llamagate/internal/packaging"
	"github.com/llamagate/llamagate/internal/registry"
	"github.com/llamagate/llamagate/internal/startup"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "import":
		handleImport()
	case "export":
		handleExport()
	case "list":
		handleList()
	case "remove":
		handleRemove()
	case "enable":
		handleEnable()
	case "disable":
		handleDisable()
	case "migrate":
		handleMigrate()
	case "sync":
		handleSync()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `LlamaGate CLI - Manage extensions and agentic modules

Usage: llamagate <command> [arguments]

Commands:
  import extension <zip-file>        Import an extension from zip file
  import agentic-module <zip-file>   Import an agentic module from zip file
  import <zip-file>                  Auto-detect type and import

  export extension <id> --out <zip>  Export an extension to zip file
  export agentic-module <id> --out <zip>  Export a module to zip file

  list extensions                    List all installed extensions
  list agentic-modules               List all installed modules

  remove extension <id>              Remove an extension
  remove agentic-module <id>         Remove a module

  enable extension <id>              Enable an extension
  disable extension <id>             Disable an extension
  enable agentic-module <id>         Enable a module
  disable agentic-module <id>        Disable a module

  migrate                            Migrate legacy extensions to new layout
  sync                               Sync registry with filesystem

Examples:
  llamagate import extension my-ext.zip
  llamagate export extension my-ext --out my-ext.zip
  llamagate list extensions
  llamagate remove extension my-ext
`)
}

func handleImport() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Error: import requires a zip file path\n")
		os.Exit(1)
	}

	var zipPath string

	// Check if type is specified (for future use, but Import auto-detects)
	if len(os.Args) >= 4 && (os.Args[2] == "extension" || os.Args[2] == "agentic-module" || os.Args[2] == "ext" || os.Args[2] == "module" || os.Args[2] == "am") {
		zipPath = os.Args[3]
	} else {
		// Auto-detect type
		zipPath = os.Args[2]
	}

	result, err := packaging.Import(zipPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error importing: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully imported %s '%s' (v%s)\n", result.Type, result.Name, result.Version)
	fmt.Printf("  ID: %s\n", result.ID)
	fmt.Printf("  Path: %s\n", result.Path)
	fmt.Printf("  Enabled: %v\n", result.Enabled)
	if !result.Enabled {
		fmt.Printf("  Note: Extension is disabled. Enable it with: llamagate enable extension %s\n", result.ID)
	}
}

func handleExport() {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	outPath := fs.String("out", "", "Output zip file path (required)")
	fs.Parse(os.Args[2:])

	if *outPath == "" {
		fmt.Fprintf(os.Stderr, "Error: --out flag is required\n")
		os.Exit(1)
	}

	args := fs.Args()
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: export requires type and id\n")
		os.Exit(1)
	}

	var itemType registry.ItemType
	switch args[0] {
	case "extension", "ext":
		itemType = registry.ItemTypeExtension
	case "agentic-module", "module", "am":
		itemType = registry.ItemTypeAgenticModule
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid type '%s'. Use 'extension' or 'agentic-module'\n", args[0])
		os.Exit(1)
	}

	id := args[1]

	if err := packaging.Export(id, *outPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error exporting: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully exported %s '%s' to %s\n", itemType, id, *outPath)
}

func handleList() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Error: list requires 'extensions' or 'agentic-modules'\n")
		os.Exit(1)
	}

	itemTypeStr := os.Args[2]
	var itemType registry.ItemType

	switch itemTypeStr {
	case "extensions", "ext":
		itemType = registry.ItemTypeExtension
	case "agentic-modules", "modules", "am":
		itemType = registry.ItemTypeAgenticModule
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid type '%s'. Use 'extensions' or 'agentic-modules'\n", itemTypeStr)
		os.Exit(1)
	}

	reg, err := registry.NewRegistry()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	items := reg.List(itemType)
	if len(items) == 0 {
		fmt.Printf("No %s installed.\n", itemTypeStr)
		return
	}

	fmt.Printf("Found %d %s:\n\n", len(items), itemTypeStr)
	for _, item := range items {
		status := "disabled"
		if item.Enabled {
			status = "enabled"
		}
		fmt.Printf("  %s (v%s) [%s]\n", item.Name, item.Version, status)
		fmt.Printf("    ID: %s\n", item.ID)
		fmt.Printf("    Path: %s\n", item.SourcePath)
		fmt.Println()
	}
}

func handleRemove() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Error: remove requires type and id\n")
		os.Exit(1)
	}

	var itemType registry.ItemType
	switch os.Args[2] {
	case "extension", "ext":
		itemType = registry.ItemTypeExtension
	case "agentic-module", "module", "am":
		itemType = registry.ItemTypeAgenticModule
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid type '%s'. Use 'extension' or 'agentic-module'\n", os.Args[2])
		os.Exit(1)
	}

	id := os.Args[3]

	if err := packaging.Remove(id, itemType); err != nil {
		fmt.Fprintf(os.Stderr, "Error removing: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully removed %s '%s'\n", itemType, id)
}

func handleEnable() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Error: enable requires type and id\n")
		os.Exit(1)
	}

	var itemType registry.ItemType
	switch os.Args[2] {
	case "extension", "ext":
		itemType = registry.ItemTypeExtension
	case "agentic-module", "module", "am":
		itemType = registry.ItemTypeAgenticModule
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid type '%s'. Use 'extension' or 'agentic-module'\n", os.Args[2])
		os.Exit(1)
	}

	id := os.Args[3]

	reg, err := registry.NewRegistry()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := reg.SetEnabled(id, true); err != nil {
		fmt.Fprintf(os.Stderr, "Error enabling: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully enabled %s '%s'\n", itemType, id)
	fmt.Printf("  Note: Changes take effect on next server restart\n")
}

func handleDisable() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Error: disable requires type and id\n")
		os.Exit(1)
	}

	var itemType registry.ItemType
	switch os.Args[2] {
	case "extension", "ext":
		itemType = registry.ItemTypeExtension
	case "agentic-module", "module", "am":
		itemType = registry.ItemTypeAgenticModule
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid type '%s'. Use 'extension' or 'agentic-module'\n", os.Args[2])
		os.Exit(1)
	}

	id := os.Args[3]

	reg, err := registry.NewRegistry()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := reg.SetEnabled(id, false); err != nil {
		fmt.Fprintf(os.Stderr, "Error disabling: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully disabled %s '%s'\n", itemType, id)
	fmt.Printf("  Note: Changes take effect on next server restart\n")
}

func handleMigrate() {
	legacyDir := "extensions"
	if len(os.Args) >= 3 {
		legacyDir = os.Args[2]
	}

	fmt.Printf("Migration: Scanning for legacy extensions in %s...\n", legacyDir)

	result, err := migration.MigrateLegacyExtensions(legacyDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during migration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nMigration complete:\n")
	fmt.Printf("  Migrated extensions: %d\n", result.MigratedExtensions)
	fmt.Printf("  Migrated modules: %d\n", result.MigratedModules)
	if len(result.Failed) > 0 {
		fmt.Printf("  Failed: %d\n", len(result.Failed))
		for _, failure := range result.Failed {
			fmt.Printf("    - %s\n", failure)
		}
	}
}

func handleSync() {
	reg, err := registry.NewRegistry()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := startup.SyncRegistry(reg); err != nil {
		fmt.Fprintf(os.Stderr, "Error syncing registry: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Registry synchronized with filesystem")
}
