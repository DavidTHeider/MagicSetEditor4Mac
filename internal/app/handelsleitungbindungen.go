package app

import "context"

// **Der Handelsvorgang am Fenster** —
// #1637.
//
// **Neun Bindungen und nicht eine.** Eröffnen, Vorschlagen und Schreiten
// sind drei Entscheidungen eines Menschen; der Stand ist die Auskunft
// darüber, wo der Vorgang steht, und die Vorgänge sind die Auskunft, welche
// es mit einem Gegenüber gibt; Schicken, Bestätigen und Wiederaufnahme
// gehören dem Ausgang — und der ist eine andere Frage als der Vorgang.
//
// # Die neunte kam mit UX-08.AK-04
//
// **Vorher kannte das Fenster nur den Handel, dessen Kennung jemand
// tippte** (#1680).
// Wer einen Handel empfing, erfuhr sie nirgends.
//
// # Die achte kam später, und sie schloss ein Loch
//
// **Zustellen trägt, was im Ausgang liegt** — und ohne sie konnte ein
// Fenster einen Handel **annehmen, aber keinen hinausbringen**
// (#1634).
// `HandelSchicken` legt in den Ausgang; getragen hat das niemand. Ein
// Angebot lag damit für immer da, und der, dem es galt, sah nie etwas.

// HandelEroeffnen legt die erste Revision eines Vorgangs an.
func (b *Bindungen) HandelEroeffnen(
	ctx context.Context,
	wunsch Handelswunsch,
) HandelsvorgangAntwort {
	return b.app(ctx).HandelEroeffnen(b.ruf(ctx), wunsch)
}

// HandelVorschlagen trägt die nächste Revision ein.
func (b *Bindungen) HandelVorschlagen(
	ctx context.Context,
	vorgang string, positionen []HandelspositionDTO,
) HandelsvorgangAntwort {
	return b.app(ctx).HandelVorschlagen(b.ruf(ctx), vorgang, positionen)
}

// HandelSchritt geht einen Schritt an einem Vorgang — an der Revision, die
// `gesehen` nennt (#2496).
func (b *Bindungen) HandelSchritt(
	ctx context.Context,
	vorgang, schritt, wer, gesehen string,
) HandelsvorgangAntwort {
	return b.app(ctx).HandelSchritt(b.ruf(ctx), vorgang, schritt, wer, gesehen)
}

// HandelStand gibt die letzte Revision eines Vorgangs — **rein lesend**.
func (b *Bindungen) HandelStand(ctx context.Context, vorgang string) HandelsvorgangAntwort {
	return b.app(ctx).HandelStand(b.ruf(ctx), vorgang)
}

// HandelVorgaenge nennt die Handel mit einem Gegenüber — **rein lesend**.
func (b *Bindungen) HandelVorgaenge(ctx context.Context, gegenueber string) HandelsvorgaengeAntwort {
	return b.app(ctx).HandelVorgaenge(b.ruf(ctx), gegenueber)
}

// HandelSchicken legt den heutigen Stand in den Ausgang.
func (b *Bindungen) HandelSchicken(
	ctx context.Context,
	vorgang, empfaenger string,
) VersandAntwort {
	return b.app(ctx).HandelSchicken(b.ruf(ctx), vorgang, empfaenger)
}

// HandelPostbestaetigung räumt eine Nachricht weg, die angekommen ist.
func (b *Bindungen) HandelPostbestaetigung(ctx context.Context, nachrichtID string) VersandAntwort {
	return b.app(ctx).HandelPostbestaetigung(b.ruf(ctx), nachrichtID)
}

// HandelWiederaufnahme sagt, was nach einem Abriss noch einmal hinausmuss.
func (b *Bindungen) HandelWiederaufnahme(
	ctx context.Context,
	vorgang, empfaenger string, seiner HandelsfortschrittDTO,
) WiederaufnahmeAntwort {
	return b.app(ctx).HandelWiederaufnahme(
		b.ruf(ctx), vorgang, empfaenger, seiner)
}

