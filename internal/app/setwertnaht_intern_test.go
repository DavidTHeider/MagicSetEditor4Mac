package app

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Naht reicht durch und tut sonst nichts.**
//
// # Warum sie überhaupt da ist
//
// Gelesen wird über den `SetLeser`, geschrieben über den `Kartenschreiber` —
// denselben, der auch Kartenfelder schreibt, denn er führt den Verlauf. Zwei
// Schreiber wären zwei Urheber im Verlauf für dieselbe Handlung.
//
// Die Oberfläche stellt beides als **eine** Frage; hier steht, was aus zweien
// eine macht. Eine Entscheidung liegt darin nicht, und genau das prüft dieser
// Test: was hineingeht, kommt heraus.
func TestDieSetwertnahtReichtDurch(t *testing.T) {
	leser := &setwertleserprobe{
		werte: []domain.Value{{Field: "title", Text: "Probe"}},
		kopf:  domain.Setkopf{MSEFassung: "2.1.2"},
	}
	schreiber := &setwertschreiberprobe{}

	naht := setwertnaht{leser: leser, schreiber: schreiber}

	werte, err := naht.Setwerte(probekontext(t), 7)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if leser.setID != 7 || len(werte) != 1 || werte[0].Text != "Probe" {
		t.Errorf("gelesen für %d: %+v", leser.setID, werte)
	}

	// **Und der Setkopf ebenso** (#1325): dieselbe Naht, dieselbe Frage.
	kopf, err := naht.Setkopf(probekontext(t), 9)
	if err != nil {
		t.Fatalf("den Kopf lesen: %v", err)
	}

	if leser.setID != 9 || kopf.MSEFassung != "2.1.2" {
		t.Errorf("gelesen für %d: %+v", leser.setID, kopf)
	}

	if _, _, fehler := naht.SetwertSetzenMitVorwert(
		probekontext(t), 7, "set code", "PRB"); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if schreiber.setID != 7 || schreiber.feld != "set code" || schreiber.wert != "PRB" {
		t.Errorf("%+v", schreiber)
	}
}

// **Und ein Fehler kommt durch** — von beiden Seiten.
func TestDieSetwertnahtReichtJedenFehlerDurch(t *testing.T) {
	kaputt := errors.New("die Platte ist voll")

	naht := setwertnaht{
		leser:     &setwertleserprobe{fehler: kaputt},
		schreiber: &setwertschreiberprobe{fehler: kaputt},
	}

	if _, err := naht.Setwerte(probekontext(t), 7); !errors.Is(err, kaputt) {
		t.Errorf("beim Lesen: %v", err)
	}

	if _, _, err := naht.SetwertSetzenMitVorwert(
		probekontext(t), 7, "set code", "PRB"); !errors.Is(err, kaputt) {
		t.Errorf("beim Schreiben: %v", err)
	}
}

type setwertleserprobe struct {
	werte  []domain.Value
	fehler error
	setID  int64

	// kopf ist, was `Setkopf` zurückgibt (#1325).
	kopf domain.Setkopf
}

func (s *setwertleserprobe) Setwerte(_ context.Context, setID int64) ([]domain.Value, error) {
	s.setID = setID

	return s.werte, s.fehler
}

func (s *setwertleserprobe) Setkopf(
	_ context.Context, setID int64,
) (domain.Setkopf, error) {
	s.setID = setID

	return s.kopf, s.fehler
}

type setwertschreiberprobe struct {
	fehler error

	setID int64
	feld  string
	wert  string
}

func (s *setwertschreiberprobe) SetwertSetzenMitVorwert(
	_ context.Context, setID int64, feld, wert string,
) (vorher string, hatte bool, fehler error) {
	s.setID, s.feld, s.wert = setID, feld, wert

	return "", false, s.fehler
}
