// **Ein anderer Nutzer wird man ausdrücklich** — Werkstück 17 von Phase 4.12
// (#1400),
// entschieden in
// [0074](../../docs/entscheidungen/0074-wer-arbeitet-steht-fest.md).
//
// # Warum das nicht der Profilwechsel nebenan ist
//
// **Weil eine Geste nicht zwei Bedeutungen haben soll.** Ein Profilwechsel
// heißt „ich sehe da hinein" und prüft **meine** Freigaben; ein Nutzerwechsel
// heißt „ich bin jetzt jemand anderes" und prüft **dessen** Ausweis. Wer
// beides an eine Geste hängte, machte aus einer abgelehnten Freigabe
// stillschweigend einen Identitätswechsel — und niemand sähe, wann er geschah.
//
// Der Auftraggeber hat sich am 2026-09-06 für die getrennte Geste entschieden.

package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// NutzerWechseln stellt um, wer arbeitet.
//
// # Der Ausweis ist der des **Ziels**
//
// **Nicht der eigene.** Wer den eigenen fragte, ließe jeden zu jedem werden,
// der seinen eigenen Code kennt — die Prüfung schützte dann den Falschen.
//
// **Wer sich selbst wählt, wird nicht gefragt.** Ein Ausweis für das, was
// ohnehin gilt, wäre eine Hürde ohne Zweck — und eine, die jemanden
// aussperrte, dessen Faktor gerade klemmt.
//
// **Und wer keinen Faktor führt, sperrt niemanden aus.** Dieselbe Zusage wie
// beim Profilwechsel seit Werkstück 7: ein Bestand ohne zweite Faktoren ist
// nicht geschützt, und das ist genau, wofür es sie gibt.
//
// # Der Zeiger zieht mit
//
// **Wer ein anderer Mensch wird, arbeitet in dessen Profil.** Bliebe der
// Zeiger stehen, stünde der neue Mensch in einem Profil, für das er womöglich
// keine Freigabe hat — und sähe Daten, die ihn nichts angehen.
func (v *Nutzerverwaltung) NutzerWechseln(
	ctx context.Context, nutzerID int64, nachweis string,
) error {
	werArbeitet, fehler := v.WerArbeitet(ctx)
	if fehler != nil {
		return fehler
	}

	if werArbeitet == nutzerID {
		return nil
	}

	ziel, suchfehler := v.profileVon(ctx, nutzerID)
	if suchfehler != nil {
		return suchfehler
	}

	if pruefung := v.faktorPruefen(ctx, nutzerID, nachweis); pruefung != nil {
		return pruefung
	}

	// **Beide Zeiger in einem Zug** — Befund 1 des Zweitblicks über die
	// Phasen 4.0 bis 4.13 (2026-09-07).
	//
	// Hier standen zwei Rufe hintereinander. Scheiterte der zweite, war der
	// erste durch: es arbeitete ein anderer Mensch, während das Profil des
	// alten gezeigt wurde — eine gemischte Identität, und der Ruf meldete
	// dazu einen Fehler.
	//
	// **Der Profilzeiger steht danach im ersten Profil des neuen Menschen**;
	// welches das ist, entscheidet dieselbe Reihenfolge, die `Profilzeiger`
	// als Rückfall nimmt.
	return v.zeiger.ZeigerSetzen(ctx, nutzerID, ziel.ID)
}

// Anmeldeprofil nennt das Profil, auf das eine Anmeldung zeigt — das erste
// des Nutzers, in derselben Reihenfolge, die `Profilzeiger` als Rückfall
// nimmt. **Es liest nur**: geschrieben wird in [Sitzung.Anmelden], zusammen
// mit dem Zustand (Nachlauf über ID-02, Befund 2). Vorher stellte es die
// Zeiger selbst, und scheiterte danach das Festhalten, standen sie schon.
func (v *Nutzerverwaltung) Anmeldeprofil(ctx context.Context, nutzerID int64) (int64, error) {
	ziel, fehler := v.profileVon(ctx, nutzerID)

	return ziel.ID, fehler
}

// profileVon gibt das erste Profil eines Nutzers — und weist ab, wenn es ihn
// nicht gibt.
//
// **Über die Übersicht und nicht über eine eigene Abfrage**, aus demselben
// Grund wie `profilSuchen` nebenan: sie liest ohnehin jeden Nutzer.
func (v *Nutzerverwaltung) profileVon(ctx context.Context, nutzerID int64) (
	domain.Profil, error,
) {
	zeilen, fehler := v.Uebersicht(ctx)
	if fehler != nil {
		return domain.Profil{}, fehler
	}

	for _, zeile := range zeilen {
		if zeile.Nutzer.ID != nutzerID {
			continue
		}

		if len(zeile.Profile) == 0 {
			// **Ein Nutzer ohne Profil ist kein Ziel.** Wer dorthin wechselte,
			// stünde ohne Arbeitsumgebung da — und der Profilzeiger fiele auf
			// ein fremdes zurück.
			return domain.Profil{}, fmt.Errorf(
				"%w: der Nutzer %d führt kein Profil", domain.ErrValidation, nutzerID)
		}

		return zeile.Profile[0], nil
	}

	return domain.Profil{}, fmt.Errorf(
		"%w: den Nutzer %d führt der Bestand nicht", domain.ErrNotFound, nutzerID)
}
