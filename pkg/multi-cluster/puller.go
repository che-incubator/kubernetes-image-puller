package multicluster

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/redhat-developer/kubernetes-image-puller/cfg"
	"github.com/redhat-developer/kubernetes-image-puller/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// CacheImages starts and maintains a daemonset to ensure images are
// cached.
func CacheImages() {
	// Set up kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Shared config to use osoproxy
	config.BearerToken = utils.GetServiceAccountToken(
		cfg.ServiceAccountID,
		cfg.ServiceAccountSecret,
		cfg.OidcProvider,
	)
	config.Host = cfg.ProxyURL
	config.TLSClientConfig = rest.TLSClientConfig{
		Insecure: true,
	}

	var wg sync.WaitGroup
	wg.Add(len(cfg.ImpersonateUsers))
	for _, user := range cfg.ImpersonateUsers {
		var shutdownChan = make(chan os.Signal, 1)
		signal.Notify(shutdownChan, syscall.SIGTERM)

		configCopy := *config
		go cacheImagesForUser(user, &configCopy, shutdownChan, &wg)
	}
	wg.Wait()
	log.Printf("Shutting down cleanly")
}

func cacheImagesForUser(impersonateUser string,
	config *rest.Config,
	shutdownChan chan os.Signal,
	wg *sync.WaitGroup) {

	log.Printf("Starting caching process for impersonate user '%s'", impersonateUser)
	config.Impersonate = rest.ImpersonationConfig{
		UserName: impersonateUser,
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf(err.Error())
	}

	// Clean up existing deployment if necessary
	utils.DeleteDaemonsetIfExists(clientset)
	// Create daemonset to cache images
	utils.CacheImages(clientset)
	utils.LogNumNodesScheduled(clientset, impersonateUser)

	for {
		select {
		case <-shutdownChan:
			log.Printf("Received SIGTERM, deleting daemonset")
			utils.DeleteDaemonsetIfExists(clientset)
			wg.Done()
		case <-time.After(time.Duration(cfg.CachingInterval) * time.Hour):
			log.Printf("Checking daemonset for user '%s'", impersonateUser)
			utils.RefreshCache(clientset)
			utils.LogNumNodesScheduled(clientset, impersonateUser)
		}
	}
}
