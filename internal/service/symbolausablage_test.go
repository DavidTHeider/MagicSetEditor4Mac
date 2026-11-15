package service_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein frisch gezeichnetes Symbol steht in der Vorschau** —
// #890,
// Werkstück 4 von Phase X.
//
// # Warum es dafür etwas Eigenes braucht
//
// Eine gezeichnete `.mse-symbol` liegt im **Bildverzeichnis** des Sets — sie
// kommt erst beim Sichern ins Paket. Die Symbolquelle des Planbauers kennt nur
// Pakete; ein Verweis auf eine Datei, die es dort noch nicht gibt, gäbe eine
// leere Gestalt, und die Karte zeigte MSEs Standardquadrat.
//
// **Das ist derselbe Weg, den ein eingesetztes Bild schon nimmt**:
// `Vorlage.Bilderverzeichnis` steht neben `Vorlage.Setpaket`, und `bildebene`
// bekommt beide. Der Symbolweg bekam bis hierher nur das Paket.
func TestEineGezeichneteZeichnungStehtVorDemSichern(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()

	zeichnung := msepaket.Symbolzeichnung{Teile: []msepaket.Symbolteil{{
		Art: msepaket.TeilForm, Name: "Raute", Kombination: msepaket.KombinationVereinen,
		Punkte: []msepaket.Kontrollpunkt{
			{Stelle: msepaket.Vektor{X: 0.5, Y: 0.05}},
			{Stelle: msepaket.Vektor{X: 0.95, Y: 0.5}},
			{Stelle: msepaket.Vektor{X: 0.5, Y: 0.95}},
		},
	}}}

	if err := os.WriteFile(filepath.Join(ort, "symbol1.mse-symbol"),
		msepaket.SchreibeSymbolzeichnung(zeichnung), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	quelle := service.SymbolquelleAusAblage(ort, nil, "")

	hat := quelle.Symbolgestalt("symbol1.mse-symbol")
	if len(hat.Teile) != 1 || hat.Teile[0].Name != "Raute" {
		t.Fatalf("die Gestalt ist %+v", hat.Teile)
	}

	if len(hat.Teile[0].Punkte) != 3 {
		t.Errorf("%d Punkte", len(hat.Teile[0].Punkte))
	}
}

// **Was nicht in der Ablage liegt, geht dahinter weiter.**
//
// Ein Set, dessen Symbol schon im Paket steht, darf nicht dadurch verlieren,
// dass jemand die Anwendung gestartet hat.
func TestWasNichtInDerAblageLiegtGehtDahinterWeiter(t *testing.T) {
	t.Parallel()

	dahinter := &symbolquellenprobe{gestalt: msepaket.Symbolzeichnung{
		Teile: []msepaket.Symbolteil{{Art: msepaket.TeilForm, Name: "Aus dem Paket"}},
	}}

	quelle := service.SymbolquelleAusAblage(t.TempDir(), dahinter, "")

	hat := quelle.Symbolgestalt("/irgendein.mse-set/symbol1")
	if len(hat.Teile) != 1 || hat.Teile[0].Name != "Aus dem Paket" {
		t.Fatalf("die Gestalt ist %+v", hat.Teile)
	}

	if dahinter.verweis != "/irgendein.mse-set/symbol1" {
		t.Errorf("dahinter gefragt wurde nach %q", dahinter.verweis)
	}
}

// **Ohne Ablage und ohne Quelle dahinter bleibt die Gestalt leer.**
//
// Beides kommt vor: eine Anwendung ohne Benutzerverzeichnis, und ein
// Ausfuhrlauf ohne Bildquelle. Ein Absturz an dieser Stelle nähme der Karte
// alles andere mit (§8).
func TestOhneAblageUndOhneQuelleBleibtDieGestaltLeer(t *testing.T) {
	t.Parallel()

	for name, quelle := range map[string]service.Symbolquelle{
		"ohne beides":         service.SymbolquelleAusAblage("", nil, ""),
		"ohne Quelle":         service.SymbolquelleAusAblage(t.TempDir(), nil, ""),
		"Datei gibt es nicht": service.SymbolquelleAusAblage(t.TempDir(), nil, ""),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if hat := quelle.Symbolgestalt("gibtsnicht.mse-symbol"); len(hat.Teile) != 0 {
				t.Errorf("die Gestalt ist %+v", hat.Teile)
			}
		})
	}
}

// symbolquellenprobe merkt sich, wonach gefragt wurde.
type symbolquellenprobe struct {
	gestalt msepaket.Symbolzeichnung
	verweis string
}

func (s *symbolquellenprobe) Symbolgestalt(verweis string) msepaket.Symbolzeichnung {
	s.verweis = verweis

	return s.gestalt
}

