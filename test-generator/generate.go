package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"

	"github.com/mit-dci/dlc-oracle-go"
)

var Log = log.New(os.Stdout,
	"INFO: ",
	log.Ldate|log.Ltime|log.Lshortfile)

func main() {
	Log.Printf("Generate files for testing on libraries in other languages.\n")

	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0755)
	var privKey [32]byte
	rand.Read(privKey[:])
	privKeyFile, _ := os.Create("testdata/privkey.hex")
	defer privKeyFile.Close()
	privKeyFile.WriteString(fmt.Sprintf("%x\n", privKey))

	pubKey := dlcoracle.PublicKeyFromPrivateKey(privKey)

	otsKeys, _ := os.Create("testdata/one-time-signing-keys.hex")
	messages, _ := os.Create("testdata/messages.hex")
	sigs, _ := os.Create("testdata/signatures.hex")
	sGsFromSig, _ := os.Create("testdata/signature-pubkeys-from-sig.hex")
	sGsFromMsg, _ := os.Create("testdata/signature-pubkeys-from-message.hex")
	defer otsKeys.Close()
	defer messages.Close()
	defer sigs.Close()
	defer sGsFromSig.Close()
	defer sGsFromMsg.Close()

	for i := 0; i < 100000; i++ {
		otsKey, _ := dlcoracle.GenerateOneTimeSigningKey()
		otsKeys.WriteString(fmt.Sprintf("%x\n", otsKey))

		rPoint := dlcoracle.PublicKeyFromPrivateKey(otsKey)

		var message [32]byte
		rand.Read(message[:])
		messages.WriteString(fmt.Sprintf("%x\n", message))

		sig, err := dlcoracle.ComputeSignature(privKey, otsKey, message[:])
		if err != nil {
			panic(err)
		}
		sigs.WriteString(fmt.Sprintf("%x\n", sig))

		sGFromSig := dlcoracle.PublicKeyFromPrivateKey(sig)
		sGsFromSig.WriteString(fmt.Sprintf("%x\n", sGFromSig))

		sGFromMsg, err := dlcoracle.ComputeSignaturePubKey(pubKey, rPoint, message[:])
		if err != nil {
			panic(err)
		}
		sGsFromMsg.WriteString(fmt.Sprintf("%x\n", sGFromMsg))

		if i%100 == 0 {
			fmt.Printf("\rWriting test files ... [%d/100000]", i)
		}

	}
	fmt.Println("\rWriting test files ... 100% completed\nDone.")

}
