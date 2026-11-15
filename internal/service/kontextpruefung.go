package service

// **Was vor dem Senden noch einmal geprüft wird** — Werkstück A16 von
// Phase 5.3
// (#1515).
//
// # Warum ein zweites Mal
//
// **Weil zwischen Zusammenstellen und Senden Zeit vergeht.** Das Issue zählt
// auf, was sich dabei ändern kann: „Revidierte Ziele, Profil-/Providerwechsel,
// Tokenwiderruf, gelöschte Quellen und abgelaufene Pakete vor
// Veröffentlichung/Senden erneut prüfen."
//
// **Ein Paket ist eine Momentaufnahme.** Es sagt, was zu einem Zeitpunkt
// galt; ob es noch gilt, sagt diese Datei.
//
// # Und was hier nicht geprüft wird
//
// **Ob die Belege die Aussage tragen.** Das kann kein Programm — „ID-Gültigkeit
// beweist noch keine inhaltliche Folgerichtigkeit", und der Satz steht in
// `belegpruefung.go`, weil er dort gebraucht wird.
//
// **Gelöschte Quellen prüft diese Datei nicht** — und das ist eine
// Entscheidung (#2436). Hier stand `PruefeBelegeNoch`: ein Beleg, dessen
// Quelle verschwunden ist, hätte seinen Ausschnitt verloren und nur die
// Herkunft behalten. Gerufen hat es niemand, und es gab nichts zu rufen: **der
// Index trägt im Betrieb nur öffentliche Quellen** — `NeuerScryfallkorpus` ist
// der einzige Korpus, gelesen am 2026-10-16 —, und Kontextpakete werden nicht
// gespeichert. Eine eigene Karte erreicht kein Paket, also auch keine
// Historie. **Wer einen Korpus eigener Inhalte baut, baut diese Prüfung mit**:
// eine gelöschte oder widerrufene Quelle darf danach keinen Ausschnitt mehr
// tragen.

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sendeziel ist, wohin ein Paket gehen soll.
type Sendeziel struct {
	Anbieter string
	Modell   string
	Laufzeit string

	// Zielrevision ist der Stand der Karte **jetzt**.
	Zielrevision string
}

// PruefeVorDemSenden sagt, ob ein Paket noch gilt.
//
// **Fünf Absagen, und jede sagt etwas anderes.** Wer sie zusammenfasste,
// ließe jemanden raten, ob er das Paket neu bauen, das Profil wechseln oder
// die Karte zurücksetzen muss.
func (d *Kontextdienst) PruefeVorDemSenden(
	ctx context.Context, manifest domain.Kontextmanifest, ziel Sendeziel,
) error {
	if manifest.Abgelaufen(d.jetzt()) {
		return domain.MitGrund(domain.ErrConflict, nil,
			"dieses Kontextpaket ist seit %v abgelaufen — es wird neu "+
				"zusammengestellt, damit die Belege zum heutigen Stand passen",
			d.jetzt().Sub(manifest.Laeuftab).Round(time.Second))
	}

	if fehler := d.pruefeProfil(ctx, manifest); fehler != nil {
		return fehler
	}

	if !manifest.PasstZu(domain.Kontextmanifest{
		Zielanbieter: ziel.Anbieter, Zielmodell: ziel.Modell,
		Laufzeit: ziel.Laufzeit, Zaehlweise: manifest.Zaehlweise,
	}) {
		return domain.MitGrund(domain.ErrConflict, nil,
			"dieses Kontextpaket ist für %s/%s budgetiert, gesendet werden "+
				"soll es an %s/%s — dieselbe Aufgabe genügt nicht, das "+
				"Fenster ist ein anderes",
			manifest.Zielanbieter, manifest.Zielmodell,
			ziel.Anbieter, ziel.Modell)
	}

	if ziel.Zielrevision != "" && manifest.Zielrevision != "" &&
		ziel.Zielrevision != manifest.Zielrevision {
		return domain.MitGrund(domain.ErrConflict, nil,
			"die Zielkarte stand bei der Zusammenstellung auf %q und steht "+
				"jetzt auf %q — ein Vorschlag daraus ginge gegen einen Stand, "+
				"den es nicht mehr gibt",
			manifest.Zielrevision, ziel.Zielrevision)
	}

	return nil
}

// pruefeProfil hält das Paket bei dem, für den es gebaut wurde.
func (d *Kontextdienst) pruefeProfil(
	ctx context.Context, manifest domain.Kontextmanifest,
) error {
	profil, fehler := d.profil.Profil(ctx)
	if fehler != nil {
		return fehler
	}

	if profil != manifest.ProfilID {
		return domain.MitGrund(domain.ErrConflict, nil,
			"dieses Kontextpaket gehört einem anderen Profil — seine Belege "+
				"stammen aus einer Sicht, die hier nicht gilt")
	}

	return nil
}
