// Die Schlüsselwortnaht — Lesen und Schreiben an einer Stelle
// (#1062).
//
// **Eine eigene Datei, weil sie `internal/storage` kennt.** Das ist den
// Composition Roots vorbehalten (§2), und der Schichtenwächter zählt nach —
// dieselbe Aufteilung wie bei `skriptnaht.go` und `wortlistennaht.go`.

package app

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// SchluesselwortnahtUeber baut die Naht über einem Bestand.
//
// **Sie steht hier und nicht in `cmd/`**: was zusammengesetzt wird, gehört an
// eine Stelle — dieselbe Überlegung wie bei [Schreibnaht] (#710).
func SchluesselwortnahtUeber(db *sql.DB) Schluesselwortablage {
	return SchluesselwortnahtAls(db, domain.UrheberFenster)
}

// SchluesselwortnahtAls baut sie unter einem anderen Urheber.
//
// **Weil die Änderung jetzt im Verlauf steht** (#1077). Bis dahin schrieb die
// Pflege an `history` vorbei, und wer sie ausführte, war belanglos; eine
// Rücknahme, die nicht sagt, wessen Änderung sie umkehrt, ist im
// Mehr-Schreiber-Betrieb eine Vermutung
// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)).
func SchluesselwortnahtAls(db *sql.DB, wer domain.Urheber) Schluesselwortablage {
	return schluesselwortnaht{
		leser:  storage.NeuerSetLeser(db),
		ablage: storage.NeueSchluesselwortablage(db).Als(wer),
	}
}

// Schluesselwortablage ist, was die Pflege braucht — gleichlautend mit
// `service.Schluesselwortablage`.
type Schluesselwortablage interface {
	Schluesselwortleser
	Schluesselwortschreiber
}

// schluesselwortnaht führt Lesen und Schreiben zusammen.
//
// **Zwei Quellen, eine Schnittstelle**: gelesen wird über denselben Setleser
// wie alles andere, geschrieben über die eigene Ablage. `service` soll nicht
// wissen, dass es zwei sind.
type schluesselwortnaht struct {
	leser  Schluesselwortleser
	ablage Schluesselwortschreiber
}

// Schluesselwortleser gibt die Schlüsselwörter eines Sets.
type Schluesselwortleser interface {
	Schluesselwoerter(ctx context.Context, setID int64) ([]*domain.Keyword, error)
}

// Schluesselwortschreiber pflegt sie.
type Schluesselwortschreiber interface {
	SchluesselwortAnlegen(ctx context.Context, setID int64, modus string) (int64, error)
	SchluesselwortSetzen(ctx context.Context, setID, wortID int64, feld, wert string) error
	SchluesselwortLoeschen(ctx context.Context, setID, wortID int64) error
}

func (n schluesselwortnaht) Schluesselwoerter(ctx context.Context, setID int64) (
	[]*domain.Keyword, error,
) {
	return n.leser.Schluesselwoerter(ctx, setID)
}

func (n schluesselwortnaht) SchluesselwortAnlegen(
	ctx context.Context, setID int64, modus string,
) (int64, error) {
	return n.ablage.SchluesselwortAnlegen(ctx, setID, modus)
}

func (n schluesselwortnaht) SchluesselwortSetzen(
	ctx context.Context, setID, wortID int64, feld, wert string,
) error {
	return n.ablage.SchluesselwortSetzen(ctx, setID, wortID, feld, wert)
}

func (n schluesselwortnaht) SchluesselwortLoeschen(
	ctx context.Context, setID, wortID int64,
) error {
	return n.ablage.SchluesselwortLoeschen(ctx, setID, wortID)
}

// Schluesselwortweg ist die Pflege, wie eine Fläche ohne Fenster sie ruft.
//
// **Ein eigener Typ und nicht [service.Schluesselwortpflege]**: `cmd/mse`,
// `cmd/mse-mcp` und `cmd/mse-rest` kennen `service` nicht (§2), und der
// Schichtenwächter zählt nach.
type Schluesselwortweg struct{ pflege *service.Schluesselwortpflege }

// SchluesselwortpflegeUeber baut die volle Pflege über einem Bestand.
//
// **Mit dem Zeichenwerk**, denn die Schlüsselwörter des **Spiels** kommen aus
// dem Spielpaket — und welches das ist, sagt der Stil des Sets. Ohne es sähe
// eine Fläche ohne Fenster nur die eigenen, und dieselbe Frage bekäme zwei
// Antworten (§16).
func SchluesselwortpflegeUeber(db *sql.DB) (*Schluesselwortweg, error) {
	return SchluesselwortpflegeAls(db, domain.UrheberFenster)
}

// SchluesselwortpflegeAls baut sie unter einem anderen Urheber.
//
// **Jede Fläche nennt sich selbst** (#1077): `mse` schreibt als
// `kommandozeile`, der Dienst als `dienst`, der MCP-Server als `server`.
func SchluesselwortpflegeAls(db *sql.DB, wer domain.Urheber) (*Schluesselwortweg, error) {
	werk, err := Zeichennaht(db)
	if err != nil {
		return nil, err
	}

	return &Schluesselwortweg{pflege: service.NeueSchluesselwortpflege(
		SchluesselwortnahtAls(db, wer), werk.Stile)}, nil
}

// Liste gibt die Schlüsselwörter des Sets und die seines Spiels.
func (w *Schluesselwortweg) Liste(ctx context.Context, setID int64) (
	SchluesselwortlisteDTO, error,
) {
	liste, err := w.pflege.Liste(ctx, setID)
	if err != nil {
		return SchluesselwortlisteDTO{}, err
	}

	return SchluesselwortlisteDTO{
		Woerter: schluesselwortDTO(liste.Woerter), Modi: liste.Modi,
	}, nil
}

// Anlegen hängt ein leeres ans Ende und gibt seine Kennung.
func (w *Schluesselwortweg) Anlegen(ctx context.Context, setID int64) (int64, error) {
	return w.pflege.Anlegen(ctx, setID)
}

// Setzen schreibt ein Feld.
func (w *Schluesselwortweg) Setzen(
	ctx context.Context, setID, wortID int64, feld, wert string,
) error {
	return w.pflege.Setzen(ctx, setID, wortID, feld, wert)
}

// Loeschen nimmt eines weg.
func (w *Schluesselwortweg) Loeschen(ctx context.Context, setID, wortID int64) error {
	return w.pflege.Loeschen(ctx, setID, wortID)
}

// SchluesselwortlisteDTO ist die Liste für eine Fläche ohne Fenster.
type SchluesselwortlisteDTO struct {
	Woerter []SchluesselwortDTO `json:"woerter"`
	Modi    []string            `json:"modi"`
}

// Parameterhilfen gibt beide Menüs zu einem Schlüsselwort des Sets
// (#1342).
func (w *Schluesselwortweg) Parameterhilfen(ctx context.Context, setID, wortID int64) (
	service.Parameterhilfen, error,
) {
	return w.pflege.Parameterhilfen(ctx, setID, wortID)
}
