package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Fassung zählt jede Änderung am Turnier** (§7, Regel 2).
//
// Sie stand seit Werkstück 1 in der Tabelle, im Verbund und im DTO — und
// wurde von niemandem erhöht und von niemandem geprüft. Das Fenster bekam
// eine Zahl, die immer 1 war, und hätte darauf gebaut (Befund 7 des
// Nachlaufs über Phase 4.7).
//
// **Sie deckt das Turnier als Ganzes**, nicht die einzelne Zeile: eine
// Anmeldung, ein Ausstieg, eine Runde und ein Ergebnis erhöhen dieselbe Zahl.
// Das ist gewollt — wer ein Turnier offen hat, will wissen, ob sich **etwas**
// geändert hat, und nicht, was.

// execer ist, was sowohl `*sql.DB` als auch `*sql.Tx` können.
//
// **Damit ist die Buchführung an beiden Orten dieselbe.** Eine zweite Fassung
// derselben Zeile — eine für den Zug, eine daneben — liefe auseinander (§10).
type execer interface {
	ExecContext(ctx context.Context, abfrage string, werte ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, abfrage string, werte ...any) *sql.Row
}

// fortschreiben erhöht die Fassung eines Turniers.
//
// **Das Profil kommt vom Rufer** — dieselbe Regel wie bei
// [warumNichtFortgeschrieben] darunter, und seit #1815 aus demselben Grund:
// jeder Rufer hat es schon gelesen, um zu fragen, ob ihm Turnier, Teilnehmer
// oder Paarung gehören. Ein zweites Lesen hier wäre ein Zweig, den kein
// Aufruf betreten kann (§3).
//
// **Und sie schreibt die Zeile des Protokolls** — DOM-04.AK-04 (#1699),
// Vermerk 0139. Jede Änderung geht hier durch; eine zweite Stelle für das
// Protokoll liefe auseinander. **Ein abgeschlossenes Turnier trifft sie
// nicht**: die zweite Reihe hinter der Regel im Dienst.
func fortschreiben(
	ctx context.Context, wo execer, turnierID, profilID int64, eintrag domain.Protokolleintrag,
) error {
	// **Jede Wurzelabfrage klemmt auf ein Profil** (#1351, Vermerk 0066).
	ergebnis, err := wo.ExecContext(ctx,
		`UPDATE turniere SET fassung = fassung + 1, geaendert = datetime('now')
		  WHERE id = ? AND profil_id = ? AND abgeschlossen IS NULL`, turnierID, profilID)
	if err != nil {
		return uebersetze(err, "die Fassung fortschreiben")
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return zaehlfehler
	}

	if zeilen == 0 {
		return warumNichtFortgeschrieben(ctx, wo, turnierID, 0, profilID, false)
	}

	return protokollieren(ctx, wo, turnierID, eintrag)
}

// fortschreibenGegen erhöht die Fassung — **aber nur, wenn sie stimmt**.
//
// Der Rückgabewert ist die neue Fassung. Passt der genannte Stand nicht, ist
// es ein [domain.ErrConflict] und die Zeile bleibt, wie sie war.
func fortschreibenGegen(
	ctx context.Context, wo execer, turnierID, fassung, profilID int64,
	eintrag domain.Protokolleintrag,
) (int64, error) {
	ergebnis, err := wo.ExecContext(ctx,
		`UPDATE turniere SET fassung = fassung + 1, geaendert = datetime('now')
		  WHERE id = ? AND fassung = ? AND profil_id = ? AND abgeschlossen IS NULL`,
		turnierID, fassung, profilID)
	if err != nil {
		return 0, uebersetze(err, "die Fassung fortschreiben")
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return 0, zaehlfehler
	}

	if zeilen == 0 {
		return 0, warumNichtFortgeschrieben(ctx, wo, turnierID, fassung, profilID, false)
	}

	return fassung + 1, protokollieren(ctx, wo, turnierID, eintrag)
}

// warumNichtFortgeschrieben sagt, ob das Turnier fehlt oder überholt ist.
//
// **Zwei Antworten auf denselben Fehlschlag**, dieselbe Form wie am Deck:
// „gibt es nicht" heißt aufräumen, „jemand anderes war schneller" heißt
// nachladen (§8).
//
// **Das Profil kommt vom Rufer.** Es hier ein zweites Mal aus dem Kontext zu
// holen wäre ein Zweig, den kein Aufruf betreten kann: der Rufer hat es schon
// gelesen und wäre sonst gar nicht bis hierher gekommen (§3).
//
// **Und seit DOM-04.AK-04 (#1699) eine dritte**: das Turnier steht nicht so,
// wie der Schreibweg es verlangt — abgeschlossen, wo es offen sein muss,
// oder offen, wo es abgeschlossen sein muss. `erwartetAbgeschlossen` sagt,
// welches der Schreibweg braucht.
func warumNichtFortgeschrieben(
	ctx context.Context, wo execer, turnierID, fassung, profilID int64,
	erwartetAbgeschlossen bool,
) error {
	var (
		stand         int64
		name          string
		abgeschlossen bool
	)

	err := wo.QueryRowContext(ctx,
		`SELECT fassung, name, abgeschlossen IS NOT NULL
		   FROM turniere WHERE id = ? AND profil_id = ?`,
		turnierID, profilID).Scan(&stand, &name, &abgeschlossen)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return fmt.Errorf("%w: das Turnier %d gibt es nicht",
			domain.ErrNotFound, turnierID)
	case err != nil:
		return uebersetze(err, "die Fassung nachschlagen")
	case abgeschlossen && !erwartetAbgeschlossen:
		return domain.TurnierIstAbgeschlossen(name)
	case !abgeschlossen && erwartetAbgeschlossen:
		return fmt.Errorf("%w: das Turnier »%s« ist nicht abgeschlossen",
			domain.ErrConflict, name)
	}

	return fmt.Errorf(
		"%w: dieses Turnier steht bei Fassung %d, geschrieben wurde gegen %d "+
			"— jemand anderes war schneller",
		domain.ErrConflict, stand, fassung)
}
