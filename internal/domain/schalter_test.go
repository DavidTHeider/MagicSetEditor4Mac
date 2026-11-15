package domain_test

// **Bereiche und Funktionen abschalten** — OPS-10
// (#1908,
// Vermerk 0158).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Vier Bereiche und sechs Funktionen, vorgabemäßig alle an** (AK-01).
func TestDasSchalterverzeichnisIstVollstaendigUndAn(t *testing.T) {
	t.Parallel()

	erwartet := map[string]domain.Schalterart{
		"sammeln": domain.SchalterBereich, "spielen": domain.SchalterBereich,
		"handeln": domain.SchalterBereich, "ressourcen": domain.SchalterBereich,
		"ki": domain.SchalterFunktion, "netz": domain.SchalterFunktion,
		"markt": domain.SchalterFunktion, "turniere": domain.SchalterFunktion,
		"dienst": domain.SchalterFunktion, "werkzeugserver": domain.SchalterFunktion,
	}

	verzeichnis := domain.Schalterverzeichnis()
	if len(verzeichnis) != len(erwartet) {
		t.Fatalf("%d Schalter, erwartet %d", len(verzeichnis), len(erwartet))
	}

	for _, schalter := range verzeichnis {
		if erwartet[schalter.Kennung] != schalter.Art || !schalter.Vorgabe || schalter.Reife != domain.ReifeStabil {
			t.Errorf("%+v", schalter)
		}
	}
}

// **Jeder Schalter ist eine Einstellung am Rechner** — an oder aus,
// vorgabemäßig an.
func TestJederSchalterIstEineEinstellung(t *testing.T) {
	t.Parallel()

	for _, schalter := range domain.Schalterverzeichnis() {
		schluessel := domain.SchalterEinstellung(schalter.Kennung)

		if fehler := domain.PruefeEinstellung(schluessel, domain.SchalterAus); fehler != nil {
			t.Errorf("%s: aus abgewiesen: %v", schluessel, fehler)
		}

		if fehler := domain.PruefeEinstellung(schluessel, "vielleicht"); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%s: vielleicht angenommen: %v", schluessel, fehler)
		}

		if ebene, kennt := domain.EbeneVon(schluessel); !kennt || ebene != domain.EbeneGeraet {
			t.Errorf("%s: Ebene %q", schluessel, ebene)
		}
	}
}

// **Ein Stand ohne Eintrag gilt mit seiner Vorgabe**, ein unbekannter Wert
// auch — ausgeschaltet ist nur, was „aus“ sagt.
func TestDerSchalterstandFolgtDerVorgabe(t *testing.T) {
	t.Parallel()

	stand := domain.SchalterstandAus(map[string]string{
		domain.SchalterEinstellung("spielen"): domain.SchalterAus,
		domain.SchalterEinstellung("ki"):      "krumm",
	})

	if !stand.Aus("spielen") || stand.Aus("ki") || stand.Aus("handeln") || stand.Aus("gibtsnicht") {
		t.Errorf("Stand: %+v", stand)
	}
}

// **Ein Weg ist abgeschaltet, wenn jede Fähigkeit, die ihn nennt, an einem
// abgeschalteten Schalter hängt** — ein Weg, den auch der Kern braucht,
// bleibt offen (AK-02).
func TestEinWegIstNurAbgeschaltetWennJedeFaehigkeitEsIst(t *testing.T) {
	t.Parallel()

	tafel := domain.Wegetafel{
		domain.FlaecheFenster: {
			"PartieBefehl":  {{"spielen"}},
			"Marktangebote": {{"handeln", "markt"}},
			"Geteilt":       {{"spielen"}, {"turniere"}},
		},
	}

	spielenAus := domain.SchalterstandAus(map[string]string{domain.SchalterEinstellung("spielen"): domain.SchalterAus})

	fehler := tafel.Pruefe(domain.FlaecheFenster, "PartieBefehl", spielenAus)

	var abgeschaltet *domain.Abschaltung
	if !errors.Is(fehler, domain.ErrAbgeschaltet) || !errors.As(fehler, &abgeschaltet) || abgeschaltet.Schalter != "spielen" {
		t.Fatalf("PartieBefehl: %v", fehler)
	}

	for _, offen := range []string{"Marktangebote", "Geteilt", "Unbekannt"} {
		if fehler := tafel.Pruefe(domain.FlaecheFenster, offen, spielenAus); fehler != nil {
			t.Errorf("%s: %v", offen, fehler)
		}
	}

	marktAus := domain.SchalterstandAus(map[string]string{domain.SchalterEinstellung("markt"): domain.SchalterAus})
	if fehler := tafel.Pruefe(domain.FlaecheFenster, "Marktangebote", marktAus); !errors.Is(fehler, domain.ErrAbgeschaltet) {
		t.Errorf("Marktangebote ohne Markt: %v", fehler)
	}

	if fehler := tafel.Pruefe(domain.FlaecheMCP, "PartieBefehl", spielenAus); fehler != nil {
		t.Errorf("eine andere Fläche: %v", fehler)
	}
}

