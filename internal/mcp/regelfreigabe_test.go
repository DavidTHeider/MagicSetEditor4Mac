// **Das Werkzeug über die Regelfreigabe** — Werkstück G4 von Phase 5.5
// (#1461).
package mcp_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// freigabeprobe ist eine Naht, die mitschreibt.
type freigabeprobe struct {
	gesehen struct {
		spiel   string
		strenge string
		sitze   int

		// **Modus und Abwahl kamen mit PLAY-02** (#1710) — geprüft wird
		// das Argument und nicht der bloße Ruf (§10).
		modus      string
		abgewaehlt []string
	}

	fehler error
}

func (r *freigabeprobe) Regelfreigabe(
	_ context.Context, spiel, strenge string, sitze int,
	modus string, abgewaehlt []string,
) (mcp.Regelfreigabeangaben, error) {
	r.gesehen.spiel, r.gesehen.strenge, r.gesehen.sitze = spiel, strenge, sitze
	r.gesehen.modus, r.gesehen.abgewaehlt = modus, abgewaehlt

	if r.fehler != nil {
		return mcp.Regelfreigabeangaben{}, r.fehler
	}

	return mcp.Regelfreigabeangaben{
		Erlaubt: true, Regelwerk: "sandkasten", Fassung: 1,
		Wirksam: "2026-08-19", Hoechste: "begleitet",
		Matrix: []mcp.Regelbereichsangaben{{
			Bereich: "zonen", Grad: "automatisch",
			Quelle:  "CR 400 (Fassung 2026-08-19)",
			Grenzen: []string{"keine Zusatzzonen"},
		}},
		Luecken:     []string{"combat"},
		Hindernisse: []string{},
	}, nil
}

// rufeRegelfreigabe sucht das Werkzeug und ruft es.
func rufeRegelfreigabe(
	t *testing.T, naht mcp.Regelquellenleser, argumente string,
) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Regelfreigabe: naht}) {
		if werkzeug.Name != "lesen.regelfreigabe" {
			continue
		}

		// **Lesend und idempotent** — sie rechnet und ändert nichts.
		if werkzeug.Art != mcp.Lesend || !werkzeug.Idempotent {
			t.Errorf("lesen.regelfreigabe ist als %v geführt", werkzeug.Art)
		}

		return werkzeug.Ruf(t.Context(), []byte(argumente))
	}

	t.Fatal("lesen.regelfreigabe steht nicht in der Liste")

	return "", nil
}

// **Die Angaben kommen an, und die Antwort trägt Quelle und Grenzen.**
func TestDasRegelfreigabewerkzeugNenntQuelleUndGrenzen(t *testing.T) {
	t.Parallel()

	naht := &freigabeprobe{}

	aus, fehler := rufeRegelfreigabe(t, naht,
		`{"spiel":"magic","strenge":"sandkasten","sitze":2}`)
	if fehler != nil {
		t.Fatalf("lesen.regelfreigabe: %v", fehler)
	}

	switch {
	case naht.gesehen.spiel != "magic":
		t.Errorf("Spiel %q", naht.gesehen.spiel)
	case naht.gesehen.strenge != "sandkasten":
		t.Errorf("Strenge %q", naht.gesehen.strenge)
	case naht.gesehen.sitze != 2:
		t.Errorf("Sitze %d", naht.gesehen.sitze)
	}

	// **Quelle und Grenze stehen in der Antwort** — ohne sie wäre die
	// Zusage eine Behauptung, und die Grenze fehlte ganz.
	for _, erwartet := range []string{
		"CR 400", "2026-08-19", "keine Zusatzzonen", "begleitet", "combat",
	} {
		if !strings.Contains(aus, erwartet) {
			t.Errorf("die Antwort nennt %q nicht: %s", erwartet, aus)
		}
	}
}

// **Was die Naht klagt, kommt zurück.**
func TestDasRegelfreigabewerkzeugReichtDenFehlerDurch(t *testing.T) {
	t.Parallel()

	naht := &freigabeprobe{fehler: errors.New("kein Regelwerk")}

	if _, fehler := rufeRegelfreigabe(t, naht,
		`{"spiel":"magic","strenge":"streng","sitze":2}`); fehler == nil {
		t.Fatal("kein Fehler")
	}
}

// **Unlesbare Angaben sind ein Fehler und kein leerer Ruf.**
//
// **Sonst fragte das Werkzeug mit lauter Nullwerten** — Spiel leer, Sitze
// null — und bekäme eine Absage, deren Grund mit der Frage nichts zu tun hat.
func TestDasRegelfreigabewerkzeugWeistUnlesbareAngabenAb(t *testing.T) {
	t.Parallel()

	naht := &freigabeprobe{}

	_, fehler := rufeRegelfreigabe(t, naht, `{"sitze":"zwei"}`)
	if fehler == nil {
		t.Fatal("kein Fehler")
	}

	if naht.gesehen.spiel != "" {
		t.Error("die Naht wurde trotzdem gefragt")
	}
}

// **Ohne Naht gibt es das Werkzeug nicht.**
//
// **Und das ist kein stiller Ausfall**: `mse-mcp` verdrahtet sie, das Fenster
// braucht sie nicht. Ein Werkzeug, das ohne Naht in der Liste stünde, fiele
// beim Ruf um.
func TestOhneNahtGibtEsDasRegelfreigabewerkzeugNicht(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if werkzeug.Name == "lesen.regelfreigabe" {
			t.Fatal("das Werkzeug steht ohne Naht in der Liste")
		}
	}
}
