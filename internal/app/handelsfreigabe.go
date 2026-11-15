package app

// **Ein Mensch gibt einen verbindlichen Handelsschritt frei** — Werkstück
// `T10` von Phase 5.6
// (#1475).
//
// # Warum es das gibt
//
// **Weil `Verbuchen` und `Bestaetigen` seit T10 einen Nachweis verlangen**
// und niemand ihn ausstellen konnte. Derselbe Befund wie bei der KI-Freigabe
// (#1539):
// der Mechanismus stand fertig da, er war nur nicht angeschlossen — und eine
// Zusage, die keine Fläche einlösen kann, ist keine Zusage, sondern eine
// Sperre.
//
// # Warum die Freigabe einem Schritt gilt und nicht einem Handel
//
// **Weil Verbuchen und Bestätigen zwei Entscheidungen sind.** Wer den
// Bestandsübergang gewollt hat, hat damit nicht die zweite Unterschrift
// gewollt; sie kommt später, und zwischen beiden liegt die Gegenseite. Ein
// gemeinsamer Zweck ließe das eine Ja für beides gelten.
//
// # Und warum es sie über MCP und REST nicht gibt
//
// **Weil ein Modell kein Mensch ist und ein Port kein Bildschirm.**
// `Nachweise.Ausstellen` weist jeden Modus außer `fenster` ab; eine Bindung
// daneben, die es doch täte, hebelte genau die Zusage aus, für die es
// Nachweise gibt
// ([0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md)).

import (
	"context"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Handelsfreigabefrist ist, wie lange eine erteilte Freigabe gilt.
//
// **Fünf Minuten**, und die Zahl folgt dem Zweck — dieselbe wie bei der
// KI-Freigabe: sie deckt den Weg vom Knopf bis zum Ruf und nicht den
// Nachmittag. Wer länger braucht, gibt sie neu.
const Handelsfreigabefrist = 5 * time.Minute

// Handelsfreigabeschritte sind die zwei Schritte, für die es eine Freigabe
// gibt — und ihre Zwecke.
//
// **Zwei und nicht drei.** Das Abgleichen bindet nichts: es meldet der
// Gegenseite, dass hier fertig ist. Wer dafür einen Menschen fragte, fragte
// ihn nach einer Entscheidung, die er nicht trifft.
func Handelsfreigabeschritte() map[string]string {
	return map[string]string{
		"verbuchen":   service.HandelsverbuchungZweck,
		"bestaetigen": service.HandelsabschlussZweck,
	}
}

// HandelFreigabeErteilen stellt den Nachweis für **einen** Schritt aus.
//
// **Die Revision wird hier gelesen und nicht vom Aufrufer genannt.** Ein
// Fenster, das sie mitschickte, nennte einen Stand, den es von irgendwoher
// hat; hier steht der, den der Bestand **jetzt** führt. Ändert sich der
// Vorgang zwischen Freigabe und Einlösung, gilt sie nicht mehr — niemand hat
// den neuen Stand gesehen ([domain.Nachweis]).
//
// **Und einen Vorgang, den es nicht gibt, gibt es auch nicht freizugeben.**
//
// # Warum `wer` dabeisteht
//
// **Weil `bestaetigen` die Stimme einer benannten Partei setzt** (Befund 5
// des Zweitblicks über Phase 5.6). Ein Nachweis, der nur den Vorgang nennt,
// gilt für **jede** Stimme: ein Mensch gibt einmal frei, und derselbe
// Nachweis setzt die Unterschrift des Gegenübers. Seit dem 2026-09-19 hängt
// er an der Partei, und wer zwei Stimmen setzt, gibt zweimal frei — wofür,
// steht jedes Mal dabei.
//
// **Für `verbuchen` und `abgleichen` bleibt `wer` leer**: sie setzen keine
// Stimme, sie bewegen den eigenen Bestand.
func (a *App) HandelFreigabeErteilen(
	ctx context.Context, vorgang, schritt, wer string,
) NachweisAntwort {
	ctx = a.imProfil(ctx)

	zweck, bekannt := Handelsfreigabeschritte()[schritt]
	if !bekannt {
		return NachweisAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: %q ist kein Schritt, für den es eine Freigabe gibt",
			domain.ErrValidation, schritt))}
	}

	db := a.laufendeVerbindung()
	if db == nil {
		return NachweisAntwort{Fehler: Uebersetze(ctx, fehlenderBestand())}
	}

	// **Über den Bestand und nicht über `a.autorisierung`** — denselben Weg
	// nimmt die Verbuchungsstelle, die den Nachweis einlöst. Wer beide aus
	// verschiedenen Quellen bezöge, stellte in der einen Tafel aus und suchte
	// in der anderen.
	//
	// **Ausgestellt wird trotzdem nur am Fenster**: `Nachweise.Ausstellen`
	// weist jeden anderen Modus ab, und daran ändert die Quelle nichts.
	angebot, fehler := a.handelsablage().Vorgang(ctx, vorgang)
	if fehler != nil {
		return NachweisAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	nachweis, fehler := AutorisierungAus(db).Ausstellen(ctx, service.Nachweisantrag{
		// **Derselbe Zweck, den die Verbuchungsstelle einlöst.** Ein anderer
		// wäre ein Nachweis, der zu nichts passt — `Einloesen` hält beide
		// gegeneinander.
		Zweck: zweck, Ziel: service.Handelsfreigabeziel(vorgang, wer),
		Revision: int64(angebot.Revision),
		Frist:    Handelsfreigabefrist,
	})
	if fehler != nil {
		return NachweisAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return NachweisAntwort{
		Kennung:  nachweis.Kennung,
		Laeuftab: nachweis.Ablauf.UTC().Format(time.RFC3339),
	}
}

// HandelFreigabeErteilen ist die Fensterbindung für die Freigabe.
func (b *Bindungen) HandelFreigabeErteilen(
	ctx context.Context,
	vorgang, schritt, wer string,
) NachweisAntwort {
	return b.app(ctx).HandelFreigabeErteilen(
		b.ruf(ctx), vorgang, schritt, wer)
}
