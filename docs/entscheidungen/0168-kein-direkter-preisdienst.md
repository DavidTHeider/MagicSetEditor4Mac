# 0168 — Kein direkter Preisdienst: Cardmarket und TCGplayer nehmen keine neuen Zugänge an

**Stand:** angenommen am 2026-10-04 · Recherche aus AI-04
(#1745),
Entscheidung 3 in [`INTEGRATIONEN.md`](../INTEGRATIONEN.md), geführt als
#2050.
Marke `new`: MSE2 kennt keine Preise.

## Die Frage

Euro-Preise kommen heute über **drei** Umwege, nicht zwei, wie
`INTEGRATIONEN.md` bis heute schrieb: MTGJSON mit der Reihe `cardmarket`
([0093](0093-eine-preisreihe-hat-sieben-teile.md)), CardTrader
([0095](0095-was-cardtrader-sagt-und-was-davon-gilt.md)) und Scryfall, dessen
Preise `eur` und `eur_foil` als eigene Reihen in den Bestand gehen
(`internal/scryfall/preise.go`). Gefragt war:
gibt es einen **direkten** Preisdienst, der für dieses Programm zugänglich
ist — und wenn ja, zu welchen Bedingungen?

## Was nachgelesen ist — am 2026-10-04, wörtlich per `curl`

| Kandidat | Zugang | Fundstelle | Was dort steht |
| --- | --- | --- | --- |
| **Cardmarket** | **geschlossen** | <https://help.cardmarket.com/en/cardmarket-api> | Anträge auf Zugang zur API werden derzeit nicht angenommen. Bestehende Nutzer dürfen ihre eigenen Zugangsdaten nur mit einem eigens beauftragten Entwickler teilen, nicht mit fremder Software, die den Bestand ändert; ein Missbrauch kann den Zugang kosten |
| | Dokumentation | <https://api.cardmarket.com/ws/documentation/API_2.0:Main_Page> | `410` mit dem Verweis auf `apiv2.cardmarket.com`; **dort `503`**, „Maintenance is active" — an diesem Tag nicht lesbar |
| | Partner | <https://help.cardmarket.com/en/api-partnerships> | genannt sind TCG PowerTools (Bestandsverwaltung) und Scryfall — kein Weg für Dritte, Preise zu beziehen |
| **TCGplayer** | **geschlossen** | <https://docs.tcgplayer.com/docs/getting-started> | Voraussetzung ist ein API-Entwicklerschlüssel, und neue API-Zugänge werden derzeit nicht vergeben; Bestehende sind an die Nutzungsbedingungen gebunden, samt Pflichtangaben zur Herkunft |

**Nicht umgangen**: `www.cardmarket.com` antwortet einem Abruf ohne
Browser mit `403` hinter einer Botsperre. Eine solche Sperre zu umgehen wäre
genau das Scraping, das AK-03 von #2050 ausschließt — und es ist kein Weg,
auf dem eine Anbindung stehen dürfte.

## Der Vergleich nach den Spalten aus `INTEGRATIONEN.md`

| Kandidat | Nutzen | Machbarkeit | Kosten | Daten nach außen | Wartung | Ausstieg |
| --- | --- | --- | --- | --- | --- | --- |
| Cardmarket direkt | mittel: Euro-Preise ohne Umweg | **nein** — keine neuen Zugänge | — | — | — | — |
| TCGplayer direkt | nicht nachgelesen — ohne Zugang stellt sich die Frage nicht | **nein** — keine neuen Zugänge | — | — | — | — |
| **Wie heute**: MTGJSON, CardTrader, Scryfall | Euro aus drei Quellen, je mit Herkunft (0093) | gebaut | keine | Kartenkennungen, je Dienst wie dokumentiert | drei Klienten, schon da | jeder einzeln abschaltbar |

## Die Entscheidung

**Abgelehnt — kein direkter Preisdienst.** Beide Kandidaten nehmen
heute keine neuen Zugänge an; ein Bauissue entsteht nicht. Die Euro-Preise
bleiben bei den drei Quellen, die gebaut sind, und `INTEGRATIONEN.md` nennt
jetzt alle drei.

## Was die Entscheidung kippt

**Eine wieder geöffnete Anmeldung** bei Cardmarket oder TCGplayer. Dann ist
die Recherche zu wiederholen — Bedingungen, Grenzen, was über den Bestand
hinaus nach außen geht —, und **erst danach** ein Bauissue (§16, Satz 1).
Nachzusehen ist dort, wo hier die Fundstellen stehen.

## Was das kostet

**Keine Preise aus erster Hand.** Was MTGJSON, CardTrader und Scryfall
weiterreichen, ist abgeleitet und hat ihren Stand — wie alt, sagt jede
Preisreihe selbst (0093).
