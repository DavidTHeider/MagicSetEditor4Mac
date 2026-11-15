package app_test

// **Ein offenes Sideboard erreicht die Flächen** — Befund 1 des Nachlaufs
// über Phase 4.6.
//
// # Was hier geprüft wird und was nicht
//
// **Hier: die Abbildung** — dass die Bindung das Feld überhaupt nennt. Diese
// Schicht arbeitet gegen Doppel (§3), und ein Doppel legt nichts ab; was
// durch **Bestand und Fläche** wieder herauskommt, steht deshalb in
// `internal/flaechen/formatsideboard_test.go`.
//
// Der Befund lag an **zwei** Grenzen zugleich: die Tabelle kannte das Feld
// nicht, und das DTO kannte es auch nicht. Ein Test je Schicht hätte beides
// nicht gezeigt — zwei grüne Seiten einer Grenze sind kein Nachweis für die
// Grenze.

import "testing"

// **Die Bauart »limitiert« kommt mit offenem Sideboard über die Bindung.**
func TestDieBauartLimitiertNenntIhrOffenesSideboard(t *testing.T) {
	anwendung, _, _, _, _ := deckanwendung(t)

	antwort := anwendung.Bauarten()
	if antwort.Fehler != nil {
		t.Fatalf("die Bauarten: %v", antwort.Fehler)
	}

	var gefunden bool

	for _, bauart := range antwort.Bauarten {
		if bauart.Bauart != "limitiert" {
			continue
		}

		gefunden = true

		if !bauart.Format.SideboardOffen {
			t.Errorf("»limitiert« nennt kein offenes Sideboard: %+v", bauart.Format)
		}
	}

	if !gefunden {
		t.Fatal("die Bauart »limitiert« steht nicht in der Liste")
	}
}
