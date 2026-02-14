package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/otiai10/copy"
)

func (s *State) PrepareTemp() error {
	if s.StateS3Path == "" {
		return fmt.Errorf("StateS3Path is empty")
	}

	tmpFolderNameSuffix := s.OrgName + s.StateS3Path + s.UnitName
	cmdTempDirFullPath := os.TempDir() + "/iacconsole-" + GetMD5Hash(tmpFolderNameSuffix)

	// Create temp directory if it doesn't exist
	if err := os.MkdirAll(cmdTempDirFullPath, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}

	// Copy options to exclude certain files/directories
	opt := copy.Options{
		Skip: func(info os.FileInfo, src string, dest string) (bool, error) {
			base := filepath.Base(src)
			return base == ".terraform" || base == "unit_manifest.json", nil
		},
	}

	// Copy the unit directory to temp directory
	if err := copy.Copy(s.UnitPath, cmdTempDirFullPath, opt); err != nil {
		os.RemoveAll(cmdTempDirFullPath)
		return fmt.Errorf("failed to copy unit to tempdir: %v", err)
	}

	if s.SharedModulesPath != "" {
		// Remove existing symlink if it exists
		sharedModulesLink := filepath.Join(cmdTempDirFullPath, "shared-modules")
		os.Remove(sharedModulesLink) // Ignore error as file might not exist

		// Create new symlink
		if err := os.Symlink(s.SharedModulesPath, sharedModulesLink); err != nil {
			os.RemoveAll(cmdTempDirFullPath)
			return fmt.Errorf("failed to create symlink for shared_modules: %v", err)
		}
		log.Println("iacconsole symlinked shared_modules to tempdir : " + s.SharedModulesPath)
	}

	s.CmdWorkTempDir = cmdTempDirFullPath
	log.Println("iacconsole prepared unit in temp dir: " + s.CmdWorkTempDir)
	return nil
}
