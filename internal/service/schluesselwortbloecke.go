// Die `keyword:`-Blöcke beim **Schreiben**
// (#1062).
//
// # Derselbe Weg wie bei den Karten
//
// Der Baum kommt aus der **Quelldatei** (0011), und was der Bestand führt,
// wird hineingelegt. So bleibt stehen, was niemand angefasst hat — Kommentare,
// Abstände, Zeilenenden, ein zweites `set info:`. Die Byte-Zusage aus 0002
// hängt daran.

package service

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// schluesselwortfelder sind die fünf Zeilen eines Blocks, in MSEs Reihenfolge.
//
// `IMPLEMENT_REFLECTION(Keyword)`, `src/data/keyword.cpp`, Zeilen 86–93,
// Fassung f9d9356 vom 2021-07-06. Die Reihenfolge zählt nur für **neue**
// Blöcke; ein vorhandener behält die seine.
var schluesselwortfelder = []struct {
	schluessel string
	lies       func(*domain.Keyword) string
}{
	{"keyword", func(k *domain.Keyword) string { return k.Word }},
	{"match", func(k *domain.Keyword) string { return k.Match }},
	{"reminder", func(k *domain.Keyword) string { return k.Reminder }},
	{"rules", func(k *domain.Keyword) string { return k.Rules }},
	{"mode", func(k *domain.Keyword) string { return k.Mode }},
}

// schluesselwoerterAngleichen bringt die `keyword:`-Blöcke auf den Stand des
// Bestands.
//
// **Ein Block ohne Schlüsselwort ist ein gelöschtes und geht mit.** Ihn stehen
// zu lassen schriebe eines zurück, das jemand entfernt hat — und die Rücknahme
// dafür heißt ⌘Z, nicht Speichern.
//
// **Ein neues geht hinter das letzte vorhandene**, und wo es keines gab, ans
// Ende: so schreibt MSE selbst (`IMPLEMENT_REFLECTION(Set)`, `src/data/set.cpp`,
// Zeilen 184–204, Fassung f9d9356 vom 2021-07-06 — `keywords` steht **hinter**
// den Karten).
func schluesselwoerterAngleichen(baum *msepaket.Baum, woerter []*domain.Keyword) {
	// **Nach Stelle, nicht nach Reihenfolge in der Liste.** Nach einer
	// Löschung ist die Liste kürzer als die Datei; wer mitzählt, schreibt in
	// den Block des gelöschten Wortes und nimmt dessen fremde Zeilen mit,
	// während der eigene Block wegfällt. Angestrichen vom Zweitblick durch
	// Codex über Phase W; seit ENG-03 geht dieser Weg durch dieselbe Stelle
	// wie der der Karten (`blockangleich.go`).
	nachStelle := make(map[int]*domain.Keyword, len(woerter))
	for _, wort := range woerter {
		nachStelle[wort.Reihenfolge] = wort
	}

	bloeckeAngleichen(baum, istSchluesselwortblock, nachStelle, schluesselwortAngleichen,
		func(ab int, zeilenende string) []*msepaket.Knoten {
			return neueSchluesselwortbloecke(woerter, ab, zeilenende)
		})
}

// istSchluesselwortblock sagt, ob ein Wurzelknoten ein Schlüsselwort ist.
func istSchluesselwortblock(knoten *msepaket.Knoten) bool {
	return knoten.Zeilenart == msepaket.ArtEintrag && knoten.Schluessel == "keyword"
}

