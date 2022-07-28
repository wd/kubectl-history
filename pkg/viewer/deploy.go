package viewer

import (
	"context"
	"fmt"
	"sort"

	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
	"strconv"
)

var ReplicasetGkV = schema.GroupVersionKind{
	Group:   "apps",
	Version: "v1",
	Kind:    "ReplicaSet",
}

type DeployViewer struct {
	clientset  *kubernetes.Clientset
	deployment *appsv1.Deployment
	selector   labels.Selector
}

func NewDeployViewer(clientset *kubernetes.Clientset, deployName string, namespace string) (KindViewer, error) {
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deployName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Get deployment error: %s", err)
	}

	selector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		return nil, fmt.Errorf("Failed to get selector for deployment: %s", err)
	}

	return &DeployViewer{
		clientset:  clientset,
		deployment: deployment,
		selector:   selector,
	}, nil
}

func (r *DeployViewer) list() ([]*appsv1.ReplicaSet, error) {
	origList, err := r.clientset.AppsV1().ReplicaSets(r.deployment.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: r.selector.String()})

	if err != nil {
		return nil, fmt.Errorf("Get rs failed for deployment: %v", r.deployment)
	}

	var rsList []*appsv1.ReplicaSet
	for _, res := range origList.Items {
		obj := res
		if metav1.IsControlledBy(&obj, r.deployment) {
			obj.SetGroupVersionKind(ReplicasetGkV)
			rsList = append(rsList, &obj)
		}
	}

	sort.Slice(rsList, func(i, j int) bool {
		return getRsRev(rsList[i]) > getRsRev(rsList[j])
	})
	return rsList, nil
}

func (r *DeployViewer) List(isShowDetail bool) (table.Writer, error) {
	rsList, err := r.list()
	if err != nil {
		return nil, err
	}

	podMap := make(map[string][]*corev1.Pod)
	if isShowDetail {
		podMap, err = listPods(r.clientset, r.deployment, r.selector, fmt.Sprintf("%s-", r.deployment.Name), "pod-template-hash")
		if err != nil {
			return nil, err
		}
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Create Time", "Name", "Desired", "Availiabe", "Ready"})

	for _, rs := range rsList {
		name := rs.Name
		avaReplica := rs.Status.AvailableReplicas
		readyReplica := rs.Status.ReadyReplicas
		descireRplica := *rs.Spec.Replicas
		createTime := rs.ObjectMeta.CreationTimestamp.Time.String()
		rev := getRsRev(rs)
		t.AppendRow(table.Row{
			rev,
			createTime,
			name,
			descireRplica,
			avaReplica,
			readyReplica,
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

func (r *DeployViewer) Diff(origOldRev, origNewRev int64) (*string, error) {
	rsList, err := r.list()
	if err != nil {
		return nil, err
	}

	return resourceDiff(
		rsList,
		origOldRev,
		origNewRev,
		getRsRev,
		getRsDiffString,
	)
}

func getRsRev(rs *appsv1.ReplicaSet) int64 {
	return getInt(rs.Annotations[RevisionAnnotation])
}

func getInt(s string) (i int64) {
	var err error
	if i, err = strconv.ParseInt(s, 10, 64); err != nil {
		panic(fmt.Errorf("can't convert string %s to int64: %s", s, err))
	}
	return i
}

func getRsDiffString(rs *appsv1.ReplicaSet) string {
	unstructedRs, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(rs.DeepCopyObject())
	removeKeysIfExists(unstructedRs, []string{"status"})
	removeKeysIfExists(unstructedRs["metadata"].(map[string]interface{}), []string{"managedFields", "creationTimestamp"})
	newYaml, _ := yaml.Marshal(unstructedRs)
	return string(newYaml)
}

func removeKeysIfExists(m map[string]interface{}, keys []string) {
	for _, key := range keys {
		if _, ok := m[key]; ok {
			delete(m, key)
		}
	}
}
