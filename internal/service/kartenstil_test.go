package service_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Karte darf ihren eigenen Stil tragen.**
//
//	const StyleSheet& Set::stylesheetFor(const CardP& card) {
//	  if (card && card->stylesheet) return *card->stylesheet;
//	  else                          return *stylesheet;
//	}
//
// `src/data/set.cpp`, Zeilen 90–93, Fassung f9d9356 vom 2021-07-06. Der Stil
// der Karte gewinnt, sonst gilt der des Sets.
//
// **Woran es aufgefallen ist:** in `Archenemy.mse-set` — einem kanonischen
// MSE2-Set — steht `stylesheet: archenemy` **70-mal an den Karten** und
// `m15-altered` einmal am Set. Alle 70 Karten wurden über die falsche Vorlage
// gezeichnet (#395).
func TestEineKarteTraegtIhrenEigenenStil(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{
		{ID: 7, Style: "probe-schlicht", Game: "probe"},
	}}

	speicher := service.NeuerStilspeicher(quelle, saatwurzel())

	karte := &domain.CardRow{ID: 1, SetID: 7, Values: []domain.Value{
		{Field: "stylesheet", Text: "probe-bild"},
	}}

	vorlage, befunde, da := speicher.StilFuerKarte(t.Context(), karte)
	if !da {
		t.Fatalf("kein Stil gefunden, Befunde: %v", befunde)
	}

	// `probe-bild` führt ein Feld `rahmen`, `probe-schlicht` nicht.
	if _, hat := vorlage.Stil.Feld("rahmen"); !hat {
		t.Errorf("der Stil führt kein Feld `rahmen` — es ist der des Sets, nicht der der Karte")
	}
}

// **Ohne eigenen Stil gilt der des Sets** — die zweite Hälfte derselben Regel.
func TestOhneEigenenStilGiltDerDesSets(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{
		{ID: 7, Style: "probe-bild", Game: "probe"},
	}}

	speicher := service.NeuerStilspeicher(quelle, saatwurzel())

	karte := &domain.CardRow{ID: 1, SetID: 7}

	vorlage, befunde, da := speicher.StilFuerKarte(t.Context(), karte)
	if !da {
		t.Fatalf("kein Stil gefunden, Befunde: %v", befunde)
	}

	if _, hat := vorlage.Stil.Feld("rahmen"); !hat {
		t.Error("der Stil des Sets ist nicht angekommen")
	}
}

// **Die Setangaben bleiben die des Sets**, auch wenn die Karte ihren eigenen
// Stil trägt: `set.symbol` und `set.copyright` hängen am Set, nicht am
// Stylesheet.
func TestDerEigeneStilBehaeltDieSetangaben(t *testing.T) {
	quelle := &koepfequelle{
		koepfe: []domain.SetHead{{
			ID: 7, Style: "probe-schlicht", Game: "probe",
			Origin: "/Orte/Mein Set.mse-set",
		}},
		werte: []domain.Value{{Field: "title", Text: "Probeserie"}},
	}

	karte := &domain.CardRow{ID: 1, SetID: 7, Values: []domain.Value{
		{Field: "stylesheet", Text: "probe-bild"},
	}}

	vorlage, _, da := service.NeuerStilspeicher(quelle, saatwurzel()).
		StilFuerKarte(t.Context(), karte)
	if !da {
		t.Fatal("kein Stil gefunden")
	}

	if vorlage.Setpaket != "/Orte/Mein Set.mse-set" {
		t.Errorf("Setpaket %q", vorlage.Setpaket)
	}

	if len(vorlage.Setangaben) != 1 || vorlage.Setangaben[0].Text != "Probeserie" {
		t.Errorf("Setangaben %+v", vorlage.Setangaben)
	}
}

// **Ein zweiter Zugriff kommt aus dem Gedächtnis.**
//
// Ein Set hat wenige Stile und viele Karten; wer je Karte das Paket öffnete,
// zerteilte dieselbe Datei siebzigmal. Der Beleg ist die Zahl der Anfragen an
// den Bestand: der Stil wird gemerkt, die Setangaben nicht.
func TestDerKartenstilWirdGemerkt(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{
		{ID: 7, Style: "probe-schlicht", Game: "probe"},
	}}

	speicher := service.NeuerStilspeicher(quelle, saatwurzel())

	karte := &domain.CardRow{ID: 1, SetID: 7, Values: []domain.Value{
		{Field: "stylesheet", Text: "probe-bild"},
	}}

	for range 3 {
		if _, _, da := speicher.StilFuerKarte(t.Context(), karte); !da {
			t.Fatal("kein Stil gefunden")
		}
	}

	if quelle.gefragt != 3 {
		t.Errorf("%d Anfragen an den Bestand, erwartet drei — je Aufruf eine für den Setkopf",
			quelle.gefragt)
	}
}

