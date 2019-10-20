/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const mainPageHelp = ` 
Welcome to AFX CLI!
---------------------
Use 'afx --help' to see available commands or go to https://aka.ms/cli.

Telemetry
---------
The AFX CLI collects usage data in order to improve your experience.
The data is anonymous and does not include commandline argument values.
The data is collected by 'Some Watcher'.

You can change your telemetry settings with 'afx configure.

____________________  __
___    |__  ____/_  |/ /
__  /| |_  /_   __    / 
_  ___ |  __/   _    |  
/_/  |_/_/      /_/|_|  
                      
Welcome to the cool new AFX CLI!

Use 'afx --version' to display the current version.
Here are the base commands:
`
const afxArt = `
____________________  __
___    |__  ____/_  |/ /
__  /| |_  /_   __    / 
_  ___ |  __/   _    |  
/_/  |_/_/      /_/|_|  
`

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "afx",
	Short: "A brief description of your application",
	Long:  afxArt,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(mainPageHelp)

		for _, item := range cmd.Commands() {
			fmt.Printf("   %-16s: %s\n", item.Use, item.Short)
		}
		fmt.Println("")
		/*
				configure         : Manage AFX CLI configuration. This command is interactive.
			  login             : Log in to AFX.

		*/

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.newApp.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("version", "v", false, "Display server and cli version.")
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

		// Search config in home directory with name ".newApp" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".newApp")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
