package domain

// **Die Zustandsmatrix eines Handels** — Werkstück `T1` von Phase 5.6
// (#1466).
//
// # Warum eine Tabelle und kein `switch`
//
// **Weil die Matrix eine Aussage ist und kein Ablauf.** Wer sie als
// Verzweigung schriebe, könnte sie nicht vorzeigen — und genau das verlangt
// das Issue: erlaubter Akteur, Voraussetzung, Persistenz und
// Wiederholungsantwort **je Feld**. Ein Wächter zählt sie deshalb nach
// (`handelsmatrix_test.go`): jeder Zustand wird erreicht, und aus jedem führt
// ein Weg heraus — oder er ist ausdrücklich ein Ende.
//
// # Die Wiederholung ist Teil der Matrix und kein Sonderfall
//
// **Über eine Leitung geht eine Nachricht verloren und kommt noch einmal.**
// Ein zweites „ich stimme zu" darf keine zweite Zustimmung werden und keinen
// Fehler geben; es bekommt **dieselbe Antwort** (§7, Regel 4). Das steht
// nicht neben der Matrix, sondern in [Handelsangebot.Schreite].

import "slices"

// Handelnder sagt, wer einen Schritt gehen darf.
type Handelnder string

const (
	// HandelnderInitiator ist, wer den Vorgang eröffnet hat.
	HandelnderInitiator Handelnder = "initiator"

	// HandelnderGegenseite ist jede andere Partei.
	HandelnderGegenseite Handelnder = "gegenseite"

	// HandelnderJede ist jede Partei des Vorgangs.
	HandelnderJede Handelnder = "jede"
)

// Handelsschritt ist eine Handlung an einem Vorgang.
type Handelsschritt string

const (
	// SchrittAnbieten schickt den Entwurf hinaus.
	SchrittAnbieten Handelsschritt = "anbieten"

	// SchrittGegenangebot antwortet mit einer eigenen Revision.
	SchrittGegenangebot Handelsschritt = "gegenangebot"

	// SchrittZustimmen bindet **diese** Revision.
	SchrittZustimmen Handelsschritt = "zustimmen"

	// SchrittVorbereiten hält die Exemplare fest (T2).
	SchrittVorbereiten Handelsschritt = "vorbereiten"

	// SchrittUebergeben ist der Gang zum Tisch — von Hand zu Hand.
	//
	// **Beide bestätigen ihn** (T6, #1471).
	// Solange eine Bestätigung fehlt, hat einer Karten weggegeben und der
	// andere weiß nichts davon; ein Zustand, den eine Seite allein
	// herstellte, hieße „übergeben" und meinte „ich sage, dass ich übergeben
	// habe". Das Issue verlangt vor dem Abschluss ausdrücklich „lokale **und
	// empfangene** Übergabebestätigung".
	//
	// **Bis zum 2026-09-18 genügte eine Stimme** — angestrichen beim Bau von
	// `T6`, dem ersten Werkstück, das aus diesem Zustand heraus etwas tut.
	SchrittUebergeben Handelsschritt = "uebergeben"

	// SchrittVerbuchen schreibt **diesen** Bestand fort (T6).
	SchrittVerbuchen Handelsschritt = "verbuchen"

	// SchrittAbgleichen wartet auf die Gegenbestätigung.
	SchrittAbgleichen Handelsschritt = "abgleichen"

	// SchrittBestaetigen ist der einzige Weg nach „beidseitig verbucht".
	SchrittBestaetigen Handelsschritt = "bestaetigen"

	// SchrittAblehnen beendet ihn von der Gegenseite her.
	SchrittAblehnen Handelsschritt = "ablehnen"

	// SchrittAbbrechen beendet ihn von der eigenen Seite her.
	SchrittAbbrechen Handelsschritt = "abbrechen"

	// SchrittAblaufen hält fest, dass die Frist um ist.
	//
	// **Die Frist rechnet niemand hier aus** — wann sie um ist, entscheidet
	// T5 (#1470).
	// Diese Matrix sagt nur, dass es den Zustand gibt und wie man ihn
	// erreicht.
	SchrittAblaufen Handelsschritt = "ablaufen"

	// SchrittKlaeren nimmt einen Vorgang aus dem Ablauf heraus.
	SchrittKlaeren Handelsschritt = "klaeren"
)

// Handelsuebergang ist ein Feld der Zustandsmatrix.
type Handelsuebergang struct {
	Schritt Handelsschritt
	Von     Handelszustand
	Nach    Handelszustand
	Wer     Handelnder

	// Alle sagt, dass der Zustand erst wechselt, wenn **jede** Partei den
	// Schritt getan hat.
	//
	// **Eine einzelne Zustimmung bewegt nichts**, ein einzelner Abbruch
	// bewegt alles — das ist der Unterschied, den dieses Feld trägt.
	Alle bool
}

