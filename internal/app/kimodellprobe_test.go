package app_test

// **Ein Modell messen, statt seinem Anbieter zu glauben** — Werkstück A9 von
// Phase 5.3
// (#1448).
//
// # Der Weg, um den es geht
//
// **Ohne Messung ist OpenAI ein Schalter, der nichts tut.** `/v1/models`
// nennt keine Fähigkeiten, ohne Beleg gehört ein Modell nicht in den Katalog
// (0097), und ohne Katalogeintrag lässt sich keiner wählen. Die Probe ist
// der dritte Weg, und dieser Lauf fährt ihn ganz.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// derProbealias ist das Modell, das diese Läufe messen.
//
// **Einer und keiner mehr**: welches Modell gemessen wird, ist für keinen
// dieser Läufe die Frage — sie prüfen, **ob** gemessen werden darf und was
// danach feststeht.
const derProbealias = "gpt-probe"

// einModellOhneBeleg ist, was ein Anbieter meldet, der nichts belegt.
func einModellOhneBeleg() domain.KIModell {
	return domain.KIModell{
		Alias: derProbealias, Anbieter: "openai", Ort: domain.KIOrtCloud,
		Laufzeit: "Responses-API /v1/responses",
	}
}

// eineOpenAIWelt baut eine Anwendung mit eingeschaltetem OpenAI.
func eineOpenAIWelt(t *testing.T, quelle app.KILaufzeitquelle) kiwelt {
	t.Helper()

	welt := einekiwelt(t)
	welt.anwendung.MitKIAnbietern(
		func(string) app.KILaufzeitquelle { return welt.laufzeit },
		map[string]func(adresse string) app.KILaufzeitquelle{
			"openai": func(string) app.KILaufzeitquelle { return quelle },
		})

	if fehler := storage.NeueEinstellungen(welt.db).Setzen(probekontext(t),
		domain.DienstEinstellung(domain.DienstOpenAI),
		domain.DienstAn); fehler != nil {
		t.Fatalf("einschalten: %v", fehler)
	}

	welt.anwendung.MitKosten(stummeskijournal{})
	welt.anwendung.MitKIFreigabe(
		"der Mensch hat zugestimmt", domain.Betrag{Wert: 100, Waehrung: "EUR"})

	return welt
}

// **Ein gemessenes Modell steht danach im Katalog — und vorher nicht.**
//
// **Beide Hälften in einem Lauf**, denn die zweite allein sagte nichts: ein
// Katalog mit einem fähigen Modell könnte es schon vorher gewesen sein.
func TestEinGemessenesModellStehtDanachImKatalog(t *testing.T) {
	t.Parallel()

	quelle := &laufzeitdoppel{
		fassung: "Responses-API /v1/responses",
		modelle: []domain.KIModell{einModellOhneBeleg()},
		antwort: app.KIRohantwort{
			Inhalt: `{"probe": "bereit"}`, Modell: "gpt-probe-2026-09-01",
		},
	}

	welt := eineOpenAIWelt(t, quelle)

	vorher := welt.anwendung.KIModelle(probekontext(t))
	if vorher.Fehler != nil {
		t.Fatalf("%+v", vorher.Fehler)
	}

	for _, modell := range vorher.Modelle {
		if modell.Anbieter == "openai" && len(modell.Faehigkeiten) > 0 {
			t.Fatalf("vor der Messung steht schon etwas fest: %+v", modell)
		}
	}

	probe := welt.anwendung.KIModellPruefen(probekontext(t), "openai", "gpt-probe", "")
	if probe.Fehler != nil {
		t.Fatalf("%+v", probe.Fehler)
	}

	// **Zwei Fähigkeiten**: Text und feste Form.
	if len(probe.Proben) != 2 {
		t.Fatalf("die Messungen sind %+v", probe.Proben)
	}

	nachher := welt.anwendung.KIModelle(probekontext(t))
	if nachher.Fehler != nil {
		t.Fatalf("%+v", nachher.Fehler)
	}

	for _, modell := range nachher.Modelle {
		if modell.Anbieter != "openai" {
			continue
		}

		if len(modell.Faehigkeiten) != 2 {
			t.Fatalf("nach der Messung: %+v", modell)
		}

		// **Der Beleg reist mit** — „geprüft" ohne das Wie ist eine
		// Behauptung (§10).
		if modell.Faehigkeiten[0].Beleg == "" || modell.Faehigkeiten[0].Datum == "" {
			t.Errorf("der Beleg ist %+v", modell.Faehigkeiten[0])
		}

		return
	}

	t.Fatal("das Modell steht nicht im Katalog")
}

