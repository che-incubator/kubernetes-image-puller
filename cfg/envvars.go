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

package cfg

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// Env vars used for configuration
const (
	intervalEnvVar          = "CACHING_INTERVAL_HOURS"
	daemonsetNameEnvVar     = "DAEMONSET_NAME"
	namespaceEnvVar         = "NAMESPACE"
	imagesEnvVar            = "IMAGES"
	cachingMemRequestEnvVar = "CACHING_MEMORY_REQUEST"
	cachingMemLimitEnvVar   = "CACHING_MEMORY_LIMIT"
	cachingCpuRequestEnvVar = "CACHING_CPU_REQUEST"
	cachingCpuLimitEnvVar   = "CACHING_CPU_LIMIT"
	nodeSelectorEnvVar      = "NODE_SELECTOR"
	imagePullSecretsEnvVar  = "IMAGE_PULL_SECRETS"
	affinityEnvVar          = "AFFINITY"
	kipImageEnvVar          = "KIP_IMAGE"
)

// Default values where applicable
const (
	defaultDeploymentName    = "kubernetes-image-puller"
	defaultDaemonsetName     = "kubernetes-image-puller"
	defaultNamespace         = "k8s-image-puller"
	defaultCachingMemRequest = "1Mi"
	defaultCachingMemLimit   = "5Mi"
	defaultCachingInterval   = 1
	defaultCachingCpuRequest = ".05"
	defaultCachingCpuLimit   = ".2"
	defaultNodeSelector      = "{}"
	defaultImagePullSecret   = ""
	defaultAffinity          = "{}"
	defaultImage             = "quay.io/eclipse/kubernetes-image-puller:next"
)

func getCachingInterval() int {
	cachingIntervalStr := getEnvVarOrExit(intervalEnvVar)
	interval, err := strconv.Atoi(cachingIntervalStr)
	if err != nil {
		log.Printf(
			"Could not parse env var %s to integer. Value is %s. Using default of %d",
			intervalEnvVar,
			cachingIntervalStr,
			defaultCachingInterval)
		return defaultCachingInterval
	}
	return interval
}

func processImagesEnvVar() map[string]string {
	rawImages := getEnvVarOrExit(imagesEnvVar)
	rawImages = strings.TrimSpace(rawImages)
	images := strings.Split(rawImages, ";")
	for i, image := range images {
		images[i] = strings.TrimSpace(image)
	}
	// If last element is empty, remove it
	if images[len(images)-1] == "" {
		images = images[:len(images)-1]
	}

	log.Printf("Processing images from configuration...")
	var imagesMap = make(map[string]string)
	for _, image := range images {
		log.Printf("Image: %s", image)
		nameAndImage := strings.Split(image, "=")
		if len(nameAndImage) != 2 {
			log.Printf("Malformed image name/tag: %s. Ignoring.", image)
			continue
		}
		imagesMap[nameAndImage[0]] = nameAndImage[1]
	}
	return imagesMap
}

func processNodeSelectorEnvVar() map[string]string {
	rawNodeSelector := getEnvVarOrDefault(nodeSelectorEnvVar, defaultNodeSelector)
	nodeSelector := make(map[string]string)
	err := json.Unmarshal([]byte(rawNodeSelector), &nodeSelector)
	if err != nil {
		log.Fatalf("Failed to unmarshal node selector json: %s", err)
	}
	return nodeSelector
}

func processImagePullSecretsEnvVar() []string {
	rawImagePullSecrets := getEnvVarOrDefault(imagePullSecretsEnvVar, defaultImagePullSecret)
	rawImagePullSecrets = strings.TrimSpace(rawImagePullSecrets)
	pullSecrets := strings.Split(rawImagePullSecrets, ";")
	for i, secret := range pullSecrets {
		pullSecrets[i] = strings.TrimSpace(secret)
	}
	// If last element is empty, remove it
	if pullSecrets[len(pullSecrets)-1] == "" {
		pullSecrets = pullSecrets[:len(pullSecrets)-1]
	}

	return pullSecrets
}

func processAffinityEnvVar() *corev1.Affinity {
	rawAffinity := getEnvVarOrDefault(affinityEnvVar, defaultAffinity)
	affinity := &corev1.Affinity{}
	if err := json.Unmarshal([]byte(rawAffinity), affinity); err != nil {
		log.Fatalf("Failed to unmarshal affinity json: %s", err)
	}
	return affinity
}

func getEnvVarOrExit(envVar string) string {
	val := os.Getenv(envVar)
	if val == "" {
		log.Fatalf("Env var %s unset. Aborting", envVar)
	}
	return val
}

func getEnvVarOrDefault(envVar, defaultValue string) string {
	val := os.Getenv(envVar)
	if val == "" {
		log.Printf("No value found for %s. Using default value of %s", envVar, defaultValue)
		val = defaultValue
	}
	return val
}

func getEnvVarOrDefaultBool(envVar string, defaultValue bool) bool {
	envvar := os.Getenv(envVar)
	val, err := strconv.ParseBool(envvar)
	if err != nil {
		val = defaultValue
	}
	return val
}
