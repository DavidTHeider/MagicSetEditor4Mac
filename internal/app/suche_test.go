package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

type suchprobe struct {
	treffer  []domain.SearchHit
	gekuerzt bool
	werte    map[int64][]domain.Value
	felder   []string
	kontext  context.Context //nolint:containedctx // die Probe merkt sich, was ankam
	frage    domain.SearchQuery

	// sets sind die Sets, die es gibt (#1336).
	//
	// **Leer heißt: keines** — und dann weist die Suche jede `--set=`-Angabe
	// ab. Wer ein Set meint, legt es hier an.
	sets []domain.SetHead

	fehler       error
	werteFehler  error
	felderFehler error
}

func (s *suchprobe) Suche(_ context.Context, frage domain.SearchQuery) (
	domain.SearchResult, error,
) {
	s.frage = frage

	return domain.SearchResult{Hits: s.treffer, Truncated: s.gekuerzt}, s.fehler
}

func (s *suchprobe) Kartenwerte(context.Context, []int64) (map[int64][]domain.Value, error) {
	return s.werte, s.werteFehler
}

func (s *suchprobe) Felder(ctx context.Context, setID int64) ([]string, error) {
	s.kontext = ctx
	s.frage.SetID = setID

	return s.felder, s.felderFehler
}

func mitSucher(probe *suchprobe) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitSucher(probe)

	return anwendung
}

func TestSuchenGibtTrefferMitKartennamen(t *testing.T) {
	probe := &suchprobe{
		treffer: []domain.SearchHit{
			{CardID: 5, SetID: 1, Number: 0, Field: "text", Value: "Feuerball verursacht Schaden."},
		},
		werte: map[int64][]domain.Value{
			5: {{Field: "name", Text: "Feuerball"}, {Field: "typ", Text: "Zauber"}},
		},
	}

	antwort := mitSucher(probe).Suchen(probekontext(t), "Schaden", "", "")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if len(antwort.Treffer) != 1 {
		t.Fatalf("%d Treffer", len(antwort.Treffer))
	}

	// **Der Kartenname steht dabei, nicht nur die Fundstelle.** Sonst lässt eine
	// Trefferliste einen raten, zu welcher Karte sie gehört.
	if antwort.Treffer[0].Kartenname != "Feuerball" {
		t.Errorf("%+v", antwort.Treffer[0])
	}

	if antwort.Treffer[0].KarteID != "5" || antwort.Treffer[0].SetID != "1" {
		t.Errorf("%+v", antwort.Treffer[0])
	}
}

// Die Filter kommen so an, wie sie gesetzt wurden.
func TestDieFilterKommenAn(t *testing.T) {
	// **Das Set muss es geben** (#1336): eine Kennung, die auf nichts zeigt,
	// wird seit dieser Änderung abgewiesen statt als leerer Filter behandelt.
	probe := &suchprobe{sets: []domain.SetHead{{ID: 3}}}

	mitSucher(probe).Suchen(probekontext(t), "feuer", "name", "3")

	if probe.frage.Text != "feuer" || probe.frage.Field != "name" || probe.frage.SetID != 3 {
		t.Errorf("%+v", probe.frage)
	}
}

// Ohne Set bleibt der Setfilter leer — und die Kennung wird nicht gedeutet.
func TestOhneSetBleibtDerSetfilterLeer(t *testing.T) {
	probe := &suchprobe{}

	if antwort := mitSucher(probe).Suchen(probekontext(t), "a", "", ""); antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if probe.frage.SetID != 0 {
		t.Errorf("SetID %d", probe.frage.SetID)
	}
}

func TestEineUnlesbareSetkennungWirdAbgewiesen(t *testing.T) {
	befund := mitSucher(&suchprobe{}).Suchen(probekontext(t), "a", "", "abc").Fehler
	if befund == nil || befund.Feld != "id" {
		t.Errorf("%+v", befund)
	}
}

// **Eine leere Antwort ist `[]`, nicht `null`** — durch die Serialisierung
// geprüft (§3).
func TestEineLeereSuchantwortIstEineListe(t *testing.T) {
	antwort := mitSucher(&suchprobe{}).Suchen(probekontext(t), "gibtsnicht", "", "")

	gebaut, err := json.Marshal(antwort.Treffer)
	if err != nil {
		t.Fatalf("serialisieren: %v", err)
	}

	if string(gebaut) != "[]" {
		t.Errorf("%s", gebaut)
	}
}

