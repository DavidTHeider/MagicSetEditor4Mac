package app

// **Was am Tisch schiefgeht** — Werkstück G5 von Phase 5.5
// (#1462).
//
// **Aus `partietischlauf_intern_test.go` herausgelöst**, als sie an die
// Zeilengrenze aus §2 stieß. Der Schnitt folgt der Frage: dort läuft es
// **durch** — ein Gast setzt sich, zieht, und beide sehen dasselbe —, hier
// geht es schief: die Leitung reißt ab, das Recht fällt, eine Nachricht ist
// unlesbar.
//
// # Und das ist kein Nebenschauplatz
//
// **Fehlerwege sind die Hälfte der Zusage** (§10). „Ein entzogener Zweck hält
// an" ist genauso viel wert wie „ein Zug kommt an" — und teurer zu prüfen,
// weil niemand im Betrieb absichtlich dorthin gerät.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Aufstehen räumt den Tisch, und ein abgerissener räumt sich selbst.**
//
// # Drei Zusagen in einem Lauf, und sie hängen zusammen
//
//   - **Wer aufsteht, sitzt danach nicht mehr** — sonst zeigte das Fenster
//     einen Tisch, an dem niemand ist.
//   - **Ein Wunsch über eine tote Leitung klagt**, statt still in einer
//     Schlange zu landen. Er ist eingereiht, bevor geschickt wird; ohne diese
//     Meldung hielte der Mensch ihn für gewirkt.
//   - **Und wenn die Leitung endet, räumt die Schleife selbst.** Sonst bliebe
//     ein Tisch stehen, den niemand mehr bedient.
func TestEinAbgerissenerTischRaeumtSichSelbst(t *testing.T) {
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
		_, fehler := wirt.anwendung.sitzungsverwaltung().Annehmen(
			wirt.ctx, wirt.leitung)
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

	// **Die Leitung reißt ab.** Der Griff weiß noch nichts davon.
	if fehler := gast.leitung.Schliesse(); fehler != nil {
		t.Fatalf("Schliesse: %v", fehler)
	}

	antwort := gast.anwendung.PartieWuenschen(gast.ctx, domain.Spielbefehl{
		Kennung: "b1", Art: "objekt.tappen",
		Nutzlast: map[string]any{"objekt": objekt},
	})
	if antwort.Fehler == nil {
		t.Error("der Wunsch über die tote Leitung kam durch")
	}

	// **Und die Schleife räumt den Tisch ab.**
	warteAufLeerenTisch(t, gast)

	// **Aufstehen bleibt trotzdem eine Antwort und keine Klage.**
	if auf := gast.anwendung.PartieBeitrittBeenden(gast.ctx); auf.Fehler != nil {
		t.Errorf("Aufstehen: %+v", auf.Fehler)
	}
}

// **Und wer aufsteht, während die Leitung steht, schließt sie.**
func TestWerAufstehtSchliesstDieSitzung(t *testing.T) {
	t.Parallel()

	wirt, gast := zweiAnwendungen(t)

	MitPartienleitung(wirt.anwendung, wirt.anwendung.db)

	partieID, _ := einTischBeimWirt(t, wirt)

	if antwort := wirt.anwendung.PartieFreigeben(
		wirt.ctx, partieID, "anna", []string{"bert"},
	); antwort.Fehler != nil {
		t.Fatalf("PartieFreigeben: %+v", antwort.Fehler)
	}

	angenommen := make(chan error, 1)

	go func() {
		_, fehler := wirt.anwendung.sitzungsverwaltung().Annehmen(
			wirt.ctx, wirt.leitung)
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

	if auf := gast.anwendung.PartieBeitrittBeenden(gast.ctx); auf.Fehler != nil {
		t.Fatalf("Aufstehen: %+v", auf.Fehler)
	}

	if stand := gast.anwendung.PartieBeitrittStand(gast.ctx); stand.Laeuft {
		t.Errorf("danach sitzt noch jemand: %+v", stand)
	}
}

// warteAufLeerenTisch wartet, bis die Schleife den Tisch abgeräumt hat.
//
// **Auf die Wirkung und nicht auf eine Frist** (§3): sie läuft nebenher, und
// wann sie merkt, dass die Leitung weg ist, sagt niemand.
func warteAufLeerenTisch(t *testing.T, gast *seite) {
	t.Helper()

	frist := time.After(5 * time.Second)

	for {
		if !gast.anwendung.PartieBeitrittStand(gast.ctx).Laeuft {
			return
		}

		select {
		case <-frist:
			t.Fatal("der abgerissene Tisch steht immer noch")
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// **Ein entzogenes Recht hält den Wunsch auf.**
//
// **Das ist die Zusage aus N1**: eine Freigabe kann zurückgenommen werden,
// während die Sitzung offen ist, und dann geht nichts mehr hinaus. Ohne diese
// Prüfung wäre „Zweck entziehen" eine Angabe ohne Wirkung — und genau diese
// Gestalt hat N7 an anderer Stelle angestrichen (§10).
func TestEinEntzogenesRechtHaeltDenWunschAuf(t *testing.T) {
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
		_, fehler := wirt.anwendung.sitzungsverwaltung().Annehmen(
			wirt.ctx, wirt.leitung)
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

	// **Erst jetzt fällt das Recht** — die Sitzung steht schon.
	if antwort := gast.anwendung.PeerzweckEntziehen(
		gast.ctx, string(wirt.kennung), string(domain.ZweckPartie),
	); antwort.Fehler != nil {
		t.Fatalf("PeerzweckEntziehen: %+v", antwort.Fehler)
	}

	antwort := gast.anwendung.PartieWuenschen(gast.ctx, domain.Spielbefehl{
		Kennung: "b1", Art: "objekt.tappen",
		Nutzlast: map[string]any{"objekt": objekt},
	})

	if antwort.Fehler == nil {
		t.Error("der Wunsch ging trotz entzogenem Recht hinaus")
	}
}

// **Eine unlesbare Nachricht des Wirts hält den Gast nicht auf** (§8).
//
// **Sie ist ein Befund über eine Nachricht.** Wer dafür den Tisch verließe,
// ließe den Wirt die Sitzung mit einem einzigen kaputten Byte beenden.
func TestEineUnlesbareNachrichtDesWirtsHaeltDenGastNichtAuf(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)

	sitzung := eineAngenommene(t, domain.ZweckPartie)
	sitzung.gibt = [][]byte{[]byte("{kein json")}
	sitzung.lesefehler = errors.New("danach ist Schluss")

	_, lauf := anwendung.beitrittMerken(sitzung, "p1")

	anwendung.amTischSitzen(t.Context(), sitzung, "p1",
		service.NeuePartiesitzung("p1", "gast", sitzung,
			service.NeueReihenfolge(erstePartieepoche)), lauf)

	if sitzung.gelesen != 1 {
		t.Errorf("gelesen wurde %dmal — die Schleife lief nicht weiter",
			sitzung.gelesen)
	}
}
