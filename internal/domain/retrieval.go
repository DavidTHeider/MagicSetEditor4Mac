package domain

// **Was ein indexierbares Dokument über sich weiß** — Werkstück A15 von
// Phase 5.3
// (#1514).
//
// # Warum so viele Felder
//
// **Weil ein Treffer auflösbar sein muss.** Das Ziel des Issues steht in
// einem Satz: „mit auflösbarer Herkunft". Wer einen Abschnitt findet, soll
// sagen können, aus welcher Quelle er stammt, in welchem Stand sie war, wann
// sie geholt wurde und welcher Zeichenbereich gemeint ist — sonst ist der
// Treffer eine Behauptung, und die Recherche in A16 baute darauf auf.
//
// # Und warum Abrufzeit und Quelldatum zwei Felder sind
//
// **Weil sie zwei verschiedene Dinge sagen.** Das Issue: „Abrufdatum ist
// nicht Gültigkeitsdatum." Eine Regel gilt seit dem 26. Juli; geladen wurde
// sie heute. Wer beides zusammenlegte, hätte einen Regelstand, der sich beim
// Nachladen ändert, ohne dass sich die Regel geändert hat.
//
// # Der Index ist abgeleitet
//
// **Er ersetzt keine Quelldatenbank** (Issue, erster Satz). Was hier steht,
// ist eine Kopie mit Herkunftsangabe; wer sie wegwirft, verliert nichts, was
// sich nicht wieder aufbauen ließe.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"time"
)

// Retrievalquelle sagt, woher ein Dokument stammt.
type Retrievalquelle string

// Die fünf Quellen des Pflichtkorpus.
//
// **Sie stehen einzeln da und nicht als „offiziell" zusammen.** Ein
// Oracle-Text, eine Regelstelle und ein Ruling haben verschiedene Stände,
// verschiedene Urheber und verschiedene Nutzungsbedingungen; eine gemeinsame
// Marke verlöre genau das.
const (
	// QuelleOracle ist der Regeltext einer Karte, wie Scryfall ihn führt.
	QuelleOracle Retrievalquelle = "oracle"

	// QuelleRuling ist eine Einzelfallauskunft — mit Urheber und Datum.
	QuelleRuling Retrievalquelle = "ruling"

	// QuelleRegelwerk ist ein Abschnitt der Comprehensive Rules.
	QuelleRegelwerk Retrievalquelle = "regelwerk"

	// QuelleEigeneKarte ist eine Karte aus dem eigenen Bestand.
	QuelleEigeneKarte Retrievalquelle = "eigenekarte"

	// QuelleEigenerText ist eine Setnotiz oder ein Settext.
	QuelleEigenerText Retrievalquelle = "eigenertext"
)

var retrievalquellen = []Retrievalquelle{
	QuelleOracle, QuelleRuling, QuelleRegelwerk,
	QuelleEigeneKarte, QuelleEigenerText,
}

// Retrievalquellen gibt die fünf in ihrer Reihenfolge.
func Retrievalquellen() []Retrievalquelle { return slices.Clone(retrievalquellen) }

// Retrievalsicht sagt, wer ein Dokument sehen darf.
type Retrievalsicht string

// Die zwei Sichtbarkeiten.
//
// **Zwei und nicht drei.** „Geteilt" gibt es nicht: was nicht öffentlich ist,
// gehört einem Profil, und wer es sehen darf, entscheidet die Autorisierung
// und nicht der Index.
const (
	SichtbarOeffentlich Retrievalsicht = "oeffentlich"
	SichtbarProfil      Retrievalsicht = "profil"
)

// Rulingurheber sagt, wer eine Einzelfallauskunft gegeben hat.
//
// **Drei Werte, und `unbekannt` ist einer davon.** Der vorhandene Adapter
// unterscheidet sie schon (`internal/scryfall/rulings.go`); hier bleiben sie
// erhalten, statt zu „offiziell" zusammenzufallen.
type Rulingurheber string

// Die drei Urheber.
const (
	UrheberWotc      Rulingurheber = "wotc"
	UrheberScryfall  Rulingurheber = "scryfall"
	UrheberUnbekannt Rulingurheber = "unbekannt"
)

