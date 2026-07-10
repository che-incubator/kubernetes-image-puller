package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
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

	defaultCommand         = []string{"/kip/sleep"}
	defaultArgs            = []string{"720h"}
	defaultVolumeMounts    = []corev1.VolumeMount{{Name: "kip", MountPath: "/kip"}}
	defaultSecurityContext = corev1.SecurityContext{
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{"ALL"},
		},
		ReadOnlyRootFilesystem:   ptr.To(true),
		AllowPrivilegeEscalation: ptr.To(false),
	}
)

func TestGetDaemonsetPodSecurityContextKubernetes(t *testing.T) {
	t.Setenv("IMAGES", "che-code=quay.io/che-incubator/che-code:next;base-developer-image=quay.io/devfile/base-developer-image:ubi9-latest")
	t.Setenv("CACHING_INTERVAL_HOURS", "1")

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "test", UID: "test-uid"},
	}
	ds := getDaemonset(deployment, false)

	ctx := ds.Spec.Template.Spec.SecurityContext
	assert.NotNil(t, ctx, "PodSecurityContext should be set")
	assert.Nil(t, ctx.RunAsNonRoot, "RunAsNonRoot should not be set at pod level")
	assert.Nil(t, ctx.RunAsUser, "RunAsUser should not be set at pod level")
	assert.Nil(t, ctx.RunAsGroup, "RunAsGroup should not be set at pod level")
	assert.Equal(t, int64(65532), *ctx.FSGroup, "FSGroup should be 65532")
	assert.Equal(t, corev1.SeccompProfileTypeRuntimeDefault, ctx.SeccompProfile.Type, "SeccompProfile should be RuntimeDefault")

	initContainer := ds.Spec.Template.Spec.InitContainers[0]
	initCtx := initContainer.SecurityContext
	assert.NotNil(t, initCtx, "InitContainer SecurityContext should be set")
	assert.True(t, *initCtx.RunAsNonRoot, "InitContainer RunAsNonRoot should be true")
	assert.Equal(t, int64(65532), *initCtx.RunAsUser, "InitContainer RunAsUser should be 65532")
	assert.Equal(t, int64(65532), *initCtx.RunAsGroup, "InitContainer RunAsGroup should be 65532")
	assert.True(t, *initCtx.ReadOnlyRootFilesystem, "InitContainer ReadOnlyRootFilesystem should be true")
	assert.False(t, *initCtx.AllowPrivilegeEscalation, "InitContainer AllowPrivilegeEscalation should be false")
	assert.Equal(t, []corev1.Capability{"ALL"}, initCtx.Capabilities.Drop, "InitContainer should drop all capabilities")
}

func TestGetDaemonsetPodSecurityContextOpenShift(t *testing.T) {
	t.Setenv("IMAGES", "che-code=quay.io/che-incubator/che-code:next;base-developer-image=quay.io/devfile/base-developer-image:ubi9-latest")
	t.Setenv("CACHING_INTERVAL_HOURS", "1")

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "test", UID: "test-uid"},
	}
	ds := getDaemonset(deployment, true)

	ctx := ds.Spec.Template.Spec.SecurityContext
	assert.NotNil(t, ctx, "PodSecurityContext should be set")
	assert.Nil(t, ctx.FSGroup, "FSGroup should not be set on OpenShift")
	assert.Equal(t, corev1.SeccompProfileTypeRuntimeDefault, ctx.SeccompProfile.Type, "SeccompProfile should be RuntimeDefault")

	initContainer := ds.Spec.Template.Spec.InitContainers[0]
	initCtx := initContainer.SecurityContext
	assert.NotNil(t, initCtx, "InitContainer SecurityContext should be set")
	assert.True(t, *initCtx.RunAsNonRoot, "InitContainer RunAsNonRoot should be true")
	assert.Nil(t, initCtx.RunAsUser, "RunAsUser should not be set on OpenShift")
	assert.Nil(t, initCtx.RunAsGroup, "RunAsGroup should not be set on OpenShift")
	assert.True(t, *initCtx.ReadOnlyRootFilesystem, "InitContainer ReadOnlyRootFilesystem should be true")
	assert.False(t, *initCtx.AllowPrivilegeEscalation, "InitContainer AllowPrivilegeEscalation should be false")
	assert.Equal(t, []corev1.Capability{"ALL"}, initCtx.Capabilities.Drop, "InitContainer should drop all capabilities")
}

func TestGetDaemonsetEmptyDirVolume(t *testing.T) {
	t.Setenv("IMAGES", "che-code=quay.io/che-incubator/che-code:next;base-developer-image=quay.io/devfile/base-developer-image:ubi9-latest")
	t.Setenv("CACHING_INTERVAL_HOURS", "1")

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "test", UID: "test-uid"},
	}
	ds := getDaemonset(deployment, false)

	volumes := ds.Spec.Template.Spec.Volumes
	assert.Len(t, volumes, 1, "Should have exactly one volume")
	assert.Equal(t, "kip", volumes[0].Name)
	assert.NotNil(t, volumes[0].EmptyDir, "Volume should be EmptyDir")
	assert.Equal(t, resource.MustParse("50Mi"), *volumes[0].EmptyDir.SizeLimit, "SizeLimit should be 50Mi")
}

// This is the only function that does not require a kubernetes client.  The rest of the tests are in ./e2e
func TestGetContainers(t *testing.T) {
	type testcase struct {
		want   []corev1.Container
		name   string
		images string
	}

	testcases := []testcase{
		{
			name: "one container",
			want: []corev1.Container{{
				Name:            "che-code",
				Image:           "quay.io/che-incubator/che-code:next",
				Command:         defaultCommand,
				Args:            defaultArgs,
				ImagePullPolicy: corev1.PullAlways,
				Resources:       defaultResourceRequirements,
				VolumeMounts:    defaultVolumeMounts,
				SecurityContext: &defaultSecurityContext,
			}},
			images: "che-code=quay.io/che-incubator/che-code:next",
		}, {
			name: "two containers",
			want: []corev1.Container{{
				Name:            "che-code",
				Image:           "quay.io/che-incubator/che-code:next",
				Command:         defaultCommand,
				Args:            defaultArgs,
				ImagePullPolicy: corev1.PullAlways,
				Resources:       defaultResourceRequirements,
				VolumeMounts:    defaultVolumeMounts,
				SecurityContext: &defaultSecurityContext,
			}, {
				Name:            "base-developer-image",
				Image:           "quay.io/devfile/base-developer-image:ubi9-latest",
				Command:         defaultCommand,
				Args:            defaultArgs,
				ImagePullPolicy: corev1.PullAlways,
				Resources:       defaultResourceRequirements,
				VolumeMounts:    defaultVolumeMounts,
				SecurityContext: &defaultSecurityContext,
			}},
			images: "che-code=quay.io/che-incubator/che-code:next;base-developer-image=quay.io/devfile/base-developer-image:ubi9-latest",
		},
	}
	for _, c := range testcases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("IMAGES", c.images)
			t.Setenv("CACHING_INTERVAL_HOURS", "1")
			got := getContainers()
			assert.ElementsMatch(t, c.want, got, "Should contain the same elements, order is not guaranteed")
		})
	}
}
