package storage

// **Was im Bestand herumliegt — gezählt, benannt und mit einer Revision** —
// Werkstück O8 von Phase 5.2
// (#1432).
//
// # Warum die Analyse hier steht und nicht in `service`
//
// **Weil sie SQL ist.** Was eine verwaiste Zeile ist, steht als Abfrage da;
// eine Schicht darüber, die dieselbe Frage über zwölf Repositories stellte,
// wäre zehnmal langsamer und einmal falsch.
//
// **Was `service` entscheidet, ist die Reihenfolge**: analysieren, dem
// Menschen zeigen, abgleichen, sichern, ausführen, nachsehen.
//
// # Und warum die Revision zweimal geprüft wird
//
// Einmal in `domain.Pflegeabgleich` gegen den Plan, den ein Mensch **gesehen**
// hat — das ist die Zusage „zwischen Vorschau und Ausführung geänderte Daten
// stoppen die Maßnahme". Und einmal **in derselben Transaktion** wie das
// Löschen, gegen den Augenblick des Schreibens. Das sind zwei verschiedene
// Fragen: die erste heißt „hat sich seit dem Anschauen etwas geändert?", die
// zweite „ändert sich gerade etwas?".

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// pflegestelle ist eine Auffälligkeit samt ihren zwei Abfragen.
//
// **Zwei und nicht eine.** `kennungen` sagt, **welche** Zeilen es betrifft —
// daraus entsteht die Revision und die Beispiele; `loeschen` räumt sie weg.
// Eine Abfrage für beides ginge nicht: die eine liest, die andere schreibt.
type pflegestelle struct {
	art    string
	klasse domain.Pflegeklasse

	// kennungen gibt je betroffener Zeile eine lesbare Kennung.
	kennungen string

	// loeschen ist leer, wo nichts gelöscht wird — bei [domain.PflegeNurInformieren].
	loeschen string

	// zeitgebunden sagt, dass die Abfrage einen Zeitpunkt braucht.
	zeitgebunden bool

	// profilgebunden sagt, dass die Abfrage auf das laufende Profil klemmt.
	//
	// **Die Ausnahme und nicht die Regel** — und beides hat einen Grund. Die
	// meisten Stellen zählen **verwaiste** Zeilen, also gerade die, die zu
	// keinem Profil mehr gehören; eine Klausel schlösse sie aus. Wo eine
	// Zeile ihr Profil aber hat — eine beendete Partie zum Beispiel —, wäre
	// eine Pflege ohne Klausel der Griff in einen fremden Bestand.
	profilgebunden bool

	// klageIstBefund macht aus einem Fehler der Abfrage einen Befund.
	//
	// **Nur wo die Abfrage SQLite über die Datei selbst fragt** — Hinweis 10
	// des Nachlaufs zum Zweitblick über Phase 5.2. `pragma_integrity_check`
	// meldet leichte Schäden als Zeile und schwere, indem es **scheitert**:
	// `database disk image is malformed`. Nachgemessen am 2026-09-11 an einem
	// Bestand mit zweihundert umgedrehten Byte — die ganze Analyse endete mit
	// einer rohen SQLite-Meldung (§8), und der Befund, für den es diese Stelle
	// gibt, entstand nie.
	//
	// **Für die zehn anderen Stellen gilt das nicht**: dort wäre ein Fehler
	// eine fehlende Tafel oder eine geschlossene Verbindung, und daraus einen
	// Befund über den Bestand zu machen verschlucke ihn.
	klageIstBefund bool
}

