// **Massenvorgänge und Preise in Masse, für alle vier Flächen** —
// Werkstück 6 von Phase 4.5
// (#1292).
//
// **Zwei Nähte in einer Datei, weil sie zusammen gerufen werden**: wer
// fünfzig Zeilen ändert, will danach ihre Preise — und beide gehören zur
// Frage „wie führe ich eine große Sammlung".

package app

import (
	"context"
	"database/sql"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// MassenzeileDTO ist ein Exemplar samt der Fassung, aus der es gelesen wurde.
//
// **Die Fassung geht mit** (§7, Regel 2): ein Massenvorgang über bloße
// Kennungen überschriebe jede Änderung, die jemand seit dem Laden gemacht
// hat — und zwar fünfzigfach auf einmal.
type MassenzeileDTO struct {
	ID      int64 `json:"id"`
	Fassung int64 `json:"fassung"`
}

// MassenaenderungDTO sagt, was an den genannten Zeilen zu ändern ist.
//
// **Zeiger heißen „unverändert".** Ein leerer Lagerort ist eine gültige
// Angabe — „nimm die Ortsangabe weg" —, und ohne Zeiger wäre sie von „lass ihn
// in Ruhe" nicht zu unterscheiden.
type MassenaenderungDTO struct {
	Zeilen []MassenzeileDTO `json:"zeilen"`

	Menge    *int    `json:"menge"`
	Zustand  *string `json:"zustand"`
	Lagerort *string `json:"lagerort"`
	Wunsch   *bool   `json:"wunsch"`
}

// MassenberichtDTO sagt, was geschehen ist.
type MassenberichtDTO struct {
	Geaendert int      `json:"geaendert"`
	Befunde   []string `json:"befunde"`

	// Vorher nennt je geänderter Zeile den Stand davor samt neuer Fassung —
	// damit sich eine Änderung zurücknehmen lässt (DOM-03.AK-03, #1698).
	// Ein Löschen nennt keinen.
	Vorher []MassenstandDTO `json:"vorher"`
}

// MassenstandDTO ist eine Zeile vor der Änderung, mit der Fassung danach.
type MassenstandDTO struct {
	ID       int64  `json:"id"`
	Fassung  int64  `json:"fassung"`
	Menge    int    `json:"menge"`
	Zustand  string `json:"zustand"`
	Lagerort string `json:"lagerort"`
	Wunsch   bool   `json:"wunsch"`
}

// PreisschubDTO sagt, was ein Schub geholt hat und was bleibt.
type PreisschubDTO struct {
	Geholt int `json:"geholt"`

	// Offen ist die Zahl der Drucke, die noch fehlen — der Fortschritt.
	Offen int `json:"offen"`

	// Fehlgeschlagen ist die Zahl der Versuche, die nichts eingetragen haben.
	//
	// **Sie steht neben `Offen`, damit die keine Auskunft erfinden muss**: ein
	// Fehlschlag ließ `Offen` sinken, obwohl der Druck offen blieb, und bei
	// genau einem vollen Schub voller Fehlschläge meldete der Bericht
	// »fertig« (Nachlauf über Phase 4.5, Befund 7).
	Fehlgeschlagen int `json:"fehlgeschlagen"`

	// Fertig sagt, ob nichts mehr bleibt.
	//
	// **Mitgeschickt und nicht drüben gerechnet**: vier Flächen, die
	// `offen == 0` je selbst prüfen, prüfen es irgendwann verschieden.
	Fertig bool `json:"fertig"`

	// Steht sagt, dass dieser Schub nichts bewegt hat, obwohl Arbeit bleibt.
	//
	// **Der Satz, an dem ein Aufrufer aufhört zu rufen.** Aus demselben Grund
	// mitgeschickt wie `Fertig`.
	Steht bool `json:"steht"`

	Befunde []string `json:"befunde"`
}

// Massennaht führt Vorgänge über eine Auswahl — **ohne Netz**.
type Massennaht struct {
	pflege *service.Massenpflege
	preise *service.Preispflege
}

// MassennahtUeber baut sie über einer Datenbank.
//
// **Die Preise brauchen ein Zuhause**, die Massenvorgänge nicht. Ohne das
// Zuhause bleibt `preise` leer, und der Schub sagt es selbst — die
// Massenvorgänge gehen weiter (§16, Regel 5).
func MassennahtUeber(db *sql.DB) *Massennaht {
	naht := &Massennaht{
		pflege: service.NeueMassenpflege(storage.NeuerSammlungsspeicher(db)),
	}

	// **Über die eine Stelle, an der ein Klient entsteht** (Phase 4.8) — dort
	// wird auch der Dienstschalter geprüft. Ohne ein Zuhause **oder** mit
	// abgeschaltetem Dienst bleibt die Preisnaht leer, und der Schub sagt es
	// beim ersten Versuch.
	klient, err := Scryfallklient(NeueEinstellungswahl(storage.NeueEinstellungen(db)).MitRiegel(AutorisierungAus(db)))
	if err != nil {
		return naht
	}

	naht.preise = service.NeuePreispflege(storage.NeuerPreisspeicher(db),
		storage.NeuerSammlungsspeicher(db), klient, time.Now)

	return naht
}

// Aendern schreibt die genannte Änderung an alle genannten Zeilen.
func (m *Massennaht) Aendern(ctx context.Context, dto MassenaenderungDTO) (
	*MassenberichtDTO, error,
) {
	aenderung := service.Massenaenderung{
		Zeilen:   massenzeilen(dto.Zeilen),
		Menge:    dto.Menge,
		Lagerort: dto.Lagerort,
		Wunsch:   dto.Wunsch,
	}

	if dto.Zustand != nil {
		zustand := domain.Zustand(*dto.Zustand)
		aenderung.Zustand = &zustand
	}

	bericht, err := m.pflege.Aendern(ctx, aenderung)
	if err != nil {
		return nil, err
	}

	return alsMassenbericht(bericht), nil
}

// Loeschen nimmt die genannten Zeilen heraus.
func (m *Massennaht) Loeschen(ctx context.Context, zeilen []MassenzeileDTO) (
	*MassenberichtDTO, error,
) {
	bericht, err := m.pflege.Loeschen(ctx, massenzeilen(zeilen))
	if err != nil {
		return nil, err
	}

	return alsMassenbericht(bericht), nil
}

// PreiseInMasse holt einen Schub Preise.
//
// **`hoechstensStunden` und keine Dauer**: über die Naht wäre eine
// `time.Duration` eine Zahl in Nanosekunden, und niemand auf der anderen Seite
// wüsste, welche Einheit gemeint ist.
func (m *Massennaht) PreiseInMasse(ctx context.Context, hoechstensStunden int) (
	*PreisschubDTO, error,
) {
	if m.preise == nil {
		return nil, keinePreise()
	}

	schub, err := m.preise.PreiseInMasse(ctx, preisalter(hoechstensStunden))
	if err != nil {
		return nil, err
	}

	return &PreisschubDTO{
		Geholt: schub.Geholt, Offen: schub.Offen,
		Fehlgeschlagen: schub.Fehlgeschlagen,
		Fertig:         schub.Fertig(), Steht: schub.Steht(),
		Befunde: befundtexte(schub.Befunde),
	}, nil
}

// massenzeilen bildet die DTOs auf die Zeilen des Dienstes ab.
func massenzeilen(zeilen []MassenzeileDTO) []service.Massenzeile {
	gebaut := make([]service.Massenzeile, 0, len(zeilen))
	for _, zeile := range zeilen {
		gebaut = append(gebaut, service.Massenzeile{
			ID: zeile.ID, Fassung: zeile.Fassung,
		})
	}

	return gebaut
}

// alsMassenbericht bildet den Bericht ab.
func alsMassenbericht(bericht service.Massenbericht) *MassenberichtDTO {
	dto := &MassenberichtDTO{
		Geaendert: bericht.Geaendert, Befunde: befundtexte(bericht.Befunde),
		Vorher: make([]MassenstandDTO, 0, len(bericht.Vorher)),
	}

	for _, stand := range bericht.Vorher {
		dto.Vorher = append(dto.Vorher, MassenstandDTO{
			ID: stand.ID, Fassung: stand.Fassung, Menge: stand.Menge,
			Zustand: string(stand.Zustand), Lagerort: stand.Lagerort, Wunsch: stand.Wunsch,
		})
	}

	return dto
}

// preisalter ist das Höchstalter eines Preises aus einer Zahl von Stunden —
// **sättigend** (Befund 1 des Nachlaufs über J05, #2269): mehr, als eine Dauer
// fasst, heißt „jeder vorhandene Preis ist frisch genug", nicht 25 Minuten.
func preisalter(stunden int) time.Duration {
	return domain.ZeitspanneHoechstens(int64(stunden), time.Hour)
}
