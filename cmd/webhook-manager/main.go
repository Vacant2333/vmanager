package main

import (
	"os"

	"k8s.io/component-base/cli"
	"k8s.io/klog/v2"
	controllerruntime "sigs.k8s.io/controller-runtime"

	"vacant.sh/vmanager/cmd/webhook-manager/app"
)

func main() {
	ctx := controllerruntime.SetupSignalHandler()
	controllerruntime.SetLogger(klog.Background())
	cmd := app.NewWebhookManagerCommand(ctx)

	os.Exit(cli.Run(cmd))
}
