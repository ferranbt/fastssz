
# FastSSZ

Clone:

```
$ git clone git@github.com:ferranbt/fastssz.git --recursive
```

Regenerate the test spec encodings:

```
$ make build-spec-tests
```

Generate encodings for a specific package:

```
$ go run sszgen/main.go --path ./ethereumapis/eth/v1alpha1 [--objs BeaconBlock,Eth1Data]
```

Optionally, you can specify the objs you want to generate. Otherwise, it will generate encodings for all structs in the package. Note that if a struct does not have 'ssz' tags when required (i.e size of arrays), the generator will fail.

Test the spectests:

```
$ go test -v ./spectests/... -run TestSpec
```

Run the fuzzer with BeaconBlockBody

```
$ FUZZ_TESTS=True go test -v ./spectests/... -run TestFuzz
```

Benchmark (BeaconBlock):

```
$ go test -v ./spectests/... -run=XXX -bench=.
goos: linux
goarch: amd64
pkg: github.com/ferranbt/fastssz/spectests
BenchmarkMarshalGoSSZ-4       	    1366	    753160 ns/op	  115112 B/op	    8780 allocs/op
BenchmarkMarshalFast-4        	  240765	      5093 ns/op	   18432 B/op	       1 allocs/op
BenchmarkMarshalSuperFast-4   	  377835	      3041 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnMarshalGoSSZ-4     	     847	   1395097 ns/op	  144608 B/op	    8890 allocs/op
BenchmarkUnMarshalFast-4      	   43824	     27190 ns/op	   31024 B/op	     577 allocs/op
PASS
ok  	github.com/ferranbt/fastssz/spectests	6.608s
```
