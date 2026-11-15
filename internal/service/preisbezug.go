package service

// **Preise aus einem Bulkbezug** — Werkstück A3 von Phase 5.3
// (#1442).
//
// # Warum ein eigener Anwendungsfall neben [Preispflege.Holen]
//
// `Holen` fragt **einen** Druck bei einer Quelle, die je Druck antwortet.
// MTGJSON antwortet nicht je Druck: es schickt **eine** Datei mit allem, und
// was daraus gebraucht wird, sagt der Bestand. Das ist ein anderer Ablauf,
// eine andere Grenze und eine andere Kostenklasse — ein `Holen`, das beides
// könnte, wäre eine Naht mit zwei Bedeutungen.
//
// # Was hier **nicht** geschieht
//
// **Es wird nichts gewählt.** Welche Quelle gefragt wird, sagt der Aufrufer,
// und es gibt keine Vorgabe: „Quellenwahl explizit, keine automatische
// Preisstrategie" steht so im Issue
// ([0093](../../docs/entscheidungen/0093-eine-preisreihe-hat-sieben-teile.md)).

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Bulkpreisquelle holt eine ganze Preisdatei.
//
// **Der Crosswalk geht hinein und nicht heraus**: die Datei führt jede Karte,
// die es gibt; gelesen wird, was der Bestand kennt. Wer sie erst ganz läse und
// danach filterte, legte ein Gigabyte in den Arbeitsspeicher.
type Bulkpreisquelle interface {
	Holen(ctx context.Context, umfang string, crosswalk map[string]string) (
		[]domain.Preisstand, []domain.Befund, error)

	// Name ist die **Familie**, unter der alles von dieser Quelle läuft —
	// siehe [domain.Preisquelle.GehoertZu].
	Name() domain.Preisquelle
}

// Krosswalkquelle sagt, welche fremde Kennung zu welchem Druck gehört.
//
// **Als Schnittstelle** (§2): die Antwort kommt aus `storage`, und dieser
// Anwendungsfall soll ohne Bestand prüfbar bleiben.
type Krosswalkquelle interface {
	Krosswalk(ctx context.Context) (map[string]string, error)
}

// Bulkbilanz ist, was aus einem Bezug wurde.
type Bulkbilanz struct {
	// Drucke ist die Zahl der Drucke, nach denen gefragt wurde.
	Drucke int

	// Staende ist die Zahl der abgelegten Preisstände.
	Staende int

	// Befunde sind die Angaben, die nicht zu deuten waren (§8).
	Befunde []domain.Befund
}

// Preisbezug holt eine ganze Preisdatei und legt ab, was darin steht.
type Preisbezug struct {
	ablage    Preisablage
	quelle    Bulkpreisquelle
	krosswalk Krosswalkquelle
	jetzt     func() time.Time
}

// NeuerPreisbezug baut ihn.
func NeuerPreisbezug(
	ablage Preisablage, quelle Bulkpreisquelle,
	krosswalk Krosswalkquelle, jetzt func() time.Time,
) *Preisbezug {
	return &Preisbezug{ablage: ablage, quelle: quelle, krosswalk: krosswalk, jetzt: jetzt}
}

// Holen fragt die Quelle und legt ab, was sie zu bekannten Drucken nennt.
//
// **Ohne Crosswalk wird nicht gefragt** (§16, Satz 4). Die Datei wiegt bis zu
// einem Gigabyte; wer keine Karte besitzt, die MTGJSON kennt, hat nichts davon
// — und ein Abruf wäre Last für den Anbieter und Wartezeit für niemanden.
//
// **Der Abrufzeitpunkt kommt von hier, der Beobachtungstag von der Quelle**
// (0093): wann gefragt wurde, weiß der Dienst; welchen Tag die Quelle meint,
// nur sie.
func (p *Preisbezug) Holen(ctx context.Context, umfang string) (Bulkbilanz, error) {
	karte, fehler := p.krosswalk.Krosswalk(ctx)
	if fehler != nil {
		return Bulkbilanz{}, fehler
	}

	if len(karte) == 0 {
		return Bulkbilanz{}, nil
	}

	staende, befunde, holfehler := p.quelle.Holen(ctx, umfang, karte)
	if holfehler != nil {
		return Bulkbilanz{}, holfehler
	}

	erhoben := p.jetzt()
	for i := range staende {
		staende[i].Erhoben = erhoben
	}

	// **Alle oder keiner** — dieselbe Zusage wie bei [Preispflege.Holen]: ein
	// halber Bezug wäre ein Stand, den es so nie gab, und niemand könnte ihn
	// als halb erkennen.
	if fehler := p.ablage.Eintragen(ctx, staende...); fehler != nil {
		return Bulkbilanz{}, fehler
	}

	return Bulkbilanz{
		Drucke: len(karte), Staende: len(staende), Befunde: befunde,
	}, nil
}
