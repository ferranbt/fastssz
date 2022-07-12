
# FastSSZ

Clone:

```
$ git clone git@github.com:ferranbt/fastssz.git
```

Download the eth2.0 spec tests

```
$ make get-spec-tests
```

Regenerate the test spec encodings:

```
$ make build-spec-tests
```

Generate encodings for a specific package:

```
$ go run sszgen/*.go --path ./ethereumapis/eth/v1alpha1 [--objs BeaconBlock,Eth1Data]
```

Optionally, you can specify the objs you want to generate. Otherwise, it will generate encodings for all structs in the package. Note that if a struct does not have 'ssz' tags when required (i.e size of arrays), the generator will fail.

By default, it generates a file with the prefix '_encoding.go' for each file that contains a generated struct. Optionally, you can combine all the outputs in a single file with the 'output' flag.

```
$ go run sszgen/*.go --path ./ethereumapis/eth/v1alpha1 --output ./ethereumapis/eth/v1alpha1/encoding.go
```

Test the spectests:

```
$ go test -v ./spectests/... -run TestSpec
```

Run the fuzzer:

```
$ FUZZ_TESTS=True go test -v ./spectests/... -run TestFuzz
```

To install the generator run:

```
$ go get github.com/ferranbt/fastssz/sszgen
```

Benchmark (BeaconBlock):

```
$ go test -v ./spectests/... -run=XXX -bench=.
goos: linux
goarch: amd64
pkg: github.com/ferranbt/fastssz/spectests
cpu: AMD Ryzen 5 2400G with Radeon Vega Graphics    
BenchmarkMarshalFast
BenchmarkMarshalFast-8        	  268454	      4166 ns/op	    8192 B/op	       1 allocs/op
BenchmarkMarshalSuperFast
BenchmarkMarshalSuperFast-8   	  883546	      1226 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnMarshalFast
BenchmarkUnMarshalFast-8      	   67159	     17772 ns/op	   11900 B/op	     210 allocs/op
BenchmarkHashTreeRootFast
BenchmarkHashTreeRootFast-8   	   24508	     45571 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/ferranbt/fastssz/spectests	5.501s
```

# Package reference

To reference a struct from another package use the '--include' flag to point to that package.

Example:

```
$ go run sszgen/*.go --path ./example2
$ go run sszgen/*.go --path ./example 
[ERR]: could not find struct with name 'Checkpoint'
$ go run sszgen/*.go --path ./example --include ./example2
```

There are some caveats required to use this functionality.
- If multiple input paths import the same package, all of them need to import it with the same alias if any.
- If the folder of the package is not the same as the name of the package, any input file that imports this package needs to do it with an alias.