// HandelZustellen trägt den Ausgang zu einem Gegenüber über die Leitung.
//
// **Getrennt von `HandelSchicken`, und das ist die ganze Trennung**: jenes
// entscheidet, **was** hinausgeht, dieses baut die Leitung und trägt, **was
// dasteht**. Wer beides in einen Knopf legte, verlöre den Fall, für den es
// den Ausgang gibt — dass drüben gerade niemand da ist.
func (b *Bindungen) HandelZustellen(ctx context.Context, kennung string) Zustellungsbericht {
	return b.app(ctx).Handelszustellung(b.ruf(ctx), kennung)
}

// HandelSichtAnfragen fragt ein Gegenüber über die Leitung nach dem, was es
// diesem Profil zum Handeln freigibt (P11, #2379).
func (b *Bindungen) HandelSichtAnfragen(ctx context.Context, kennung string) FreigabesichtAntwort {
	return b.app(ctx).HandelSichtAnfragen(b.ruf(ctx), kennung)
}

// HandelFesthalten hält eine benannte Besitzzeile für einen Vorgang fest.
//
// **Vier Werte und kein Verbund**, anders als beim Eröffnen: ein Mensch
// benennt **eine** Zeile, und ein Verbund mit vier Feldern wäre eine Hülle
// um vier Werte.
func (b *Bindungen) HandelFesthalten(
	ctx context.Context,
	vorgang string, exemplarID int64, menge int, fassung int64,
) ZusagenAntwort {
	return b.app(ctx).HandelFesthalten(
		b.ruf(ctx), vorgang, exemplarID, menge, fassung)
}

// Handelszusagen gibt, was für einen Vorgang festgehalten ist.
func (b *Bindungen) Handelszusagen(ctx context.Context, vorgang string) ZusagenAntwort {
	return b.app(ctx).Handelszusagen(b.ruf(ctx), vorgang)
}

// HandelUebersicht gibt die Vorgänge dieses Profils — gefiltert und
// geordnet (TRADE-01, #1715).
func (b *Bindungen) HandelUebersicht(ctx context.Context, filter HandelsfilterDTO) HandelsuebersichtAntwort {
	return b.app(ctx).HandelUebersicht(b.ruf(ctx), filter)
}

// HandelEntwurfAendern schreibt einen Entwurf neu.
func (b *Bindungen) HandelEntwurfAendern(ctx context.Context, wunsch Handelswunsch) HandelsvorgangAntwort {
	return b.app(ctx).HandelEntwurfAendern(b.ruf(ctx), wunsch)
}

// HandelKopieren legt aus einem Vorgang einen neuen Entwurf an.
func (b *Bindungen) HandelKopieren(ctx context.Context, vorgang string) HandelsvorgangAntwort {
	return b.app(ctx).HandelKopieren(b.ruf(ctx), vorgang)
}

// HandelEntwurfLoeschen löscht einen Entwurf.
func (b *Bindungen) HandelEntwurfLoeschen(ctx context.Context, vorgang string) SchreibAntwort {
	return b.app(ctx).HandelEntwurfLoeschen(b.ruf(ctx), vorgang)
}

// HandelArchivieren legt einen beendeten Vorgang ins Archiv oder holt ihn
// heraus.
func (b *Bindungen) HandelArchivieren(ctx context.Context, vorgang string, an bool) SchreibAntwort {
	return b.app(ctx).HandelArchivieren(b.ruf(ctx), vorgang, an)
}

// HandelProtokoll gibt, wie ein Vorgang dorthin kam, wo er steht — **rein
// lesend** (TRADE-03.AK-03, #1717).
func (b *Bindungen) HandelProtokoll(ctx context.Context, vorgang string) HandelsprotokollAntwort {
	return b.app(ctx).HandelProtokoll(b.ruf(ctx), vorgang)
}

// HandelPreisanhalt rechnet, was die Seiten eines Entwurfs ungefähr wert
// sind — **rein lesend** (TRADE-02b, #1951).
func (b *Bindungen) HandelPreisanhalt(ctx context.Context, positionen []HandelspositionDTO) PreisanhaltAntwort {
	return b.app(ctx).HandelPreisanhalt(b.ruf(ctx), positionen)
}
