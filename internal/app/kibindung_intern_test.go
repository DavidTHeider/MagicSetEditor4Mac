package app

// **Die KI-Bindungen am Fenster** — Werkstück A7 von Phase 5.3
// (#1446).
//
// **Eine eigene Datei**, weil `kifehlerwege_intern_test.go` mit ihr über
// die Zeilengrenze aus §2 ginge.

import "testing"

// **Die Bindung am Fenster reicht die Kennung als Zahl weiter.**
//
// **Und weist ab, was keine ist** — §8 verlangt den Fehler am verursachenden
// Feld, und bei zwei Kennungen in einem Aufruf wäre „id" keiner.
func TestDieVorschlagsbindungPruefeDieKennung(t *testing.T) {
	t.Parallel()

	bindungen := &Bindungen{anwendung: NeueApp("")}

	einBezug := KIVorschlagsbezug{
		Anbieter: "ollama", Alias: "gemma4:8b", Auftrag: "x",
	}

	if antwort := bindungen.KIVorschlagen(t.Context(), "vier", einBezug); antwort.Fehler == nil {
		t.Fatal("die Kennung »vier« kam durch")
	}

	// **Eine gültige Kennung kommt an der Anwendung an** — und scheitert
	// dort an etwas anderem: ohne Bestand gibt es keine Laufzeit. Das ist
	// die Gegenprobe: ohne sie sagte der Lauf nur, dass irgendetwas
	// abgewiesen wird.
	antwort := bindungen.KIVorschlagen(t.Context(), "4", einBezug)

	if antwort.Fehler == nil {
		t.Fatal("ohne Bestand kam ein Vorschlag")
	}

	if antwort.Fehler.Feld == "karteId" {
		t.Errorf("die gültige Kennung wurde abgewiesen: %+v", antwort.Fehler)
	}

	if uebernahme := bindungen.KIVorschlagUebernehmen(t.Context(),
		"ki-4-1", []string{"name"}); uebernahme.Fehler == nil {
		t.Fatal("ohne Bestand ließ sich übernehmen")
	}
}
