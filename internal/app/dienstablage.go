package app

// **Die Ablage des Dienstes** — FMT-03.AK-03
// (#1692).
//
// Zwei Verzeichnisse neben dem Bestand, und sie haben verschiedene Regeln:
//
//   - **Eingang**: was ein Aufrufer übertragen hat, **nach seinem Inhalt
//     benannt und dauerhaft**. Ein Set liest seine Bilder später aus seiner
//     Quelle (`vorlage.Setpaket = kopf.Origin`), und eine Quelle, die nach
//     dem Einfahren verschwände, ließe ein Set ohne Bilder zurück. Derselbe
//     Inhalt bekommt denselben Ort — so gilt „eine Quelle, ein Set" auch für
//     eine zweite Übertragung.
//   - **Ausgang**: was eine Ausgabe ohne `ziel` geschrieben hat, **eine
//     Stunde lang** zum Abholen. Der Name ist gewürfelt; wer ihn nicht aus
//     einer Antwort kennt, rät ihn nicht.
//   - **Vorschau**: was nur angesehen wird, liegt dort für **eine** Anfrage
//     und geht danach (#1944). Nicht im Eingang, denn dort teilt derselbe
//     Inhalt denselben Ort — die Vorschau nähme sonst womöglich einem Set
//     seine Quelle weg.
//
// **Der Eingang räumt höchstens stündlich ab**, was älter als eine Stunde
// ist und worauf kein Set irgendeines Profils mehr verweist — eine
// Vorschau von früher, die Quelle eines gelöschten Sets. Die Stunde schützt
// ein Einfahren, das gerade liest, und die Wiederholung eines Auftrags, die
// so lange möglich ist. Wer nicht sicher sagen kann, ob verwiesen wird,
// behält die Datei.
//
// **Abholen und Vergessen nehmen nur einen schlichten Dateinamen** — und
// gehen über `os.Root`, damit ein Name wie `../eingang/…` nicht aus dem
// Ausgang herausführt, auch wenn jemand ihn doch durchließe.

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ausgangsfrist ist, wie lange eine Ausgabe zum Abholen liegt; eingangsfrist,
// wie lange eine Quelle mindestens bleibt und wie oft der Eingang aufräumt —
// so lange, wie ein beendeter Auftrag wiederholt werden kann.
const (
	ausgangsfrist = time.Hour
	eingangsfrist = time.Hour
)

// Dienstablage hält übertragene Dateien und Ausgaben des Dienstes.
type Dienstablage struct {
	eingang, ausgang, vorschau string

	// verwiesen sagt, ob ein Set auf eine Quelle verweist; ohne ihn räumt
	// der Eingang nichts ab.
	verwiesen func(ctx context.Context, pfad string) (bool, error)

	sperre      sync.Mutex
	aufgeraeumt time.Time
}

// NeueDienstablage baut die Ablage unter `ort`.
func NeueDienstablage(ort string) *Dienstablage {
	return &Dienstablage{
		eingang: filepath.Join(ort, "eingang"), ausgang: filepath.Join(ort, "ausgang"),
		vorschau: filepath.Join(ort, "vorschau"),
	}
}

// MitVerweis gibt der Ablage die Frage, ob ein Set auf eine Quelle verweist.
func (a *Dienstablage) MitVerweis(verwiesen func(ctx context.Context, pfad string) (bool, error)) *Dienstablage {
	a.verwiesen = verwiesen

	return a
}

// Ablegen schreibt einen Inhalt in den Eingang und gibt seinen Ort.
func (a *Dienstablage) Ablegen(ctx context.Context, endung string, inhalt io.Reader) (string, error) {
	if err := os.MkdirAll(a.eingang, 0o700); err != nil {
		return "", fmt.Errorf("den Eingang anlegen: %w", err)
	}

	bau, err := os.CreateTemp(a.eingang, ".eingang-*")
	if err != nil {
		return "", fmt.Errorf("im Eingang anlegen: %w", err)
	}

	summe := sha256.New()
	_, kopierfehler := io.Copy(io.MultiWriter(bau, summe), inhalt)

	if err := errors.Join(kopierfehler, bau.Close()); err != nil {
		_ = os.Remove(bau.Name()) //nolint:errcheck // Wegwerfdatei

		return "", err
	}

	ziel := filepath.Join(a.eingang, hex.EncodeToString(summe.Sum(nil))[:32]+endung)

	// **Derselbe Inhalt darf sich selbst ersetzen** — der Name ist die
	// Prüfsumme, und das Umbenennen ist atomar: wer die alte Datei gerade
	// liest, liest sie zu Ende.
	if err := os.Rename(bau.Name(), ziel); err != nil {
		_ = os.Remove(bau.Name()) //nolint:errcheck // Wegwerfdatei

		return "", fmt.Errorf("in den Eingang legen: %w", err)
	}

	a.eingangAufraeumen(ctx)

	return ziel, nil
}

