package domain_test

// **Der Ablauf eines Zuges** — Werkstück G9 von Phase 5.5
// (#1592).
//
// # Wogegen hier geprüft wird
//
// **Gegen die Comprehensive Rules, Fassung 2026-08-19**, und jede Nummer
// steht bei ihrer Zusicherung. Eine erfundene sähe aus wie ein Beleg (§11).
//
//   - CR 500.1 — fünf Phasen in dieser Reihenfolge, **jede in jedem Zug**.
//   - CR 501.1 — die Anfangsphase hat drei Abschnitte.
//   - CR 506.1 — die Kampfphase hat fünf.
//   - CR 512.1 — die Endphase hat zwei.
//   - CR 502.4 — im Enttappen bekommt **niemand** Priorität.
//   - CR 514.3 — im Aufräumen normalerweise auch nicht.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Zug besteht aus zwölf Abschnitten** — fünf Phasen, zehn Schritte,
// und die beiden Hauptphasen haben keine.
//
// **Gezählt und nicht geschätzt**: CR 500.1 zerlegt Anfangs-, Kampf- und
// Endphase in Schritte und die beiden Hauptphasen nicht — 3 + 1 + 5 + 1 + 2.
func TestEinZugHatZwoelfAbschnitte(t *testing.T) {
	t.Parallel()

	folge := domain.Zugfolge()

	if len(folge) != 12 {
		t.Fatalf("die Zugfolge hat %d Abschnitte", len(folge))
	}

	erwartet := []domain.Abschnitt{
		{Phase: domain.PhaseAnfang, Schritt: domain.SchrittEnttappen},
		{Phase: domain.PhaseAnfang, Schritt: domain.SchrittVersorgung},
		{Phase: domain.PhaseAnfang, Schritt: domain.SchrittZiehen},
		{Phase: domain.PhaseHauptVorKampf},
		{Phase: domain.PhaseKampf, Schritt: domain.SchrittKampfbeginn},
		{Phase: domain.PhaseKampf, Schritt: domain.SchrittAngreifer},
		{Phase: domain.PhaseKampf, Schritt: domain.SchrittBlocker},
		{Phase: domain.PhaseKampf, Schritt: domain.SchrittKampfschaden},
		{Phase: domain.PhaseKampf, Schritt: domain.SchrittKampfende},
		{Phase: domain.PhaseHauptNachKampf},
		{Phase: domain.PhaseEnde, Schritt: domain.SchrittEnde},
		{Phase: domain.PhaseEnde, Schritt: domain.SchrittAufraeumen},
	}

	for stelle, abschnitt := range erwartet {
		if folge[stelle] != abschnitt {
			t.Errorf("an Stelle %d steht %+v, erwartet %+v",
				stelle, folge[stelle], abschnitt)
		}
	}
}

// **Eine Hauptphase hat keinen Schritt** — CR 505.1 nennt sie „phases", und
// CR 500.1 zählt nur Anfang, Kampf und Ende zu denen mit Schritten.
//
// **Das ist kein Detail**: wer ihr einen Schritt gäbe, bekäme im Zug
// dreizehn Abschnitte und zählte die Priorität falsch.
func TestEineHauptphaseHatKeinenSchritt(t *testing.T) {
	t.Parallel()

	for _, abschnitt := range domain.Zugfolge() {
		haupt := abschnitt.Phase == domain.PhaseHauptVorKampf ||
			abschnitt.Phase == domain.PhaseHauptNachKampf

		if haupt && abschnitt.Schritt != "" {
			t.Errorf("%v trägt den Schritt %q", abschnitt.Phase, abschnitt.Schritt)
		}

		if !haupt && abschnitt.Schritt == "" {
			t.Errorf("%v hat keinen Schritt", abschnitt.Phase)
		}
	}
}

// **Zehn Abschnitte vergeben Priorität, zwei nicht.**
//
// CR 502.4 (Enttappen) und CR 514.3 (Aufräumen) nehmen sie ausdrücklich aus —
// und CR 117.3a sagt dasselbe von der anderen Seite: im Enttappen bekommt sie
// niemand.
func TestZweiAbschnitteVergebenKeinePrioritaet(t *testing.T) {
	t.Parallel()

	ohne := []domain.Abschnitt{}

	for _, abschnitt := range domain.Zugfolge() {
		if !abschnitt.VergibtPrioritaet() {
			ohne = append(ohne, abschnitt)
		}
	}

	if len(ohne) != 2 {
		t.Fatalf("%d Abschnitte ohne Priorität: %+v", len(ohne), ohne)
	}

	if ohne[0].Schritt != domain.SchrittEnttappen ||
		ohne[1].Schritt != domain.SchrittAufraeumen {
		t.Errorf("ohne Priorität sind %+v", ohne)
	}
}

// **Nach dem Aufräumen fängt ein neuer Zug an.**
func TestNachDemAufraeumenFaengtEinNeuerZugAn(t *testing.T) {
	t.Parallel()

	stelle := domain.Abschnitt{
		Phase: domain.PhaseEnde, Schritt: domain.SchrittAufraeumen,
	}

	naechster, neuerZug := stelle.Naechster()

	if !neuerZug {
		t.Error("nach dem Aufräumen ging es im selben Zug weiter")
	}

	if naechster != domain.Zugfolge()[0] {
		t.Errorf("der neue Zug fängt bei %+v an", naechster)
	}
}

// **Und dazwischen geht es der Reihe nach weiter, ohne einen zu überspringen.**
//
// **CR 500.1**: jede Phase findet in jedem Zug statt, auch wenn nichts
// geschieht. Ein Ablauf, der leere Abschnitte überspränge, wäre schneller und
// falsch.
func TestJederAbschnittFolgtDemVorigen(t *testing.T) {
	t.Parallel()

	folge := domain.Zugfolge()
	gelaufen := []domain.Abschnitt{folge[0]}

	for stelle := 0; stelle < len(folge)-1; stelle++ {
		naechster, neuerZug := gelaufen[stelle].Naechster()

		if neuerZug {
			t.Fatalf("bei %+v fing schon ein neuer Zug an", gelaufen[stelle])
		}

		gelaufen = append(gelaufen, naechster)
	}

	for stelle := range folge {
		if gelaufen[stelle] != folge[stelle] {
			t.Errorf("Schritt %d lief nach %+v statt %+v",
				stelle, gelaufen[stelle], folge[stelle])
		}
	}
}

// **Ein unbekannter Abschnitt fängt von vorn an.**
//
// **Nicht „bleibt stehen"**: ein Zustand aus einem Bestand, den jemand von
// Hand geändert hat, darf eine Partie nicht anhalten (§8) — und der Anfang
// eines Zuges ist der sichere Ausgang.
func TestEinUnbekannterAbschnittFaengtVonVornAn(t *testing.T) {
	t.Parallel()

	stelle := domain.Abschnitt{Phase: "erfunden", Schritt: "auch"}

	naechster, neuerZug := stelle.Naechster()

	if !neuerZug || naechster != domain.Zugfolge()[0] {
		t.Errorf("ein unbekannter Abschnitt führte zu %+v (neuer Zug: %v)",
			naechster, neuerZug)
	}
}
