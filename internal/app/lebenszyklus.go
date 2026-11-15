package app

// **Wie die Anwendung endet** — abbrechen und beenden.
//
// Eine eigene Datei, weil `app.go` mit ihr über 400 Zeilen ginge (§2). Der
// Schnitt folgt der Frage: dort wird zusammengesetzt, hier abgeräumt.

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Schlussfrist ist, wie lange der Schluss **insgesamt** dauern darf — BK-02
// (#2110):
// das Merken, das Beenden der eigenen Läufe und das Warten auf sie. Danach
// geht der Bestand trotzdem zu, und was noch lief, findet ihn geschlossen.
const Schlussfrist = 5 * time.Second

// Bestenfallsfrist ist der Teil der [Schlussfrist], den das Merken bekommt —
// **höchstens**. Gemessen an der Hüllenprobe `bk07a` (2026-10-06): mit der
// ganzen Frist nahm ein Merken, das an einem gesperrten Bestand hing, den
// eigenen Läufen ihre Zeit, und keiner wurde mehr beendet.
const Bestenfallsfrist = time.Second

// ErrSchlussNachFrist sagt, dass der Schluss nicht alles in seiner Frist
// geschafft hat — und was dabei verloren ging, steht dahinter.
var ErrSchlussNachFrist = errors.New("der Schluss kam nicht in seiner Frist zu Ende")

// errAnwendungEndet weist ab, was nach dem Beginn des Schlusses kommt — als
// Abbruch, denn genau das ist es.
var errAnwendungEndet = fmt.Errorf("%w: die Anwendung endet", context.Canceled)

// errBestandWirdGetauscht weist ab, was während eines Tauschs des Bestands
// kommt — A02 (#2216).
//
// **Belegt und kein Abbruch**: der Zustand vergeht von allein, und wer es
// gleich noch einmal versucht, kommt durch.
var errBestandWirdGetauscht = fmt.Errorf("%w: der Bestand wird gerade getauscht",
	domain.ErrBelegt)

// Die drei Lagen des Tors.
//
// **Drei und nicht zwei** — A02: bis dahin gab es nur „begonnen", und der
// Tausch des Bestands nahm denselben Schluss wie das Ende des Programms.
// Das Tor ging zu und **nie wieder auf**, denn `Starten` wusste nicht, dass
// es nach einem Tausch lief und nicht nach dem Ende.
const (
	// torOffen lässt jede gebundene Methode durch.
	torOffen int32 = iota

	// torTausch hält an, solange der Bestand getauscht wird — und geht
	// danach wieder auf.
	torTausch

	// torEndet hält an, weil das Programm endet. **Von hier führt kein Weg
	// zurück**, auch nicht das Ende eines Tauschs, der vorher begann.
	torEndet
)

// schlusslage hält, was der Schluss braucht — BK-07a
// (#2115).
type schlusslage struct {
	// tor ist eine der drei Lagen [torOffen], [torTausch], [torEndet].
	tor atomic.Int32

	// wechsel hält einen Tausch des Bestands — den ganzen, vom Schließen bis
	// zum letzten Öffnen ([App.imTausch]) — und das Ende auseinander. Befund
	// des Zweitblicks über Phase A: überlappten Wiederöffnen und Ende, ging
	// der Bestand nach dem Schluss wieder auf und blieb offen; Restbefund des
	// Nachlaufs: überlappten Schließen zum Tausch und Ende, räumten beide ab.
	wechsel sync.Mutex

	// epoche zählt die Tausche des Bestands (J05, #2269, zu #2292). Eine
	// Meldung des Doppelklicks trägt die ihre, die Antworten von Neuanfang und
	// Zurückspielen die neue — und das Fenster verwirft eine Kennung aus
	// einem Bestand, den es nicht mehr zeigt. **Die Sperre oben ordnet Go,
	// nicht das Fenster**: Wails stellt Ereignisse in eine eigene Schlange.
	epoche atomic.Uint64

	// kinder zählt die Läufe, die die Anwendung selbst startet und die über
	// den Bestand gehen — Tisch des Gastes, Sitzung eines Gastes beim Wirt.
	// **Gezählt wird beim Start**, und kein neuer kommt nach dem Beginn des
	// Schlusses: das Tor weist den Beitritt ab, und die Gäste kommen über
	// den Horchposten, der vorher endet.
	kinder sync.WaitGroup

	// startmarke ist die Marke dieser Sitzung (F04, #2343) — `ende` nimmt
	// sie beim sauberen Schluss weg, und nur dort.
	startmarke sitzungsbeginn
}

