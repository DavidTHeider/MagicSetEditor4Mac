// **Die umgelenkte Freigabeadresse** —
// #1100.
//
// Innen, weil die Umlenkung innen liegt: von außen ist nur zu sehen, dass eine
// Anfrage irgendwo hingeht. Dieselbe Datei wie `scryfallumgebung_intern_test.go`
// und aus demselben Grund.
package app

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
)

// **Eine gesetzte Umgebungsvariable schlägt die echte Adresse.**
func TestDieFreigabeadresseFolgtDerUmgebung(t *testing.T) {
	t.Setenv(Fassungsumgebung, "http://127.0.0.1:1/probe")

	if ort := umgelenkteFreigabeadresse(); ort != "http://127.0.0.1:1/probe" {
		t.Errorf("die Adresse ist %q", ort)
	}
}

// **Und ohne sie steht die echte da.**
func TestDieFreigabeadresseIstSonstGitHub(t *testing.T) {
	t.Setenv(Fassungsumgebung, "")

	if ort := umgelenkteFreigabeadresse(); ort != aktualisierung.Adresse {
		t.Errorf("die Adresse ist %q", ort)
	}
}
