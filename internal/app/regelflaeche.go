package app

// **Die Regelfreigabe ohne Fenster** — Werkstück G4 von Phase 5.5
// (#1461).
//
// # Warum es diese Fläche gibt
//
// **§16**: was das Fenster kann, können Kommandozeile und MCP — oder es steht
// begründet, warum nicht. Hier gibt es keinen Grund: die Freigabe rechnet über
// ein Regelprofil und eine Strenge, und beides hat mit einem Fenster nichts zu
// tun.
//
// **Und das Issue verlangt es ausdrücklich**: „Turniersperre lässt sich weder
// über UI noch REST/CLI/MCP umgehen." Eine Sperre, die nur das Fenster kennt,
// ist keine.

import "database/sql"

// RegelflaecheUeber baut eine Anwendung, die nach der Regelfreigabe gefragt
// werden kann.
//
// **Mit Bestand, obwohl die Rechnung keinen braucht**: der Riegel vor jedem
// Weg fragt die Rolle, und die steht im Bestand. Ohne ihn käme jeder durch.
func RegelflaecheUeber(db *sql.DB) *App {
	anwendung := NeueApp("")
	anwendung.db = db

	return anwendung
}
