package main

import (
	"./buildcache"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/Sirupsen/logrus"
)

func main() {
	if err := mainCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

var (
	mainCmd = &cobra.Command{
		Use:          os.Args[0],
		Short:        "Make docker build great again!",
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			logrus.SetOutput(os.Stderr)
			flag, err := cmd.Flags().GetString("log-level")
			if err != nil {
				logrus.Fatal(err)
			}
			level, err := logrus.ParseLevel(flag)
			if err != nil {
				logrus.Fatal(err)
			}
			logrus.SetLevel(level)
			logrus.SetOutput(os.Stdout)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			registryAddr, err := cmd.Flags().GetString("registry-addr")
			if err != nil {
				return err
			}

			pushImage, _ := cmd.Flags().GetString("push")
			pullImage, _ := cmd.Flags().GetString("pull")
			if pushImage == "" && pullImage == "" {
				return fmt.Errorf("push flag or pull flag is needed")
			}

			if pushImage != "" && pullImage != "" {
				return fmt.Errorf("can't run push and pull at the same time")
			}

			if pushImage != "" {
				return buildcache.Push(pushImage, registryAddr)
			}

			return nil
		},
	}
)

func init() {
	mainCmd.Flags().String("registry-addr", "localhost:5000", "Address of docker registry, default is localhost:5000.")
	mainCmd.Flags().StringP("log-level", "l", "debug", "Log level (options \"debug\", \"info\", \"warn\", \"error\", \"fatal\", \"panic\")")
	mainCmd.Flags().String("push", "", "Push image with cache")
	mainCmd.Flags().String("pull", "", "Pull image with cache")
}
