package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/scriptogre/op-dotenv/internal"
	"github.com/urfave/cli/v3"
)

// Build information set by GoReleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd := &cli.Command{
		Name:                  "op-dotenv",
		Usage:                 "Convert .env files to 1Password items and vice versa",
		Description:           "Push .env files to 1Password items or pull 1Password items to .env files. Supports sections via comments.",
		Version:               version,
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "vault",
				Aliases: []string{"v"},
				Usage:   "Override vault name (will prompt if not specified)",
			},
			&cli.StringFlag{
				Name:    "item",
				Aliases: []string{"i"},
				Usage:   "Override item name (defaults to current directory name)",
			},
		},
		Commands: []*cli.Command{
			{
				Name:        "push",
				Usage:       "Convert .env file to 1Password item",
				Description: "Create or update a 1Password item from a local .env file. Comments become sections.",
				ArgsUsage:   "[env-file]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Usage:   "Force overwrite without confirmation",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					// Determine file path
					filePath := ".env"
					if cmd.NArg() > 0 {
						filePath = cmd.Args().Get(0)
					}

					// Create app and execute push
					app, err := internal.NewApp()
					if err != nil {
						return err
					}

					vault := cmd.String("vault")
					item := cmd.String("item")
					force := cmd.Bool("force")

					return app.Push(filePath, vault, item, force)
				},
			},
			{
				Name:        "pull",
				Usage:       "Convert 1Password item to .env file",
				Description: "Create or update a local .env file from a 1Password item. Sections become comments.",
				ArgsUsage:   "[env-file]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Usage:   "Force overwrite without confirmation",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					// Determine file path
					filePath := ".env"
					if cmd.NArg() > 0 {
						filePath = cmd.Args().Get(0)
					}

					// Create app and execute pull
					app, err := internal.NewApp()
					if err != nil {
						return err
					}

					vault := cmd.String("vault")
					item := cmd.String("item")

					return app.Pull(filePath, vault, item)
				},
			},
			{
				Name:        "config",
				Usage:       "Show current configuration",
				Description: "Display the current vault and item configuration for this directory",
				Aliases:     []string{"cfg"},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					config, err := internal.LoadConfig()
					if err != nil {
						return fmt.Errorf("failed to load config: %w", err)
					}

					workingDir, _ := os.Getwd()
					
					// Check if this project has any stored configuration
					if _, exists := config.Projects[workingDir]; exists {
						fmt.Printf("Current configuration for %s:\n", workingDir)
						fmt.Printf("  Vault: %s\n", config.GetVault(workingDir, "Environments"))
						fmt.Printf("  Item:  %s\n", config.GetItem(workingDir, filepath.Base(workingDir)))
					} else {
						fmt.Printf("No configuration found for %s.\n", workingDir)
						fmt.Printf("Default values will be used:\n")
						fmt.Printf("  Vault: %s\n", "Environments")
						fmt.Printf("  Item:  %s\n", filepath.Base(workingDir))
					}

					return nil
				},
			},
			{
				Name:        "clean",
				Usage:       "Remove all configuration data",
				Description: "Delete the configuration file and all stored preferences",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					app, err := internal.NewApp()
					if err != nil {
						return err
					}

					return app.Clean()
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
