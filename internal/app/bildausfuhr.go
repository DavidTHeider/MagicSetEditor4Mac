package app

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Bildausfuehrer schreibt Karten als Bilder — die Naht zu `internal/service`.
type Bildausfuehrer interface {
	Namen(ctx context.Context, setID int64, vorlage string,
		format service.Bildformat, verhalten string,
		auswahl domain.Kartenauswahl) (
		[]service.Bildname, error,
	)

	Schreibe(ctx context.Context, karteID int64, verzeichnis, datei string,
		format service.Bildformat, zielbreite int, verhalten string,
		guete render.Guete,
	) (*service.Bildschreibung, error)
}

// BildnameDTO ist der Dateiname, unter dem eine Karte herauskommt.
type BildnameDTO struct {
	// KarteID als Zeichenkette — eine `int64` überlebt den Weg durch
	// JavaScript nicht unbeschadet.
	KarteID string `json:"karteID"`
	Nummer  int    `json:"nummer"`
	Datei   string `json:"datei"`
}

// BildnamenAntwort trägt die Namen eines Laufs oder einen Fehler.
type BildnamenAntwort struct {
	Namen  []BildnameDTO `json:"namen"`
	Fehler *Fehler       `json:"fehler"`
}

// BildAntwort ist der Ausgang eines geschriebenen Bilds.
type BildAntwort struct {
	Pfad   string `json:"pfad"`
	Breite int    `json:"breite"`
	Hoehe  int    `json:"hoehe"`

	// Befunde sind die Stellen, an denen dem Bild etwas fehlt (§8).
	Befunde []string `json:"befunde"`

	// Uebersprungen sagt, dass gar nichts geschrieben wurde — `altbehalten`
	// und eine Datei, die schon dalag (#491).
	Uebersprungen bool `json:"uebersprungen"`

	Fehler *Fehler `json:"fehler"`
}

// Bildnamen gibt für jede Karte eines Sets den Dateinamen nach Vorlage.
//
// # Warum die Namen einzeln herauskommen
//
// **Der Stapellauf läuft in der Oberfläche, nicht hier** — und das ist keine
// Bequemlichkeit, sondern die einzige Bauart, die
// #49s zweite
// Zusage einlöst: „schreitet sichtbar fort und lässt sich abbrechen". Ein Lauf
// im Hintergrund müsste dem Fenster sagen können, wo er steht — und diese Naht
// gibt es nicht (#162).
//
// Die **Namen** entstehen trotzdem hier und auf einmal: Eindeutigkeit ist eine
// Aussage über den ganzen Lauf, und wer sie je Aufruf bildete, müsste gegen das
// Dateisystem prüfen statt gegen die Liste.
//
// **`auswahl` leer heißt: das ganze Set** — Befund 2 des Zweitblicks über
// Phase 3.6. MSE2 stellt die Wahl vor der Bildausgabe genauso
// (`SetWindow::onFileExportImages`, `src/gui/set/window.cpp`, Zeilen 668–673,
// Fassung f9d9356 vom 2021-07-06).
func (a *App) Bildnamen(
	ctx context.Context, setID, vorlage, format, verhalten string, auswahl []string,
) BildnamenAntwort {
	ctx = a.imProfil(ctx)

	leer := BildnamenAntwort{Namen: []BildnameDTO{}}

	kennung, err := strconv.ParseInt(setID, 10, 64)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	gewaehlt, auswahlbefund := kartenauswahlAus(auswahl)
	if auswahlbefund != nil {
		leer.Fehler = auswahlbefund

		return leer
	}

	namen, err := a.bildausfuhr.Namen(
		ctx, kennung, vorlage, service.Bildformat(format), verhalten, gewaehlt)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	for _, name := range namen {
		leer.Namen = append(leer.Namen, BildnameDTO{
			KarteID: strconv.FormatInt(name.KarteID, 10),
			Nummer:  name.Nummer,
			Datei:   name.Datei,
		})
	}

	return leer
}

// BildSchreiben legt eine Karte als Bild in ein Verzeichnis.
func (a *App) BildSchreiben(
	ctx context.Context, karteID, verzeichnis, datei, format string,
	zielbreite int, verhalten, guete string,
) BildAntwort {
	ctx = a.imProfil(ctx)

	leer := BildAntwort{Befunde: []string{}}

	gewaehlt, fehler := GueteAusName(guete)
	if fehler != nil {
		leer.Fehler = Uebersetze(ctx, fehler)

		return leer
	}

	kennung, err := strconv.ParseInt(karteID, 10, 64)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	geschrieben, err := a.bildausfuhr.Schreibe(
		ctx, kennung, verzeichnis, datei, service.Bildformat(format), zielbreite,
		verhalten, gewaehlt)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	for _, befund := range geschrieben.Befunde {
		leer.Befunde = append(leer.Befunde, befund.String())
	}

	leer.Pfad, leer.Breite, leer.Hoehe = geschrieben.Pfad, geschrieben.Breite, geschrieben.Hoehe
	leer.Uebersprungen = geschrieben.Uebersprungen

	return leer
}

// MitBildausfuhr setzt die Naht zur Bildausgabe.
func (a *App) MitBildausfuhr(ausfuhr Bildausfuehrer) { a.bildausfuhr = ausfuhr }
