package app

// **Was zu einem Gegenüber hinausgeht** — der Aufrufer für
// `service.Handelspost.Senden` aus `T5`
// (#1470),
// nachgetragen mit
// #1637.
//
// # Was „schicken" hier heißt und was nicht
//
// **Es legt die Nachricht in den Ausgang.** Über die Leitung geht sie damit
// noch nicht — „abgeschickt" heißt nicht „angekommen", und was hier
// hineingeht, bleibt liegen, bis die Gegenseite es bestätigt (`T5`). Dass der
// Ausgang von selbst leerläuft, verlangt eine Fläche für die Leitung, und die
// steht als
// #1634.
//
// # Und warum die Nachlieferung nur Kennungen nennt
//
// **Weil eine Diagnose sagen soll, ob etwas klemmt, und nicht, worum es
// geht** — derselbe Satz wie an [PoststandDTO]. Was nachzuliefern ist, steht
// im Ausgang; wer es hinausträgt, braucht die Kennung und nicht eine zweite
// Sicht auf den Handel.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// VersandAntwort nennt die Nachricht, die nun im Ausgang liegt.
type VersandAntwort struct {
	// NachrichtID ist die Kennung, an der die Bestätigung sie wegräumt.
	NachrichtID string `json:"nachrichtId"`

	Fehler *Fehler `json:"fehler"`
}

// HandelsfortschrittDTO ist, was eine Seite als bestätigt kennt.
type HandelsfortschrittDTO struct {
	Vorgang string `json:"vorgang"`

	// Revision ist die höchste bestätigte — **0 heißt: noch keine**.
	Revision int `json:"revision"`

	// Abdruck ist der Abdruck dieser Revision — er ist die Aussage, nicht
	// die Zahl.
	Abdruck string `json:"abdruck"`

	// Folge ist die höchste Folge, die von der anderen Seite bestätigt
	// angekommen ist.
	Folge int64 `json:"folge"`
}

// NachlieferungDTO ist eine Nachricht, die noch einmal hinausmuss.
type NachlieferungDTO struct {
	NachrichtID string `json:"nachrichtId"`
	Vorgang     string `json:"vorgang"`
	Empfaenger  string `json:"empfaenger"`
	Folge       int64  `json:"folge"`

	// Art ist die Nachrichtenart — `vorschlag`, `annahme`, `uebergabe`, …
	Art string `json:"art"`
}

// WiederaufnahmeAntwort sagt, was nach einem Abriss fehlt.
type WiederaufnahmeAntwort struct {
	Nachzuliefern []NachlieferungDTO `json:"nachzuliefern"`

	// Klaerung heißt: beide Seiten haben ordentlich gearbeitet, und
	// trotzdem stehen zwei Verläufe da. **Kein Fehler** — der Grund steht in
	// [domain.Nachzuliefern].
	Klaerung bool `json:"klaerung"`

	Fehler *Fehler `json:"fehler"`
}

// HandelSchicken legt den heutigen Stand eines Vorgangs in den Ausgang.
func (a *App) HandelSchicken(
	ctx context.Context, vorgang, empfaenger string,
) VersandAntwort {
	ctx = a.imProfil(ctx)

	nachrichtID, fehler := a.handelsversand().Schicken(ctx, vorgang, empfaenger)
	if fehler != nil {
		return VersandAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return VersandAntwort{NachrichtID: nachrichtID}
}

// HandelPostbestaetigung räumt eine Nachricht weg, die angekommen ist.
func (a *App) HandelPostbestaetigung(
	ctx context.Context, nachrichtID string,
) VersandAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.handelsversand().Bestaetigt(ctx, nachrichtID); fehler != nil {
		return VersandAntwort{
			NachrichtID: nachrichtID, Fehler: Uebersetze(ctx, fehler),
		}
	}

	return VersandAntwort{NachrichtID: nachrichtID}
}

// HandelWiederaufnahme sagt, was nach einem Abriss noch einmal hinausmuss.
//
// **Der eigene Fortschritt wird gelesen und nicht genannt** — nur der des
// Gegenübers kommt von außen, und er ist eine Behauptung, die gegen den
// eigenen Stand gehalten wird.
func (a *App) HandelWiederaufnahme(
	ctx context.Context, vorgang, empfaenger string,
	seiner HandelsfortschrittDTO,
) WiederaufnahmeAntwort {
	ctx = a.imProfil(ctx)

	offen, klaerung, fehler := a.handelsversand().Wiederaufnehmen(
		ctx, vorgang, empfaenger, domain.Handelsfortschritt{
			Vorgang: vorgang, Revision: seiner.Revision,
			Abdruck: seiner.Abdruck, Folge: seiner.Folge,
		})
	if fehler != nil {
		return WiederaufnahmeAntwort{
			Nachzuliefern: []NachlieferungDTO{},
			Fehler:        Uebersetze(ctx, fehler),
		}
	}

	return WiederaufnahmeAntwort{
		Nachzuliefern: alsNachlieferungen(offen), Klaerung: klaerung,
	}
}

// alsNachlieferungen macht aus den offenen Nachrichten ihre Kennungen.
func alsNachlieferungen(
	offen []service.Ausgangsnachricht,
) []NachlieferungDTO {
	dtos := make([]NachlieferungDTO, 0, len(offen))

	for _, zeile := range offen {
		dtos = append(dtos, NachlieferungDTO{
			NachrichtID: zeile.NachrichtID, Vorgang: zeile.Vorgang,
			Empfaenger: zeile.Empfaenger, Folge: zeile.Folge,
			Art: string(zeile.Nachricht.Umschlag.Art),
		})
	}

	return dtos
}
