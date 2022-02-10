package system

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/layer5io/meshery/mesheryctl/internal/cli/root/config"
	"github.com/layer5io/meshery/mesheryctl/pkg/utils"
	meshkitutils "github.com/layer5io/meshkit/utils"
	meshkitkube "github.com/layer5io/meshkit/utils/kubernetes"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// runPortForward is used for port-forwarding Meshery UI via `system dashboard`
	runPortForward bool

	// defaultHost is the default host used for port-forwarding via `system dashboard`
	defaultHost = "localhost"

	// defaultPort is for port-forwarding via `system dashboard`
	defaultPort = 9081

	podPort = 8080
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Open Meshery UI in browser.",
	Args:  cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// check if meshery is running or not
		mctlCfg, err := config.GetMesheryCtl(viper.GetViper())
		if err != nil {
			return errors.Wrap(err, "error processing config")
		}
		currCtx, err := mctlCfg.GetCurrentContext()
		if err != nil {
			return err
		}
		running, _ := utils.IsMesheryRunning(currCtx.GetPlatform())
		if !running {
			return errors.New(`meshery server is not running. run "mesheryctl system start" to start meshery`)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		mctlCfg, err := config.GetMesheryCtl(viper.GetViper())
		if err != nil {
			return errors.Wrap(err, "error processing config")
		}
		currCtx, err := mctlCfg.GetCurrentContext()
		if err != nil {
			return err
		}
		log.Debug("fetching Meshery-UI endpoint")

		switch currCtx.GetPlatform() {
		case "docker":
			break
		case "kubernetes":
			client, err := meshkitkube.New([]byte(""))
			if err != nil {
				return err
			}

			// Run port forwarding for accessing Meshery UI
			if runPortForward {
				signals := make(chan os.Signal, 1)
				signal.Notify(signals, os.Interrupt)
				defer signal.Stop(signals)

				portforward, err := utils.NewPortForward(
					cmd.Context(),
					client,
					utils.MesheryNamespace,
					"meshery",
					defaultHost,
					defaultPort,
					podPort,
					false,
				)
				if err != nil {
					return fmt.Errorf("failed to initialize port-forward: %s\n", err)

				}

				if err = portforward.Init(); err != nil {
					// TODO: consider falling back to an ephemeral port if defaultPort is taken
					return fmt.Errorf("error running port-forward: %s\n", err)
				}
				log.Info("Starting Port-forwarding for Meshery UI")

				mesheryURL := portforward.URLFor("")

				// ticker for keeping connection alive with pod each 10 seconds
				ticker := time.NewTicker(10 * time.Second)
				go func() {
					for {
						select {
						case <-signals:
							portforward.Stop()
							ticker.Stop()
							return
						case <-ticker.C:
							keepConnectionAlive(mesheryURL)
						}
					}
				}()

				log.Info("Meshery UI available at: ", mesheryURL)
				log.Info("Opening Meshery UI in the default browser.")
				err = utils.NavigateToBrowser(mesheryURL)
				if err != nil {
					log.Warn("Failed to open Meshery in browser, please point your browser to " + currCtx.GetEndpoint() + " to access Meshery.")
				}

				<-portforward.GetStop()
				return nil
			}

			var mesheryEndpoint string
			var endpoint *meshkitutils.Endpoint
			clientset := client.KubeClient
			var opts meshkitkube.ServiceOptions
			opts.Name = "meshery"
			opts.Namespace = utils.MesheryNamespace
			opts.APIServerURL = client.RestConfig.Host

			endpoint, err = meshkitkube.GetServiceEndpoint(context.TODO(), clientset, &opts)
			if err != nil {
				return err
			}

			mesheryEndpoint = fmt.Sprintf("%s://%s:%d", utils.EndpointProtocol, endpoint.Internal.Address, endpoint.Internal.Port)
			currCtx.SetEndpoint(mesheryEndpoint)
			if !meshkitutils.TcpCheck(&meshkitutils.HostPort{
				Address: endpoint.Internal.Address,
				Port:    endpoint.Internal.Port,
			}, nil) {
				currCtx.SetEndpoint(fmt.Sprintf("%s://%s:%d", utils.EndpointProtocol, endpoint.External.Address, endpoint.External.Port))
				if !meshkitutils.TcpCheck(&meshkitutils.HostPort{
					Address: endpoint.External.Address,
					Port:    endpoint.External.Port,
				}, nil) {
					u, _ := url.Parse(opts.APIServerURL)
					if meshkitutils.TcpCheck(&meshkitutils.HostPort{
						Address: u.Hostname(),
						Port:    endpoint.External.Port,
					}, nil) {
						mesheryEndpoint = fmt.Sprintf("%s://%s:%d", utils.EndpointProtocol, u.Hostname(), endpoint.External.Port)
						currCtx.SetEndpoint(mesheryEndpoint)
					}
				}
			}

			if err == nil {
				err = config.UpdateContextInConfig(viper.GetViper(), currCtx, mctlCfg.GetCurrentContextName())
				if err != nil {
					return err
				}
			}

		}

		log.Info("Opening Meshery (" + currCtx.GetEndpoint() + ") in browser.")
		err = utils.NavigateToBrowser(currCtx.GetEndpoint())
		if err != nil {
			log.Warn("Failed to open Meshery in browser, please point your browser to " + currCtx.GetEndpoint() + " to access Meshery.")
		}

		return nil
	},
}

// keepConnectionAlive to stop being timed out with port forwarding
func keepConnectionAlive(url string) {
	_, err := http.Get(url)
	if err != nil {
		log.Debugf("connection request failed %v", err)
	}
	log.Debugf("connection request success")
}

func init() {
	dashboardCmd.Flags().BoolVarP(&runPortForward, "port-forward", "", false, "(optional) Use port forwarding to access Meshery UI")
}
