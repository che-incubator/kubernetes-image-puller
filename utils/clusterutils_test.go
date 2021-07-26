package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	defaultMemoryLimit          = resource.MustParse("5Mi")
	defaultMemoryRequest        = resource.MustParse("1Mi")
	defaultCpuLimit             = resource.MustParse(".2")
	defaultCpuRequest           = resource.MustParse(".05")
	defaultResourceRequirements = corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			"memory": defaultMemoryLimit,
			"cpu":    defaultCpuLimit,
		},
		Requests: corev1.ResourceList{
			"memory": defaultMemoryRequest,
			"cpu":    defaultCpuRequest,
		},
	}

	defaultCommand      = []string{"/kip/sleep"}
	defaultArgs         = []string{"720h"}
	defaultVolumeMounts = []corev1.VolumeMount{{Name: "kip", MountPath: "/kip"}}
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
				Name:            "che-theia",
				Image:           "eclipse/che-theia:nightly",
				Command:         defaultCommand,
				Args:            defaultArgs,
				ImagePullPolicy: corev1.PullAlways,
				Resources:       defaultResourceRequirements,
				VolumeMounts:    defaultVolumeMounts,
			}, {
				Name:            "che-plugin-registry",
				Image:           "quay.io/eclipse/che-plugin-registry:nightly",
				Command:         defaultCommand,
				Args:            defaultArgs,
				ImagePullPolicy: corev1.PullAlways,
				Resources:       defaultResourceRequirements,
				VolumeMounts:    defaultVolumeMounts,
			}},
			images: "che-theia=eclipse/che-theia:nightly;che-plugin-registry=quay.io/eclipse/che-plugin-registry:nightly",
		}, {
			name: "four containers",
			want: []corev1.Container{{
				Name:            "che-sidecar-java",
				Image:           "quay.io/eclipse/che-sidecar-java:nightly",
				Command:         defaultCommand,
				Args:            defaultArgs,
				ImagePullPolicy: corev1.PullAlways,
				Resources:       defaultResourceRequirements,
				VolumeMounts:    defaultVolumeMounts,
			}, {
				Name:            "che-plugin-registry",
				Image:           "quay.io/eclipse/che-plugin-registry:nightly",
				Command:         defaultCommand,
				Args:            defaultArgs,
				ImagePullPolicy: corev1.PullAlways,
				Resources:       defaultResourceRequirements,
				VolumeMounts:    defaultVolumeMounts,
			}, {
				Name:            "che-devfile-registry",
				Image:           "quay.io/eclipse/che-devfile-registry:nightly",
				Command:         defaultCommand,
				Args:            defaultArgs,
				ImagePullPolicy: corev1.PullAlways,
				Resources:       defaultResourceRequirements,
				VolumeMounts:    defaultVolumeMounts,
			}, {
				Name:            "che-theia",
				Image:           "quay.io/eclipse/che-theia:nightly",
				Command:         defaultCommand,
				Args:            defaultArgs,
				ImagePullPolicy: corev1.PullAlways,
				Resources:       defaultResourceRequirements,
				VolumeMounts:    defaultVolumeMounts,
			}},
			images: "che-sidecar-java=quay.io/eclipse/che-sidecar-java:nightly;che-plugin-registry=quay.io/eclipse/che-plugin-registry:nightly;che-devfile-registry=quay.io/eclipse/che-devfile-registry:nightly;che-theia=quay.io/eclipse/che-theia:nightly",
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
