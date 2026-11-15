// **Die Formatprüfung** — Werkstück 5 von Phase 4.6
// (#1293).
package service_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Deck aus echten Drucken bekommt für jeden die Antwort von Scryfall.**
func TestEchteDruckeWerdenBeiScryfallGefragt(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.drucke.legalitaeten[druckA] = map[string]string{"standard": "legal"}
	werk.drucke.legalitaeten[druckB] = map[string]string{"standard": "banned"}

	werk.legeDruck(t, druckA, 4)
	werk.legeDruck(t, druckB, 1)

	urteil, befunde, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	if len(befunde) != 0 {
		t.Errorf("es gab Befunde: %v", befunde)
	}

	if len(urteil.Urteile) != 2 {
		t.Fatalf("%d Urteile, erwartet 2", len(urteil.Urteile))
	}

	gefunden := map[domain.Zulaessigkeit]bool{}
	for _, eines := range urteil.Urteile {
		gefunden[eines.Zulaessigkeit] = true

		// **Die Quelle steht dabei** — das verlangt die Abnahme des Issues
		// ausdrücklich: „mit dem Grund und der Quelle".
		if eines.Quelle == "" {
			t.Errorf("ein Urteil ohne Quelle: %+v", eines)
		}
	}

	if !gefunden[domain.ZulaessigLegal] || !gefunden[domain.ZulaessigVerboten] {
		t.Errorf("%+v", urteil.Urteile)
	}

	if urteil.Zulaessig() {
		t.Error("ein Deck mit einer verbotenen Karte gilt als zulässig")
	}
}

// **Eine eigene Karte ohne Verweis ist `ungeprüft` — und nennt den Grund.**
//
// **Nicht `nicht zulässig`.** Das ist die Zusage aus Vermerk 0061, und sie ist
// der Grund, warum es diesen Weg gibt.
func TestEineEigeneKarteOhneVerweisIstUngeprueft(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.legeKarte(t, 1)

	urteil, _, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	if len(urteil.Urteile) != 1 {
		t.Fatalf("%d Urteile, erwartet 1", len(urteil.Urteile))
	}

	eines := urteil.Urteile[0]

	if eines.Zulaessigkeit != domain.ZulaessigUngeprueft {
		t.Errorf("Zulässigkeit %q", eines.Zulaessigkeit)
	}

	if eines.Grund == "" {
		t.Error("ungeprüft ohne Grund — dann sagt es nichts")
	}

	if eines.Quelle != "" {
		t.Errorf("ungeprüft mit Quelle %q — von wem denn?", eines.Quelle)
	}

	// **Und die Karte selbst steht dem Deck nicht entgegen.**
	//
	// Geprüft wird hier **die Karte**, nicht das ganze Deck: dieses hat eine
	// Karte und das Format verlangt sechzig, also sagt `Zulaessig()` schon
	// wegen der Größe nein. Der erste Anlauf dieses Laufs hat beides
	// verwechselt und hätte auch dann bestanden, wenn `ungeprüft` als
	// Ablehnung gezählt hätte.
	ohneGroessen := domain.Deckurteil{Urteile: urteil.Urteile}
	if !ohneGroessen.Zulaessig() {
		t.Error("eine eigene Karte machte das Deck unzulässig")
	}

	if urteil.Ungeprueft() != 1 {
		t.Errorf("Ungeprueft() = %d, erwartet 1", urteil.Ungeprueft())
	}
}

// **Eine eigene Karte mit Verweis wird über ihren Druck gefragt.**
func TestEineEigeneKarteMitVerweisWirdUeberDenDruckGefragt(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.verweise.verweise[eigeneKarte] = domain.Kartenverweis{Printing: druckA}
	werk.drucke.legalitaeten[druckA] = map[string]string{"standard": "legal"}

	werk.legeKarte(t, 1)

	urteil, _, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	eines := urteil.Urteile[0]

	if eines.Zulaessigkeit != domain.ZulaessigLegal {
		t.Errorf("Zulässigkeit %q", eines.Zulaessigkeit)
	}

	// **Und die Herkunft steht dabei** — der Vermerk verlangt es für genau
	// diesen Fall: „wie oben, und die Herkunft steht dabei".
	if eines.Quelle == "" {
		t.Error("der Verweis kam ohne Herkunft")
	}
}

// **Ein Format ohne Scryfallnamen lässt jede Karte ungeprüft.**
//
// Ein eigenes Küchenformat hat keinen Eintrag in `legalities`, und eine
// erfundene Antwort wäre schlimmer als keine.
func TestEinFormatOhneScryfallnamen(t *testing.T) {
	werk := neuesPruefwerk(t)
	werk.formatOhneScryfallnamen(t)

	werk.drucke.legalitaeten[druckA] = map[string]string{"standard": "legal"}
	werk.legeDruck(t, druckA, 1)

	urteil, _, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	if urteil.Urteile[0].Zulaessigkeit != domain.ZulaessigUngeprueft {
		t.Errorf("%+v", urteil.Urteile[0])
	}

	// **Und Scryfall wurde gar nicht erst gefragt** — eine Anfrage, deren
	// Antwort niemand lesen kann, ist verbrauchtes Kontingent (§16, Regel 4).
	if werk.drucke.gefragt != 0 {
		t.Errorf("Scryfall wurde %d-mal gefragt", werk.drucke.gefragt)
	}
}

// **Ein Druck, den Scryfall zu diesem Format nicht führt, ist ungeprüft.**
func TestEinFormatDasScryfallZuDiesemDruckNichtFuehrt(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.drucke.legalitaeten[druckA] = map[string]string{"modern": "legal"}
	werk.legeDruck(t, druckA, 1)

	urteil, _, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	if urteil.Urteile[0].Zulaessigkeit != domain.ZulaessigUngeprueft {
		t.Errorf("%+v", urteil.Urteile[0])
	}
}

