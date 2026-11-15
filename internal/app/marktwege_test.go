package app_test

// **Die Wege der Marktnaht, die kein Erfolgsfall betritt** — A4, #1443.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ohne eingesetzten Bund wird der echte gebaut — und er hat nichts.**
//
// # Warum das gefahrlos ist
//
// `schluesselbund.Anmeldebund()` **baut** den Bund und liest nichts; gelesen
// wird erst beim Fragen, und `schluesselbund.Bundort` lenkt das auf einen
// Bund in einem Wegwerfverzeichnis um. Ohne diese Umlenkung fragte der Lauf
// den echten Schlüsselbund dieses Rechners (§3).
func TestOhneEingesetztenBundWirdDerEchteGebaut(t *testing.T) {
	t.Setenv(schluesselbund.Bundort, probebundOrt(t))
	t.Setenv("HOME", t.TempDir())

	antwort := app.MarktAnwendungUeber(frischerBestand(t)).
		Marktangebote(probekontext(t), "cardtrader", "abc", "", "", nil)

	if antwort.Fehler == nil {
		t.Fatal("ein leerer Bund gab eine Auskunft")
	}
}

// **Ohne umgelenkte Adresse gilt die Vorgabe des Pakets.**
//
// **Und trotzdem geht nichts hinaus**: der Dienst ist aus, und die Erlaubnis
// am Netzklienten fragt **vor** der Anfrage — das ist in
// `internal/netz/erlaubnis_test.go` gemessen, nicht hier behauptet.
func TestOhneUmlenkungGiltDieVorgabeadresse(t *testing.T) {
	db := frischerBestand(t)

	if fehler := storage.NeueEinstellungen(db).Setzen(probekontext(t),
		domain.DienstEinstellung(domain.DienstCardTrader), domain.DienstAus,
	); fehler != nil {
		t.Fatalf("den Dienst abschalten: %v", fehler)
	}

	antwort := app.MarktAnwendungUeber(db).
		MitMarktbund(marktbund{marke: "geheim"}).
		Marktangebote(probekontext(t), "cardtrader", "abc", "", "", nil)

	if antwort.Fehler == nil {
		t.Fatal("der abgeschaltete Dienst gab eine Auskunft")
	}
}

// **Ein klemmender Bund ist ein Fehler und kein fehlender Zugang.**
func TestEinKlemmenderBundIstEinFehler(t *testing.T) {
	t.Parallel()

	antwort := app.MarktAnwendungUeber(frischerBestand(t)).
		MitMarktbund(klemmenderBund{}).
		Marktangebote(probekontext(t), "cardtrader", "abc", "", "", nil)

	if antwort.Fehler == nil {
		t.Fatal("ein klemmender Bund gab eine Auskunft")
	}
}

// **Ein Fehler bei den Blueprints reist weiter.**
func TestEinFehlerBeiDenBlueprintsReistWeiter(t *testing.T) {
	db := frischerBestand(t)
	marktbestand(t, db)

	server := halbeGegenstelle(t, "/blueprints/export")
	t.Setenv(app.CardTraderUmgebung, server.URL)

	antwort := app.MarktAnwendungUeber(db).
		MitMarktbund(marktbund{marke: "geheim"}).
		Marktangebote(probekontext(t), "cardtrader",
			"11111111-1111-4111-8111-111111111111", "", "", nil)

	if antwort.Fehler == nil {
		t.Fatal("ein 500 bei den Blueprints kam als Erfolg zurück")
	}
}

// **Und ein Fehler am Marktplatz ebenso.**
func TestEinFehlerAmMarktplatzReistWeiter(t *testing.T) {
	db := frischerBestand(t)
	marktbestand(t, db)

	server := halbeGegenstelle(t, "/marketplace/products")
	t.Setenv(app.CardTraderUmgebung, server.URL)

	antwort := app.MarktAnwendungUeber(db).
		MitMarktbund(marktbund{marke: "geheim"}).
		Marktangebote(probekontext(t), "cardtrader",
			"11111111-1111-4111-8111-111111111111", "", "", nil)

	if antwort.Fehler == nil {
		t.Fatal("ein 500 am Marktplatz kam als Erfolg zurück")
	}
}

// **Eine gemerkte Fassung, die keine Zahl ist, wird abgewiesen.**
//
// # Wie so etwas entsteht
//
// **Durch einen Bestand aus einer anderen Hand.** Die Kennungstabelle nimmt
// Text; wer einen Bestand von Hand ändert oder aus einer fremden Quelle
// einliest, kann dort alles stehen haben. Ein `strconv` ohne Prüfung ergäbe
// hier stillschweigend die Fassung 0 — und die gibt es nicht.
func TestEineFassungDieKeineZahlIstWirdAbgewiesen(t *testing.T) {
	db := frischerBestand(t)
	marktbestand(t, db)

	if fehler := storage.NeueDrucke(db).Sichern(probekontext(t), &domain.Druck{
		Name: "Probe", Setcode: "ptkdf", Nummer: "1", Sprache: "en",
		Kennungen: []domain.Druckkennung{
			{
				System: domain.SystemScryfall, Semantik: domain.SemantikDruck,
				Wert: "11111111-1111-4111-8111-111111111111",
			},
			{
				System: domain.SystemCardTrader, Semantik: domain.SemantikDruck,
				Wert: "keine Zahl",
			},
		},
	}); fehler != nil {
		t.Fatalf("den Druck sichern: %v", fehler)
	}

	server := marktgegenstelle(t)
	t.Setenv(app.CardTraderUmgebung, server.URL)

	antwort := app.MarktAnwendungUeber(db).
		MitMarktbund(marktbund{marke: "geheim"}).
		Marktangebote(probekontext(t), "cardtrader",
			"11111111-1111-4111-8111-111111111111", "", "", nil)

	if antwort.Fehler == nil {
		t.Fatal("eine Fassung ohne Zahl kam durch")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "keine Zahl") {
		t.Errorf("die Meldung ist %q", antwort.Fehler.Meldung)
	}
}

// halbeGegenstelle antwortet wie CardTrader — außer auf einem Weg.
func halbeGegenstelle(t *testing.T, kaputt string) *httptest.Server {
	t.Helper()

	gut := marktgegenstelle(t)

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if gestellt.URL.Path == kaputt {
				schreiben.WriteHeader(http.StatusInternalServerError)

				return
			}

			gut.Config.Handler.ServeHTTP(schreiben, gestellt)
		}))

	t.Cleanup(server.Close)

	return server
}

// klemmenderBund antwortet mit einem Fehler.
type klemmenderBund struct{}

var errBundKlemmt = errors.New("der Bund klemmt")

func (klemmenderBund) Lies(_ context.Context, _, _ string) (string, error) {
	return "", errBundKlemmt
}
