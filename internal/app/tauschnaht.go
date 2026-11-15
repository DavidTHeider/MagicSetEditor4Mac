package app

// **Die Naht des Tauschabgleichs** — Werkstück `T4` von Phase 5.6
// (#1469).
//
// # Warum sie in einer eigenen Datei steht
//
// **Weil nur hier `internal/storage` stehen darf** (§2) — derselbe Grund wie
// bei [freigabenaht]. Nebenan steht, **was** ein Abgleich ist; hier steht,
// **woraus** er gebaut wird.
//
// # Warum die Preisreihe hier gewählt wird und nicht im Dienst
//
// **Weil sie eine Entscheidung dessen ist, der tauscht.** Das Werk nimmt sie
// entgegen und erfindet keine (A3); welche gilt, steht in den Vorgaben dieses
// Profils, und die zu lesen ist Sache der Naht.

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// tauschwerk baut den Abgleich für diesen Ruf.
//
// **Mit der Naht zu den Zusagen** (T2, #1467): was zugesagt ist, steht nicht
// als überzählig da und wird deshalb auch nicht angeboten.
func (a *App) tauschwerk() *service.Tauschwerk {
	db := a.laufendeVerbindung()

	return service.NeuesTauschwerk(
		service.NeueSammlungssichten(storage.NeuerSammlungsspeicher(db)).
			MitReservierungen(storage.NeueReservierungsablage(db)),
		storage.NeueDrucke(db),
	)
}
