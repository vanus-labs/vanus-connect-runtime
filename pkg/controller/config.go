package controller

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	clientset "github.com/vanus-labs/vanus-connect-runtime/pkg/client/clientset/versioned"
)

// Config is the controller conf
type Config struct {
	KubeConfigFile     string
	KubeRestConfig     *rest.Config
	KubeFactoryClient  kubernetes.Interface
	VanusFactoryClient clientset.Interface
}

// ParseFlags parses cmd args then init kubeclient and conf
// TODO: validate configuration
func NewConfig() (*Config, error) {
	config := &Config{}
	if err := config.initKubeFactoryClient(); err != nil {
		return nil, err
	}

	klog.Infof("config is  %+v", config)
	return config, nil
}

func (config *Config) initKubeFactoryClient() error {
	var cfg *rest.Config
	var err error
	config.KubeConfigFile = GetKubeConfigFromEnv()
	cfg, err = GetInClusterOrKubeConfig(config.KubeConfigFile)
	if err != nil {
		klog.Errorf("failed to build kubeconfig %v", err)
		return err
	}
	cfg.QPS = 1000
	cfg.Burst = 2000

	config.KubeRestConfig = cfg

	VanusClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("init vanus client failed %v", err)
		return err
	}
	config.VanusFactoryClient = VanusClient

	cfg.ContentType = "application/vnd.kubernetes.protobuf"
	cfg.AcceptContentTypes = "application/vnd.kubernetes.protobuf,application/json"
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("init kubernetes client failed %v", err)
		return err
	}
	config.KubeFactoryClient = kubeClient
	return nil
}

func GetKubeConfigFromEnv() string {
	home := os.Getenv("HOME")
	if home != "" {
		fpath := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(fpath); err != nil {
			return ""
		}
		return fpath
	}
	return ""
}

// try list authprovider
// kubeconfig
// serviceAccount
func GetInClusterOrKubeConfig(kubeconfig string) (config *rest.Config, rerr error) {
	config, rerr = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if rerr != nil {
		klog.Errorf("auth from kubeconfig failed:%v", rerr)
		return nil, rerr
	}
	return config, nil
}
