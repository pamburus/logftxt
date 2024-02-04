#!/bin/sh

set -e

rm -f *.out
go test -cpuprofile logftxt-color-same.cpu.out -bench . -benchmem -test.bench BenchmarkEncoder/logftxt/Color/SameLoggerID
go test -cpuprofile logftxt-color-new.cpu.out -bench . -benchmem -test.bench BenchmarkEncoder/logftxt/Color/NewLoggerID
go test -cpuprofile logftxt-no-color-same.cpu.out -bench . -benchmem -test.bench BenchmarkEncoder/logftxt/NoColor/SameLoggerID
go test -cpuprofile logftxt-no-color-new.cpu.out -bench . -benchmem -test.bench BenchmarkEncoder/logftxt/NoColor/NewLoggerID
go test -cpuprofile logftext-color-same.cpu.out -bench . -benchmem -test.bench BenchmarkEncoder/logftext/Color/SameLoggerID
go test -cpuprofile logftext-color-new.cpu.out -bench . -benchmem -test.bench BenchmarkEncoder/logftext/Color/NewLoggerID
go test -cpuprofile logftext-no-color-same.cpu.out -bench . -benchmem -test.bench BenchmarkEncoder/logftext/NoColor/SameLoggerID
go test -cpuprofile logftext-no-color-new.cpu.out -bench . -benchmem -test.bench BenchmarkEncoder/logftext/NoColor/NewLoggerID
