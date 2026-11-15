package flaechen_test

// **Die Helfer, mit denen drei Flächen verglichen werden** — Werkstück G1
// von Phase 5.5
// (#1458).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan stehen die Zusagen** — dass dieselbe Partie an jeder Fläche
// gleich entsteht, gleich aussieht und ihre Bibliothek überall geheim hält.
// Hier steht, **womit** das geprüft wird: was an einer frischen Partie
// stimmen muss, wie zwei Bilder verglichen werden und wie aus der Ausgabe
// einer Fläche ein Bild wird. `partie_test.go` stand mit beidem bei 415
// Zeilen (§2).

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// pruefeFrischePartie sichert zu, was beim Anlegen entstanden ist.
//
// **Die Zahlen sind gerechnet und nicht abgelesen**: zwei Decks mit je
// [deckmenge] Stück derselben Karte geben zwei Sitze und
// `2 * deckmenge` Objekte. Wer sie aus einem grünen Lauf abschriebe, hätte
// sie nie geprüft.
func pruefeFrischePartie(t *testing.T, bild partiebild) {
	t.Helper()

	switch {
	case bild.Stand != "laeuft":
		t.Errorf("Stand %q, erwartet »laeuft«", bild.Stand)
	case bild.Wirt != "anna":
		t.Errorf("Wirt %q — ohne Angabe ist es der erste Sitz", bild.Wirt)
	case bild.Fuer != "anna":
		t.Errorf("gerechnet für %q, erwartet den Wirt", bild.Fuer)
	case bild.Spiel != "magic":
		t.Errorf("Spiel %q", bild.Spiel)
	case bild.Fassung != "1":
		t.Errorf("Fassung %q — eine frische Partie steht bei 1", bild.Fassung)
	}

	if len(bild.Sitze) != 2 {
		t.Fatalf("%d Sitze, erwartet zwei: %+v", len(bild.Sitze), bild.Sitze)
	}

	for _, sitz := range bild.Sitze {
		if sitz.Karten != deckmenge {
			t.Errorf("der Sitz %q bringt %d Karten mit, erwartet %d",
				sitz.Kennung, sitz.Karten, deckmenge)
		}
	}

	if len(bild.Objekte) != 2*deckmenge {
		t.Errorf("%d Objekte, erwartet %d — je Stück eines und nicht je Zeile",
			len(bild.Objekte), 2*deckmenge)
	}

	// **Vor dem Beginn nur der Beginn** — CR 103.8 (Fassung 2026-09-25): eine
	// frische Partie bietet dem Wirt die Starthand an und, was jederzeit
	// geht — aufgeben (CR 104.3a) und die zwei Würfe, mit denen CR 103.1 den
	// Startspieler bestimmen lässt. **Bis #1851 waren es fünfundvierzig
	// Sandkastenbefehle**: eine Partie, in der niemand anfing, spielte ohne
	// Beginn.
	//
	// **Was nach dem Beginn angeboten wird**, zählt
	// `service.TestOhnePrioritaetBleibtNurDieBuchfuehrung` — hier zählt, dass
	// **jede Fläche dieselbe Liste** gibt (§16), und welche.
	if !slices.Equal(bild.Erlaubt, []string{
		"spieler.starthand", "zug.aufgeben", "spieler.muenzwurf",
		"spieler.wuerfeln",
	}) {
		t.Errorf("erlaubt: %v, erwartet den Beginn und was jederzeit geht",
			bild.Erlaubt)
	}
}

// gleichesBild verlangt, dass zwei Flächen dieselbe Partie beschreiben.
//
// **Bis auf die Kennung**: sie ist je Partie neu und sagt über die Auskunft
// nichts aus. Alles andere wird verglichen, und zwar als Wert — Einrückung
// und Feldreihenfolge sind Gestalt (derselbe Satz wie bei [gleicheAuskunft]).
func gleichesBild(t *testing.T, was string, eine, andere partiebild) {
	t.Helper()

	eine.ID, andere.ID = "", ""

	if !reflect.DeepEqual(eine, andere) {
		t.Errorf("%s: die beiden Flächen sagen Verschiedenes:\n%+v\n%+v",
			was, eine, andere)
	}
}

