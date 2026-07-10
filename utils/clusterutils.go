//
// Copyright (c) 2019-2024 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/che-incubator/kubernetes-image-puller/cfg"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"
)

const (
	kipVolumeName         = "kip"
	kipVolumeMountPath    = "/kip"
	copySleepCommand      = "cp /bin/sleep /kip/sleep"
	containerSleepCommand = "/kip/sleep"
	sleepDuration         = "720h"
)

var (
	propagationPolicy             = metav1.DeletePropagationForeground
	terminationGracePeriodSeconds = int64(1)

	nonRootUID         = int64(65532)
	nonRootGID         = int64(65532)
	seccompProfile     = corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault}
	kipVolumeSizeLimit = resource.MustParse("50Mi")

	// Volume mount to copy the sleep binary into.
	// To allow the image puller to cache scratch images, an initContainer copies
	// the sleep binary to this volume mount. As a result, every container has
	// access to the sleep binary via this volume mount.
	containerVolumeMounts = []corev1.VolumeMount{
		{
			Name:      kipVolumeName,
			MountPath: kipVolumeMountPath,
		},
	}
)

// Set up watch on daemonset
func watchDaemonset(clientset *kubernetes.Clientset) watch.Interface {
	cfg := cfg.GetConfig()
	watch, err := clientset.AppsV1().DaemonSets(cfg.Namespace).Watch(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", cfg.DaemonsetName),
		//IncludeUninitialized: true,
	})
	if err != nil {
		log.Fatalf("Failed to set up watch on daemonsets: %s", err.Error())
	}
	return watch
}

func getImagePullerDeployment(clientset *kubernetes.Clientset) *appsv1.Deployment {
	cfg := cfg.GetConfig()
	deploymentName := os.Getenv("DEPLOYMENT_NAME")
	if deploymentName == "" {
		log.Fatalf("DEPLOYMENT_NAME is not set for the image puller deployment")
	}

	deployment, err := clientset.AppsV1().Deployments(cfg.Namespace).Get(context.Background(), deploymentName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Failed to get Deployment: %v", err)
	}
	return deployment
}

func getOwnerReferenceFromDeployment(deployment *appsv1.Deployment) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Name:       deployment.Name,
		UID:        deployment.UID,
	}
}

func isOpenShift(clientset *kubernetes.Clientset) bool {
	apiGroups, err := clientset.Discovery().ServerGroups()
	if err != nil {
		log.Printf("Failed to discover API groups, assuming vanilla Kubernetes: %v", err)
		return false
	}

	for _, group := range apiGroups.Groups {
		if group.Name == "route.openshift.io" {
			return true
		}
	}

	return false
}

func getDaemonset(deployment *appsv1.Deployment, openShift bool) *appsv1.DaemonSet {
	cfg := cfg.GetConfig()

	imgPullSecrets := []corev1.LocalObjectReference{}
	for _, secretName := range cfg.ImagePullSecrets {
		imgPullSecrets = append(imgPullSecrets, corev1.LocalObjectReference{
			Name: secretName,
		})
	}

	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cfg.DaemonsetName,
			Annotations: cfg.DaemonsetAnnotations,
			OwnerReferences: []metav1.OwnerReference{
				getOwnerReferenceFromDeployment(deployment),
			},
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"test":                   "daemonset-test",
					"app.kubernetes.io/name": "image-puller",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"test":                   "daemonset-test",
						"app.kubernetes.io/name": "image-puller",
					},
					Name: "test-po",
				},
				Spec: corev1.PodSpec{
					SecurityContext:               podSecurityContext(openShift),
					NodeSelector:                  cfg.NodeSelector,
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					InitContainers: []corev1.Container{{
						Name:            "copy-sleep",
						Image:           cfg.ImagePullerImage,
						ImagePullPolicy: corev1.PullAlways,
						Command:         []string{"/bin/sh"},
						Args:            []string{"-c", copySleepCommand},
						VolumeMounts:    containerVolumeMounts,
						Resources:       getContainerResources(cfg),
						SecurityContext: initContainerSecurityContext(openShift),
					}},
					Containers:       getContainers(),
					ImagePullSecrets: imgPullSecrets,
					Affinity:         cfg.Affinity,
					Volumes: []corev1.Volume{{
						Name: kipVolumeName,
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{
								SizeLimit: &kipVolumeSizeLimit,
							},
						},
					}},
					Tolerations: cfg.Tolerations,
				},
			},
		},
	}
}

func podSecurityContext(openShift bool) *corev1.PodSecurityContext {
	ctx := &corev1.PodSecurityContext{
		SeccompProfile: &seccompProfile,
	}

	if !openShift {
		ctx.FSGroup = &nonRootGID
	}

	return ctx
}

func initContainerSecurityContext(openShift bool) *corev1.SecurityContext {
	ctx := &corev1.SecurityContext{
		RunAsNonRoot: ptr.To(true),
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{"ALL"},
		},
		ReadOnlyRootFilesystem:   ptr.To(true),
		AllowPrivilegeEscalation: ptr.To(false),
	}

	if !openShift {
		ctx.RunAsUser = &nonRootUID
		ctx.RunAsGroup = &nonRootGID
	}
	return ctx
}

