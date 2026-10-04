// ABOUTME: Immutable request-scoped database handles sharing the underlying connection pool.
// ABOUTME: Cancellation state belongs to a borrowed store copy, never the shared application store.
package storage

import "context"

func (s *SQLiteStore) WithContext(ctx context.Context) Store {
	copy := *s
	copy.requestCtx = ctx
	return &copy
}
func (s *PostgresStore) WithContext(ctx context.Context) Store {
	copy := *s
	copy.requestCtx = ctx
	return &copy
}
func (s *SQLiteStore) operationContext() context.Context {
	if s.requestCtx != nil {
		return s.requestCtx
	}
	return context.Background()
}
func (s *PostgresStore) operationContext() context.Context {
	if s.requestCtx != nil {
		return s.requestCtx
	}
	return context.Background()
}
