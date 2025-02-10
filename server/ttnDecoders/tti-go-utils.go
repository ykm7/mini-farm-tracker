package main

import (
	"encoding/base64"
	"fmt"
	"log"
)

/*
*
Simple function to test parsing of the "raw" uplink TTI frm_payload to the expected bytes
Created when attempting to identify issue with S2120 messages but being parsed into anything useful
*/
func frmPayloadFormatter(s string) {
	//

	// 1. Base64 Decode
	decodedBytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Fatalf("Error decoding base64: %v", err)
		return
	}

	fmt.Printf("Decoded Bytes: %X\n", decodedBytes)
}

func main() {
	frmPayload := "SgEYMAABwnVMABBLADYAAAAAJsRMABwAAAAA"
	frmPayloadFormatter(frmPayload)
}
