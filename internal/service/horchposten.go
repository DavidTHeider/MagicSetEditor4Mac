package service

// **Der Horchposten** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Warum die Schleife hier steht und nicht im Transport
//
// **Weil sie eine Entscheidung über Sitzungen ist und nicht über Bytes.** Was
// bei einem krummen Gegenüber geschieht, wie viele Handschläge nebeneinander
// laufen dürfen und wann aufgehört wird — das sind fachliche Fragen. Ein
// Blatt, das sie beantwortete, hätte dieselbe Regel zweimal, sobald es einen
// zweiten Transport gibt.
//
// # Und warum er nicht von selbst anfängt
//
// **Weil das Issue es verlangt**: „nur nach bewusster Auswahl … sichtbar
// werden; keine unbemerkte Dauerwerbung aller Profile." Wer nicht horcht, ist
// nicht erreichbar, und das ist der Ausgangszustand.

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Grenzen eines Horchpostens — Befund Z3 des Zweitblicks über Phase 6.7
// (NET-Teil).
//
// **Ein Platz je Handschlag**, und für eine **Paarung** bis zu ihrem Ende:
// sie ist die Sitzung, die jeder öffnen darf, der nicht gesperrt ist
// ([domain.Peerkontakt.Darf]), und damit die, die ein Fremder in Menge
// öffnen könnte. Wer über der Grenze kommt, wird gleich geschlossen.
//
// **Eine Frist je Handschlag**: wer nach zehn Sekunden nichts gesagt hat,
// sagt nichts mehr. Ohne sie hielte jedes schweigende Gegenüber einen Platz
// für immer.
const (
	HoechstensHandschlaege = 32
	Handschlagfrist        = 10 * time.Second

	// Annehmer ist, wie viele Verbindungen **zugleich** angenommen und
	// gesichert werden — das Kontingent vor TLS (SEC-04,
	// #2139).
	// Ein Annehmer, an dem ein Schweiger hängt, hält nur sich auf; erst wer
	// alle acht zugleich schweigen lässt, hält den Posten bis zur Frist des
	// Transports auf — und das ist eine Last, die er tragen muss.
	Annehmer = 8
)

// ErrZuVieleHandschlaege meldet eine Verbindung über [HoechstensHandschlaege].
var ErrZuVieleHandschlaege = fmt.Errorf(
	"%w: zu viele Handschläge zugleich — die Verbindung wurde geschlossen",
	domain.ErrGedrosselt)

// Hereinkommend ist, was ein Transport zum Horchen liefert.
//
// **Kein Gegenüber und keine Adresse** — dieselbe Zusage wie bei
// [Verbindung]: wer da ist, sagt der Handschlag.
type Hereinkommend interface {
	// Nimm wartet auf die nächste Verbindung.
	//
	// **Die Fehlerklasse ist die Zusage** (§8): `ErrNotFound` heißt „es
	// kommt nichts mehr" und beendet den Posten; alles andere heißt „diese
	// eine Verbindung ist nichts geworden" und wird gemeldet.
	Nimm(ctx context.Context) (Verbindung, error)

	// Adresse ist, wo dieser Lauscher wirklich steht.
	//
	// **Die wirkliche und nicht die gewünschte.** Wer `:0` angibt, bekommt
	// vom Betriebssystem einen Port; ohne diesen Weg müsste ein Aufrufer
	// die gewünschte Adresse weitergeben, und die nennt keinen.
	Adresse() string

	// Schliesse legt den Lauscher nieder.
	Schliesse() error
}

// Horchposten macht aus hereinkommenden Verbindungen Sitzungen.
type Horchposten struct {
	sitzungen *Peersitzungen

	// melden nimmt entgegen, was niemanden sonst erreicht — `nil` heißt:
	// schweigen.
	//
	// **Eine Entscheidung des Aufrufers und kein vergessenes Feld.** Ein
	// Sturz beim Melden wäre schlimmer als ein ungehörter Befund; deshalb
	// wird hier geprüft und nicht vorausgesetzt.
	//
	// **Sie wird aus mehreren Goroutinen gerufen** — jeder Handschlag läuft
	// nebenher, und was er meldet, meldet er von dort. Wer hier etwas
	// sammelt, sammelt es unter einem Schloss.
	melden func(error)

	// aufnahme bekommt jede Sitzung, die entstanden ist — `nil` heißt: sie
	// steht in der Übersicht, und mehr geschieht nicht.
	//
	// # Warum der Posten nicht selbst entscheidet, was folgt
	//
	// **Weil das am Zweck hängt und nicht am Horchen.** Eine Sitzung für
	// eine Partie braucht eine Empfangsschleife, eine für einen Handschlag
	// gar nichts. Wer die Entscheidung hierher legte, machte aus dem Posten
	// eine Stelle, die jede Fachrichtung kennt (§2).
	//
	// **Sie läuft in der Goroutine des Handschlags** — also nebenher, und
	// aus mehreren zugleich. Wer darin etwas sammelt, sammelt es unter einem
	// Schloss; wer darin lange bleibt, hält **diesen** Gast auf und keinen
	// anderen.
	aufnahme func(context.Context, Angenommene)

	// frist ist die je Handschlag — null heißt [Handschlagfrist].
	frist time.Duration
}

