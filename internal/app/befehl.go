package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"sync"
)

// **Die Kommandozentrale ruft die Werkzeuge des MCP-Servers** — Werkstück U12
// von Phase 5.1 (#1423).
//
// # Kein zweiter Katalog
//
// Was ein Sprachmodell über MCP kann, ist die Liste der Fachbefehle dieser
// Anwendung — mit Namen, Art, Schema und Erklärung, gehalten von Wächtern
// gegen `docs/FAEHIGKEITEN.md`. Die Kommandozentrale zeigt **diese** Liste und
// ruft **diese** Werkzeuge, im eigenen Prozess: dieselbe Fachantwort wie über
// MCP und die Kommandozeile, keine Shell, kein Schattenbefehl.
//
// # Warum die Quelle von außen kommt
//
// `internal/mcpnaehte` füllt die Nähte und kennt dafür `app`; `app` kann es
// deshalb nicht importieren. Der Wurzelknoten des Fensters (`main.go`) reicht
// die Quelle herein — wie `cmd/mse-mcp` dem Server seine Nähte gibt.

// Befehl ist ein Werkzeug des MCP-Servers, wie das Fenster es ruft.
type Befehl struct {
	Name       string
	Schreibend bool
	Idempotent bool
	Erklaerung string
	Schema     map[string]any
	Ruf        func(ctx context.Context, argumente json.RawMessage) (string, error)
}

// Befehlsquelle füllt die Befehle für einen Bestand.
type Befehlsquelle func(db *sql.DB, pfad string) ([]Befehl, error)

// befehlsnaehte ist, was die Kommandozentrale am App hält — eingebettet über
// die Nähte der Übersicht, weil `anwendung.go` an der Zeilengrenze steht.
type befehlsnaehte struct {
	befehlsquelle Befehlsquelle
	befehle       []Befehl

	// **Ein Lauf zur Zeit** (§7, Regel 4): der Riegel hält den Abbruch des
	// laufenden Befehls — oder nil, wenn keiner läuft.
	befehlsriegel   sync.Mutex
	befehlAbbrechen context.CancelFunc
}

// MitBefehlsquelle setzt die Quelle — vor dem Start; die Befehle entstehen
// mit dem Bestand.
func (a *App) MitBefehlsquelle(quelle Befehlsquelle) *App {
	a.befehlsquelle = quelle

	return a
}

// MitBefehlen setzt die Befehle unmittelbar — für Tests und für eine Quelle,
// die schon gelaufen ist.
func (a *App) MitBefehlen(befehle []Befehl) *App {
	a.befehle = befehle

	return a
}

// befehleAus füllt die Befehle aus der Quelle — ein Fehler dabei hält den
// Start nicht auf: die Kommandozentrale sagt dann, dass sie leer ist.
func (a *App) befehleAus(db *sql.DB) {
	if a.befehlsquelle == nil {
		return
	}

	befehle, err := a.befehlsquelle(db, a.datenbankpfad)
	if err != nil {
		slog.Error("die Befehle der Kommandozentrale stehen nicht bereit", "grund", err.Error())

		return
	}

	a.befehle = befehle
}