// Zwischenlegen legt einen Inhalt für **eine** Anfrage ab — die Vorschau
// (#1944). `weg` nimmt ihn wieder weg.
func (a *Dienstablage) Zwischenlegen(_ context.Context, endung string, inhalt io.Reader) (pfad string, weg func(), err error) {
	if anlegefehler := os.MkdirAll(a.vorschau, 0o700); anlegefehler != nil {
		return "", nil, fmt.Errorf("die Vorschau anlegen: %w", anlegefehler)
	}

	datei, err := os.CreateTemp(a.vorschau, "*"+endung)
	if err != nil {
		return "", nil, fmt.Errorf("in der Vorschau anlegen: %w", err)
	}

	weg = func() { _ = os.Remove(datei.Name()) } //nolint:errcheck // Wegwerfdatei

	_, kopierfehler := io.Copy(datei, inhalt)
	if fehler := errors.Join(kopierfehler, datei.Close()); fehler != nil {
		weg()

		return "", nil, fehler
	}

	return datei.Name(), weg, nil
}

// eingangAufraeumen nimmt, worauf niemand mehr verweist — höchstens einmal
// je Frist, und nur, was älter als die Frist ist.
func (a *Dienstablage) eingangAufraeumen(ctx context.Context) {
	a.sperre.Lock()
	faellig := a.verwiesen != nil && time.Since(a.aufgeraeumt) > eingangsfrist
	if faellig {
		a.aufgeraeumt = time.Now()
	}
	a.sperre.Unlock()

	if !faellig {
		return
	}

	eintraege, _ := os.ReadDir(a.eingang) //nolint:errcheck // ohne Liste nichts abzuräumen

	for _, eintrag := range eintraege {
		pfad := filepath.Join(a.eingang, eintrag.Name())
		info, err := eintrag.Info()

		// **Halbe Übertragungen gehören dem, der gerade schreibt.**
		if err != nil || strings.HasPrefix(eintrag.Name(), ".") || time.Since(info.ModTime()) <= eingangsfrist {
			continue
		}

		if verwiesen, err := a.verwiesen(ctx, pfad); err == nil && !verwiesen {
			_ = os.Remove(pfad) //nolint:errcheck // der nächste Lauf versucht es wieder
		}
	}
}

// Ausgangsort gibt einen freien Ort für eine Ausgabe — und räumt ab, was
// länger als die Frist liegt.
func (a *Dienstablage) Ausgangsort(_ context.Context, endung string) (name, pfad string, err error) {
	if err := os.MkdirAll(a.ausgang, 0o700); err != nil {
		return "", "", fmt.Errorf("den Ausgang anlegen: %w", err)
	}

	a.aufraeumen()

	name = strings.ToLower(rand.Text()[:16]) + endung

	return name, filepath.Join(a.ausgang, name), nil
}

// aufraeumen nimmt, was älter als die Frist ist — ein Fehler dabei hält
// keine Ausgabe auf.
func (a *Dienstablage) aufraeumen() {
	eintraege, _ := os.ReadDir(a.ausgang) //nolint:errcheck // ohne Liste nichts abzuräumen

	for _, eintrag := range eintraege {
		if info, err := eintrag.Info(); err == nil && time.Since(info.ModTime()) > ausgangsfrist {
			_ = os.Remove(filepath.Join(a.ausgang, eintrag.Name())) //nolint:errcheck // der nächste Lauf versucht es wieder
		}
	}
}

// Abholen öffnet eine Ausgabe.
func (a *Dienstablage) Abholen(_ context.Context, name string) (io.ReadCloser, error) {
	wurzel, err := a.ausgangswurzel(name)
	if err != nil {
		return nil, err
	}

	defer func() { _ = wurzel.Close() }() //nolint:errcheck // nur geöffnet

	datei, err := wurzel.Open(name)
	if err != nil {
		return nil, nichtDa(name, err)
	}

	return datei, nil
}

// Vergessen nimmt eine Ausgabe weg.
func (a *Dienstablage) Vergessen(_ context.Context, name string) error {
	wurzel, err := a.ausgangswurzel(name)
	if err != nil {
		return err
	}

	defer func() { _ = wurzel.Close() }() //nolint:errcheck // nur geöffnet

	if err := wurzel.Remove(name); err != nil {
		return nichtDa(name, err)
	}

	return nil
}

// ausgangswurzel prüft den Namen und öffnet den Ausgang als Wurzel.
func (a *Dienstablage) ausgangswurzel(name string) (*os.Root, error) {
	if name == "" || name != filepath.Base(name) || strings.HasPrefix(name, ".") {
		return nil, fmt.Errorf("%w: eine Ausgabe %q gibt es nicht", domain.ErrNotFound, name)
	}

	wurzel, err := os.OpenRoot(a.ausgang)
	if err != nil {
		return nil, fmt.Errorf("den Ausgang öffnen: %w", err)
	}

	return wurzel, nil
}

// nichtDa macht aus einer fehlenden Datei ein „gibt es nicht".
func nichtDa(name string, fehler error) error {
	if errors.Is(fehler, fs.ErrNotExist) {
		return fmt.Errorf("%w: die Ausgabe %q gibt es nicht — abgeholt, vergessen oder älter als eine Stunde",
			domain.ErrNotFound, name)
	}

	return fmt.Errorf("die Ausgabe %q: %w", name, fehler)
}
