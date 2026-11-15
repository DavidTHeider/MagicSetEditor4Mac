package app

// **Das Gespräch am Tisch, von einem Ende zum anderen** — Werkstück G5 von
// Phase 5.5 (#1462).
//
// # Wofür es diesen Lauf gibt
//
// **Weil beide Lücken nur hier sichtbar werden.** Jedes Stück hatte seinen
// eigenen Lauf, und jeder war grün:
//
//   - Der Wirt reichte Chat und Präsenz **nirgendwohin** — `MitNebenstrom`
//     stand auf der Gastseite und fehlte auf der Wirtsseite. Ein Satz des
//     Gastes erreichte den Wirt und verschwand still.
//   - Der Wirt hielt **keine Sitzungsliste**. Er konnte antworten und sonst
//     nichts; von sich aus etwas sagen ging gar nicht.
//
// **Beides sieht man keinem Einzelteil an.** Der Wirt nimmt entgegen, das
// Fenster kann melden, die Runde kann rundrufen — zusammengesteckt war es
// nicht. Genau dafür steht der Satz aus §16: ein Wächter sieht, dass etwas da
// ist, nicht dass es dasselbe tut.

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Satz des Gastes kommt beim Wirt an — und einer des Wirts beim Gast.**
func TestAmTischWirdInBeideRichtungenGeredet(t *testing.T) {
	t.Parallel()

	wirt, gast := einTischMitGast(t)

	wirt.altesVerwerfen()
	gast.altesVerwerfen()

	if antwort := gast.anwendung.PartieSagen(
		gast.ctx, "ich bin da"); antwort.Fehler != nil {
		t.Fatalf("PartieSagen (Gast): %+v", antwort.Fehler)
	}

	// **Der Sprecher ist der Sitz, den der Wirt vergeben hat** — nicht der
	// Platzhalter, den der Gast geschrieben hat.
	gesagt := warteAufNebenbei(t, wirt.nebenbei, "chat")
	if gesagt.Von != "bert" || gesagt.Text != "ich bin da" {
		t.Errorf("beim Wirt steht %q von %q", gesagt.Text, gesagt.Von)
	}

	if antwort := wirt.anwendung.PartieSagen(
		wirt.ctx, "willkommen"); antwort.Fehler != nil {
		t.Fatalf("PartieSagen (Wirt): %+v", antwort.Fehler)
	}

	zurueck := warteAufNebenbei(t, gast.nebenbei, "chat")
	if zurueck.Von != "anna" || zurueck.Text != "willkommen" {
		t.Errorf("beim Gast steht %q von %q", zurueck.Text, zurueck.Von)
	}
}

// **Ein Zeiger des Wirts erreicht den Gast — mit Sitz, Art, Objekt und Ort.**
func TestEinZeigerDesWirtsErreichtDenGast(t *testing.T) {
	t.Parallel()

	wirt, gast := einTischMitGast(t)

	gast.altesVerwerfen()

	if antwort := wirt.anwendung.PartieZeigen(
		wirt.ctx, string(domain.PraesenzSchwebt), "k1", 120, 240,
	); antwort.Fehler != nil {
		t.Fatalf("PartieZeigen: %+v", antwort.Fehler)
	}

	gezeigt := warteAufNebenbei(t, gast.nebenbei, "praesenz")

	switch {
	case gezeigt.Von != "anna":
		t.Errorf("der Zeiger gehört %q statt anna", gezeigt.Von)
	case gezeigt.Zeigerart != string(domain.PraesenzSchwebt):
		t.Errorf("er ist ein %q", gezeigt.Zeigerart)
	case gezeigt.Objekt != "k1":
		t.Errorf("er zeigt auf %q", gezeigt.Objekt)
	case gezeigt.X != 120 || gezeigt.Y != 240:
		t.Errorf("er steht bei %v/%v statt 120/240", gezeigt.X, gezeigt.Y)
	}
}

// **Wer an keinem Tisch sitzt, redet nicht.**
func TestOhneTischGehtKeinSatzHinaus(t *testing.T) {
	t.Parallel()

	allein := PeerflaecheUeber(bestandFuerDenRiegel(t))
	ctx := allein.imProfil(riegelkontext(t))

	if antwort := allein.PartieSagen(ctx, "jemand da?"); antwort.Fehler == nil {
		t.Error("ein Satz ging hinaus, obwohl kein Tisch offensteht")
	}

	if antwort := allein.PartieZeigen(
		ctx, string(domain.PraesenzZeiger), "", 0, 0,
	); antwort.Fehler == nil {
		t.Error("ein Zeiger ging hinaus, obwohl kein Tisch offensteht")
	}
}

// einTischMitGast baut den Tisch aus `partietischlauf_intern_test.go` auf.
//
// **Herausgelöst und nicht kopiert**: zwei Aufbauten liefen beim ersten
// Unterschied auseinander, und dann prüfte dieser Lauf einen anderen Tisch als
// jener.
func einTischMitGast(t *testing.T) (wirt, gast *seite) {
	t.Helper()

	wirt, gast = zweiAnwendungen(t)

	MitPartienleitung(wirt.anwendung, wirt.anwendung.db)

	partieID, _ := einTischBeimWirt(t, wirt)

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

	antwort := gast.anwendung.PartieBeitreten(
		gast.ctx, string(wirt.kennung), partieID, PartiemitbringDTO{})

	if fehler := <-angenommen; fehler != nil {
		t.Fatalf("Annehmen: %v", fehler)
	}

	if antwort.Fehler != nil {
		t.Fatalf("PartieBeitreten: %+v", antwort.Fehler)
	}

	warteAufSitz(t, wirt)

	return wirt, gast
}

