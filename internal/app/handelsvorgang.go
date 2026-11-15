package app

// **Ein Handelsvorgang am Fenster** —
// #1637.
//
// # Warum eine eigene Gestalt und nicht [domain.Handelsangebot]
//
// **Weil der Abdruck dazugehört und im Verbund nicht steht.** Er ist die
// Kennung einer Revision — eine Zustimmung nennt ihn, statt „ja" zu sagen —,
// und er wird **gerechnet**. Wer das Angebot roh hinausreichte, ließe jede
// Fläche ihn selbst ausrechnen; zwei Rechnungen derselben Art laufen beim
// ersten neuen Feld auseinander.
//
// **Und weil Zeiten und Aufzählungen hier zu Zeichenketten werden** — die
// Wails-Naht ist innen, jede andere Fläche baut ihre eigene Sicht
// ([0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// HandelspositionDTO ist eine Zeile eines Angebots.
type HandelspositionDTO struct {
	// Von ist die Partei, deren Bestand die Position verlässt.
	Von string `json:"von"`

	Printing string `json:"printing"`
	DruckID  int64  `json:"druckId"`
	Menge    int    `json:"menge"`

	// Guete ist der Zustand des Exemplars — `NM`, `LP`, …
	Guete   string `json:"guete"`
	Sprache string `json:"sprache"`
	Finish  string `json:"finish"`

	// Abbild ist der Abdruck des Exemplarstands beim Anbieten.
	Abbild string `json:"abbild"`

	// Bewertung ist der eingefrorene Wert in der kleinsten Währungseinheit —
	// sie bindet nur mit [HandelsvorgangDTO.BewertungGeteilt].
	Bewertung int64 `json:"bewertung"`
}

// HandelsvorgangDTO ist **eine** Revision, wie das Fenster sie sieht.
type HandelsvorgangDTO struct {
	Vorgang  string `json:"vorgang"`
	Revision int    `json:"revision"`

	// Eltern ist der Abdruck der vorigen Revision — leer in der ersten.
	Eltern string `json:"eltern"`

	// Abdruck ist die Kennung **dieser** Revision. Er wird gerechnet und
	// nicht gespeichert.
	Abdruck string `json:"abdruck"`

	Art       string `json:"art"`
	Initiator string `json:"initiator"`

	Parteien []string `json:"parteien"`

	// Ich ist die Partei **dieses** Profils — die, für die Lage, Naechster
	// und Moegliche gelten, und die, deren Stimme `naechster` nennt
	// (#2026).
	// Leer, wo es keine Identität gibt.
	//
	// **Gesetzt und nicht gefolgert**: ein Fenster, das sie aus den
	// Parteien herausrechnete, nähme „die, die nicht das Gegenüber ist" —
	// eine zweite Fassung derselben Aussage, und bei mehr als zwei Parteien
	// eine falsche. Ohne sie wusste das Fenster, **dass** seine Stimme zum
	// Abschluss fehlt, und konnte sie nicht setzen.
	Ich string `json:"ich"`

	Zustand string `json:"zustand"`

	Positionen       []HandelspositionDTO `json:"positionen"`
	BewertungGeteilt bool                 `json:"bewertungGeteilt"`

	// Zustimmungen sind die Parteien, die **dieser** Revision zugestimmt
	// haben.
	Zustimmungen []string `json:"zustimmungen"`

	// Ende sagt, ob aus diesem Zustand nichts mehr folgt.
	//
	// **Gerechnet und nicht geraten**: die vier Enden stehen in
	// [domain.HandelEnde], und eine Fläche, die sie nachbaute, hätte die
	// zweite Fassung derselben Aussage.
	Ende bool `json:"ende"`

	// Beschluss ist der Schritt, dem die Zustimmungen gelten — leer, wenn
	// keiner offen ist (TRADE-03, #1717). **Gerechnet**, aus demselben Grund
	// wie `Ende`.
	Beschluss string `json:"beschluss"`

	// Lage, Naechster, Moegliche und Phase sagen, wo der Vorgang für
	// **dieses** Profil steht und was es jetzt tun darf (TRADE-02.AK-02,
	// #1716). Sie kommen aus `domain` — eine Fläche, die sie nachrechnete,
	// hätte eine zweite Fassung der Matrix.
	Lage      string   `json:"lage"`
	Naechster string   `json:"naechster"`
	Moegliche []string `json:"moegliche"`
	Phase     string   `json:"phase"`
}

// HandelsvorgangAntwort trägt den Vorgang samt seinem Fehler.
type HandelsvorgangAntwort struct {
	Vorgang HandelsvorgangDTO `json:"vorgang"`
	Fehler  *Fehler           `json:"fehler"`
}

// HandelsvorgaengeAntwort sind die Handel mit einem Gegenüber — UX-08.AK-04
// (#1680).
// **Auch eine Absage trägt eine Liste**, eine leere.
type HandelsvorgaengeAntwort struct {
	Vorgaenge []HandelsvorgangDTO `json:"vorgaenge"`
	Fehler    *Fehler             `json:"fehler"`
}

