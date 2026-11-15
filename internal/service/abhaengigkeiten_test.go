package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Die Abhängigkeitsauflösung (#57).
//
// **Über den Katalog und nicht über die Platte**: die Angaben stehen seit #54
// im Bestand, und ein zweiter Lauf über 324 Pakete wäre dieselbe Arbeit noch
// einmal.

// mitPaketen baut einen Verwalter über einem festen Katalog.
func mitPaketen(t *testing.T, pakete ...domain.Paket) *service.Paketverwalter {
	t.Helper()

	return service.NeuerPaketverwalter(&ablage{pakete: pakete}, t.TempDir(), t.TempDir())
}

// **Eine fehlende Abhängigkeit ist ein Befund, kein Abbruch.**
func TestEineFehlendeAbhaengigkeitWirdBenannt(t *testing.T) {
	verwalter := mitPaketen(t, domain.Paket{
		Art: "mse-style", Name: "magic-m15", Pfad: "/p",
		Braucht: []domain.Bedarf{{Paket: "magic.mse-game", Version: "2009-07-23"}},
	})

	befunde, err := verwalter.Abhaengigkeiten(t.Context())
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	if len(befunde) != 1 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	if befunde[0].Art != service.AbhaengigkeitFehlt ||
		befunde[0].Paket != "magic-m15.mse-style" ||
		befunde[0].Braucht != "magic.mse-game" {
		t.Errorf("Befund: %+v", befunde[0])
	}
}

// **Eine zu alte Abhängigkeit wird benannt, bevor eine Vorlage stillschweigend
// falsch rendert.**
//
// Ein Stil rechnet mit den Funktionen seines Spiels; ist das Spiel älter als
// verlangt, fehlen sie — und was dabei herauskommt, ist eine Karte, der man es
// nicht ansieht.
func TestEineZuAlteAbhaengigkeitWirdBenannt(t *testing.T) {
	verwalter := mitPaketen(t,
		domain.Paket{Art: "mse-game", Name: "magic", Version: "2009-07-23", Pfad: "/p"},
		domain.Paket{
			Art: "mse-style", Name: "magic-m15", Pfad: "/p",
			Braucht: []domain.Bedarf{{Paket: "magic.mse-game", Version: "2024-01-01"}},
		},
	)

	befunde, err := verwalter.Abhaengigkeiten(t.Context())
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	if len(befunde) != 1 || befunde[0].Art != service.AbhaengigkeitZuAlt {
		t.Fatalf("Befunde: %+v", befunde)
	}

	if befunde[0].Verlangt != "2024-01-01" || befunde[0].Da != "2009-07-23" {
		t.Errorf("Befund: %+v", befunde[0])
	}
}

// **Eine neuere Abhängigkeit ist keine.**
//
// MSE-Versionen sind Datumsangaben, und `depends on:` nennt ein **Mindest**maß.
// Wer eine neuere Fassung abwiese, machte jede Aktualisierung zum Fehler.
func TestEineNeuereAbhaengigkeitIstKeinBefund(t *testing.T) {
	verwalter := mitPaketen(t,
		domain.Paket{Art: "mse-game", Name: "magic", Version: "2024-05-26", Pfad: "/p"},
		domain.Paket{
			Art: "mse-style", Name: "magic-m15", Pfad: "/p",
			Braucht: []domain.Bedarf{{Paket: "magic.mse-game", Version: "2009-07-23"}},
		},
	)

	befunde, err := verwalter.Abhaengigkeiten(t.Context())
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde: %+v", befunde)
	}
}

// **Dieselbe Version genügt** — `depends on:` nennt ein Mindestmaß.
func TestDieselbeVersionGenuegt(t *testing.T) {
	verwalter := mitPaketen(t,
		domain.Paket{Art: "mse-game", Name: "magic", Version: "2009-07-23", Pfad: "/p"},
		domain.Paket{
			Art: "mse-style", Name: "magic-m15", Pfad: "/p",
			Braucht: []domain.Bedarf{{Paket: "magic.mse-game", Version: "2009-07-23"}},
		},
	)

	befunde, err := verwalter.Abhaengigkeiten(t.Context())
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde: %+v", befunde)
	}
}

// **Ohne verlangte Version gibt es nichts zu vergleichen.**
//
// `depends on:` ohne `version:` kommt vor; daraus ein Urteil zu machen hieße,
// eine Regel zu erfinden — auch eine Ablehnung braucht Deckung (§10).
func TestOhneVerlangteVersionKeinUrteil(t *testing.T) {
	verwalter := mitPaketen(t,
		domain.Paket{Art: "mse-game", Name: "magic", Version: "2009-07-23", Pfad: "/p"},
		domain.Paket{
			Art: "mse-style", Name: "magic-m15", Pfad: "/p",
			Braucht: []domain.Bedarf{{Paket: "magic.mse-game"}},
		},
	)

	befunde, err := verwalter.Abhaengigkeiten(t.Context())
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde: %+v", befunde)
	}
}

