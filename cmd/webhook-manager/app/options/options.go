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

	CertDir  string
	CertName string
	KeyName  string

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

	fs.StringVar(&o.CertDir, "cert-dir", "", "The directory that contains the server key and certificate.")
	fs.StringVar(&o.CertName, "tls-cert-file-name", "tls.crt", "The name of server certificate.")
	fs.StringVar(&o.KeyName, "tls-private-key-file-name", "tls.key", "The name of server key.")

	fs.StringVar(&o.BindAddress, "bind-address", defaultBindAddress,
		"The IP address on which to listen for the --secure-port port.")
	fs.IntVar(&o.SecurePort, "secure-port", defaultPort,
		"The secure port on which to serve HTTPS.")
}

func (o *Options) Validate() field.ErrorList {
	errList := field.ErrorList{}

	if o.CertDir == "" {
		errList = append(errList, field.Required(field.NewPath("cert-dir"), "must specify --cert-dir"))
	}

	return errList
}
