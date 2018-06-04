# Testdata generator

This utility will output a testdata folder, containing files needed to run the tests on libraries in other languages. This verifies to some degreet that these other libraries produce the same output as the implementation in Go, which is used in LIT for executing the Discreet Log Contracts.

## How to use

Build the executable using:

```
go get github.com/mit-dci/dlc-oracle-go/test-generator
cd $GOPATH/src/github.com/mit-dci/dlc-oracle-go/test-generator
go build
```

Then run the executable using:

```
./test-generator
```

The folder `testdata` that is created, should be copied into the folder containing the `test` sample from any of the other libraries such as :

[NodeJS]()
[.NET Core]()

Then execute that test sample, and verify that no errors are encountered while running it.