// warteAufSitz wartet, bis der Wirt den Platz des Gastes eingetragen hat.
//
// **Auf die Wirkung und nicht auf den Ruf** (§3): `PartieBeitreten` kehrt
// zurück, sobald die Leitung steht — dass der Wirt daraus einen Platz gemacht
// hat, geschieht in seiner eigenen Goroutine.
func warteAufSitz(t *testing.T, wirt *seite) {
	t.Helper()

	frist := time.After(5 * time.Second)

	for {
		if len(wirt.anwendung.PartieFreigabeStand(wirt.ctx).Vergeben) > 0 {
			return
		}

		select {
		case <-frist:
			t.Fatal("der Wirt hat keinen Sitz vergeben")
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// warteAufNebenbei wartet auf die erste Meldung einer Art.
func warteAufNebenbei(
	t *testing.T, gemeldet chan NebenbeiDTO, art string,
) NebenbeiDTO {
	t.Helper()

	frist := time.After(5 * time.Second)

	for {
		select {
		case eines := <-gemeldet:
			if eines.Art == art {
				return eines
			}

		case <-frist:
			t.Fatalf("es kam nichts als %q an", art)
		}
	}
}

// **Beide Wege fragen nach der Rolle, bevor etwas hinausgeht.**
//
// **`externSenden` und nicht `lokalSchreiben`**: ein Satz verlässt diesen
// Rechner, und was ihn hinauslässt, ist eine Entscheidung über die Reichweite
// von Daten (N7, #1568).
func TestJederGespraechswegFragtNachDerRolle(t *testing.T) {
	t.Parallel()

	for name, fahre := range map[string]func(*App) PartienebenbeiAntwort{
		"PartieSagen": func(a *App) PartienebenbeiAntwort {
			return a.PartieSagen(t.Context(), "hallo")
		},
		"PartieZeigen": func(a *App) PartienebenbeiAntwort {
			return a.PartieZeigen(
				t.Context(), string(domain.PraesenzZeiger), "", 0, 0)
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wache := &zaehlendeWache{}
			anwendung := freigabeanwendung(t).MitWache(wache)

			if antwort := fahre(anwendung); antwort.Fehler == nil {
				t.Fatal("der Weg kam durch")
			}

			if len(wache.gefragt) == 0 ||
				wache.gefragt[0] != domain.WirkungExternSenden {
				t.Errorf("gefragt wurde nach %v statt nach externSenden",
					wache.gefragt)
			}
		})
	}
}

// **Ein Gast zeigt, und der Wirt sieht es — mit einem Stand, der steigt.**
//
// # Warum der Stand hier geprüft wird und nicht nur im Dienst
//
// **Weil er in der Anwendung liegt und nicht im Gast.** `service.Partiegast`
// wird je Ruf neu gebaut; ein Zähler darin finge bei jedem Zeigen wieder bei
// eins an, und der Wirt wiese alles nach dem ersten als überholt ab. Dass er
// die einzelne Handlung überlebt, sieht man nur an zwei Rufen nacheinander.
func TestEinGastZeigtUndDerWirtSiehtEs(t *testing.T) {
	t.Parallel()

	wirt, gast := einTischMitGast(t)

	wirt.altesVerwerfen()

	for _, wohin := range [][2]float64{{10, 20}, {30, 40}} {
		if antwort := gast.anwendung.PartieZeigen(
			gast.ctx, string(domain.PraesenzZeiger), "", wohin[0], wohin[1],
		); antwort.Fehler != nil {
			t.Fatalf("PartieZeigen: %+v", antwort.Fehler)
		}
	}

	erster := warteAufNebenbei(t, wirt.nebenbei, "praesenz")
	if erster.Von != "bert" || erster.X != 10 {
		t.Errorf("der erste Zeiger steht bei %v von %q", erster.X, erster.Von)
	}

	// **Der zweite kommt an, und das ist die Aussage dieses Laufs.** Wäre der
	// Zähler je Ruf frisch, hielte der Wirt ihn für eine Wiederholung.
	zweiter := warteAufNebenbei(t, wirt.nebenbei, "praesenz")
	if zweiter.X != 30 || zweiter.Y != 40 {
		t.Errorf("der zweite Zeiger steht bei %v/%v", zweiter.X, zweiter.Y)
	}
}

// **Was keine Gestalt hat, geht gar nicht erst hinaus.**
func TestEinGastSchicktNichtsFormlosesVomTisch(t *testing.T) {
	t.Parallel()

	_, gast := einTischMitGast(t)

	if antwort := gast.anwendung.PartieSagen(
		gast.ctx, "   "); antwort.Fehler == nil {
		t.Error("ein leerer Satz ging hinaus")
	}

	if antwort := gast.anwendung.PartieZeigen(
		gast.ctx, "winken", "", 0, 0); antwort.Fehler == nil {
		t.Error("eine erfundene Zeigerart ging hinaus")
	}
}
