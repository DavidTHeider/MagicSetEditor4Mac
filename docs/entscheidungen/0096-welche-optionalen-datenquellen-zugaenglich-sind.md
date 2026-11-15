# 0096 — Welche optionalen Datenquellen zugänglich sind

**Stand:** angenommen · **Datum:** 2026-09-11 · **Werkstück:** A5 von Phase 5.3
(#1444)

## Worum es geht

A5 ist kein Bauauftrag, sondern zuerst ein **Entscheidungspunkt**: das Issue
sagt es selbst — „Ein negatives Ergebnis ist keine Implementierung". Gefragt
ist je Anbieter eine Untercheckliste mit Zugang, tarifabhängiger Fähigkeit,
exakten Endpunkten, Budgeteinheit und Status.

**Alles hier ist am 2026-09-11 nachgelesen** (§16, Satz 1). Wo eine Angabe
fehlt, steht das da — und nicht eine plausible.

## Scrydex — Kontostand

| | |
| --- | --- |
| Quelle | <https://scrydex.com/docs/getting-started/rate-limits> |
| Endpunkt | `GET /account/v1/usage` |
| Antwort | `total_credits`, `remaining_credits`, `used_credits`, `overage_credit_rate` |
| Frist | „Usage info is updated every 20-30 minutes." |
| Grenze | „100 requests per second", für alle Tarife gleich |
| Budgeteinheit | **Credits**, nicht Anfragen; Überschreitung kostet Gebühren |
| Status | **bauen** |

**Der Satz über die 20 bis 30 Minuten entscheidet die Budgetfrage.** Das Issue
verlangt: „harte Overage-Aussage nur bei belastbarer Anbietersperre". Eine
Zahl, die eine halbe Stunde alt sein kann, ist keine Sperre. Die lokale
Zählung bleibt die Grenze, die gilt; der Anbieterrest steht **daneben** und
sagt, wie alt er ist.

## Scrydex — Preisverlauf

| | |
| --- | --- |
| Quelle | <https://scrydex.com/docs/magicthegathering/price-history> |
| Endpunkt | `GET /magicthegathering/v1/cards/<id>/price_history` |
| Wahlfrei | `days`, `start_date`, `end_date`, `variant`, `condition`, `company`, `grade`, `is_perfect`, `is_error`, `is_signed`, `page`, `page_size` |
| Antwort | `date` (`YYYY-MM-DD`) und je Zeile `prices[]` mit `variant`, `condition`, `type` (raw/graded), `low`, `market`, `currency`, `is_perfect`, `is_signed`, `is_error` |
| Status | **bauen** |

**`low` und `market` sind zwei Aussagen, keine eine.** Beide sind Forderungen
— das niedrigste Angebot und der Marktpreis —, und wer sie in eine Reihe
legte, mischte zwei Fragen. Sie werden über den Quellentrenner getrennt:
`scrydex:low` und `scrydex:market`, dieselbe Gestalt wie `mtgjson:cardkingdom`
([0093](0093-eine-preisreihe-hat-sieben-teile.md)).

**Die Erhaltung wird wieder nicht übersetzt** — dieselbe Regel wie bei
CardTrader ([0095](0095-was-cardtrader-sagt-und-was-davon-gilt.md)).

## Scrydex Vision — das Gate ist erfüllt, gebaut wird hier nichts

| | |
| --- | --- |
| Quelle | <https://scrydex.com/docs/vision/overview> |
| Magic? | **ja, ausdrücklich**: „Identify cards from Pokémon, Magic: The Gathering, Lorcana, One Piece, Riftbound, and Gundam." |
| Endpunkt | `POST /vision/v1/cards/identify` |
| Kosten | „Vision is a premium API and costs **5 credits per request**." |
| Status | **Gate erfüllt → A13** (#1452) |

Das Issue sagt: „Scrydex Vision liefert A13 nur bei dokumentierter
Magic-Fähigkeit." Sie ist dokumentiert. **Gebaut wird sie dort und nicht
hier** — A5 ist die Datenquelle, A13 die Kartenerkennung.

## MTGGraphQL — zurückgestellt, und der Grund hat sich unterwegs geändert

| | |
| --- | --- |
| Quelle | <https://mtgjson.com/mtggraphql/> |
| Endpunkt | `https://graphql.mtgjson.com/`, `authorization: Bearer <Access Token>` |
| Grenze | 1.000 je IP und Stunde, 500 je Marke und Stunde |
| Zugang | **Beta**, „available to Patreon subscribers during rollout" |
| Preisfelder | `provider`, `date`, `cardType`, `listType`, `price` — **dokumentiert**, im Beispiel der Seite ausgeschrieben |
| **Nicht** dokumentiert | eine **Währung** am Preis; und die Cross-IDs (`scryfallId`, `mcmId`, `tcgplayerProductId`, `cardtraderId`) kommen auf der Seite nicht vor |
| Status | **zurückgestellt** |

**Die erste Einschätzung war zu pessimistisch, und das gehört hierher.** Eine
Zusammenfassung der Seite nannte die Preisfelder nicht; erst das
**ausgeschriebene Beispiel** zeigt sie — `cards { prices { provider date
cardType listType price } }`. Wer eine Quelle über eine Zusammenfassung liest,
liest sie nicht (§10).

**Zurückgestellt wird trotzdem, und aus zwei Gründen:**

1. **Am Preis steht keine Währung.** Das Beispiel nennt `price: 5.5` und
   sonst nichts; die Bulkdatei derselben Quelle trägt `"currency"` je Anbieter,
   die GraphQL-Antwort nicht. Ein Preis ohne Währung ist keine Aussage (§16,
   Pflichtfelder streng) — und ob es ein Feld `currency` gibt, lässt sich ohne
   Marke nicht messen. Es anzunehmen wäre genau der Fehler, den §15 verbietet.
2. **Dieselben Preise gibt es schon, und vollständiger.** A3
   (#1442)
   holt MTGJSONs Preise als Bulkdatei — dieselben fünf Anbieter, **mit**
   Währung, ohne Patreon-Marke. Ein zweiter Weg zu denselben Daten mit
   weniger Information wäre Arbeit, die etwas verschlechtert.

**Was ein Mensch mit Marke in einer Anfrage klären kann**: `currency` in die
Auswahl schreiben. Gibt es das Feld, nennt GraphQL es; gibt es keines, sagt es
das mit Namen. Dann ist es eine Messung — bis dahin bleibt es hier stehen.

**Die Cross-IDs bleiben ebenso zurückgestellt**, und aus dem ersten Grund
allein: sie stehen nirgends. Wer sie erfände, hätte eine Zuordnung, die
niemand nachprüfen kann.

## TCG API (tcgapi.dev) — zurückgestellt

| | |
| --- | --- |
| Quelle | <https://tcgapi.dev/> und <https://tcgapi.dev/docs> |
| Magic? | ja — „Daily-updated market prices for Pokemon, Magic: The Gathering, Yu-Gi-Oh!, and 50+ more games" |
| Grundadresse | `api.tcgapi.dev` |
| Anmeldung | `X-API-Key` |
| Tarife | Free 100/Tag, Hobby 1.000, Starter 2.500, Pro 10.000, Business 50.000 |
| Historie | nur Pro und Business, „weekly data points" ab März 2025 |
| Status | **zurückgestellt** |

**Drei Gründe, und keiner ist Bequemlichkeit:**

1. **Die Wege widersprechen sich.** Das Menü nennt `/api/games/`, `/api/sets/`,
   `/api/cards/`, `/api/prices/`, `/api/search/`, `/api/bulk/`; das einzige
   ausgeschriebene Beispiel lautet `/v1/search?q=…&game=…`. Welcher Präfix
   gilt, sagt die Seite nicht.
2. **Es gibt kein dokumentiertes Preisobjekt** und keine Fehlergestalt für 401
   und 429. Ein Adapter dagegen wäre erfundene Fachlogik (§15).
3. **Der freie Tarif ist „Non-commercial use only".** Dieses Programm steht
   unter GPL-2.0-or-later; wer es weitergibt, darf es kommerziell nutzen. Ein
   eingebauter Weg, der im freien Tarif genau das verbietet, wäre eine Falle
   für die Nutzenden — und die Bedingung steht beim Anbieter, nicht bei uns.

## Cardmarket und TCGplayer — nicht zugänglich

Das Issue verlangt es wörtlich: „aktueller rechtlich und technisch nutzbarer
Zugang nachweisen, sonst **keine Attrappenadapter oder Schlüsseloberflächen**."

**Es gibt hier keinen Zugang** — kein Händlerkonto bei Cardmarket, keine
TCGplayer-Partnerkennung. Damit wird nichts gebaut, auch keine leere
Einstellungsfläche: eine Zugangsmaske für einen Weg, den es nicht gibt, sähe
aus wie eine Fähigkeit.

**Die Kennungen bleiben trotzdem lesbar**: `SystemCardmarket` und
`SystemTCGplayer` stehen seit A2 in `domain`, und CardTrader liefert sie seit
A4 mit. Das ist Identität, nicht Marktzugang.

## Scrydex-Webhooks — abgelehnt

Das Issue macht sie zum eigenen Gate: „Scrydex-Webhooks eigenes Gate **ohne
stillen Webserver**."

**Ein Webhook verlangt eine Adresse, die von außen erreichbar ist.** Diese
Anwendung ist ein Programm auf einem Schreibtischrechner hinter einem Router;
sie müsste einen Server aufmachen und ihn öffentlich erreichbar machen. Das
ist eine Fläche nach außen, die niemand verlangt hat, und sie stünde offen,
während niemand hinsieht.

**Verworfen, nicht vertagt.** Wer Preise will, fragt danach — der Verlauf oben
gibt sie.

## Was daraus als Arbeit folgt

1. `internal/scrydex` bekommt **Kontostand** und **Preisverlauf**, beide an
   allen vier Flächen (§16).
2. Die Zusage „zwei gleichzeitige Anfragen können die Grenze nicht
   überschreiten" bekommt einen **Lauf**, der sie misst.

   **Er hat unterwegs etwas anderes gezeigt, als er sollte.** Die Erwartung
   war: ohne `_txlock=immediate` wird die Grenze gesprengt. Gemessen am
   2026-09-11 hält sie auch dann — was umfällt, ist die **Meldung**: zwei von
   drei Aufträgen scheitern mit „der Bestand ist gerade belegt", wo einer von
   ihnen hätte gebucht werden dürfen. Der Lauf prüft deshalb beides, und die
   zweite Zusicherung ist die empfindliche.
3. **Fünf Gates** stehen hier und im Dienstkatalog: TCG API und MTGGraphQL
   zurückgestellt, Cardmarket und TCGplayer nicht zugänglich, die
   Scrydex-Webhooks abgelehnt. Scrydex Vision ist erfüllt und gehört nach A13.
