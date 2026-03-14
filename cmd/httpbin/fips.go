//go:build fips

package main

import _ "crypto/tls/fipsonly" // Restricts TLS to FIPS-approved settings when built with GOEXPERIMENT=boringcrypto
