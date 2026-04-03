package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the version of the built application",
	Long:  `Shows the version of the built application, based on Go's internal build primitives.`,
	Run: func(cmd *cobra.Command, args []string) {
		version := "dev"
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" {
					version = setting.Value
					break
				}
			}
			if info.Main.Version != "(devel)" && info.Main.Version != "" {
				version = info.Main.Version
			}
		}
		fmt.Printf("%s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
