package cmd

import "github.com/jiajia556/chainkit/pkg/utils"

// Execute is the CLI entrypoint.
func Execute() {
	if err := GetRootCmd().Execute(); err != nil {
		utils.OutputFatal(err)
	}
}
