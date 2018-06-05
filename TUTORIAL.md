# Building a Discreet Log Contract Oracle in Go

In 2017, [Tadge Dryja](https://twitter.com/tdryja) published a [paper](https://adiabat.github.io/dlc.pdf) on Discreet Log Contracts. 

By creating a Discreet Log Contract, Alice can form a contract paying Bob some time in the future, based on preset conditions, without committing any details of those conditions to the blockchain. Therefore it is discreet in the sense that no external observer can learn its existence from the public ledger. This contract depends on an external entity or entities publishing a signed message at some point in the future (before the expiration of the contract). The contents of this signed message determine the division of the funds committed to the contract. This external entity is called an “oracle”. Using Discreet Log Contracts, the signature published by the oracle gives each participant of the contract the possibility to claim the amount from the contract that is due him without the need for cooperation from the other party. 

This tutorial will describe you how to build a Discreet Log Contract "oracle". This tutorial describes how to do this in Go, but you can also use [NodeJS](http://google.nl) or [.NET Core](https://google.nl)

### Set up a new project

Firstly, set up a new empty project and include the correct libraries. We start by creating the project folder and add the main program file to it.

```
mkdir $GOROOT/src/tutorial
cd $GOROOT/src/tutorial
touch $GOROOT/src/tutorial/tutorial.go
```

Next, open the file `tutorial.go` in your favorite editor, and add this to it:

```
package main

import (
	
)

func main() {
	
}
```

### Generate and save the oracle's private key

Next, we'll need to have a private key. This private key is used in conjunction with a unique one-time-signing key for each message. The private key of the oracle never changes, and its public key is incorporated into Discreet Log Contracts that people form. So if we lose access to this key, people's contracts might be unable to settle. In this example, we'll store the key in a simple hexadecimal format on disk. This is not secure, and should not be considered for production scenarios. However, to illustrate the working of the library it is sufficient.

So we add this function to the `tutorial.go` file:

```
func getOrCreateKey() ([32]byte, error) {
	// Initialize the byte array that will hold the generated key
	var priv [32]byte
		
	// Check if the privatekey.hex file exists
	_, err := os.Stat("privatekey.hex");
	if err != nil {
		if os.IsNotExist(err) {
			// If not, generate a new private key by reading 32 random bytes
			rand.Read(priv[:])
			
			// Convert the key in to a hexadecimal format
			keyhex := fmt.Sprintf("%x\n", priv[:])
			
			// Save the hexadecimal value into the file
			err := ioutil.WriteFile("privatekey.hex", []byte(keyhex), 0600)
			
			if err != nil {
				// Unable the save the key file, return the error
				return priv, err
			}
		} else {
			// Some other error occurred while checking the file's existence, return the error
			return priv, err
		}
	}
	
	// At this point, the file either existed or is created. Read the private key from the file
	keyhex, err := ioutil.ReadFile("privatekey.hex")
	if err != nil {
		// Unable to read the key file, return the error
		return priv, err
	}
	
	// Trim any whitespace from the file's contents
	keyhex = []byte(strings.TrimSpace(string(keyhex)))
	
	// Decode the hexadecimal format into a byte array
	key, err := hex.DecodeString(string(keyhex))
	if err != nil {
		// Unable to decode the hexadecimal format, return the error 
		return priv, err
	}
	
	// Copy the variable-width byte array key into priv ([32]byte)
	copy(priv[:], key[:])
	
	// Return the key
	return priv, nil
}
```

and then we adjust the `main()` function to use it, and add the necessary imports:

```
package main

import (
	"crypto/rand"
	"encoding/hex"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	privateKey, err := getOrCreateKey()
	if err != nil {
		panic(err)
	}
}
```

### Derive and print out the public key

Next, we'll use the DLC library to generate the public key from the private key and print it out to the console:

```
(...)

import (
	(...)
	
	"github.com/mit-dci/dlc-oracle-go"
)

func main() {
	(...)
	
	// Print out the public key for the oracle
	pubKey := dlcoracle.PublicKeyFromPrivateKey(privateKey)
	fmt.Printf("Oracle public key: %x\n", pubKey)

}
```

In your terminal window, compile the application and run it:

```
go build .
./tutorial
```

The program should show an output similar to this:

```
Oracle public key: 032382208af6ab43b8d21e71edf73ec8f37070a9b4fc4547f406bcd744b4e8e7b3
```

### Create a loop that publishes oracle values

Next, we'll add a loop to the oracle that will take the following steps:

* Generate a new one-time signing key
* Print out the public key to that signing key (the "R Point")
* Wait 1 minute
* Sign a random numerical value with the one-time key 
* Print out the value and signature

Using the oracle's public key and the "R Point" (public key to the one-time signing key), people can use LIT to form a Discreet Log Contract, and use your signature when you publish it to settle the contract.

So for a regular DLC use case, you would publish your oracle's public key and the R-point for each time / value you will publish onto a website or some other form of publication, so that people can obtain the keys and use them in their contracts. When the time arrives you have determined the correct value, and sign it, you publish both the value and your signature so the contract participants can use those values to settle the contract.

As for the one-time signing key, this has the same security requirements as the oracle's private key. If this key is lost, contracts that depend on it cannot be settled. It is therefore important to save this key somewhere safe. Just keeping it in memory as we do in this example is not good practice for production scenarios. 

One last note on the one-time signing key: The reason that it's named this, is that you can only use it once. Even though there's no technical limitation of you producing two signatures with that key, doing so using the signature scheme DLC uses will allow people to derive your oracle's private key from the data you published.

OK, back to the code. So, first we add the generation of the one-time signing key and printing out the corresponding public key (R Point) to the loop.

```
(...)
import (
    (...)
    "time"
    
)

func main() {
	(...)
	for {
		// Generate a new one-time signing key
		privPoint, err := dlcoracle.GenerateOneTimeSigningKey()
		if err != nil {
			panic(err)
		}

		// Generate the public key to the one-time signing key (R-point) and print it out
		rPoint := dlcoracle.PublicKeyFromPrivateKey(privPoint)
		fmt.Printf("R-Point for next publication: %x\n", rPoint)

		// Sleep for 1 minute
		time.Sleep(time.Second * 60)
	}
}
```

Go ahead and build the code once more and run it again. You'll see an output similar to this:

```
Oracle public key: 032382208af6ab43b8d21e71edf73ec8f37070a9b4fc4547f406bcd744b4e8e7b3
R-Point for next publication: 0377e45381f0545af64f413be03e1a236ab64274bc4c81cf81c73add104c360976
```

If you let the program run, after a minute it will print out another R-Point. 

Next step is to actually generate a random value and sign it. Using the DLC library, this is quite easy:

```
(...)
import (
    (...)
    mathrand "math/rand"
    
)

func main() {
	(...)
	for {
		(...)
		// Generate random value between 10000 and 20000
		randomValue := uint64(mathrand.Int31n(10000) + 10000)

		// Generate message to sign. Uses the same encoding as expected by LIT when settling the contract
		message := dlcoracle.GenerateNumericMessage(randomValue)

		// Sign the message
		signature, err := dlcoracle.ComputeSignature(privateKey, privPoint, message)
		if err != nil {
			panic(err)
		}

		// Print out the value and signature
		fmt.Printf("Signed message. Value: %d\nSignature: %x\n", randomValue, signature)
	}
}
```

Next, build and run your code again. It will print out something like this (you'll have to wait 60 seconds for the value to be published. If you want, you can change that interval to make it output the data quicker).

```
Oracle public key: 032382208af6ab43b8d21e71edf73ec8f37070a9b4fc4547f406bcd744b4e8e7b3
R-Point for next publication: 02f650ab605757ef687ffa1ab0e493ed8e8c054adf6ff010605656ed3a992e6ef4
Signed message. Value: 18081
Signature: 8b044f80589821a8a6d1837752dfdda5809236b7ce72e63e789ae9288c2937dc
R-Point for next publication: 03126cdf6340bd1277abe2f2bef6ab21c594e44398876113efebb2e6a855ec0ae1
```

### Done!

And that's all there is to it. Next steps you could take involve changing how you secure the private key(s), how you publish your public key and the R-points (to something other than your console), and to sign actual real-world values using this set-up. If you publish interesting data feeds using this mechanism, people can base real Discreet Log Contracts on them. If you created any cool oracles, feel free to send a pull request to our [samples repository](https://github.com/mit-dci/dlc-oracle-go-samples) so we can include them for other people to enjoy. You'll also find the complete code for this tutorial there as [one of the samples](https://github.com/mit-dci/dlc-oracle-go-samples/tree/master/tutorial)