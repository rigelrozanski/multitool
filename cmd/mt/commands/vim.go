package commands

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rigelrozanski/common"
	"github.com/spf13/cobra"
)

// tools intended to be used with vim
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
	DebugPrintsCmd = &cobra.Command{
		Use:   "debug-prints [name] [source-file] [lineno]",
		Short: "add prints to all the possible end points within a function",
		Args:  cobra.ExactArgs(3),
		RunE:  debugPrintsCmd,
	}
	RemoveDebugPrintsCmd = &cobra.Command{
		Use:   "remove-debug-prints [source-file] [lineno]",
		Short: "remove debug prints",
		Args:  cobra.ExactArgs(2),
		RunE:  removeDebugPrintsCmd,
	}
	ColumnSpacesCmd = &cobra.Command{
		Use:   "column-width [source-file] [lineno-start] [lineno-end] [column-characters]",
		Short: "add spaces up to the column specified",
		Args:  cobra.ExactArgs(4),
		RunE:  columnSpacesCmd,
	}
)

func init() {
	VIMCmd.AddCommand(CreateTestCmd)
	VIMCmd.AddCommand(DebugPrintsCmd)
	VIMCmd.AddCommand(RemoveDebugPrintsCmd)
	VIMCmd.AddCommand(ColumnSpacesCmd)
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

func debugPrintsCmd(cmd *cobra.Command, args []string) error {

	name := args[0]
	sourceFile := args[1]
	startLineNo, err := strconv.Atoi(args[2])
	if err != nil {
		return err
	}

	if !common.FileExists(sourceFile) {
		return errors.New("file don't exist")
	}

	lines, err := common.ReadLines(sourceFile)
	if err != nil {
		return err
	}
	lines = insertPrints(lines, startLineNo, name)

	err = common.WriteLines(lines, sourceFile)
	if err != nil {
		return err
	}
	return nil
}

func insertPrints(lines []string, startLineNo int, name string) []string {

	debugNo := 0
	for i := startLineNo; i < len(lines); i++ {
		line := lines[i]
		if len(line) == 0 { // skip blank lines
			continue
		}
		if strings.HasPrefix(line, "}") { // reached the end of the function
			break
		}

		if strings.Contains(line, "}") || strings.Contains(line, "{") { // reached the end of the function

			outputStr := fmt.Sprintf("fmt.Println(\"wackydebugoutput %v %v\")", name, debugNo)
			debugNo++

			// insert the line
			lines = append(lines[:i+1], append([]string{outputStr}, lines[i+1:]...)...)
			i++ // skip a line
			continue
		}
	}

	return lines
}

func removeDebugPrintsCmd(cmd *cobra.Command, args []string) error {

	sourceFile := args[0]
	startLineNo, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}

	if !common.FileExists(sourceFile) {
		return errors.New("file don't exist")
	}

	lines, err := common.ReadLines(sourceFile)
	if err != nil {
		return err
	}
	lines = removePrints(lines, startLineNo)

	err = common.WriteLines(lines, sourceFile)
	if err != nil {
		return err
	}
	return nil
}

func removePrints(lines []string, startLineNo int) []string {
	for i := startLineNo; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "}") { // reached the end of the function
			break
		}

		// remove the line
		if strings.Contains(line, "wackydebugoutput") {
			lines = append(lines[:i], lines[i+1:]...)
		}
	}

	return lines
}

func columnSpacesCmd(cmd *cobra.Command, args []string) error {

	sourceFile := args[0]
	startLineNo, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}
	endLineNo, err := strconv.Atoi(args[2])
	if err != nil {
		return err
	}
	columnChars, err := strconv.Atoi(args[3])
	if err != nil {
		return err
	}

	if !common.FileExists(sourceFile) {
		return errors.New("file don't exist")
	}

	lines, err := common.ReadLines(sourceFile)
	if err != nil {
		return err
	}
	for i := startLineNo; i <= endLineNo; i++ {
		line := lines[i]
		len := len(line)
		if len >= columnChars {
			continue
		}
		for j := len; j <= columnChars; j++ {
			line += " "
		}
		lines[i] = line
	}

	err = common.WriteLines(lines, sourceFile)
	if err != nil {
		return err
	}
	return nil
}
