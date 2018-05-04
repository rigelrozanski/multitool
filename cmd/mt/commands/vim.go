package commands

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/rigelrozanski/common"
	"github.com/spf13/cobra"
)

// Lock2yamlCmd represents the lock2yaml command
var (
	VIMCmd = &cobra.Command{
		Use:   "vim",
		Short: "tools intended to be called from vim",
	}
	CreateTestCmd = &cobra.Command{
		Use:   "create-test [name] [source-file]",
		Short: "create a go-lang test shell in the correct file",
		Args:  cobra.ExactArgs(2),
		RunE:  createTestCmd,
	}
)

func init() {
	VIMCmd.AddCommand(CreateTestCmd)
	RootCmd.AddCommand(VIMCmd)
}

func createTestCmd(cmd *cobra.Command, args []string) error {

	fnName := args[0]
	sourceFile := args[1]

	// construct the test file
	base := filepath.Base(sourceFile)
	name := strings.Split(base, ".")[0] + "_test.go"
	dir := filepath.Dir(sourceFile)
	testFile := path.Join(dir, name)

	testFnStr := fmt.Sprintf("\nfunc Test%v(t *testing.T) { \n\n}", fnName)

	var lines []string
	if common.FileExists(testFile) {
		var err error
		lines, err = common.ReadLines(testFile)
		if err != nil {
			return err
		}
		lines = append(lines, testFnStr)
	} else {
		sourceLines, err := common.ReadLines(sourceFile)
		if err != nil {
			return err
		}
		lines = []string{
			sourceLines[0], //package
			"\nimport \"testing\"",
			testFnStr,
		}
	}

	fmt.Println(testFile)
	return common.WriteLines(lines, testFile)
}
