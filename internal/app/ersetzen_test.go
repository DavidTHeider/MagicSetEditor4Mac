package app_test

// **Suchen und Ersetzen über die Naht** — Werkstück 2 von Phase 3.7
// (#1168).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ersetzungsquelle ist ein Doppel des Anwendungsfalls.
type ersetzungsquelle struct {
	schau     domain.Ersetzungsvorschau
	fehler    error
	getan     bool
	gefragt   domain.Ersetzungsfrage
	gefragtes int64
}

func (e *ersetzungsquelle) Vorschau(
	_ context.Context, setID int64, frage domain.Ersetzungsfrage,
) (domain.Ersetzungsvorschau, error) {
	e.gefragt, e.gefragtes = frage, setID

	return e.schau, e.fehler
}

func (e *ersetzungsquelle) Ausfuehren(
	ctx context.Context, setID int64, frage domain.Ersetzungsfrage,
) (domain.Ersetzungsvorschau, error) {
	e.getan = true

	return e.Vorschau(ctx, setID, frage)
}

// eineStelle ist eine Vorschau mit genau einer Stelle.
func eineStelle() domain.Ersetzungsvorschau {
	return domain.Ersetzungsvorschau{
		Stellen: []domain.Ersetzungsstelle{{
			KarteID: 7, Nummer: 2, Feld: "rule text",
			Vorher: "Fliegend", Nachher: "Rennend", Treffer: 1,
		}},
		Karten: 1, Treffer: 1,
	}
}

// mitErsetzung baut eine Anwendung mit dem Doppel.
func mitErsetzung(quelle app.Ersetzungsquelle) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitErsetzung(quelle)

	return anwendung
}

// dieFrage ist die Frage, wie das Fenster sie stellt.
var dieFrage = app.ErsetzungsfrageDTO{
	Text: "Flieg", Ersatz: "Renn", Feld: "rule text",
	GrossKleinBeachten: true, GanzesWort: true,
}

// **Die Vorschau reicht die ganze Frage durch — und schreibt nichts.**
func TestDieVorschauReichtDieFrageDurch(t *testing.T) {
	t.Parallel()

	quelle := &ersetzungsquelle{schau: eineStelle()}

	antwort := mitErsetzung(quelle).ErsetzungVorschau(probekontext(t), "1", dieFrage)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if quelle.getan {
		t.Error("die Vorschau hat ausgeführt")
	}

	if quelle.gefragtes != 1 || quelle.gefragt.Text != "Flieg" ||
		quelle.gefragt.Ersatz != "Renn" || quelle.gefragt.Feld != "rule text" ||
		!quelle.gefragt.GrossKleinBeachten || !quelle.gefragt.GanzesWort {
		t.Errorf("gefragt wurde %d / %+v", quelle.gefragtes, quelle.gefragt)
	}

	// **Die Nummer ist eins-basiert**, wie die Liste zählt.
	if len(antwort.Stellen) != 1 || antwort.Stellen[0].KarteID != "7" ||
		antwort.Stellen[0].Nummer != 3 || antwort.Stellen[0].Nachher != "Rennend" {
		t.Errorf("Stellen: %+v", antwort.Stellen)
	}

	if antwort.Karten != 1 || antwort.Treffer != 1 || antwort.Weitere {
		t.Errorf("%d Karten, %d Treffer, weitere=%v",
			antwort.Karten, antwort.Treffer, antwort.Weitere)
	}
}

// **Ausführen geht denselben Weg — über die andere Methode.**
func TestAusfuehrenGehtDenAnderenWeg(t *testing.T) {
	t.Parallel()

	quelle := &ersetzungsquelle{schau: eineStelle()}

	if antwort := mitErsetzung(quelle).ErsetzungAusfuehren(
		probekontext(t), "1", dieFrage); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if !quelle.getan {
		t.Error("ausgeführt wurde nicht")
	}
}

// **Eine unbrauchbare Kennung erreicht die Naht nicht** — und die Antwort
// trägt trotzdem eine Liste.
func TestEineUnbrauchbareKennungErreichtDasErsetzenNicht(t *testing.T) {
	t.Parallel()

	quelle := &ersetzungsquelle{schau: eineStelle()}

	antwort := mitErsetzung(quelle).ErsetzungVorschau(probekontext(t), "abc", dieFrage)
	if antwort.Fehler == nil || antwort.Fehler.Feld != "setID" {
		t.Fatalf("%+v", antwort.Fehler)
	}

	// **`[]` und nicht `null`** — eine leere Liste, die als `null` ankommt,
	// kostet jeden Leser im Fenster einen Zweig (§3).
	if antwort.Stellen == nil {
		t.Error("die Antwort trägt keine Liste")
	}

	if quelle.gefragtes != 0 {
		t.Error("die Naht wurde trotzdem gefragt")
	}
}

