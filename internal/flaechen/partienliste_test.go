package flaechen_test

// **Die Liste der eigenen Partien an jeder Fläche** —
// #1600.
//
// # Wogegen hier geprüft wird
//
// **Dass es sie überhaupt an allen vier gibt** (§16) — und dass sie an jeder
// **dasselbe** sagt. Ein Wächter sieht, dass eine Fähigkeit vorhanden ist;
// dass sie dasselbe tut, sieht nur ein Lauf, der sie dreimal fährt.
//
// **Und dass eine Liste keine Sicht ist** (G1): sie nennt Sitze, Stand und
// die **Zahl** der Objekte — keinen Kartennamen. Wer sie um Objekte
// erweiterte, umginge die Projektion, statt sie zu benutzen.

import (
	"encoding/json"
	"net/http"
	"testing"
)

// listenzeile ist eine Partie, wie eine Liste sie zeigt.
type listenzeile struct {
	ID      string `json:"id"`
	Fassung string `json:"fassung"`
	Spiel   string `json:"spiel"`
	Stand   string `json:"stand"`

	Sitze []struct {
		Kennung string `json:"kennung"`
		Name    string `json:"name"`
	} `json:"sitze"`

	Objekte int `json:"objekte"`
	Zug     int `json:"zug"`
}

// **Dieselbe Liste an jeder Fläche.**
func TestDieselbePartienlisteAnJederFlaeche(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	partieID := bildAusCLI(t, mussKommandozeile(t, umgebung, "partie-anlegen",
		"magic", "anna:Anna:"+ersterDeck, "bert:Bert:"+zweiterDeck, "--json")).ID

	ausCLI := zeilenAusCLI(t, mussKommandozeile(t, umgebung, "partien", "--json"))
	ausMCP := zeilenAusFlach(t, mussMCP(t, umgebung, "lesen.partien",
		map[string]any{}))
	ausREST := zeilenAusFlach(t, mussREST(t, an, http.MethodGet, "/partien", ""))

	for name, zeilen := range map[string][]listenzeile{
		"Kommandozeile": ausCLI, "MCP": ausMCP, "REST": ausREST,
	} {
		if len(zeilen) != 1 {
			t.Fatalf("%s nennt %d Partien", name, len(zeilen))
		}

		zeile := zeilen[0]

		switch {
		case zeile.ID != partieID:
			t.Errorf("%s nennt die Partie %q", name, zeile.ID)
		case zeile.Stand != "laeuft":
			t.Errorf("%s nennt den Stand %q", name, zeile.Stand)
		case len(zeile.Sitze) != 2:
			t.Errorf("%s nennt %d Sitze", name, len(zeile.Sitze))
		case zeile.Sitze[0].Kennung != "anna" || zeile.Sitze[0].Name != "Anna":
			t.Errorf("%s nennt den ersten Sitz %+v", name, zeile.Sitze[0])
		case zeile.Objekte == 0:
			t.Errorf("%s zählt kein Objekt", name)
		}
	}

	if ausCLI[0].Objekte != ausMCP[0].Objekte ||
		ausCLI[0].Objekte != ausREST[0].Objekte {
		t.Errorf("die Flächen zählen verschieden: %d, %d, %d",
			ausCLI[0].Objekte, ausMCP[0].Objekte, ausREST[0].Objekte)
	}
}

// **Und ohne Partie gibt jede Fläche eine leere Liste** — keinen Fehler.
//
// **Weil „du hast keine“ eine Antwort ist** und kein Ausfall (§8). Ein
// Fehler schickte jemanden auf die Suche nach einer Ursache, die es nicht
// gibt.
func TestOhnePartieGibtJedeFlaecheEineLeereListe(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)

	for name, zeilen := range map[string][]listenzeile{
		"Kommandozeile": zeilenAusCLI(t, mussKommandozeile(t, umgebung,
			"partien", "--json")),
		"MCP": zeilenAusFlach(t, mussMCP(t, umgebung, "lesen.partien",
			map[string]any{})),
		"REST": zeilenAusFlach(t, mussREST(t, an, http.MethodGet, "/partien", "")),
	} {
		if len(zeilen) != 0 {
			t.Errorf("%s nennt %d Partien", name, len(zeilen))
		}
	}
}

// zeilenAusCLI liest die Liste aus der Hülle der Kommandozeile.
func zeilenAusCLI(t *testing.T, roh string) []listenzeile {
	t.Helper()

	var huelle struct {
		Partien []listenzeile `json:"partien"`
	}

	if err := json.Unmarshal([]byte(roh), &huelle); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	return huelle.Partien
}

// zeilenAusFlach liest die Liste aus einer flachen Antwort.
func zeilenAusFlach(t *testing.T, roh string) []listenzeile {
	t.Helper()

	var zeilen []listenzeile

	if err := json.Unmarshal([]byte(roh), &zeilen); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	return zeilen
}
