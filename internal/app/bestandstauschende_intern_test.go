package app

// **Der Tausch des Bestands und das Ende des Programms** — A02 und A05
// (#2216,
// #2219).
// Aus `bestandstauschtor_intern_test.go` herausgelöst (§2, Zeilengrenze —
// Restbefund des Nachlaufs über Phase A).
//
// **Nicht parallel**: zwei der Fälle warten Fristen ab.

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"
)

// **Ein echter Schluss mitten im Tausch bleibt ein Schluss.** Wer danach
// wieder öffnete, ließe ein Programm, das gerade endet, mit offenem Tor und
// offenem Bestand zurück.
func TestEinSchlussImTauschBleibtEinSchluss(t *testing.T) {
	a, _ := tauschbereit(t)

	if fehler := a.zumTauschSchliessen(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if fehler := a.Beenden(); fehler != nil {
		t.Fatalf("beenden: %v", fehler)
	}

	if fehler := a.wiederOeffnen(t.Context()); !errors.Is(fehler, context.Canceled) {
		t.Errorf("das Wiederöffnen nach dem Schluss gab %v, erwartet war ein Abbruch", fehler)
	}

	if a.db != nil {
		t.Error("nach dem Schluss wurde der Bestand wieder geöffnet")
	}

	abgewiesen := torAntwort(t, a)
	if abgewiesen == nil || abgewiesen.Schluessel != "anwendungEndet" {
		t.Errorf("nach dem Schluss antwortet das Tor %+v, erwartet war anwendungEndet", abgewiesen)
	}
}

// **Ein Tausch, dessen Schließen die Frist überschreitet, lässt die
// Anwendung nicht ohne Bestand zurück** — Befund des Zweitblicks über
// Phase A: der Schluss schließt den Bestand auch nach der Frist, der
// Bestandsweg brach dann ab, ohne wieder zu öffnen, und das Tor blieb auf
// „Tausch".
//
// **Nicht parallel**: er wartet die [Schlussfrist] ab.
func TestEinTauschUeberDieFristLaesstDieAnwendungNichtOhneBestand(t *testing.T) {
	a, _ := tauschbereit(t)

	// **Ein Gast, der nicht endet** — das Warten auf ihn läuft in die Frist.
	a.lauf.kinder.Add(1)
	t.Cleanup(a.lauf.kinder.Done)

	if _, fehler := a.neuanfangImTausch(t.Context()); !errors.Is(fehler, ErrSchlussNachFrist) {
		t.Fatalf("der Neuanfang gab %v, erwartet war der Schluss nach der Frist", fehler)
	}

	if abgewiesen := torAntwort(t, a); abgewiesen != nil {
		t.Fatalf("nach dem abgebrochenen Tausch weist das Tor ab: %+v", abgewiesen)
	}

	if a.db == nil || a.db.PingContext(t.Context()) != nil {
		t.Error("nach dem abgebrochenen Tausch hat die Anwendung keinen offenen Bestand")
	}
}

// **Ein Ende während des Wiederöffnens wartet den Tausch ab und schließt
// danach** — Befund des Zweitblicks über Phase A: der Test darüber führt
// Schluss und Wiederöffnen nacheinander aus; überlappen sie, ging der
// Bestand nach dem Schluss wieder auf und blieb offen.
func TestEinEndeWaehrendDesWiederoeffnensSchliesstDanach(t *testing.T) {
	a, _ := tauschbereit(t)
	angefangen, weiter := haltendesOeffnen(a)

	getauscht := make(chan error, 1)
	go func() { getauscht <- a.imTausch(schliessenUndOeffnen(a)) }()

	<-angefangen

	beendet := make(chan error, 1)
	go func() { beendet <- a.Beenden() }()

	// **Das Ende kommt, während das Öffnen hängt** — dann erst geht es weiter.
	time.Sleep(200 * time.Millisecond)
	close(weiter)

	// **Das Ende bricht den Tausch ab** — wie in den Läufen darunter; gemessen
	// wird, dass danach geschlossen ist.
	if fehler := <-getauscht; !errors.Is(fehler, context.Canceled) {
		t.Fatalf("der Tausch endete mit %v, erwartet war der Abbruch", fehler)
	}

	if fehler := <-beendet; fehler != nil {
		t.Fatalf("beenden: %v", fehler)
	}

	if a.db != nil {
		t.Error("nach dem Ende ist der Bestand wieder offen")
	}

	if a.lauf.tor.Load() != torEndet {
		t.Errorf("das Tor steht auf %d, erwartet war das Ende", a.lauf.tor.Load())
	}
}

// **Wartet das Ende nicht bis zum Schluss des Tauschs, schließt das
// Wiederöffnen selbst** — Befund 1 des zweiten Durchgangs des Zweitblicks
// über Phase A: die Frist des Endes lief ab, während das Öffnen hing, und
// danach ging der Bestand hinter einem geschlossenen Tor auf und blieb offen.
func TestEinEndeOhneGeduldLaesstDasWiederoeffnenSchliessen(t *testing.T) {
	a, _ := tauschbereit(t)
	angefangen, weiter := haltendesOeffnen(a)

	getauscht := make(chan error, 1)
	go func() { getauscht <- a.imTausch(schliessenUndOeffnen(a)) }()

	<-angefangen

	frist, aufhoeren := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer aufhoeren()

	if fehler := a.Schliessen(frist, nil); !errors.Is(fehler, ErrSchlussNachFrist) {
		t.Errorf("das Ende gab %v, erwartet war der Schluss nach der Frist", fehler)
	}

	close(weiter)

	if fehler := <-getauscht; !errors.Is(fehler, context.Canceled) {
		t.Errorf("der Tausch gab %v, erwartet war der Abbruch durch das Ende", fehler)
	}

	if a.db != nil {
		t.Error("nach dem Ende ist der Bestand wieder offen")
	}
}

// **Ein Ende während des Schließens zum Tausch räumt nicht mit ab** —
// Restbefund des Nachlaufs über Phase A: das Schließen zum Tausch nahm
// `wechsel` nicht, und ein Ende in dieser Zeit räumte gleichzeitig ab —
// beide lasen und setzten `a.db`. **Der ganze Rahmen hält die Sperre**; das
// Ende beginnt erst, wenn der Tausch durch ist.
func TestEinEndeWaehrendDesSchliessensZumTauschRaeumtNichtMitAb(t *testing.T) {
	a, _ := tauschbereit(t)

	// **Ein Gast, der erst auf ein Zeichen endet** — das Schließen zum Tausch
	// wartet auf ihn.
	a.lauf.kinder.Add(1)

	drin := make(chan struct{})
	getauscht := make(chan error, 1)

	go func() {
		getauscht <- a.imTausch(func() error {
			close(drin)

			return schliessenUndOeffnen(a)()
		})
	}()

	<-drin

	bestenfalls := make(chan struct{})
	frist, aufhoeren := context.WithTimeout(context.Background(), 10*time.Second)
	defer aufhoeren()

	beendet := make(chan error, 1)
	go func() { beendet <- a.Schliessen(frist, func(context.Context) { close(bestenfalls) }) }()

	select {
	case <-bestenfalls:
		t.Error("das Ende räumte ab, während der Tausch noch schloss")
	case <-time.After(300 * time.Millisecond):
	}

	a.lauf.kinder.Done()

	if fehler := <-getauscht; !errors.Is(fehler, context.Canceled) {
		t.Errorf("der Tausch gab %v, erwartet war der Abbruch durch das Ende", fehler)
	}

	if fehler := <-beendet; fehler != nil {
		t.Errorf("beenden: %v", fehler)
	}

	if a.db != nil {
		t.Error("nach dem Ende ist der Bestand offen")
	}
}

// **Ein Tausch, der die Sperre erst nach dem Ende bekommt, fasst nichts
// an** — sonst tauschte er Dateien unter einem Programm, das gerade endet.
func TestEinTauschNachDemEndeFasstNichtsAn(t *testing.T) {
	a, _ := tauschbereit(t)

	a.lauf.wechsel.Lock()

	freigeben := sync.OnceFunc(a.lauf.wechsel.Unlock)
	t.Cleanup(freigeben)

	angefasst := false
	getauscht := make(chan error, 1)

	go func() {
		getauscht <- a.imTausch(func() error {
			angefasst = true

			return nil
		})
	}()

	for ende := time.Now().Add(5 * time.Second); a.lauf.tor.Load() != torTausch; time.Sleep(time.Millisecond) {
		if time.Now().After(ende) {
			t.Fatal("der Tausch stellte das Tor nicht um")
		}
	}

	a.lauf.tor.Store(torEndet)
	freigeben()

	if fehler := <-getauscht; !errors.Is(fehler, context.Canceled) {
		t.Errorf("der Tausch gab %v, erwartet war der Abbruch durch das Ende", fehler)
	}

	if angefasst {
		t.Error("der Tausch lief nach dem Ende noch")
	}
}

// **Nach dem Ende wird nicht mehr getauscht** — und die Antwort sagt, dass
// das Programm endet, nicht dass gerade getauscht wird.
func TestNachDemEndeWirdNichtMehrGetauscht(t *testing.T) {
	a, _ := tauschbereit(t)

	if fehler := a.Beenden(); fehler != nil {
		t.Fatalf("beenden: %v", fehler)
	}

	if fehler := a.imTausch(func() error { return nil }); !errors.Is(fehler, errAnwendungEndet) {
		t.Errorf("der Tausch nach dem Ende gab %v, erwartet war das Ende", fehler)
	}
}

// haltendesOeffnen lässt das nächste Öffnen hängen, bis `weiter` schließt.
func haltendesOeffnen(a *App) (angefangen, weiter chan struct{}) {
	angefangen, weiter = make(chan struct{}), make(chan struct{})
	echt := a.oeffnen
	a.oeffnen = func(ctx context.Context, pfad string) (*sql.DB, error) {
		close(angefangen)
		<-weiter

		return echt(ctx, pfad)
	}

	return angefangen, weiter
}

// schliessenUndOeffnen ist der Kern eines Tauschs ohne den Tausch.
func schliessenUndOeffnen(a *App) func() error {
	return func() error {
		if fehler := a.zumTauschSchliessen(); fehler != nil {
			return fehler
		}

		return a.wiederOeffnen(context.Background())
	}
}
