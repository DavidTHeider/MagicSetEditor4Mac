package app

// **Das Betriebsband: was sich von selbst meldet** — OPS02
// (#2348),
// Befund A-07 des Audits 0.9.0.
//
// # Was hier steht und was nicht
//
// **Gemessen wird hier** — der Platz neben dem Bestand, das Alter der
// Sicherung, der Dienst beim Anmelden, die Neustarts des Webprozesses.
// **Wann eine Lage ein Hinweis ist und wohin er führt**, steht in
// `domain/betriebsband.go`.
//
// # Ein Gedächtnis je Anwendung
//
// **Die Schwellen haben eine Vorgeschichte**, damit nichts flattert: der
// Platz wird erst über einer zweiten Schwelle wieder gut, ein abstürzender
// Dienst nimmt den Hinweis nicht in seinen Millisekunden zurück. Das hält
// [betriebsgedaechtnis] unter einer Sperre — zwei Abrufe zugleich sähen
// sonst dieselbe Vorgeschichte und schrieben zwei verschiedene.
//
// # Höchstens alle 30 Sekunden, und billig
//
// Das Fenster fragt alle 30 s (`haken/useBetriebsband.ts`). **Jeder Abruf
// misst seinen Aufwand selbst** (`aufwandMikros`); `launchctl print` kostet
// 3 bis 5 ms (gemessen am 2026-10-15) und läuft nur, wenn die plist des
// Dienstes daliegt.

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// BetriebshinweisDTO ist ein Hinweis des Bands — `art` wählt den Satz im
// Katalog, `werte` füllt ihn, `ziel` sagt, wohin er führt.
type BetriebshinweisDTO struct {
	Art   string            `json:"art"`
	Ziel  string            `json:"ziel"`
	Werte map[string]string `json:"werte"`
}

// BetriebsbandAntwort trägt die Hinweise eines Abrufs.
type BetriebsbandAntwort struct {
	Hinweise []BetriebshinweisDTO `json:"hinweise"`

	// Ungemessen nennt, was sich nicht fragen ließ — `platz`, `sicherung`,
	// `dienst`. **Ein gescheiterter Nachweis gibt keinen neuen Hinweis**; ein
	// stehender bleibt.
	Ungemessen []string `json:"ungemessen"`

	AufwandMikros int64   `json:"aufwandMikros"`
	Fehler        *Fehler `json:"fehler"`
}

// Was sich nicht fragen ließ.
const (
	ungemessenPlatz     = "platz"
	ungemessenSicherung = "sicherung"
	ungemessenDienst    = "dienst"
)

// betriebsquellen sind die Messungen — eingesetzt, damit jede Lage ohne
// volle Platte und ohne launchd prüfbar ist.
type betriebsquellen struct {
	platz     func() (frei, gesamt uint64, fehler error)
	sicherung func(ctx context.Context) (SicherungsstandDTO, error)
	dienst    func(ctx context.Context) (domain.Skriptdienstlage, error)
	neustarts func() int
}

// betriebsgedaechtnis ist die Vorgeschichte der Schwellen und der Zähler
// der Webprozesse.
type betriebsgedaechtnis struct {
	sperre sync.Mutex

	platzKnapp bool
	frei       uint64

	// sicherungAlt und sicherungTage sind die letzte gelungene Messung der
	// Sicherung — ein Lesefehler ersetzt sie nicht (Zweitblick über Phase 4,
	// Befund 10).
	sicherungAlt  bool
	sicherungTage int

	dienst        domain.Skriptdienstlage
	dienstBeendet bool

	// laufSeit ist, seit wann derselbe Lauf des Dienstes gesehen wird —
	// **an der monotonen Uhr**, die `time.Now` mitträgt (Zweitblick über
	// Phase 4, Befund 11). laufGesehen sagt, ob es ihn gibt; uhr ist
	// einsetzbar, sonst `time.Now`.
	laufSeit    time.Time
	laufGesehen bool
	uhr         func() time.Time

	webneustarts atomic.Int64
}

