package storage

import "context"

// Setstand gibt den Änderungszähler eines Sets — PERF02
// (#2352).
// Ihn heben die Auslöser aus `0099_setstand.sql` bei jedem Weg, der eine
// gezeichnete Karte verändern kann; eine Problemprüfung vergleicht ihn mit
// dem, mit dem sie begann.
func (l *SetLeser) Setstand(ctx context.Context, setID int64) (int64, error) {
	// **Nur aus dem gezeigten Profil** (#1815), wie jeder Leseweg eines Sets.
	if fehler := setImProfil(ctx, l.db, setID); fehler != nil {
		return 0, fehler
	}

	var stand int64
	if err := l.db.QueryRowContext(ctx, `SELECT zaehler FROM set_aenderungen WHERE set_id = ?`, setID).Scan(&stand); err != nil {
		return 0, uebersetze(err, "den Stand des Sets lesen")
	}

	return stand, nil
}
