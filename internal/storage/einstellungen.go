package storage

import (
	"context"
	"database/sql"
)

// Einstellungen ist die Ablage der Anwendungseinstellungen (#245).
//
// **Schlüssel und Wert**, nicht eine Spalte je Einstellung: sonst kostete jede
// neue Einstellung eine Migration. Geprüft werden die Paare in
// `internal/domain`, nicht im Schema — ein `CHECK` hier wäre eine zweite
// Wahrheit.
type Einstellungen struct {
	db *sql.DB
}

// NeueEinstellungen gibt die Ablage über dieser Datenbank.
func NeueEinstellungen(db *sql.DB) *Einstellungen {
	return &Einstellungen{db: db}
}

// Alle gibt die gesetzten Einstellungen.
//
// **Leer und nicht nil.** Ein `nil` durchquert die Naht als `null`, und die
// Oberfläche prüfte dann darauf; eine leere Abbildung sagt dasselbe und braucht
// keinen zweiten Zweig.
func (e *Einstellungen) Alle(ctx context.Context) (map[string]string, error) {
	// **Jede Wurzelabfrage klemmt auf ein Profil** (#1351, Vermerk 0066).
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, err := e.db.QueryContext(ctx,
		`SELECT schluessel, wert FROM settings WHERE profil_id = ?`, profilID)
	if err != nil {
		return nil, uebersetze(err, "die Einstellungen lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // der Fehler kommt über zeilen.Err

	gefunden := make(map[string]string)

	for zeilen.Next() {
		var schluessel, wert string

		if err := zeilen.Scan(&schluessel, &wert); err != nil {
			return nil, uebersetze(err, "eine Einstellung lesen")
		}

		gefunden[schluessel] = wert
	}

	if err := zeilen.Err(); err != nil {
		return nil, uebersetze(err, "die Einstellungen lesen")
	}

	return gefunden, nil
}

// Setzen schreibt eine Einstellung fort.
func (e *Einstellungen) Setzen(ctx context.Context, schluessel, wert string) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	// **Der Konflikt nennt beide Spalten** — der Schlüssel ist seit
	// Wanderung 0035 `(profil_id, schluessel)`.
	if _, err := e.db.ExecContext(ctx,
		`INSERT INTO settings (schluessel, wert, profil_id) VALUES (?, ?, ?)
		 ON CONFLICT (profil_id, schluessel) DO UPDATE SET wert = excluded.wert`,
		schluessel, wert, profilID); err != nil {
		return uebersetze(err, "die Einstellung %q setzen", schluessel)
	}

	return nil
}
