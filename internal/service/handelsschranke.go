package service

// **Die Handelsrichtlinie greift, wo gehandelt wird** — #2435.
//
// Gefunden bei der Erreichbarkeitsinventur von SEC04 (#2357): die fünf
// Schranken unter `richtlinie.handel` standen gefaltet in `domain`
// ([domain.HandelsrichtlinieAus], T1, #1466), die Einstellungsseite sagte
// „eine Schranke, keine Vorliebe" — und kein Dienst fragte sie. Ein Kopf, der
// mehr verspricht, als sein Rumpf hält (§10).
//
// # Wo sie greift
//
// **An jedem Weg, auf dem ein Handel entsteht oder der Einigung näher
// kommt**: Eröffnen, Entwurf ändern, Kopieren, Vorschlagen, die Schritte
// `anbieten`, `gegenangebot` und `zustimmen` — von hier wie von drüben
// ([Handelspost]). Dazu das Festhalten einer Zusage (samt ihrer Frist), das
// Verleihen eines Decks und das Freigeben einer Sicht. Die Prüfung selbst
// steht in `domain` ([domain.Handelsrichtlinie.Erlaubt],
// [domain.Handelsrichtlinie.ErlaubtArt], [domain.Handelsrichtlinie.ErlaubtFrist]);
// hier steht nur, **wann** gefragt wird.
//
// # Wo sie nicht greift
//
// **An den Auswegen**: Ablehnen, Abbrechen, Ablaufen und Klären bleiben
// offen. Eine Schranke, die man umlegt, während ein Handel läuft, soll ihn
// beenden lassen und nicht festhalten. **Und nach der Einigung**:
// `vorbereiten`, `uebergeben` und die Verbuchung führen zu Ende, was beide
// Seiten schon zugesagt haben — Exemplare stehen dann fest, und ein Halt
// mittendrin ließe sie liegen, bis jemand abbricht. Das ist eine Entscheidung
// dieses Dienstes und keine Vorgabe von T1.
//
// # Und was gilt, wenn keine Schranke eingesetzt ist
//
// **Die Richtlinie ohne Einstellung** ([domain.Standardhandelsrichtlinie]):
// handeln ja, Markt nein. Ein Dienst ohne eingesetzte Schranke ließe sonst den
// Markt zu, den das Programm ohne Einstellung ausschließt.

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Handelsschranke gibt die Handelsrichtlinie, die gerade gilt — gefaltet
// über alle Ebenen (Vermerk 0079).
type Handelsschranke interface {
	Handelsrichtlinie(ctx context.Context) (domain.Handelsrichtlinie, error)
}

// ohneEinstellung ist die Schranke, wenn keine eingesetzt ist.
type ohneEinstellung struct{}

func (ohneEinstellung) Handelsrichtlinie(context.Context) (domain.Handelsrichtlinie, error) {
	return domain.Standardhandelsrichtlinie(), nil
}

// angebotErlaubt fragt die Schranke nach einem Angebot.
func angebotErlaubt(ctx context.Context, schranke Handelsschranke, angebot domain.Handelsangebot) error {
	richtlinie, fehler := schranke.Handelsrichtlinie(ctx)
	if fehler != nil {
		return fehler
	}

	return richtlinie.Erlaubt(angebot)
}

// artErlaubt fragt sie nach einer Vertragsart ohne Positionen.
func artErlaubt(ctx context.Context, schranke Handelsschranke, art domain.Handelsart) error {
	richtlinie, fehler := schranke.Handelsrichtlinie(ctx)
	if fehler != nil {
		return fehler
	}

	return richtlinie.ErlaubtArt(art)
}

// zusagefrist fragt sie nach einer Zusage an einem Angebot und gibt die
// Frist, die sie tragen darf.
func zusagefrist(
	ctx context.Context, schranke Handelsschranke, angebot domain.Handelsangebot,
	jetzt, gewuenscht time.Time,
) (time.Time, error) {
	richtlinie, fehler := schranke.Handelsrichtlinie(ctx)
	if fehler != nil {
		return time.Time{}, fehler
	}

	if fehler := richtlinie.Erlaubt(angebot); fehler != nil {
		return time.Time{}, fehler
	}

	return richtlinie.Zusagefrist(jetzt, gewuenscht)
}

// einigungErlaubt fragt die Schranke, wenn ein Stand **auf dem Weg zur
// Einigung** steht — angeboten, in Verhandlung, angenommen. Ein Ausweg und
// was nach der Einigung kommt, fragt sie nicht.
func einigungErlaubt(ctx context.Context, schranke Handelsschranke, angebot domain.Handelsangebot) error {
	switch angebot.Zustand {
	case domain.HandelAngeboten, domain.HandelVerhandlung, domain.HandelAngenommen:
		return angebotErlaubt(ctx, schranke, angebot)
	}

	return nil
}

// einigungVerbuchbar fragt auf demselben Weg, ob jede Zeile drüben zu buchen
// ist (#2495) — **nur für die eigenen Schritte**: was von drüben kommt, wird
// angenommen, wie es ist, und erst die eigene Zustimmung hält an.
func einigungVerbuchbar(angebot domain.Handelsangebot) error {
	switch angebot.Zustand {
	case domain.HandelAngeboten, domain.HandelVerhandlung, domain.HandelAngenommen:
		return angebot.PruefeVerbuchbar()
	}

	return nil
}

// MitHandelsschranke setzt die Handelsrichtlinie, die gilt — **auch für das,
// was von drüben kommt**: eine Schranke gilt unabhängig davon, wer fragt.
func (p *Handelspost) MitHandelsschranke(schranke Handelsschranke) *Handelspost {
	p.schranke = schranke

	return p
}
