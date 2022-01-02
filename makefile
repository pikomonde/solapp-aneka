
run-solana:
	solana-test-validator

# clean-bpf-c:
# 	V=1 make -C ./bpf/program-c clean

build-bpf-c-aneka_hello:
	V=1 make -C ./bpf/program-c aneka_hello
build-bpf-c-aneka_guess_number:
	V=1 make -C ./bpf/program-c aneka_guess_number

deploy-bpf-aneka_hello:
	solana program deploy output/program/aneka_hello.so
deploy-bpf-aneka_guess_number:
	solana program deploy output/program/aneka_guess_number.so

run-cli-aneka_hello:
	go run cmd/cli/aneka_hello/app.go
run-cli-aneka_guess_number:
	go run cmd/cli/aneka_guess_number/app.go
