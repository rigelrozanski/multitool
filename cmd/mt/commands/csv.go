package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/rigelrozanski/common"
	"github.com/spf13/cobra"
)

// Lock2yamlCmd represents the lock2yaml command
var (
	CSVCmd = &cobra.Command{
		Use:   "csv",
		Short: "csv processing commands",
	}
	LastColOnlyCmd = &cobra.Command{
		Use:   "last-col-only [read-file] [write-file]",
		Short: "make each line only the last column",
		Args:  cobra.ExactArgs(2),
		RunE:  lastColOnlyCmd,
	}
)

func init() {
	CSVCmd.AddCommand(LastColOnlyCmd)
	RootCmd.AddCommand(CSVCmd)
}

func lastColOnlyCmd(cmd *cobra.Command, args []string) error {

	newFilePath := args[1]
	var writeLines []string

	file, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		split := strings.Split(line, ",")
		last := split[len(split)-1]
		writeLines = append(writeLines, last)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if err := common.WriteLines(writeLines, newFilePath); err != nil {
		return err
	}

	fmt.Println("completed")
	return nil
}
