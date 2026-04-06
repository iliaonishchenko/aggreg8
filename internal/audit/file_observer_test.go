package audit

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileObserver_Notify(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "audit_test_*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	observer, err := NewFileObserver(tmpFile.Name())
	require.NoError(t, err)
	defer observer.Close()

	event1 := AuditEvent{
		TS:        1000,
		Metrics:   []string{"Alloc", "Frees"},
		IPAddress: "192.168.0.1",
	}
	event2 := AuditEvent{
		TS:        2000,
		Metrics:   []string{"HeapSys"},
		IPAddress: "10.0.0.1",
	}

	require.NoError(t, observer.Notify(event1))
	require.NoError(t, observer.Notify(event2))

	data, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	assert.Len(t, lines, 2)

	var got1 AuditEvent
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &got1))
	assert.Equal(t, event1, got1)

	var got2 AuditEvent
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &got2))
	assert.Equal(t, event2, got2)
}

func TestFileObserver_Notify_InvalidPath(t *testing.T) {
	_, err := NewFileObserver("/nonexistent/dir/audit.log")
	assert.Error(t, err)
}
