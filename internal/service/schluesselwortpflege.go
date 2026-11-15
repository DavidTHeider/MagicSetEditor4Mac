// **Schlüsselwörter sehen und pflegen** —
// #1062.
//
// # Die Liste zeigt beides
//
// Die des **Sets** und die des **Spiels**, und der Unterschied ist der halbe
// Zweck: die des Spiels sind unveränderlich, die des Sets nicht. MSE hält ihn
// über `Keyword::fixed` (`src/data/keyword.hpp`, Zeile 104) und zeigt die
// festen grau (`KeywordList::OnGetItemAttr`, `src/gui/control/keyword_list.cpp`,
// Zeilen 225–235). Fassung f9d9356 vom 2021-07-06.
//
// # Und sie sagt, welches nicht greift
//
//	valid = !match_re.matches(_(""));
//
// Ein Schlüsselwort, dessen Muster auf den leeren Text passt, ließe MSE hängen
// (`Keyword::prepare`, Zeilen 205–209). MSE zeigt ein solches **rot**; hier
// steht dieselbe Auskunft an [Schluesselwortsicht.Gueltig].

package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Schluesselwortablage ist, was die Pflege zum Lesen und Schreiben braucht.
type Schluesselwortablage interface {
	Schluesselwoerter(ctx context.Context, setID int64) ([]*domain.Keyword, error)

	SchluesselwortAnlegen(ctx context.Context, setID int64, modus string) (int64, error)
	SchluesselwortSetzen(ctx context.Context, setID, wortID int64, feld, wert string) error
	SchluesselwortLoeschen(ctx context.Context, setID, wortID int64) error
}

// Setstilquelle gibt den Stil eines Sets — und damit sein Spiel.
//
// **Schmaler als [Stilquelle]**, und das ist §2: die Schnittstelle steht dort,
// wo sie gebraucht wird. Die Pflege fragt nie nach dem Stil einer **Karte**;
// wer die volle verlangte, schlösse jeden Aufrufer aus, der sie nicht hat.
type Setstilquelle interface {
	StilFuerSet(ctx context.Context, setID int64) (Vorlage, []domain.Befund, bool)
}

// Schluesselwortsicht ist ein Schlüsselwort, wie die Oberfläche es zeigt.
type Schluesselwortsicht struct {
	// ID ist die Kennung im Bestand — **0 bei einem des Spiels**: das gehört
	// keinem Set und lässt sich nicht ändern.
	ID int64

	Wort       string
	Muster     string
	Erinnerung string
	Regeln     string
	Modus      string

	// Fest sagt, dass es aus dem **Spiel** kommt.
	Fest bool

	// Gueltig sagt, dass sein Muster greift.
	Gueltig bool
}

// Schluesselwortliste ist, was eine Seite braucht, um sie zu zeigen.
type Schluesselwortliste struct {
	Woerter []Schluesselwortsicht

	// Modi sind die Namen der `keyword mode:`-Blöcke des Spiels — die Auswahl
	// eines Schlüsselworts.
	Modi []string
}

// Schluesselwortpflege zeigt und pflegt die Schlüsselwörter eines Sets.
type Schluesselwortpflege struct {
	ablage Schluesselwortablage
	stile  Setstilquelle
}

// NeueSchluesselwortpflege baut sie.
func NeueSchluesselwortpflege(
	ablage Schluesselwortablage, stile Setstilquelle,
) *Schluesselwortpflege {
	return &Schluesselwortpflege{ablage: ablage, stile: stile}
}

