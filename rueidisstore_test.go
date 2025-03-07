package rueidisstore

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/redis/rueidis"
)

func TestFind(t *testing.T) {
	opt, err := rueidis.ParseURL(os.Getenv("SCS_REDIS_TEST_DSN"))
	if err != nil {
		t.Fatal(err)
	}
	client, err := rueidis.NewClient(opt)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()
	r := New(client)

	err = client.Do(ctx, client.B().Flushdb().Build()).Error()
	if err != nil {
		t.Fatal(err)
	}

	err = client.Do(ctx, client.B().Set().Key(r.prefix+"session_token").Value("encoded_data").ExSeconds(60).Build()).Error()
	if err != nil {
		t.Fatal(err)
	}

	b, found, err := r.FindCtx(ctx, "session_token")
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if bytes.Equal(b, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", b, []byte("encoded_data"))
	}
}

func TestSaveNew(t *testing.T) {
	opt, err := rueidis.ParseURL(os.Getenv("SCS_REDIS_TEST_DSN"))
	if err != nil {
		t.Fatal(err)
	}
	client, err := rueidis.NewClient(opt)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()
	r := New(client)

	err = client.Do(ctx, client.B().Flushdb().Build()).Error()
	if err != nil {
		t.Fatal(err)
	}

	err = r.CommitCtx(ctx, "session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	data, err := client.Do(ctx, client.B().Get().Key(r.prefix+"session_token").Build()).AsBytes()
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(data, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", data, []byte("encoded_data"))
	}
}

func TestFindMissing(t *testing.T) {
	opt, err := rueidis.ParseURL(os.Getenv("SCS_REDIS_TEST_DSN"))
	if err != nil {
		t.Fatal(err)
	}
	client, err := rueidis.NewClient(opt)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()
	r := New(client)

	err = client.Do(ctx, client.B().Flushdb().Build()).Error()
	if err != nil {
		t.Fatal(err)
	}

	_, found, err := r.FindCtx(ctx, "missing_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestSaveUpdated(t *testing.T) {
	opt, err := rueidis.ParseURL(os.Getenv("SCS_REDIS_TEST_DSN"))
	if err != nil {
		t.Fatal(err)
	}
	client, err := rueidis.NewClient(opt)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()
	r := New(client)

	err = client.Do(ctx, client.B().Flushdb().Build()).Error()
	if err != nil {
		t.Fatal(err)
	}

	err = client.Do(ctx, client.B().Set().Key(r.prefix+"session_token").Value("encoded_data").ExSeconds(100).Build()).Error()
	if err != nil {
		t.Fatal(err)
	}

	err = r.CommitCtx(ctx, "session_token", []byte("new_encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	data, err := client.Do(ctx, client.B().Get().Key(r.prefix+"session_token").Build()).AsBytes()
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(data, []byte("new_encoded_data")) == false {
		t.Fatalf("got %v: expected %v", data, []byte("new_encoded_data"))
	}
}

func TestExpiry(t *testing.T) {
	opt, err := rueidis.ParseURL(os.Getenv("SCS_REDIS_TEST_DSN"))
	if err != nil {
		t.Fatal(err)
	}
	client, err := rueidis.NewClient(opt)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()
	r := New(client)

	err = client.Do(ctx, client.B().Flushdb().Build()).Error()
	if err != nil {
		t.Fatal(err)
	}

	err = r.CommitCtx(ctx, "session_token", []byte("encoded_data"), time.Now().Add(5*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	_, found, _ := r.FindCtx(ctx, "session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(5 * time.Second)
	_, found, _ = r.FindCtx(ctx, "session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestDelete(t *testing.T) {
	opt, err := rueidis.ParseURL(os.Getenv("SCS_REDIS_TEST_DSN"))
	if err != nil {
		t.Fatal(err)
	}
	client, err := rueidis.NewClient(opt)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()
	r := New(client)

	err = client.Do(ctx, client.B().Flushdb().Build()).Error()
	if err != nil {
		t.Fatal(err)
	}

	err = client.Do(ctx, client.B().Set().Key(r.prefix+"session_token").Value("encoded_data").ExSeconds(100).Build()).Error()
	if err != nil {
		t.Fatal(err)
	}

	err = r.DeleteCtx(ctx, "session_token")
	if err != nil {
		t.Fatal(err)
	}

	data, err := client.Do(ctx, client.B().Get().Key(r.prefix+"session_token").Build()).AsBytes()
	if !rueidis.IsRedisNil(err) {
		t.Fatal(err)
	}
	if data != nil {
		t.Fatalf("got %v: expected %v", data, nil)
	}
}

func TestAll(t *testing.T) {
	opt, err := rueidis.ParseURL(os.Getenv("SCS_REDIS_TEST_DSN"))
	if err != nil {
		t.Fatal(err)
	}
	client, err := rueidis.NewClient(opt)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()
	r := New(client)

	err = client.Do(ctx, client.B().Flushdb().Build()).Error()
	if err != nil {
		t.Fatal(err)
	}

	sessions := make(map[string][]byte)
	for i := 0; i < 4; i++ {
		key := fmt.Sprintf("token_%v", i)
		val := []byte(key)
		err = client.Do(ctx, client.B().Set().Key(r.prefix+key).Value(key).ExSeconds(100).Build()).Error()
		if err != nil {
			t.Fatal(err)
		}
		sessions[key] = val
	}

	gotSessions, err := r.AllCtx(ctx)
	if err != nil {
		t.Fatal(err)
	}

	for k := range sessions {
		err = r.DeleteCtx(ctx, k)
		if err != nil {
			t.Fatal(err)
		}
	}
	if reflect.DeepEqual(sessions, gotSessions) == false {
		t.Fatalf("got %v: expected %v", gotSessions, sessions)
	}
}
