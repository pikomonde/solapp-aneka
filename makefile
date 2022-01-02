
run-solana:
	solana-test-validator

# clean-bpf-c:
# 	V=1 make -C ./bpf/program-c clean

build-bpf-c-aneka_hello:
	V=1 make -C ./bpf/program-c aneka_hello

deploy-bpf-aneka_hello:
	solana program deploy output/program/aneka_hello.so

run-cli-aneka_hello:
	go run cmd/cli/app.go
