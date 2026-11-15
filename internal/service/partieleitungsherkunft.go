package service

// **Was über die Leitung kommt, nennt Sichtkennungen** — #2480, Vermerk
// 0201, Punkte 6 und 7.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// leitungsherkunft ist der Schlüssel des Merkmals im Kontext.
type leitungsherkunft struct{}

// MitLeitungsherkunft sagt der Partieleitung, dass ein Befehl von einem Gast
// über die Leitung kommt. **Ein Gast kennt keine Kennung**: er sieht nur
// Sichtkennungen, und eine Kennung in seinem Befehl ist erraten.
func MitLeitungsherkunft(ctx context.Context) context.Context {
	return context.WithValue(ctx, leitungsherkunft{}, true)
}

// vonDerLeitung sagt, ob ein Befehl über die Leitung kam.
func vonDerLeitung(ctx context.Context) bool {
	gekommen, _ := ctx.Value(leitungsherkunft{}).(bool) //nolint:errcheck // ohne Merkmal nicht von der Leitung

	return gekommen
}

// keineKennungVonDerLeitung weist einen Befehl von der Leitung ab, der eine
// Kennung nennt — **ohne sie zu nennen**: der Satz geht an den Gast zurück.
func keineKennungVonDerLeitung(
	ctx context.Context, partie domain.Partie, befehl domain.Spielbefehl,
) error {
	if !vonDerLeitung(ctx) {
		return nil
	}

	if _, nennt := partie.NenntKennung(befehl.Nutzlast); nennt {
		return domain.MitGrund(domain.ErrValidation, nil,
			"ein Gast nennt ein Objekt bei dem Namen, unter dem er es sieht")
	}

	return nil
}

// bereinigt ist ein Fehler, dessen Satz keine Kennung nennt — **und dessen
// Klasse bleibt**: der Wirt erkennt an ihr ein Urteil ([einUrteil]).
type bereinigt struct {
	satz  string
	innen error
}

func (b bereinigt) Error() string { return b.satz }

func (b bereinigt) Unwrap() error { return b.innen }

// ohneKennungen gibt einen Fehler, der keine Kennung dieser Partie nennt.
func ohneKennungen(partie domain.Partie, fehler error) error {
	if fehler == nil {
		return nil
	}

	return bereinigt{satz: partie.OhneKennungen(fehler.Error()), innen: fehler}
}

// MitGeheimnisquelle setzt, woher eine Partie ihr Geheimnis bekommt — für
// den Lauf, der denselben Anfang zweimal braucht.
func (l *Partieleitung) MitGeheimnisquelle(quelle func() string) *Partieleitung {
	l.geheimnis = quelle

	return l
}
