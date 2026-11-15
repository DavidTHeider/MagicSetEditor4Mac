package main

// **Die Verdrahtung der Freigaben im Dienst** — Werkstück 11 von Phase 4.12
// (#1351).

import (
	"testing"
)

// **Der ganze Weg über die Naht.**
func TestDieFreigabennahtDesDienstesErreichtDieAnwendung(t *testing.T) {
	naht := fremdfreigaben{db: frischerTokenbestand(t)}

	stand, fehler := naht.Freigaben(probekontext(t), "1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if stand.EigeneRolle != "verwalter" {
		t.Errorf("EigeneRolle = %q", stand.EigeneRolle)
	}

	if stand.Freigaben == nil {
		t.Error("die Liste ist nil")
	}
}

// **Und jeder Weg reicht sein Nein durch.**
func TestDieFreigabennahtDesDienstesReichtIhrNeinDurch(t *testing.T) {
	naht := fremdfreigaben{db: frischerTokenbestand(t)}

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
