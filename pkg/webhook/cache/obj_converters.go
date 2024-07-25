package cache

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

func convertToReplicaSet(obj interface{}) *appsv1.ReplicaSet {
	replicaSet, ok := obj.(*appsv1.ReplicaSet)
	if !ok {
		klog.Errorf("Cannot convert %v to *appsv1.ReplicaSet", obj)
		return nil
	}

	return replicaSet
}

func convertToDeployment(obj interface{}) *appsv1.Deployment {
	deployment, ok := obj.(*appsv1.Deployment)
	if !ok {
		klog.Errorf("Cant convert obj to *appsv1.Deployment: %v", obj)
		return nil
	}

	return deployment
}

func convertToStatefulSet(obj interface{}) *appsv1.StatefulSet {
	statefulSet, ok := obj.(*appsv1.StatefulSet)
	if !ok {
		klog.Errorf("Cant convert obj to *appsv1.StatefulSet: %v", obj)
		return nil
	}

	return statefulSet
}

func convertToPod(obj interface{}) *corev1.Pod {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		klog.Errorf("Cant convert obj to *corev1.Pod: %v", obj)
		return nil
	}

	return pod
}
