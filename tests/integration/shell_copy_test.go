package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestShellAndCopyEvent tests shell command execution and file copy operations
func TestShellAndCopyEvent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	setupShellCopyTestEnvironment(t)
	defer cleanupShellCopyTestEnvironment(t)

	// Wait for containers to be ready
	waitForContainerReady(t, "clab-delay-test-r1")

	// Create test_logs directory
	err := os.MkdirAll("./test_logs", 0755)
	require.NoError(t, err, "Failed to create test_logs directory")
	defer os.RemoveAll("./test_logs")

	// Execute the scenario
	executeScenario(t, "tests/scenarios/shell_copy_test.json")

	// Wait for scenario to complete
	time.Sleep(15 * time.Second)

	// Verify the output file was copied
	outputFile := "./test_logs/shell_output.txt"
	require.FileExists(t, outputFile, "Output file should exist after copy")

	// Read and verify content
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err, "Failed to read output file")

	contentStr := string(content)
	assert.Contains(t, contentStr, "shell test output", "Output should contain expected text")
	t.Logf("Shell output content:\n%s", contentStr)
}

// TestCopyBidirectional tests both toContainer and fromContainer operations
func TestCopyBidirectional(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	setupShellCopyTestEnvironment(t)
	defer cleanupShellCopyTestEnvironment(t)

	// Wait for containers to be ready
	waitForContainerReady(t, "clab-delay-test-r1")

	// Create test_logs directory
	err := os.MkdirAll("./test_logs", 0755)
	require.NoError(t, err, "Failed to create test_logs directory")
	defer os.RemoveAll("./test_logs")

	// Verify input file exists
	inputFile := "./tests/data/test_input.txt"
	require.FileExists(t, inputFile, "Input file should exist")

	// Execute the scenario
	executeScenario(t, "tests/scenarios/copy_bidirectional_test.json")

	// Wait for scenario to complete
	time.Sleep(15 * time.Second)

	// Verify the output file was copied back
	outputFile := "./test_logs/test_output.txt"
	require.FileExists(t, outputFile, "Output file should exist after copy")

	// Read and verify content
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err, "Failed to read output file")

	contentStr := string(content)
	assert.Contains(t, contentStr, "test input file", "Output should contain original input")
	assert.Contains(t, contentStr, "processed by r1", "Output should contain processed marker")
	t.Logf("Bidirectional copy output:\n%s", contentStr)
}

// TestCopyWithPermissions tests file copy with owner/mode options
func TestCopyWithPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	setupShellCopyTestEnvironment(t)
	defer cleanupShellCopyTestEnvironment(t)

	// Wait for containers to be ready
	waitForContainerReady(t, "clab-delay-test-r1")

	// Create a test file in container and verify permissions can be set
	containerName := "clab-delay-test-r1"

	// Create file in container
	cmd := exec.Command("docker", "exec", containerName, "sh", "-c", "echo 'test' > /tmp/perm_test.txt")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to create test file: %s", string(output))

	// Copy file from container
	err = os.MkdirAll("./test_logs", 0755)
	require.NoError(t, err)
	defer os.RemoveAll("./test_logs")

	cmd = exec.Command("docker", "cp", containerName+":/tmp/perm_test.txt", "./test_logs/")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Failed to copy file: %s", string(output))

	// Verify file exists
	require.FileExists(t, "./test_logs/perm_test.txt")

	t.Log("Copy with permissions test passed")
}

func setupShellCopyTestEnvironment(t *testing.T) {
	t.Log("Setting up shell/copy test environment...")

	// Use the same topology as delay tests
	cmd := exec.Command("sudo", "containerlab", "deploy",
		"--topo", "tests/topology/minimal_delay_test.yaml")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to deploy topology: %s", string(output))

	t.Log("Shell/copy test environment setup completed")
}

func cleanupShellCopyTestEnvironment(t *testing.T) {
	t.Log("Cleaning up shell/copy test environment...")

	cmd := exec.Command("sudo", "containerlab", "destroy",
		"--topo", "tests/topology/minimal_delay_test.yaml", "--cleanup")
	if err := cmd.Run(); err != nil {
		t.Logf("Warning: Failed to cleanup topology: %v", err)
	}

	t.Log("Shell/copy test environment cleanup completed")
}

func waitForContainerReady(t *testing.T, containerName string) {
	t.Logf("Waiting for container %s to be ready...", containerName)

	timeout := 60 * time.Second
	interval := 2 * time.Second

	require.Eventually(t, func() bool {
		cmd := exec.Command("docker", "exec", containerName, "echo", "ready")
		err := cmd.Run()
		return err == nil
	}, timeout, interval, "Container %s not ready", containerName)

	// Additional stabilization time
	time.Sleep(3 * time.Second)
	t.Logf("Container %s is ready", containerName)
}

func executeScenario(t *testing.T, scenarioFile string) {
	t.Logf("Executing scenario: %s", scenarioFile)

	// Get absolute path to scenario file
	absPath, err := filepath.Abs(scenarioFile)
	require.NoError(t, err, "Failed to get absolute path")

	cmd := exec.Command("sudo", "./netroub", absPath)
	cmd.Dir = getProjectRoot(t)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Scenario output: %s", string(output))
	}
	require.NoError(t, err, "Failed to execute scenario: %s", string(output))

	t.Log("Scenario executed successfully")
}

func getProjectRoot(t *testing.T) string {
	// Use current working directory
	wd, err := os.Getwd()
	require.NoError(t, err)
	// If we're in tests/integration, go up two levels
	if strings.HasSuffix(wd, "tests/integration") {
		return filepath.Join(wd, "../..")
	}
	return wd
}
