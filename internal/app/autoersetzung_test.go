package app_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Die Naht des Autoersetzers — und die Umrechnung, die sie leisten muss.
//
// # Warum die Umrechnung hier steht und nicht im Fenster
//
// Go zählt **Bytes**, JavaScript zählt **UTF-16-Einheiten**. `selectionStart`
// eines Eingabefeldes ist eine UTF-16-Einheit; `service.Autoersetzer` rechnet
// in Bytes. Wer die Umrechnung dem Fenster überließe, müsste sie dort in
// TypeScript nachbauen — und zwei Zählweisen an zwei Orten laufen auseinander,
// sobald jemand ein Zeichen außerhalb der Grundebene tippt
// (#897).

// **Die Stelle kommt und geht in UTF-16-Einheiten.**
func TestDieStelleWirdZwischenZaehlweisenUmgerechnet(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitAutoersetzung(ersetzungsprobe{regeln: []msepaket.AutoErsetzung{
		{Muster: "(C)", Ersatz: "©", GanzesWort: true},
	}})

	faelle := []struct {
		name    string
		text    string
		stelle  int
		nachher string
		danach  int
	}{
		{"reines ASCII", "(C)", 3, "©", 1},

		// **Ein Emoji wiegt in UTF-16 zwei Einheiten und in UTF-8 vier Bytes.**
		// Wer die Stelle ungerechnet durchreichte, landete mitten im Zeichen.
		{"hinter einem Emoji", "🂡(C)", 5, "🂡©", 3},
		{"hinter einem Umlaut", "ä(C)", 4, "ä©", 2},

		// **Die Marke steht mitten im Text**, nicht am Ende — der Fall, in dem
		// die Umrechnung die Schleife wirklich durchläuft.
		{"mitten im Text", "(C)x", 3, "©x", 1},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			antwort := anwendung.AutoErsetzen(context.Background(), "1", fall.text, fall.stelle)

			if antwort.Fehler != nil {
				t.Fatalf("Fehler: %+v", antwort.Fehler)
			}

			if antwort.Text != fall.nachher || antwort.Stelle != fall.danach {
				t.Errorf("%q@%d → %q@%d, erwartet %q@%d",
					fall.text, fall.stelle, antwort.Text, antwort.Stelle,
					fall.nachher, fall.danach)
			}
		})
	}
}

// **Eine unbrauchbare Setkennung wird benannt abgewiesen** (§8).
func TestEineUnbrauchbareKennungWirdAbgewiesen(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitAutoersetzung(ersetzungsprobe{})

	antwort := anwendung.AutoErsetzen(context.Background(), "keine Zahl", "(C)", 3)
	if antwort.Fehler == nil {
		t.Fatal("die Kennung kam durch")
	}

	// **Das Feld heißt `id`** — so benennt `kennungAus` jede Kennung, und
	// eine zweite Schreibweise hier wäre eine zweite Wahrheit.
	if antwort.Fehler.Feld != "id" {
		t.Errorf("der Fehler nennt das Feld %q", antwort.Fehler.Feld)
	}
}

// **Und ein Fehler beim Holen der Regeln lässt den Text stehen.**
//
// Ein Spielpaket, das nicht daliegt, ist der Normalfall auf einem frischen
// Rechner. Wer dabei den Text verlöre, machte aus einer fehlenden Vorlage
// einen Datenverlust.
func TestOhneRegelnBleibtDerTextStehen(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitAutoersetzung(ersetzungsprobe{fehler: context.Canceled})

	antwort := anwendung.AutoErsetzen(context.Background(), "1", "(C)", 3)
	if antwort.Fehler == nil {
		t.Fatal("der Fehler wurde verschluckt")
	}

	if antwort.Text != "(C)" || antwort.Stelle != 3 {
		t.Errorf("%q@%d — Text und Stelle müssen unverändert zurückkommen",
			antwort.Text, antwort.Stelle)
	}
}

// ersetzungsprobe ist das Double der Regelquelle.
type ersetzungsprobe struct {
	regeln []msepaket.AutoErsetzung
	fehler error
}

func (e ersetzungsprobe) Autoersetzungen(context.Context, int64) (
	[]msepaket.AutoErsetzung, error,
) {
	return e.regeln, e.fehler
}

// **Und die Bindung reicht durch, die Wails hinhält.**
//
// Die Naht, nicht die Rechnung: dass `Bindungen.AutoErsetzen` dieselbe Antwort
// gibt wie `App.AutoErsetzen`, ist das, was das Fenster erreicht (§10).
func TestDieBindungReichtDieErsetzungDurch(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitAutoersetzung(ersetzungsprobe{regeln: []msepaket.AutoErsetzung{
		{Muster: "(C)", Ersatz: "©", GanzesWort: true},
	}})

	antwort := app.NeueBindungen(anwendung).AutoErsetzen(t.Context(), "1", "(C)", 3)
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if antwort.Text != "©" || antwort.Stelle != 1 {
		t.Errorf("%q@%d, erwartet \"©\"@1", antwort.Text, antwort.Stelle)
	}
}

