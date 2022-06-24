// Copyright api7.ai
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package mysql

import (
	"context"
	"github.com/api7/etcd-adapter/internal/extend"
	"github.com/api7/etcd-adapter/kine/drivers/generic"
	mysqldriver "github.com/api7/etcd-adapter/kine/drivers/mysql"
	"github.com/api7/etcd-adapter/kine/server"
	"github.com/api7/etcd-adapter/kine/tls"
)

// Options contains settings for controlling the connection to MySQL.
type Options struct {
	DSN      string
	ConnPool generic.ConnectionPoolConfig
}

type mysqlCache struct {
	server.Backend
}

// NewMySQLCache returns a server.Backend interface which was implemented with
// the MySQL backend. The first argument `ctx` is used to control the lifecycle of
// mysql connection pool.
func NewMySQLCache(ctx context.Context, options *Options) (server.Backend, extend.Extend, error) {
	dsn := options.DSN
	backend, err := mysqldriver.New(ctx, dsn, tls.Config{}, options.ConnPool, nil)
	if err != nil {
		return nil, nil, err
	}
	mc := &mysqlCache{
		Backend: backend,
	}
	return mc, nil, nil
}

func (m *mysqlCache) Start(ctx context.Context) error {
	return m.Backend.Start(ctx)
}
