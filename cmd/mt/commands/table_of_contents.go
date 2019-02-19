package commands

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/rigelrozanski/common"
	"github.com/spf13/cobra"
)

// file commands
var (
	TocCmd = &cobra.Command{
		Use:   "toc [optional-folder]",
		Short: "generate a markdown table of contents",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "./"
			if len(args) == 1 {
				dir = args[0]
			}
			files, err := ioutil.ReadDir(dir)
			if err != nil {
				return err
			}
			tocNumber := 1
			for _, f := range files {
				if f.IsDir() {
					continue
				}
				name := f.Name()
				if string(name[0]) == "." {
					continue
				}
				if strings.ToLower(name) == "readme.md" {
					continue
				}

				p := path.Join(dir, name)
				if strings.ToLower(path.Ext(p)) != ".md" {
					continue
				}
				mdContents, err := common.ReadLines(p)
				if err != nil {
					return err
				}

				parsed := parseHeaders(mdContents)

				// print title
				fileTitle := name
				if len(parsed[1]) == 1 {
					fileTitle = parsed[1][0]
				}
				fmt.Printf("%v. **[%v](%v)**\n", tocNumber, fileTitle, p)
				tocNumber++

				for _, subTitle := range parsed[2] {
					linking := strings.ToLower(subTitle)
					linking = strings.Replace(linking, " ", "-", -1)
					fmt.Printf("    - [%v](%v#%v)\n", subTitle, p, linking)
				}
			}
			return nil
		},
	}
)

func init() {
	RootCmd.AddCommand(TocCmd)
}

// TODO move to rigelrozanski/common
// TODO make suborders
func parseHeaders(markdown []string) map[int][]string {
	headers := make(map[int][]string)
	for _, line := range markdown {
		for level := 1; level <= 6; level++ {

			headerPrefix := strings.Repeat("#", level) + " "

			if strings.HasPrefix(line, headerPrefix) {
				withoutPrefix := strings.TrimLeft(line, headerPrefix)
				headers[level] = append(headers[level], withoutPrefix)
			}
		}
	}
	return headers
}