// **Und dann lässt sich damit auch vorschlagen.**
//
// **Das ist die Probe auf den ganzen Weg** (§10, erste Dimension): ohne sie
// sagte der Lauf darüber nur, dass eine Zeile im Katalog steht — nicht, dass
// jemand sie benutzen kann.
func TestNachDerMessungLaesstSichVorschlagen(t *testing.T) {
	t.Parallel()

	quelle := &laufzeitdoppel{
		fassung: "Responses-API /v1/responses",
		modelle: []domain.KIModell{einModellOhneBeleg()},
		antwort: app.KIRohantwort{
			Inhalt: `{"probe": "bereit"}`, Modell: "gpt-probe-2026-09-01",
		},
	}

	welt := eineOpenAIWelt(t, quelle)

	// **Vorher geht es nicht** — und zwar, weil nichts belegt ist.
	if antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"openai", "gpt-probe", "Schreibe.", ""); antwort.Fehler == nil {
		t.Fatal("ein ungeprüftes Modell hat vorgeschlagen")
	}

	if probe := welt.anwendung.KIModellPruefen(
		probekontext(t), "openai", "gpt-probe", ""); probe.Fehler != nil {
		t.Fatalf("%+v", probe.Fehler)
	}

	quelle.antwort = app.KIRohantwort{
		Inhalt: `{"name": "Blitzschlag"}`, Modell: "gpt-probe-2026-09-01",
	}

	antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"openai", "gpt-probe", "Schreibe.", "")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Anbieter != "openai" || len(antwort.Felder) == 0 {
		t.Errorf("die Antwort ist %+v", antwort.KIVorschlagDTO)
	}
}

// **Eine Messung, die die Form nicht hält, belegt nichts.**
func TestEineMessungOhneFormBelegtNichts(t *testing.T) {
	t.Parallel()

	welt := eineOpenAIWelt(t, &laufzeitdoppel{
		fassung: "Responses-API /v1/responses",
		modelle: []domain.KIModell{einModellOhneBeleg()},
		antwort: app.KIRohantwort{Inhalt: "gerne doch!", Modell: "gpt-probe"},
	})

	if probe := welt.anwendung.KIModellPruefen(
		probekontext(t), "openai", "gpt-probe", ""); probe.Fehler == nil {
		t.Fatal("die formlose Antwort galt als Messung")
	}

	katalog := welt.anwendung.KIModelle(probekontext(t))
	for _, modell := range katalog.Modelle {
		if modell.Anbieter == "openai" && len(modell.Faehigkeiten) > 0 {
			t.Errorf("trotzdem belegt: %+v", modell)
		}
	}
}

