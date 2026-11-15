package main

// **Die Verdrahtung der Profilvorgaben im Dienst** — Werkstück 15 von
// Phase 4.12
// (#1351).

import (
	"os"
	"path/filepath"
	"testing"
)

// **Alle fünf Arten kommen an — auch die leeren.**
func TestDieVorgabennahtDesDienstesErreichtDieAnwendung(t *testing.T) {
	db, zuhause := frischerTokenbestandMitZuhause(t)
	naht := fremdvorgaben{db: db}

	mitSpielpaket(t, zuhause)

	zeilen, fehler := naht.Vorgaben(probekontext(t), "1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(zeilen) != 5 {
		t.Fatalf("es sind %d Arten: %+v", len(zeilen), zeilen)
	}

	// **Jede Zeile trägt Listen und kein `nil`** (§3) — sonst stünde in der
	// Antwort `null` statt einer leeren Liste.
	for _, zeile := range zeilen {
		if zeile.Werte == nil || zeile.Moeglich == nil || zeile.Verwaist == nil {
			t.Errorf("%s trägt eine nil-Liste: %+v", zeile.Art, zeile)
		}
	}
}

// **Das Setzen gibt den ganzen Stand zurück.**
func TestDieVorgabennahtDesDienstesSetzt(t *testing.T) {
	naht := fremdvorgaben{db: frischerTokenbestand(t)}

	gesetzt, fehler := naht.VorgabeSetzen(
		probekontext(t), "1", "lagerort", []string{})
	if fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	if len(gesetzt.Vorgaben) != 5 || gesetzt.Nachladefehler != nil {
		t.Errorf("zurück kamen %d Arten (%+v)", len(gesetzt.Vorgaben), gesetzt.Nachladefehler)
	}
}

// **Und beide Wege reichen ihr Nein durch.**
func TestDieVorgabennahtDesDienstesReichtIhrNeinDurch(t *testing.T) {
	naht := fremdvorgaben{db: frischerTokenbestand(t)}

	if _, fehler := naht.Vorgaben(probekontext(t), "4711"); fehler == nil {
		t.Error("ein Profil, das es nicht gibt, kam durch")
	}

	if _, fehler := naht.VorgabeSetzen(
		probekontext(t), "1", "kartenrücken", []string{"1"}); fehler == nil {
		t.Error("eine Art, die es nicht gibt, kam durch")
	}
}

// mitSpielpaket legt ein Spielpaket im Zuhause dieses Laufs an.
//
// **Damit das Angebot nicht leer ist.** Ohne ein Paket gibt jede der fünf
// Arten eine leere Liste — und die Abbildung der Wahlen bliebe unbetreten
// (§3): ein Verbund, der die Beschriftung wegwürfe, wäre grün.
func mitSpielpaket(t *testing.T, zuhause string) {
	t.Helper()

	// **Derselbe Weg wie `app.Paketverzeichnis()`**, von Hand nachgebaut: der
	// Lauf setzt `HOME` selbst und kann ihn der Fassade nicht abfragen.
	ort := filepath.Join(zuhause, "Library", "Application Support",
		"MagicSetEditor4Mac", "pakete", "probe.mse-game")

	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ort, "game"),
		[]byte("mse version: 2.0.0\nshort name: Probe\n"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}
}
