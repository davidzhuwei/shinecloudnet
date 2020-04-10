#!/usr/bin/make -f

PACKAGES_NOSIMULATION=$(shell go list ./... | grep -v '/simulation')
PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
VERSION := $(shell echo $(shell git describe --tags --always) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
LEDGER_ENABLED ?= true
SDK_PACK := $(shell go list -m github.com/shinecloudfoundation/shinecloudnet | sed  's/ /\@/g')
BINDIR ?= $(GOPATH)/bin
SIMAPP = ./simapp
MOCKS_DIR = $(CURDIR)/tests/mocks

export GO111MODULE = on

all: tools install lint check

update-swagger-docs: statik
	$(BINDIR)/statik -src=client/lcd/swagger-ui -dest=client/lcd -f -m
	@if [ -n "$(git status --porcelain)" ]; then \
        echo "\033[91mSwagger docs are out of sync!!!\033[0m";\
        exit 1;\
    else \
    	echo "\033[92mSwagger docs are in sync\033[0m";\
    fi
.PHONY: update-swagger-docs

# process build tags

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support)
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq ($(WITH_CLEVELDB),yes)
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags

ldflags = -X github.com/shinecloudfoundation/shinecloudnet/version.Name=shine \
		  -X github.com/shinecloudfoundation/shinecloudnet/version.ServerName=scloud \
		  -X github.com/shinecloudfoundation/shinecloudnet/version.ClientName=scloudcli \
		  -X github.com/shinecloudfoundation/shinecloudnet/version.Version=$(VERSION) \
		  -X github.com/shinecloudfoundation/shinecloudnet/version.Commit=$(COMMIT) \
		  -X "github.com/shinecloudfoundation/shinecloudnet/version.BuildTags=$(build_tags_comma_sep)"

ifeq ($(WITH_CLEVELDB),yes)
  ldflags += -X github.com/shinecloudfoundation/shinecloudnet/types.DBBackend=cleveldb
  ldflags += $(LDFLAGS)
endif
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

# The below include contains the tools target.
include contrib/devtools/Makefile

build: go.sum
ifeq ($(OS),Windows_NT)
	go build -mod=readonly $(BUILD_FLAGS) -o build/scloud.exe ./cmd/scloud
	go build -mod=readonly $(BUILD_FLAGS) -o build/scloudcli.exe ./cmd/scloudcli
else
	go build -mod=readonly $(BUILD_FLAGS) -o build/scloud ./cmd/scloud
	go build -mod=readonly $(BUILD_FLAGS) -o build/scloudcli ./cmd/scloudcli
endif

build-linux: go.sum
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

build-contract-tests-hooks:
ifeq ($(OS),Windows_NT)
	go build -mod=readonly $(BUILD_FLAGS) -o build/contract_tests.exe ./cmd/contract_tests
else
	go build -mod=readonly $(BUILD_FLAGS) -o build/contract_tests ./cmd/contract_tests
endif

install: go.sum check-ledger
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/scloud
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/scloudcli

install-debug: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/sclouddebug



########################################
### Tools & dependencies

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

draw-deps:
	@# requires brew install graphviz or apt-get install graphviz
	go get github.com/RobotsAndPencils/goviz
	@goviz -i ./cmd/scloud -d 2 | dot -Tpng -o dependency-graph.png

clean:
	rm -rf snapcraft-local.yaml build/

distclean: clean
	rm -rf vendor/

########################################
### Testing

test: test_unit

test_ledger_mock:
	@go test -mod=readonly `go list github.com/shinecloudfoundation/shinecloudnet/crypto` -tags='cgo ledger test_ledger_mock'

test_ledger: test_ledger_mock
	@go test -mod=readonly -v `go list github.com/shinecloudfoundation/shinecloudnet/crypto` -tags='cgo ledger'

test_unit:
	@VERSION=$(VERSION) go test -mod=readonly $(PACKAGES_NOSIMULATION) -tags='ledger test_ledger_mock'

test_race:
	@VERSION=$(VERSION) go test -mod=readonly -race $(PACKAGES_NOSIMULATION)

test_sim_nondeterminism:
	@echo "Running non-determinism test..."
	@go test -mod=readonly $(SIMAPP) -run TestAppStateDeterminism -Enabled=true \
	    -NumBlocks=100 -BlockSize=200 -Commit=true -v -timeout 24h

test_sim_custom_genesis_fast:
	@echo "Running custom genesis simulation..."
	@echo "By default, ${HOME}/.gaiad/config/genesis.json will be used."
	@go test -mod=readonly $(SIMAPP) -run TestFullAppSimulation -Genesis=${HOME}/.gaiad/config/genesis.json \
		-Enabled=true -NumBlocks=100 -BlockSize=200 -Commit=true -Seed=99 -Period=5 -v -timeout 24h

test_sim_import_export: runsim
	@echo "Running application import/export simulation. This may take several minutes..."
	$(BINDIR)/runsim -j 4 $(SIMAPP) 50 5 TestAppImportExport

test_sim_after_import: runsim
	@echo "Running application simulation-after-import. This may take several minutes..."
	$(BINDIR)/runsim -j 4 $(SIMAPP) 50 5 TestAppSimulationAfterImport

