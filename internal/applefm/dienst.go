package applefm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Programm ist das Werkzeug von macOS 27.
const Programm = "/usr/bin/fm"

// Einrichtung sagt, ob das Gerätemodell auf diesem System aktiviert ist.
type Einrichtung string

const (
	EinrichtungBereit          Einrichtung = "bereit"
	EinrichtungFehlt           Einrichtung = "fehlt"
	EinrichtungLizenz          Einrichtung = "lizenz"
	EinrichtungNichtVerfuegbar Einrichtung = "nichtverfuegbar"
)

// Ausfuehrer führt `fm` mit diesen Argumenten aus und gibt seinen Exit-Code.
type Ausfuehrer func(ctx context.Context, argumente ...string) (int, error)

// Erkenne sagt, ob das Gerätemodell aktiviert ist — **drei Dinge zugleich**:
// das Werkzeug ist da, `fm license --status` und `fm available --model system`
// enden mit 0 (Vermerk 0166). **Gelesen wird der Exit-Code**, nicht der Text:
// die Ausgabe ist teils lokalisiert. **Gefragt wird nur der Stand** — die
// Lizenz nimmt ein Mensch an, für alle Nutzer der Maschine, mit `sudo`.
func Erkenne(ctx context.Context, vorhanden func(string) bool, ausfuehren Ausfuehrer) Einrichtung {
	if !vorhanden(Programm) {
		return EinrichtungFehlt
	}

	code, fehler := ausfuehren(ctx, "license", "--status")

	// **Ein Testlauf sieht kein Gerätemodell** — auf jedem Rechner gleich,
	// statt hier „Lizenz fehlt" und dort „fehlt" (§3).
	if errors.Is(fehler, ErrImLauf) {
		return EinrichtungFehlt
	}

	if fehler != nil || code != 0 {
		return EinrichtungLizenz
	}

	if code, fehler := ausfuehren(ctx, "available", "--model", Modell); fehler != nil || code != 0 {
		return EinrichtungNichtVerfuegbar
	}

	return EinrichtungBereit
}

// exitcodeVon führt einen fertigen Befehl aus — getrennt von [Exitcode],
// damit ein Lauf den Weg mit einem harmlosen Programm prüft.
func exitcodeVon(befehl *exec.Cmd) (int, error) {
	befehl.Env = append(os.Environ(), "NO_COLOR=1")

	fehler := befehl.Run()

	var beendet *exec.ExitError
	if errors.As(fehler, &beendet) {
		return beendet.ExitCode(), nil
	}

	if fehler != nil {
		return -1, fmt.Errorf("%s: %w", befehl.Path, fehler)
	}

	return 0, nil
}

// Vorhanden sagt, ob eine Datei da ist — der echte Prüfer für [Erkenne].
func Vorhanden(pfad string) bool {
	_, fehler := os.Stat(pfad)

	return fehler == nil
}

// ErrImLauf sagt, dass ein Testlauf `/usr/bin/fm` nicht startet (§3).
var ErrImLauf = errors.New("ein Testlauf startet das Gerätemodell nicht")

// gesperrt sagt, ob `fm` in diesem Lauf nicht starten darf: **in einem
// Testlauf** — erkannt am Namen des Programms, derselbe Riegel wie am
// Schlüsselbund (`internal/schluesselbund`).
//
// **Nur der Echtlauf öffnet ihn** — übersetzt mit `-tags fmlauf`
// ([echtlauf]); im gewöhnlichen Bau gibt es keinen Schalter dafür.
//
// **Und ein Kindprozess eines Laufs**: `mse`, `mse-mcp` und `mse-rest`
// heißen nicht `*.test`. Gemessen am 2026-10-04 im Flächenlauf: ohne diese
// Zeile hielten MCP und REST das echte Modell dieses Macs für bereit und
// gingen bis zur Frage. `internal/flaechen` setzt [Sperre] an der einen
// Stelle, an der jeder Kindprozess vorbeikommt — dieselbe Form wie beim
// Schlüsselbund.
func gesperrt() bool { return gesperrtFuer(os.Args[0], os.Getenv) }

// gesperrtFuer ist [gesperrt] über eingesetzte Angaben — prüfbar, ohne
// `os.Args` eines laufenden Tests zu verändern.
func gesperrtFuer(programm string, umgebung func(string) string) bool {
	return !echtlauf && (strings.HasSuffix(programm, ".test") || umgebung(Sperre) == "an")
}

// Sperre ist die Umgebungsvariable, die `fm` für einen Kindprozess eines
// Laufs sperrt.
const Sperre = "MSE_FM_GESPERRT"

// socketgrenze ist die Länge, die ein Unix-Socket unter macOS fasst
// (`sun_path`, 104 Bytes samt Endzeichen).
const socketgrenze = 104

// PruefeSocketpfad weist einen Pfad ab, den ein Unix-Socket nicht fasst —
// **gemessen**: mit 121 Bytes legte `fm serve` keinen Socket an und lauschte
// still auf allen Schnittstellen (Vermerk 0166).
func PruefeSocketpfad(pfad string) error {
	if len(pfad) >= socketgrenze {
		return fmt.Errorf("%w: der Socketpfad hat %d Bytes, ein Unix-Socket fasst %d",
			domain.ErrValidation, len(pfad), socketgrenze-1)
	}

	return nil
}

