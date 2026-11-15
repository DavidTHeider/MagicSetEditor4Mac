package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Zusatzfelder eines Stils zeichnen mit — und stehen als `extra_card`.**
//
// `magic-m15` zeichnet den Kasten für Stärke und Widerstandskraft nicht über
// ein Kartenfeld, sondern über ein eigenes Feld der Vorlage:
//
//	extra card field:
//		name: pt box
//		script: card.card_color
//	extra card style:
//		pt box:
//			left: 273 … visible: { card.pt != "" } … image: { module_ptbox() }
//
// `src/data/stylesheet.cpp`, Zeilen 116–125, Fassung f9d9356 vom 2021-07-06.
// Ohne sie fehlt der Kasten auf **jeder** Kreatur, und `extra_card.…` ist ein
// unbekannter Name — 232-mal über die 190 Stile von `M15-Magic-Pack`
// (#270).
func TestDieZusatzfelderEinesStilsZeichnenMit(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Zusatzfelder: []domain.Field{
				{Name: "pt box", Kind: domain.KindChoice, Skript: "card.farbe"},
			},
			Zusatzstile: []msepaket.Feldstil{{
				Feld: "pt box", Links: "10", Oben: "20", Breite: "30", Hoehe: "12",
				Darstellung: "image", Bild: `{ extra_card.pt_box + ".png" }`,
			}},
		},
		Kartenfelder: []domain.Field{{Name: "farbe", Kind: domain.KindText}},
	}

	karte := &domain.CardRow{
		ID: 1, Values: []domain.Value{{Field: "farbe", Text: "rot"}},
	}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(plan.Ebenen) != 2 {
		t.Fatalf("%d Ebenen, erwartet zwei", len(plan.Ebenen))
	}

	// **Der Wert kommt aus dem Skript des Zusatzfeldes**, und der Bildpfad
	// liest ihn über `extra_card`.
	if pfad := plan.Ebenen[1].Bild.Pfad; !strings.HasSuffix(pfad, "rot.png") {
		t.Errorf("das Bild ist %q, erwartet rot.png", pfad)
	}

	if rahmen := plan.Ebenen[1].Rahmen; rahmen.Links != 10 || rahmen.Breite != 30 {
		t.Errorf("Rahmen %+v", rahmen)
	}
}

// **Ein gescheitertes Zusatzfeld wird benannt und lässt den Rest stehen.**
//
// Dieselbe Regel wie bei den Kartenfeldern (§8): der Wert bleibt, wie er war,
// die übrigen Felder werden trotzdem gerechnet, und die Meldung geht mit.
func TestEinGescheitertesZusatzfeldWirdBenannt(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Zusatzfelder: []domain.Field{
				{Name: "kaputt", Skript: "nirgends_bekannt()"},
				{Name: "heil", Skript: `"da"`},
			},
			Zusatzstile: []msepaket.Feldstil{{
				Feld: "heil", Links: "0", Oben: "0", Breite: "30", Hoehe: "12",
				Darstellung: "image", Bild: `{ extra_card.heil + ".png" }`,
			}},
		},
	}

	plan, befunde, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), &domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "nirgends_bekannt") {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(plan.Ebenen) != 2 || !strings.HasSuffix(plan.Ebenen[1].Bild.Pfad, "da.png") {
		t.Errorf("das heile Feld fehlt: %+v", plan.Ebenen)
	}
}

// **Ohne Zusatzfelder bleibt `extra_card` leer, nicht unbekannt.**
//
// `magic.mse-game` schreibt `extra_card.vorthos_box or else ""` — der Name muss
// dastehen, sonst scheitert das Kartenfeld, statt seinen Ersatzwert zu nehmen.
func TestOhneZusatzfelderBleibtExtraCardLeer(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Felder: []msepaket.Feldstil{{
				Feld: "name", Links: "0", Oben: "0", Breite: "50", Hoehe: "10",
				Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "8"},
			}},
		},
		Kartenfelder: []domain.Field{
			{Name: "name", Kind: domain.KindText, Skript: `extra_card.gibtsnicht or else "leer"`},
		},
	}

	plan, befunde, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), &domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(plan.Ebenen) < 2 || plan.Ebenen[1].Klartext() != "leer" {
		t.Errorf("gezeichnet wird %q", plan.Ebenen[len(plan.Ebenen)-1].Klartext())
	}
}

// **Ein Abbruch wirkt auch zwischen den Zusatzfeldern.**
//
// Sie stehen am Ende des Laufs, und jedes rechnet ein Skript. Wer während des
// Tippens die Karte wechselt, soll nicht darauf warten, dass die alte fertig
// gerechnet wird (§7). Der Kontext ist **erst beim n-ten Blick** abgebrochen —
// ein von Anfang an abgebrochener käme nie bis hierher.
func TestEinAbbruchWirktZwischenDenZusatzfeldern(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Zusatzfelder: []domain.Field{
				{Name: "eins", Skript: `"a"`},
				{Name: "zwei", Skript: `"b"`},
			},
		},
	}

	for verbleibend := 1; verbleibend < 30; verbleibend++ {
		zaehler := verbleibend

		plan, _, err := service.NeuerPlanbauer().Stilplan(
			spaeterAbbruch{Context: t.Context(), verbleibend: &zaehler},
			&domain.CardRow{ID: 1}, vorlage)
		if err == nil && plan == nil {
			t.Fatalf("bei %d: weder Plan noch Fehler", verbleibend)
		}
	}
}

// **Ein Zusatzfeld, dessen Skript einen Fehlerwert ergibt, wird benannt.**
//
// Dieselbe Regel wie beim Kartenfeld: MSEs Zuweisung ruft `toString()`, und
// bei einem `ScriptDelayedError` wirft die (`src/script/value.cpp`, Zeile 148,
// Fassung f9d9356 vom 2021-07-06). Der Wert bleibt leer
// (#345).
func TestEinZusatzfeldMitFehlerwertWirdBenannt(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Zusatzfelder: []domain.Field{{Name: "kasten", Skript: "card.gibtsnicht"}},
			Zusatzstile: []msepaket.Feldstil{{
				Feld: "kasten", Links: "0", Oben: "0", Breite: "30", Hoehe: "12",
				Darstellung: "image", Bild: `{ extra_card.kasten + ".png" }`,
			}},
		},
	}

	_, befunde, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), &domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "gibtsnicht") {
		t.Errorf("Befunde: %v", befunde)
	}
}
