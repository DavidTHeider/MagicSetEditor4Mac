package app

// **Die Protokollweiche** — OPS-04
// (#1721).
//
// # Warum eine Weiche
//
// **Weil das Protokoll vor dem Bestand entsteht und seine Einstellungen im
// Bestand liegen** (#848). `main` setzt die Weiche mit der Vorgabe als
// Standard; sobald der Bestand offen ist, richtet `App` sie nach den
// Einstellungen ein. Jeder Schreiber — auch einer, der vorher mit `With`
// gebunden wurde — hängt an der Weiche und nicht an einem Kanal, und
// schreibt danach dorthin, wo es gilt.
//
// # Drei Kanäle, und einer, der ausfällt, nimmt die anderen nicht mit
//
// **AK-02**: hier stand in `main` ein Protokoll, das bei einer
// unbeschreibbaren Datei ganz verstummte — auch das Unified Log. Jetzt
// entsteht jeder Kanal für sich, und der Stand sagt, welcher warum fehlt;
// die Diagnose zeigt ihn.

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Namen der Kanäle, wie der Stand sie führt.
const (
	KanalDatei   = "datei"
	KanalSystem  = "system"
	KanalAusgabe = "ausgabe"
)

// Protokollkanal ist ein Kanal und ob er schreibt.
type Protokollkanal struct {
	Name  string `json:"name"`
	Aktiv bool   `json:"aktiv"`

	// Grund sagt, warum ein gewählter Kanal nicht schreibt — leer, wenn er
	// schreibt oder abgewählt ist.
	Grund string `json:"grund"`
}

// Protokollstand ist, was das Protokoll gerade tut.
type Protokollstand struct {
	Stufe        string           `json:"stufe"`
	Format       string           `json:"format"`
	Ausfuehrlich bool             `json:"ausfuehrlich"`
	Kanaele      []Protokollkanal `json:"kanaele"`

	// Datei ist die Datei, in die geschrieben wird — leer ohne Dateikanal.
	Datei string `json:"datei"`
}

// Protokollweiche leitet jede Zeile an die Kanäle, die gerade gelten.
type Protokollweiche struct {
	schloss sync.RWMutex
	innen   slog.Handler
	stand   Protokollstand
	zu      func() error
	kanaele verteiler

	bestandsordner string
	system         io.Writer
	ausgabe        io.Writer
	heim           func() (string, error)
}

// NeueProtokollweiche baut die Weiche mit der Vorgabe — die Datei neben dem
// Bestand und das Systemprotokoll.
func NeueProtokollweiche(bestandsordner string) *Protokollweiche {
	return neueWeicheMit(bestandsordner, systemprotokoll(), os.Stderr, os.UserHomeDir)
}

func neueWeicheMit(bestandsordner string, system, ausgabe io.Writer, heim func() (string, error)) *Protokollweiche {
	weiche := &Protokollweiche{bestandsordner: bestandsordner, system: system, ausgabe: ausgabe, heim: heim}

	// **Die Startweiche dreht nicht** — Nachlauf über Phase 6.4, N2. Sie
	// entsteht vor dem Bestand, also mit der Vorgabe; drehte sie damit, wirkte
	// eine Grenze über 4 MB nie, und eine Aufbewahrung über 1 auch nicht — die
	// Vorgabe überschrieb `mse.log.1`, bevor die Einstellung galt. Gedreht
	// wird beim ersten [Protokollweiche.Einrichten] mit der echten Regel.
	anfang := domain.ProtokollregelAus(nil)
	anfang.GrenzeBytes = math.MaxInt64
	weiche.Einrichten(anfang)

	return weiche
}

// Logger gibt einen Schreiber, der an der Weiche hängt.
func (w *Protokollweiche) Logger() *slog.Logger { return slog.New(&weichenhandler{weiche: w}) }

// Stand gibt, was das Protokoll gerade tut.
func (w *Protokollweiche) Stand() Protokollstand {
	w.schloss.RLock()
	defer w.schloss.RUnlock()

	return mitSchreibfehlern(w.stand, w.kanaele)
}

// Einrichten baut die Kanäle nach einer Regel neu und schließt die alten.
//
// **Erst gebaut, dann getauscht, dann geschlossen** — zwischen zwei Zeilen
// gibt es keinen Augenblick ohne Kanal. **Außer bei derselben Datei**: sie
// zu drehen verlangt, dass niemand sie hält, und der Halter ist die alte
// Datei dieser Weiche (Befund 6 des Zweitblicks über Phase 6.4 — am
// Vorgabeort griff die eingestellte Drehung deshalb nie). Dann schließt die
// alte zuerst; die anderen Kanäle schreiben bis dahin weiter.
func (w *Protokollweiche) Einrichten(regel domain.Protokollregel) {
	if ordner, _ := w.dateiordner(regel); regel.Datei && ordner != "" &&
		w.Stand().Datei == filepath.Join(ordner, Protokollname) {
		w.dateiSchliessen()
	}

	innen, stand, zu, kanaele := w.kanaeleBauen(regel)

	w.schloss.Lock()
	alt := w.zu
	w.innen, w.stand, w.zu, w.kanaele = innen, stand, zu, kanaele
	w.schloss.Unlock()

	// **Die alte Datei schließt still** — ihre Zeilen sind geschrieben, und
	// ein Fehler beim Schließen hätte keinen Empfänger mehr als das Protokoll,
	// das gerade wechselt.
	if alt != nil && alt() != nil {
		return
	}
}

