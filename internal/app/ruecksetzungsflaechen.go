package app

// **Die Rücksetzung für die Flächen ohne Fenster** — Werkstück O4 von Phase
// 5.2 (#1428).
//
// # Keine dritte Wurzel, und zwar zweimal
//
// **Der Neuanfang hängt am `service.Bestandsweg`**, und den baut
// `BestandsnahtUeber` — dieselbe Überlegung wie bei `ruecksprungflaechen.go`.
//
// **Und das Löschen hängt an der `service.Nutzerverwaltung`**, die
// `nutzerflaeche` baut. Der erste Anlauf baute hier eine eigene, und der
// Nahtwächter hat es angestrichen: ihr fehlten `MitBeiwerk`, `MitFaktor` und
// `MitFreigaben` — ein über diese Fläche gelöschter Mensch hätte sein Bild,
// sein Abzeichen und sein TOTP-Geheimnis liegen lassen. **Derselbe Befund wie
// am 2026-09-07** (Zweitblick über die Phasen 4.0 bis 4.13, Befund 2), nur an
// einer neuen Fläche — und wieder gefunden vom Wächter und nicht von mir.

import (
	"database/sql"
)

// RuecksetzungsAnwendungUeber baut eine Anwendung, die zurücksetzen kann.
//
// **Ohne Auftragsschau**: ein Befehl führt kein Auftragsbuch — was er tut, tut
// er selbst, und es läuft nichts daneben, das er stoppen müsste.
func RuecksetzungsAnwendungUeber(db *sql.DB, bestandsort string) *App {
	anwendung := NutzerflaecheUeber(db).app
	naht := BestandsnahtUeber(&db, bestandsort)

	anwendung.MitRuecksetzung(RuecksetzungUeber(
		func() *sql.DB { return db },
		naht.Weg,
		anwendung.nutzer,
		anwendung.autorisierung,
	))

	return anwendung
}

// RuecksetzungsAnwendungOhneTausch baut sie für den Dienst — **ohne den
// Neuanfang** (G09, #2273).
//
// **Der Dienst hält eine Verbindung über seine ganze Laufzeit**, und jede
// seiner Nähte hat sie sich beim Aufbau geholt — deshalb spielt er auch nicht
// zurück (`internal/rest/bestandssicherung.go`). Ein Neuanfang ist ein Tausch:
// bis G09 schloss er diese Verbindung, und jede Anfrage danach traf sie
// geschlossen. Ohne den Weg nennt die Vorschau
// [domain.RuecksetzungOhneTausch], und die Ausführung hält an, bevor der
// Nachweis vorgezeigt ist.
func RuecksetzungsAnwendungOhneTausch(db *sql.DB) *App {
	anwendung := NutzerflaecheUeber(db).app

	anwendung.MitRuecksetzung(ruecksetzungMit(
		func() *sql.DB { return db }, anwendung.nutzer, anwendung.autorisierung, nil))

	return anwendung
}
