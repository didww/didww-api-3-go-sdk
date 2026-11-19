package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	didww "github.com/didww/didww-api-3-go-sdk/v4"
)

// redisTokenStore is a didww.TokenStore backed by Redis, as the README
// suggests for sharing tokens between processes. It speaks the Redis protocol
// (RESP) over one connection with the standard library only, so the examples
// need no Redis client package; use a Redis client of your choice in your code.
type redisTokenStore struct {
	mu     sync.Mutex
	conn   net.Conn
	rd     *bufio.Reader
	prefix string

	hits, sets int
	keys       map[string]bool
}

var _ didww.TokenStore = (*redisTokenStore)(nil)

// dialRedis connects to a redis://[user:password@]host[:port][/db] URL.
func dialRedis(ctx context.Context, rawURL, prefix string) (*redisTokenStore, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "redis" {
		return nil, fmt.Errorf("unsupported Redis URL scheme %q", u.Scheme)
	}
	addr := u.Host
	if u.Port() == "" {
		addr = net.JoinHostPort(u.Hostname(), "6379")
	}
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	store := &redisTokenStore{conn: conn, rd: bufio.NewReader(conn), prefix: prefix, keys: map[string]bool{}}
	if password, ok := u.User.Password(); ok {
		auth := []string{"AUTH", u.User.Username(), password}
		if u.User.Username() == "" {
			auth = []string{"AUTH", password}
		}
		if _, err = store.do(ctx, auth...); err != nil {
			_ = conn.Close()
			return nil, err
		}
	}
	if db := strings.TrimPrefix(u.Path, "/"); db != "" && db != "0" {
		if _, err = store.do(ctx, "SELECT", db); err != nil {
			_ = conn.Close()
			return nil, err
		}
	}
	return store, nil
}

func (r *redisTokenStore) Get(ctx context.Context, key string) (t didww.Token, ok bool, err error) {
	reply, err := r.do(ctx, "GET", r.prefix+key)
	if err != nil || reply == nil {
		return didww.Token{}, false, err
	}
	data, isString := reply.(string)
	if !isString {
		return didww.Token{}, false, fmt.Errorf("unexpected Redis reply %v", reply)
	}
	err = json.Unmarshal([]byte(data), &t)
	if err != nil {
		return didww.Token{}, false, err
	}
	r.mu.Lock()
	r.hits++
	r.mu.Unlock()
	return t, true, nil
}

func (r *redisTokenStore) Set(ctx context.Context, key string, t didww.Token, ttl time.Duration) error { //nolint:gocritic // signature of didww.TokenStore
	data, err := json.Marshal(t) //nolint:gosec // a token store keeps the access token, protect it like the API key
	if err != nil {
		return err
	}
	args := []string{"SET", r.prefix + key, string(data)}
	if ttl > 0 {
		args = append(args, "PX", strconv.FormatInt(ttl.Milliseconds(), 10))
	}
	if _, err = r.do(ctx, args...); err != nil {
		return err
	}
	r.mu.Lock()
	r.sets++
	r.keys[key] = true
	r.mu.Unlock()
	return nil
}

func (r *redisTokenStore) Delete(ctx context.Context, key string) error {
	_, err := r.do(ctx, "DEL", r.prefix+key)
	return err
}

// counts returns how many Gets found a token and how many tokens were stored.
func (r *redisTokenStore) counts() (hits, sets int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.hits, r.sets
}

// close deletes the keys this store wrote and closes the connection.
func (r *redisTokenStore) close(ctx context.Context) error {
	r.mu.Lock()
	keys := make([]string, 0, len(r.keys))
	for key := range r.keys {
		keys = append(keys, key)
	}
	r.mu.Unlock()
	var errs []error
	for _, key := range keys {
		errs = append(errs, r.Delete(ctx, key))
	}
	errs = append(errs, r.conn.Close())
	return errors.Join(errs...)
}

// do sends one command and reads its reply: a string, an int64, nil for a
// missing value, or an error for a Redis error reply.
func (r *redisTokenStore) do(ctx context.Context, args ...string) (any, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(10 * time.Second)
	}
	if err := r.conn.SetDeadline(deadline); err != nil {
		return nil, err
	}

	var cmd strings.Builder
	fmt.Fprintf(&cmd, "*%d\r\n", len(args))
	for _, arg := range args {
		fmt.Fprintf(&cmd, "$%d\r\n%s\r\n", len(arg), arg)
	}
	if _, err := io.WriteString(r.conn, cmd.String()); err != nil {
		return nil, err
	}
	return r.readReply()
}

func (r *redisTokenStore) readReply() (any, error) {
	line, err := r.rd.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSuffix(line, "\r\n")
	if line == "" {
		return nil, errors.New("empty Redis reply")
	}
	switch line[0] {
	case '+':
		return line[1:], nil
	case '-':
		return nil, fmt.Errorf("redis: %s", line[1:])
	case ':':
		return strconv.ParseInt(line[1:], 10, 64)
	case '$':
		n, err := strconv.Atoi(line[1:])
		if err != nil || n < 0 {
			return nil, err // -1 is a missing value
		}
		buf := make([]byte, n+2)
		if _, err = io.ReadFull(r.rd, buf); err != nil {
			return nil, err
		}
		return string(buf[:n]), nil
	}
	return nil, fmt.Errorf("unsupported Redis reply %q", line)
}
