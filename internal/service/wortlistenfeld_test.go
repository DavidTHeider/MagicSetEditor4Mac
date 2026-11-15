package service_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// karteMitTyp baut eine Karte mit einem ausgezeichneten Typfeld.
func karteMitTyp(roh string) *domain.CardRow {
	return &domain.CardRow{
		ID: 7, SetID: 1,
		Values: []domain.Value{
			{Field: "name", Text: "Grizzly Bears"},
			{Field: "type", Text: roh},
		},
	}
}

// vorlageMitListen baut eine Vorlage, die nur die Wortlisten trägt.
func vorlageMitListen(listen ...msepaket.Wortliste) service.Vorlage {
	return service.Vorlage{Wortlisten: listen}
}

// **Ein Feld mit Marken gibt seine Bereiche samt Menü** —
// #1074.
//
// Das ist die Naht: die Stellen kommen aus dem Wert, die Listen aus dem
// Spielpaket, der Zustand aus beidem.
func TestEinFeldGibtSeineWortlistenbereiche(t *testing.T) {
	planer := service.NeuerPlanbauer()

	karte := karteMitTyp("<word-list-type-en>Legendary Creature</word-list-type-en>" +
		"<sep-soft> — </sep-soft><word-list-class-en>Bear</word-list-class-en>")

	vorlage := vorlageMitListen(
		msepaket.Wortliste{Name: "type-en", Woerter: typenliste()},
		msepaket.Wortliste{Name: "class-en", Woerter: []msepaket.Wortlisteneintrag{
			{Name: "Bear"}, {Name: "Elf"},
		}},
	)

	bereiche, befunde, err := planer.Wortlistenfeld(
		context.Background(), karte, "type", vorlage)
	if err != nil {
		t.Fatalf("Wortlistenfeld: %v", err)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde: %+v", befunde)
	}

	if len(bereiche) != 2 {
		t.Fatalf("%d Bereiche: %+v", len(bereiche), bereiche)
	}

	if bereiche[0].Liste != "type-en" || bereiche[0].Text != "Legendary Creature" {
		t.Errorf("der erste Bereich ist %+v", bereiche[0])
	}

	if !bereiche[0].Menue.Kaestchen {
		t.Error("die Typenliste braucht Kästchen")
	}

	if bereiche[0].Menue.Gewaehlt < 0 ||
		bereiche[0].Menue.Eintraege[bereiche[0].Menue.Gewaehlt].Name != "Creature" {
		t.Errorf("gewählt ist %d in %+v", bereiche[0].Menue.Gewaehlt, bereiche[0].Menue.Eintraege)
	}

	if bereiche[1].Liste != "class-en" || bereiche[1].Menue.Gewaehlt != 0 {
		t.Errorf("der zweite Bereich ist %+v", bereiche[1])
	}
}

// **Eine Marke ohne Liste ist ein Befund, kein Abbruch.**
//
//	queue_message(MESSAGE_WARNING, _ERROR_1_("word list type not found", name));
//
// `src/gui/value/text.cpp`, Zeile 1372, Fassung f9d9356 vom 2021-07-06. MSE
// meldet und macht weiter — der Bereich fällt weg, die anderen bleiben.
func TestEineMarkeOhneListeIstEinBefund(t *testing.T) {
	planer := service.NeuerPlanbauer()

	karte := karteMitTyp("<word-list-gibtsnicht>X</word-list-gibtsnicht>" +
		"<word-list-type-en>Creature</word-list-type-en>")

	bereiche, befunde, err := planer.Wortlistenfeld(context.Background(), karte, "type",
		vorlageMitListen(msepaket.Wortliste{Name: "type-en", Woerter: typenliste()}))
	if err != nil {
		t.Fatalf("Wortlistenfeld: %v", err)
	}

	if len(bereiche) != 1 || bereiche[0].Liste != "type-en" {
		t.Errorf("übrig sind %+v", bereiche)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde: %+v", len(befunde), befunde)
	}

	if befunde[0].Ort != "type" {
		t.Errorf("der Befund steht bei %q: %+v", befunde[0].Ort, befunde[0])
	}
}

// **Ein Feld ohne Marken gibt nichts** — und keinen Befund.
//
// Die allermeisten Felder haben keine Wortliste; eine Meldung je Feld ohne
// eine wäre eine Liste, die niemand liest (§8).
func TestEinFeldOhneMarkenGibtNichts(t *testing.T) {
	planer := service.NeuerPlanbauer()

	for _, fall := range []string{"Legendary Creature", ""} {
		bereiche, befunde, err := planer.Wortlistenfeld(context.Background(),
			karteMitTyp(fall), "type", vorlageMitListen())
		if err != nil || len(bereiche) != 0 || len(befunde) != 0 {
			t.Errorf("%q gab %+v / %+v / %v", fall, bereiche, befunde, err)
		}
	}
}

