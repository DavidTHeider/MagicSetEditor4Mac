package app

// **Fertige Konfigurationsbruchstücke für Agenten** — Werkstück A14 von
// Phase 5.3 (#1453).
//
// # Wofür
//
// **Weil `mse-mcp` in keinem Suchpfad liegt.** Es steckt im Programmbündel,
// und eine Anleitung, die `mse-mcp` schreibt, führt bei niemandem zum Ziel.
// Was hier entsteht, ist der **volle Pfad dieses Rechners** in der Gestalt,
// die der jeweilige Agent liest.
//
// # Und was hier ausdrücklich nicht geschieht
//
// **Es wird nichts geschrieben und nichts gestartet.** Das Issue verlangt es
// wörtlich: „Bestehende Konfiguration nicht automatisch
// überschreiben/Agent starten." Eine Anwendung, die in `~/.claude.json`
// schriebe, überschriebe die Arbeit eines Menschen an einer Datei, die ihm
// gehört — und ein Fehlgriff dort ist nicht rückgängig zu machen.
//
// # Und kein Geheimnis
//
// **`mse-mcp` liest seine Zugangsdaten selbst** — aus dem Schlüsselbund von
// macOS (0063). Ein `env`-Block mit einem Wert darin wäre der eine Weg, diese
// Entscheidung zu unterlaufen: ein Schlüssel in einer Konfigurationsdatei
// liegt dort im Klartext und wandert mit ihr in jedes Repositorium. Ein
// Wächter hält das.
//
// # Warum die Gestalten nachgelesen sind
//
// **Weil eine Konvention aus dem Gedächtnis kein Prüfergebnis ist** (§1, §10).
// Beide sind am 2026-09-13 an ihrer Primärquelle abgerufen, und das Datum
// steht am Bruchstück — wer sie morgen nachsieht, soll wissen, wo und wann.

import (
	"path/filepath"
	"strings"
)

// AgentenbruchstueckDTO ist ein fertiges Stück Konfiguration.
type AgentenbruchstueckDTO struct {
	// Agent ist `claude` oder `codex` — die Marke, unter der das Fenster
	// seinen Namen nachschlägt (§9).
	Agent string `json:"agent"`

	// Datei ist der Ort, an den es gehört.
	Datei string `json:"datei"`

	// Form ist `json` oder `toml` — für die Hervorhebung im Fenster.
	Form string `json:"form"`

	// Inhalt ist das Bruchstück selbst.
	Inhalt string `json:"inhalt"`

	// Befehl ist der Weg über die Kommandozeile des Agenten — dieselbe
	// Wirkung, ohne eine Datei von Hand zu öffnen.
	Befehl string `json:"befehl"`

	// Nachgelesen nennt Quelle und Abrufdatum.
	Nachgelesen string `json:"nachgelesen"`
}

// Agentenname ist der Name, unter dem dieses Programm bei einem Agenten steht.
//
// **Klein und ohne Punkt**: beide Konventionen nehmen ihn als Schlüssel einer
// Tabelle, und ein Punkt darin wäre in TOML eine zweite Ebene.
const Agentenname = "magicseteditor4mac"

// Agentenbruchstuecke baut sie für den Ort, an dem die Werkzeuge liegen.
//
// **Öffentlich, damit ein Lauf sie ohne Bündel prüfen kann** — dieselbe
// Gestalt wie `technikangabenNeben` daneben.
func Agentenbruchstuecke(ort string) []AgentenbruchstueckDTO {
	server := filepath.Join(ort, "mse-mcp")

	return []AgentenbruchstueckDTO{
		claudebruchstueck(server),
		codexbruchstueck(server),
	}
}

// claudebruchstueck gibt die Gestalt, die Claude Code liest.
//
// **Nachgelesen am 2026-09-13** unter `https://code.claude.com/docs/en/mcp`:
// die Datei heißt `~/.claude.json`, die Wurzel `mcpServers`, und **`type`
// fehlt absichtlich** — für einen Stdio-Server ist es nicht nötig und wird
// dort ausdrücklich weggelassen.
func claudebruchstueck(server string) AgentenbruchstueckDTO {
	return AgentenbruchstueckDTO{
		Agent: "claude", Datei: "~/.claude.json", Form: "json",
		Inhalt: `{
  "mcpServers": {
    "` + Agentenname + `": {
      "command": "` + fuerJSON(server) + `",
      "args": []
    }
  }
}`,
		Befehl: "claude mcp add --transport stdio " + Agentenname +
			" -- " + zitiert(server),
		Nachgelesen: "code.claude.com/docs/en/mcp, abgerufen am 2026-09-13",
	}
}

// codexbruchstueck gibt die Gestalt, die Codex liest.
//
// **Nachgelesen am 2026-09-13** unter
// `https://learn.chatgpt.com/docs/extend/mcp?surface=cli`: die Datei heißt
// `~/.codex/config.toml`, die Tabelle `[mcp_servers.<name>]`.
func codexbruchstueck(server string) AgentenbruchstueckDTO {
	return AgentenbruchstueckDTO{
		Agent: "codex", Datei: "~/.codex/config.toml", Form: "toml",
		Inhalt: "[mcp_servers." + Agentenname + "]\n" +
			"command = " + fuerTOML(server) + "\n" +
			"args = []",
		Befehl:      "codex mcp add " + Agentenname + " -- " + zitiert(server),
		Nachgelesen: "learn.chatgpt.com/docs/extend/mcp, abgerufen am 2026-09-13",
	}
}

// fuerJSON macht aus einem Pfad einen JSON-Zeichenkettenrumpf.
//
// **Von Hand und nicht über `encoding/json`**: das Bruchstück ist
// **eingerückter** Text zum Ansehen, und ein Marshal über die ganze Gestalt
// gäbe entweder eine Zeile oder eine Einrückung, die niemand wählt. Zu
// entkommen sind dabei genau zwei Zeichen, und ein Dateipfad unter macOS
// trägt keines von beiden — die Zeile steht für den Fall, dass doch.
func fuerJSON(pfad string) string {
	ersetzt := strings.ReplaceAll(pfad, `\`, `\\`)

	return strings.ReplaceAll(ersetzt, `"`, `\"`)
}

// fuerTOML macht daraus eine TOML-Zeichenkette.
//
// **Einfach in Anführungszeichen** (basic string): dieselben zwei Zeichen
// sind zu entkommen wie in JSON, und TOML schreibt sie gleich.
func fuerTOML(pfad string) string { return `"` + fuerJSON(pfad) + `"` }
