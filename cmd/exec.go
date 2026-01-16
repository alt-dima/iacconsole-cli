package cmd

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/alt-dima/iacconsole-cli/utils"
	"github.com/spf13/cobra"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute OpenTofu commands within a synthesized environment",
	Args:  cobra.MinimumNArgs(1),
	Long:  `Execute OpenTofu commands within a synthesized environment from inventory and parameters after --`,
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		//Creating signal to be handled and send to the child tofu/terraform
		sigs := make(chan os.Signal, 2)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		var err error

		// Creating Session State and filling with values
		s := &utils.State{}

		IACCONSOLE_API_URL := os.Getenv("IACCONSOLE_API_URL")
		if IACCONSOLE_API_URL != "" {
			// validate URL format and remove trailing slash if present
			if strings.HasSuffix(IACCONSOLE_API_URL, "/") {
				IACCONSOLE_API_URL = strings.TrimRight(IACCONSOLE_API_URL, "/")
			}

			// Basic validation for IACCONSOLE_API_URL format
			if !strings.HasPrefix(IACCONSOLE_API_URL, "https://") {
				log.Fatalf("Error: IACCONSOLE_API_URL must start with https://")
			}

			// Check if URL contains credentials and correct domain
			urlParts := strings.Split(strings.TrimPrefix(IACCONSOLE_API_URL, "https://"), "@")
			if len(urlParts) != 2 || urlParts[1] != "api.iacconsole.com" {
				log.Fatalf("Error: IACCONSOLE_API_URL must be in format https://ACCOUNTID:PASSWORD@api.iacconsole.com")
			}

			// Validate credential part has both account ID and password
			credParts := strings.Split(urlParts[0], ":")
			if len(credParts) != 2 || credParts[0] == "" || credParts[1] == "" {
				log.Fatalf("Error: IACCONSOLE_API_URL credentials must include both ACCOUNTID and PASSWORD")
			}
		}

		s.UnitName, _ = cmd.Flags().GetString("unit")
		s.OrgName, _ = cmd.Flags().GetString("org")
		s.Workspace, _ = cmd.Flags().GetString("workspace")
		s.IacconsoleApiUrl = IACCONSOLE_API_URL
		s.DimensionsFlags, _ = cmd.Flags().GetStringSlice("dimension")
		s.UnitPath, _ = filepath.Abs(s.GetStringFromViperByOrgOrDefault("units_path") + "/" + s.OrgName + "/" + s.UnitName)
		if s.GetStringFromViperByOrgOrDefault("shared_modules_path") != "" {
			s.SharedModulesPath, _ = filepath.Abs(s.GetStringFromViperByOrgOrDefault("shared_modules_path"))
		}
		if s.GetStringFromViperByOrgOrDefault("inventory_path") != "" {
			s.InventoryPath, _ = filepath.Abs(s.GetStringFromViperByOrgOrDefault("inventory_path") + "/" + s.OrgName)
		}

		s.ParseUnitManifest("unit_manifest.json")
		s.ParseDimensions()

		backendiacconsoleConfig := s.SetupBackendConfig()

		s.PrepareTemp()

		s.GenerateVarsByDims()
		s.GenerateVarsByDimOptional("defaults")
		s.GenerateVarsByEnvVars()
		s.GenerateVarsByDimAndData("config", "backend", backendiacconsoleConfig)

		//Local variables for child execution
		forceCleanTempDir, _ := cmd.Flags().GetBool("clean")
		var backendConfig []string
		for param, value := range backendiacconsoleConfig {
			backendConfig = append(backendConfig, "-backend-config="+param+"="+value.(string))
		}
		cmdArgs := args
		if args[0] == "init" {
			cmdArgs = append(cmdArgs, backendConfig...)
		}
		cmdToExec := s.GetStringFromViperByOrgOrDefault("cmd_to_exec")

		// Starting child and Waiting for it to finish, passing signals to it
		log.Println("excuting: " + cmdToExec + " " + strings.Join(cmdArgs, " "))
		execChildCommand := exec.Command(cmdToExec, cmdArgs...)
		execChildCommand.Dir = s.CmdWorkTempDir
		execChildCommand.Env = os.Environ()
		execChildCommand.Stdin = os.Stdin
		execChildCommand.Stdout = os.Stdout
		execChildCommand.Stderr = os.Stderr
		err = execChildCommand.Start()
		if err != nil {
			log.Fatalf("cmd.Start() failed with %s\n", err)
		}

		go func() {
			sig := <-sigs
			log.Println("Got singnal +" + sig.String())
			if err := execChildCommand.Process.Signal(sig); err != nil {
				log.Printf("Failed to send signal to child process: %v", err)
			}
		}()

		err = execChildCommand.Wait()
		exitCodeFinal := 0
		if err != nil && execChildCommand.ProcessState.ExitCode() < 0 {
			exitCodeFinal = 1
			log.Println(cmdToExec + " failed " + err.Error())
		} else if execChildCommand.ProcessState.ExitCode() == 143 {
			exitCodeFinal = 0
		} else {
			exitCodeFinal = execChildCommand.ProcessState.ExitCode()
		}

		if (exitCodeFinal == 0 && (args[0] == "apply" || args[0] == "destroy")) || forceCleanTempDir {
			os.RemoveAll(s.CmdWorkTempDir)
			log.Println("removed temp dir: " + s.CmdWorkTempDir)
		}

		log.Printf("%v finished with code %v", cmdToExec, exitCodeFinal)
		os.Exit(exitCodeFinal)
	},
}

func init() {
	rootCmd.AddCommand(execCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// execCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	execCmd.Flags().StringSliceP("dimension", "d", []string{}, "specify dimensions from invetory like dim:name")
	//viper.BindPFlag("account", execCmd.Flags().Lookup("account"))
	execCmd.Flags().StringP("unit", "u", "", "specify unit")
	//viper.BindPFlag("unit", execCmd.Flags().Lookup("unit"))
	execCmd.Flags().StringP("org", "o", "", "specify org")
	execCmd.Flags().StringP("workspace", "w", "master", "specify workspace for IaCConsole DB")
	execCmd.Flags().BoolP("clean", "c", false, "remove tmp after execution")
	//viper.BindPFlag("org", execCmd.Flags().Lookup("org"))
	if err := execCmd.MarkFlagRequired("unit"); err != nil {
		log.Fatalf("Error marking flag 'unit' as required: %v", err)
	}
	if err := execCmd.MarkFlagRequired("org"); err != nil {
		log.Fatalf("Error marking flag 'org' as required: %v", err)
	}
}
