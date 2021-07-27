package cfg

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
)

func TestEnvVars(t *testing.T) {

	defer unsetEnv()

	os.Setenv("IMAGES", "che-theia=quay.io/eclipse/che-theia:nightly")
	os.Setenv("CACHING_INTERVAL_HOURS", "5")

	type testcase struct {
		name string
		env  map[string]string
		want Config
	}

	cases := []testcase{
		{
			name: "default",
			env:  map[string]string{},
			want: Config{
				DaemonsetName: "kubernetes-image-puller",
				Namespace:     "k8s-image-puller",
				Images: map[string]string{
					"che-theia": "quay.io/eclipse/che-theia:nightly",
				},
				CachingMemRequest: "1Mi",
				CachingMemLimit:   "5Mi",
				CachingCpuRequest: ".05",
				CachingCpuLimit:   ".2",
				CachingInterval:   5,
				NodeSelector:      map[string]string{},
				ImagePullSecrets:  []string{},
				Affinity:          &v1.Affinity{},
				ImagePullerImage:  "quay.io/eclipse/kubernetes-image-puller:next",
			},
		},
		{
			name: "overrides",
			env: map[string]string{
				"DAEMONSET_NAME":      "custom-daemonset-name",
				"NAMESPACE":           "my-namespace",
				"NODE_SELECTOR":       "{\"type\": \"compute\"}",
				"CACHING_CPU_REQUEST": ".055",
				"IMAGE_PULL_SECRETS":  "secret1; secret2",
				"AFFINITY":            `{"nodeAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":{"nodeSelectorTerms":[{"matchExpressions":[{"key":"kubernetes.io/e2e-az-name","operator":"In","values":["e2e-az1","e2e-az2"]}]}]}}}`,
				"KIP_IMAGE":           "quay.io/my-repo/kubernetes-image-puller:next",
			},
			want: Config{
				DaemonsetName: "custom-daemonset-name",
				Namespace:     "my-namespace",
				Images: map[string]string{
					"che-theia": "quay.io/eclipse/che-theia:nightly",
				},
				CachingMemRequest: "1Mi",
				CachingMemLimit:   "5Mi",
				CachingCpuRequest: ".055",
				CachingCpuLimit:   ".2",
				CachingInterval:   5,
				NodeSelector: map[string]string{
					"type": "compute",
				},
				ImagePullSecrets: []string{"secret1", "secret2"},
				Affinity: &v1.Affinity{
					NodeAffinity: &v1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
							NodeSelectorTerms: []v1.NodeSelectorTerm{
								{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "kubernetes.io/e2e-az-name",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"e2e-az1", "e2e-az2"},
										},
									},
								},
							},
						},
					},
				},
				ImagePullerImage: "quay.io/my-repo/kubernetes-image-puller:next",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for k, v := range c.env {
				os.Setenv(k, v)
			}
			cfg := GetConfig()
			if d := cmp.Diff(c.want, cfg); d != "" {
				t.Errorf("Diff (-want, +got): %s", d)
			}
		})
	}
}

func unsetEnv() {
	os.Unsetenv("IMAGES")
	os.Unsetenv("CACHING_INTERVAL_HOURS")
}
