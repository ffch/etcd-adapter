ENV_INSTALL            ?= install
ENV_INST_PREFIX        ?= /usr/local

test:
	@go test ./...

bench:
	@go test -bench '^Benchmark' ./...

gofmt:
	@find . -name "*.go" | xargs gofmt -w

.PHONY: build
build:
	go build

### install:		Install apisix-seed
.PHONY: install
install:
	$(ENV_INSTALL) -d $(ENV_INST_PREFIX)/apisix-etcd
	$(ENV_INSTALL) -d $(ENV_INST_PREFIX)/apisix-etcd/log
	$(ENV_INSTALL) -d $(ENV_INST_PREFIX)/apisix-etcd/config
	$(ENV_INSTALL) etcd-adapter $(ENV_INST_PREFIX)/apisix-etcd/
	$(ENV_INSTALL) config/config.yaml $(ENV_INST_PREFIX)/apisix-etcd/config/
	$(ENV_INSTALL) config/miot-etcd_client.pem $(ENV_INST_PREFIX)/apisix-etcd/config/
	$(ENV_INSTALL) config/miot-etcd_client_key.pem $(ENV_INST_PREFIX)/apisix-etcd/config/
	$(ENV_INSTALL) config/miot-etcd_ca.pem $(ENV_INST_PREFIX)/apisix-etcd/config/

lint:
	@golangci-lint run
