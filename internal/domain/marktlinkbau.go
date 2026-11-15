package domain

// **Wie ein Marktlink entsteht** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// # Die Reihenfolge, und warum sie eine ist
//
// Das Issue nennt sie: „validierter Produktlink aus A2/A3, belegter
// Provider-Identifier, genau definierter Suchlink, Anbieterstartseite mit
// ehrlicher Beschriftung." Je Anbieter wird von oben nach unten genommen, was
// zu haben ist — und **nur eines**: drei Links auf denselben Anbieter sind
// drei Wege zu derselben Frage, und der genaueste ist die Antwort.
//
// # Und warum ein fehlender Name keine Suche ergibt
//
// **Weil eine Suche ohne Suchwort keine ist.** Sie führte auf die Liste
// **aller** Karten, und das sieht aus wie ein Treffer. Wer keinen Namen hat,
// bekommt die Startseite — mit der Beschriftung, die dazu gehört.

import (
	"net/url"
	"strings"
)

// Marktlinks baut je Anbieter den genauesten Weg, der zu haben ist.
//
// **Die Kennungen kommen aus `A2`** (die Druckzeile dieses Bestands), der
// Name aus derselben Zeile. Was nicht dasteht, wird nicht erfunden.
func Marktlinks(kennungen []Druckkennung, name string) []Marktlink {
	links := make([]Marktlink, 0, len(Linkanbieterliste()))

	for _, anbieter := range Linkanbieterliste() {
		links = append(links, linkFuer(anbieter, kennungen, name))
	}

	return links
}

// linkFuer baut den einen Link eines Anbieters.
func linkFuer(
	anbieter Linkanbieter, kennungen []Druckkennung, name string,
) Marktlink {
	if adresse := produktlink(anbieter, kennungen); adresse != "" {
		return Marktlink{
			Anbieter: anbieter, Art: LinkProdukt,
			Herkunft: HerkunftKennung, Adresse: adresse,
		}
	}

	if adresse := suchlink(anbieter, name); adresse != "" {
		return Marktlink{
			Anbieter: anbieter, Art: LinkSuche,
			Herkunft: HerkunftName, Adresse: adresse,
		}
	}

	return Marktlink{
		Anbieter: anbieter, Art: LinkStartseite, Herkunft: HerkunftKeine,
		Adresse: "https://" + LinkhostFuer(anbieter) + "/",
	}
}

// produktlink baut den Weg auf **diesen** Druck — oder gibt nichts.
//
// # Warum Cardmarket hier fehlt
//
// **Weil seine Kennung nicht trägt.** Gemessen am 2026-09-19:
// `…/Products/Singles?idProduct=241866` landet auf der Sammelseite und nicht
// auf der Karte; der Weg dorthin geht über einen Namensschlüssel, den dieses
// Programm nicht führt. Ein Produktlink, der keiner ist, wäre schlimmer als
// keiner — deshalb bekommt Cardmarket die **Suche**.
//
// # Und warum Gatherer hier steht
//
// **Weil es der einzige Weg dorthin ist.** Gatherer ist kein Marktplatz,
// sondern die Regelauskunft des Herausgebers; es kennt eine Karte an ihrer
// **Multiverse-ID** und an nichts sonst. Ohne sie ist es **nicht verfügbar** —
// das Issue sagt es wörtlich, und eine Suche wäre die Erfindung, die es
// ausschließt.
func produktlink(anbieter Linkanbieter, kennungen []Druckkennung) string {
	wert := kennungswert(anbieter, kennungen)
	if wert == "" {
		return ""
	}

	switch anbieter {
	case LinkCardTrader:
		return "https://www.cardtrader.com/cards/" + url.PathEscape(wert)

	case LinkTCGplayer:
		return "https://www.tcgplayer.com/product/" + url.PathEscape(wert)

	case LinkGatherer:
		return "https://gatherer.wizards.com/Pages/Card/Details.aspx?" +
			url.Values{"multiverseid": {wert}}.Encode()
	}

	return ""
}

// kennungssystem sagt, unter welchem System ein Anbieter seine Kennung führt.
func kennungssystem(anbieter Linkanbieter) Kennungssystem {
	switch anbieter {
	case LinkCardTrader:
		return SystemCardTrader
	case LinkTCGplayer:
		return SystemTCGplayer
	case LinkGatherer:
		return SystemGatherer
	case LinkCardmarket:
		return SystemCardmarket
	}

	return ""
}

// kennungswert sucht die Produktkennung eines Anbieters.
//
// **Nur `SemantikProdukt` und `SemantikDruck`.** Ein Konzept nennt mehrere
// Drucke, ein Kunstwerk noch mehr; wer damit einen Produktlink baute, führte
// auf irgendeinen davon.
func kennungswert(anbieter Linkanbieter, kennungen []Druckkennung) string {
	system := kennungssystem(anbieter)
	if system == "" {
		return ""
	}

	for _, kennung := range kennungen {
		passt := kennung.Semantik == SemantikProdukt ||
			kennung.Semantik == SemantikDruck

		if kennung.System == system && passt && kennung.Wert != "" {
			return kennung.Wert
		}
	}

	return ""
}

// suchlink baut den Weg auf eine Liste — oder gibt nichts.
func suchlink(anbieter Linkanbieter, name string) string {
	geglaettet := strings.TrimSpace(name)
	if geglaettet == "" {
		return ""
	}

	switch anbieter {
	case LinkCardmarket:
		return "https://www.cardmarket.com/en/Magic/Products/Search?" +
			url.Values{"searchString": {geglaettet}}.Encode()

	case LinkCardTrader:
		return "https://www.cardtrader.com/en/search?" +
			url.Values{"q": {geglaettet}}.Encode()

	case LinkTCGplayer:
		return "https://www.tcgplayer.com/search/magic/product?" +
			url.Values{"q": {geglaettet}}.Encode()

	case LinkCardKingdom:
		return "https://www.cardkingdom.com/catalog/search?" +
			url.Values{"search": {"header"}, "filter[name]": {geglaettet}}.Encode()

	case LinkMTGGoldfish:
		return "https://www.mtggoldfish.com/q?" +
			url.Values{"query_string": {geglaettet}}.Encode()
	}

	// **Gatherer sucht nicht.** Es ist eine Referenz und kein Marktplatz;
	// ohne Multiverse-ID ist es nicht verfügbar.
	return ""
}
