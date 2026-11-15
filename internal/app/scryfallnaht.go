// Die Naht zum Scryfall-Abgleich für die Flächen ohne Fenster
// (#1005).
//
// # Warum sie hier steht und nicht in jedem Composition Root
//
// `cmd/mse`, `cmd/mse-mcp` und `cmd/mse-rest` kennen `service` nicht (§2) —
// und sollen es nicht. Sie verdrahten `app`-Nähte; was ein Anwendungsfall ist,
// weiß `service`. Ohne diese Datei müsste jeder von ihnen `service.Vorschlag`
// nennen, und drei Kopien der Übersetzung liefen auseinander.
//
// **Dieselbe Gestalt wie `Zeichennaht` und `Setschreibnaht`**: eine Funktion,
// die aus einer `*sql.DB` etwas Benutzbares macht.

package app

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Fremdabgleich ist der Abgleich, wie ihn eine Fläche ohne Fenster benutzt.
//
// **Zwei Schritte** (§16, Regel 6): `Vorschlag` schreibt nichts, `Uebernehmen`
// schreibt die genannten Felder — und gibt ihre vorigen Werte zurück, damit
// die Änderung umkehrbar ist.
type Fremdabgleich struct {
	db       *sql.DB
	abgleich *service.Scryfallabgleich

	// anlage macht aus einem Namen eine Karte (#1032).
	//
	// **An derselben Naht wie der Abgleich**: derselbe Klient, derselbe
	// Zwischenspeicher, dieselbe Drossel. Ein zweiter Klient hätte eine
	// zweite Drossel, und zwei Drosseln unter einer Grenze sind keine (#1024).
	anlage *service.Scryfallanlage
}

// Scryfallabgleichnaht baut den Abgleich über einem Bestand.
func Scryfallabgleichnaht(
	db *sql.DB, urheber domain.Urheber,
) (*Fremdabgleich, error) {
	schreiber, err := Schreibnaht(db, urheber)
	if err != nil {
		return nil, err
	}

	abgleich, nahtfehler := Scryfallnaht(NeueEinstellungswahl(storage.NeueEinstellungen(db)).MitRiegel(AutorisierungAus(db)),
		schreiber, storage.NeueKartenverweise(db))
	if nahtfehler != nil {
		return nil, nahtfehler
	}

	return &Fremdabgleich{
		db: db, abgleich: abgleich,
		anlage: service.NeueScryfallanlage(abgleich, schreiber),
	}, nil
}

// Anlegen macht aus einem Kartennamen eine Karte in diesem Set (#1032).
//
// **Das Spiel kommt aus dem Set**, nicht vom Aufrufer: es steht am Set, und
// eine Fläche, die es mitgeben müsste, könnte es falsch mitgeben.
func (f *Fremdabgleich) Anlegen(ctx context.Context, setID int64, name string) (
	AnlageDTO, error,
) {
	spiel, fehler := Setspiel(ctx, f.db, setID)
	if fehler != nil {
		return AnlageDTO{}, fehler
	}

	ergebnis, anlegefehler := f.anlage.Anlegen(ctx, setID, spiel, name)

	return alsAnlageDTO(ergebnis), anlegefehler
}

// AnlageDTO ist, was eine Fläche ohne Fenster vom Anlegen zurückbekommt.
//
// **Ein eigener Verbund und nicht [service.Anlageergebnis]**: `cmd/mse`,
// `cmd/mse-mcp` und `cmd/mse-rest` kennen `service` nicht (§2). Ohne diesen
// Verbund müsste jeder von ihnen die Schicht überspringen, die ihn vor
// Änderungen an der Fachlichkeit schützt.
type AnlageDTO struct {
	// KarteID ist die angelegte Karte — **auch wenn das Füllen scheitert**.
	KarteID int64 `json:"karteID"`

	// Felder sind die gefüllten Feldnamen.
	Felder []string `json:"felder"`

	// Befunde sind die Auffälligkeiten der Antwort.
	Befunde []string `json:"befunde"`

	Herkunft string `json:"herkunft"`

	// AlterStunden sagt, wie alt der Stand ist (§16, Regel 5).
	AlterStunden float64 `json:"alterStunden"`
}

