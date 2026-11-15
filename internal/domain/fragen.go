package domain

import "slices"

// **Die Fragen, die ein fremder Dienst beantwortet** — Werkstück 8 von
// Phase 4.8 (#1321),
// entschieden in [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md),
// Entscheidung 1.
//
// # Warum es sie neben [Dienst] gibt
//
// **Weil ein Schalter und eine Wahl zwei verschiedene Aussagen sind.** „Dieser
// Dienst darf gefragt werden" ist die eine; „für diese Frage nimm jenen
// Dienst" ist die andere. Wer nur Schalter hätte, wüsste bei zwei
// eingeschalteten Diensten nicht, wer antwortet — und die Antwort entschiede
// die Reihenfolge im Code.
//
// # Warum es heute nur eine Frage mit Wahl gibt
//
// **Weil eine Wahl mit einer Möglichkeit keine ist.** Bilder, Setangaben,
// Regeln, Manakosten und Sammeldaten beantwortet heute nur Scryfall — sie
// stehen als Fragen im Vertragskatalog (`dienstvertrag.go`), aber ohne
// Einstellung. Sobald ein zweiter Dienst einen Vertrag dafür hat, hat die
// Frage eine Wahl — und bekommt Schalter, Vorgabe und Prüfung aus derselben
// Quelle.
//
// **Und eine Frage, die alle zugleich beantworten, bekommt trotzdem keine** —
// siehe [Frage.Mehrstimmig]. Die **Preise** haben seit A4 drei Dienste und
// trotzdem keine Wahl.

// Frage ist etwas, das ein fremder Dienst beantworten kann.
type Frage string

// FrageSuche ist „finde Karten zu einem Text".
const FrageSuche Frage = "suche"

// Mehrstimmig sagt, ob **alle** eingeschalteten Dienste diese Frage
// nebeneinander beantworten — statt einer anstelle der anderen.
//
// # Warum das nicht dasselbe ist wie »mehrere können es«
//
// Bei der **Suche** antwortet genau einer: wer „Lightning Bolt" tippt,
// bekommt eine Trefferliste, und zwei Dienste ergäben zwei Listen, zwischen
// denen niemand entschieden hat. Dafür gibt es die Wahl.
//
// Beim **KI-Vorschlag** ebenso, und aus einem verwandten Grund: die Modelle
// aller eingeschalteten Anbieter stehen nebeneinander im Katalog, und wer
// fragt, nennt Anbieter **und** Modell in derselben Bedienhandlung (A8,
// #1447). Eine Einstellung »frage.kivorschlag« entschiede etwas, das längst
// entschieden ist.
//
// Bei **Preisen** antworten alle nebeneinander. Scryfall schätzt, MTGJSON
// sammelt fünf Anbieter, CardTrader nennt die Angebote einzelner Verkäufer —
// und jede dieser Aussagen ist eine **eigene Reihe**
// ([0093](../../docs/entscheidungen/0093-eine-preisreihe-hat-sieben-teile.md)),
// keine Alternative zu den anderen. Zwei Preise nebeneinander sind kein
// Widerspruch, den jemand auflösen müsste; zwei Trefferlisten wären einer.
//
// **Ohne diese Unterscheidung entstünde eine Einstellung, die niemand
// liest** — `frage.preise` mit drei Werten, und keine Stelle im Code, die sie
// fragt. Ein Schalter ohne Wirkung ist schlimmer als keiner (§10), und er
// wäre hier von selbst entstanden: [Fragen] leitet die Wahl daraus ab, dass
// mehr als ein Dienst antworten kann (A4, #1443,
// [0095](../../docs/entscheidungen/0095-was-cardtrader-sagt-und-was-davon-gilt.md)).
func (f Frage) Mehrstimmig() bool {
	return f == FragePreise || f == FrageKIVorschlag
}

// Fragen gibt die Fragen, die eine **Wahl** haben — in Katalogreihenfolge.
//
// **Eine Wahl mit einer Möglichkeit ist keine** (oben). Welche Dienste eine
// Frage beantworten, steht seit Werkstück A1 von Phase 5.3
// (#1440)
// im Vertragskatalog ([Dienstvertraege]) und nicht mehr in einer Liste
// daneben: eine Frage kommt hierher, sobald ein zweiter Dienst einen Vertrag
// dafür hat — und bekommt Schalter, Vorgabe und Prüfung aus derselben Quelle.
func Fragen() []Frage {
	var alle []Frage

	for _, vertrag := range dienstvertraege {
		if vertrag.Frage.Mehrstimmig() {
			continue
		}

		if len(DiensteFuer(vertrag.Frage)) > 1 && !slices.Contains(alle, vertrag.Frage) {
			alle = append(alle, vertrag.Frage)
		}
	}

	return alle
}

// FrageEinstellung gibt den Einstellungsschlüssel einer Frage.
//
// **Ein anderes Präfix als bei den Schaltern** (`dienst.`): sie stehen im
// selben Bereich und beantworten zwei verschiedene Fragen. Wer sie gleich
// benennte, hätte in der Einstellungsreihe zwei Zeilen, die gleich aussehen
// und Verschiedenes tun.
func FrageEinstellung(frage Frage) string { return "frage." + string(frage) }

// DiensteFuer gibt die Dienste, die eine Frage beantworten können — in
// Katalogreihenfolge, und **der erste ist die Vorgabe**.
//
// **Aus den Verträgen und nicht aus einer Tabelle**: ein Dienst steht hier,
// weil es für ihn und diese Frage einen Vertrag gibt — also einen gebauten
// Weg. Ein Dienst, der einen Endpunkt *nennt*, den hier niemand ruft, steht
// nicht zur Wahl (#1440: kein Universalanbieter mit Methoden, die »nicht
// unterstützt« melden).
func DiensteFuer(frage Frage) []Dienst {
	var dienste []Dienst

	for _, vertrag := range dienstvertraege {
		if vertrag.Frage == frage {
			dienste = append(dienste, vertrag.Dienst)
		}
	}

	return dienste
}
