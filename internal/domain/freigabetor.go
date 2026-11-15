package domain

// **Das Tor vor jeder Sicht** — Werkstück `T3` von Phase 5.6
// (#1468).
//
// # Warum es **ein** Tor gibt und nicht eines je Aufrufer
//
// **Weil das Issue die Prüfung an fünf Stellen verlangt** — vor jeder
// Projektion, vor jeder Antwort, vor dem Zwischenspeicher, vor der
// Wiederaufnahme, vor Ausfuhr und Beleg. Fünf Stellen, die je selbst
// entscheiden, entscheiden irgendwann verschieden, und die erste, die es
// merkt, ist die, an der etwas hinausging.
//
// # Und warum das Nein ein [ErrErlaubnis] ist
//
// **Weil es keine Auskunft über den Bestand geben darf.** „Gibt es nicht" und
// „darfst du nicht" sind für ein fremdes Gegenüber dieselbe Antwort; wer sie
// unterschiede, ließe die Sammlung abfragen, ohne sie zu sehen. Und weil zwei
// Flächen ihn schon lesen: REST macht `403` daraus, `internal/app` den Code
// `erlaubnis`.

import (
	"fmt"
	"slices"
	"time"
)

// FreigabenFuer gibt die Freigaben, die einem Kontakt für einen Zweck gelten.
//
// # Warum mehrere und nicht eine
//
// **Weil die fünf Umfänge einzeln erteilt werden.** Das Issue listet sie als
// eigene Schalter: jemand gibt seine Dubletten frei **und** seine Wünsche, und
// beides gilt nebeneinander. Wer hier eine einzelne Freigabe zurückgäbe,
// müsste entscheiden, welche — und die andere fiele still weg.
//
// # Die Reihenfolge der Fragen ist die Reihenfolge der Gründe
//
// **Erst der Kontakt, dann die Freigabe.** Ein gesperrter Kontakt bekommt
// nichts, gleich welche Freigaben herumliegen — das ist die Zusage aus `N1`
// („blockiert bleibt auch nach Spiel→Handeln blockiert"), und sie wäre
// gebrochen, sobald eine Freigabe sie überstimmen könnte.
//
// **Und eine Paarung ist keine Freigabe.** [Peerkontakt.Darf] lässt sie für
// [ZweckPaarung] durch, weil sie das Kennenlernen ist; für jeden anderen
// Zweck verlangt sie einen bestätigten Kontakt mit dem Recht darauf.
func FreigabenFuer(
	kontakt Peerkontakt, freigaben []Sammlungsfreigabe, zweck Zweck,
	jetzt time.Time,
) ([]Sammlungsfreigabe, error) {
	if !kontakt.Darf(zweck) || zweck == ZweckPaarung {
		return nil, fmt.Errorf(
			"%w: dieser Kontakt hat für %q keine Erlaubnis", ErrErlaubnis, zweck)
	}

	gelten := []Sammlungsfreigabe{}

	for _, freigabe := range freigaben {
		if freigabe.Empfaenger != kontakt.Kennung || freigabe.Zweck != zweck {
			continue
		}

		if freigabe.Gilt(jetzt) {
			gelten = append(gelten, freigabe)
		}
	}

	if len(gelten) == 0 {
		return nil, fmt.Errorf(
			"%w: für %q ist nichts freigegeben", ErrErlaubnis, zweck)
	}

	return gelten, nil
}

// Sichtmenge sagt, mit welcher Menge ein Bündel von Freigaben ein Exemplar
// zeigt.
//
// **Die größte gewinnt.** Wer seine Dubletten **und** seine ganze Sammlung
// freigibt, hat die Sammlung freigegeben; die engere Freigabe daneben nimmt
// davon nichts weg. Der umgekehrte Satz — die engste gewinnt — wäre die Regel
// für eine **Schranke** (0118) und ist hier falsch: eine Freigabe ist eine
// Erlaubnis, und zwei Erlaubnisse verbieten einander nicht.
func Sichtmenge(freigaben []Sammlungsfreigabe, exemplar Exemplar) (int, bool) {
	groesste, gezeigt := 0, false

	for _, freigabe := range freigaben {
		if menge, zeigt := freigabe.Zeigt(exemplar); zeigt {
			gezeigt = true
			groesste = max(groesste, menge)
		}
	}

	return groesste, gezeigt
}

// Validate hält eine Freigabe gegen das, was eine sein kann.
//
// **Sie prüft auch, was **zu wenig** ist.** Eine Auswahl ohne Auswahl sieht
// aus wie eine enge Freigabe und ist eine, die nichts zeigt — wer sie erteilt,
// glaubt, er habe etwas freigegeben.
func (f Sammlungsfreigabe) Validate() error {
	if f.Empfaenger == "" {
		return fmt.Errorf(
			"%w: eine Freigabe ohne Empfänger ist keine", ErrValidation)
	}

	if !IstFreigabeumfang(f.Umfang) {
		return NewFieldError("umfang", SchluesselValidierung,
			"Den Umfang %q gibt es nicht.", string(f.Umfang))
	}

	if !slices.Contains(DieZwecke, f.Zweck) || f.Zweck == ZweckPaarung {
		return NewFieldError("zweck", SchluesselValidierung,
			"Für den Zweck %q lässt sich nichts freigeben.", string(f.Zweck))
	}

	return f.pruefeUmfangsangaben()
}

// MitVorgaben setzt, was eine Freigabe ohne Angabe bedeutet — UXR-F18
// (#2033).
//
// **Eine Behaltemenge von null heißt bei „Dubletten" eins.** Mit null wäre
// jedes Stück eine Dublette, und die Sicht gäbe die ganze Sammlung frei — wer
// das will, nimmt [UmfangSammlung]. Die Vorgabe ist dieselbe wie bei
// `mse dubletten` und in der Ordnung der Sammlung; sie steht hier, damit
// Fenster, Kommandozeile und Dienst für dieselbe Eingabe dieselbe Freigabe
// erteilen (§16). Eine negative Angabe bleibt, damit [Sammlungsfreigabe.Validate]
// sie anstreicht.
func (f Sammlungsfreigabe) MitVorgaben() Sammlungsfreigabe {
	if f.Umfang == UmfangDubletten && f.Behalten == 0 {
		f.Behalten = 1
	}

	return f
}

// pruefeUmfangsangaben hält die Angaben gegen den gewählten Umfang.
//
// **Eine Angabe, die zum Umfang nicht passt, ist ein Befund und kein
// Nebenbei.** Wer eine Behaltemenge an eine Sammlungsfreigabe schreibt, hat
// etwas anderes gemeint, als er erteilt — und niemand sieht es ihm an.
func (f Sammlungsfreigabe) pruefeUmfangsangaben() error {
	if f.Umfang == UmfangAuswahl && len(f.Auswahl) == 0 {
		return NewFieldError("auswahl", SchluesselValidierung,
			"Eine Auswahl ohne gewählte Drucke gibt nichts frei.")
	}

	if f.Umfang != UmfangAuswahl && len(f.Auswahl) > 0 {
		return NewFieldError("auswahl", SchluesselValidierung,
			"Gewählte Drucke ergeben nur bei einer Auswahl einen Sinn.")
	}

	if f.Umfang != UmfangDubletten && f.Behalten != 0 {
		return NewFieldError("behalten", SchluesselValidierung,
			"Eine Behaltemenge ergibt nur bei Dubletten einen Sinn.")
	}

	if f.Behalten < 0 {
		return NewFieldError("behalten", SchluesselValidierung,
			"%d ist keine Stückzahl, die man behält.", f.Behalten)
	}

	return nil
}
