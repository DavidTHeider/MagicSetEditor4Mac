// Die Prüfungen darüber, dass der Setschreiber Feldzeilen **entfernt und
// ergänzt** — nicht nur ihre Werte neu setzt.
//
// Von `setschreiber_test.go` abgetrennt, weil die Datei über die Zeilengrenze
// aus AGENTS.md §2 gewachsen wäre. Der Schnitt folgt der Frage: dort steht, was
// aus einer geänderten Zeile wird, hier, was aus einer wird, die es nicht mehr
// gibt oder noch nicht gab
// (#191).
package service_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// dreizeilenquelle ist eine `.mse-set` mit drei Feldern in einer Karte.
//
// **Mit einem Kommentar über der mittleren Zeile**, und das ist Absicht: was
// mit ihm geschieht, wenn seine Zeile fortfällt, ist die Entwurfsfrage aus
// #191 und keine Nebensache.
const dreizeilenquelle = "\ufeffmse version: 2.0.0\n" +
	"game: probe\n" +
	"card:\n" +
	"\tname:   Feuerball\n" +
	"\t# zur Farbe\n" +
	"\tfarbe: rot\n" +
	"\ttext: Drei Schaden\n" +
	"unverstandener block:\n" +
	"\tirgendwas: bleibt liegen"

// dreizeilenbestand gibt den Bestand zu [dreizeilenquelle].
func dreizeilenbestand(werte ...domain.Value) quellenDouble {
	return quellenDouble{
		leserDouble: leserDouble{
			koepfe: []domain.SetHead{{ID: 7, Name: "Mein Probeset", Game: "probe"}},
			karten: []*domain.CardRow{{ID: 1, Number: 0, Values: werte}},
		},
		quelle: &domain.Quelle{Datei: "set", Inhalt: []byte(dreizeilenquelle)},
	}
}

// geschriebenesSet schreibt das Set und gibt die Hauptdatei zurück.
func geschriebenesSet(t *testing.T, bestand quellenDouble) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "wieder.mse-set")

	if _, err := service.NeuerSetschreiber(bestand, &ortsschreiber{}).Schreibe(t.Context(), 7, ort); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	gelesen, err := os.ReadFile(filepath.Join(ort, "set")) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	return string(gelesen)
}

// **Ein Feld, das der Bestand nicht mehr führt, verliert seine Zeile.**
//
// # Warum das mehr ist als Ordnungsliebe
//
// Bis hierher setzte `karteAngleichen` nur den Wert **vorhandener** Zeilen neu.
// Wer ein Feld entfernte und speicherte, bekam eine Datei, in der es weiter
// stand — **die Entfernung erreichte die Datei nicht.** Die Datei ist die
// Wahrheit, die MSE2 später liest.
func TestEinEntferntesFeldVerliertSeineZeile(t *testing.T) {
	gelesen := geschriebenesSet(t, dreizeilenbestand(
		domain.Value{Field: "name", Text: "Feuerball"},
		domain.Value{Field: "text", Text: "Drei Schaden"}))

	if strings.Contains(gelesen, "farbe") {
		t.Errorf("die Zeile steht noch da:\n%s", gelesen)
	}

	// **Und sonst hat sich nichts bewegt** — der doppelte Abstand hinter
	// `name:`, der unverstandene Block, die BOM.
	erwartet := strings.Replace(dreizeilenquelle, "\tfarbe: rot\n", "", 1)
	if gelesen != erwartet {
		t.Errorf("Abweichung:\n%q\nerwartet\n%q", gelesen, erwartet)
	}
}

// **Der Kommentar über einer entfernten Zeile bleibt stehen.**
//
// # Die Entscheidung, und sie ist eine
//
// Ein Kommentar **könnte** zu der Zeile darunter gehören — `# zur Farbe` über
// `farbe: rot` sieht ganz danach aus. Nur steht das nirgends: das Paketformat
// kennt keine Bindung zwischen Kommentar und Eintrag, und der Zerteiler hängt
// ihn an die offene Ebene, nicht an seinen Nachbarn (`docs/FORMATE.md`).
//
// **Ihn mitzunehmen hieße, eine Bindung zu erfinden und danach Text zu
// löschen, den jemand geschrieben hat.** Das ist die teurere Richtung: ein
// übrig gebliebener Kommentar ist sichtbar und in einer Sekunde entfernt; ein
// verschwundener ist fort.
func TestEinKommentarUeberEinerEntferntenZeileBleibt(t *testing.T) {
	gelesen := geschriebenesSet(t, dreizeilenbestand(
		domain.Value{Field: "name", Text: "Feuerball"},
		domain.Value{Field: "text", Text: "Drei Schaden"}))

	if !strings.Contains(gelesen, "# zur Farbe") {
		t.Errorf("der Kommentar ist fort:\n%s", gelesen)
	}
}

