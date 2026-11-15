package mcpnaehte

// **Die Verdrahtung der Profilvorgaben** — Werkstück 15 von Phase 4.12
// (#1351).
//
// # Was dieser Lauf hält
//
// Dass die Wurzel den Weg **erreicht** und die Abbildung die Felder nicht
// vertauscht — nicht, was der Dienst dahinter tut. Der Arbeitsablauf über alle
// Flächen steht in `internal/flaechen`.

import (
	"os"
	"path/filepath"
	"testing"
)

// **Alle fünf Arten kommen an — auch die leeren.**
func TestDieVorgabennahtErreichtDieAnwendung(t *testing.T) {
	db, zuhause := frischerBestandMitZuhause(t)
	naht := vorgaben{db: db}

	mitSpielpaket(t, zuhause)

	zeilen, fehler := naht.Vorgaben(probekontext(t), "1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(zeilen) != 5 {
		t.Fatalf("es sind %d Arten: %+v", len(zeilen), zeilen)
	}

	// **Jede Zeile trägt Listen und kein `nil`** (§3): über die JSON-Naht
	// würde daraus `null`, und ein Modell griffe in etwas, das es nicht gibt.
	for _, zeile := range zeilen {
		if zeile.Werte == nil || zeile.Moeglich == nil || zeile.Verwaist == nil {
			t.Errorf("%s trägt eine nil-Liste: %+v", zeile.Art, zeile)
		}
	}

	// **Mehrwertig kommt durch** — sonst sähe ein Modell die Präferenzdecks
	// als einwertig und schriebe einen Wert, wo mehrere gehen.
	for _, zeile := range zeilen {
		if zeile.Art == "deck" && !zeile.Mehrwertig {
			t.Errorf("`deck` gilt als einwertig: %+v", zeile)
		}
	}
}

// **Das Setzen gibt den ganzen Stand zurück.**
func TestDieVorgabennahtSetztUndGibtDenStand(t *testing.T) {
	naht := vorgaben{db: frischerBestand(t)}

	// **Ein leeres `werte` räumt ab** — der einzige Fall, der ohne Angebot
	// durchgeht, und dieser Bestand hat keines.
	gesetzt, fehler := naht.VorgabeSetzen(probekontext(t), "1", "deck", []string{})
	if fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	if len(gesetzt.Vorgaben) != 5 || gesetzt.Nachladefehler != nil {
		t.Errorf("zurück kamen %d Arten (%+v)", len(gesetzt.Vorgaben), gesetzt.Nachladefehler)
	}
}

// **Und beide Wege reichen ihr Nein durch.**
//
// **Ohne diese Fälle bliebe die Hälfte jeder Naht unbetreten** — und eine
// Naht, die nur den Erfolgsweg kennt, meldet einen Fehlschlag als Erfolg (§10).
func TestDieVorgabennahtReichtIhrNeinDurch(t *testing.T) {
	naht := vorgaben{db: frischerBestand(t)}

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
