// **Der Zug des Stolpertreibers** — Rollback und Commit.
//
// **Eine eigene Datei**, weil `stolperabfragen_test.go` mit dem
// Handelsabschluss über die Zeilengrenze ging (§2). Der Schnitt folgt der
// Frage: dort, **was eine Abfrage antwortet** — hier, **was aus einer
// Transaktion wird**.
package storage

type stolpertx stolperconn

// Rollback stolpert nur an der Stelle, die dafür da ist.
//
// **Sonst verdeckte er den Fehler, der ihn ausgelöst hat**: `verwerfen` hängt
// seinen Fehler an den eigentlichen an, und ein Treiber, der immer stolpert,
// machte aus jedem Fehlerweg zwei.
func (t stolpertx) Rollback() error {
	aufzeichnen(t.stelle, "ROLLBACK")

	if t.stelle == beimZurueckrollen {
		return errGestolpert
	}

	return nil
}

func (t stolpertx) Commit() error {
	aufzeichnen(t.stelle, "COMMIT")

	// **`beimNutzercommit` stand seit dem 2026-09-04 nicht hier** — angelegt
	// in Werkstück 1 von Phase 4.12 (05ac67d2), registriert als
	// `stolper-nutzer-commit`, und nie abgefragt. Wer ihn genommen hätte,
	// hätte einen Treiber bekommen, der **nicht stolpert**: der Lauf wäre
	// glatt durchgegangen, und der Test hätte den Fehlerweg bestätigt, ohne
	// ihn zu betreten.
	//
	// Gefunden hat es der erste Test, der ihn wirklich benutzt — er wurde rot
	// (`AGENTS.md` §10: ein Wächter belegt nichts, bevor eine gezielte
	// Verletzung ihn rot gemacht hat).
	if t.stelle == beimCommit || t.stelle == beimFeldcommit ||
		t.stelle == beimVerlaufcommit || t.stelle == beimSetcommit ||
		t.stelle == beimKartecommit || t.stelle == beimWortcommit ||
		t.stelle == beimSetloeschcommit || t.stelle == beimRundencommit ||
		t.stelle == beimNutzercommit || t.stelle == beimErwartungcommit ||
		t.stelle == beimDruckcommit || t.stelle == beimOffenlegungCommit ||
		t.stelle == beimBindungCommit || t.stelle == beimAbschlusscommit {
		return errGestolpert
	}

	return nil
}
