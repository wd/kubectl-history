package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wd/kubectl-history/pkg/viewer"

	"github.com/jedib0t/go-pretty/v6/table"
	"k8s.io/klog/v2"
)

var isShowDetail bool

func init() {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("list %s NAME", SupportedResources),
		Short: "List all the revisions of the resource",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("Need resource type and resource name\n\n%s", cmd.UsageString())
			}
			if len(args) > 2 {
				return fmt.Errorf("Too many args\n\n%s", cmd.UsageString())
			}
			return nil
		},
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return prepare()
		},
		RunE: runList,
	}
	cmd.Flags().BoolVarP(&isShowDetail, "detail", "d", false, "Show more details")
	rootCmd.AddCommand(cmd)
}

func runList(cmd *cobra.Command, args []string) error {
	var resourceViewer viewer.KindViewer
	var err error

	kind, name := args[0], args[1]
	klog.V(3).Infof("Parsed namespace=%v kind=%v name=%v", namespace, kind, name)

	switch kind {
	case "deploy", "deployment":
		resourceViewer, err = viewer.NewDeployViewer(clientSet, name, namespace)
	case "daemonset", "ds":
		resourceViewer, err = viewer.NewDSViewer(clientSet, name, namespace)
	case "statefulset", "sts":
		resourceViewer, err = viewer.NewStsViewer(clientSet, name, namespace)
	default:
		return fmt.Errorf("Resource type %s is not supported.", kind)
	}
	if err != nil {
		return err
	}

	t, err := resourceViewer.List(isShowDetail)
	if err != nil {
		return err
	} else {
		t.Style().Options = table.Options{
			DrawBorder:      false,
			SeparateColumns: true,
			SeparateFooter:  false,
			SeparateHeader:  true,
			SeparateRows:    false,
		}

		t.Render()
	}
	return nil
}
