package viewer

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
	"sort"
)

var ControllerRevisionGKV = schema.GroupVersionKind{
	Group:   "apps",
	Version: "v1",
	Kind:    "ControllerRevision",
}

func listControllerRevison(clientset kubernetes.Clientset, owner metav1.Object, selector labels.Selector) ([]*appsv1.ControllerRevision, error) {
	options := metav1.ListOptions{
		LabelSelector: selector.String(),
	}
	crList, err := clientset.AppsV1().ControllerRevisions(owner.GetNamespace()).List(context.TODO(), options)
	if err != nil {
		return nil, fmt.Errorf("Get controller reversion error: %s", err)
	}

	var resList []*appsv1.ControllerRevision
	for _, res := range crList.Items {
		obj := res
		if metav1.IsControlledBy(&obj, owner) {
			obj.SetGroupVersionKind(ControllerRevisionGKV)
			resList = append(resList, &obj)
		}
	}

	sort.Slice(resList, func(i, j int) bool {
		return resList[i].Revision > resList[j].Revision
	})

	return resList, nil
}

func getCrRev(cr *appsv1.ControllerRevision) int64 {
	return cr.Revision
}

func getCrDiffString(cr *appsv1.ControllerRevision) string {
	unstructedRs, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(cr.DeepCopyObject())
	newYaml, _ := yaml.Marshal(unstructedRs["data"])
	return string(newYaml)
}
