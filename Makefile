.PHONY: bench
bench:
	go test  . -benchmem -bench BenchmarkFair -count 1 -run _ -cpuprofile cpu.out -memprofile mem.out

.PHONY: bench-values
bench-values:
	go test  . -benchmem -bench BenchmarkValue -count 1 -run _ -cpuprofile cpu.out -memprofile mem.out

.PHONY: test
test:
	go test . -count 1 -v