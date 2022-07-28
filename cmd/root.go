package cmd

import (
	"os"

	"fmt"

	"flag"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

const (
	SupportedResources = "deployment|deploy|daemonset|ds|statefulset|sts"
)

var (
	cf           *genericclioptions.ConfigFlags
	clientConfig clientcmd.ClientConfig
	clientSet    *kubernetes.Clientset
	namespace    string

	rootCmd = &cobra.Command{
		Use:   "kubectl-v",
		Short: "List and diff versions of deployment/daemonset/statefulset",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
)

func init() {
	klog.InitFlags(nil)

	flags := pflag.NewFlagSet("kubectl-v", pflag.ExitOnError)
	flags.AddGoFlagSet(flag.CommandLine)
	pflag.CommandLine = flags

	// hide all glog flags except for -v
	flags.VisitAll(func(f *pflag.Flag) {
		if f.Name != "v" {
			flags.Lookup(f.Name).Hidden = true
		}
	})

	cf = genericclioptions.NewConfigFlags(true)
	streams := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
	rootCmd.SetOutput(streams.ErrOut)
	flags.AddFlagSet(rootCmd.PersistentFlags())

	flags.StringVar(cf.KubeConfig, "kubeconfig", "", "Path to the kubeconfig file to use for CLI requests.")
	flags.StringVarP(cf.Namespace, "namespace", "n", "", "If present, the namespace scope for this CLI request")
}

func getClientSet() (*kubernetes.Clientset, error) {
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return clientSet, nil
}

func getNamespace() string {
	if v := *cf.Namespace; v != "" {
		return v
	}
	clientConfig := cf.ToRawKubeConfigLoader()
	defaultNamespace, _, err := clientConfig.Namespace()
	if err != nil {
		defaultNamespace = "default"
	}
	return defaultNamespace
}

func prepare() error {
	clientConfig = cf.ToRawKubeConfigLoader()

	var err error
	clientSet, err = getClientSet()
	if err != nil {
		return fmt.Errorf("Init kubernetes client with error: %s", err)
	}

	namespace = getNamespace()
	return nil
}

func Execute() {
	defer klog.Flush()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
