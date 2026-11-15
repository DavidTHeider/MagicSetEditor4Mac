package service_test

// **Die Läufe über die Spieltexturen** — Werkstück G2 von Phase 5.5
// (#1459).
//
// **Die Vorrichtungen stehen nebenan** in `texturwerkstatt_test.go`; diese
// Datei ging sonst über die Zeilengrenze aus §2.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Siebenmal dieselbe Karte wird einmal gezeichnet.**
//
// **Die Abnahme des Issues, wörtlich**: „Sieben gleiche Karten nutzen eine
// Textur." Gezählt wird der **Rasterlauf** und nicht der Ruf — ein Dienst, der
// siebenmal zeichnet und dieselbe Adresse gäbe, käme sonst durch.
func TestSiebenGleicheKartenWerdenEinmalGezeichnet(t *testing.T) {
	t.Parallel()

	werk, _, rasterer, _ := eineWerkstatt(t)

	erste, fehler := werk.Textur(t.Context(), eineTexturkarte(1),
		service.SeiteVorne, service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	for range 6 {
		weitere, fehler := werk.Textur(t.Context(), eineTexturkarte(1),
			service.SeiteVorne, service.GueteTisch)
		if fehler != nil {
			t.Fatalf("Textur: %v", fehler)
		}

		if weitere.Adresse != erste.Adresse {
			t.Errorf("zwei Adressen: %q und %q", erste.Adresse, weitere.Adresse)
		}
	}

	if rasterer.gezeichnet != 1 {
		t.Errorf("%dmal gezeichnet, erwartet einmal", rasterer.gezeichnet)
	}
}

// **Und dieselbe Kartenkennung mit anderem Inhalt trifft nicht.**
//
// **Der Fall, an dem ein Cache über der Zeilennummer zerbricht** — das Issue
// verlangt ihn ausdrücklich.
func TestDieselbeKennungMitAnderemInhaltTrifftNicht(t *testing.T) {
	t.Parallel()

	werk, plaene, rasterer, _ := eineWerkstatt(t)

	erste, fehler := werk.Textur(t.Context(), eineTexturkarte(7),
		service.SeiteVorne, service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	// **Dieselbe Kennung, anderer Inhalt** — der Plan zeichnet etwas anderes.
	plaene.plan = texturplan("rahmen.png", "Beleren")

	zweite, fehler := werk.Textur(t.Context(), eineTexturkarte(7),
		service.SeiteVorne, service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	if erste.Adresse == zweite.Adresse {
		t.Errorf("dieselbe Adresse für zwei Inhalte: %q", erste.Adresse)
	}

	if rasterer.gezeichnet != 2 {
		t.Errorf("%dmal gezeichnet, erwartet zweimal", rasterer.gezeichnet)
	}
}

// **Die Güte entscheidet über die Vergrößerung und die Rasterstufe.**
func TestDieGueteEntscheidetUeberDieVergroesserung(t *testing.T) {
	t.Parallel()

	werk, _, rasterer, _ := eineWerkstatt(t)

	if _, fehler := werk.Textur(t.Context(), eineTexturkarte(1),
		service.SeiteVorne, service.GueteMiniatur); fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	miniatur := rasterer.breite

	if _, fehler := werk.Textur(t.Context(), eineTexturkarte(1),
		service.SeiteVorne, service.GueteAusfuhr); fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	if rasterer.breite <= miniatur {
		t.Errorf("Ausfuhr ist %d breit, Miniatur %d", rasterer.breite, miniatur)
	}
}

// **Eine Textur nennt ihre Maße und ihren Schlüssel.**
//
// **Den Schlüssel, damit das Fenster teilen kann**: zwei Objekte mit demselben
// Schlüssel sind dieselbe Textur, und das ist die Frage, die eine
// Referenzzählung beantwortet (#1459).
func TestEineTexturNenntIhreMasseUndIhrenSchluessel(t *testing.T) {
	t.Parallel()

	werk, _, _, _ := eineWerkstatt(t)

	textur, fehler := werk.Textur(t.Context(), eineTexturkarte(1),
		service.SeiteVorne, service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	switch {
	case textur.Schluessel == "":
		t.Error("die Textur hat keinen Schlüssel")
	case textur.Breite <= 0 || textur.Hoehe <= 0:
		t.Errorf("Maße %d×%d", textur.Breite, textur.Hoehe)
	case textur.Adresse == "":
		t.Error("die Textur hat keine Adresse")
	case textur.Guete != service.GueteTisch:
		t.Errorf("Güte %q", textur.Guete)
	}
}

// **Ein Plan ohne Breite gibt die Vorlage in ihren Punkten.**
//
// **Der Zweig gegen die Division durch null** — ein Behelfsplan kann ohne
// Maße dastehen, und ein Rasterer mit dem Faktor `+Inf` zeichnete nichts, das
// ein Mensch deuten könnte.
func TestEinPlanOhneBreiteBekommtDenFaktorEins(t *testing.T) {
	t.Parallel()

	plaene := &planprobe{plan: &render.Plan{}}
	rasterer := &rasterprobe{}

	werk := service.NeueSpieltexturen(plaene, rasterer,
		ressourcenprobe{}, service.NeueTexturfachprobe())

	textur, fehler := werk.Textur(t.Context(), eineTexturkarte(1),
		service.SeiteVorne, service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	// **Der Rasterer bekam den Faktor eins.**
	//
	// **Gemessen an der Vergrößerung selbst und nicht an der Breite des
	// Bildes**: ein Plan ohne Breite ergibt auch mal eins genommen kein
	// Bild, und wer daraus den Faktor läse, läse aus einer Null.
	if rasterer.vergroesserung != 1 {
		t.Errorf("gezeichnet mit dem Faktor %v, erwartet eins",
			rasterer.vergroesserung)
	}

	// **Und die Maße sind die des Bildes** — was der Rasterer aus einem Plan
	// ohne Maße macht, entscheidet er; diese Naht gibt es weiter.
	if textur.Breite != rasterer.breite || textur.Hoehe != rasterer.hoehe {
		t.Errorf("Maße %d×%d, gezeichnet %d×%d",
			textur.Breite, textur.Hoehe, rasterer.breite, rasterer.hoehe)
	}
}

// **Die Befunde des Plans stehen an der Textur — auch beim Treffer.**
//
// **Auch beim Treffer**, denn sie sagen, *warum* eine Karte behilflich
// aussieht; das gilt beim zweiten Blick so wie beim ersten. Ein Cache, der
// sie verschwiege, machte aus einem erklärten Mangel einen unerklärten.
func TestDieBefundeStehenAuchBeimTreffer(t *testing.T) {
	t.Parallel()

	plaene := &planprobe{
		plan:    texturplan("rahmen.png", "MPlantin"),
		befunde: []domain.Befund{{Datei: "stil.mse-style", Meldung: "fehlt"}},
	}

	// **Und der Rasterer hat auch einen** — die Befunde des Plans sagen,
	// *warum* eine Karte behilflich aussieht, die des Rasterers, was beim
	// Zeichnen fehlte. Beide gehören an die Textur, und der Lauf hielte ohne
	// den zweiten nur die Hälfte.
	rasterer := &rasterprobe{
		befunde: []render.Befund{{Meldung: "die Schrift fehlt"}},
	}

	werk := service.NeueSpieltexturen(plaene, rasterer,
		ressourcenprobe{
			bilder:    map[string]string{"rahmen.png": "bild-1"},
			schriften: map[string]string{"MPlantin": "schrift-1"},
		}, service.NeueTexturfachprobe())

	erste, fehler := werk.Textur(t.Context(), eineTexturkarte(1),
		service.SeiteVorne, service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	zweite, fehler := werk.Textur(t.Context(), eineTexturkarte(1),
		service.SeiteVorne, service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	// **Beim ersten Mal beide**, beim Treffer die des Plans — der Rasterer
	// lief dort gar nicht.
	if len(erste.Befunde) != 2 || len(zweite.Befunde) != 1 {
		t.Errorf("Befunde: %v und %v", erste.Befunde, zweite.Befunde)
	}
}

// **Ein Treffer nennt dieselben Maße wie der Fehltreffer.**
//
// # Wofür es diesen Lauf gibt
//
// **Weil der Treffer bis zum 2026-09-15 gerechnet hat, statt zu lesen.** Er
// nahm die Zielbreite der Güte und leitete die Höhe aus dem
// Seitenverhältnis des Plans ab; der Fehltreffer nahm die Maße des
// **gezeichneten Bildes**. Beide Wege gaben fast dieselben Zahlen — und
// „fast" heißt hier: derselbe Rücken meldete 512 × 512 und 512 × 714.
//
// **Das Issue sagt „unveränderter Inhalt trifft ihn."** Ein Treffer, der
// andere Zahlen nennt als der Fehltreffer, ist keiner — das Fenster legte
// die Karte beim zweiten Blick in ein anderes Kästchen.
//
// **Also führt das Fach die Maße mit**, und dieser Lauf hält es fest.
func TestEinTrefferNenntDieselbenMasseWieDerFehltreffer(t *testing.T) {
	t.Parallel()

	werk, _, _, _ := eineWerkstatt(t)

	erste, fehler := werk.Textur(t.Context(), eineTexturkarte(1),
		service.SeiteVorne, service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	zweite, fehler := werk.Textur(t.Context(), eineTexturkarte(1),
		service.SeiteVorne, service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	if erste.Breite != zweite.Breite || erste.Hoehe != zweite.Hoehe {
		t.Errorf("Fehltreffer %d×%d, Treffer %d×%d",
			erste.Breite, erste.Hoehe, zweite.Breite, zweite.Hoehe)
	}

	// **Und die Zahlen sind nicht zufällig gleich, weil beide null wären.**
	if erste.Breite <= 0 || erste.Hoehe <= 0 {
		t.Errorf("Maße %d×%d", erste.Breite, erste.Hoehe)
	}

	// **Die Gegenprobe gegen ein Quadrat** (§10): wären Breite und Höhe
	// gleich, hielte dieser Lauf auch eine Vorrichtung für richtig, die die
	// Gestalt des Plans nicht nachbildet — und genau die hat den Fehler
	// verdeckt.
	if erste.Breite == erste.Hoehe {
		t.Errorf("%d×%d ist quadratisch — die Vorrichtung bildet den Plan "+
			"nicht nach, und dieser Lauf misst nichts", erste.Breite,
			erste.Hoehe)
	}
}

// **Was auf dem Weg klagt, kommt als Fehler zurück.**
//
// **Je Weg gestolpert und nicht auf jedem** (§10): eine Vorrichtung, die
// überall klagt, klagt schon beim ersten — und alles danach bleibt ungemessen.
//
// **Die Ressourcenquelle steht hier nicht.** Sie klagt nicht: eine fehlende
// Datei ist ein Wert, und der Grund steht an [service.Texturschluessel].
func TestWasAufDemTexturwegKlagtKommtDurch(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("der Bestand ist zu")

	for name, stellen := range map[string]func(*planprobe, *rasterprobe){
		"der Planbauer": func(p *planprobe, _ *rasterprobe) { p.klagt = kaputt },
		"der Rasterer":  func(_ *planprobe, r *rasterprobe) { r.klagt = kaputt },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			plaene := &planprobe{plan: texturplan("rahmen.png", "MPlantin")}
			rasterer := &rasterprobe{}

			stellen(plaene, rasterer)

			werk := service.NeueSpieltexturen(plaene, rasterer,
				ressourcenprobe{}, service.NeueTexturfachprobe())

			_, fehler := werk.Textur(t.Context(), eineTexturkarte(1),
				service.SeiteVorne, service.GueteTisch)

			if !errors.Is(fehler, kaputt) {
				t.Errorf("Fehler %v, erwartet die Klage von %s", fehler, name)
			}
		})
	}
}
