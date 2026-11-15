//go:build fremdpakete

// Der Nachweis für #240:
// **eine Karte aus einem bezogenen Paket zeichnet ihren Rahmen.**
//
//	MSE_FREMDPAKETE="$HOME/Library/Application Support/MagicSetEditor4Mac/pakete/M15-Magic-Pack/data" \
//	  go test -tags fremdpakete ./internal/service/ -run Rahmen -v
package service_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **`magic-m15` zeichnet den Rahmen einer roten Karte.**
//
// # Warum dieser Test und nicht die Zählung nebenan
//
// `TestDerPlanbauerUeberFremdpakete` zählt Bildebenen über 190 Stile. Die Zahl
// stieg von 310 auf 795 und sagte trotzdem nicht, ob **die eine** Karte
// zeichnet, um die es geht: eine Ebene mehr an einem Nebenfeld sieht in einer
// Summe genauso aus wie der Rahmen.
//
// Hier steht die Zusage einzeln und benannt: `card color` ergibt eine Bildebene
// über die volle Kartenfläche, und ihr Pfad zeigt auf die Rahmendatei der
// **roten** Karte.
//
// # Der Weg dorthin, damit ein späterer Leser ihn wiedererkennt
//
// `card color → image` ruft `{ module_card() }`. `module_card` steht in
// `magic-blends.mse-include/new-blends:985` und wird über zwei Ketten erreicht:
//
//  1. `magic.mse-game/game` → `init script: include file: script`
//  2. `magic.mse-game/script:34` → `include file: /magic-blends.mse-include/new-blends`
//
// Beide Ketten waren beschädigt, und beide Schäden zeigt diese Zusage an: ein
// relativer Verweis suchte im Paket des Laufs statt im Paket der Datei (#241),
// und ein `include file:` ließ seine eingerückten Nachbarzeilen fallen —
// darunter `trim_colors`, ohne das `color_background` nicht rechnet.
//
// **Nachgeprüft, nicht behauptet.** Am 2026-08-16 wurde die zweite Behebung
// (`ergebnis = append(ergebnis, e.knotenliste(kind.Kinder)...)` in
// `einbinden.go`) versuchsweise entfernt; beide Zusagen hier wurden rot, mit
// „keine Bildebene über die volle Kartenfläche". Ein Wächter ohne diese
// Gegenprobe ist eine Behauptung darüber, was er merken würde (§10).
func TestDerRahmenEinerRotenKarteZeichnet(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	vorlage := m15Vorlage(t, wurzel)

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "name", Text: "Feuerball"},
		{Field: "casting cost", Text: "2R"},
		{Field: "card color", Text: "red"},
		{Field: "type", Text: "Instant"},
		{Field: "text", Text: "Deal three damage to any target."},
	}}

	plan, _, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	// **Der Rahmen ist die Ebene über die ganze Karte.** Der weiße Grund ist
	// eine Fläche, keine Bildebene — er zählt hier nicht mit.
	var rahmen string

	for _, ebene := range plan.Ebenen {
		if ebene.Bild.Pfad != "" && ebene.Rahmen.Breite == 375 && ebene.Rahmen.Hoehe == 523 {
			rahmen = ebene.Bild.Pfad
		}
	}

	if rahmen == "" {
		t.Fatal("keine Bildebene über die volle Kartenfläche — der Rahmen fehlt")
	}

	// **Die rote Rahmendatei**, nicht irgendeine. `color_background` wählt sie
	// über die Farbe der Karte; ein Rahmen in der falschen Farbe wäre ein
	// Ergebnis, das aussieht wie ein richtiges.
	if !strings.HasSuffix(rahmen, "/rcard.jpg") {
		t.Errorf("Rahmen %q — erwartet war die Datei der roten Karte", rahmen)
	}
}

