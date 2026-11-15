// **Nachweise ausstellen und einlösen** — Werkstück O1 von Phase 5.2
// (#1425),
// nach [Vermerk 0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md).
//
// # Zwei Zusagen
//
//   - **Ein Nachweis entsteht dort, wo der Mensch ist.** Ausgestellt wird am
//     Fenster, ohne Token, von einer Rolle, die mindestens mitarbeitet. Eine
//     Maschine — Dienst, MCP, Kommandozeile — bekommt keinen: sie könnte ihn
//     sich sonst selbst geben, und das ist genau das `bestaetigt`-Feld, das
//     kein Nachweis sein soll.
//   - **Ein Nachweis wird einmal eingelöst**, und die Ablage entscheidet das
//     atomar: wer als Zweiter kommt, bekommt [domain.ErrConflict] und hier
//     die Absage [domain.ErrErlaubnis] — die Eingabe ist richtig, was fehlt,
//     ist die Erlaubnis.
//
// # Was hier **nicht** entschieden wird
//
// **Ob die Wirkung an sich erlaubt ist.** Das sagt die Rolle
// ([domain.Akteur.Darf]), und [Autorisierung.PruefeMitNachweis] fragt sie
// **zuerst** — ein Nachweis, der an einer fehlenden Rolle scheitert, bleibt
// unverbraucht.

package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Nachweisablage ist die Naht zum Bestand.
type Nachweisablage interface {
	NachweisAnlegen(ctx context.Context, nachweis domain.Nachweis) error
	Nachweis(ctx context.Context, kennung string) (domain.Nachweis, error)

	// NachweisEinloesen verbraucht einen Nachweis — **atomar**: ist er
	// schon verbraucht, kommt [domain.ErrConflict]; gibt es ihn nicht,
	// [domain.ErrNotFound].
	NachweisEinloesen(ctx context.Context, kennung string) error
}

// Nachweisantrag ist, was ein Mensch am Fenster freigibt.
type Nachweisantrag struct {
	Zweck    string
	Ziel     string
	Revision int64
	Grenze   domain.Kostengrenze

	// Frist ist, wie lange der Nachweis gilt — mehr als ein Tag ist keine
	// Freigabe für das Eine, sondern eine Vollmacht.
	Frist time.Duration
}

// Nachweisbindung ist der Teil der Bindung, den der Aufruf beisteuert;
// Nutzer und Profil kommen vom Akteur.
type Nachweisbindung struct {
	Zweck    string
	Ziel     string
	Revision int64
	Kosten   int64
	Waehrung string
}

// laengsteFrist ist die Obergrenze für einen Nachweis.
const laengsteFrist = 24 * time.Hour

// Nachweise stellt aus und löst ein.
type Nachweise struct {
	ablage  Nachweisablage
	uhr     func() time.Time
	kennung func() string
}

// NeueNachweise baut den Dienst über einer Ablage — mit der echten Uhr und
// zufälligen Kennungen.
func NeueNachweise(ablage Nachweisablage) *Nachweise {
	return &Nachweise{ablage: ablage, uhr: time.Now, kennung: nachweiskennung}
}

// MitUhr setzt die Uhr — für Tests, die einen Ablauf sehen wollen.
func (n *Nachweise) MitUhr(uhr func() time.Time) *Nachweise {
	n.uhr = uhr

	return n
}

// MitKennung setzt die Quelle der Kennungen.
func (n *Nachweise) MitKennung(kennung func() string) *Nachweise {
	n.kennung = kennung

	return n
}

// nachweiskennung gibt 128 Zufallsbits als Hexzeichen — eine eigene Quelle
// neben `zufallskennung` aus der Kontoanmeldung, weil jene auf eine feste
// Länge kürzt und base32 schreibt; hier zählt nur, dass niemand sie rät.
//
// **Ohne Fehlerrückgabe**, weil `crypto/rand.Read` seit Go 1.24 keinen
// Fehler mehr gibt — dieselbe Stelle wie dort.
func nachweiskennung() string {
	roh := make([]byte, 16)
	rand.Read(roh) //nolint:errcheck // gibt seit Go 1.24 keinen Fehler

	return hex.EncodeToString(roh)
}