// **Ein Feld, das die Karte nicht führt, gibt nichts.**
func TestEinUnbekanntesFeldGibtNichts(t *testing.T) {
	planer := service.NeuerPlanbauer()

	bereiche, befunde, err := planer.Wortlistenfeld(context.Background(),
		karteMitTyp("<word-list-type-en>Creature</word-list-type-en>"),
		"gibtsnicht", vorlageMitListen())
	if err != nil || len(bereiche) != 0 || len(befunde) != 0 {
		t.Errorf("gab %+v / %+v / %v", bereiche, befunde, err)
	}
}

// **Ein Skript sieht die Karte.**
//
// `all_sub_types(setting: "is_land")` läuft in MSE im Kontext des Editors
// (`Context& ctx = tve.getContext();`, `src/gui/value/text.cpp`, Zeile 184,
// Fassung f9d9356 vom 2021-07-06). Ohne diese Umgebung gäbe jede skriptgebaute
// Liste dasselbe — nämlich nichts.
func TestEinSkriptImMenueSiehtDieKarte(t *testing.T) {
	planer := service.NeuerPlanbauer()

	vorlage := vorlageMitListen(msepaket.Wortliste{
		Name:    "land",
		Woerter: []msepaket.Wortlisteneintrag{{Skript: `"," + card.name`}},
	})

	bereiche, _, err := planer.Wortlistenfeld(context.Background(),
		karteMitTyp("<word-list-land></word-list-land>"), "type", vorlage)
	if err != nil {
		t.Fatalf("Wortlistenfeld: %v", err)
	}

	if len(bereiche) != 1 {
		t.Fatalf("%d Bereiche", len(bereiche))
	}

	if zeigeMenue(bereiche[0].Menue) != "|Grizzly Bears" {
		t.Errorf("das Menü ist %q", zeigeMenue(bereiche[0].Menue))
	}
}

// **Ein Skript, das scheitert, gibt einen Befund und keine Wörter.**
func TestEinGescheitertesSkriptGibtEinenBefund(t *testing.T) {
	planer := service.NeuerPlanbauer()

	vorlage := vorlageMitListen(msepaket.Wortliste{
		Name:    "land",
		Woerter: []msepaket.Wortlisteneintrag{{Skript: `gibtsnicht(`}},
	})

	bereiche, befunde, err := planer.Wortlistenfeld(context.Background(),
		karteMitTyp("<word-list-land></word-list-land>"), "type", vorlage)
	if err != nil {
		t.Fatalf("Wortlistenfeld: %v", err)
	}

	if len(bereiche) != 1 || len(bereiche[0].Menue.Eintraege) != 1 ||
		bereiche[0].Menue.Eintraege[0].Name != "" {
		t.Errorf("die Bereiche sind %+v", bereiche)
	}

	if len(befunde) != 1 || befunde[0].Ort != "type" {
		t.Errorf("die Befunde sind %+v", befunde)
	}
}

// **Die Marken stehen im gerechneten Wert, nicht im gespeicherten** —
// #1074.
//
// # Der Befund, für den dieser Lauf steht
//
// `magic.mse-game/card_fields` gibt dem Feld `super type` ein
// `script: super_type_filter(value, …)`, und **erst dieses Skript** legt
// `<word-list-type-en>…</word-list-type-en>` um den Text. Gemessen am
// 2026-08-24 an `Vanguard.mse-set`: gespeichert steht dort `Vanguard`, in der
// Umgebung `<word-list-type-en>Vanguard</…>` — und **keines** der zehn
// kanonischen Sets führt die Marken in der Datei.
//
// Der erste Anlauf las den gespeicherten Wert. Alle Tests waren grün, die
// Abdeckung voll — und auf **keiner echten Karte** wäre eine Auswahl
// erschienen (§10, erste Dimension).
func TestDieMarkenKommenAusDemWertskript(t *testing.T) {
	planer := service.NeuerPlanbauer()

	vorlage := vorlageMitListen(
		msepaket.Wortliste{Name: "type-en", Woerter: typenliste()})

	// **Das Feld rechnet seinen Wert** — wie im Pack, nur kürzer.
	vorlage.Kartenfelder = []domain.Field{{
		Name:   "type",
		Skript: `"<word-list-type-en>" + value + "</word-list-type-en>"`,
	}}

	// **Und gespeichert steht der blanke Text.**
	karte := karteMitTyp("Creature")

	bereiche, _, err := planer.Wortlistenfeld(
		context.Background(), karte, "type", vorlage)
	if err != nil {
		t.Fatalf("Wortlistenfeld: %v", err)
	}

	if len(bereiche) != 1 || bereiche[0].Text != "Creature" {
		t.Fatalf("die Bereiche sind %+v", bereiche)
	}

	if bereiche[0].Menue.Gewaehlt != 2 {
		t.Errorf("gewählt ist %d", bereiche[0].Menue.Gewaehlt)
	}

	// **Und geschrieben wird der gerechnete Wert samt Marken** — so tut MSE
	// es: dort ist der gerechnete Wert der Wert.
	neu, _, wahlfehler := planer.Wortwahl(
		context.Background(), karte, "type", vorlage, 0, []int{3})
	if wahlfehler != nil {
		t.Fatalf("Wortwahl: %v", wahlfehler)
	}

	if neu != "<word-list-type-en>Instant</word-list-type-en>" {
		t.Errorf("daraus wird %q", neu)
	}
}
