package main

import (
	"Tolstoy/agent"
	"net"
	"os/exec"
	"sync"
	"testing"
	"time"
)

func startTolstoy(t *testing.T, wg *sync.WaitGroup) *exec.Cmd {
	defer wg.Done()

	cmd := exec.Command("go", "run", "../main.go", "-config", "../broker/config.yaml")
	err := cmd.Start()
	if err != nil {
		t.Errorf("Failed to start the server during delivery test: %v", err)
		return nil
	}

	// Wait for the server to become available
	for i := 0; i < 10; i++ {
		conn, err := net.Dial("tcp", "localhost:8080")
		if err == nil {
			conn.Close()
			return cmd
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Errorf("Server never became ready")
	return nil
}

func TestFanoutDelivery(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	cmd := startTolstoy(t, &wg)
	if cmd == nil {
		t.Fatalf("Could not start server")
	}
	defer cmd.Process.Kill()

	wg.Wait()

	pub, err := agent.NewAgent("localhost:8080")
	if err != nil {
		t.Fatalf("Failed to start the publisher: %v", err)
	}

	sub, err := agent.NewAgent("localhost:8080")
	if err != nil {
		t.Fatalf("Failed to start the subscriber: %v", err)
	}

	var actual []string
	sub.Subscribe("mytopic", func(topic, message string) {
		actual = append(actual, message)
	})

	expected := []string{"first message", "second message", "third message"}
	for _, msg := range expected {
		pub.Publish("mytopic", msg)
		time.Sleep(100 * time.Millisecond) // Give some time for delivery
	}

	// Check results
	if len(actual) != len(expected) {
		t.Fatalf("Expected %d messages, got %d", len(expected), len(actual))
	}
	for i := range expected {
		if actual[i] != expected[i] {
			t.Errorf("Expected message %q, got %q", expected[i], actual[i])
		}
	}

	pub.Terminate()
	sub.Terminate()
}

