package storage

// **Löschen und Kopieren einer Partie** — PLAY-03.AK-05
// (#1711).
//
// # Eine eigene Datei
//
// **Der Schnitt folgt der Frage** (§2): `partien.go` beantwortet, *was* in
// einer Partie steht und wie sie fortschreibt; hier steht, wie eine
// **verschwindet** und wie aus einer zwei werden. Der Anlass war auch die
// Zeilengrenze — `partien.go` stand bei 293 Zeilen.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Loeschen nimmt eine Partie aus dem Bestand.
//
// **Die Ereignisse gehen mit** — `ON DELETE CASCADE` an `partieereignisse`
// und `partiefreigaben` (Migration 0064). Ein Strom ohne seine Partie wäre
// eine Kette, an deren Anfang nichts steht.
//
// **Ob sie gelöscht werden darf, entscheidet [domain.Loeschbar]** und nicht
// diese Ebene: die Regel gehört zur Partie und nicht zur Tafel, in der sie
// liegt (§2).
//
// **Eine Partie, die es nicht gibt, ist ein Befund und kein stiller
// Erfolg.** Wer zweimal löscht, hat beim zweiten Mal etwas anderes gemeint.
func (a *Partieablage) Loeschen(ctx context.Context, kennung string) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	ergebnis, fehler := a.db.ExecContext(ctx,
		`DELETE FROM partien WHERE id = ? AND profil_id = ?`,
		kennung, profilID)
	if fehler != nil {
		return uebersetze(fehler, "die Partie löschen")
	}

	// **Der Fehler von `RowsAffected` ist keiner, der eintreten kann** —
	// derselbe Grund wie bei `alsPartiejson`: der Treiber zählt, was er
	// gerade selbst getan hat. Er wird trotzdem gelesen, weil ein
	// verworfener Fehler wie eine geprüfte Zusage aussähe (§10).
	getroffen, zaehlfehler := ergebnis.RowsAffected()
	if zaehlfehler != nil {
		return uebersetze(zaehlfehler, "die gelöschten Zeilen zählen")
	}

	if getroffen == 0 {
		return fmt.Errorf("%w: die Partie %q gibt es in diesem Profil nicht",
			domain.ErrNotFound, kennung)
	}

	return nil
}

// Kopieren legt eine Partie als neuen Übungszustand daneben.
//
// **Ohne Ereignisse und ohne Einwilligungen** — der Grund steht bei
// [domain.Kopiere]. Hier ist es eine Folge davon, dass nur die eine Zeile
// geschrieben wird.
//
// **Die neue Kennung kommt von außen.** Woher eine Partiekennung stammt,
// entscheidet der Dienst; eine Tafel, die sie erfände, hätte zwei Quellen
// für dieselbe Zusage.
func (a *Partieablage) Kopieren(
	ctx context.Context, kennung, neueKennung string,
) (domain.Partie, error) {
	vorlage, fehler := a.Partie(ctx, kennung)
	if fehler != nil {
		return domain.Partie{}, fehler
	}

	return a.Anlegen(ctx, domain.Kopiere(vorlage, neueKennung))
}

// Offenlegungszahl sagt, für wie viele Sitze eine Einwilligung vorliegt.
//
// **Eine Zahl und keine Liste** — was darin steht, geht die Löschregel nichts
// an; dass es etwas gibt, schon ([domain.Loeschbar]).
func (a *Partieablage) Offenlegungszahl(
	ctx context.Context, partieID string,
) (int, error) {
	// **Nur aus dem gezeigten Profil** (#1815).
	if fehler := partieImProfil(ctx, a.db, partieID); fehler != nil {
		return 0, fehler
	}

	var zahl int

	// **Kein Sonderfall für „keine Zeile"**: `count(*)` gibt immer eine,
	// auch wenn sie null trägt. Ein Zweig für `sql.ErrNoRows` stünde hier
	// ungeprüft da und sähe aus wie eine geprüfte Zusage (§10).
	fehler := a.db.QueryRowContext(ctx,
		`SELECT count(*) FROM partiefreigaben WHERE partie_id = ?`,
		partieID).Scan(&zahl)
	if fehler != nil {
		return 0, uebersetze(fehler, "die Freigaben zählen")
	}

	return zahl, nil
}
