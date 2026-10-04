// ABOUTME: Simulates a successful commit followed by a lost acknowledgement.
// ABOUTME: HTTP restore tests must not promise unchanged data on this failure.
package main

import (
	"context"
	"fmt"
	"omnicollect/storage"
)

type lostCommitAckStore struct{ storage.Store }

func (s *lostCommitAckStore) WithContext(ctx context.Context) storage.Store {
	return &lostCommitAckStore{s.Store.WithContext(ctx)}
}
func (s *lostCommitAckStore) Restore(ctx context.Context, snapshot storage.Snapshot, mode string) (storage.RestoreResult, error) {
	result, err := s.Store.Restore(ctx, snapshot, mode)
	if err != nil {
		return result, err
	}
	return result, fmt.Errorf("commit acknowledgement lost")
}
