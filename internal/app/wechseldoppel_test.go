// Das Doppel des Profilwechsels — herausgezogen aus `freigaben_test.go`,
// weil es dort über die Zeilengrenze aus §2 ging.
//
// **Der Schnitt folgt der Frage**: dort stehen die Läufe über Freigaben, hier
// steht die Vorrichtung, die sie **und** die Läufe über die nur lesende
// Verbindung, den Nutzerwechsel und die Bindungen benutzen.
package app_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// wechseldoppel sagt, in welchem Profil gearbeitet wird.
type wechseldoppel struct {
	gezeigt domain.Profil
	fehler  error

	// nachWechsel klemmt erst, **nachdem** der Nutzerwechsel geglückt ist —
	// Werkstück 17 von Phase 4.12.
	//
	// **Getrennt von [wechseldoppel.fehler]**, weil `App.NutzerWechseln`
	// danach noch einmal fragt, welches Profil jetzt gezeigt wird. Ein
	// gemeinsamer Fehler ließe den Ruf nie so weit kommen, und der zweite
	// Zweig bliebe unbetreten (§3).
	nachWechsel error

	// gewechselt zählt die Nutzerwechsel.
	gewechselt int

	// rolle ist, was [wechseldoppel.RolleImGezeigten] gibt — leer heißt
	// Verwalter, die Lage der meisten Läufe hier.
	rolle domain.Rolle

	// rollenfehler klagt **nur** an der Rollenfrage — Werkstück 16 von Phase
	// 4.12.
	//
	// **Getrennt von [wechseldoppel.fehler]**, weil `leserverbindungRichten`
	// erst nach einem geglückten Wechsel fragt. Ein gemeinsamer Fehler ließe
	// den Ruf nie so weit kommen, und der Zweig dahinter bliebe unbetreten —
	// dieselbe Gestalt wie bei [freigabenfuehrer.datenfehler].
	rollenfehler error
}

func (d *wechseldoppel) Gezeigt(context.Context) (domain.Profil, error) {
	if d.gewechselt > 0 && d.nachWechsel != nil {
		return domain.Profil{}, d.nachWechsel
	}

	return d.gezeigt, d.fehler
}

// Arbeitsprofil ist hier dasselbe wie [wechseldoppel.Gezeigt] — A14 (#1453).
//
// **Weil dieses Doppel keinen Kontext liest.** Was die beiden unterscheidet,
// ist genau das: die Ablage sieht am Ruf ein Profil oder nicht. Wo ein Fall
// den Unterschied braucht, steht ein echter Bestand (`nutzernaht_test.go`).
func (d *wechseldoppel) Arbeitsprofil(ctx context.Context) (domain.Profil, error) {
	return d.Gezeigt(ctx)
}

func (d *wechseldoppel) Wechseln(_ context.Context, _ int64, _ string) (
	domain.Profil, error,
) {
	return d.gezeigt, d.fehler
}

// WerArbeitet gibt den Eigner des gezeigten Profils — Werkstück 17 von Phase
// 4.12.
//
// **Für dieses Doppel ist das dasselbe**, und das ist Absicht: die Läufe hier
// prüfen die Freigabenseite und nicht den Identitätswechsel. Wer den prüft,
// steht in `internal/service/nutzerwechsel_test.go`.
func (d *wechseldoppel) WerArbeitet(context.Context) (int64, error) {
	return d.gezeigt.NutzerID, d.fehler
}

// RolleImGezeigten meldet Verwalter — Werkstück 16 von Phase 4.12.
//
// **Das ist die Lage, in der die Läufe hier stehen**: sie prüfen die
// Freigabenseite im eigenen Profil, und dort ist man Verwalter. Wer den
// Nur-Lese-Weg prüft, steht in `nurlesen_test.go`.
func (d *wechseldoppel) RolleImGezeigten(context.Context) (domain.Rolle, error) {
	if d.rollenfehler != nil {
		return "", d.rollenfehler
	}

	if d.rolle != "" {
		return d.rolle, d.fehler
	}

	return domain.RolleVerwalter, d.fehler
}

func (d *wechseldoppel) Anmeldeprofil(context.Context, int64) (int64, error) {
	if d.fehler != nil {
		return 0, d.fehler
	}

	d.gewechselt++

	return 1, nil
}

func (d *wechseldoppel) NutzerWechseln(_ context.Context, _ int64, _ string) error {
	if d.fehler != nil {
		return d.fehler
	}

	d.gewechselt++

	return nil
}
