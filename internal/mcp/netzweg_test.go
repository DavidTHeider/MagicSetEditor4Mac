// **Das Werkzeug über den Netzweg** — Werkstück N5 von Phase 5.4
// (#1571).
package mcp_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// netzwegprobe ist eine Naht, die mitschreibt.
type netzwegprobe struct {
	gefragt bool
	fehler  error
}

func (n *netzwegprobe) Netzstand(context.Context) (mcp.Netzstandangaben, error) {
	n.gefragt = true

	if n.fehler != nil {
		return mcp.Netzstandangaben{}, n.fehler
	}

	return mcp.Netzstandangaben{
		ProxyHTTPS: mcp.Netzfeld{
			Wert: "http://proxy.firma:3128", Herkunft: "umgebung",
		},
		FristVerbindung: 15, FristKopf: 30, FristStillstand: 60,
		ProxyBenutzt: true,
	}, nil
}

// rufeNetzweg sucht das Werkzeug und ruft es.
func rufeNetzweg(t *testing.T, naht mcp.Netzweg) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Netzweg: naht}) {
		if werkzeug.Name == "lesen.netzweg" {
			if werkzeug.Art != mcp.Lesend || !werkzeug.Idempotent {
				t.Errorf("lesen.netzweg ist als %v geführt", werkzeug.Art)
			}

			return werkzeug.Ruf(t.Context(), nil)
		}
	}

	t.Fatal("lesen.netzweg steht nicht in der Liste")

	return "", nil
}

// **Der Stand kommt durch — samt Herkunft.**
func TestDasNetzwegwerkzeugNenntDieHerkunft(t *testing.T) {
	naht := &netzwegprobe{}

	aus, fehler := rufeNetzweg(t, naht)
	if fehler != nil {
		t.Fatalf("lesen.netzweg: %v", fehler)
	}

	if !naht.gefragt {
		t.Error("die Naht wurde nicht gerufen")
	}

	// **Die Herkunft ist die halbe Auskunft**: ein Wert aus der Umgebung
	// lässt sich über `schreiben.einstellung` nicht ändern.
	for _, erwartet := range []string{
		`"wert": "http://proxy.firma:3128"`, `"herkunft": "umgebung"`,
		`"fristKopf": 30`, `"proxyBenutzt": true`,
	} {
		if !strings.Contains(aus, erwartet) {
			t.Errorf("%s fehlt in %s", erwartet, aus)
		}
	}
}

// **Ein Fehler der Naht kommt an.**
func TestDasNetzwegwerkzeugReichtSeinenFehlerDurch(t *testing.T) {
	if _, fehler := rufeNetzweg(t,
		&netzwegprobe{fehler: errors.New("der Bestand schweigt")}); fehler == nil {
		t.Error("der Fehler wurde verschluckt")
	}
}

// **Ohne Naht gibt es das Werkzeug nicht.**
func TestOhneNetzwegnahtGibtEsDasWerkzeugNicht(t *testing.T) {
	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if werkzeug.Name == "lesen.netzweg" {
			t.Error("lesen.netzweg steht da, obwohl keine Naht eingerichtet ist")
		}
	}
}
