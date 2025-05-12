package main

import (
	"Tolstoy/agent"
	"net"
	"os/exec"
	"sync"
	"testing"
	"time"
	"strings"
)
func startTolstoy(t *testing.T) *exec.Cmd {
	cmd := exec.Command("go", "run", "../main.go", "-config", "../broker/config.yaml")

	err := cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start the server: %v", err)
	}

	// Wait for the server to be available
	for i := 0; i < 20; i++ {
		conn, err := net.Dial("tcp", "localhost:8080")
		if err == nil {
			conn.Close()
			return cmd
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatal("Server never became ready")
	return nil
}

func TestFanoutDelivery(t *testing.T) {
	cmd := startTolstoy(t)
	defer cmd.Process.Kill()

	pub, err := agent.NewAgent("localhost:8080")
	if err != nil {
		t.Fatalf("Failed to start publisher: %v", err)
	}
	defer pub.Terminate()

	sub, err := agent.NewAgent("localhost:8080")
	if err != nil {
		t.Fatalf("Failed to start subscriber: %v", err)
	}
	defer sub.Terminate()

	var mu sync.Mutex
	actual := []string{}
	done := make(chan struct{})

	expected := []string{"first message", "second message", "third message"}

	sub.Subscribe("mytopic", func(topic, message string) {
		mu.Lock()
		actual = append(actual, message)
		if len(actual) == len(expected) {
			close(done)
		}
		mu.Unlock()
	})

	// Give some time for subscription to register
	time.Sleep(200 * time.Millisecond)

	for _, msg := range expected {
		pub.Publish("mytopic", msg)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for all messages")
	}

	mu.Lock()
	defer mu.Unlock()
	if len(actual) != len(expected) {
		t.Fatalf("Expected %d messages, got %d", len(expected), len(actual))
	}
	for i := range expected {
		if actual[i] != expected[i] {
			t.Errorf("Expected %q, got %q", expected[i], actual[i])
		}
	}
	killPort(t,":8080")
}
func killPort(t *testing.T, port string) {
	cmd := exec.Command("lsof", "-t", "-i", ":"+port)
	output, err := cmd.Output()
	if err != nil {
		t.Logf("Could not find process on port %s: %v", port, err)
		return
	}

	pid := strings.TrimSpace(string(output))
	t.Logf("Killing process on port %s with PID %s", port, pid)

	killCmd := exec.Command("kill", "-9", pid)
	if err := killCmd.Run(); err != nil {
		t.Logf("Failed to kill PID %s: %v", pid, err)
	}
}
