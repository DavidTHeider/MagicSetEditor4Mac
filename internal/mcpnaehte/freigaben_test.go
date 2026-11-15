package mcpnaehte

// **Die Verdrahtung der Freigaben** — Werkstück 11 von Phase 4.12
// (#1351).

import (
	"testing"
)

// **Der ganze Weg über die Naht.**
func TestDieFreigabennahtErreichtDieAnwendung(t *testing.T) {
	naht := freigaben{db: frischerBestand(t)}

	stand, fehler := naht.Freigaben(probekontext(t), "1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	// **Der Eigner ist Verwalter — ohne Zeile im Bestand.**
	if stand.EigeneRolle != "verwalter" {
		t.Errorf("EigeneRolle = %q", stand.EigeneRolle)
	}

	if stand.Sichtbarkeit != "privat" {
		t.Errorf("Sichtbarkeit = %q", stand.Sichtbarkeit)
	}

	// **Und die Liste ist eine Liste** (§3): über die JSON-Naht würde aus
	// `nil` sonst `null`.
	if stand.Freigaben == nil {
		t.Error("die Liste ist nil")
	}
}

// **Und jeder Weg reicht sein Nein durch.**
func TestDieFreigabennahtReichtIhrNeinDurch(t *testing.T) {
	naht := freigaben{db: frischerBestand(t)}

	if _, fehler := naht.Freigaben(probekontext(t), "4711"); fehler == nil {
		t.Error("ein Profil, das es nicht gibt, kam durch")
	}

	if _, fehler := naht.FreigabeSetzen(
		probekontext(t), "1", "1", "gott"); fehler == nil {
		t.Error("eine Rolle, die es nicht gibt, kam durch")
	}

	if _, fehler := naht.FreigabeLoeschen(
		probekontext(t), "1", "4711"); fehler == nil {
		t.Error("eine Freigabe, die es nicht gibt, ließ sich zurückziehen")
	}
}
