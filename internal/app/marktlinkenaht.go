package app

// **Die Naht der Marktlinks** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// **Weil nur hier `internal/storage` stehen darf** (§2). Nebenan steht, **was**
// ein Marktlink ist; hier steht, **woraus** er gebaut wird: aus den Kennungen
// eines Drucks (`A2`).

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// linkwerk baut die Wege für diesen Ruf.
func (a *App) linkwerk() *service.Linkwerk {
	return service.NeuesLinkwerk(storage.NeueDrucke(a.laufendeVerbindung()))
}

// MitOeffner setzt den Plattformöffner — für Tests und den Composition Root.
//
// **Einer für alle vier Flächen** — derselbe, den die Anmeldung nimmt: `open`
// und nicht `runtime.BrowserOpenURL`, denn jener bräuchte einen echten
// Wails-Kontext, und die drei Flächen ohne Fenster haben keinen.
//
// **Ohne Rückfall in der Naht.** Er wird beim Start gesetzt; ein Rückfall
// wäre ein Zweig, den im Betrieb nie jemand betritt (§10). Wer keinen setzt,
// bekommt die Meldung „diese Fläche kann keinen Browser öffnen" — und das ist
// eine Auskunft und kein Absturz.
func (a *App) MitOeffner(oeffnen service.Oeffner) { a.browseroeffner = oeffnen }
