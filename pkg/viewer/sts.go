package viewer

import (
	"context"
	"fmt"

	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type StsViewer struct {
	clientset   *kubernetes.Clientset
	namespace   string
	statefulset *appsv1.StatefulSet
	selector    labels.Selector
}

func NewStsViewer(clientset *kubernetes.Clientset, name string, namespace string) (KindViewer, error) {
	sts, err := clientset.AppsV1().StatefulSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Get rs error: %s", err)
	}

	selector, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		return nil, fmt.Errorf("Failed to create selector for sts: %v", sts)
	}

	return &StsViewer{
		clientset:   clientset,
		namespace:   namespace,
		statefulset: sts,
		selector:    selector,
	}, nil
}

func (r *StsViewer) list() ([]*appsv1.ControllerRevision, error) {
	return listControllerRevison(*r.clientset, r.statefulset, r.selector)
}

func (r *StsViewer) List(isShowDetail bool) (table.Writer, error) {
	crList, err := r.list()
	if err != nil {
		return nil, err
	}
	podMap := make(map[string][]*corev1.Pod)
	if isShowDetail {
		podMap, err = listPods(r.clientset, r.statefulset, r.selector, "", "controller-revision-hash")
		if err != nil {
			return nil, err
		}
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Create Time", "Name"})

	for _, cr := range crList {
		name := cr.Name
		createTime := cr.ObjectMeta.CreationTimestamp.Time.String()
		rev := cr.Revision
		t.AppendRow(table.Row{
			rev,
			createTime,
			name,
		}, table.RowConfig{})

		if podList, ok := podMap[name]; ok {
			for _, pod := range podList {
				t.AppendRow(table.Row{
					"",
					"",
					fmt.Sprintf("└─%s", pod.Name),
				}, table.RowConfig{})
			}
		}

	}
	return t, nil
}

func (r *StsViewer) Diff(oldRev int64, newRev int64) (*string, error) {
	crList, err := r.list()
	if err != nil {
		return nil, err
	}
	return resourceDiff(
		crList,
		oldRev,
		newRev,
		getCrRev,
		getCrDiffString,
	)
}
