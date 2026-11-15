package domain_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Satz für Menschen und der Grund fürs Protokoll sind zweierlei** —
// Werkstück 1 von Phase 4.13
// (#1352).
//
// # Wofür es das gibt
//
// **`fmt.Errorf("…: %w", ursache)` setzt die Ursache in den Satz.** Damit
// steht sie in `Error()`, und `Error()` ist das, was `internal/app` als
// `Meldung` weitergibt — an Kommandozeile, MCP und REST unübersetzt. So
// erreichte ein `open /Users/…/bestand.db: no such file or directory` die
// Oberfläche, und §8 verlangt das Gegenteil.
//
// **Sie soll aber nicht verschwinden.** Wer einen Fehler sucht, braucht sie —
// deshalb bleibt sie in der Kette und geht ins Protokoll.
func TestEinGrundStehtInDerKetteUndNichtImSatz(t *testing.T) {
	t.Parallel()

	ursache := errors.New("open /Users/wer/bestand.db: no such file or directory")
	fehler := domain.MitGrund(
		domain.ErrValidation, ursache, "die Bilddatei ließ sich nicht öffnen")

	// **Der Satz nennt das Wahrzeichen und den eigenen Text** — dieselbe
	// Gestalt wie bisher, nur ohne den technischen Schwanz.
	erwartet := "Validierung fehlgeschlagen: die Bilddatei ließ sich nicht öffnen"
	if fehler.Error() != erwartet {
		t.Errorf("Error() = %q, erwartet war %q", fehler.Error(), erwartet)
	}

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Error("das Wahrzeichen ging verloren — errors.Is findet es nicht")
	}

	if !errors.Is(fehler, ursache) {
		t.Error("der Grund ging verloren — er soll in der Kette bleiben")
	}

	if grund := domain.GrundVon(fehler); !errors.Is(grund, ursache) {
		t.Errorf("GrundVon = %v, erwartet war %v", grund, ursache)
	}
}

// **Und er überlebt das Einwickeln auf dem Weg nach oben.**
//
// **Das ist der Fall, für den es ihn gibt**: `service` und `app` schlagen
// ihren eigenen Satz darum, und der Grund muss unten trotzdem auffindbar
// bleiben.
func TestEinGrundUeberlebtDasEinwickeln(t *testing.T) {
	t.Parallel()

	ursache := errors.New("sqlite3: SQL logic error: no such table: karten")
	unten := domain.MitGrund(domain.ErrNotFound, ursache, "die Karte gibt es nicht")
	oben := fmt.Errorf("die Karte lesen: %w", unten)

	if grund := domain.GrundVon(oben); !errors.Is(grund, ursache) {
		t.Errorf("GrundVon = %v, erwartet war %v", grund, ursache)
	}

	erwartet := "die Karte lesen: nicht gefunden: die Karte gibt es nicht"
	if oben.Error() != erwartet {
		t.Errorf("Error() = %q, erwartet war %q", oben.Error(), erwartet)
	}
}

// **Ohne Grund ist es ein gewöhnlicher umschlossener Fehler.**
//
// **Kein eigener Zweig am Aufrufort**: wer eine Stelle umstellt, an der es
// nichts einzuwickeln gibt, soll dieselbe Form schreiben können.
func TestOhneGrundBleibtNurDerSatz(t *testing.T) {
	t.Parallel()

	fehler := domain.MitGrund(domain.ErrConflict, nil, "das Set hat sich geändert")

	if !errors.Is(fehler, domain.ErrConflict) {
		t.Error("das Wahrzeichen ging verloren")
	}

	if grund := domain.GrundVon(fehler); grund != nil {
		t.Errorf("GrundVon = %v, erwartet war nichts", grund)
	}

	erwartet := "Konflikt mit dem gegenwärtigen Stand: das Set hat sich geändert"
	if fehler.Error() != erwartet {
		t.Errorf("Error() = %q, erwartet war %q", fehler.Error(), erwartet)
	}
}

// **Und ein Fehler ohne verdeckten Grund gibt keinen her.**
func TestEinGewoehnlicherFehlerHatKeinenVerdecktenGrund(t *testing.T) {
	t.Parallel()

	if grund := domain.GrundVon(fmt.Errorf("%w: nichts", domain.ErrValidation)); grund != nil {
		t.Errorf("GrundVon = %v, erwartet war nichts", grund)
	}

	if grund := domain.GrundVon(nil); grund != nil {
		t.Errorf("GrundVon(nil) = %v, erwartet war nichts", grund)
	}
}

// **Der Satz kommt heraus, die fremde Meldung nicht.**
//
// **Für die Stellen, an denen ein Fehler kein Fehler bleibt** — die halbe
// Suche in A15 ist die erste. Ein `%w`-gewickelter Pfad darf dort nicht
// hinaus (§8).
func TestSatzVonGibtNurDenEigenenSatz(t *testing.T) {
	t.Parallel()

	eigener := domain.MitGrund(domain.ErrConflict,
		errors.New("/Users/jemand/geheim.db: locked"),
		"die Laufzeit hat nicht geantwortet")

	if satz := domain.SatzVon(eigener); satz != "die Laufzeit hat nicht geantwortet" {
		t.Errorf("der Satz ist %q", satz)
	}

	// **Und er trägt das Wahrzeichen nicht mit** — anders als `Error()`.
	if strings.Contains(domain.SatzVon(eigener), "Konflikt") {
		t.Errorf("das Wahrzeichen steht im Satz: %q", domain.SatzVon(eigener))
	}

	for name, fremder := range map[string]error{
		"ein roher Fehler": errors.New("/Users/jemand/geheim.db: locked"),
		"gar keiner":       nil,
		"ein gewickelter": fmt.Errorf("beim Lesen: %w",
			errors.New("/Users/jemand/geheim.db: locked")),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if satz := domain.SatzVon(fremder); satz != "" {
				t.Errorf("es kam %q heraus", satz)
			}
		})
	}
}
