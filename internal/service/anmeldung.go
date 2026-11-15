// **Die Anmeldung über einen Anbieter** — Werkstück 4 von Phase 4.12
// (#1351),
// nach Vermerk [0066](../../docs/entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md).
//
// # Was hier entschieden wird
//
// **Was aus einer fremden Kennung wird.** Ist sie bekannt, gehört sie einem
// Nutzer; ist sie neu, entsteht einer. Das ist eine fachliche Entscheidung und
// steht deshalb hier — nicht in `internal/anmeldung`, das nur den Fluss kennt,
// und nicht in `internal/app`, das nichts entscheiden soll (§2).
//
// # Was hier **nicht** steht
//
// **Kein OAuth.** Adressen, PKCE, `state` und `nonce` sind Sache des Blatts.
// Dieser Dienst bekommt eine fertige [domain.Fremdkennung] und fragt nur noch:
// wer ist das hier?

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Kontenablage ist die Naht zum Bestand.
type Kontenablage interface {
	NutzerZuKonto(ctx context.Context, anbieter, subjekt string) (int64, error)
	NutzerAnlegen(ctx context.Context, neu domain.Nutzer) (domain.Nutzer, error)
	KontoVerbinden(ctx context.Context, neu domain.Anbieterkonto) (domain.Anbieterkonto, error)
}

// Anmeldung führt eine fremde Kennung auf einen Nutzer.
type Anmeldung struct{ ablage Kontenablage }

// NeueAnmeldung baut sie über einer Ablage.
func NeueAnmeldung(ablage Kontenablage) *Anmeldung { return &Anmeldung{ablage: ablage} }

// Herein gibt den Nutzer zu dieser Kennung — und legt ihn an, wenn es ihn
// noch nicht gibt.
//
// # Warum ein neuer Nutzer entsteht statt einer Rückfrage
//
// **Weil eine Rückfrage eine Hürde ohne Gegenwert wäre.** Wer sich über GitHub
// anmeldet, hat gesagt, wer er ist; ihn danach zu bitten, seinen Namen
// einzutippen, fragt zweimal dasselbe.
func (a *Anmeldung) Herein(ctx context.Context, kennung domain.Fremdkennung) (int64, error) {
	if strings.TrimSpace(kennung.Subjekt) == "" {
		return 0, fmt.Errorf(
			"%w: die Anmeldung nennt keine Kennung für diesen Menschen",
			domain.ErrValidation)
	}

	nutzerID, fehler := a.ablage.NutzerZuKonto(ctx, kennung.Anbieter, kennung.Subjekt)

	switch {
	case fehler == nil:
		return nutzerID, nil

	// **„Gibt es nicht" ist der Normalfall**, nicht die Panne: so sieht die
	// erste Anmeldung aus. Jeder andere Fehler reißt ab.
	case !errors.Is(fehler, domain.ErrNotFound):
		return 0, fehler
	}

	return a.neuerNutzer(ctx, kennung)
}

// neuerNutzer legt einen Nutzer an und verbindet die Kennung.
func (a *Anmeldung) neuerNutzer(ctx context.Context, kennung domain.Fremdkennung) (
	int64, error,
) {
	angelegt, fehler := a.ablage.NutzerAnlegen(ctx,
		domain.Nutzer{Name: namensvorschlag(kennung)})

	// **Ein belegter Name hält niemanden auf.** Zwei Menschen heißen bei zwei
	// Anbietern gleich, und der Anzeigename ist im Bestand eindeutig
	// (Werkstück 1) — der zweite käme sonst nicht herein, obwohl an seiner
	// Anmeldung nichts falsch ist.
	//
	// **Die Abweichung ist sichtbar**: er sieht seinen Namen mit dem Anbieter
	// dahinter und kann ihn ändern.
	if errors.Is(fehler, domain.ErrConflict) {
		angelegt, fehler = a.ablage.NutzerAnlegen(ctx, domain.Nutzer{
			Name: namensvorschlag(kennung) + " (" + kennung.Anbieter + ")",
		})
	}

	if fehler != nil {
		return 0, fehler
	}

	if _, verbindfehler := a.ablage.KontoVerbinden(ctx, domain.Anbieterkonto{
		NutzerID:    angelegt.ID,
		Anbieter:    kennung.Anbieter,
		Subjekt:     kennung.Subjekt,
		Anzeigename: kennung.Anzeigename,
	}); verbindfehler != nil {
		return 0, verbindfehler
	}

	return angelegt.ID, nil
}

// namensvorschlag gibt den Namen, unter dem ein neuer Nutzer entsteht.
//
// **Ein Rückfall, weil ein Nutzer einen Namen haben muss**
// ([domain.Nutzer.Validate]) — und wer sein Profil beim Anbieter nicht
// ausgefüllt hat, soll deswegen nicht draußen bleiben.
func namensvorschlag(kennung domain.Fremdkennung) string {
	if name := strings.TrimSpace(kennung.Anzeigename); name != "" {
		return name
	}

	return kennung.Anbieter + ":" + kennung.Subjekt
}
