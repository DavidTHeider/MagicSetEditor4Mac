package flaechen_test

// **Einen Handel wählen, statt seine Kennung zu tippen — über beide
// Flächen** — UX-08.AK-04
// (#1680).
//
// Eröffnet wird an der Kommandozeile **ohne** Kennung; die Leitung vergibt
// eine. Dann nennen Kommandozeile und Dienst die Handel mit dem Gegenüber —
// und beide denselben. §16: dieselbe Fähigkeit, überall dieselbe Antwort.

import (
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// derVorgaengeweg ist der Endpunkt, wie er im Katalog steht.
const derVorgaengeweg = "GET /peer/kontakte/{kennung}/handel"

// vorgaengeweg setzt das Gegenüber in den Weg ein.
func vorgaengeweg(gegenueber string) string {
	return strings.ReplaceAll(pfadAus(derVorgaengeweg), "{kennung}", gegenueber)
}

func TestDieHandelMitEinemGegenueberSindUeberBeideFlaechenDieselben(t *testing.T) {
	umgebung := append(mitEigenerSaat(t),
		schluesselbund.Bundort+"="+probebund(t))
	an := starteDienst(t, umgebung)

	mussKommandozeile(t, umgebung, "peer-anlegen")

	// **Ein leerer Vorgang lässt die Kennung vergeben.**
	eroeffnet := mitEingabe(t, umgebung, diePositionen,
		"handel", "eroeffnen", "", "tausch", dasGegenueberImHandel)

	kennung, _, _ := strings.Cut(eroeffnet, "\t")
	if !strings.HasPrefix(kennung, "h-") {
		t.Fatalf("vergeben wurde %q", eroeffnet)
	}

	if aus := mussKommandozeile(t, umgebung, "handel", "vorgaenge",
		dasGegenueberImHandel); !strings.Contains(aus, kennung) {
		t.Errorf("die Kommandozeile nennt %q nicht: %s", kennung, aus)
	}

	if aus := mussREST(t, an, http.MethodGet,
		vorgaengeweg(dasGegenueberImHandel), "",
	); !strings.Contains(aus, `"vorgang":"`+kennung+`"`) {
		t.Errorf("der Dienst nennt %q nicht: %s", kennung, aus)
	}

	// **Mit einem anderen keiner** — über beide Flächen.
	if aus := mussKommandozeile(t, umgebung, "handel", "vorgaenge",
		"bjemandanders"); strings.TrimSpace(aus) != "" {
		t.Errorf("die Kommandozeile nennt für einen anderen: %s", aus)
	}

	if aus := mussREST(t, an, http.MethodGet,
		vorgaengeweg("bjemandanders"), "",
	); !strings.Contains(aus, `"vorgaenge":[]`) {
		t.Errorf("der Dienst nennt für einen anderen: %s", aus)
	}
}
