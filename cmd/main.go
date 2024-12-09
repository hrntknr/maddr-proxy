package main

import (
	maddrproxy "github.com/hrntknr/maddr-proxy/pkg/maddr-proxy"
	"github.com/spf13/cobra"
)

var flagWatch bool
var flagIface []string
var flagGw []string
var flagUseHostMinAsGw bool
var setupRouteCmd = &cobra.Command{
	Use: "setup-route",
	Run: func(cmd *cobra.Command, args []string) {
		if err := maddrproxy.SetupRoute(flagWatch, flagIface, flagGw, flagUseHostMinAsGw); err != nil {
			panic(err)
		}
	},
}

var flagListen string
var flagPassword []string
var flagSetupRoute bool
var flagSetupRouteIface []string
var flagSetupRouteGw []string
var flagSetupRouteUseHostMinAsGw bool
var proxyCmd = &cobra.Command{
	Use: "proxy",
	Run: func(cmd *cobra.Command, args []string) {
		if flagSetupRoute {
			go func() {
				if err := maddrproxy.SetupRoute(true, flagSetupRouteIface, flagSetupRouteGw, flagSetupRouteUseHostMinAsGw); err != nil {
					panic(err)
				}
			}()
		}
		if err := maddrproxy.NewProxy(flagPassword).ListenAndServe(flagListen); err != nil {
			panic(err)
		}
	},
}

func main() {
	rootCmd := &cobra.Command{
		Use: "maddr-proxy",
	}
	setupRouteCmd.Flags().BoolVarP(&flagWatch, "watch", "w", false, "watch")
	setupRouteCmd.Flags().StringSliceVarP(&flagIface, "iface", "i", []string{"en.*", "eth.*"}, "interface")
	setupRouteCmd.Flags().StringSliceVarP(&flagGw, "gw", "g", []string{}, "gateway")
	setupRouteCmd.Flags().BoolVarP(&flagUseHostMinAsGw, "use-host-min-as-gw", "", true, "use host min as gateway")
	rootCmd.AddCommand(setupRouteCmd)
	proxyCmd.Flags().StringVarP(&flagListen, "listen", "l", ":1080", "listen address")
	proxyCmd.Flags().StringSliceVarP(&flagPassword, "password", "p", []string{}, "password")
	proxyCmd.Flags().BoolVarP(&flagSetupRoute, "setup-route", "", false, "setup route")
	proxyCmd.Flags().StringSliceVarP(&flagSetupRouteIface, "setup-route-iface", "", []string{"en.*", "eth.*"}, "interface")
	proxyCmd.Flags().StringSliceVarP(&flagSetupRouteGw, "setup-route-gw", "", []string{}, "gateway")
	proxyCmd.Flags().BoolVarP(&flagSetupRouteUseHostMinAsGw, "setup-route-use-host-min-as-gw", "", true, "use host min as gateway")
	rootCmd.AddCommand(proxyCmd)
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
