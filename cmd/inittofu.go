package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/alt-dima/iacconsole-cli/utils"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new tofugu working directory",
	Long:  `Create a new tofugu working directory with default structure and .tofugu config file`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get target directory, default to current directory
		targetDir, _ := cmd.Flags().GetString("target-dir")
		if targetDir == "" {
			var err error
			targetDir, err = os.Getwd()
			if err != nil {
				log.Fatalf("Failed to get current directory: %v", err)
			}
		}

		tofuguConfigPath := filepath.Join(targetDir, ".tofugu")
		if _, err := os.Stat(tofuguConfigPath); err == nil {
			// File exists
			overwrite, _ := cmd.Flags().GetBool("force")
			if !overwrite {
				log.Fatalf(".tofugu file already exists. Use --force to overwrite")
			}
			log.Printf("Overwriting existing .tofugu file")
		}

		utils.CreateExampleTofuguConfigFile(tofuguConfigPath)

		// Create example organization structure if requested
		createExample, _ := cmd.Flags().GetBool("with-example")
		useToasterDB, _ := cmd.Flags().GetBool("toaster")
		var exampleCmd string
		if createExample {
			exampleCmd = utils.CreateExampleStructure(targetDir, useToasterDB)
		}

		fmt.Println("\nTofugu workspace initialized successfully!")
		fmt.Println("\nTo use this workspace:")
		fmt.Println("1. Make sure you have OpenTofu installed: https://opentofu.org/docs/intro/install/")
		fmt.Printf("2. Navigate to: %s\n", targetDir)
		fmt.Println("3. Run a command like:")
		if createExample {
			fmt.Print(exampleCmd)
		} else {
			fmt.Println("   tofugu cook -o your-org -d dimension:value -t tofi -- init")
			fmt.Println("   tofugu cook -o your-org -d dimension:value -t tofi -- plan")
			fmt.Println("   tofugu cook -o your-org -d dimension:value -t tofi -- apply")
		}
		fmt.Println("\nYou can customize the .tofugu file for your specific needs. Add S3 backend configuration if required.")
		fmt.Println("Do not to forget to try Infrastructure layers/dimensions configurations Storage: https://toaster.altuhov.su/")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Define flags for the init command
	initCmd.Flags().StringP("target-dir", "d", "", "Target directory for the new tofugu workspace (defaults to current directory)")
	initCmd.Flags().BoolP("toaster", "t", true, "Use ToasterDB toaster.altuhov.su for inventory data")
	initCmd.Flags().BoolP("force", "f", false, "Force overwrite of existing .tofugu file")
	initCmd.Flags().BoolP("with-example", "e", true, "Create an example organization structure with sample files")
}
