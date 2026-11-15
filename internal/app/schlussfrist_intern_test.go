package app

// **Der Schluss hat eine Gesamtfrist** — BK-02.AK-01
// (#2110).
//
// Beim Schließen merkt sich das Fenster seine Größe (`FenstergroesseMerken`,
// aus dem Haken der Hülle). Das geht über die Profilnaht — eine
// `sync.RWMutex`, die keinen Kontext kennt — und über den Pool mit **einer**
// Verbindung. Hält jemand beides, hielt das Merken das Fenster offen, und
// eine Frist um das SQL allein half nicht: es kam nie bis dorthin.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestDerSchlussKehrtTrotzBesetztemPoolUndProfilschlossZurueck(t *testing.T) {
	t.Parallel()

	anwendung := gestartet(t)

	griff, fehler := anwendung.db.Conn(t.Context())
	if fehler != nil {
		t.Fatal(fehler)
	}

	anwendung.umschaltung.Lock()

	freigegeben := false
	freigeben := func() {
		if !freigegeben {
			freigegeben = true

			anwendung.umschaltung.Unlock()
			_ = griff.Close() //nolint:errcheck // Rückgabe an den Pool
		}
	}
	t.Cleanup(freigeben)

	// **Erst das Hindernis belegen**: das Merken allein, ohne Frist, kommt
	// nicht zurück — so lief es bis BK-02 im Haken der Hülle.
	gemerkt := make(chan struct{})

	go func() {
		anwendung.FenstergroesseMerken(context.Background(), 1200, 800)
		close(gemerkt)
	}()

	select {
	case <-gemerkt:
		t.Fatal("das Merken kam an Pool und Profilschloss vorbei — der Lauf prüft nichts")
	case <-time.After(300 * time.Millisecond):
	}

	frist, aufhoeren := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer aufhoeren()

	angefangen := time.Now()

	fehler = anwendung.Schliessen(frist, func(ctx context.Context) {
		anwendung.FenstergroesseMerken(ctx, 1200, 800)
	})

	if dauer := time.Since(angefangen); dauer > 2*time.Second {
		t.Errorf("der Schluss brauchte %v bei einer Frist von 500 ms", dauer)
	}

	// **Der Verlust ist benannt** — als Bestenfalls-Verlust und nicht still.
	if !errors.Is(fehler, ErrSchlussNachFrist) {
		t.Errorf("Schliessen gab %v", fehler)
	}

	freigeben()
}

// **Bestenfalls heißt: in der Frist oder gar nicht** — und ein Verlust wird
// gesagt.
func TestBestenfallsWartetHoechstensBisZurFrist(t *testing.T) {
	t.Parallel()

	if !Bestenfalls(t.Context(), "ein Schritt, der gleich fertig ist", func() {}) {
		t.Error("ein Schritt ohne Hindernis ging verloren")
	}

	frist, aufhoeren := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer aufhoeren()

	loslassen := make(chan struct{})
	defer close(loslassen)

	if Bestenfalls(frist, "ein Schritt, der hängt", func() { <-loslassen }) {
		t.Error("ein hängender Schritt galt als erledigt")
	}
}

// **Hängen die eigenen Läufe, endet der Schluss trotzdem in seiner Frist** —
// und nennt, was nicht fertig wurde. Hier hält der Test das Schloss des
// Tischbeitritts, an dem das Beenden der eigenen Läufe vorbeimuss.
func TestHaengendeEigeneLaeufeHaltenDenSchlussNichtAuf(t *testing.T) {
	t.Parallel()

	anwendung := gestartet(t)

	anwendung.partie.beitritt.schloss.Lock()

	frist, aufhoeren := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer aufhoeren()

	fehler := anwendung.Schliessen(frist, nil)

	anwendung.partie.beitritt.schloss.Unlock()

	if !errors.Is(fehler, ErrSchlussNachFrist) || !strings.Contains(fehler.Error(), "das Beenden der eigenen Läufe") {
		t.Errorf("Schliessen gab %v, erwartet war der Verlust beim Beenden der eigenen Läufe", fehler)
	}
}
