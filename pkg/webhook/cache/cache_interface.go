package cache

import (
	corev1 "k8s.io/api/core/v1"

	podaffinity "vacant.sh/vmanager/pkg/definitions/pod-affinity"
)

type Interface interface {
	Run(stopCh <-chan struct{})
	DetermineNewPodAffinityPreference(pod *corev1.Pod) podaffinity.PodAffinitySettingName
}
