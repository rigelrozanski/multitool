package commands

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

// Lock2yamlCmd represents the lock2yaml command
var (
	SlackCleanupCmd = &cobra.Command{
		Use:   "slack [names]",
		Short: "From/to clipboard - slack text, delete extra lines and timestamps",
		RunE:  slackCleanupCmd,
	}
)

func init() {
	RootCmd.AddCommand(SlackCleanupCmd)
}

func slackCleanupCmd(cmd *cobra.Command, args []string) error {

	// get the text
	text, err := clipboard.ReadAll()
	if err != nil {
		return err
	}

	//delete timestamps
	re := regexp.MustCompile(`\[(\d|\d\d)\:\d\d(|\s[A-Z]{2})\]`)
	text = re.ReplaceAllString(text, "")

	//delete blank lines
	re = regexp.MustCompile(`(?m)^\s*$[\r\n]*|[\r\n]+\s+\z`)
	text = re.ReplaceAllString(text, "")

	//delete all @ symbols
	text = strings.Replace(text, "@", "", -1)

	//Add a space for all names, make the name bold
	for _, arg := range args {
		text = strings.Replace(text, "\n"+arg+"\n", "\n\n**"+arg+"**\n", -1)
		text = strings.Replace(text, "\n"+arg+" \n", "\n\n**"+arg+"**\n", -1)
		text = strings.Replace(text, "\n"+arg+"  \n", "\n\n**"+arg+"**\n", -1)
	}

	err = clipboard.WriteAll(text)
	if err != nil {
		return err
	}

	fmt.Println("Processed successfully")
	return nil
}
