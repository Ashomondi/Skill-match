package clients

import (
	"context"
	"testing"
	"time"
)

func TestPoolOptionsWithDefaults(t *testing.T) {
	opts := (PoolOptions{}).withDefaults()
	if opts.MaxConns != defaultMaxConns {
		t.Fatalf("expected default MaxConns %d, got %d", defaultMaxConns, opts.MaxConns)
	}
	if opts.MinConns != defaultMinConns {
		t.Fatalf("expected default MinConns %d, got %d", defaultMinConns, opts.MinConns)
	}
	if opts.MaxConnLifetime != defaultMaxConnLifetime {
		t.Fatalf("expected default MaxConnLifetime, got %v", opts.MaxConnLifetime)
	}
	if opts.ConnectTimeout != defaultConnectTimeout {
		t.Fatalf("expected default ConnectTimeout, got %v", opts.ConnectTimeout)
	}
}

func TestPoolOptionsKeepsOverrides(t *testing.T) {
	opts := (PoolOptions{MaxConns: 5, MinConns: 1, ConnectTimeout: 2 * time.Second}).withDefaults()
	if opts.MaxConns != 5 {
		t.Fatalf("expected MaxConns 5, got %d", opts.MaxConns)
	}
	if opts.MinConns != 1 {
		t.Fatalf("expected MinConns 1, got %d", opts.MinConns)
	}
	if opts.ConnectTimeout != 2*time.Second {
		t.Fatalf("expected ConnectTimeout 2s, got %v", opts.ConnectTimeout)
	}
}

func TestNewPoolRejectsEmptyDSN(t *testing.T) {
	if _, err := NewPool(context.Background(), "", PoolOptions{}); err != ErrEmptyDSN {
		t.Fatalf("expected ErrEmptyDSN, got %v", err)
	}
}