// neueSchluesselwortbloecke gibt die Blöcke, für die es in der Datei keinen gibt.
func neueSchluesselwortbloecke(
	woerter []*domain.Keyword, ab int, zeilenende string,
) []*msepaket.Knoten {
	// **Neu ist, was hinter der letzten Stelle der Datei liegt** — nicht, was
	// hinter der Listenlänge liegt. Nach einer Löschung ist die Liste kürzer
	// als die Datei, und das Abzählen fände gar nichts mehr: von Alpha und
	// Beta bleibt Beta, ein neues Gamma trägt die Stelle 2, und die Liste hat
	// zwei Einträge — `woerter[2:]` ist leer, und Gamma stünde nie in der
	// Datei.
	neue := make([]*msepaket.Knoten, 0, len(woerter))

	for _, wort := range woerter {
		if wort.Reihenfolge < ab {
			continue
		}

		knoten := &msepaket.Knoten{
			Zeilenart:  msepaket.ArtEintrag,
			Schluessel: "keyword",
			Zeilenende: zeilenende,
		}

		schluesselwortAngleichen(knoten, wort)
		neue = append(neue, knoten)
	}

	return neue
}

// schluesselwortAngleichen bringt einen Block auf den Stand eines Eintrags.
//
// Drei Dinge, und sie sind nicht dasselbe: **Werte setzen**, **Zeilen
// entfernen**, **Zeilen ergänzen** — dieselbe Dreiteilung wie bei einer Karte
// (#191), und aus demselben Grund: wer ein Feld leert und speichert, bekäme
// sonst eine Datei, in der es weiter steht.
func schluesselwortAngleichen(knoten *msepaket.Knoten, wort *domain.Keyword) {
	behalten := make([]*msepaket.Knoten, 0, len(knoten.Kinder))

	gesetzt := map[string]bool{}

	for _, kind := range knoten.Kinder {
		lies, gehoertDazu := feldlesenFuer(kind.Schluessel)
		if !gehoertDazu {
			behalten = append(behalten, kind)

			continue
		}

		// **Ein leerer Wert nimmt die Zeile weg.** Eine Zeile `match:` ohne
		// alles ist für MSE ein leeres Muster — dasselbe wie keine Zeile —,
		// und sie stünde als Rest von etwas da, das jemand gelöscht hat.
		neu := lies(wort)
		if neu == "" {
			continue
		}

		gesetzt[kind.Schluessel] = true
		behalten = append(behalten, kind)

		// **Eine unveränderte Zeile bleibt, wie sie war** — dieselbe Regel wie
		// in [werteSetzen] und aus demselben Grund: die Rohzeile trägt
		// Einrückung und Abstände, und `keyword:  Flug` mit zwei Leerzeichen
		// ist dieselbe Angabe wie mit einem. Wer sie überschriebe, bräche die
		// Byte-Zusage aus 0002 bei jedem Speichern.
		if wieInDerQuelle(neu, kind) {
			continue
		}

		// **Der alte Unterblock geht mit, wenn er der Wert war** — bei einem
		// mehrzeiligen Erinnerungstext steht hinter dem Doppelpunkt nichts.
		if kind.Wert == "" {
			kind.Kinder = nil
		}

		kind.Wert, kind.Roh = neu, ""
	}

	behalten = append(behalten, fehlendeZeilen(knoten, wort, gesetzt)...)
	knoten.Kinder = behalten
}

// feldlesenFuer gibt den Leser eines Schlüssels — und ob es einer ist.
func feldlesenFuer(schluessel string) (func(*domain.Keyword) string, bool) {
	for _, feld := range schluesselwortfelder {
		if feld.schluessel == schluessel {
			return feld.lies, true
		}
	}

	return nil, false
}

// fehlendeZeilen gibt die Zeilen, die der Block noch nicht hat.
//
// **In MSEs Reihenfolge** und nur für nichtleere Werte.
func fehlendeZeilen(
	knoten *msepaket.Knoten, wort *domain.Keyword, gesetzt map[string]bool,
) []*msepaket.Knoten {
	var neue []*msepaket.Knoten

	for _, feld := range schluesselwortfelder {
		wert := feld.lies(wort)
		if wert == "" || gesetzt[feld.schluessel] {
			continue
		}

		neue = append(neue, &msepaket.Knoten{
			Zeilenart:  msepaket.ArtEintrag,
			Schluessel: feld.schluessel,
			Wert:       wert,
			Zeilenende: zeilenendeVon(knoten.Kinder),
		})
	}

	return neue
}
