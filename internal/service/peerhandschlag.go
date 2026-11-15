package service

// **Der Handschlag über eine Verbindung** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Warum diese Datei erst jetzt entsteht
//
// **Weil N1 zwei Hälften gebaut hat, die sich nicht kannten.** N1.3 hat den
// Handschlag, N1.5 den Lebenszyklus — und `Oeffne` prüfte das Recht und
// wählte, mehr nicht. Eine [peer.Sitzung] entstand oberhalb von
// `internal/peer` in **keiner** Schicht. Sichtbar wurde das, als der erste
// Transport da war: ohne ihn gab es nichts, worüber der Wechsel hätte laufen
// können.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// handschlagFuehren führt die anbietende Seite durch die drei Schritte.
//
// # Die Reihenfolge ist der Beleg
//
//  1. Angebot hinaus — mit **unserer** Zufallszahl und dem Kanal dieser
//     Verbindung.
//  2. Annahme herein — sie trägt unsere Zahl zurück und ist vom Gegenüber
//     unterschrieben. Erst damit steht fest, dass dort **jetzt** jemand ist.
//  3. Abschluss hinaus — er trägt **seine** Zahl zurück, und damit weiß auch
//     er es.
//
// **`Erwartet` steht im Wunsch.** Wer wählt, weiß, wen er meint; eine Antwort
// von jemand anderem endet dann im Handschlag und nicht erst in der Sitzung.
func (p *Peersitzungen) handschlagFuehren(
	ctx context.Context, ich peer.Identitaet, verbindung Verbindung,
	kennung peer.Kennung, zweck domain.Zweck,
) (peer.Sitzung, error) {
	anbahnung, angebot, fehler := ich.Anbieten(peer.Wunsch{
		Zweck:    zweck,
		Rolle:    RolleAnbietend,
		Kanal:    verbindung.Kanal(),
		Erwartet: kennung,
	}, p.zufall)
	if fehler != nil {
		return peer.Sitzung{}, fehler
	}

	if befund := verbindung.Sende(ctx, angebot); befund != nil {
		return peer.Sitzung{}, befund
	}

	annahme, fehler := verbindung.Empfange(ctx)
	if fehler != nil {
		return peer.Sitzung{}, fehler
	}

	sitzung, abschluss, fehler := anbahnung.Schliessen(annahme)
	if fehler != nil {
		return peer.Sitzung{}, fehler
	}

	if befund := eineIdentitaet(verbindung, sitzung.Gegenueber); befund != nil {
		return peer.Sitzung{}, befund
	}

	if befund := verbindung.Sende(ctx, abschluss); befund != nil {
		return peer.Sitzung{}, befund
	}

	return sitzung, nil
}

// Annehmen nimmt einen Handschlag entgegen, der hereinkommt.
//
// # Warum das Recht **zwischen** Annahme und Antwort geprüft wird
//
// **Weil vorher niemand weiß, wer da ist.** Erst das Angebot sagt, welche
// Kennung und welcher Zweck — und erst dann lässt sich fragen, ob dieses
// Profil das erlaubt hat.
//
// **Und die Antwort geht danach hinaus, nicht davor.** Wer zuerst antwortete,
// hätte sich einem Fremden gegenüber ausgewiesen, dem er nichts erlaubt: der
// wüsste damit, dass es dieses Profil gibt und unter welcher Kennung.
//
// # Warum die Verbindung hereingereicht wird
//
// **Weil ein Wähler wählt.** Bis zu diesem Schnitt rief `Annehmen` selbst
// `Waehle(ctx, "", "")` — mit leerer Kennung und leerem Protokoll, weil es
// beides nicht kennt. Das war ein Wähler, der als Horcher benutzt wurde, und
// die zwei leeren Zeichenketten haben es gesagt. Wer horcht, ist der
// Transport; er nimmt an und reicht herein.
func (p *Peersitzungen) Annehmen(
	ctx context.Context, verbindung Verbindung,
) (*Sitzungsgriff, error) {
	ich, fehler := p.ausweis.Ausweis(ctx)
	if fehler != nil {
		return nil, fehler
	}

	p.schloss.Lock()
	epoche := p.epoche
	p.schloss.Unlock()

	griff, fehler := p.entgegennehmen(ctx, ich, verbindung, epoche)
	if fehler != nil {
		_ = verbindung.Schliesse() //nolint:errcheck // der Weg endet hier ohnehin

		return nil, fehler
	}

	if fehler := p.eintragen(griff); fehler != nil {
		return nil, fehler
	}

	return griff, nil
}

