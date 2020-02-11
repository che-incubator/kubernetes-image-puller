package main

import (
	"os"
	"path"
	"time"

	singleCluster "github.com/che-incubator/kubernetes-image-puller/pkg/single-cluster"
	"github.com/che-incubator/kubernetes-image-puller/utils"

	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func setup(t *testing.T) *kubernetes.Clientset {
	kubeConfigVar := os.Getenv("KUBECONFIG")
	var config *rest.Config
	var err error
	if kubeConfigVar != "" {
		t.Logf("No kubeconfig")
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigVar)
	} else {
		t.Logf("Using ~/.kube/config")
		config, err = clientcmd.BuildConfigFromFlags("", path.Join(os.Getenv("HOME"), ".kube", "config"))
	}
	if err != nil {
		t.Errorf("Error creating rest config: %v", err)
	}
	c := kubernetes.NewForConfigOrDie(config)
	_, err = c.CoreV1().Namespaces().Create(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "k8s-image-puller",
		},
	})
	if err != nil {
		t.Errorf("Error creating namespace %v", err)
	}
	go singleCluster.CacheImages()
	return c
}

func teardown(clientset *kubernetes.Clientset, t *testing.T) {
	utils.DeleteDaemonsetIfExists(clientset)
	err := clientset.CoreV1().Namespaces().Delete("k8s-image-puller", metav1.NewDeleteOptions(30))
	if err != nil {
		t.Errorf("Could not delete namespace: %v", err)
	}
}

func TestSingleClusterCacheImages(t *testing.T) {
	clientset := setup(t)
	defer teardown(clientset, t)

	dsListChan := make(chan appsv1.DaemonSetList)

	go func() {
		for {
			daemonsets, err := clientset.AppsV1().DaemonSets("k8s-image-puller").List(metav1.ListOptions{})
			if err != nil {
				t.Errorf("Error listing daemonsets: %v", err)
			}
			if len(daemonsets.Items) != 0 {
				if daemonsets.Items[0].Status.NumberReady != 0 {
					t.Logf("Got a ready daemonset")
					dsListChan <- *daemonsets
				}
			}
			time.Sleep(2 * time.Second)
		}
	}()

	var gotDaemonsets appsv1.DaemonSetList
	select {
	case gotDS := <-dsListChan:
		gotDaemonsets = gotDS
	case <-time.After(120 * time.Second):
		t.Errorf("Timeout waiting for daemonsets")
	}

	if len(gotDaemonsets.Items) != 1 {
		t.Errorf("Wanted 1 daemonset but got %v", len(gotDaemonsets.Items))
	}

	if len(gotDaemonsets.Items) == 1 {
		if gotDaemonsets.Items[0].Name != "kubernetes-image-puller" {
			t.Errorf("Expected daemonset named kubernetes-image-puller, got %v", gotDaemonsets.Items[0].Name)
		}
	}
}
