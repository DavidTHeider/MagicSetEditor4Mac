package service_test

// **Der Abgleich zweier Sammlungen** — Werkstück `T4` von Phase 5.6
// (#1469).
//
// # Was hier gemessen wird
//
// **Dass die fremde Seite ein Schnappschuss ist und keine Abfrage.** Was ein
// Gegenüber zeigt, hat es freigegeben (T3); dieses Werkstück verarbeitet, was
// angekommen ist, und greift nirgends nach.
//
// **Dass das Konzept aus dem eigenen Druckverzeichnis kommt.** Ein
// druckunabhängiger Wunsch trifft über die Kennung, die mehrere Drucke teilen
// — und die steht hier, nicht in der fremden Zeile.
//
// **Und dass ein Abbruch abbricht.** Der Vorschlag entsteht auf Tastendruck;
// wer den Tisch wechselt, während gerechnet wird, bekommt kein Ergebnis von
// vorhin.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

const (
	shivanDruck = "shivan-lea"
	boltDruck   = "bolt-m10"
	shivanNeu   = "shivan-m10"
)

// druckverzeichnisdoppel gibt die Drucke, die der Abgleich auflösen soll.
type druckverzeichnisdoppel struct {
	drucke []domain.Druck
	fehler error

	// abbrechen bricht den Kontext **während** des Lesens ab — damit sich der
	// zweite Blick auf den Kontext messen lässt (§3, Gegenprobe am Schaden).
	abbrechen func()
}

func (d druckverzeichnisdoppel) Alle(_ context.Context) ([]domain.Druck, error) {
	if d.abbrechen != nil {
		d.abbrechen()
	}

	return d.drucke, d.fehler
}

// stolperndeSammlung scheitert erst beim n-ten Lauf.
//
// **Erst beim n-ten**, weil beide Sichten dieselbe Ablage lesen: ein Doppel,
// das immer scheitert, misst nur den ersten von zwei Wegen.
type stolperndeSammlung struct {
	sammlungsdoppel

	abLauf int
	fehler error
	laeufe int
}

func (s *stolperndeSammlung) Liste(ctx context.Context) ([]*domain.Exemplar, error) {
	s.laeufe++
	if s.laeufe >= s.abLauf {
		return nil, s.fehler
	}

	return s.sammlungsdoppel.Liste(ctx)
}

// einDruckMitKonzept baut eine Druckzeile samt ihren zwei Kennungen.
func einDruckMitKonzept(printing, konzept string) domain.Druck {
	return domain.Druck{
		Name: "Shivan Dragon",
		Kennungen: []domain.Druckkennung{
			{System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: printing},
			{System: domain.SystemScryfall, Semantik: domain.SemantikKonzept, Wert: konzept},
		},
	}
}

// meinExemplar baut eine eigene Zeile.
func meinExemplar(printing string, menge int, wunsch bool) *domain.Exemplar {
	return &domain.Exemplar{
		ID: int64(len(printing)), Printing: printing, Menge: menge,
		Zustand: domain.ZustandNearMint, Sprache: "de",
		Finish: domain.FinishNonfoil, Wunsch: wunsch,
		Name: "Shivan Dragon",
	}
}

// eineFremdzeile baut, was von drüben angekommen ist.
func eineFremdzeile(printing string, menge int) domain.Sichtkarte {
	return domain.Sichtkarte{
		Printing: printing, Menge: menge, Zustand: domain.ZustandNearMint,
		Sprache: "de", Finish: domain.FinishNonfoil, Bezeichnung: "Shivan Dragon",
	}
}

// einTauschwerk stellt den Abgleich über die zwei Doppel.
func einTauschwerk(eigene []*domain.Exemplar, drucke []domain.Druck) *service.Tauschwerk {
	return service.NeuesTauschwerk(
		service.NeueSammlungssichten(&sammlungsdoppel{zeilen: eigene}),
		druckverzeichnisdoppel{drucke: drucke},
	)
}

func TestMeineDubletteTrifftSeinenWunsch(t *testing.T) {
	t.Parallel()

	werk := einTauschwerk(
		[]*domain.Exemplar{meinExemplar(shivanDruck, 3, false)},
		[]domain.Druck{einDruckMitKonzept(shivanDruck, "konzept-shivan")},
	)

	abgleich, fehler := werk.Abgleichen(context.Background(), service.Fremdstand{
		Kennung:  "bgegenueber",
		Wuensche: []domain.Sichtkarte{eineFremdzeile(shivanDruck, 1)},
	}, 1, domain.Tauschumfang{})
	if fehler != nil {
		t.Fatalf("Abgleichen: %v", fehler)
	}

	if len(abgleich.Gebe) != 1 || abgleich.Gebe[0].Menge != 1 {
		t.Fatalf("es kam %+v", abgleich.Gebe)
	}

	// **Ohne Gegenleistung gibt es keinen Vorschlag** — der leere Tausch ist
	// keiner.
	if abgleich.HatVorschlag {
		t.Error("es kam ein einseitiger Vorschlag")
	}
}