// Handelswunsch ist, womit ein Vorgang eröffnet wird.
//
// # Warum er das Gegenüber nennt und nicht die Parteien
//
// **Weil der Initiator dieser Rechner ist und nicht das, was jemand
// hinschreibt.** Er kommt aus dem Ausweis; wer ihn im Wunsch nennen dürfte,
// eröffnete einen Vorgang im Namen eines anderen — und die Sequenz eines
// Handels hängt genau an dieser Stelle ([domain.Handelsangebot.Fortgeschrieben]).
type Handelswunsch struct {
	// Vorgang ist die Kennung des Handels über alle Revisionen.
	Vorgang string `json:"vorgang"`

	// Art ist `teilen`, `deckleihe`, `tausch` oder `markt`.
	Art string `json:"art"`

	// Gegenueber sind die anderen Parteien — mindestens eine.
	Gegenueber []string `json:"gegenueber"`

	Positionen       []HandelspositionDTO `json:"positionen"`
	BewertungGeteilt bool                 `json:"bewertungGeteilt"`
}

// alsHandelsvorgangDTO macht aus einer Revision ihre Gestalt fürs Fenster.
func alsHandelsvorgangDTO(angebot domain.Handelsangebot, ich string) HandelsvorgangDTO {
	moegliche := make([]string, 0, 4)
	for _, schritt := range angebot.Moegliche(ich) {
		moegliche = append(moegliche, string(schritt))
	}

	return HandelsvorgangDTO{
		Vorgang: angebot.Vorgang, Revision: angebot.Revision,
		Eltern: angebot.Eltern, Abdruck: angebot.Abdruck(),
		Art: string(angebot.Art), Initiator: angebot.Initiator,
		Parteien: alsListe(angebot.Parteien), Ich: ich,
		Zustand: string(angebot.Zustand),
		// **Eine leere Liste und kein `null`** — sonst fällt aus der Antwort
		// `null` heraus, wo eine Fläche eine Liste erwartet (§3).
		Positionen:       alsPositionsDTOs(angebot.Positionen),
		BewertungGeteilt: angebot.BewertungGeteilt,
		Zustimmungen:     alsListe(angebot.Zustimmungen),
		Ende:             domain.HandelEnde(angebot.Zustand),
		Beschluss:        string(angebot.OffenerBeschluss()),
		Lage:             string(angebot.Lage(ich)),
		Naechster:        string(angebot.Naechster(ich)),
		Moegliche:        moegliche,
		Phase:            string(angebot.Zustand.Phase()),
	}
}

// leererHandelsvorgang ist die Gestalt, die ein Fehlerweg zurückgibt.
//
// **Mit leeren Listen und nicht mit `nil`.** Eine Absage ist eine Antwort,
// und sie geht durch dieselbe Serialisierung wie ein Erfolg; wer sie als
// Nullwert zurückgäbe, schickte dem Fenster `null` statt einer leeren Liste —
// und dort sieht das aus wie „konnte nicht beantwortet werden" statt wie
// „nichts da" (§3, gehalten von `TestKeineBindungGibtEineListeAlsNull`).
func leererHandelsvorgang() HandelsvorgangDTO {
	return HandelsvorgangDTO{
		Parteien: []string{}, Positionen: []HandelspositionDTO{},
		Zustimmungen: []string{}, Moegliche: []string{},
	}
}

// handelsabsage gibt die Antwort eines Fehlerwegs.
func handelsabsage(ctx context.Context, fehler error) HandelsvorgangAntwort {
	return HandelsvorgangAntwort{
		Vorgang: leererHandelsvorgang(), Fehler: Uebersetze(ctx, fehler),
	}
}

// alsPositionsDTOs macht aus den Zeilen ihre Gestalt fürs Fenster.
func alsPositionsDTOs(
	positionen []domain.Handelsposition,
) []HandelspositionDTO {
	dtos := make([]HandelspositionDTO, 0, len(positionen))

	for _, position := range positionen {
		dtos = append(dtos, HandelspositionDTO{
			Von: position.Von, Printing: position.Printing,
			DruckID: position.DruckID, Menge: position.Menge,
			Guete: position.Guete, Sprache: position.Sprache,
			Finish: position.Finish, Abbild: position.Abbild,
			Bewertung: position.Bewertung,
		})
	}

	return dtos
}

// alsPositionen macht aus der Gestalt des Fensters die Zeilen.
func alsPositionen(
	dtos []HandelspositionDTO,
) []domain.Handelsposition {
	positionen := make([]domain.Handelsposition, 0, len(dtos))

	for _, dto := range dtos {
		positionen = append(positionen, domain.Handelsposition{
			Von: dto.Von, Printing: dto.Printing,
			DruckID: dto.DruckID, Menge: dto.Menge,
			Guete: dto.Guete, Sprache: dto.Sprache,
			Finish: dto.Finish, Abbild: dto.Abbild,
			Bewertung: dto.Bewertung,
		})
	}

	return positionen
}