// **Und der Rahmen folgt der Farbe.** Die Gegenprobe: ohne sie hielte der Test
// auch eine Umsetzung für richtig, die immer dieselbe Datei nimmt.
func TestDerRahmenFolgtDerFarbe(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	vorlage := m15Vorlage(t, wurzel)

	for farbe, datei := range map[string]string{
		"white": "wcard.jpg",
		"blue":  "ucard.jpg",
		"black": "bcard.jpg",
		"green": "gcard.jpg",
	} {
		t.Run(farbe, func(t *testing.T) {
			karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
				{Field: "name", Text: "Probe"},
				{Field: "card color", Text: farbe},
			}}

			plan, _, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			for _, ebene := range plan.Ebenen {
				if strings.HasSuffix(ebene.Bild.Pfad, datei) {
					return
				}
			}

			t.Errorf("keine Ebene mit %s", datei)
		})
	}
}

// **Der Name beginnt bei 32, nicht bei 50** — die Probe auf die Vorgabewerte
// (#276).
//
// `magic-m15.mse-style` schreibt
// `left: { if card.card_symbol == "none" then 32 else 50 }`. Eine Karte ohne
// Symbol führt kein Feld `card symbol`; ohne die Vorgabewerte des Spielpakets
// war der Wert `nil`, der Vergleich falsch und der Name um 18 Punkte zu weit
// rechts. Am Vergleichsbild einer echten MSE-Ausgabe war es zu sehen.
func TestDerNameFolgtDemVorgabewertDesKartensymbols(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	vorlage := m15Vorlage(t, wurzel)

	if len(vorlage.Kartenfelder) == 0 {
		t.Fatal("das Spielpaket hat keine Kartenfelder geliefert")
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "name", Text: "Feuerball"},
		{Field: "card color", Text: "red"},
	}}

	plan, _, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Klartext() != "Feuerball" {
			continue
		}

		if ebene.Rahmen.Links != 32 {
			t.Errorf("der Name beginnt bei %v, erwartet waren 32", ebene.Rahmen.Links)
		}

		return
	}

	t.Fatal("keine Textebene mit dem Kartennamen")
}

// m15Vorlage lädt `magic-m15.mse-style` samt Spielpaket und Kartenfeldern.
//
// **Ohne Symbolschriften**, und bis zum 2026-08-16 stand hier, sie kämen mit.
// Sie kamen nie: die zurückgegebene [service.Vorlage] führt `Stil`,
// `Spielskript` und `Kartenfelder` und sonst nichts. Ein Helfer, dessen Kopf
// mehr verspricht als sein Rumpf liefert, lässt jeden Leser die Zusage darüber
// für geprüft halten — angestrichen vom Zweitblick über #274/#275/#277/#278
// (#292).
//
// **Der Weg durch die Naht steht daneben**, nicht hier:
// `app.TestDerRahmenDurchDieNaht` geht über die laufende Anwendung. Dieser
// Wächter hält den Planbauer fest, und das ist eine andere und ebenfalls
// nötige Zusage.
func m15Vorlage(t *testing.T, wurzel string) service.Vorlage {
	t.Helper()

	paket, err := msepaket.Oeffne(filepath.Join(wurzel, "magic-m15.mse-style"))
	if err != nil {
		t.Fatalf("Stilpaket: %v", err)
	}

	_ = paket.ModuleEinbinden(msepaket.NeuerPaketleser(wurzel)) //nolint:errcheck // Befunde sind hier kein Abbruch
	stil, _ := paket.Stilmodell()
	_ = paket.Schliessen() //nolint:errcheck // das Modell ist kopiert

	spiel, err := msepaket.Oeffne(filepath.Join(wurzel, "magic.mse-game"))
	if err != nil {
		t.Fatalf("Spielpaket: %v", err)
	}

	_ = spiel.ModuleEinbinden(msepaket.NeuerPaketleser(wurzel)) //nolint:errcheck // dito
	spielskript := spiel.Initskript()
	spielmodell, _ := spiel.Spielmodell()
	_ = spiel.Schliessen() //nolint:errcheck // Skript und Modell sind kopiert

	return service.Vorlage{
		Stil: stil, Spielskript: spielskript, Kartenfelder: spielmodell.Fields,
	}
}
