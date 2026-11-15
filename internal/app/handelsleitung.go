package app

// **Ein Handel von der Eröffnung bis zur Übergabe** —
// #1637.
//
// # Warum es das gibt
//
// **Weil `service.Handelsleitung` aus `T1` keinen Aufrufer hatte.**
// `Eroeffnen`, `Vorschlagen`, `Schreiten` und `Stand` standen gebaut,
// geprüft und zu 100 % abgedeckt da — und **unerreichbar**. Ein Paket ohne
// Aufrufer ist keine Anwendungsfunktion (§10, erste Dimension); derselbe
// Befund wie seinerzeit bei `internal/svg`.
//
// # Warum der Initiator aus dem Ausweis kommt
//
// **Weil an ihm die Sequenz hängt.** Wer den Vorgang eröffnet hat, nummeriert
// die Revisionen ([domain.Handelsangebot.Fortgeschrieben]); ein Aufrufer, der
// den Namen hinschreiben dürfte, eröffnete einen Handel im Namen eines
// anderen — und die Gegenseite hätte zwei Stellen, die sequenzieren.
//
// # Und was hier **nicht** geschieht
//
// **Die drei Abschlussschritte.** `verbuchen`, `abgleichen` und
// `bestaetigen` bewegen Bestand und Beleg und verlangen einen Nachweis (T10,
// #1475);
// sie gehen über [App.Handelsverbuchung] und ihre beiden Geschwister. Die
// Leitung weist sie ab, und zwar in `service` — nicht hier, sonst hinge die
// Zusage an der Fläche.
//
// **Das Festhalten der Exemplare steht daneben** und nicht hier:
// **welche** Besitzzeilen weggehen, sagt ein Angebot nicht — es nennt einen
// Druck, keine Zeile. Das wählt ein Mensch über `HandelFesthalten`
// (#1643),
// und `vorbereiten` verlangt seit dem 2026-09-19, dass gedeckt ist, was von
// hier weggeht.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// HandelEroeffnen legt die erste Revision eines Vorgangs an.
//
// **Mit diesem Rechner als Initiator** — der Grund steht im Kopf.
func (a *App) HandelEroeffnen(
	ctx context.Context, wunsch Handelswunsch,
) HandelsvorgangAntwort {
	ctx = a.imProfil(ctx)

	ich, fehler := a.peerwerk().Ausweis(ctx)
	if fehler != nil {
		return handelsabsage(ctx, fehler)
	}

	mir := string(ich.Kennung())

	angebot := domain.Handelsangebot{
		Vertragsfassung: domain.Handelsvertragsfassung,
		Vorgang:         wunsch.Vorgang,
		Revision:        1,
		Art:             domain.Handelsart(wunsch.Art),
		Initiator:       mir,
		Parteien:        append([]string{mir}, wunsch.Gegenueber...),
		Zustand:         domain.HandelEntwurf,
		Positionen:      alsPositionen(wunsch.Positionen),

		BewertungGeteilt: wunsch.BewertungGeteilt,
	}

	angelegt, fehler := a.handelsleitung().Eroeffnen(ctx, angebot)
	if fehler != nil {
		return handelsabsage(ctx, fehler)
	}

	return HandelsvorgangAntwort{Vorgang: alsHandelsvorgangDTO(angelegt, a.eigeneKennung(ctx))}
}

// HandelVorschlagen trägt die nächste Revision ein.
//
// **Die Eltern werden gelesen und nicht genannt.** Ein Aufrufer, der den
// Elternabdruck mitschickte, nennte einen Stand, den er von irgendwoher hat;
// hier setzt der Vorschlag auf dem auf, was der Bestand **jetzt** führt —
// und ist er inzwischen weitergezogen, sagt es die Leitung (§7, Regel 2).
func (a *App) HandelVorschlagen(
	ctx context.Context, vorgang string, positionen []HandelspositionDTO,
) HandelsvorgangAntwort {
	ctx = a.imProfil(ctx)

	// **Vorschlagen tut dieser Rechner** — und aus `angeboten` darf das nur
	// die Gegenseite (#1975).
	ich, fehler := a.oderIch(ctx, "")
	if fehler != nil {
		return handelsabsage(ctx, fehler)
	}

	leitung := a.handelsleitung()

	heute, fehler := leitung.Stand(ctx, vorgang)
	if fehler != nil {
		return handelsabsage(ctx, fehler)
	}

	naechste, fehler := leitung.Vorschlagen(
		ctx, heute.Fortgeschrieben(alsPositionen(positionen)), ich)
	if fehler != nil {
		return handelsabsage(ctx, fehler)
	}

	return HandelsvorgangAntwort{Vorgang: alsHandelsvorgangDTO(naechste, a.eigeneKennung(ctx))}
}

