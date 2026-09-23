package e2e_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

func readInferenceManifest(t *testing.T, name string) map[string]any {
	t.Helper()
	file, err := os.Open(filepath.Join("prerequisites", "inference", name))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, file.Close()) })

	manifest := map[string]any{}
	require.NoError(t, utilyaml.NewYAMLOrJSONDecoder(file, 4096).Decode(&manifest))
	return manifest
}

func TestLeaderWorkerSetOperatorManifest(t *testing.T) {
	manifest := readInferenceManifest(t, "lwsoperator.yaml")
	require.Equal(t, "operator.openshift.io/v1", manifest["apiVersion"])
	require.Equal(t, "LeaderWorkerSetOperator", manifest["kind"])

	metadata, ok := manifest["metadata"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "cluster", metadata["name"])
	require.Equal(t, "openshift-lws-operator", metadata["namespace"])
}

func TestConnectivityLinkOperatorConfiguration(t *testing.T) {
	require.Equal(t, "rhcl-operator", connectivityLinkOpName)
	require.Equal(t, "openshift-operators", connectivityLinkOpNamespace)
	require.Equal(t, "stable", connectivityLinkOpChannel)
	require.Equal(t, "redhat-operators", connectivityLinkOpSource)
}

func TestDSCIManifest(t *testing.T) {
	manifest := readInferenceManifest(t, "dsci.yaml")
	require.Equal(t, "dscinitialization.opendatahub.io/v2", manifest["apiVersion"])
	require.Equal(t, "DSCInitialization", manifest["kind"])

	spec, ok := manifest["spec"].(map[string]any)
	require.True(t, ok)
	_, hasAlerting := spec["alerting"]
	require.False(t, hasAlerting)
}

func TestKuadrantManifest(t *testing.T) {
	manifest := readInferenceManifest(t, "kuadrant.yaml")
	require.Equal(t, "kuadrant.io/v1beta1", manifest["apiVersion"])
	require.Equal(t, "Kuadrant", manifest["kind"])

	metadata, ok := manifest["metadata"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "kuadrant", metadata["name"])
	require.Equal(t, "kuadrant-system", metadata["namespace"])
}
