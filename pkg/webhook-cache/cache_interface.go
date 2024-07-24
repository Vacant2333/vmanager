package webhook_cache

type Interface interface {
	Run(stopCh <-chan struct{})
}
