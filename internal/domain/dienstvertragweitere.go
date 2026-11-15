package domain

// Die Verträge mit den drei neueren Diensten — Werkstück A1 von Phase 5.3
// (#1440). **Alle drei nachgelesen am 2026-09-02** (0056, und je Paket im
// Kommentar an `Adresse` und `Abstand`).
//
// # Drei Grade von Gewissheit
//
// **magicthegathering.io** hat einen Lauf gegen sein Original — und der hat
// beim ersten Mal die Grenze angestrichen: die Seite sagt 5.000 je Stunde,
// der Dienst schickt `ratelimit-limit: 1000`. **MTGJSON** und **Scrydex**
// haben keinen: für das eine braucht es eine Patreon-Marke, für das andere
// ein Konto samt Team, und beides gibt es hier nicht. Was in ihren Verträgen
// steht, ist die **Auskunft des Anbieters**, keine Messung — und der Katalog
// trägt für sie keinen Nachweis.

// weitereVertraege gibt die Verträge mit magicthegathering.io, MTGJSON und Scrydex.
func weitereVertraege() []Dienstvertrag {
	return []Dienstvertrag{
		{
			Dienst: DienstMTGAPI, Frage: FrageSuche, Wirkung: WirkungExternSenden,
			Basisadresse: "https://api.magicthegathering.io/v1",
			Endpunkte:    []string{"/cards"},
			Zugang:       "keiner — kein Schlüssel, kein Konto",
			Tarif:        "kostenlos",
			Grenze: "die Seite nennt 5.000 Anfragen je Stunde, der Dienst schickt " +
				"ratelimit-limit: 1000 (gemessen am 2026-09-02); gedrosselt auf eine " +
				"je fünf Sekunden",
			Frist:       "keine Zwischenspeicherung — jede Suche fragt",
			Bedingungen: "https://docs.magicthegathering.io/",
			Abgerufen:   "2026-09-02",
			Beleg:       "https://docs.magicthegathering.io/",
			// **Gelaufen am 2026-09-09** mit Fassung 0.4.0 (drei Tests, grün):
			// der Dienst antwortet, die Zuordnung liest ihn, keine unbekannten
			// Zustände. Wer den Lauf wiederholt, trägt Datum und Fassung nach.
			Nachweis: &Livenachweis{
				Lauf:      "internal/mtgapi/netzlauf_test.go",
				Datum:     "2026-09-09",
				Fassung:   "0.4.0",
				Endpunkte: []string{"/cards"},
			},
		},
		{
			Dienst: DienstMTGJSON, Frage: FrageSuche, Wirkung: WirkungExternSenden,
			Basisadresse: "https://graphql.mtgjson.com/",
			Endpunkte:    []string{"/"},
			Zugang: "authorization: Bearer <Marke> — die Marke gibt es nur für " +
				"Patreon-Unterstützer, die Schnittstelle ist im Beta",
			Tarif: "Patreon-Unterstützung",
			Grenze: "1.000 Anfragen je IP und Stunde, 500 je Marke und Stunde; " +
				"gedrosselt auf eine je zehn Sekunden",
			Frist:       "keine Zwischenspeicherung — jede Suche fragt",
			Bedingungen: "https://mtgjson.com/mtggraphql/",
			Abgerufen:   "2026-09-02",
			Beleg:       "https://mtgjson.com/mtggraphql/",
		},
		{
			Dienst: DienstScrydex, Frage: FrageSuche, Wirkung: WirkungExternSenden,
			Basisadresse: "https://api.scrydex.com",
			Endpunkte:    []string{"/magicthegathering/v1/cards"},
			Zugang: "zwei Kopfzeilen, X-Api-Key und X-Team-ID — Konto, Team und " +
				"Schlüssel nötig",
			Tarif: "Kontingente, die Geld kosten: jede Anfrage verbraucht Credits",
			Grenze: "100 Anfragen je Sekunde; gedrosselt auf eine je Sekunde, weil " +
				"die Kontingente die Grenze sind, die zählt",
			Frist:       "keine Zwischenspeicherung — jede Suche fragt",
			Bedingungen: "https://scrydex.com/docs/getting-started/authentication",
			Abgerufen:   "2026-09-02",
			Beleg:       "https://scrydex.com/docs/magicthegathering/cards",
		},
	}
}
