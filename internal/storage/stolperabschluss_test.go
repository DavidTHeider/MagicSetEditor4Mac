package storage

// **Die Fehlerwege des Abschlusses** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Warum es dafür einen eigenen Treiber braucht
//
// **Weil SQLite diese Wege nicht hergibt.** Eine Anweisung, die in einer
// offenen Transaktion scheitert, ein Zählwerk, das beim Nachzählen klagt,
// eine Zeile, die sich nicht lesen lässt, eine Uhr, die keine ist — das sind
// genau die Stellen, an denen ein **halber** Abschluss zurückbliebe. Ohne
// diesen Lauf stünden sie als ungeprüfte Zusagen da (§10).
//
// # Und warum die Helfer einzeln gefahren werden
//
// **Weil der ganze Weg an der ersten Stelle stehen bliebe.** Wer
// [Abschlussablage.Abschliessen] über einen Stolpertreiber fährt, prüft die
// erste Stelle und erreicht die zwanzigste nie. Die Helfer nehmen einen
// [abfrager] entgegen — sie sind einzeln zu fahren, und genau dafür ist der
// Schnitt gemacht.

import (
	"database/sql"
	"database/sql/driver"
	"strings"
)

// belegspalten sind die zwölf, die [Abschlussablage.Beleg] liest —
// **in einem Lesen** mit dem Fremdbeleg (#1650) und dem jüngsten Befund
// (#2500).
var belegspalten = []string{
	"abdruck", "form", "unterschrift", "beidseitig", "gebucht",
	"von", "empfangen",
	"befund_von", "befund_abdruck", "mangel", "grund", "befund_empfangen",
}

// Die Stellen, an denen ein Abschluss stolpern kann.
const (
	beimBesitzlesen stolperstelle = iota + 2700
	beimDeckungsscan
	beimDeckungsende
	beimMindern
	beimMindernzaehlen
	beimBestandloeschen
	beimLoeschzaehlen
	beimZugangsdruck
	beimZugangschreiben
	beimJournalschreiben
	beimJournalscan
	beimJournalende
	beimJournaluhr
	beimBelegschreiben
	beimBeleguhr
	beimFremdbeleguhr
	beimBefunduhr
	beimGegenzeichnen
	beimGegenzaehlen
	beimFremdbelegschreiben
	beimZusagenloesen
	beimVermerkschreiben
	beimBelegausgang
	beimRevisionschreiben
	beimZuwenigVorrat
	beimStummenZaehler
	beimDeckungsfragen
	beimVorgangsfragen
	beimAbschlussbeginn
	beimAbschlusscommit
	beimBelegsuche
	beimAbschlussende
)

func init() {
	for name, stelle := range map[string]stolperstelle{
		"stolper-besitz-lesen":    beimBesitzlesen,
		"stolper-deckung-scan":    beimDeckungsscan,
		"stolper-deckung-ende":    beimDeckungsende,
		"stolper-mindern":         beimMindern,
		"stolper-mindern-zaehlen": beimMindernzaehlen,
		"stolper-bestand-loesch":  beimBestandloeschen,
		"stolper-loesch-zaehlen":  beimLoeschzaehlen,
		"stolper-zugang-druck":    beimZugangsdruck,
		"stolper-zugang":          beimZugangschreiben,
		"stolper-journal-schreib": beimJournalschreiben,
		"stolper-journal-scan":    beimJournalscan,
		"stolper-journal-ende":    beimJournalende,
		"stolper-journal-uhr":     beimJournaluhr,
		"stolper-beleg-schreib":   beimBelegschreiben,
		"stolper-beleg-uhr":       beimBeleguhr,
		"stolper-fremdbeleg-uhr":  beimFremdbeleguhr,
		"stolper-befund-uhr":      beimBefunduhr,
		"stolper-gegenzeichnen":   beimGegenzeichnen,
		"stolper-gegen-zaehlen":   beimGegenzaehlen,
		"stolper-fremdbeleg":      beimFremdbelegschreiben,
		"stolper-zusagen-loesen":  beimZusagenloesen,
		"stolper-vermerk":         beimVermerkschreiben,
		"stolper-belegausgang":    beimBelegausgang,
		"stolper-revision":        beimRevisionschreiben,
		"stolper-zuwenig":         beimZuwenigVorrat,
		"stolper-stumm":           beimStummenZaehler,
		"stolper-deckung-fragen":  beimDeckungsfragen,
		"stolper-vorgangszusagen": beimVorgangsfragen,
		"stolper-abschlussbeginn": beimAbschlussbeginn,
		"stolper-abschlusscommit": beimAbschlusscommit,
		"stolper-belegsuche":      beimBelegsuche,
	} {
		sql.Register(name, stolpertreiber{stelle: stelle})
	}
}

