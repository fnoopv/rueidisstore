package rueidisstore

import (
	"context"
	"errors"
	"time"

	"github.com/redis/rueidis"
)

// RueidisStore the session store
type RueidisStore struct {
	client rueidis.Client
	prefix string
}

// New returns a new RueidisStore instance
func New(client rueidis.Client) *RueidisStore {
	return NewWithPrefix(client, "scs:session:")
}

// NewWithPrefix returns a new RueidisStore instance
func NewWithPrefix(client rueidis.Client, prefix string) *RueidisStore {
	return &RueidisStore{
		client: client,
		prefix: prefix,
	}
}

// FindCtx returns the data for a given session token from the RueidisStore instance.
// If the session token is not found or is expired, the returned exists flag
// will be set to false.
func (ru *RueidisStore) FindCtx(ctx context.Context, token string) ([]byte, bool, error) {
	b, err := ru.client.Do(ctx, ru.client.B().Get().Key(ru.prefix+token).Build()).AsBytes()
	if rueidis.IsRedisNil(err) {
		return nil, false, nil
	}
	if rueidis.IsParseErr(err) {
		return nil, false, errors.New("parse error")
	}

	return b, true, nil
}

// CommitCtx adds a session token and data to the RueidisStore instance with the
// given expiry time. If the session token already exists then the data and
// expiry time are updated.
func (ru *RueidisStore) CommitCtx(ctx context.Context, token string, b []byte, expiry time.Time) error {
	err := ru.client.Do(
		ctx,
		ru.client.B().Set().Key(ru.prefix+token).Value(string(b)).Ex(expiry.Sub(time.Now())).Build(),
	).Error()

	return err
}

// DeleteCtx removes a session token and corresponding data from the RueidisStore
// instance.
func (ru *RueidisStore) DeleteCtx(ctx context.Context, token string) error {
	return ru.client.Do(ctx, ru.client.B().Del().Key(ru.prefix+token).Build()).Error()
}

// AllCtx returns a map containing the token and data for all active (i.e.
// not expired) sessions in the RueidisStore instance.
func (ru *RueidisStore) AllCtx(ctx context.Context) (map[string][]byte, error) {
	sessions := make(map[string][]byte)

	for _, node := range ru.client.Nodes() {
		var scan rueidis.ScanEntry

		for more := true; more; more = scan.Cursor != 0 {
			entries, err := node.Do(ctx, node.B().Scan().Cursor(scan.Cursor).Match(ru.prefix+"*").Build()).AsScanEntry()

			if rueidis.IsRedisNil(err) {
				return nil, nil
			}

			if err != nil {
				return nil, err
			}

			scan = entries

			for _, entry := range entries.Elements {
				token := entry[len(ru.prefix):]
				data, exists, err := ru.FindCtx(ctx, token)
				if err != nil {
					return nil, err
				}
				if exists {
					sessions[token] = data
				}
			}
		}
	}

	return sessions, nil
}

// Find don't use
func (ru *RueidisStore) Find(token string) ([]byte, bool, error) {
	panic("missing context arg")
}

// Commit don't use
func (ru *RueidisStore) Commit(token string, b []byte, expiry time.Time) error {
	panic("missing context arg")
}

// Delete don't use
func (ru *RueidisStore) Delete(token string) error {
	panic("missing context arg")
}
