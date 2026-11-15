package app

// **Der freiwillige Plan läuft im Fenster** — OPS01
// (#2347).
//
// # Wann
//
// **Kurz nach dem Start, dann stündlich gefragt, und gleich nach einer
// Ordnerwahl** — gesichert wird nur, wenn der Plan fällig ist
// (`domain.SicherungsplanFaellig`, einmal die Woche). Ohne Ordner tut er
// nichts: der Plan ist freiwillig.
//
// # Wie lange
//
// **So lange, wie ein Bestand offen ist.** Der Schluss und jeder Tausch des
// Bestands beenden ihn über [App.eigeneLaeufeBeenden] — und warten auf ihn —,
// und [App.starten] startet ihn über dem neuen Bestand wieder. Ein Lauf, der
// gerade sichert, wird abgebrochen; seine Nachbardateien räumt er selbst. So
// landet nie ein Vermerk in einem Bestand, den die Sicherung nicht enthält
// (§7, Regel 2).
//
// # Was er nicht kann
//
// **Sichern, solange das Fenster zu ist.** Der Dienst unter launchd führt den
// Plan nicht aus; wer ihn dort braucht, sichert über `mse bestand-sichern`
// aus einem eigenen Zeitplan. Die Sperre neben dem Bestand hält auch dann
// zwei Pläne auseinander.

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Planendung hängt an den Pfad des Bestands — die Sperre des Plans.
const Planendung = ".plan"

// planverzug und plantakt sind, wann der Plan nach dem Start zuerst und
// danach wieder fragt. **Nach dem Start erst nach zwei Minuten**: der Start
// ist gemessen (≤ 2 s, `LEISTUNGSSCHNITTE.md`), und eine Sicherung gleich
// dahinter träfe die ersten Schritte am Fenster.
var (
	planverzug = 2 * time.Minute
	plantakt   = time.Hour
)

// sicherungsplanlauf hält den laufenden Plan.
type sicherungsplanlauf struct {
	mu        sync.Mutex
	aufhoeren context.CancelFunc
	fertig    chan struct{}
	anstoss   chan struct{}

	// fehler ist, woran der letzte Lauf scheiterte — leer nach einem
	// geglückten oder einem, der nichts zu tun hatte.
	fehler error
}

// sicherungsplanStarten startet den Plan über dem offenen Bestand.
func (a *App) sicherungsplanStarten() {
	plan := service.NeuerSicherungsplan(a.bestandsweg(),
		func(ctx context.Context) (string, error) { return planordnerAus(a.imProfil(ctx), a.einstellungen) },
		plansperre(a.datenbankpfad), domain.FlaecheFenster)

	ctx, aufhoeren := context.WithCancel(context.Background())
	fertig, anstoss := make(chan struct{}), make(chan struct{}, 1)

	a.planlauf.mu.Lock()
	a.planlauf.aufhoeren, a.planlauf.fertig, a.planlauf.anstoss = aufhoeren, fertig, anstoss
	a.planlauf.mu.Unlock()

	go func() {
		defer close(fertig)
		defer domain.Auffangen(a.planfehlerMerken)

		uhr := time.NewTimer(planverzug)
		defer uhr.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-uhr.C:
			case <-anstoss:
			}

			lauf, fehler := plan.Laufen(ctx)
			if lauf.Gesichert != "" {
				slog.InfoContext(ctx, "nach Plan gesichert", "datei", lauf.Gesichert, "anzahl", len(lauf.Entfernt))
			}

			a.planfehlerMerken(fehler)
			uhr.Reset(plantakt)
		}
	}()
}

// sicherungsplanAnstossen fragt den Plan gleich — nach einer Ordnerwahl.
func (a *App) sicherungsplanAnstossen() {
	a.planlauf.mu.Lock()
	defer a.planlauf.mu.Unlock()

	// **Ein Anstoß genügt**: liegt schon einer, wartet der Plan ohnehin. Unter
	// dem Schloss sendet nur einer, und der Platz ist da, wenn die Länge null
	// ist — das Senden wartet nie.
	if a.planlauf.anstoss != nil && len(a.planlauf.anstoss) == 0 {
		a.planlauf.anstoss <- struct{}{}
	}
}

// sicherungsplanBeenden beendet den Plan und wartet auf ihn.
func (a *App) sicherungsplanBeenden() {
	a.planlauf.mu.Lock()
	aufhoeren, fertig := a.planlauf.aufhoeren, a.planlauf.fertig
	a.planlauf.aufhoeren, a.planlauf.fertig, a.planlauf.anstoss = nil, nil, nil
	a.planlauf.mu.Unlock()

	if aufhoeren == nil {
		return
	}

	aufhoeren()
	<-fertig
}

// planfehlerMerken hält fest, woran der Plan zuletzt scheiterte.
func (a *App) planfehlerMerken(fehler error) {
	a.planlauf.mu.Lock()
	defer a.planlauf.mu.Unlock()

	a.planlauf.fehler = fehler
}

// sicherungsplanfehler übersetzt den letzten Fehler des Plans — `nil` ohne.
func (a *App) sicherungsplanfehler(ctx context.Context) *Fehler {
	a.planlauf.mu.Lock()
	fehler := a.planlauf.fehler
	a.planlauf.mu.Unlock()

	if fehler == nil {
		return nil
	}

	return Uebersetze(ctx, fehler)
}

// plansperre nimmt die Sperre des Plans neben dem Bestand, **ohne zu warten**
// — zwei Prozesse über demselben Bestand sichern nicht beide nach Plan.
//
// **Ein belegter Griff ist die Antwort und kein Fehlschlag** — dieselbe
// Lesart wie bei der Sperre des Suchindex (`retrievalsperre.go`).
func plansperre(bestandspfad string) func() (func(), bool, error) {
	return func() (func(), bool, error) {
		griff, fehler := dateisperre.Oeffnen(bestandspfad + Planendung)
		if fehler != nil {
			return nil, false, fehler
		}

		// **Ein Erbe ist kein zweiter Plan** (#2401): ein Kind, das dieser
		// Prozess gerade startet, hält die Sperre bis zu seinem `exec`.
		if belegt := griff.AlleinOhneErbe(context.Background()); belegt != nil {
			_ = griff.Schliessen() //nolint:errcheck // der Griff ist ohnehin unbrauchbar

			return nil, true, nil //nolint:nilerr // siehe oben
		}

		return func() { _ = griff.Schliessen() }, false, nil //nolint:errcheck // die Sperre fällt mit dem Griff
	}
}
