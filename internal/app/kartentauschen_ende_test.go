package app_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

// **Die neue Ordnung steht in der geschriebenen Datei**
// (#628).
//
// Der Weg durch alle Schichten: tauschen, speichern, nachsehen. Ein Test gegen
// die Datenbank sagte nur, dass `cards.reihenfolge` sich geändert hat; ob die
// Karten **in der Datei** die Plätze tauschen, entscheidet der Schreiber — und
// der ordnet die `card:`-Blöcke über genau diese Spalte zu.
func TestEineNeueOrdnungStehtInDerGeschriebenenDatei(t *testing.T) {
	anwendung, setID := geoeffnetesProbeset(t)

	karten := anwendung.Karten(t.Context(), setID)
	if karten.Fehler != nil {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	if antwort := anwendung.KartenTauschen(t.Context(), setID,
		karten.Karten[0].ID, karten.Karten[1].ID); antwort.Fehler != nil {
		t.Fatalf("tauschen: %+v", antwort.Fehler)
	}

	geschrieben := string(gespeichertesSet(t, anwendung, setID))

	// Aus „Feuerball, Eisball, Doppelt" wird „Eisball, Feuerball, Doppelt".
	if strings.Index(geschrieben, "Eisball") > strings.Index(geschrieben, "Feuerball") {
		t.Errorf("die Ordnung ist unverändert:\n%s", geschrieben)
	}

	// **Und alles andere ist unangetastet** — ein Tausch ist keine
	// Sammeländerung.
	for _, muss := range []string{"Doppelt", "unverstandener block:", "set info:"} {
		if !strings.Contains(geschrieben, muss) {
			t.Errorf("%q fehlt:\n%s", muss, geschrieben)
		}
	}
}

// **Und ⌘Z stellt die alte Ordnung byteweise wieder her**
// (#48).
//
// Das ist die schärfste Probe: die Stellen sind Zahlen im Bestand, die Ordnung
// ist die Folge der Blöcke in der Datei. Ein Tausch, der nur die Zahlen
// zurückdreht, aber die Werte an den falschen Blöcken lässt, fiele hier auf.
func TestEinZurueckgenommenerTauschHinterlaesstKeineSpur(t *testing.T) {
	anwendung, setID := geoeffnetesProbeset(t)

	karten := anwendung.Karten(t.Context(), setID)
	if karten.Fehler != nil {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	if antwort := anwendung.KartenTauschen(t.Context(), setID,
		karten.Karten[1].ID, karten.Karten[2].ID); antwort.Fehler != nil {
		t.Fatalf("tauschen: %+v", antwort.Fehler)
	}

	zurueck := anwendung.Zuruecknehmen(t.Context(), setID, "")
	if zurueck.Fehler != nil || !zurueck.Bewegt {
		t.Fatalf("⌘Z: %+v, bewegt=%v", zurueck.Fehler, zurueck.Bewegt)
	}

	vorher := lies(t, filepath.Join(probesetpfad(t), "set"))
	nachher := gespeichertesSet(t, anwendung, setID)

	if !bytes.Equal(nachher, vorher) {
		t.Errorf("die Datei ist eine andere:\n%s", unterschied(vorher, nachher))
	}
}
