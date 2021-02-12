package cfg

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestProcessImagesEnvVar(t *testing.T) {
	type testcase struct {
		name   string
		images string
		want   map[string]string
	}

	testcases := []testcase{
		{
			name:   "one image",
			images: "che-theia=quay.io/eclipse/che-theia:nightly",
			want:   map[string]string{"che-theia": "quay.io/eclipse/che-theia:nightly"},
		},
		{
			name:   "three images",
			images: "image1=my/image1:dev;image2=my/image2:next;image3=my/image3:stage",
			want: map[string]string{
				"image1": "my/image1:dev",
				"image2": "my/image2:next",
				"image3": "my/image3:stage",
			},
		},
	}

	for _, c := range testcases {
		t.Run(c.name, func(t *testing.T) {
			defer os.Clearenv()
			os.Setenv("IMAGES", c.images)
			got := processImagesEnvVar()

			if d := cmp.Diff(c.want, got); d != "" {
				t.Errorf("(-want, +got): %s", d)
			}
		})
	}
}

func TestProcessNodeSElectorEnvVar(t *testing.T) {
	type testcase struct {
		name              string
		nodeSelector      string
		isNodeSelectorSet bool
		want              map[string]string
	}

	testcases := []testcase{
		{
			name:              "default node selector, NODE_SELECTOR set",
			nodeSelector:      "{}",
			isNodeSelectorSet: true,
			want:              map[string]string{},
		},
		{
			name:              "compute type, NODE_SELECTOR set",
			nodeSelector:      "{\"type\": \"compute\"}",
			isNodeSelectorSet: true,
			want: map[string]string{
				"type": "compute",
			},
		},
		{
			name:              "default env var, NODE_SELECTOR not set",
			nodeSelector:      "{\"this\": \"shouldn't be set\"}",
			isNodeSelectorSet: false,
			want:              map[string]string{},
		},
	}

	for _, c := range testcases {
		t.Run(c.name, func(t *testing.T) {
			defer os.Clearenv()
			if c.isNodeSelectorSet {
				os.Setenv("NODE_SELECTOR", c.nodeSelector)
			}
			got := processNodeSelectorEnvVar()

			if d := cmp.Diff(c.want, got); d != "" {
				t.Errorf("(-want, +got): %s", d)
			}
		})
	}
}

func TestGetEnvVarOrDefaultBool(t *testing.T) {
	defer os.Clearenv()
	os.Setenv("DEFINED_ENV_VAR", "true")
	assert.True(t, getEnvVarOrDefaultBool("DEFINED_ENV_VAR", false), "When a variable is defined it should return its value")
	assert.True(t, getEnvVarOrDefaultBool("UNDEFINED_ENV_VAR", true), "When a variable is not defined it should return the default value")
}

func TestGetEnvVarOrDefault(t *testing.T) {
	defer os.Clearenv()
	os.Setenv("DEFINED_ENV_VAR", "foo")
	assert.Equal(t, getEnvVarOrDefault("DEFINED_ENV_VAR", "bar"), "foo", "When a variable is defined it should return it's set value")
	assert.Equal(t, getEnvVarOrDefault("UNDEFINED_ENV_VAR", "bar"), "bar", "When a variable is undefined it should return the default value")
}

func Test_processAffinityEnvVar(t *testing.T) {
	type testcase struct {
		name          string
		affinity      string
		isAffinitySet bool
		want          *v1.Affinity
	}

	tests := []testcase{
		{
			name:          "default affinity, AFFINITY set",
			affinity:      "{}",
			isAffinitySet: true,
			want:          &v1.Affinity{},
		},
		{
			name:          "requiredDuringSchedulingIgnoredDuringExecution, AFFINITY set",
			affinity:      "{\"nodeAffinity\":{\"requiredDuringSchedulingIgnoredDuringExecution\":{\"nodeSelectorTerms\":[{\"matchExpressions\":[{\"key\":\"kubernetes.io/e2e-az-name\",\"operator\":\"In\",\"values\":[\"e2e-az1\",\"e2e-az2\"]}]}]}}}",
			isAffinitySet: true,
			want: &v1.Affinity{
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
		},
		{
			name:          "default env var, AFFINITY not set",
			affinity:      "{\"this\": \"shouldn't be set\"}",
			isAffinitySet: false,
			want:          &v1.Affinity{},
		},
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			defer os.Clearenv()
			if c.isAffinitySet {
				os.Setenv("AFFINITY", c.affinity)
			}
			got := processAffinityEnvVar()

			if d := cmp.Diff(c.want, got); d != "" {
				t.Errorf("(-want, +got): %s", d)
			}
		})
	}
}
