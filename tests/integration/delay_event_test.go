package integration

import (
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDelayEvent tests the delay event functionality in a real Docker environment
func TestDelayEvent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Phase 1: Setup and wait for network readiness
	setupTestEnvironment(t)
	defer cleanupTestEnvironment(t)

	waitForNetworkReady(t, []string{"r1", "r2"})

	// Phase 2: Baseline measurement
	baselineRTT := measureRTT(t, "r1", "r2", 10)
	t.Logf("Baseline RTT: mean=%.2fms, count=%d", calculateMean(baselineRTT), len(baselineRTT))

	// Phase 3: Execute delay event
	executeDelayScenario(t, "delay_50ms_scenario.json")

	// Phase 4: Wait for delay to take effect, then measure
	time.Sleep(2 * time.Second)
	delayedRTT := measureRTT(t, "r1", "r2", 10)
	t.Logf("Delayed RTT: mean=%.2fms, count=%d", calculateMean(delayedRTT), len(delayedRTT))

	// Phase 5: Verify delay was applied
	assertDelayApplied(t, baselineRTT, delayedRTT, 50.0)
}

// TestMultipleDelayScenarios tests various delay scenarios in parallel
func TestMultipleDelayScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	scenarios := []struct {
		name     string
		delay    int
		duration string
	}{
		{"delay_10ms", 10, "15s"},
		{"delay_50ms", 50, "15s"},
		{"delay_100ms", 100, "15s"},
	}

	for _, scenario := range scenarios {
		scenario := scenario // capture loop variable
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()
			testDelayScenario(t, scenario.delay, scenario.duration)
		})
	}
}

func testDelayScenario(t *testing.T, delayMs int, duration string) {
	setupTestEnvironment(t)
	defer cleanupTestEnvironment(t)

	waitForNetworkReady(t, []string{"r1", "r2"})

	// Measure baseline
	baselineRTT := measureRTT(t, "r1", "r2", 5)

	// Create and execute scenario
	scenarioFile := createDelayScenario(t, delayMs, duration)
	executeDelayScenario(t, scenarioFile)

	// Measure with delay applied
	time.Sleep(2 * time.Second)
	delayedRTT := measureRTT(t, "r1", "r2", 5)

	// Verify
	assertDelayApplied(t, baselineRTT, delayedRTT, float64(delayMs))
}

func setupTestEnvironment(t *testing.T) {
	t.Log("Setting up test environment...")
	
	// Deploy minimal topology using containerlab
	cmd := exec.Command("sudo", "containerlab", "deploy", 
		"--topo", "tests/topology/minimal_delay_test.yaml")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to deploy topology: %s", string(output))
	
	t.Log("Test environment setup completed")
}

func cleanupTestEnvironment(t *testing.T) {
	t.Log("Cleaning up test environment...")
	
	// Destroy topology
	cmd := exec.Command("sudo", "containerlab", "destroy", 
		"--topo", "tests/topology/minimal_delay_test.yaml", "--cleanup")
	if err := cmd.Run(); err != nil {
		t.Logf("Warning: Failed to cleanup topology: %v", err)
	}
	
	t.Log("Test environment cleanup completed")
}

func waitForNetworkReady(t *testing.T, hosts []string) {
	t.Log("Waiting for network to be ready...")
	
	timeout := 60 * time.Second
	interval := 2 * time.Second
	
	for _, host := range hosts {
		require.Eventually(t, func() bool {
			// Check if container is running and network is up
			cmd := exec.Command("docker", "exec", host, 
				"ping", "-c", "1", "-W", "1", "192.168.1.2")
			err := cmd.Run()
			return err == nil
		}, timeout, interval, "Host %s network not ready", host)
		
		t.Logf("Host %s is ready", host)
	}
	
	// Additional stabilization time
	time.Sleep(5 * time.Second)
	t.Log("Network is ready")
}

func measureRTT(t *testing.T, from, to string, count int) []float64 {
	targetIP := getTargetIP(to)
	
	cmd := fmt.Sprintf(`docker exec %s ping -c %d -i 0.2 %s | 
		grep 'time=' | 
		sed 's/.*time=\([0-9.]*\).*/\1/'`, from, count, targetIP)
	
	output, err := exec.Command("bash", "-c", cmd).Output()
	require.NoError(t, err, "Failed to measure RTT from %s to %s", from, to)
	
	return parseRTTValues(string(output))
}

func getTargetIP(host string) string {
	// For minimal test topology, use simple IP mapping
	switch host {
	case "r1":
		return "192.168.1.1"
	case "r2":
		return "192.168.1.2"
	default:
		return "192.168.1.1"
	}
}

func parseRTTValues(output string) []float64 {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	values := make([]float64, 0, len(lines))
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		value, err := strconv.ParseFloat(strings.TrimSpace(line), 64)
		if err == nil {
			values = append(values, value)
		}
	}
	
	return values
}

func executeDelayScenario(t *testing.T, scenarioFile string) {
	t.Logf("Executing delay scenario: %s", scenarioFile)
	
	cmd := exec.Command("sudo", "./netroub", fmt.Sprintf("tests/scenarios/%s", scenarioFile))
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to execute scenario: %s", string(output))
	
	t.Log("Delay scenario executed successfully")
}

func createDelayScenario(t *testing.T, delayMs int, duration string) string {
	scenarioFile := fmt.Sprintf("delay_%dms_test.json", delayMs)
	return scenarioFile
}

func assertDelayApplied(t *testing.T, baseline, delayed []float64, expectedDelay float64) {
	require.NotEmpty(t, baseline, "Baseline measurements should not be empty")
	require.NotEmpty(t, delayed, "Delayed measurements should not be empty")
	
	baselineMean := calculateMean(baseline)
	delayedMean := calculateMean(delayed)
	actualDelay := delayedMean - baselineMean
	
	t.Logf("RTT Analysis: baseline=%.2fms, delayed=%.2fms, actual_delay=%.2fms, expected=%.2fms",
		baselineMean, delayedMean, actualDelay, expectedDelay)
	
	// Tolerance: ±20% or ±10ms, whichever is larger
	tolerance := math.Max(expectedDelay*0.2, 10.0)
	
	assert.InDelta(t, expectedDelay, actualDelay, tolerance,
		"Expected delay %.1fms, got %.1fms (baseline: %.1fms, delayed: %.1fms)",
		expectedDelay, actualDelay, baselineMean, delayedMean)
	
	// Verify delay is statistically significant
	assert.Greater(t, actualDelay, 5.0, 
		"Delay should be at least 5ms to be measurable")
	
	// Check that variance didn't increase dramatically (jitter control)
	baselineStdDev := calculateStdDev(baseline)
	delayedStdDev := calculateStdDev(delayed)
	assert.LessOrEqual(t, delayedStdDev, baselineStdDev*3.0,
		"Delay introduced excessive jitter (baseline_stddev=%.2f, delayed_stddev=%.2f)",
		baselineStdDev, delayedStdDev)
}

func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateStdDev(values []float64) float64 {
	if len(values) <= 1 {
		return 0
	}
	
	mean := calculateMean(values)
	sumSquares := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	
	variance := sumSquares / float64(len(values)-1)
	return math.Sqrt(variance)
}