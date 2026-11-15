package domain_test

// **Was ein Filter wissen muss** — Werkstück A15 von Phase 5.3
// (#1514).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einManawert gibt einen Zeiger auf ihn.
func einManawert(wert float64) *float64 { return &wert }

// **Die Farbidentität umfasst die Farben — Regel 903.4.**
func TestDieFarbidentitaetUmfasstDieFarben(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		merkmale domain.Retrievalmerkmale
		haelt    bool
	}{
		"rot und rot-identisch": {
			domain.Retrievalmerkmale{
				Farben: []string{"R"}, Farbidentitaet: []string{"R"},
			}, true,
		},
		"farblos und rot-identisch": {
			// **Der Normalfall, nicht der Sonderfall**: ein `{R}` im
			// Regeltext färbt die Identität, nicht die Karte.
			domain.Retrievalmerkmale{Farbidentitaet: []string{"R"}}, true,
		},
		"rot, aber grün-identisch": {
			domain.Retrievalmerkmale{
				Farben: []string{"R"}, Farbidentitaet: []string{"G"},
			}, false,
		},
		"eine Farbe, die es nicht gibt": {
			domain.Retrievalmerkmale{
				Farben: []string{"P"}, Farbidentitaet: []string{"P"},
			}, false,
		},
		"gar nichts": {domain.Retrievalmerkmale{}, true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := fall.merkmale.Pruefe()
			if fall.haelt != (fehler == nil) {
				t.Fatalf("Pruefe: %v", fehler)
			}

			if fehler != nil && !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("das Wahrzeichen fehlt: %v", fehler)
			}
		})
	}
}

// **Eine Legalität ohne Stand gibt es nicht.**
func TestEineLegalitaetOhneStandGibtEsNicht(t *testing.T) {
	t.Parallel()

	ohne := domain.Retrievalmerkmale{Formate: map[string]string{"modern": "legal"}}
	if ohne.Pruefe() == nil {
		t.Error("Formate ohne Datenstand kamen durch")
	}

	if _, _, bekannt := ohne.Format("modern"); bekannt {
		t.Error("die Legalität kam trotzdem heraus")
	}

	mit := ohne
	mit.Formatstand = "2026-09-12"

	if fehler := mit.Pruefe(); fehler != nil {
		t.Fatalf("Pruefe: %v", fehler)
	}

	lage, stand, bekannt := mit.Format("Modern")
	if !bekannt || lage != "legal" || stand != "2026-09-12" {
		t.Fatalf("%q, %q, %v", lage, stand, bekannt)
	}

	// **Ein unbekanntes Format ist kein erfundener Treffer.**
	if _, _, bekannt := mit.Format("pauper"); bekannt {
		t.Error("ein nicht geführtes Format kam als bekannt zurück")
	}
}

// **Ein Manawert von null ist etwas anderes als keiner.**
func TestEinManawertVonNullIstEtwasAnderesAlsKeiner(t *testing.T) {
	t.Parallel()

	land := domain.Retrievalmerkmale{Manawert: einManawert(0), Set: "neo"}
	if land.Leer() {
		t.Error("ein Land mit Manawert 0 gilt als leer")
	}

	if (domain.Retrievalmerkmale{}).Leer() != true {
		t.Error("nichts gilt nicht als leer")
	}
}

// **Eine Kopie teilt nichts mit dem Original.**
func TestEineMerkmalskopieTeiltNichts(t *testing.T) {
	t.Parallel()

	urbild := domain.Retrievalmerkmale{
		Farben: []string{"R"}, Farbidentitaet: []string{"R"},
		Typen: []string{"creature"}, Manawert: einManawert(3),
		Formate: map[string]string{"modern": "legal"}, Formatstand: "heute",
	}

	kopie := urbild.Kopie()
	kopie.Farben[0] = "G"
	kopie.Typen[0] = "instant"
	kopie.Formate["modern"] = "banned"
	*kopie.Manawert = 7

	if urbild.Farben[0] != "R" || urbild.Typen[0] != "creature" ||
		urbild.Formate["modern"] != "legal" || *urbild.Manawert != 3 {
		t.Fatalf("das Urbild hat sich mitgeändert: %+v", urbild)
	}

	// **Und eine leere Kopie erfindet keine Map.**
	if (domain.Retrievalmerkmale{}).Kopie().Formate != nil {
		t.Error("aus nichts wurde eine Map")
	}
}

// **Ein Dokument mit widersprüchlichen Merkmalen kommt nicht durch.**
//
// **Der Weg dahin gehört mitgeprüft**: [domain.Retrievaldokument.Pruefe]
// ruft die Merkmalsprüfung, und ohne diesen Lauf stünde sie da, ohne je
// betreten zu werden.
func TestEinDokumentMitWidersprechendenMerkmalenKommtNichtDurch(t *testing.T) {
	t.Parallel()

	dokument := einDokument()
	dokument.Merkmale = domain.Retrievalmerkmale{
		Farben: []string{"R"}, Farbidentitaet: []string{"G"},
	}

	if fehler := dokument.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Pruefe: %v", fehler)
	}
}
