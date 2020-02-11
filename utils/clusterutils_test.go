package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	defaultLimit                = resource.MustParse("5Mi")
	defaultRequest              = resource.MustParse("1Mi")
	defaultResourceRequirements = corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			"memory": defaultLimit,
		},
		Requests: corev1.ResourceList{
			"memory": defaultRequest,
		},
	}

	defaultCommand = []string{"sleep"}
	defaultArgs    = []string{"30d"}
)

// This is the only function that does not require a kubernetes client.  The rest of the tests are in ./e2e
func TestGetContainers(t *testing.T) {
	type testcase struct {
		want   []corev1.Container
		name   string
		images string
	}

	testcases := []testcase{
		{
			name: "two containers",
			want: []corev1.Container{{
				Name:            "che-server",
				Image:           "eclipse/che-server:nightly",
				Command:         defaultCommand,
				Args:            defaultArgs,
				ImagePullPolicy: corev1.PullAlways,
				Resources:       defaultResourceRequirements,
			}, {
				Name:            "che-plugin-registry",
				Image:           "quay.io/eclipse/che-plugin-registry:nightly",
				Command:         defaultCommand,
				Args:            defaultArgs,
				ImagePullPolicy: corev1.PullAlways,
				Resources:       defaultResourceRequirements,
			}},
			images: "che-server=eclipse/che-server:nightly;che-plugin-registry=quay.io/eclipse/che-plugin-registry:nightly",
		}, {
			name: "four containers",
			want: []corev1.Container{{
				Name:            "che-server",
				Image:           "eclipse/che-server:nightly",
				Command:         defaultCommand,
				Args:            defaultArgs,
				ImagePullPolicy: corev1.PullAlways,
				Resources:       defaultResourceRequirements,
			}, {
				Name:            "che-plugin-registry",
				Image:           "quay.io/eclipse/che-plugin-registry:nightly",
				Command:         defaultCommand,
				Args:            defaultArgs,
				ImagePullPolicy: corev1.PullAlways,
				Resources:       defaultResourceRequirements,
			}, {
				Name:            "che-devfile-registry",
				Image:           "quay.io/eclipse/che-devfile-registry:nightly",
				Command:         defaultCommand,
				Args:            defaultArgs,
				ImagePullPolicy: corev1.PullAlways,
				Resources:       defaultResourceRequirements,
			}, {
				Name:            "che-machine-exec",
				Image:           "quay.io/eclipse/che-machine-exec:nightly",
				Command:         defaultCommand,
				Args:            defaultArgs,
				ImagePullPolicy: corev1.PullAlways,
				Resources:       defaultResourceRequirements,
			}},
			images: "che-server=eclipse/che-server:nightly;che-plugin-registry=quay.io/eclipse/che-plugin-registry:nightly;che-devfile-registry=quay.io/eclipse/che-devfile-registry:nightly;che-machine-exec=quay.io/eclipse/che-machine-exec:nightly",
		},
	}
	for _, c := range testcases {
		t.Run(c.name, func(t *testing.T) {
			defer os.Clearenv()
			os.Setenv("IMAGES", c.images)
			os.Setenv("CACHING_INTERVAL_HOURS", "1")
			got := getContainers()
			assert.ElementsMatch(t, c.want, got, "Should contain the same elements, order is not guaranteed")
		})
	}
}