// gleicheBilderZweierPartien vergleicht zwei **verschiedene** Partien aus
// denselben Decks — **ohne die Namen ihrer Objekte** (#2480): jede trägt ihr
// eigenes Geheimnis und damit eigene Sichtkennungen, und verborgene Objekte
// stehen nach ihnen geordnet. Verglichen wird, was auf dem Tisch liegt.
func gleicheBilderZweierPartien(t *testing.T, was string, eine, andere partiebild) {
	t.Helper()

	gleichesBild(t, was, ohneNamen(eine), ohneNamen(andere))
}

// ohneNamen gibt ein Bild ohne die Sichtkennungen seiner Objekte, in einer
// Reihenfolge, die nicht an ihnen hängt.
func ohneNamen(bild partiebild) partiebild {
	objekte := slices.Clone(bild.Objekte)
	for stelle := range objekte {
		objekte[stelle].Kennung = ""
	}

	slices.SortFunc(objekte, func(a, b bildobjekt) int {
		return strings.Compare(fmt.Sprint(a), fmt.Sprint(b))
	})

	bild.Objekte = objekte

	return bild
}

// verschiedeneKennungen verlangt, dass drei Partien drei Partien sind.
func verschiedeneKennungen(t *testing.T, kennungen ...string) {
	t.Helper()

	gesehen := map[string]bool{}

	for _, kennung := range kennungen {
		if kennung == "" {
			t.Fatalf("eine Partie ohne Kennung: %v", kennungen)
		}

		if gesehen[kennung] {
			t.Errorf("die Kennung %q kam zweimal: %v", kennung, kennungen)
		}

		gesehen[kennung] = true
	}
}

// bildAusCLI liest die Antwort der Kommandozeile.
func bildAusCLI(t *testing.T, roh string) partiebild {
	t.Helper()

	var huelle struct {
		Partie     partiebild     `json:"partie"`
		Erlaubt    []string       `json:"erlaubt"`
		Handlungen []bildhandlung `json:"handlungen"`
	}

	if err := json.Unmarshal([]byte(roh), &huelle); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	bild := huelle.Partie
	bild.Erlaubt = huelle.Erlaubt
	bild.Handlungen = huelle.Handlungen

	return bild
}

// bildAusFlach liest die Antwort von MCP und Dienst.
func bildAusFlach(t *testing.T, roh string) partiebild {
	t.Helper()

	var bild partiebild

	if err := json.Unmarshal([]byte(roh), &bild); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	return bild
}

// deckmenge ist, wie viele Stücke derselben Karte in jedem Deck liegen.
//
// **Mehr als eines**, denn „je Stück ein Objekt und nicht je Zeile" ist bei
// einem Stück keine Aussage. **Und mehr als sieben** (#1851): seit der
// Beginn vom Aufbau an läuft, zieht jeder Sitz eine Starthand, bevor
// gespielt wird (CR 103.5), und aus drei Karten wurde keine. Die drei über
// der Starthand lassen dem nächsten Zug etwas zu ziehen.
const deckmenge = 10

// zweiDecks legt zwei spielbare Decks an und gibt ihre Kennungen.
//
// **Über die Kommandozeile für alle drei Flächen.** Was hier entsteht, ist
// die **Vorbereitung** und nicht das Geprüfte; sie über drei Wege zu bauen
// brächte drei Fehlerquellen in einen Lauf, der eine andere Frage stellt.
func zweiDecks(t *testing.T, umgebung []string) (erster, zweiter string) {
	t.Helper()

	menge := strconv.Itoa(deckmenge)

	for _, name := range []string{"Aggro", "Kontrolle"} {
		kennung := strings.TrimSpace(mussKommandozeile(t, umgebung,
			"deck", "neu", name, "magic"))

		mussKommandozeile(t, umgebung,
			"deck", "legen", kennung, druckkennung, menge, "hauptdeck")

		if erster == "" {
			erster = kennung
		} else {
			zweiter = kennung
		}
	}

	return erster, zweiter
}

// sitzwuensche gibt die zwei Sitze, wie MCP und Dienst sie lesen.
func sitzwuensche(erster, zweiter string) []map[string]string {
	return []map[string]string{
		{"kennung": "anna", "name": "Anna", "deckID": erster},
		{"kennung": "bert", "name": "Bert", "deckID": zweiter},
	}
}