// **Und der ganze Weg trägt: ein gezeichnetes Symbol steht im Plan.**
//
// Das ist die Abnahme von Werkstück 4 und der Grund, warum die vorigen Läufe
// nicht genügen: sie prüfen die Quelle. **Dieser prüft, dass der Planbauer sie
// benutzt** — und genau die Zeile hat bis zum 2026-08-25 gefehlt.
//
// **Ohne sie wäre er nicht rot, sondern falsch grün**: `symbol_variation`
// setzt bei leerem Verweis MSEs Standardsymbol ein, ein Quadrat mit **vier**
// Punkten. Deshalb hat die Raute hier **drei** — eine Zahl, die das Quadrat
// nicht haben kann.
func TestEinGezeichnetesSymbolStehtImPlan(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()

	raute := msepaket.Symbolzeichnung{Teile: []msepaket.Symbolteil{{
		Art: msepaket.TeilForm, Name: "Raute", Kombination: msepaket.KombinationVereinen,
		Punkte: []msepaket.Kontrollpunkt{
			{Stelle: msepaket.Vektor{X: 0.5, Y: 0.05}},
			{Stelle: msepaket.Vektor{X: 0.95, Y: 0.5}},
			{Stelle: msepaket.Vektor{X: 0.05, Y: 0.5}},
		},
	}}}

	if err := os.WriteFile(filepath.Join(ort, "symbol1.mse-symbol"),
		msepaket.SchreibeSymbolzeichnung(raute), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
			Symbolfassungen: []msepaket.Symbolfassung{
				{Feld: "symbol", Name: "rare", Fuellart: "solid"},
			},
			Felder: []msepaket.Feldstil{{
				Feld: "rarity", Links: "0", Oben: "0", Breite: "40", Hoehe: "20",
				Darstellung: "image",
				Bild:        `{ symbol_variation(symbol: set.symbol, variation: "rare") }`,
			}},
		},
		Setfelder:         []domain.Field{{Name: "symbol", Kind: domain.KindImage}},
		Setangaben:        []domain.Value{{Field: "symbol", Text: "symbol1.mse-symbol"}},
		Bilderverzeichnis: ort,
	}

	plan, befunde, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), &domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	ebene := bildebeneAus(t, plan)
	if len(ebene.Bild.Symbol) != 1 {
		t.Fatalf("die Symbolgestalt ist %+v", ebene.Bild.Symbol)
	}

	if hat := len(ebene.Bild.Symbol[0].Punkte); hat != 3 {
		t.Errorf("die Gestalt hat %d Punkte — vier wären MSEs Standardquadrat", hat)
	}
}

// **Ein Symbol, das im Setpaket liegt, kommt auch an** —
// #890.
//
// # Der Befund
//
// `set.symbol` trägt in einer echten Datei einen **schlichten** Namen —
// `symbol1.mse-symbol` —, und die Datei liegt im `.mse-set` daneben. Die
// Ablagenquelle sah nur in der Anwendungsablage nach und reichte denselben
// rohen Namen weiter; die Paketquelle verlangt aber einen Setpaketverweis oder
// einen Pfad mit führendem `/`.
//
// **Ein importiertes Set zeigte damit sein Symbol nicht** — nur ein hier
// gezeichnetes, weil das in der Ablage liegt.
//
// Angestrichen vom Zweitblick über Phase X. Er hat dazu angemerkt, dass der
// vorige Rückfalltest mit `/irgendein.mse-set/symbol1` einen **künstlich
// qualifizierten** Wert benutzte und den Betriebsfall damit verfehlte.
func TestEinSymbolAusDemSetpaketKommtAn(t *testing.T) {
	t.Parallel()

	setpaket := t.TempDir()

	zeichnung := msepaket.Symbolzeichnung{Teile: []msepaket.Symbolteil{{
		Art: msepaket.TeilForm, Name: "Aus dem Paket",
		Punkte: []msepaket.Kontrollpunkt{{Stelle: msepaket.Vektor{X: 0.5, Y: 0.5}}},
	}}}

	if err := os.WriteFile(filepath.Join(setpaket, "symbol1.mse-symbol"),
		msepaket.SchreibeSymbolzeichnung(zeichnung), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	// **Die echte Bildquelle und kein Double.** Sie ist die, die im Betrieb
	// dahintersteht, und ihre Anforderung an den Verweis ist der ganze Punkt.
	quelle := service.SymbolquelleAusAblage(t.TempDir(),
		msepaket.NeueBildquelle(), setpaket)

	hat := quelle.Symbolgestalt("symbol1.mse-symbol")
	if len(hat.Teile) != 1 || hat.Teile[0].Name != "Aus dem Paket" {
		t.Fatalf("die Gestalt ist %+v", hat.Teile)
	}
}

// **Ein Verweis, der seinen Ort schon trägt, bleibt unverändert.**
//
// Drei Fälle, und jeder kommt vor: ein Setpaketverweis, wie ihn die Ausfuhr
// baut; ein Pfad mit führendem `/`, der ein **anderes** Paket meint; und ein
// Set, dessen Ort niemand kennt.
func TestEinQualifizierterVerweisBleibt(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct{ setpaket, verweis string }{
		"ein Setpaketverweis": {
			setpaket: "/irgendwo/Set.mse-set",
			verweis:  msepaket.SetpaketVerweis("/woanders/Anderes.mse-set", "s.mse-symbol"),
		},
		"ein fremdes Paket": {
			setpaket: "/irgendwo/Set.mse-set",
			verweis:  "/fremd.mse-style/s.mse-symbol",
		},
		"kein Setort": {setpaket: "", verweis: "s.mse-symbol"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			probe := &symbolquellenprobe{}

			service.SymbolquelleAusAblage(t.TempDir(), probe, fall.setpaket).
				Symbolgestalt(fall.verweis)

			if probe.verweis != fall.verweis {
				t.Errorf("dahinter kam %q an, hineingegeben wurde %q",
					probe.verweis, fall.verweis)
			}
		})
	}
}
