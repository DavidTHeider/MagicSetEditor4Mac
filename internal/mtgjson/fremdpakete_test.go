//go:build fremdpakete

// Der Lauf über **bezogenes** MTG JSON.
//
// # Warum das hinter einem Build-Tag steht
//
// `testdaten/` enthält keine fremden Daten (§11,
// [0005](../../docs/entscheidungen/0005-keine-fremden-vorlagen.md)):
// MTG-JSON-Dateien sind Kartentexte und Namen von Wizards of the Coast. Unsere
// Beispieldateien sind **selbst geschrieben** — sie prüfen, dass der Leser tut,
// was wir für richtig halten, nicht, dass er echte Dateien liest.
//
// Dieser Lauf schließt die Lücke, ohne sie ins Repository zu holen. Er läuft
// nicht in der Pipeline, zählt nicht zur Abdeckung und wird von Hand angestoßen:
//
//	MSE_MTGJSON=~/Downloads/AllPrintings.json \
//	  go test -tags fremdpakete ./internal/mtgjson/ -run Fremd -v
//
// **Jeder Test hier heißt `TestFremd…`, und das ist kein Geschmack.** Der erste
// Anlauf nannte sie nach ihrer Sache — `TestDieSchluesseltabelle…`,
// `TestDerRundlauf…` — und der dokumentierte Befehl mit `-run Fremd` filterte
// sie damit **weg**. Übrig blieb einer von dreien, und der Lauf sah grün aus.
// Der Zweitblick über Phase 9 hat es gefunden.
//
// `MSE_MTGJSON` darf eine Datei **oder** ein Verzeichnis sein — beide Gestalten
// sind zugesagt, und **beide gehören geprüft**: ein Verzeichnis mit einzelnen
// Sets ist der Fall, den `magic-set-generator` ebenso nimmt.
//
// **Vor jeder Freigabe** (`docs/RELEASE.md`): was hier rot wird, ist ein
// Lesefehler an echten Daten.
package mtgjson_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

func TestFremdesMTGJSONWirdGelesen(t *testing.T) {
	ort := os.Getenv("MSE_MTGJSON")
	if ort == "" {
		t.Skip("MSE_MTGJSON ist nicht gesetzt — keine bezogene MTG-JSON-Quelle")
	}

	quelle, err := mtgjson.Lies(ort, mtgjson.ZuordnungMagic)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(quelle.Sets) == 0 {
		t.Fatal("kein einziges Set")
	}

	karten := 0
	for _, set := range quelle.Sets {
		karten += len(set.Karten)
	}

	if karten == 0 {
		t.Fatal("keine einzige Karte")
	}

	t.Logf("%d Sets, %d Karten, %d Befunde", len(quelle.Sets), karten, len(quelle.Befunde))

	// **Die Befunde werden aufgezählt, nicht gezählt.**
	//
	// Eine Zahl sagt „irgendetwas kennt die Zuordnung nicht"; die Namen sagen,
	// **was** — und genau die gehören in die Zuordnung, wenn sie sich häufen.
	// Sie sind kein Fehlschlag: ein unbekanntes Feld wird mitgeführt (§8).
	unbekannt := map[string]int{}
	for _, befund := range quelle.Befunde {
		unbekannt[befund.Ort]++
	}

	if len(unbekannt) > 0 {
		t.Logf("%d verschiedene Stellen mit unbekannten Feldern", len(unbekannt))
	}

	// Die erste Karte muss wenigstens einen Namen tragen — sonst hat die
	// Zuordnung gar nicht gegriffen, und der Lauf wäre grün ohne Aussage.
	erste := quelle.Sets[0].Karten[0]

	benannt := false

	for _, wert := range erste.Werte {
		if wert.Field == "name" && wert.Text != "" {
			benannt = true
		}
	}

	if !benannt {
		t.Errorf("die erste Karte trägt keinen Namen: %+v", erste.Werte)
	}
}

// **Die Tabelle der bekannten Schlüssel wird gegen die bezogene Datei
// gehalten.**
//
// Das ist der Beleg, dass sie keine Erfindung ist. Ein maschinenlesbares Schema
// gibt es nicht (nachgesehen am 2026-08-15: weder mtgjson.com noch
// `mtgjson/mtgjson`), und eine Tabelle, die nur an unserer eigenen Ausgabe
// geprüft wird, prüft unsere Vorstellung vom Format — genau das, was #142
// ausschließt.
func TestFremdeDatenBestaetigenDieSchluesseltabelle(t *testing.T) {
	ort := os.Getenv("MSE_MTGJSON")
	if ort == "" {
		t.Skip("MSE_MTGJSON ist nicht gesetzt — keine bezogene MTG-JSON-Quelle")
	}

	arten := mtgjson.Schluesselarten()
	gesehen := map[string]bool{}

	// **Datei oder Verzeichnis** — beide Gestalten sind zugesagt, und beide
	// gehören geprüft. Der erste Anlauf las mit `os.ReadFile` und scheiterte an
	// einem Verzeichnis; der Zweitblick über Phase 9 hat es angestrichen.
	for _, datei := range jsonDateien(t, ort) {
		inhalt, err := os.ReadFile(datei) // #nosec G304 -- der Pfad kommt aus der Umgebung dieses Laufs
		if err != nil {
			t.Fatalf("lesen: %v", err)
		}

		// Erstens: die bezogene Datei besteht **denselben** Prüfer wie unsere
		// Ausgabe. Wäre er nach unseren Vorstellungen gebaut, fiele sie durch.
		if befunde := mtgjson.Pruefe(inhalt, datei); len(befunde) > 0 {
			t.Errorf("%s fällt durch den eigenen Prüfer:\n%v",
				datei, befunde[:min(5, len(befunde))])
		}

		// Zweitens: jeder Schlüssel der Tabelle, den die Datei führt, hat dort
		// die Art, die die Tabelle behauptet.
		for nummer, karte := range kartenAus(t, inhalt, datei) {
			for feld, roh := range karte {
				art, bekannt := arten[feld]
				if !bekannt {
					continue
				}

				gesehen[feld] = true

				if !mtgjson.PasstZurArt(roh, art) {
					t.Errorf("%s, Karte %d, %q: die Tabelle sagt %q, die Datei sagt %s",
						datei, nummer+1, feld, art, roh)
				}
			}
		}
	}

	t.Logf("%d von %d Tabelleneinträgen an echten Daten gesehen", len(gesehen), len(arten))
}

