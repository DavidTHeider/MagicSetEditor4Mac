package app

// Die Naht des Autoersetzers — und die Umrechnung, die sie leisten muss.
//
// # Der Anlass
//
// `msepaket` liest die `auto replace:`-Regeln jedes Spielpakets, und
// `service.Autoersetzer` weiß, wann sie greifen. Bis zum 2026-08-23 konnte
// niemand sie erreichen
// (#897).
//
// # Warum die Umrechnung hier steht
//
// **Go zählt Bytes, JavaScript zählt UTF-16-Einheiten.** `selectionStart`
// eines Eingabefeldes ist eine UTF-16-Einheit; der Ersetzer rechnet in Bytes.
// Wer die Umrechnung dem Fenster überließe, müsste sie dort in TypeScript
// nachbauen — und zwei Zählweisen an zwei Orten laufen auseinander, sobald
// jemand ein Zeichen außerhalb der Grundebene tippt.
//
// Das ist genau die Sorte Naht, die §10 meint: jede Schicht für sich ist
// richtig, und der Weg durch sie hindurch trägt nur, wenn eine von beiden die
// Zählweise wechselt.

import (
	"context"
	"unicode/utf16"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Autoersetzungsquelle gibt die Regeln des Spiels, zu dem ein Set gehört.
type Autoersetzungsquelle interface {
	Autoersetzungen(ctx context.Context, setID int64) ([]msepaket.AutoErsetzung, error)
}

// AutoersetzungAntwort trägt Text und Schreibmarke — in UTF-16-Einheiten.
type AutoersetzungAntwort struct {
	Text   string  `json:"text"`
	Stelle int     `json:"stelle"`
	Fehler *Fehler `json:"fehler"`
}

// MitAutoersetzung setzt die Naht — für Tests und den Composition Root.
func (a *App) MitAutoersetzung(quelle Autoersetzungsquelle) { a.autoersetzung = quelle }

// AutoErsetzen wendet die Regeln des Spiels an der Schreibmarke an.
//
// **Text und Stelle kommen auch im Fehlerfall zurück, unverändert.** Ein
// Spielpaket, das nicht daliegt, ist auf einem frischen Rechner der Normalfall;
// wer dabei den Text verlöre, machte aus einer fehlenden Vorlage einen
// Datenverlust (§7 Regel 3).
func (a *App) AutoErsetzen(
	ctx context.Context, setID, text string, stelle int,
) AutoersetzungAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(setID)
	if befund != nil {
		return AutoersetzungAntwort{Text: text, Stelle: stelle, Fehler: befund}
	}

	regeln, err := a.autoersetzung.Autoersetzungen(ctx, kennung)
	if err != nil {
		return AutoersetzungAntwort{Text: text, Stelle: stelle, Fehler: Uebersetze(ctx, err)}
	}

	// **Eine Stelle außerhalb des Textes wird nicht gerechnet.** Sie kommt aus
	// dem Fenster, und was von dort kommt, ist fremde Eingabe (§8).
	inBytes, gueltig := byteStelle(text, stelle)
	if !gueltig {
		return AutoersetzungAntwort{Text: text, Stelle: stelle}
	}

	nachher, danach := service.NeuerAutoersetzer(regeln).AnStelle(text, inBytes)

	return AutoersetzungAntwort{Text: nachher, Stelle: utf16Stelle(nachher, danach)}
}

// byteStelle rechnet eine UTF-16-Einheit in eine Byte-Stelle um.
//
// Der zweite Rückgabewert sagt, ob die Stelle im Text liegt — eine Stelle
// **mitten** in einem Ersatzpaar gibt es nicht, und sie zu runden hieße, still
// etwas anderes zu meinen als das Fenster.
func byteStelle(text string, einheiten int) (int, bool) {
	if einheiten < 0 {
		return 0, false
	}

	gezaehlt := 0

	for byteStelle, zeichen := range text {
		if gezaehlt == einheiten {
			return byteStelle, true
		}

		gezaehlt += len(utf16.Encode([]rune{zeichen}))
	}

	if gezaehlt == einheiten {
		return len(text), true
	}

	return 0, false
}

// utf16Stelle rechnet eine Byte-Stelle in UTF-16-Einheiten um.
func utf16Stelle(text string, byteStelle int) int {
	if byteStelle > len(text) {
		byteStelle = len(text)
	}

	gezaehlt := 0

	for _, zeichen := range text[:byteStelle] {
		gezaehlt += len(utf16.Encode([]rune{zeichen}))
	}

	return gezaehlt
}

// Utf16StelleFuerTest gibt [utf16Stelle] für den Test frei.
//
// **Weil der Kappzweig von außen nicht erreichbar ist**: der Ersetzer gibt nur
// Stellen in seinem eigenen Ergebnis zurück. Er steht trotzdem da, weil eine
// Stelle hinter dem Text eine Bereichsverletzung wäre und kein falsches
// Ergebnis — und ein Zweig, den niemand ausführt, ist keiner (§3).
func Utf16StelleFuerTest(text string, byteStelle int) int {
	return utf16Stelle(text, byteStelle)
}
