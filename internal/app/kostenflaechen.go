package app

// **Das Kostenjournal für die Flächen ohne Fenster** — Werkstück O12 von
// Phase 5.2 (#1439).
//
// **Über die Nutzerfläche** — dieselbe Überlegung wie bei
// `ruecksetzungsflaechen.go`: der Wächter fragt den Akteur, und den baut
// `nutzerflaeche` samt Freigaben und Nachweisen.

import (
	"database/sql"
	"time"
)

// KostenAnwendungUeber baut eine Anwendung, die das Journal führt.
func KostenAnwendungUeber(db *sql.DB) *App {
	anwendung := NutzerflaecheUeber(db).app

	anwendung.MitKosten(KostenwacheUeber(
		func() *sql.DB { return db }, anwendung.autorisierung, time.Now))

	return anwendung
}
