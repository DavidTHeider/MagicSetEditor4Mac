// **Die Kartensuche bei Scryfall** — Werkstück 1 von Phase 3.6
// (#1167).
package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// kartensuchprobe gibt eine feste Auskunft und merkt sich die Frage.
type kartensuchprobe struct {
	gefragt  string
	auskunft *service.Suchauskunft
	fehler   error
}

func (s *kartensuchprobe) Suche(_ context.Context, frage string) (*service.Suchauskunft, error) {
	s.gefragt = frage

	return s.auskunft, s.fehler
}

// **Die Treffer kommen als DTO an — samt Alter und »es gibt mehr«.**
func TestDieSuchtrefferKommenAn(t *testing.T) {
	t.Parallel()

	probe := &kartensuchprobe{auskunft: &service.Suchauskunft{
		Treffer: []service.Suchtreffer{{
			Kennung: "abc", Name: "Lightning Bolt",
			Setcode: "lea", Setname: "Limited Edition Alpha", Nummer: "161",
			Seltenheit: "common", Manakosten: "{R}", Typzeile: "Instant",
			Bild: "https://bilder/klein.jpg",
		}},
		Weitere: true,
		Alter:   3 * time.Hour,
		Befunde: []domain.Befund{{Ort: "karte", Meldung: "ein Feld war unklar"}},
	}}

	anwendung := app.NeueApp("")
	anwendung.MitScryfallsuche(probe)

	antwort := anwendung.ScryfallSuche(t.Context(), "bolt")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.gefragt != "bolt" {
		t.Errorf("gefragt wurde %q", probe.gefragt)
	}

	erwartet := app.SuchtrefferDTO{
		Kennung: "abc", Name: "Lightning Bolt",
		Setcode: "lea", Setname: "Limited Edition Alpha", Nummer: "161",
		Seltenheit: "common", Manakosten: "{R}", Typzeile: "Instant",
		Bild: "https://bilder/klein.jpg",
	}

	if len(antwort.Treffer) != 1 || antwort.Treffer[0] != erwartet {
		t.Errorf("die Treffer sind %+v", antwort.Treffer)
	}

	if !antwort.Weitere || antwort.AlterStunden != 3 || len(antwort.Befunde) != 1 {
		t.Errorf("die Herkunft ging verloren: %+v", antwort)
	}
}

// **Ohne Naht gibt es eine Meldung und keine leere Liste.**
//
// Eine Suche, die schweigend nichts findet, sieht aus wie eine, die nichts
// gefunden hat — und der Unterschied ist der zwischen „diese Karte gibt es
// nicht" und „ich kann gerade nicht suchen" (§8).
func TestOhneSuchnahtGibtEsEineMeldung(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").ScryfallSuche(probekontext(t), "bolt")

	if antwort.Fehler == nil {
		t.Fatal("die fehlende Naht kam durch")
	}

	if antwort.Treffer == nil || antwort.Befunde == nil {
		t.Error("die Listen sind nil statt leer — das Fenster liefe darüber")
	}
}

// **Ein Fehler der Suche kommt an.**
func TestEinFehlerDerSucheErreichtDasFenster(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitScryfallsuche(&kartensuchprobe{fehler: errSuche})

	if antwort := anwendung.ScryfallSuche(t.Context(), "bolt"); antwort.Fehler == nil {
		t.Error("der Fehler kam nicht an")
	}
}

// **Und die Fläche ohne Fenster gibt dasselbe** (§16).
func TestDieFremdsucheGibtDasselbe(t *testing.T) {
	t.Parallel()

	probe := &kartensuchprobe{auskunft: &service.Suchauskunft{
		Treffer: []service.Suchtreffer{{Kennung: "abc", Name: "Lightning Bolt"}},
		Alter:   2 * time.Hour,
	}}

	ergebnis, err := (&app.Fremdsuche{}).MitSuchquelle(probe).
		Suche(probekontext(t), "bolt")
	if err != nil {
		t.Fatalf("Suche: %v", err)
	}

	if len(ergebnis.Treffer) != 1 || ergebnis.Treffer[0].Name != "Lightning Bolt" {
		t.Errorf("die Treffer sind %+v", ergebnis.Treffer)
	}

	if ergebnis.AlterStunden != 2 {
		t.Errorf("das Alter ist %d", ergebnis.AlterStunden)
	}
}

// **Und ein Fehler ebenso.**
func TestEinFehlerDerFremdsucheKommtDurch(t *testing.T) {
	t.Parallel()

	_, err := (&app.Fremdsuche{}).MitSuchquelle(&kartensuchprobe{fehler: errSuche}).
		Suche(probekontext(t), "bolt")

	if !errors.Is(err, errSuche) {
		t.Errorf("der Fehler ist %v", err)
	}
}

// **Die Bindung reicht durch.**
func TestDieSuchbindungReichtDurch(t *testing.T) {
	t.Parallel()

	probe := &kartensuchprobe{auskunft: &service.Suchauskunft{
		Treffer: []service.Suchtreffer{{Kennung: "abc", Name: "Lightning Bolt"}},
	}}

	stand := bindungsprobe(t)
	stand.Anwendung.MitScryfallsuche(probe)

	antwort := stand.Bindungen.ScryfallSuche(t.Context(), "bolt")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Treffer) != 1 {
		t.Errorf("die Treffer sind %+v", antwort.Treffer)
	}

	if !strings.Contains(antwort.Treffer[0].Name, "Lightning") {
		t.Errorf("der Name ist %q", antwort.Treffer[0].Name)
	}
}

// errSuche ist der Fehler dieser Läufe.
var errSuche = errors.New("Scryfall antwortet nicht")

// **Die Suchnaht kommt ohne Bestand zusammen.**
//
// Dieselbe Eigenschaft wie bei der Auskunftsnaht, und aus demselben Grund:
// eine Suche liest nichts aus der Datenbank. Einen `*sql.DB` zu verlangen
// hieße, dass `mse scryfall-suche` ohne geöffneten Bestand nicht liefe.
func TestDieSuchnahtKommtOhneBestandZusammen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	naht, fehler := app.Fremdsuchnaht(probekontext(t), frischerBestand(t))
	if fehler != nil {
		t.Fatalf("Fremdsuchnaht: %v", fehler)
	}

	if naht == nil {
		t.Fatal("die Naht ist leer")
	}
}

// **Ohne Zuhause kommt sie nicht zusammen.**
//
// Der Zwischenspeicher liegt im Benutzerverzeichnis; ohne eines gibt es keinen
// Ort dafür — und eine Suche ohne Zwischenspeicher ginge bei jedem Tastendruck
// ins Netz.
func TestOhneZuhauseSagtEsDieSucheUndNichtDieNaht(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("MSE_BESTAND", "")

	// **Die Naht kommt zusammen** — sie entscheidet nichts mehr (Befund 1 des
	// Zweitblicks über Phase 4.8). Ohne Zuhause gibt es keinen
	// Scryfall-Zwischenspeicher und damit keinen Klienten; gesagt wird das
	// beim Suchen, mit dem Namen des gewählten Dienstes.
	naht, fehler := app.Fremdsuchnaht(probekontext(t), frischerBestand(t))
	if fehler != nil {
		t.Fatalf("die Naht: %v", fehler)
	}

	if _, suchfehler := naht.Suche(probekontext(t), "Blitz"); suchfehler == nil {
		t.Error("ohne Zuhause glückte eine Suche")
	}
}
