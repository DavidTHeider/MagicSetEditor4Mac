package app

// **Das eigene Konto bei einem Marktplatz** — Werkstück `T11` von Phase 5.6
// (#1476).
//
// # Ein Ruf und nicht drei
//
// **Weil drei Rufe drei Reihenfolgen hätten.** Konto, eigene Angebote und
// Bestellungen gehören zu einer Übersicht; wer sie einzeln anböte, ließe jede
// Fläche selbst entscheiden, ob sie den Zugang vor oder nach den Daten prüft.
//
// # Und warum es nur diesen einen gibt
//
// **Weil dieses Programm dort nichts schreibt.** Angebot anlegen, Warenkorb,
// Kauf und Versand kennt der Anbieter; warum keiner davon hier steht, sagt
// [0122](../../docs/entscheidungen/0122-ein-fremdes-konto-wird-gelesen.md).

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// EigenangebotDTO ist ein Posten, den das eigene Konto anbietet.
type EigenangebotDTO struct {
	ID          int64  `json:"id"`
	BlueprintID int64  `json:"blueprintID"`
	Name        string `json:"name"`
	Menge       int    `json:"menge"`

	// PreisCent und Waehrung gehören zusammen — **beides oder keines**.
	PreisCent int64  `json:"preisCent"`
	Waehrung  string `json:"waehrung"`

	// Eigennotiz ist der Faden zum eigenen Lager.
	Eigennotiz   string `json:"eigennotiz"`
	Beschreibung string `json:"beschreibung"`
}

// MarktbestellungDTO ist ein Kauf, an dem das eigene Konto beteiligt ist.
type MarktbestellungDTO struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`

	// Zustand ist einer der acht des Anbieters — **unübersetzt**, damit ein
	// Vergleich ein Vergleich bleibt.
	Zustand string `json:"zustand"`

	// UeberNullkonto ist `via_cardtrader_zero`.
	UeberNullkonto bool `json:"ueberNullkonto"`

	// Rolle ist `seller` oder `buyer`.
	Rolle  string `json:"rolle"`
	Stueck int    `json:"stueck"`

	// Bezahlt ist der Zeitpunkt der Zahlung, wie der Anbieter ihn schreibt.
	Bezahlt string `json:"bezahlt"`

	// BewegtBestand sagt, ob diese Bestellung nach der Regel des Anbieters
	// ein Stück aus dem eigenen Lager genommen hat
	// ([domain.BewegtDenBestand]).
	//
	// **Es ist eine Auskunft und keine Buchung.** Dieses Programm bucht
	// daraus nichts; wer es tun will, tut es von Hand, und dann steht hier,
	// was zu tun wäre.
	BewegtBestand bool `json:"bewegtBestand"`
}

// KontostandDTO ist die Übersicht über das eigene Konto.
type KontostandDTO struct {
	// Name und NutzerID sind, was der Anbieter über den Zugang sagt.
	//
	// **Ohne das geteilte Geheimnis**: der Anbieter schreibt „Do not share
	// this information" daneben, und dieses Programm liest es gar nicht erst.
	Name     string `json:"name"`
	NutzerID int64  `json:"nutzerID"`

	Angebote     []EigenangebotDTO    `json:"angebote"`
	Bestellungen []MarktbestellungDTO `json:"bestellungen"`

	// Befunde sind die Zeilen, die nicht aufgingen — etwa ein Zustand, den
	// dieses Programm nicht kennt.
	Befunde []string `json:"befunde"`
}

// KontostandAntwort trägt die Übersicht samt ihrem Fehler.
type KontostandAntwort struct {
	Stand  *KontostandDTO `json:"stand"`
	Fehler *Fehler        `json:"fehler"`
}

// Marktkonto gibt die Übersicht über das eigene Konto bei einem Marktplatz.
//
// **`quelle` ist der Marktplatz** — heute gibt es einen, und der Name steht
// im Aufruf, damit ein zweiter ihn nicht still verdrängt.
//
// **`seite` und `menge` sind die Blätterangaben des Anbieters**: null heißt
// „nimm seine Vorgabe" und nicht „nimm alles".
func (a *App) Marktkonto(
	ctx context.Context, quelle string, seite, menge int,
) KontostandAntwort {
	ctx = a.imProfil(ctx)

	schau, fehler := a.kontoschau(ctx, quelle)
	if fehler != nil {
		return KontostandAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	stand, lesefehler := schau.Lesen(ctx, seite, menge)
	if lesefehler != nil {
		return KontostandAntwort{Fehler: Uebersetze(ctx, lesefehler)}
	}

	return KontostandAntwort{Stand: alsKontostand(stand)}
}

// Marktkonto ist die Fensterbindung für die Übersicht.
func (b *Bindungen) Marktkonto(
	ctx context.Context,
	quelle string, seite, menge int,
) KontostandAntwort {
	return b.app(ctx).Marktkonto(b.ruf(ctx), quelle, seite, menge)
}

// alsKontostand bildet die Übersicht auf ihre Gestalt für die Flächen ab.
func alsKontostand(stand kontostand) *KontostandDTO {
	angebote := make([]EigenangebotDTO, 0, len(stand.Angebote))
	for _, angebot := range stand.Angebote {
		angebote = append(angebote, EigenangebotDTO{
			ID: angebot.ID, BlueprintID: angebot.BlueprintID,
			Name: angebot.Name, Menge: angebot.Menge,
			PreisCent: angebot.PreisCent, Waehrung: angebot.Waehrung,
			Eigennotiz: angebot.Eigennotiz, Beschreibung: angebot.Beschreibung,
		})
	}

	bestellungen := make([]MarktbestellungDTO, 0, len(stand.Bestellungen))
	for _, bestellung := range stand.Bestellungen {
		bestellungen = append(bestellungen, MarktbestellungDTO{
			ID: bestellung.ID, Code: bestellung.Code,
			Zustand:        string(bestellung.Zustand),
			UeberNullkonto: bestellung.UeberNullkonto,
			Rolle:          bestellung.Rolle, Stueck: bestellung.Stueck,
			Bezahlt: bestellung.Bezahlt,
			BewegtBestand: domain.BewegtDenBestand(
				bestellung.Zustand, bestellung.UeberNullkonto),
		})
	}

	return &KontostandDTO{
		Name: stand.Konto.Name, NutzerID: stand.Konto.NutzerID,
		Angebote: angebote, Bestellungen: bestellungen,
		Befunde: befundtexte(stand.Befunde),
	}
}

// keinMarktplatz sagt, dass dieser Name keiner ist.
func keinMarktplatz(genannt, bekannt string) error {
	return fmt.Errorf(
		"%w: den Marktplatz %q kennt diese Anwendung nicht — es gibt %q",
		domain.ErrValidation, genannt, bekannt)
}
