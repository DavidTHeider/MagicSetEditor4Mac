// **Wer in welches Profil darf** — Werkstück 11 von Phase 4.12
// (#1351).
//
// # Die eine Frage, die dieser Dienst beantwortet
//
// **„Welche Rolle hat dieser Mensch an diesem Profil?"** Alles andere folgt
// daraus: der Wechsel verlangt [domain.RolleMitarbeiter], das Vergeben einer
// Freigabe [domain.RolleVerwalter], und die Übersicht zeigt, was mindestens
// [domain.RolleLeser] hergibt.
//
// # Wer „dieser Mensch" ist
//
// **Der Eigner des gezeigten Profils.** Diese Anwendung hat keinen Begriff
// einer Sitzung neben dem Profilzeiger — und einen zu erfinden hieße, zwei
// Wahrheiten darüber zu führen, wer gerade arbeitet.

package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Freigabenablage ist die Naht zum Bestand.
type Freigabenablage interface {
	Freigaben(ctx context.Context, profilID int64) ([]domain.Profilfreigabe, error)
	FreigabenFuer(ctx context.Context, nutzerID int64) (
		[]domain.Profilfreigabe, error)
	FreigabeSetzen(ctx context.Context, freigabe domain.Profilfreigabe) error
	FreigabeLoeschen(ctx context.Context, profilID, nutzerID int64) error
}

// Freigabenverwaltung sagt, wer was darf.
type Freigabenverwaltung struct {
	ablage  Freigabenablage
	profile Nutzerablage
}

// NeueFreigabenverwaltung baut sie.
func NeueFreigabenverwaltung(
	ablage Freigabenablage, profile Nutzerablage,
) *Freigabenverwaltung {
	return &Freigabenverwaltung{ablage: ablage, profile: profile}
}

// Rolle gibt die Rolle eines Nutzers an einem Profil.
//
// # Die Reihenfolge ist die Aussage
//
//  1. **Der Eigner ist Verwalter seines eigenen Profils** — ohne Zeile im
//     Bestand. Eine Freigabe an sich selbst wäre eine Zeile, die niemand
//     löschen dürfte, und ein Bestand, in dem sie fehlt, wäre kaputt.
//  2. **Sonst zählt die Freigabe**, wenn es eine gibt.
//  3. **Sonst die Sichtbarkeit**: ein Profil, das `bestand` ist, gibt jedem
//     [domain.RolleLeser].
//  4. **Sonst gar keine** — und das ist kein Fehler, sondern die Antwort.
func (v *Freigabenverwaltung) Rolle(
	ctx context.Context, nutzerID int64, profil domain.Profil,
) (domain.Rolle, error) {
	rolle, _, fehler := v.rolleUndQuelle(ctx, nutzerID, profil)

	return rolle, fehler
}

// Eintrittsrolle gibt die Rolle, die zum **Hineingehen** berechtigt.
//
// # Warum die Sichtbarkeit dabei nicht zählt
//
// **Weil sie etwas anderes verspricht.** [domain.SichtbarImBestand] sagt: „der
// Name steht in der Übersicht, der Inhalt nicht" — so steht es dort seit
// Werkstück 11, und so hat es jeder gelesen, der die Einstellung gesetzt hat.
//
// **Bis Werkstück 16 machte das keinen Unterschied**: `leser` kam ohnehin
// nicht hinein, und beide Wege endeten am selben Nein. Seit `leser`
// hineinkommt, endeten sie es nicht mehr — und eine Sichtbarkeit, die
// plötzlich den Inhalt öffnet, wäre eine stillschweigende Änderung
// bestehender Semantik (§15, Regel 4).
//
// **Wer den Inhalt öffnen will, vergibt eine Freigabe.** Das ist eine
// Handlung, und sie steht in einer Liste, die man ansehen kann.
func (v *Freigabenverwaltung) Eintrittsrolle(
	ctx context.Context, nutzerID int64, profil domain.Profil,
) (domain.Rolle, error) {
	rolle, ausSichtbarkeit, fehler := v.rolleUndQuelle(ctx, nutzerID, profil)
	if fehler != nil || ausSichtbarkeit {
		return "", fehler
	}

	return rolle, nil
}

