// **Ein Wechsel hält an, was währenddessen arbeitet** — Befund 1 des
// Zweitblicks über die Phasen 4.0 bis 4.13 (2026-09-07).
//
// # Der Befund
//
// `ProfilWechseln` legte die neue Kennung in `profil` ab und baute **danach**
// die Repositories auf die rollenrichtige Verbindung um. Wails-Bindungen
// laufen ausdrücklich parallel; eine, die dazwischen lief, nahm das **neue**
// Profil und die **alte** Verbindung.
//
// **Sichtbar wäre das als Schreiben in einem Leserprofil** — für die Dauer
// eines Verbindungsaufbaus.
package app

import (
	"sync"
	"testing"
	"time"
)

// **Solange umgeschaltet wird, kommt niemand durch `imProfil`.**
//
// # Wie der Fall hergestellt wird
//
// **Über die Sperre selbst und nicht über einen echten Wechsel**: `Wechseln`
// braucht einen Bestand, einen zweiten Faktor und ein Ziel — und geprüft wird
// hier nicht der Wechsel, sondern dass er **anhält**. Ein Lauf, der den ganzen
// Weg führe, prüfte fünf Dinge und sagte über dieses eine nichts.
func TestWaehrendDerUmschaltungKommtNiemandDurch(t *testing.T) {
	t.Parallel()

	anwendung := &App{}
	anwendung.profil.Store(1)

	anwendung.umschaltung.Lock()

	durch := make(chan struct{})

	go func() {
		_ = anwendung.imProfil(t.Context())
		close(durch)
	}()

	// **Er darf nicht durchkommen**, solange die Sperre steht.
	select {
	case <-durch:
		t.Fatal("eine Bindung kam durch, während umgeschaltet wurde")
	case <-time.After(50 * time.Millisecond):
	}

	anwendung.umschaltung.Unlock()

	select {
	case <-durch:
	case <-time.After(2 * time.Second):
		t.Fatal("die Bindung kam auch nach der Umschaltung nicht durch")
	}
}

// **Und der Wechsel selbst klemmt nicht gegen sich.**
//
// **Ein `RWMutex` ist nicht wiedereintrittsfähig**: wer ihn schreibend hält
// und dann `imProfil` ruft, wartet auf sich selbst. Deshalb gibt es
// [App.imProfilOhneSperre] — und deshalb gibt es diesen Lauf.
func TestDerWechselKlemmtNichtGegenSichSelbst(t *testing.T) {
	t.Parallel()

	anwendung := &App{}
	anwendung.profil.Store(7)

	fertig := make(chan struct{})

	go func() {
		anwendung.umschaltung.Lock()
		_ = anwendung.imProfilOhneSperre(t.Context())
		anwendung.umschaltung.Unlock()
		close(fertig)
	}()

	select {
	case <-fertig:
	case <-time.After(2 * time.Second):
		t.Fatal("der Wechsel wartet auf sich selbst")
	}
}

// **Und zwei Wechsel zugleich laufen nacheinander.**
//
// **Das ist die zweite Hälfte des Befunds**: `useVorgang` im Fenster hält je
// Komponenteninstanz ein eigenes Set, und Nutzerwahl und Profilwahl sind zwei
// Instanzen — sie schließen einander nicht aus. REST und MCP bieten dieselben
// Wege ohnehin ohne gemeinsame Sperre an. **Die Serialisierung gehört deshalb
// hierher**, wo alle vier Flächen durchkommen.
func TestZweiUmschaltungenLaufenNacheinander(t *testing.T) {
	t.Parallel()

	anwendung := &App{}

	var (
		zugleich sync.WaitGroup
		hoechste int
		jetzt    int
		wache    sync.Mutex
	)

	for range 8 {
		zugleich.Add(1)

		go func() {
			defer zugleich.Done()

			anwendung.umschaltung.Lock()
			defer anwendung.umschaltung.Unlock()

			wache.Lock()
			jetzt++

			if jetzt > hoechste {
				hoechste = jetzt
			}

			wache.Unlock()

			time.Sleep(time.Millisecond)

			wache.Lock()
			jetzt--
			wache.Unlock()
		}()
	}

	zugleich.Wait()

	if hoechste != 1 {
		t.Errorf("%d Umschaltungen liefen zugleich — erwartet war eine", hoechste)
	}
}