// **Eine Stelle außerhalb des Textes wird nicht gerechnet** (§8).
//
// Sie kommt aus dem Fenster, und was von dort kommt, ist fremde Eingabe. Ein
// Sonderfall ist die Stelle **mitten in einem Ersatzpaar**: die gibt es in
// UTF-16 nicht, und sie zu runden hieße, still etwas anderes zu meinen als das
// Fenster.
func TestEineUnmoeglicheStelleWirdNichtGerechnet(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitAutoersetzung(ersetzungsprobe{regeln: []msepaket.AutoErsetzung{
		{Muster: "(C)", Ersatz: "©"},
	}})

	faelle := map[string]struct {
		text   string
		stelle int
	}{
		"vor dem Anfang":             {"(C)", -1},
		"hinter dem Ende":            {"(C)", 9},
		"mitten in einem Ersatzpaar": {"🂡", 1},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			antwort := anwendung.AutoErsetzen(
				context.Background(), "1", fall.text, fall.stelle)

			if antwort.Text != fall.text || antwort.Stelle != fall.stelle {
				t.Errorf("%q@%d — Text und Stelle müssen stehen bleiben",
					antwort.Text, antwort.Stelle)
			}
		})
	}
}

// **Die Rückrechnung kappt eine Stelle hinter dem Text.**
//
// Sie kann nicht von außen kommen — der Ersetzer gibt nur Stellen in seinem
// eigenen Ergebnis. Der Wächter steht trotzdem da, weil eine Stelle hinter dem
// Text eine Bereichsverletzung wäre und kein falsches Ergebnis.
func TestDieRueckrechnungKapptHinterDemText(t *testing.T) {
	if gerechnet := app.Utf16StelleFuerTest("ab", 99); gerechnet != 2 {
		t.Errorf("%d, erwartet 2", gerechnet)
	}
}

// **Eine eigene Regel greift beim Tippen** — die Abnahme aus
// #1341.
//
// # Warum dieser Fall hier steht und nicht im Dienst
//
// Weil er die **Naht** prüft: `AutoErsetzen` fragt eine Quelle nach den
// Regeln, und seit #1341 ist das die Pflege — nicht mehr der Stilwähler.
// Ohne diese Verdrahtung wäre eine eigene Regel eine Zeile im Bestand, die
// nichts tut, und **jeder Test einer Schicht für sich bliebe grün**.
func TestEineEigeneRegelGreiftBeimTippen(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")

	// **Der Weg über die Pflege**, wie ihn `werkzeugnaehte.go` verdrahtet.
	anwendung.MitAutoersetzung(app.NeuerAutoersetzungsweg(
		service.NeueAutoersetzungspflege(
			&ersetzablageprobe{regeln: []domain.Autoersetzungsregel{
				{Muster: "(W)", Ersatz: "weiss", GanzesWort: true, Aktiv: true},
			}, an: true},
			&spielprobeErsetzungApp{spiel: "probe"})))

	antwort := anwendung.AutoErsetzen(t.Context(), "1", "Text (W)", 8)
	if antwort.Fehler != nil {
		t.Fatalf("%v", antwort.Fehler)
	}

	if antwort.Text != "Text weiss" {
		t.Errorf("heraus kam %q", antwort.Text)
	}
}

// **Und der ausgeschaltete Schalter hält sie an.**
//
//	if (!gs.use_auto_replace) return false;
//
// `src/gui/value/text.cpp`, Zeile 1034, Fassung f9d9356 vom 2021-07-06.
func TestDerAusgeschalteteSchalterHaeltDieErsetzungAn(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitAutoersetzung(app.NeuerAutoersetzungsweg(
		service.NeueAutoersetzungspflege(
			&ersetzablageprobe{regeln: []domain.Autoersetzungsregel{
				{Muster: "(W)", Ersatz: "weiss", GanzesWort: true, Aktiv: true},
			}, an: false},
			&spielprobeErsetzungApp{spiel: "probe"})))

	antwort := anwendung.AutoErsetzen(t.Context(), "1", "Text (W)", 8)
	if antwort.Fehler != nil {
		t.Fatalf("%v", antwort.Fehler)
	}

	if antwort.Text != "Text (W)" {
		t.Errorf("heraus kam %q — der Schalter stand auf aus", antwort.Text)
	}
}

// ersetzablageprobe ist die Ablage, die die Pflege braucht.
type ersetzablageprobe struct {
	regeln []domain.Autoersetzungsregel
	an     bool
}

func (e *ersetzablageprobe) Regeln(context.Context, string) (
	[]domain.Autoersetzungsregel, error,
) {
	return e.regeln, nil
}

func (e *ersetzablageprobe) Setzen(
	context.Context, string, domain.Autoersetzungsregel,
) error {
	return nil
}

func (e *ersetzablageprobe) Loeschen(context.Context, string, string) error { return nil }

func (e *ersetzablageprobe) Zuruecksetzen(context.Context, string) error { return nil }

func (e *ersetzablageprobe) Schalter(context.Context, string) (bool, error) {
	return e.an, nil
}

func (e *ersetzablageprobe) SchalterSetzen(context.Context, string, bool) error {
	return nil
}

// spielprobeErsetzungApp gibt Spielnamen und Paketregeln.
type spielprobeErsetzungApp struct{ spiel string }

func (s *spielprobeErsetzungApp) SpielVon(context.Context, int64) (string, error) {
	return s.spiel, nil
}

func (s *spielprobeErsetzungApp) Paketregeln(context.Context, int64) (
	[]msepaket.AutoErsetzung, error,
) {
	return nil, nil
}
