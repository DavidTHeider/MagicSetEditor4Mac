package storage

// **Nur vom Stand aus, den es noch gibt** — TRADE-03.AK-01
// (#1717).
//
// # Woran „noch gibt" gemessen wird
//
// **Am Zustand und an den Stimmen** der Revision, von der aus geschrieben
// wird — das sind die zwei Dinge, die ein Schritt an einer Revision ändert.
// Die Positionen ändert er nicht; wer sie ändert, legt eine neue Revision an,
// und die entsteht **einmal**: ein `INSERT` ohne `DO UPDATE`, und wer
// zweimal dieselbe anlegen will, kam vom alten Stand.
//
// **Schreiben bleibt daneben**, aber kein Dienstweg ruft es mehr: Eröffnen
// und Kopieren legen über einen leeren Vorstand an (Befund N2 des
// Nachlaufs über Phase 6.8). **Briefkasten und Abschluss schreiben seit dem
// Zweitblick über Phase 6.8 ebenfalls hierüber** (Befund Z1, #1669) — in
// ihrer eigenen Transaktion, gegen den Stand, den ihr Dienst gelesen hat.
// Vorher schrieben sie unbedingt, und eine Stimme, die gegen `angeboten`
// gelesen war, setzte einen angenommenen Handel still zurück.

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Fortschreiben legt `danach` ab — nur, wenn `vorher` noch der Stand ist.
func (a *Handelsablage) Fortschreiben(
	ctx context.Context, vorher, danach domain.Handelsangebot,
) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	return revisionFortschreiben(ctx, a.db, profilID, vorher, danach)
}

// revisionFortschreiben ist derselbe Schritt über einen [abfrager] — damit
// er auch **in** einer Transaktion läuft.
//
// **Ein leerer Vorstand** (Revision 0) heißt: es gab keinen; dann wird
// angelegt, wenn es die Revision noch nicht gibt.
func revisionFortschreiben(
	ctx context.Context, an abfrager, profilID int64, vorher, danach domain.Handelsangebot,
) error {
	roh, _ := json.Marshal(danach) //nolint:errcheck // Grund bei revisionSchreiben

	var (
		ergebnis interface{ RowsAffected() (int64, error) }
		fehler   error
	)

	if danach.Revision == vorher.Revision {
		stimmen, _ := json.Marshal(stimmenOderLeer(vorher.Zustimmungen)) //nolint:errcheck // eine Liste aus Zeichenketten

		ergebnis, fehler = an.ExecContext(ctx,
			`UPDATE handelsrevisionen
			    SET zustand = ?, angebot = ?, geaendert = datetime('now')
			  WHERE profil_id = ? AND vorgang = ? AND revision = ?
			    AND zustand = ?
			    AND coalesce(json_extract(angebot, '$.Zustimmungen'), '[]') = json(?)`,
			string(danach.Zustand), string(roh), profilID, danach.Vorgang,
			danach.Revision, string(vorher.Zustand), string(stimmen))
	} else {
		// **Und eine neue Revision nur auf dem Vorgänger, der gelesen wurde**
		// (Befund N1 des Nachlaufs über Phase 6.8): die freie Zielzeile
		// allein belegte nicht, dass er noch gilt. **Eine Anweisung**, damit
		// Prüfung und Anlage nicht auseinanderfallen; ohne Vorstand
		// (Revision 0) prüft `EXISTS` gegen eine Bedingung, die immer gilt.
		stimmen, _ := json.Marshal(stimmenOderLeer(vorher.Zustimmungen)) //nolint:errcheck // eine Liste aus Zeichenketten

		ergebnis, fehler = an.ExecContext(ctx,
			`INSERT INTO handelsrevisionen
			   (profil_id, vorgang, revision, abdruck, eltern, fassung, zustand,
			    angebot)
			 SELECT ?, ?, ?, ?, ?, ?, ?, ?
			  WHERE ? = 0 OR EXISTS (
			        SELECT 1 FROM handelsrevisionen
			         WHERE profil_id = ? AND vorgang = ? AND revision = ?
			           AND zustand = ?
			           AND coalesce(json_extract(angebot, '$.Zustimmungen'), '[]') = json(?))
			 ON CONFLICT (profil_id, vorgang, revision) DO NOTHING`,
			profilID, danach.Vorgang, danach.Revision, danach.Abdruck(),
			danach.Eltern, danach.Vertragsfassung, string(danach.Zustand),
			string(roh),
			vorher.Revision, profilID, vorher.Vorgang, vorher.Revision,
			string(vorher.Zustand), string(stimmen))
	}

	if fehler != nil {
		return uebersetze(fehler, "eine Revision fortschreiben")
	}

	if getroffen, _ := ergebnis.RowsAffected(); getroffen == 0 { //nolint:errcheck // SQLite meldet die Zahl immer
		return fmt.Errorf("%w: der Handel %s hat sich inzwischen geändert — "+
			"lade ihn neu", domain.ErrUeberholt, danach.Vorgang)
	}

	return nil
}

// stimmenOderLeer gibt eine Liste und nie `null` — so steht sie im Bestand.
func stimmenOderLeer(stimmen []string) []string {
	if stimmen == nil {
		return []string{}
	}

	return slices.Clone(stimmen)
}
