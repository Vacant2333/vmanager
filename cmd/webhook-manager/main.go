package main

import (
	"os"

	"k8s.io/component-base/cli"
	"k8s.io/klog/v2"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func main() {
	ctx := controllerruntime.SetupSignalHandler()
	controllerruntime.SetLogger(klog.Background())
	cmd := app.NewWebhookCommand(ctx)
	code := cli.Run(cmd)
	os.Exit(code)
}
