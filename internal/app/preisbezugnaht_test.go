package app_test

// **Die Naht zum Bulkbezug** — Werkstück A3 von Phase 5.3
// (#1442).
//
// **Ohne Netz**: der Lauf geht gegen einen frischen Bestand, und der hat
// keinen Druck mit MTGJSON-Kennung. Damit greift die Zusage aus §16, Satz 4 —
// „ohne Crosswalk wird nicht gefragt" —, und der Weg bis dorthin ist
// vollständig gefahren.

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein frischer Bestand fragt nicht — und sagt es.**
func TestDerBulkbezugFragtOhneKrosswalkNicht(t *testing.T) {
	db := frischerBestand(t)

	antwort := app.PreisbulkAnwendungUeber(db).
		MitBulkadresse("https://example.invalid").
		Preisbulk(probekontext(t), "mtgjson", "heute")

	if antwort.Fehler != nil {
		t.Fatalf("Preisbulk: %+v", antwort.Fehler)
	}

	if antwort.Drucke != 0 || antwort.Staende != 0 {
		t.Errorf("die Bilanz ist %+v", antwort.BulkbilanzDTO)
	}

	if antwort.Quelle != "mtgjson" || antwort.Umfang != "heute" {
		t.Errorf("die Bilanz nennt %q und %q", antwort.Quelle, antwort.Umfang)
	}
}

// **Eine Quelle, die es nicht gibt, geht gar nicht erst an den Bestand.**
func TestEineUnbekannteQuelleWirdVorDemBestandAbgewiesen(t *testing.T) {
	antwort := app.PreisbulkAnwendungUeber(nil).
		Preisbulk(probekontext(t), "cardmarket", "heute")

	if antwort.Fehler == nil {
		t.Fatal("es kam durch")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "cardmarket") {
		t.Errorf("die Meldung ist %q", antwort.Fehler.Meldung)
	}
}

// **Ohne offenen Bestand sagt sie das** (§8).
func TestDerBulkbezugOhneBestandMeldetSich(t *testing.T) {
	antwort := app.NeueApp("").Preisbulk(probekontext(t), "mtgjson", "heute")

	if antwort.Fehler == nil {
		t.Error("es kam durch")
	}
}

// **Mit einem Druck, den MTGJSON kennt, geht der ganze Weg.**
//
// # Warum das hier steht und nicht in `internal/flaechen`
//
// Der Crosswalk kommt aus dem Bestand, und ein Druck mit **zwei** Kennungen
// entsteht nur über die Ablage — keine Fläche ohne Fenster legt einen an. Was
// hier gefahren wird, ist deshalb die Naht: Bestand → Crosswalk → Klient →
// Ablage, mit einer eigenen Gegenstelle statt mtgjson.com (§3).
func TestDerBulkbezugHoltUndLegtAb(t *testing.T) {
	db := frischerBestand(t)

	if fehler := storage.NeueDrucke(db).Sichern(probekontext(t), &domain.Druck{
		Name: "Probe", Setcode: "m11", Nummer: "1", Sprache: "en",
		Kennungen: []domain.Druckkennung{
			{
				System: domain.SystemScryfall, Semantik: domain.SemantikDruck,
				Wert: "11111111-1111-4111-8111-111111111111",
			},
			{
				System: domain.SystemMTGJSON, Semantik: domain.SemantikDruck,
				Wert: "22222222-2222-4222-8222-222222222222",
			},
		},
	}); fehler != nil {
		t.Fatalf("den Druck sichern: %v", fehler)
	}

	// **Der Dienst muss an sein** — die Vorgabe für die neuen ist „aus"
	// ([0063](../../docs/entscheidungen/0063-neue-dienste-sind-aus.md)), und
	// das ist genau die Zusage, die der Riegel hält.
	if fehler := storage.NeueEinstellungen(db).Setzen(probekontext(t),
		domain.DienstEinstellung(domain.DienstMTGJSON), domain.DienstAn,
	); fehler != nil {
		t.Fatalf("den Dienst anschalten: %v", fehler)
	}

	server := bulkgegenstelle(t)
	t.Setenv(app.MTGJSONPreisumgebung, server.URL)

	antwort := app.PreisbulkAnwendungUeber(db).
		Preisbulk(probekontext(t), "mtgjson", "heute")

	if antwort.Fehler != nil {
		t.Fatalf("Preisbulk: %+v", antwort.Fehler)
	}

	if antwort.Drucke != 1 || antwort.Staende != 1 {
		t.Errorf("die Bilanz ist %+v", antwort.BulkbilanzDTO)
	}

	// **Und der Befund der Quelle reist mit** (§8).
	if len(antwort.Befunde) != 1 ||
		!strings.Contains(antwort.Befunde[0], "mond") {
		t.Errorf("die Befunde sind %v", antwort.Befunde)
	}
}

// **Und ein abgeschalteter Dienst hält ihn auf** (§16, Regel 6).
//
// **Das ist die Gegenprobe zum Lauf darüber**: ohne sie sagte er nur, dass der
// Weg geht — nicht, dass der Schalter ihn erreicht.
func TestEinAbgeschalteterDienstHaeltDenBulkbezugAuf(t *testing.T) {
	db := frischerBestand(t)

	if fehler := storage.NeueDrucke(db).Sichern(probekontext(t), &domain.Druck{
		Name: "Probe", Setcode: "m11", Nummer: "1", Sprache: "en",
		Kennungen: []domain.Druckkennung{
			{
				System: domain.SystemScryfall, Semantik: domain.SemantikDruck,
				Wert: "11111111-1111-4111-8111-111111111111",
			},
			{
				System: domain.SystemMTGJSON, Semantik: domain.SemantikDruck,
				Wert: "22222222-2222-4222-8222-222222222222",
			},
		},
	}); fehler != nil {
		t.Fatalf("den Druck sichern: %v", fehler)
	}

	server := bulkgegenstelle(t)
	t.Setenv(app.MTGJSONPreisumgebung, server.URL)

	// **Kein Schalter gesetzt** — und die Vorgabe ist „aus".
	antwort := app.PreisbulkAnwendungUeber(db).
		Preisbulk(probekontext(t), "mtgjson", "heute")

	if antwort.Fehler == nil {
		t.Fatal("der abgeschaltete Dienst wurde gefragt")
	}
}

// bulkgegenstelle antwortet mit einer gepackten Preisdatei.
//
// **Ein Preis und ein unbekannter Markt**: so ist in einem Lauf beides
// geprüft — was ankommt und was als Befund danebensteht.
func bulkgegenstelle(t *testing.T) *httptest.Server {
	t.Helper()

	const inhalt = `{
	  "meta": {"date": "2026-09-11"},
	  "data": {
	    "22222222-2222-4222-8222-222222222222": {
	      "mond":  {"haendler": {"currency": "USD", "retail": {"normal": {"2026-09-10": 1.00}}}},
	      "paper": {"cardkingdom": {"currency": "USD", "retail": {"normal": {"2026-09-10": 4.17}}}}
	    }
	  }
	}`

	var puffer bytes.Buffer

	schreiber := gzip.NewWriter(&puffer)

	if _, fehler := schreiber.Write([]byte(inhalt)); fehler != nil {
		t.Fatalf("packen: %v", fehler)
	}

	if fehler := schreiber.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write(puffer.Bytes()) //nolint:errcheck // die Gegenstelle
		}))

	t.Cleanup(server.Close)

	return server
}
