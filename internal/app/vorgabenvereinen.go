package app

// **Zwei Domänen der Vorgaben vereinen, die neue gewinnt** — #2424, Befund 3
// des Nachlaufs über Phase 3 von 0.9.0.
//
// `defaults import` **ersetzt** eine Domäne. Scheiterte ein erster Umzug vor
// dem Import, stand die neue ohne Marke da; AppKit und WebKit schrieben
// hinein, und der nächste Start importierte die alte ganz — was die neue
// Fassung geschrieben hatte, war fort. Jetzt wird vorher vereint: jeder
// Schlüssel der neuen Domäne bleibt, aus der alten kommt nur, was fehlt.
//
// # Warum hier und nicht mit `plutil`
//
// **`plutil` kennt kein Vereinen**, und einen Wert über `defaults write`
// zurückzuschreiben verlöre seine Art (Daten, Datum, Wörterbuch). Hier werden
// beide Domänen als XML gelesen (`plutil -convert xml1`), und **der Wert jedes
// Schlüssels reist als der Text, der er war** — nichts wird gedeutet.

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
)

// plisteintrag ist ein Schlüssel des obersten Wörterbuchs und sein Wert als
// roher XML-Text.
type plisteintrag struct {
	schluessel string
	wert       []byte
}

// plisteintraege liest das oberste Wörterbuch einer XML-Plist.
func plisteintraege(roh []byte) ([]plisteintrag, error) {
	leser := xml.NewDecoder(bytes.NewReader(roh))
	tiefe := 0

	var (
		eintraege  []plisteintrag
		schluessel *string
	)

	for {
		vorher := leser.InputOffset()

		marke, fehler := leser.Token()
		if errors.Is(fehler, io.EOF) {
			return eintraege, nil
		}

		if fehler != nil {
			return nil, fmt.Errorf("die Vorgaben lesen: %w", fehler)
		}

		switch element := marke.(type) {
		case xml.StartElement:
			if tiefe < 2 {
				tiefe++

				continue
			}

			// **Im obersten Wörterbuch**: abwechselnd ein Schlüssel und sein Wert.
			if element.Name.Local == "key" && schluessel == nil {
				var name string
				if fehler := leser.DecodeElement(&name, &element); fehler != nil {
					return nil, fmt.Errorf("einen Schlüssel der Vorgaben lesen: %w", fehler)
				}

				schluessel = &name

				continue
			}

			if schluessel == nil {
				return nil, errors.New("die Vorgaben tragen einen Wert ohne Schlüssel")
			}

			if fehler := leser.Skip(); fehler != nil {
				return nil, fmt.Errorf("den Wert von %s lesen: %w", *schluessel, fehler)
			}

			eintraege = append(eintraege, plisteintrag{schluessel: *schluessel, wert: roh[vorher:leser.InputOffset()]})
			schluessel = nil
		case xml.EndElement:
			tiefe--
		}
	}
}

// vorgabenVereinen gibt die alte Domäne mit jedem Schlüssel der neuen
// darüber — und mit der Marke des Umzugs, die `markenwert` trägt.
func vorgabenVereinen(alt, neu []byte, markenwert string) ([]byte, error) {
	alte, fehler := plisteintraege(alt)
	if fehler != nil {
		return nil, fehler
	}

	neue, fehler := plisteintraege(neu)
	if fehler != nil {
		return nil, fehler
	}

	var text strings.Builder

	text.WriteString(xml.Header + `<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">` +
		"\n<plist version=\"1.0\">\n<dict>\n")

	schreibe := func(schluessel string, wert []byte) {
		text.WriteString("\t<key>")
		_ = xml.EscapeText(&text, []byte(schluessel)) //nolint:errcheck // ein strings.Builder schreibt ohne Fehler
		text.WriteString("</key>\n\t")
		text.Write(wert)
		text.WriteString("\n")
	}

	gesehen := map[string]bool{umzugsmarke: true}

	for _, eintrag := range neue {
		if !gesehen[eintrag.schluessel] {
			gesehen[eintrag.schluessel] = true
			schreibe(eintrag.schluessel, eintrag.wert)
		}
	}

	for _, eintrag := range alte {
		if !gesehen[eintrag.schluessel] {
			gesehen[eintrag.schluessel] = true
			schreibe(eintrag.schluessel, eintrag.wert)
		}
	}

	var wert strings.Builder

	_ = xml.EscapeText(&wert, []byte(markenwert)) //nolint:errcheck // dito
	schreibe(umzugsmarke, []byte("<string>"+wert.String()+"</string>"))
	text.WriteString("</dict>\n</plist>\n")

	return []byte(text.String()), nil
}