// **Eine vergessene Messung ist aus dem Katalog verschwunden.**
func TestEineVergesseneMessungIstAusDemKatalogFort(t *testing.T) {
	t.Parallel()

	welt := eineOpenAIWelt(t, &laufzeitdoppel{
		fassung: "Responses-API /v1/responses",
		modelle: []domain.KIModell{einModellOhneBeleg()},
		antwort: app.KIRohantwort{
			Inhalt: `{"probe": "bereit"}`, Modell: "gpt-probe-2026-09-01",
		},
	})

	if probe := welt.anwendung.KIModellPruefen(
		probekontext(t), "openai", "gpt-probe", ""); probe.Fehler != nil {
		t.Fatalf("%+v", probe.Fehler)
	}

	if fort := welt.anwendung.KIModellVergessen(
		probekontext(t), "openai", "gpt-probe"); fort.Fehler != nil {
		t.Fatalf("%+v", fort.Fehler)
	}

	katalog := welt.anwendung.KIModelle(probekontext(t))
	for _, modell := range katalog.Modelle {
		if modell.Anbieter == "openai" && len(modell.Faehigkeiten) > 0 {
			t.Errorf("die vergessene Messung steht noch da: %+v", modell)
		}
	}
}

// **Ohne Freigabe wird gar nicht erst gemessen.**
//
// **Eine Messung ist eine Anfrage und keine Ausnahme** (A8): sie kostet
// Geld, und ohne Journal und Nachweis geht sie nicht hinaus.
func TestOhneFreigabeWirdNichtGemessen(t *testing.T) {
	t.Parallel()

	quelle := &laufzeitdoppel{
		fassung: "Responses-API /v1/responses",
		modelle: []domain.KIModell{einModellOhneBeleg()},
		antwort: app.KIRohantwort{Inhalt: `{"probe": "bereit"}`},
	}

	welt := einekiwelt(t)
	welt.anwendung.MitKIAnbietern(
		func(string) app.KILaufzeitquelle { return welt.laufzeit },
		map[string]func(adresse string) app.KILaufzeitquelle{
			"openai": func(string) app.KILaufzeitquelle { return quelle },
		})

	if fehler := storage.NeueEinstellungen(welt.db).Setzen(probekontext(t),
		domain.DienstEinstellung(domain.DienstOpenAI),
		domain.DienstAn); fehler != nil {
		t.Fatalf("einschalten: %v", fehler)
	}

	welt.anwendung.MitKosten(stummeskijournal{})

	if probe := welt.anwendung.KIModellPruefen(
		probekontext(t), "openai", "gpt-probe", ""); probe.Fehler == nil {
		t.Fatal("ohne Freigabe wurde gemessen")
	}

	// **Die Gegenprobe zur Zusage**: null Anfragen heißt null.
	if len(quelle.schemen) != 0 {
		t.Errorf("OpenAI bekam %d Fragen", len(quelle.schemen))
	}
}

// **Eine Messung mit einer erfundenen Freigabe geht nicht hinaus.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil die Messung ihren Höchstbetrag aus dem Nachweis liest** (A8, und
// seit #1539
// aus dem echten): eine Kennung, die es nicht gibt, hat keinen — und dann
// wird nichts gebucht und nichts gefragt. Das Issue verlangt es als eigene
// Zeile: „Unter Local-only, fehlendem Scope oder Budget entstehen null
// Anbieteranfragen."
func TestEineMessungMitErfundenerFreigabeGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	quelle := &laufzeitdoppel{
		fassung: "Responses-API /v1/responses",
		modelle: []domain.KIModell{einModellOhneBeleg()},
	}

	welt := eineOpenAIWelt(t, quelle)

	// **Die Freigabe des Fensters weg** — sonst nähme `budgetFuer` sie, und
	// die mitgegebene Kennung käme nie zur Ablage.
	welt.anwendung.MitKIFreigabe("", domain.Betrag{})

	antwort := welt.anwendung.KIModellPruefen(probekontext(t), "openai",
		derProbealias, "gibtesnicht")
	if antwort.Fehler == nil {
		t.Fatal("die Messung ging mit einer erfundenen Freigabe hinaus")
	}

	// **Und null Anbieteranfragen** — das ist die Abnahme, nicht der Fehler.
	if len(quelle.adressen) != 0 {
		t.Errorf("gefragt wurde %v", quelle.adressen)
	}
}
