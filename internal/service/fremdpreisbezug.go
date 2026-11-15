package service

// **Preisverlauf von einer fremden Quelle** — Werkstück A5 von Phase 5.3
// (#1444).
//
// # Warum ein eigener Anwendungsfall neben [Preisbezug]
//
// **Weil die Frage eine andere ist.** [Preisbezug] holt **eine Datei mit
// allem** und filtert sie am Bestand; hier wird **eine Karte** gefragt, und
// vorher muss feststehen, wie die Quelle sie nennt. Das ist ein anderer
// Ablauf, eine andere Grenze und eine andere Kostenklasse.
//
// # Und warum nicht neben [Marktabruf]
//
// **Weil ein Verlauf kein Angebot ist.** Ein Marktangebot ist ein Posten, den
// es morgen nicht mehr gibt; ein Verlauf ist eine Reihe über die Zeit und
// gehört in die Ablage
// ([0093](../../docs/entscheidungen/0093-eine-preisreihe-hat-sieben-teile.md)).
// Die **Auflösung** ist bei beiden dieselbe Überlegung — und dieselbe Regel
// aus A2: mehrdeutig ist ein Konflikt.

import (
	"context"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Fremdpreisquelle ist, was dieser Anwendungsfall von einer Quelle braucht.
//
// **In `domain`-Begriffen** (§2, §16): die Gestalt eines Anbieters gehört
// nicht in die Fachlogik, und ein Lauf hier soll ohne Netz auskommen.
type Fremdpreisquelle interface {
	// Familie ist der Name, unter dem alles von dieser Quelle läuft.
	Familie() domain.Preisquelle

	// System ist das Kennungssystem, unter dem sie ihre Karten führt.
	System() domain.Kennungssystem

	// Fassungen gibt die Kartenfassungen, die zu einem Druck passen könnten.
	//
	// **Die Quelle sucht, dieser Dienst entscheidet.** Wie gesucht wird —
	// nach Namen, nach Setcode, nach beidem —, weiß nur der Anbieter; was
	// aus mehreren Treffern folgt, ist eine fachliche Regel und steht hier.
	Fassungen(ctx context.Context, druck domain.Druck) ([]domain.Fremdfassung, error)

	// Verlauf gibt die Preisstände einer Fassung, schon auf den Druck bezogen.
	Verlauf(ctx context.Context, fassung, alias string, wunsch domain.Angebotswunsch) (
		[]domain.Preisstand, []domain.Befund, error)
}

// Preiseintrag ist der Teil der Ablage, den dieser Anwendungsfall braucht.
//
// **Eine Methode und nicht vier**: [Preisablage] kann auch lesen, und ein
// Bezug, der den Verlauf abfragen dürfte, könnte sich selbst bestätigen.
type Preiseintrag interface {
	Eintragen(ctx context.Context, staende ...domain.Preisstand) error
}

// Fremdpreisbilanz ist, was ein Bezug ergeben hat.
type Fremdpreisbilanz struct {
	// Fassung ist die Kennung der Karte beim Anbieter.
	Fassung string

	// Zugeordnet sagt, ob sie gerade erst gefunden wurde.
	Zugeordnet bool

	// Staende ist die Zahl der abgelegten Preisstände.
	Staende int

	Befunde []domain.Befund
}

// Fremdpreisbezug holt einen Preisverlauf und legt ihn ab.
type Fremdpreisbezug struct {
	drucke Marktregister
	quelle Fremdpreisquelle
	ablage Preiseintrag
	jetzt  func() time.Time
}

// NeuerFremdpreisbezug baut ihn.
func NeuerFremdpreisbezug(
	drucke Marktregister, quelle Fremdpreisquelle, ablage Preiseintrag,
) *Fremdpreisbezug {
	return &Fremdpreisbezug{drucke: drucke, quelle: quelle, ablage: ablage, jetzt: time.Now}
}

// Holen holt den Verlauf eines Drucks und legt ihn ab.
func (f *Fremdpreisbezug) Holen(
	ctx context.Context, printing string, wunsch domain.Angebotswunsch,
) (Fremdpreisbilanz, error) {
	druck, fehler := f.drucke.NachScryfall(ctx, printing)
	if fehler != nil {
		return Fremdpreisbilanz{}, fehler
	}

	if druck == nil {
		return Fremdpreisbilanz{}, fmt.Errorf(
			"%w: den Druck %q gibt es in diesem Bestand nicht",
			domain.ErrNotFound, printing)
	}

	bilanz, fehler := f.fassungFuer(ctx, druck)
	if fehler != nil {
		return Fremdpreisbilanz{}, fehler
	}

	staende, befunde, holfehler := f.quelle.Verlauf(
		ctx, bilanz.Fassung, druck.Scryfall(), wunsch)
	if holfehler != nil {
		return Fremdpreisbilanz{}, holfehler
	}

	bilanz.Befunde = append(bilanz.Befunde, befunde...)

	// **Der Abrufzeitpunkt kommt von hier, der Beobachtungstag von der
	// Quelle** (0093): wann gefragt wurde, weiß der Dienst; welchen Tag die
	// Quelle meint, nur sie.
	erhoben := f.jetzt()
	for i := range staende {
		staende[i].Erhoben = erhoben
	}

	// **Alle oder keiner** — dieselbe Zusage wie beim Bulkbezug: ein halber
	// Verlauf wäre ein Stand, den es so nie gab.
	if ablagefehler := f.ablage.Eintragen(ctx, staende...); ablagefehler != nil {
		return Fremdpreisbilanz{}, ablagefehler
	}

	bilanz.Staende = len(staende)

	return bilanz, nil
}

// fassungFuer findet die Kartenfassung eines Drucks beim Anbieter.
//
// **Die gemerkte zuerst**: eine Suche kostet eine Anfrage, und die Zuordnung
// ändert sich nicht.
func (f *Fremdpreisbezug) fassungFuer(ctx context.Context, druck *domain.Druck) (
	Fremdpreisbilanz, error,
) {
	if gemerkt, gibts := druck.Kennung(f.quelle.System(), domain.SemantikDruck); gibts {
		return Fremdpreisbilanz{Fassung: gemerkt}, nil
	}

	fassungen, fehler := f.quelle.Fassungen(ctx, *druck)
	if fehler != nil {
		return Fremdpreisbilanz{}, fehler
	}

	switch len(fassungen) {
	case 0:
		return Fremdpreisbilanz{}, fmt.Errorf(
			"%w: %s kennt %q aus %q nicht",
			domain.ErrNotFound, f.quelle.Familie(), druck.Name, druck.Setcode)
	case 1:
	default:
		// **Mehrdeutig ist ein Konflikt, kein Automerge** (A2, #1441): zwei
		// Treffer heißen, dass der Anbieter etwas doppelt führt oder anders
		// schneidet als dieser Bestand. Das ist eine Frage an einen Menschen.
		return Fremdpreisbilanz{}, fmt.Errorf(
			"%w: %s führt zu %q aus %q %d Fassungen — welche gemeint ist, "+
				"entscheidet dieses Programm nicht",
			domain.ErrConflict, f.quelle.Familie(), druck.Name,
			druck.Setcode, len(fassungen))
	}

	bilanz := Fremdpreisbilanz{Fassung: fassungen[0].Kennung, Zugeordnet: true}

	// **Gemerkt ist Bestenfalls-Arbeit** (§7, Regel 3): scheitert es, kommt es
	// als Befund mit und hält den Bezug nicht auf.
	if merkfehler := f.merken(ctx, druck, fassungen[0].Kennung); merkfehler != nil {
		bilanz.Befunde = append(bilanz.Befunde, domain.Befund{
			Datei: druck.Name, Art: domain.ArtUmgebung,
			Meldung: fmt.Sprintf(
				"Die Zuordnung zu %s ließ sich nicht merken: %v. "+
					"Der nächste Abruf sucht sie wieder.", f.quelle.Familie(), merkfehler),
		})
	}

	return bilanz, nil
}

// merken schreibt die gefundene Kennung an den Druck.
func (f *Fremdpreisbezug) merken(
	ctx context.Context, druck *domain.Druck, fassung string,
) error {
	druck.Kennungen = append(druck.Kennungen, domain.Druckkennung{
		System: f.quelle.System(), Semantik: domain.SemantikDruck, Wert: fassung,
	})

	return f.drucke.Sichern(ctx, druck)
}
