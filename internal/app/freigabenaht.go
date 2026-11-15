package app

// **Die Naht der freigegebenen Sichten** — Werkstück `T3` von Phase 5.6
// (#1468).
//
// # Warum sie in einer eigenen Datei steht
//
// **Weil nur hier `internal/storage` stehen darf** (§2): `internal/app`
// enthält keine fachlichen Entscheidungen, und die benannte Ausnahme sind die
// **Composition Roots** — die Stellen, an denen die Repositories für einen Ruf
// entstehen. Nebenan steht, **was** eine Freigabe ist und wie sie an die
// Oberfläche kommt; hier steht, **woraus** sie gebaut wird.
//
// # Und warum die Verbindung geholt und nicht festgehalten wird
//
// **Weil ein Profilwechsel sie tauscht.** Dieselbe Form wie bei
// [kontaktnaht]: wer sie beim Bauen festhielte, schriebe nach einem Wechsel
// in den Bestand von vorhin.

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// freigabepflege baut die Pflege für diesen Ruf.
//
// **Mit der Naht zu den Zusagen** (T2, #1467, 0120): was einem Handel zugesagt
// ist, steht nicht in der Sicht — sonst böte sie es einem zweiten Partner an.
func (a *App) freigabepflege() *service.Freigabepflege {
	db := a.laufendeVerbindung()

	return service.NeueFreigabepflege(
		storage.NeueFreigabeablage(db),
		storage.NeueKontaktablage(db),
		storage.NeuerSammlungsspeicher(db),
		a.uhr,
	).MitReservierungen(storage.NeueReservierungsablage(db)).
		MitHandelsschranke(einstellungsschranke{a: a})
}
