package app

// **CLI und MCP per Knopf einrichten und widerrufen** — OPS-09, Schnitt 1
// (#1907).
//
// # Welche Ziele es gibt
//
//   - **befehlszeile**: der `PATH`-Eintrag in `~/.zprofile` — die
//     Anmeldeshell unter macOS ist zsh. Nur, wenn `~/.local/bin` im `PATH`
//     der Anmeldeshell fehlt; die Verknüpfungen selbst legt weiterhin
//     [App.Befehlszeile] an (OPS-03).
//   - **codex**: der Eintrag für `mse-mcp` in `~/.codex/config.toml`,
//     nachgelesen am 2026-10-01 an `codex mcp add --help` (CLI 0.159.2):
//     `[mcp_servers.<name>]` mit `command`. Nur, wenn Codex hier ist.
//
// **Claude Desktop, Claude Code und der Dienst als LaunchAgent** sind
// Schnitt 2 (#1961): JSON-Dateien, die ihre Klienten selbst umschreiben, und
// ein Dienst, der beim Anmelden startet, brauchen jeder eine eigene Antwort
// auf „byte-gleich".
//
// # Wer schreibt
//
// **Nur das Fenster** — dieselbe Entscheidung wie bei OPS-03: ein fremder
// Rufer über REST oder MCP schriebe in ein fremdes Zuhause. Lesend steht der
// Stand an allen vier Flächen (§16).

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// SchnittstelleDTO ist ein Ziel samt Stand und Vorschau.
type SchnittstelleDTO struct {
	Name  string `json:"name"`
	Stand string `json:"stand"`
	Datei string `json:"datei"`

	// Art ist `""` (ein Block), `json`, `befehl` oder `agent` (OPS-09b).
	Art string `json:"art"`

	// Zeilen sind die, die einrichten hinzufügt — samt ihren Marken
	// (AK-01: jede Datei und jede Zeile vorher).
	Zeilen []string `json:"zeilen"`
}

// SchnittstellenAntwort ist der Stand aller Ziele.
type SchnittstellenAntwort struct {
	Schnittstellen []SchnittstelleDTO `json:"schnittstellen"`

	// NurEintrag nennt das Ziel, dessen Datei nach dem Widerrufen nicht
	// wieder ist wie vorher — nur der eigene Eintrag ging (OPS-09b).
	NurEintrag string `json:"nurEintrag"`

	Fehler *Fehler `json:"fehler"`
}

// schnittstellenfelder sind die Nähte — für Läufe mit eigenem Zuhause.
type schnittstellenfelder struct {
	heim       func() (string, error)
	werkzeug   func() (string, error)
	pfad       func(context.Context) string
	ausfuehren befehlsausfuehrung
}

// echteSchnittstellenorte sind die des Betriebs — gesetzt in [NeueApp].
func echteSchnittstellenorte() schnittstellenfelder {
	return schnittstellenfelder{
		heim: os.UserHomeDir, werkzeug: Werkzeugort, pfad: anmeldepfad, ausfuehren: befehlAusfuehren,
	}
}

// anmeldepfad ist der PATH der Anmeldeshell — der des Fensters ist der von
// `launchd` (Befund 9 des Zweitblicks über Phase 6.4).
func anmeldepfad(ctx context.Context) string {
	return shellpfad(ctx, (*exec.Cmd).Output, os.Getenv("PATH"))
}

// MitSchnittstellenorten setzt Zuhause, Werkzeugort und den PATH der
// Anmeldeshell — für Läufe, die nie das echte Zuhause anfassen dürfen.
func (a *App) MitSchnittstellenorten(
	heim, werkzeug func() (string, error), pfad func(context.Context) string,
) *App {
	a.schnittstellen = schnittstellenfelder{
		heim: heim, werkzeug: werkzeug, pfad: pfad, ausfuehren: a.schnittstellen.ausfuehren,
	}

	return a
}

// MitSchnittstellenbefehlen setzt, wie `claude` und `launchctl` gerufen
// werden — ein Lauf ruft keines von beiden wirklich (OPS-09b, #1961).
func (a *App) MitSchnittstellenbefehlen(ausfuehren func(ctx context.Context, befehl []string) error) *App {
	a.schnittstellen.ausfuehren = ausfuehren

	return a
}