// alsAnlageDTO übersetzt das Ergebnis des Dienstes.
func alsAnlageDTO(ergebnis service.Anlageergebnis) AnlageDTO {
	return AnlageDTO{
		KarteID: ergebnis.KarteID, Felder: felderAus(ergebnis.Uebernommen),
		Befunde: befundtexte(ergebnis.Befunde), Herkunft: ergebnis.Herkunft,
		AlterStunden: ergebnis.Alter.Hours(),
	}
}

// Vorschlag sagt, was sich ändern würde — und ändert nichts.
func (f *Fremdabgleich) Vorschlag(
	ctx context.Context, setID, karteID int64, name string,
) (ScryfallvorschlagDTO, error) {
	jetzige, spiel, fehler := Kartenstand(ctx, f.db, setID, karteID)
	if fehler != nil {
		return ScryfallvorschlagDTO{}, fehler
	}

	vorschlag, holfehler := f.abgleich.VorschlagFuerKarte(ctx, spiel, jetzige, karteID, name)
	if holfehler != nil {
		return ScryfallvorschlagDTO{}, holfehler
	}

	return alsVorschlagDTO(vorschlag), nil
}

// Uebernehmen schreibt die genannten Felder und gibt ihre vorigen Werte.
//
// # Warum der Vorschlag neu geholt wird
//
// Weil ein Vorschlag, den der Aufrufer zurückschickte, ein Wert wäre, den er
// bestimmt — dann könnte jemand **jeden** Text in ein Feld schreiben und die
// Herkunft „von Scryfall" daranhängen. Der Zwischenspeicher liefert dieselbe
// Antwort ohne Netzgang, solange sie frisch ist.
func (f *Fremdabgleich) Uebernehmen(
	ctx context.Context, setID, karteID int64, name string, felder []string,
) ([]FeldvorschlagDTO, error) {
	jetzige, spiel, fehler := Kartenstand(ctx, f.db, setID, karteID)
	if fehler != nil {
		return nil, fehler
	}

	vorschlag, holfehler := f.abgleich.VorschlagFuerKarte(ctx, spiel, jetzige, karteID, name)
	if holfehler != nil {
		return nil, holfehler
	}

	uebernommen, schreibfehler := f.abgleich.Uebernehmen(ctx, setID, karteID, vorschlag, felder)
	if schreibfehler != nil {
		return nil, schreibfehler
	}

	gestalt := make([]FeldvorschlagDTO, 0, len(uebernommen))
	for _, eine := range uebernommen {
		gestalt = append(gestalt, FeldvorschlagDTO{Feld: eine.Feld, Alt: eine.Alt, Neu: eine.Neu})
	}

	return gestalt, nil
}

// Fremdsatz liest einen ganzen Satz, wie es eine Fläche ohne Fenster tut
// (#1033).
type Fremdsatz struct{ einlesen *service.Satzimport }

// ScryfallsatznahtUeber baut den Satzimport über einem Bestand.
func ScryfallsatznahtUeber(db *sql.DB) (*Fremdsatz, error) {
	leser := storage.NeuerSetLeser(db)

	einlesen, err := Scryfallsatznaht(NeueEinstellungswahl(storage.NeueEinstellungen(db)).MitRiegel(AutorisierungAus(db)),
		storage.NeuerSetSpeicher(db), leser)
	if err != nil {
		return nil, err
	}

	return &Fremdsatz{einlesen: einlesen}, nil
}

// SatzDTO ist, was eine Fläche ohne Fenster vom Einlesen zurückbekommt.
//
// **Ein eigener Verbund und nicht [service.Ergebnis]**: `cmd/mse`,
// `cmd/mse-mcp` und `cmd/mse-rest` kennen `service` nicht (§2).
type SatzDTO struct {
	SetID      int64    `json:"setID"`
	Karten     int      `json:"karten"`
	Befunde    []string `json:"befunde"`
	SchonOffen bool     `json:"schonOffen"`
}

// Einlesen holt den Satz und legt ihn als Set ab.
func (f *Fremdsatz) Einlesen(ctx context.Context, code, spiel string) (SatzDTO, error) {
	ergebnis, fehler := f.einlesen.Einlesen(ctx, code, spiel)

	// **Auch im Fehlerfall mitgeben** (#155): was durchkam, steht im Bestand.
	if ergebnis == nil {
		return SatzDTO{Befunde: []string{}}, fehler
	}

	return SatzDTO{
		SetID: ergebnis.SetID, Karten: ergebnis.Karten,
		Befunde: befundtexte(ergebnis.Befunde), SchonOffen: ergebnis.SchonOffen,
	}, fehler
}