var rulingurheber = []Rulingurheber{UrheberWotc, UrheberScryfall, UrheberUnbekannt}

// Retrievaldokument ist ein indexierbarer Textabschnitt samt Herkunft.
type Retrievaldokument struct {
	// Kennung benennt genau diesen Abschnitt — sie ist der Schlüssel im Index.
	Kennung string

	Quelle Retrievalquelle

	// Bezugsadresse ist, wo das Original steht — leer bei eigenen Inhalten.
	Bezugsadresse string

	Spiel string

	// Kennungssystem, Fachkennung und Semantik sind die Identität nach A2
	// ([Druckkennung]) — leer bei einer Regelstelle, die keine Karte nennt.
	//
	// **Kein Kartenname als Identität.** Das Issue sagt es scharf: „keine
	// Deduplizierung nach Namen und keine künstlichen Scryfall-IDs für
	// Eigenkarten."
	Kennungssystem Kennungssystem
	Fachkennung    string
	Semantik       Semantik

	// Regelnummer ist die Subnummer einer Regelstelle (`205.4a`).
	Regelnummer string

	// Urheber ist der Urheber eines Rulings.
	Urheber Rulingurheber

	Sprache string
	Titel   string
	Text    string

	// Quellrevision ist der Stand der Quelle, aus der dieser Abschnitt kommt.
	Quellrevision string

	// Quelldatum ist das **Gültigkeitsdatum** — nicht die Abrufzeit.
	Quelldatum string

	// Abgerufen ist, wann geholt wurde.
	Abgerufen time.Time

	// Inhaltshash ist der Abdruck des Textes (`sha256:…`).
	//
	// **Er ist kein Geheimnisschutz** (Issue). Er sagt, ob zwei Abschnitte
	// denselben Text tragen, und sonst nichts.
	Inhaltshash string

	// Abschnitt ist die laufende Nummer innerhalb des Dokuments.
	Abschnitt int

	// Von und Bis sind der Zeichenbereich im Quelltext.
	Von int
	Bis int

	// Merkmale sind die strukturierten Angaben, nach denen gefiltert wird —
	// leer bei allem, was keine Karte ist.
	Merkmale Retrievalmerkmale

	ProfilID     int64
	Sichtbarkeit Retrievalsicht
}

// Pruefe sagt, ob das Dokument vollständig und in sich stimmig ist.
func (d Retrievaldokument) Pruefe() error {
	if fehler := d.pruefeGrundangaben(); fehler != nil {
		return fehler
	}

	if fehler := d.pruefeHerkunft(); fehler != nil {
		return fehler
	}

	if fehler := d.Merkmale.Pruefe(); fehler != nil {
		return fehler
	}

	return d.pruefeSicht()
}

// pruefeGrundangaben hält, was jedes Dokument braucht.
func (d Retrievaldokument) pruefeGrundangaben() error {
	for name, wert := range map[string]string{
		"eine Kennung":       d.Kennung,
		"ein Spiel":          d.Spiel,
		"eine Sprache":       d.Sprache,
		"einen Text":         d.Text,
		"eine Quellrevision": d.Quellrevision,
		"einen Inhaltshash":  d.Inhaltshash,
	} {
		if strings.TrimSpace(wert) == "" {
			return fmt.Errorf("%w: ein Retrievaldokument braucht %s",
				ErrValidation, name)
		}
	}

	if !slices.Contains(retrievalquellen, d.Quelle) {
		return fmt.Errorf("%w: %q ist keine Retrievalquelle", ErrValidation, d.Quelle)
	}

	if d.Abgerufen.IsZero() {
		return fmt.Errorf(
			"%w: ein Retrievaldokument braucht eine Abrufzeit", ErrValidation)
	}

	if d.Abschnitt < 0 || d.Von < 0 || d.Bis < d.Von {
		return fmt.Errorf(
			"%w: der Abschnitt %d deckt den Bereich %d–%d — das ist keiner",
			ErrValidation, d.Abschnitt, d.Von, d.Bis)
	}

	return nil
}

