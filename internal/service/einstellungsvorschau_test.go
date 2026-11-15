// **Die Vorschau der Einfuhr** — Werkstück O3 von Phase 5.2
// (#1427):
// vorher, nachher, Herkunft und die Schranke, die eine Datei nicht lockert.
// Die Werkbank steht in `einstellungsaustausch_test.go`; diese Datei hält die
// Läufe über die Vorschau, damit jene unter der Zeilengrenze bleibt (§2).
package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Vorschau zeigt vorher, nachher und wessen Entscheidung der alte Wert
// war — und ändert nichts.**
func TestDieVorschauZeigtDenUnterschied(t *testing.T) {
	t.Parallel()

	bank := werkbankBauen()
	bank.ablage.werte[domain.EinstellungSprache] = "de"

	vorschau, fehler := bank.austausch.Vorschau(t.Context(), alsDatei(t,
		map[string]map[string]string{
			string(domain.EbeneProfil): {domain.EinstellungSprache: "en"},
		}))
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if len(vorschau.Zeilen) != 1 {
		t.Fatalf("es kamen %+v", vorschau.Zeilen)
	}

	zeile := vorschau.Zeilen[0]
	if zeile.Vorher != "de" || zeile.Nachher != "en" ||
		zeile.Herkunft != domain.GeltungProfil {
		t.Errorf("%+v", zeile)
	}

	if vorschau.Revision == "" {
		t.Error("die Vorschau nennt keine Revision")
	}

	if bank.schreiber.rufe != 0 {
		t.Error("die Vorschau hat geschrieben")
	}
}

// **Was schon so dasteht, ist keine Zeile — wird aber gezählt.**
func TestWasSchonSoDastehtWirdGezaehlt(t *testing.T) {
	t.Parallel()

	bank := werkbankBauen()
	bank.ablage.werte[domain.EinstellungSprache] = "en"

	vorschau, fehler := bank.austausch.Vorschau(t.Context(), alsDatei(t,
		map[string]map[string]string{
			string(domain.EbeneProfil): {domain.EinstellungSprache: "en"},
		}))
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if len(vorschau.Zeilen) != 0 || vorschau.Unveraendert != 1 {
		t.Errorf("%d Zeilen, %d unverändert",
			len(vorschau.Zeilen), vorschau.Unveraendert)
	}
}

// **Die Herkunft eines geerbten Werts ist der Mensch.**
func TestDieHerkunftEinesGeerbtenWerts(t *testing.T) {
	t.Parallel()

	bank := werkbankBauen()
	bank.nutzerwerte.werte[domain.EinstellungErscheinungsbild] = "hell"

	vorschau, fehler := bank.austausch.Vorschau(t.Context(), alsDatei(t,
		map[string]map[string]string{
			string(domain.EbeneProfil): {
				domain.EinstellungErscheinungsbild: "dunkel",
			},
		}))
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if len(vorschau.Zeilen) != 1 {
		t.Fatalf("%+v", vorschau.Zeilen)
	}

	if vorschau.Zeilen[0].Herkunft != domain.GeltungNutzer {
		t.Errorf("die Herkunft ist %q", vorschau.Zeilen[0].Herkunft)
	}
}

// **Eine Schranke lässt sich nicht durch eine Datei aufweichen.**
func TestEineDateiLockertKeineSchranke(t *testing.T) {
	t.Parallel()

	schalter := domain.DienstEinstellung(domain.DienstScryfall)

	bank := werkbankBauen()
	bank.nutzerwerte.werte[domain.RichtlinieNurLokal] = domain.NurLokalAn

	vorschau, fehler := bank.austausch.Vorschau(t.Context(), alsDatei(t,
		map[string]map[string]string{
			string(domain.EbeneProfil): {schalter: domain.DienstAn},
		}))
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if len(vorschau.Zeilen) != 0 {
		t.Errorf("der Dienst wurde angeschaltet: %+v", vorschau.Zeilen)
	}

	if len(vorschau.Befunde) != 1 {
		t.Fatalf("%+v", vorschau.Befunde)
	}
}

// **Und die Gegenprobe: dieselbe Zeile ohne Schranke kommt durch.**
//
// Ohne sie sagte der Lauf darüber nur, dass nichts geschieht — nicht, dass es
// die Schranke war.
func TestOhneSchrankeKommtDerSchalterDurch(t *testing.T) {
	t.Parallel()

	schalter := domain.DienstEinstellung(domain.DienstScryfall)

	bank := werkbankBauen()
	bank.ablage.werte[schalter] = domain.DienstAus

	vorschau, fehler := bank.austausch.Vorschau(t.Context(), alsDatei(t,
		map[string]map[string]string{
			string(domain.EbeneProfil): {schalter: domain.DienstAn},
		}))
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if len(vorschau.Zeilen) != 1 || len(vorschau.Befunde) != 0 {
		t.Errorf("%+v / %+v", vorschau.Zeilen, vorschau.Befunde)
	}
}

// **Jede Zeile sagt, wann sie wirkt** — OPS-01.AK-02
// (#1718):
// der Import zeigt den Neustartbedarf. Die Auskunft ist dieselbe wie an der
// Einstellung selbst (`domain.EinstellungsWirkung`), damit Vorschau und
// Einstellungsseite nicht zweierlei sagen.
func TestJedeZeileDerVorschauSagtWannSieWirkt(t *testing.T) {
	t.Parallel()

	bank := werkbankBauen()

	vorschau, fehler := bank.austausch.Vorschau(t.Context(), alsDatei(t,
		map[string]map[string]string{
			string(domain.EbeneProfil): {domain.EinstellungSprache: "en"},
		}))
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if len(vorschau.Zeilen) != 1 || vorschau.Zeilen[0].Wirkung != domain.WirkungSofort {
		t.Errorf("es kamen %+v", vorschau.Zeilen)
	}
}