// Schliessen schließt die Datei, wenn eine offen ist — **und das Unified
// Log**: seine abgelegten Zeilen gehen dabei noch hinaus, höchstens fünf
// Sekunden lang (`systemkanal.go`, PERF03).
func (w *Protokollweiche) Schliessen() error {
	w.schloss.Lock()
	defer w.schloss.Unlock()

	var system error
	if kanal, schliessbar := w.system.(io.Closer); schliessbar {
		system = kanal.Close()
	}

	if w.zu == nil {
		return system
	}

	zu := w.zu
	w.zu = nil

	return errors.Join(zu(), system)
}

func (w *Protokollweiche) aktuell() slog.Handler {
	w.schloss.RLock()
	defer w.schloss.RUnlock()

	return w.innen
}

// dateiSchliessen schließt die Datei und lässt die anderen Kanäle stehen.
func (w *Protokollweiche) dateiSchliessen() {
	w.schloss.Lock()
	defer w.schloss.Unlock()

	if w.zu != nil && w.zu() == nil {
		w.zu = nil
	}
}

// kanaeleBauen baut die Kanäle einer Regel.
func (w *Protokollweiche) kanaeleBauen(regel domain.Protokollregel) (
	slog.Handler, Protokollstand, func() error, verteiler,
) {
	stand := Protokollstand{Stufe: regel.Stufe, Format: regel.Format, Ausfuehrlich: regel.Ausfuehrlich()}
	schreiber := verteiler{}
	zu := func() error { return nil }

	if regel.Datei {
		datei, pfad, grund := w.dateiOeffnen(regel)
		stand.Kanaele = append(stand.Kanaele, Protokollkanal{Name: KanalDatei, Aktiv: datei != nil, Grund: grund})

		if datei != nil {
			schreiber = append(schreiber, &kanalschreiber{name: KanalDatei, ziel: datei})
			stand.Datei = pfad
			zu = datei.Close
		}
	} else {
		stand.Kanaele = append(stand.Kanaele, Protokollkanal{Name: KanalDatei})
	}

	for _, kanal := range []struct {
		name     string
		gewaehlt bool
		ziel     io.Writer
	}{{KanalSystem, regel.System, w.system}, {KanalAusgabe, regel.Ausgabe, w.ausgabe}} {
		stand.Kanaele = append(stand.Kanaele, Protokollkanal{Name: kanal.name, Aktiv: kanal.gewaehlt})
		if kanal.gewaehlt {
			schreiber = append(schreiber, &kanalschreiber{name: kanal.name, ziel: kanal.ziel})
		}
	}

	// **Ein Schreiber je Kanal, jeder mit seiner Stufe** — Nachlauf über
	// Phase 6.4, N5: das Unified Log nimmt ab „info". Auf der ausführlichen
	// Stufe steht jeder Ruf des Fensters, und dort startete jede Zeile einen
	// eigenen `logger`.
	handler := make([]slog.Handler, 0, len(schreiber))

	for _, kanal := range schreiber {
		stufe := stufeAus(regel.Stufe)
		if kanal.name == KanalSystem {
			stufe = max(stufe, slog.LevelInfo)
		}

		optionen := &slog.HandlerOptions{Level: stufe}

		var einer slog.Handler = slog.NewTextHandler(kanal, optionen)
		if regel.Format == domain.ProtokollformatJSON {
			einer = slog.NewJSONHandler(kanal, optionen)
		}

		handler = append(handler, einer)
	}

	var innen slog.Handler = slog.NewMultiHandler(handler...)

	if len(regel.Komponenten) > 0 {
		innen = &komponentenfilter{innen: innen, komponenten: regel.Komponenten}
	}

	return &geschwaerzt{innen: innen}, stand, zu, schreiber
}

// dateiordner gibt den Ordner der Datei — oder den Grund, warum es keinen gibt.
func (w *Protokollweiche) dateiordner(regel domain.Protokollregel) (ordner, grund string) {
	if regel.Ort == "" {
		if w.bestandsordner == "" {
			return "", "der Ort des Bestands ist unbekannt"
		}

		return w.bestandsordner, ""
	}

	zuhause, fehler := w.heim()
	if fehler != nil {
		return "", "das Zuhause für »" + regel.Ort + "« ließ sich nicht bestimmen"
	}

	return domain.OrdnerAus(regel.Ort, zuhause), ""
}

