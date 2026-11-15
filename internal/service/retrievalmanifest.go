package service

// **Das Manifest einer neuen Generation** — Werkstück A15 von Phase 5.3
// (#1514).
//
// **Es steht neben dem Aufbau**, weil es eine eigene Frage beantwortet: dort
// „wie entsteht eine Generation?", hier „was sagt sie über sich?". Zusammen
// liefen beide über die Zeilengrenze aus §2.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// manifestFuer setzt das Manifest aus Wunsch, Raum und Korpusständen zusammen.
func manifestFuer(
	ctx context.Context, einbetter Einbetter,
	wunsch Aufbauwunsch, quelle Korpusquelle,
) (domain.Indexmanifest, error) {
	staende, fehler := quelle.Staende(ctx)
	if fehler != nil {
		return domain.Indexmanifest{}, fehler
	}

	raum := einbetter.Raum()

	manifest := domain.Indexmanifest{
		Schemafassung:  wunsch.Schemafassung,
		Zerteilfassung: wunsch.Zerteilfassung,
		Indexfassung:   wunsch.Indexfassung,
		Modell:         raum.Modell,
		Modellrevision: raum.Modellrevision,
		Artefakthash:   raum.Artefakthash,
		Laufzeit:       raum.Laufzeit,
		Abstandsmass:   raum.Abstandsmass,
		Vertrag:        raum.Vertrag,
		Korpusstaende:  staende,
		// **Die Suchkonfiguration ist keine Wahl des Aufrufers.** Wer sie
		// setzen dürfte, schriebe ein Verfahren ins Manifest, das der Dienst
		// gar nicht fährt.
		Suchkonfiguration: Rangfusionsname,
		Rerankerabdruck:   wunsch.Rerankerabdruck,
	}

	return manifest, manifest.Pruefe()
}
