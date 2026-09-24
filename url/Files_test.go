package url

import (
	"testing"
	"time"
)

func TestFilesAddFieldDoesNotDeadlock(t *testing.T) {
	files := NewFiles()
	done := make(chan struct{})
	go func() {
		files.AddField("field", "value")
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("AddField did not return")
	}

	if got := files.Get("field"); got == nil || got["value"] != "value" {
		t.Fatalf("stored field = %#v, want value", got)
	}
}
