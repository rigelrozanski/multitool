package commands

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/knetic/govaluate"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "mt [expr]",
	Short: "multitool, a collection of handy lil tools",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		// First attempt to calculate like math
		var mathErr error
		argsJoined := strings.Join(args, "")
		amountExpr, err := govaluate.NewEvaluableExpression(argsJoined)
		if err != nil {
			mathErr = err
		}
		if mathErr == nil {
			amountI, err := amountExpr.Evaluate(map[string]interface{}{})
			if err != nil {
				mathErr = err
			}
			if mathErr == nil {
				fmt.Println(amountI.(float64))
				return nil
			}
		}

		// Next attempt a conversion operation if correct args
		if len(args) >= 4 && len(args) <= 6 {
			err := convertCmd(nil, args)
			if err == nil {
				return nil
			}
		}

		return errors.New("could not resolve command")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config",
		"", "config file (default is $HOME/.multitool.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".multitool" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".multitool")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
