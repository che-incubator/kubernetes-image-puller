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

package main

import (
	"log"
	"os"

	"github.com/redhat-developer/kubernetes-image-puller/cfg"
	multicluster "github.com/redhat-developer/kubernetes-image-puller/pkg/multi-cluster"
	singlecluster "github.com/redhat-developer/kubernetes-image-puller/pkg/single-cluster"
)

func main() {
	log.SetOutput(os.Stdout)

	if cfg.MultiCluster {
		multicluster.CacheImages()
	} else {
		log.Printf("Running in single-cluster mode")
		singlecluster.CacheImages()
	}
}
