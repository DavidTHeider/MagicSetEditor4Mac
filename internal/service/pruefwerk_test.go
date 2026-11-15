package service_test

// **Die Vorrichtung der Formatprüfung** — Werkstück 5 von Phase 4.6
// (#1293).
//
// **Eine eigene Datei**, damit `formatpruefung_test.go` unter der Zeilengrenze
// aus §2 bleibt.

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

const (
	druckA = "a1b2c3d4-0000-0000-0000-00000000000a"
	druckB = "a1b2c3d4-0000-0000-0000-00000000000b"
)

// zulaessigkeitsdoppel antwortet **je Kennung** mit abgelegten
// Zulässigkeiten.
//
// **Neben `druckdoppel` und nicht statt seiner**: jenes gibt auf jeden Ruf
// dieselbe Karte zurück, und eine Prüfung über zwei verschiedene Drucke
// bekäme damit zweimal dieselbe Antwort — eine Vorrichtung, die großzügiger
// wäre als der Betrieb (§3).
type zulaessigkeitsdoppel struct {
	legalitaeten map[string]map[string]string

	// typzeilen sagt, was für eine Karte ein Druck ist — daran hängt, ob die
	// Höchstzahl gleicher Karten ihn ausnimmt (#1370).
	typzeilen map[string]string

	// manawerte sind die Zahlen, die Scryfall als `cmc` führt (U10, #1421).
	manawerte map[string]int

	// namen sagt, wie ein Druck heißt — daran hängt, **welche Karte** er ist
	// (Befund 1 des Zweitblicks über Phase 4.6).
	namen map[string]string

	// orakel sagt, welche **Karte hinter den Drucken** ein Druck ist.
	//
	// Zwei englische Namen können dieselbe Karte sein — „Universes Beyond"
	// und „Universes Within" teilen sich eine `oracle_id`, und CR 100.2a
	// zählt sie zusammen
	// (#1375).
	orakel map[string]string

	// gefragt zählt die Einzelrufe — für die Zusage, dass nicht gefragt
	// wird, wo die Antwort niemandem nützt.
	gefragt int

	// gesammelt zählt die Sammelrufe, schuebe hält ihre Größen.
	//
	// **Beides**, denn die Zusage aus Befund 6 hat zwei Hälften: dass
	// überhaupt gebündelt wird, und dass kein Schub über [scryfall.Sammelgrenze]
	// geht. Nur die Zahl der Rufe zu zählen ließe einen Ruf mit hundert
	// Kennungen als Erfolg durchgehen — und Scryfall antwortete darauf mit
	// einem Fehler.
	gesammelt int
	schuebe   []int

	// unbekannt sind Drucke, die der Sammelweg unter `not_found` führt.
	unbekannt map[string]bool

	fehler error

	// sammelfehler lässt **nur** den Sammelweg scheitern — für die Zusage,
	// dass dann der Einzelweg einspringt (§16, Regel 5).
	sammelfehler error
}

// Sammlung beantwortet einen ganzen Schub — wie `POST /cards/collection`.
func (d *zulaessigkeitsdoppel) Sammlung(ctx context.Context, kennungen []scryfall.Kennung) (
	scryfall.Sammelantwort, time.Duration, []domain.Befund, error,
) {
	d.gesammelt++
	d.schuebe = append(d.schuebe, len(kennungen))

	if d.sammelfehler != nil {
		return scryfall.Sammelantwort{}, 0, nil, d.sammelfehler
	}

	if d.fehler != nil {
		return scryfall.Sammelantwort{}, 0, nil, d.fehler
	}

	if len(kennungen) > scryfall.Sammelgrenze {
		return scryfall.Sammelantwort{}, 0, nil, fmt.Errorf(
			"%w: %d Kennungen — Scryfall nimmt höchstens %d",
			domain.ErrValidation, len(kennungen), scryfall.Sammelgrenze)
	}

	antwort := scryfall.Sammelantwort{
		Karten: []*scryfall.Karte{}, NichtGefunden: []scryfall.Kennung{},
	}

	for _, kennung := range kennungen {
		if d.unbekannt[kennung.ID] {
			antwort.NichtGefunden = append(antwort.NichtGefunden, kennung)

			continue
		}

		// **Über den Einzelweg, damit beide dieselbe Karte geben.** Eine
		// Vorrichtung, in der der Sammelweg großzügiger antwortet als der
		// einzelne, prüfte die Bündelung und nicht das Ergebnis (§3).
		karte, _, _, _ := d.KarteNachKennung(ctx, kennung.ID) //nolint:errcheck // gibt hier nie einen
		antwort.Karten = append(antwort.Karten, karte)
	}

	// **Der Einzelweg hat dabei mitgezählt** — das soll er nicht: `gefragt`
	// beantwortet die Frage, ob **einzeln** gefragt wurde.
	d.gefragt -= len(antwort.Karten)

	return antwort, 0, nil, nil
}