// entgegennehmen führt die annehmende Seite durch die drei Schritte.
func (p *Peersitzungen) entgegennehmen(
	ctx context.Context, ich peer.Identitaet, verbindung Verbindung,
	epoche uint64,
) (*Sitzungsgriff, error) {
	angebot, fehler := verbindung.Empfange(ctx)
	if fehler != nil {
		return nil, fehler
	}

	entgegennahme, annahme, fehler := ich.Entgegennehmen(angebot,
		peer.Bereitschaft{Rolle: RolleAnnehmend, Kanal: verbindung.Kanal()},
		p.zufall, p.gesehen)
	if fehler != nil {
		return nil, fehler
	}

	// **Eine Identität für beide Ebenen** (SEC-16) — vor dem Recht: wer
	// mit einem Zeugnis anklopft und mit einer anderen Kennung unterschreibt,
	// ist keiner der Kontakte, sondern zwei.
	if befund := eineIdentitaet(verbindung, entgegennahme.Gegenueber()); befund != nil {
		return nil, befund
	}

	// **Hier und nicht früher, und hier und nicht später.**
	if befund := p.darfEintreten(ctx, entgegennahme); befund != nil {
		return nil, befund
	}

	if befund := verbindung.Sende(ctx, annahme); befund != nil {
		return nil, befund
	}

	abschluss, fehler := verbindung.Empfange(ctx)
	if fehler != nil {
		return nil, fehler
	}

	sitzung, fehler := entgegennahme.Abschliessen(abschluss)
	if fehler != nil {
		return nil, fehler
	}

	return p.griffAus(verbindung, sitzung,
		string(sitzung.Gegenueber)+"/"+string(sitzung.Zweck), epoche), nil
}

// darfEintreten fragt die Kontakte, ob dieses Gegenüber diesen Zweck darf.
func (p *Peersitzungen) darfEintreten(
	ctx context.Context, entgegennahme peer.Entgegennahme,
) error {
	kennung, zweck := entgegennahme.Gegenueber(), entgegennahme.Zweck()

	darf, fehler := p.rechte.Darf(ctx, kennung, zweck)
	if fehler != nil {
		return fehler
	}

	if !darf {
		return fmt.Errorf(
			"%w: %s darf %q nicht", domain.ErrValidation,
			peer.Fingerabdruck(kennung), zweck)
	}

	return nil
}

// Sitzung gibt, worauf sich beide Seiten geeinigt haben.
func (g *Sitzungsgriff) Sitzung() peer.Sitzung { return g.sitzung }

// eineIdentitaet hält die Kennung des Handschlags gegen die des Zeugnisses —
// SEC-16 (#2145,
// Vermerk 0176).
//
// **Ein Zeugnis und eine Unterschrift von zwei Schlüsseln** heißt: wer da ist,
// hält beide oder benutzt einen fremden als Orakel. Bis dahin prüfte niemand
// beides gegeneinander — die horchende Seite verlangte ein Zeugnis
// (`RequireAnyClientCert`) und ließ dann jede gültige Unterschrift durch.
//
// **Ohne Zeugnis gibt es nichts zu vergleichen** — eine Röhre im selben
// Prozess; dort sagt allein der Handschlag, wer da ist.
func eineIdentitaet(verbindung Verbindung, handschlag peer.Kennung) error {
	zeugnis := verbindung.Gegenstelle()
	if zeugnis == "" || zeugnis == handschlag {
		return nil
	}

	return fmt.Errorf("%w: das Zeugnis gehört %s, der Handschlag %s — eine Verbindung hat eine Identität",
		domain.ErrValidation, peer.Fingerabdruck(zeugnis), peer.Fingerabdruck(handschlag))
}
