// Die Fehlerpfade der Verdrahtung.
//
// Paketintern, weil sie an der Naht `oeffnen` hängen. Der Grund ist derselbe wie
// in `internal/storage`: eine Datenbank, die sich nicht schließen lässt, kommt
// mit SQLite auf einer echten Platte nicht zustande — im Betrieb aber schon,
// und dann entscheidet der Zweig darüber, ob die Anwendung es meldet.
package app

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"image"
	"io"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

var errNichtSchliessbar = errors.New("diese Verbindung lässt sich planmäßig nicht schließen")

func init() { sql.Register("app-nicht-schliessbar", sperrtreiber{}) }

type sperrtreiber struct{}

func (sperrtreiber) Open(string) (driver.Conn, error) { return sperrconn{}, nil }

type sperrconn struct{}

func (sperrconn) Prepare(string) (driver.Stmt, error) { return sperrstmt{}, nil }
func (sperrconn) Begin() (driver.Tx, error)           { return nil, errNichtSchliessbar }
func (sperrconn) Close() error                        { return errNichtSchliessbar }

type sperrstmt struct{}

func (sperrstmt) Close() error                               { return nil }
func (sperrstmt) NumInput() int                              { return -1 }
func (sperrstmt) Exec([]driver.Value) (driver.Result, error) { return driver.RowsAffected(0), nil }
func (sperrstmt) Query([]driver.Value) (driver.Rows, error)  { return leereZeilen{}, nil }

type leereZeilen struct{}

func (leereZeilen) Columns() []string         { return []string{"name"} }
func (leereZeilen) Close() error              { return nil }
func (leereZeilen) Next([]driver.Value) error { return io.EOF }

func TestStartenMeldetEineNichtOeffenbareDatenbank(t *testing.T) {
	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.oeffnen = func(_ context.Context, _ string) (*sql.DB, error) {
		return nil, errNichtSchliessbar
	}

	if err := anwendung.Starten(t.Context()); err == nil {
		t.Error("eine nicht öffenbare Datenbank wurde angenommen")
	}
}

func TestStartenMeldetEineGescheiterteMigration(t *testing.T) {
	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.oeffnen = func(_ context.Context, _ string) (*sql.DB, error) {
		return sql.Open("app-nicht-schliessbar", "")
	}

	if err := anwendung.Starten(t.Context()); err == nil {
		t.Error("eine gescheiterte Migration wurde angenommen")
	}
}

// Der Bestand fehlt, obwohl die Migration als angewendet verbucht ist.
//
// So sieht eine Datenbank aus, die jemand von Hand angefasst oder aus einer
// fremden Fassung mitgebracht hat: die Buchführung sagt „schon migriert", der
// Bestand ist aber nicht da. Ohne diese Prüfung liefe die Anwendung darauf
// weiter.
func TestStartenMeldetEinenUnbrauchbarenBestand(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "bestand.db")

	db, err := storage.Oeffnen(probekontext(t), pfad)
	if err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}
	for _, anweisung := range []string{
		`CREATE TABLE migrationen (name TEXT PRIMARY KEY, angewendet TEXT NOT NULL) STRICT`,
		`INSERT INTO migrationen VALUES ('0001_grundlage.sql', '2026-08-08T10:00:00.000Z')`,
	} {
		if _, fehler := db.ExecContext(probekontext(t), anweisung); fehler != nil {
			t.Fatalf("vorbereiten: %v", fehler)
		}
	}
	if fehler := db.Close(); fehler != nil {
		t.Fatalf("vorbereiten: %v", fehler)
	}

	if fehler := NeueApp(pfad).Starten(probekontext(t)); fehler == nil {
		t.Error("ein Bestand ohne Tabelle wurde angenommen")
	}
}

func TestBeendenMeldetEineNichtSchliessbareDatenbank(t *testing.T) {
	anwendung := NeueApp("egal")

	db, err := sql.Open("app-nicht-schliessbar", "")
	if err != nil {
		t.Fatalf("Sperrtreiber: %v", err)
	}

	// Erst eine Abfrage, damit überhaupt eine Verbindung im Pool liegt.
	// `sql.DB.Close` schließt nur, was da ist — ohne Verbindung gibt es nichts
	// zu schließen und damit auch keinen Fehler. Der erste Anlauf dieses Tests
	// war genau deshalb grün und prüfte nichts.
	verbindungHerstellen(t, db)

	anwendung.db = db

	if fehler := anwendung.Beenden(); !errors.Is(fehler, errNichtSchliessbar) {
		t.Errorf("Fehler %v, erwartet war der Schließfehler", fehler)
	}
}

