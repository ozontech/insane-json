.PHONY: bench-decode
bench-decode:
	go test  . -benchmem -bench BenchmarkDecode -count 1 -run _ -cpuprofile cpu.out -memprofile mem.out

.PHONY: bench-encode
bench-encode:
	go test  . -benchmem -bench BenchmarkEncode -count 1 -run _ -cpuprofile cpu.out -memprofile mem.out

.PHONY: bench-ws
bench-ws:
	go test  . -benchmem -bench BenchmarkWS -count 1 -run _ -cpuprofile cpu.out -memprofile mem.out

.PHONY: bench-dig
bench-dig:
	go test  . -benchmem -bench BenchmarkDig -count 1 -run _ -cpuprofile cpu.out -memprofile mem.out

.PHONY: bench-decode-int
bench-decode-int:
	go test  . -benchmem -bench BenchmarkDecodeInt -count 1 -run _ -cpuprofile cpu.out -memprofile mem.out

.PHONY: bench-escape-string
bench-escape-string:
	go test  . -benchmem -bench BenchmarkEscapeString -count 1 -run _ -cpuprofile cpu.out -memprofile mem.out

.PHONY: test
test:
	go test . -count 1 -v