// **Ohne Naht sagt die Bindung es, statt zu schweigen.**
func TestOhneErsetzungsnahtSagenDieBindungenEs(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").ErsetzungVorschau(probekontext(t), "1", dieFrage)
	if antwort.Fehler == nil {
		t.Fatal("ohne Naht kam kein Fehler")
	}

	if antwort.Stellen == nil {
		t.Error("die Antwort trägt keine Liste")
	}
}

// **Ein Fehler des Dienstes wird übersetzt.**
func TestEinErsetzungsfehlerWirdUebersetzt(t *testing.T) {
	t.Parallel()

	quelle := &ersetzungsquelle{fehler: domain.ErrBelegt}

	antwort := mitErsetzung(quelle).ErsetzungAusfuehren(probekontext(t), "1", dieFrage)
	if antwort.Fehler == nil || antwort.Fehler.Code == "" {
		t.Fatalf("%+v", antwort.Fehler)
	}
}

// **Die Naht für die Flächen ohne Fenster trägt beide Wege durch.**
func TestDieErsetzungsnahtTraegtBeideWege(t *testing.T) {
	t.Parallel()

	quelle := &ersetzungsquelle{schau: eineStelle()}
	naht := (&app.Fremdersetzung{}).MitErsetzungsquelle(quelle)

	schau, fehler := naht.Vorschau(probekontext(t), 1, domain.Ersetzungsfrage{Text: "a"})
	if fehler != nil || schau.Treffer != 1 {
		t.Fatalf("Vorschau: %+v / %v", schau, fehler)
	}

	if quelle.getan {
		t.Error("die Vorschau hat ausgeführt")
	}

	if _, err := naht.Ausfuehren(
		probekontext(t), 1, domain.Ersetzungsfrage{Text: "a"}); err != nil {
		t.Fatalf("Ausführen: %v", err)
	}

	if !quelle.getan {
		t.Error("ausgeführt wurde nicht")
	}
}

// **Und ein Fehler kommt an beiden Wegen durch.**
func TestDieErsetzungsnahtReichtIhrenFehlerDurch(t *testing.T) {
	t.Parallel()

	naht := (&app.Fremdersetzung{}).MitErsetzungsquelle(
		&ersetzungsquelle{fehler: domain.ErrBelegt})

	if _, fehler := naht.Vorschau(
		probekontext(t), 1, domain.Ersetzungsfrage{Text: "a"}); !errors.Is(
		fehler, domain.ErrBelegt) {
		t.Errorf("Vorschau: %v", fehler)
	}

	if _, fehler := naht.Ausfuehren(
		probekontext(t), 1, domain.Ersetzungsfrage{Text: "a"}); !errors.Is(
		fehler, domain.ErrBelegt) {
		t.Errorf("Ausführen: %v", fehler)
	}
}

// **Ohne Benutzerverzeichnis wird gemeldet statt geraten.**
//
// Dieselbe Zusage wie bei der Ablagenaht daneben: der Stilwähler braucht die
// Paketwurzeln, und die stehen im Zuhause.
func TestDieErsetzungsflaechennahtMeldetEinenFehlendenOrt(t *testing.T) {
	t.Setenv("HOME", "")

	if _, fehler := app.Ersetzungsflaechennaht(
		nil, domain.UrheberKommandozeile); fehler == nil {
		t.Error("ohne Zuhause kam eine Naht")
	}
}

// **Und mit Zuhause kommt sie zusammen und trägt.**
func TestDieErsetzungsflaechennahtTraegtUeberEinenEchtenBestand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	setID, karteID := saatsetImBestand(t, db)

	naht, nahtfehler := app.Ersetzungsflaechennaht(db, domain.UrheberKommandozeile)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	// Das Saatset trägt in `text` einen Wortlistenwert; er ist ein Textfeld,
	// und genau darin soll gesucht werden.
	frage := domain.Ersetzungsfrage{Text: "Kreatur", Ersatz: "Artefakt"}

	schau, fehler := naht.Vorschau(probekontext(t), setID, frage)
	if fehler != nil {
		t.Fatalf("Vorschau: %v", fehler)
	}

	if schau.Treffer == 0 {
		t.Fatalf("die Vorschau fand nichts — die Karte %d trägt »Kreatur«", karteID)
	}

	// **Und die Vorschau hat nicht geschrieben.**
	getan, ausfuehrfehler := naht.Ausfuehren(probekontext(t), setID, frage)
	if ausfuehrfehler != nil {
		t.Fatalf("Ausführen: %v", ausfuehrfehler)
	}

	if getan.Treffer != schau.Treffer {
		t.Errorf("%d Treffer ausgeführt, %d in der Vorschau",
			getan.Treffer, schau.Treffer)
	}

	danach, nachfehler := naht.Vorschau(probekontext(t), setID, frage)
	if nachfehler != nil {
		t.Fatalf("Nachschau: %v", nachfehler)
	}

	if danach.Treffer != 0 {
		t.Errorf("nach dem Ersetzen stehen noch %d Treffer", danach.Treffer)
	}
}
