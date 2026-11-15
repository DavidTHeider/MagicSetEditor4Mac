package storage

// **Der Deltaimport** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Warum eine neue Generation und trotzdem inkrementell
//
// **Weil beides im Issue steht.** „Inkrementelle Änderungen" und „neuer Index
// entsteht separat und wird erst nach Prüfung atomar aktiviert" sind kein
// Widerspruch: was sich nicht geändert hat, wird **kopiert** statt neu
// eingebettet — das spart den teuren Teil —, und sichtbar wird alles
// zusammen.
//
// **Der teure Teil ist die Einbettung.** Ein Vektor kostet einen Lauf durch
// das Modell; eine Zeile zu kopieren kostet einen `INSERT`. Bei 30.000
// Abschnitten, von denen sich 200 geändert haben, ist das der Unterschied
// zwischen Minuten und Sekunden.

import (
	"context"
	"database/sql"
	"errors"
)

// Uebernehmen kopiert die Abschnitte der aktiven Generation in eine neue.
//
// # Wofür es das gibt
//
// **Für den Deltaimport.** Das Issue verlangt „inkrementelle Änderungen" und
// zugleich „neuer Index entsteht separat und wird erst nach Prüfung atomar
// aktiviert" — das ist kein Widerspruch, sondern diese Zeile: was sich nicht
// geändert hat, wird **kopiert** statt neu eingebettet, und sichtbar wird
// alles zusammen.
//
// **Der Vektor wandert mit.** Er stammt aus demselben Vektorraum; dass das
// stimmt, prüft der Aufrufer am Manifest ([domain.Vektorraum.PasstZu]) und
// nicht diese Tafel.
//
// **`ausser` nennt, was neu geschrieben wird** — geändert, gelöscht oder
// hinzugekommen. Was darin steht, wird nicht kopiert; wer es danach nicht
// schreibt, hat es entfernt.
func (r *Retrievalindex) Uebernehmen(
	ctx context.Context, nach int64, ausser []string,
) (int, error) {
	uebernommen := 0

	fehler := imRetrievalzug(ctx, r.db, func(tx *sql.Tx) error {
		// **Über den Zug und nicht über die Verbindung.** SQLite hält beim
		// Schreiben eine Verbindung; ein Lauf daneben wartet auf eine, die er
		// selbst blockiert — und wartet still (siehe [zeilenleser]).
		aktiv, fehler := aktiveGenerationIn(ctx, tx)
		if fehler != nil {
			return fehler
		}

		gezaehlt, fehler := uebernimmZeilen(ctx, tx, aktiv, nach, ausser)
		uebernommen = gezaehlt

		return fehler
	})

	return uebernommen, fehler
}

// uebernimmZeilen kopiert Dokument, Volltext und Vektor in einem Zug.
//
// **Drei Anweisungen und nicht eine**: die Volltexttafel ist virtuell und
// kennt die Kennung des Dokuments nur als Spalte, und der Vektor hängt an der
// **neuen** Zeilennummer. Ein `INSERT … SELECT` über alle drei gäbe es nicht.
func uebernimmZeilen(
	ctx context.Context, tx *sql.Tx, von, nach int64, ausser []string,
) (int, error) {
	uebergangen := map[string]bool{}
	for _, kennung := range ausser {
		uebergangen[kennung] = true
	}

	quellen, fehler := zuUebernehmen(ctx, tx, von, uebergangen)
	if fehler != nil {
		return 0, fehler
	}

	for _, quelle := range quellen {
		if fehler := kopiereZeile(ctx, tx, quelle, nach); fehler != nil {
			return 0, fehler
		}
	}

	return len(quellen), nil
}

// zuUebernehmen sammelt die Zeilennummern, die kopiert werden.
//
// **Die Zeilen werden erst zu Ende gelesen und dann kopiert.** Ein `INSERT`
// in dieselbe Tafel, während die Abfrage darüber noch läuft, ist bei SQLite
// ein Lauf mit sich selbst — die neuen Zeilen kämen in der Abfrage wieder
// hoch, und der Kopiervorgang liefe ohne Ende.
func zuUebernehmen(
	ctx context.Context, tx *sql.Tx, von int64, uebergangen map[string]bool,
) ([]int64, error) {
	zeilen, fehler := tx.QueryContext(ctx, `
		SELECT id, kennung FROM retrieval_dokumente
		WHERE generation = ? ORDER BY id`, von)
	if fehler != nil {
		return nil, uebersetze(fehler, "die alte Generation lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird unten

	quellen := make([]int64, 0)

	for zeilen.Next() {
		var (
			id      int64
			kennung string
		)

		if fehler := zeilen.Scan(&id, &kennung); fehler != nil {
			return nil, uebersetze(fehler,
				"eine Zeile der alten Generation lesen")
		}

		if !uebergangen[kennung] {
			quellen = append(quellen, id)
		}
	}

	return quellen, uebersetze(zeilen.Err(), "die alte Generation lesen")
}

// kopiereZeile hängt eine Zeile samt Volltext und Vektor an die neue Generation.
func kopiereZeile(ctx context.Context, tx *sql.Tx, quelle, nach int64) error {
	ergebnis, fehler := tx.ExecContext(ctx, `
		INSERT INTO retrieval_dokumente
			(generation, kennung, quelle, bezugsadresse, spiel, kennungssystem,
			 fachkennung, semantik, regelnummer, urheber, sprache, titel, text,
			 quellrevision, quelldatum, abgerufen, inhaltshash, merkmale,
			 abschnitt, von, bis, profil_id, sichtbarkeit)
		SELECT ?, kennung, quelle, bezugsadresse, spiel, kennungssystem,
			 fachkennung, semantik, regelnummer, urheber, sprache, titel, text,
			 quellrevision, quelldatum, abgerufen, inhaltshash, merkmale,
			 abschnitt, von, bis, profil_id, sichtbarkeit
		FROM retrieval_dokumente WHERE id = ?`, nach, quelle)
	if fehler != nil {
		return uebersetze(fehler, "eine Zeile übernehmen")
	}

	kennung, zaehlfehler := eingefuegt(ergebnis)

	// **Dieselbe verbundene Form wie beim Schreiben** (§10): drei `if` wären
	// drei Zweige, von denen der erste nie betreten wird.
	_, volltext := tx.ExecContext(ctx, `
		INSERT INTO retrieval_volltext (titel, text, dokument_id)
		SELECT titel, text, ? FROM retrieval_dokumente WHERE id = ?`,
		kennung, quelle)

	_, vektor := tx.ExecContext(ctx, `
		INSERT INTO retrieval_vektoren (dokument_id, vektor)
		SELECT ?, vektor FROM retrieval_vektoren WHERE dokument_id = ?`,
		kennung, quelle)

	return uebersetze(errors.Join(zaehlfehler, volltext, vektor),
		"eine Zeile übernehmen")
}
