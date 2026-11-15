package service

// **Was der Wirt über seinen eigenen Stand weiß** — Werkstück G5 von Phase
// 5.5 (#1462).
//
// # Warum das eine eigene Datei ist
//
// **Weil `partieleitung.go` an der Zeilengrenze steht** (§2) — und weil der
// Schnitt trägt: dort steht, was eine Partie **tut**, hier, wo sie **steht**.
// Beides sind Fragen an denselben Bestand, aber nicht dieselbe Frage.
//
// # Und warum der Wirt ein Fenster nachreicht statt alles
//
// **Weil ein Tischbild eine Nachricht ist und ein Strom beliebig viele.**
// [domain.Wirtsstand.AelteteFolge] hieß im ersten Entwurf „die älteste, die
// noch vorliegt" — im Bestand liegt aber **alles** vor, und damit wäre der
// Abzugsweg toter Code. Die Bedeutung ist ausdrücklich verschoben (§15.4):
// **die älteste, die der Wirt noch nachreicht.**

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Nachreichfenster ist, wie viele Ereignisse ein Wirt höchstens nachreicht.
//
// **Zweihundert, und die Zahl ist eine Setzung mit einem Grund**: ein
// Tischbild ist **eine** Nachricht von begrenzter Größe; zweihundert
// Ereignisse sind mehr Verkehr als das Bild, und sie kosten den Gast
// zweihundert Schritte, an deren Ende dasselbe steht. Wer weiter zurückliegt,
// bekommt deshalb das Bild.
const Nachreichfenster = 200

// Wirtsstand gibt, wo dieser Wirt im Strom einer Partie steht.
//
// **Aus der Sicht eines Sitzes**, denn der Abdruck ist sichtbezogen: die
// Kette ist für alle gleich, die privaten Ergänzungen sind es nicht
// ([domain.Ereignis.Sichtabdruck]).
func (l *Partieleitung) Wirtsstand(
	ctx context.Context, kennung, fuer string,
) (domain.Wirtsstand, error) {
	if fehler := SitzbindungAus(ctx).Erlaubt(kennung, fuer); fehler != nil {
		return domain.Wirtsstand{}, fehler
	}

	if _, fehler := l.partien.Partie(ctx, kennung); fehler != nil {
		return domain.Wirtsstand{}, fehler
	}

	_, letzte, fehler := l.partien.LetzteKette(ctx, kennung)
	if fehler != nil {
		return domain.Wirtsstand{}, fehler
	}

	stand := domain.Wirtsstand{
		Epoche:       domain.Partieepoche,
		Folge:        folgeZahl(letzte),
		AelteteFolge: aeltesteNachgereichte(letzte),
	}

	if letzte == 0 {
		return stand, nil
	}

	abdruck, gerechnet := l.sichtabdruck(ctx, kennung, fuer, letzte)
	if gerechnet != nil {
		return domain.Wirtsstand{}, gerechnet
	}

	stand.Kette = abdruck

	return stand, nil
}

// Nachreichen gibt die Ereignisse ab einer Folge — für **einen** Sitz.
//
// **Gefiltert und nicht roh** (G1): was ein Sitz nicht sehen darf, sieht er
// auch beim Nachreichen nicht. Der Weg dahin ist derselbe wie beim Verlauf,
// und das ist die Zusage: es gibt **einen** Weg von der Wahrheit zur Sicht.
//
// **Ohne die Einwilligungen, und das ist der Unterschied zum Verlauf.** Eine
// [domain.Offenlegung] gilt der **Nachlese**; hier holt jemand nach, was er
// während des Spiels verpasst hat — und das ist der laufende Tisch. Wer sie
// hier mitrechnete, zeigte einem Zurückkehrenden ein fremdes Blatt, während
// noch gespielt wird.
func (l *Partieleitung) Nachreichen(
	ctx context.Context, kennung, fuer string, abFolge int64,
) ([]domain.Ereignis, error) {
	if fehler := SitzbindungAus(ctx).Erlaubt(kennung, fuer); fehler != nil {
		return nil, fehler
	}

	if _, fehler := l.partien.Partie(ctx, kennung); fehler != nil {
		return nil, fehler
	}

	strom, fehler := l.partien.Ereignisse(ctx, kennung, abFolge, 0)
	if fehler != nil {
		return nil, fehler
	}

	return domain.RueckschauFuer(strom, fuer, nil), nil
}

// sichtabdruck rechnet den Abdruck des Ereignisses mit dieser Folge.
func (l *Partieleitung) sichtabdruck(
	ctx context.Context, kennung, fuer string, folge int64,
) (string, error) {
	strom, fehler := l.partien.Ereignisse(ctx, kennung, folge, folge)
	if fehler != nil {
		return "", fehler
	}

	if len(strom) == 0 {
		// **Der Bestand nennt eine Folge, die er nicht hat.** Das ist kein
		// Netzfehler und keine Frage an den Gast, sondern ein Widerspruch im
		// eigenen Haus — und er wird benannt (§8).
		return "", fmt.Errorf(
			"%w: die Partie %q nennt die Folge %d und hat sie nicht",
			domain.ErrConflict, kennung, folge)
	}

	return strom[len(strom)-1].Sichtabdruck(fuer)
}

// folgeZahl macht aus einer Bestandsfolge eine Folgenummer.
//
// **Mit einer Grenze nach unten**: eine negative Folge gibt es nicht, und eine
// vorzeichenlose Umwandlung machte daraus eine sehr große Zahl — der Gast
// stünde dann für immer „hinter" dem Wirt.
func folgeZahl(folge int64) uint64 {
	if folge < 0 {
		return 0
	}

	return uint64(folge)
}

// aeltesteNachgereichte gibt die älteste Folge, die noch nachgereicht wird.
func aeltesteNachgereichte(letzte int64) uint64 {
	if letzte <= Nachreichfenster {
		return 1
	}

	return folgeZahl(letzte - Nachreichfenster + 1)
}