// Schnittstellen liest den Stand — **rein lesend**.
func (a *App) Schnittstellen(ctx context.Context) SchnittstellenAntwort {
	ctx = a.imProfil(ctx)

	ziele, fehler := a.schnittstellenziele(ctx)
	if fehler != nil {
		return SchnittstellenAntwort{Schnittstellen: []SchnittstelleDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	return SchnittstellenAntwort{Schnittstellen: alsSchnittstellen(ziele)}
}

// SchnittstelleEinrichten richtet ein Ziel ein.
func (a *App) SchnittstelleEinrichten(ctx context.Context, name string) SchnittstellenAntwort {
	ctx = a.imProfil(ctx)

	return a.anSchnittstelle(ctx, name, service.SchnittstelleEinrichten)
}

// SchnittstelleWiderrufen nimmt den eigenen Eintrag wieder heraus.
func (a *App) SchnittstelleWiderrufen(ctx context.Context, name string) SchnittstellenAntwort {
	ctx = a.imProfil(ctx)

	return a.anSchnittstelle(ctx, name, service.SchnittstelleWiderrufen)
}

func (a *App) anSchnittstelle(
	ctx context.Context, name string,
	handlung func(context.Context, service.Schnittstellenziel) (service.Schnittstellenausgang, error),
) SchnittstellenAntwort {
	bytegleich := true

	ziele, fehler := a.schnittstellenziele(ctx)
	if fehler == nil {
		fehler = fmt.Errorf("%w: %q ist keine Schnittstelle dieser Anwendung", domain.ErrValidation, name)

		for _, ziel := range ziele {
			if ziel.Name == name {
				var ausgang service.Schnittstellenausgang

				ausgang, fehler = handlung(ctx, ziel)
				bytegleich = ausgang.Bytegleich || ausgang.Stand != service.SchnittstelleOffen

				break
			}
		}
	}

	antwort := a.Schnittstellen(ctx)
	if fehler != nil {
		antwort.Fehler = Uebersetze(ctx, fehler)
	}

	// **Nur den eigenen Eintrag genommen** — die Datei hatte sich seither
	// geändert, und ihr Stand davor kam nicht zurück (OPS-09b, #1961).
	if fehler == nil && !bytegleich {
		antwort.NurEintrag = name
	}

	return antwort
}

// schnittstellenziele baut die Ziele aus Zuhause, Werkzeugort und PATH.
func (a *App) schnittstellenziele(ctx context.Context) ([]service.Schnittstellenziel, error) {
	orte := a.schnittstellen

	heim, fehler := orte.heim()
	if fehler != nil {
		return nil, fmt.Errorf("das Zuhause finden: %w", fehler)
	}

	werkzeug, fehler := orte.werkzeug()
	if fehler != nil {
		return nil, fmt.Errorf("den Ort der Werkzeuge finden: %w", fehler)
	}

	roh := orte.pfad(ctx)
	pfad := filepath.SplitList(roh)
	lokal := filepath.Join(heim, ".local", "bin")

	return append([]service.Schnittstellenziel{
		{
			Name: "befehlszeile", Datei: filepath.Join(heim, ".zprofile"), Vorhanden: true,
			Block:    "export PATH=\"$HOME/.local/bin:$PATH\"\n",
			Unnoetig: enthaeltOrt(pfad, lokal),
		},
		befehleZiel(heim, werkzeug, roh),
		{
			Name: "codex", Datei: filepath.Join(heim, ".codex", "config.toml"),
			Vorhanden: ordnerDa(filepath.Join(heim, ".codex")) || programmImPfad(pfad, "codex"),
			Block: fmt.Sprintf("[mcp_servers.magicseteditor4mac]\ncommand = %q\n",
				filepath.Join(werkzeug, "mse-mcp")),
			Anspruch: "[mcp_servers.magicseteditor4mac]",
		},
	}, zweiteZiele(heim, werkzeug, pfad, orte.ausfuehren)...), nil
}

func alsSchnittstellen(ziele []service.Schnittstellenziel) []SchnittstelleDTO {
	dtos := make([]SchnittstelleDTO, 0, len(ziele))

	for _, ziel := range ziele {
		dtos = append(dtos, SchnittstelleDTO{
			Name: ziel.Name, Stand: string(service.SchnittstelleLesen(ziel)), Datei: ziel.Datei,
			Art: string(ziel.Art), Zeilen: vorschauzeilen(ziel),
		})
	}

	return dtos
}

func enthaeltOrt(pfad []string, ort string) bool {
	for _, eintrag := range pfad {
		if filepath.Clean(eintrag) == filepath.Clean(ort) {
			return true
		}
	}

	return false
}

func ordnerDa(ort string) bool {
	angaben, fehler := os.Stat(ort)

	return fehler == nil && angaben.IsDir()
}

func programmImPfad(pfad []string, name string) bool { return programmIn(pfad, name) != "" }
