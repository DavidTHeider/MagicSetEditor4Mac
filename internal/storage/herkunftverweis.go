package storage

// **Verweist irgendein Set auf eine Quelle?** — die einzige Frage an `sets`,
// die mit Absicht über **alle** Profile geht (#1944). Sie steht in einer
// eigenen Datei, damit die Ausnahme im Wächter über die Wurzelabfragen eine
// Datei nennt und nicht die übrigen Abfragen mitnimmt.

import "context"

// HerkunftVerwiesen sagt, ob **irgendein** Set auf eine Quelldatei verweist —
// über alle Profile. Der Dienst räumt seinen Eingang damit auf (#1944), und
// eine Datei, die ein fremdes Profil noch braucht, darf nicht fort; deshalb
// fragt diese Stelle als einzige ohne Profil.
func (l *SetLeser) HerkunftVerwiesen(ctx context.Context, herkunft string) (bool, error) {
	var verwiesen bool

	if err := l.db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM sets WHERE herkunft = ? AND herkunft <> '')`, herkunft).
		Scan(&verwiesen); err != nil {
		return false, uebersetze(err, "nachsehen, ob ein Set auf %s verweist", herkunft)
	}

	return verwiesen, nil
}
