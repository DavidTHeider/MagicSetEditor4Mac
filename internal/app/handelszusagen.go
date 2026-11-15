package app

// **Was für einen Handel festgehalten wird** —
// #1643.
//
// # Warum das eine eigene Fläche braucht
//
// **Weil ein Angebot die Besitzzeile nicht nennt.** Es nennt einen **Druck**;
// wer drei `NM`-Exemplare desselben Drucks in drei Lagerorten hat, muss
// sagen, welches weggeht. Das rät kein Weg — und einer, der es täte, legte
// die falsche Karte auf den Tisch.
//
// # Und warum die Antwort die ganze Liste trägt
//
// **Weil eine Zusage allein nicht sagt, ob es reicht.** Die Fläche zeigt, was
// für diesen Vorgang zusammengekommen ist; `vorbereiten` verlangt, dass es
// deckt, was weggeht.

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// HandelszusageDTO ist eine festgehaltene Besitzzeile.
type HandelszusageDTO struct {
	ExemplarID int64 `json:"exemplarId"`
	Menge      int   `json:"menge"`

	// Fassung ist die Fassung der Besitzzeile, gegen die gerechnet wurde.
	Fassung int64 `json:"fassung"`

	// Revision ist die Angebotsrevision, der die Zusage gilt.
	Revision int `json:"revision"`

	// Frist ist leer, wo es keine gibt — der Handel selbst ist die Frist.
	Frist string `json:"frist,omitempty"`

	// Uebergeben heißt: sie ist über den Tisch gegangen.
	Uebergeben bool `json:"uebergeben"`
}

// ZusagenAntwort ist, was für einen Vorgang festgehalten ist.
type ZusagenAntwort struct {
	Vorgang string             `json:"vorgang"`
	Zusagen []HandelszusageDTO `json:"zusagen"`

	// Nachladefehler sagt: die Zusage **ist festgehalten**, nur die Liste
	// danach war nicht zu lesen (§7, Regel 3; #1970) — dann ist `zusagen`
	// leer und `fehler` auch.
	Nachladefehler *Fehler `json:"nachladefehler,omitempty"`

	Fehler *Fehler `json:"fehler"`
}

// HandelFesthalten hält eine benannte Besitzzeile für einen Vorgang fest.
//
// **Die Fassung kommt aus dem Ruf** (§7, Regel 2): sie ist die, die ein
// Mensch gesehen hat. Hat inzwischen jemand geschrieben, ist die Zusage ein
// Konflikt und kein stiller Erfolg.
func (a *App) HandelFesthalten(
	ctx context.Context, vorgang string, exemplarID int64,
	menge int, fassung int64,
) ZusagenAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungSchreiben); fehler != nil {
		return leereZusagen(ctx, vorgang, fehler)
	}

	// **Wer weggibt, ist dieser Rechner** — und der Vertrag sagt, wie viel
	// er weggibt (Befund 1 des Zweitblicks über Phase 5.6).
	ich, fehler := a.oderIch(ctx, "")
	if fehler != nil {
		return leereZusagen(ctx, vorgang, fehler)
	}

	if _, fehler := a.zusagenwerk().Festhalten(ctx, service.Zusagenwunsch{
		Vorgang: vorgang, Wer: ich, ExemplarID: exemplarID,
		Menge: menge, Fassung: fassung,
	}); fehler != nil {
		return leereZusagen(ctx, vorgang, fehler)
	}

	// **Geschrieben ist geschrieben** (§7, Regel 3; #1970): scheitert das
	// Nachlesen, steht sein Fehler daneben — wer `fehler` läse, hielte
	// dieselbe Besitzzeile ein zweites Mal fest.
	antwort := a.Handelszusagen(ctx, vorgang)
	antwort.Nachladefehler, antwort.Fehler = antwort.Fehler, nil

	return antwort
}

// Handelszusagen gibt, was für einen Vorgang festgehalten ist.
func (a *App) Handelszusagen(
	ctx context.Context, vorgang string,
) ZusagenAntwort {
	ctx = a.imProfil(ctx)

	zusagen, fehler := a.zusagenwerk().Zusagen(ctx, vorgang)
	if fehler != nil {
		return leereZusagen(ctx, vorgang, fehler)
	}

	antwort := ZusagenAntwort{
		Vorgang: vorgang,
		Zusagen: make([]HandelszusageDTO, 0, len(zusagen)),
	}

	for _, zusage := range zusagen {
		antwort.Zusagen = append(antwort.Zusagen, alsZusage(zusage))
	}

	return antwort
}

// alsZusage übersetzt eine Zusage für das Fenster.
func alsZusage(zusage domain.Reservierung) HandelszusageDTO {
	sicht := HandelszusageDTO{
		ExemplarID: zusage.ExemplarID, Menge: zusage.Menge,
		Fassung: zusage.Fassung, Revision: zusage.Revision,
		Uebergeben: zusage.Uebergeben,
	}

	if !zusage.Frist.IsZero() {
		sicht.Frist = zusage.Frist.UTC().Format(time.RFC3339)
	}

	return sicht
}

// leereZusagen gibt die Absage — **mit einer leeren Liste und ohne `null`**.
func leereZusagen(
	ctx context.Context, vorgang string, fehler error,
) ZusagenAntwort {
	return ZusagenAntwort{
		Vorgang: vorgang, Zusagen: []HandelszusageDTO{},
		Fehler: Uebersetze(ctx, fehler),
	}
}