// Prozess ist ein laufendes `fm serve`.
type Prozess interface{ Beende() error }

// Starter startet `fm` mit diesen Argumenten.
type Starter func(ctx context.Context, argumente ...string) (Prozess, error)

// Dienst hält **ein** `fm serve` je Anwendungsprozess — gestartet bei
// Bedarf, beendet mit [Dienst.Beende].
type Dienst struct {
	ordner  string
	starter Starter
	frist   time.Duration

	// lebt und verwaisteBeenden räumen ab, was ein Absturz zurückließ —
	// eingesetzt, damit kein Lauf einen fremden Prozess fragt.
	lebt             func(kennung int) bool
	verwaisteBeenden func(ctx context.Context, pfad string) error

	sperre  sync.Mutex
	prozess Prozess
	klient  *Klient

	// abgeschaltet hält jeden Start auf — „für diese Sitzung aus"
	// (AI-05.AK-04). **Unter derselben Sperre wie der Start** (Z2 des
	// Zweitblicks über Phase 6.10): ein Schalter in der Anwendung, gesetzt
	// vor dem Beenden, ließ einen Vorschlag, der den Anbieter schon gewählt
	// hatte, `fm serve` danach neu starten.
	abgeschaltet bool

	// startAbbruch bricht einen Start ab, der gerade auf `/health` wartet —
	// **unter einer eigenen Sperre**, denn der Start hält [Dienst.sperre]
	// bis zu [startfrist] (BK-02, #2110).
	startsperre  sync.Mutex
	startAbbruch context.CancelFunc
}

// ErrAbgeschaltet heißt: das Gerätemodell ist für diese Sitzung aus.
var ErrAbgeschaltet = fmt.Errorf("%w: das Gerätemodell ist für diese Sitzung abgeschaltet",
	domain.ErrValidation)

// NeuerDienst baut einen Dienst über einem Ordner, in dem der Socket liegt —
// **kurz**, etwa unter `$TMPDIR`.
func NeuerDienst(ordner string, starter Starter) *Dienst {
	return &Dienst{
		ordner: ordner, starter: starter, frist: startfrist,
		lebt: prozessLebt, verwaisteBeenden: verwaistenBeenden,
	}
}

// eigenerOrdner legt den Ordner des Sockets an — oder nimmt einen
// vorhandenen, **wenn er ein Ordner mit `0700` ist**. Einer mit fremden
// Rechten, ein Symlink oder eine Datei an seiner Stelle wird abgewiesen und
// nicht still umgestellt: wer ihn vorbereitet hat, ist nicht dieses Programm.
func eigenerOrdner(ordner string) (string, error) {
	if fehler := os.Mkdir(ordner, 0o700); fehler != nil && !errors.Is(fehler, os.ErrExist) {
		return "", fmt.Errorf("den Ordner des Sockets anlegen: %w", fehler)
	}

	info, fehler := os.Lstat(ordner)
	if fehler != nil || !info.IsDir() || info.Mode().Perm() != 0o700 {
		return "", fmt.Errorf("%w: %s ist kein Ordner mit Rechten 0700 — er wird nicht benutzt",
			domain.ErrValidation, ordner)
	}

	return ordner, nil
}

// startfrist ist, wie lange auf `/health` gewartet wird — gemessen waren
// 0,15 s; zehn Sekunden lassen einem kalten Modell Zeit.
const startfrist = 10 * time.Second

