package app_test

// **Die Autoersetzungen an der Bindungsfassade** — Werkstück E von Phase 4.11
// (#1341).

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// pflegequelle ist ein Doppel für `app.Autoersetzungspflegequelle`.
type pflegequelle struct {
	regeln []app.AutoersetzungsregelDTO
	an     bool
	fehler error

	setID      int64
	gesetzt    app.AutoersetzungsregelDTO
	geloescht  string
	zurueck    bool
	geschaltet *bool
}

func (p *pflegequelle) Liste(_ context.Context, setID int64) (
	[]app.AutoersetzungsregelDTO, bool, error,
) {
	p.setID = setID

	return p.regeln, p.an, p.fehler
}

func (p *pflegequelle) Setzen(
	_ context.Context, setID int64, regel app.AutoersetzungsregelDTO,
) error {
	p.setID, p.gesetzt = setID, regel

	return p.fehler
}

func (p *pflegequelle) Loeschen(_ context.Context, setID int64, muster string) error {
	p.setID, p.geloescht = setID, muster

	return p.fehler
}

func (p *pflegequelle) Zuruecksetzen(_ context.Context, setID int64) error {
	p.setID, p.zurueck = setID, true

	return p.fehler
}

func (p *pflegequelle) SchalterSetzen(_ context.Context, setID int64, an bool) error {
	p.setID, p.geschaltet = setID, &an

	return p.fehler
}

// **Die Regeln und der Schalter kommen durch.**
func TestDieAutoersetzungenKommenAnDieOberflaeche(t *testing.T) {
	t.Parallel()

	quelle := &pflegequelle{
		regeln: []app.AutoersetzungsregelDTO{{Muster: "(W)", Eigen: true}},
		an:     true,
	}

	anwendung := app.NeueApp("")
	anwendung.MitAutoersetzungspflege(quelle)

	antwort := anwendung.Autoersetzungen(t.Context(), "7")
	if antwort.Fehler != nil {
		t.Fatalf("%v", antwort.Fehler)
	}

	if quelle.setID != 7 || len(antwort.Regeln) != 1 || !antwort.An {
		t.Errorf("die Antwort lautet %+v", antwort)
	}
}

// **Und die vier Schreibwege.**
func TestDieVierAutoersetzungsschreibwege(t *testing.T) {
	t.Parallel()

	quelle := &pflegequelle{}

	anwendung := app.NeueApp("")
	anwendung.MitAutoersetzungspflege(quelle)

	if antwort := anwendung.AutoersetzungSetzen(
		probekontext(t), "7", "(W)", "weiss", true, true); antwort.Fehler != nil {
		t.Fatalf("Setzen: %v", antwort.Fehler)
	}

	if quelle.gesetzt.Muster != "(W)" || quelle.gesetzt.Ersatz != "weiss" {
		t.Errorf("gesetzt wurde %+v", quelle.gesetzt)
	}

	if antwort := anwendung.AutoersetzungLoeschen(
		probekontext(t), "7", "(W)"); antwort.Fehler != nil || quelle.geloescht != "(W)" {
		t.Errorf("Löschen: %+v / %q", antwort.Fehler, quelle.geloescht)
	}

	if antwort := anwendung.AutoersetzungenZuruecksetzen(
		probekontext(t), "7"); antwort.Fehler != nil || !quelle.zurueck {
		t.Errorf("Zurücksetzen: %+v / %v", antwort.Fehler, quelle.zurueck)
	}

	if antwort := anwendung.AutoersetzungSchalten(
		probekontext(t), "7", false); antwort.Fehler != nil {
		t.Fatalf("Schalten: %v", antwort.Fehler)
	}

	if quelle.geschaltet == nil || *quelle.geschaltet {
		t.Errorf("geschaltet wurde %v", quelle.geschaltet)
	}
}

// **Ohne Naht sagen alle fünf es** — und die Liste ist leer, nicht `nil`.
func TestOhneNahtSagenDieAutoersetzungsbindungenEs(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")

	antwort := anwendung.Autoersetzungen(t.Context(), "7")
	if antwort.Fehler == nil || antwort.Regeln == nil {
		t.Errorf("ohne Naht kam %+v", antwort)
	}

	if a := anwendung.AutoersetzungSetzen(
		probekontext(t), "7", "x", "y", true, true); a.Fehler == nil {
		t.Error("Setzen ohne Naht kam durch")
	}

	if a := anwendung.AutoersetzungLoeschen(t.Context(), "7", "x"); a.Fehler == nil {
		t.Error("Löschen ohne Naht kam durch")
	}

	if a := anwendung.AutoersetzungenZuruecksetzen(t.Context(), "7"); a.Fehler == nil {
		t.Error("Zurücksetzen ohne Naht kam durch")
	}

	if a := anwendung.AutoersetzungSchalten(t.Context(), "7", true); a.Fehler == nil {
		t.Error("Schalten ohne Naht kam durch")
	}
}

// **Jede Bindung prüft ihre Kennung.**
func TestJedeAutoersetzungsbindungPrueftIhreKennung(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitAutoersetzungspflege(&pflegequelle{})

	if a := anwendung.Autoersetzungen(t.Context(), "keine Zahl"); a.Fehler == nil {
		t.Error("Lesen nahm die Zeichenkette")
	}

	if a := anwendung.AutoersetzungSetzen(
		probekontext(t), "keine Zahl", "x", "y", true, true); a.Fehler == nil {
		t.Error("Setzen nahm die Zeichenkette")
	}

	if a := anwendung.AutoersetzungLoeschen(
		probekontext(t), "keine Zahl", "x"); a.Fehler == nil {
		t.Error("Löschen nahm die Zeichenkette")
	}

	if a := anwendung.AutoersetzungenZuruecksetzen(
		probekontext(t), "keine Zahl"); a.Fehler == nil {
		t.Error("Zurücksetzen nahm die Zeichenkette")
	}

	if a := anwendung.AutoersetzungSchalten(
		probekontext(t), "keine Zahl", true); a.Fehler == nil {
		t.Error("Schalten nahm die Zeichenkette")
	}
}

// **Ein Fehler der Naht kommt durch.**
func TestEinFehlerDerAutoersetzungspflegeKommtDurch(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitAutoersetzungspflege(&pflegequelle{
		fehler: fmt.Errorf("%w: kein Spielpaket", domain.ErrNotFound),
	})

	antwort := anwendung.Autoersetzungen(t.Context(), "7")
	if antwort.Fehler == nil || !strings.Contains(antwort.Fehler.Meldung, "Spielpaket") {
		t.Errorf("der Fehler lautet %+v", antwort.Fehler)
	}

	if a := anwendung.AutoersetzungSetzen(
		probekontext(t), "7", "x", "y", true, true); a.Fehler == nil {
		t.Error("der Fehler kam beim Setzen nicht durch")
	}
}

// **Und eine leere Liste bleibt eine Liste.**
func TestEineLeereRegellisteBleibtEineListe(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitAutoersetzungspflege(&pflegequelle{an: true})

	if antwort := anwendung.Autoersetzungen(t.Context(), "7"); antwort.Regeln == nil {
		t.Error("die Liste ist nil")
	}
}
