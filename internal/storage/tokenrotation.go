// **Rotation, letzte Nutzung und Sammelwiderruf im Bestand** — Werkstück O2
// von Phase 5.2 (#1426),
// nach [Vermerk 0080](../../docs/entscheidungen/0080-maschinenzugaenge-und-widerrufbare-sitzungen.md).
//
// # Warum die Rotation ein `UPDATE` ist und kein `INSERT`
//
// **Weil das Token dasselbe bleibt** — Kennung, Bezeichnung, Platz in jeder
// Liste. Nur das Geheimnis wechselt: das geltende rückt in die Spalte des
// Vorgängers, das neue tritt an seine Stelle, und die Frist sagt, wie lange
// der Vorgänger noch zählt. Ein zweiter Ruf schiebt den alten Vorgänger
// heraus — es gibt genau einen, und das ist die Zusage aus dem Issue:
// wiederholte Rotation häuft nichts an.

package storage

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// TokenRotieren setzt ein neues gehütetes Geheimnis und hebt das bisherige
// bis zur genannten Frist auf.
func (a *Zugriffstokenablage) TokenRotieren(
	ctx context.Context, tokenID int64, gehuetet, vorgaengerBis string,
) (domain.Zugriffstoken, error) {
	ergebnis, err := a.db.ExecContext(ctx,
		`UPDATE zugriffstoken
		    SET vorgaenger_gehuetet = gehuetet, gehuetet = ?, vorgaenger_bis = ?,
		        fassung = fassung + 1
		  WHERE id = ?`, gehuetet, vorgaengerBis, tokenID)
	if err != nil {
		return domain.Zugriffstoken{}, uebersetze(err, "das Token %d rotieren", tokenID)
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return domain.Zugriffstoken{}, zaehlfehler
	}

	if zeilen == 0 {
		return domain.Zugriffstoken{}, fmt.Errorf(
			"%w: das Token %d gibt es nicht", domain.ErrNotFound, tokenID)
	}

	rotiert, _, lesefehler := a.TokenMitKennung(ctx, tokenID)

	return rotiert, lesefehler
}

// TokenNutzungVermerken schreibt an, wann ein Token zuletzt angenommen wurde.
//
// **Ohne die Fassung zu heben**: ein Vermerk ist keine Änderung, gegen die
// jemand einträgt (§7, Regel 2) — wer die Bezeichnung ändert, während das
// Token gerade benutzt wird, soll keinen Konflikt sehen.
func (a *Zugriffstokenablage) TokenNutzungVermerken(
	ctx context.Context, tokenID int64, zeitpunkt string,
) error {
	ergebnis, err := a.db.ExecContext(ctx,
		`UPDATE zugriffstoken SET letzte_nutzung = ? WHERE id = ?`, zeitpunkt, tokenID)
	if err != nil {
		return uebersetze(err, "die Nutzung des Tokens %d vermerken", tokenID)
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return zaehlfehler
	}

	if zeilen == 0 {
		return fmt.Errorf("%w: das Token %d gibt es nicht", domain.ErrNotFound, tokenID)
	}

	return nil
}

// TokenAlleLoeschen zieht jedes Token eines Profils zurück und sagt, wie
// viele es waren.
//
// **Null ist kein Fehler**: wer alle Zugänge eines Profils schließt, das
// keine hat, hat danach, was er wollte.
func (a *Zugriffstokenablage) TokenAlleLoeschen(
	ctx context.Context, profilID int64,
) (int64, error) {
	ergebnis, err := a.db.ExecContext(ctx,
		`DELETE FROM zugriffstoken WHERE profil_id = ?`, profilID)
	if err != nil {
		return 0, uebersetze(err, "die Token des Profils %d zurückziehen", profilID)
	}

	return betroffen(ergebnis)
}
