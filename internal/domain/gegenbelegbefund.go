package domain

// **Ein Gegenbeleg, der nicht passt, wird festgehalten** —
// #2500.
//
// # Warum es das gibt
//
// **Weil er bis dahin still blieb.** Passte ein Beleg der Gegenseite nicht,
// endete die Zustellung mit einem Fehler: beim Absender stand weiter
// „hinausgegangen", beim Empfänger nichts, und beide Fenster sagten
// dauerhaft „Nur du hast unterschrieben". Jetzt steht am eigenen Beleg,
// **dass** einer kam, **woran** er scheiterte und **von wem**.
//
// # Warum der Mangel ein Wort ist und der Grund ein Satz
//
// **Das Wort übersetzt das Fenster**, der Satz ist für das Protokoll und für
// die Kommandozeile (Vermerk 0075): „der Beleg trägt 2 Zeilen und meiner 3"
// sagt einem Menschen am Fenster in seiner Sprache nichts, wenn er es nicht
// spricht — „die Zeilen weichen ab" schon.

import (
	"fmt"
	"slices"
	"time"
)

// Gegenbelegmangel ist, woran ein Gegenbeleg scheiterte.
type Gegenbelegmangel string

// Die fünf Mängel — einer je Prüfung, in der Reihenfolge, in der geprüft
// wird ([Belegurkunde.PasstZu] und davor die Unterschrift).
const (
	MangelUnterschrift Gegenbelegmangel = "unterschrift"
	MangelVorgang      Gegenbelegmangel = "vorgang"
	MangelRevision     Gegenbelegmangel = "revision"
	MangelParteien     Gegenbelegmangel = "parteien"
	MangelZeilen       Gegenbelegmangel = "zeilen"
)

// DieGegenbelegmaengel sind alle fünf — das Schema hält dieselbe Liste
// (`handelsgegenbelegbefunde`, docs/INVARIANTEN.md).
var DieGegenbelegmaengel = []Gegenbelegmangel{
	MangelUnterschrift, MangelVorgang, MangelRevision, MangelParteien, MangelZeilen,
}

// Gegenbelegbefund hält fest, dass ein Gegenbeleg abgewiesen wurde.
type Gegenbelegbefund struct {
	// Vorgang ist der **hiesige** Handel, gegen dessen Beleg geprüft wurde.
	Vorgang string

	// Von ist die Peerkennung des Absenders — aus der Leitung, wie bei
	// [Fremdbeleg.Von].
	Von string

	// Abdruck ist der, den der abgewiesene Beleg **nannte** — eine
	// Behauptung, und deshalb kein Fremdschlüssel.
	Abdruck string

	Mangel Gegenbelegmangel

	// Grund ist der Satz der Prüfung, für das Protokoll.
	Grund string

	// Empfangen ist die eigene Uhr beim Abweisen.
	Empfangen time.Time
}

// Pruefe sagt, ob der Befund vollständig ist.
func (b Gegenbelegbefund) Pruefe() error {
	switch {
	case b.Vorgang == "":
		return NewFieldError("vorgang", SchluesselValidierung, "Ein Befund nennt den Handel, zu dem er gehört.")
	case b.Von == "":
		return NewFieldError("von", SchluesselValidierung, "Ein Befund nennt, von wem der Gegenbeleg kam.")
	case !slices.Contains(DieGegenbelegmaengel, b.Mangel):
		return NewFieldError("mangel", SchluesselValidierung, "%q ist kein Mangel eines Gegenbelegs.", b.Mangel)
	case b.Grund == "":
		return NewFieldError("grund", SchluesselValidierung, "Ein Befund nennt seinen Grund.")
	}

	return nil
}

// Belegabweichung sagt, woran ein Gegenbeleg scheiterte — **der Mangel für
// das Fenster, der Satz für das Protokoll** (#2500).
//
// **Sie bleibt eine Validierung**: wer nur `errors.Is(…, ErrValidation)`
// fragt, fragt weiter richtig.
type Belegabweichung struct {
	Mangel Gegenbelegmangel
	Satz   string
}

// abweichung baut eine mit ihrem Satz.
func abweichung(mangel Gegenbelegmangel, satz string, werte ...any) *Belegabweichung {
	return &Belegabweichung{Mangel: mangel, Satz: fmt.Sprintf(satz, werte...)}
}

func (a *Belegabweichung) Error() string { return ErrValidation.Error() + ": " + a.Satz }

// Unwrap macht sie zu einer Validierung.
func (a *Belegabweichung) Unwrap() error { return ErrValidation }
