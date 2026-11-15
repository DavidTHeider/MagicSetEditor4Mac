// Command lauf misst — und belegt, dass gemessen wurde (QA-01,
// #1733).
//
//	go run ./internal/laufbeleg/lauf -plan docs/belege/fremdpakete.plan.json
//	go run ./internal/laufbeleg/lauf -plan docs/belege/fremdpakete.plan.json -pruefen
//
// Das erste misst und schreibt den Beleg, das zweite hält ihn gegen den
// heutigen Stand. **Ein Werkzeug der Entwicklung**: kein ausgeliefertes
// Programm kennt es.
package main

import "os"

func main() {
	os.Exit(lauf(os.Args[1:], os.Stdout, os.Stderr, echterKasten()))
}
