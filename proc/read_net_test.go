package proc

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewNetIPSocket(t *testing.T) {
	expected := NetIPSocket{
		{Sl: 0, St: 10, TxQueue: 0, RxQueue: 0, UID: 101, Inode: 24695},
		{Sl: 1, St: 10, TxQueue: 0, RxQueue: 0, UID: 1000, Inode: 32001},
		{Sl: 2, St: 1, TxQueue: 0, RxQueue: 0, UID: 1000, Inode: 188938},
		{Sl: 3, St: 1, TxQueue: 0, RxQueue: 0, UID: 1000, Inode: 188936},
		{Sl: 4, St: 1, TxQueue: 0, RxQueue: 0, UID: 1000, Inode: 166671},
		{Sl: 5, St: 1, TxQueue: 0, RxQueue: 0, UID: 1000, Inode: 194784},
	}

	actual, err := NewNetIPSocket("../fixtures/14804/net/tcp")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Errorf("NetIPSocket differs: (-got +want)\n%s", diff)
	}
}

func TestNewNetIPSocketTCP_NoFile(t *testing.T) {
	_, err := NewNetIPSocket("../fixtures/14804/net/tcp_not_found")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestNewNetIPSocketTCP_InvalidFile(t *testing.T) {
	_, err := NewNetIPSocket("read_net_test.go")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
