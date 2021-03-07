package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rigelrozanski/common"
	"github.com/rigelrozanski/thranch/quac"
)

func init() {
	RootCmd.AddCommand(SampleCmd)
}

// sample
var SampleCmd = &cobra.Command{
	Use: "sample",
	Run: func(cmd *cobra.Command, args []string) {

		quac.Initialize(os.ExpandEnv("$HOME/.thranch_config"))
		urlsClumped := quac.GetForApp("sample")
		urls := strings.Split(urlsClumped, "\n")

		for _, url := range urls {

			fmt.Printf("sampling' %v - ", url)
			out, err := common.Execute(fmt.Sprintf("scdl -l %v", url))
			if err != nil {
				fmt.Printf("\terr: %v\n", err)
			} else {
				fmt.Printf("\tresult: %v\n", out)
			}
		}
	},
}