// pruefeHerkunft hält, was **je Quelle** dazugehört.
func (d Retrievaldokument) pruefeHerkunft() error {
	if d.Kennungssystem != "" && !kennt(kennungssysteme, d.Kennungssystem) {
		return fmt.Errorf("%w: %q ist kein Kennungssystem",
			ErrValidation, d.Kennungssystem)
	}

	if d.Semantik != "" && !kennt(semantiken, d.Semantik) {
		return fmt.Errorf("%w: %q ist keine Semantik", ErrValidation, d.Semantik)
	}

	// **Eine Regelstelle ohne Nummer ist keine.** Die Hierarchie der
	// Comprehensive Rules ist die Auflösbarkeit: `205.4a` findet jemand
	// wieder, „irgendwo in Kapitel 205" nicht.
	if d.Quelle == QuelleRegelwerk && strings.TrimSpace(d.Regelnummer) == "" {
		return fmt.Errorf(
			"%w: eine Regelstelle braucht ihre Nummer", ErrValidation)
	}

	// **Und ein Ruling ohne Urheber ist eines, das »offiziell« behauptet.**
	if d.Quelle == QuelleRuling && !slices.Contains(rulingurheber, d.Urheber) {
		return fmt.Errorf(
			"%w: ein Ruling braucht seinen Urheber — %q ist keiner",
			ErrValidation, d.Urheber)
	}

	// **Ein Stand ohne Datum gilt seit wann?** Regelwerk und Ruling tragen
	// beide eines; ein Oracle-Text hat nur seine Revision.
	if (d.Quelle == QuelleRegelwerk || d.Quelle == QuelleRuling) &&
		strings.TrimSpace(d.Quelldatum) == "" {
		return fmt.Errorf(
			"%w: %q braucht ihr Gültigkeitsdatum — die Abrufzeit ist keines",
			ErrValidation, d.Quelle)
	}

	return nil
}

// pruefeSicht hält Sichtbarkeit und Profil zusammen.
func (d Retrievaldokument) pruefeSicht() error {
	switch d.Sichtbarkeit {
	case SichtbarOeffentlich:
		if d.ProfilID != 0 {
			return fmt.Errorf(
				"%w: ein öffentliches Dokument gehört keinem Profil — hier steht %d",
				ErrValidation, d.ProfilID)
		}

	case SichtbarProfil:
		if d.ProfilID == 0 {
			return fmt.Errorf(
				"%w: ein profilgebundenes Dokument braucht sein Profil",
				ErrValidation)
		}

	default:
		return fmt.Errorf("%w: %q ist keine Sichtbarkeit",
			ErrValidation, d.Sichtbarkeit)
	}

	return nil
}

// Retrievalhash gibt den Abdruck eines Textes.
//
// **Er ist kein Geheimnisschutz** (Issue). Wer den Text kennt, kennt den
// Abdruck; wozu er da ist, ist der Vergleich zweier Abschnitte.
func Retrievalhash(text string) string {
	summe := sha256.Sum256([]byte(text))

	return "sha256:" + hex.EncodeToString(summe[:])
}

// Retrievaltreffer ist ein Fund samt seiner Stelle.
//
// # Ein Rang und kein Punktwert
//
// **Weil sich Punktwerte nicht zusammenführen lassen.** BM25 gibt eine Zahl
// ohne Obergrenze, Cosinus eine zwischen −1 und 1, ein Reranker eine dritte
// Skala. Das Issue sagt es als Verbot: „BM25, Cosinus und Rerankwerte nicht
// roh addieren und nicht als Wahrscheinlichkeit ausgeben."
//
// **Was sich zusammenführen lässt, sind Ränge** — und genau das tut die
// Rangfusion.
type Retrievaltreffer struct {
	Dokument Retrievaldokument

	// Rang ist die Position in **dieser** Liste, ab eins.
	Rang int

	// Ausschnitt ist die Textstelle, die getroffen hat — leer, wo es keine
	// einzelne gibt.
	Ausschnitt string
}
