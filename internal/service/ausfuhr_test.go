package service_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// leserDouble ist ein In-Memory-Double des lesenden Bestands (§3).
type leserDouble struct {
	koepfe []domain.SetHead
	karten []*domain.CardRow
	werte  []domain.Value
	fehler error

	// kartenfehler scheitert **nur** beim Lesen der Karten. Ein Double, das
	// überall gleich scheitert, kann den zweiten Fehlerweg nicht zeigen.
	kartenfehler error

	// wertefehler scheitert nur beim Lesen der Setangaben — derselbe Grund.
	wertefehler error

	// woerter sind die Schlüsselwörter des Sets (#1062), und
	// woertfehler scheitert **nur** beim Lesen von ihnen.
	woerter     []*domain.Keyword
	woertfehler error

	// format sind die fünf Formatangaben aus dem Dateikopf (F02), und
	// formatfehler scheitert **nur** beim Lesen von ihnen — derselbe Grund wie
	// bei `wertefehler`: sie werden nacheinander gelesen.
	format       domain.Setkopf
	formatfehler error

	// stilwerte sind die setweiten Stiloptionen (F05, #2247), und
	// stilwertefehler scheitert **nur** beim Lesen von ihnen.
	stilwerte       []domain.Value
	stilwertefehler error

	// zeilenfehler scheitert **nur** an der Bestandszeile — die liest der
	// Schreiber für den Stilblock eigens (F05).
	zeilenfehler error
}

// Stilwerte gibt die setweiten Stiloptionen (F05, #2247).
func (l leserDouble) Stilwerte(context.Context, int64) ([]domain.Value, error) {
	return l.stilwerte, l.stilwertefehler
}

// Schluesselwoerter gibt die Schlüsselwörter des Sets (#1062).
func (l leserDouble) Schluesselwoerter(context.Context, int64) ([]*domain.Keyword, error) {
	return l.woerter, l.woertfehler
}

func (l leserDouble) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return l.koepfe, l.fehler
}

// Setzeile gibt die Bestandszeile eines Sets (F02,
// #1746).
//
// **Aus derselben Liste wie [leserDouble.Setkoepfe]** und nicht aus einem
// zweiten Feld daneben: ein Double, in dem dasselbe Set zwei Namen haben kann,
// prüft am Ende das Double.
func (l leserDouble) Setzeile(_ context.Context, setID int64) (domain.SetHead, error) {
	if l.fehler != nil {
		return domain.SetHead{}, l.fehler
	}

	if l.zeilenfehler != nil {
		return domain.SetHead{}, l.zeilenfehler
	}

	// **Ein Double ohne Köpfe kennt jedes Set, ohne Spiel und Stil** — der
	// Schreiber liest die Zeile seit F05 für jeden Kopf; ein Lauf, der sich
	// für ihn nicht interessiert, soll ihn nicht anlegen müssen. **Eine leere
	// Liste dagegen kennt keines** (`setneubau_test.go`).
	if l.koepfe == nil {
		return domain.SetHead{ID: setID}, nil
	}

	for _, kopf := range l.koepfe {
		if kopf.ID == setID {
			return kopf, nil
		}
	}

	return domain.SetHead{}, fmt.Errorf("%w: das Set %d gibt es nicht",
		domain.ErrNotFound, setID)
}

// Setkopf gibt die fünf Formatangaben — leer, wo nichts gesetzt ist.
func (l leserDouble) Setkopf(context.Context, int64) (domain.Setkopf, error) {
	if l.formatfehler != nil {
		return domain.Setkopf{}, l.formatfehler
	}

	return l.format, l.fehler
}

func (l leserDouble) Setwerte(context.Context, int64) ([]domain.Value, error) {
	if l.wertefehler != nil {
		return nil, l.wertefehler
	}

	return l.werte, l.fehler
}

func (l leserDouble) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	if l.kartenfehler != nil {
		return nil, l.kartenfehler
	}

	return l.karten, l.fehler
}

func probenbestand() leserDouble {
	return leserDouble{
		koepfe: []domain.SetHead{{ID: 7, Name: "Probeset", Game: "magic", Cards: 1}},
		karten: []*domain.CardRow{{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "name", Text: "Flammenball"},
			{Field: "casting cost", Text: "2R"},
			{Field: "rarity", Text: "mythic rare"},
		}}},
	}
}

// Ein Set aus dem Bestand geht als MTG JSON hinaus.
func TestSetGehtAlsMTGJSONHinaus(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "PRB.json")

	ausfuhr, err := service.NeuerExporteur(probenbestand()).
		MTGJSON(t.Context(), 7, "magic", ort)
	if err != nil {
		t.Fatalf("MTGJSON: %v", err)
	}

	if ausfuhr.Karten != 1 {
		t.Errorf("Karten: %d, erwartet 1", ausfuhr.Karten)
	}

	inhalt, err := os.ReadFile(ort) // #nosec G304 -- der Pfad stammt aus t.TempDir()
	if err != nil {
		t.Fatalf("die Datei fehlt: %v", err)
	}

	// **Gültig, nicht nur vorhanden**: derselbe Prüfer, den auch eine bezogene
	// Datei besteht.
	if befunde := mtgjson.Pruefe(inhalt, ort); len(befunde) > 0 {
		t.Errorf("die Ausgabe ist nicht gültig: %v", befunde)
	}

	if !strings.Contains(string(inhalt), `"mythic"`) {
		t.Errorf("die Seltenheit ist nicht zurückgewandelt worden:\n%s", inhalt)
	}
}

