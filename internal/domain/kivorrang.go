package domain

// **Wer einen Vorschlag zuerst bekommt** — AI-05.AK-05
// (#2042),
// Vermerk [0166](../../docs/entscheidungen/0166-was-fm-serve-belegt.md).
//
// # Die Regel
//
// **Eine ausdrückliche Wahl gewinnt.** Ohne sie fragt eine **kleine
// Textaufgabe** zuerst das Gerätemodell des Betriebssystems und fällt dann
// auf die **lokalen** Anbieter zurück; eine größere geht nie an das Gerät.
// **Ein Cloudanbieter kommt von selbst nie dran** — er bleibt hinter Journal
// und Freigabe wie bisher.
//
// # Was hier nicht steht
//
// **Ob ein Anbieter wirklich antwortet.** `Bereit` sagt, dass gefragt werden
// darf und die Einrichtung stimmt; ob die Antwort kommt, zeigt erst der Ruf,
// und dann gilt der nächste in der Reihe.

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// KIWeg sagt, wie ein Anbieter an einen Vorschlag kam.
type KIWeg string

const (
	// KIWegGewaehlt heißt: jemand hat ihn genannt.
	KIWegGewaehlt KIWeg = "gewaehlt"

	// KIWegVorrang heißt: er stand ohne Wahl vorn.
	KIWegVorrang KIWeg = "vorrang"

	// KIWegRueckfall heißt: ein Anbieter vor ihm hat nicht geliefert.
	KIWegRueckfall KIWeg = "rueckfall"
)

// KIVorrang ist der Anbietername, mit dem ein Ruf den Vorrang verlangt.
const KIVorrang = "vorrang"

// kleineTextfelder sind die Felder, deren Vorschlag eine kleine Textaufgabe
// ist — **mindestens** Kartenname und Flavortext, wie das Issue sie nennt.
// Die Namen sind die der MSE-Spiele (`name`, `flavor text`); ein anderes
// kurzes Feld kommt hinzu, wenn seine Länge gemessen ist.
var kleineTextfelder = map[string]bool{"name": true, "flavor text": true}

// KIKleineAuftragsgrenze ist, wie viele Zeichen der Auftrag einer kleinen
// Aufgabe höchstens hat.
//
// **Gemessen am 2026-10-03** mit `fm count-tokens` (Vermerk 0166): das
// Fenster des Gerätemodells fasst 8.009 Eingangstokens und nicht 9.009; ein
// Satz braucht auf Deutsch 4,1 Zeichen je Token, auf Englisch 4,6, auf
// Chinesisch **1,3** — das ist der ungünstigste gemessene Fall. 2.000
// Zeichen sind dort rund 1.540 Tokens, und Anweisung, Schema und Antwort
// haben daneben mehr als drei Viertel des Fensters.
const KIKleineAuftragsgrenze = 2000

// IstKleineTextaufgabe sagt, ob ein Vorschlag eine kleine Textaufgabe ist:
// er fragt nur nach kleinen Textfeldern, sein Auftrag bleibt unter der
// Grenze, und er hängt **keine Belegfrage** an — deren Stellen bestimmt der
// Index und nicht der Auftrag.
func IstKleineTextaufgabe(felder []string, auftrag, belegfrage string) bool {
	if len(felder) == 0 || strings.TrimSpace(belegfrage) != "" ||
		utf8.RuneCountInString(auftrag) > KIKleineAuftragsgrenze {
		return false
	}

	for _, feld := range felder {
		if !kleineTextfelder[feld] {
			return false
		}
	}

	return true
}

// KIAnbieterlage ist, was die Wahl über einen Anbieter wissen muss.
type KIAnbieterlage struct {
	Name string
	Ort  KIOrt

	// Bereit heißt: eingeschaltet und eingerichtet.
	Bereit bool

	// Geraet heißt: das Modell des Betriebssystems — es nimmt nur kleine
	// Aufgaben.
	Geraet bool
}

// KIVersuch ist ein Anbieter in der Reihe samt dem Weg, auf dem er drankam.
type KIVersuch struct {
	Anbieter string
	Weg      KIWeg
}

// KIReihenfolge gibt, wen ein Vorschlag in welcher Reihenfolge fragt.
func KIReihenfolge(gewaehlt string, klein bool, lagen []KIAnbieterlage) []KIVersuch {
	if gewaehlt != "" && gewaehlt != KIVorrang {
		return []KIVersuch{{Anbieter: gewaehlt, Weg: KIWegGewaehlt}}
	}

	var reihe []KIVersuch

	for _, name := range KIRueckfaelle("", klein, lagen) {
		weg := KIWegRueckfall
		if len(reihe) == 0 {
			weg = KIWegVorrang
		}

		reihe = append(reihe, KIVersuch{Anbieter: name, Weg: weg})
	}

	return reihe
}

// KIRueckfaelle nennt die Anbieter, die nach `ausser` gefragt werden dürfen
// — das Gerät zuerst, wenn die Aufgabe klein ist, dann die lokalen, **nie**
// die Cloud.
func KIRueckfaelle(ausser string, klein bool, lagen []KIAnbieterlage) []string {
	var namen []string

	if klein {
		for _, lage := range lagen {
			if lage.Geraet && lage.Bereit && lage.Name != ausser {
				namen = append(namen, lage.Name)
			}
		}
	}

	for _, lage := range lagen {
		if !lage.Geraet && lage.Bereit && lage.Ort == KIOrtLokal && lage.Name != ausser {
			namen = append(namen, lage.Name)
		}
	}

	return namen
}

// PruefeVorrangwunsch weist ein Modell ab, das ohne Anbieter genannt ist —
// es wäre nicht eindeutig, denn zwei Anbieter dürfen denselben Namen führen.
func PruefeVorrangwunsch(alias string) error {
	if alias != "" {
		return fmt.Errorf("%w: mit dem Vorrang wählt das Programm Anbieter und "+
			"Modell — %q passt nur zu einem genannten Anbieter", ErrValidation, alias)
	}

	return nil
}
