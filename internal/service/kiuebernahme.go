package service

// **Die Übernahme: Auswahl und Herkunft** — Werkstück A11 von Phase 5.3
// (#1450).
//
// # Eine eigene Datei
//
// `kiauftragswerk.go` stand mit ihr über der Zeilengrenze aus §2, und der
// Schnitt folgt der Frage: dort steht, **wie ein Auftrag entsteht und was
// ihn gültig hält** — hier, **was beim Schreiben geschieht**.

import (
	"context"
	"fmt"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Uebernehmen schreibt die **ausgewählten** Felder — und prüft vorher alles.
//
// # Warum eine Auswahl und nicht alles
//
// **Das Issue verlangt es** (A11, #1450): „Nur ausgewählte gültige Felder
// werden geändert." Ein Vorschlag über acht Felder ist selten in allen acht
// gut; wer ihn nur ganz oder gar nicht übernehmen könnte, würfe sieben gute
// Werte weg oder nähme einen schlechten mit.
//
// **Leer heißt nicht »alles«**, sondern ist ein Fehler — das wäre die
// stillste Art, eine Karte zu überschreiben.
// **`profil` ist der Mensch, der jetzt fragt** (Befund 1 des Nachlaufs über
// Phase 5.3). Ein Vorschlag gehört dem Profil, das ihn angelegt hat; das Buch
// überlebt einen Profilwechsel, und ohne diese Grenze schriebe eine Übernahme
// in Profil B die Karte von A.
func (w *KIAuftragswerk) Uebernehmen(
	ctx context.Context, kennung string, auswahl []string, profil int64,
) (KIUebernahme, error) {
	lauf, gibts := w.lauf(kennung)
	if !gibts || lauf.auftrag.ProfilID != profil {
		// **Dieselbe Meldung für beides.** „Es gibt ihn, aber nicht für dich"
		// wäre eine Auskunft über ein fremdes Profil; „in diesem Profil gibt
		// es ihn nicht" ist wahr und sagt, was zu tun ist.
		return KIUebernahme{}, fmt.Errorf("%w: den Auftrag %q gibt es nicht",
			domain.ErrNotFound, kennung)
	}

	if lauf.vorschlag == nil {
		return KIUebernahme{}, fmt.Errorf(
			"%w: zu %q gibt es keinen Vorschlag", domain.ErrNotFound, kennung)
	}

	gewaehlt, fehler := gewaehlteFelder(lauf.vorschlag.Felder, auswahl)
	if fehler != nil {
		return KIUebernahme{}, fehler
	}

	if lauf.auftrag.Abgelaufen(w.jetzt()) {
		return KIUebernahme{}, fmt.Errorf(
			"%w: die Frist für %q ist abgelaufen — ein Vorschlag von gestern "+
				"trifft den Stand von heute vielleicht nicht mehr",
			domain.ErrConflict, kennung)
	}

	if fehler := w.pruefeStand(ctx, lauf.auftrag); fehler != nil {
		return KIUebernahme{}, fehler
	}

	// **Werte und Herkunft gehen in einem Ruf hinaus** ([KIKarten]) — und
	// **mit dem Stand, auf dem der Auftrag steht**: `pruefeStand` darüber
	// hat ihn gelesen, geschrieben wird er unter derselben Bedingung.
	if fehler := w.karten.FelderSetzen(ctx, lauf.auftrag.Zielkarte, gewaehlt,
		w.herkunftFuer(lauf, gewaehlt), lauf.auftrag.Zielrevision); fehler != nil {
		return KIUebernahme{}, fehler
	}

	uebernahme := KIUebernahme{Felder: len(gewaehlt), Namen: feldnamenAus(gewaehlt)}

	// **Die Kennung ist konsumierbar** (Issue): zweimal übernehmen hieße,
	// denselben Vorschlag auf einen Stand zu legen, der ihn schon trägt.
	//
	// **Und erst nach dem Schreiben** — scheitert es, bleibt der Auftrag
	// offen, und ein zweiter Versuch ist möglich.
	w.buch.verbrauche(kennung)

	return uebernahme, nil
}

// gewaehlteFelder gibt die ausgewählten Werte — und weist ab, was fehlt.
//
// **Erst prüfen, dann schreiben**: eine Auswahl mit einem Tippfehler darin
// darf nicht dazu führen, dass die Felder davor schon stehen.
func gewaehlteFelder(
	felder map[string][]string, auswahl []string,
) (map[string][]string, error) {
	if len(auswahl) == 0 {
		return nil, fmt.Errorf(
			"%w: es ist kein Feld ausgewählt — eine Übernahme sagt, was sie "+
				"übernimmt", domain.ErrValidation)
	}

	gewaehlt := make(map[string][]string, len(auswahl))

	for _, name := range auswahl {
		werte, dabei := felder[name]
		if !dabei {
			return nil, fmt.Errorf(
				"%w: der Vorschlag trägt kein Feld %q — ausgewählt werden "+
					"kann nur, was darin steht", domain.ErrNotFound, name)
		}

		gewaehlt[name] = werte
	}

	return gewaehlt, nil
}

// herkunftFuer baut je übernommenem Feld seine Herkunft.
//
// **Sie entsteht hier und nicht in der Ablage**: was ein Wert bedeutet —
// welcher Auftrag, welches Modell, welche Aufgabe —, weiß dieses Werk, und
// die Ablage schreibt, was sie bekommt.
func (w *KIAuftragswerk) herkunftFuer(
	lauf *kiLauf, gewaehlt map[string][]string,
) []domain.KIFeldherkunft {
	jetzt := w.jetzt()

	herkunft := make([]domain.KIFeldherkunft, 0, len(gewaehlt))

	for _, name := range feldnamenAus(gewaehlt) {
		herkunft = append(herkunft, domain.KIFeldherkunft{
			Feld: name, Anbieter: lauf.auftrag.Anbieter,
			Modell: lauf.vorschlag.Modell, Aufgabe: lauf.vorschlag.Aufgabe,
			Auftrag: lauf.auftrag.Kennung, Zeitpunkt: jetzt,
			Bezug: domain.KIInhaltsbezug(gewaehlt[name]),
		})
	}

	return herkunft
}

// feldnamenAus gibt die Feldnamen in fester Ordnung.
//
// **Sortiert** — eine Karte hat keine Reihenfolge, und eine Herkunft, deren
// Zeilen bei jedem Lauf anders stehen, ist nicht zu vergleichen.
func feldnamenAus(felder map[string][]string) []string {
	namen := make([]string, 0, len(felder))
	for name := range felder {
		namen = append(namen, name)
	}

	slices.Sort(namen)

	return namen
}
