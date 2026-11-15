package service

// **Einen Handel wählen, statt seine Kennung zu tippen** — UX-08.AK-04
// (#1680):
// *„Normalnutzende müssen keine Netzwerk- oder Transaktionskennungen
// eingeben, soweit diese technisch ableitbar sind."*
//
// **Zwei Wege dahin, und beide stehen hier**: die Kennung eines neuen
// Handels vergibt die Leitung ([Handelsleitung.Eroeffnen]), und die Handel
// mit einem Gegenüber sind eine Liste ([Handelsleitung.Vorgaenge]). Wer einen
// Handel empfängt, erfuhr seine Kennung bis dahin nirgends im Fenster.

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// MitKennungen setzt, wer die Kennung eines neuen Handels vergibt — für den
// Test, der wissen muss, welche es wird.
func (l *Handelsleitung) MitKennungen(kennungen func() string) *Handelsleitung {
	l.kennungen = kennungen

	return l
}

// handelskennung ist die Kennung eines neuen Handels: `h-` und acht
// zufällige Bytes.
//
// **Zufällig und nicht fortlaufend**: zwei Rechner eröffnen Handel, ohne
// einander zu fragen, und eine Zahl, die jeder bei eins beginnt, träfe sich.
// **Der Vorsatz sagt, was sie ist**, wenn sie jemand in einem Protokoll
// liest. `crypto/rand` scheitert nicht — seit Go 1.24 hält es das Programm
// an, statt einen Fehler zu geben.
func handelskennung() string {
	roh := make([]byte, 8)
	_, _ = rand.Read(roh) //nolint:errcheck // der Grund steht darüber

	return "h-" + hex.EncodeToString(roh)
}

// Vorgaenge gibt die Handel, an denen ein Gegenüber beteiligt ist — je
// Vorgang die letzte Revision, zuletzt geänderte zuerst.
//
// **Wer beteiligt ist, sagt das Angebot** ([domain.Handelsangebot.Parteien])
// und nicht eine zweite Liste daneben. **Keiner ist eine leere Liste**:
// „mit ihr noch kein Handel" ist eine Antwort.
func (l *Handelsleitung) Vorgaenge(
	ctx context.Context, gegenueber string,
) ([]domain.Handelsangebot, error) {
	if strings.TrimSpace(gegenueber) == "" {
		return nil, domain.MitGrund(domain.ErrValidation, nil,
			"ohne Gegenüber gibt es keine Liste seiner Handel")
	}

	alle, fehler := l.ablage.Letzte(ctx)
	if fehler != nil {
		return nil, fehler
	}

	mit := []domain.Handelsangebot{}

	for _, angebot := range alle {
		if slices.Contains(angebot.Parteien, gegenueber) {
			mit = append(mit, angebot)
		}
	}

	return mit, nil
}
