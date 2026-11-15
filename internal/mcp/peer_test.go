package mcp

// **Das Werkzeug zur eigenen Peeridentität** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Was hier gemessen wird
//
// **Dass es lesend ist und bleibt.** Ein Modell darf fragen, wer dieses
// Programm ist; anlegen und ablösen sind Handlungen für einen Menschen, und
// eine Rotation entwertet jede Bestätigung, die ein Kontakt gegeben hat.
//
// **Und dass „keine Naht" kein Werkzeug ergibt** — sonst stünde es im Katalog
// und antwortete mit einem Sturz.

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// peerprobe gibt eine feste Identität — oder einen Fehler.
type peerprobe struct {
	lage   Peeridentitaet
	fehler error
}

func (p peerprobe) Peeridentitaet(context.Context) (Peeridentitaet, error) {
	return p.lage, p.fehler
}

// **Das Werkzeug nennt die Identität und ändert nichts.**
func TestDasPeerwerkzeugNenntDieIdentitaet(t *testing.T) {
	t.Parallel()

	werkzeuge := peerwerkzeug(peerprobe{lage: Peeridentitaet{
		Stand: "bereit", Kennung: "bprobe",
		Fingerabdruck: "1111-2222-3333-4444-5555-6666-7777-8888",
	}})

	if len(werkzeuge) != 1 {
		t.Fatalf("es sind %d Werkzeuge", len(werkzeuge))
	}

	eines := werkzeuge[0]

	if eines.Name != "lesen.peeridentitaet" || eines.Art != Lesend ||
		!eines.Idempotent {
		t.Fatalf("das Werkzeug ist %+v", eines)
	}

	antwort, fehler := eines.Ruf(t.Context(), nil)
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	for _, teil := range []string{
		`"stand": "bereit"`, `"kennung": "bprobe"`,
		`"fingerabdruck": "1111-2222-3333-4444-5555-6666-7777-8888"`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("in %q fehlt %q", antwort, teil)
		}
	}
}

// **Ein Fehler der Naht kommt durch.**
func TestDasPeerwerkzeugReichtSeinenFehlerDurch(t *testing.T) {
	t.Parallel()

	werkzeuge := peerwerkzeug(peerprobe{fehler: errors.New("kein Bestand")})

	if _, fehler := werkzeuge[0].Ruf(t.Context(), nil); fehler == nil {
		t.Fatal("der Fehler kam nicht durch")
	}
}

// **Ohne Naht gibt es kein Werkzeug.**
//
// **Die Gegenprobe** (§10): ein Werkzeug ohne Naht stünde im Katalog und
// stürzte beim ersten Ruf — ein Modell liest den Katalog als Zusage.
func TestOhnePeernahtGibtEsKeinWerkzeug(t *testing.T) {
	t.Parallel()

	if werkzeuge := peerwerkzeug(nil); werkzeuge != nil {
		t.Fatalf("ohne Naht kamen %d Werkzeuge", len(werkzeuge))
	}
}