// rolleUndQuelle sagt zusätzlich, ob die Rolle **nur** aus der Sichtbarkeit
// kommt.
//
// **Ein Zug und nicht zwei**: die Freigaben werden einmal gelesen, und die
// beiden Fragen darüber unterscheiden sich in einem Wahrheitswert.
func (v *Freigabenverwaltung) rolleUndQuelle(
	ctx context.Context, nutzerID int64, profil domain.Profil,
) (rolle domain.Rolle, ausSichtbarkeit bool, fehler error) {
	if profil.NutzerID == nutzerID {
		return domain.RolleVerwalter, false, nil
	}

	freigaben, lesefehler := v.ablage.Freigaben(ctx, profil.ID)
	if lesefehler != nil {
		return "", false, lesefehler
	}

	for _, freigabe := range freigaben {
		if freigabe.NutzerID == nutzerID {
			return freigabe.Rolle, false, nil
		}
	}

	if domain.Sichtbarkeit(profil.Sichtbarkeit) == domain.SichtbarImBestand {
		return domain.RolleLeser, true, nil
	}

	return "", false, nil
}

// Verlangt weist ab, wenn die Rolle nicht reicht.
//
// **Ein eigener Ruf und keine Wiederholung an jeder Stelle**: die Absage soll
// überall gleich klingen, und der Satz, der sagt, was verlangt war, gehört an
// **eine** Stelle.
func (v *Freigabenverwaltung) Verlangt(
	ctx context.Context, nutzerID int64, profil domain.Profil, verlangt domain.Rolle,
) error {
	rolle, fehler := v.Rolle(ctx, nutzerID, profil)
	if fehler != nil {
		return fehler
	}

	if !rolle.Reicht(verlangt) {
		return domain.Freigabefehler(verlangt)
	}

	return nil
}

// Freigaben gibt die Freigaben eines Profils.
//
// **Ein Durchreichen und keine Entscheidung.** Wer sie **sehen** darf, ist
// eine Frage der Fläche — sie fragt vorher nach der eigenen Rolle.
func (v *Freigabenverwaltung) Freigaben(ctx context.Context, profilID int64) (
	[]domain.Profilfreigabe, error,
) {
	return v.ablage.Freigaben(ctx, profilID)
}

// FreigabeSetzen legt eine Freigabe an oder schreibt ihre Rolle fort.
//
// # Zwei Zusagen, die hier und nicht im Bestand stehen
//
//  1. **Ein Profil kann nicht an seinen eigenen Eigner freigegeben werden.**
//     Er ist [domain.RolleVerwalter] ohne Zeile; eine Zeile daneben wäre eine,
//     die ihm etwas nehmen könnte, das er ohnehin hat.
//  2. **Das Profil muss es geben.** Der Fremdschlüssel sagt es auch — aber als
//     SQLite-Meldung, und die erreicht die Oberfläche nie (§8).
func (v *Freigabenverwaltung) FreigabeSetzen(
	ctx context.Context, freigabe domain.Profilfreigabe,
) error {
	if fehler := freigabe.Validate(); fehler != nil {
		return fehler
	}

	profil, fehler := v.profilSuchen(ctx, freigabe.ProfilID)
	if fehler != nil {
		return fehler
	}

	if profil.NutzerID == freigabe.NutzerID {
		return domain.NewFieldError("nutzerID", domain.SchluesselValidierung,
			"Diesem Profil gehört dieser Mensch schon — er verwaltet es.")
	}

	return v.ablage.FreigabeSetzen(ctx, freigabe)
}

// FreigabeLoeschen zieht eine Freigabe zurück.
func (v *Freigabenverwaltung) FreigabeLoeschen(
	ctx context.Context, profilID, nutzerID int64,
) error {
	return v.ablage.FreigabeLoeschen(ctx, profilID, nutzerID)
}

// profilSuchen holt ein Profil über die Nutzerübersicht.
func (v *Freigabenverwaltung) profilSuchen(ctx context.Context, profilID int64) (
	domain.Profil, error,
) {
	nutzer, fehler := v.profile.Nutzer(ctx)
	if fehler != nil {
		return domain.Profil{}, fehler
	}

	for _, einer := range nutzer {
		profile, lesefehler := v.profile.Profile(ctx, einer.ID)
		if lesefehler != nil {
			return domain.Profil{}, lesefehler
		}

		for _, profil := range profile {
			if profil.ID == profilID {
				return profil, nil
			}
		}
	}

	return domain.Profil{}, fmt.Errorf(
		"%w: das Profil %d führt der Bestand nicht", domain.ErrNotFound, profilID)
}
