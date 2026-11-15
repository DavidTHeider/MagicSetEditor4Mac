package app_test

// **Der Dienstkatalog** — Werkstück A1 von Phase 5.3
// (#1440).
//
// Was hier geprüft wird: dass jede Fähigkeit aus `domain.Dienstvertraege`
// mit Status, Grund und Vertrag ankommt; dass ein Hindernis bei **einem**
// Dienst die anderen nicht anhält; dass der Riegel jede Fähigkeit sperrt,
// die er sperren muss; dass eine Wahl nur dort zählt, wo es eine gibt — und
// dass der Katalog **nichts schickt**: er ist eine Auskunft über die
// Bereitschaft, kein Versuch.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// katalogwahl ist Schalter, Wahl und Riegel in einem — wie [app.Einstellungswahl].
type katalogwahl struct {
	an     map[domain.Dienst]bool
	frage  domain.Dienst
	sperre error
	fehler error
}

func (k *katalogwahl) IstAn(_ context.Context, dienst domain.Dienst) (bool, error) {
	return k.an[dienst], k.fehler
}

func (k *katalogwahl) FuerFrage(_ context.Context, _ domain.Frage) (domain.Dienst, error) {
	return k.frage, k.fehler
}

func (k *katalogwahl) NachAussen(context.Context) error { return k.sperre }

// stoerbund liest ein Konto nicht — für alle anderen ist er ein [bundprobe].
type stoerbund struct {
	bundprobe
	gestoert string
	fehler   error
}

func (s stoerbund) Lies(ctx context.Context, dienst, konto string) (string, error) {
	if strings.HasPrefix(konto, s.gestoert+".") {
		return "", s.fehler
	}

	return s.bundprobe.Lies(ctx, dienst, konto)
}

func eintrag(t *testing.T, katalog []app.DienstfaehigkeitDTO, dienst, frage string) app.DienstfaehigkeitDTO {
	t.Helper()

	for _, eintrag := range katalog {
		if eintrag.Dienst == dienst && eintrag.Frage == frage {
			return eintrag
		}
	}

	t.Fatalf("%s/%s steht nicht im Katalog", dienst, frage)

	return app.DienstfaehigkeitDTO{}
}

func TestDerKatalogNenntJedeFaehigkeitMitStatusUndVertrag(t *testing.T) {
	t.Parallel()

	wahl := &katalogwahl{
		an:    map[domain.Dienst]bool{domain.DienstScryfall: true, domain.DienstScrydex: true},
		frage: domain.DienstScryfall,
	}
	bund := bundprobe{werte: map[string]string{"scrydex.schluessel": "geheim"}}

	katalog, err := app.Dienstkatalogeintraege(probekontext(t), wahl, bund)
	if err != nil {
		t.Fatalf("der Katalog: %v", err)
	}

	if len(katalog) != len(domain.Dienstvertraege()) {
		t.Fatalf("%d Einträge, erwartet %d", len(katalog), len(domain.Dienstvertraege()))
	}

	suche := eintrag(t, katalog, "scryfall", "suche")
	if suche.Status != domain.StatusBereit || !suche.HatWahl || !suche.Gewaehlt || suche.Grund != "" {
		t.Errorf("scryfall/suche: %+v — erwartet bereit, mit Wahl und gewählt", suche)
	}

	if suche.Wirkung != string(domain.WirkungExternSenden) {
		t.Errorf("die Wirkung ist %q, erwartet externSenden", suche.Wirkung)
	}

	if suche.Vertrag.Basisadresse != "https://api.scryfall.com" ||
		suche.Vertrag.Abgerufen == "" || suche.Vertrag.Beleg == "" ||
		len(suche.Vertrag.Endpunkte) == 0 {
		t.Errorf("der Vertrag kam nicht mit: %+v", suche.Vertrag)
	}

	if suche.Nachweis == nil {
		t.Fatal("scryfall/suche hat keinen Nachweis")
	}

	// **Ein Lauf aus einer früheren Fassung ist kein Nachweis über diese.**
	if suche.Nachweis.Fassung == app.Fassung && !suche.Nachweis.GiltFuerDieseFassung ||
		suche.Nachweis.Fassung != app.Fassung && suche.Nachweis.GiltFuerDieseFassung {
		t.Errorf("der Nachweis %+v sagt Falsches über die Fassung %s",
			suche.Nachweis, app.Fassung)
	}

	bild := eintrag(t, katalog, "scryfall", "bild")
	if bild.Status != domain.StatusBereit || bild.HatWahl || bild.Nachweis != nil {
		t.Errorf("scryfall/bild: %+v — erwartet bereit, ohne Wahl, ohne Nachweis", bild)
	}

	if aus := eintrag(t, katalog, "mtgapi", "suche"); aus.Status != domain.StatusAus ||
		!strings.Contains(aus.Grund, "abgeschaltet") {
		t.Errorf("mtgapi/suche: %+v — erwartet aus", aus)
	}

	scrydex := eintrag(t, katalog, "scrydex", "suche")
	if scrydex.Status != domain.StatusZugangFehlt || scrydex.Angabe != "team" ||
		!strings.Contains(scrydex.Grund, "team") {
		t.Errorf("scrydex/suche: %+v — erwartet zugangFehlt wegen team", scrydex)
	}

	if suche.Angabe != "" {
		t.Errorf("scryfall/suche nennt eine fehlende Angabe %q", suche.Angabe)
	}

	if scrydex.Nachweis != nil {
		t.Error("Scrydex trägt einen Nachweis — es gibt hier kein Konto")
	}
}

