package laufbeleg

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
)

// Ereignis ist eine Zeile aus `go test -json`.
//
// **Die Felder von `TestEvent` und `BuildEvent`**, wie `cmd/test2json` sie in
// Go 1.27.1 beschreibt — nachgelesen im Kopf von `src/cmd/test2json/main.go`
// der installierten Toolchain am 2026-09-22. Die Namen sind fremd und bleiben
// deshalb englisch (AGENTS.md §9).
//
// **`Time` fehlt mit Absicht.** Die Dokumentation sagt, ein Ergebnis aus dem
// Zwischenspeicher trage keine Zeit — gemessen trägt es eine, die des
// Abspielens. Wer daran erkennen wollte, ob gemessen wurde, läge falsch; das
// verrät nur die Paketzeile `(cached)`. Anfang und Ende hält der Lauf selbst
// fest.
type Ereignis struct {
	Action      string  `json:"Action"`
	Package     string  `json:"Package"`
	Test        string  `json:"Test"`
	Output      string  `json:"Output"`
	OutputType  string  `json:"OutputType"`
	Elapsed     float64 `json:"Elapsed"`
	FailedBuild string  `json:"FailedBuild"`
	ImportPath  string  `json:"ImportPath"`
}

// laengsteZeile ist die längste Zeile, die gelesen wird. Ein `t.Log` über
// eine ganze Karte kann lang werden; sechzehn MiB sind mehr, als je kam.
const laengsteZeile = 16 << 20

// LiesEreignisse liest einen Strom — und gibt zurück, was keine Zeile von
// `go test -json` war.
//
// **Eine fremde Zeile ist kein Lesefehler**, sondern ein Befund über den Lauf:
// `go test -json` schreibt nur JSON, und was dazwischensteht, kam von
// woanders. Sie wird zurückgegeben und nicht verworfen.
func LiesEreignisse(quelle io.Reader) ([]Ereignis, []string, error) {
	ereignisse := []Ereignis{}
	fremd := []string{}

	zeilen := bufio.NewScanner(quelle)
	zeilen.Buffer(make([]byte, 0, 64<<10), laengsteZeile)

	for zeilen.Scan() {
		zeile := zeilen.Bytes()
		if len(bytes.TrimSpace(zeile)) == 0 {
			continue
		}

		var ereignis Ereignis
		if json.Unmarshal(zeile, &ereignis) != nil || ereignis.Action == "" {
			fremd = append(fremd, string(zeile))

			continue
		}

		ereignisse = append(ereignisse, ereignis)
	}

	return ereignisse, fremd, zeilen.Err()
}