func TestSeinAngebotTrifftMeinenWunsch(t *testing.T) {
	t.Parallel()

	werk := einTauschwerk(
		[]*domain.Exemplar{meinExemplar(boltDruck, 2, true)},
		[]domain.Druck{einDruckMitKonzept(boltDruck, "konzept-bolt")},
	)

	abgleich, fehler := werk.Abgleichen(context.Background(), service.Fremdstand{
		Kennung: "bgegenueber",
		Angebot: []domain.Sichtkarte{eineFremdzeile(boltDruck, 5)},
	}, 1, domain.Tauschumfang{})
	if fehler != nil {
		t.Fatalf("Abgleichen: %v", fehler)
	}

	if len(abgleich.Bekomme) != 1 || abgleich.Bekomme[0].Menge != 2 {
		t.Fatalf("es kam %+v", abgleich.Bekomme)
	}
}

// **Beide Seiten zusammen ergeben einen Vorschlag.**
func TestZweiSeitenErgebenEinenVorschlag(t *testing.T) {
	t.Parallel()

	werk := einTauschwerk(
		[]*domain.Exemplar{
			meinExemplar(shivanDruck, 3, false),
			meinExemplar(boltDruck, 1, true),
		},
		[]domain.Druck{
			einDruckMitKonzept(shivanDruck, "konzept-shivan"),
			einDruckMitKonzept(boltDruck, "konzept-bolt"),
		},
	)

	abgleich, fehler := werk.Abgleichen(context.Background(), service.Fremdstand{
		Kennung:  "bgegenueber",
		Angebot:  []domain.Sichtkarte{eineFremdzeile(boltDruck, 4)},
		Wuensche: []domain.Sichtkarte{eineFremdzeile(shivanDruck, 2)},
	}, 1, domain.Tauschumfang{})
	if fehler != nil {
		t.Fatalf("Abgleichen: %v", fehler)
	}

	if !abgleich.HatVorschlag {
		t.Fatal("es kam kein Vorschlag")
	}

	if abgleich.Vorschlag.Gebe.Stueck != 2 || abgleich.Vorschlag.Bekomme.Stueck != 1 {
		t.Errorf("es sind %d hin und %d her",
			abgleich.Vorschlag.Gebe.Stueck, abgleich.Vorschlag.Bekomme.Stueck)
	}
}

// **Das Konzept kommt aus dem eigenen Verzeichnis** — und macht aus zwei
// Drucken derselben Karte einen Treffer, sobald der Wunsch es zulässt.
func TestEinDruckunabhaengigerWunschTrifftEinenAnderenDruck(t *testing.T) {
	t.Parallel()

	wunsch := meinExemplar(shivanDruck, 1, true)
	wunsch.Wunschart = domain.WunschJederDruck

	werk := einTauschwerk(
		[]*domain.Exemplar{wunsch},
		[]domain.Druck{
			einDruckMitKonzept(shivanDruck, "konzept-shivan"),
			einDruckMitKonzept(shivanNeu, "konzept-shivan"),
		},
	)

	abgleich, fehler := werk.Abgleichen(context.Background(), service.Fremdstand{
		Kennung: "bgegenueber",
		Angebot: []domain.Sichtkarte{eineFremdzeile(shivanNeu, 1)},
	}, 1, domain.Tauschumfang{})
	if fehler != nil {
		t.Fatalf("Abgleichen: %v", fehler)
	}

	if len(abgleich.Bekomme) != 1 {
		t.Fatalf("es kamen %d Treffer, Befunde: %v", len(abgleich.Bekomme), abgleich.Befunde)
	}

	// **Die Gegenprobe**: derselbe Wunsch, genau gemeint, trifft nicht.
	wunsch.Wunschart = domain.WunschGenau

	genau, fehler := werk.Abgleichen(context.Background(), service.Fremdstand{
		Kennung: "bgegenueber",
		Angebot: []domain.Sichtkarte{eineFremdzeile(shivanNeu, 1)},
	}, 1, domain.Tauschumfang{})
	if fehler != nil {
		t.Fatalf("Abgleichen: %v", fehler)
	}

	if len(genau.Bekomme) != 0 {
		t.Errorf("der genaue Wunsch hat getroffen: %+v", genau.Bekomme)
	}
}

// **Ein Abbruch bricht ab** — und gibt kein Ergebnis von vorhin.
func TestEinAbbruchBrichtDenAbgleichAb(t *testing.T) {
	t.Parallel()

	werk := einTauschwerk(
		[]*domain.Exemplar{meinExemplar(shivanDruck, 3, false)},
		[]domain.Druck{einDruckMitKonzept(shivanDruck, "konzept-shivan")},
	)

	beendet, abbrechen := context.WithCancel(context.Background())
	abbrechen()

	_, fehler := werk.Abgleichen(beendet, service.Fremdstand{
		Kennung:  "bgegenueber",
		Wuensche: []domain.Sichtkarte{eineFremdzeile(shivanDruck, 1)},
	}, 1, domain.Tauschumfang{})

	if !errors.Is(fehler, context.Canceled) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Ein Fehler des Verzeichnisses kommt durch** und wird nicht zu einem
// leeren Abgleich.
func TestEinFehlerDesVerzeichnissesKommtDurch(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("das Verzeichnis ist zu")

	werk := service.NeuesTauschwerk(
		service.NeueSammlungssichten(&sammlungsdoppel{}),
		druckverzeichnisdoppel{fehler: gestolpert},
	)

	if _, fehler := werk.Abgleichen(context.Background(),
		service.Fremdstand{Kennung: "b"}, 1, domain.Tauschumfang{}); !errors.Is(fehler, gestolpert) {
		t.Errorf("es kam %v", fehler)
	}
}
