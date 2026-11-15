package service

// **Das eigene Konto bei einem Marktplatz, nur gelesen** — Werkstück `T11`
// von Phase 5.6
// (#1476).
//
// # Was diese Stelle tut — und was ausdrücklich nicht
//
// **Sie liest, und sie bucht nichts.** Ein Angebot dort ist kein Verkauf,
// ein Warenkorb ist keine Bestellung, und eine Bestellung ist keine Karte,
// die schon den Besitzer gewechselt hat. Ob und wann aus einer fremden
// Bestellung eine Bewegung im eigenen Bestand wird, entscheidet ein Mensch —
// über die vorhandenen Schreibwege und nicht von hier aus.
//
// **Sie sagt trotzdem, was der Fall wäre**: je Bestellung steht daneben, ob
// sie nach der Regel des Anbieters ein Stück aus dem eigenen Lager genommen
// hat ([domain.BewegtDenBestand]). Das ist eine Auskunft und keine Buchung —
// und ohne sie müsste jeder sie selbst herleiten, wo sie doch **eine** Regel
// ist.
//
// # Zwei Schalter und ein Riegel
//
// **Der Dienst muss an sein** — er fängt aus an (0063) —, **die Kontoschau
// muss an sein** ([domain.Kontoschau], fängt ebenfalls aus an), und der
// Aufruf muss hinausdürfen ([domain.WirkungExternSenden]). Wer den
// Marktplatz eingeschaltet hat, hat damit nicht sein Konto freigegeben.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Kontoquelle ist, was die Schau von einem Marktplatz braucht.
//
// **Drei lesende Rufe und keine vierte Tür** (§2): die Schnittstelle steht
// hier, weil dieser Anwendungsfall sie braucht — und sie hat kein Verfahren,
// das schreibt.
type Kontoquelle interface {
	Konto(ctx context.Context) (domain.Marktkonto, error)
	EigeneAngebote(ctx context.Context, erweiterung int64) (
		[]domain.Eigenangebot, error)
	Bestellungen(ctx context.Context, seite, menge int) (
		[]domain.Marktbestellung, []domain.Befund, error)
}

// Kontoschalter sagt, ob die Kontoschau erlaubt ist.
type Kontoschalter interface {
	Wert(ctx context.Context, schluessel string) (string, error)
}

// Kontoschau liest das eigene Konto bei einem Marktplatz.
type Kontoschau struct {
	quelle   Kontoquelle
	schalter Kontoschalter
	wache    Pruefer
}

// NeueKontoschau baut sie.
func NeueKontoschau(
	quelle Kontoquelle, schalter Kontoschalter, wache Pruefer,
) *Kontoschau {
	return &Kontoschau{quelle: quelle, schalter: schalter, wache: wache}
}

// Kontostand ist, was aus dem Konto zu sehen ist.
type Kontostand struct {
	Konto        domain.Marktkonto
	Angebote     []domain.Eigenangebot
	Bestellungen []domain.Marktbestellung
	Befunde      []domain.Befund
}

// Lesen gibt Konto, eigene Angebote und Bestellungen — in einem Zug.
//
// **Drei Rufe und eine Antwort**: wer sie einzeln anböte, ließe jede Fläche
// die Reihenfolge selbst wählen, und eine davon fragte das Konto nach den
// Angeboten, bevor sie weiß, ob der Zugang überhaupt trägt.
//
// **Der Zugang zuerst.** `/info` ist der Weg, den der Anbieter selbst zum
// Prüfen nennt; scheitert er, erspart das zwei Anfragen, von denen feststeht,
// dass sie scheitern (§16, Regel 4).
func (k *Kontoschau) Lesen(ctx context.Context, seite, menge int) (
	Kontostand, error,
) {
	if fehler := k.darf(ctx); fehler != nil {
		return Kontostand{}, fehler
	}

	konto, fehler := k.quelle.Konto(ctx)
	if fehler != nil {
		return Kontostand{}, fehler
	}

	angebote, fehler := k.quelle.EigeneAngebote(ctx, 0)
	if fehler != nil {
		return Kontostand{}, fehler
	}

	bestellungen, befunde, fehler := k.quelle.Bestellungen(ctx, seite, menge)
	if fehler != nil {
		return Kontostand{}, fehler
	}

	return Kontostand{
		Konto: konto, Angebote: angebote, Bestellungen: bestellungen,
		Befunde: befunde,
	}, nil
}

// darf prüft die zwei Schalter und den Riegel.
//
// **In dieser Reihenfolge**: der Schalter ist eine Einstellung dieses
// Profils, der Riegel eine Frage an den Akteur. Wer ausgeschaltet hat, soll
// das lesen und nicht „dir fehlt eine Rolle".
func (k *Kontoschau) darf(ctx context.Context) error {
	stellung, fehler := k.schalter.Wert(ctx, domain.Kontoschau)
	if fehler != nil {
		return fehler
	}

	if stellung != domain.KontoschauAn {
		return fmt.Errorf("%w: die Kontoschau ist aus — sie steht unter "+
			"Einstellungen → Dienste und fängt ausgeschaltet an, weil ein "+
			"eingeschalteter Marktplatz noch keine Freigabe für das eigene "+
			"Konto ist", domain.ErrErlaubnis)
	}

	if k.wache == nil {
		return nil
	}

	return k.wache.Pruefe(ctx, domain.WirkungExternSenden)
}
