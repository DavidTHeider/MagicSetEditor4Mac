package msescript

import (
	"fmt"
	"strings"
)

// **Jede Meldung trägt einen Schlüssel** —
// #1178.
//
// # Der Befund
//
// Seit #1164
// erreicht der Satz des Zerteilers das Fenster: `skriptFehler` trägt ihn als
// Platzhalter, und über dem angestrichenen Ausdruck steht „Der Ausdruck ist
// nicht auswertbar: etwas fehlt" statt „Es ist ein unerwarteter Fehler
// aufgetreten".
//
// **Der eingesetzte Satz blieb aber deutsch, auch im englischen Katalog.**
//
// # Warum ein eigener Typ und kein `domain.Schluessel`
//
// **`internal/msescript` kennt nichts aus dem Projekt** (§2), und das ist
// Absicht: es ist ein Sprachwerkzeug, und eine Kartenentität darin wäre eine
// Vermischung, die den Interpreter unprüfbar machte. Ein eigener Schlüsseltyp
// hier fasst die Regel nicht an — das Issue nennt ihn ausdrücklich als den
// Weg.
//
// # Warum die deutschen Sätze hierbleiben
//
// **§9: Fehlermeldungen sind deutsch.** Kommandozeile, MCP und der Dienst
// bekommen den Satz, wie er hier steht; das Fenster bekommt zusätzlich den
// Schlüssel und setzt seinen eigenen Katalogtext ein. Zwei Wege, ein Satz —
// und der deutsche bleibt der Rückfall, wenn ein Schlüssel im Katalog fehlt.

// Meldungsschluessel benennt eine Meldung des Zerteilers oder Auswerters.
//
// **Als Zeichenkette und nicht als Zahl**, aus demselben Grund wie bei [Art]:
// eine fehlgeschlagene Zusicherung meldet sonst „erwartet 3, bekam 7".
type Meldungsschluessel string

