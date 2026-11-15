package service

// **CLI und MCP per Knopf einrichten und widerrufen** — OPS-09, Schnitt 1
// (#1907),
// Marke `new`: MSE2 richtet nichts außerhalb seines Programms ein.
//
// # Was ein Ziel ist
//
// **Eine Datei, in die ein Block gehört** — der `PATH`-Eintrag in
// `~/.zprofile`, der Eintrag für `mse-mcp` in `~/.codex/config.toml`. Der
// Block steht zwischen zwei Marken ([domain.BlockEinsetzen]); was außerhalb
// davon steht, gehört dem Menschen und wird nie angefasst.
//
// # Die sechs Stände
//
// **nichtGefunden** (der Klient ist nicht auf diesem Rechner), **unnoetig**
// (es fehlt nichts — `~/.local/bin` steht schon im `PATH`), **offen**,
// **eingerichtet**, **veraltet** (der eigene Block nennt einen anderen Ort —
// einrichten repariert ihn) und **fremd** (derselbe Name steht außerhalb des
// Blocks, oder ein halber Block, oder die Datei lässt sich nicht lesen —
// dann wird nichts geschrieben).

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Schnittstellendateigrenze ist, wie groß eine Zieldatei sein darf, damit
// dieses Programm sie liest — ein Mebibyte. Darüber gilt sie als fremd.
const Schnittstellendateigrenze = 1 << 20

// Schnittstellenart sagt, wie ein Ziel geschrieben wird — OPS-09b
// (#1961).
type Schnittstellenart string

// Die fünf Arten — die fünfte, [ArtVerknuepfung], steht in
// `schnittstellenverknuepfung.go`.
const (
	// ArtBlock ist ein markierter Block in einer Textdatei.
	ArtBlock Schnittstellenart = ""

	// ArtJSON ist ein Eintrag unter `mcpServers` einer JSON-Datei, die
	// dieses Programm selbst schreibt (Claude Desktop).
	ArtJSON Schnittstellenart = "json"

	// ArtBefehl ist ein Eintrag, den der Klient mit seinem eigenen Befehl
	// schreibt (Claude Code) — gelesen wird er aus seiner Datei.
	ArtBefehl Schnittstellenart = "befehl"

	// ArtAgent ist ein LaunchAgent: eine plist, die `launchctl` lädt.
	ArtAgent Schnittstellenart = "agent"
)

// Schnittstellenziel ist ein Ort, an den ein Eintrag gehört.
type Schnittstellenziel struct {
	Name  string
	Art   Schnittstellenart
	Datei string
	Block string

	// Schluessel ist der Name unter `mcpServers` (JSON, Befehl).
	Schluessel string

	// Programm ist der Werkzeugserver, den der Eintrag nennt — daran
	// erkennt dieses Programm seinen eigenen (JSON, Befehl).
	Programm string

	// Gedaechtnis ist die Datei, in der der Stand vor dem Einrichten steht,
	// damit widerrufen byte-gleich zurückgeht (JSON).
	Gedaechtnis string

	// Einrichtbefehle und Widerrufbefehle sind die Befehle, die der Klient
	// selbst ausführt (Befehl, Agent) — je ein Aufruf mit Argumenten.
	Einrichtbefehle, Widerrufbefehle [][]string

	// Altlasten sind LaunchAgents dieser Anwendung unter einem früheren
	// Etikett (Agent) — Zweitblick über Phase 4, Befund 7.
	Altlasten []Altagent

	// Begleitordner entsteht vor den Befehlen mit 0700 — dort schreibt der
	// LaunchAgent seine Ansage hin, und niemand sonst soll sie lesen (Agent).
	Begleitordner string

	// Ansage ist die Datei, in die der Dienst je Start seine Marke schreibt
	// (Agent, SEC02 #2355) — leer heißt: keine.
	Ansage string

	// Ausfuehren führt einen Befehl aus — eingesetzt, damit ein Lauf ihn
	// nachstellen kann, statt `claude` oder `launchctl` zu rufen.
	Ausfuehren func(ctx context.Context, befehl []string) error

	// Vorhanden sagt, ob der Klient auf diesem Rechner ist.
	Vorhanden bool

	// Unnoetig sagt, dass ohne den Block nichts fehlt.
	Unnoetig bool

	// Anspruch ist eine Zeile, die außerhalb des Blocks heißt: jemand
	// anders hat diesen Eintrag schon geschrieben — leer heißt: keine.
	Anspruch string

	// Verknuepfung trägt Stand und Schritte eines Ziels der
	// [ArtVerknuepfung] (OPS03, #2349).
	Verknuepfung *Verknuepfungsziel

	// zwischendurch läuft unmittelbar vor dem Vergleich, der jedem Ersetzen
	// vorausgeht — nur für Läufe, die einen anderen Schreiber nachstellen.
	zwischendurch func()
}