func (d *zulaessigkeitsdoppel) KarteNachKennung(_ context.Context, kennung string) (
	*scryfall.Karte, time.Duration, []domain.Befund, error,
) {
	d.gefragt++

	if d.fehler != nil {
		return nil, 0, nil, d.fehler
	}

	// **Kanonisch klein**, wie Scryfall antwortet: nachgemessen am
	// Nachlauf über Phase 3.4 —
	// `/cards/516E137A-…` antwortet mit `516e137a-…`. Ein Doppel, das die
	// Schreibweise der Frage zurückgäbe, wäre großzügiger als der Betrieb
	// (§3) und ließe eine fehlende Glättung durchgehen.
	kennung = strings.ToLower(kennung)

	name := d.namen[kennung]
	if name == "" {
		name = "Druck " + kennung
	}

	return &scryfall.Karte{
		Kennung: kennung, Name: name,
		Orakelkennung: d.orakel[kennung],
		Typzeile:      d.typzeilen[kennung],
		Manawert:      d.manawerte[kennung],
		Legalitaeten:  d.legalitaeten[kennung],
	}, 0, nil, nil
}

// pruefwerk hält alles zusammen, was eine Prüfung braucht.
type pruefwerk struct {
	decks    *deckdoppel
	formate  *formatdoppel
	drucke   *zulaessigkeitsdoppel
	verweise *verweisdoppel

	// karten ist die Naht zu den Feldern einer **eigenen** Karte
	// (#1622).
	karten *kartennamendoppel

	sammlung *sammlungsbestandsdoppel

	bau      *service.Deckbau
	pruefung *service.Formatpruefung
	abgleich *service.Habenabgleich

	deck   domain.Deck
	format domain.Format
}

// neuesPruefwerk baut ein Deck mit Format, bereit zum Legen.
func neuesPruefwerk(t *testing.T) *pruefwerk {
	t.Helper()

	werk := &pruefwerk{
		decks:   &deckdoppel{},
		formate: &formatdoppel{},
		drucke: &zulaessigkeitsdoppel{
			legalitaeten: map[string]map[string]string{},
			typzeilen:    map[string]string{},
			namen:        map[string]string{},
			manawerte:    map[string]int{},
			orakel:       map[string]string{},
			unbekannt:    map[string]bool{},
		},
		verweise: neuesVerweisdoppel(),
		karten: &kartennamendoppel{
			namen: map[int64]string{}, felder: map[string]string{},
		},
		sammlung: &sammlungsbestandsdoppel{je: map[string][]*domain.Exemplar{}, jeZeile: map[int64][]*domain.Exemplar{}},
	}

	// **Mit den Drucknamen** (U10, #1421): dieselbe Antwort trägt Name,
	// Kosten, Typzeile und Manawert — ohne sie bliebe jeder Platz namenlos,
	// und ein Lauf über die Typzeile prüfte nichts.
	werk.bau = service.NeuerDeckbau(werk.decks, &kartenspieldoppel{spiel: "magic"},
		werk.formate).MitDrucknamen(werk.drucke)
	werk.pruefung = service.NeueFormatpruefung(werk.decks, werk.formate,
		werk.drucke, werk.verweise).MitKartenfeldern(werk.karten)
	werk.abgleich = service.NeuerHabenabgleich(werk.decks, werk.sammlung, werk.verweise)

	format, err := werk.formate.Anlegen(t.Context(), domain.Format{
		Name: "Standard", Spiel: "magic", Mindestgroesse: 60,
		Sideboardgroesse: 15, HoechstzahlGleich: 4, Scryfallname: "standard",
	})
	if err != nil {
		t.Fatalf("das Format: %v", err)
	}

	werk.format = format
	werk.deck = deckAnlegenImDoppel(t, werk.bau)

	gestellt, err := werk.bau.FormatWaehlen(t.Context(), werk.deck.ID, format.ID)
	if err != nil {
		t.Fatalf("das Format wählen: %v", err)
	}

	werk.deck = gestellt

	return werk
}

// eigeneKarte ist die Kennung, unter der die Läufe eine eigene Karte legen.
const eigeneKarte = 7