// **Ein Deck ohne Format wird nicht geprüft, sondern abgewiesen.**
//
// Es ist eine Frage ohne Gegenstand: „ist dieses Deck zulässig?" — wozu?
func TestEinDeckOhneFormatWirdAbgewiesen(t *testing.T) {
	werk := neuesPruefwerk(t)
	werk.formatHerunter(t)

	if _, _, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID); !errors.Is(
		err, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet ErrValidation", err)
	}
}

// **Die Größen werden mitgeprüft** — und tragen den Formatnamen.
func TestDieGroessenWerdenMitgeprueft(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.drucke.legalitaeten[druckA] = map[string]string{"standard": "legal"}
	werk.legeDruck(t, druckA, 1)

	urteil, _, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	if len(urteil.Groessen) == 0 {
		t.Fatal("ein Deck mit einer Karte kam durch ein Format mit Mindestgröße 60")
	}

	if urteil.Formatname != "Standard" {
		t.Errorf("Formatname %q", urteil.Formatname)
	}

	if urteil.Hauptdeck != 1 {
		t.Errorf("Hauptdeck %d", urteil.Hauptdeck)
	}
}

// **Ein Fehler bei Scryfall hält die Prüfung nicht an** — er wird zum Befund.
//
// **Dieselbe Haltung wie in §16, Regel 5**: ohne Netz arbeitet die Anwendung
// weiter. Eine Prüfung, die an einer nicht erreichbaren Gegenstelle abbricht,
// sagt über die anderen neunundfünfzig Karten nichts.
func TestEinFehlerBeiScryfallWirdZumBefund(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.drucke.fehler = errors.New("die Gegenstelle schweigt")
	werk.legeDruck(t, druckA, 1)

	urteil, befunde, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("die Prüfung brach ab: %v", err)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1", len(befunde))
	}

	if urteil.Urteile[0].Zulaessigkeit != domain.ZulaessigUngeprueft {
		t.Errorf("%+v", urteil.Urteile[0])
	}
}

// **Und ein Deck, das es nicht gibt, ist ein fachliches Nein.**
func TestEinFehlendesDeckLaesstSichNichtPruefen(t *testing.T) {
	werk := neuesPruefwerk(t)

	if _, _, err := werk.pruefung.Pruefe(t.Context(), 999); !errors.Is(
		err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", err)
	}
}

// **Ein Format, das der Bestand nicht hergibt, hält die Prüfung an.**
//
// **Anders als ein Fehler bei Scryfall**: ohne das Format ist gar nicht zu
// sagen, wonach gefragt wird. Es gibt nichts, was sich als `ungeprüft`
// melden ließe.
func TestEinUnlesbaresFormatHaeltDiePruefungAn(t *testing.T) {
	werk := neuesPruefwerk(t)

	kaputt := errors.New("der Formatspeicher schweigt")
	werk.formate.fehler = kaputt

	if _, _, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID); !errors.Is(
		err, kaputt) {
		t.Errorf("Fehler %v, erwartet den Speicherfehler", err)
	}
}

// **Ein Verweis, der sich nicht lesen lässt, wird zum Befund.**
//
// Dieselbe Haltung wie bei einer schweigenden Gegenstelle: die Karte bleibt
// ungeprüft, die anderen werden geprüft.
func TestEinUnlesbarerVerweisWirdZumBefund(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.legeKarte(t, 1)
	werk.verweise.fehler = errors.New("die Verweisablage schweigt")

	urteil, befunde, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("die Prüfung brach ab: %v", err)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1", len(befunde))
	}

	if urteil.Urteile[0].Zulaessigkeit != domain.ZulaessigUngeprueft {
		t.Errorf("%+v", urteil.Urteile[0])
	}
}

// **Zwei Namen mit einer Orakelkennung zählen zusammen** —
// #1375.
//
// Die Naht, nicht die Regel: dass `domain.gruppe` nach der Orakelkennung
// gruppiert, prüft `internal/domain`. Hier steht die Frage, ob die Kennung
// **aus Scryfalls Antwort** bis dorthin kommt — zwei grüne Seiten einer Naht
// belegen die Naht nicht.
//
// CR 100.2a: für den Deckbau gelten Karten mit austauschbaren Namen als
// derselbe englische Name. Vier von einem und vier
// vom anderen Namen sind acht derselben Karte.
func TestZweiNamenMitEinerOrakelkennungZaehlenZusammen(t *testing.T) {
	werk := neuesPruefwerk(t)

	for druck, name := range map[string]string{
		druckA: "Mind Flayer, the Shadow",
		druckB: "Arvinox, the Mind Flail",
	} {
		werk.drucke.legalitaeten[druck] = map[string]string{"standard": "legal"}
		werk.drucke.namen[druck] = name
		werk.drucke.orakel[druck] = "9abecd23-a2d9-43c7-8d01-f8afdc53026e"

		werk.legeDruck(t, druck, 4)
	}

	urteil, _, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	gleiche := 0

	for _, befund := range urteil.Groessen {
		if befund.Feld == "gleiche" {
			gleiche++

			if !strings.Contains(befund.Meldung, "8") {
				t.Errorf("die Meldung nennt nicht acht: %q", befund.Meldung)
			}
		}
	}

	if gleiche != 1 {
		t.Errorf("%d Befunde zur Höchstzahl, erwartet war einer: %v",
			gleiche, urteil.Groessen)
	}
}
