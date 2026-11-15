package app

// **Ein eigenes Zeugnis trägt jede gewählte Adresse** — G08
// (#2272).
//
// Die erste Adresse geht ins Zeugniswerk, und jede weitere wird ebenso
// gegen das Blatt gehalten: sonst lauschte der Dienst auf einer Adresse mit
// einem Zeugnis, das die Gegenseite abweist. `0.0.0.0` prüft nichts — so
// war es vorher, und eine unbestimmte Adresse hat keinen Namen im Zeugnis.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
)

func TestEinEigenesZeugnisMussJedeAdresseTragen(t *testing.T) {
	t.Parallel()

	anwendung, setze := einDienst(t)
	pki := eineHausPKI(t, "Haus-Wurzel")

	setze(domain.DienstTLSZertifikat, pki.Zertifikat)
	setze(domain.DienstTLSSchluessel, pki.Schluessel)
	setze(domain.DienstTLSAnker, pki.Anker)

	if _, _, fehler := anwendung.Dienstzeugnis(riegelkontext(t), []string{"0.0.0.0"}, leitung.Zeugnisdateien{}, stumm); fehler != nil {
		t.Fatalf("unbestimmt: %v", fehler)
	}

	_, _, fehler := anwendung.Dienstzeugnis(riegelkontext(t), []string{"0.0.0.0", "192.0.2.10"}, leitung.Zeugnisdateien{}, stumm)
	if fehler == nil || !strings.Contains(fehler.Error(), "192.0.2.10") {
		t.Errorf("eine Adresse, die das Zeugnis nicht nennt, kam durch: %v", fehler)
	}
}
