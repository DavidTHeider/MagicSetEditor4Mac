// **Beide Ebenen in einem Zug** — Werkstück O3 von Phase 5.2
// (#1427).
//
// # Warum eine Transaktion und nicht zwei Aufrufe
//
// **Weil ein halber Import schlimmer ist als keiner** (§8). Eine Datei bringt
// Zeilen für zwei Tafeln mit: `settings` hängt am Profil, `nutzereinstellungen`
// am Menschen. Wer sie nacheinander schriebe, hinterließe bei einem Fehler
// dazwischen einen Stand, den niemand gewählt hat und den keine Vorschau je
// gezeigt hat.
//
// # Und warum der Mensch aus einer Unterabfrage kommt
//
// **Weil zwei Abfragen zwei Stände wären** (§7, Regel 2) — dieselbe
// Überlegung wie bei [Nutzerablage.NutzerwerteZumProfil]: wer erst den Eigner
// holte und dann schriebe, träfe womöglich einen anderen Menschen, als die
// Vorschau gezeigt hat.

package storage

import (
	"context"
	"database/sql"
	"fmt"
	"maps"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// EinstellungenUebernehmen schreibt Profil- und Nutzerwerte in einem Zug.
//
// **`erwartet` ist die Revision, die der Vorschau zugrunde lag** — Befund 11
// des Zweitblicks über Phase 5.2. `service` vergleicht sie auch, aber **vor**
// der Transaktion: zwischen seinem Lesen und diesem Schreiben liegt ein
// Fenster, in dem ein anderer Prozess schreiben kann, und dessen Änderung
// wäre danach weg. Hier wird sie **im selben Geschäft** noch einmal gerechnet,
// aus den Zeilen, auf die gleich geschrieben wird.
//
// **Leer heißt: nicht vergleichen.** Diesen Weg gibt es für Aufrufer, die
// keine Vorschau gesehen haben — heute keiner; die Zusage steht hier, damit
// ein künftiger sie nicht unbemerkt bekommt.
func (e *Einstellungen) EinstellungenUebernehmen(
	ctx context.Context, profil, nutzer map[string]string, erwartet string,
) (fehler error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	// **Geprüft wird auch hier**, obwohl `service` es schon getan hat: diese
	// Tafel ist die letzte Stelle vor der Platte, und `NutzereinstellungSetzen`
	// nebenan tut dasselbe. Eine Zusage, die nur eine Ebene hoch gilt, bricht
	// an dem Tag, an dem jemand einen zweiten Aufrufer schreibt.
	if pruefung := pruefeAustauschwerte(profil, nutzer); pruefung != nil {
		return pruefung
	}

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "die Einstellungen übernehmen")
	}

	defer func() { fehler = zurueckgenommen(tx, fehler) }()

	if revisionsfehler := revisionPruefen(ctx, tx, profilID, erwartet); revisionsfehler != nil {
		return revisionsfehler
	}

	if schreibfehler := imLaufSchreiben(ctx, tx, profilID, profil, nutzer); schreibfehler != nil {
		return schreibfehler
	}

	return bestaetige(tx, "die Einstellungen übernehmen")
}

// pruefeAustauschwerte urteilt über jedes Paar, bevor eine Zeile entsteht.
func pruefeAustauschwerte(profil, nutzer map[string]string) error {
	for _, schluessel := range slices.Sorted(maps.Keys(profil)) {
		if befund := domain.PruefeEinstellung(
			schluessel, profil[schluessel]); befund != nil {
			return befund
		}
	}

	for _, schluessel := range slices.Sorted(maps.Keys(nutzer)) {
		if befund := domain.PruefeNutzereinstellung(
			schluessel, nutzer[schluessel]); befund != nil {
			return befund
		}
	}

	return nil
}

// imLaufSchreiben setzt beide Ebenen innerhalb der Transaktion.
//
// **In fester Reihenfolge** — eine `map` durchläuft Go absichtlich wechselnd,
// und zwei Läufe, die dieselben Zeilen in anderer Ordnung schreiben, sperren
// einander bei einem zweiten Schreiber in anderer Reihenfolge.
func imLaufSchreiben(
	ctx context.Context, tx *sql.Tx, profilID int64, profil, nutzer map[string]string,
) error {
	for _, schluessel := range slices.Sorted(maps.Keys(profil)) {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO settings (schluessel, wert, profil_id) VALUES (?, ?, ?)
			 ON CONFLICT (profil_id, schluessel) DO UPDATE SET wert = excluded.wert`,
			schluessel, profil[schluessel], profilID); err != nil {
			return uebersetze(err, "die Einstellung %q setzen", schluessel)
		}
	}

	for _, schluessel := range slices.Sorted(maps.Keys(nutzer)) {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO nutzereinstellungen (nutzer_id, schluessel, wert)
			 VALUES ((SELECT nutzer_id FROM profile WHERE id = ?), ?, ?)
			 ON CONFLICT (nutzer_id, schluessel) DO UPDATE SET wert = excluded.wert`,
			profilID, schluessel, nutzer[schluessel]); err != nil {
			return uebersetze(err, "die Einstellung %q des Menschen setzen", schluessel)
		}
	}

	return nil
}

// revisionPruefen hält den Stand im Geschäft gegen die erwartete Revision.
//
// **Gerechnet aus denselben Zeilen, auf die gleich geschrieben wird** — das
// ist der Unterschied zur Prüfung in `service`: die sieht einen Stand von
// vorhin, diese den, der gilt.
func revisionPruefen(ctx context.Context, tx *sql.Tx, profilID int64, erwartet string) error {
	if erwartet == "" {
		return nil
	}

	var nutzerID int64

	if err := tx.QueryRowContext(ctx,
		`SELECT nutzer_id FROM profile WHERE id = ?`, profilID).Scan(&nutzerID); err != nil {
		return uebersetze(err, "den Menschen zum Profil lesen")
	}

	imProfil, profilfehler := werteImLauf(ctx, tx,
		`SELECT schluessel, wert FROM settings WHERE profil_id = ?`, profilID)
	if profilfehler != nil {
		return profilfehler
	}

	beimMenschen, nutzerfehler := werteImLauf(ctx, tx,
		`SELECT schluessel, wert FROM nutzereinstellungen WHERE nutzer_id = ?`, nutzerID)
	if nutzerfehler != nil {
		return nutzerfehler
	}

	if domain.Standmarke(profilID, nutzerID, imProfil, beimMenschen) != erwartet {
		return fmt.Errorf(
			"%w: die Einstellungen haben sich geändert, seit du hingesehen "+
				"hast. Es wurde nichts übernommen — sieh sie dir noch einmal an",
			domain.ErrConflict)
	}

	return nil
}

// werteImLauf liest Schlüssel und Wert einer Tafel im Geschäft.
func werteImLauf(ctx context.Context, tx *sql.Tx, abfrage string, werte ...any) (
	map[string]string, error,
) {
	zeilen, err := tx.QueryContext(ctx, abfrage, werte...)
	if err != nil {
		return nil, uebersetze(err, "den Stand der Einstellungen lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // der Fehler kommt über zeilen.Err

	gelesen := map[string]string{}

	for zeilen.Next() {
		var schluessel, wert string

		if err := zeilen.Scan(&schluessel, &wert); err != nil {
			return nil, uebersetze(err, "den Stand der Einstellungen lesen")
		}

		gelesen[schluessel] = wert
	}

	return gelesen, uebersetze(zeilen.Err(), "den Stand der Einstellungen lesen")
}
