package app

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Token wirkt nur an seinem Profil, und nie am Menschen** — QA-09
// (#1740), Befund P2. Der Lauf am gebauten Dienst steht in
// `internal/flaechen/tokenfremdprofil_test.go`; hier stehen die Wachen an
// den drei Stellen, an denen sie greifen.
//
// **Mit einer leeren Anwendung**: die Wache steht vor jeder Frage an
// Nutzerverwaltung und Freigaben. Kommt ein Token durch sie hindurch, fragte
// die Anwendung danach den **Menschen** — genau das war der Befund.
func TestEinTokenWirktNurAnSeinemProfil(t *testing.T) {
	t.Parallel()

	alsToken := service.MitTokenumfang(storage.MitProfil(t.Context(), 7), domain.ScopeSchreiben)
	leer := &App{}

	for name, fehler := range map[string]error{
		"Rolle am fremden Profil": func() error {
			_, fehler := leer.rolleUndUrteil(alsToken, 1, domain.RolleMitarbeiter)

			return fehler
		}(),
		"über ein fremdes Profil": leer.ueberProfil(alsToken, 1),
		"über einen Menschen":     leer.ueberNutzer(alsToken, 1),
	} {
		if !errors.Is(fehler, domain.ErrErlaubnis) {
			t.Errorf("%s: %v — erwartet war eine Absage an der Erlaubnis", name, fehler)
		}
	}
}

// **Und ohne Token oder im eigenen Profil ändert sich nichts** — die
// Gegenprobe (§10).
func TestOhneTokenOderImEigenenProfilFragtDieWacheNichts(t *testing.T) {
	t.Parallel()

	alsToken := service.MitTokenumfang(storage.MitProfil(t.Context(), 7), domain.ScopeSchreiben)
	ohneToken := storage.MitProfil(t.Context(), 7)

	if fehler := nurAnSeinemProfil(alsToken, 7); fehler != nil {
		t.Errorf("im eigenen Profil: %v", fehler)
	}

	if fehler := nurAnSeinemProfil(ohneToken, 1); fehler != nil {
		t.Errorf("ohne Token an einem anderen Profil: %v", fehler)
	}

	if fehler := keinTokenAmMenschen(ohneToken); fehler != nil {
		t.Errorf("ohne Token am Menschen: %v", fehler)
	}
}

// **Was einem Menschen gehört, liest ein Token nicht** — Konten und
// persönliche Einstellungen, die zwei Lese-Ausnahmen neben den Wachen oben.
func TestEinTokenLiestKeinesMenschenKontenUndEinstellungen(t *testing.T) {
	t.Parallel()

	alsToken := service.MitTokenumfang(storage.MitProfil(t.Context(), 7), domain.ScopeLesen)
	leer := &App{}

	konten := leer.Anbieterkonten(alsToken, "1")
	if konten.Fehler == nil || konten.Konten == nil || len(konten.Konten) != 0 {
		t.Errorf("Konten: %+v — erwartet war eine leere Liste mit Absage", konten)
	}

	einstellungen := leer.Nutzereinstellungen(alsToken, "1")
	if einstellungen.Fehler == nil || einstellungen.Einstellungen == nil ||
		len(einstellungen.Einstellungen) != 0 {
		t.Errorf("Einstellungen: %+v — erwartet war eine leere Liste mit Absage", einstellungen)
	}
}
