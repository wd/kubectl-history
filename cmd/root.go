package cmd

import (
	"flag"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	SupportedResources = "deployment|deploy|daemonset|ds|statefulset|sts"
)

var (
	Version = "git-head"

	cf           *genericclioptions.ConfigFlags
	clientConfig clientcmd.ClientConfig
	clientSet    *kubernetes.Clientset
	namespace    string

	rootCmd = &cobra.Command{
		Use:   "kubectl-history",
		Short: "List and diff versions of deployment/daemonset/statefulset",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
)

func init() {
	klog.InitFlags(nil)
	rootCmd.Version = Version
	flags := pflag.NewFlagSet("kubectl-history", pflag.ExitOnError)
	flags.AddGoFlagSet(flag.CommandLine)
	pflag.CommandLine = flags

	// hide all glog flags except `-v`
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

	flags.StringVarP(cf.KubeConfig, "kubeconfig", "f", "", "If present, path to the kubeconfig file to be used for Cli requests.")
	flags.StringVarP(cf.Context, "context", "c", "c", "If present, the context scope for the Cli request")
	flags.StringVarP(cf.Namespace, "namespace", "n", "", "If present, the namespace scope for this Cli request")
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
