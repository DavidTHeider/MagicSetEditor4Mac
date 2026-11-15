// **Karten aus einem Skript, für alle vier Flächen** — Werkstück B von Phase
// 4.11 (#1337,
// #1349).
//
// **Ohne Netz.** Ein Kartenskript rechnet aus dem Spielpaket und schreibt in
// den Bestand; es fragt niemanden.

package app

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// KartenskriptDTO ist ein `add cards script:`-Block, wie eine Fläche ihn zeigt.
//
// **Ohne den Skripttext.** Er ist bis zu neunzehntausend Zeichen lang und für
// niemanden lesbar, der ihn nicht ausführt; ein Menü braucht Namen und
// Beschreibung.
type KartenskriptDTO struct {
	Name string `json:"name"`

	// Kuerzel ist der Buchstabe, den MSEs `&` markiert — leer, wenn keiner
	// dasteht.
	Kuerzel string `json:"kuerzel"`

	Beschreibung string `json:"beschreibung"`
}

// KartenskriptlaufDTO sagt, was ein Lauf angelegt hat.
type KartenskriptlaufDTO struct {
	// Karten sind die Kennungen der angelegten Karten, in ihrer Reihenfolge.
	Karten []string `json:"karten"`

	// Befunde ist, was unterwegs auffiel — ein Spielskript, das nicht lief,
	// etwa.
	Befunde []string `json:"befunde"`
}

// Kartenskriptnaht führt die Kartenskripte eines Sets.
type Kartenskriptnaht struct{ dienst *service.Kartenskriptdienst }

// KartenskripteUeber baut sie über einer Datenbank.
//
// **Sie bestimmt die Wurzeln selbst**, wie [Zeichennaht] und aus demselben
// Grund: eine Fläche ohne Fenster kennt kein „zuletzt gewähltes Verzeichnis".
func KartenskripteUeber(db *sql.DB) (*Kartenskriptnaht, error) {
	pakete, paketfehler := Paketverzeichnis()

	bilder, bildfehler := Bilderverzeichnis()

	if fehler := errors.Join(paketfehler, bildfehler); fehler != nil {
		return nil, fehler
	}

	wurzeln := paketwurzelnFuer(pakete)

	leser := storage.NeuerSetLeser(db)

	stile := service.NeuerStilspeicher(leser, wurzeln...).
		MitBildablage(service.NeueBildablage(bilder)).MitStilwerten(leser)

	return &Kartenskriptnaht{
		// **Die Karten rechnen ihre Werte** (#2516), wie eine angelegte.
		dienst: service.NeuerKartenskriptdienst(
			storage.NeuerKartenschreiber(db), service.NeueVorlagenauskunft(stile)).
			MitRechnung(stile, nil, nil),
	}, nil
}

// NeueKartenskriptnaht baut sie über einem fertigen Dienst — die Naht fürs
// Fenster und für Tests.
func NeueKartenskriptnaht(dienst *service.Kartenskriptdienst) *Kartenskriptnaht {
	return &Kartenskriptnaht{dienst: dienst}
}

// Angebot nennt die Kartenskripte, die zu diesem Set gehören.
func (k *Kartenskriptnaht) Angebot(ctx context.Context, setID int64) (
	[]KartenskriptDTO, error,
) {
	skripte, fehler := k.dienst.Angebot(ctx, setID)
	if fehler != nil {
		return nil, fehler
	}

	gebaut := make([]KartenskriptDTO, 0, len(skripte))

	for _, skript := range skripte {
		gebaut = append(gebaut, KartenskriptDTO{
			Name:         skript.Name,
			Kuerzel:      skript.Kuerzel,
			Beschreibung: skript.Beschreibung,
		})
	}

	return gebaut, nil
}

// Ausfuehren führt das benannte Skript und hängt seine Karten ans Set.
func (k *Kartenskriptnaht) Ausfuehren(ctx context.Context, setID int64, name string) (
	*KartenskriptlaufDTO, error,
) {
	lauf, fehler := k.dienst.Ausfuehren(ctx, setID, name)
	if fehler != nil {
		return nil, fehler
	}

	kennungen := make([]string, 0, len(lauf.Karten))
	for _, kennung := range lauf.Karten {
		kennungen = append(kennungen, kennungText(kennung))
	}

	return &KartenskriptlaufDTO{
		Karten: kennungen, Befunde: befundtexte(lauf.Befunde),
	}, nil
}