// Eine Karte ohne Namensfeld bekommt einen leeren Namen, keinen Absturz.
func TestEineKarteOhneNamensfeld(t *testing.T) {
	probe := &suchprobe{
		treffer: []domain.SearchHit{{CardID: 5, Field: "typ", Value: "Zauber"}},
		werte:   map[int64][]domain.Value{5: {{Field: "typ", Text: "Zauber"}}},
	}

	antwort := mitSucher(probe).Suchen(probekontext(t), "Zauber", "", "")
	if antwort.Fehler != nil || antwort.Treffer[0].Kartenname != "" {
		t.Errorf("%+v", antwort)
	}
}

func TestSuchfehlerKommenAn(t *testing.T) {
	roh := errors.New("SQL logic error: no such table: card_suche")

	befund := mitSucher(&suchprobe{fehler: roh}).Suchen(probekontext(t), "a", "", "").Fehler
	if befund == nil || befund.Code != app.CodeUnbekannt {
		t.Fatalf("%+v", befund)
	}

	// **Die rohe Meldung erreicht die Oberfläche nie** (§8).
	if befund.Meldung == roh.Error() {
		t.Errorf("die rohe Meldung kam durch: %q", befund.Meldung)
	}
}

// Ein Fehler beim Nachschlagen der Namen kommt ebenfalls an.
func TestEinFehlerBeimNamenNachschlagenKommtAn(t *testing.T) {
	probe := &suchprobe{
		treffer:     []domain.SearchHit{{CardID: 5}},
		werteFehler: errors.New("nicht lesbar"),
	}

	if mitSucher(probe).Suchen(probekontext(t), "a", "", "").Fehler == nil {
		t.Error("der Fehler ging verloren")
	}
}

func TestSuchfelderKommenAusDemBestand(t *testing.T) {
	probe := &suchprobe{felder: []string{"name", "typ"}}

	antwort := mitSucher(probe).Suchfelder(probekontext(t), "3")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if len(antwort.Felder) != 2 || probe.frage.SetID != 3 {
		t.Errorf("%+v / %+v", antwort, probe.frage)
	}

	// Ohne Set über alle Sets.
	mitSucher(probe).Suchfelder(probekontext(t), "")

	if probe.frage.SetID != 0 {
		t.Errorf("SetID %d", probe.frage.SetID)
	}
}

func TestSuchfelderMeldenBefunde(t *testing.T) {
	if befund := mitSucher(&suchprobe{}).Suchfelder(probekontext(t), "abc").Fehler; befund == nil ||
		befund.Feld != "id" {
		t.Errorf("%+v", befund)
	}

	probe := &suchprobe{felderFehler: errors.New("nicht lesbar")}
	if mitSucher(probe).Suchfelder(probekontext(t), "").Fehler == nil {
		t.Error("der Fehler ging verloren")
	}
}

// **Eine leere Feldliste ist `[]`, nicht `null`.**
func TestEineLeereFeldlisteIstEineListe(t *testing.T) {
	gebaut, err := json.Marshal(mitSucher(&suchprobe{}).Suchfelder(probekontext(t), "").Felder)
	if err != nil {
		t.Fatalf("serialisieren: %v", err)
	}

	if string(gebaut) != "[]" {
		t.Errorf("%s", gebaut)
	}
}

// **Die Kürzung kommt am Frontend an.** Eine stille sähe aus wie ein
// vollständiges Ergebnis.
func TestDieKuerzungWirdDurchgereicht(t *testing.T) {
	probe := &suchprobe{
		treffer:  []domain.SearchHit{{CardID: 1, Field: "name", Value: "a"}},
		gekuerzt: true,
	}

	if antwort := mitSucher(probe).Suchen(probekontext(t), "a", "", ""); !antwort.Gekuerzt {
		t.Errorf("%+v", antwort)
	}
}

// Setkoepfe gibt die Sets, die es gibt (#1336).
func (s *suchprobe) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return s.sets, s.fehler
}
