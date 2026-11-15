package app_test

// **Das eigene Konto am Fenster** — Werkstück `T11` von Phase 5.6
// (#1476).
//
// # Was hier geprüft wird
//
//   - **Der Marktplatz zuerst**: ein Name, den es nicht gibt, öffnet keinen
//     Schlüsselbund.
//   - **Ohne Bestand und ohne Marke gibt es keine Auskunft** — und beide
//     Absagen sagen, was fehlt.
//   - **Und die Kontoschau fängt aus an**: mit Marke und mit Bestand ist sie
//     immer noch aus, und genau das ist die Zusage dieses Werkstücks.
//
// # Was er **nicht** prüft
//
// **Den geglückten Abruf.** Er braucht eine echte Zugangsmarke; es gibt
// keine, und eine zu beschaffen ist keine Maßnahme, die ein Lauf trifft.
// Was der Adapter aus einer Antwort macht, misst `internal/cardtrader`
// gegen selbst geschriebene Beispiele — **und das ist keine
// Anbieterabnahme** ([0122](../../docs/entscheidungen/0122-ein-fremdes-konto-wird-gelesen.md)).

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Marktplatz, den es nicht gibt, öffnet keinen Schlüsselbund.**
func TestEinUnbekannterMarktplatzHatKeinKonto(t *testing.T) {
	t.Parallel()

	antwort := app.MarktAnwendungUeber(frischerBestand(t)).
		Marktkonto(probekontext(t), "kartenkiosk", 0, 0)

	if antwort.Fehler == nil {
		t.Fatal("der fremde Marktplatz kam durch")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "kartenkiosk") {
		t.Errorf("die Meldung ist %q", antwort.Fehler.Meldung)
	}
}

// **Ohne Bestand gibt es kein Konto.**
func TestOhneBestandGibtEsKeinMarktkonto(t *testing.T) {
	t.Parallel()

	if antwort := app.NeueApp("").
		Marktkonto(probekontext(t), "cardtrader", 0, 0); antwort.Fehler == nil {
		t.Fatal("ohne Bestand kam eine Auskunft")
	}
}

// **Ohne Zugangsmarke wird nicht gefragt.**
//
// Der Schlüsselbund ist eingesetzt und leer — ohne ihn führe der Lauf in den
// echten Bund dieses Rechners (§3).
func TestOhneMarkeGibtEsKeinMarktkonto(t *testing.T) {
	t.Parallel()

	antwort := app.MarktAnwendungUeber(frischerBestand(t)).
		MitMarktbund(marktbund{}).
		Marktkonto(probekontext(t), "cardtrader", 0, 0)

	if antwort.Fehler == nil {
		t.Fatal("ohne Marke kam eine Auskunft")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "marke") {
		t.Errorf("die Meldung ist %q", antwort.Fehler.Meldung)
	}
}

// **Und mit Marke ist die Kontoschau immer noch aus.**
//
// **Das ist die Zusage dieses Werkstücks**, und sie steht hier als Messung:
// der Marktplatz einzuschalten und eine Marke einzutragen reicht **nicht**.
// Ohne diesen Lauf sähe der vorige aus, als wäre die Marke der einzige
// Riegel.
func TestMitMarkeIstDieKontoschauImmerNochAus(t *testing.T) {
	t.Parallel()

	antwort := app.MarktAnwendungUeber(frischerBestand(t)).
		MitMarktbund(marktbund{marke: "geheim"}).
		Marktkonto(probekontext(t), "cardtrader", 0, 0)

	if antwort.Fehler == nil {
		t.Fatal("die Kontoschau war an")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "Kontoschau") {
		t.Errorf("die Meldung nennt die Kontoschau nicht: %q",
			antwort.Fehler.Meldung)
	}
}

// **Die Bindung geht denselben Weg wie die App-Methode.**
//
// Ohne diesen Lauf bliebe `Bindungen.Marktkonto` ungerufen — und eine
// Bindung, die kein Lauf betritt, ist Code mit voller Abdeckung, den niemand
// benutzt (§10, erste Dimension).
func TestDieBindungDesMarktkontos(t *testing.T) {
	t.Parallel()

	anwendung := app.MarktAnwendungUeber(frischerBestand(t)).
		MitMarktbund(marktbund{})

	anwendung.Bereit(probekontext(t))

	if antwort := app.NeueBindungen(anwendung).
		Marktkonto(t.Context(), "cardtrader", 0, 0); antwort.Fehler == nil {
		t.Error("Bindungen.Marktkonto gab keinen Fehler")
	}
}

