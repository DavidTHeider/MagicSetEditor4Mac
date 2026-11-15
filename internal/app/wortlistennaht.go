// Die Wortlisten für die Flächen ohne Fenster
// (#1074).
//
// # Warum die Auswahl nicht am Fenster hängen bleibt
//
// §16: was das Fenster kann, können Kommandozeile und MCP — oder es steht
// begründet, warum nicht. Bei der **Symbolpalette** steht es: sie gibt
// gerasterte Kacheln, und eine Kachel ist ohne Bildschirm keine Antwort.
//
// **Eine Wortliste ist Text.** Für ein Sprachmodell ist sie sogar die bessere
// Hälfte: sie sagt, welche Untertypen dieses Spiel **kennt**, statt es raten zu
// lassen — und ein geratener Kartentyp ist genau das, was §15 verhindern soll.

package app

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Wortlistennaht ist die Auswahl, wie eine Fläche ohne Fenster sie ruft.
type Wortlistennaht struct {
	karten    *storage.SetLeser
	planer    *service.Planbauer
	schreiber Kartenschreibnaht
}

// WortlistennahtUeber baut sie über einem Bestand.
//
// **Mit dem vollen Zeichenwerk**, wie [SkriptnahtUeber] und aus demselben
// Grund: ein Wort mit `script:` rechnet in der Umgebung, in der auch gezeichnet
// wird. Eine ärmere gäbe eine andere Liste als das Fenster.
//
// **Und mit derselben Schreibnaht wie MCP** (#710): ein zweiter Weg zum
// Bestand liefe beim ersten Sonderfall auseinander — genau das war der Befund,
// der `Schreibnaht` entstehen ließ.
func WortlistennahtUeber(db *sql.DB, urheber domain.Urheber) (*Wortlistennaht, error) {
	// **Beide auf einmal, ein Zweig** — dieselbe Überlegung wie in
	// [Zeichennaht]: beide scheitern an derselben Ursache (dem
	// Paketverzeichnis), und ein zweites `if` darauf wäre ein Zweig, den kein
	// Aufruf betreten kann (§3). Beide Fehler zu melden ist obendrein die
	// bessere Auskunft.
	werk, werkfehler := Zeichennaht(db)

	schreiber, schreibfehler := Schreibnaht(db, urheber)

	if fehler := errors.Join(werkfehler, schreibfehler); fehler != nil {
		return nil, fehler
	}

	return &Wortlistennaht{
		karten:    storage.NeuerSetLeser(db),
		planer:    werk.Planbauer,
		schreiber: schreiber,
	}, nil
}

// Wortlisten gibt, was ein Feld zur Auswahl anbietet.
func (w *Wortlistennaht) Wortlisten(ctx context.Context, karteID int64, feld string) (
	[]WortlistenbereichDTO, error,
) {
	karte, err := w.karten.KarteLesen(ctx, karteID)
	if err != nil {
		return nil, err
	}

	// **Kein eigener Zweig für den Fehler.** Hier ist nichts zu tun, als ihn
	// weiterzugeben — und ein `if` darauf wäre von keinem Aufruf zu betreten:
	// der einzige Fehler des Planbaus ist der Abbruch, und den fängt
	// `KarteLesen` eine Zeile früher (§3).
	bereiche, _, planfehler := w.planer.Kartenwortlisten(ctx, karte, feld)

	return bereicheDTO(bereiche), planfehler
}

// Waehlen schreibt den Wert, der nach einer Wahl dasteht.
//
// **Rechnen und Schreiben in einem Aufruf** (§7, Regel 4) — dieselbe Zusage
// wie im Fenster.
func (w *Wortlistennaht) Waehlen(
	ctx context.Context, setID, karteID int64, feld string, bereich int, weg []int,
) error {
	if befund := domain.PruefeFeldname(feld); befund != nil {
		return befund
	}

	karte, err := w.karten.KarteLesen(ctx, karteID)
	if err != nil {
		return err
	}

	wert, _, wahlfehler := w.planer.Kartenwortwahl(ctx, karte, feld, bereich, weg)
	if wahlfehler != nil {
		return wahlfehler
	}

	_, _, schreibfehler := w.schreiber.FeldSetzenMitVorwert(ctx, setID, karteID, feld, wert)

	return schreibfehler
}
