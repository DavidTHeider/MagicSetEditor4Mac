package app

import (
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Guthaben und Sperre für die Wurzelknoten** — Werkstück O5 von Phase 5.2
// (#1429). `rest` und `mcp` kennen `service` nicht (§2); die Programme, die
// sie starten, holen beides hier — mit den Grenzen des Maschinenhaushalts.

// NeueKontingente sind die Guthaben je Principal und global dieses Prozesses
// — mit den Grenzen aus [GrenzenUeber], damit der Schalter aus dem Bestand
// gilt (QA-08.AK-03).
func NeueKontingente(grenzen domain.Haushaltsgrenzen, uhr func() time.Time) *service.Kontingente {
	return service.NeueKontingente(grenzen, uhr)
}

// NeueZugangssperre ist die Sperre je vorgezeigtem Nachweis.
func NeueZugangssperre(uhr func() time.Time) *service.Zugangssperre {
	return service.NeueZugangssperre(uhr)
}
