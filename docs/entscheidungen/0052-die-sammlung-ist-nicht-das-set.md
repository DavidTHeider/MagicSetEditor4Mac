# 0052 — Die Sammlung ist nicht das Set

**Stand:** entschieden · **Datum:** 2026-08-27 ·
**Betrifft:** #1165 ·
**Ergänzt** [0002](0002-sqlite-als-arbeitsformat.md),
[0040](0040-der-lebenszyklus-einer-karte-gehoert-dazu.md)

## Die Frage

Phase 3.4 soll eine **Sammlung** und **Preise** bauen. Bevor die erste
Migration entsteht, ist zu entscheiden, **wo diese Daten leben** — und ob eine
Sammlung überhaupt an einem Set hängt.

Die Frage ist nicht akademisch: wer Sammlungseigenschaften in die Kartenfelder
schreibt, bekommt sie beim Ausfahren, beim Sichern und beim Weitergeben nicht
mehr heraus.

## Erst gesucht, dann entschieden

**Es gibt dafür keine Vorlage in MSE2** — und §15 verlangt, dass gesagt wird,
**wo** gesucht wurde. Gesucht wurde am 2026-08-27 in der Fassung `f9d9356`:

| Wo | Womit | Ergebnis |
| --- | --- | --- |
| `src/` insgesamt | `collection`, `inventory`, `owned`, `price`, `wishlist` | 149 Treffer, **alle** technisch: `collection` ist ein C++-Container in `to_value.hpp`, `owned` steht in Kommentaren über Zeigerbesitz |
| `src/data/` — 21 Kopfdateien, das ganze Datenmodell | `quantity`, `amount`, `condition`, `foil` | nichts über Besitz. `match_condition` gehört zu Schlüsselwörtern |
| `src/data/card.hpp`, die Klasse `Card` | gelesen | Feldwerte, Notizen, zwei Zeitstempel, ein abweichender Stil, Stildaten, Zusatzdaten, Schlüsselwortstatistik — **keine Menge, kein Zustand, kein Preis** |
| `src/gui/set/window.cpp:148–154`, Fassung `a7336d6` vom 2020-06-14 — die **sieben** Reiter | die `addPanel`-Rufe gelesen | `CardsPanel`, `StylePanel`, `SetInfoPanel`, `KeywordsPanel`, `StatsPanel`, `RandomPackPanel`, `ConsolePanel`. **Kein Sammlungsreiter** |
| [`fenhl/magic-set-generator`](https://github.com/fenhl/magic-set-generator), Fassung `1b8fb7f` | `price`, `collection`, `owned` | nur Rust-Idiom (`to_owned`, `BTreeSet`) |

**Hier stand „die acht Reiter", und „Fenster" war einer davon.** Das ist
keiner: `window.cpp` **hält** die Reiter, es ist keiner. Angestrichen vom
Zweitblick über Phase 3.4 — und der Befund trifft die Zeile härter, als die
Zahl aussieht. Diese Tabelle ist eine **Messung**; ihre Aussage trägt so weit
wie das, was wirklich nachgelesen wurde (§10). Eine aus dem Verzeichnisnamen
erschlossene Aufzählung, ausgegeben als gelesene, ist genau der Fehler, gegen
den §15 geschrieben ist. Gelesen sind jetzt die Zeilen, und sie stehen da.

**Am Ergebnis ändert es nichts**: unter den sieben ist keiner für den Besitz,
und das war die Frage.

**Damit ist die Sammlung `new`** im Sinne von §15 — und das ist hier
aufgeschrieben, damit die Marke nicht eine unterlassene Suche verdeckt.

**MSE2 kennt nicht einmal `foil` in seinem Kern**: es steht weder in `src/data/`
noch in `src/render/`. Was ein Datenpaket „foil" nennt, ist ein Stilfeld über
das **Aussehen** einer Karte, keine Aussage über ein Exemplar.

## Die Entscheidung

**Drei Dinge, die einander nicht berühren.**

| | Was es ist | Wo es lebt |
| --- | --- | --- |
| **Die Karte** | ein Entwurf: Felder, Bild, Stil | `cards`, `card_values` — und in der `.mse-set` |
| **Das Exemplar** | *ich besitze drei davon, near mint, englisch, nonfoil* | eigene Tabellen, **nie** in der `.mse-set` |
| **Der Preis** | *am 27.08. sagte Scryfall 4,17 €* | eigene Tabelle, Momentaufnahmen mit Zeitstempel |

### Ein Exemplar hängt an einem Druck, nicht an einer Karte

**Und das ist der Kern.** Eine Sammlung führt echte Karten — ein bestimmter
Druck aus einem bestimmten Set, in einer Sprache, foil oder nicht. Eine Karte
in dieser Anwendung ist ein **Entwurf**; sie existiert nicht als Papier.

Die Bezugsgröße ist deshalb die **Scryfall-Kennung eines Drucks**
(`id`, eine UUID je Druck) und nicht die Kennung einer Karte im Bestand.

**Eine Verbindung zu einer entworfenen Karte gibt es trotzdem** — als
Verweis, nicht als Identität: wer eine eigene Karte auf einen echten Druck
bezieht (Phase 3.6, „Kartenreferenzen"), sieht danach auch, ob er ihn besitzt.
Der Verweis darf fehlen, und er darf auf beiden Seiten verschwinden, ohne dass
die andere Seite etwas verliert.

**Ohne Set geht es also auch.** Eine Sammlung ist benutzbar, ohne dass ein
einziges Set offen ist — sie ist kein Reiter **im** Set, sondern ein eigener
Ort.

### Ein Preis ist eine Beobachtung

Kein Feld an einem Exemplar, sondern eine Zeile mit **Quelle, Währung, Wert,
Zeitpunkt**. Drei Gründe:

1. **Ein Preis altert.** Ein Feld verrät nicht, wie alt es ist; eine
   Momentaufnahme trägt ihren Zeitpunkt.
2. **Ein Preis hat eine Herkunft.** „4,17 €" ohne „laut wem" ist keine Auskunft
   (§16, Satz 6).
3. **Aus Momentaufnahmen wird ein Verlauf**, aus einem Feld nicht.

### Die Bedingungen der Quelle, nachgelesen

**Am 2026-08-27 unter [scryfall.com/docs/api](https://scryfall.com/docs/api)
und im Fußbereich derselben Seite** (§16, Satz 1):

| Was | Wortlaut, verkürzt |
| --- | --- |
| Was `prices` ist | „An object containing **daily** price information for this card, including `usd`, `usd_foil`, `usd_etched`, `eur`, `eur_foil`, `eur_etched`, and `tix` prices, **as strings**" |
| Wie verlässlich | „Card prices … represent **daily estimates and/or market values provided by our affiliates**. **Absolutely no guarantee** is made for any price information. See stores for final prices and details" |
| Was nicht erlaubt ist | „You may not simply **repackage, republish, or proxy** Scryfall data. Your software must create additional value for end-users" |
| Und ebenso wenig | „You may not **paywall** access to Scryfall data" |

**Drei Folgen daraus, und alle drei sind bindend:**

- **Als Zeichenkette gelesen, nicht als Gleitkommazahl.** Scryfall schickt
  `"4.17"`; wer daraus ein `float64` macht, hat eine Zahl, die sich nicht mehr
  genau zurückschreiben lässt. Gerechnet wird in **Hundertsteln als
  Ganzzahl**.
- **Der Vorbehalt reist mit.** Wo ein Preis steht, steht auch, woher er kommt
  und wann er erhoben wurde. Eine Portfoliosumme ohne diesen Satz wäre eine
  Behauptung, die die Quelle ausdrücklich nicht macht.
- **Keine Weitergabe als Datensatz.** Preise gehen nicht in eine Ausfuhr, nicht
  in eine `.mse-set`, nicht in MTG JSON. Was sie dürfen: auf **diesem** Rechner
  eine Übersicht tragen.

## Verworfen

| Verworfen | Warum |
| --- | --- |
| **Sammlungsangaben als Kartenfelder** (`quantity`, `condition` als `card field:` im Spielpaket) | Sie landeten in der `.mse-set` und damit bei jedem, dem man ein Set schickt. Und sie wären **je Spielpaket** zu definieren — eine Sammlung, die es nur gibt, wenn das Vorlagenpaket daran gedacht hat |
| **Preis als Feld am Exemplar** | Ein Feld ohne Zeitpunkt ist entweder immer frisch oder immer verdächtig. Der Verlauf wäre nicht nachzubauen, und ein Ausfall der Quelle sähe aus wie „Preis 0" |
| **Eine eigene Datenbankdatei für die Sammlung** | Zwei Dateien, zwei Sperren, zwei Sicherungswege — und die Sicherung aus Phase 3.1 müsste beide kennen. Der Bestand ist schon der Ort für alles, was **diesem Rechner** gehört ([0002](0002-sqlite-als-arbeitsformat.md)); die Trennung gehört ins **Schema**, nicht in die Dateiablage |
| **Die Sammlung an ein Set binden** | Es gibt keine fachliche Beziehung. Wer 300 Karten besitzt und drei Sets entwirft, müsste sich für eines entscheiden — oder die Sammlung dreimal führen |
| **Preise beim Öffnen automatisch holen** | Ein Programm, das beim Start Netzverkehr für 300 Karten auslöst, überrascht seine Nutzenden — dasselbe Argument wie beim Paketbezug (#56). Geholt wird auf **Anweisung**, und der letzte Stand bleibt bis dahin stehen (§16, Satz 5) |
| **Eine zweite Preisquelle neben Scryfall** | Nicht heute. Die Tabelle trägt eine **Quelle** je Zeile, damit eine zweite ohne Migration danebenpasst; gebaut wird sie erst, wenn jemand sie braucht — und ihre Bedingungen sind dann zu lesen, nicht zu vermuten |

## Folgen

- **Ein neuer Schemaschritt** mit eigenen Tabellen. Die Kreativtabellen werden
  nicht angefasst — kein Feld, keine Spalte, kein `NOT NULL` weniger.
- **Die Ausfuhr bleibt, wie sie ist.** Ein Set ein- und auszufahren muss
  Sammlungs- und Preisdaten **unberührt** lassen; das ist ein Abnahmekriterium
  von #1165 und wird als Lauf festgehalten.
- **Die Sicherung nimmt sie mit** (Phase 3.1) — sie sichert den Bestand, und
  darin liegen sie.
- **Optional heißt wirklich optional.** Wer die Sammlung nicht benutzt, merkt
  nichts von ihr: keine leere Spalte an einer Karte, kein Reiter, der auf
  Daten wartet.
- **§16 gilt auch hier**: was das Fenster kann, können Kommandozeile, MCP und
  REST — oder es steht begründet, warum nicht.
