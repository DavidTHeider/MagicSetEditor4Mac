package domain

// **Wohin eine Karte außerhalb dieses Programms führt** — Werkstück `T8` von
// Phase 5.6
// (#1473).
//
// # Warum die Adressen hier stehen und nicht geraten werden
//
// **Weil jede von ihnen am 2026-09-19 aufgerufen wurde** (§16, Regel 1: „die
// Bedingungen sind nachgelesen, nicht erinnert"). Was dabei herauskam, steht
// in `docs/NETZEIGENSCHAFTEN.md` mit Datum; die Tabelle unten ist die kurze
// Fassung:
//
//	Anbieter      Produkt                        Suche
//	Cardmarket    — (idProduct trägt nicht)      /en/Magic/Products/Search
//	CardTrader    /cards/<blueprintId>           /en/search
//	TCGplayer     /product/<productId>           /search/magic/product
//	Card Kingdom  —                              /catalog/search
//	Gatherer      Details.aspx?multiverseid=     — (Referenz, kein Markt)
//	MTGGoldfish   —                              /q
//
// **Der Unterschied zwischen Produkt und Suche ist der Punkt.** Ein
// Produktlink führt auf **diese** Karte; ein Suchlink führt auf eine Liste,
// in der sie vermutlich steht. Wer beides gleich beschriftete, versprach eine
// Genauigkeit, die er nicht hat.
//
// # Und was ein Link nicht ist
//
// **Keine zugesicherte Verkaufsfunktion.** Ein Marktplatz darf eine Anmeldung
// verlangen, ein Angebot kann weg sein, ein Produkt kann umgezogen sein.
// „Ohne API-Schlüssel" heißt: dieses Programm braucht keinen, um die Adresse
// zu **bauen** — nicht, dass dort ohne Konto etwas zu kaufen wäre.

import (
	"fmt"
	"net/url"
	"strings"
)

// Linkanbieter ist, wohin ein Link führt.
type Linkanbieter string

// Die Anbieter, die dieses Programm kennt.
//
// **Eine Aufzählung und kein freier Text.** Eine Adresse, die aus einer
// Eingabe entsteht, ist ein Weg nach überallhin; was hier steht, ist
// nachgelesen und geprüft.
const (
	LinkCardmarket  Linkanbieter = "cardmarket"
	LinkCardTrader  Linkanbieter = "cardtrader"
	LinkTCGplayer   Linkanbieter = "tcgplayer"
	LinkCardKingdom Linkanbieter = "cardkingdom"
	LinkGatherer    Linkanbieter = "gatherer"
	LinkMTGGoldfish Linkanbieter = "mtggoldfish"
)

// Linkart sagt, wie genau ein Link trifft.
type Linkart string

const (
	// LinkProdukt führt auf **diesen** Druck.
	LinkProdukt Linkart = "produkt"

	// LinkSuche führt auf eine Liste, in der er stehen sollte.
	LinkSuche Linkart = "suche"

	// LinkStartseite führt zum Anbieter und sonst nirgendwohin.
	//
	// **Sie ist die ehrlichste Antwort, wenn es keine andere gibt** — und
	// sie sagt das auch: ein Suchlink ohne Suchwort wäre eine Auskunft, die
	// keine ist.
	LinkStartseite Linkart = "startseite"
)

// Linkherkunft sagt, woraus ein Link gebaut wurde.
type Linkherkunft string

const (
	// HerkunftKennung — aus einem belegten Anbieterbezeichner.
	HerkunftKennung Linkherkunft = "kennung"

	// HerkunftName — aus dem Kartennamen.
	HerkunftName Linkherkunft = "name"

	// HerkunftKeine — aus nichts; es ist die Startseite.
	HerkunftKeine Linkherkunft = "keine"
)

// Marktlink ist ein Weg nach draußen.
type Marktlink struct {
	Anbieter Linkanbieter
	Art      Linkart
	Herkunft Linkherkunft

	// Adresse ist vollständig und geprüft.
	Adresse string
}

