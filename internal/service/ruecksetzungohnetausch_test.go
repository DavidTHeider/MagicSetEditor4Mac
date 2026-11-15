package service_test

// **Ohne Tausch ist die ganze Anwendung ein Hindernis** — G09
// (#2273).
//
// Eine Fläche, die ihren Bestand über ihre ganze Laufzeit hält, gibt keinen
// Neuanfang mit. Die Vorschau sagt es vorher, und die Ausführung hält an,
// **bevor** der Nachweis vorgezeigt ist: ein Mensch hat ihn am Fenster für
// etwas gegeben, das hier nicht geschieht.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestOhneTauschIstDieGanzeAnwendungEinHindernis(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{}
	wacht := wache()
	setzung := service.NeueRuecksetzung(doppel, doppel, wacht, nil)

	vorschau, fehler := setzung.Vorschau(t.Context(),
		domain.RuecksetzungAnwendung, service.Ruecksetzungsziel{})
	if fehler != nil {
		t.Fatalf("Vorschau: %v", fehler)
	}

	if vorschau.Bereit() || len(vorschau.Hindernisse) != 1 ||
		vorschau.Hindernisse[0].Art != domain.RuecksetzungOhneTausch {
		t.Errorf("die Vorschau nennt %v — erwartet allein %q",
			vorschau.Hindernisse, domain.RuecksetzungOhneTausch)
	}

	_, fehler = setzung.Zuruecksetzen(t.Context(),
		domain.RuecksetzungAnwendung, service.Ruecksetzungsziel{}, "abc123", vorschau.Revision)

	switch {
	case !errors.Is(fehler, domain.ErrConflict):
		t.Errorf("die Ausführung: %v — erwartet ein Konflikt", fehler)
	case !strings.Contains(fehler.Error(), "Kommandozeile"):
		t.Errorf("der Satz sagt nicht, wo es geht: %v", fehler)
	}

	if wacht.kennung != "" {
		t.Errorf("der Nachweis %q wurde vorgezeigt, obwohl nichts geschieht", wacht.kennung)
	}

	// **Die übrigen fünf brauchen keinen Tausch** — die Gegenprobe, dass das
	// Hindernis an der Art hängt und nicht an der Fläche.
	if _, fehler := setzung.Zuruecksetzen(t.Context(),
		domain.RuecksetzungLayout, service.Ruecksetzungsziel{}, "", ""); fehler != nil {
		t.Errorf("das Layout ohne Tausch: %v", fehler)
	}
}
