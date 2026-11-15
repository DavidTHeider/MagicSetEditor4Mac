// **Der technische Bereich** — Werkstück 5 von Phase 3.3
// (#1164).
package app_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Der Ort ist das Verzeichnis des laufenden Programms.**
//
// Dort legt `scripts/release.sh` die drei Werkzeuge ab — neben dem Fenster in
// `…/Contents/MacOS/`.
func TestDerTechnikortIstDasProgrammverzeichnis(t *testing.T) {
	antwort := app.NeueApp("").Technikangaben(t.Context(), "")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Ort == "" {
		t.Error("der Ort fehlt")
	}
}

// **Die Zeilen tragen den vollen Pfad.**
//
// Sie liegen in keinem `PATH`; eine Zeile mit bloßem `mse` liefe ins Leere
// oder — schlimmer — in ein fremdes Programm desselben Namens.
func TestDieBefehleTragenDenVollenPfad(t *testing.T) {
	antwort := app.NeueApp("").Technikangaben(t.Context(), "")

	if len(antwort.Zeilen) == 0 {
		t.Fatal("es gibt keine Zeile")
	}

	for _, zeile := range antwort.Zeilen {
		if !strings.Contains(zeile.Befehl, antwort.Ort) {
			t.Errorf("%q nennt den Ort nicht", zeile.Befehl)
		}

		if zeile.Schluessel == "" {
			t.Errorf("%q hat keinen Schlüssel", zeile.Befehl)
		}
	}
}

// **Ein Set bringt zwei weitere Zeilen mit — und ohne Set gibt es sie nicht.**
func TestEinSetBringtWeitereZeilen(t *testing.T) {
	ohne := app.NeueApp("").Technikangaben(t.Context(), "")
	mit := app.NeueApp("").Technikangaben(t.Context(), "7")

	if len(mit.Zeilen) <= len(ohne.Zeilen) {
		t.Fatalf("mit Set %d Zeilen, ohne %d", len(mit.Zeilen), len(ohne.Zeilen))
	}

	gefunden := false

	for _, zeile := range mit.Zeilen {
		if strings.HasSuffix(zeile.Befehl, "karten 7") {
			gefunden = true
		}
	}

	if !gefunden {
		t.Errorf("keine Zeile nennt das Set: %+v", mit.Zeilen)
	}
}

// **Ein Ort mit Leerzeichen kommt in Hochkommata.**
//
// Unter macOS ist das der Normalfall: `/Applications/MagicSetEditor4Mac.app/…`
// geht ohne, `~/Library/Application Support/…` nicht. Eine Zeile, die dort
// zerfiele, wäre eine Vorlage, die beim Einfügen kaputtgeht.
//
// **Der Lauf hat bis zum 2026-08-27 die falsche Schreibweise verlangt** — ein
// doppeltes Anführungszeichen, das seit der Berichtigung der Quotierung nicht
// mehr kommt. Aufgefallen ist es nicht, weil er auf einem Rechner ohne
// Leerzeichen im Pfad **übersprungen** wird: ein Auschecken in
// `~/Meine Projekte/` wäre rot geworden, obwohl die Ausgabe stimmt.
// Angestrichen vom Nachlauf des Zweitblicks über Phase 3.3.
func TestEinOrtMitLeerzeichenKommtInHochkommata(t *testing.T) {
	antwort := app.NeueApp("").Technikangaben(t.Context(), "")

	if !strings.Contains(antwort.Ort, " ") {
		t.Skip("dieses Programm liegt in einem Pfad ohne Leerzeichen")
	}

	for _, zeile := range antwort.Zeilen {
		if !strings.Contains(zeile.Befehl, "'") {
			t.Errorf("%q steht ohne Hochkommata da", zeile.Befehl)
		}
	}
}

// **Leere Listen und kein `nil`** — auch auf dem Fehlerweg.
func TestDieTechnikangabenGebenLeereListen(t *testing.T) {
	antwort := app.NeueApp("").Technikangaben(t.Context(), "")

	if antwort.Vorhanden == nil || antwort.Zeilen == nil {
		t.Errorf("Vorhanden %v, Zeilen %v", antwort.Vorhanden, antwort.Zeilen)
	}
}

// **Und die Bindung am Fenster geht denselben Weg.**
func TestDieTechnikbindungGehtDenselbenWeg(t *testing.T) {
	if antwort := app.NeueBindungen(app.NeueApp("")).Technikangaben(t.Context(), ""); antwort.Ort == "" {
		t.Error("die Bindung gab keinen Ort")
	}
}
