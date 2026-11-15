package service_test

// **Die Vorgangskennung im Kontext** — Werkstück O9 von Phase 5.2
// (#1433).
//
// **Hier und nicht nur in `internal/app`**: die Naht gehört diesem Paket, und
// eine Naht, die nur über ihren Aufrufer geprüft wird, ist ungeprüft, sobald
// der Aufrufer wegfällt. Gefunden hat es die Abdeckungsschwelle der Pipeline —
// 8.741 von 8.745 Anweisungen, und `go test` rundet das auf „100,0 %".

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDerKontextTraegtDieVorgangskennung(t *testing.T) {
	t.Parallel()

	ctx := service.MitVorgang(t.Context(), domain.Vorgangskennung("0a1b2c3d"))

	if kennung := service.VorgangAus(ctx); kennung != "0a1b2c3d" {
		t.Errorf("VorgangAus = %q", kennung)
	}
}

// **Ohne Grenze ist sie leer, und das ist ein gültiger Wert.** Ein Test, der
// einen nackten Kontext baut, hat keine durchlaufen; eine erfundene Kennung
// behauptete einen Vorgang, den es nicht gibt.
func TestOhneGrenzeGibtEsKeineVorgangskennung(t *testing.T) {
	t.Parallel()

	if kennung := service.VorgangAus(t.Context()); kennung != "" {
		t.Errorf("VorgangAus = %q, erwartet leer", kennung)
	}
}

// **Und der zweite Ruf überschreibt den ersten.** Das ist keine Zusage an die
// Anwendung — dort würfelt nur die Grenze —, sondern die Eigenschaft eines
// Kontexts, auf die sich [service.MitVorgang] verlässt.
func TestEineZweiteKennungGiltAbDaWoSieGesetztWurde(t *testing.T) {
	t.Parallel()

	erster := service.MitVorgang(t.Context(), domain.Vorgangskennung("11111111"))
	zweiter := service.MitVorgang(erster, domain.Vorgangskennung("22222222"))

	if kennung := service.VorgangAus(erster); kennung != "11111111" {
		t.Errorf("der erste Kontext trägt %q", kennung)
	}

	if kennung := service.VorgangAus(zweiter); kennung != "22222222" {
		t.Errorf("der zweite Kontext trägt %q", kennung)
	}
}
