package main

import (
	"strings"

	maddrproxy "github.com/hrntknr/maddr-proxy/pkg/maddr-proxy"
	"github.com/spf13/cobra"
)

type Config struct {
	Listen   string
	Password string
}

var flagWatch bool
var flagIface string
var setupRouteCmd = &cobra.Command{
	Use: "setup-route",
	Run: func(cmd *cobra.Command, args []string) {
		ifaceMatch := strings.Split(flagIface, ",")
		if len(ifaceMatch) == 1 && ifaceMatch[0] == "" {
			ifaceMatch = []string{}
		}
		if err := maddrproxy.SetupRoute(flagWatch, ifaceMatch); err != nil {
			panic(err)
		}
	},
}

var flagListen string
var flagPassword string
var flagSetupRoute bool
var flagSetupRouteIface string
var proxyCmd = &cobra.Command{
	Use: "proxy",
	Run: func(cmd *cobra.Command, args []string) {
		if flagSetupRoute {
			ifaceMatch := strings.Split(flagIface, ",")
			if len(ifaceMatch) == 1 && ifaceMatch[0] == "" {
				ifaceMatch = []string{}
			}
			go func() {
				if err := maddrproxy.SetupRoute(true, ifaceMatch); err != nil {
					panic(err)
				}
			}()
		}
		passwords := strings.Split(flagPassword, ",")
		if len(passwords) == 1 && passwords[0] == "" {
			passwords = []string{}
		}
		if err := maddrproxy.NewProxy(passwords).ListenAndServe(flagListen); err != nil {
			panic(err)
		}
	},
}

func main() {
	rootCmd := &cobra.Command{
		Use: "maddr-proxy",
	}
	setupRouteCmd.Flags().BoolVarP(&flagWatch, "watch", "w", false, "watch")
	setupRouteCmd.Flags().StringVarP(&flagIface, "iface", "i", "en.*,eth.*", "interface match")
	rootCmd.AddCommand(setupRouteCmd)
	proxyCmd.Flags().StringVarP(&flagListen, "listen", "l", ":1080", "listen address")
	proxyCmd.Flags().StringVarP(&flagPassword, "password", "p", "", "password")
	proxyCmd.Flags().BoolVarP(&flagSetupRoute, "setup-route", "", false, "setup route")
	proxyCmd.Flags().StringVarP(&flagSetupRouteIface, "setup-route-iface", "", "en.*,eth.*", "interface match")
	rootCmd.AddCommand(proxyCmd)
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
