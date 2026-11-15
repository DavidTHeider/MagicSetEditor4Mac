package app

// **Das Kostenjournal externer Aufträge** — Werkstück O12 von Phase 5.2
// (#1439).
//
// # Vier Zahlen, überall dieselben
//
// Verfügbar, reserviert, verbraucht, ungeklärt. Die Abnahme verlangt es
// wörtlich: „UI und vier Flächen zeigen denselben verfügbaren/reservierten/
// verbrauchte/ungeklärten Stand, ohne Secrets oder Rohprompts im Ledger."

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Kostenfuehrer ist die Naht zum Dienst.
type Kostenfuehrer interface {
	Stand(ctx context.Context, konto string, zeitraum domain.Zeitraum) (
		domain.Haushaltsstand, error)
	Posten(ctx context.Context, konto string, zeitraum domain.Zeitraum) (
		[]domain.Kostenposten, error)
	BudgetSetzen(ctx context.Context, konto string, zeitraum domain.Zeitraum,
		betrag domain.Betrag) error
	Reservieren(ctx context.Context, auftrag service.Kostenauftrag, nachweis string) (
		domain.Kostenposten, error)
	Abschliessen(ctx context.Context, kennung string, betrag domain.Betrag) (
		domain.Kostenposten, error)
	Freigeben(ctx context.Context, kennung string) (domain.Kostenposten, error)
	AlsUnklarMerken(ctx context.Context, kennung, grund string) (domain.Kostenposten, error)
}

// BetragDTO ist Geld in der kleinsten Einheit seiner Währung.
//
// **Als Zahl und Währung und nicht als Text.** „12,00 €" wäre eine
// Formatierung, und wer sie zurückrechnet, rundet.
type BetragDTO struct {
	Wert     int64  `json:"wert"`
	Waehrung string `json:"waehrung"`
}

// HaushaltsstandAntwort sind die vier Zahlen samt ihrer Grenze.
type HaushaltsstandAntwort struct {
	Konto    string `json:"konto"`
	Zeitraum string `json:"zeitraum"`

	Grenze     BetragDTO `json:"grenze"`
	Reserviert BetragDTO `json:"reserviert"`
	Verbraucht BetragDTO `json:"verbraucht"`
	Unklar     BetragDTO `json:"unklar"`
	Verfuegbar BetragDTO `json:"verfuegbar"`

	Fehler *Fehler `json:"fehler"`
}

// KostenpostenDTO ist ein Auftrag im Journal.
//
// **Ohne Prompt und ohne Zugangsdaten** — was hier steht, ist die Aufgabe als
// Marke.
type KostenpostenDTO struct {
	Kennung string `json:"kennung"`
	Konto   string `json:"konto"`
	Aufgabe string `json:"aufgabe"`

	Hoechstbetrag BetragDTO `json:"hoechstbetrag"`
	Betrag        BetragDTO `json:"betrag"`

	// Stand ist eine **Marke** — vier Flächen bauen daraus ihren eigenen Satz.
	Stand string `json:"stand"`
	Grund string `json:"grund"`

	Begonnen string `json:"begonnen"`
	Beendet  string `json:"beendet"`
}

// KostenpostenAntwort ist die Liste der letzten Aufträge.
type KostenpostenAntwort struct {
	Posten []KostenpostenDTO `json:"posten"`
	Fehler *Fehler           `json:"fehler"`
}

// Kostenstand sagt, was an einem Konto steht.
func (a *App) Kostenstand(ctx context.Context, konto, zeitraum string) HaushaltsstandAntwort {
	ctx = a.imProfil(ctx)

	stand, fehler := a.kosten.Stand(ctx, konto, domain.Zeitraum(zeitraum))
	if fehler != nil {
		return HaushaltsstandAntwort{
			Konto: konto, Zeitraum: zeitraum, Fehler: Uebersetze(ctx, fehler),
		}
	}

	return HaushaltsstandAntwort{
		Konto: stand.Konto,

		// **Der Zeitraum kommt aus der Antwort und nicht aus der Frage** —
		// wer nichts nennt, bekommt den Tag, und die Auskunft sagt es.
		Zeitraum:   string(stand.Zeitraum),
		Grenze:     betragDTO(stand.Grenze),
		Reserviert: betragDTO(stand.Reserviert),
		Verbraucht: betragDTO(stand.Verbraucht),
		Unklar:     betragDTO(stand.Unklar),
		Verfuegbar: betragDTO(stand.Verfuegbar),
	}
}

// Kostenposten gibt die letzten Aufträge eines Kontos.
func (a *App) Kostenposten(ctx context.Context, konto, zeitraum string) KostenpostenAntwort {
	ctx = a.imProfil(ctx)

	posten, fehler := a.kosten.Posten(ctx, konto, domain.Zeitraum(zeitraum))
	if fehler != nil {
		return KostenpostenAntwort{
			Posten: []KostenpostenDTO{}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	uebersetzt := make([]KostenpostenDTO, 0, len(posten))
	for _, einer := range posten {
		uebersetzt = append(uebersetzt, KostenpostenDTO{
			Kennung: einer.Kennung, Konto: einer.Konto, Aufgabe: einer.Aufgabe,
			Hoechstbetrag: betragDTO(einer.Hoechstbetrag),
			Betrag:        betragDTO(einer.Betrag),
			Stand:         string(einer.Stand), Grund: einer.Grund,
			Begonnen: alsZeitpunkt(einer.Begonnen),
			Beendet:  alsZeitpunkt(einer.Beendet),
		})
	}

	return KostenpostenAntwort{Posten: uebersetzt}
}

// KostenbudgetSetzen legt die Grenze eines Kontos fest.
//
// **Null nimmt sie weg** — und dann geht kein kostenpflichtiger Auftrag mehr
// hinaus. Das ist die sichere Vorgabe und kein Nebeneffekt.
func (a *App) KostenbudgetSetzen(
	ctx context.Context, konto, zeitraum string, wert int64, waehrung string,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	fehler := a.kosten.BudgetSetzen(ctx, konto, domain.Zeitraum(zeitraum),
		domain.Betrag{Wert: wert, Waehrung: waehrung})
	if fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return SchreibAntwort{}
}

// MitKosten setzt die Naht.
func (a *App) MitKosten(fuehrer Kostenfuehrer) { a.kosten = fuehrer }

// betragDTO übersetzt einen Betrag.
func betragDTO(betrag domain.Betrag) BetragDTO {
	return BetragDTO{Wert: betrag.Wert, Waehrung: betrag.Waehrung}
}

// Kostenstand ist die Fensterbindung.
func (b *Bindungen) Kostenstand(ctx context.Context, konto, zeitraum string) HaushaltsstandAntwort {
	return b.app(ctx).Kostenstand(b.ruf(ctx), konto, zeitraum)
}

// Kostenposten ist die Fensterbindung für die Liste.
func (b *Bindungen) Kostenposten(ctx context.Context, konto, zeitraum string) KostenpostenAntwort {
	return b.app(ctx).Kostenposten(b.ruf(ctx), konto, zeitraum)
}

// KostenbudgetSetzen ist die Fensterbindung für die Grenze.
func (b *Bindungen) KostenbudgetSetzen(
	ctx context.Context,
	konto, zeitraum string, wert int64, waehrung string,
) SchreibAntwort {
	return b.app(ctx).KostenbudgetSetzen(b.ruf(ctx), konto, zeitraum, wert, waehrung)
}