// Create the daemonset, using to-be-cached images as init containers. Blocks
// until daemonset is ready.
func createDaemonsetOrDie(clientset *kubernetes.Clientset) {
	cfg := cfg.GetConfig()
	thisDeployment := getImagePullerDeployment(clientset)
	isOpenShiftInfra := isOpenShift(clientset)
	toCreate := getDaemonset(thisDeployment, isOpenShiftInfra)
	dsWatch := watchDaemonset(clientset)
	defer dsWatch.Stop()
	watchChan := dsWatch.ResultChan()

	_, err := clientset.AppsV1().DaemonSets(cfg.Namespace).Create(context.Background(), toCreate, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("Failed to create daemonset: %s", err.Error())
	} else {
		log.Printf("Created daemonset")
	}
	watchErr := waitDaemonsetReady(watchChan)
	if watchErr != nil {
		log.Printf("Unable to watch daemonset for readiness, falling back to manually checking.")
		checkDaemonsetReadiness(clientset)
	} else {
		log.Printf("Daemonset ready.")
	}
}

// Wait for daemonset to be ready (MODIFIED event with all nodes scheduled)
func waitDaemonsetReady(c <-chan watch.Event) error {
	log.Printf("Waiting for daemonset to be ready")
	for ev := range c {
		switch ev.Type {
		case watch.Modified:
			daemonset := ev.Object.(*appsv1.DaemonSet)
			if daemonset.Status.NumberReady == daemonset.Status.DesiredNumberScheduled {
				log.Printf("%d/%d nodes ready in daemonset", daemonset.Status.NumberReady, daemonset.Status.DesiredNumberScheduled)
				return nil
			}
			log.Printf("%d/%d nodes ready", daemonset.Status.NumberReady, daemonset.Status.DesiredNumberScheduled)
		case watch.Deleted, watch.Error:
			log.Fatalf("Error occurred while waiting for daemonset to be ready -- event %s detected", ev.Type)
		default:
			log.Printf("Unexpected watch event type while waiting for daemonset ready: %s", ev.Type)
		}
	}
	log.Printf("WARN: Watch closed before daemonset ready")
	return fmt.Errorf("watch closed before daemonset ready")
}

func checkDaemonsetReadiness(clientset *kubernetes.Clientset) {
	cfg := cfg.GetConfig()
	// Loop 30 times, sleeping for 3 seconds each time -- 90 seconds total wait.
	for i := 0; i < 30; i++ {
		ds, err := clientset.AppsV1().DaemonSets(cfg.Namespace).Get(context.Background(), cfg.DaemonsetName, metav1.GetOptions{
			// IncludeUninitialized: true,
		})
		if err != nil {
			log.Printf("WARN: could not get daemonset: %s", err)
			return
		}
		if ds.Status.DesiredNumberScheduled == 0 {
			// We've received a daemonset during initialization
			continue
		}
		log.Printf("%d/%d nodes ready in daemonset", ds.Status.NumberReady, ds.Status.DesiredNumberScheduled)
		if ds.Status.NumberReady == ds.Status.DesiredNumberScheduled {
			log.Printf("All nodes running")
			return
		}
		time.Sleep(3 * time.Second)
	}
	log.Printf("Maximum duration for readiness checking exceeded.")
}

// Delete daemonset with metadata.name daemonsetName. Blocks until daemonset
// is deleted.
func deleteDaemonset(clientset *kubernetes.Clientset) {
	log.Println("Deleting daemonset")
	cfg := cfg.GetConfig()

	dsWatch := watchDaemonset(clientset)
	defer dsWatch.Stop()
	watchChan := dsWatch.ResultChan()

	err := clientset.AppsV1().DaemonSets(cfg.Namespace).Delete(context.Background(), cfg.DaemonsetName, metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	if err != nil {
		log.Fatalf("Failed to delete daemonset %s", err.Error())
	} else {
		log.Printf("Deleted daemonset %s", cfg.DaemonsetName)
	}
	waitDaemonsetDeleted(watchChan)
}

// Use watch channel to wait for DELETED event on daemonset, then return
func waitDaemonsetDeleted(c <-chan watch.Event) {
	for ev := range c {
		switch ev.Type {
		case watch.Deleted:
			log.Printf("Daemonset deleted")
			return
		case watch.Error:
			log.Printf("Error event received while waiting for daemonset deletion: %s", ev.Type)
		default:
			log.Printf("Unexpected watch event type while waiting for daemonset deletion: %s", ev.Type)
		}
	}
	log.Printf("WARN: Watch closed before daemonset deleted")
}

// Get array of all images in containers to be cached.
func getContainers() []corev1.Container {
	cfg := cfg.GetConfig()
	images := cfg.Images
	containers := make([]corev1.Container, len(images))
	idx := 0

	for name, image := range images {
		containers[idx] = corev1.Container{
			Name:            name,
			Image:           image,
			Command:         []string{containerSleepCommand},
			Args:            []string{sleepDuration},
			Resources:       getContainerResources(cfg),
			ImagePullPolicy: corev1.PullAlways,
			VolumeMounts:    containerVolumeMounts,
			SecurityContext: &corev1.SecurityContext{
				Capabilities: &corev1.Capabilities{
					Drop: []corev1.Capability{"ALL"},
				},
				ReadOnlyRootFilesystem:   ptr.To(true),
				AllowPrivilegeEscalation: ptr.To(false),
			},
		}
		idx++
	}
	return containers
}

func getContainerResources(cfg cfg.Config) corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			"memory": resource.MustParse(cfg.CachingMemLimit),
			"cpu":    resource.MustParse(cfg.CachingCpuLimit),
		},
		Requests: corev1.ResourceList{
			"memory": resource.MustParse(cfg.CachingMemRequest),
			"cpu":    resource.MustParse(cfg.CachingCpuRequest),
		},
	}
}