// Ein Set, das es nicht gibt, wird benannt statt leer geschrieben.
func TestEinUnbekanntesSetWirdBenannt(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "PRB.json")

	_, err := service.NeuerExporteur(probenbestand()).MTGJSON(t.Context(), 99, "magic", ort)
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler: %v — erwartet war »nicht gefunden«", err)
	}

	if _, err := os.Stat(ort); err == nil {
		t.Error("es ist trotzdem eine Datei entstanden")
	}
}

// Ein Spiel ohne Zuordnung wird abgewiesen, nicht geraten.
func TestAusfuhrOhneZuordnungWirdAbgewiesen(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "PRB.json")

	if _, err := service.NeuerExporteur(probenbestand()).
		MTGJSON(t.Context(), 7, "gibtsnicht", ort); err == nil {
		t.Fatal("kein Fehler")
	}
}

// Ein Fehler des Bestands hält die Ausfuhr an.
func TestEinFehlerDesBestandsHaeltDieAusfuhrAn(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "PRB.json")

	bestand := probenbestand()
	bestand.fehler = errors.New("nicht lesbar")

	if _, err := service.NeuerExporteur(bestand).MTGJSON(t.Context(), 7, "magic", ort); err == nil {
		t.Fatal("kein Fehler")
	}
}

// Ein Fehler beim Lesen der Karten hält die Ausfuhr an.
//
// **Ein eigener Weg, nicht derselbe wie beim Setkopf**: wer nur den ersten
// prüft, hat den zweiten nie ausgeführt.
func TestEinFehlerBeimKartenlesenHaeltDieAusfuhrAn(t *testing.T) {
	bestand := probenbestand()
	bestand.kartenfehler = errors.New("die Karten sind nicht lesbar")

	_, err := service.NeuerExporteur(bestand).
		MTGJSON(t.Context(), 7, "magic", filepath.Join(t.TempDir(), "PRB.json"))
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(err.Error(), "die Karten sind nicht lesbar") {
		t.Errorf("Fehler: %q — der Grund fehlt", err)
	}
}

// Ein Pfad, an den sich nicht schreiben lässt, wird gemeldet.
func TestEinUnschreibbarerPfadWirdGemeldet(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "gibtsnicht", "PRB.json")

	if _, err := service.NeuerExporteur(probenbestand()).
		MTGJSON(t.Context(), 7, "magic", ort); err == nil {
		t.Fatal("kein Fehler")
	}
}

// **Was beim Schreiben verlorengeht, steht am Ergebnis** (§8, #142).
func TestDieVerlusteStehenAmErgebnis(t *testing.T) {
	bestand := probenbestand()
	bestand.karten = []*domain.CardRow{{ID: 1, Values: []domain.Value{
		{Field: "name", Text: "Ebene"},
		{Field: "super type", Text: "Basic Land"},
	}}}

	ausfuhr, err := service.NeuerExporteur(bestand).
		MTGJSON(t.Context(), 7, "magic", filepath.Join(t.TempDir(), "PRB.json"))
	if err != nil {
		t.Fatalf("MTGJSON: %v", err)
	}

	if len(ausfuhr.Befunde) == 0 {
		t.Error("kein Befund — der Verlust wird stillschweigend hingenommen")
	}
}

// **Der Rundlauf über den Bestand verliert den Setcode nicht.**
//
// # Woher diese Prüfung kommt
//
// Aus dem Zweitblick über Phase 9. Der Import legt `code` als Setwert ab; die
// Ausfuhr las nur Setköpfe und Karten — und der Schreiber griff mangels Code
// auf den **Namen** zurück. Aus `PRB` wurde `Probeset`.
//
// Gesehen hat das kein Test: die Rundläufe der Fachbibliothek gehen an SQLite
// vorbei, und die Prüfungen der Ausfuhr benutzten ein Set, dessen Code und Name
// gleich waren. Deshalb sind sie hier **verschieden**.
func TestDieAusfuhrBehaeltDenSetcode(t *testing.T) {
	bestand := probenbestand()
	bestand.werte = []domain.Value{{Field: "code", Text: "PRB"}}

	ort := filepath.Join(t.TempDir(), "PRB.json")

	if _, err := service.NeuerExporteur(bestand).
		MTGJSON(t.Context(), 7, "magic", ort); err != nil {
		t.Fatalf("MTGJSON: %v", err)
	}

	inhalt, err := os.ReadFile(ort) // #nosec G304 -- der Pfad stammt aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if !strings.Contains(string(inhalt), `"code": "PRB"`) {
		t.Errorf("der Setcode fehlt in der Ausgabe:\n%s", inhalt)
	}
}

// Ein Fehler beim Lesen der Setangaben hält die Ausfuhr an.
//
// **Ein eigener Weg**: der Setkopf kommt durch, die Angaben nicht. Wer nur den
// ersten prüft, hat den zweiten nie ausgeführt.
func TestEinFehlerBeimLesenDerSetangabenHaeltAn(t *testing.T) {
	bestand := probenbestand()
	bestand.wertefehler = errors.New("die Angaben sind nicht lesbar")

	_, err := service.NeuerExporteur(bestand).
		MTGJSON(t.Context(), 7, "magic", filepath.Join(t.TempDir(), "PRB.json"))
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(err.Error(), "die Angaben sind nicht lesbar") {
		t.Errorf("Fehler: %q — der Grund fehlt", err)
	}
}
