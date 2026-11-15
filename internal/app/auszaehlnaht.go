// **Auszählen nach einer erklärten Dimension, für alle vier Flächen** —
// Werkstück C von Phase 4.11
// (#1338,
// #1349).
//
// **Ohne Netz.** Eine Dimension rechnet aus dem Spielpaket über den Bestand;
// sie fragt niemanden.

package app

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// AuszaehldimensionDTO ist ein `statistics dimension:`-Block, wie eine Fläche
// ihn zeigt.
//
// **Ohne den Skripttext.** Er ist für niemanden lesbar, der ihn nicht
// ausführt; ein Menü braucht Name und Beschreibung.
type AuszaehldimensionDTO struct {
	Name         string `json:"name"`
	Beschreibung string `json:"beschreibung"`

	// Numerisch sagt, dass die Gruppen eine Achse bilden und keine
	// Aufzählung — die Oberfläche zeichnet sie dann in ihrer Reihenfolge.
	Numerisch bool `json:"numerisch"`
}

// AuszaehlgruppeDTO ist ein Balken: ein Wert und wie viele Karten ihn tragen.
type AuszaehlgruppeDTO struct {
	Name   string `json:"name"`
	Karten int    `json:"karten"`

	// Farbe ist die im Paket erklärte Farbe, in MSEs Schreibweise — leer, wo
	// keine dasteht.
	Farbe string `json:"farbe"`

	// KartenIDs sind die Karten des Balkens, aufsteigend — UXR-F6
	// (#2013).
	//
	// **Eine Liste und nie `null`**: eine Gruppe ohne Karten nennt eine leere,
	// und die Seite filtert damit, ohne zu prüfen, ob es sie gibt. Als
	// Zeichenketten, wie jede Kennung an dieser Naht.
	KartenIDs []string `json:"kartenIDs"`
}

// AuszaehlungDTO ist das Ergebnis einer Auszählung.
type AuszaehlungDTO struct {
	Dimension    string              `json:"dimension"`
	Beschreibung string              `json:"beschreibung"`
	Numerisch    bool                `json:"numerisch"`
	Gruppen      []AuszaehlgruppeDTO `json:"gruppen"`
	Befunde      []string            `json:"befunde"`
}

// Auszaehlnaht zählt ein Set nach einer erklärten Dimension aus.
type Auszaehlnaht struct{ dienst *service.Auszaehldienst }

// AuszaehlungUeber baut sie über einer Datenbank.
//
// **Sie bestimmt die Wurzeln selbst**, wie [Kartenskriptnaht] und aus
// demselben Grund: eine Fläche ohne Fenster kennt kein „zuletzt gewähltes
// Verzeichnis".
func AuszaehlungUeber(db *sql.DB) (*Auszaehlnaht, error) {
	pakete, paketfehler := Paketverzeichnis()

	bilder, bildfehler := Bilderverzeichnis()

	if fehler := errors.Join(paketfehler, bildfehler); fehler != nil {
		return nil, fehler
	}

	wurzeln := paketwurzelnFuer(pakete)

	leser := storage.NeuerSetLeser(db)

	stile := service.NeuerStilspeicher(leser, wurzeln...).
		MitBildablage(service.NeueBildablage(bilder)).MitStilwerten(leser)

	return NeueAuszaehlnaht(service.NeuerAuszaehldienst(
		service.NeueVorlagenauszaehlung(stile, leser))), nil
}

// NeueAuszaehlnaht baut sie über einem fertigen Dienst — die Naht fürs Fenster
// und für Tests.
func NeueAuszaehlnaht(dienst *service.Auszaehldienst) *Auszaehlnaht {
	return &Auszaehlnaht{dienst: dienst}
}

// Angebot nennt die Dimensionen, die das Spiel dieses Sets erklärt.
func (a *Auszaehlnaht) Angebot(ctx context.Context, setID int64) (
	[]AuszaehldimensionDTO, error,
) {
	angaben, fehler := a.dienst.Angebot(ctx, setID)
	if fehler != nil {
		return nil, fehler
	}

	gebaut := make([]AuszaehldimensionDTO, 0, len(angaben))

	for _, angabe := range angaben {
		gebaut = append(gebaut, AuszaehldimensionDTO{
			Name:         angabe.Name,
			Beschreibung: angabe.Beschreibung,
			Numerisch:    angabe.Numerisch,
		})
	}

	return gebaut, nil
}

// Auszaehlen zählt das Set nach der benannten Dimension aus.
func (a *Auszaehlnaht) Auszaehlen(ctx context.Context, setID int64, name string) (
	*AuszaehlungDTO, error,
) {
	auszaehlung, fehler := a.dienst.Auszaehlen(ctx, setID, name)
	if fehler != nil {
		return nil, fehler
	}

	return &AuszaehlungDTO{
		Dimension:    auszaehlung.Dimension,
		Beschreibung: auszaehlung.Beschreibung,
		Numerisch:    auszaehlung.Numerisch,
		Gruppen:      alsGruppen(auszaehlung.Gruppen),
		Befunde:      befundtexte(auszaehlung.Befunde),
	}, nil
}

// alsGruppen übersetzt die Gruppen der Domäne.
func alsGruppen(gruppen []domain.Auszaehlgruppe) []AuszaehlgruppeDTO {
	gebaut := make([]AuszaehlgruppeDTO, 0, len(gruppen))

	for _, gruppe := range gruppen {
		gebaut = append(gebaut, AuszaehlgruppeDTO{
			Name: gruppe.Name, Karten: gruppe.Karten, Farbe: gruppe.Farbe,
			KartenIDs: kennungstexte(gruppe.Kartenkennungen),
		})
	}

	return gebaut
}