// legeDruck legt einen echten Druck ins Hauptdeck.
//
// **Ins Hauptdeck und nicht wahlweise**: was das Sideboard angeht, prüft
// `internal/domain` — hier geht es um die Antwort je Karte, und die hängt
// nicht am Bereich.
func (p *pruefwerk) legeDruck(t *testing.T, printing string, menge int) {
	t.Helper()

	if _, err := p.bau.PlatzLegen(t.Context(), p.deck.ID, domain.Deckplatz{
		Printing: printing, Menge: menge, Bereich: domain.BereichHauptdeck,
	}); err != nil {
		t.Fatalf("legen: %v", err)
	}
}

// legeKarte legt eine eigene Karte ins Hauptdeck.
func (p *pruefwerk) legeKarte(t *testing.T, menge int) {
	t.Helper()

	if _, err := p.bau.PlatzLegen(t.Context(), p.deck.ID, domain.Deckplatz{
		KarteID: eigeneKarte, Menge: menge, Bereich: domain.BereichHauptdeck,
	}); err != nil {
		t.Fatalf("legen: %v", err)
	}
}

// formatOhneScryfallnamen nimmt dem Format seine Brücke zu Scryfall.
func (p *pruefwerk) formatOhneScryfallnamen(t *testing.T) {
	t.Helper()

	p.format.Scryfallname = ""

	if _, err := p.formate.Aendern(t.Context(), p.format); err != nil {
		t.Fatalf("das Format ändern: %v", err)
	}
}

// formatHerunter nimmt dem Deck sein Format.
func (p *pruefwerk) formatHerunter(t *testing.T) {
	t.Helper()

	if _, err := p.bau.FormatWaehlen(t.Context(), p.deck.ID, 0); err != nil {
		t.Fatalf("herunterstellen: %v", err)
	}
}

// sammlungsbestandsdoppel gibt die Exemplare eines Drucks.
type sammlungsbestandsdoppel struct {
	je     map[string][]*domain.Exemplar
	fehler error

	// jeZeile sind die Exemplare je Druckzeile ohne Alias (U9, #1420).
	jeZeile      map[int64][]*domain.Exemplar
	zeilenfehler error
}

// Einzeln gibt ein Exemplar nach seiner Kennung — wie der Bestand (U9, #1420).
func (s *sammlungsbestandsdoppel) Einzeln(_ context.Context, id int64) (*domain.Exemplar, error) {
	if s.fehler != nil {
		return nil, s.fehler
	}

	for _, zeilen := range s.je {
		for _, exemplar := range zeilen {
			if exemplar.ID == id {
				return exemplar, nil
			}
		}
	}

	for _, zeilen := range s.jeZeile {
		for _, exemplar := range zeilen {
			if exemplar.ID == id {
				return exemplar, nil
			}
		}
	}

	return nil, fmt.Errorf("%w: Exemplar %d gibt es nicht", domain.ErrNotFound, id)
}

func (s *sammlungsbestandsdoppel) NachDruck(_ context.Context, druckID int64) (
	[]*domain.Exemplar, error,
) {
	if s.zeilenfehler != nil {
		return nil, s.zeilenfehler
	}

	return s.jeZeile[druckID], nil
}

func (s *sammlungsbestandsdoppel) Fuer(_ context.Context, printing string) (
	[]*domain.Exemplar, error,
) {
	if s.fehler != nil {
		return nil, s.fehler
	}

	return s.je[printing], nil
}

// besitze legt Exemplare eines Drucks in die Sammlung.
//
//nolint:unparam // druckB besitzt heute niemand — die Naht bleibt trotzdem am Druck, sonst wäre sie eine über den Bestand
func (p *pruefwerk) besitze(printing string, menge int) {
	p.sammlung.je[printing] = append(p.sammlung.je[printing],
		&domain.Exemplar{Printing: printing, Menge: menge})
}

// wuensche legt einen **Wunsch** ab — und der ist kein Besitz (0057).
//
//nolint:unparam // wie bei besitze: die Naht bleibt am Druck, auch wenn heute nur druckA gewünscht wird
func (p *pruefwerk) wuensche(printing string, menge int) {
	p.sammlung.je[printing] = append(p.sammlung.je[printing],
		&domain.Exemplar{Printing: printing, Menge: menge, Wunsch: true})
}

// legeDruckIns legt einen Druck in einen genannten Bereich.
func (p *pruefwerk) legeDruckIns(
	t *testing.T, printing string, menge int, bereich domain.Deckbereich,
) {
	t.Helper()

	if _, err := p.bau.PlatzLegen(t.Context(), p.deck.ID, domain.Deckplatz{
		Printing: printing, Menge: menge, Bereich: bereich,
	}); err != nil {
		t.Fatalf("legen: %v", err)
	}
}
