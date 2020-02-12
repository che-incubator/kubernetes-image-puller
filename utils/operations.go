//
// Copyright (c) 2019 Red Hat, Inc.
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
	"log"

	"github.com/che-incubator/kubernetes-image-puller/cfg"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CacheImages creates the daemonset responsible for ensuring images are cached
func CacheImages(clientset *kubernetes.Clientset) {
	log.Printf("Starting caching process")
	// Create daemonset, wait for it to be ready
	if err := createDaemonset(clientset); err != nil {
		log.Printf("Could not create Daemonset: %v", err)
	}
	log.Printf("Daemonset ready.")
}

// RefreshCache forces a refresh of all pods in the daemonset, to ensure images
// with mutable tags (e.g. nightlies) are up-to-date.
func RefreshCache(clientset *kubernetes.Clientset) {
	log.Printf("Refreshing cached images")
	DeleteDaemonsetIfExists(clientset)
	if err := createDaemonset(clientset); err != nil {
		log.Printf("Could not create Daemonset: %v", err)
	}
	log.Printf("Refreshed images")
}

// EnsureDaemonsetExists checks that the daemonset is still present, and
// recreates it if necessary
func EnsureDaemonsetExists(clientset *kubernetes.Clientset) {
	log.Printf("Checking that daemonset exists.")

	cfg := cfg.GetConfig()
	daemonset, err :=
		clientset.
			AppsV1().
			DaemonSets(cfg.Namespace).
			Get(cfg.DaemonsetName, metav1.GetOptions{})
	if err != nil || daemonset == nil {
		log.Printf("Recreating daemonset due to error")
		DeleteDaemonsetIfExists(clientset)
		CacheImages(clientset)
	}
}

// DeleteDaemonsetIfExists first checks if the daemonset exists, and deletes
// it if it does. Useful for ensuring no daemonset is already present from a
// previous rollout.
func DeleteDaemonsetIfExists(clientset *kubernetes.Clientset) {
	cfg := cfg.GetConfig()
	daemonset, err :=
		clientset.
			AppsV1().
			DaemonSets(cfg.Namespace).
			Get(cfg.DaemonsetName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		log.Fatalf("Error getting daemonset: %v", statusError.ErrStatus.Message)
	} else if err != nil {
		log.Fatalf(err.Error())
	}
	if daemonset != nil {
		deleteDaemonset(clientset)
		log.Printf("Deleted existing daemonset")
	}
}

// LogNumNodesScheduled logs the basic status of the daemonset.
func LogNumNodesScheduled(clientset *kubernetes.Clientset, user string) {
	cfg := cfg.GetConfig()
	daemonset, err :=
		clientset.
			AppsV1().
			DaemonSets(cfg.Namespace).
			Get(cfg.DaemonsetName, metav1.GetOptions{})
	if err != nil {
		log.Printf("Failed to get daemonset for user '%s': %s", user, err)
	}
	if daemonset != nil {
		log.Printf("Daemonset for user '%s': Ready: %d, Desired: %d",
			user,
			daemonset.Status.NumberReady,
			daemonset.Status.DesiredNumberScheduled)
	}
}