// **Ein Feld, das der Bestand neu führt, bekommt seine Zeile — an seiner
// Stelle.**
//
// # Wohin, und woher wir das wissen
//
// Aus der Reihenfolge des Bestands: `card_values.reihenfolge` sagt, wo das Feld
// in der Karte steht, und seit
// #160
// überlebt sie eine Rücknahme. Die neue Zeile kommt **hinter die des
// Vorgängers** — das ist die einzige Stelle, die sich aus dem ableitet, was
// dasteht, statt sie zu erfinden.
//
// Hat der Vorgänger selbst keine Zeile, wird weiter zurückgesehen; hat keiner
// eine, kommt sie an den Anfang des Kartenblocks.
func TestEinNeuesFeldBekommtSeineZeileAnSeinerStelle(t *testing.T) {
	gelesen := geschriebenesSet(t, dreizeilenbestand(
		domain.Value{Field: "name", Text: "Feuerball"},
		domain.Value{Field: "kosten", Text: "2R"},
		domain.Value{Field: "farbe", Text: "rot"},
		domain.Value{Field: "text", Text: "Drei Schaden"}))

	// **Hinter den Kommentar, nicht davor** — und das ist keine Nachlässigkeit,
	// sondern das Format: der Zerteiler hängt einen Kommentar an die gerade
	// offene Ebene, also an den Eintrag **darüber** (`docs/FORMATE.md`).
	// `# zur Farbe` ist damit ein Kind von `name`, und wer hinter `name`
	// einfügt, fügt hinter dessen Block ein.
	//
	// Anders herum ginge nur, indem man den Kommentar dem Eintrag darunter
	// zuschlüge — eine Bindung, die das Format nicht kennt.
	erwartet := strings.Replace(dreizeilenquelle,
		"\t# zur Farbe\n", "\t# zur Farbe\n\tkosten: 2R\n", 1)

	if gelesen != erwartet {
		t.Errorf("Abweichung:\n%q\nerwartet\n%q", gelesen, erwartet)
	}
}

// **Ein neues Feld ganz vorn kommt an den Anfang des Kartenblocks.**
func TestEinNeuesFeldOhneVorgaengerKommtNachVorn(t *testing.T) {
	gelesen := geschriebenesSet(t, dreizeilenbestand(
		domain.Value{Field: "nummer", Text: "001"},
		domain.Value{Field: "name", Text: "Feuerball"},
		domain.Value{Field: "farbe", Text: "rot"},
		domain.Value{Field: "text", Text: "Drei Schaden"}))

	erwartet := strings.Replace(dreizeilenquelle,
		"card:\n", "card:\n\tnummer: 001\n", 1)

	if gelesen != erwartet {
		t.Errorf("Abweichung:\n%q\nerwartet\n%q", gelesen, erwartet)
	}
}

// **Eine neue Zeile nimmt das Zeilenende, das der Kartenblock schon benutzt.**
//
// Eine echte `.mse-set` trägt `\n` und `\r\n` **in derselben Datei**
// (`docs/FORMATE.md`); eine neue Zeile mit dem falschen wäre ein Unterschied,
// den niemand angeordnet hat.
//
// **Und wo gar keines dasteht, gilt `\n`.** Der Fall kommt vor: die letzte
// Zeile einer Datei darf ohne Umbruch enden, und ein Kartenblock, der aus
// genau ihr besteht, hat keinen Anhaltspunkt.
func TestEineNeueZeileNimmtDasZeilenendeDesBlocks(t *testing.T) {
	faelle := map[string]struct {
		quelle, erwartet string
	}{
		"der Block benutzt CRLF": {
			quelle:   "card:\r\n\tname: Feuerball\r\n",
			erwartet: "card:\r\n\tname: Feuerball\r\n\tfarbe: rot\r\n",
		},
		// **Die letzte Zeile darf ohne Umbruch enden** (`docs/FORMATE.md`) —
		// und die alte bekommt trotzdem eines, sobald sie einen Nachfolger hat.
		// Ohne das klebten beide aneinander: `name: Feuerball\tfarbe: rot`.
		//
		// Die **neue** letzte Zeile bekommt eines, und das ist der Schreiber:
		// er gibt selbst gebauten Knoten seit Phase 2 ein `\n`, damit eine
		// erzeugte Zeile nie an die nächste stößt. Ein zusätzlicher Umbruch am
		// Dateiende ist der Preis, und er ist der kleinere.
		"der Block hat gar keines": {
			quelle:   "card:\n\tname: Feuerball",
			erwartet: "card:\n\tname: Feuerball\n\tfarbe: rot\n",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			bestand := dreizeilenbestand(
				domain.Value{Field: "name", Text: "Feuerball"},
				domain.Value{Field: "farbe", Text: "rot"})
			bestand.quelle = &domain.Quelle{Datei: "set", Inhalt: []byte(fall.quelle)}

			if gelesen := geschriebenesSet(t, bestand); gelesen != fall.erwartet {
				t.Errorf("Abweichung:\n%q\nerwartet\n%q", gelesen, fall.erwartet)
			}
		})
	}
}
