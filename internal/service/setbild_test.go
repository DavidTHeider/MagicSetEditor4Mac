package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// kunstvorlage baut eine Vorlage mit **einem** Bildfeld namens `kunst`.
//
// `standard` ist das `default:` des Feldstils — leer heißt „der Stil nennt
// keines", und das ist der Fall, den [TestOhneStandardbildZeichnetDieKarteTrotzdem]
// prüft.
func kunstvorlage(setpaket, standard string) service.Vorlage {
	feld := flaeche("kunst")
	feld.Darstellung = "image"
	feld.Standardbild = standard

	return service.Vorlage{
		Stil:         bildstil(feld),
		Setpaket:     setpaket,
		Kartenfelder: []domain.Field{{Name: "kunst", Kind: domain.KindImage}},
	}
}

// kunstkarte ist eine Karte, die für `kunst` einen Dateinamen führt.
func kunstkarte(datei string) *domain.CardRow {
	return &domain.CardRow{ID: 1, Values: []domain.Value{{Field: "kunst", Text: datei}}}
}

// kunstplan baut den Plan und verlangt, dass es keinen Fehler gab.
func kunstplan(
	t *testing.T, karte *domain.CardRow, vorlage service.Vorlage,
) (*render.Plan, []domain.Befund) {
	t.Helper()

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	return plan, befunde
}

// **Das Bild der Karte kommt aus dem Set-Paket.**
//
// Bis #380
// stand hier ein Befund und das Ersatzbild — das Kunstfenster **jeder** Karte
// blieb damit leer.
func TestDasBildDerKarteKommtAusDemSetpaket(t *testing.T) {
	vorlage := kunstvorlage("/Orte/Mein Set.mse-set", "ersatz.png")

	plan, befunde := kunstplan(t, kunstkarte("image1"), vorlage)

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "noch nicht gedeutet") {
			t.Errorf("die alte Meldung steht noch da: %s", befund.Meldung)
		}
	}

	gezeichnet := ebenen(t, plan)
	if len(gezeichnet) != 1 {
		t.Fatalf("%d Ebenen, erwartet eine", len(gezeichnet))
	}

	erwartet := msepaket.SetpaketVerweis("/Orte/Mein Set.mse-set", "image1")
	if gezeichnet[0].Bild.Pfad != erwartet {
		t.Errorf("Pfad %q, erwartet %q", gezeichnet[0].Bild.Pfad, erwartet)
	}
}

// **Und es wird gestreckt, nicht eingepasst.**
//
//	if (image_load_file(image, *image_file)) { image.Rescale(w, h); }
//
// `ImageValueViewer::draw`, `src/render/value/image.cpp`, Zeilen 36–39,
// Fassung f9d9356 vom 2021-07-06. Eine Einpassung ließe zwei Streifen des
// Kartenrahmens durchscheinen, wo MSE die Illustration zieht.
func TestDasBildDerKarteWirdGestreckt(t *testing.T) {
	plan, _ := kunstplan(t, kunstkarte("image1"), kunstvorlage("/Orte/S.mse-set", ""))

	gezeichnet := ebenen(t, plan)
	if len(gezeichnet) != 1 {
		t.Fatalf("%d Ebenen, erwartet eine", len(gezeichnet))
	}

	if gezeichnet[0].Passung != nil {
		t.Errorf("Passung %+v, erwartet keine — ein Bildfeld streckt", *gezeichnet[0].Passung)
	}
}

// **Ohne `default:` zeichnet die Karte ihr Bild trotzdem.**
//
// Die Klasse hinter dem Fall: `default:` ist der **Marker** eines Feldes, an
// dem die Karte ihr Bild mitbringt — aber nicht seine Bedingung. Ein Stil, der
// kein Ersatzbild nennt, ließ das Kunstfenster sonst leer, weil die Bildebene
// gar nicht erst gebaut wurde.
func TestOhneStandardbildZeichnetDieKarteTrotzdem(t *testing.T) {
	plan, befunde := kunstplan(t, kunstkarte("image1"), kunstvorlage("/Orte/S.mse-set", ""))

	gezeichnet := ebenen(t, plan)
	if len(gezeichnet) != 1 {
		t.Fatalf("%d Ebenen, erwartet eine; Befunde: %v", len(gezeichnet), befunde)
	}

	if gezeichnet[0].Art != render.ArtBild {
		t.Errorf("Art %q, erwartet %q", gezeichnet[0].Art, render.ArtBild)
	}
}

// **`default:` gilt weiter, wenn die Karte kein Bild führt.**
//
//	if (!image.Ok() && style().default_image.isReady()) { … }
//
// Dieselbe Fundstelle, Zeilen 41–46: das Ersatzbild ist das, was **fehlt** —
// nicht das, was gilt.
func TestOhneEigenesBildGiltDasStandardbild(t *testing.T) {
	plan, befunde := kunstplan(t, kunstkarte(""), kunstvorlage("/Orte/S.mse-set", "ersatz.png"))

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	gezeichnet := ebenen(t, plan)
	if len(gezeichnet) != 1 {
		t.Fatalf("%d Ebenen, erwartet eine", len(gezeichnet))
	}

	if gezeichnet[0].Bild.Pfad != "/probe.mse-style/ersatz.png" {
		t.Errorf("Pfad %q, erwartet das Ersatzbild", gezeichnet[0].Bild.Pfad)
	}
}

// **Ein Set ohne bekannten Ort ist ein Befund, kein stilles Ersatzbild.**
//
// Der Ort steht in `domain.SetHead.Origin`. Fehlt er, ist das Bild der Karte
// nicht aufzulösen — und das gehört gesagt (§8), sonst sieht die Karte aus wie
// eine ohne Illustration.
func TestOhneSetpaketBleibtEinBefund(t *testing.T) {
	plan, befunde := kunstplan(t, kunstkarte("image1"), kunstvorlage("", "ersatz.png"))

	if len(befunde) == 0 {
		t.Fatal("kein Befund, obwohl der Ort des Sets fehlt")
	}

	gezeichnet := ebenen(t, plan)
	if len(gezeichnet) != 1 || gezeichnet[0].Bild.Pfad != "/probe.mse-style/ersatz.png" {
		t.Errorf("gezeichnet wurde %+v, erwartet das Ersatzbild", gezeichnet)
	}
}

// **Der Ort des Sets kommt aus dem Bestand bis zur Vorlage.**
//
// Die Naht: `domain.SetHead.Origin` steht seit dem Import da, und ohne diesen
// Schritt trägt die Vorlage ihn nicht — dann bliebe das Kunstfenster leer,
// obwohl alles andere steht (§10, die Nähte).
func TestDieVorlageTraegtDenOrtDesSets(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{
		{ID: 7, Style: "probe-schlicht", Game: "probe", Origin: "/Orte/Mein Set.mse-set"},
	}}

	vorlage, befunde, da := service.NeuerStilspeicher(quelle, saatwurzel()).
		StilFuerSet(t.Context(), 7)
	if !da {
		t.Fatalf("kein Stil gefunden, Befunde: %v", befunde)
	}

	if vorlage.Setpaket != "/Orte/Mein Set.mse-set" {
		t.Errorf("Setpaket %q, erwartet den Ort aus dem Setkopf", vorlage.Setpaket)
	}
}
