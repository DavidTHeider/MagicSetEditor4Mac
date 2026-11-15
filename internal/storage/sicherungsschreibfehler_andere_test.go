//go:build !darwin && !linux

package storage_test

import "os"

// grenzschalter und sichereUnterGrenze gibt es nur, wo es `RLIMIT_FSIZE` gibt
// (`sicherungsschreibfehler_test.go`); hier hält `TestMain` nur den Namen.
const grenzschalter = "MSE_SICHERUNGSGRENZE"

func sichereUnterGrenze(string) { os.Exit(2) }