// MitHandschlagfrist setzt die Frist je Handschlag — für Läufe.
func (h *Horchposten) MitHandschlagfrist(frist time.Duration) *Horchposten {
	h.frist = frist

	return h
}

// NeuerHorchposten baut ihn über einer Sitzungsverwaltung.
func NeuerHorchposten(
	sitzungen *Peersitzungen, melden func(error),
) *Horchposten {
	return &Horchposten{sitzungen: sitzungen, melden: melden}
}

// MitAufnahme setzt, was mit einer entstandenen Sitzung geschieht.
func (h *Horchposten) MitAufnahme(
	aufnahme func(context.Context, Angenommene),
) *Horchposten {
	h.aufnahme = aufnahme

	return h
}

// Bediene nimmt Verbindungen entgegen, bis nichts mehr kommt.
//
// # Warum jeder Handschlag nebenher läuft
//
// **Weil er über das Netz geht.** Er braucht drei Wege hin und her, und ein
// Gegenüber, das dabei trödelt, hielte sonst jeden Nächsten auf — eine Sperre,
// die jeder von außen auslösen kann.
//
// # Und warum der Abbruch den Lauscher niederlegt
//
// **Weil `Nimm` sonst nie zurückkäme.** Ein `Accept` kennt keinen Kontext: es
// kehrt zurück, wenn eine Verbindung kommt — oder wenn der Lauscher
// geschlossen wird, und sonst nie. Wer ihn erst **nach** der Schleife
// niederlegte, hätte einen Verklemmer: der Abbruch kommt an, die Schleife
// steht in `Accept`, und wer auf ihr Ende wartet, wartet für immer.
//
// **Deshalb schließt der Abbruch, und der Rückweg schließt noch einmal.** Das
// eine bricht das Warten, das andere gilt dem Weg ohne Abbruch; ein `sync.Once`
// macht daraus einen einzigen Ruf, damit ein Lauscher, der beim zweiten Mal
// klagte, nicht zweimal gemeldet wird.
func (h *Horchposten) Bediene(ctx context.Context, herein Hereinkommend) {
	var einmal sync.Once

	schliessen := func() {
		einmal.Do(func() {
			if fehler := herein.Schliesse(); fehler != nil {
				h.melde(fehler)
			}
		})
	}

	// **Die Reihenfolge der drei ist tragend und nicht zufällig.**
	// Aufgeschobenes läuft rückwärts: erst wird auf die laufenden
	// Handschläge gewartet, dann der Wächter über den Abbruch abbestellt,
	// dann der Lauscher niedergelegt. Wer sie vertauschte, verließe diesen
	// Ruf, während noch Sitzungen entstehen — und ein Aufrufer, der danach
	// die Übersicht liest, läse einen Stand, der sich noch ändert.
	defer schliessen()
	defer context.AfterFunc(ctx, schliessen)()

	var laeuft sync.WaitGroup

	defer laeuft.Wait()

	plaetze := make(chan struct{}, HoechstensHandschlaege)

	// **Mehrere nehmen zugleich an** (SEC-04): `Nimm` führt beim Transport
	// den TLS-Handschlag, und der Reihe nach hielt ein Schweiger jeden
	// folgenden bis zur Frist auf. `ende` endet, sobald einer hört, dass
	// nichts mehr kommt — und legt dann den Lauscher nieder, damit die
	// übrigen aus ihrem `Accept` zurückkehren. **Die Handschläge bekommen
	// `ctx` und nicht `ende`**: was angenommen ist, wird zu Ende geführt.
	ende, beenden := context.WithCancel(ctx)
	defer beenden()
	defer context.AfterFunc(ende, schliessen)()

	var (
		annehmer          sync.WaitGroup
		zusammenbruch     sync.Once
		zusammengebrochen any
	)

	for range Annehmer {
		annehmer.Add(1)

		go func() {
			defer annehmer.Done()

			// **Ein Zusammenbruch im Annehmer gehört dem, der bedient**
			// (Vermerk 0171): bis SEC-04 lief `Nimm` in dessen Goroutine, und
			// dort fängt ihn der Starter. Hier wird er festgehalten, der Posten
			// beendet und der Zusammenbruch danach dort weitergeworfen.
			defer func() {
				if wert := recover(); wert != nil {
					zusammenbruch.Do(func() { zusammengebrochen = wert })
					beenden()
				}
			}()

			h.nimmAn(ctx, ende, beenden, herein, plaetze, &laeuft)
		}()
	}

	annehmer.Wait()

	if zusammengebrochen != nil {
		panic(zusammengebrochen)
	}
}

