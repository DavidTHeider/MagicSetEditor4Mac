package app

// **Das Fenster des Wirts hört seinen Tisch unter eigenem Namen** — aus dem
// Zweitblick über Phase 6.11 (#1672, zu Z1).
//
// # Warum eigene Namen
//
// Seit #1627 meldet [service.Tischfunk] dem Fenster des Wirts, was ein Gast
// an seinem Tisch tut. Es kam unter **denselben** Namen an wie beim Gast —
// und dieselbe Anwendung kann zugleich Wirt eines Tisches und Gast an einem
// anderen sein. Ein Gast, der ohne Partie beitritt, nimmt die des ersten
// Bildes; das konnte das Bild des eigenen Tisches sein. Unter eigenem Namen
// hört ihn der Haken des Gastes gar nicht erst.

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDerWirtHoertSeinenTischUnterEigenemNamen(t *testing.T) {
	t.Parallel()

	wirt, gast := zweiAnwendungen(t)
	MitPartienleitung(wirt.anwendung, wirt.anwendung.db)

	partieID, objekt := einTischBeimWirt(t, wirt)

	if antwort := wirt.anwendung.PartieFreigeben(
		wirt.ctx, partieID, "anna", []string{"bert"},
	); antwort.Fehler != nil {
		t.Fatalf("PartieFreigeben: %+v", antwort.Fehler)
	}

	angenommen := make(chan error, 1)

	go func() {
		griff, fehler := wirt.anwendung.sitzungsverwaltung().Annehmen(
			wirt.ctx, wirt.leitung)
		if fehler == nil {
			wirt.anwendung.partiesitzungAufnehmen(wirt.ctx, griff)
		}

		angenommen <- fehler
	}()

	if antwort := gast.anwendung.PartieBeitreten(
		gast.ctx, string(wirt.kennung), partieID, PartiemitbringDTO{},
	); antwort.Fehler != nil {
		t.Fatalf("PartieBeitreten: %+v", antwort.Fehler)
	}

	if fehler := <-angenommen; fehler != nil {
		t.Fatalf("Annehmen: %v", fehler)
	}

	wirt.altesVerwerfen()

	if antwort := gast.anwendung.PartieWuenschen(gast.ctx, domain.Spielbefehl{
		Kennung: "b1", Art: "objekt.tappen",
		Nutzlast: map[string]any{"objekt": objekt},
	}); antwort.Fehler != nil {
		t.Fatalf("PartieWuenschen: %+v", antwort.Fehler)
	}

	gesehen := sammleBis(t, wirt.meldungen, EreignisEigenesTischbild)

	for _, name := range []string{EreignisPartietischbild, EreignisPartiezug} {
		if gesehen[name] {
			t.Errorf("das Fenster des Wirts bekam %q — den Namen des Gastes", name)
		}
	}

	if !gesehen[EreignisEigenerZug] {
		t.Errorf("der Zug des Gastes kam nicht als %q an: %v", EreignisEigenerZug, gesehen)
	}
}

// sammleBis nimmt Meldungen, bis `ende` kam, und sagt, welche es waren.
func sammleBis(t *testing.T, gemeldet chan string, ende string) map[string]bool {
	t.Helper()

	gesehen := map[string]bool{}
	frist := time.After(5 * time.Second)

	for {
		select {
		case ereignis := <-gemeldet:
			gesehen[ereignis] = true
			if ereignis == ende {
				return gesehen
			}

		case <-frist:
			t.Fatalf("%q wurde nicht gemeldet: %v", ende, gesehen)
		}
	}
}

// **Der Wirt erfährt, dass ein Gast Platz nimmt und wieder geht** — gefunden
// mit zwei Fenstern (Hüllenprobe `nat04`, S12.NAT-04): Bo saß am Tisch und
// hatte schon eine Münze geworfen, und die Freigabe des Wirts sagte „Noch
// sitzt niemand". Sie las ihren Stand nur auf „Nachsehen".
func TestDerWirtHoertDieBesetzungSeinesTisches(t *testing.T) {
	t.Parallel()

	wirt, gast := zweiAnwendungen(t)
	MitPartienleitung(wirt.anwendung, wirt.anwendung.db)

	partieID, _ := einTischBeimWirt(t, wirt)

	if antwort := wirt.anwendung.PartieFreigeben(
		wirt.ctx, partieID, "anna", []string{"bert"},
	); antwort.Fehler != nil {
		t.Fatalf("PartieFreigeben: %+v", antwort.Fehler)
	}

	wirt.altesVerwerfen()

	angenommen := make(chan error, 1)

	go func() {
		griff, fehler := wirt.anwendung.sitzungsverwaltung().Annehmen(
			wirt.ctx, wirt.leitung)
		if fehler == nil {
			wirt.anwendung.partiesitzungAufnehmen(wirt.ctx, griff)
		}

		angenommen <- fehler
	}()

	if antwort := gast.anwendung.PartieBeitreten(
		gast.ctx, string(wirt.kennung), partieID, PartiemitbringDTO{},
	); antwort.Fehler != nil {
		t.Fatalf("PartieBeitreten: %+v", antwort.Fehler)
	}

	if fehler := <-angenommen; fehler != nil {
		t.Fatalf("Annehmen: %v", fehler)
	}

	warteAufMeldung(t, wirt.meldungen, EreignisPartiefreigabe)

	gast.anwendung.PartieBeitrittBeenden(gast.ctx)

	warteAufMeldung(t, wirt.meldungen, EreignisPartiefreigabe)
}

// **Ohne Fenster meldet sich die Besetzung nirgends hin** — an der
// Kommandozeile gibt es keines, und das ist kein Fehler.
func TestDieBesetzungOhneFensterGehtInsLeere(t *testing.T) {
	t.Parallel()

	(&App{}).besetzungMelden("p1")
}