const (
	// Der Lexer.
	SchluesselFremdesZeichen     Meldungsschluessel = "fremdesZeichen"
	SchluesselEinzelnesGleich    Meldungsschluessel = "einzelnesGleich"
	SchluesselOffeneZeichenkette Meldungsschluessel = "offeneZeichenkette"

	// SchluesselZuTiefGeschachtelt ist eine Schachtelung über
	// [HoechsteSchachtelung] (BK-08, #2134) — beim Zerteilen, nicht beim
	// Auswerten ([SchluesselZuTief]).
	SchluesselZuTiefGeschachtelt Meldungsschluessel = "zuTiefGeschachtelt"

	// Der Zerteiler und das Skriptnetz.
	SchluesselKeinSkript  Meldungsschluessel = "keinSkript"
	SchluesselSkriptkreis Meldungsschluessel = "skriptkreis"
	SchluesselKeinVerweis Meldungsschluessel = "keinVerweis"

	// Der Auswerter.
	SchluesselKeineZahl          Meldungsschluessel = "keineZahl"
	SchluesselFormUnbekannt      Meldungsschluessel = "formUnbekannt"
	SchluesselNameUnbekannt      Meldungsschluessel = "nameUnbekannt"
	SchluesselKeineFunktion      Meldungsschluessel = "keineFunktion"
	SchluesselOhneRumpf          Meldungsschluessel = "ohneRumpf"
	SchluesselNichtWahrheitswert Meldungsschluessel = "nichtWahrheitswert"
	SchluesselVorzeichenZahl     Meldungsschluessel = "vorzeichenZahl"
	SchluesselZaehlschleifeZahl  Meldungsschluessel = "zaehlschleifeZahl"
	SchluesselRechnenLinks       Meldungsschluessel = "rechnenLinks"
	SchluesselRechnenRechts      Meldungsschluessel = "rechnenRechts"
	SchluesselVergleichLinks     Meldungsschluessel = "vergleichLinks"
	SchluesselVergleichRechts    Meldungsschluessel = "vergleichRechts"
	SchluesselKeineListe         Meldungsschluessel = "keineListe"
	SchluesselZuTief             Meldungsschluessel = "zuTief"
	SchluesselSchleifeSteht      Meldungsschluessel = "schleifeSteht"
	SchluesselFunktionVerbinden  Meldungsschluessel = "funktionVerbinden"

	// Der Zerteiler — was an einer Stelle fehlt.
	//
	// **Zwei Schlüssel für zwanzig Sätze.** Sie sind alle von der Form „nach X
	// fehlt Y" oder „hier fehlt X"; ein Schlüssel je Satz wäre zwanzigmal
	// dieselbe Regel mit einem anderen Wort darin, und beide Kataloge müssten
	// sie zwanzigmal tragen.
	SchluesselFehltNach Meldungsschluessel = "fehltNach"
	SchluesselFehltHier Meldungsschluessel = "fehltHier"

	// SchluesselKlammerOffen heißt: hier beginnt eine Klammer, die niemand
	// schließt — und die Stelle ist die **öffnende**.
	//
	// **Ein eigener Schlüssel und nicht `fehltHier`**, weil er eine andere
	// Stelle trägt. Die anderen zwanzig Sätze zeigen dorthin, wo der
	// Zerteiler stehen blieb; dieser zeigt dorthin, wo jemand etwas zu tun
	// hat. Am Quellende gemeldet sagt eine fehlende Klammer nichts: sie steht
	// irgendwo darüber.
	SchluesselKlammerOffen Meldungsschluessel = "klammerOffen"

	// SchluesselAusdruckGehtWeiter heißt: der Ausdruck ist zu Ende, der Text
	// nicht.
	SchluesselAusdruckGehtWeiter Meldungsschluessel = "ausdruckGehtWeiter"

	// SchluesselZuweisungBrauchtNamen heißt: links von `:=` steht kein Name.
	SchluesselZuweisungBrauchtNamen Meldungsschluessel = "zuweisungBrauchtNamen"

	// SchluesselDurchgereicht trägt einen Satz, der schon fertig ist.
	//
	// **Er kommt aus der Standardbibliothek** — `break_text`, `filter_list`
	// und ihre 46 Geschwister melden ihre eigenen Sätze, und die nennen
	// Funktionsnamen und Argumentnamen. Ein Katalogeintrag je Funktion wäre
	// eine zweite Abschrift der Standardbibliothek in zwei Sprachen.
	//
	// **Das ist eine Auslassung mit Grund und kein Versehen.** Die Abnahme von
	// #1178
	// verlangt, dass ein **Syntaxfehler** im englischen Fenster englisch
	// steht; ein Laufzeitfehler einer Bibliotheksfunktion ist keiner. Das
	// Fenster zeigt dann den deutschen Satz — der Rückfall, den es ohnehin
	// braucht.
	SchluesselDurchgereicht Meldungsschluessel = "durchgereicht"

	// **Die sieben, die bis QA-03 einen Satz statt eines Schlüssels trugen**
	// (#1656) — und deshalb »hier steht %s« zeigten, ohne den Wert.
	SchluesselBedingungWahrheit Meldungsschluessel = "bedingungWahrheit"
	SchluesselFuerJedenListe    Meldungsschluessel = "fuerJedenListe"
	SchluesselWahrheitLinks     Meldungsschluessel = "wahrheitLinks"
	SchluesselWahrheitRechts    Meldungsschluessel = "wahrheitRechts"
	SchluesselBindenFunktion    Meldungsschluessel = "bindenFunktion"
)