// linkhosts sind die Wirte, auf die ein Link zeigen darf.
//
// **Auf den Wirt genau und nicht auf ein Suffix.** Ein Vergleich mit
// `strings.HasSuffix` ließe `cardmarket.com.beispiel.tld` durch — genau der
// Angriff, gegen den die Liste steht.
var linkhosts = map[Linkanbieter]string{
	LinkCardmarket:  "www.cardmarket.com",
	LinkCardTrader:  "www.cardtrader.com",
	LinkTCGplayer:   "www.tcgplayer.com",
	LinkCardKingdom: "www.cardkingdom.com",
	LinkGatherer:    "gatherer.wizards.com",
	LinkMTGGoldfish: "www.mtggoldfish.com",
}

// Linkanbieterliste gibt die Anbieter in ihrer Reihenfolge.
func Linkanbieterliste() []Linkanbieter {
	return []Linkanbieter{
		LinkCardmarket, LinkCardTrader, LinkTCGplayer, LinkCardKingdom,
		LinkGatherer, LinkMTGGoldfish,
	}
}

// LinkhostFuer gibt den Wirt eines Anbieters — leer, wenn es keinen gibt.
func LinkhostFuer(anbieter Linkanbieter) string { return linkhosts[anbieter] }

// PruefeMarktlink weist ab, was nicht hinausgehen darf.
//
// # Die fünf Bedingungen, und warum jede von ihnen
//
//  1. **HTTPS.** Ein `http`-Link ginge im Klartext; ein `file`- oder
//     `javascript`-Link ginge gar nicht ins Netz, sondern in dieses
//     Programm.
//  2. **Der Wirt ist der erlaubte, Zeichen für Zeichen.** Kein Suffix, kein
//     Präfix, keine Unterdomäne: `www.cardmarket.com.beispiel.tld` gehört
//     jemand anderem, und `evil-cardmarket.com` erst recht.
//  3. **Keine Benutzerangabe.** `https://opfer@boese.example/` zeigt in der
//     Adresszeile den Teil vor dem `@` — das ist der ganze Trick.
//  4. **Kein eigener Port.** Ein Marktplatz auf `:8080` ist keiner.
//  5. **Kein Geheimnis in der Abfrage.** Was in einer Adresse steht, steht
//     im Verlauf des Browsers, im Protokoll des Anbieters und in jeder
//     Weiterleitung.
func PruefeMarktlink(link Marktlink) error {
	wirt := linkhosts[link.Anbieter]
	if wirt == "" {
		return fmt.Errorf("%w: %q ist kein Anbieter, den dieses Programm kennt",
			ErrValidation, link.Anbieter)
	}

	gelesen, fehler := url.Parse(link.Adresse)
	if fehler != nil {
		return MitGrund(ErrValidation, fehler, "die Adresse lesen")
	}

	switch {
	case gelesen.Scheme != "https":
		return fmt.Errorf("%w: %q ist nicht https", ErrValidation, link.Adresse)

	case gelesen.Host != wirt:
		return fmt.Errorf("%w: %q liegt nicht auf %s", ErrValidation,
			gelesen.Host, wirt)

	case gelesen.User != nil:
		return fmt.Errorf("%w: eine Adresse mit Benutzerangabe geht nicht "+
			"hinaus — sie zeigt in der Adresszeile etwas anderes, als sie "+
			"aufruft", ErrValidation)
	}

	return pruefeAbfrage(gelesen)
}

// pruefeAbfrage weist eine Abfrage ab, die ein Geheimnis tragen könnte.
//
// **Auf den Namen des Feldes und nicht auf seinen Wert.** Ein Geheimnis
// erkennt man nicht an seiner Gestalt — ein Schlüssel sieht aus wie ein
// Suchwort. Was es verrät, ist der Name, unter dem es steht.
func pruefeAbfrage(gelesen *url.URL) error {
	verboten := []string{
		"token", "key", "secret", "password", "passwort", "auth",
		"session", "api_key", "apikey", "signature",
	}

	for name := range gelesen.Query() {
		klein := strings.ToLower(name)

		for _, wort := range verboten {
			if strings.Contains(klein, wort) {
				return fmt.Errorf(
					"%w: %q gehört in keine Adresse, die hinausgeht",
					ErrValidation, name)
			}
		}
	}

	return nil
}
