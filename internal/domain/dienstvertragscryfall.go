package domain

// Die Verträge mit Scryfall — Werkstück A1 von Phase 5.3 (#1440).
//
// **Nachgelesen am 2026-08-23 unter `https://scryfall.com/docs/api`**
// (`internal/scryfall/klient.go`, derselbe Tag): höchstens zehn Anfragen je
// Sekunde, für `/cards/search`, `/cards/named`, `/cards/random` und
// `/cards/collection` zwei; ein `User-Agent` und ein `Accept` sind Pflicht;
// Antworten sollen mindestens 24 Stunden zwischengespeichert werden.
// **Gedrosselt wird darunter** (§16, Regel 4): 125 ms und 600 ms.
//
// Die Bedingungen der Preise stehen in 0052 (nachgelesen am 2026-08-27):
// tägliche Schätzwerte ohne Gewähr, keine Weitergabe als Datensatz.
//
// # Der Nachweis
//
// `internal/scryfall/netzlauf_test.go` fährt Karte nach Name und Kennung,
// Regeln, Setangaben und die Symbologie gegen `api.scryfall.com`;
// `netzlaufschub_test.go` die Drucksuche und die Sammlung. **Datum und
// Fassung stehen am Lauf**, nicht am Dienst: die Bulk-Dateien und die Bilder
// ruft kein Lauf, und der Katalog sagt das — die zwei Verträge tragen keinen
// Nachweis.

const (
	scryfalladresse     = "https://api.scryfall.com"
	scryfallbedingungen = "https://scryfall.com/docs/api"
	scryfallabgerufen   = "2026-08-23"
	scryfallzugang      = "keiner — Pflicht sind ein User-Agent, der die Anwendung nennt, und ein Accept"
	scryfalltarif       = "kostenlos; die Daten dürfen nicht weiterverkauft, hinter eine Bezahlschranke gestellt oder als Datensatz weitergegeben werden"
	scryfallgrenze      = "höchstens 10 Anfragen je Sekunde, gedrosselt auf 8"
	scryfallkartenlimit = "höchstens 2 Anfragen je Sekunde auf den Kartenwegen, gedrosselt auf 1,7"
	scryfallfrist       = "24 Stunden — das Mindestmaß aus den Bedingungen"
	// scryfallbildgrenze: „The direct file origins located at *.scryfall.io"
	// haben laut Bedingungen keine der Grenzen — der User-Agent geht trotzdem mit.
	scryfallbildgrenze = "keine — die Dateiherkünfte *.scryfall.io sind von den Grenzen ausgenommen; die Bildadresse kommt aus der Kartenantwort"
)

// scryfallnachweis ist der letzte Lauf gegen den echten Dienst.
//
// **Ein Wert und nicht je Vertrag einer**: es ist ein Lauf. Welche
// Endpunkte er deckt, steht in ihm — und ein Vertrag, dessen Endpunkte er
// nicht ruft, trägt ihn nicht.
//
// **Gelaufen am 2026-09-09** mit Fassung 0.4.0: `netzlauf_test.go` (Karte
// nach Name und Kennung, Regeln, Setangaben, Symbologie),
// `netzlaufschub_test.go` (Drucksuche und Sammlung: 100 Drucke in 2
// Anfragen, 1,14 s) und `zulaessigkeitnetz_test.go`. **Der Schub war unter
// Last zweimal rot** (14,9 s und 7,8 s bei Lastmittel 10–25, rohes curl
// dabei 0,1–0,5 s) und auf leerer Maschine grün — die Abnahme »unter fünf
// Sekunden« misst die Maschine mit. Wer den Lauf wiederholt, trägt Datum
// und Fassung nach.
var scryfallnachweis = Livenachweis{
	Lauf:    "internal/scryfall/netzlauf_test.go",
	Datum:   "2026-09-09",
	Fassung: "0.4.0",
	Endpunkte: []string{
		"/cards/search", "/cards/named", "/cards/{id}", "/cards/collection",
		"/cards/{id}/rulings", "/sets/{code}", "/symbology",
	},
}

// scryfallvertraege gibt die Verträge mit Scryfall.
func scryfallvertraege() []Dienstvertrag {
	vertrag := func(frage Frage, grenze, beleg string, endpunkte ...string) Dienstvertrag {
		var nachweis *Livenachweis

		for _, endpunkt := range endpunkte {
			for _, gelaufen := range scryfallnachweis.Endpunkte {
				if endpunkt == gelaufen {
					kopie := scryfallnachweis
					nachweis = &kopie
				}
			}
		}

		return Dienstvertrag{
			Dienst: DienstScryfall, Frage: frage, Wirkung: WirkungExternSenden,
			Basisadresse: scryfalladresse, Endpunkte: endpunkte,
			Zugang: scryfallzugang, Tarif: scryfalltarif, Grenze: grenze,
			Frist: scryfallfrist, Bedingungen: scryfallbedingungen,
			Abgerufen: scryfallabgerufen, Beleg: beleg, Nachweis: nachweis,
		}
	}

	return []Dienstvertrag{
		vertrag(FrageSuche, scryfallkartenlimit,
			"https://scryfall.com/docs/api/cards/search", "/cards/search"),
		vertrag(FrageKartendaten, scryfallkartenlimit,
			"https://scryfall.com/docs/api/cards",
			"/cards/named", "/cards/{id}", "/cards/collection", "/cards/search"),
		vertrag(FrageBild, scryfallbildgrenze,
			"https://scryfall.com/docs/api/images", "*.scryfall.io"),
		vertrag(FrageRegeln, scryfallgrenze,
			"https://scryfall.com/docs/api/rulings", "/cards/{id}/rulings"),
		vertrag(FrageManakosten, scryfallgrenze,
			"https://scryfall.com/docs/api/card-symbols", "/symbology"),
		vertrag(FrageSetangaben, scryfallgrenze,
			"https://scryfall.com/docs/api/sets", "/sets/{code}"),
		vertrag(FragePreise, scryfallgrenze,
			"https://scryfall.com/docs/api/cards", "/cards/{id}"),
		vertrag(FrageSammeldaten, scryfallgrenze,
			"https://scryfall.com/docs/api/bulk-data", "/bulk-data", "data.scryfall.io"),
	}
}
