package main

import (
	"os"
	"path"
	"time"

	singleCluster "github.com/che-incubator/kubernetes-image-puller/pkg/single-cluster"
	"github.com/che-incubator/kubernetes-image-puller/utils"

	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	namespace               = os.Getenv("NAMESPACE")
	kubeConfig              = os.Getenv("KUBECONFIG")
	daemonsetName           = os.Getenv("DAEMONSET_NAME")
	daemonsetTimeoutSeconds = time.Duration(120) * time.Second
)

func getKubeConfig(t *testing.T) (*rest.Config, error) {
	var config *rest.Config
	var err error
	if kubeConfig != "" {
		t.Logf("No kubeconfig")
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
	} else {
		t.Logf("Using ~/.kube/config")
		config, err = clientcmd.BuildConfigFromFlags("", path.Join(os.Getenv("HOME"), ".kube", "config"))
	}
	if err != nil {
		t.Errorf("Error creating rest config")
		return nil, err
	}
	return config, nil
}

func getClientset(t *testing.T) (*kubernetes.Clientset, error) {
	config, err := getKubeConfig(t)
	if err != nil {
		return nil, err
	}

	c := kubernetes.NewForConfigOrDie(config)
	return c, nil
}

func TestSingleClusterCacheImages(t *testing.T) {

	type testCase struct {
		name   string
		images string
	}

	testCases := []testCase{
		{name: "test images", images: "che-theia=quay.io/eclipse/che-theia:next;che-plugin-registry=quay.io/eclipse/che-plugin-registry:next"},
		{name: "test scratch image", images: "che-machine-exec=quay.io/eclipse/che-machine-exec:next;"},
		{name: "test volume mount image", images: "volume-mount-image=quay.io/dkwon17/volume-mount:latest"},
	}

	clientset, err := getClientset(t)
	if err != nil {
		t.Fatalf(err.Error())
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			os.Setenv("IMAGES", test.images)
			cacheImages(t, clientset)
		})
	}
}

func cacheImages(t *testing.T, clientset *kubernetes.Clientset) {
	go singleCluster.CacheImages()

	dsListChan := make(chan appsv1.DaemonSetList)

	go func() {
		for {
			daemonsets, err := clientset.AppsV1().DaemonSets(namespace).List(metav1.ListOptions{})
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
	case <-time.After(daemonsetTimeoutSeconds):
		checkPods(t, clientset)
		t.Errorf("Timeout waiting for a ready daemonset (%v)", daemonsetTimeoutSeconds)
	}

	if len(gotDaemonsets.Items) != 1 {
		t.Errorf("Wanted 1 ready daemonset but got %v", len(gotDaemonsets.Items))
	}

	if len(gotDaemonsets.Items) == 1 {
		if gotDaemonsets.Items[0].Name != daemonsetName {
			t.Errorf("Expected daemonset named %v, got %v", daemonsetName, gotDaemonsets.Items[0].Name)
		}
	}

	utils.DeleteDaemonsetIfExists(clientset)
}

func checkPods(t *testing.T, clientset *kubernetes.Clientset) {
	pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		t.Errorf("Error listing pods: %v", err)
	}
	for _, pod := range pods.Items {
		if pod.ObjectMeta.OwnerReferences[0].Name == daemonsetName {
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.State.Waiting != nil {
					t.Error(getWaitingErrorMessage(containerStatus))
				}
			}
		}
	}
}

func getWaitingErrorMessage(containerStatus v1.ContainerStatus) string {
	container := "Waiting container '" + containerStatus.Name + "'"
	reason := ""
	message := ""
	if len(containerStatus.State.Waiting.Reason) > 0 {
		reason = ", reason: '" + containerStatus.State.Waiting.Reason + "'"
	}
	if len(containerStatus.State.Waiting.Message) > 0 {
		message = ", message: '" + containerStatus.State.Waiting.Message + "'"
	}
	return container + reason + message
}
