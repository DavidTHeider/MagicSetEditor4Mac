package service

// **Dieselbe Filterrechnung wird nicht zweimal gemacht** —
// #1288.

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// packspeicher hält die zuletzt gefilterten Karten je Set.
//
// # Warum
//
// Das Filtern kostet, was alle Karten des Sets kosten: **3,1 s** bei 107
// Karten im Fenster, gemessen am 2026-08-29 an der laufenden Brücke. Die
// Oberfläche ruft erst `Packarten` und dann beim Klick `PackZiehen` — beide
// gehen durch dieselbe Rechnung.
//
// **MSE2 rechnet auch nur einmal**: `PackGenerator` baut seine
// `PackInstance`-Liste beim Aufsetzen und behält sie, bis sich das Set ändert
// (`src/data/pack.cpp`, Fassung f9d9356 vom 2021-07-06).
//
// **Gemerkt werden die gefilterten Karten und nicht die Zahlen.** Ein Speicher
// über Zahlen bediente `Angebot` und ließe `Ziehen` von vorn anfangen —
// angestrichen vom Nachlauf über Phase 4.1, Befund 5.3.
//
// # Warum der Schlüssel der Inhalt ist und kein Zähler
//
// **Weil ein Zähler eine Ungültigkeitsregel braucht und ein Inhalt keine.**
// Ändert sich eine Karte, eine Packart oder eine beteiligte Vorlage, ändert
// sich der Schlüssel — es gibt keinen Stand, den jemand zu erhöhen vergessen
// kann.
//
// **Was er kostet, ist gemessen und nicht geschätzt**: 2 ms für die Vorlage
// und unter einer Millisekunde für 107 Karten, einmal erhoben am 2026-08-29
// an `Vanguard.mse-set` über die gebaute Kommandozeile, gegen rund 3.300 ms
// Rechnung. **Eine Messung, ein Rechner, ein Set** — wer daraus „er kostet
// nichts" macht, sagt mehr, als dasteht (Nachlauf über Phase 4.1, Befund 4).
//
// # Was der Schlüssel **nicht** umfasst
//
//   - **Schriften und Bilder.** Sie hängen am Composition Root und ändern sich
//     in einer Sitzung nicht. Ein ausgetauschtes Bild änderte theoretisch die
//     Inhaltsmessung und damit `card_style`; **im vollen Pack liest kein
//     einziger der zwölf Packfilter `card_style`, eine Schrift oder ein
//     Bildmaß** — nachgesehen am 2026-08-29 an `622160f9`.
//   - **Die Uhrzeit.** MSE-Script hat keine Zeitfunktion.
//
// Was er dagegen **umfasst**, seit dem Nachlauf über Phase 4.1: die Kennung
// und das Set jeder Karte, jeden Feldwert **längenpräfigiert**, und die
// Vorlage jedes beteiligten Stilbogens — auch die einer Karte mit eigenem.
type packspeicher struct {
	mu        sync.Mutex
	eintraege map[int64]packeintrag
}

// packeintrag ist eine gerechnete Filterung samt ihrem Schlüssel.
type packeintrag struct {
	schluessel [sha256.Size]byte
	gefiltert  map[string][]*domain.CardRow
	befunde    []domain.Befund
}

// GrenzePackspeicher ist die Zahl der Sets, deren Filterung vorgehalten wird.
//
// **Acht, und die Zahl ist eine Aussage über Fenster und nicht über Speicher**:
// so viele Sets hat niemand gleichzeitig offen. Ein Eintrag hält Zeiger auf
// Karten, die ohnehin schon dastehen.
const GrenzePackspeicher = 8

// hole gibt eine gerechnete Filterung, wenn der Schlüssel noch stimmt.
func (s *packspeicher) hole(setID int64, schluessel [sha256.Size]byte) (
	map[string][]*domain.CardRow, []domain.Befund, bool,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	eintrag, da := s.eintraege[setID]
	if !da || eintrag.schluessel != schluessel {
		return nil, nil, false
	}

	return eintrag.gefiltert, eintrag.befunde, true
}

