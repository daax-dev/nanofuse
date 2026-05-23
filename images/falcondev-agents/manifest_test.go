package falcondevagents_test

import (
	"testing"

	"github.com/daax-dev/nanofuse/internal/layerbuild"
)

func TestFalconDevAgentsManifest(t *testing.T) {
	manifest, err := layerbuild.ParseManifest("image.manifest.yaml")
	if err != nil {
		t.Fatalf("failed to parse manifest: %v", err)
	}

	// Validate manifest structure
	if err := layerbuild.ValidateManifest(manifest); err != nil {
		t.Fatalf("manifest validation failed: %v", err)
	}

	// Check basic properties
	if manifest.Name != "falcondev-agents" {
		t.Errorf("expected name 'falcondev-agents', got '%s'", manifest.Name)
	}

	if manifest.Version != "1.0" {
		t.Errorf("expected version '1.0', got '%s'", manifest.Version)
	}

	// Verify kernel config
	if manifest.Kernel.Version == "" {
		t.Error("kernel version is required")
	}
	if manifest.Kernel.Source == "" {
		t.Error("kernel source is required")
	}
	if manifest.Kernel.Cmdline == "" {
		t.Error("kernel cmdline is required")
	}

	// Verify layers
	expectedLayers := []string{"base-os", "python-runtime", "node-runtime", "recording-agent", "agent-tools"}
	if len(manifest.Layers) != len(expectedLayers) {
		t.Errorf("expected %d layers, got %d", len(expectedLayers), len(manifest.Layers))
	}

	layerNames := make(map[string]bool)
	for _, layer := range manifest.Layers {
		layerNames[layer.Name] = true
	}

	for _, expected := range expectedLayers {
		if !layerNames[expected] {
			t.Errorf("missing expected layer: %s", expected)
		}
	}

	// Verify output config
	if manifest.Output.Format != "ext4" {
		t.Errorf("expected output format 'ext4', got '%s'", manifest.Output.Format)
	}
}

func TestFalconDevAgentsConditions(t *testing.T) {
	manifest, err := layerbuild.ParseManifest("image.manifest.yaml")
	if err != nil {
		t.Fatalf("failed to parse manifest: %v", err)
	}

	// Test with recording enabled (default)
	activeLayers := layerbuild.EvaluateConditions(manifest, nil)

	// All layers should be active by default
	if len(activeLayers) != 5 {
		t.Errorf("expected 5 active layers with default conditions, got %d", len(activeLayers))
	}

	// Test with recording disabled
	env := map[string]string{"INCLUDE_RECORDING": "false"}
	activeLayers = layerbuild.EvaluateConditions(manifest, env)

	// recording-agent should be excluded
	if len(activeLayers) != 4 {
		t.Errorf("expected 4 active layers with INCLUDE_RECORDING=false, got %d", len(activeLayers))
	}

	for _, layer := range activeLayers {
		if layer.Name == "recording-agent" {
			t.Error("recording-agent should be excluded when INCLUDE_RECORDING=false")
		}
	}
}

func TestFalconDevAgentsDependencies(t *testing.T) {
	manifest, err := layerbuild.ParseManifest("image.manifest.yaml")
	if err != nil {
		t.Fatalf("failed to parse manifest: %v", err)
	}

	activeLayers := layerbuild.EvaluateConditions(manifest, nil)

	sorted, err := layerbuild.ResolveDependencies(activeLayers)
	if err != nil {
		t.Fatalf("dependency resolution failed: %v", err)
	}

	// base-os must come first
	if sorted[0].Name != "base-os" {
		t.Errorf("expected base-os first, got %s", sorted[0].Name)
	}

	// agent-tools must come after python-runtime and node-runtime
	var pythonIdx, nodeIdx, agentToolsIdx int
	for i, layer := range sorted {
		switch layer.Name {
		case "python-runtime":
			pythonIdx = i
		case "node-runtime":
			nodeIdx = i
		case "agent-tools":
			agentToolsIdx = i
		}
	}

	if agentToolsIdx < pythonIdx || agentToolsIdx < nodeIdx {
		t.Error("agent-tools must come after python-runtime and node-runtime")
	}
}

func TestFalconDevAgentsLayerTypes(t *testing.T) {
	manifest, err := layerbuild.ParseManifest("image.manifest.yaml")
	if err != nil {
		t.Fatalf("failed to parse manifest: %v", err)
	}

	// Verify each layer has a valid type
	expectedTypes := map[string]layerbuild.LayerType{
		"base-os":         layerbuild.LayerTypeBase,
		"python-runtime":  layerbuild.LayerTypeRuntime,
		"node-runtime":    layerbuild.LayerTypeRuntime,
		"recording-agent": layerbuild.LayerTypeFeature,
		"agent-tools":     layerbuild.LayerTypeApplication,
	}

	for _, layer := range manifest.Layers {
		expectedType, ok := expectedTypes[layer.Name]
		if !ok {
			t.Errorf("unexpected layer: %s", layer.Name)
			continue
		}
		if layer.Type != expectedType {
			t.Errorf("layer %s: expected type %s, got %s", layer.Name, expectedType, layer.Type)
		}
	}
}

func TestFalconDevAgentsSources(t *testing.T) {
	manifest, err := layerbuild.ParseManifest("image.manifest.yaml")
	if err != nil {
		t.Fatalf("failed to parse manifest: %v", err)
	}

	// Verify source types are valid
	// All layers in our manifest use local:// sources (we build them ourselves)
	for _, layer := range manifest.Layers {
		sourceType, ok := layerbuild.ParseSourceType(layer.Source)
		if !ok {
			t.Errorf("layer %s has invalid source: %s", layer.Name, layer.Source)
		}

		if sourceType != layerbuild.SourceTypeLocal {
			t.Errorf("layer %s should have local:// source, got %s", layer.Name, sourceType)
		}
	}
}
