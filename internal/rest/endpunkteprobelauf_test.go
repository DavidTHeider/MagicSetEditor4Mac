// **Der Probelauf über den Dienst** — UXR-F7
// (#2015).
//
// Zwei Wege, die **lesen**: sie sagen, was der Druckbogen und die Exportvorlage
// ergäben, und schreiben nichts. `POST`, weil die Wahl in den Rumpf gehört —
// wie bei `/ersetzung/vorschau`. Einordnung nach §16: **erweiternd**, zwei Wege
// kommen hinzu und keiner ändert seine Bedeutung.
package rest_test

import (
	"net/http"
	"strings"
	"testing"
)

// endpunkteProbelauf sind die zwei Wege, die sagen, was eine Ausgabe ergäbe.
func endpunkteProbelauf() []endpunkt {
	return []endpunkt{
		{
			name: "Bogen durchrechnen", methode: http.MethodPost,
			weg: "/sets/1/bogen/probelauf", rumpf: `{}`, erfolg: http.StatusOK,
		},
		{
			name: "Vorlage durchrechnen", methode: http.MethodPost,
			weg: "/sets/1/vorlagen/probelauf", rumpf: `{"paket":"probe-liste.mse-export-template"}`,
			erfolg: http.StatusOK,
		},
	}
}

// **Die Antwort kennzeichnet sich als Probelauf und trägt die Zahlen.**
func TestDieAntwortDesBogenprobelaufsTraegtDieZahlen(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t), endpunkteProbelauf()[0], gueltig)
	if antwort.Code != http.StatusOK {
		t.Fatalf("%d: %s", antwort.Code, antwort.Body.String())
	}

	for _, erwartet := range []string{`"probelauf":true`, `"karten":10`, `"seiten":2`, `"jeSeite":9`, `"dpi":300`} {
		if !strings.Contains(strings.ReplaceAll(antwort.Body.String(), " ", ""), erwartet) {
			t.Errorf("%s fehlt: %s", erwartet, antwort.Body.String())
		}
	}
}

// **Die Antwort der Vorlage nennt Namen und Größe, die Befunde nie als `null`.**
func TestDieAntwortDesVorlagenprobelaufsTraegtGroesseUndBefunde(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t), endpunkteProbelauf()[1], gueltig)
	if antwort.Code != http.StatusOK {
		t.Fatalf("%d: %s", antwort.Code, antwort.Body.String())
	}

	rumpf := strings.ReplaceAll(antwort.Body.String(), " ", "")
	for _, erwartet := range []string{`"probelauf":true`, `"dateiname":"ausfuhr.txt"`, `"bytes":12`, `"befunde":[]`} {
		if !strings.Contains(rumpf, erwartet) {
			t.Errorf("%s fehlt: %s", erwartet, antwort.Body.String())
		}
	}
}

// **Ein unbekanntes Feld im Rumpf wird abgewiesen — wie bei jedem Weg.**
func TestDerProbelaufWeistEinUnbekanntesFeldAb(t *testing.T) {
	t.Parallel()

	for _, endpunkt := range endpunkteProbelauf() {
		t.Run(endpunkt.name, func(t *testing.T) {
			t.Parallel()

			unfug := endpunkt
			unfug.rumpf = `{"gibtsnicht":1}`

			if antwort := ruf(t, neuerDienst(t), unfug, gueltig); antwort.Code != http.StatusBadRequest {
				t.Errorf("%d statt 400: %s", antwort.Code, antwort.Body.String())
			}
		})
	}
}

// **Eine Set-Kennung, die keine Zahl ist, weist der Probelauf ab** — vor dem
// Rumpf und ohne die Naht zu fragen, wie jeder Weg mit `{setID}`.
func TestDerProbelaufWeistEineUnbrauchbareSetKennungAb(t *testing.T) {
	t.Parallel()

	for _, weg := range endpunkteProbelauf() {
		t.Run(weg.name, func(t *testing.T) {
			t.Parallel()

			weg.weg = strings.Replace(weg.weg, "/sets/1/", "/sets/x/", 1)
			if antwort := ruf(t, neuerDienst(t), weg, gueltig); antwort.Code != http.StatusBadRequest {
				t.Errorf("%d: %s", antwort.Code, antwort.Body.String())
			}
		})
	}
}