// Schnittstellenstand ist, wie es um ein Ziel steht.
type Schnittstellenstand string

// Die sechs Stände.
const (
	SchnittstelleNichtGefunden Schnittstellenstand = "nichtGefunden"
	SchnittstelleUnnoetig      Schnittstellenstand = "unnoetig"
	SchnittstelleOffen         Schnittstellenstand = "offen"
	SchnittstelleEingerichtet  Schnittstellenstand = "eingerichtet"
	SchnittstelleVeraltet      Schnittstellenstand = "veraltet"
	SchnittstelleFremd         Schnittstellenstand = "fremd"
)

// Schnittstellenausgang ist, was einrichten oder widerrufen hinterließ.
type Schnittstellenausgang struct {
	Stand Schnittstellenstand

	// Bytegleich sagt nach dem Widerrufen, ob die Datei wieder ist wie vor
	// dem Einrichten. **Nicht**, wenn der Mensch sie seither geändert hat:
	// dann geht nur der eigene Eintrag, und das wird gesagt.
	Bytegleich bool
}

// SchnittstelleLesen liest, wie es um ein Ziel steht.
func SchnittstelleLesen(ziel Schnittstellenziel) Schnittstellenstand {
	if ziel.Art == ArtVerknuepfung {
		return ziel.Verknuepfung.Stand()
	}

	inhalt, fehler := inhaltVon(ziel)
	if fehler != nil {
		return SchnittstelleFremd
	}

	return standVon(ziel, inhalt)
}

// standVon ist der Stand eines gelesenen Inhalts.
//
// **Getrennt vom Lesen** (Zweitblick über Phase 6.12, Z4): einrichten und
// widerrufen urteilen über **denselben** Inhalt, den sie danach ändern —
// vorher lasen sie zweimal, und das Urteil konnte einem anderen Inhalt gelten
// als die Änderung.
func standVon(ziel Schnittstellenziel, inhalt []byte) Schnittstellenstand {
	switch ziel.Art {
	case ArtJSON, ArtBefehl:
		return eintragsstandAus(ziel, inhalt)
	case ArtAgent:
		return agentenstandAus(ziel, inhalt)
	}

	return standAus(ziel, inhalt)
}

func standAus(ziel Schnittstellenziel, inhalt []byte) Schnittstellenstand {
	block := domain.BlockStand(inhalt, ziel.Block)

	switch {
	case block == domain.BlockFremd:
		return SchnittstelleFremd
	case block == domain.BlockEingetragen:
		return SchnittstelleEingerichtet
	case block == domain.BlockVeraltet:
		return SchnittstelleVeraltet
	case ziel.Anspruch != "" && bytes.Contains(inhalt, []byte(ziel.Anspruch)):
		return SchnittstelleFremd
	case !ziel.Vorhanden:
		return SchnittstelleNichtGefunden
	case ziel.Unnoetig:
		return SchnittstelleUnnoetig
	}

	return SchnittstelleOffen
}

