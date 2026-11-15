package domain

// **Wogegen eine eigene Übersetzung geprüft wird** — LOC-02.AK-01 bis AK-03
// (#1688).
//
// Die Vorlage ist der **deutsche** Katalog des Fensters, flach: Pfad → Satz.
// Deutsch, weil es die Ausgangssprache ist (§9) — jeder andere Katalog ist
// eine Übersetzung davon, und eine Übersetzung einer Übersetzung prüfte
// gegen etwas, das selbst falsch sein kann.

import (
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Katalogvorlage ist der deutsche Katalog, flach: Pfad → Satz.
type Katalogvorlage map[string]string

// Sprachabgleich ist, was ein Paket gegenüber der Vorlage trägt.
type Sprachabgleich struct {
	// Gesamt ist die Zahl der Sätze der Vorlage, Uebersetzt die der
	// übersetzten, deren Quelle noch stimmt.
	Gesamt     int
	Uebersetzt int

	// Fehlend sind Pfade ohne Übersetzung — sie zeigen die Basis.
	Fehlend []string

	// Veraltet sind Pfade, deren deutscher Satz sich seit der Übersetzung
	// geändert hat. **Sie bleiben stehen** und werden markiert (AK-03).
	Veraltet []string

	// Unbekannt sind Pfade, die diese Fassung nicht kennt — aus einer
	// neueren, oder aus einer, in der es sie noch gab. **Sie bleiben im
	// Paket**; wer es weitergibt, verliert sie nicht.
	Unbekannt []string
}

// hoechstensZeichen ist die längste Übersetzung eines Satzes. Der längste
// deutsche Satz hatte am 2026-09-29 rund 1.300 Zeichen.
const hoechstensZeichen = 4000

// platzhalter findet `{name}` — dieselbe Gestalt wie in `einsetzen.ts`.
var platzhalter = regexp.MustCompile(`\{[A-Za-z]+\}`)

// unsicher findet, was in einem Satz kein Text ist: ein Tag oder Kommentar,
// eine Skript- oder Datenadresse.
var unsicher = regexp.MustCompile(`(?i)<[a-z/!?]|javascript:|vbscript:|data:`)

// wenige ist die dritte Zählform — LOC-01, `zaehlwort.ts`.
const wenige = ".wenige"

// Quelle gibt den deutschen Satz zu einem Pfad — bei der dritten Zählform
// den von `viele`, leer für einen unbekannten Pfad.
func (v Katalogvorlage) Quelle(pfad string) string {
	if satz, da := v[pfad]; da {
		return satz
	}

	if stamm, ist := strings.CutSuffix(pfad, wenige); ist {
		_, eine := v[stamm+".eine"]
		if viele, da := v[stamm+".viele"]; da && eine {
			return viele
		}
	}

	return ""
}

// PruefeUebersetzung urteilt über einen übersetzten Satz.
//
// **Der Befund steht am Pfad** — dort zeigt ihn der Assistent (§6).
func (v Katalogvorlage) PruefeUebersetzung(pfad, text string) error {
	werte := map[string]string{"pfad": pfad}

	quelle := v.Quelle(pfad)
	if quelle == "" {
		return NewFieldError(pfad, SchluesselValidierung,
			"%q gibt es im Katalog dieser Fassung nicht.", pfad).MitWerten(werte)
	}

	if fehler := pruefeText(pfad, text, ausgewogen(quelle)); fehler != nil {
		return fehler
	}

	erwartet := namenDerPlatzhalter(quelle)
	if !slices.Equal(namenDerPlatzhalter(text), erwartet) {
		werte["erwartet"] = strings.Join(erwartet, " ")

		return NewFieldError(pfad, SchluesselPlatzhalterAbweichend,
			"Diese Übersetzung nennt andere Platzhalter als ihre Vorlage: %s.",
			werte["erwartet"]).MitWerten(werte)
	}

	return nil
}

// PruefeFremdeUebersetzung urteilt über einen Satz unter einem Pfad, den
// diese Fassung nicht kennt — er stammt aus einer neueren. **Ohne Vorlage
// bleibt nur die Frage, ob er Text ist**; Platzhalter prüft die Fassung, die
// ihn kennt.
func PruefeFremdeUebersetzung(pfad, text string) error {
	return pruefeText(pfad, text, true)
}

// pruefeText urteilt über Länge und Auszeichnung eines Satzes.
func pruefeText(pfad, text string, schliessend bool) error {
	werte := map[string]string{"pfad": pfad}

	if strings.TrimSpace(text) == "" || utf8.RuneCountInString(text) > hoechstensZeichen {
		return NewFieldError(pfad, SchluesselValidierung,
			"Eine Übersetzung hat zwischen einem und %d Zeichen.", hoechstensZeichen).
			MitWerten(werte)
	}

	if !sicher(text) || ausgewogen(text) != schliessend {
		return NewFieldError(pfad, SchluesselAuszeichnungUnsicher,
			"Diese Übersetzung trägt, was kein Text ist, oder eine Auszeichnung, die nicht schließt.").
			MitWerten(werte)
	}

	return nil
}

// Abgleich stellt ein Paket neben die Vorlage.
func (v Katalogvorlage) Abgleich(p Sprachpaket) Sprachabgleich {
	abgleich := Sprachabgleich{
		Gesamt: len(v), Fehlend: []string{}, Veraltet: []string{}, Unbekannt: []string{},
	}

	for pfad, satz := range v {
		switch text, da := p.Texte[pfad]; {
		case !da:
			abgleich.Fehlend = append(abgleich.Fehlend, pfad)
		case text.Quelle != satz:
			abgleich.Veraltet = append(abgleich.Veraltet, pfad)
		default:
			abgleich.Uebersetzt++
		}
	}

	for pfad, text := range p.Texte {
		quelle := v.Quelle(pfad)

		switch {
		case quelle == "":
			abgleich.Unbekannt = append(abgleich.Unbekannt, pfad)
		case strings.HasSuffix(pfad, wenige) && text.Quelle != quelle:
			abgleich.Veraltet = append(abgleich.Veraltet, pfad)
		}
	}

	// **Geordnet**, damit zwei Läufe über dasselbe Paket dasselbe sagen.
	slices.Sort(abgleich.Fehlend)
	slices.Sort(abgleich.Veraltet)
	slices.Sort(abgleich.Unbekannt)

	return abgleich
}

// namenDerPlatzhalter gibt die Platzhalter eines Satzes, jeden einmal und
// geordnet. **Zweimal derselbe ist erlaubt** — eine Sprache darf einen Wert
// wiederholen, wo die andere ihn einmal nennt.
func namenDerPlatzhalter(satz string) []string {
	namen := platzhalter.FindAllString(satz, -1)
	slices.Sort(namen)

	return slices.Compact(namen)
}

// sicher sagt, ob ein Satz nur Text ist — Zeilenumbruch und Tabulator
// eingeschlossen, jedes andere Steuerzeichen nicht.
func sicher(satz string) bool {
	if unsicher.MatchString(satz) {
		return false
	}

	return !strings.ContainsFunc(satz, func(r rune) bool {
		return unicode.IsControl(r) && r != '\n' && r != '\t'
	})
}

// ausgewogen sagt, ob Fettdruck und Code in einem Satz schließen.
func ausgewogen(satz string) bool {
	return strings.Count(satz, "**")%2 == 0 && strings.Count(satz, "`")%2 == 0
}
