package cassandra

import (
	"fmt"
	"github.com/rook/rook/cmd/rook/rook"
	rookInformers "github.com/rook/rook/pkg/client/informers/externalversions"
	"github.com/rook/rook/pkg/operator/cassandra/controller"
	"github.com/rook/rook/pkg/operator/k8sutil"
	"github.com/rook/rook/pkg/util/flags"
	"github.com/spf13/cobra"
	"k8s.io/apiserver/pkg/server"
	kubeInformers "k8s.io/client-go/informers"
	"time"
)

const resyncPeriod = time.Second * 30

var operatorCmd = &cobra.Command{
	Use:   "operator",
	Short: "Runs the cassandra operator to deploy and manage cassandra in Kubernetes",
	Long: `Runs the cassandra operator to deploy and manage cassandra in kubernetes clusters.
https://github.com/rook/rook`,
}

func init() {
	flags.SetFlagsFromEnv(operatorCmd.Flags(), rook.RookEnvVarPrefix)

	operatorCmd.RunE = startOperator
}

func startOperator(cmd *cobra.Command, args []string) error {
	rook.SetLogLevel()
	rook.LogStartupInfo(operatorCmd.Flags())

	kubeClient, _, rookClient, err := rook.GetClientset()
	if err != nil {
		rook.TerminateFatal(fmt.Errorf("failed to get k8s clients. %+v", err))
	}

	logger.Infof("starting cassandra operator")

	// Using the current image version to deploy other rook pods
	pod, err := k8sutil.GetRunningPod(kubeClient)
	if err != nil {
		rook.TerminateFatal(fmt.Errorf("failed to get pod. %+v\n", err))
	}

	rookImage, err := k8sutil.GetContainerImage(pod, "")
	if err != nil {
		rook.TerminateFatal(fmt.Errorf("failed to get container image. %+v\n", err))
	}

	kubeInformerFactory := kubeInformers.NewSharedInformerFactory(kubeClient, resyncPeriod)
	rookInformerFactory := rookInformers.NewSharedInformerFactory(rookClient, resyncPeriod)

	c := controller.New(
		rookImage,
		kubeClient,
		rookClient,
		rookInformerFactory.Cassandra().V1alpha1().Clusters(),
		kubeInformerFactory.Apps().V1().StatefulSets(),
		kubeInformerFactory.Core().V1().Services(),
		kubeInformerFactory.Core().V1().Pods(),
		kubeInformerFactory.Core().V1().ServiceAccounts(),
		kubeInformerFactory.Rbac().V1().Roles(),
		kubeInformerFactory.Rbac().V1().RoleBindings(),
	)

	// Create a channel to receive OS signals
	stopCh := server.SetupSignalHandler()

	// Start the informer factories
	go kubeInformerFactory.Start(stopCh)
	go rookInformerFactory.Start(stopCh)

	// Start the controller
	if err = c.Run(1, stopCh); err != nil {
		logger.Fatalf("Error running controller: %s", err.Error())
	}

	return nil
}
