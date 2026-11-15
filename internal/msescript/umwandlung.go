package msescript

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Die Umwandlungsregeln von MSE — portiert, nicht erschlossen.
//
// # Herkunft
//
// `src/script/value.cpp` aus twanvl/MagicSetEditor2, Stand 2026-08-16, gelesen
// nach dem Lizenzwechsel auf GPL-2.0
// ([0019](../../docs/entscheidungen/0019-gpl-2-0-statt-mit.md)). Die Angabe
// steht hier, weil `AGENTS.md` §11 sie verlangt: die GPL will, dass eine
// Bearbeitung als solche kenntlich ist, und ein späterer Leser muss sehen
// können, was aus einem Paket erschlossen und was übernommen wurde.
//
// # Was übernommen ist
//
// Die **Regel**, nicht der Code. Der Aufbau von MSE ist eine Klassenhierarchie
// mit `virtual toInt/toDouble/toBool` je Werttyp; hier steht dieselbe Aussage
// als Funktion über einem Verbund, weil [Wert] so gebaut ist.
//
//	| Art            | Zahl        | Text        | Wahrheit          |
//	| ---            | ---         | ---         | ---               |
//	| `nil`          | 0           | ""          | falsch            |
//	| Zahl           | sie selbst  | Ziffern     | **Fehler**        |
//	| Zeichenkette   | geparst     | sie selbst  | fünf Wörter       |
//	| Wahrheitswert  | **Fehler**  | true/false  | er selbst         |
//
// **Die beiden Fehler sind der Punkt.** Eine Umwandlung, die immer gelingt,
// verdeckt jeden Tippfehler: `card.gibtsnicht * 2` gäbe 0, und niemand sähe es.
// MSE lässt `nil` durch — ein Feld ohne Wert ist der Normalfall einer frischen
// Karte — und wirft für alles, was jemand gemeint haben könnte.
//
// Fassung f9d9356 vom 2021-07-06.

// ErrUmwandlung meldet einen Wert, der sich in der verlangten Art nicht lesen
// lässt.
//
// **Mit dem Wert in der Meldung.** Ein Lauf über 190 Stile meldete sonst „ein
// Wert war keine Zahl"; welcher von 5.950, stünde nirgends.
var ErrUmwandlung = errors.New("lässt sich nicht umwandeln")

// AlsZahl gibt die Zahl, die MSE aus diesem Wert macht.
//
// Herkunft: `ScriptNil::toDouble` (0.0), `ScriptString::toDouble` (geparst oder
// `ScriptErrorConversion`), `ScriptBool` (**keine** Umwandlung — der Kommentar
// dort lautet „bools don't autoconvert to int").
func (w Wert) AlsZahl() (float64, error) {
	if w.IstFehler() {
		return 0, w.alsFehler()
	}

	switch w.Art {
	case ArtZahlwert:
		return w.Zahl, nil

	// **Der Fall, der ein Drittel aller Befunde ausmachte.** 1.606 Feldstile
	// scheiterten an `font size: … * …` mit einer leeren Seite, 327 an einem
	// Vergleich zweier leerer Seiten. Nachgemessen am 2026-08-16 über
	// `M15-Magic-Pack`.
	case ArtNichts:
		return 0, nil

	case ArtTextwert:
		zahl, err := strconv.ParseFloat(strings.TrimSpace(w.Text), 64)
		if err != nil {
			return 0, fmt.Errorf("%q %w in eine Zahl", w.Text, ErrUmwandlung)
		}

		return zahl, nil

	case ArtWahrheitswert, ArtAbbildung, ArtFunktion, ArtFarbe:
	}

	return 0, fmt.Errorf("%s %w in eine Zahl", w.Art, ErrUmwandlung)
}

// wahrheitsworte sind die fünf Wörter, die `ScriptString::toBool` kennt.
//
// **Alles andere wirft**, und das ist kein Versehen von MSE: „nicht leer, also
// wahr" ließe `if card.name then …` für jede Karte wahr sein, deren Name nicht
// zufällig `no` lautet — und für die ohne Namen falsch. Ein Feldinhalt ist
// keine Bedingung.
var wahrheitsworte = map[string]bool{
	"yes": true, "true": true,
	"no": false, "false": false, "": false,
}

// AlsWahrheitswert gibt den Wahrheitswert, den MSE aus diesem Wert macht.
//
// Herkunft: `ScriptNil::toBool` (falsch), `ScriptString::toBool` (die fünf
// Wörter, sonst `ScriptErrorConversion`), `ScriptBool::toBool`.
//
// **Eine Zahl steht hier nicht** — `ScriptInt` überschreibt `toBool` nicht und
// fällt damit auf die werfende Grundfassung zurück. Dass dieses Programm sie in
// der Bedingung eines `if` trotzdem zulässt, ist eine bewusste Abweichung; sie
// steht in `auswerter_formen.go` und in `docs/FORMATE.md`.
func (w Wert) AlsWahrheitswert() (bool, error) {
	if w.IstFehler() {
		return false, w.alsFehler()
	}

	switch w.Art {
	case ArtWahrheitswert:
		return w.Wahrheit, nil

	case ArtNichts:
		return false, nil

	case ArtTextwert:
		if trifft, bekannt := wahrheitsworte[strings.TrimSpace(w.Text)]; bekannt {
			return trifft, nil
		}

		return false, fmt.Errorf("%q %w in einen Wahrheitswert", w.Text, ErrUmwandlung)

	case ArtZahlwert, ArtAbbildung, ArtFunktion, ArtFarbe:
	}

	return false, fmt.Errorf("%s %w in einen Wahrheitswert", w.Art, ErrUmwandlung)
}
