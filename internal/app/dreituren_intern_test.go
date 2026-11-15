//go:build fremdpakete

// **Die drei Türen** — von Hand, über den Scryfall-Abgleich, und was dabei
// nicht durchpasst.
//
// Eigene Datei, weil `dreiquellen_intern_test.go` mit ihnen über 400 Zeilen
// ginge (§2). Der Schnitt folgt der Frage: dort steht, **was verglichen
// wird**, hier, **wie die Karten entstehen**.
package app

import (
	"context"
	"image"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// ausDenselbenWerten legt eine Karte an und tippt jeden Wert hinein.
func ausDenselbenWerten(
	t *testing.T, anwendung *App, setID string, werte map[string]string,
) string {
	t.Helper()

	angelegt := anwendung.KarteAnlegen(context.Background(), setID)
	if angelegt.Fehler != nil {
		t.Fatalf("KarteAnlegen: %+v", angelegt.Fehler)
	}

	for _, feld := range setzeAlle(t, anwendung, setID, angelegt.KarteID, werte) {
		t.Logf("von Hand nicht setzbar: %s", feld)
	}

	return angelegt.KarteID
}

// setzeAlle schreibt jeden Wert über [App.FeldSetzen].
//
// **In fester Reihenfolge.** Eine Map läuft in Go zufällig; zwei Läufe
// schrieben dieselben Werte in verschiedener Folge, und ein Unterschied wäre
// dann nicht wiederholbar.
func setzeAlle(
	t *testing.T, anwendung *App, setID, karteID string, werte map[string]string,
) (abgelehnt []string) {
	t.Helper()

	felder := make([]string, 0, len(werte))
	for feld := range werte {
		felder = append(felder, feld)
	}

	sort.Strings(felder)

	for _, feld := range felder {
		antwort := anwendung.FeldSetzen(
			context.Background(), setID, karteID, feld, werte[feld])
		if antwort.Fehler != nil {
			// **Ein abgelehntes Feld ist kein Fehlschlag.** `time created`
			// etwa rechnet die Vorlage aus und speichert es nicht — von Hand
			// ist es gar nicht zu setzen, und es steht auf allen drei Karten
			// gleich, **weil** die Vorlage es rechnet.
			//
			// **Gesagt wird es trotzdem**: eine Handkarte, der die Hälfte
			// fehlt, verglichen sich prächtig (§10, keine stillen
			// Deckelungen).
			abgelehnt = append(abgelehnt, feld+" ("+antwort.Fehler.Meldung+")")
		}
	}

	return abgelehnt
}

// ueberDenAbgleich legt eine Karte über die Scryfall-Naht an.
//
// **Gegen eine Gegenstelle im Prozess** (§3): die Adresse wird umgelenkt und
// `HOME` bleibt, wie es ist — die Vorlagen und Schriften stehen darunter.
// Was der Dienst schickt, ist hier gerade **das, was in der Paketkarte
// steht**, damit der Vergleich die Naht prüft und nicht Scryfalls Inhalte.
func ueberDenAbgleich(
	t *testing.T, anwendung *App, setID string, werte map[string]string,
) string {
	t.Helper()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, err := schreiben.Write([]byte(scryfallkarteAus(werte))); err != nil {
				t.Errorf("die Gegenstelle konnte nicht schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	vorher := MitScryfalladresse(gegenstelle.URL)
	t.Cleanup(func() { MitScryfalladresse(vorher) })

	// **Ohne Bildfeld**: das Bild käme von einem zweiten Aufruf ins Netz, und
	// die Illustration ist für diesen Vergleich ohnehin nicht die Frage.
	ergebnis := anwendung.ScryfallKarteAnlegen(
		context.Background(), setID, werte["name"], "", "")
	if ergebnis.Fehler != nil {
		t.Fatalf("ScryfallKarteAnlegen: %+v", ergebnis.Fehler)
	}

	return ergebnis.KarteID
}

// scryfallkarteAus baut die Antwort, die zu diesen Feldwerten führt.
func scryfallkarteAus(werte map[string]string) string {
	return `{
		"object": "card",
		"id": "00000000-0000-4000-8000-000000000007",
		"name": ` + alsJSONtext(werte["name"]) + `,
		"layout": "normal",
		"mana_cost": ` + alsJSONtext(werte["casting cost"]) + `,
		"type_line": ` + alsJSONtext(typzeileAus(werte)) + `,
		"oracle_text": ` + alsJSONtext(werte["rule text"]) + `,
		"rarity": "rare",
		"scryfall_uri": "https://beispiel.invalid/karte/7"
	}`
}

// alsJSONtext setzt eine Zeichenkette in JSON-Anführungszeichen.
//
// **Von Hand und nicht über `encoding/json`**: die Antwort ist eine Vorlage
// mit Löchern, kein Verbund — und ein Verbund dafür wäre ein zweites Modell
// der Scryfall-Antwort neben dem in `internal/scryfall`.
func alsJSONtext(wert string) string {
	ersetzt := strings.NewReplacer(
		`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`, "\t", `\t`,
	).Replace(wert)

	return `"` + ersetzt + `"`
}

// gezeichnet rastert eine Karte in Bildschirmauflösung.
func gezeichnet(
	t *testing.T, anwendung *App, rasterer *render.Rasterer, karteID string,
) *image.NRGBA {
	t.Helper()

	plan := planZurKarte(t, anwendung, karteID)

	bild, _, err := rasterer.Zeichne(context.Background(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	return bild
}
