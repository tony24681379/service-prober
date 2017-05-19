// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tony24681379/k8s-prober/prober"
)

type options struct {
	Debug  bool
	Config string
	Port   string
}

var opts = options{}

// RootCmd represents the base command when called without any subcommands
func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "k8s-prober",
		Short: "Kubernetes liveness and readiness probe tool",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if opts.Debug == true {
				logrus.SetLevel(logrus.DebugLevel)
				fmt.Println("debug")
			}
		}, Run: func(cmd *cobra.Command, args []string) {
			if opts.Config == "" {
				cmd.Help()
			} else {
				runProber(opts)
			}
		},
	}
	return cmd
}

// Execute adds all child commands to the root command sets flags appropriately.
func execute(rootCmd *cobra.Command) {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

//InitCmd init cobra command
func InitCmd() {
	rootCmd := newRootCmd()
	initProgramFlag(rootCmd)
	execute(rootCmd)
}

func initProgramFlag(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().BoolVarP(&opts.Debug, "debug", "D", false, "Enable debug mode")
	flags := rootCmd.Flags()
	flags.StringVar(&opts.Config, "config", "", "config file")
	flags.StringVar(&opts.Port, "port", "10000", "serve port")
}

func runProber(opts options) {
	err := prober.Prober(opts.Config, opts.Port)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}