test_sim_custom_genesis_multi_seed: runsim
	@echo "Running multi-seed custom genesis simulation..."
	@echo "By default, ${HOME}/.gaiad/config/genesis.json will be used."
	$(BINDIR)/runsim -g ${HOME}/.gaiad/config/genesis.json $(SIMAPP) 400 5 TestFullAppSimulation

test_sim_multi_seed_long: runsim
	@echo "Running multi-seed application simulation. This may take awhile!"
	$(BINDIR)/runsim -j 4 $(SIMAPP) 500 50 TestFullAppSimulation

test_sim_multi_seed_short: runsim
	@echo "Running multi-seed application simulation. This may take awhile!"
	$(BINDIR)/runsim -j 4 $(SIMAPP) 50 10 TestFullAppSimulation

test_sim_benchmark_invariants:
	@echo "Running simulation invariant benchmarks..."
	@go test -mod=readonly $(SIMAPP) -benchmem -bench=BenchmarkInvariants -run=^$ \
	-Enabled=true -NumBlocks=1000 -BlockSize=200 \
	-Commit=true -Seed=57 -v -timeout 24h

.PHONY: test \
test_sim_nondeterminism \
test_sim_custom_genesis_fast \
test_sim_import_export \
test_sim_after_import \
test_sim_custom_genesis_multi_seed \
test_sim_multi_seed \
test_sim_multi_seed_short \
test_sim_benchmark_invariants

SIM_NUM_BLOCKS ?= 500
SIM_BLOCK_SIZE ?= 200
SIM_COMMIT ?= true

test_sim_benchmark:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	@go test -mod=readonly -benchmem -run=^$$ $(SIMAPP) -bench ^BenchmarkFullAppSimulation$$  \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h

test_sim_profile:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	@go test -mod=readonly -benchmem -run=^$$ $(SIMAPP) -bench ^BenchmarkFullAppSimulation$$ \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h -cpuprofile cpu.out -memprofile mem.out

test_cover:
	@export VERSION=$(VERSION); bash -x tests/test_cover.sh

lint: golangci-lint
	$(BINDIR)/golangci-lint run
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -d -s
	go mod verify
.PHONY: lint

format: tools
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs misspell -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs goimports -w -local github.com/shinecloudfoundation/shinecloudnet
.PHONY: format

check: check-unit check-build
check-all: check check-race check-cover

check-unit:
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./...

check-race:
	@VERSION=$(VERSION) go test -mod=readonly -race -tags='ledger test_ledger_mock' ./...

check-cover:
	@go test -mod=readonly -timeout 30m -race -coverprofile=coverage.txt -covermode=atomic -tags='ledger test_ledger_mock' ./...

check-build: build
	@go test -mod=readonly -p 4 `go list ./cli_test/...` -tags=cli_test -v

benchmark:
	@go test -mod=readonly -bench=. ./...


########################################
### Local validator nodes using docker and docker-compose

build-docker-scloudnode:
	$(MAKE) -C networks/local

# Run a 4-node testnet locally
localnet-start: localnet-stop
	@if ! [ -f build/node0/scloud/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/scloud:Z tendermint/scloudnode testnet --v 4 -o . --starting-ip-address 192.168.10.2 ; fi
	docker-compose up -d

# Stop testnet
localnet-stop:
	docker-compose down

localnet-reset:
	docker-compose down
	docker run --rm -v $(CURDIR)/build:/scloud:Z tendermint/scloudnode unsafe-reset-all --home /scloud/node0/scloud
	docker run --rm -v $(CURDIR)/build:/scloud:Z tendermint/scloudnode unsafe-reset-all --home /scloud/node1/scloud
	docker run --rm -v $(CURDIR)/build:/scloud:Z tendermint/scloudnode unsafe-reset-all --home /scloud/node2/scloud
	docker run --rm -v $(CURDIR)/build:/scloud:Z tendermint/scloudnode unsafe-reset-all --home /scloud/node3/scloud

setup-contract-tests-data:
	echo 'Prepare data for the contract tests'
	rm -rf /tmp/contract_tests ; \
	mkdir /tmp/contract_tests ; \
	cp "${GOPATH}/pkg/mod/${SDK_PACK}/client/lcd/swagger-ui/swagger.yaml" /tmp/contract_tests/swagger.yaml ; \
	./build/scloud init --home /tmp/contract_tests/.scloud --chain-id lcd contract-tests ; \
	tar -xzf lcd_test/testdata/state.tar.gz -C /tmp/contract_tests/

start-shine: setup-contract-tests-data
	./build/scloud --home /tmp/contract_tests/.scloud start &
	@sleep 2s

setup-transactions: start-shine
	@bash ./lcd_test/testdata/setup.sh

run-lcd-contract-tests:
	@echo "Running Scloud LCD for contract tests"
	./build/scloudcli rest-server --laddr tcp://0.0.0.0:8080 --home /tmp/contract_tests/.scloudcli --node http://localhost:26657 --chain-id lcd --trust-node true

contract-tests: setup-transactions
	@echo "Running Scloud LCD for contract tests"
	dredd && pkill scloud

# include simulations
include sims.mk

.PHONY: all build-linux install install-debug \
	go-mod-cache draw-deps clean build \
	setup-transactions setup-contract-tests-data start-shine run-lcd-contract-tests contract-tests \
	check check-all check-build check-cover check-ledger check-unit check-race

