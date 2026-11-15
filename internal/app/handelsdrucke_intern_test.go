package app

// **Die Karte einer Handelszeile an der Naht** — P10
// (#2378).

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// derEigeneDruck ist ein Druck des eigenen Bestands mit Scryfall-Alias.
const derEigeneDruck = "00000000-0000-4000-8000-0000000000a1"

// derGeholteDruck ist ein Druck, den Scryfall schon einmal beantwortet hat.
const derGeholteDruck = "00000000-0000-4000-8000-0000000000b1"

// einGeholterDruck legt eine Antwort von `/cards/:id` in den Zwischenspeicher
// dieses Zuhauses — über den Klienten und damit unter seinem Schlüssel.
func einGeholterDruck(t *testing.T) {
	t.Helper()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, fehler := schreiben.Write([]byte(`{"object":"card","id":"` + derGeholteDruck +
				`","name":"Sturmrufer","lang":"en","layout":"normal","set":"zwe","set_name":"Zweite",` +
				`"collector_number":"3","rarity":"rare",` +
				`"image_uris":{"small":"https://cards.scryfall.io/small/sturmrufer.jpg"}}`)); fehler != nil {
				t.Errorf("schreiben: %v", fehler)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	ort, fehler := Scryfallverzeichnis()
	if fehler != nil {
		t.Fatal(fehler)
	}

	if _, _, _, fehler := scryfall.NeuerKlient(gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()),
		netz.NeuerSpeicher(ort, scryfall.Speicherfrist, time.Now), time.Now, "prüfung").
		KarteNachKennung(t.Context(), derGeholteDruck); fehler != nil {
		t.Fatalf("Druck holen: %v", fehler)
	}
}

func TestDieZeilenDesAnhaltsNennenIhreKarte(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	anwendung, ich := eineLeitungsflaeche(t)
	ctx := riegelkontext(t)

	einGeholterDruck(t)

	if fehler := storage.NeueDrucke(anwendung.db).Sichern(anwendung.imProfil(ctx), &domain.Druck{
		Name: "Blitzschlag", Setcode: "m10", Nummer: "146", Sprache: "de",
		Kennungen: []domain.Druckkennung{{
			System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: derEigeneDruck,
		}},
	}); fehler != nil {
		t.Fatalf("Druck sichern: %v", fehler)
	}

	antwort := anwendung.HandelPreisanhalt(ctx, []HandelspositionDTO{
		{Von: ich, Printing: derEigeneDruck, Menge: 2},
		{Von: dasLeitungsgegenueber, Printing: derGeholteDruck, Menge: 1},
		{Von: dasLeitungsgegenueber, DruckID: 1, Menge: 1},
	})
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	erwartet := []HandelsdruckDTO{
		{Name: "Blitzschlag", Setcode: "m10", Nummer: "146"},
		{Name: "Sturmrufer", Setcode: "zwe", Nummer: "3", Bild: "https://cards.scryfall.io/small/sturmrufer.jpg"},
		{Eigen: true},
	}

	for stelle, zeile := range antwort.Anhalt.Zeilen {
		if zeile.Druck != erwartet[stelle] {
			t.Errorf("Zeile %d: %+v, erwartet %+v", stelle+1, zeile.Druck, erwartet[stelle])
		}
	}

	// **Auch über einen Vorgang**, also an Kommandozeile, Dienst und MCP.
	if angelegt := anwendung.HandelEroeffnen(ctx, einLeitungswunsch()); angelegt.Fehler != nil {
		t.Fatalf("Eröffnen: %+v", angelegt.Fehler)
	}

	if von := anwendung.HandelPreisanhaltVon(ctx, derLeitungsvorgang); von.Fehler != nil ||
		len(von.Anhalt.Zeilen) != 1 || von.Anhalt.Zeilen[0].Druck.Eigen {
		t.Fatalf("über den Vorgang: %+v", von)
	}
}

// **Ohne Zwischenspeicher bleiben Namen** — nur das Bild fehlt.
func TestOhneZuhauseNenntDerAnhaltDenEigenenDruck(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	anwendung, ich := eineLeitungsflaeche(t)
	ctx := riegelkontext(t)

	if fehler := storage.NeueDrucke(anwendung.db).Sichern(anwendung.imProfil(ctx), &domain.Druck{
		Name: "Blitzschlag", Setcode: "m10", Nummer: "146",
		Kennungen: []domain.Druckkennung{{
			System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: derEigeneDruck,
		}},
	}); fehler != nil {
		t.Fatalf("Druck sichern: %v", fehler)
	}

	t.Setenv("HOME", "")

	antwort := anwendung.HandelPreisanhalt(ctx, []HandelspositionDTO{{Von: ich, Printing: derEigeneDruck, Menge: 1}})
	if antwort.Fehler != nil || antwort.Anhalt.Zeilen[0].Druck != (HandelsdruckDTO{
		Name: "Blitzschlag", Setcode: "m10", Nummer: "146",
	}) {
		t.Fatalf("Antwort: %+v", antwort)
	}
}

// **Ein Lesefehler der Drucke ist eine Absage** — mit leeren Listen.
func TestEinFehlenderDruckbestandIstEineAbsage(t *testing.T) {
	anwendung, ich := eineLeitungsflaeche(t)

	if _, fehler := anwendung.db.ExecContext(t.Context(), `DROP TABLE druckkennungen`); fehler != nil {
		t.Fatal(fehler)
	}

	antwort := anwendung.HandelPreisanhalt(riegelkontext(t), []HandelspositionDTO{
		{Von: ich, Printing: "bolt-m10", Menge: 1},
	})
	if antwort.Fehler == nil || antwort.Anhalt.Zeilen == nil {
		t.Fatalf("Antwort: %+v", antwort)
	}
}
