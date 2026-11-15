package service_test

// **Der Deltaimport** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Der Satz, um den es geht
//
// **„Inkrementelle Änderungen" und „atomar aktiviert" sind kein
// Widerspruch.** Was sich nicht geändert hat, wird **kopiert** statt neu
// eingebettet — der teure Teil ist das Modell, nicht der `INSERT` —, und
// sichtbar wird alles zusammen.
//
// **Und wer mit einem fremden Modell fortschreibt, legte zwei Vektorräume in
// eine Generation.** Dann sind die Abstände zwischen alten und neuen
// Abschnitten Zahlen ohne Bedeutung — und der Index sähe aus wie einer.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Fortschreibung übernimmt, was die Quelle nicht liefert.**
func TestEineFortschreibungUebernimmtDasUebrige(t *testing.T) {
	t.Parallel()

	schreiber := &schreiberdoppel{uebernommen: 28}

	bericht, fehler := eineVerwaltung(&retrievalablagedoppel{
		manifest: einRetrievalmanifest(),
	}, schreiber, dieQuelleMit(einRetrievalmanifest().Modell),
		&korpusoeffnerdoppel{}).Aufbauen(t.Context(), service.Aufbauauftrag{
		Modell: einRetrievalmanifest().Modell, Fortschreiben: true,
	})
	if fehler != nil {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	// **Zwei geschrieben, achtundzwanzig übernommen.**
	if bericht.Geschrieben != 30 {
		t.Fatalf("der Bericht zählt %d", bericht.Geschrieben)
	}

	// **Und übergangen wurde genau, was aus der Quelle kam** — sonst stünde
	// ein Abschnitt zweimal in derselben Generation.
	if len(schreiber.ausser) != 2 {
		t.Fatalf("übergangen wurde %v", schreiber.ausser)
	}
}

// **Ohne Fortschreibung wird nichts übernommen.**
func TestOhneFortschreibungWirdNichtsUebernommen(t *testing.T) {
	t.Parallel()

	schreiber := &schreiberdoppel{uebernommen: 28}

	bericht, fehler := eineVerwaltung(&retrievalablagedoppel{
		manifest: einRetrievalmanifest(),
	}, schreiber, dieQuelleMit(einRetrievalmanifest().Modell),
		&korpusoeffnerdoppel{}).Aufbauen(t.Context(), service.Aufbauauftrag{
		Modell: einRetrievalmanifest().Modell,
	})
	if fehler != nil {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	if bericht.Geschrieben != 2 || schreiber.ausser != nil {
		t.Fatalf("der Bericht zählt %d, übergangen wurde %v",
			bericht.Geschrieben, schreiber.ausser)
	}
}

// **Eine Fortschreibung mit fremdem Modell wird abgewiesen.**
func TestEineFortschreibungMitFremdemModellWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	// **Gleiche Dimension, anderes Modell** — genau der Fall, den eine
	// Längenprüfung durchließe.
	fremd := einRetrievalmanifest().Raum()
	fremd.Modell = "ein-anderes"

	quelle := &einbetterquellendoppel{nach: map[string]service.Einbetter{
		"ein-anderes": &retrievaleinbetterdoppel{
			raum: fremd, vektor: []float32{1, 0, 0},
		},
	}}

	schreiber := &schreiberdoppel{}

	_, fehler := eineVerwaltung(&retrievalablagedoppel{
		manifest: einRetrievalmanifest(),
	}, schreiber, quelle, &korpusoeffnerdoppel{}).Aufbauen(t.Context(),
		service.Aufbauauftrag{Modell: "ein-anderes", Fortschreiben: true})

	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "Vektorraum") {
		t.Errorf("die Meldung ist %q", fehler.Error())
	}

	// **Und es ist nichts angelegt worden.**
	if schreiber.naechste != 0 {
		t.Errorf("es wurde %d angelegt", schreiber.naechste)
	}
}

// **Ohne aktiven Index gibt es nichts fortzuschreiben.**
func TestOhneAktivenIndexGibtEsNichtsFortzuschreiben(t *testing.T) {
	t.Parallel()

	for name, ablage := range map[string]*retrievalablagedoppel{
		"kein Index": {kein: true},
		"Bestand unlesbar": {manifestfehler: domain.MitGrund(
			domain.ErrConflict, nil, "die Datei ist zu")},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := eineVerwaltung(ablage, &schreiberdoppel{},
				dieQuelleMit("all-minilm"), &korpusoeffnerdoppel{}).Aufbauen(
				t.Context(), service.Aufbauauftrag{
					Modell: "all-minilm", Fortschreiben: true,
				}); fehler == nil {
				t.Fatal("die Fortschreibung kam durch")
			}
		})
	}
}

// **Scheitert die Übernahme, wird die Generation verworfen.**
func TestEineGescheiterteUebernahmeVerwirftDieGeneration(t *testing.T) {
	t.Parallel()

	schreiber := &schreiberdoppel{
		uebernahmefehler: domain.MitGrund(domain.ErrConflict, nil,
			"die alte Generation ist weg"),
	}

	if _, fehler := eineVerwaltung(&retrievalablagedoppel{
		manifest: einRetrievalmanifest(),
	}, schreiber, dieQuelleMit(einRetrievalmanifest().Modell),
		&korpusoeffnerdoppel{}).Aufbauen(t.Context(), service.Aufbauauftrag{
		Modell: einRetrievalmanifest().Modell, Fortschreiben: true,
	}); fehler == nil {
		t.Fatal("die gescheiterte Übernahme kam durch")
	}

	if len(schreiber.verworfen) != 1 || len(schreiber.aktiviert) != 0 {
		t.Fatalf("verworfen %v, aktiviert %v",
			schreiber.verworfen, schreiber.aktiviert)
	}
}
