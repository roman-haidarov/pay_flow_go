package cache

import (
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
)

func TestNew_WithURL_DSN(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rc, err := New("redis://" + mr.Addr())
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer rc.Close()
}

func TestNew_WithHostPort_DSN(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rc, err := New(mr.Addr()) // host:port fall-back path
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer rc.Close()
}

func TestNew_EmptyDSN_Error(t *testing.T) {
	if _, err := New(""); err == nil {
		t.Fatal("expected error on empty DSN")
	}
}

func TestClose_Idempotent(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rc, err := New("redis://" + mr.Addr())
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	if err := rc.Close(); err != nil {
		t.Fatalf("close error: %v", err)
	}

	_ = rc.Close()
}
