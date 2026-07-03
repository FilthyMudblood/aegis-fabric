package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/FilthyMudblood/aegis-fabric/internal/policyplane/tlsconfig"
)

func main() {
	out := flag.String("out", "./.afp-tls", "output directory for dev mTLS material")
	flag.Parse()

	if err := tlsconfig.GenerateDevMaterial(*out); err != nil {
		fmt.Fprintf(os.Stderr, "generate tls material: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("generated AFP policy mTLS material in %s\n", *out)
}