// Liste gibt die Schlüsselwörter des Sets **und** die des Spiels.
//
// **Die des Sets zuerst** — so ordnet MSE sie (`KeywordList::getItems`,
// Zeilen 153–165), und die Reihenfolge ist keine Zier: sie entscheidet, welches
// zuerst auf einen Text passt.
func (p *Schluesselwortpflege) Liste(ctx context.Context, setID int64) (
	Schluesselwortliste, error,
) {
	eigene, err := p.ablage.Schluesselwoerter(ctx, setID)
	if err != nil {
		return Schluesselwortliste{}, err
	}

	vorlage := p.spielvorlage(ctx, setID)

	woerter := make([]Schluesselwortsicht, 0, len(eigene)+len(vorlage.Rohwoerter))

	for _, wort := range eigene {
		woerter = append(woerter, Schluesselwortsicht{
			ID: wort.ID, Wort: wort.Word, Muster: wort.Match,
			Erinnerung: wort.Reminder, Regeln: wort.Rules, Modus: wort.Mode,
			Gueltig: musterGreift(wort.Match, vorlage.Parameterarten),
		})
	}

	for _, wort := range vorlage.Rohwoerter {
		woerter = append(woerter, Schluesselwortsicht{
			Wort: wort.Wort, Muster: wort.Muster,
			Erinnerung: wort.Erinnerung, Regeln: wort.Regeln, Modus: wort.Modus,
			Fest:    true,
			Gueltig: musterGreift(wort.Muster, vorlage.Parameterarten),
		})
	}

	return Schluesselwortliste{Woerter: woerter, Modi: modusnamen(vorlage.Modi)}, nil
}

// Anlegen hängt ein leeres Schlüsselwort ans Ende.
//
// **Mit dem Vorgabemodus des Spiels** — `AddKeywordAction`, Zeilen 20–32:
// MSE sucht den Modus mit `is default: true` und setzt sonst nichts.
func (p *Schluesselwortpflege) Anlegen(ctx context.Context, setID int64) (int64, error) {
	return p.ablage.SchluesselwortAnlegen(ctx, setID, vorgabemodus(p.spielvorlage(ctx, setID)))
}

// Setzen schreibt ein Feld eines Schlüsselworts des Sets.
func (p *Schluesselwortpflege) Setzen(
	ctx context.Context, setID, wortID int64, feld, wert string,
) error {
	return p.ablage.SchluesselwortSetzen(ctx, setID, wortID, feld, wert)
}

// Loeschen nimmt eines weg.
func (p *Schluesselwortpflege) Loeschen(ctx context.Context, setID, wortID int64) error {
	return p.ablage.SchluesselwortLoeschen(ctx, setID, wortID)
}

// spielvorlage gibt die Vorlage des Sets — leer, wenn es keine gibt.
//
// **Ohne Stilpaket bleiben die des Sets**, und das ist kein Fehler: ein Set,
// dessen Spiel nicht vorliegt, hat trotzdem seine eigenen Schlüsselwörter, und
// sie zu verschweigen wäre die schlechtere Antwort (§8).
func (p *Schluesselwortpflege) spielvorlage(ctx context.Context, setID int64) Vorlage {
	if p.stile == nil {
		return Vorlage{}
	}

	vorlage, _, da := p.stile.StilFuerSet(ctx, setID)
	if !da {
		return Vorlage{}
	}

	return vorlage
}

// musterGreift sagt, ob ein Muster benutzbar ist.
//
// **Über dieselbe Vorbereitung, die auch auf der Karte läuft** ([NeueWortliste]):
// eine zweite Prüfung daneben liefe beim ersten Sonderfall auseinander. Was
// dort herausfällt, ist genau das, was MSE abschaltet.
func musterGreift(muster string, arten []msepaket.Schluesselwortparameterart) bool {
	if muster == "" {
		return false
	}

	liste, _ := NeueWortliste([]msepaket.Schluesselwort{{Muster: muster}}, arten)

	return liste.Anzahl() == 1
}

// vorgabemodus gibt den Modus mit `is default: true` — sonst nichts.
func vorgabemodus(vorlage Vorlage) string {
	for _, modus := range vorlage.Modi {
		if modus.Vorgabe {
			return modus.Name
		}
	}

	return ""
}

// modusnamen gibt die Namen der Modi.
func modusnamen(modi []msepaket.Schluesselwortmodus) []string {
	namen := make([]string, 0, len(modi))
	for _, modus := range modi {
		namen = append(namen, modus.Name)
	}

	return namen
}