// abrufen misst einmal und gibt die Hinweise.
func (g *betriebsgedaechtnis) abrufen(ctx context.Context, q betriebsquellen) BetriebsbandAntwort {
	beginn := time.Now()
	antwort := BetriebsbandAntwort{Hinweise: []BetriebshinweisDTO{}, Ungemessen: []string{}}

	g.sperre.Lock()
	defer g.sperre.Unlock()

	if frei, gesamt, fehler := q.platz(); fehler != nil {
		antwort.Ungemessen = append(antwort.Ungemessen, ungemessenPlatz)
	} else {
		g.platzKnapp, g.frei = domain.PlatzKnapp(frei, gesamt, g.platzKnapp), frei
	}

	lage := domain.Betriebslage{PlatzKnapp: g.platzKnapp, FreiBytes: g.frei, Webneustarts: q.neustarts()}

	if stand, fehler := q.sicherung(ctx); fehler != nil {
		antwort.Ungemessen = append(antwort.Ungemessen, ungemessenSicherung)
	} else {
		g.sicherungAlt, g.sicherungTage = stand.Stufe == string(domain.SicherungsstufeAlt), stand.Tage
	}

	lage.SicherungAlt, lage.SicherungTage = g.sicherungAlt, g.sicherungTage

	jetzt, fehler := q.dienst(ctx)
	if fehler != nil {
		antwort.Ungemessen = append(antwort.Ungemessen, ungemessenDienst)
		jetzt = g.dienst
		jetzt.Unbekannt = true
	}

	g.dienstBeendet = domain.DienstBeendet(jetzt, g.dienst, g.dienstBeendet, g.gelaufen(jetzt, fehler == nil))
	if fehler == nil {
		g.dienst = jetzt
	}

	lage.DienstBeendet, lage.DienstAusgang = g.dienstBeendet, g.dienst.LetzterAusgang

	for _, hinweis := range domain.Betriebshinweise(lage) {
		antwort.Hinweise = append(antwort.Hinweise, BetriebshinweisDTO{
			Art: string(hinweis.Art), Ziel: hinweis.Ziel, Werte: hinweis.Werte,
		})
	}

	antwort.AufwandMikros = time.Since(beginn).Microseconds()

	return antwort
}

// gelaufen gibt, wie lange derselbe Lauf schon gesehen wird — **derselbe**
// heißt: er läuft, und `Laeufe` hat sich seit dem letzten Abruf nicht
// geändert. Ein Abruf ohne Antwort ändert nichts daran.
func (g *betriebsgedaechtnis) gelaufen(jetzt domain.Skriptdienstlage, gefragt bool) time.Duration {
	uhr := g.uhr
	if uhr == nil {
		uhr = time.Now
	}

	switch {
	case !gefragt:
	case !jetzt.Laeuft:
		g.laufGesehen = false
	case !g.laufGesehen || jetzt.Laeufe != g.dienst.Laeufe:
		g.laufSeit, g.laufGesehen = uhr(), true
	}

	if !g.laufGesehen {
		return 0
	}

	return uhr().Sub(g.laufSeit)
}

// Muster über `launchctl print` — **nur eine Einrückung**: tiefer
// eingerückte Zeilen gehören den Teilen eines Dienstes.
var (
	laeufeMuster  = regexp.MustCompile(`(?m)^\truns = (\d+)$`)
	ausgangMuster = regexp.MustCompile(`(?m)^\tlast exit code = (\d+)`)
	signalMuster  = regexp.MustCompile(`(?m)^\tlast terminating signal = [^:\n]*: (\d+)$`)
)

// dienstlageAus liest die Lage eines geladenen Dienstes. **Ein Signal zählt
// wie in der Shell als 128 plus seine Nummer** — beendet ist beendet, und
// ein `SIGKILL` ist kein sauberes Ende.
func dienstlageAus(ausgabe string) (domain.Skriptdienstlage, error) {
	pid, fehler := pidAus(ausgabe)
	if fehler != nil {
		return domain.Skriptdienstlage{}, fehler
	}

	lage := domain.Skriptdienstlage{Eingerichtet: true, Laeuft: pid != 0}

	if treffer := laeufeMuster.FindStringSubmatch(ausgabe); treffer != nil {
		lage.Laeufe, _ = strconv.Atoi(treffer[1]) //nolint:errcheck // das Muster sind Ziffern
	}

	if treffer := ausgangMuster.FindStringSubmatch(ausgabe); treffer != nil {
		lage.LetzterAusgang, _ = strconv.Atoi(treffer[1]) //nolint:errcheck // dito
	} else if treffer := signalMuster.FindStringSubmatch(ausgabe); treffer != nil {
		signal, _ := strconv.Atoi(treffer[1]) //nolint:errcheck // dito
		lage.LetzterAusgang = 128 + signal
	}

	return lage, nil
}

