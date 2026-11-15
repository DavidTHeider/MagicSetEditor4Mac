package service_test

// **Die Doppel des Vertragslaufs** — abgetrennt am 2026-09-11, weil
// `kivertragslauf_test.go` mit ihnen über die Zeilengrenze aus §2 ging
// (A6, #1445).
//
// # Der Schnitt folgt der Frage
//
// Dort stehen **die Läufe** — was geprüft wird und warum. Hier stehen die
// **Anbieter, die sich danebenbenehmen**: jede Flagge ist ein Verhalten, das
// ein echter Anbieter zeigen kann, und deshalb steht neben jeder, welches.

import (
	"context"
	"math"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einbettungsprobe ist eine Quelle, die sich auf Wunsch danebenbenimmt.
type einbettungsprobe struct {
	laenge      int
	keineZahl   bool
	ohneNorm    bool
	wechselhaft bool
	blind       bool
	halbiert    bool

	// kuerzerAb lässt ab diesem Ruf einen Vektor weg — für die Fälle, in
	// denen ein Anbieter **zwischen** zwei Rufen seine Form wechselt.
	kuerzerAb int

	// laengerAb gibt ab diesem Ruf Vektoren anderer Länge.
	laengerAb int
	fehler    error

	// fehlerAb sagt, ab dem wievielten Ruf der Fehler kommt — 0 heißt: sofort.
	fehlerAb int

	rufe int
}

func (e *einbettungsprobe) Einbetten(_ context.Context, texte []string) (
	[][]float32, error,
) {
	e.rufe++

	if e.fehler != nil && e.rufe >= max(e.fehlerAb, 1) {
		return nil, e.fehler
	}

	if e.halbiert || (e.kuerzerAb > 0 && e.rufe >= e.kuerzerAb) {
		texte = texte[:1]
	}

	vektoren := make([][]float32, 0, len(texte))

	for _, text := range texte {
		vektoren = append(vektoren, e.einVektor(text))
	}

	return vektoren, nil
}

// einVektor baut einen Vektor — normalisiert, sofern nicht anders verlangt.
func (e *einbettungsprobe) einVektor(text string) []float32 {
	laenge := 4
	if e.laenge > 0 {
		laenge = e.laenge
	}

	if e.laengerAb > 0 && e.rufe >= e.laengerAb {
		laenge++
	}

	vektor := make([]float32, laenge)

	// **Die Richtung hängt am Text** — und zwar über **zwei** Stellen.
	//
	// Mit nur einer wäre jeder normalisierte Vektor `[1,0,0,0]`, und der Lauf
	// über die Anweisungen prüfte nichts: die Normalisierung löschte gerade
	// den Unterschied, den er sucht. Das ist beim Schreiben dieses Doppels
	// passiert, und es ist genau der Fall, für den es den Lauf gibt.
	vektor[1] = 1

	if !e.blind {
		vektor[0] = float32(quersumme(text)%7) + 1
	} else {
		vektor[0] = 1
	}

	if e.wechselhaft {
		vektor[0] += float32(e.rufe)
	}

	if e.keineZahl {
		vektor[0] = float32(math.NaN())

		return vektor
	}

	if e.ohneNorm {
		vektor[0] *= 3

		return vektor
	}

	// **Auf Länge eins bringen** — das sagt der Vertrag zu.
	laengeQuadrat := float64(0)
	for _, wert := range vektor {
		laengeQuadrat += float64(wert) * float64(wert)
	}

	betrag := float32(math.Sqrt(laengeQuadrat))
	for i := range vektor {
		vektor[i] /= betrag
	}

	return vektor
}

// neuordnungsprobe bewertet Kandidaten — auf Wunsch auch erfundene.
type neuordnungsprobe struct {
	erfindet bool

	// fremd tauscht einen Kandidatennamen aus — die Zahl stimmt, der Name
	// nicht.
	fremd bool

	fehler error
}

func (n *neuordnungsprobe) Neuordnen(_ context.Context, _ string, kandidaten []string) (
	[]domain.KIBewertung, error,
) {
	if n.fehler != nil {
		return nil, n.fehler
	}

	bewertungen := make([]domain.KIBewertung, 0, len(kandidaten)+1)

	for i, kandidat := range kandidaten {
		if e := i == 0 && n.fremd; e {
			kandidat = "nie geschickt"
		}

		bewertungen = append(bewertungen, domain.KIBewertung{
			Kandidat: kandidat, Wert: 1 / float64(i+1),
		})
	}

	if n.erfindet {
		bewertungen = append(bewertungen, domain.KIBewertung{
			Kandidat: "nie geschickt", Wert: 0.5,
		})
	}

	return bewertungen, nil
}

// quersumme ist ein grober Fingerabdruck eines Textes.
//
// **Er muss nicht gut sein** — er muss nur zwei verschiedene Texte
// unterscheiden, und die beiden Probetexte tun das.
func quersumme(text string) int {
	summe := 0

	for _, zeichen := range text {
		summe += int(zeichen)
	}

	return summe
}
