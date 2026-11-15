package service

// **Massenvorgänge über eine Auswahl** — Werkstück 6 von Phase 4.5
// (#1292).
//
// # Der Befund
//
// Bis hierher ändert man ein Exemplar. Wer fünfzig Karten aus einer Kiste in
// einen Ordner räumt, klickt fünfzigmal — und wer eine ganze Kiste verkauft
// hat, löscht fünfzigmal.
//
// # Warum jede Zeile ihre Fassung mitbringt
//
// **§7, Regel 2**: eine Antwort gehört zu dem Stand, aus dem sie stammt. Ein
// Massenvorgang über bloße Kennungen überschriebe jede Änderung, die jemand
// seit dem Laden der Liste gemacht hat — und zwar fünfzigfach auf einmal.
//
// **Was das kostet, steht ausgesprochen**: der Aufrufer muss die Fassungen
// mitführen. Er hat sie: sie stehen in jeder Zeile, die er zeigt.
//
// # Warum die Änderungen Zeiger sind
//
// Weil „nicht ändern" und „auf den Nullwert setzen" verschiedene Dinge sind.
// Ein leerer Lagerort ist eine gültige Angabe — „nimm die Ortsangabe weg" —,
// und ohne Zeiger wäre sie von „lass den Lagerort in Ruhe" nicht zu
// unterscheiden.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Massenzeile ist ein Exemplar samt der Fassung, aus der es gelesen wurde.
type Massenzeile struct {
	ID      int64
	Fassung int64
}

// Massenaenderung sagt, was an den genannten Zeilen zu ändern ist.
//
// **Ein `nil` heißt „unverändert".** Ein Wert heißt „setze auf diesen" — auch
// wenn er der Nullwert ist.
type Massenaenderung struct {
	Zeilen []Massenzeile

	Menge    *int
	Zustand  *domain.Zustand
	Lagerort *string
	Wunsch   *bool
}

// Massenbericht sagt, was geschehen ist.
type Massenbericht struct {
	// Geaendert ist die Zahl der Zeilen, die geschrieben wurden.
	Geaendert int

	// Befunde nennt die Zeilen, die nicht durchkamen — mit ihrer Kennung.
	//
	// **Eine Zeile, deren Fassung veraltet ist, hält die anderen nicht auf**
	// (§8). Wer sie übergeht, sieht in diesem Bericht, welche es war.
	Befunde []domain.Befund

	// Vorher nennt je geschriebener Zeile, was sie **vor** der Änderung trug
	// — DOM-03.AK-03 (#1698). **Damit sich eine Änderung zurücknehmen
	// lässt**: die Sammlung hat keinen Verlauf wie ein Set, und die alten
	// Werte kennt nur dieser Schreibvorgang.
	Vorher []Massenstand
}

// Massenstand ist eine Zeile, wie sie vor der Änderung war — **samt der
// Fassung, die sie danach trägt**. Eine Rücknahme schreibt die alten Werte
// gegen diese Fassung; hat seither jemand anderes geschrieben, wird es zum
// Befund statt zum Überschreiben (§7, Regel 2).
type Massenstand struct {
	ID       int64
	Fassung  int64
	Menge    int
	Zustand  domain.Zustand
	Lagerort string
	Wunsch   bool
}

// Massenpflege führt Vorgänge über eine Auswahl.
type Massenpflege struct {
	haushaltswahl

	ablage Sammlungsablage
}

// NeueMassenpflege baut sie.
func NeueMassenpflege(ablage Sammlungsablage) *Massenpflege {
	return &Massenpflege{ablage: ablage}
}

