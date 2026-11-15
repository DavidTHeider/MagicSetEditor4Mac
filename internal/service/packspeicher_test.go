package service_test

// **Dasselbe Angebot wird nicht zweimal gerechnet** —
// #1288.
//
// # Woran man es sieht
//
// Ein Skript hat keine Nebenwirkung, an der ein Test zählen könnte. Gezählt
// wird deshalb, wie oft der Lauf **den Stil einer Karte** nachschlägt: das tut
// er einmal je Karte, und nur dann, wenn er wirklich filtert. Bleibt der
// Zähler beim zweiten Aufruf stehen, wurde nicht gerechnet.
//
// **Und das ist beobachtbares Verhalten und keine Buchführung**: an der
// laufenden Wails-Brücke gemessen (`Vanguard.mse-set`, 107 Karten,
// `m15-altered`, volles Pack) fielen drei Aufrufe hintereinander auf
// **3.078 ms · 12 ms · 13 ms** — der erste kalt, die beiden anderen warm.
//
// **Die Zahl 1.469 → 3 ms stand hier als „über die Kommandozeile" und war in
// ihrer Reichweite falsch**: jeder CLI-Aufruf baut einen eigenen
// `Packbereich`, und über zwei Prozesse kann kein Speicher treffen. Gemessen
// war eine Schleife **innerhalb** eines Prozesses. Angestrichen vom Nachlauf
// über Phase 4.1, Befund 4.

