package domain_test

// **Eine Kostentafel für jede Fläche** — Folgebefund F1 des Lastlaufs QA-08
// (#2003).
//
// Bis zum 2026-10-02 buchte REST nach einer kurzen Liste von Wegen und MCP
// nach Wörtern im Namen eines Werkzeugs. Dieselbe Fähigkeit kostete deshalb
// je Fläche etwas anderes: eine Karte zeichnen an beiden 2 statt 5, ein Bogen
// über REST 2 und über MCP 8 — und eine Vorschau der Einstellungen über MCP 5,
// weil ihr Name „vorschau" enthält.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Was die Tafel sagt, gilt an jedem ihrer Wege** — gleich, welche Art der
// Weg an seiner Fläche hat.
func TestDieKostentafelGiltAnJedemIhrerWege(t *testing.T) {
	t.Parallel()

	if len(domain.Kostentafel()) == 0 {
		t.Fatal("die Kostentafel ist leer")
	}

	for _, zeile := range domain.Kostentafel() {
		for flaeche, wege := range zeile.Wege {
			for _, weg := range wege {
				for _, schreibend := range []bool{false, true} {
					if klasse := domain.KostenklasseFuer(flaeche, weg, schreibend); klasse != zeile.Klasse {
						t.Errorf("%s %q (%s): %s, die Tafel sagt %s",
							flaeche, weg, zeile.Faehigkeit, klasse, zeile.Klasse)
					}
				}
			}
		}
	}
}

// **Die teuren Operationen aus dem Befund** — an REST, MCP und der
// Kommandozeile dieselbe Klasse.
func TestDieTeurenOperationenKostenUeberallDasselbe(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		was    string
		wege   map[domain.Flaeche]string
		klasse domain.Kostenklasse
	}{
		{"eine Karte zeichnen", map[domain.Flaeche]string{
			domain.FlaecheREST: "POST /karten/{}/bild",
			domain.FlaecheMCP:  "schreiben.karte.zeichnen",
			domain.FlaecheCLI:  "mse zeichnen",
		}, domain.KostenRendern},
		{"einen Bogen legen", map[domain.Flaeche]string{
			domain.FlaecheREST: "POST /sets/{}/bogen",
			domain.FlaecheMCP:  "schreiben.set.bogen",
			domain.FlaecheCLI:  "mse bogen",
		}, domain.KostenAusfuhr},
		{"MTG JSON schreiben", map[domain.Flaeche]string{
			domain.FlaecheREST: "POST /sets/{}/mtgjson",
			domain.FlaecheMCP:  "schreiben.mtgjson.schreiben",
			domain.FlaecheCLI:  "mse mtgjson-schreiben",
		}, domain.KostenAusfuhr},
		{"ein Paket einlesen", map[domain.Flaeche]string{
			domain.FlaecheREST: "POST /pakete/eingelesen",
			domain.FlaecheMCP:  "schreiben.paket.einlesen",
			domain.FlaecheCLI:  "mse installer-einlesen",
		}, domain.KostenEinfuhr},
		{"eine Sammlung einlesen", map[domain.Flaeche]string{
			domain.FlaecheREST: "POST /sammlung/einlesen",
			domain.FlaecheMCP:  "schreiben.sammlung.einlesen",
			domain.FlaecheCLI:  "mse sammlung-einlesen",
		}, domain.KostenEinfuhr},
	}

	for _, fall := range faelle {
		for flaeche, weg := range fall.wege {
			if klasse := domain.KostenklasseFuer(flaeche, weg, true); klasse != fall.klasse {
				t.Errorf("%s über %s (%s): %s, erwartet %s", fall.was, flaeche, weg, klasse, fall.klasse)
			}
		}
	}
}

// **Was nicht in der Tafel steht, kostet nach seiner Art** — und ein Wort im
// Namen macht nichts teuer.
func TestWasNichtInDerTafelStehtKostetNachSeinerArt(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name       string
		flaeche    domain.Flaeche
		weg        string
		schreibend bool
		will       domain.Kostenklasse
	}{
		{"ein lesendes Werkzeug", domain.FlaecheMCP, "lesen.sets", false, domain.KostenLesen},
		{"ein schreibendes Werkzeug", domain.FlaecheMCP, "schreiben.feld", true, domain.KostenSchreiben},
		{"ein lesender Weg", domain.FlaecheREST, "GET /sets", false, domain.KostenLesen},
		{"ein schreibender Weg", domain.FlaecheREST, "PUT /sets/{}/karten/{}/felder/{}", true, domain.KostenSchreiben},
		{"»vorschau« rastert nichts", domain.FlaecheMCP, "lesen.einstellungen.vorschau", false, domain.KostenLesen},
		{"»ausfuhr« der Einstellungen ist klein", domain.FlaecheMCP, "lesen.einstellungen.ausfuhr", false, domain.KostenLesen},
		{"ein Turnier ist kein Set", domain.FlaecheMCP, "schreiben.turnier.einlesen", true, domain.KostenSchreiben},
		{"ein Profilbild wegnehmen", domain.FlaecheMCP, "schreiben.nutzer.bild.loeschen", true, domain.KostenSchreiben},
		{"ein unbekannter Weg", domain.FlaecheREST, "", false, domain.KostenLesen},
	}

	for _, fall := range faelle {
		if klasse := domain.KostenklasseFuer(fall.flaeche, fall.weg, fall.schreibend); klasse != fall.will {
			t.Errorf("%s: %s, erwartet %s", fall.name, klasse, fall.will)
		}
	}
}