// imAbschluss sagt, ob eine Stelle zu diesem Werkstück gehört.
//
// **Ohne diese Frage nähme der Treiber fremden Läufen ihre Anweisungen ab.**
// `INSERT INTO handelsbriefkasten` gibt es auch in T5, und ein Handler, der
// jede Anweisung beantwortet, sobald ihr Vorsatz passt, machte aus einem
// fremden Fehlerweg einen glatten Lauf (§10).
func imAbschluss(stelle stolperstelle) bool {
	return stelle >= beimBesitzlesen && stelle <= beimAbschlussende
}

// abschlussschreiben beantwortet die Anweisungen des Abschlusses.
func abschlussschreiben(
	stelle stolperstelle, abfrage string,
) (driver.Result, bool, error) {
	if !imAbschluss(stelle) {
		return nil, false, nil
	}

	geglaettet := strings.TrimSpace(abfrage)

	for _, fall := range []struct {
		vorsatz  string
		stolpert stolperstelle
		zaehlt   stolperstelle
	}{
		{"UPDATE sammlung SET menge = menge - ?", beimMindern, beimMindernzaehlen},
		{"DELETE FROM sammlung", beimBestandloeschen, beimLoeschzaehlen},
		{"INSERT INTO sammlung", beimZugangschreiben, 0},
		{"INSERT INTO handelsjournal", beimJournalschreiben, 0},
		{"INSERT INTO handelsbelege", beimBelegschreiben, 0},
		{"UPDATE handelsbelege", beimGegenzeichnen, beimGegenzaehlen},
		{"INSERT INTO handelsfremdbelege", beimFremdbelegschreiben, 0},
		{"DELETE FROM reservierungen", beimZusagenloesen, 0},
		{"INSERT INTO handelsbriefkasten", beimVermerkschreiben, 0},
		{"INSERT INTO handelsausgang", beimBelegausgang, 0},
		{"INSERT INTO handelsrevisionen", beimRevisionschreiben, 0},
	} {
		if !strings.HasPrefix(geglaettet, fall.vorsatz) {
			continue
		}

		switch stelle {
		case fall.stolpert:
			return nil, true, errGestolpert
		case fall.zaehlt:
			return klagendeZaehlung{}, true, nil
		case beimStummenZaehler:
			return stummeWirkung{}, true, nil
		}

		return leereWirkung{}, true, nil
	}

	return nil, false, nil
}

// abschlussabfrage beantwortet die Abfragen des Abschlusses.
func abschlussabfrage(
	stelle stolperstelle, abfrage string,
) (driver.Rows, bool, error) {
	if !imAbschluss(stelle) {
		return nil, false, nil
	}

	geglaettet := strings.TrimSpace(abfrage)

	switch {
	// **Ob der Vorgang schon einen Beleg hat** (C05, #2230) — er hat keinen,
	// sonst käme keine Stelle dahinter je dran.
	case strings.HasPrefix(geglaettet, "SELECT COUNT(*) FROM handelsbelege"):
		if stelle == beimBelegsuche {
			return nil, true, errGestolpert
		}

		return &werteZeile{spalten: []string{"anzahl"}, werte: []driver.Value{int64(0)}}, true, nil

	// **Mit der Variante** (#2492): `besitzInTx` liest sie für das
	// Festhalten mit, und der Abschluss teilt sich den Ableser.
	case strings.HasPrefix(geglaettet, "SELECT menge, fassung, wunsch, printing"):
		spalten := []string{
			"menge", "fassung", "wunsch", "printing", "druck_id", "zustand",
			"sprache", "finish",
		}

		if stelle == beimBesitzlesen {
			return &krummeZeile{spalten: spalten}, true, nil
		}

		// **Ein Stück statt vier** — die Zusage deckt mehr, als dasteht.
		// Dass das überhaupt geht, ist der Grund für die Prüfung: eine
		// Zusage rechnet gegen eine Fassung und nicht gegen den Vorrat.
		vorrat := int64(4)
		if stelle == beimZuwenigVorrat {
			vorrat = 1
		}

		// **Vier Stück, Fassung eins, kein Wunsch** — der Stand, gegen den
		// die Stellen dahinter geprüft werden.
		return &werteZeile{
			spalten: spalten,
			werte: []driver.Value{
				vorrat, int64(1), int64(0), "bolt-m10", int64(0), "nearMint", "de", "nonfoil",
			},
		}, true, nil

	case strings.HasPrefix(geglaettet, "SELECT menge, fassung FROM reservierungen"):
		spalten := []string{"menge", "fassung"}

		switch stelle {
		case beimDeckungsscan:
			return &krummeZeile{spalten: spalten}, true, nil
		case beimDeckungsende:
			return &abbrechendeZeile{spalten: spalten}, true, nil
		case beimDeckungsfragen:
			// **Eine Abfrage, die gar nicht erst läuft** — die Verbindung ist
			// fort. Sie ist der Grund dafür, dass hier überhaupt gefragt wird.
			return nil, true, errGestolpert
		}

		// **Vier gedeckte Stück in Fassung eins** — genug für jede Probe.
		return &werteZeile{
			spalten: spalten,
			werte:   []driver.Value{int64(4), int64(1)},
		}, true, nil

	// **Die Zusagen eines Vorgangs** — ihre Zeilen prüft der Nachbar über den
	// Weg je Exemplar; hier fehlt die Abfrage selbst.
	case strings.HasPrefix(geglaettet, "SELECT vorgang, exemplar_id, revision"):
		if stelle == beimVorgangsfragen {
			return nil, true, errGestolpert
		}

		return &leerezeile{spalten: 7}, true, nil

	case strings.HasPrefix(geglaettet, "SELECT id, art, anlass"):
		return journalzeilen(stelle), true, nil

	// **Zwölf Spalten seit dem 2026-10-19**: der Beleg wird mit seinem
	// Fremdbeleg (#1650) und dem jüngsten Befund (#2500) in **einem** Lesen
	// geholt. **Je Uhr eine Stelle**, und jede liefert alle zwölf — sonst
	// scheiterte das Lesen an der Spaltenzahl und nicht an der Uhr.
	case strings.HasPrefix(geglaettet, "SELECT b.abdruck, b.form"):
		if zeile := belegzeile(stelle); zeile != nil {
			return &werteZeile{spalten: belegspalten, werte: zeile}, true, nil
		}

		return &leerezeile{spalten: len(belegspalten)}, true, nil

	// **Der Druck des Zugangs** — `druckZeile` legt ihn an, wenn es ihn nicht
	// gibt, und der Stolpertreiber gibt ihn her oder verweigert ihn.
	case strings.HasPrefix(geglaettet, "INSERT INTO drucke"):
		if stelle == beimZugangsdruck {
			return &leerezeile{spalten: 1}, true, nil
		}

		return &werteZeile{
			spalten: []string{"id"}, werte: []driver.Value{int64(7)},
		}, true, nil
	}

	return nil, false, nil
}