// SchnittstelleEinrichten setzt den Eintrag — oder repariert einen
// veralteten. **Idempotent**: wer zweimal einrichtet, ändert beim zweiten
// Mal nichts.
func SchnittstelleEinrichten(ctx context.Context, ziel Schnittstellenziel) (Schnittstellenausgang, error) {
	if ziel.Art == ArtVerknuepfung {
		return verknuepfungEinrichten(ziel)
	}

	inhalt, fehler := inhaltVon(ziel)
	if fehler != nil {
		return Schnittstellenausgang{Stand: SchnittstelleFremd}, fehler
	}

	stand := standVon(ziel, inhalt)

	switch stand {
	case SchnittstelleEingerichtet:
		return Schnittstellenausgang{Stand: stand}, nil
	case SchnittstelleFremd:
		return Schnittstellenausgang{Stand: stand}, fmt.Errorf(
			"%w: dort steht schon ein fremder Eintrag — er bleibt, wie er ist", domain.ErrConflict)
	case SchnittstelleNichtGefunden, SchnittstelleUnnoetig:
		return Schnittstellenausgang{Stand: stand}, fmt.Errorf("%w: %s braucht hier keinen Eintrag (%s)",
			domain.ErrValidation, ziel.Name, stand)
	}

	switch ziel.Art {
	case ArtJSON:
		fehler = jsonEinrichten(ziel, inhalt)
	case ArtBefehl, ArtAgent:
		fehler = befehleEinrichten(ctx, ziel, stand)
	default:
		fehler = zielErsetzen(ziel, inhalt, domain.BlockEinsetzen(inhalt, ziel.Block))
	}

	if fehler != nil {
		return Schnittstellenausgang{Stand: SchnittstelleLesen(ziel)}, fehler
	}

	return Schnittstellenausgang{Stand: SchnittstelleLesen(ziel)}, nil
}

// SchnittstelleWiderrufen nimmt den eigenen Eintrag — byte-gleich zum Stand
// davor, wo es geht. **Eine Datei, die dabei leer wird, geht** — sie
// entstand mit dem Einrichten.
func SchnittstelleWiderrufen(ctx context.Context, ziel Schnittstellenziel) (Schnittstellenausgang, error) {
	if ziel.Art == ArtVerknuepfung {
		return verknuepfungWiderrufen(ziel)
	}

	inhalt, fehler := inhaltVon(ziel)
	if fehler != nil {
		return Schnittstellenausgang{Stand: SchnittstelleFremd}, fehler
	}

	stand := standVon(ziel, inhalt)

	switch stand {
	case SchnittstelleFremd:
		return Schnittstellenausgang{Stand: stand}, fmt.Errorf(
			"%w: dort steht kein eigener Eintrag, nur ein fremder — er bleibt", domain.ErrConflict)
	case SchnittstelleEingerichtet, SchnittstelleVeraltet:
	default:
		return Schnittstellenausgang{Stand: stand, Bytegleich: true}, nil
	}

	bytegleich := true

	switch ziel.Art {
	case ArtJSON:
		bytegleich, fehler = jsonWiderrufen(ziel, inhalt)
	case ArtBefehl, ArtAgent:
		fehler = befehleWiderrufen(ctx, ziel)
	default:
		fehler = blockWiderrufen(ziel, inhalt)
	}

	return Schnittstellenausgang{Stand: SchnittstelleLesen(ziel), Bytegleich: bytegleich && fehler == nil}, fehler
}

// blockWiderrufen nimmt den markierten Block heraus.
func blockWiderrufen(ziel Schnittstellenziel, inhalt []byte) error {
	uebrig := domain.BlockEntfernen(inhalt)
	if len(uebrig) == 0 {
		return zielEntfernen(ziel, inhalt)
	}

	return zielErsetzen(ziel, inhalt, uebrig)
}

// zielErsetzen schreibt `neu` — aber nur, wenn in der Datei noch steht, was
// gelesen wurde.
//
// **Eine Änderung zwischen Lesen und Schreiben gewinnt** (Zweitblick über
// Phase 6.12, Z4): die Datei gehört dem Menschen, und wer sie eben
// geändert hat, verlöre seine Änderung an einen Stand von vorhin. **Was das
// nicht hält**: einen Schreiber, der genau zwischen diesem Vergleich und dem
// Umbenennen schreibt — zwei Programme, die dieselbe Datei ohne Absprache
// schreiben, kann keines von beiden ganz trennen.
func zielErsetzen(ziel Schnittstellenziel, erwartet, neu []byte) error {
	if fehler := unveraendert(ziel, erwartet); fehler != nil {
		return fehler
	}

	return zielSchreiben(ziel.Datei, neu)
}