// **Und eine nicht schließbare Leserverbindung wird gemeldet, nicht
// verschluckt** — Werkstück 16 von Phase 4.12.
//
// **Paketintern und mit demselben Sperrtreiber**: `sql.DB.Close` einer echten
// SQLite-Verbindung scheitert auf einer gesunden Platte nicht, und der Zweig
// darüber entscheidet im Betrieb, ob jemand davon erfährt.
//
// **Gemeldet und nicht zurückgegeben**: der Aufrufer ist ein Profilwechsel,
// und der darf daran nicht scheitern. Was bleibt, ist die Zeile im Protokoll —
// und dass die zweite Verbindung danach fort ist.
func TestEineNichtSchliessbareLeserverbindungWirdGemeldet(t *testing.T) {
	anwendung := NeueApp("egal")

	db, err := sql.Open("app-nicht-schliessbar", "")
	if err != nil {
		t.Fatalf("Sperrtreiber: %v", err)
	}

	verbindungHerstellen(t, db)

	anwendung.db = db
	anwendung.nurLesen = db

	anwendung.leserverbindungSchliessen()

	if anwendung.nurLesen != nil {
		t.Error("die zweite Verbindung blieb stehen, obwohl sie hergegeben wurde")
	}
}

// verbindungHerstellen erzwingt eine Verbindung im Pool.
//
// `sql.DB.Close` schließt nur, was da ist — ohne Verbindung gibt es nichts zu
// schließen und damit auch keinen Fehler. Der erste Anlauf des Tests darüber
// war genau deshalb grün und prüfte nichts.
func verbindungHerstellen(t *testing.T, db *sql.DB) {
	t.Helper()

	zeilen, err := db.QueryContext(probekontext(t), "SELECT name")
	if err != nil {
		t.Fatalf("Verbindung herstellen: %v", err)
	}
	defer func() {
		if fehler := zeilen.Close(); fehler != nil {
			t.Errorf("Zeilen schließen: %v", fehler)
		}
	}()

	if fehler := zeilen.Err(); fehler != nil {
		t.Fatalf("Zeilen lesen: %v", fehler)
	}
}

// Eine gescheiterte Kodierung wird gemeldet.
//
// Mit einem `bytes.Buffer` kann `png.Encode` nicht scheitern; im Betrieb
// schreibt irgendwann jemand auf etwas, das voll ist. Deshalb die Naht.
func TestEineGescheiterteKodierungWirdGemeldet(t *testing.T) {
	anwendung := NeueApp("")
	anwendung.MitEditor(
		karteschreiberProbe{}, karteleserProbe{}, planbauerProbe{}, rastererProbe{},
	)
	anwendung.packen = func(io.Writer, image.Image) error { return errKodierung }

	if anwendung.Vorschau(t.Context(), "1", 1).Fehler == nil {
		t.Error("die gescheiterte Kodierung ging verloren")
	}
}

// **An der Palette wird sie nicht gemeldet, sondern übergangen.**
//
// Und das ist kein Widerspruch zum Fall darüber: eine Vorschau **ist** ihr
// Bild, eine Palette ist sechzig Schaltflächen. Ein Symbol, das sich nicht
// packen lässt, ist ein Eintrag ohne Kachel — genau das, was MSE zeigt, wenn
// es keine findet (`InsertSymbolMenu::makeMenuItem`, `src/data/symbol_font.cpp`,
// Zeilen 508–518, Fassung f9d9356 vom 2021-07-06).
func TestEineGescheiterteKodierungLaesstDiePaletteStehen(t *testing.T) {
	anwendung := NeueApp("")
	anwendung.MitEditor(
		karteschreiberProbe{}, karteleserProbe{}, mitPalettenplan{}, rastererProbe{},
	)
	anwendung.packen = func(io.Writer, image.Image) error { return errKodierung }

	antwort := anwendung.Symbolpalette(t.Context(), "1", "kosten")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Eintraege) != 1 || antwort.Eintraege[0].Bild != "" {
		t.Errorf("%+v", antwort.Eintraege)
	}
}

// mitPalettenplan gibt eine Palette mit genau einem zeichenbaren Symbol.
type mitPalettenplan struct{ planbauerProbe }

func (mitPalettenplan) Kartenpalette(context.Context, *domain.CardRow, string) (
	[]service.Paletteneintrag, bool, []domain.Befund, error,
) {
	return []service.Paletteneintrag{{
		Art: "code", Code: "W", Beschriftung: "W",
		Plan: &render.Plan{Breite: 2, Hoehe: 2},
	}}, true, nil, nil
}

