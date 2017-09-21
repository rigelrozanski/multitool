package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/rigelrozanski/common"
)

// GenTestCmd represents the gentest command
var (
	GenTestCmd = &cobra.Command{
		Use:   "gentest [filepath]",
		Short: "generate a basic test file provided",
		RunE:  genTestCmd,
	}

	//flags
	flagUnexposed = "unexposed"
)

func init() {
	GenTestCmd.Flags().BoolP(flagUnexposed, "u", false,
		"also generate function endpoints for unexposed variables")
	viper.BindPFlag(flagUnexposed, GenTestCmd.Flags().Lookup(flagUnexposed))
	RootCmd.AddCommand(GenTestCmd)
}

func genTestCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("Please specify the filepath")
	}

	//get files
	yaml, err := common.ReadLines("glide.yaml")
	if err != nil {
		return err
	}
	lock, err := common.ReadLines(viper.GetString(flagUnexposed))
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
