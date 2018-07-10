package commands

import (
	"fmt"
	"strings"

	"github.com/rigelrozanski/common"
	wb "github.com/rigelrozanski/wb/lib"
	"github.com/spf13/cobra"
)

// ripper bud
var RipCmd = &cobra.Command{
	Use: "rip",
	Run: ripCmd,
}

func init() {
	RootCmd.AddCommand(RipCmd)
}

func ripCmd(cmd *cobra.Command, args []string) {
	urls, found := wb.GetWB("rip")
	if !found {
		fmt.Println("can't find wb rip")
		return
	}
	for _, url := range urls {

		flags := ""
		if strings.Contains(url, "bandcamp") {
			flags += "-b "
		}

		fmt.Printf("rippin' %v\n", url)
		out, err := common.Execute(fmt.Sprintf("soundscrape %v%v", flags, url))
		fmt.Printf("result/err: %v/%v\n", out, err)
	}
}