// **Eine krumme Version ist ein eigener Befund, kein Urteil.**
//
// Ein Vergleich gegen etwas, das kein Datum ist, wäre eine Behauptung
// (Phase 2, Punkt 3).
func TestEineKrummeVersionWirdBenanntStattVerglichen(t *testing.T) {
	faelle := map[string]struct{ vorhanden, verlangt string }{
		"die installierte ist krumm": {vorhanden: "übermorgen", verlangt: "2024-01-01"},
		"die verlangte ist krumm":    {vorhanden: "2024-01-01", verlangt: "bald"},
		"die installierte fehlt":     {vorhanden: "", verlangt: "2024-01-01"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			verwalter := mitPaketen(t,
				domain.Paket{Art: "mse-game", Name: "magic", Version: fall.vorhanden, Pfad: "/p"},
				domain.Paket{
					Art: "mse-style", Name: "magic-m15", Pfad: "/p",
					Braucht: []domain.Bedarf{{Paket: "magic.mse-game", Version: fall.verlangt}},
				},
			)

			befunde, err := verwalter.Abhaengigkeiten(t.Context())
			if err != nil {
				t.Fatalf("prüfen: %v", err)
			}

			if len(befunde) != 1 || befunde[0].Art != service.AbhaengigkeitUnlesbar {
				t.Errorf("Befunde: %+v", befunde)
			}
		})
	}
}

// **Ein Zyklus wird erkannt** — auch über drei Pakete.
func TestEinZyklusWirdErkannt(t *testing.T) {
	verwalter := mitPaketen(t,
		domain.Paket{
			Art: "mse-style", Name: "eins", Pfad: "/p",
			Braucht: []domain.Bedarf{{Paket: "zwei.mse-style"}},
		},
		domain.Paket{
			Art: "mse-style", Name: "zwei", Pfad: "/p",
			Braucht: []domain.Bedarf{{Paket: "drei.mse-style"}},
		},
		domain.Paket{
			Art: "mse-style", Name: "drei", Pfad: "/p",
			Braucht: []domain.Bedarf{{Paket: "eins.mse-style"}},
		},
	)

	befunde, err := verwalter.Abhaengigkeiten(t.Context())
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	zyklen := 0

	for _, befund := range befunde {
		if befund.Art == service.AbhaengigkeitZyklus {
			zyklen++
		}
	}

	if zyklen != 1 {
		t.Errorf("Befunde: %+v", befunde)
	}
}

// **Ein Paket, das sich selbst braucht, ist auch ein Zyklus.**
func TestEinSelbstbezugIstEinZyklus(t *testing.T) {
	verwalter := mitPaketen(t, domain.Paket{
		Art: "mse-style", Name: "eins", Pfad: "/p",
		Braucht: []domain.Bedarf{{Paket: "eins.mse-style"}},
	})

	befunde, err := verwalter.Abhaengigkeiten(t.Context())
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	if len(befunde) != 1 || befunde[0].Art != service.AbhaengigkeitZyklus {
		t.Errorf("Befunde: %+v", befunde)
	}
}

// **Ein Baum ohne Kreis gibt keinen Zyklusbefund** — auch wenn zwei Wege zum
// selben Paket führen.
func TestEinBaumOhneKreisGibtKeinenZyklus(t *testing.T) {
	verwalter := mitPaketen(t,
		domain.Paket{Art: "mse-game", Name: "magic", Version: "2009-07-23", Pfad: "/p"},
		domain.Paket{
			Art: "mse-style", Name: "eins", Pfad: "/p",
			Braucht: []domain.Bedarf{{Paket: "magic.mse-game"}, {Paket: "zwei.mse-style"}},
		},
		domain.Paket{
			Art: "mse-style", Name: "zwei", Pfad: "/p",
			Braucht: []domain.Bedarf{{Paket: "magic.mse-game"}},
		},
	)

	befunde, err := verwalter.Abhaengigkeiten(t.Context())
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde: %+v", befunde)
	}
}

// **Ein Fehler des Bestands kommt durch.**
func TestDerBestandsfehlerKommtDurchDieAbhaengigkeiten(t *testing.T) {
	kaputt := errors.New("der Katalog schweigt")
	verwalter := service.NeuerPaketverwalter(&ablage{fehler: kaputt}, t.TempDir(), t.TempDir())

	if _, err := verwalter.Abhaengigkeiten(t.Context()); !errors.Is(err, kaputt) {
		t.Errorf("Fehler: %v", err)
	}
}
