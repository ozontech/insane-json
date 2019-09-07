.PHONY: bench
bench:
	go test  . -benchmem -bench Benchmark -count 1 -run _ -cpuprofile cpu.out -memprofile mem.out

.PHONY: test
test:
	go test . -count 1 -v