// Klient gibt einen Klienten an einem laufenden Dienst — und startet ihn,
// wenn er noch nicht läuft.
func (d *Dienst) Klient(ctx context.Context) (*Klient, error) {
	d.sperre.Lock()
	defer d.sperre.Unlock()

	if d.abgeschaltet {
		return nil, ErrAbgeschaltet
	}

	// **Ein gemerkter Klient muss noch antworten** — Z3 des Zweitblicks über
	// Phase 6.10: stürzte `fm serve` ab, gingen bis zum Beenden der Anwendung
	// alle Fragen an einen toten Socket. Ein toter wird aufgeräumt und neu
	// gestartet; `/health` kostet am lokalen Socket unter einer Millisekunde.
	//
	// **Und nur ein Ausfall des Dienstes zählt** — N3 des Nachlaufs: ist bloß
	// der Kontext dieses Rufs abgebrochen oder abgelaufen, scheitert die
	// Frage ebenso, und ein gesunder Dienst, den ein anderer Ruf gerade
	// nutzt, wurde beendet.
	if d.klient != nil {
		fehler := d.klient.Gesund(ctx)
		if fehler == nil {
			return d.klient, nil
		}

		if ctx.Err() != nil {
			return nil, fmt.Errorf("das Gerätemodell fragen: %w", ctx.Err())
		}

		_ = d.beendeGesperrt() //nolint:errcheck // er ist schon fort; gezählt wird der neue Start
	}

	ordner, fehler := eigenerOrdner(filepath.Join(d.ordner, "mse-fm"))
	if fehler != nil {
		return nil, fehler
	}

	d.aufraeumen(ctx, ordner)

	// **Ein Stop bricht den Start ab** und wartet nicht auf ihn: ohne das
	// hielt `Beende` die Sperre, bis die Startfrist ablief — und mit ihm das
	// Fenster, das jemand schließen wollte (BK-02, #2110).
	ctx, abbrechen := context.WithCancel(ctx)
	defer abbrechen()

	d.startsperre.Lock()
	d.startAbbruch = abbrechen
	d.startsperre.Unlock()

	defer func() {
		d.startsperre.Lock()
		d.startAbbruch = nil
		d.startsperre.Unlock()
	}()

	pfad := filepath.Join(ordner, fmt.Sprintf("fm-%d.sock", os.Getpid()))
	if zuLang := PruefeSocketpfad(pfad); zuLang != nil {
		return nil, zuLang
	}

	_ = os.Remove(pfad) //nolint:errcheck // ein verwaister Socket eines früheren Laufs

	prozess, fehler := d.starter(ctx, "serve", "--socket", pfad)
	if fehler != nil {
		return nil, fehler
	}

	klient := NeuerKlient(netz.NeuerKlientUeberSocket(netz.UeblicheFristen(), pfad), Socketadresse)
	if fehler := wartenBisGesund(ctx, klient, pfad, d.frist); fehler != nil {
		_ = prozess.Beende() //nolint:errcheck // der Fehler davor zählt

		return nil, fehler
	}

	d.prozess, d.klient = prozess, klient

	return klient, nil
}

// wartenBisGesund wartet, bis der **Socket** dasteht und `/health` antwortet.
// Ohne Socket wird nicht gefragt: dann lauschte `fm` vielleicht auf TCP.
func wartenBisGesund(ctx context.Context, klient *Klient, pfad string, frist time.Duration) error {
	ende := time.Now().Add(frist)

	for time.Now().Before(ende) {
		if info, fehler := os.Stat(pfad); fehler == nil && info.Mode()&os.ModeSocket != 0 &&
			klient.Gesund(ctx) == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("das Gerätemodell starten: %w", ctx.Err())
		case <-time.After(50 * time.Millisecond):
		}
	}

	return fmt.Errorf("%w: das Gerätemodell antwortete nicht binnen %s", domain.ErrValidation, frist)
}

// Beende beendet einen laufenden Dienst — ohne laufenden tut es nichts.
func (d *Dienst) Beende() error {
	d.startAbbrechen()

	d.sperre.Lock()
	defer d.sperre.Unlock()

	return d.beendeGesperrt()
}

// startAbbrechen bricht einen laufenden Start ab — ohne laufenden tut es
// nichts.
func (d *Dienst) startAbbrechen() {
	d.startsperre.Lock()
	defer d.startsperre.Unlock()

	if d.startAbbruch != nil {
		d.startAbbruch()
	}
}

// Abschalten beendet einen laufenden Dienst und hält jeden weiteren Start
// auf, bis [Dienst.Anschalten] — **in einem Zug**: wer danach fragt, bekommt
// [ErrAbgeschaltet], auch wenn er den Anbieter vorher gewählt hat.
func (d *Dienst) Abschalten() error {
	d.startAbbrechen()

	d.sperre.Lock()
	defer d.sperre.Unlock()

	d.abgeschaltet = true

	return d.beendeGesperrt()
}

// Abgeschaltet sagt, ob Starts gerade aufgehalten werden.
func (d *Dienst) Abgeschaltet() bool {
	d.sperre.Lock()
	defer d.sperre.Unlock()

	return d.abgeschaltet
}

// Anschalten lässt Starts wieder zu.
func (d *Dienst) Anschalten() {
	d.sperre.Lock()
	defer d.sperre.Unlock()

	d.abgeschaltet = false
}

// beendeGesperrt beendet den Prozess; die Sperre hält der Rufer.
func (d *Dienst) beendeGesperrt() error {
	if d.prozess == nil {
		return nil
	}

	fehler := d.prozess.Beende()
	d.prozess, d.klient = nil, nil

	return fehler
}

// starte startet einen fertigen Befehl — getrennt von [Prozessstarter], aus
// demselben Grund wie [exitcodeVon].
func starte(befehl *exec.Cmd) (Prozess, error) {
	befehl.Env = append(os.Environ(), "NO_COLOR=1")

	if fehler := befehl.Start(); fehler != nil {
		return nil, fmt.Errorf("%s starten: %w", befehl.Path, fehler)
	}

	return befehlsprozess{befehl: befehl}, nil
}

// befehlsprozess beendet ein gestartetes Programm und wartet auf es.
type befehlsprozess struct{ befehl *exec.Cmd }

func (p befehlsprozess) Beende() error {
	if fehler := p.befehl.Process.Kill(); fehler != nil {
		return fmt.Errorf("das Gerätemodell beenden: %w", fehler)
	}

	_ = p.befehl.Wait() //nolint:errcheck // ein beendeter Prozess meldet sein Signal

	return nil
}