// Aendern schreibt die genannte Änderung an alle genannten Zeilen.
//
// **So viele wie möglich** (§8). Eine Zeile, die inzwischen jemand anderes
// geändert hat, wird zum Befund; die übrigen gehen durch.
func (m *Massenpflege) Aendern(ctx context.Context, aenderung Massenaenderung) (
	Massenbericht, error,
) {
	freigeben, platzfehler := m.platz(ctx, domain.RessourceMasse)
	if platzfehler != nil {
		return Massenbericht{}, platzfehler
	}

	defer freigeben()

	if len(aenderung.Zeilen) == 0 {
		return Massenbericht{}, fmt.Errorf(
			"%w: ohne eine einzige Zeile gibt es nichts zu ändern",
			domain.ErrValidation)
	}

	if aenderung.leer() {
		return Massenbericht{}, fmt.Errorf(
			"%w: es ist nicht gesagt, was sich ändern soll", domain.ErrValidation)
	}

	bericht := Massenbericht{Befunde: []domain.Befund{}, Vorher: []Massenstand{}}

	// **Die ganze Liste einmal**, nicht je Zeile ein Nachschlagen: bei fünfzig
	// Zeilen wären das fünfzig Abfragen — derselbe Fehler, den das Portfolio
	// gemacht hat (docs/SAMMLUNG.md).
	vorhandene, err := m.nachKennung(ctx)
	if err != nil {
		return bericht, err
	}

	for _, zeile := range aenderung.Zeilen {
		if fehler := m.eineAendern(ctx, zeile, aenderung, vorhandene); fehler != nil {
			bericht.Befunde = append(bericht.Befunde, domain.Befund{
				Ort:     fmt.Sprintf("Exemplar %d", zeile.ID),
				Meldung: fehler.Error(),
				Art:     domain.ArtEingabe,
			})

			continue
		}

		// **Die Fassung danach ist die gelesene plus eins** — jedes Schreiben
		// hebt sie um genau eins (`domain.Exemplar.Fassung`).
		vorher := vorhandene[zeile.ID]
		bericht.Vorher = append(bericht.Vorher, Massenstand{
			ID: zeile.ID, Fassung: zeile.Fassung + 1, Menge: vorher.Menge,
			Zustand: vorher.Zustand, Lagerort: vorher.Lagerort, Wunsch: vorher.Wunsch,
		})
		bericht.Geaendert++
	}

	return bericht, nil
}

// eineAendern schreibt eine Zeile.
func (m *Massenpflege) eineAendern(
	ctx context.Context, zeile Massenzeile, aenderung Massenaenderung,
	vorhandene map[int64]*domain.Exemplar,
) error {
	gefunden, da := vorhandene[zeile.ID]
	if !da {
		return fmt.Errorf("%w: Exemplar %d gibt es nicht", domain.ErrNotFound, zeile.ID)
	}

	geaendert := *gefunden
	geaendert.Fassung = zeile.Fassung

	if aenderung.Menge != nil {
		geaendert.Menge = *aenderung.Menge
	}

	if aenderung.Zustand != nil {
		geaendert.Zustand = *aenderung.Zustand
	}

	if aenderung.Lagerort != nil {
		geaendert.Lagerort = *aenderung.Lagerort
	}

	if aenderung.Wunsch != nil {
		geaendert.Wunsch = *aenderung.Wunsch
	}

	if fehler := vorbereiten(&geaendert); fehler != nil {
		return fehler
	}

	return m.ablage.Aendern(ctx, &geaendert)
}

// Loeschen nimmt die genannten Zeilen heraus.
//
// **Gegen die gelesene Fassung** (§7, Regel 2): ohne sie träfe ein Löschen
// auch eine Zeile, die inzwischen jemand anderes geändert hat.
func (m *Massenpflege) Loeschen(ctx context.Context, zeilen []Massenzeile) (
	Massenbericht, error,
) {
	freigeben, platzfehler := m.platz(ctx, domain.RessourceMasse)
	if platzfehler != nil {
		return Massenbericht{}, platzfehler
	}

	defer freigeben()

	if len(zeilen) == 0 {
		return Massenbericht{}, fmt.Errorf(
			"%w: ohne eine einzige Zeile gibt es nichts zu löschen",
			domain.ErrValidation)
	}

	bericht := Massenbericht{Befunde: []domain.Befund{}}

	for _, zeile := range zeilen {
		if fehler := m.ablage.Loeschen(ctx, zeile.ID, zeile.Fassung); fehler != nil {
			bericht.Befunde = append(bericht.Befunde, domain.Befund{
				Ort:     fmt.Sprintf("Exemplar %d", zeile.ID),
				Meldung: fehler.Error(),
				Art:     domain.ArtEingabe,
			})

			continue
		}

		bericht.Geaendert++
	}

	return bericht, nil
}

// nachKennung liest die Sammlung und legt sie nach Kennung ab.
func (m *Massenpflege) nachKennung(ctx context.Context) (
	map[int64]*domain.Exemplar, error,
) {
	exemplare, err := m.ablage.Liste(ctx)
	if err != nil {
		return nil, err
	}

	nach := make(map[int64]*domain.Exemplar, len(exemplare))
	for _, exemplar := range exemplare {
		nach[exemplar.ID] = exemplar
	}

	return nach, nil
}

// leer sagt, ob die Änderung gar nichts ändert.
func (a Massenaenderung) leer() bool {
	return a.Menge == nil && a.Zustand == nil &&
		a.Lagerort == nil && a.Wunsch == nil
}