// belegzeile ist die Zeile des Belegs an einer Stelle, die eine Uhr kippt —
// `nil` an jeder anderen.
func belegzeile(stelle stolperstelle) []driver.Value {
	ohneBefund := []driver.Value{nil, nil, nil, nil, nil}

	switch stelle {
	case beimBeleguhr:
		return append([]driver.Value{"a", "f", "u", int64(0), "keine Uhr", nil, nil}, ohneBefund...)
	case beimFremdbeleguhr:
		return append([]driver.Value{
			"a", "f", "u", int64(1), "2026-09-19 12:00:00", "bert", "keine Uhr",
		}, ohneBefund...)
	case beimBefunduhr:
		return []driver.Value{
			"a", "f", "u", int64(0), "2026-09-19 12:00:00", nil, nil,
			"bert", "a1", "zeilen", "zwei statt drei", "keine Uhr",
		}
	}

	return nil
}

// journalzeilen beantwortet die Leseabfrage des Journals.
func journalzeilen(stelle stolperstelle) driver.Rows {
	spalten := []string{
		"id", "art", "anlass", "vorgang", "bezug", "grund", "printing",
		"menge", "zustand", "sprache", "finish", "exemplar_id", "lagerort",
		"notiz", "gebucht",
	}

	switch stelle {
	case beimJournalscan:
		return &krummeZeile{spalten: spalten}

	case beimJournalende:
		return &abbrechendeZeile{spalten: spalten}

	case beimJournaluhr:
		// **Eine Zeile, deren Uhr keine ist.** Sie kommt aus einer Datei, die
		// jemand von Hand berührt hat — SQLite hält den Text, nicht sein
		// Format.
		return &werteZeile{
			spalten: spalten,
			werte: []driver.Value{
				int64(1), "zugang", "kauf", "", int64(0), "", "bolt-m10",
				int64(1), "near_mint", "de", "nonfoil", int64(0), "", "",
				"keine Uhr",
			},
		}
	}

	return &leerezeile{spalten: len(spalten)}
}

// stummeWirkung trifft keine Zeile — und klagt darüber nicht.
//
// **Den Fall gibt SQLite in einer offenen Transaktion nicht her**: die Fassung
// wurde eben gelesen. Der Zweig dahinter ist trotzdem der, an dem ein zweiter
// Schreiber laut würde, statt still zu buchen.
type stummeWirkung struct{}

func (stummeWirkung) LastInsertId() (int64, error) { return 0, nil }
func (stummeWirkung) RowsAffected() (int64, error) { return 0, nil }
