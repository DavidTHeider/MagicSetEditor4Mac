package app

// **Ein Zugriffstoken wirkt nur an seinem Profil** — QA-09
// (#1740),
// Befund P2 in `docs/PENTEST.md`.
//
// # Der Befund
//
// **Die Rolle am Zielprofil fragte den Menschen, nicht das Token** (#1817):
// `rolleUndUrteil` liest, wer am Rechner arbeitet, und dessen Rolle — die
// des Eigners. Ein Automat mit einem Token für **sein** Profil legte sich
// damit am 2026-10-02 ein schreibendes Token **im Profil des Menschen** an,
// samt Klartext, setzte dort Vorgaben und zog dort Token zurück; über die
// Wege am Menschen benannte er ihn um. Gemessen am gebauten Dienst
// (`internal/flaechen/tokenfremdprofil_test.go`).
//
// # Die Regel
//
// **Ein Token reicht nie weiter als sein Profil** (`AUTORISIERUNG.md`) —
// und was einem Menschen gehört, verwaltet es gar nicht: Nutzer, Faktoren,
// Konten, Personendaten. Das geschieht am Fenster, an der Kommandozeile oder
// mit der Prozessmarke, also dort, wo jemand den Prozess selbst gestartet
// hat. **Ohne Token ändert sich nichts.**

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// nurAnSeinemProfil weist ein Token ab, das an einem anderen Profil wirken
// will als dem, dem es gehört.
//
// **Das eigene Profil ist das gezeigte**: die Tokenwache hängt es an den
// Kontext ([Tokenwache.Pruefe]), und es gewinnt.
func nurAnSeinemProfil(ctx context.Context, profilID int64) error {
	if service.TokenumfangAus(ctx) == "" || imGezeigten(ctx, profilID) {
		return nil
	}

	return fmt.Errorf("%w: ein Zugriffstoken wirkt nur an seinem eigenen Profil",
		domain.ErrErlaubnis)
}

// keinTokenAmMenschen weist ein Token ab, das einen Menschen verwalten will.
func keinTokenAmMenschen(ctx context.Context) error {
	if service.TokenumfangAus(ctx) == "" {
		return nil
	}

	return fmt.Errorf("%w: ein Zugriffstoken verwaltet keinen Menschen — "+
		"das geht am Fenster, an der Kommandozeile oder mit der Prozessmarke",
		domain.ErrErlaubnis)
}
