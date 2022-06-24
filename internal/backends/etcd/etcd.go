package etcd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/api7/etcd-adapter/internal/extend"
	"github.com/api7/etcd-adapter/kine/server"
	"github.com/api7/gopkg/pkg/log"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
	"strconv"
	"time"
)

var (
	DirPlaceholder = []byte("init_dir")
)

type Options struct {
	Host     []string
	Prefix   string
	Timeout  int
	User     string
	Password string
	Tls      TlsConfig
}

type TlsConfig struct {
	CertFile string
	KeyFile  string
	CaFile   string
	Verify   bool
}

type EtcdV3 struct {
	client          *clientv3.Client
	conf            clientv3.Config
	timeout         time.Duration
	currentRevision int64
}

func (s *EtcdV3) LeaseGrant(ctx context.Context, id string, ttl int64) (*extend.GrantRes, error) {
	resp, err := s.client.Lease.Grant(ctx, ttl)
	if err != nil {
		return nil, err
	}
	res := &extend.GrantRes{
		ID:    strconv.FormatInt(int64(resp.ID), 10),
		TTL:   resp.TTL,
		Error: resp.Error,
		GrantHeaderRes: extend.GrantHeaderRes{
			ClusterId: resp.ResponseHeader.ClusterId,
			MemberId:  resp.ResponseHeader.MemberId,
			Revision:  resp.ResponseHeader.Revision,
			RaftTerm:  resp.ResponseHeader.RaftTerm,
		},
	}
	return res, nil
}

func (s *EtcdV3) Count(ctx context.Context, prefix string) (int64, int64, error) {
	resp, err := s.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return s.currentRevision, 0, err
	}
	return s.currentRevision, resp.Count, nil
}

func (s *EtcdV3) Watch(ctx context.Context, key string, revision int64) <-chan []*server.Event {
	eventChan := s.client.Watch(ctx, key, clientv3.WithPrefix())
	ch := make(chan []*server.Event, 1)
	go func() {
		defer close(ch)
		for event := range eventChan {
			var events []*server.Event
			for _, ev := range event.Events {
				// We use a placeholder to mark a key to be a directory. So we need to skip the hack here.
				if bytes.Equal(ev.Kv.Value, DirPlaceholder) {
					continue
				}

				key := string(ev.Kv.Key)
				kv := &server.KeyValue{
					Key:            key,
					CreateRevision: ev.Kv.CreateRevision,
					ModRevision:    ev.Kv.ModRevision,
					Value:          ev.Kv.Value,
					Lease:          ev.Kv.Lease,
				}
				typ := &server.Event{
					Delete: false,
					Create: false,
					KV:     kv,
				}
				switch ev.Type {
				case clientv3.EventTypePut:
					typ.Create = true
				case clientv3.EventTypeDelete:
					typ.Delete = true
				}
				events = append(events, typ)
			}
			if len(events) > 0 {
				ch <- events
			}
		}
	}()

	return ch
}

func (s *EtcdV3) DbSize(ctx context.Context) (int64, error) {
	return 0, nil
}

func NewEtcdCache(ctx context.Context, options *Options) (server.Backend, extend.Extend, error) {
	timeout := time.Duration(options.Timeout)
	s := &EtcdV3{timeout: timeout, currentRevision: 0}

	if s.timeout == 0 {
		s.timeout = 10 * time.Second
	}
	config := clientv3.Config{
		Endpoints:            options.Host,
		DialTimeout:          timeout,
		DialKeepAliveTimeout: timeout,
		Username:             options.User,
		Password:             options.Password,
	}

	if options.Tls.Verify {
		tlsInfo := transport.TLSInfo{
			CertFile:      options.Tls.CertFile,
			KeyFile:       options.Tls.KeyFile,
			TrustedCAFile: options.Tls.CaFile,
		}
		tlsConf, err := tlsInfo.ClientConfig()
		if err != nil {
			return nil, nil, err
		}
		config.TLS = tlsConf
	}

	s.conf = config
	cli, err := clientv3.New(s.conf)
	if err != nil {
		log.Errorf("etcd init failed: %s", err)
		return nil, nil, err
	}

	s.client = cli
	return s, s, nil
}

func (s *EtcdV3) Start(ctx context.Context) error {
	return nil
}

// Get a value given its key
func (s *EtcdV3) Get(ctx context.Context, key string, revision int64) (int64, *server.KeyValue, error) {
	if revision < 0 {
		revision = s.currentRevision
	}
	resp, err := s.client.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithRev(revision))
	if err != nil {
		log.Errorf("etcd get key[%s] failed: %s", key, err)
		return revision, nil, fmt.Errorf("etcd get key[%s] failed: %s", key, err)
	}
	if resp.Count == 0 {
		log.Warnf("etcd get key[%s] is not found", key)
		return revision, nil, nil
	}
	kv := &server.KeyValue{
		Key:            key,
		CreateRevision: resp.Kvs[0].CreateRevision,
		ModRevision:    resp.Kvs[0].ModRevision,
		Value:          resp.Kvs[0].Value,
		Lease:          resp.Kvs[0].Lease,
	}
	return revision, kv, nil
}
func (s *EtcdV3) List(ctx context.Context, prefix, startKey string, limit, revision int64) (revRet int64, kvRet []*server.KeyValue, errRet error) {
	resp, err := s.client.Get(ctx, prefix, clientv3.WithPrefix(), clientv3.WithRev(revision), clientv3.WithLimit(limit))
	if err != nil {
		return revision, nil, err
	}

	var vals []*server.KeyValue
	for _, kv := range resp.Kvs {
		vals = append(vals, &server.KeyValue{
			Key:            string(kv.Key),
			CreateRevision: kv.CreateRevision,
			ModRevision:    kv.ModRevision,
			Value:          kv.Value,
			Lease:          kv.Lease,
		})
	}

	return revision, vals, nil
}

func (s *EtcdV3) Create(ctx context.Context, key string, value []byte, lease int64) (int64, error) {
	resp, err := s.client.Txn(ctx).
		Then(clientv3.OpPut(key, string(value))).
		Commit()
	if err != nil {
		return s.currentRevision, err
	}
	if !resp.Succeeded {
		return s.currentRevision, fmt.Errorf("key exists")
	}
	return resp.Header.Revision, nil
}

func (s *EtcdV3) Update(ctx context.Context, key string, value []byte, revision, lease int64) (int64, *server.KeyValue, bool, error) {
	resp, err := s.client.Txn(ctx).
		Then(clientv3.OpPut(key, string(value))).
		Commit()
	if err != nil {
		return revision, nil, false, err
	}
	if !resp.Succeeded {
		return revision, nil, false, fmt.Errorf("revision %d doesnt match", revision)
	}
	kv := &server.KeyValue{
		Key:            key,
		CreateRevision: revision,
		ModRevision:    revision,
	}
	return resp.Header.Revision, kv, true, nil
}

func (s *EtcdV3) Delete(ctx context.Context, key string, revision int64) (int64, *server.KeyValue, bool, error) {
	resp, err := s.client.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(key), "=", revision)).
		Then(clientv3.OpDelete(key)).
		Else(clientv3.OpGet(key)).
		Commit()
	if err != nil {
		return revision, nil, false, err
	}
	if !resp.Succeeded {
		return revision, nil, false, fmt.Errorf("revision %d doesnt match", revision)
	}
	kv := &server.KeyValue{
		Key:            key,
		CreateRevision: revision,
		ModRevision:    revision,
	}
	return resp.Header.Revision, kv, true, nil
}

func (s *EtcdV3) Close() error {
	return s.client.Close()
}