// Handelsmatrix gibt die Übergänge, die es gibt.
//
// **Die Reihenfolge ist der Weg**: erst der übliche, dann die Ausgänge.
func Handelsmatrix() []Handelsuebergang {
	return []Handelsuebergang{
		{SchrittAnbieten, HandelEntwurf, HandelAngeboten, HandelnderInitiator, false},
		{SchrittGegenangebot, HandelAngeboten, HandelVerhandlung, HandelnderGegenseite, false},
		{SchrittGegenangebot, HandelVerhandlung, HandelVerhandlung, HandelnderJede, false},
		{SchrittZustimmen, HandelAngeboten, HandelAngenommen, HandelnderJede, true},
		{SchrittZustimmen, HandelVerhandlung, HandelAngenommen, HandelnderJede, true},
		{SchrittVorbereiten, HandelAngenommen, HandelVorbereitet, HandelnderInitiator, false},
		{SchrittUebergeben, HandelVorbereitet, HandelUebergabeOffen, HandelnderJede, true},
		{SchrittVerbuchen, HandelUebergabeOffen, HandelLokalVerbucht, HandelnderJede, false},
		{SchrittAbgleichen, HandelLokalVerbucht, HandelAbgleichOffen, HandelnderJede, false},
		{SchrittBestaetigen, HandelAbgleichOffen, HandelBeidseitigVerbucht, HandelnderJede, true},

		{SchrittAblehnen, HandelAngeboten, HandelAbgelehnt, HandelnderGegenseite, false},
		{SchrittAblehnen, HandelVerhandlung, HandelAbgelehnt, HandelnderJede, false},
		{SchrittAbbrechen, HandelEntwurf, HandelAbgebrochen, HandelnderInitiator, false},
		{SchrittAbbrechen, HandelAngeboten, HandelAbgebrochen, HandelnderInitiator, false},
		// **Und nach der Einigung ein Ausweg für jede Partei** (TRADE-03,
		// #1717): ein Gegenüber, das verschwindet, hielte sonst Vorgang und
		// Zusagen fest. Nach begonnener Übergabe ist es die Klärung — die
		// Karten können schon unterwegs sein.
		{SchrittAbbrechen, HandelAngenommen, HandelAbgebrochen, HandelnderJede, false},
		{SchrittAbbrechen, HandelVorbereitet, HandelAbgebrochen, HandelnderJede, false},
		{SchrittKlaeren, HandelUebergabeOffen, HandelKlaerung, HandelnderJede, false},
		{SchrittAbbrechen, HandelKlaerung, HandelAbgebrochen, HandelnderJede, false},
		{SchrittAblaufen, HandelAngeboten, HandelAbgelaufen, HandelnderJede, false},
		{SchrittKlaeren, HandelLokalVerbucht, HandelKlaerung, HandelnderJede, false},
		{SchrittKlaeren, HandelAbgleichOffen, HandelKlaerung, HandelnderJede, false},
	}
}

// HandelEnde sagt, ob aus diesem Zustand nichts mehr folgt.
//
// **Vier Enden und nicht eines**: „beidseitig verbucht" ist das gute, die
// drei anderen sind die verschiedenen Arten, es nicht zu werden — und sie
// sind unterscheidbar, weil sie verschiedene Antworten auf „warum" geben.
func HandelEnde(zustand Handelszustand) bool {
	return slices.Contains([]Handelszustand{
		HandelBeidseitigVerbucht, HandelAbgelehnt,
		HandelAbgebrochen, HandelAbgelaufen,
	}, zustand)
}

// Abschlussschritte sind die drei, die mehr bewegen als den Zustand.
//
// # Warum sie benannt werden müssen
//
// **Weil die Zustandsmaschine sie gehen könnte, ohne dass geschieht, wofür es
// sie gibt.** `verbuchen` bewegt den **Bestand**, `bestaetigen` zeichnet den
// **Beleg** gegen, und `abgleichen` steht zwischen beiden. Wer einen davon
// über die bloße Matrix ginge, bekäme einen Vorgang, der `lokalVerbucht`
// heißt, ohne dass eine Karte den Besitzer gewechselt hat — oder einen Beleg,
// der einseitig bleibt, während der Zustand `beidseitigVerbucht` sagt.
//
// **Gegangen werden sie deshalb nur über die Verbuchungsstelle** (T6,
// #1471);
// die Leitung aus `T1` weist sie ab
// (#1637).
//
// **Als Funktion und nicht als Variable** — dieselbe Überlegung wie bei
// [Handelsmatrix]: eine Aufzählung, die sich ändern lässt, ist keine.
func Abschlussschritte() []Handelsschritt {
	return []Handelsschritt{
		SchrittVerbuchen, SchrittAbgleichen, SchrittBestaetigen,
	}
}

// IstAbschlussschritt sagt, ob dieser Schritt zur Verbuchung gehört.
func IstAbschlussschritt(schritt Handelsschritt) bool {
	return slices.Contains(Abschlussschritte(), schritt)
}

// Handelsschritte gibt jeden Schritt, den die Matrix kennt — einmal.
//
// # Warum aus der Matrix erhoben und nicht danebengeschrieben
//
// **Weil eine zweite Liste beim nächsten Übergang still zu kurz wäre.** Sie
// wird an den Flächen gebraucht: ein unbekannter Schritt ist dort eine
// **Beanstandung** und kein Konflikt — „das gibt es nicht" und „von hier aus
// geht das nicht" sind zwei Antworten, und wer sie zusammenlegte, ließe einen
// Klienten wiederholen, bis der Zustand passt ([Handelsangebot.Schreite]).
//
// **In der Reihenfolge des Weges**, weil die Matrix sie so führt.
func Handelsschritte() []Handelsschritt {
	schritte := make([]Handelsschritt, 0, len(Handelsmatrix()))

	for _, uebergang := range Handelsmatrix() {
		if !slices.Contains(schritte, uebergang.Schritt) {
			schritte = append(schritte, uebergang.Schritt)
		}
	}

	return schritte
}

// IstHandelsschritt sagt, ob es diesen Schritt gibt.
func IstHandelsschritt(schritt Handelsschritt) bool {
	return slices.Contains(Handelsschritte(), schritt)
}