// lege merkt eine Filterung und räumt an der Grenze.
//
// **Sie räumt ganz, nicht klug** — dieselbe Wahl wie beim Baumspeicher
// (`msescript.GrenzeBaumspeicher`), und aus demselben Grund: wer acht Sets
// zugleich offen hat, hat keinen Arbeitssatz mehr, den eine Strategie rettete.
func (s *packspeicher) lege(
	setID int64, schluessel [sha256.Size]byte,
	gefiltert map[string][]*domain.CardRow, befunde []domain.Befund,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.eintraege == nil || len(s.eintraege) >= GrenzePackspeicher {
		s.eintraege = make(map[int64]packeintrag, GrenzePackspeicher)
	}

	s.eintraege[setID] = packeintrag{
		schluessel: schluessel, gefiltert: gefiltert, befunde: befunde,
	}
}

// packschluessel baut den Inhaltsschlüssel einer Filterung.
//
// # Warum jede Länge davorsteht
//
// **Weil ein Trennzeichen in einem Kartenwert stehen darf.** Der erste Anlauf
// schrieb `Feld=Text\n`, und ein Regeltext mit einem Zeilenumbruch darin
// erzeugte dieselben Bytes wie zwei andere Felder — **eine Kollision vor
// SHA-256**, kein Hashzufall. Zwei verschieden verteilte Inhalte bekamen
// denselben Schlüssel, und der zweite hätte das Angebot des ersten geerbt.
// Angestrichen vom Nachlauf über Phase 4.1, Befund 5.1.
//
// Mit der Länge davor ist die Zeichenkette **eindeutig zerlegbar**: `6:rarity`
// kann nur `rarity` heißen, gleich was daneben steht.
//
// # Warum die Vorlage jedes Stilbogens dazugehört
//
// Weil eine Karte einen **eigenen** tragen darf ([Stilquelle.StilFuerKarte]).
// Ihr Initskript entscheidet mit über den Filter; ohne sie im Schlüssel bliebe
// eine Filterung stehen, obwohl sich der Stil dieser Karte geändert hat — auch
// das aus dem Nachlauf.
//
// **Nachgeschlagen wird nur für Karten, die einen eigenen Stil nennen**, und je
// Name einmal. In einem gewöhnlichen Set ist das keine einzige.
func (p *Packbereich) packschluessel(
	ctx context.Context,
	arten []msepaket.Pack, karten []*domain.CardRow, vorlage Vorlage,
) [sha256.Size]byte {
	// **Über einen Puffer und nicht über `Fprintf` in den Hasher.** `hash.Hash`
	// gibt beim Schreiben nie einen Fehler zurück — der Linter weiß das nicht,
	// und drei `//nolint` wären drei Ausnahmen für eine Zusage, die die
	// Bibliothek ausdrücklich gibt.
	var gebaut strings.Builder

	laengeVoran(&gebaut, fmt.Sprintf("%v", vorlage))
	laengeVoran(&gebaut, fmt.Sprintf("%v", arten))

	eigene := map[string]bool{}

	for _, karte := range karten {
		fmt.Fprintf(&gebaut, "%d/%d/%d\n", karte.ID, karte.SetID, karte.Number)

		for _, wert := range karte.Values {
			laengeVoran(&gebaut, wert.Field)
			laengeVoran(&gebaut, wert.Text)
		}

		if name, _ := karte.Value(Kartenstilfeld); strings.TrimSpace(name) != "" {
			eigene[name] = true
		}
	}

	p.eigeneStileHinein(ctx, &gebaut, karten, eigene)

	return sha256.Sum256([]byte(gebaut.String()))
}

// eigeneStileHinein schreibt die Vorlage jedes eigenen Stilbogens dazu.
func (p *Packbereich) eigeneStileHinein(
	ctx context.Context, gebaut *strings.Builder,
	karten []*domain.CardRow, eigene map[string]bool,
) {
	if len(eigene) == 0 || p.planer == nil || p.planer.stile == nil {
		return
	}

	for _, karte := range karten {
		name, _ := karte.Value(Kartenstilfeld)
		if !eigene[name] {
			continue
		}

		eigene[name] = false

		vorlage, _, da := p.planer.stile.StilFuerKarte(ctx, karte)
		if !da {
			continue
		}

		laengeVoran(gebaut, name)
		laengeVoran(gebaut, fmt.Sprintf("%v", vorlage))
	}
}

// laengeVoran schreibt eine Zeichenkette mit ihrer Länge davor.
func laengeVoran(gebaut *strings.Builder, text string) {
	fmt.Fprintf(gebaut, "%d:%s", len(text), text)
}