// dateiOeffnen öffnet die Datei eines Dateikanals — gedreht nach der Regel.
//
// **Ein eigener Ort wird nicht angelegt** — Befund 2 des Zweitblicks über
// Phase 6.4: auf einem abgezogenen Laufwerk entstünde er sonst still unter
// `/Volumes` auf der eigenen Platte. Der Ordner neben dem Bestand gehört dem
// Programm und entsteht, wenn er fehlt.
func (w *Protokollweiche) dateiOeffnen(regel domain.Protokollregel) (datei *protokolldatei, pfad, grund string) {
	ordner, grund := w.dateiordner(regel)
	if ordner == "" {
		return nil, "", grund
	}

	pfad = filepath.Join(ordner, Protokollname)

	var anlegen error
	if regel.Ort == "" {
		anlegen = os.MkdirAll(ordner, 0o750)
	} else {
		_, anlegen = os.Stat(ordner)
	}

	if anlegen != nil {
		return nil, "", schwaerze(anlegen.Error())
	}

	griff := drehenMit(pfad, regel.GrenzeBytes, regel.Aufbewahrung)

	// #nosec G304 -- der Ort ist eine geprüfte Einstellung des Geräts, nicht die Angabe eines Aufrufers
	offen, fehler := os.OpenFile(pfad, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if fehler != nil {
		schliesseGriff(griff)

		return nil, "", schwaerze(fehler.Error())
	}

	return &protokolldatei{
		File: offen, griff: griff, pfad: pfad, grenze: regel.GrenzeBytes, aufbewahrung: regel.Aufbewahrung,
	}, pfad, ""
}

// stufeAus übersetzt eine Stufe der Einstellung.
func stufeAus(stufe string) slog.Level {
	switch stufe {
	case domain.ProtokollstufeFehler:
		return slog.LevelError
	case domain.ProtokollstufeWarnung:
		return slog.LevelWarn
	case domain.ProtokollstufeAusfuehrlich:
		return slog.LevelDebug
	default:
		return slog.LevelInfo
	}
}

// weichenhandler hängt an der Weiche und gibt jede Zeile dem Handler, der
// gerade gilt — samt den gebundenen Merkmalen und Gruppen.
type weichenhandler struct {
	weiche *Protokollweiche
	binden []func(slog.Handler) slog.Handler
}

func (h *weichenhandler) handler() slog.Handler {
	innen := h.weiche.aktuell()
	for _, bindung := range h.binden {
		innen = bindung(innen)
	}

	return innen
}

func (h *weichenhandler) Enabled(ctx context.Context, stufe slog.Level) bool {
	return h.weiche.aktuell().Enabled(ctx, stufe)
}

func (h *weichenhandler) Handle(ctx context.Context, satz slog.Record) error {
	return h.handler().Handle(ctx, satz) //nolint:wrapcheck // ein Protokoll reicht durch
}

func (h *weichenhandler) WithAttrs(merkmale []slog.Attr) slog.Handler {
	return &weichenhandler{weiche: h.weiche, binden: append(slices.Clone(h.binden),
		func(innen slog.Handler) slog.Handler { return innen.WithAttrs(merkmale) })}
}

func (h *weichenhandler) WithGroup(name string) slog.Handler {
	return &weichenhandler{weiche: h.weiche, binden: append(slices.Clone(h.binden),
		func(innen slog.Handler) slog.Handler { return innen.WithGroup(name) })}
}

// komponentenfilter lässt eine Zeile unter »info« nur aus den genannten
// Komponenten durch — ab »info« jede.
//
// **Die Komponente ist das Paket, aus dem die Zeile stammt** — der letzte
// Teil seines Pfads, ohne `_test`: `service`, `storage`, `app`.
type komponentenfilter struct {
	innen       slog.Handler
	komponenten []string
}

func (f *komponentenfilter) Enabled(ctx context.Context, stufe slog.Level) bool {
	return f.innen.Enabled(ctx, stufe)
}

func (f *komponentenfilter) Handle(ctx context.Context, satz slog.Record) error {
	if satz.Level < slog.LevelInfo && !slices.Contains(f.komponenten, komponenteVon(satz.PC)) {
		return nil
	}

	return f.innen.Handle(ctx, satz) //nolint:wrapcheck // ein Protokoll reicht durch
}

func (f *komponentenfilter) WithAttrs(merkmale []slog.Attr) slog.Handler {
	return &komponentenfilter{innen: f.innen.WithAttrs(merkmale), komponenten: f.komponenten}
}

func (f *komponentenfilter) WithGroup(name string) slog.Handler {
	return &komponentenfilter{innen: f.innen.WithGroup(name), komponenten: f.komponenten}
}

// komponenteVon gibt das Paket einer Stelle — `service` aus
// `…/internal/service.(*Verwalter).Setzen`.
func komponenteVon(stelle uintptr) string {
	rahmen, _ := runtime.CallersFrames([]uintptr{stelle}).Next()
	name := rahmen.Function

	if schnitt := strings.LastIndex(name, "/"); schnitt >= 0 {
		name = name[schnitt+1:]
	}

	paket, _, _ := strings.Cut(name, ".")

	return strings.TrimSuffix(paket, "_test")
}