// jsonDateien gibt die JSON-Dateien eines Ortes — er darf eine sein.
func jsonDateien(t *testing.T, ort string) []string {
	t.Helper()

	angaben, err := os.Stat(ort) //nolint:gosec // ein Ort, den der Lauf selbst genannt hat (§3)
	if err != nil {
		t.Fatalf("%s: %v", ort, err)
	}

	if !angaben.IsDir() {
		return []string{ort}
	}

	eintraege, err := os.ReadDir(ort)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	var dateien []string

	for _, eintrag := range eintraege {
		if !eintrag.IsDir() && strings.EqualFold(filepath.Ext(eintrag.Name()), ".json") {
			dateien = append(dateien, filepath.Join(ort, eintrag.Name()))
		}
	}

	if len(dateien) == 0 {
		t.Fatalf("%s führt keine JSON-Datei", ort)
	}

	return dateien
}

// kartenAus holt die rohen Karten einer Datei — einzeln oder AllSets.
func kartenAus(t *testing.T, inhalt []byte, datei string) []map[string]json.RawMessage {
	t.Helper()

	var umschlag struct {
		Daten json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(inhalt, &umschlag); err != nil {
		t.Fatalf("%s ist kein JSON: %v", datei, err)
	}

	kern := umschlag.Daten
	if kern == nil {
		kern = inhalt
	}

	var eines struct {
		Karten []map[string]json.RawMessage `json:"cards"`
	}

	if err := json.Unmarshal(kern, &eines); err == nil && eines.Karten != nil {
		return eines.Karten
	}

	var viele map[string]struct {
		Karten []map[string]json.RawMessage `json:"cards"`
	}

	if err := json.Unmarshal(kern, &viele); err != nil {
		t.Fatalf("%s führt weder ein Set noch eine Abbildung von Sets: %v", datei, err)
	}

	var alle []map[string]json.RawMessage
	for _, set := range viele {
		alle = append(alle, set.Karten...)
	}

	return alle
}

// **Der Rundlauf über bezogene Daten.**
//
// Er ist der einzige, der ihn an einer Datei fährt, die niemand für ihn
// geschrieben hat.
func TestFremdeDatenUeberstehenDenRundlauf(t *testing.T) {
	ort := os.Getenv("MSE_MTGJSON")
	if ort == "" {
		t.Skip("MSE_MTGJSON ist nicht gesetzt — keine bezogene MTG-JSON-Quelle")
	}

	erste, err := mtgjson.Lies(ort, mtgjson.ZuordnungMagic)
	if err != nil {
		t.Fatalf("erstes Lesen: %v", err)
	}

	// **Je Set ein Rundlauf.** Der Schreiber gibt ein Set je Datei aus; der
	// erste Anlauf schrieb nur das erste und verglich danach alle — und war an
	// einem Verzeichnis mit zwei Sets rot, ohne dass etwas kaputt war.
	for _, gelesen := range erste.Sets {
		set := &domain.Set{
			Name:   gelesen.Name,
			Game:   "magic",
			Values: []domain.Value{{Field: "code", Text: gelesen.Code}},
		}

		for _, karte := range gelesen.Karten {
			werte := append(append([]domain.Value{}, karte.Werte...), karte.Rest...)
			set.Cards = append(set.Cards, &domain.Card{Values: werte})
		}

		geschrieben, verluste := mtgjson.ZuordnungMagic.Schreibe(set)

		t.Logf("%s: %d Karten, %d benannte Verluste", gelesen.Code, len(set.Cards), len(verluste))

		if befunde := mtgjson.Pruefe(geschrieben, gelesen.Code); len(befunde) > 0 {
			t.Errorf("%s: die Ausgabe aus echten Daten ist nicht gültig:\n%v",
				gelesen.Code, befunde[:min(5, len(befunde))])
		}

		zweite, err := mtgjson.LiesInhalt(geschrieben, gelesen.Code, mtgjson.ZuordnungMagic)
		if err != nil {
			t.Fatalf("%s: zweites Lesen: %v", gelesen.Code, err)
		}

		for k, karteVorher := range gelesen.Karten {
			werteVorher := werteVon(karteVorher)
			werteNachher := werteVon(zweite.Sets[0].Karten[k])

			for feld, wert := range werteVorher {
				if werteNachher[feld] != wert {
					t.Errorf("%s, Karte %d, Feld %q: %q, erwartet %q",
						gelesen.Code, k+1, feld, werteNachher[feld], wert)
				}
			}
		}
	}
}

// werteVon legt die Werte einer Karte als Abbildung aus.
func werteVon(karte *mtgjson.Karte) map[string]string {
	werte := map[string]string{}
	for _, wert := range append(append([]domain.Value{}, karte.Werte...), karte.Rest...) {
		werte[wert.Field] = wert.Text
	}

	return werte
}
