# 0095 — Was CardTrader sagt, und was davon hier gilt

**Stand:** angenommen · **Datum:** 2026-09-11 · **Werkstück:** A4 von Phase 5.3
(#1443)

## Worum es geht

CardTrader ist der erste Dienst hier, der einen **Marktplatz** führt: nicht
eine Schätzung je Karte, sondern die Angebote einzelner Verkäufer, jedes mit
Preis, Menge, Sprache, Ausführung und Erhaltung. Das ist eine andere Art von
Auskunft als bei Scryfall oder MTGJSON, und sie bringt drei Fragen mit, die
eine Entscheidung brauchen, bevor eine Zeile Code entsteht.

**Vier Dinge stehen hier, weil §16 Satz 1 sie verlangt**: die Bedingungen sind
**nachgelesen**, nicht erinnert.

## Was nachgelesen ist

Am **2026-09-11** unter <https://www.cardtrader.com/en/docs/api/full/reference>:

| | |
| --- | --- |
| Grundadresse | `https://api.cardtrader.com/api/v2` |
| Anmeldung | `Authorization: Bearer <Marke>`, die Marke aus den Profileinstellungen |
| `GET /games` | `[{id, name, display_name}]` — Magic ist `id: 1` |
| `GET /expansions` | `[{id, game_id, code, name}]` |
| `GET /blueprints/export?expansion_id=N` | `expansion_id` ist **Pflicht**; ohne oder ungültig: `404` |
| `GET /marketplace/products` | **eines** von `expansion_id` oder `blueprint_id`; dazu `foil` und `language` |

Ein Blueprint:

```json
{ "id": 89, "name": "Beast // Plant", "version": null,
  "game_id": 1, "category_id": 2, "expansion_id": 8, "image_url": "…",
  "editable_properties": [ … ],
  "scryfall_id": "381854e2-7369-473e-a604-4dd7c010fc89",
  "card_market_ids": [362677], "tcg_player_id": null }
```

Ein Angebot — und **die Antwort ist nach Blueprint geordnet**, nicht flach:

```json
{ "10050": [ { "id": 101862104, "blueprint_id": 10050, "quantity": 1,
    "price": {"cents": 2, "currency": "USD"},
    "properties_hash": {"condition": "Moderately Played", "signed": false,
                        "mtg_foil": false, "mtg_language": "en",
                        "altered": false},
    "expansion": {"id": 92, "code": "ptkdf", "name_en": "Tarkir Dragonfury"},
    "user": {…}, "graded": false, "on_vacation": false, "bundle_size": 1 } ] }
```

## Die Grenzen — und der Widerspruch, den das Issue nennt

Drei Sätze stehen in der Referenz:

- „All API endpoints combined have a rate limit of **200 requests per 10
  seconds**."
- am Marktplatz: „This endpoint has a rate limit of **10 requests per second**."
- an den **Jobs**: „This endpoint has a rate limit of **1 request per second**."

**Das Issue nennt „Marketplace widersprüchlich 1/s und 10/s" — und es hat
recht.** Am 2026-09-19 im Wortlaut nachgelesen: der Abschnitt zu
`GET /marketplace/products` trägt **beide** Sätze, „This API is also limited
to 1 call per second" und „This endpoint has a rate limit of 10 requests per
second". Der Widerspruch steht in der Quelle und nicht im Zusammenfassen.

> **Hier stand das Gegenteil**, und der Zweitblick über Phase 5.6 hat es
> angestrichen (Befund 8). Die erste Zurückweisung stützte sich auf eine
> **Zusammenfassung** der Seite statt auf ihren Wortlaut — und eine
> Zusammenfassung ordnet einen Satz dem Abschnitt zu, in den er zu passen
> scheint. Die 1/s stehen **auch** an `GET /jobs/:uuid`; das machte sie hier
> nicht falsch.
>
> **Eine Ablehnung braucht Deckung** (§10) — auch die Ablehnung eines
> Befunds. Nachgelesen wird der Wortlaut, nicht die Auskunft über ihn.

**Gedrosselt wird auf 1/s**, und damit ist die engere der beiden Angaben die,
die gilt:

- Was dieses Programm fragt, sind wenige Abrufe je Bedienhandlung — ein
  Blueprint, ein Marktplatz. Zwischen 1/s und 10/s liegt hier keine spürbare
  Wartezeit.
- Eine falsch geratene Grenze kostet einen Ausschluss, eine zu vorsichtige
  kostet nichts. §16, Regel 4: **gedrosselt wird unter der genannten Zahl**.

Die globale Grenze von 200/10 s bleibt zusätzlich: sie gilt über alle
Endpunkte zusammen, und 1/s liegt weit darunter.

## Die Erhaltung wird **nicht** übersetzt

CardTrader kennt sechs Stufen: `Near Mint`, `Slightly Played`, `Moderately
Played`, `Played`, `Heavily Played`, `Poor`. Dieses Programm kennt sieben —
`mint`, `nearMint`, `excellent`, `good`, `lightPlayed`, `played`, `poor`
(`domain.Zustand`, die Skala von Cardmarket).

**Die beiden Skalen sind nicht ineinander überführbar.** Sechs Stufen auf
sieben abzubilden verlangte Gleichsetzungen, die niemand nachprüfen kann — ist
„Slightly Played" ein `excellent` oder ein `good`? Die Frage hat keine
belegbare Antwort, und eine erfundene wäre genau der Fehler, vor dem §15 warnt.

**Deshalb reist die Erhaltung als Text mit**, so wie die Quelle sie nennt.
`domain.Preisstand.Zustand` ist eine Zeichenkette und Teil der **Reihe**
([0093](0093-eine-preisreihe-hat-sieben-teile.md)): „CardTrader, Near Mint" ist
eine andere Reihe als „CardTrader, Played", und beide sind mit keiner Reihe
von Cardmarket vergleichbar. Das ist keine Lücke, sondern die ehrliche Form
der Auskunft.

**Was das kostet, gehört dazu**: eine Sammlung, deren Exemplare in
`domain.Zustand` stehen, findet zu einem Exemplar in `excellent` keinen
CardTrader-Preis „für genau diesen Zustand". Sie findet die Reihen, die es
gibt, und wer vergleicht, vergleicht sichtbar über Skalen hinweg.

## Die Bild-URL ist flüchtig

Das Issue verlangt, „signierte/expirierende Bild-URLs nicht ungeprüft zu
übernehmen". **Nachgelesen sagt die Referenz dazu nichts** — weder dass
`image_url` signiert ist noch dass sie abläuft.

Eine Auskunft, die es nicht gibt, wird nicht erfunden (§10). Behandelt wird
sie deshalb als **flüchtig**: sie wird nicht in den Bestand geschrieben und
nicht in einen Export gereicht. Sie reist mit einer Antwort und stirbt mit
ihr. Wäre sie doch signiert, hätte ein abgelegtes Exemplar nach Ablauf auf
nichts gezeigt — und niemand hätte gesehen, wann es passiert ist.

## Nur Lesen, und die Liste steht im Code

Die Positivliste war `GET /games`, `GET /expansions`,
`GET /blueprints/export`, `GET /marketplace/products` — **und sonst nichts**.
Kein `POST`, kein `PUT`, kein `PATCH`, kein `DELETE`.

**Seit `T11` (#1476)
sind es sieben**: `GET /info`, `GET /products/export` und `GET /orders` kamen
dazu — das **eigene** Konto, hinter einer eigenen Freigabe, die
ausgeschaltet anfängt. **An der Zusage ändert das nichts**: alle sieben sind
`GET`. Was CardTrader zum Schreiben anbietet, ist bekannt und bleibt
ungebaut; die achtzehn Wege stehen namentlich in
[0122](0122-ein-fremdes-konto-wird-gelesen.md).

**Die Liste ist eine Aufzählung im Code und kein Vorsatz**, und ein
Server-Spion in `internal/flaechen` weist nach, dass über alle vier Flächen
zusammen genau null schreibende Methoden hinausgehen. Ein zweiter in
`internal/cardtrader/spion_test.go` ruft **jeden** der sieben Wege und zählt
die Methoden nach: eine, und sie heißt `GET`.

## Preise sind mehrstimmig — und bekommen deshalb keine Wahl

Der Vertragskatalog leitet bisher eine **Wahl** daraus ab, dass mehr als ein
Dienst eine Frage beantwortet: „für diese Frage nimm jenen Dienst"
([0063](0063-was-ein-dienst-hier-ist.md)). Das ist bei der **Suche** richtig —
wer „Lightning Bolt" tippt, will eine Trefferliste, und zwei Dienste ergäben
zwei, zwischen denen niemand entschieden hat.

**Bei Preisen ist es falsch.** Scryfall schätzt, MTGJSON sammelt fünf Anbieter,
CardTrader nennt Angebote einzelner Verkäufer — und jede dieser Aussagen ist
eine **eigene Reihe** ([0093](0093-eine-preisreihe-hat-sieben-teile.md)), keine
Alternative zu den anderen. Zwei Preise nebeneinander sind kein Widerspruch,
den jemand auflösen müsste; zwei Trefferlisten wären einer.

Ohne diese Unterscheidung wäre mit diesem Werkstück eine Einstellung
`frage.preise` **von selbst entstanden**, mit drei Werten und keiner Stelle im
Code, die sie liest. Ein Schalter ohne Wirkung ist schlimmer als keiner (§10).

`domain.Frage.Mehrstimmig()` sagt es jetzt, und `domain.Fragen()` übergeht
mehrstimmige Fragen.

## Was dabei aufgefallen ist: MTGJSONs Vertrag fehlte

Der Katalog ist die **Auskunft an Nutzende** darüber, was ein Dienst hier tut
und unter welchen Bedingungen — und er stand seit A3
(#1442) zu
kurz: der Bulkbezug holt Preise bei MTGJSON, während der Katalog sagte,
MTGJSON könne hier „suchen und sonst nichts". Eine Auskunft, die zu wenig
sagt, ist so falsch wie eine, die zu viel verspricht.

Nachgetragen mit diesem Werkstück, nachgelesen am 2026-09-11:
MIT-Lizenz, keine genannte Anfragegrenze, und „Builds kick off at 1:00AM EST
and go live at 9:00AM EST" — die Dateien entstehen **einmal täglich**, ein
zweiter Abruf am selben Tag bringt dasselbe.

## Verworfen

**Die Erhaltung auf `domain.Zustand` abbilden.** Sie machte Preise über
Quellen hinweg vergleichbar — aber nur scheinbar, und der Schein wäre nicht
mehr zu entfernen, sobald ein Verlauf darauf steht.

**Angebote als Exemplare der Sammlung führen.** Ein Angebot ist die Aussage
eines fremden Verkäufers, kein Besitz; wer beides in eine Tabelle legte,
könnte „meine Karten" nicht mehr beantworten.

**Blueprint und Verkaufsprodukt gleichsetzen.** Ein Blueprint ist die
Kartenfassung, ein Produkt ist das, was ein bestimmter Verkäufer anbietet —
mit eigener `id`. Das Issue nennt es ausdrücklich, und die Trennung steht in
den Typen: `Blueprint.ID` und `Angebot.ProduktID` sind zwei Felder.
