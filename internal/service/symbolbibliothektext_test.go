package service_test

// **Ein Symbol als Text hinaus und herein** — DOM-02
// (#1697).
// Für MCP und REST, die an keiner Platte stehen, die ihr Gegenüber sieht —
// dieselbe Abwägung wie bei der Turnierdatei (Vermerk 0140).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestEinSymbolGehtAlsTextHinausUndHerein(t *testing.T) {
	bibliothek, _, _ := bibliotheksplatz(t)
	krone := kroneAufgenommen(t, bibliothek)

	gelesen, err := bibliothek.Symbol(t.Context(), krone.ID)
	if err != nil || gelesen.Inhalt != krone.Inhalt || gelesen.Name != "Krone" {
		t.Fatalf("lesen: %+v, %v", gelesen, err)
	}

	// **Derselbe Text ist dasselbe Symbol** — der Abdruck stimmt überein.
	reif, err := bibliothek.EinlesenText(t.Context(), " Reif ", []byte(gelesen.Inhalt))
	if err != nil || reif.Name != "Reif" || reif.Abdruck != krone.Abdruck {
		t.Fatalf("einlesen: %+v, %v", reif, err)
	}

	if _, err := bibliothek.Symbol(t.Context(), 99); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein fehlendes: %v", err)
	}
}

func TestEinTextSagtWasMitIhmNichtStimmt(t *testing.T) {
	bibliothek, _, _ := bibliotheksplatz(t)

	var feldfehler *domain.FieldError

	// **Ohne Namen** geht es nicht — ein Text hat keinen Dateinamen.
	if _, err := bibliothek.EinlesenText(t.Context(), " ", []byte("part:\n")); !errors.As(err, &feldfehler) ||
		feldfehler.Field != "name" {
		t.Errorf("ohne Namen: %v", err)
	}

	// **Zu viele Teile** kommen mit ihrem Schlüssel.
	viel := strings.Repeat("part:\n\ttype: shape\n", msepaket.GrenzeSymbolteile+1)
	if _, err := bibliothek.EinlesenText(t.Context(), "Viel", []byte(viel)); !errors.As(err, &feldfehler) ||
		feldfehler.Key != domain.SchluesselSymbolZuVieleTeile {
		t.Errorf("zu viele Teile: %v", err)
	}

	// **Zu groß** ist ein eigener Schlüssel, mit der Grenze.
	gross := make([]byte, msepaket.GrenzeSymboldatei+1)
	if _, err := bibliothek.EinlesenText(t.Context(), "Gross", gross); !errors.As(err, &feldfehler) ||
		feldfehler.Key != domain.SchluesselSymbolZuGross || feldfehler.Werte["grenze"] != "1 MiB" {
		t.Errorf("zu groß: %v", err)
	}
}
