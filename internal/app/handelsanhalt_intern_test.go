package app

// **Der Preisanhalt an der Naht** — TRADE-02b
// (#1951).

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// einPreis legt einen genannten Verkaufspreis in den Bestand.
func einPreis(t *testing.T, anwendung *App, printing, waehrung string, wert int64) {
	t.Helper()

	stand := domain.Preisstand{
		Printing: printing, Quelle: domain.Preisquelle("scryfall"),
		Markt: domain.MarktPapier, Art: domain.PreisartVerkauf, Lage: domain.PreisGenannt,
		Waehrung: waehrung, Finish: domain.FinishNonfoil, Wert: wert,
		Erhoben: time.Date(2026, 9, 20, 8, 0, 0, 0, time.UTC),
	}

	if fehler := storage.NeuerPreisspeicher(anwendung.db).Eintragen(
		anwendung.imProfil(riegelkontext(t)), stand); fehler != nil {
		t.Fatalf("Preis eintragen: %v", fehler)
	}
}

func TestDerAnhaltEinesEntwurfsKommtAnDerNahtAn(t *testing.T) {
	anwendung, ich := eineLeitungsflaeche(t)
	ctx := riegelkontext(t)

	einPreis(t, anwendung, "shivan-lea", "EUR", 1250)
	einPreis(t, anwendung, "bolt-m10", "EUR", 150)

	// **Die Vorgabe wird gelesen** — hier USD, ohne eine Zeile in USD: dann
	// gilt die Währung mit den meisten Zeilen.
	if fehler := storage.NeueEinstellungen(anwendung.db).Setzen(
		anwendung.imProfil(ctx), domain.VorgabeWaehrung, "USD"); fehler != nil {
		t.Fatal(fehler)
	}

	antwort := anwendung.HandelPreisanhalt(ctx, []HandelspositionDTO{
		{Von: ich, Printing: "bolt-m10", Menge: 2, Finish: "nonfoil"},
		{Von: dasLeitungsgegenueber, Printing: "shivan-lea", Menge: 1, Finish: "nonfoil"},
		{Von: dasLeitungsgegenueber, Printing: "unbekannt", Menge: 1},
	})

	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	anhalt := antwort.Anhalt
	if len(anhalt.Zeilen) != 3 || len(anhalt.Zeilen[2].Preise) != 0 ||
		anhalt.Zeilen[0].Preise[0].Wert != 300 || anhalt.Zeilen[0].Preise[0].Erhoben == "" {
		t.Fatalf("Zeilen: %+v", anhalt.Zeilen)
	}

	if len(anhalt.Seiten) != 2 || anhalt.Seiten[0].Von != ich || anhalt.Seiten[1].OhnePreis != 1 {
		t.Fatalf("Seiten: %+v", anhalt.Seiten)
	}

	if len(anhalt.Differenzen) != 1 || anhalt.Differenzen[0].Wert != 1250-300 ||
		anhalt.Differenzen[0].Vollstaendig || anhalt.Hauptwaehrung != "EUR" {
		t.Fatalf("Differenz: %+v, Hauptwährung %q", anhalt.Differenzen, anhalt.Hauptwaehrung)
	}
}

func TestDerAnhaltEinesVorgangsNimmtSeinePositionen(t *testing.T) {
	anwendung, _ := eineLeitungsflaeche(t)
	ctx := riegelkontext(t)

	einPreis(t, anwendung, "shivan-lea", "EUR", 1250)

	if angelegt := anwendung.HandelEroeffnen(ctx, einLeitungswunsch()); angelegt.Fehler != nil {
		t.Fatalf("Eröffnen: %+v", angelegt.Fehler)
	}

	antwort := anwendung.HandelPreisanhaltVon(ctx, derLeitungsvorgang)
	if antwort.Fehler != nil || len(antwort.Anhalt.Zeilen) != 1 ||
		antwort.Anhalt.Differenzen[0].Wert != 1250 || !antwort.Anhalt.Differenzen[0].Vollstaendig {
		t.Fatalf("Antwort: %+v", antwort)
	}

	// **Ein Vorgang, den es nicht gibt, ist eine Absage** — mit leeren
	// Listen und nicht mit `null`.
	fehlt := anwendung.HandelPreisanhaltVon(ctx, "v-gibt-es-nicht")
	if fehlt.Fehler == nil || fehlt.Anhalt.Zeilen == nil || fehlt.Anhalt.Seiten == nil ||
		fehlt.Anhalt.Differenzen == nil {
		t.Fatalf("Absage: %+v", fehlt)
	}
}

func TestDerAnhaltOhneProfilIstEineAbsage(t *testing.T) {
	antwort := PeerflaecheUeber(bestandFuerDenRiegel(t)).HandelPreisanhalt(t.Context(), []HandelspositionDTO{
		{Von: "ich", Printing: "bolt-m10", Menge: 1},
	})

	if antwort.Fehler == nil || antwort.Anhalt.Zeilen == nil {
		t.Fatalf("Antwort: %+v", antwort)
	}
}

func TestEinFehlenderPreisbestandIstEineAbsage(t *testing.T) {
	anwendung, ich := eineLeitungsflaeche(t)

	// **Die Einstellungen lesen sich, die Preise nicht** — so geht der
	// Fehler aus der Rechnung selbst zurück.
	if _, fehler := anwendung.db.ExecContext(t.Context(), `DROP TABLE preisstaende`); fehler != nil {
		t.Fatal(fehler)
	}

	antwort := anwendung.HandelPreisanhalt(riegelkontext(t), []HandelspositionDTO{
		{Von: ich, Printing: "bolt-m10", Menge: 1},
	})
	if antwort.Fehler == nil || antwort.Anhalt.Differenzen == nil {
		t.Fatalf("Antwort: %+v", antwort)
	}
}
