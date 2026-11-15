package mcp_test

// **Ein Sitz lässt sich an ein Profil binden** — UXR-F21
// (#2041).
//
// **Bis dahin verwarf dieses Werkzeug `profilID` ohne ein Wort**: der
// Entzifferer ist nachsichtig, und eine Partie mit gebundenem Sitz sah aus
// wie eine ohne. Hier steht, dass die Angabe die Werkstatt erreicht und dass
// das Schema sie nennt — ein Modell kennt nur, was dort steht. Dass dieselbe
// Partie an allen Flächen denselben eigenen Sitz hat, hält
// `internal/flaechen/sitzprofil_test.go`.

import (
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Die Profilkennung kommt bis zur Werkstatt — je Sitz.**
func TestDasWerkzeugPartieAnlegenReichtDasProfilDesSitzesDurch(t *testing.T) {
	t.Parallel()

	probe := &partienprobe{}

	if _, err := rufePartiewerkzeug(t, probe, "schreiben.partie.anlegen",
		`{"spiel":"magic","sitze":[`+
			`{"kennung":"sitz-1","name":"Anna","deckID":"1"},`+
			`{"kennung":"sitz-2","name":"Bert","deckID":"2","profilID":"3"}]}`); err != nil {
		t.Fatalf("schreiben.partie.anlegen: %v", err)
	}

	if len(probe.gesehen.sitze) != 2 ||
		probe.gesehen.sitze[0].ProfilID != "" ||
		probe.gesehen.sitze[1].ProfilID != "3" {
		t.Errorf("gerufen mit %+v", probe.gesehen.sitze)
	}
}

// karteAus deutet einen Schemateil als Karte — oder bricht ab.
func karteAus(t *testing.T, wert any) map[string]any {
	t.Helper()

	karte, da := wert.(map[string]any)
	if !da {
		t.Fatalf("das Schema hat an dieser Stelle %T statt einer Karte", wert)
	}

	return karte
}

// **Das Schema nennt `profilID` — und verlangt es nicht.**
func TestDasSchemaVonPartieAnlegenNenntDasProfil(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: probenbestand(), Schreiber: probenbestand(),
		Partien: &partienprobe{},
	}) {
		if werkzeug.Name != "schreiben.partie.anlegen" {
			continue
		}

		sitze := karteAus(t, karteAus(t, werkzeug.Schema["properties"])["sitze"])
		je := karteAus(t, sitze["items"])

		eigenschaft := karteAus(t, karteAus(t, je["properties"])["profilID"])

		beschreibung, da := eigenschaft["description"].(string)
		if !da || !strings.Contains(beschreibung, "lesen.partien") {
			t.Errorf("das Schema nennt profilID ohne Wirkung: %v", eigenschaft)
		}

		pflichten, da := je["required"].([]string)
		if !da || slices.Contains(pflichten, "profilID") ||
			!slices.Contains(pflichten, "deckID") {
			t.Errorf("Pflichtfelder: %v — profilID ist freiwillig, deckID nicht", pflichten)
		}

		return
	}

	t.Fatal("das Werkzeug schreiben.partie.anlegen gibt es nicht")
}