// **Die eingebaute Tafel kennt jede Fläche** — und einen Weg, der an einem
// Bereich hängt.
func TestDieEingebauteWegetafel(t *testing.T) {
	t.Parallel()

	tafel := domain.EingebauteWegetafel()
	stand := domain.SchalterstandAus(map[string]string{domain.SchalterEinstellung("spielen"): domain.SchalterAus})

	for _, flaeche := range []domain.Flaeche{domain.FlaecheFenster, domain.FlaecheMCP, domain.FlaecheCLI, domain.FlaecheREST} {
		if len(tafel[flaeche]) == 0 {
			t.Errorf("%s: keine Wege", flaeche)
		}
	}

	if fehler := tafel.Pruefe(domain.FlaecheFenster, "PartieBefehl", stand); !errors.Is(fehler, domain.ErrAbgeschaltet) {
		t.Errorf("PartieBefehl bei abgeschaltetem Spielen: %v", fehler)
	}

	if fehler := tafel.Pruefe(domain.FlaecheFenster, "Sets", stand); fehler != nil {
		t.Errorf("Sets gehört zum Kern: %v", fehler)
	}
}

// **Die Abweisung nennt den Schalter und den Weg zurück.**
func TestDieAbschaltungNenntIhrenSchalter(t *testing.T) {
	t.Parallel()

	satz := (&domain.Abschaltung{Schalter: "markt"}).Error()
	if !strings.Contains(satz, `"markt"`) || !strings.Contains(satz, "Bereiche und Funktionen") {
		t.Errorf("Satz: %q", satz)
	}
}

// **Sind alle Fähigkeiten eines Wegs aus, nennt er die erste** — damit
// dieselbe Lage überall denselben Namen trägt.
func TestEinGanzAbgeschalteterWegNenntDieErsteFaehigkeit(t *testing.T) {
	t.Parallel()

	tafel := domain.Wegetafel{domain.FlaecheCLI: {"mse deck": {{"spielen"}, {"turniere"}}}}
	stand := domain.SchalterstandAus(map[string]string{
		domain.SchalterEinstellung("turniere"): domain.SchalterAus,
		domain.SchalterEinstellung("spielen"):  domain.SchalterAus,
	})

	var abgeschaltet *domain.Abschaltung
	if !errors.As(tafel.Pruefe(domain.FlaecheCLI, "mse deck", stand), &abgeschaltet) || abgeschaltet.Schalter != "spielen" {
		t.Errorf("%+v", abgeschaltet)
	}
}

// **Ohne eigene Wahl weist kein Weg ab** — OPS-10.AK-01, an der eingebauten
// Tafel und über jede ihrer Flächen.
func TestOhneWahlWeistKeinWegAb(t *testing.T) {
	t.Parallel()

	tafel := domain.EingebauteWegetafel()
	stand := domain.SchalterstandAus(nil)

	if !stand.AllesAn() {
		t.Fatal("ohne Wahl ist ein Schalter aus")
	}

	for flaeche, wege := range tafel {
		for weg := range wege {
			if fehler := tafel.Pruefe(flaeche, weg, stand); fehler != nil {
				t.Errorf("%s %s: %v", flaeche, weg, fehler)
			}
		}
	}
}

// **Ein Endpunkt trifft seine Zeile, gleich wie seine Platzhalter heißen** —
// der Katalog schreibt für Menschen, `net/http` für sich.
func TestEinEndpunktWirdOhnePlatzhalternamenGefuehrt(t *testing.T) {
	t.Parallel()

	for roh, erwartet := range map[string]string{
		"GET /sets/{setID}/karten":      "GET /sets/{}/karten",
		"GET /sets/{id}/karten":         "GET /sets/{}/karten",
		"PUT /pakete/{paket}/{pfad...}": "PUT /pakete/{}/{}",
		"GET /grenzen":                  "GET /grenzen",
	} {
		if weg := domain.RESTWeg(roh); weg != erwartet {
			t.Errorf("%q → %q, erwartet %q", roh, weg, erwartet)
		}
	}
}

// schalterschluessel gibt die Einstellung je Schalter, in Verzeichnisordnung.
func schalterschluessel() []string {
	verzeichnis := domain.Schalterverzeichnis()
	schluessel := make([]string, 0, len(verzeichnis))

	for _, schalter := range verzeichnis {
		schluessel = append(schluessel, domain.SchalterEinstellung(schalter.Kennung))
	}

	return schluessel
}
