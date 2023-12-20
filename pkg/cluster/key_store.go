// SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
// SPDX-FileCopyrightText: 2023 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"fmt"
	"os"
	"path/filepath"
)

// KeyStore is a store for public keys.
type KeyStore struct {
	basePath string
}

// NewKeyStore creates a new KeyStore
func NewKeyStore(basePath string) *KeyStore {
	return &KeyStore{
		basePath: basePath,
	}
}

// Init initializes the key store, creating the store directory if needed.
func (s *KeyStore) Init() error {
	return os.MkdirAll(s.basePath, 0760)
}

func fileExists(path string) bool {
	// XXX: There's a subtle bug: if stat fails for another reason that the file
	// not existing, we return the file exists.
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (s *KeyStore) keyPath(name string) string {
	return filepath.Join(s.basePath, name)
}

func (s *KeyStore) keyExists(name string) bool {
	return fileExists(s.keyPath(name))
}

// Store adds the key to the store.
func (s *KeyStore) Store(name, key string) error {
	if s.keyExists(name) {
		return fmt.Errorf("key store: store: key '%s' already exists", name)
	}

	if err := os.WriteFile(s.keyPath(name), []byte(key), 0644); err != nil {
		return fmt.Errorf("key store: write: %w", err)
	}

	return nil
}

// Get retrieves a key from the store.
func (s *KeyStore) Get(name string) ([]byte, error) {
	if !s.keyExists(name) {
		return nil, fmt.Errorf("key store: get: unknown key '%s'", name)
	}
	return os.ReadFile(s.keyPath(name))
}

// Remove removes a key from the store.
func (s *KeyStore) Remove(name string) error {
	if !s.keyExists(name) {
		return fmt.Errorf("key store: remove: unknown key '%s'", name)
	}
	if err := os.Remove(s.keyPath(name)); err != nil {
		return fmt.Errorf("key store: remove: %w", err)
	}
	return nil
}