// Ausstellen legt einen Nachweis für diesen Akteur an.
func (n *Nachweise) Ausstellen(
	ctx context.Context, wer domain.Akteur, antrag Nachweisantrag,
) (domain.Nachweis, error) {
	if wer.Modus != domain.ModusFenster || wer.Token != "" {
		return domain.Nachweis{}, fmt.Errorf(
			"%w: ein Nachweis wird am Fenster ausgestellt, von einem Menschen — "+
				"nicht über %s", domain.ErrErlaubnis, string(wer.Modus))
	}

	if !wer.Rolle.Reicht(domain.RolleMitarbeiter) || wer.NurLesend {
		return domain.Nachweis{}, fmt.Errorf(
			"%w: einen Nachweis stellt aus, wer mindestens mitarbeitet", domain.ErrErlaubnis)
	}

	if antrag.Frist <= 0 || antrag.Frist > laengsteFrist {
		return domain.Nachweis{}, domain.NewFieldError("frist", domain.SchluesselValidierung,
			"Ein Nachweis gilt zwischen einer Sekunde und einem Tag — nicht %s.",
			antrag.Frist.String())
	}

	jetzt := n.uhr()
	nachweis := domain.Nachweis{
		Kennung:     n.kennung(),
		NutzerID:    wer.NutzerID,
		ProfilID:    wer.ProfilID,
		Zweck:       antrag.Zweck,
		Ziel:        antrag.Ziel,
		Revision:    antrag.Revision,
		Grenze:      antrag.Grenze,
		Ausgestellt: jetzt,
		Ablauf:      jetzt.Add(antrag.Frist),
	}

	if fehler := nachweis.Pruefe(); fehler != nil {
		return domain.Nachweis{}, fehler
	}

	if fehler := n.ablage.NachweisAnlegen(ctx, nachweis); fehler != nil {
		return domain.Nachweis{}, fehler
	}

	return nachweis, nil
}

// Einloesen verbraucht einen Nachweis, wenn er zur Bindung passt.
func (n *Nachweise) Einloesen(
	ctx context.Context, kennung string, bindung domain.Nachweisbindung,
) error {
	nachweis, fehler := n.ablage.Nachweis(ctx, kennung)
	if fehler != nil {
		return fehler
	}

	if bindungsfehler := nachweis.Gilt(bindung, n.uhr()); bindungsfehler != nil {
		return bindungsfehler
	}

	fehler = n.ablage.NachweisEinloesen(ctx, kennung)
	if errors.Is(fehler, domain.ErrConflict) {
		return fmt.Errorf("%w: dieser Nachweis ist schon eingelöst", domain.ErrErlaubnis)
	}

	return fehler
}

// MitNachweisen gibt der Autorisierung den Nachweisdienst.
func (a *Autorisierung) MitNachweisen(nachweise *Nachweise) *Autorisierung {
	a.nachweise = nachweise

	return a
}

// PruefeMitNachweis weist ab, was dieser Aufruf nicht darf — und löst für
// das, was einen Nachweis braucht, den vorgezeigten ein.
//
// **Die Rolle zuerst, dann der Nachweis.** Wer die Wirkung nicht darf, bekommt
// die Absage der Rolle, und sein Nachweis bleibt unverbraucht — er hat ihn
// womöglich für einen berechtigten Aufruf.
func (a *Autorisierung) PruefeMitNachweis(
	ctx context.Context, wirkung domain.Wirkung, kennung string, bindung Nachweisbindung,
) error {
	return gezaehlt(ctx, wirkung, a.pruefeMitNachweis(ctx, wirkung, kennung, bindung))
}

func (a *Autorisierung) pruefeMitNachweis(
	ctx context.Context, wirkung domain.Wirkung, kennung string, bindung Nachweisbindung,
) error {
	akteur, fehler := a.Akteur(ctx)
	if fehler != nil {
		return fehler
	}

	if fehler := akteur.Darf(wirkung); fehler != nil {
		return fehler
	}

	if !wirkung.BrauchtNachweis() && kennung == "" {
		return nil
	}

	if a.nachweise == nil {
		return fmt.Errorf("%w: %q braucht einen Nachweis, und diese Fläche stellt "+
			"keinen aus", domain.ErrErlaubnis, string(wirkung))
	}

	return a.nachweise.Einloesen(ctx, kennung, domain.Nachweisbindung{
		NutzerID: akteur.NutzerID,
		ProfilID: akteur.ProfilID,
		Zweck:    bindung.Zweck,
		Ziel:     bindung.Ziel,
		Revision: bindung.Revision,
		Kosten:   bindung.Kosten,
		Waehrung: bindung.Waehrung,
	})
}
