package service

// **Der Korpus aus einer Scryfall-Sammeldatei** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Warum ein Abschnitt je Seite und nicht je Karte
//
// **Weil eine zweiseitige Karte zwei Regeltexte hat.** Das Issue verlangt
// „Karten-/Regelchunks an fachlichen Grenzen" und „Parent-/Face-/Regelbezüge
// bleiben rekonstruierbar". Ein gemeinsamer Vektor über Vorder- und
// Rückseite läge zwischen beiden und neben keiner; die Bezüge bleiben über
// die **Orakelkennung**, die beide Seiten teilen.
//
// # Und warum die Identität aus A2 kommt
//
// **Keine Deduplizierung nach Namen** (Issue). Die Fachkennung ist die
// `oracle_id`, das System ist `scryfall`, die Semantik ist `konzept` — ein
// Oracle-Text gehört zum Konzept und nicht zu einem Druck. Zwei Drucke
// derselben Karte tragen denselben Text und werden deshalb **einmal**
// indiziert.
//
// # Was hier nicht steht
//
// **Rulings, Regelwerk und eigene Inhalte.** Sie sind Teil des
// Pflichtkorpus, kommen aber aus anderen Quellen mit anderen Ständen; jede
// bekommt ihre eigene [Korpusquelle]. Was sie teilen, ist der
// [domain.Retrievaldokument]-Vertrag — und das ist der Sinn des Vertrags.

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// Zerteilfassung nennt, wie Abschnitte gebildet und normalisiert werden.
//
// **Sie steht im Manifest und wird hochgezählt, wenn sich hier etwas
// ändert.** Ein Index, dessen Abschnitte anders geschnitten sind, ist ein
// anderer Index — auch bei gleichem Modell.
const Zerteilfassung = "karten-seiten-1"

// KorpusstapelGroesse ist, wie viele Abschnitte ein Stapel trägt.
//
// **Zweihundertsechsundfünfzig**, und die Zahl ist gewählt: sie hält den
// Speicher klein und den Aufbau flüssig. Eingebettet wird in kleineren
// Teilen ([StapelHoechstens]) — das eine ist die Lesegröße, das andere die
// Auftragsgröße.
const KorpusstapelGroesse = 256

// Scryfallkorpus liest Oracle-Abschnitte aus einer Sammeldatei.
type Scryfallkorpus struct {
	leser *scryfall.Bulkleser
	stand domain.Korpusstand

	// abgerufen ist, wann die Sammeldatei geholt wurde.
	//
	// **Nicht „jetzt", und das ist der Unterschied, den das Issue
	// ausdrücklich nennt**: „Abrufdatum ist nicht Gültigkeitsdatum." Eine
	// Datei von gestern, heute indiziert, wurde gestern abgerufen — und der
	// Dienst hat keine Uhr, die das erfinden dürfte.
	abgerufen time.Time

	// spiel steht in jedem Dokument — eine Sammeldatei von Scryfall trägt
	// Magic und sonst nichts.
	spiel string

	befunde []domain.Befund
	fertig  bool
}

// NeuerScryfallkorpus baut ihn über einem offenen Leser.
//
// **Der Leser kommt von außen.** Wer ihn hier öffnete, machte jeden Test von
// einer Datei abhängig — und `service` von einem Dateisystem.
func NeuerScryfallkorpus(
	leser *scryfall.Bulkleser, stand domain.Korpusstand, abgerufen time.Time,
) *Scryfallkorpus {
	return &Scryfallkorpus{
		leser: leser, stand: stand, abgerufen: abgerufen, spiel: "magic",
	}
}

// Staende nennt den Stand dieser Sammeldatei.
func (s *Scryfallkorpus) Staende(context.Context) ([]domain.Korpusstand, error) {
	if s.stand.Quelle != domain.QuelleOracle {
		return nil, domain.MitGrund(domain.ErrValidation, nil,
			"eine Scryfall-Sammeldatei trägt Oracle-Texte, der Stand nennt %q",
			s.stand.Quelle)
	}

	return []domain.Korpusstand{s.stand}, nil
}

// Befunde sind die Zeilen, die übergangen wurden.
//
// **Sie werden gesammelt und nicht geworfen** (§8): ein Set, von dem 300
// Karten lesbar sind und zwei nicht, wird mit 300 Karten und zwei benannten
// Befunden geöffnet.
func (s *Scryfallkorpus) Befunde() []domain.Befund { return s.befunde }

// Stapel gibt die nächsten Abschnitte.
func (s *Scryfallkorpus) Stapel(ctx context.Context) (
	[]domain.Retrievaldokument, error,
) {
	if s.fertig {
		return nil, nil
	}

	abschnitte := make([]domain.Retrievaldokument, 0, KorpusstapelGroesse)

	for len(abschnitte) < KorpusstapelGroesse {
		if fehler := ctx.Err(); fehler != nil {
			return nil, fehler
		}

		karte, befunde, fehler := s.leser.Naechste()
		if fehler != nil {
			return nil, fehler
		}

		s.befunde = append(s.befunde, befunde...)

		if karte == nil && befunde == nil {
			s.fertig = true

			break
		}

		if karte == nil {
			continue
		}

		abschnitte = append(abschnitte, s.abschnitteVon(karte)...)
	}

	return abschnitte, nil
}

