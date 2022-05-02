package commands

import (
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(HtmlServeCmd)
}

// sample
var HtmlServeCmd = &cobra.Command{
	Use:     "html-serve",
	Aliases: []string{"srv", "serve"},
	Run: func(cmd *cobra.Command, args []string) {
		http.Handle("/", http.FileServer(http.Dir(".")))
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	},
}
