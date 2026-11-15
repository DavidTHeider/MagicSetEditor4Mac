package storage

import (
	"context"
	"database/sql"
)

// Exportoptionen ist die Ablage der Optionswerte je Exportvorlage (#492).
//
// **Je Vorlage, nicht je Set** — so hält MSE sie
// (`exportOptionsFor`, `src/data/settings.hpp`, Zeilen 180–186, Fassung
// f9d9356 vom 2021-07-06). Ein Feld `creator` in zwei Vorlagen sind zwei
// Werte.
type Exportoptionen struct {
	db *sql.DB
}

// NeueExportoptionen gibt die Ablage über dieser Datenbank.
func NeueExportoptionen(db *sql.DB) *Exportoptionen {
	return &Exportoptionen{db: db}
}

// Fuer gibt die gesetzten Optionen einer Vorlage.
//
// **Leer und nicht nil** — derselbe Grund wie bei den Einstellungen: ein `nil`
// durchquert die Naht als `null`, und die Oberfläche prüfte dann darauf.
func (e *Exportoptionen) Fuer(ctx context.Context, vorlage string) (map[string]string, error) {
	// **Jede Wurzelabfrage klemmt auf ein Profil** (#1351, Vermerk 0066).
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, err := e.db.QueryContext(ctx,
		`SELECT feld, wert FROM export_options
		  WHERE vorlage = ? AND profil_id = ?`, vorlage, profilID)
	if err != nil {
		return nil, uebersetze(err, "die Optionen von %q lesen", vorlage)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // der Fehler kommt über zeilen.Err

	gefunden := make(map[string]string)

	for zeilen.Next() {
		var feld, wert string

		if err := zeilen.Scan(&feld, &wert); err != nil {
			return nil, uebersetze(err, "eine Option lesen")
		}

		gefunden[feld] = wert
	}

	if err := zeilen.Err(); err != nil {
		return nil, uebersetze(err, "die Optionen von %q lesen", vorlage)
	}

	return gefunden, nil
}

// Setzen schreibt eine Option fort.
func (e *Exportoptionen) Setzen(ctx context.Context, vorlage, feld, wert string) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	if _, err := e.db.ExecContext(ctx,
		`INSERT INTO export_options (vorlage, feld, wert, profil_id)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT (profil_id, vorlage, feld) DO UPDATE SET wert = excluded.wert`,
		vorlage, feld, wert, profilID); err != nil {
		return uebersetze(err, "die Option %q von %q setzen", feld, vorlage)
	}

	return nil
}
