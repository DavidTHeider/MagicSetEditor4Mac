package app

// **Die Rolle am Zielprofil** —
// #1817.
//
// # Der Befund
//
// Zugriffstoken und Vorgaben nehmen die Kennung des Profils, an dem sie
// wirken, vom Rufer. Die Matrix verlangt dort „Mitarbeiter" — durchgesetzt
// über `query_only`, und das gehört zum **gezeigten** Profil. Aus dem eigenen
// Profil heraus entstand so am 2026-09-23 ein schreibendes Token für das
// Profil eines anderen, samt Klartext; Rotieren und Sammelwiderruf gingen
// ebenso.
//
// # Die Regel
//
// **Wer an einem Profil wirkt, das er nennt, braucht die Rolle dort** —
// dieselbe Frage, die `FreigabeSetzen` seit Werkstück 11 von Phase 4.12
// stellt. Wer ein fremdes Profil ohne Faktor verwalten will, **wird** zuerst
// dessen Eigner; die Bindung selbst bleibt streng.

import (
	"context"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// rolleAm weist ab, wer am genannten Profil nicht mindestens `verlangt` hat.
func (a *App) rolleAm(ctx context.Context, profilID int64, verlangt domain.Rolle) error {
	_, fehler := a.rolleUndUrteil(ctx, profilID, verlangt)

	return fehler
}

// zumSchreibenAm verlangt für einen **Schreibweg** die Rolle „Mitarbeiter" am
// genannten Profil — und sagt im gezeigten Profil, was jeder andere
// Schreibweg dort sagt.
//
// **„Nur lesen" statt „dafür braucht es die Rolle"**, wenn der Mensch im
// gezeigten Profil gar nicht schreibt: dieselbe Antwort wie
// `Nutzerverwaltung.UeberProfil` und jeder Weg über die Nur-Lese-Verbindung
// (ID-03). Zwei Sätze für denselben Zustand ließen die Fläche raten, welcher
// gilt.
func (a *App) zumSchreibenAm(ctx context.Context, profilID int64) error {
	rolle, fehler := a.rolleUndUrteil(ctx, profilID, domain.RolleMitarbeiter)
	if !errors.Is(fehler, domain.ErrErlaubnis) {
		return fehler
	}

	if imGezeigten(ctx, profilID) && (domain.Akteur{Rolle: rolle}).Darf(domain.WirkungSchreiben) != nil {
		return fmt.Errorf("%w: in diesem Profil wird nur gelesen", domain.ErrNurLesend)
	}

	return fehler
}

// rolleUndUrteil gibt die Rolle des arbeitenden Menschen am genannten Profil
// und das Urteil über `verlangt`.
//
// # Was ein fehlender Führer hier heißt
//
// **Ohne Nutzerverwaltung gibt es kein fremdes Profil** — dieselbe Regel wie
// bei [App.ueberNutzer]. Beide Wurzeln setzen sie; eine Anwendung, die
// niemanden kennt, kennt auch kein Profil, an dem jemand die Rolle fehlte.
func (a *App) rolleUndUrteil(ctx context.Context, profilID int64, verlangt domain.Rolle) (
	domain.Rolle, error,
) {
	// **Ein Token wirkt nur an seinem Profil** (QA-09, #1740) — vor der
	// Rollenfrage, denn sie geht sonst über den Menschen am Rechner.
	if fehler := nurAnSeinemProfil(ctx, profilID); fehler != nil {
		return "", fehler
	}

	if a.nutzer == nil || a.freigaben == nil {
		return "", nil
	}

	profil, fehler := a.profilMitKennung(ctx, profilID)
	if fehler != nil {
		return "", fehler
	}

	arbeitend, fehler := a.arbeitenderNutzer(ctx)
	if fehler != nil {
		return "", fehler
	}

	rolle, fehler := a.freigaben.Rolle(ctx, arbeitend, profil)
	if fehler != nil {
		return "", fehler
	}

	if !rolle.Reicht(verlangt) {
		return rolle, domain.Freigabefehler(verlangt)
	}

	return rolle, nil
}

// rolleAmToken fragt dasselbe an dem Profil, dem ein Token gehört — **nicht**
// an dem, das gezeigt ist. Jeder Weg mit einer Tokenkennung schreibt.
func (a *App) rolleAmToken(ctx context.Context, tokenID int64) error {
	if a.nutzer == nil || a.freigaben == nil {
		return nil
	}

	profilID, fehler := a.token.ProfilDes(ctx, tokenID)
	if fehler != nil {
		return fehler
	}

	return a.zumSchreibenAm(ctx, profilID)
}

// lesbarAm lässt das Lesen von Vorgaben und Freigaben im gezeigten Profil zu
// — wer darin steht, liest, was es zeigt — und verlangt an jedem anderen
// dieselbe Rolle wie das Setzen.
func (a *App) lesbarAm(ctx context.Context, profilID int64) error {
	if imGezeigten(ctx, profilID) {
		return nil
	}

	return a.rolleAm(ctx, profilID, domain.RolleMitarbeiter)
}

// lesbarMit sagt dasselbe wie [App.lesbarAm] — mit einer Rolle, die der
// Rufer schon nachgeschlagen hat.
func lesbarMit(ctx context.Context, profilID int64, eigene domain.Rolle) error {
	if imGezeigten(ctx, profilID) || eigene.Reicht(domain.RolleMitarbeiter) {
		return nil
	}

	return domain.Freigabefehler(domain.RolleMitarbeiter)
}

// imGezeigten sagt, ob ein Ruf in genau diesem Profil steht.
func imGezeigten(ctx context.Context, profilID int64) bool {
	gezeigt, fehler := gezeigtesProfil(ctx)

	return fehler == nil && gezeigt == profilID
}