import (
	"context"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// zaehlendeStile zählt, wie oft ein Stil je Karte nachgeschlagen wird.
type zaehlendeStile struct {
	mu      sync.Mutex
	vorlage service.Vorlage
	rufe    int
}

func (z *zaehlendeStile) StilFuerSet(context.Context, int64) (
	service.Vorlage, []domain.Befund, bool,
) {
	return z.vorlage, nil, true
}

func (z *zaehlendeStile) StilFuerKarte(context.Context, *domain.CardRow) (
	service.Vorlage, []domain.Befund, bool,
) {
	z.mu.Lock()
	defer z.mu.Unlock()

	z.rufe++

	return z.vorlage, nil, true
}

func (z *zaehlendeStile) stand() int {
	z.mu.Lock()
	defer z.mu.Unlock()

	return z.rufe
}

// **Der zweite Aufruf rechnet nicht noch einmal.**
func TestDasAngebotWirdNichtZweimalGerechnet(t *testing.T) {
	t.Parallel()

	stile := &zaehlendeStile{vorlage: packvorlage()}

	bereich := service.NeuerPackbereich(nil,
		setkarten{karten: []*domain.CardRow{packkarte(1, "common")}},
		service.NeuerPlanbauer().MitStilen(stile))

	erstes, _, err := bereich.Angebot(t.Context(), 1)
	if err != nil {
		t.Fatalf("Angebot: %v", err)
	}

	nachErstem := stile.stand()
	if nachErstem == 0 {
		t.Fatal("der erste Lauf hat keinen Stil nachgeschlagen — der Zähler " +
			"misst die falsche Stelle")
	}

	zweites, _, err := bereich.Angebot(t.Context(), 1)
	if err != nil {
		t.Fatalf("zweites Angebot: %v", err)
	}

	if stile.stand() != nachErstem {
		t.Errorf("der zweite Lauf hat %d Stile nachgeschlagen — erwartet war "+
			"keiner, das Angebot stand schon da", stile.stand()-nachErstem)
	}

	if len(erstes) != len(zweites) {
		t.Errorf("die beiden Antworten sind verschieden: %+v gegen %+v",
			erstes, zweites)
	}
}

// **Eine geänderte Karte macht das Gemerkte ungültig.**
//
// **Ohne Stand und ohne Zähler**: der Schlüssel ist der Inhalt. Was sich
// ändert, ändert ihn — und wer eine Ungültigkeitsregel schreiben müsste,
// könnte sie vergessen.
func TestEineGeaenderteKarteRechnetDasAngebotNeu(t *testing.T) {
	t.Parallel()

	stile := &zaehlendeStile{vorlage: packvorlage()}
	karte := packkarte(1, "common")

	bereich := service.NeuerPackbereich(nil,
		setkarten{karten: []*domain.CardRow{karte}},
		service.NeuerPlanbauer().MitStilen(stile))

	erstes, _, err := bereich.Angebot(t.Context(), 1)
	if err != nil {
		t.Fatalf("Angebot: %v", err)
	}

	nachErstem := stile.stand()

	// **Dieselbe Karte, ein anderer Wert** — und damit ein anderer Schlüssel.
	karte.Values[0].Text = "rare"

	zweites, _, err := bereich.Angebot(t.Context(), 1)
	if err != nil {
		t.Fatalf("zweites Angebot: %v", err)
	}

	if stile.stand() == nachErstem {
		t.Error("der zweite Lauf hat nichts nachgeschlagen — das Gemerkte galt " +
			"weiter, obwohl die Karte eine andere Seltenheit hat")
	}

	if gleicheAngebote(erstes, zweites) {
		t.Errorf("beide Antworten sind gleich: %+v — die Karte ist von "+
			"»common« auf »rare« gewechselt", erstes)
	}
}

// gleicheAngebote sagt, ob zwei Angebote dieselben Zahlen tragen.
func gleicheAngebote(eines, anderes []service.Packangebot) bool {
	if len(eines) != len(anderes) {
		return false
	}

	for i := range eines {
		if eines[i] != anderes[i] {
			return false
		}
	}

	return true
}

// **Zwei verschieden verteilte Inhalte bekommen nicht denselben Schlüssel.**
//
// # Der Befund
//
// Der erste Anlauf schrieb `Feld=Text\n` in den Schlüssel. Ein Zeilenumbruch
// darf in einem Kartenwert stehen — und dann erzeugten zwei **verschiedene**
// Karten dieselben Bytes. Das ist eine Kollision **vor** SHA-256 und kein
// Hashzufall: die zweite Karte hätte die Filterung der ersten geerbt
// (Nachlauf über Phase 4.1, Befund 5.1).
//
// **Der Fall ist nicht ausgedacht.** `card.notes` trägt in jedem kanonischen
// Set mehrzeiligen Text („Sets: [pmoa];\n\nFormats: ;"), und `rule text` erst
// recht.
func TestZweiKarteninhalteHabenNichtDenselbenSchluessel(t *testing.T) {
	t.Parallel()

	// **Dieselben Bytes bei der ursprünglichen Verkettung** `\t<Feld>=<Text>\n`:
	// ein Feld `a` mit dem Text `1\n\tb=2` schreibt sich genau wie zwei Felder
	// `a` mit `1` und `b` mit `2`.
	eine := &domain.CardRow{
		ID: 1, SetID: 1,
		Values: []domain.Value{
			{Field: "rarity", Text: "common"},
			{Field: "a", Text: "1\n\tb=2"},
		},
	}

	andere := &domain.CardRow{
		ID: 1, SetID: 1,
		Values: []domain.Value{
			{Field: "rarity", Text: "common"},
			{Field: "a", Text: "1"},
			{Field: "b", Text: "2"},
		},
	}

	stile := &zaehlendeStile{vorlage: packvorlage()}

	// **Ein Bereich und zwei Inhalte.** Zwei Bereiche hätten je einen leeren
	// Speicher — dann könnte gar nichts kollidieren, und der Lauf prüfte
	// nichts. Der erste Anlauf machte genau das, und die Gegenprobe blieb
	// grün; §10 sagt, was dann zu tun ist.
	quelle := &wechselndeKarten{karten: []*domain.CardRow{eine}}

	bereich := service.NeuerPackbereich(nil, quelle,
		service.NeuerPlanbauer().MitStilen(stile))

	if _, _, err := bereich.Angebot(t.Context(), 1); err != nil {
		t.Fatalf("Angebot: %v", err)
	}

	nachErstem := stile.stand()

	// **Dasselbe Set, ein anderer Inhalt.** Der Speicher darf nicht treffen.
	quelle.setze([]*domain.CardRow{andere})

	if _, _, err := bereich.Angebot(t.Context(), 1); err != nil {
		t.Fatalf("zweites Angebot: %v", err)
	}

	if stile.stand() == nachErstem {
		t.Error("der zweite Inhalt hat nichts nachgeschlagen — er bekam den " +
			"Schlüssel des ersten")
	}
}

// wechselndeKarten gibt Karten, die sich zwischen zwei Läufen ändern.
type wechselndeKarten struct {
	mu     sync.Mutex
	karten []*domain.CardRow
}

func (w *wechselndeKarten) setze(karten []*domain.CardRow) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.karten = karten
}

func (w *wechselndeKarten) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.karten, nil
}

func (w *wechselndeKarten) Setpackarten(context.Context, int64) ([]*domain.PackArt, error) {
	return nil, nil
}
