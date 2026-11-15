# 0093 — Eine Preisreihe hat sieben Teile

**Stand:** angenommen · **Datum:** 2026-09-11 · **Werkstück:** A3 von Phase 5.3
(#1442)

## Worum es geht

Bis hierher war **Scryfall** die einzige Preisquelle, und eine Preisreihe war
„Quelle, Währung, Ausführung". Das reichte, weil Scryfall je Karte **einen**
Preis je Währung nennt und sonst nichts. Der Code wusste das: `service`
schrieb `domain.PreisquelleScryfall` fest hin, das Fenster ergänzte „Papier,
Verkauf" beim Zeichnen, und der Dienst erriet dieselben drei aus dem Weg.

Mit **MTGJSON** daneben ist jede dieser drei Stellen ein Fehler, den niemand
sieht.

## Was MTGJSON wirklich führt

Nachgelesen am 2026-09-11 unter
<https://mtgjson.com/data-models/price/price-formats/>,
<https://mtgjson.com/data-models/price/price-list/> und
<https://mtgjson.com/data-models/price/price-points/>:

```
"<uuid>": {
  "paper": { "cardkingdom": { "currency": "USD",
                              "retail":  { "normal": {"2026-09-10": 4.17} },
                              "buylist": { "normal": {"2026-09-10": 2.10} } },
             "cardmarket": …, "cardsphere": …, "tcgplayer": … },
  "mtgo":  { "cardhoarder": … }
}
```

**Dieselbe Karte, vier Anbieter, zwei Märkte, zwei Richtungen** — und jeder
Preis unter einem **Datum**, das die Quelle nennt.

## Die Entscheidung

Eine Preisreihe besteht aus **sieben** Teilen:

| Teil | Werte | Woher |
| --- | --- | --- |
| **Quelle** | `scryfall`, `mtgjson`, ein Händler | wer geantwortet hat |
| **Markt** | `paper`, `mtgo` | MTGJSONs `PriceFormats` |
| **Preisart** | `retail`, `buylist` | MTGJSONs `PriceList` |
| **Währung** | `usd`, `eur`, `tix`, … | die Quelle |
| **Ausführung** | `nonfoil`, `foil`, `etched` | Scryfall/MTGJSON |
| **Zustand** | leer heißt unbekannt | heute **niemand** |
| **Sprache** | leer heißt unbekannt | heute **niemand** |

**Die Werte sind das Vokabular des Anbieters** (§9), die Typnamen deutsch. Eine
Übersetzung in der Tafel machte jede Zeile, die von dort kommt, unlesbar für
den, der beide Seiten vergleicht.

### Warum Zustand und Sprache schon da sind, obwohl sie leer bleiben

Das ist die Stelle, an der dieser Vermerk gegen §10 verstoßen könnte: ein Feld
ohne Verbraucher ist keine Anwendungsfunktion.

Sie stehen trotzdem im Schlüssel, und der Grund ist **A4** (#1443): CardTrader
liefert Angebote, die sich **genau darin** unterscheiden — dieselbe Karte, drei
Zustände, vier Sprachen, vier Preise. Ohne die beiden Spalten liefen sie in
**einer** Reihe zusammen, und „der neueste Preis" wäre der neueste von zwölf.
Eine zweite Wanderung über dieselbe Tafel wäre der teurere Weg.

**Was das kostet, gehört gesagt**: bis A4 gebaut ist, sind zwei Spalten leer.

## Drei Zustände statt eines `bool`

`Weggefallen bool` beantwortete zwei Fragen zugleich. Seither:

| Lage | Heißt |
| --- | --- |
| `genannt` | die Quelle nennt einen Wert |
| `weggefallen` | die Quelle hat **geantwortet** und nennt hier nichts |
| `ungeprueft` | die Antwort deckte diese Reihe **nicht ab** |

Der Unterschied zwischen den letzten beiden ist die Zusage aus §16, Satz 5:
bei `weggefallen` hat jemand nachgesehen und nichts gefunden, bei `ungeprueft`
hat niemand nachgesehen — und dann bleibt der letzte bekannte Wert gültig und
sagt, wie alt er ist.

## Eine Quelle widerruft nur ihre eigenen Reihen

**Der Fehler, den das verhindert**, ist der teuerste dieses Werkstücks: bis
hierher hieß es „was die Antwort nicht nennt, ist weggefallen", und die Antwort
kam von Scryfall. Mit MTGJSON in derselben Tafel hätte ein Scryfall-Abruf
dessen Reihen widerrufen — mit einem Wert von null, der nichts bedeutet, und
ohne eine Meldung.

**Welche Quelle geantwortet hat, steht nicht in der Antwort**, wenn sie leer
ist. Deshalb hat `service.Preisquellenklient` eine Methode `Name()` bekommen:
„die Quelle nennt hier nichts mehr" ist erst dann eine Aussage, wenn feststeht,
**welche** Quelle das war.

## Zwei Zeitpunkte, nicht einer

* **Erhoben** — wann *wir* gefragt haben. Weiß der Dienst, er hält die Uhr.
* **Beobachtet** — welchen Tag die *Quelle* nennt. Weiß nur sie.

Scryfall nennt **keinen** („daily estimates … absolutely no guarantee"), und
dort bleibt das Feld leer. Einen aus dem Abrufzeitpunkt abzuleiten wäre eine
Genauigkeit, die die Quelle nicht hergibt (§16, Satz 6).

## Was das an den Flächen ändert

**Brechend, und zwar mit Absicht** (§16). Die Reihe wird genannt, nicht
erraten:

| Fläche | Vorher | Jetzt |
| --- | --- | --- |
| Fenster | `Preisverlauf(printing, waehrung, finish)` | `Preisverlauf(reihe)` — die Seite nimmt sie aus dem Stand, den jemand angeklickt hat |
| CLI | `mse preisverlauf <druck> <waehrung> <finish>` | `mse preisverlauf <druck> <quelle> <markt> <art> <waehrung> <finish> [zustand] [sprache]` |
| REST | `GET /drucke/{printing}/preise/{waehrung}/{finish}` | `GET /drucke/{printing}/preisverlauf?quelle=…&markt=…&art=…&waehrung=…&finish=…` |
| MCP | `{printing, waehrung, finish}` | dieselben sieben Felder |

**Fünf davon sind Pflicht**, Zustand und Sprache nicht: leer ist dort eine
Angabe („unbekannt"), und heute füllt sie keine Quelle.

**Als Abfrage und nicht im Weg** (REST): sieben Teile in einem Pfad sind keine
Adresse mehr, und zwei davon sind meist leer — ein Pfad mit zwei leeren
Segmenten ist keiner.

## Das Portfolio rechnet enger

Summiert wird seither **nur** `paper` + `retail` + `genannt`. Ein
Ankaufspreis ist, was ein Händler **zahlt**; er liegt deutlich darunter, und
eine Sammlung, die beides mischte, hätte zwei Werte in einer Summe. Dass es
niemandem aufgefallen wäre, liegt daran, dass beide Zahlen aussehen wie Preise.

## Was verworfen wurde

**Ein verträglicher Weg mit Vorgaben.** `GET …/preise/{waehrung}/{finish}`
hätte weiter funktioniert, wenn die Naht „Scryfall, Papier, Verkauf" ergänzte.
Genau das verbietet das Issue: „Quellenwahl explizit, keine automatische
Preisstrategie". Eine Vorgabe ist eine Strategie, die aussieht wie keine.

**Ein `float64` für den Betrag.** MTGJSON schreibt Zahlen, Scryfall
Zeichenketten — beide werden über denselben Textweg gelesen
([domain.PreisAusText]) und als **Hundertstel** geführt. Ein `float64`
dazwischen wäre Rundungsdrift in einer Geldrechnung.

## Reichweite

**Geprüft ist die Buchführung, nicht ein echter zweiter Anbieter.** MTGJSONs
Bulkanschluss ist das nächste Werkstück; was hier steht, ist die Gestalt, in
der seine Antworten Platz haben — gemessen an seiner **Dokumentation**, nicht
an einer Antwort seines Servers.