// Pflegeanalyse sieht nach, was im Bestand herumliegt — **rein lesend**.
func Pflegeanalyse(ctx context.Context, db *sql.DB, jetzt time.Time) (
	[]domain.Pflegebefund, error,
) {
	befunde := []domain.Pflegebefund{}

	for _, stelle := range pflegestellen() {
		kennungen, fehler := stelle.lesen(ctx, db, jetzt)
		if fehler != nil {
			if !stelle.klageIstBefund {
				return nil, fehler
			}

			// **Ein Scheitern ist hier die Antwort** und kein Abbruch: die
			// Frage lautete, ob die Datei heil ist.
			kennungen = []string{stelle.art + ":" + fehler.Error()}
		}

		if len(kennungen) == 0 {
			continue
		}

		befunde = append(befunde, domain.Pflegebefund{
			Art: stelle.art, Klasse: stelle.klasse, Anzahl: len(kennungen),
			Revision: revisionAus(kennungen), Beispiele: beispiele(kennungen),
		})
	}

	return befunde, nil
}

// leser ist, was `lesen` braucht — eine Verbindung oder eine Transaktion.
type leser interface {
	QueryContext(ctx context.Context, abfrage string, werte ...any) (*sql.Rows, error)
}

// lesen holt die Kennungen der betroffenen Zeilen.
func (p pflegestelle) lesen(ctx context.Context, aus leser, jetzt time.Time) (
	[]string, error,
) {
	werte, fehler := p.werte(ctx, jetzt)
	if fehler != nil {
		return nil, fehler
	}

	return p.lesenMit(ctx, aus, werte)
}