// dienstlageUeber fragt launchd nach dem Dienst, dessen plist `plist` ist —
// **ohne plist gar nicht**, denn dann ist keiner eingerichtet.
func dienstlageUeber(
	ctx context.Context, plist, etikett string, drucken func(context.Context, string) ([]byte, error),
) (domain.Skriptdienstlage, error) {
	if _, fehler := os.Stat(plist); errors.Is(fehler, fs.ErrNotExist) {
		return domain.Skriptdienstlage{}, nil
	} else if fehler != nil {
		return domain.Skriptdienstlage{}, fmt.Errorf("die plist des Dienstes prüfen: %w", fehler)
	}

	ctx, ende := context.WithTimeout(ctx, halterfrist)
	defer ende()

	ausgabe, fehler := drucken(ctx, "gui/"+strconv.Itoa(os.Getuid())+"/"+etikett)

	var beendet *exec.ExitError
	if errors.As(fehler, &beendet) && beendet.ExitCode() == nichtGeladenCode && ctx.Err() == nil {
		return domain.Skriptdienstlage{Eingerichtet: true}, nil
	}

	if fehler != nil {
		return domain.Skriptdienstlage{}, fmt.Errorf("launchctl print: %w", fehler)
	}

	return dienstlageAus(string(ausgabe))
}

// dienstlageBeider fragt beide eigenen Etiketten nach ihrem Dienst — das
// heutige und das von vor 0.9.0 ([altdienstlabel]), dieses nur, wenn seine
// plist von hier stammt.
//
// **Ein beendeter geht vor** (Zweitblick über Phase 4, Befund 7, und sein
// Nachlauf, Befund 8): blieb der Kennungsumzug unvollständig, stehen zwei
// plists da, und eine neue, die nicht geladen ist, verdeckte einen alten
// Dienst, der abstürzt — wie ein gesunder neuer dessen Fehler. Sonst gilt
// der heutige, solange es ihn gibt.
func dienstlageBeider(
	ctx context.Context, heim string, drucken func(context.Context, string) ([]byte, error),
) (domain.Skriptdienstlage, error) {
	neu, neufehler := dienstlageUeber(ctx, filepath.Join(heim, "Library", "LaunchAgents", dienstlabel+".plist"), dienstlabel, drucken)

	alt := altdienst(heim, "")
	if !service.AltlastEigen(alt) {
		return neu, neufehler
	}

	altlage, altfehler := dienstlageUeber(ctx, alt.Datei, altdienstlabel, drucken)
	if altfehler == nil && (altlage.Beendet() || !neu.Eingerichtet && neufehler == nil) {
		return altlage, nil
	}

	return neu, errors.Join(neufehler, altfehler)
}

// betriebsquellen sind die Messungen des Betriebs.
func (a *App) betriebsquellen() betriebsquellen {
	return betriebsquellen{
		platz: func() (uint64, uint64, error) { return platzAuf(filepath.Dir(a.datenbankpfad)) },
		sicherung: func(ctx context.Context) (SicherungsstandDTO, error) {
			if a.db == nil {
				return SicherungsstandDTO{}, errors.New("betriebsband: der Bestand ist nicht offen")
			}

			return sicherungsstandUeber(ctx, a.bestandsweg(), a.einstellungen, time.Now())
		},
		dienst: func(ctx context.Context) (domain.Skriptdienstlage, error) {
			heim, fehler := a.schnittstellen.heim()
			if fehler != nil {
				return domain.Skriptdienstlage{}, fmt.Errorf("betriebsband: %w", fehler)
			}

			return dienstlageBeider(ctx, heim, launchctlDrucken)
		},
		neustarts: func() int { return int(a.betrieb.webneustarts.Load()) },
	}
}

// Betriebsband gibt die Hinweise des Betriebs.
func (a *App) Betriebsband(ctx context.Context) BetriebsbandAntwort {
	ctx = a.imProfil(ctx)

	return a.betrieb.abrufen(ctx, a.betriebsquellen())
}

// Betriebsband ist die Fensterbindung.
func (b *Bindungen) Betriebsband(ctx context.Context) BetriebsbandAntwort {
	return b.app(ctx).Betriebsband(b.ruf(ctx))
}

// WebprozessNeuGestartet zählt einen Webprozess, den WebKit neu starten
// musste — die Hülle ruft es, wenn er endete (`events.Mac.
// WebViewWebContentProcessDidTerminate`), und Wails lädt die Seite danach neu.
func (a *App) WebprozessNeuGestartet(ctx context.Context) {
	ctx = a.imProfil(ctx)
	anzahl := a.betrieb.webneustarts.Add(1)

	slog.WarnContext(ctx, "der Webprozess endete und wurde neu geladen", "anzahl", anzahl)
}
