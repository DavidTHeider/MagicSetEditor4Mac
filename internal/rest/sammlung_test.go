// **Die Sammlung über die Rückschleife** — Werkstück 2 und 3 von Phase 3.4
// (#1165).
//
// # Was hier steht und nicht in der Endpunktliste
//
// Jene prüft die **vier Klassen**: dass jeder Weg antwortet, seinen Fehler
// durchreicht, seine Kennung und seinen Rumpf prüft. Hier stehen die Zusagen
// **dieser** Fähigkeit: dass eine Kennung aus dem Weg gilt und nicht die aus
// dem Rumpf, und dass ein kaputter Rumpf abgewiesen wird, bevor jemand
// schreibt.
package rest_test

import (
	"net/http"
	"strings"
	"testing"
)

// **Die Kennung kommt aus dem Weg, nicht aus dem Rumpf.**
//
// # Warum das eine eigene Zusage ist
//
// Eine Kennung im Rumpf, die von der im Weg abweicht, wäre zwei Aussagen über
// dasselbe — und der Dienst müsste entscheiden, welche gilt. Er entscheidet
// sich für den Weg; ohne diesen Lauf wäre das eine Behauptung.
func TestDieKennungKommtAusDemWeg(t *testing.T) {
	t.Parallel()

	dienst := neuerDienst(t)

	antwort := ruf(t, dienst, endpunkt{
		methode: http.MethodPut, weg: "/sammlung/7",
		rumpf:  `{"id":99,"printing":"abc","menge":1}`,
		erfolg: http.StatusOK,
	}, gueltig)

	if !strings.Contains(antwort.Body.String(), `"id":7`) {
		t.Errorf("die Antwort ist %s", antwort.Body.String())
	}
}

// **Eine Kennung, die keine Zahl ist, wird abgewiesen — vor dem Schreiben.**
func TestEineKennungOhneZahlWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	dienst := neuerDienst(t)

	for name, ziel := range map[string]endpunkt{
		"ändern": {
			methode: http.MethodPut, weg: "/sammlung/keinezahl",
			rumpf: `{"printing":"abc","menge":1}`, erfolg: http.StatusOK,
		},
		"löschen": {
			methode: http.MethodDelete, weg: "/sammlung/keinezahl",
			erfolg: http.StatusOK,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, dienst, ziel, gueltig)
			pruefeAbweisung(t, antwort, http.StatusBadRequest, "validierung")
		})
	}
}

// **Ein kaputter Rumpf kommt gar nicht erst bis zur Naht.**
func TestEinKaputterRumpfDerSammlungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	dienst := neuerDienst(t)

	for name, ziel := range map[string]endpunkt{
		"anlegen": {
			methode: http.MethodPost, weg: "/sammlung",
			rumpf: `{"menge":`, erfolg: http.StatusCreated,
		},
		"ändern": {
			methode: http.MethodPut, weg: "/sammlung/1",
			rumpf: `{"menge":`, erfolg: http.StatusOK,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, dienst, ziel, gueltig)
			pruefeAbweisung(t, antwort, http.StatusBadRequest, "validierung")
		})
	}
}

// **Die Preise tragen ihre Angaben — und die Übersicht ihre Zahl.**
func TestDiePreiswegeTragenIhreAngaben(t *testing.T) {
	t.Parallel()

	dienst := neuerDienst(t)

	for name, fall := range map[string]struct {
		ziel  endpunkt
		nennt string
	}{
		"Preise": {
			endpunkt{
				methode: http.MethodGet, weg: "/drucke/probe-druck/preise",
				erfolg: http.StatusOK,
			}, "alterStunden",
		},
		"Verlauf": {
			endpunkt{
				methode: http.MethodGet,
				weg: "/drucke/probe-druck/preisverlauf?quelle=scryfall" +
					"&markt=paper&art=retail&waehrung=EUR&finish=nonfoil",
				erfolg: http.StatusOK,
			}, "scryfall",
		},
		"Holen": {
			endpunkt{
				methode: http.MethodPost, weg: "/drucke/probe-druck/preise",
				erfolg: http.StatusOK,
			}, "befunde",
		},
		"Portfolio": {
			endpunkt{methode: http.MethodGet, weg: "/portfolio", erfolg: http.StatusOK},
			"ohnePreis",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, dienst, fall.ziel, gueltig)

			if !strings.Contains(antwort.Body.String(), fall.nennt) {
				t.Errorf("die Antwort nennt %q nicht: %s", fall.nennt, antwort.Body.String())
			}
		})
	}
}

// **Ein Löschen ohne Fassung wird abgewiesen.**
//
// # Warum das eine eigene Zusage ist
//
// Ein `DELETE` hat keinen Rumpf; die Fassung steht deshalb in der **Abfrage**.
// Fehlt sie, wäre die naheliegende Antwort, ohne sie zu löschen — und genau
// das nähme eine fremde Änderung mit, die der Rufende nie gesehen hat
// (Befund 10 des Nachlaufs über Phase 3.4).
//
// **Die Meldung sagt, wo sie hingehört**, denn ein Aufrufer, der sie im Weg
// sucht, findet sie nie.
func TestEinLoeschenOhneFassungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, weg := range map[string]string{
		"gar keine":  "/sammlung/1",
		"leer":       "/sammlung/1?fassung=",
		"keine Zahl": "/sammlung/1?fassung=neulich",
		"null":       "/sammlung/1?fassung=0",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), endpunkt{
				methode: http.MethodDelete, weg: weg,
			}, gueltig)

			if antwort.Code != http.StatusBadRequest {
				t.Errorf("Status %d, erwartet war 400: %s",
					antwort.Code, antwort.Body.String())
			}

			if !strings.Contains(antwort.Body.String(), "fassung") {
				t.Errorf("die Meldung nennt das Feld nicht: %s", antwort.Body.String())
			}
		})
	}
}
