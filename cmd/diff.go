package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wd/kubectl-v/pkg/viewer"
	"k8s.io/klog/v2"
	"strconv"
)

func init() {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("diff %s NAME [old] [new]", SupportedResources),
		Short: "Show a diff for different reversions of the resource",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("Need resource type and resource name\n\n%s", cmd.UsageString())
			}

			if len(args) > 4 {
				return fmt.Errorf("Too many args\n\n%s", cmd.UsageString())
			}
			return nil
		},
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return prepare()
		},
		RunE: runDiff,
	}
	rootCmd.AddCommand(cmd)
}

func runDiff(cmd *cobra.Command, args []string) error {
	var resourceViewer viewer.KindViewer
	var err error

	kind, name := args[0], args[1]
	oldRev, newRev := int64(-1), int64(0) // default value
	if len(args) >= 3 {
		oldRev, err = strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("Parse old reversion faild: %s", err)
		}
	}
	if len(args) >= 4 {
		newRev, err = strconv.ParseInt(args[3], 10, 64)

		if err != nil {
			return fmt.Errorf("Parse new reversion faild: %s", err)
		}
	}
	if oldRev == newRev {
		return fmt.Errorf("Old rev=%d, new rev=%d makes no sense", oldRev, newRev)
	}
	klog.V(3).Infof("Parsed namespace=%v kind=%v name=%v, old rev=%v, new rev=%v", namespace, kind, name, oldRev, newRev)

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

	diff, err := resourceViewer.Diff(oldRev, newRev)
	if err != nil {
		return err
	} else {
		fmt.Print(*diff)
	}
	return nil
}