// **Ein Abbruch kostet nichts** — er ist die eine Anfrage, die Last nimmt.
func TestEinAbbruchKostetNichts(t *testing.T) {
	t.Parallel()

	if kosten := domain.KostenAbbruch.Kosten(); kosten != 0 {
		t.Errorf("ein Abbruch kostet %d", kosten)
	}

	for _, weg := range []string{"DELETE /auftraege/{}", "DELETE /index/aufbau"} {
		if klasse := domain.KostenklasseFuer(domain.FlaecheREST, weg, true); klasse != domain.KostenAbbruch {
			t.Errorf("%s: %s", weg, klasse)
		}
	}
}

// **Die Tafel als Auskunft** — so, wie jede Fläche sie in `grenzen` nennt:
// mit den Kosten der Klasse und leeren Listen statt keiner.
func TestDieKostentafelAlsAuskunft(t *testing.T) {
	t.Parallel()

	auskunft := domain.KostentafelAlsAuskunft()
	if len(auskunft) != len(domain.Kostentafel()) {
		t.Fatalf("%d Zeilen, die Tafel hat %d", len(auskunft), len(domain.Kostentafel()))
	}

	for stelle, zeile := range domain.Kostentafel() {
		gesagt := auskunft[stelle]
		if gesagt.Faehigkeit != zeile.Faehigkeit || gesagt.Klasse != string(zeile.Klasse) ||
			gesagt.Kosten != zeile.Klasse.Kosten() || gesagt.REST == nil || gesagt.MCP == nil || gesagt.CLI == nil {
			t.Errorf("Zeile %d: %+v", stelle, gesagt)
		}
	}

	if erste := auskunft[0]; erste.Kosten != 5 || erste.CLI[0] != "mse zeichnen" {
		t.Errorf("die erste Zeile: %+v", erste)
	}
}

// **Ein Weg kann unter mehreren Schreibweisen gefragt werden** — #2020.
//
// Der Dienst trägt die Rechnungen des Symboleditors als **ein** Muster ein,
// `POST /symbol/{rechnung}`; `FAEHIGKEITEN.md` und die Tafel nennen jede mit
// ihrem Wert, `POST /symbol/grundform`. Wer nur das Muster fragte, fände keine.
// Der erste Weg, den die Tafel kennt, entscheidet; kennt sie keinen, entscheidet
// die Art.
func TestEinWegKannUnterMehrerenSchreibweisenInDerTafelStehen(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name       string
		wege       []string
		schreibend bool
		will       domain.Kostenklasse
	}{
		{"nur der Wert steht in der Tafel", []string{"POST /symbol/{}", "POST /symbol/grundform"}, true, domain.KostenLesen},
		{"nur das Muster steht in der Tafel", []string{"POST /karten/{}/bild", "POST /karten/7/bild"}, true, domain.KostenRendern},
		{"der erste Treffer entscheidet", []string{"POST /karten/{}/bild", "POST /symbol/grundform"}, true, domain.KostenRendern},
		{"keiner steht darin, schreibend", []string{"POST /nirgends/{}", "POST /nirgends/eins"}, true, domain.KostenSchreiben},
		{"keiner steht darin, lesend", []string{"GET /nirgends/{}", "GET /nirgends/eins"}, false, domain.KostenLesen},
		{"gar keine Schreibweise, schreibend", nil, true, domain.KostenSchreiben},
	}

	for _, fall := range faelle {
		if klasse := domain.KostenklasseFuerSchreibweisen(domain.FlaecheREST, fall.schreibend, fall.wege...); klasse != fall.will {
			t.Errorf("%s: %s, erwartet %s", fall.name, klasse, fall.will)
		}
	}

	// **Eine Schreibweise ist dasselbe wie der Weg selbst.**
	if domain.KostenklasseFuer(domain.FlaecheREST, "POST /symbol/grundform", true) !=
		domain.KostenklasseFuerSchreibweisen(domain.FlaecheREST, true, "POST /symbol/grundform") {
		t.Error("KostenklasseFuer und KostenklasseFuerSchreibweisen antworten verschieden")
	}
}