// nimmAn ist ein Annehmer: er wartet auf die nächste Verbindung und gibt sie
// in einen Handschlag, bis nichts mehr kommt.
func (h *Horchposten) nimmAn(
	ctx, ende context.Context, beenden context.CancelFunc, herein Hereinkommend,
	plaetze chan struct{}, laeuft *sync.WaitGroup,
) {
	for {
		if ende.Err() != nil {
			return
		}

		verbindung, fehler := herein.Nimm(ende)

		switch {
		case errors.Is(fehler, domain.ErrNotFound):
			beenden()

			return
		case fehler != nil:
			h.melde(fehler)

			continue
		}

		select {
		case plaetze <- struct{}{}:
		default:
			_ = verbindung.Schliesse() //nolint:errcheck // abgewiesen ist abgewiesen
			h.melde(ErrZuVieleHandschlaege)

			continue
		}

		laeuft.Add(1)

		go func() {
			defer laeuft.Done()

			h.bedieneEine(ctx, verbindung, plaetze)
		}()
	}
}

// bedieneEine führt einen Handschlag in seiner Frist und reicht die Sitzung
// an die Aufnahme — den Platz gibt sie nach dem Handschlag frei, bei einer
// Paarung erst nach der Aufnahme.
//
// # Und ein Zusammenbruch gibt frei, was der Handschlag hielt
//
// **BK-05a** (#2111):
// eine Panik im Handschlag oder in der Aufnahme fing niemand, und der Prozess
// endete. Jetzt wird sie gemeldet, und **frei wird, was diesem Handschlag
// gehört** — der Platz, und je nachdem, wie weit er kam, die Verbindung oder
// die Sitzung. Eine Aufnahme, die zusammenbrach, ist keine: ihre Sitzung
// bliebe sonst in der Übersicht stehen, ohne dass jemand von ihr liest.
func (h *Horchposten) bedieneEine(
	ctx context.Context, verbindung Verbindung, plaetze chan struct{},
) {
	var einmal sync.Once

	frei := func() { einmal.Do(func() { <-plaetze }) }

	var griff *Sitzungsgriff

	// **Der Platz zuerst, die Meldung danach** — aufgeschoben wird rückwärts,
	// und wer die Meldung liest, soll den Platz schon wieder vorfinden.
	defer domain.Auffangen(func(zusammenbruch error) {
		if griff != nil {
			_ = griff.Schliesse() //nolint:errcheck // der Zusammenbruch ist die Meldung
		} else {
			_ = verbindung.Schliesse() //nolint:errcheck // der Zusammenbruch ist die Meldung
		}

		h.melde(zusammenbruch)
	})
	defer frei()

	frist := h.frist
	if frist == 0 {
		frist = Handschlagfrist
	}

	handschlag, abbrechen := context.WithTimeout(ctx, frist)
	defer abbrechen()

	angenommen, fehler := h.sitzungen.Annehmen(handschlag, verbindung)

	abbrechen()

	if fehler != nil {
		frei()
		h.melde(fehler)

		return
	}

	griff = angenommen

	if griff.Stand().Zweck != domain.ZweckPaarung {
		frei()
	}

	if h.aufnahme != nil {
		h.aufnahme(ctx, griff)
	}
}

// melde gibt einen Befund weiter, wenn jemand zuhört.
func (h *Horchposten) melde(fehler error) {
	if h.melden != nil {
		h.melden(fehler)
	}
}
