package options

import (
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	defaultBindAddress = "0.0.0.0"
	defaultPort        = 8443
)

type Options struct {
	Master     string
	KubeConfig string

	BindAddress string
	SecurePort  int
}

// NewOptions return a new webhook-manager options.
func NewOptions() *Options {
	return &Options{}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Master, "master", "", "The address of the Kubernetes API server (overrides any value in kubeconfig).")
	fs.StringVar(&o.KubeConfig, "kubeconfig", "", "Path to kubeconfig file with authorization and master location information.")

	fs.StringVar(&o.BindAddress, "bind-address", defaultBindAddress,
		"The IP address on which to listen for the --secure-port port.")
	fs.IntVar(&o.SecurePort, "secure-port", defaultPort,
		"The secure port on which to serve HTTPS.")
}

func (o *Options) Validate() field.ErrorList {
	errList := field.ErrorList{}

	if o.KubeConfig == "" {
		errList = append(errList, field.Required(field.NewPath("kubeconfig"), "must specify a kubeconfig path."))
	}

	return errList
}