// abschnitteVon macht aus einer Karte ihre Abschnitte.
func (s *Scryfallkorpus) abschnitteVon(karte *scryfall.Karte) []domain.Retrievaldokument {
	if karte.Orakelkennung == "" {
		// **Ohne Orakelkennung keine Identität** — und ohne Identität kein
		// auflösbarer Treffer. Ein Token oder eine Kunstkarte gehört nicht
		// in einen Regelkorpus.
		return nil
	}

	if len(karte.Seiten) == 0 {
		return []domain.Retrievaldokument{s.abschnitt(karte, karte, 0)}
	}

	abschnitte := make([]domain.Retrievaldokument, 0, len(karte.Seiten))

	for stelle := range karte.Seiten {
		abschnitte = append(abschnitte,
			s.abschnitt(karte, karte.MitSeite(stelle), stelle))
	}

	return abschnitte
}

// abschnitt baut einen — `seite` trägt Titel und Text, `karte` die Identität.
func (s *Scryfallkorpus) abschnitt(
	karte, seite *scryfall.Karte, nummer int,
) domain.Retrievaldokument {
	text := abschnittstext(seite)

	return domain.Retrievaldokument{
		Kennung: fmt.Sprintf("scryfall-oracle:%s:%d", karte.Orakelkennung, nummer),
		Quelle:  domain.QuelleOracle,
		// **Die Adresse des Drucks und nicht des Konzepts.** Scryfall gibt
		// nur diese heraus; sie führt zu einem Druck, dessen Text dieser ist.
		Bezugsadresse:  karte.Adresse,
		Spiel:          s.spiel,
		Kennungssystem: domain.SystemScryfall,
		Fachkennung:    karte.Orakelkennung,
		Semantik:       domain.SemantikKonzept,
		Sprache:        karte.Sprache,
		Titel:          seite.Name,
		Text:           text,
		Quellrevision:  s.stand.Revision,
		Quelldatum:     s.stand.Datum,
		Abgerufen:      s.abgerufen,
		Inhaltshash:    domain.Retrievalhash(text),
		Merkmale:       s.merkmaleVon(karte),
		Abschnitt:      nummer,
		Bis:            len(text),
		Sichtbarkeit:   domain.SichtbarOeffentlich,
	}
}

// abschnittstext setzt zusammen, was durchsucht wird.
//
// **Typzeile und Regeltext, kein Aromatext.** Der Aromatext ist Literatur;
// er trägt dieselben Wörter wie hundert andere Karten und zöge jede
// semantische Frage zu sich.
func abschnittstext(seite *scryfall.Karte) string {
	teile := make([]string, 0, 3)

	for _, teil := range []string{seite.Manakosten, seite.Typzeile, seite.Regeltext} {
		if strings.TrimSpace(teil) != "" {
			teile = append(teile, teil)
		}
	}

	return strings.Join(teile, "\n")
}

// merkmaleVon liest die strukturierten Angaben aus der Karte.
//
// **An einer Stelle je Spiel** (§16): wer die Zuordnung über den Code
// verteilte, hätte bei der nächsten Änderung bei Scryfall fünf Stellen.
func (s *Scryfallkorpus) merkmaleVon(karte *scryfall.Karte) domain.Retrievalmerkmale {
	manawert := float64(karte.Manawert)

	merkmale := domain.Retrievalmerkmale{
		Farben:         karte.Farben,
		Farbidentitaet: karte.Farbidentitaet,
		Typen:          typwoerter(karte.Typzeile),
		Manawert:       &manawert,
		Set:            karte.Set,
		Formate:        karte.Legalitaeten,
	}

	if len(karte.Legalitaeten) > 0 {
		// **Der Datenstand der Legalitäten ist der der Sammeldatei.** Eine
		// Bannliste ändert sich, ohne dass sich die Karte ändert; ohne Stand
		// gibt es deshalb gar keine Auskunft
		// ([domain.Retrievalmerkmale.Format]).
		merkmale.Formatstand = s.stand.Revision
	}

	return merkmale
}

// typwoerter zerlegt eine Typzeile in ihre Wörter, klein geschrieben.
//
// **Mit dem Gedankenstrich als Trenner und nicht als Wort.** „Legendary
// Creature — Goblin Shaman" ergibt fünf Wörter, nach denen sich einzeln
// filtern lässt.
func typwoerter(zeile string) []string {
	ersetzt := strings.NewReplacer("—", " ", "-", " ", "//", " ").Replace(zeile)

	woerter := make([]string, 0, 6)

	for _, wort := range strings.Fields(ersetzt) {
		woerter = append(woerter, strings.ToLower(wort))
	}

	if len(woerter) == 0 {
		return nil
	}

	return woerter
}
