package store

import (
	"context"
	"fmt"
	"strings"
)

type BackendType uint8

const (
	EigenDABackendType BackendType = iota
	MemoryBackendType
	S3BackendType
	RedisBackendType

	Unknown
)

var (
	ErrProxyOversizedBlob   = fmt.Errorf("encoded blob is larger than max blob size")
	ErrEigenDAOversizedBlob = fmt.Errorf("blob size cannot exceed")
)

func (b BackendType) String() string {
	switch b {
	case EigenDABackendType:
		return "EigenDA"
	case MemoryBackendType:
		return "Memory"
	case S3BackendType:
		return "S3"
	case RedisBackendType:
		return "Redis"
	case Unknown:
		fallthrough
	default:
		return "Unknown"
	}
}

func StringToBackendType(s string) BackendType {
	lower := strings.ToLower(s)

	switch lower {
	case "eigenda":
		return EigenDABackendType
	case "memory":
		return MemoryBackendType
	case "s3":
		return S3BackendType
	case "redis":
		return RedisBackendType
	case "unknown":
		fallthrough
	default:
		return Unknown
	}
}

// Used for E2E tests
type Stats struct {
	Entries int
	Reads   int
}

type Store interface {
	// Backend returns the backend type provider of the store.
	BackendType() BackendType
	// Verify verifies the given key-value pair.
	Verify(ctx context.Context, key []byte, value []byte) error
}

type GeneratedKeyStore interface {
	Store
	// Get retrieves the given key if it's present in the key-value data store.
	Get(ctx context.Context, key []byte) ([]byte, error)
	// Put inserts the given value into the key-value data store.
	Put(ctx context.Context, value []byte) (key []byte, err error)
}

type PrecomputedKeyStore interface {
	Store
	// Get retrieves the given key if it's present in the key-value data store.
	Get(ctx context.Context, key []byte) ([]byte, error)
	// Put inserts the given value into the key-value data store.
	Put(ctx context.Context, key []byte, value []byte) error
}
