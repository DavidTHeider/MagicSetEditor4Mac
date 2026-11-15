// **Wie etwas aussieht, wo die Uhr steht und womit bezahlt wird** —
// LOC-03.AK-03
// (#1689).
//
// # Warum vier Einstellungen und nicht eine
//
// **Weil die Sprache nur eine davon ist.** Das Issue verlangt, dass Sprache,
// regionale Darstellung, Zeitzone und Währung **getrennt** einstellbar sind —
// und dass die volle Locale erhalten bleibt: aus `de-CH` darf nicht `de`
// werden, sonst sieht eine Schweizer Zahl aus wie eine deutsche.
//
// **Leer heißt „wie das System"**, bei allen dreien. Eine Vorgabe hier
// überschriebe, was der Rechner sagt, sobald jemand die Anwendung zum
// ersten Mal öffnet — derselbe Grund, aus dem die Sprache keine hat.

package domain

import (
	"strings"
	"time"

	"golang.org/x/text/language"
)

const (
	// EinstellungDarstellung ist die Locale, in der Zahlen, Beträge und
	// Zeitpunkte erscheinen — `de-CH`, `en-GB`, `ja-JP`.
	EinstellungDarstellung = "darstellung"

	// EinstellungZeitzone ist die Zone, in der ein Zeitpunkt gezeigt wird —
	// ein Name aus der Zonendatenbank wie `Europe/Berlin`.
	EinstellungZeitzone = "zeitzone"

	// VorgabeWaehrung ist die Währung, mit der eine neue Sammlungszeile
	// anfängt. **Umgerechnet wird nichts** (AK-02): sie steht vor, sie
	// rechnet nicht.
	VorgabeWaehrung = "vorgabe.waehrung"
)

// PruefeDarstellung urteilt über eine Locale.
//
// **Sie muss so geschrieben sein, wie sie gelesen wird.** `x/text` nimmt
// »de_CH« und »de-ch« entgegen, `Intl` im Fenster wirft bei »de_CH« einen
// `RangeError` — gemessen am 2026-09-22. Eine Einstellung, die das Fenster
// nicht versteht, wäre eine, die still nichts bewirkt; deshalb wird die
// kanonische Schreibweise **genannt** und nicht still übernommen (AK-03).
func PruefeDarstellung(wert string) error {
	if wert == "" {
		return nil
	}

	gelesen, fehler := language.Parse(wert)
	if fehler != nil {
		return NewFieldError("wert", SchluesselValidierung,
			"»%s« ist keine Locale — erwartet wird eine wie de-CH oder en-GB.",
			wert)
	}

	if kanonisch := gelesen.String(); kanonisch != wert {
		return NewFieldError("wert", SchluesselValidierung,
			"»%s« schreibt sich %s.", wert, kanonisch)
	}

	return nil
}

// PruefeZeitzone urteilt über den Namen einer Zeitzone.
//
// **Gegen die Zonendatenbank dieses Rechners** — dieselbe, aus der die
// Anzeige im Fenster liest. **»Local« ist kein Name**, sondern Gos Wort für
// „das System", und das System heißt hier: leer.
func PruefeZeitzone(wert string) error {
	if wert == "" {
		return nil
	}

	if wert == "Local" {
		return NewFieldError("wert", SchluesselValidierung,
			"»Local« ist kein Zonenname — für die Zone des Systems bleibt das Feld leer.")
	}

	if _, fehler := time.LoadLocation(wert); fehler != nil {
		return NewFieldError("wert", SchluesselValidierung,
			"»%s« ist keine Zeitzone — erwartet wird ein Name wie Europe/Berlin.",
			wert)
	}

	return nil
}

// PruefeWaehrungscode urteilt über einen Währungscode nach ISO 4217.
//
// **Klein geschrieben wird erklärt und nicht still großgeschrieben** — die
// Einstellung steht so da, wie sie gespeichert ist, und »eur« neben »EUR«
// wären zwei Schreibweisen für eine Sache.
func PruefeWaehrungscode(wert string) error {
	if wert == "" {
		return nil
	}

	if gross := strings.ToUpper(wert); gross != wert && istWaehrungscode(gross) {
		return NewFieldError("wert", SchluesselValidierung,
			"»%s« — ein Währungscode steht in Großbuchstaben: %s.", wert, gross)
	}

	if !istWaehrungscode(wert) {
		return NewFieldError("wert", SchluesselValidierung,
			"»%s« ist kein Währungscode — der hat drei Buchstaben, wie EUR.", wert)
	}

	return nil
}
