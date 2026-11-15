package service

// **Wie aus Treffern Belege werden** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Die Zählweise ist geschätzt, und sie sagt es
//
// **Vier Zeichen je Token.** Das ist die grobe Faustregel für lateinische
// Schrift; sie stimmt für kein Modell genau, und deshalb steht sie **als
// Name im Manifest** (`zeichen/4`). Wer eine genaue Zahl braucht, zählt mit
// dem Zerteiler des Zielmodells — und trägt einen anderen Namen ein.
//
// **Ohne den Namen wäre die Zahl bedeutungslos.** Das Issue verlangt die
// „Tokenzählmethode" ausdrücklich im Manifest; eine Zahl ohne sie sagte
// niemandem, wie nah sie an der Wahrheit liegt.
//
// # Und warum gekürzt wird, statt wegzulassen
//
// **Weil ein halber Beleg noch ein Beleg ist.** Ein Regeltext, dessen letzter
// Satz fehlt, trägt seine Regelnummer und sein Datum; ein weggelassener trägt
// nichts. Gekürzt wird **sichtbar** — der Ausschnitt endet mit einem Zeichen,
// das es sagt.

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Zaehlweise nennt, wie Tokens geschätzt werden.
const Zaehlweise = "zeichen/4"

// zeichenJeToken ist die Faustregel.
const zeichenJeToken = 4

// Kuerzungszeichen steht am Ende eines gekürzten Ausschnitts.
//
// **Sichtbar und nicht still** (Issue: „token-/bytebezogene Kürzung ist
// nachvollziehbar"): wer einen Beleg liest, soll sehen, dass er nicht ganz
// ist.
const Kuerzungszeichen = " […]"

// Mindestbeleg ist, wie viele Zeichen ein Ausschnitt mindestens trägt.
//
// **Achtzig**, und die Zahl ist gewählt: darunter steht kein Satz mehr da,
// und ein Beleg ohne Satz belegt nichts. Wer so wenig Platz hat, bekommt
// **einen Beleg weniger** statt eines unbrauchbaren.
const Mindestbeleg = 80

// Tokenschaetzung gibt die geschätzte Zahl der Tokens eines Textes.
func Tokenschaetzung(text string) int {
	if text == "" {
		return 0
	}

	return (len([]rune(text)) + zeichenJeToken - 1) / zeichenJeToken
}

// belegeAus wählt aus den Treffern, was ins Budget passt.
//
// **Die Reihenfolge der Treffer bleibt die Reihenfolge der Belege.** Sie ist
// die Ordnung der Rangfusion, und sie hier umzustellen hieße, eine zweite
// Ordnung über eine bestehende zu legen.
func belegeAus(
	treffer []domain.Recherchetreffer, wunsch Kontextwunsch, verfuegbar int,
) (belege []domain.Beleg, verbraucht int) {
	belege = make([]domain.Beleg, 0, len(treffer))

	for _, einzelner := range treffer {
		if wunsch.HoechstensBelege > 0 && len(belege) >= wunsch.HoechstensBelege {
			break
		}

		ausschnitt, kosten := passendGekuerzt(
			volltextVon(einzelner), verfuegbar-verbraucht)
		if ausschnitt == "" {
			// **Kein Platz mehr** — und die Liste bricht ab, statt den
			// nächsten zu versuchen: die Treffer stehen nach Rang, und ein
			// späterer wäre kürzer und schlechter.
			break
		}

		belege = append(belege, belegAus(einzelner, ausschnitt, len(belege)))
		verbraucht += kosten
	}

	return belege, verbraucht
}

// volltextVon gibt den Text, aus dem der Ausschnitt geschnitten wird.
//
// **Der ganze Abschnitt und nicht das Suchsnippet.** Das Snippet zeigt, warum
// etwas gefunden wurde; der Beleg soll tragen, was daraus folgt — und dafür
// braucht es den Satz und nicht drei Wörter daraus.
func volltextVon(treffer domain.Recherchetreffer) string {
	if strings.TrimSpace(treffer.Text) != "" {
		return treffer.Titel + "\n" + treffer.Text
	}

	return treffer.Titel
}

// passendGekuerzt kürzt einen Text auf das, was noch ins Budget passt.
func passendGekuerzt(text string, uebrig int) (ausschnitt string, kosten int) {
	kosten = Tokenschaetzung(text)
	if kosten <= uebrig {
		return text, kosten
	}

	// **Was übrig ist, in Zeichen** — abzüglich des Kürzungszeichens.
	zeichen := uebrig*zeichenJeToken - len(Kuerzungszeichen)
	if zeichen < Mindestbeleg {
		return "", 0
	}

	gekuerzt := string([]rune(text)[:min(zeichen, len([]rune(text)))]) +
		Kuerzungszeichen

	return gekuerzt, Tokenschaetzung(gekuerzt)
}

// belegAus macht aus einem Treffer einen Beleg.
func belegAus(
	treffer domain.Recherchetreffer, ausschnitt string, stelle int,
) domain.Beleg {
	return domain.Beleg{
		Kennung:        domain.BelegkennungFuer(stelle),
		Quellkennung:   treffer.Kennung,
		Quelle:         treffer.Quelle,
		Kennungssystem: treffer.Kennungssystem,
		Fachkennung:    treffer.Fachkennung,
		Semantik:       treffer.Semantik,
		Titel:          treffer.Titel,
		Ausschnitt:     ausschnitt,
		Von:            treffer.Von,
		Bis:            treffer.Bis,
		Inhaltshash:    domain.Belegabdruck(ausschnitt),
		Sprache:        treffer.Sprache,
		Regelnummer:    treffer.Regelnummer,
		Urheber:        treffer.Urheber,
		Bezugsadresse:  treffer.Bezugsadresse,
		Quellrevision:  treffer.Quellrevision,
		Quelldatum:     treffer.Quelldatum,
		Abgerufen:      treffer.Abgerufen,
	}
}

// kontexttext setzt zusammen, was das Modell zu sehen bekommt.
//
// **Mit der Belegkennung vor jedem Ausschnitt**, damit ein Modell sie nennen
// kann — und mit der Herkunft dahinter, damit es sie nicht erfinden muss.
func kontexttext(belege []domain.Beleg) string {
	teile := make([]string, 0, len(belege))

	for _, beleg := range belege {
		teile = append(teile, "["+string(beleg.Kennung)+"] "+
			beleg.Ausschnitt+"\n("+herkunftszeile(beleg)+")")
	}

	return strings.Join(teile, "\n\n")
}

// herkunftszeile sagt, woher ein Beleg stammt.
func herkunftszeile(beleg domain.Beleg) string {
	teile := []string{string(beleg.Quelle)}

	for _, angabe := range []string{
		beleg.Regelnummer, beleg.Quellrevision, beleg.Quelldatum,
	} {
		if strings.TrimSpace(angabe) != "" {
			teile = append(teile, angabe)
		}
	}

	return strings.Join(teile, " · ")
}

// gewuerfelteKennung gibt eine Paketkennung.
//
// **Gewürfelt und nicht gezählt**: eine laufende Nummer sagte, wie viele
// Pakete es gibt, und ließe eines erraten.
func gewuerfelteKennung() string {
	roh := make([]byte, 16)

	// **Der Fehler steht im Rückgabewert und nicht in einem `if`**:
	// `crypto/rand.Read` scheitert seit Go 1.24 nicht mehr, und ein Zweig
	// dafür stünde ungeprüft da (§10) — dieselbe Gestalt wie im Bildvorrat.
	_, _ = rand.Read(roh) //nolint:errcheck // siehe oben

	return "kp-" + hex.EncodeToString(roh)
}
