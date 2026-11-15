// **Die Stilwahl und die Stiloptionen für die Flächen ohne Fenster**
// (#1107).
//
// # Warum es diese Naht gibt
//
// `App.Stile` und ihre fünf Geschwister hängen an einem gestarteten `App` —
// und das gibt es nur mit Fenster. Dieselbe Lage wie bei [SetangabennahtUeber],
// und derselbe Weg hinaus: der Stilwähler wird hier gebaut, wie ihn `app.go`
// und `mcpnaht.go` bauen.
//
// **Und der Wechsel bleibt zweistufig**: erst [Stilweg.Folgen], dann
// [Stilweg.Wechseln]. Ein Wechsel kann Feldwerte ungedeutet lassen und
// Stiloptionen verlieren; im Fenster steht das in einer Rückfrage, und eine
// Fläche ohne Rückfrage muss es wenigstens **fragen** können (§16, Regel 6).

package app

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Stilweg liest und wechselt Stile und ihre Optionen.
type Stilweg struct {
	stile   Stilwaehler
	katalog Katalogauffrischer
}

// Katalogauffrischer ist der Teil des Paketwegs, den die Stilwahl braucht.
//
// **Nur die eine Methode**: die Stilwahl frischt auf und liest sonst nichts
// aus dem Katalog. Eine engere Schnittstelle ist auch die, die ein Test
// belegen kann, ohne einen halben Paketverwalter nachzubauen.
type Katalogauffrischer interface {
	Auffrischen(ctx context.Context) ([]domain.Befund, error)
}

// MitKatalog setzt die Naht zum Paketkatalog — für Tests.
func (w *Stilweg) MitKatalog(katalog Katalogauffrischer) *Stilweg {
	w.katalog = katalog

	return w
}

// MitWaehler setzt die Naht zum Stilwähler — für Tests.
func (w *Stilweg) MitWaehler(waehler Stilwaehler) *Stilweg {
	w.stile = waehler

	return w
}

// StilnahtUeber baut den Weg zur Stilwahl über einer Datenbank.
//
// **Der Katalog wird hier mitgebaut und nicht über [PaketnahtUeber] geholt.**
// Beide scheitern an derselben Bedingung — kein Paketverzeichnis —, und ein
// zweiter `if err != nil` daneben wäre ein Zweig, den kein Aufruf je betritt.
func StilnahtUeber(db *sql.DB) (*Stilweg, error) {
	pakete, err := Paketverzeichnis()
	if err != nil {
		return nil, err
	}

	return &Stilweg{
		stile:   stilwaehlerMit(db, pakete),
		katalog: paketwegMit(db, pakete),
	}, nil
}

// stilwaehlerUeber baut den Stilwähler — **einmal für beide Nähte**.
//
// Er stand am 2026-08-25 in [SetangabennahtUeber] und hier gleichlautend; eine
// zweite Fassung derselben Verdrahtung läuft irgendwann auseinander, und dann
// sieht dieselbe Fähigkeit an zwei Flächen andere Pakete.
func stilwaehlerUeber(db *sql.DB) (Stilwaehler, error) {
	pakete, err := Paketverzeichnis()
	if err != nil {
		return nil, err
	}

	return stilwaehlerMit(db, pakete), nil
}

// stilwaehlerMit baut ihn über einem schon aufgelösten Paketverzeichnis.
//
// **Und seit dem 2026-08-27 ist er die Stelle für alle** (#1168): [Schreibnaht]
// baute ihn gleichlautend daneben, und das Ersetzen brauchte ihn als dritte —
// es liest die **Feldarten** daraus, um nur Textfelder zu durchsuchen. Drei
// Fassungen derselben Verdrahtung liefen auseinander, und dann sähe dieselbe
// Fähigkeit an drei Flächen andere Pakete.
func stilwaehlerMit(db *sql.DB, pakete string) Stilwaehler {
	leser := storage.NeuerSetLeser(db)
	wurzeln := paketwurzelnFuer(pakete)

	return service.NeuerStilwaehler(leser,
		storage.NeuerPaketkatalog(db), storage.NeueStilablage(db), leser, wurzeln...)
}

// Stile nennt die wählbaren Stile eines Sets.
//
// **Mit einem Auffrischen davor**, wie das Fenster es beim Start tut. Ohne
// diesen Zug gäbe die Kommandozeile auf einem frischen Rechner eine leere
// Liste — dieselbe Frage, zwei Antworten (§16). Gefunden am 2026-08-25, als
// `mse stile` gegen die vollständige Saat `{"stile": []}` gab.
//
// **Und ein Befund hält sie nicht auf** (§8): ein unlesbares Paket unter
// vielen soll die übrigen nicht verstecken.
func (w *Stilweg) Stile(ctx context.Context, setID int64) ([]StilDTO, error) {
	if _, err := w.katalog.Auffrischen(ctx); err != nil {
		return nil, err
	}

	angebote, err := w.stile.Stile(ctx, setID)
	if err != nil {
		return nil, err
	}

	stile := make([]StilDTO, 0, len(angebote))
	for _, angebot := range angebote {
		stile = append(stile, StilDTO{
			Paket:        angebot.Paket,
			Spiel:        angebot.Spiel,
			Version:      angebot.Version,
			EigenesSpiel: angebot.EigenesSpiel,
			Gewaehlt:     angebot.Gewaehlt,
		})
	}

	return stile, nil
}

// Folgen sagt, was ein Wechsel kostet — **bevor** er stattfindet.
func (w *Stilweg) Folgen(ctx context.Context, setID int64, paket string) (*FolgenDTO, error) {
	folgen, err := w.stile.Folgen(ctx, setID, paket)
	if err != nil {
		return nil, err
	}

	return &FolgenDTO{
		UngedeuteteFelder: folgen.UngedeuteteFelder,
		VerloreneOptionen: folgen.VerloreneOptionen,
		Spielwechsel:      folgen.Spielwechsel,
	}, nil
}

// Wechseln setzt den Stil eines Sets.
func (w *Stilweg) Wechseln(ctx context.Context, setID int64, paket string) error {
	return w.stile.Wechseln(ctx, setID, paket)
}

// Optionen gibt die Optionen des gewählten Stils samt Werten.
func (w *Stilweg) Optionen(ctx context.Context, setID int64) ([]StiloptionDTO, error) {
	gelesen, err := w.stile.Optionen(ctx, setID)
	if err != nil {
		return nil, err
	}

	optionen := make([]StiloptionDTO, 0, len(gelesen))

	// **Über den Index**: eine Stiloption ist über 128 Byte groß, seit ein
	// Farbfeld seine Wahlfarben mitführt.
	//
	// **Dieselbe Abbildung wie im Fenster** ([alsStiloptionDTO]).
	for i := range gelesen {
		optionen = append(optionen, alsStiloptionDTO(&gelesen[i]))
	}

	return optionen, nil
}

// OptionSetzen schreibt den Wert einer Stiloption.
func (w *Stilweg) OptionSetzen(ctx context.Context, setID int64, feld, wert string) error {
	return w.stile.OptionSetzen(ctx, setID, feld, wert)
}

// OptionZuruecksetzen stellt eine Stiloption auf ihre Vorgabe zurück.
//
// **Ein eigener Weg und kein leerer Wert an [Stilweg.OptionSetzen]**: leer ist
// bei einer Textoption ein gültiger Wert, und zwei Bedeutungen an einem Aufruf
// wären später nicht mehr auseinanderzuhalten.
func (w *Stilweg) OptionZuruecksetzen(ctx context.Context, setID int64, feld string) error {
	return w.stile.OptionZuruecksetzen(ctx, setID, feld)
}