// **Ein Stil, den es nicht gibt, wird benannt** (§8) — und die Karte bekommt
// den Behelfsplan, keinen Absturz.
func TestEinUnbekannterKartenstilWirdBenannt(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{
		{ID: 7, Style: "probe-schlicht", Game: "probe"},
	}}

	karte := &domain.CardRow{ID: 1, SetID: 7, Values: []domain.Value{
		{Field: "stylesheet", Text: "gibtesnicht"},
	}}

	_, befunde, da := service.NeuerStilspeicher(quelle, saatwurzel()).
		StilFuerKarte(t.Context(), karte)
	if da {
		t.Fatal("ein Stil, den es nicht gibt, wurde gefunden")
	}

	if len(befunde) == 0 {
		t.Error("kein Befund")
	}
}

// **Ohne Paketverzeichnis wird das gesagt, nicht verschwiegen.**
func TestOhneVerzeichnisNenntDerKartenstilDenGrund(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{
		{ID: 7, Style: "probe-schlicht", Game: "probe"},
	}}

	karte := &domain.CardRow{ID: 1, SetID: 7, Values: []domain.Value{
		{Field: "stylesheet", Text: "probe-bild"},
	}}

	_, befunde, da := service.NeuerStilspeicher(quelle).StilFuerKarte(t.Context(), karte)
	if da || len(befunde) == 0 {
		t.Fatalf("da=%v, Befunde: %v", da, befunde)
	}

	if !strings.Contains(befunde[0].Meldung, "Paketverzeichnis") {
		t.Errorf("Befund %+v", befunde[0])
	}
}

// **Ein Set, das der Bestand nicht führt, ebenso.**
func TestEinUnbekanntesSetNenntDenGrundAuchBeimKartenstil(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{}}

	karte := &domain.CardRow{ID: 1, SetID: 7, Values: []domain.Value{
		{Field: "stylesheet", Text: "probe-bild"},
	}}

	_, befunde, da := service.NeuerStilspeicher(quelle, saatwurzel()).
		StilFuerKarte(t.Context(), karte)
	if da || len(befunde) == 0 {
		t.Fatalf("da=%v, Befunde: %v", da, befunde)
	}
}

// **Und ein Bestand, der gerade nicht lesbar ist.**
func TestEinUnlesbarerBestandNenntDenGrundBeimKartenstil(t *testing.T) {
	quelle := &koepfequelle{fehler: errKaputt}

	karte := &domain.CardRow{ID: 1, SetID: 7, Values: []domain.Value{
		{Field: "stylesheet", Text: "probe-bild"},
	}}

	_, befunde, da := service.NeuerStilspeicher(quelle, saatwurzel()).
		StilFuerKarte(t.Context(), karte)
	if da || len(befunde) == 0 {
		t.Fatalf("da=%v, Befunde: %v", da, befunde)
	}
}

// **Ein Fehler bei den Setangaben hält den Stil nicht auf** (§8).
func TestEinFehlerBeiDenSetangabenHaeltDenKartenstilNichtAuf(t *testing.T) {
	quelle := &koepfequelle{
		koepfe:      []domain.SetHead{{ID: 7, Style: "probe-schlicht", Game: "probe"}},
		wertefehler: errKaputt,
	}

	karte := &domain.CardRow{ID: 1, SetID: 7, Values: []domain.Value{
		{Field: "stylesheet", Text: "probe-bild"},
	}}

	_, befunde, da := service.NeuerStilspeicher(quelle, saatwurzel()).
		StilFuerKarte(t.Context(), karte)
	if !da {
		t.Fatal("der Stil fiel mit den Setangaben aus")
	}

	if len(befunde) == 0 {
		t.Error("kein Befund über die Setangaben")
	}
}

// errKaputt ist der Fehler, den ein Bestand meldet, der nicht lesbar ist.
var errKaputt = errors.New("kaputt")
