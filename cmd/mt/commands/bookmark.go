package commands

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/rigelrozanski/common"
	"github.com/rigelrozanski/thranch/quac"
)

// speedy todolists
var (
	Bookmark = &cobra.Command{
		Use:   "bm [url] <tag1,tag2,tag3,...>",
		Short: "bookmark the url with the following tags",
		Args:  cobra.RangeArgs(1, 2),
		RunE:  BookmarksCmd,
	}
)

func init() {
	RootCmd.AddCommand(Bookmark)
}

// TODO auto search website for keywords
// display these keywords before commiting them in here,
// user must approve
func BookmarksCmd(cmd *cobra.Command, args []string) error {

	title, err := common.GetUrlTitle(args[0])
	if err != nil {
		return err
	}
	aline := title

	aline += " " + args[0]
	if len(args) == 2 {
		aline += " " + args[1]
	}

	quac.Initialize(os.ExpandEnv("$HOME/.thranch_config"))
	err = quac.AppendLineForApp("mt-bookmarks", aline)
	if err != nil {
		return err
	}

	return nil
}
