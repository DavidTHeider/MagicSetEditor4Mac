package service

// **Was eine Rücksetzung tut** — Werkstück O4 von Phase 5.2
// (#1428).
//
// Aus `ruecksetzung.go` herausgelöst, als es über 400 Zeilen ging (§2). Der
// Schnitt folgt der Frage: dort steht, **was betroffen wäre** — die Vorschau,
// ihr Umfang, ihre Marke —, hier, **was geschieht**, wenn jemand sie
// bestätigt.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// gebundenAn baut die Prüfung, die im Schreibgeschäft noch einmal nachsieht.
//
// **Die dritte Hälfte von Befund 9** des Nachlaufs zum Zweitblick über Phase
// 5.2: der Vergleich stand vor den Schreibaufrufen, nicht in ihrem Geschäft.
// Zwischen dem Zählen und dem Löschen liegt eine Lücke, und wer sie trifft,
// verliert etwas, das in keiner Vorschau stand.
//
// **Der Vergleich oben bleibt trotzdem stehen.** Er weist ab, **bevor** ein
// Nachweis eingelöst ist; dieser hier ist der, der die Zusage hält.
//
// **Sie wird nur für die löschenden Arten gebaut**, und die verlangen eine
// Revision (siehe [Ruecksetzung.Zuruecksetzen]). Ein Zweig für „keine genannt"
// wäre unerreichbar — ein toter Zweig ist schlimmer als keiner, weil er
// behauptet, es gebe einen Fall, den niemand betreten kann (§3).
//
// **Dass die Ablage eine fehlende Bindung verträgt, bleibt trotzdem geprüft**:
// die vier anderen Flächen löschen ohne Vorschau und geben gar keine mit
// (`storage.gebundenerUmfang`).
func gebundenAn(
	art domain.Ruecksetzungsart, revision string,
) func(profile []domain.Profil, umfaenge [][]domain.Betroffenes) error {
	return func(profile []domain.Profil, umfaenge [][]domain.Betroffenes) error {
		if domain.Ruecksetzungsmarke(art, gesamtumfang(art, profile, umfaenge)) == revision {
			return nil
		}

		return fmt.Errorf(
			"%w: seit der Vorschau hat sich geändert, was betroffen wäre. Es "+
				"wurde nichts zurückgesetzt — sieh es dir noch einmal an",
			domain.ErrConflict)
	}
}

// Ruecksetzungsbilanz ist, was aus einer Rücksetzung wurde.
type Ruecksetzungsbilanz struct {
	Art domain.Ruecksetzungsart

	// Zurueckgesetzt ist die Zahl der zurückgenommenen Wahlen.
	Zurueckgesetzt int

	// Beiseite ist der Ort des alten Bestands — nur beim Neuanfang.
	Beiseite string
}

// Zuruecksetzen führt die Rücksetzung aus.
//
// **Sie sieht vorher noch einmal nach.** Zwischen Vorschau und Bestätigung
// liegt eine Entscheidung eines Menschen, und in dieser Zeit kann eine
// Rückkehr entstehen — dieselbe Überlegung wie bei der Revision der
// Datenpflege ([0089](../../docs/entscheidungen/0089-ein-befund-ist-kein-auftrag.md)).
func (r *Ruecksetzung) Zuruecksetzen(
	ctx context.Context, art domain.Ruecksetzungsart, ziel Ruecksetzungsziel,
	nachweis, revision string,
) (Ruecksetzungsbilanz, error) {
	vorschau, ziel, fehler := r.vorschauUndZiel(ctx, art, ziel)
	if fehler != nil {
		return Ruecksetzungsbilanz{}, fehler
	}

	// **Gelöscht wird, was jemand gesehen hat** — Befund 12 des Zweitblicks
	// über Phase 5.2. Vorher rechnete dieser Weg eine neue Vorschau und
	// löschte auf **deren** Stand; was zwischen dem Hinsehen und dem Klicken
	// dazukam, ging ohne Zustimmung mit.
	//
	if revision != "" && revision != vorschau.Revision {
		return Ruecksetzungsbilanz{}, fmt.Errorf(
			"%w: seit der Vorschau hat sich geändert, was betroffen wäre. Es "+
				"wurde nichts zurückgesetzt — sieh es dir noch einmal an",
			domain.ErrConflict)
	}

	if !vorschau.Bereit() {
		return Ruecksetzungsbilanz{}, fmt.Errorf(
			"%w: %s", domain.ErrConflict, satzZuHindernis(vorschau.Hindernisse[0]))
	}

	// **Was Daten löscht, braucht sie** — Befund 9 des Nachlaufs zum
	// Zweitblick über Phase 5.2. Eine leere Revision hob die Bindung vorher
	// vollständig auf; für die Arten, die etwas wegnehmen, ist das kein Weg,
	// sondern eine Tür daneben. Für die drei, die nur eine **Wahl**
	// zurücknehmen (Layout, Profil- und Nutzereinstellungen), bleibt sie
	// wahlfrei: dort gibt es nichts, was zwischen Vorschau und Klick
	// dazukommen und ungefragt verschwinden könnte.
	//
	// **Und sie steht hinter den Hindernissen**, nicht davor: wer ein fremdes
	// Profil nennt, soll das erfahren und nicht, dass eine Revision fehlt.
	if revision == "" && art.LoeschtDaten() {
		return Ruecksetzungsbilanz{}, domain.NewFieldError(
			"revision", domain.SchluesselValidierung,
			"Zum Zurücksetzen gehört die Revision aus der Vorschau.")
	}

	// **Die Erlaubnis nach der Vorschau und vor der Wirkung.** Davor wäre ein
	// Nachweis für etwas eingelöst, das an einem Hindernis scheitert; danach
	// wäre es zu spät.
	if absage := r.erlaubt(ctx, art, nachweis); absage != nil {
		return Ruecksetzungsbilanz{}, absage
	}

	bilanz := Ruecksetzungsbilanz{Art: art}

	switch art {
	case domain.RuecksetzungLayout:
		bilanz.Zurueckgesetzt, fehler = r.ablage.EinstellungenZuruecksetzen(
			ctx, domain.Layouteinstellungen())
	case domain.RuecksetzungProfileinstellungen:
		bilanz.Zurueckgesetzt, fehler = r.ablage.EinstellungenZuruecksetzen(ctx, nil)
	case domain.RuecksetzungNutzereinstellungen:
		bilanz.Zurueckgesetzt, fehler = r.ablage.NutzereinstellungenZuruecksetzen(
			ctx, ziel.NutzerID)
	case domain.RuecksetzungProfil:
		fehler = r.loescher.ProfilLoeschen(ctx, ziel.ProfilID, ziel.Name,
			gebundenAn(art, revision))
	case domain.RuecksetzungNutzer:
		fehler = r.loescher.NutzerLoeschen(ctx, ziel.NutzerID, ziel.Name,
			gebundenAn(art, revision))
	case domain.RuecksetzungAnwendung:
		bilanz.Beiseite, fehler = r.neuanfang(ctx)
	}

	if fehler != nil {
		return bilanz, fehler
	}

	return bilanz, nil
}
