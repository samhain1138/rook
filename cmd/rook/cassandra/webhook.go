package cassandra

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/rook/rook/cmd/rook/rook"
	"github.com/rook/rook/pkg/util/flags"
	"github.com/rook/rook/pkg/webhook"
	"github.com/spf13/cobra"
	"k8s.io/apiserver/pkg/server"
	"net/http"
)

var webhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Runs the cassandra operator to deploy and manage cassandra in Kubernetes",
	Long: `Runs the cassandra operator to deploy and manage cassandra in kubernetes clusters.
https://github.com/rook/rook`,
}

var whConfig webhook.WebhookConfig

func init() {
	flags.SetFlagsFromEnv(webhookCmd.Flags(), rook.RookEnvVarPrefix)
	webhookCmd.Flags().Int32Var(&whConfig.Port, "port", 443, "Webhook server port.")
	webhookCmd.Flags().StringVar(&whConfig.TLSCertFile, "tlsCertFile", "/etc/webhook/certs/cert.pem", "File containing the x509 Certificate for HTTPS.")
	webhookCmd.Flags().StringVar(&whConfig.TLSKeyFile, "tlsKeyFile", "/etc/webhook/certs/key.pem", "File containing the x509 private key to --tlsCertFile.")

	webhookCmd.RunE = startAdmissionWebhook
}

func startAdmissionWebhook(cmd *cobra.Command, args []string) error {

	// Load X509 Keypair
	pair, err := tls.LoadX509KeyPair(whConfig.TLSCertFile, whConfig.TLSKeyFile)
	if err != nil {
		logger.Fatalf("Failed to load TLS Keypair with config %v: %s", spew.Sdump(whConfig), err.Error())
	}

	whServer := webhook.WebhookServer{
		Server: &http.Server{
			Addr:      fmt.Sprintf(":%d", whConfig.Port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
		AdmissionFns: &webhook.CassandraAdmission{},
	}

	// Define HTTP server and server handle
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", whServer.Serve)
	whServer.Server.Handler = mux

	// Start webhook server in new routine
	go func() {
		if err := whServer.Server.ListenAndServeTLS("", ""); err != nil {
			logger.Fatalf("Failed to listen and serve webhook server: %v", err)
		}
	}()

	logger.Info("Server started...")

	// Listen for OS shutdown signal
	stopCh := server.SetupSignalHandler()
	<-stopCh

	logger.Info("Got OS shutdown signal, shutting down gracefully...")
	whServer.Server.Shutdown(context.Background())
	return nil
}