// **Mit beiden Schaltern und einer Marke kommt die Übersicht.**
//
// **Gegen eine eigene Gegenstelle** und nicht gegen CardTrader: was dieser
// Lauf belegt, ist der Weg durch die Schichten — nicht, dass der Anbieter
// heute so antwortet ([0122](../../docs/entscheidungen/0122-ein-fremdes-konto-wird-gelesen.md)).
func TestMitBeidenSchalternKommtDasMarktkonto(t *testing.T) {
	db := frischerBestand(t)

	// **Beide Schalter** — und das ist der Punkt: der Dienst allein genügt
	// nicht, und die Kontoschau allein auch nicht.
	for schluessel, wert := range map[string]string{
		domain.DienstEinstellung(domain.DienstCardTrader): domain.DienstAn,
		domain.Kontoschau: domain.KontoschauAn,
	} {
		if fehler := storage.NeueEinstellungen(db).Setzen(probekontext(t),
			schluessel, wert); fehler != nil {
			t.Fatalf("%s einschalten: %v", schluessel, fehler)
		}
	}

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			switch {
			case strings.HasPrefix(anfrage.URL.Path, "/info"):
				schreibeRoh(t, schreiben,
					`{"id":3,"name":"Test App","user_id":51242}`)

			case strings.HasPrefix(anfrage.URL.Path, "/products/export"):
				schreibeRoh(t, schreiben, `[{"id":1,"name_en":"Tranquil Cove",`+
					`"quantity":2,"price_cents":250,"price_currency":"EUR",`+
					`"user_data_field":"Kiste 4"}]`)

			default:
				schreibeRoh(t, schreiben, `[{"id":9,"code":"AB12",`+
					`"state":"hub_pending","via_cardtrader_zero":true,`+
					`"order_as":"seller","size":3}]`)
			}
		}))

	t.Cleanup(server.Close)
	t.Setenv(app.CardTraderUmgebung, server.URL)

	antwort := app.MarktAnwendungUeber(db).
		MitMarktbund(marktbund{marke: "geheim"}).
		Marktkonto(probekontext(t), "cardtrader", 1, 20)
	if antwort.Fehler != nil {
		t.Fatalf("Marktkonto: %+v", antwort.Fehler)
	}

	if antwort.Stand.Name != "Test App" || antwort.Stand.NutzerID != 51242 {
		t.Errorf("das Konto kam als %+v an", antwort.Stand)
	}

	if len(antwort.Stand.Angebote) != 1 ||
		antwort.Stand.Angebote[0].Eigennotiz != "Kiste 4" {
		t.Errorf("die Angebote kamen als %+v an", antwort.Stand.Angebote)
	}

	if len(antwort.Stand.Bestellungen) != 1 {
		t.Fatalf("die Bestellungen kamen als %+v an", antwort.Stand.Bestellungen)
	}

	// **Und die Regel ist gerechnet** — `hub_pending` über CardTrader Zero
	// hat ein Stück genommen. Das ist die Auskunft, für die es dieses
	// Werkstück gibt.
	if !antwort.Stand.Bestellungen[0].BewegtBestand {
		t.Errorf("die Bestellung sagt, sie habe nichts bewegt: %+v",
			antwort.Stand.Bestellungen[0])
	}
}

// **Ein Schlüsselbund, der klagt, gibt kein Konto.**
//
// **Und die Klage kommt an**: ein leeres Konto neben einem klemmenden Bund
// sähe aus wie ein Konto ohne Angebote (§8).
func TestEinKlagenderBundGibtKeinMarktkonto(t *testing.T) {
	t.Parallel()

	db := frischerBestand(t)

	if fehler := storage.NeueEinstellungen(db).Setzen(probekontext(t),
		domain.Kontoschau, domain.KontoschauAn); fehler != nil {
		t.Fatalf("die Kontoschau einschalten: %v", fehler)
	}

	antwort := app.MarktAnwendungUeber(db).MitMarktbund(klagenderBund{}).
		Marktkonto(probekontext(t), "cardtrader", 0, 0)

	if antwort.Fehler == nil {
		t.Fatal("der klagende Bund gab ein Konto")
	}
}

// klagenderBund ist ein Schlüsselbund, der klemmt.
type klagenderBund struct{}

func (klagenderBund) Lies(context.Context, string, string) (string, error) {
	return "", errors.New("der Schlüsselbund klemmt")
}
