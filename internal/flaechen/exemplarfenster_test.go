package flaechen_test

// **Ein Exemplar am Fenster** — C04
// (#2229),
// Zweitblick über Phase C, Befund 3: `TestEinExemplarUeberJedeFlaeche` fuhr
// Kommandozeile, MCP und Dienst, nicht das Fenster.
//
// **Die Eingabe kommt als JSON**, wie die Oberfläche sie über Wails schickt —
// aus derselben Tabelle wie an MCP und Dienst (`vollesExemplar`) —, und die
// Antwort geht als JSON zurück an denselben Leser (`ersteAusHuelle`).
//
// # Eine begründete Abweichung: `foil`
//
// **Das Fenster führt den abgekündigten Wahrheitswert nicht.** `foil` ist seit
// T2 durch `finish` ersetzt und steht an Kommandozeile, MCP und Dienst nur
// noch für Gegenüber, die die neue Fassung nicht kennen
// ([0119](../../docs/entscheidungen/0119-eine-ausfuehrung-statt-eines-wahrheitswerts.md));
// die Oberfläche schickt und liest `finish` (`frontend/src/antwortensammlung.ts`).
// Deshalb schickt dieser Fall `finish`, und `foil` wird aus `finish` gelesen —
// **nachdem** geprüft ist, dass das Fenster es wirklich nicht nennt.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// alsExemplar macht aus der Tabelle das, was Wails dem Fenster übergäbe.
func alsExemplar(t *testing.T, werte map[string]any) app.ExemplarDTO {
	t.Helper()

	// **Was die Oberfläche schickt**: `finish`, nicht `foil`.
	werte["finish"] = "foil"
	delete(werte, "foil")

	var exemplar app.ExemplarDTO
	if fehler := json.Unmarshal([]byte(alsRumpf(t, werte)), &exemplar); fehler != nil {
		t.Fatalf("die Tabelle passt nicht auf ExemplarDTO: %v", fehler)
	}

	return exemplar
}

// exemplarfallFenster ist der Fall „Fenster" in [TestEinExemplarUeberJedeFlaeche]
// — derselbe Typ wie die drei übrigen.
func exemplarfallFenster() struct {
	aufnehmen func(t *testing.T, umgebung []string)
	lesen     func(t *testing.T, umgebung []string) flaechenexemplar
	aendern   func(t *testing.T, umgebung []string, gelesen flaechenexemplar)
} {
	return struct {
		aufnehmen func(t *testing.T, umgebung []string)
		lesen     func(t *testing.T, umgebung []string) flaechenexemplar
		aendern   func(t *testing.T, umgebung []string, gelesen flaechenexemplar)
	}{
		aufnehmen: func(t *testing.T, umgebung []string) {
			t.Helper()

			antwort := fensterAus(t, umgebung).ExemplarAnlegen(t.Context(), alsExemplar(t, vollesExemplar()))
			if antwort.Fehler != nil {
				t.Fatalf("ExemplarAnlegen: %+v", antwort.Fehler)
			}
		},
		lesen: func(t *testing.T, umgebung []string) flaechenexemplar {
			t.Helper()

			antwort := fensterAus(t, umgebung).Sammlung(t.Context())
			if antwort.Fehler != nil {
				t.Fatalf("Sammlung: %+v", antwort.Fehler)
			}

			roh, fehler := json.Marshal(antwort)
			if fehler != nil {
				t.Fatalf("Sammlung serialisieren: %v", fehler)
			}

			if strings.Contains(string(roh), `"foil":`) {
				t.Fatalf("das Fenster nennt den abgekündigten Wahrheitswert doch: %s", roh)
			}

			gelesen := ersteAusHuelle(t, string(roh))
			gelesen.Foil = gelesen.Finish == "foil"

			return gelesen
		},
		aendern: func(t *testing.T, umgebung []string, gelesen flaechenexemplar) {
			t.Helper()

			werte := vollesExemplar()
			werte["id"] = gelesen.ID
			werte["fassung"] = gelesen.Fassung
			werte["menge"] = 3

			if fehler := fensterAus(t, umgebung).ExemplarAendern(t.Context(), alsExemplar(t, werte)); fehler != nil {
				t.Fatalf("ExemplarAendern: %+v", fehler)
			}
		},
	}
}