// lesenMit liest mit schon bestimmten Platzhaltern.
//
// **Getrennt von [pflegestelle.lesen], weil [pflegestelle.ausfuehren] sie
// vorher braucht.** Die löschende Abfrage bekommt dieselben Werte wie die
// lesende; wer sie zweimal bestimmte, hätte einen zweiten Fehlerweg für
// dieselbe Frage — und der zweite wäre unbetreten, weil der erste ihn schon
// gefangen hat (§10).
func (p pflegestelle) lesenMit(
	ctx context.Context, aus leser, werte []any,
) ([]string, error) {
	zeilen, fehler := aus.QueryContext(ctx, p.kennungen, werte...)
	if fehler != nil {
		return nil, uebersetze(fehler, "den Bestand auf %q durchsehen", p.art)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird über Err()

	kennungen := []string{}

	for zeilen.Next() {
		var kennung string
		if lesefehler := zeilen.Scan(&kennung); lesefehler != nil {
			return nil, uebersetze(lesefehler, "eine Zeile zu %q lesen", p.art)
		}

		kennungen = append(kennungen, kennung)
	}

	if fehler := zeilen.Err(); fehler != nil {
		return nil, uebersetze(fehler, "den Bestand auf %q durchsehen", p.art)
	}

	return kennungen, nil
}

// werte gibt die Platzhalter einer Stelle — **in fester Reihenfolge**.
//
// **Erst das Profil, dann die Zeit.** Zwei Abfragen derselben Stelle — die
// lesende und die löschende — müssen dieselbe Reihenfolge erwarten; sie an
// zwei Stellen zusammenzusetzen war genau die Doppelung, an der die zweite
// beim nächsten Merkmal abweicht (§10).
func (p pflegestelle) werte(
	ctx context.Context, jetzt time.Time,
) ([]any, error) {
	var werte []any

	if p.profilgebunden {
		profilID, fehler := ProfilAus(ctx)
		if fehler != nil {
			return nil, fehler
		}

		werte = append(werte, profilID)
	}

	if p.zeitgebunden {
		werte = append(werte, Zeitstempel(jetzt))
	}

	return werte, nil
}

// revisionAus rechnet die Prüfsumme über die betroffenen Zeilen.
//
// **Über die Zeilen und nicht über die Zahl.** Zwei Zeilen weg und zwei dazu
// ergäben dieselbe Zahl und einen anderen Bestand.
func revisionAus(kennungen []string) string {
	geordnet := make([]string, len(kennungen))
	copy(geordnet, kennungen)
	sort.Strings(geordnet)

	summe := sha256.Sum256([]byte(strings.Join(geordnet, "\n")))

	return hex.EncodeToString(summe[:])
}

// Pflegebeispiele ist, wie viele Zeilen ein Befund beispielhaft nennt.
//
// **Fünf, und die Zahl ist eine Wahl.** Genug, um zu sehen, worum es geht;
// wenig genug, dass eine Antwort nicht die halbe Tabelle trägt.
const Pflegebeispiele = 5

func beispiele(kennungen []string) []string {
	if len(kennungen) <= Pflegebeispiele {
		return kennungen
	}

	return kennungen[:Pflegebeispiele]
}

// PflegeAusfuehren räumt die gewählten Stellen weg — **in einer Transaktion**.
//
// **Die Revision wird darin noch einmal geprüft.** Zwischen dem Abgleich gegen
// die Vorschau und diesem Schreiben liegt ein Augenblick, und in ihm kann eine
// andere Fläche schreiben. Was hier gegen die mitgegebene Revision gehalten
// wird, ist der Bestand **im Augenblick des Löschens**.
//
// **Und ein Konflikt an einer Stelle nimmt den anderen nichts**: jede Maßnahme
// bekommt ihre eigene Transaktion. Eine gemeinsame hieße, dass eine geänderte
// Zeile fünf erledigte Maßnahmen zurückdreht.
func PflegeAusfuehren(
	ctx context.Context, db *sql.DB, jetzt time.Time, revisionen map[string]string,
) ([]domain.Pflegezeile, error) {
	nachArt := map[string]pflegestelle{}
	for _, stelle := range pflegestellen() {
		nachArt[stelle.art] = stelle
	}

	arten := make([]string, 0, len(revisionen))
	for art := range revisionen {
		arten = append(arten, art)
	}

	sort.Strings(arten)

	zeilen := make([]domain.Pflegezeile, 0, len(arten))

	for _, art := range arten {
		stelle, kennt := nachArt[art]
		if !kennt || stelle.loeschen == "" {
			return nil, fmt.Errorf(
				"%w: %q ist keine Maßnahme, die etwas wegräumt", domain.ErrValidation, art)
		}

		zeilen = append(zeilen, stelle.ausfuehren(ctx, db, jetzt, revisionen[art]))
	}

	return zeilen, nil
}

// ausfuehren räumt eine Stelle weg — oder sagt, warum nicht.
func (p pflegestelle) ausfuehren(
	ctx context.Context, db *sql.DB, jetzt time.Time, erwartet string,
) domain.Pflegezeile {
	zeile := domain.Pflegezeile{Art: p.art}

	werte, wertefehler := p.werte(ctx, jetzt)
	if wertefehler != nil {
		zeile.Ausgang, zeile.Grund = domain.PflegeGescheitert, wertefehler.Error()

		return zeile
	}

	geschaeft, fehler := db.BeginTx(ctx, nil)
	if fehler != nil {
		zeile.Ausgang, zeile.Grund = domain.PflegeGescheitert, fehler.Error()

		return zeile
	}

	defer func() { _ = geschaeft.Rollback() }() //nolint:errcheck // nach einem Commit ohne Wirkung

	kennungen, lesefehler := p.lesenMit(ctx, geschaeft, werte)
	if lesefehler != nil {
		zeile.Ausgang, zeile.Grund = domain.PflegeGescheitert, lesefehler.Error()

		return zeile
	}

	if len(kennungen) == 0 {
		zeile.Ausgang = domain.PflegeSchonWeg

		return zeile
	}

	if revisionAus(kennungen) != erwartet {
		zeile.Ausgang, zeile.Grund = domain.PflegeKonflikt, domain.PflegegrundGeaendert

		return zeile
	}

	if _, schreibfehler := geschaeft.ExecContext(ctx, p.loeschen, werte...); schreibfehler != nil {
		zeile.Ausgang, zeile.Grund = domain.PflegeGescheitert, schreibfehler.Error()

		return zeile
	}

	if abschlussfehler := geschaeft.Commit(); abschlussfehler != nil {
		zeile.Ausgang, zeile.Grund = domain.PflegeGescheitert, abschlussfehler.Error()

		return zeile
	}

	zeile.Ausgang, zeile.Anzahl = domain.PflegeErledigt, len(kennungen)

	return zeile
}
