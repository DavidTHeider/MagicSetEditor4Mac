package domain

// **Die Verträge für Preise** — Werkstück A4 von Phase 5.3
// (#1443).
//
// # Zwei Verträge, und einer ist eine Nachtragung
//
// **CardTrader** kommt mit diesem Werkstück dazu: der erste **Marktplatz**
// hier, und die einzige Quelle, die Angebote einzelner Verkäufer nennt.
//
// **MTGJSON hätte seinen schon in A3 bekommen sollen.** Dort ist der Bulkbezug
// der Preise entstanden (#1442),
// und der Katalog sagte weiter, MTGJSON könne hier „suchen und sonst nichts".
// Das ist genau die Lücke, die der Katalog verhindern soll: er ist die
// **Auskunft an Nutzende** darüber, was ein Dienst hier tut und unter welchen
// Bedingungen — und eine Auskunft, die zu wenig sagt, ist so falsch wie eine,
// die zu viel verspricht (§10). Aufgefallen beim Eintragen von CardTrader.

// marktvertraege gibt die Verträge für [FragePreise] neben Scryfall.
func marktvertraege() []Dienstvertrag {
	return []Dienstvertrag{
		{
			Dienst: DienstMTGJSON, Frage: FragePreise, Wirkung: WirkungExternSenden,
			Basisadresse: "https://mtgjson.com/api/v5",
			Endpunkte:    []string{"AllPricesToday.json.gz", "AllPrices.json.gz"},
			Zugang:       "keiner — die Dateien liegen offen",
			Tarif:        "kostenlos, MIT-Lizenz",
			Grenze: "die Seite nennt keine Anfragegrenze (abgerufen 2026-09-11); " +
				"die Grenze ist die Größe: AllPrices wiegt ausgepackt über ein " +
				"Gigabyte, und geholt wird nur auf ausdrückliche Anforderung",
			Frist: "die Dateien entstehen einmal täglich — „Builds kick off at " +
				"1:00AM EST and go live at 9:00AM EST“; ein zweiter Abruf am " +
				"selben Tag bringt dasselbe",
			Bedingungen: "https://mtgjson.com/faq/",
			Abgerufen:   "2026-09-11",
			Beleg:       "https://mtgjson.com/downloads/all-files/",
		},
		{
			Dienst: DienstScrydex, Frage: FragePreise, Wirkung: WirkungExternSenden,
			Basisadresse: "https://api.scrydex.com",
			Endpunkte: []string{
				"/magicthegathering/v1/cards", "/price_history",
				"/account/v1/usage",
			},
			Zugang: "zwei Kopfzeilen, X-Api-Key und X-Team-ID — Konto, Team und " +
				"Schlüssel nötig",
			Tarif: "Kontingente, die Geld kosten: jede Anfrage verbraucht Credits, " +
				"und über den Tarif hinaus fallen Gebühren an",
			Grenze: "100 Anfragen je Sekunde; gedrosselt auf eine je Sekunde, weil " +
				"die Kontingente die Grenze sind, die zählt. Der Reststand steht " +
				"unter /account/v1/usage und ist dort bis zu 30 Minuten alt — " +
				"eine Sperre ist er damit nicht",
			Frist: "keine Zwischenspeicherung; der Verlauf trägt je Zeile den Tag, " +
				"den die Quelle nennt",
			Bedingungen: "https://scrydex.com/docs/getting-started/authentication",
			Abgerufen:   "2026-09-11",
			Beleg:       "https://scrydex.com/docs/magicthegathering/price-history",
		},
		{
			Dienst: DienstCardTrader, Frage: FragePreise, Wirkung: WirkungExternSenden,
			Basisadresse: "https://api.cardtrader.com/api/v2",
			Endpunkte: []string{
				"/games", "/expansions", "/blueprints/export", "/marketplace/products",
			},
			Zugang: "authorization: Bearer <Marke> — die Marke steht in den " +
				"Profileinstellungen eines CardTrader-Kontos",
			Tarif: "kostenlos zu lesen; was über den Marktplatz gekauft wird, " +
				"kostet — dieses Programm kauft nicht",
			Grenze: "200 Anfragen je 10 Sekunden über alle Endpunkte, am " +
				"Marktplatz 10 je Sekunde; gedrosselt auf eine je Sekunde " +
				"(0095)",
			Frist: "keine Zwischenspeicherung — ein Marktpreis ist eine " +
				"Momentaufnahme, und eine alte sähe aus wie eine neue",
			Bedingungen: "https://www.cardtrader.com/en/docs/api/full/reference",
			Abgerufen:   "2026-09-11",
			Beleg:       "https://www.cardtrader.com/en/docs/api/full/reference",
		},
		{
			// **Das eigene Konto, nur lesend** — Werkstück T11 von Phase 5.6
			// (#1476).
			//
			// **Ein zweiter Vertrag und kein größerer erster**: die
			// Marktwege fragen nach **fremden** Angeboten, diese hier nach
			// **eigenen** — andere Daten, andere Folgen, und deshalb ein
			// eigener Schalter ([0122](../../docs/entscheidungen/0122-ein-fremdes-konto-wird-gelesen.md)).
			Dienst: DienstCardTrader, Frage: FrageKonto,
			Wirkung:      WirkungExternSenden,
			Basisadresse: "https://api.cardtrader.com/api/v2",
			Endpunkte:    []string{"/info", "/products/export", "/orders"},
			Zugang: "authorization: Bearer <Marke> — dieselbe Marke wie am " +
				"Marktplatz, aus den Profileinstellungen des Kontos",
			Tarif: "kostenlos zu lesen; dieses Programm schreibt dort nichts " +
				"und kauft nichts",
			Grenze: "200 Anfragen je 10 Sekunden über alle Endpunkte; " +
				"gedrosselt auf eine je Sekunde. Der Abruf der eigenen " +
				"Angebote kann Sekunden dauern — der Anbieter empfiehlt eine " +
				"Lesefrist von 120 bis 180 Sekunden",
			Frist: "keine Zwischenspeicherung — was im eigenen Konto steht, " +
				"ändert sich, während man hinsieht",
			Bedingungen: "https://www.cardtrader.com/en/docs/api/full/reference",
			Abgerufen:   "2026-09-19",
			Beleg:       "https://www.cardtrader.com/en/docs/api/full/reference",
		},
	}
}