// Beenden beendet die Anwendung — in der [Schlussfrist].
func (a *App) Beenden() error {
	ctx, aufhoeren := context.WithTimeout(context.Background(), Schlussfrist)
	defer aufhoeren()

	return a.Schliessen(ctx, nil)
}

// Schliessen beendet die Anwendung in der Frist von `frist` — **in dieser
// Reihenfolge**, und alle zusammen in dieser Frist:
//
//  1. Das Tor geht zu: keine neue Arbeit mehr aus dem Fenster.
//  2. `bestenfalls` — was beim Schließen noch gemerkt werden soll, etwa die
//     Fenstergröße — in höchstens [Bestenfallsfrist]. Ein Verlust hier
//     kostet Komfort und sonst nichts.
//  3. Die eigenen Läufe enden: der Tisch als Gast, Horchen, Fernhorchen, die
//     Peersitzungen (und mit ihnen die Sitzungen der Gäste beim Wirt), die
//     Dienstposten, das Gerätemodell, das KI-Auftragswerk.
//  4. Gewartet wird, bis Tisch und Gäste zu Ende sind.
//  5. **Zuletzt** der Bestand.
//
// **Läuft die Frist ab, geht der Bestand trotzdem zu** (BK-02, #2110): ein
// Schluss, der an einer Sperre hängen bleibt, hält das Fenster offen, das
// jemand schließen wollte. Was dann noch läuft, bekommt **keine neue**
// Verbindung mehr — `sql: database is closed`. **Eine Transaktion, die schon
// offen ist, kann aber noch zu Ende schreiben** (`database/sql` entzieht eine
// ausgeliehene Verbindung nicht; gemessen vom Zweitblick über A und B,
// Befund 3): ganz oder gar nicht, wie jede Transaktion. Der Rückgabewert
// nennt, was in der Frist nicht fertig wurde.
//
// **Zweimal schließen ist keines**: der zweite Schluss findet nichts mehr vor.
//
// **Dies ist das Ende des Programms**, und das Tor geht dabei für immer zu.
// Wer den Bestand nur tauschen will, nimmt [App.imTausch] — bis
// A02 nahm der Tausch diesen Weg, und das Fenster blieb danach gesperrt.
func (a *App) Schliessen(frist context.Context, bestenfalls func(context.Context)) error {
	a.lauf.tor.Store(torEndet)

	// **Ein Tausch, der gerade läuft, wird abgewartet** — in der Frist, und
	// danach räumt der Schluss auch den eben geöffneten Bestand ab.
	//
	// **Kommt es in der Frist nicht zu Ende, fasst der Schluss den Bestand
	// nicht an** (Befund 1 des zweiten Durchgangs des Zweitblicks über Phase
	// A): schlösse er ohne die Sperre, liefen Schließen und Verdrahten
	// gleichzeitig. Das Wiederöffnen sieht das Ende, wenn es fertig ist, und
	// schließt dann selbst ([App.wiederOeffnen]).
	if !sperreMitKontext(frist, a.lauf.wechsel.TryLock, a.lauf.wechsel.Lock, a.lauf.wechsel.Unlock) {
		slog.WarnContext(frist, "der Schluss kam nicht in seiner Frist zu Ende",
			"grund", "ein Tausch des Bestands läuft noch und schließt selbst")

		return fmt.Errorf("%w: ein Tausch des Bestands", ErrSchlussNachFrist)
	}

	defer a.lauf.wechsel.Unlock()

	fehler := a.abraeumen(frist, bestenfalls)

	// **Die Marke zuletzt, und nur hier** (F04, #2343): erst wenn der
	// Bestand zu ist, war der Schluss sauber. Ein Schluss nach der Frist ist
	// trotzdem einer — geordnet und mit benanntem Verlust, kein Abbruch.
	//
	// **Nicht in [App.abraeumen]**: dort stand sie bis zum Zweitblick über
	// Phase 1 von 0.9.0 (Befund 1), und jeder Tausch des Bestands —
	// Zurückspielen, Verschlüsseln, Neuanfang, ein abgebrochener Start —
	// nahm sie mit. Die Sitzung lief weiter, und ein späterer Absturz blieb
	// ohne Hinweis. Die Marke gehört dem Prozess, nicht dem Bestand.
	return errors.Join(fehler, a.sitzungBeenden())
}