// **Ein Anbieterfehler bleibt beim Anbieter** (#1440): der Schlüsselbund,
// der für Scrydex nicht antwortet, hält weder Scryfall noch MTGJSON an.
func TestEinGestoerterSchluesselbundStoertNurSeinenDienst(t *testing.T) {
	t.Parallel()

	wahl := &katalogwahl{
		an: map[domain.Dienst]bool{
			domain.DienstScryfall: true, domain.DienstMTGJSON: true, domain.DienstScrydex: true,
		},
		frage: domain.DienstMTGJSON,
	}
	bund := stoerbund{
		bundprobe: bundprobe{werte: map[string]string{"mtgjson.marke": "patreon"}},
		gestoert:  "scrydex",
		fehler:    errors.New("der Schlüsselbund ist gesperrt"),
	}

	katalog, err := app.Dienstkatalogeintraege(probekontext(t), wahl, bund)
	if err != nil {
		t.Fatalf("der Katalog: %v", err)
	}

	if scrydex := eintrag(t, katalog, "scrydex", "suche"); scrydex.Status != domain.StatusGestoert ||
		!strings.Contains(scrydex.Grund, "gesperrt") {
		t.Errorf("scrydex/suche: %+v — erwartet gestört mit Grund", scrydex)
	}

	if mtgjson := eintrag(t, katalog, "mtgjson", "suche"); mtgjson.Status != domain.StatusBereit {
		t.Errorf("mtgjson/suche: %+v — erwartet bereit, die Marke liegt", mtgjson)
	}

	if scryfall := eintrag(t, katalog, "scryfall", "regeln"); scryfall.Status != domain.StatusBereit {
		t.Errorf("scryfall/regeln: %+v — erwartet bereit", scryfall)
	}
}

