package app

// **Was nicht mitgehen kann, setzt niemanden an einen Tisch** — P9b
// (#2377).
// Ein Deck, das sich nicht lesen lässt, ist ein Fehler **hier** und keine
// Absage des Wirts; gewählt wird dann gar nicht erst.

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

func TestWasNichtMitgehenKannSetztNiemanden(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		mit   func(deckID string) PartiemitbringDTO
		code  Fehlercode
		grund string
	}{
		"keine Zahl":     {func(string) PartiemitbringDTO { return PartiemitbringDTO{DeckID: "x"} }, CodeValidierung, "mitgebrachte Deck"},
		"kein Deck":      {func(string) PartiemitbringDTO { return PartiemitbringDTO{DeckID: "999"} }, CodeNichtGefunden, ""},
		"Deck und Leihe": {func(id string) PartiemitbringDTO { return PartiemitbringDTO{DeckID: id, LeiheID: "l1"} }, CodeValidierung, "nicht beides"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anwendung := MitPartienleitung(freigabeanwendung(t), bestandFuerDenRiegel(t))
			deck, fehler := storage.NeueDeckablage(anwendung.db).Anlegen(riegelkontext(t),
				domain.Deck{Name: "Eigen", Spiel: "magic"})
			if fehler != nil {
				t.Fatalf("Deck anlegen: %v", fehler)
			}

			antwort := anwendung.PartieBeitreten(t.Context(), "ein-gegenueber", "",
				fall.mit(strconv.FormatInt(deck.ID, 10)))
			if antwort.Fehler == nil || antwort.Fehler.Code != fall.code ||
				!strings.Contains(antwort.Fehler.Meldung+fmt.Sprint(antwort.Fehler.Werte), fall.grund) {
				t.Fatalf("die Antwort: %+v", antwort.Fehler)
			}

			if anwendung.PartieBeitrittStand(t.Context()).Laeuft {
				t.Error("trotzdem am Tisch")
			}
		})
	}
}

// **Lässt sich die Partie am eigenen Tisch nicht lesen, gibt es kein Bild** —
// und der Grund geht zurück.
func TestOhneLesbarePartieFragtDasFensterNiemanden(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t).MitPartienaht(&unbekannteLeitung{})
	anwendung.spielbilder = vorschauweg.NeuesTexturfach(1 << 20)
	anwendung.partie.freigabe.offen = service.NeuePartiefreigabe("weg", []string{"bert"})
	anwendung.partie.freigabe.wirt = "anna"

	antwort := anwendung.Fremdkartenbild(t.Context(), "weg", "o1", "b1", "vorne", "tisch")
	if antwort.Fehler == nil || antwort.Fehler.Code != CodeNichtGefunden {
		t.Errorf("die Antwort: %+v", antwort)
	}
}

// **Der eigene Tisch geht vor einer Partie, die der Gast nur genannt hat** —
// Nachlauf zum Zweitblick über Phase 7, R2. Nennt ein Gast beim Beitreten die
// Kennung seines eigenen offenen Tisches, entsteht sofort eine fremde
// Sendeseite; bis der fremde Wirt antwortet, gingen Bildrufe des eigenen
// Tisches sonst dorthin.
func TestDerEigeneTischGehtVorEinerGenanntenPartie(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t).MitPartienaht(&unbekannteLeitung{})
	anwendung.spielbilder = vorschauweg.NeuesTexturfach(1 << 20)
	anwendung.partie.freigabe.offen = service.NeuePartiefreigabe("weg", []string{"bert"})
	anwendung.partie.freigabe.wirt = "anna"
	anwendung.beitrittMerken(eineAngenommene(t, domain.ZweckPartie), "weg")

	vorher := time.Now()

	antwort := anwendung.Fremdkartenbild(t.Context(), "weg", "o1", "b1", "vorne", "tisch")
	if antwort.Fehler == nil || antwort.Fehler.Code != CodeNichtGefunden || time.Since(vorher) > Bildfrist/2 {
		t.Errorf("die Antwort: %+v nach %v", antwort, time.Since(vorher))
	}
}
