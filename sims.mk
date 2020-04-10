#!/usr/bin/make -f

########################################
### Simulations

SIMAPP = github.com/shinecloudfoundation/shinecloudnet/app

sim-scloud-nondeterminism:
	@echo "Running nondeterminism test..."
	@go test -mod=readonly $(SIMAPP) -run TestAppStateDeterminism -Enabled=true \
		-NumBlocks=100 -BlockSize=200 -Commit=true -v -timeout 24h

sim-scloud-custom-genesis-fast:
	@echo "Running custom genesis simulation..."
	@echo "By default, ${HOME}/.scloud/config/genesis.json will be used."
	@go test -mod=readonly $(SIMAPP) -run TestFullScloudSimulation -Genesis=${HOME}/.scloud/config/genesis.json \
		-Enabled=true -NumBlocks=100 -BlockSize=200 -Commit=true -Seed=99 -Period=5 -v -timeout 24h

sim-scloud-fast:
	@echo "Running quick Scloud simulation. This may take several minutes..."
	@go test -mod=readonly $(SIMAPP) -run TestFullScloudSimulation -Enabled=true -NumBlocks=100 -BlockSize=200 -Commit=true -Seed=99 -Period=5 -v -timeout 24h

sim-scloud-import-export: runsim
	@echo "Running Scloud import/export simulation. This may take several minutes..."
	$(GOPATH)/bin/runsim $(SIMAPP) 25 5 TestScloudImportExport

sim-scloud-simulation-after-import: runsim
	@echo "Running Scloud simulation-after-import. This may take several minutes..."
	$(GOPATH)/bin/runsim $(SIMAPP) 25 5 TestScloudSimulationAfterImport

sim-scloud-custom-genesis-multi-seed: runsim
	@echo "Running multi-seed custom genesis simulation..."
	@echo "By default, ${HOME}/.scloud/config/genesis.json will be used."
	$(GOPATH)/bin/runsim $(SIMAPP) -g ${HOME}/.scloud/config/genesis.json 400 5 TestFullScloudSimulation

sim-scloud-multi-seed: runsim
	@echo "Running multi-seed Scloud simulation. This may take awhile!"
	$(GOPATH)/bin/runsim $(SIMAPP) 400 5 TestFullScloudSimulation

sim-benchmark-invariants:
	@echo "Running simulation invariant benchmarks..."
	@go test -mod=readonly $(SIMAPP) -benchmem -bench=BenchmarkInvariants -run=^$ \
	-Enabled=true -NumBlocks=1000 -BlockSize=200 \
	-Commit=true -Seed=57 -v -timeout 24h

SIM_NUM_BLOCKS ?= 500
SIM_BLOCK_SIZE ?= 200
SIM_COMMIT ?= true
sim-scloud-benchmark:
	@echo "Running Scloud benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	@go test -mod=readonly -benchmem -run=^$$ $(SIMAPP) -bench ^BenchmarkFullScloudSimulation$$  \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h

sim-scloud-profile:
	@echo "Running Scloud benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	@go test -mod=readonly -benchmem -run=^$$ $(SIMAPP) -bench ^BenchmarkFullScloudSimulation$$ \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h -cpuprofile cpu.out -memprofile mem.out


.PHONY: runsim sim-scloud-nondeterminism sim-scloud-custom-genesis-fast sim-scloud-fast sim-scloud-import-export \
	sim-scloud-simulation-after-import sim-scloud-custom-genesis-multi-seed sim-scloud-multi-seed \
	sim-benchmark-invariants sim-scloud-benchmark sim-scloud-profile
