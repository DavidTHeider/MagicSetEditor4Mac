package app

// **Wer geht, kann wiederkommen** — gefunden am Messlauf „Zwei Macs auf
// einem" zu Phase 6.7
// (#1668).
//
// Ein Gast stand auf und setzte sich gleich wieder: der Wirt wies ihn ab —
// „zu … läuft schon eine Sitzung für partie". **Die Schleife endete, der
// Griff blieb in der Übersicht**, auf beiden Seiten: der Wirt schloss ihn
// nach dem Ende nie, und der Gast nicht, wenn ihm die Leitung riss. Damit
// sperrte eine tote Sitzung jede neue zum selben Gegenüber, bis das Profil
// wechselte — und die Rückkehr nach einem Abriss, für die es `Pausiert`
// gibt, ging nie.

import (
	"fmt"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// warteAufKeinePartiesitzung wartet, bis die Übersicht keine Partie mehr
// führt — auf die Wirkung und nicht auf eine Frist (§3).
func warteAufKeinePartiesitzung(t *testing.T, wer *seite, rolle string) {
	t.Helper()

	frist := time.After(5 * time.Second)

	for {
		if !fuehrtPartie(wer) {
			return
		}

		select {
		case <-frist:
			t.Fatalf("beim %s steht die Partiesitzung nach ihrem Ende noch in der Übersicht", rolle)
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// gastAmTischDesWirts setzt den Gast an den Tisch — der Wirt nimmt an, wie
// es der Horchposten täte.
func gastAmTischDesWirts(t *testing.T) (wirt, gast *seite) {
	t.Helper()

	wirt, gast, _, _ = gastAnEinemTisch(t, 0)

	return wirt, gast
}

// gastAnEinemTisch setzt den Gast an einen Tisch, an dem der Wirt vorher
// `zuege` Mal getappt und enttappt hat — und gibt Partie und eine Karte des
// Gastes.
func gastAnEinemTisch(t *testing.T, zuege int) (wirt, gast *seite, partieID, objekt string) {
	t.Helper()

	wirt, gast = zweiAnwendungen(t)

	MitPartienleitung(wirt.anwendung, wirt.anwendung.db)

	partieID, objekt = einTischBeimWirt(t, wirt)
	stellen := ueberDieFlaeche(wirt.ctx, PartieflaecheUeber(wirt.anwendung.db))
	eigenes := inDerBibliothek(t, wirt.anwendung.Partie(wirt.ctx, partieID, "anna").Partie.Objekte, "anna")

	for zug := range zuege {
		art := "objekt.tappen"
		if zug%2 == 1 {
			art = "objekt.enttappen"
		}

		if antwort := stellen(domain.Spielbefehl{
			PartieID: partieID, Kennung: fmt.Sprintf("vorher-%d", zug), Absender: "anna",
			Art: art, Nutzlast: map[string]any{"objekt": eigenes},
		}); antwort.Fehler != nil {
			t.Fatalf("Zug %d des Wirts: %+v", zug, antwort.Fehler)
		}
	}

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
			wirt.anwendung.sitzungAufnehmen(wirt.ctx, griff)
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

	// **Die Schleife läuft erst, wenn der Wirt begrüßt hat** — sein erstes
	// Tischbild. Wer vorher aufsteht, prüft die gescheiterte Begrüßung, und
	// die räumte schon immer (der erste Anlauf dieses Laufs war deshalb grün).
	warteAufMeldung(t, gast.meldungen, EreignisPartietischbild)

	// **Erst der Beleg, dass die Übersicht sie führt** — sonst bewiese ein
	// leeres Ergebnis danach nichts (§3).
	for _, wer := range []struct {
		seite *seite
		rolle string
	}{{wirt, "Wirt"}, {gast, "Gast"}} {
		if !fuehrtPartie(wer.seite) {
			t.Fatalf("beim %s steht die Partiesitzung gar nicht in der Übersicht", wer.rolle)
		}
	}

	return wirt, gast, partieID, objekt
}

func fuehrtPartie(wer *seite) bool {
	for _, stand := range wer.anwendung.sitzungsverwaltung().Alle() {
		if stand.Zweck == domain.ZweckPartie {
			return true
		}
	}

	return false
}

// **Steht der Gast auf, gibt der Wirt die Sitzung frei.**
func TestNachDemAufstehenGibtDerWirtDieSitzungFrei(t *testing.T) {
	t.Parallel()

	wirt, gast := gastAmTischDesWirts(t)

	if auf := gast.anwendung.PartieBeitrittBeenden(gast.ctx); auf.Fehler != nil {
		t.Fatalf("Aufstehen: %+v", auf.Fehler)
	}

	warteAufKeinePartiesitzung(t, gast, "Gast")
	warteAufKeinePartiesitzung(t, wirt, "Wirt")
}

// **Reißt die Leitung, gibt auch der Gast die Sitzung frei** — sonst wiese
// ihn sein eigenes Programm bei der Rückkehr ab.
func TestNachDemAbrissGibtDerGastDieSitzungFrei(t *testing.T) {
	t.Parallel()

	wirt, gast := gastAmTischDesWirts(t)

	if fehler := gast.leitung.Schliesse(); fehler != nil {
		t.Fatalf("Schliesse: %v", fehler)
	}

	warteAufLeerenTisch(t, gast)

	if stand := gast.anwendung.PartieBeitrittStand(gast.ctx); !stand.Pausiert {
		t.Errorf("nach dem Abriss nicht pausiert: %+v", stand)
	}

	warteAufKeinePartiesitzung(t, gast, "Gast")
	warteAufKeinePartiesitzung(t, wirt, "Wirt")
}

// **Wer sich an einen laufenden Tisch setzt, bleibt sitzen.** Der Wirt
// reicht ihm nach, was fehlt — Antwort, jedes Ereignis, dann das Bild —, und
// legte das alles in eine Schlange von acht, bevor er sie leerte: ab gut
// sechs Zügen war sie voll, und der Wirt legte die Sitzung nieder. Gefunden
// am Messlauf, als der Gast im zweiten Weg zurückkam.
func TestEinGastSetztSichAnEinenLaufendenTisch(t *testing.T) {
	t.Parallel()

	wirt, gast, partieID, objekt := gastAnEinemTisch(t, 2*Sitzungstiefe)

	gast.altesVerwerfen()

	if antwort := gast.anwendung.PartieWuenschen(gast.ctx, domain.Spielbefehl{
		Kennung: "b1", Art: "objekt.tappen", Nutzlast: map[string]any{"objekt": objekt},
	}); antwort.Fehler != nil {
		t.Fatalf("PartieWuenschen: %+v", antwort.Fehler)
	}

	if !warteAufGetapptesObjekt(t, wirt, partieID, objekt) {
		t.Fatal("der Zug des Gastes kam beim Wirt nicht an — der Tisch hat ihn beim Nachreichen verloren")
	}
}