func TestDerRiegelSperrtJedeEingeschalteteFaehigkeit(t *testing.T) {
	t.Parallel()

	wahl := &katalogwahl{
		an:     map[domain.Dienst]bool{domain.DienstScryfall: true, domain.DienstMTGAPI: true},
		frage:  domain.DienstScryfall,
		sperre: errors.New("der Leser darf nicht nach draußen"),
	}

	katalog, err := app.Dienstkatalogeintraege(probekontext(t), wahl, bundprobe{})
	if err != nil {
		t.Fatalf("der Katalog: %v", err)
	}

	for _, fall := range katalog {
		switch fall.Dienst {
		case "scryfall", "mtgapi":
			if fall.Status != domain.StatusGesperrt || !strings.Contains(fall.Grund, "Leser") {
				t.Errorf("%s/%s: %+v — erwartet gesperrt", fall.Dienst, fall.Frage, fall)
			}
		default:
			// **Aus gewinnt**: ein abgeschalteter Dienst braucht keinen Riegel.
			if fall.Status != domain.StatusAus {
				t.Errorf("%s/%s: %+v — erwartet aus", fall.Dienst, fall.Frage, fall)
			}
		}
	}
}

// **Kein stiller Anbieterwechsel** (0063, #1440): wer für die Suche
// magicthegathering.io gewählt hat, sieht Scryfalls Suche als »nicht
// gewählt« — und Scryfalls Bilder weiter als bereit, denn dort gibt es
// keine Wahl.
func TestEinAndererGewaehlterDienstMachtNichtBereit(t *testing.T) {
	t.Parallel()

	wahl := &katalogwahl{
		an:    map[domain.Dienst]bool{domain.DienstScryfall: true, domain.DienstMTGAPI: true},
		frage: domain.DienstMTGAPI,
	}

	katalog, err := app.Dienstkatalogeintraege(probekontext(t), wahl, bundprobe{})
	if err != nil {
		t.Fatalf("der Katalog: %v", err)
	}

	if suche := eintrag(t, katalog, "scryfall", "suche"); suche.Status != domain.StatusNichtGewaehlt ||
		suche.Gewaehlt || !strings.Contains(suche.Grund, "gewählt") {
		t.Errorf("scryfall/suche: %+v — erwartet nicht gewählt", suche)
	}

	if bild := eintrag(t, katalog, "scryfall", "bild"); bild.Status != domain.StatusBereit {
		t.Errorf("scryfall/bild: %+v — erwartet bereit", bild)
	}

	if mtgapi := eintrag(t, katalog, "mtgapi", "suche"); mtgapi.Status != domain.StatusBereit ||
		!mtgapi.Gewaehlt {
		t.Errorf("mtgapi/suche: %+v — erwartet bereit und gewählt", mtgapi)
	}
}

// **Eine Wahl, die keine Frage beantwortet, nimmt die Vorgabe** — der
// Schalter der Tests kennt keine Fragen, und ein Katalog darüber sagt, was
// die Einstellung sagte, hätte sie niemand angefasst.
func TestOhneFragewahlGiltDieVorgabe(t *testing.T) {
	t.Parallel()

	wahl := &schaltbareWahl{an: map[domain.Dienst]bool{
		domain.DienstScryfall: true, domain.DienstMTGAPI: true,
	}}

	katalog, err := app.Dienstkatalogeintraege(probekontext(t), wahl, bundprobe{})
	if err != nil {
		t.Fatalf("der Katalog: %v", err)
	}

	if suche := eintrag(t, katalog, "scryfall", "suche"); suche.Status != domain.StatusBereit || !suche.Gewaehlt {
		t.Errorf("scryfall/suche: %+v — erwartet bereit als Vorgabe", suche)
	}

	if mtgapi := eintrag(t, katalog, "mtgapi", "suche"); mtgapi.Status != domain.StatusNichtGewaehlt {
		t.Errorf("mtgapi/suche: %+v — erwartet nicht gewählt", mtgapi)
	}
}