var errKodierung = errors.New("planmäßig gescheitert")

type karteschreiberProbe struct{}

func (karteschreiberProbe) FeldSetzenMitVorwert(
	context.Context, int64, int64, string, string,
) (vorher string, hatte bool, fehler error) {
	return "", false, nil
}

// FeldSetzenVonAussen schreibt wie oben und merkt sich die Herkunft (#888).
func (karteschreiberProbe) FelderVonAussenMitHerkunft(
	ctx context.Context, setID, karteID int64, werte []domain.Value,
	_ string, _ []domain.KIFeldherkunft,
	_ map[string][]domain.KIBildparameter, _ string,
) (vorher string, hatte bool, fehler error) {
	return (karteschreiberProbe{}).FelderVonAussen(ctx, setID, karteID, werte, "")
}

func (karteschreiberProbe) FelderVonAussen(
	ctx context.Context, setID, karteID int64, werte []domain.Value, _ string,
) (vorher string, hatte bool, fehler error) {
	for _, wert := range werte {
		if _, _, fehler := (karteschreiberProbe{}).FeldSetzenMitVorwert(
			ctx, setID, karteID, wert.Field, wert.Text); fehler != nil {
			return "", false, fehler
		}
	}

	return "", false, nil
}

func (karteschreiberProbe) FeldSetzenVonAussen(
	ctx context.Context, setID, karteID int64, feld, wert, herkunft string,
) (vorher string, hatte bool, fehler error) {
	_ = herkunft

	return karteschreiberProbe{}.FeldSetzenMitVorwert(ctx, setID, karteID, feld, wert)
}

type karteleserProbe struct{}

func (karteleserProbe) KarteLesen(context.Context, int64) (*domain.CardRow, error) {
	return &domain.CardRow{ID: 1}, nil
}

func (karteleserProbe) Herkuenfte(context.Context, int64) (map[string]string, error) {
	return nil, nil
}

type planbauerProbe struct{}

func (planbauerProbe) Kartenplan(context.Context, *domain.CardRow) (
	*render.Plan, []domain.Befund, error,
) {
	return &render.Plan{Breite: 2, Hoehe: 2}, nil, nil
}

func (planbauerProbe) Kartenpalette(context.Context, *domain.CardRow, string) (
	[]service.Paletteneintrag, bool, []domain.Befund, error,
) {
	return nil, false, nil, nil
}

func (planbauerProbe) Kartenwortlisten(context.Context, *domain.CardRow, string) (
	[]service.Wortlistenbereich, []domain.Befund, error,
) {
	return nil, nil, nil
}

// **Ohne Wahlbilder** — die Stiloptionen prüft `stilwahlbilder_test.go`.
func (planbauerProbe) Optionsbilder(context.Context, int64, string) (
	[]service.Optionsbild, []domain.Befund, error,
) {
	return nil, nil, nil
}

func (planbauerProbe) Kartenwortwahl(
	context.Context, *domain.CardRow, string, int, []int,
) (string, []domain.Befund, error) {
	return "", nil, nil
}

type rastererProbe struct{}

func (rastererProbe) Zeichne(context.Context, *render.Plan, float64, render.Guete) (
	*image.NRGBA, []render.Befund, error,
) {
	return image.NewNRGBA(image.Rect(0, 0, 2, 2)), nil, nil
}

func (rastererProbe) SymbolAlsBild(render.Symbol, render.Symbolmasse) (
	bild *image.NRGBA, befund string,
) {
	return image.NewNRGBA(image.Rect(0, 0, 2, 2)), ""
}

func (rastererProbe) BildAusBaum(render.Bildbaum, int, int) (
	bild *image.NRGBA, befund string,
) {
	return image.NewNRGBA(image.Rect(0, 0, 2, 2)), ""
}

// GruppeVonAussenMitHerkunft gibt es an diesem Doppel **nur der
// Schnittstelle wegen** (A13.1b, #1452).
//
// **Sie scheitert laut.** Ein Doppel, das hier still gutginge, ließe einen
// Lauf über Gruppen grün werden, der gar keine geschrieben hat — §3: eine
// Vorrichtung, die großzügiger ist als der Betrieb.
func (karteschreiberProbe) GruppeVonAussenMitHerkunft(
	context.Context, []domain.Kartenschritt,
) error {
	return errors.New("dieses Doppel schreibt keine Gruppen")
}
