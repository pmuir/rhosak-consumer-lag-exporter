/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
	"os"
	"rhosak-consumer-lag-exporter/pkg/exporters"
	"strings"

	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export the consumer lag",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		clientID, _ := cmd.Flags().GetString("clientId")
		if clientID == "" {
			clientID = os.Getenv("CLIENT_ID")
		}
		clientSecret, _ := cmd.Flags().GetString("clientSecret")
		if clientSecret == "" {
			clientSecret = os.Getenv("CLIENT_SECRET")
		}
		tokenURL, _ := cmd.Flags().GetString("tokenURL")
		bootstrapServer, _ := cmd.Flags().GetStringArray("bootstrapServers")
		if bootstrapServer == nil || len(bootstrapServer) == 0 {
			bootstrapServer = strings.Split(os.Getenv("BOOTSTRAP_SERVERS"), ";")
		}
		serve, _ := cmd.Flags().GetBool("serve")
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		err := exporters.Export(clientID, clientSecret, tokenURL, bootstrapServer, exporters.Prometheus, serve, host, port)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().String("clientId", "", "Client ID to use, or use the environment variable CLIENT_ID")
	exportCmd.Flags().String("clientSecret", "", "Client Secret to use, or use the environment variable CLIENT_SECRET")
	exportCmd.Flags().String("tokenURL", "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token", "The token endpoint to use")
	exportCmd.Flags().StringArray("bootstrapServer", make([]string, 0), "The bootstrap server, or use a semi-colon separated list in the environment variable BOOTSTRAP_SERVERS")
	exportCmd.Flags().Bool("serve", false, "If true, will run an http server")
	exportCmd.Flags().Int("port", 7843, "The port to run the http server on")
	exportCmd.Flags().String("host", "localhost", "The host to run the server on")
}