// HandelSchritt geht einen Schritt an einem Vorgang.
//
// **Ein leeres `wer` ist dieser Rechner** — derselbe Weg wie bei der
// Verbuchung. Wer eine andere Partei nennt, trägt **deren** Stimme ein: über
// eine Leitung kommt die Zustimmung des Gegenübers an, und sie gehört
// eingetragen, ohne dass dieser Rechner sie wäre.
//
// **`gesehen` ist der Abdruck, den der Schritt meint** (#2496): ist der
// Vorgang inzwischen eine andere Revision, ist das ein Konflikt und kein
// Schritt. Leer heißt: nicht gefragt.
func (a *App) HandelSchritt(
	ctx context.Context, vorgang, schritt, wer, gesehen string,
) HandelsvorgangAntwort {
	ctx = a.imProfil(ctx)

	wer, fehler := a.oderIch(ctx, wer)
	if fehler != nil {
		return handelsabsage(ctx, fehler)
	}

	// **Wer dieser Rechner ist**, entscheidet, ob die Deckung hier zu prüfen
	// ist — die Stimme des Gegenübers hat ihre Exemplare drüben.
	gegangen, fehler := a.handelsleitung().SchreitenAls(
		ctx, vorgang, gesehen, domain.Handelsschritt(schritt), wer,
		a.eigeneKennung(ctx))
	if fehler != nil {
		return handelsabsage(ctx, fehler)
	}

	return HandelsvorgangAntwort{Vorgang: alsHandelsvorgangDTO(gegangen, a.eigeneKennung(ctx))}
}

// HandelStand gibt die letzte Revision eines Vorgangs — **rein lesend**.
func (a *App) HandelStand(
	ctx context.Context, vorgang string,
) HandelsvorgangAntwort {
	ctx = a.imProfil(ctx)

	stand, fehler := a.handelsleitung().Stand(ctx, vorgang)
	if fehler != nil {
		return handelsabsage(ctx, fehler)
	}

	return HandelsvorgangAntwort{Vorgang: alsHandelsvorgangDTO(stand, a.eigeneKennung(ctx))}
}

// HandelVorgaenge nennt die Handel mit einem Gegenüber — **rein lesend**,
// je Vorgang die letzte Revision, zuletzt geänderte zuerst (UX-08.AK-04,
// #1680).
// Wer einen Handel empfängt, wählt ihn hier, statt seine Kennung zu tippen.
func (a *App) HandelVorgaenge(
	ctx context.Context, gegenueber string,
) HandelsvorgaengeAntwort {
	ctx = a.imProfil(ctx)

	vorgaenge, fehler := a.handelsleitung().Vorgaenge(ctx, gegenueber)
	if fehler != nil {
		return HandelsvorgaengeAntwort{
			Vorgaenge: []HandelsvorgangDTO{}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	ich := a.eigeneKennung(ctx)

	dtos := make([]HandelsvorgangDTO, 0, len(vorgaenge))
	for _, vorgang := range vorgaenge {
		dtos = append(dtos, alsHandelsvorgangDTO(vorgang, ich))
	}

	return HandelsvorgaengeAntwort{Vorgaenge: dtos}
}

// oderIch gibt die genannte Partei — oder die eigene Kennung.
//
// **Der Ausweis wird nur geholt, wenn er gebraucht wird.** Er kommt aus dem
// Schlüsselbund, und der fragt zurück; ein Ruf, der die Partei ohnehin nennt,
// soll niemanden einen Dialog wegklicken lassen.
func (a *App) oderIch(ctx context.Context, wer string) (string, error) {
	if wer != "" {
		return wer, nil
	}

	ich, fehler := a.peerwerk().Ausweis(ctx)
	if fehler != nil {
		return "", fehler
	}

	return string(ich.Kennung()), nil
}