// meldungsmuster gibt zu jedem Schlüssel seinen deutschen Satz.
//
// **Hier und nicht verteilt.** Ein Muster neben seiner Fundstelle wäre eine
// zweite Liste, und der Vertragstest gegen die Kataloge des Fensters läse
// dann den halben Bestand.
var meldungsmuster = map[Meldungsschluessel]string{
	SchluesselFremdesZeichen:     "das Zeichen %s gehört nicht zur Sprache",
	SchluesselEinzelnesGleich:    "ein einzelnes `=`; zugewiesen wird mit `:=`, verglichen mit `==`",
	SchluesselOffeneZeichenkette: "eine Zeichenkette ohne schließendes Anführungszeichen",
	SchluesselZuTiefGeschachtelt: "der Ausdruck ist tiefer geschachtelt als %s Ebenen",

	SchluesselKeinSkript:  "es gibt kein Skript namens %s",
	SchluesselSkriptkreis: "diese Felder lesen sich im Kreis: %s",
	SchluesselKeinVerweis: "mit `+` lässt sich kein Verweis auf ein anderes Skript verketten",

	SchluesselKeineZahl:          "%s ist keine Zahl",
	SchluesselFormUnbekannt:      "die Form %s lässt sich nicht auswerten",
	SchluesselNameUnbekannt:      "der Name %s ist hier nicht bekannt",
	SchluesselKeineFunktion:      "hier wird %s gerufen, das geht nur bei einer Funktion",
	SchluesselOhneRumpf:          "diese Funktion hat keinen Rumpf",
	SchluesselNichtWahrheitswert: "`not` braucht einen Wahrheitswert, hier steht %s",
	SchluesselVorzeichenZahl:     "ein Vorzeichen braucht eine Zahl, hier steht %s",
	SchluesselZaehlschleifeZahl:  "%s einer Zählschleife braucht eine Zahl, hier steht %s",
	SchluesselRechnenLinks:       "`%s` rechnet mit Zahlen, links %s",
	SchluesselRechnenRechts:      "`%s` rechnet mit Zahlen, rechts %s",
	SchluesselVergleichLinks:     "`%s` vergleicht Zahlen, links %s",
	SchluesselVergleichRechts:    "`%s` vergleicht Zahlen, rechts %s",
	SchluesselKeineListe:         "sort_list ordnet eine Liste, hier steht %s",
	SchluesselZuTief:             "die Auswertung geht mehr als %s Stufen tief — ruft sich hier etwas selbst?",
	SchluesselSchleifeSteht:      "eine Zählschleife kommt bei %s nicht voran — die Zahl ist zu groß",
	SchluesselFunktionVerbinden:  "eine Funktion lässt sich nur mit einer Funktion verbinden, hier steht %s und %s",

	SchluesselFehltNach:             "nach %s fehlt %s",
	SchluesselFehltHier:             "hier fehlt %s",
	SchluesselKlammerOffen:          "diese %s wird nicht geschlossen — %s fehlt",
	SchluesselAusdruckGehtWeiter:    "hier hört der Ausdruck auf, aber der Text geht weiter",
	SchluesselZuweisungBrauchtNamen: "zugewiesen wird an einen Namen",

	SchluesselDurchgereicht: "%s",

	SchluesselBedingungWahrheit: "eine Bedingung braucht einen Wahrheitswert, hier steht %s",
	SchluesselFuerJedenListe:    "`for each` geht eine Liste durch, hier steht %s",
	SchluesselWahrheitLinks:     "`%s` braucht Wahrheitswerte, links steht %s",
	SchluesselWahrheitRechts:    "`%s` braucht Wahrheitswerte, rechts steht %s",
	SchluesselBindenFunktion:    "`@` bindet Argumente an eine Funktion, hier steht %s",
}

// Meldungsschluesselliste gibt alle Schlüssel — für den Vertragstest gegen
// die Kataloge des Fensters.
func Meldungsschluesselliste() []Meldungsschluessel {
	alle := make([]Meldungsschluessel, 0, len(meldungsmuster))
	for schluessel := range meldungsmuster {
		alle = append(alle, schluessel)
	}

	return alle
}

// satzZu setzt den deutschen Satz eines Schlüssels zusammen.
func satzZu(schluessel Meldungsschluessel, werte []string) string {
	muster, gefuehrt := meldungsmuster[schluessel]
	if !gefuehrt {
		// **Ein unbekannter Schlüssel gibt seinen Namen.** Er kommt nicht vor —
		// die Aufzählung und die Tabelle stehen nebeneinander —, und ein
		// leerer Satz wäre die schlechtere Antwort.
		return string(schluessel)
	}

	teile := make([]any, len(werte))
	for i := range werte {
		teile[i] = werte[i]
	}

	return fmt.Sprintf(muster, teile...)
}

// PlatzhalterZahl gibt, wie viele Werte ein Satz einsetzt.
//
// **Für den Vertragstest gegen die Kataloge des Fensters.** `nach %s fehlt %s`
// braucht zwei; ein Katalogsatz mit nur einem zeigte den zweiten nie — und
// niemand merkte es, weil der Text lesbar bleibt.
func PlatzhalterZahl(schluessel Meldungsschluessel) int {
	return strings.Count(meldungsmuster[schluessel], "%s")
}
