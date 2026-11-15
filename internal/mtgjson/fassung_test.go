// **Fassung und Schema einer MTG-JSON-Datei** —
// Abnahmekriterium aus #888:
// „Fassungs- und Schemaprüfung; Identifiers, Rulings, mehrseitige Karten."
package mtgjson_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// **Der Umschlag nennt seine Fassung, und sie wird gelesen.**
//
// # Warum das keine Abweisung ist
//
// MTGJSON 5 legt `{"meta": {"date": …, "version": …}, "data": …}` um alles.
// Eine Datei **von Hand** hat den Umschlag nicht (§8, „fremde Daten sind der
// Normalfall") — sie deshalb abzuweisen hieße, den häufigsten Fall zu
// verbieten. Und eine Datei aus einer künftigen Hauptfassung soll geöffnet
// werden, nicht verweigert; was sich geändert haben mag, steht als **Befund**
// daneben.
func TestDieFassungWirdGelesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		inhalt  string
		fassung string
		befund  string
	}{
		"eine bekannte Hauptfassung": {
			inhalt: `{"meta":{"date":"2026-08-24","version":"5.2.2+20260824"},
			          "data":{"code":"PRF","name":"Prüfsatz","cards":[]}}`,
			fassung: "5.2.2+20260824",
		},
		"eine unbekannte Hauptfassung": {
			inhalt: `{"meta":{"date":"2027-01-01","version":"6.0.0+20270101"},
			          "data":{"code":"PRF","name":"Prüfsatz","cards":[]}}`,
			fassung: "6.0.0+20270101",
			befund:  "6",
		},
		"ohne Umschlag": {
			inhalt: `{"code":"PRF","name":"Prüfsatz","cards":[]}`,
		},
		"Umschlag ohne meta": {
			inhalt: `{"data":{"code":"PRF","name":"Prüfsatz","cards":[]}}`,
		},
		"eine unleserliche Fassung": {
			inhalt: `{"meta":{"version":"morgen"},
			          "data":{"code":"PRF","name":"Prüfsatz","cards":[]}}`,
			fassung: "morgen",
			befund:  "morgen",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fassung, befunde := mtgjson.Fassung([]byte(fall.inhalt), "probe.json")

			if fassung.Version != fall.fassung {
				t.Errorf("Version = %q, erwartet %q", fassung.Version, fall.fassung)
			}

			if fall.befund == "" {
				if len(befunde) != 0 {
					t.Errorf("Befunde = %+v, erwartet keine", befunde)
				}

				return
			}

			if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, fall.befund) {
				t.Errorf("Befunde = %+v, erwartet einen über %q", befunde, fall.befund)
			}
		})
	}
}

// **Das Datum kommt mit, unverändert.**
//
// Es steht in ISO 8601 da, und wer es umrechnete, verschöbe es um bis zu einen
// Tag — dieselbe Überlegung wie bei Scryfall (§3).
func TestDasBaudatumKommtMit(t *testing.T) {
	t.Parallel()

	fassung, _ := mtgjson.Fassung([]byte(
		`{"meta":{"date":"2026-08-24","version":"5.2.2"},"data":{}}`), "probe.json")

	if fassung.Datum != "2026-08-24" {
		t.Errorf("Datum = %q", fassung.Datum)
	}
}

// **Die vier neuen Schlüssel stehen in der Schematabelle.**
//
// # Warum das eine eigene Zusicherung ist
//
// Die Tabelle sagt, welche **Art** ein Schlüssel hat — und `pruefeKarte`
// meldet einen Befund, wenn eine Datei etwas anderes liefert. Ein Schlüssel,
// der **fehlt**, wird dagegen stillschweigend durchgelassen: „Unbekannte
// Schlüssel sind kein Befund", sagt die Tabelle selbst, und das ist richtig,
// weil MTGJSON wächst.
//
// **Der Preis ist, dass eine falsche Art unbemerkt bliebe.** Wer `rulings` als
// Objekt statt als Liste schickte, käme durch. Deshalb stehen die vier hier —
// sie sind die, die dieses Programm **benutzt**.
func TestDieMehrseitigenSchluesselStehenInDerTabelle(t *testing.T) {
	t.Parallel()

	erwartet := map[string]mtgjson.Art{
		"rulings":      mtgjson.ArtListe,
		"otherFaceIds": mtgjson.ArtWortliste,
		"faceName":     mtgjson.ArtZeichenkette,
		"side":         mtgjson.ArtZeichenkette,
	}

	arten := mtgjson.Schluesselarten()

	for schluessel, art := range erwartet {
		bekommen, da := arten[schluessel]
		if !da {
			t.Errorf("%q fehlt in der Schematabelle", schluessel)

			continue
		}

		if bekommen != art {
			t.Errorf("%q hat die Art %v, erwartet %v", schluessel, bekommen, art)
		}
	}
}

// **Ein Umschlag mit leerer Fassungsangabe meldet nichts.**
//
// Sie kann leer sein: MTGJSON schreibt sie erst seit v4.0.0, und eine von Hand
// geschriebene Datei kann `meta` führen und `version` nicht. Ein Befund darüber
// sagte nichts, was der Leser nicht schon weiß.
func TestEineLeereFassungsangabeMeldetNichts(t *testing.T) {
	t.Parallel()

	fassung, befunde := mtgjson.Fassung(
		[]byte(`{"meta":{"date":"2026-08-24"},"data":{}}`), "probe.json")

	if fassung.Datum != "2026-08-24" || fassung.Version != "" {
		t.Errorf("Fassung = %+v", fassung)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde = %+v", befunde)
	}
}

// **Eine Liste von Objekten wird als solche erkannt.**
//
// # Warum das eine eigene Zusicherung ist
//
// [mtgjson.ArtListe] kam mit `rulings` (#888). Ohne diesen Lauf wäre nur
// belegt, dass der Schlüssel in der Tabelle **steht** — nicht, dass die
// Prüfung ihn auch deuten kann. Ein Wächter, der die halbe Zusage hält, macht
// die andere unsichtbar (§10).
func TestEineListeVonObjektenWirdErkannt(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		roh   string
		passt bool
	}{
		"eine Liste von Objekten": {`[{"date":"2004-10-04","text":"Ein Hinweis."}]`, true},
		"eine leere Liste":        {`[]`, true},
		"eine Wortliste":          {`["kein Objekt"]`, false},
		"ein Objekt":              {`{"date":"2004-10-04"}`, false},
		"eine Zeichenkette":       {`"gar nichts"`, false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if mtgjson.PasstZurArt([]byte(fall.roh), mtgjson.ArtListe) != fall.passt {
				t.Errorf("PasstZurArt(%s, ArtListe) = %v, erwartet %v",
					fall.roh, !fall.passt, fall.passt)
			}
		})
	}
}
