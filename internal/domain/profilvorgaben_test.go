package domain_test

// **Die Vorgaben eines Profils** — Werkstück 9 von Phase 4.12
// (#1351),
// Punkte 2, 3 und 8.

import (
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Zustand einer neuen Sammlungszeile kommt aus der Domäne.**
//
// **Und nicht aus einer zweiten Liste im Fenster**: die Zustände stehen in
// `exemplar.go`, und wer sie hier abschriebe, hätte beim nächsten eine
// Auswahl, die der Bestand abweist.
func TestDieZustandsvorgabeKenntJedenZustand(t *testing.T) {
	t.Parallel()

	werte := domain.ErlaubteEinstellungswerte(domain.VorgabeZustand)

	for _, zustand := range domain.Zustaende() {
		if !slices.Contains(werte, string(zustand)) {
			t.Errorf("%q fehlt in %v", zustand, werte)
		}
	}

	// **Und »ohne« steht vorn** — es ist die Vorgabe: ein frischer Bestand
	// verhält sich, wie er sich immer verhalten hat, und das Feld bleibt leer.
	if len(werte) == 0 || werte[0] != domain.OhneVorgabe {
		t.Fatalf("die Werte sind %v", werte)
	}

	if domain.EinstellungsVorgabe(domain.VorgabeZustand) != domain.OhneVorgabe {
		t.Errorf("die Vorgabe ist %q",
			domain.EinstellungsVorgabe(domain.VorgabeZustand))
	}
}

// **Ein neues Turnier ist kein Einzelspiel, solange niemand es sagt.**
func TestDieEinzelspielvorgabeIstAus(t *testing.T) {
	t.Parallel()

	if vorgabe := domain.EinstellungsVorgabe(
		domain.VorgabeEinzelspiel); vorgabe != domain.ErsteSchritteAus {
		t.Errorf("die Vorgabe ist %q", vorgabe)
	}
}

// **Ein Basispfad ist freier Text — er hat keine Werteliste.**
//
// # Warum das eine eigene Art ist
//
// **Weil ein Verzeichnis sich nicht aufzählen lässt.** Jede andere
// Einstellung nennt ihre erlaubten Werte, und das Fenster baut daraus die
// Auswahlliste; hier gibt es nichts aufzuzählen, und ein leeres `Werte`
// **sagt** das — statt eine Auswahl zu versprechen, die leer bleibt.
func TestEinBasispfadIstFreierText(t *testing.T) {
	t.Parallel()

	if werte := domain.ErlaubteEinstellungswerte(
		domain.VorgabeHochladepfad); len(werte) != 0 {
		t.Errorf("der Pfad nennt Werte: %v", werte)
	}

	// **Und jeder vollständige Pfad kommt durch** — auch einer, den es nicht
	// gibt: ob ein Verzeichnis existiert, weiß der Rechner und nicht die
	// Domäne. **Ein relativer nicht mehr** (OPS-02.AK-01, #1719): relativ
	// wozu, stand hier nicht, und die Antwort war auf jeder Fläche eine
	// andere — `ordnerangabe_test.go` hält die Regel.
	for _, pfad := range []string{"", "/Users/wer/Bilder", "~/Bilder"} {
		if fehler := domain.PruefeEinstellung(
			domain.VorgabeHochladepfad, pfad); fehler != nil {
			t.Errorf("%q wurde abgewiesen: %v", pfad, fehler)
		}
	}
}

// **Ein Pfad mit einem Nullbyte kommt nicht durch.**
//
// **Der eine Fall, den die Domäne beurteilen kann**: ein `\x00` beendet in
// jedem Systemaufruf die Zeichenkette, und was danach steht, verschwindet
// stillschweigend.
func TestEinPfadMitNullbyteWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	fehler := domain.PruefeEinstellung(domain.VorgabeHochladepfad, "/wo/\x00/hin")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("gemeldet wurde: %v", fehler)
	}
}

// **Eine Einstellung mit Werteliste weist einen fremden Wert weiter ab.**
//
// **Die Gegenprobe zum freien Text**: ohne sie hielte der Lauf darüber auch,
// wenn `PruefeEinstellung` gar nichts mehr prüfte.
func TestEineAufzaehlungWeistWeiterhinAb(t *testing.T) {
	t.Parallel()

	if fehler := domain.PruefeEinstellung(
		domain.VorgabeZustand, "zerknittert"); fehler == nil {
		t.Error("der erfundene Zustand kam durch")
	}
}
