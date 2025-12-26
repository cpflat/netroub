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

	projectRoot := getProjectRoot(t)

	// Ensure no existing containers (cleanup before test)
	cleanupShellCopyTestEnvironment(t)

	// Create test_logs directory
	testLogsDir := filepath.Join(projectRoot, "test_logs")
	err := os.MkdirAll(testLogsDir, 0755)
	require.NoError(t, err, "Failed to create test_logs directory")
	defer os.RemoveAll(testLogsDir)

	// Execute the scenario (netroub handles deploy/destroy)
	executeScenario(t, filepath.Join(projectRoot, "tests/scenarios/shell_copy_test.json"))

	// Verify the output file was copied
	outputFile := filepath.Join(testLogsDir, "shell_output.txt")
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

	projectRoot := getProjectRoot(t)

	// Ensure no existing containers (cleanup before test)
	cleanupShellCopyTestEnvironment(t)

	// Create test_logs directory
	testLogsDir := filepath.Join(projectRoot, "test_logs")
	err := os.MkdirAll(testLogsDir, 0755)
	require.NoError(t, err, "Failed to create test_logs directory")
	defer os.RemoveAll(testLogsDir)

	// Verify input file exists
	inputFile := filepath.Join(projectRoot, "tests/data/test_input.txt")
	require.FileExists(t, inputFile, "Input file should exist")

	// Execute the scenario (netroub handles deploy/destroy)
	executeScenario(t, filepath.Join(projectRoot, "tests/scenarios/copy_bidirectional_test.json"))

	// Verify the output file was copied back
	outputFile := filepath.Join(testLogsDir, "test_output.txt")
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
// This test uses containerlab directly (not via netroub) to test docker cp functionality
func TestCopyWithPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	projectRoot := getProjectRoot(t)

	// Ensure no existing containers and deploy fresh
	cleanupShellCopyTestEnvironment(t)
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
	testLogsDir := filepath.Join(projectRoot, "test_logs")
	err = os.MkdirAll(testLogsDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(testLogsDir)

	cmd = exec.Command("docker", "cp", containerName+":/tmp/perm_test.txt", testLogsDir+"/")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Failed to copy file: %s", string(output))

	// Verify file exists
	require.FileExists(t, filepath.Join(testLogsDir, "perm_test.txt"))

	t.Log("Copy with permissions test passed")
}

func setupShellCopyTestEnvironment(t *testing.T) {
	t.Log("Setting up shell/copy test environment...")

	// Get project root directory
	projectRoot := getProjectRoot(t)
	topoPath := filepath.Join(projectRoot, "tests/topology/minimal_delay_test.yaml")

	// Use the same topology as delay tests
	cmd := exec.Command("sudo", "containerlab", "deploy", "--topo", topoPath)
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to deploy topology: %s", string(output))

	t.Log("Shell/copy test environment setup completed")
}

func cleanupShellCopyTestEnvironment(t *testing.T) {
	t.Log("Cleaning up shell/copy test environment...")

	projectRoot := getProjectRoot(t)
	topoPath := filepath.Join(projectRoot, "tests/topology/minimal_delay_test.yaml")

	cmd := exec.Command("sudo", "containerlab", "destroy", "--topo", topoPath, "--cleanup")
	cmd.Dir = projectRoot
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

	projectRoot := getProjectRoot(t)
	netroubPath := filepath.Join(projectRoot, "netroub")

	cmd := exec.Command("sudo", netroubPath, scenarioFile)
	cmd.Dir = projectRoot

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
