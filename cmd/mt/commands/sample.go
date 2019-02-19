package commands

import (
	"fmt"
	"strings"

	"github.com/rigelrozanski/common"
	wb "github.com/rigelrozanski/wb/lib"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(SampleCmd)
}

// sample
var SampleCmd = &cobra.Command{
	Use: "sample",
	Run: func(cmd *cobra.Command, args []string) {

		urls, found := wb.GetWB("sample")
		if !found {
			fmt.Println("can't find wb sample")
			return
		}
		for _, url := range urls {

			flags := ""
			if strings.Contains(url, "bandcamp") {
				flags += "-b "
			}

			fmt.Printf("sampling' %v - ", url)
			out, err := common.Execute(fmt.Sprintf("soundscrape %v%v", flags, url))
			if err != nil {
				fmt.Printf("\terr: %v\n", err)
			} else {
				fmt.Printf("\tresult: %v\n", out)
			}
		}
	},
}
