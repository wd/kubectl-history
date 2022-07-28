package viewer

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func listPods(clientset *kubernetes.Clientset, owner metav1.Object, selector labels.Selector, keyPrefix string, revLabel string) (map[string][]*corev1.Pod, error) {
	options := metav1.ListOptions{
		LabelSelector: selector.String(),
	}

	podList, err := clientset.CoreV1().Pods(owner.GetNamespace()).List(context.TODO(), options)
	if err != nil {
		return nil, fmt.Errorf("Get pods list error: %s", err)
	}

	podMap := make(map[string][]*corev1.Pod)
	for _, item := range podList.Items {
		pod := item
		rev := pod.Labels[revLabel]
		key := fmt.Sprintf("%s%s", keyPrefix, rev)
		podMap[key] = append(podMap[key], &pod)
	}
	return podMap, nil
}
