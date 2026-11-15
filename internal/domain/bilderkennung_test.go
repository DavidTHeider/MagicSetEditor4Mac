package domain_test

// **Was auf einem Kartenbild zu sehen ist** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// # Was hier gemessen wird
//
// **Dass die Evidenz getrennt bleibt.** Das Issue verlangt es wörtlich:
// „OCR/Layout/Face/Set/Kunstwerk als **getrennte** Evidenz behandeln."
//
// **Dass ein Widerspruch auffällt** — ein Wert an einer Stelle, die angeblich
// unlesbar war, ist geraten und nicht gelesen.
//
// **Und dass nirgends eine Sicherheit in Prozent steht.** Es gibt kein Feld
// dafür; dieser Lauf hält, dass auch keines nachwächst (0108).

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine gelesene Karte ergibt einen Suchbegriff aus Name, Set und Nummer.**
func TestEineGeleseneKarteErgibtEinenSuchbegriff(t *testing.T) {
	t.Parallel()

	erkennung := domain.Bilderkennung{Beobachtungen: []domain.Beobachtung{
		{Art: domain.EvidenzName, Wert: "Lightning Bolt", Lesbar: true},
		{Art: domain.EvidenzSetcode, Wert: "lea", Lesbar: true},
		{Art: domain.EvidenzNummer, Wert: "161", Lesbar: true},
		{Art: domain.EvidenzKunstwerk, Wert: "eine Hand mit Blitz", Lesbar: true},
	}}

	if befunde := domain.PruefeBilderkennung(erkennung); len(befunde) != 0 {
		t.Fatalf("die Befunde sind %+v", befunde)
	}

	begriff := erkennung.Suchbegriff()

	for _, teil := range []string{"Lightning Bolt", "lea", "161"} {
		if !strings.Contains(begriff, teil) {
			t.Errorf("der Suchbegriff %q nennt %q nicht", begriff, teil)
		}
	}

	// **Das Kunstwerk geht nicht mit** — es ist kein Feld der Suche, und eine
	// Bildbeschreibung dorthin machte aus jeder Suche eine Volltextsuche.
	if strings.Contains(begriff, "Hand") {
		t.Errorf("der Suchbegriff trägt die Bildbeschreibung: %q", begriff)
	}
}

// **Eine unlesbare Stelle verengt die Suche nicht.**
//
// **Und sie verschwindet auch nicht**: sie steht in der Antwort, damit
// jemand sieht, **warum** der Kandidat offenblieb.
func TestEineUnlesbareStelleVerengtDieSucheNicht(t *testing.T) {
	t.Parallel()

	erkennung := domain.Bilderkennung{Beobachtungen: []domain.Beobachtung{
		{Art: domain.EvidenzName, Wert: "Counterspell", Lesbar: true},
		{Art: domain.EvidenzNummer, Lesbar: false},
	}}

	if begriff := erkennung.Suchbegriff(); begriff != "Counterspell" {
		t.Errorf("der Suchbegriff ist %q", begriff)
	}
}

// **Ein Wert an einer unlesbaren Stelle ist ein Befund.**
func TestEinWertAnEinerUnlesbarenStelleIstEinBefund(t *testing.T) {
	t.Parallel()

	befunde := domain.PruefeBilderkennung(domain.Bilderkennung{
		Beobachtungen: []domain.Beobachtung{
			{Art: domain.EvidenzName, Wert: "Counterspell", Lesbar: true},
			{Art: domain.EvidenzNummer, Wert: "55", Lesbar: false},
		},
	})

	if len(befunde) != 1 || befunde[0].Art != domain.BefundWertOhneLesbarkeit {
		t.Fatalf("die Befunde sind %+v", befunde)
	}

	if befunde[0].Stelle != string(domain.EvidenzNummer) {
		t.Errorf("der Befund nennt %q", befunde[0].Stelle)
	}
}

// **Eine erfundene Evidenzart ebenso.**
func TestEineErfundeneEvidenzartIstEinBefund(t *testing.T) {
	t.Parallel()

	befunde := domain.PruefeBilderkennung(domain.Bilderkennung{
		Beobachtungen: []domain.Beobachtung{
			{Art: "hellsehen", Wert: "etwas", Lesbar: true},
		},
	})

	if len(befunde) != 1 || befunde[0].Art != domain.BefundErfundeneEvidenz {
		t.Fatalf("die Befunde sind %+v", befunde)
	}
}

// **Ein Bild ohne eine einzige Beobachtung ist ein Befund.**
//
// **Und nicht ein leeres Ergebnis**: „ich sehe nichts" ist eine Antwort, aber
// keine, aus der sich suchen lässt.
func TestEinBildOhneBeobachtungIstEinBefund(t *testing.T) {
	t.Parallel()

	befunde := domain.PruefeBilderkennung(domain.Bilderkennung{})

	if len(befunde) != 1 || befunde[0].Art != domain.BefundOhneBeobachtung {
		t.Fatalf("die Befunde sind %+v", befunde)
	}
}

// **Und eine Bilderkennung hat keine Sicherheit in Prozent.**
//
// # Warum das hier steht und nicht als Prüfung im Code
//
// **Weil es kein Feld dafür gibt.** Die Zusage ist eine über die **Gestalt**,
// und eine Gestalt lässt sich aufzählen — [domain.Bilderkennungsfelder].
// Dieselbe Bauart wie bei der Balanceanalyse (0108): wer eine Zahl wollte,
// müsste zuerst das Feld schaffen und stieße dabei auf diesen Lauf.
func TestEineBilderkennungHatKeineSicherheit(t *testing.T) {
	t.Parallel()

	for _, feld := range domain.Bilderkennungsfelder() {
		for _, verboten := range []string{
			"sicher", "wahrscheinlich", "prozent", "score", "konfidenz",
		} {
			if strings.Contains(strings.ToLower(feld), verboten) {
				t.Errorf("das Feld %q trägt %q — eine Bilderkennung sagt, "+
					"was sie gelesen hat, und nicht, wie sicher sie ist",
					feld, verboten)
			}
		}
	}

	// **Und die Gestalt ist vollständig aufgezählt.** Ein Feld, das hier
	// fehlt, entzöge sich dem Lauf darüber.
	if len(domain.Bilderkennungsfelder()) != 2 {
		t.Errorf("die Felder sind %v", domain.Bilderkennungsfelder())
	}
}