// sitzungBeenden nimmt die Marke der Sitzung weg — **genau einmal**, und nur,
// wer `wechsel` hält: [App.Schliessen], oder [App.wiederOeffnen], wenn ein
// Schluss nicht auf den Tausch warten konnte (Nachlauf über Phase 1 von
// 0.9.0, Befund 1). Dort blieb sie bis dahin liegen, und der nächste Start
// meldete einen Abbruch, den es nicht gab.
func (a *App) sitzungBeenden() error {
	ende := a.lauf.startmarke.ende
	if ende == nil {
		return nil
	}

	a.lauf.startmarke.ende = nil

	return ende()
}

// abraeumen sind die Schritte 2 bis 5 aus [App.Schliessen] — **das Tor
// stellt, wer ruft**.
func (a *App) abraeumen(frist context.Context, bestenfalls func(context.Context)) error {
	var verloren []string

	if bestenfalls != nil {
		knapp, aufhoeren := context.WithTimeout(frist, Bestenfallsfrist)

		if !innerhalb(knapp, func() { bestenfalls(knapp) }) {
			verloren = append(verloren, "was beim Schließen gemerkt werden sollte")
		}

		aufhoeren()
	}

	// **Und welcher Schritt in Panik endete**, steht mit dabei — gelesen nur,
	// wenn das Beenden in der Frist fertig wurde: sonst schreibt es noch.
	var gescheitert []string

	if innerhalb(frist, func() { gescheitert = a.eigeneLaeufeBeenden(frist) }) {
		verloren = append(verloren, gescheitert...)
	} else {
		verloren = append(verloren, "das Beenden der eigenen Läufe")
	}

	if !innerhalb(frist, a.lauf.kinder.Wait) {
		verloren = append(verloren, "das Warten auf Tisch und Gäste")
	}

	fehler := a.bestandSchliessen()

	if len(verloren) > 0 {
		slog.WarnContext(frist, "der Schluss kam nicht in seiner Frist zu Ende",
			"grund", strings.Join(verloren, ", "))

		fehler = errors.Join(fmt.Errorf("%w: %s", ErrSchlussNachFrist, strings.Join(verloren, ", ")), fehler)
	}

	return fehler
}

// innerhalb führt einen Schritt aus und wartet höchstens bis zum Ende von
// `ctx`. **Ein Schritt, der hängen bleibt, läuft weiter** — aber niemand
// wartet mehr auf ihn.
//
// **Und ein Schritt, der in Panik endet, ist auch nicht erledigt** — F02
// (#2341).
// Die Goroutine ist seine; fängt sie die Panik nicht, fängt sie niemand, und
// der Prozess endete mitten im Schluss, bevor der Bestand zu war.
func innerhalb(ctx context.Context, schritt func()) bool {
	fertig := make(chan bool, 1)

	go func() {
		erledigt := false

		defer func() {
			if wert := recover(); wert != nil {
				slog.ErrorContext(ctx, "ein Schritt brach mit einer Panik ab", "panik", fmt.Sprint(wert))
			}

			fertig <- erledigt
		}()

		schritt()

		erledigt = true
	}()

	select {
	case erledigt := <-fertig:
		return erledigt
	case <-ctx.Done():
		return false
	}
}

// Bestenfalls führt einen Schritt aus, dessen Verlust nur Komfort kostet —
// etwa das Merken der Fenstergröße —, und wartet höchstens bis zum Ende von
// `ctx` (BK-02, #2110). **Ein Verlust steht im Protokoll**, mit dem, was
// verloren ging.
func Bestenfalls(ctx context.Context, was string, schritt func()) bool {
	if innerhalb(ctx, schritt) {
		return true
	}

	slog.WarnContext(ctx, "beim Schließen ging etwas verloren", "grund", was)

	return false
}

// bestandSchliessen schließt die Datenbank — der letzte Schritt.
func (a *App) bestandSchliessen() error {
	if a.db == nil {
		return nil
	}

	// **Der Standleser hat nichts mehr zurückzugeben.** Bis zum 2026-09-09
	// hielt er eine geliehene Verbindung, die hier vor `db.Close` zurück
	// musste — seit #1485 fragt er über den Vorrat wie jeder andere, und der
	// Vorrat hat eine Verbindung (`storage.Oeffnen`).
	//
	// **Die zweite Verbindung zuerst** — Werkstück 16 von Phase 4.12. Sie
	// hängt an derselben Datei, und wer sie stehen ließe, hielte den Bestand
	// nach dem Beenden weiter.
	a.leserverbindungSchliessen()

	if err := a.db.Close(); err != nil {
		return fmt.Errorf("Datenbank schließen: %w", err)
	}

	a.db = nil

	return nil
}
