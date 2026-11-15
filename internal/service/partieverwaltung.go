package service

// **Löschen und Kopieren einer Partie** — PLAY-03.AK-05
// (#1711).
//
// # Eine eigene Datei
//
// **Der Schnitt folgt der Frage** (§2): `partieleitung.go` beantwortet, wie
// eine Partie **läuft**; hier steht, wie man eine loswird und wie aus einer
// zwei werden. Der Anlass war auch die Zeilengrenze — jene stand bei 400.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Partieverwaltungsspeicher sind die drei Wege, die eine Partie als Ganzes
// betreffen.
//
// **Eine eigene Schnittstelle und kein Anbau an [Partiespeicher]**: jene
// stand mit ihnen über der Zeilengrenze aus §2, und der Schnitt trägt — hier
// steht, was eine Partie **loswird** oder **verdoppelt**, dort, wie sie
// läuft. Sie wird dort eingebettet; für den Aufrufer ändert das nichts.
//
// **Offenlegungszahl und nicht die Liste**: was darin steht, geht die
// Löschregel nichts an; dass es etwas gibt, schon ([domain.Loeschbar]).
type Partieverwaltungsspeicher interface {
	Loeschen(ctx context.Context, kennung string) error
	Kopieren(ctx context.Context, kennung, neueKennung string) (
		domain.Partie, error)
	Offenlegungszahl(ctx context.Context, partieID string) (int, error)
}

// Loeschen nimmt eine Partie aus dem Bestand — wenn die Regel es erlaubt.
//
// **Die Regel steht in der Domäne** ([domain.Loeschbar]) und nicht hier: sie
// gehört zur Partie und nicht zu dem Ablauf, der sie anwendet (§2).
//
// **Der Satz nennt den Grund.** „Darf nicht" ohne ihn ließe jemanden raten,
// ob es am Stand liegt oder an einer Einwilligung — und beides führt zu einem
// anderen nächsten Schritt (§6).
func (l *Partieleitung) Loeschen(ctx context.Context, kennung string) error {
	if fehler := SitzbindungAus(ctx).ErlaubtVollenZustand(); fehler != nil {
		return fehler
	}

	partie, fehler := l.partien.Partie(ctx, kennung)
	if fehler != nil {
		return fehler
	}

	offenlegungen, fehler := l.partien.Offenlegungszahl(ctx, kennung)
	if fehler != nil {
		return fehler
	}

	if !domain.Loeschbar(partie, offenlegungen) {
		return fmt.Errorf("%w: %s", domain.ErrConflict, loeschgrund(partie))
	}

	return l.partien.Loeschen(ctx, kennung)
}

// loeschgrund sagt, **warum** eine Partie bleibt.
//
// **Nur der Stand geht hinein**, obwohl [domain.Loeschbar] zwei Dinge
// prüft: wer hier steht, ist schon abgewiesen — liegt es nicht am Stand,
// bleibt nur die Einwilligung übrig. Ein zweiter Parameter wäre einer, den
// niemand liest.
func loeschgrund(partie domain.Partie) string {
	if partie.Stand == domain.StandBeendet {
		return "diese Partie ist beendet und bleibt nachvollziehbar — was " +
			"mit ihr geschieht, entscheidet die Datenpflege nach ihrer Frist"
	}

	return "für diese Partie liegt eine Einwilligung vor — nimm sie zuerst " +
		"zurück, dann ist die Partie wieder ein Entwurf"
}

// Kopieren legt eine Partie als neuen Übungszustand daneben.
//
// **Jede Partie darf kopiert werden, auch eine beendete** — das ist der Sinn
// eines Übungszustands: eine Lage nachstellen und sie anders spielen. Die
// Löschregel gilt hier ausdrücklich **nicht**; sie schützt die Geschichte,
// und eine Kopie fasst die nicht an.
//
// **Die neue Kennung kommt von außen** — wie beim Anlegen, wo
// `app.NeuePartiekennung()` sie stellt. Zwei Quellen für dieselbe Zusage
// wären eine zu viel.
func (l *Partieleitung) Kopieren(
	ctx context.Context, kennung, neueKennung string,
) (domain.Partie, error) {
	// **Eine Kopie trägt den ganzen Zustand** — und eine ungebundene
	// Partie, in der das Token jede Hand läse (#1819).
	if fehler := SitzbindungAus(ctx).ErlaubtVollenZustand(); fehler != nil {
		return domain.Partie{}, fehler
	}

	return l.partien.Kopieren(ctx, kennung, neueKennung)
}

// Ausfuhr gibt eine Partie als Datei.
//
// **Ohne Sicht und ohne Zuschnitt**: was hier hinausgeht, ist der **ganze**
// Zustand — auch die Hand des Gegenübers. Das ist kein Versehen, sondern die
// Bedeutung dieser Datei: sie stellt eine Lage nach, und eine Lage ohne die
// verdeckten Teile wäre eine andere.
//
// **Wer sie bekommen darf, entscheidet die Autorisierung an der Fläche** —
// dieselbe Trennung wie beim Verlauf: hier steht der Anwendungsfall, dort
// die Frage, wer ihn stellen darf.
func (l *Partieleitung) Ausfuhr(
	ctx context.Context, kennung string,
) ([]byte, error) {
	// **Die Ausfuhr trägt jede Hand** — ein gebundenes Token bekommt sie nie
	// (#1819).
	if fehler := SitzbindungAus(ctx).ErlaubtVollenZustand(); fehler != nil {
		return nil, fehler
	}

	partie, fehler := l.partien.Partie(ctx, kennung)
	if fehler != nil {
		return nil, fehler
	}

	return domain.FormePartie(partie), nil
}

// Einfuhr legt eine Partie aus einer Datei an.
//
// **Sie geht durch [domain.Kopiere]** und nicht unmittelbar in den Bestand:
// eine eingelesene Partie ist ein **neuer Übungszustand** und nicht die
// Fortsetzung einer fremden. Deshalb trägt sie die neue Kennung, fängt bei
// Fassung 1 an und läuft — auch wenn die Datei eine beendete beschreibt.
//
// **Die Befunde gehen mit nach oben.** Ein fremdes Feld in der Datei ist
// kein Fehler, aber es wird **gesagt** (§8, §16): wer es verschlucken
// ließe, ließe jemanden glauben, es sei alles angekommen.
func (l *Partieleitung) Einfuhr(
	ctx context.Context, roh []byte, neueKennung string,
) (domain.Partie, []domain.Einfuhrbefund, error) {
	if fehler := SitzbindungAus(ctx).ErlaubtVollenZustand(); fehler != nil {
		return domain.Partie{}, nil, fehler
	}

	austausch, befunde, fehler := domain.ZerlegePartie(roh)
	if fehler != nil {
		return domain.Partie{}, nil, fehler
	}

	// **Ein eigenes Geheimnis** (#2480) — eine eingelesene Partie ist eine
	// neue, und das Geheimnis ihrer Herkunft gehört nicht zu ihr.
	kopie := domain.Kopiere(austausch.Partie, neueKennung)
	kopie.Geheimnis = l.geheimnis()

	angelegt, fehler := l.partien.Anlegen(ctx, kopie)
	if fehler != nil {
		return domain.Partie{}, nil, fehler
	}

	return angelegt, befunde, nil
}
