// **Die Herkunft einer Übernahme** —
// #888.
//
// **Eine eigene Datei, weil sie eine eigene Frage beantwortet**: nicht *was*
// übernommen wird, sondern ob danach erkennbar bleibt, **woher** es kam.
package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Übernahme trägt ihre Herkunft bis in den Schreibaufruf** —
// Abnahmekriterium aus #888:
// „Herkunft je Feld: was von dort kam, ist erkennbar."
//
// # Was dieser Test hält und was nicht
//
// Er hält, dass der Dienst die Herkunft **weiterreicht**. Dass sie im Bestand
// ankommt, hält `internal/storage/herkunft_test.go`; dass sie über die Naht
// geht, `internal/app`. Drei Tests, weil es drei Nähte sind — und jede kann
// für sich brechen.
func TestEineUebernahmeReichtIhreHerkunftWeiter(t *testing.T) {
	t.Parallel()

	schreiber := &mitschreiber{}
	abgleich := service.NeuerScryfallabgleich(fremdkarte(), nil, neuesVerweisdoppel(), schreiber)

	vorschlag := mussVorschlag(t, abgleich, map[string]string{"rule text": "Alt"}, "Prüfstein")

	if _, fehler := abgleich.Uebernehmen(
		t.Context(), 1, 2, vorschlag, []string{"rule text"}); fehler != nil {
		t.Fatalf("Übernehmen: %v", fehler)
	}

	if len(schreiber.geschrieben) != 1 {
		t.Fatalf("geschrieben wurden %d Felder", len(schreiber.geschrieben))
	}

	if schreiber.geschrieben[0].herkunft != vorschlag.Herkunft {
		t.Errorf("Herkunft = %q, erwartet %q — ohne sie ist eine Übernahme nach "+
			"dem Schreiben nicht mehr von einer Eingabe zu unterscheiden",
			schreiber.geschrieben[0].herkunft, vorschlag.Herkunft)
	}

	if schreiber.geschrieben[0].herkunft == "" {
		t.Error("die Herkunft ist leer — dann sagt sie nichts")
	}
}

// **Die Herkunft geht durch alle drei Nähte des Dienstes** —
// #888.
//
// # Warum das ein zweiter Test ist
//
// Der oben hält, dass der **Abgleich** sie weiterreicht. Hier steht der Weg
// darunter: [service.Kartenaufsicht] prüft die Zugehörigkeit und reicht
// weiter, [service.Feldweg] zerlegt einen Verbund oder schreibt roh. Jede
// Naht kann für sich brechen, und eine gebrochene fiele sonst erst im
// Bestand auf — wo niemand mehr sieht, welche es war.
func TestDieHerkunftGehtDurchAufsichtUndFeldweg(t *testing.T) {
	t.Parallel()

	t.Run("die Aufsicht reicht sie weiter", func(t *testing.T) {
		t.Parallel()

		felder := &feldprobe{}
		aufsicht := service.NeueKartenaufsicht(
			&kartensicht{karten: map[int64]int64{2: 1}}, &lebenszyklusprotokoll{},
		).MitFeldweg(felder, felder)

		if _, _, fehler := aufsicht.FeldSetzenVonAussen(t.Context(), 1, 2,
			"rule text", "Fliegend.", "https://scryfall.invalid/x"); fehler != nil {
			t.Fatalf("FeldSetzenVonAussen: %v", fehler)
		}

		if len(felder.herkuenfte) != 1 || felder.herkuenfte[0] != "https://scryfall.invalid/x" {
			t.Errorf("herkuenfte = %v", felder.herkuenfte)
		}
	})

	t.Run("eine fremde Karte kommt nicht durch", func(t *testing.T) {
		t.Parallel()

		felder := &feldprobe{}
		aufsicht := service.NeueKartenaufsicht(
			&kartensicht{karten: map[int64]int64{2: 99}}, &lebenszyklusprotokoll{},
		).MitFeldweg(felder, felder)

		if _, _, fehler := aufsicht.FeldSetzenVonAussen(t.Context(), 1, 2,
			"rule text", "Fliegend.", "https://scryfall.invalid/x"); fehler == nil {
			t.Fatal("die fremde Karte kam durch")
		}

		if len(felder.herkuenfte) != 0 {
			t.Errorf("es wurde trotzdem geschrieben: %v", felder.herkuenfte)
		}
	})
}

// **Und der Feldweg zerlegt einen übernommenen Verbund wie einen getippten.**
//
// # Warum das die eigentliche Zusage ist
//
// Ein Feld mit `save value: false` — `text` einer M15-Karte etwa — geht in
// seine Teile (#696). Ein Wert von Scryfall nimmt **denselben** Weg; würde er
// roh geschrieben, stünde er in einem Feld, das die Vorlage gar nicht liest,
// und die Karte bliebe leer. Dass beide Eingänge auf **einem** Rumpf sitzen,
// ist die Vorkehrung dagegen — dieser Test ist ihr Beleg.
func TestDerFeldwegZerlegtAuchEinenUebernommenenVerbund(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}
	weg := service.NeuerFeldweg(
		wegkarten{karte: &domain.CardRow{ID: 7, SetID: 3}},
		wegfelder{felder: m15felder()},
		ablage,
	)

	if _, _, fehler := weg.FeldSetzenVonAussen(t.Context(), 7, "text",
		"Fliegend.<sep><line>\n</line></sep>Es brennt.",
		"https://scryfall.invalid/x"); fehler != nil {
		t.Fatalf("FeldSetzenVonAussen: %v", fehler)
	}

	if ablage.reihen != 1 {
		t.Errorf("Reihen = %d, erwartet eine Transaktion", ablage.reihen)
	}

	if len(ablage.herkuenfte) != 1 || ablage.herkuenfte[0] != "https://scryfall.invalid/x" {
		t.Errorf("herkuenfte = %v — die Teile eines Verbunds tragen sie nicht",
			ablage.herkuenfte)
	}
}
