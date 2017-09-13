package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/rigelrozanski/common"
)

// Lock2yamlCmd represents the lock2yaml command
var (
	Lock2YamlCmd = &cobra.Command{
		Use:   "lock2yaml [import-root]",
		Short: "Set the glide yaml version to the hash from the lock file",
		RunE:  lock2YamlCmd,
	}

	//flags
	flagLockSrc = "lockSrc"
)

func init() {
	Lock2YamlCmd.Flags().StringP(flagLockSrc, "l", "glide.lock",
		"read an external .lock file besides the one in the working directory")
	viper.BindPFlag(flagLockSrc, Lock2YamlCmd.Flags().Lookup(flagLockSrc))
	RootCmd.AddCommand(Lock2YamlCmd)
}

func lock2YamlCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("Please specify the import root")
	}

	//get files
	yaml, err := common.ReadLines("glide.yaml")
	if err != nil {
		return err
	}
	lock, err := common.ReadLines(viper.GetString(flagLockSrc))
	//lock, err := common.ReadLines("glide.lock")
	if err != nil {
		return err
	}

	for i := 0; i < len(args); i++ {
		importRoot := args[i]
		for j := 0; j < len(yaml); j++ {
			if strings.Contains(yaml[j], importRoot) && j+1 < len(yaml) {
				if strings.Contains(yaml[j+1], "version") {
					importName, err := getPostColon(yaml[j])
					if err != nil {
						return err
					}
					lockHash, err := getLockHash(lock, importName)
					if err != nil {
						return err
					}
					yaml[j+1] = "  version: " + lockHash
				}
			}
		}
	}
	return common.WriteLines(yaml, "glide.yaml")
}

func getLockHash(lock []string, importName string) (string, error) {
	for i, line := range lock {
		if strings.Contains(line, importName) && i+1 < len(lock) {
			if strings.Contains(lock[i+1], "version") {
				return getPostColon(lock[i+1])
			}
		}
	}
	return "", fmt.Errorf("Could not determine lock hash for %v", importName)
}

func getPostColon(line string) (string, error) {
	s := strings.Split(line, ": ")
	if len(s) < 2 {
		return "", fmt.Errorf("Problem retrieving value for line: \n\v", line)
	}
	return s[1], nil
}
