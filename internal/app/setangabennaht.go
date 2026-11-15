// **Die Setangaben, die Vorbehalte und die mitgelieferten Sets für die
// Flächen ohne Fenster**
// (#1107).
//
// # Warum es diese Naht gibt
//
// `App.Setangaben` und `App.Vorbehalte` hängen an einem gestarteten `App` —
// und das gibt es nur mit Fenster. Kommandozeile, Dienst und MCP-Server
// bauen ihre Wege selbst, so wie sie es für Schlüsselwörter, Packs und
// Skripte schon tun (`SchluesselwortnahtUeber`, `PackUeber`,
// `SkriptnahtUeber`).
//
// **Und über denselben Fachweg wie das Fenster.** Wer hier
// `storage.NeuerSetSpeicher` unmittelbar nähme, umginge den Verlauf — und eine
// Rücknahme wüsste von der Änderung nichts.

package app

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Setangabenweg liest und schreibt die Angaben eines Sets.
type Setangabenweg struct {
	stile     Stilwaehler
	werte     Setwertquelle
	rohbloeck Vorbehaltsleser

	// sets sagt, ob es ein Set gibt
	// (#1336).
	sets service.Setkopfleser
}

// Vorbehaltsleser gibt die Blöcke, die beim Einfahren nicht gedeutet wurden.
//
// **Als Schnittstelle und nicht als `*storage.SetLeser`** (§2): `storage` darf
// nur der Composition Root kennen, und ein Wächter zählt es nach.
type Vorbehaltsleser interface {
	Rohbloecke(ctx context.Context, setID int64) ([]domain.RawBlock, error)
}

// SetangabennahtUeber baut den Weg zu den Setangaben über einer Datenbank.
func SetangabennahtUeber(db *sql.DB, urheber domain.Urheber) (*Setangabenweg, error) {
	stile, err := stilwaehlerUeber(db)
	if err != nil {
		return nil, err
	}

	leser := storage.NeuerSetLeser(db)

	return &Setangabenweg{
		stile: stile,
		werte: setwertnaht{
			leser:     leser,
			schreiber: storage.NeuerKartenschreiber(db).Als(urheber),
		},
		rohbloeck: leser,
		sets:      leser,
	}, nil
}

// Angaben gibt die Angabenfelder eines Sets samt ihren Werten.
func (w *Setangabenweg) Angaben(
	ctx context.Context, setID int64,
) (felder []FeldbeschreibungDTO, werte []WertDTO, fehler error) {
	beschreibungen, err := w.stile.Setfelder(ctx, setID)
	if err != nil {
		return nil, nil, err
	}

	belegt, err := w.werte.Setwerte(ctx, setID)
	if err != nil {
		return nil, nil, err
	}

	// **In der gemeldeten Sprache der Anwendung.** Ohne Fenster gibt es keine
	// Spracheinstellung; genommen wird die Vorgabe, und das ist dieselbe, die
	// ein frisch gestartetes Fenster hätte.
	return beschreibungenAus(beschreibungen, ""), setwerteAus(belegt), nil
}

// Setzen schreibt eine Angabe und nennt die vorige.
func (w *Setangabenweg) Setzen(
	ctx context.Context, setID int64, feld, wert string,
) (vorher string, hatte bool, fehler error) {
	return w.werte.SetwertSetzenMitVorwert(ctx, setID, feld, wert)
}

// Vorbehalte gibt die Blöcke, die beim Einfahren nicht gedeutet wurden.
//
// **Ein Set, das es nicht gibt, ist ein Fehler und kein „alles gedeutet"**
// (#1336). Ohne diese Frage bekämen beide Sachverhalte dieselbe Antwort.
func (w *Setangabenweg) Vorbehalte(
	ctx context.Context, setID int64,
) ([]VorbehaltDTO, error) {
	if fehler := service.SetDaSein(ctx, w.sets, setID); fehler != nil {
		return nil, fehler
	}

	bloecke, err := w.rohbloeck.Rohbloecke(ctx, setID)
	if err != nil {
		return nil, err
	}

	vorbehalte := make([]VorbehaltDTO, 0, len(bloecke))
	for _, block := range bloecke {
		vorbehalte = append(vorbehalte, VorbehaltDTO{
			Datei: block.File, Zeile: block.Line, Inhalt: block.Text,
		})
	}

	return vorbehalte, nil
}

// Beispielsetliste nennt die mitgelieferten Sets — **ohne Bestand**.
//
// **Eine freie Funktion und keine Methode**: sie liest ein Verzeichnis im
// Bündel und fragt die Datenbank nichts. Wer sie an einen Bestand bände,
// zwänge jede Fläche, einen zu öffnen, um ein Verzeichnis zu lesen.
// **Und ohne Vorgangskennung** — die einzige Stelle, an der ein Fehler dieser
// Anwendung ohne eine hinausgeht (Werkstück O9, #1433). Der Grund ist derselbe
// wie eine Zeile darüber: die drei Nähte, die hierher rufen
// (`internal/mcp.Setangabenquelle`, `internal/rest`, `beispielsetsbefehl`),
// führen absichtlich keinen Kontext, und an der Kommandozeile gibt es keinen
// Vorgang — der Befehl öffnet keinen Bestand. Eine Kennung wäre ein Griff ins
// Leere; was schiefgehen kann, ist ein unlesbares Verzeichnis, und das steht
// im Protokoll unter seinem eigenen Namen.
func Beispielsetliste() ([]BeispielsetDTO, *Fehler) {
	stumm := &App{beispielsetort: Beispielsetverzeichnis, saatort: Saatverzeichnis}

	antwort := stumm.Beispielsets(context.Background())

	return antwort.Sets, antwort.Fehler
}
