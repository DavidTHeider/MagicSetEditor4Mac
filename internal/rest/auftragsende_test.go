package rest_test

// **Das Dienstende räumt das Auftragsbuch** — BK-07b
// (#2116).
//
// Ein angenommener Auftrag überlebt seine Startantwort — absichtlich. Aber
// `http.Server.Shutdown` kennt ihn nicht: der Dienst endete, der Bestand ging
// zu, und der Auftrag schrieb weiter in eine Verbindung, die es nicht mehr
// gab. Jetzt gehört er dem Buch, und das Buch schließt mit dem Dienst.

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// wartenderLauf wartet auf seinen Abbruch und meldet dann `abgebrochen` —
// so, wie ein Rastern, das zwischen zwei Karten den Kontext fragt.
func wartenderLauf(begonnen chan<- struct{}, nachDemEnde *atomic.Bool) rest.Auftragslauf {
	return func(ctx context.Context, _ rest.Auftragseingabe, _ domain.Fortschrittsmelder) rest.Auftragsergebnis {
		close(begonnen)
		<-ctx.Done()

		// **Was nach dem Abbruch noch geschähe**, zählt hier: ein Lauf, den
		// niemand abbricht, käme an diese Stelle nie.
		nachDemEnde.Store(true)

		return rest.Auftragsergebnis{Status: 499}
	}
}

func TestDasDienstendeBrichtDenLaufendenAuftragAbUndWartet(t *testing.T) {
	t.Parallel()

	buch := rest.NeuesAuftragsbuch(time.Now)
	begonnen := make(chan struct{})

	var abgebrochen atomic.Bool

	auftrag, fehler := buch.Annehmen(t.Context(), rest.Auftragseingabe{Weg: "POST /sets"},
		wartenderLauf(begonnen, &abgebrochen))
	if fehler != nil {
		t.Fatal(fehler)
	}

	<-begonnen

	frist, aufhoeren := context.WithTimeout(t.Context(), 2*time.Second)
	defer aufhoeren()

	if fehler := buch.Schliessen(frist); fehler != nil {
		t.Fatalf("Schliessen: %v", fehler)
	}

	// **Gewartet heißt: der Lauf ist zu Ende**, nicht nur abgebrochen.
	if !abgebrochen.Load() {
		t.Error("Schliessen kehrte zurück, bevor der Lauf endete")
	}

	if stand, fehler := buch.Auftrag(t.Context(), auftrag.ID); fehler != nil || stand.Stand != "abgebrochen" {
		t.Errorf("der Auftrag steht als %q", stand.Stand)
	}

	// **Danach nimmt das Buch nichts mehr an** — ein Auftrag nach dem Ende
	// liefe gegen einen Bestand, der gleich zugeht.
	if _, fehler := buch.Annehmen(t.Context(), rest.Auftragseingabe{Weg: "POST /sets"}, fertigerLauf); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("nach dem Ende angenommen: %v", fehler)
	}

	// **Und ein zweites Schließen ist keines.**
	if fehler := buch.Schliessen(frist); fehler != nil {
		t.Errorf("zweites Schliessen: %v", fehler)
	}
}

// **Ein Lauf, der den Abbruch nicht hört, hält das Ende nicht auf** — nach
// der Frist kehrt `Schliessen` mit einem benannten Fehler zurück.
func TestEinTauberAuftragHaeltDasDienstendeNichtAuf(t *testing.T) {
	t.Parallel()

	buch := rest.NeuesAuftragsbuch(time.Now)
	begonnen, loslassen := make(chan struct{}), make(chan struct{})

	t.Cleanup(func() { close(loslassen) })

	if _, fehler := buch.Annehmen(t.Context(), rest.Auftragseingabe{Weg: "POST /sets"},
		func(context.Context, rest.Auftragseingabe, domain.Fortschrittsmelder) rest.Auftragsergebnis {
			close(begonnen)
			<-loslassen

			return rest.Auftragsergebnis{Status: http.StatusCreated}
		}); fehler != nil {
		t.Fatal(fehler)
	}

	<-begonnen

	frist, aufhoeren := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer aufhoeren()

	angefangen := time.Now()

	fehler := buch.Schliessen(frist)
	if !errors.Is(fehler, rest.ErrAuftraegeLaufenNoch) {
		t.Errorf("Schliessen gab %v", fehler)
	}

	if dauer := time.Since(angefangen); dauer > time.Second {
		t.Errorf("Schliessen brauchte %v", dauer)
	}
}

// **Der Dienst schließt sein Buch** — und ein Dienst ohne Buch hat nichts zu
// schließen.
func TestDerDienstSchliesstSeinAuftragsbuch(t *testing.T) {
	t.Parallel()

	buch := rest.NeuesAuftragsbuch(time.Now)
	dienst := rest.NeuerDienst("marke", time.Now, rest.Naehte{Auftraege: buch})

	if fehler := dienst.AuftraegeSchliessen(t.Context()); fehler != nil {
		t.Fatalf("AuftraegeSchliessen: %v", fehler)
	}

	if _, fehler := buch.Annehmen(t.Context(), rest.Auftragseingabe{Weg: "POST /sets"}, fertigerLauf); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("das Buch nimmt nach dem Schließen an: %v", fehler)
	}

	if fehler := rest.NeuerDienst("marke", time.Now, rest.Naehte{}).AuftraegeSchliessen(t.Context()); fehler != nil {
		t.Errorf("ohne Buch: %v", fehler)
	}
}
