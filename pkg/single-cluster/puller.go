package singlecluster

import (
	"log"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"

	"github.com/che-incubator/kubernetes-image-puller/cfg"
	"github.com/che-incubator/kubernetes-image-puller/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// CacheImages starts and maintains a daemonset to ensure images are
// cached.
func CacheImages() {
	// Set up kubernetes client
	// Look in
	// 1) $KUBECONFIG -- For testing
	// 2) ~/.kube/config -- For testing
	// 3) InClusterConfig

	var config *rest.Config
	var err error
	defaultKubeConfigPath := path.Join(os.Getenv("HOME"), ".kube", "config")
	if kubeConfigEnv := os.Getenv("KUBECONFIG"); kubeConfigEnv != "" {
		if config, err = clientcmd.BuildConfigFromFlags("", kubeConfigEnv); err != nil {
			log.Fatalf("Error building REST Config: %v", err)
		}
	} else if _, err := os.Stat(defaultKubeConfigPath); err == nil {
		if config, err = clientcmd.BuildConfigFromFlags("", defaultKubeConfigPath); err != nil {
			log.Fatalf("Error building REST Config: %v", err)
		}
	} else {
		if config, err = rest.InClusterConfig(); err != nil {
			log.Fatalf("Error building REST Config: %v", err)
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGTERM)

	go cacheImagesLocally(config, shutdownChan, &wg)
	wg.Wait()
	log.Printf("Shutting down cleanly")
}

func cacheImagesLocally(config *rest.Config,
	shutdownChan chan os.Signal,
	wg *sync.WaitGroup) {
	cfg := cfg.GetConfig()

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("Error creating Clientset: %v", err)
	}

	// Clean up existing deployment if necessary
	utils.DeleteDaemonsetIfExists(clientset)
	// Create daemonset to cache images
	utils.CacheImages(clientset)
	utils.LogNumNodesScheduled(clientset, "(single user mode)")

	for {
		select {
		case <-shutdownChan:
			log.Printf("Received SIGTERM, deleting daemonset")
			utils.DeleteDaemonsetIfExists(clientset)
			wg.Done()
		case <-time.After(time.Duration(cfg.CachingInterval) * time.Hour):
			utils.RefreshCache(clientset)
			utils.LogNumNodesScheduled(clientset, "(single user mode)")
		}
	}
}
