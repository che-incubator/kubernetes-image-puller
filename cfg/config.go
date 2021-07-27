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

import corev1 "k8s.io/api/core/v1"

type Config struct {
	DaemonsetName     string
	Namespace         string
	Images            map[string]string
	CachingMemRequest string
	CachingMemLimit   string
	CachingCpuRequest string
	CachingCpuLimit   string
	CachingInterval   int
	NodeSelector      map[string]string
	ImagePullSecrets  []string
	Affinity          *corev1.Affinity
	ImagePullerImage  string
}

func GetConfig() Config {
	return Config{
		DaemonsetName:     getEnvVarOrDefault(daemonsetNameEnvVar, defaultDaemonsetName),
		Namespace:         getEnvVarOrDefault(namespaceEnvVar, defaultNamespace),
		Images:            processImagesEnvVar(),
		CachingInterval:   getCachingInterval(),
		CachingMemRequest: getEnvVarOrDefault(cachingMemRequestEnvVar, defaultCachingMemRequest),
		CachingMemLimit:   getEnvVarOrDefault(cachingMemLimitEnvVar, defaultCachingMemLimit),
		CachingCpuRequest: getEnvVarOrDefault(cachingCpuRequestEnvVar, defaultCachingCpuRequest),
		CachingCpuLimit:   getEnvVarOrDefault(cachingCpuLimitEnvVar, defaultCachingCpuLimit),
		NodeSelector:      processNodeSelectorEnvVar(),
		ImagePullSecrets:  processImagePullSecretsEnvVar(),
		Affinity:          processAffinityEnvVar(),
		ImagePullerImage:  getEnvVarOrDefault(kipImageEnvVar, defaultImage),
	}
}
