package cassandra

import (
	"fmt"
	"github.com/rook/rook/cmd/rook/rook"
	"github.com/rook/rook/pkg/operator/cassandra/constants"
	"github.com/rook/rook/pkg/operator/cassandra/sidecar"
	"github.com/rook/rook/pkg/util/flags"
	"github.com/spf13/cobra"
	"k8s.io/apiserver/pkg/server"
	"os"
)

var sidecarCmd = &cobra.Command{
	Use:   "sidecar",
	Short: "Runs the cassandra sidecar to deploy and manage cassandra in Kubernetes",
	Long: `Runs the cassandra sidecar to deploy and manage cassandra in kubernetes clusters.
https://github.com/rook/rook`,
}

func init() {
	flags.SetFlagsFromEnv(operatorCmd.Flags(), rook.RookEnvVarPrefix)

	sidecarCmd.RunE = startSidecar
}

func startSidecar(cmd *cobra.Command, args []string) error {
	rook.SetLogLevel()
	rook.LogStartupInfo(operatorCmd.Flags())

	kubeClient, _, rookClient, err := rook.GetClientset()
	if err != nil {
		rook.TerminateFatal(fmt.Errorf("failed to get k8s clients. %+v", err))
	}

	podName := os.Getenv(constants.EnvVarPodName)
	if podName == "" {
		rook.TerminateFatal(fmt.Errorf("cannot detect the pod name. Please provide it using the downward API in the manifest file"))
	}
	podNamespace := os.Getenv(constants.EnvVarPodNamespace)
	if podNamespace == "" {
		rook.TerminateFatal(fmt.Errorf("cannot detect the pod namespace. Please provide it using the downward API in the manifest file"))
	}

	logger.Infof("Started rook sidecar for Cassandra.")

	mc, err := sidecar.New(
		podName,
		podNamespace,
		kubeClient,
		rookClient,
	)

	if err != nil {
		rook.TerminateFatal(fmt.Errorf("failed to initialize member controller: %s", err.Error()))
	}
	logger.Infof("Initialized Member Controller: %+v", mc)

	// Create a channel to receive OS signals
	stopCh := server.SetupSignalHandler()

	// Start the controller loop
	if err = mc.Run(1, stopCh); err != nil {
		logger.Fatalf("Error running sidecar: %s", err.Error())
	}

	return nil
}