// zielEntfernen nimmt die Datei weg — unter derselben Bedingung.
func zielEntfernen(ziel Schnittstellenziel, erwartet []byte) error {
	if fehler := unveraendert(ziel, erwartet); fehler != nil {
		return fehler
	}

	if fehler := os.Remove(echterOrt(ziel.Datei)); fehler != nil {
		return fmt.Errorf("%s entfernen: %w", ziel.Datei, fehler)
	}

	return nil
}

// unveraendert sagt, ob in der Datei noch steht, was gelesen wurde. **Eine
// Datei, die sich nicht mehr lesen lässt, hat sich auch geändert.**
func unveraendert(ziel Schnittstellenziel, erwartet []byte) error {
	if ziel.zwischendurch != nil {
		ziel.zwischendurch()
	}

	if jetzt, fehler := inhaltVon(ziel); fehler != nil || !bytes.Equal(jetzt, erwartet) {
		return fmt.Errorf("%w: %s hat sich eben geändert — versuche es noch einmal",
			domain.ErrConflict, ziel.Datei)
	}

	return nil
}

// Klientdateigrenze ist, wie groß die Datei eines Klienten sein darf, die
// dieses Programm nur **liest** — `~/.claude.json` führt Verläufe je Projekt
// und wächst über ein Mebibyte hinaus. Geschrieben wird sie hier nie.
const Klientdateigrenze = 32 << 20

// inhaltVon liest die Datei eines Ziels mit der Grenze seiner Art.
func inhaltVon(ziel Schnittstellenziel) ([]byte, error) {
	if ziel.Art == ArtBefehl {
		return zielLesenBis(ziel.Datei, Klientdateigrenze)
	}

	return zielLesen(ziel.Datei)
}

// zielLesen liest eine Zieldatei — eine fehlende ist leer, eine über
// [Schnittstellendateigrenze] wird nicht gelesen.
func zielLesen(datei string) ([]byte, error) {
	return zielLesenBis(datei, Schnittstellendateigrenze)
}

// zielLesenBis liest eine Zieldatei bis zu einer Grenze.
func zielLesenBis(datei string, grenze int64) ([]byte, error) {
	quelle, fehler := os.Open(echterOrt(datei)) // #nosec G304 -- eine Datei im Zuhause des Menschen, von der App benannt
	if errors.Is(fehler, fs.ErrNotExist) {
		return nil, nil
	}

	if fehler != nil {
		return nil, fmt.Errorf("%s lesen: %w", datei, fehler)
	}

	defer func() { _ = quelle.Close() }() //nolint:errcheck // nur gelesen

	inhalt, fehler := io.ReadAll(io.LimitReader(quelle, grenze+1))
	if fehler != nil {
		return nil, fmt.Errorf("%s lesen: %w", datei, fehler)
	}

	if int64(len(inhalt)) > grenze {
		return nil, fmt.Errorf("%w: die Datei ist größer als %d Byte — sie bleibt, wie sie ist",
			domain.ErrConflict, grenze)
	}

	return inhalt, nil
}

// zielSchreiben schreibt atomar — und dorthin, wohin ein Symlink zeigt:
// wer sein Shellprofil aus einem Dotfiles-Verzeichnis verknüpft, behält die
// Verknüpfung.
func zielSchreiben(datei string, inhalt []byte) error {
	ort := echterOrt(datei)

	if fehler := os.MkdirAll(filepath.Dir(ort), 0o700); fehler != nil {
		return fmt.Errorf("das Verzeichnis zu %s anlegen: %w", datei, fehler)
	}

	if fehler := schreibeAtomar(ort, inhalt); fehler != nil {
		return fmt.Errorf("%s schreiben: %w", datei, fehler)
	}

	return nil
}

// echterOrt löst einen Symlink auf — oder gibt den Pfad, wie er ist.
func echterOrt(datei string) string {
	if ort, fehler := filepath.EvalSymlinks(datei); fehler == nil {
		return ort
	}

	return datei
}
