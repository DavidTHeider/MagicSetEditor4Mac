// **Die Autoersetzungen eines Spiels verwalten** — Werkstück E von Phase 4.11
// (#1341,
// Vermerk [0060](../../docs/entscheidungen/0060-eigene-autoersetzungen-liegen-im-bestand.md)).
//
// **Ohne Netz.** Die Regeln kommen aus dem Spielpaket und dem Bestand.

package app

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// AutoersetzungsregelDTO ist eine Regel, wie eine Fläche sie zeigt.
type AutoersetzungsregelDTO struct {
	Muster string `json:"muster"`
	Ersatz string `json:"ersatz"`

	// GanzesWort sagt, ob nur ein ganzes Wort ersetzt wird.
	GanzesWort bool `json:"ganzesWort"`

	// Aktiv sagt, ob die Regel greift.
	Aktiv bool `json:"aktiv"`

	// Eigen sagt, dass die Regel hier angelegt wurde — nur eine eigene lässt
	// sich löschen, und nur ihr Muster lässt sich ändern.
	Eigen bool `json:"eigen"`
}

// Autoersetzungsweg zeigt und pflegt die Regeln eines Spiels.
type Autoersetzungsweg struct{ pflege *service.Autoersetzungspflege }

// AutoersetzungspflegeUeber baut sie über einem Bestand.
//
// **Über denselben Stilwähler wie das Ersetzen selbst** (#897): er öffnet das
// Spielpaket ohnehin, und ein zweiter Weg dorthin öffnete es ein zweites Mal.
func AutoersetzungspflegeUeber(db *sql.DB) (*Autoersetzungsweg, error) {
	pakete, paketfehler := Paketverzeichnis()

	if paketfehler != nil {
		return nil, paketfehler
	}

	wurzeln := paketwurzelnFuer(pakete)

	leser := storage.NeuerSetLeser(db)

	waehler := service.NeuerStilwaehler(
		leser, storage.NeuerPaketkatalog(db), storage.NeueStilablage(db),
		leser, wurzeln...)

	return NeuerAutoersetzungsweg(service.NeueAutoersetzungspflege(
		storage.NeueAutoersetzungsablage(db),
		service.NeueSpielregeln(leser, waehler))), nil
}

// NeuerAutoersetzungsweg baut ihn über einer fertigen Pflege.
func NeuerAutoersetzungsweg(pflege *service.Autoersetzungspflege) *Autoersetzungsweg {
	return &Autoersetzungsweg{pflege: pflege}
}

// Liste gibt die Regeln und den Schalter.
func (w *Autoersetzungsweg) Liste(ctx context.Context, setID int64) (
	regeln []AutoersetzungsregelDTO, an bool, fehler error,
) {
	liste, err := w.pflege.Liste(ctx, setID)
	if err != nil {
		return nil, false, err
	}

	return alsRegeln(liste.Regeln), liste.An, nil
}

// Setzen legt eine Regel an oder schreibt sie fort.
func (w *Autoersetzungsweg) Setzen(
	ctx context.Context, setID int64, regel AutoersetzungsregelDTO,
) error {
	return w.pflege.Setzen(ctx, setID, domain.Autoersetzungsregel{
		Muster: regel.Muster, Ersatz: regel.Ersatz,
		GanzesWort: regel.GanzesWort, Aktiv: regel.Aktiv,
	})
}

// Loeschen nimmt eine eigene Regel heraus.
func (w *Autoersetzungsweg) Loeschen(
	ctx context.Context, setID int64, muster string,
) error {
	return w.pflege.Loeschen(ctx, setID, muster)
}

// Zuruecksetzen vergisst alle Abweichungen dieses Spiels.
func (w *Autoersetzungsweg) Zuruecksetzen(ctx context.Context, setID int64) error {
	return w.pflege.Zuruecksetzen(ctx, setID)
}

// SchalterSetzen legt den globalen Schalter dieses Spiels um.
func (w *Autoersetzungsweg) SchalterSetzen(
	ctx context.Context, setID int64, an bool,
) error {
	return w.pflege.SchalterSetzen(ctx, setID, an)
}

// alsRegeln übersetzt die Regeln der Domäne.
func alsRegeln(regeln []domain.Autoersetzungsregel) []AutoersetzungsregelDTO {
	gebaut := make([]AutoersetzungsregelDTO, 0, len(regeln))

	for _, regel := range regeln {
		gebaut = append(gebaut, AutoersetzungsregelDTO{
			Muster: regel.Muster, Ersatz: regel.Ersatz,
			GanzesWort: regel.GanzesWort, Aktiv: regel.Aktiv, Eigen: regel.Eigen,
		})
	}

	return gebaut
}

// Autoersetzungen gibt die Regeln, die beim Tippen wirklich greifen.
//
// **Damit erfüllt der Weg auch [Autoersetzungsquelle]** — die Naht, die
// `AutoErsetzen` benutzt. Das ist der Punkt der ganzen Übung: ohne sie
// griffen weiter nur die Regeln des Pakets, und eine eigene wäre eine Zeile
// im Bestand, die nichts tut
// (#1341).
func (w *Autoersetzungsweg) Autoersetzungen(ctx context.Context, setID int64) (
	[]msepaket.AutoErsetzung, error,
) {
	return w.pflege.Wirksame(ctx, setID)
}
