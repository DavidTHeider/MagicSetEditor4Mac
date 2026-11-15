package mcp_test

// **Eine Einfuhr vorab ansehen, über MCP** — DOM-03.AK-01
// (#1698).

import (
	"errors"
	"strings"
	"testing"
)

func TestDieVorschauenStehenUeberMCP(t *testing.T) {
	t.Run("Liste", func(t *testing.T) {
		probe := &einfuhrprobe{}

		antwort := rufeRoh(t, mitEinfuhr(t, probe, "lesen.liste.vorschau"),
			`{"zeilen":[{"printing":"probe-druck","menge":2}]}`)

		if len(probe.vorgeschaut) != 1 || probe.vorgeschaut[0].Printing != "probe-druck" ||
			len(probe.eingelesen) != 0 {
			t.Errorf("vorgeschaut %+v, eingelesen %+v", probe.vorgeschaut, probe.eingelesen)
		}

		if !strings.Contains(antwort, "zusammen") {
			t.Errorf("die Antwort nennt die Art nicht: %s", antwort)
		}
	})

	t.Run("CSV", func(t *testing.T) {
		probe := &sammlungsausfuhrprobe{}

		antwort := rufeRoh(t, mitSammlungsausfuhr(t, probe, "lesen.sammlung.vorschau"),
			`{"text":"printing,menge\nprobe-druck,1\n"}`)

		if probe.vorgeschaut == "" || probe.gelesen != "" {
			t.Errorf("vorgeschaut %q, eingelesen %q", probe.vorgeschaut, probe.gelesen)
		}

		if !strings.Contains(antwort, "neu") {
			t.Errorf("die Antwort nennt die Art nicht: %s", antwort)
		}
	})
}

// **Die Zusammenführung reist mit** — und ohne sie ist sie leer, was die
// Naht als addieren liest.
func TestDieZusammenfuehrungReistUeberMCP(t *testing.T) {
	liste := &einfuhrprobe{}
	rufeRoh(t, mitEinfuhr(t, liste, "schreiben.liste.einlesen"),
		`{"zeilen":[{"printing":"probe-druck","menge":1}],"zusammenfuehrung":"auslassen"}`)

	if liste.zusammenfuehrung != "auslassen" {
		t.Errorf("die Liste bekam %q", liste.zusammenfuehrung)
	}

	csv := &sammlungsausfuhrprobe{}
	rufeRoh(t, mitSammlungsausfuhr(t, csv, "schreiben.sammlung.einlesen"),
		`{"text":"printing,menge\nprobe-druck,1\n","zusammenfuehrung":"auslassen"}`)

	if csv.zusammenfuehrung != "auslassen" {
		t.Errorf("das CSV bekam %q", csv.zusammenfuehrung)
	}
}

// **Fehler und unlesbare Angaben gehen als Fehler hinaus.**
func TestDieVorschauenReichenFehlerDurch(t *testing.T) {
	gestolpert := errors.New("die Platte klemmt")

	for _, fall := range []struct {
		name string
		ruf  func(fehler error) bool
	}{
		{"Liste, Fehler", func(fehler error) bool {
			return rufeMitFehler(t, mitEinfuhr(t, &einfuhrprobe{fehler: fehler}, "lesen.liste.vorschau"),
				`{"zeilen":[]}`)
		}},
		{"CSV, Fehler", func(fehler error) bool {
			return rufeMitFehler(t, mitSammlungsausfuhr(t, &sammlungsausfuhrprobe{fehler: fehler},
				"lesen.sammlung.vorschau"), `{"text":"x"}`)
		}},
		{"Liste, kein JSON", func(error) bool {
			return rufeMitFehler(t, mitEinfuhr(t, &einfuhrprobe{}, "lesen.liste.vorschau"), `kein JSON`)
		}},
		{"CSV, kein JSON", func(error) bool {
			return rufeMitFehler(t, mitSammlungsausfuhr(t, &sammlungsausfuhrprobe{},
				"lesen.sammlung.vorschau"), `kein JSON`)
		}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			if !fall.ruf(gestolpert) {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}
