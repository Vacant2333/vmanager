package app

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"vacant.sh/vmanager/cmd/webhook-manager/app/options"
	"vacant.sh/vmanager/pkg/webhook/cache"
	"vacant.sh/vmanager/pkg/webhook/deployment"
	"vacant.sh/vmanager/pkg/webhook/pod"
	"vacant.sh/vmanager/pkg/webhook/statefulset"
)

const ComponentName = "vmanager-webhook-manager"

func NewWebhookManagerCommand(ctx context.Context) *cobra.Command {
	// Init the flags to global flag config.
	klog.InitFlags(flag.CommandLine)

	// Build the options.
	opts := options.NewOptions()
	opts.AddFlags(pflag.CommandLine)

	cmd := &cobra.Command{
		Use:  ComponentName,
		Long: fmt.Sprintf("The %s starts a webhook server to manage the workload's locate.", ComponentName),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate the options.
			if errs := opts.Validate(); len(errs) > 0 {
				return errs.ToAggregate()
			}

			return Run(ctx, opts)
		},
		Args: cobra.NoArgs,
	}

	return cmd
}

func Run(ctx context.Context, opts *options.Options) error {
	klog.V(3).Infof("Start to run the %s.", ComponentName)

	// Build the kube config by the option parameters.
	kubeConfig, err := clientcmd.BuildConfigFromFlags(opts.Master, opts.KubeConfig)
	if err != nil {
		return err
	}

	// Build the webhook cache.
	wc, err := cache.NewWebhookCache(kubeConfig)
	if err != nil {
		return err
	}

	webhookManager, err := controllerruntime.NewManager(kubeConfig, controllerruntime.Options{
		Logger: klog.Background(),
		WebhookServer: webhook.NewServer(webhook.Options{
			Host: opts.BindAddress,
			Port: opts.SecurePort,

			CertDir:  opts.CertDir,
			CertName: opts.CertName,
			KeyName:  opts.KeyName,

			TLSOpts: []func(*tls.Config){
				func(config *tls.Config) {
					config.MinVersion = tls.VersionTLS13
				},
			},
		}),
		LeaderElection: false,
	})
	if err != nil {
		return err
	}

	// Run the WebhookCache, wait for cache sync.
	wc.Run(ctx.Done())

	klog.V(3).Infof("Registering webhook to %s.", ComponentName)
	webhookServer := webhookManager.GetWebhookServer()
	{
		decoder := admission.NewDecoder(webhookManager.GetScheme())

		webhookServer.Register("/mutate-pod", &webhook.Admission{
			Handler: &pod.Mutating{Decoder: decoder, Cache: wc},
		})
		webhookServer.Register("/validate-pod", &webhook.Admission{
			Handler: &pod.Validating{Decoder: decoder},
		})
		webhookServer.Register("/validate-deployment", &webhook.Admission{
			Handler: &deployment.Validating{Decoder: decoder},
		})
		webhookServer.Register("/validate-statefulset", &webhook.Admission{
			Handler: &statefulset.Validating{Decoder: decoder},
		})
	}

	// Block until err or context is done.
	return webhookServer.Start(ctx)
}
