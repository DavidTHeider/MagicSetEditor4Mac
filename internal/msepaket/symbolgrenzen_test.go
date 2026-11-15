package msepaket_test

// **Schadhafte und überkomplexe Symbole sind begrenzt** — DOM-02.AK-04
// (#1697):
// „Schadhafte oder überkomplexe Vektor-/Schriftdateien sind begrenzt und
// erhalten verständliche Fehler; FMT-06 gilt auch hier."

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// form gibt eine Form mit `punkte` Punkten.
func form(punkte int) msepaket.Symbolteil {
	teil := msepaket.Symbolteil{Art: msepaket.TeilForm, Name: "Form"}
	for range punkte {
		teil.Punkte = append(teil.Punkte, msepaket.Kontrollpunkt{})
	}

	return teil
}

// verschachtelt gibt eine Gruppe, `tiefe` Ebenen tief.
func verschachtelt(tiefe int) msepaket.Symbolteil {
	teil := form(3)
	for range tiefe {
		teil = msepaket.Symbolteil{Art: msepaket.TeilGruppe, Teile: []msepaket.Symbolteil{teil}}
	}

	return teil
}

func TestEinGewoehnlichesSymbolLiegtInDenGrenzen(t *testing.T) {
	if err := msepaket.PruefeSymbolumfang(msepaket.Standardsymbol()); err != nil {
		t.Errorf("das Standardsymbol: %v", err)
	}

	// **Knapp an der Tiefengrenze und mit so vielen Punkten, wie eine Datei
	// trägt, geht es noch.** Knapp an der Punktegrenze stand hier bis QA-06b
	// (#1967) — und das war ein Symbol, das sich geschrieben nicht wieder
	// öffnen ließ ([TestEinSymbolUeberDerDateigrenzeWirdNichtAbgelegt]).
	gross := msepaket.Symbolzeichnung{Teile: []msepaket.Symbolteil{
		form(12_000), verschachtelt(msepaket.GrenzeSymboltiefe - 1),
	}}
	if err := msepaket.PruefeSymbolumfang(gross); err != nil {
		t.Errorf("knapp an den Grenzen: %v", err)
	}
}

func TestZuVieleTeilePunkteOderEbenenWerdenAbgewiesen(t *testing.T) {
	vieleTeile := msepaket.Symbolzeichnung{}
	for range msepaket.GrenzeSymbolteile + 1 {
		vieleTeile.Teile = append(vieleTeile.Teile, form(3))
	}

	for name, fall := range map[string]struct {
		zeichnung  msepaket.Symbolzeichnung
		schluessel domain.Schluessel
	}{
		"Teile":  {vieleTeile, domain.SchluesselSymbolZuVieleTeile},
		"Punkte": {msepaket.Symbolzeichnung{Teile: []msepaket.Symbolteil{form(msepaket.GrenzeSymbolpunkte + 1)}}, domain.SchluesselSymbolZuVielePunkte},
		"Ebenen": {msepaket.Symbolzeichnung{Teile: []msepaket.Symbolteil{verschachtelt(msepaket.GrenzeSymboltiefe + 1)}}, domain.SchluesselSymbolZuTief},
	} {
		err := msepaket.PruefeSymbolumfang(fall.zeichnung)

		// **Ein eigener Schlüssel und die Zahlen als Werte** — der Satz im
		// Katalog sagt, was zu tun ist, in der Sprache des Fensters.
		var feldfehler *domain.FieldError
		if !errors.Is(err, domain.ErrValidation) || !errors.As(err, &feldfehler) ||
			feldfehler.Key != fall.schluessel || !strings.Contains(feldfehler.Message, "höchstens") ||
			feldfehler.Werte["grenze"] == "" || feldfehler.Werte["anzahl"] == "" {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestEineZuGrosseOderZuVerwickelteDateiWirdNichtGelesen(t *testing.T) {
	ort := t.TempDir()

	gross := filepath.Join(ort, "gross.mse-symbol")
	if err := os.WriteFile(gross, make([]byte, msepaket.GrenzeSymboldatei+1), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, _, err := msepaket.LiesSymbolzeichnung(gross); !errors.Is(err, msepaket.ErrZuGross) {
		t.Errorf("zu groß: %v", err)
	}

	// **Aus Bytes derselbe Weg** — die Bibliothek hält den Text im Bestand.
	var text strings.Builder
	text.WriteString("mse version: 0.3.5\n")
	for range msepaket.GrenzeSymbolteile + 1 {
		text.WriteString("part:\n\ttype: shape\n")
	}

	if _, _, err := msepaket.SymbolzeichnungAus("viel.mse-symbol", []byte(text.String())); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("zu viele Teile: %v", err)
	}

	if _, _, err := msepaket.SymbolzeichnungAus("gross.mse-symbol", make([]byte, msepaket.GrenzeSymboldatei+1)); !errors.Is(err, msepaket.ErrZuGross) {
		t.Errorf("zu groß aus Bytes: %v", err)
	}

	zeichnung, _, err := msepaket.SymbolzeichnungAus("klein.mse-symbol",
		msepaket.SchreibeSymbolzeichnung(msepaket.Standardsymbol()))
	if err != nil || len(zeichnung.Teile) == 0 {
		t.Errorf("ein gewöhnliches: %+v, %v", zeichnung, err)
	}
}

// **Was durch die Grenzen geht, lässt sich geschrieben wieder lesen** —
// QA-06b (#1967),
// beim Festlegen der Invariante von `FuzzSymbolzeichnung` gefunden.
//
// Ein Punkt kostet geschrieben mindestens achtzig Byte (`position:`, `lock:`,
// `line_after:`, zehn Nachkommastellen je Zahl), mit Kurve gut 170. Unter der
// Punktegrenze von 65.536 lag damit eine Zeichnung, die geschrieben fünf
// Megabyte wog: sie ging durch [msepaket.PruefeSymbolumfang], wurde
// abgelegt — und ließ sich danach nicht mehr öffnen, weil
// [msepaket.GrenzeSymboldatei] beim Lesen gilt.
//
// **Mit eigenem Schlüssel** (#1973): `symbolZuGross` sagt „diese Datei" und
// rät zu einer anderen — für ein Symbol, das gerade gezeichnet, abgelegt
// oder nachgezeichnet wurde, ist das der falsche Rat.
func TestEinSymbolUeberDerDateigrenzeWirdNichtAbgelegt(t *testing.T) {
	zuViel := msepaket.Symbolzeichnung{Teile: []msepaket.Symbolteil{form(20_000)}}

	if geschrieben := msepaket.SchreibeSymbolzeichnung(zuViel); int64(len(geschrieben)) <= msepaket.GrenzeSymboldatei {
		t.Fatalf("die Vorrichtung trägt: %d Byte liegen unter der Grenze", len(geschrieben))
	}

	var feldfehler *domain.FieldError
	if err := msepaket.PruefeSymbolumfang(zuViel); !errors.As(err, &feldfehler) ||
		feldfehler.Key != domain.SchluesselSymbolGespeichertZuGross || feldfehler.Werte["grenze"] == "" {
		t.Errorf("20.000 Punkte, geschrieben über der Dateigrenze: %v", err)
	}
}