// **Der Katalog schickt nichts** — auch nicht, wenn alles an ist. Vier
// Spione, null Anfragen.
func TestDerKatalogSchicktKeineEinzigeAnfrage(t *testing.T) {
	var anfragen atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(
		func(_ http.ResponseWriter, _ *http.Request) { anfragen.Add(1) }))
	defer server.Close()

	t.Setenv("HOME", t.TempDir())

	zurueck := app.MitDienstadressen(server.URL, server.URL, server.URL)
	defer zurueck()

	vorher := app.MitScryfalladresse(server.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	wahl := &katalogwahl{an: map[domain.Dienst]bool{}, frage: domain.DienstScryfall}
	for _, dienst := range domain.Dienste() {
		wahl.an[dienst] = true
	}

	bund := bundprobe{werte: map[string]string{
		"mtgjson.marke": "m", "scrydex.schluessel": "s", "scrydex.team": "t",
		"cardtrader.marke": "c",

		// **Anthropic seit A8** (#1447): der Katalog führt ihn wie jeden
		// anderen Dienst, und ohne Schlüssel stünde er hier als
		// `zugangFehlt` da — ein Stand, den dieser Lauf nicht prüft.
		"anthropic.schluessel": "a", "openai.schluessel": "o",
		"xai.schluessel": "x",
	}}

	katalog, err := app.Dienstkatalogeintraege(probekontext(t), wahl, bund)
	if err != nil {
		t.Fatalf("der Katalog: %v", err)
	}

	for _, fall := range katalog {
		if fall.Status != domain.StatusBereit && fall.Status != domain.StatusNichtGewaehlt {
			t.Errorf("%s/%s ist %s: %s", fall.Dienst, fall.Frage, fall.Status, fall.Grund)
		}
	}

	if gezaehlt := anfragen.Load(); gezaehlt != 0 {
		t.Errorf("der Katalog schickte %d Anfragen — er ist eine Auskunft, kein Versuch", gezaehlt)
	}
}

func TestEinUnlesbarerSchalterIstEinFehlerDesKatalogs(t *testing.T) {
	t.Parallel()

	wahl := &katalogwahl{fehler: errors.New("die Einstellungen sind nicht lesbar")}

	if _, err := app.Dienstkatalogeintraege(probekontext(t), wahl, bundprobe{}); err == nil ||
		!strings.Contains(err.Error(), "nicht lesbar") {
		t.Errorf("der Fehler ist %v — erwartet der des Schalters", err)
	}

	wahl = &katalogwahl{an: map[domain.Dienst]bool{domain.DienstScryfall: true}}
	wahl.fehler = nil

	sperrig := &fragefehlerwahl{katalogwahl: wahl, fehler: errors.New("die Wahl ist nicht lesbar")}

	if _, err := app.Dienstkatalogeintraege(probekontext(t), sperrig, bundprobe{}); err == nil ||
		!strings.Contains(err.Error(), "Wahl ist nicht lesbar") {
		t.Errorf("der Fehler ist %v — erwartet der der Wahl", err)
	}
}

// fragefehlerwahl liest den Schalter, aber nicht die Wahl.
type fragefehlerwahl struct {
	*katalogwahl
	fehler error
}

func (f *fragefehlerwahl) FuerFrage(context.Context, domain.Frage) (domain.Dienst, error) {
	return "", f.fehler
}

func TestDieBindungLiefertDenKatalog(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitSchluesselbund(neuerBundspeicher())
	anwendung.MitDienstwahl(&katalogwahl{
		an: map[domain.Dienst]bool{domain.DienstScryfall: true}, frage: domain.DienstScryfall,
	})

	antwort := anwendung.Dienstkatalog(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Eintraege) != len(domain.Dienstvertraege()) {
		t.Errorf("%d Einträge, erwartet %d", len(antwort.Eintraege), len(domain.Dienstvertraege()))
	}

	anwendung.MitDienstwahl(&katalogwahl{fehler: errors.New("kaputt")})

	if antwort := anwendung.Dienstkatalog(t.Context()); antwort.Fehler == nil {
		t.Error("ein unlesbarer Schalter kam ohne Fehler durch")
	} else if antwort.Eintraege == nil {
		t.Error("die Liste ist nil — das Fenster erwartet eine leere")
	}
}
