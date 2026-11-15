package service

// **Die Einwilligung, dass später jemand mitliest** — Werkstück G5 von Phase
// 5.5 (#1462).
//
// # Warum sie hier steht und nicht bei der Rückschau
//
// **Weil sie geschrieben wird und jene nur liest.** [Partieleitung.Verlauf]
// beantwortet „was darf dieser sehen"; hier entscheidet ein Sitz, was er
// hergibt. Zwei Richtungen derselben Sache, und die schreibende hat einen
// Eigner: **nur der Sitz selbst**.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Offenlegen trägt die Einwilligung eines Sitzes ein — und ersetzt die alte.
//
// **Ersetzen und nicht ergänzen.** Wer eine Liste schickt, meint diese Liste;
// wer einen Empfänger herausnimmt, will ihn heraus haben. Ein Hinzufügen
// hätte keine Gegenrichtung, und „zurücknehmen" wäre ein zweiter Ruf, den
// man vergessen kann.
func (l *Partieleitung) Offenlegen(
	ctx context.Context, kennung string, gabe domain.Offenlegung,
) error {
	if fehler := gabe.Pruefe(); fehler != nil {
		return fehler
	}

	// **Freigeben kann nur der eigene Sitz** — auch für ein Token (#1819).
	if fehler := SitzbindungAus(ctx).Erlaubt(kennung, gabe.Von); fehler != nil {
		return fehler
	}

	partie, fehler := l.partien.Partie(ctx, kennung)
	if fehler != nil {
		return fehler
	}

	// **Nur ein Sitz dieser Partie hat etwas herzugeben.** Ein Zuschauer
	// bekommt keine private Ergänzung; eine Einwilligung von ihm gäbe nichts
	// frei und stünde trotzdem da, als hätte sie gewirkt (§10).
	if partie.SitzVon(gabe.Von) == nil {
		return fmt.Errorf("%w: %q sitzt nicht an dieser Partie",
			domain.ErrValidation, gabe.Von)
	}

	return l.partien.Freigeben(ctx, kennung, gabe)
}

// Zuruecknehmen schließt die Tür wieder.
//
// **Es holt nichts ein.** Wer den Verlauf schon gesehen hat, hat ihn gesehen —
// derselbe Satz wie bei der Deckleihe, und derselbe Grund: dieses Programm
// hat keinen Zugriff auf fremde Kopien, und eine Zusage, die es nicht halten
// kann, ist schlimmer als keine.
func (l *Partieleitung) Zuruecknehmen(
	ctx context.Context, kennung, von string,
) error {
	if fehler := SitzbindungAus(ctx).Erlaubt(kennung, von); fehler != nil {
		return fehler
	}

	if _, fehler := l.partien.Partie(ctx, kennung); fehler != nil {
		return fehler
	}

	return l.partien.Freigabeloeschen(ctx, kennung, von)
}

// Freigaben gibt, wer wem etwas freigegeben hat.
func (l *Partieleitung) Freigaben(
	ctx context.Context, kennung string,
) ([]domain.Offenlegung, error) {
	if fehler := SitzbindungAus(ctx).ErlaubtPartie(kennung); fehler != nil {
		return nil, fehler
	}

	if _, fehler := l.partien.Partie(ctx, kennung); fehler != nil {
		return nil, fehler
	}

	return l.partien.Freigaben(ctx, kennung)
}

// Vergessen löscht das Protokoll einer Partie.
//
// **Der ausdrückliche Löschweg aus #1462.** Die Frist daneben steht in der
// Datenpflege; sie **bietet** an, dieser hier **tut**.
//
// **Eine laufende Partie wird nicht vergessen.** Wer sie löschte, während
// jemand daran sitzt, nähme dem Tisch die Grundlage mitten im Zug — und der
// Gast bekäme für seinen nächsten Wunsch „gibt es nicht". Erst beenden, dann
// vergessen.
func (l *Partieleitung) Vergessen(ctx context.Context, kennung string) error {
	if fehler := SitzbindungAus(ctx).ErlaubtVollenZustand(); fehler != nil {
		return fehler
	}

	partie, fehler := l.partien.Partie(ctx, kennung)
	if fehler != nil {
		return fehler
	}

	if partie.Stand != domain.StandBeendet {
		return fmt.Errorf(
			"%w: diese Partie läuft noch — ein Protokoll wird vergessen, "+
				"wenn es eines ist", domain.ErrConflict)
	}

	return l.partien.Vergessen(ctx, kennung)
}
