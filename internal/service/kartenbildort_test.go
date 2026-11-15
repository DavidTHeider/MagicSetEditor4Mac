package service_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein eingesetztes Bild wird auch gefunden** —
// #246.
//
// # Wofür
//
// [0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md) legt
// eingesetzte Bilder **neben** das Set. Der Rasterer sucht sie im
// Herkunftspaket (`ImageValueViewer::draw` fragt `getLocalPackage()`,
// `src/render/value/image.cpp`, Zeilen 30–33, Fassung f9d9356 vom
// 2021-07-06) — dort liegen sie nicht, und ohne diese Naht zeichnete jede
// gerade eingesetzte Illustration das Ersatzbild.
//
// # Die Ablage schlägt das Herkunftspaket
//
// Wer ein Bild einsetzt, ersetzt damit das, was vorher dastand. Läge das
// Herkunftspaket vorn, zeigte die Karte nach dem Einsetzen weiter das alte
// Bild — bis zum Speichern, und danach nie wieder.
func TestEinEingesetztesBildSchlaegtDasHerkunftspaket(t *testing.T) {
	ablage := t.TempDir()
	if err := os.WriteFile(filepath.Join(ablage, "image1"), []byte("nicht wirklich ein PNG"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	vorlage := bildvorlage()
	vorlage.Setpaket = "/gibt/es/nicht.mse-set"
	vorlage.Bilderverzeichnis = ablage

	karte := &domain.CardRow{ID: 1, Values: []domain.Value{{Field: "kunst", Text: "image1"}}}

	plan, _, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	pfad := ersterBildpfad(t, plan)
	if !strings.Contains(pfad, ablage) {
		t.Errorf("der Verweis zeigt auf %q statt in die Ablage %q", pfad, ablage)
	}
}

// **Und was dort nicht liegt, kommt weiter aus dem Herkunftspaket.**
//
// Die Bilder eines geöffneten Sets liegen dort und nirgends sonst; sie in die
// Ablage zu kopieren wären bei `Multicolor.mse-set` 229 MB, nur damit sie an
// einem zweiten Ort noch einmal stehen.
func TestEinNichtEingesetztesBildKommtAusDemHerkunftspaket(t *testing.T) {
	vorlage := bildvorlage()
	vorlage.Setpaket = "/irgendwo/Mein Set.mse-set"
	vorlage.Bilderverzeichnis = t.TempDir()

	karte := &domain.CardRow{ID: 1, Values: []domain.Value{{Field: "kunst", Text: "image1"}}}

	plan, _, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	pfad := ersterBildpfad(t, plan)
	if !strings.Contains(pfad, "Mein Set.mse-set") {
		t.Errorf("der Verweis zeigt auf %q", pfad)
	}
}

// **Ohne Ablage bleibt alles, wie es war.** Ein Set, in dem niemand ein Bild
// eingesetzt hat, hat kein Verzeichnis — und keinen Zugriff auf die Platte.
func TestOhneAblageBleibtDerHerkunftsweg(t *testing.T) {
	vorlage := bildvorlage()
	vorlage.Setpaket = "/irgendwo/Mein Set.mse-set"

	karte := &domain.CardRow{ID: 1, Values: []domain.Value{{Field: "kunst", Text: "image1"}}}

	plan, _, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if pfad := ersterBildpfad(t, plan); !strings.Contains(pfad, "Mein Set.mse-set") {
		t.Errorf("der Verweis zeigt auf %q", pfad)
	}
}

// bildvorlage baut einen Stil mit einem Bildfeld, das die Karte füllt.
func bildvorlage() service.Vorlage {
	return service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 140,
			Felder: []msepaket.Feldstil{
				{Feld: "kunst", Links: "0", Oben: "0", Breite: "100", Hoehe: "80"},
			},
		},
		Kartenfelder: []domain.Field{{Name: "kunst", Kind: domain.KindImage}},
	}
}

// ersterBildpfad gibt den Pfad des ersten Bildblatts im Plan.
func ersterBildpfad(t *testing.T, plan *render.Plan) string {
	t.Helper()

	for i := range plan.Ebenen {
		if plan.Ebenen[i].Art == render.ArtBild {
			return plan.Ebenen[i].Bild.Pfad
		}
	}

	t.Fatalf("keine Bildebene: %+v", plan.Ebenen)

	return ""
}

// **Der Stilspeicher und der Setschreiber kennen die Ablage** —
// [0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md).
//
// # Warum beide und nicht einer
//
// Der Speicher sagt dem Planbauer, **wo er suchen soll**; der Schreiber gibt
// beim Speichern weiter, **was dort liegt**. Ohne den ersten zeichnete ein eben
// eingesetztes Bild das Ersatzbild, ohne den zweiten fehlte es im
// weitergegebenen Set — zwei verschiedene Verluste aus derselben Naht.
func TestSpeicherUndSchreiberKennenDieAblage(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)

	t.Run("der Stilspeicher trägt sie in die Vorlage ein", func(t *testing.T) {
		speicher := service.NeuerStilspeicher(
			&koepfequelle{koepfe: []domain.SetHead{
				{ID: 7, Style: "probe-schlicht", Game: "probe"},
			}}, saatwurzel()).MitBildablage(ablage)

		vorlage, _, _ := speicher.StilFuerSet(t.Context(), 7)
		if vorlage.Bilderverzeichnis != ablage.Verzeichnis(7) {
			t.Errorf("%q statt %q", vorlage.Bilderverzeichnis, ablage.Verzeichnis(7))
		}
	})

	t.Run("und ohne Ablage bleibt sie leer", func(t *testing.T) {
		speicher := service.NeuerStilspeicher(
			&koepfequelle{koepfe: []domain.SetHead{
				{ID: 7, Style: "probe-schlicht", Game: "probe"},
			}}, saatwurzel())

		if vorlage, _, _ := speicher.StilFuerSet(t.Context(), 7); vorlage.Bilderverzeichnis != "" {
			t.Errorf("%q", vorlage.Bilderverzeichnis)
		}
	})
}
