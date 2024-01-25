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

type DSViewer struct {
	clientset *kubernetes.Clientset
	daemonset *appsv1.DaemonSet
	selector  labels.Selector
}

func NewDSViewer(clientset *kubernetes.Clientset, name string, namespace string) (KindViewer, error) {
	ds, err := clientset.AppsV1().DaemonSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Get DS error: %s", err)
	}

	selector, err := metav1.LabelSelectorAsSelector(ds.Spec.Selector)
	if err != nil {
		return nil, fmt.Errorf("Failed to create selector for ds: %v", ds)
	}

	return &DSViewer{
		clientset: clientset,
		daemonset: ds,
		selector:  selector,
	}, nil
}

func (r *DSViewer) list() ([]*appsv1.ControllerRevision, error) {
	return listControllerRevison(*r.clientset, r.daemonset, r.selector)
}

func (r *DSViewer) List(isShowDetail bool) (table.Writer, error) {
	crList, err := r.list()
	if err != nil {
		return nil, err
	}

	podMap := make(map[string][]*corev1.Pod)
	if isShowDetail {
		podMap, err = listPods(r.clientset, r.daemonset, r.selector, fmt.Sprintf("%s-", r.daemonset.Name), "controller-revision-hash")
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

func (r *DSViewer) Diff(origOldRev, origNewRev int64) (*string, error) {
	crList, err := r.list()
	if err != nil {
		return nil, err
	}

	return resourceDiff(
		crList,
		origOldRev,
		origNewRev,
		getCrRev,
		getCrDiffString,
	)
}
