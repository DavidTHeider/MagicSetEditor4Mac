# 0071 — Woher ein Deck kommt

**Entschieden am 2026-09-06**, Werkstück 12 von Phase 4.12
(#1351).

## Der Anlass

Die Kommentierung verlangt es wörtlich:

> es [muss] auch möglich sein, die Herkunft eines Decks zu verfolgen, das
> künftig mit anderen getauscht werden kann, ebenso eine Sammlung. Die gehört
> einem zwar dann, kommt aber woanders her.

## Portierungsmarke: **new** — und hier ist gesucht worden

§15 verlangt, dass eine `new`-Marke sagt, **wo** gesucht wurde. Am 2026-09-06
gemessen:

| Quelle | Treffer auf `deck` | Was es ist |
| --- | ---: | --- |
| `/Users/david/mse2-quelle/src` | **1** | ein deutsches Wort („deckt") in einem Kommentar |
| `/Users/david/magic-set-generator/src` | **1** | ein `TODO` über `args.decklists`, nicht umgesetzt |

**MSE2 kennt kein Deck**, und `magic-set-generator` hat eines vorgehabt und
nicht gebaut. Es gibt kein Altverhalten, das zu portieren wäre — das ist die
Lage, die Vermerk [0061](0061-was-ein-deck-hier-ist.md) für das Deck selbst
schon festgestellt hat.

## 1. Eine Herkunft darf **fehlen**, und das Fehlen ist die Auskunft

**Keine Herkunft heißt: hier entstanden.** Eine Zeile für „von mir selbst"
trüge nichts bei und wäre die Zeile, die beim Anlegen einmal vergessen wird —
danach sähe ein eigenes Deck aus wie ein übernommenes.

Der Satz steht schon in `0014_wertherkunft.sql` und gilt hier unverändert:
**eine falsche Nennung ist schlimmer als keine.**

Deshalb zwei **Tabellen** und nicht zwei Spalten: ein leerer Text ließe „hier
entstanden" und „übernommen, aber niemand hat die Quelle notiert" nicht mehr
unterscheiden; eine Zeile, die es gibt oder nicht, sagt beides ohne Zusatzregel.

## 2. Zwei Arten, und beide haben einen Aufrufer

| Art | Wer sie schreibt |
| --- | --- |
| `profil` | die **Deckübernahme** aus einem fremden Profil — neu in diesem Werkstück |
| `einfuhr` | `SammlungCSVEinlesen`, wenn der Aufrufer eine Bezeichnung mitgibt |

**Die zweite Zeile ist der Grund, warum die erste gebaut wurde.** Für die
Sammlung gab es längst einen Weg von außen; für ein Deck gab es **keinen**.
Eine Herkunftstabelle ohne Schreiber wäre eine Spalte, die niemand füllt — die
erste Dimension aus §10, und sie steht dort zuerst, weil sie schon einmal
gefehlt hat.

## 3. Warum die Übernahme zwischen **Profilen** läuft

**Weil das der Tausch ist, den diese Anwendung heute wirklich kann.** Der
Kommentar sagt „künftig mit anderen getauscht"; der Tausch zwischen zwei
Rechnern ist ausdrücklich eine spätere Iteration, und die Roadmap führt ihn
außerhalb dieser Phase.

**Zwischen zwei Profilen desselben Bestands ist der Tausch aber schon
möglich** — seit Werkstück 11 sagt eine Freigabe, wer in welches fremde Profil
darf. Und er ist **belegbar**: dieser Bestand hat die Übernahme selbst
gesehen. Das ist der Unterschied zu einer Datei, die behauptet, woher sie
kommt.

**Was der spätere Tausch erbt**: dieselbe Tabelle, dieselben zwei Arten. Ein
Deck von einem anderen Rechner käme als `einfuhr` mit der Bezeichnung, die der
Empfänger angibt — genau wie ein CSV heute.

## 4. Die Übernahme fragt **dieselbe** Erlaubnis wie der Wechsel

**Und keine strengere.** Wer aus einem Profil nehmen darf, ist genau, wer
hineingehen darf; sie ruft deshalb `Nutzerverwaltung.Zutritt` und prüft nicht
eigens.

**Warum das keine Nachlässigkeit ist**: eine Übernahme, die mehr verlangte als
der Wechsel, wäre mit einem Wechsel zu umgehen — hineingehen, ansehen,
abtippen. Eine Sperre, die man in drei Schritten umgeht, ist keine Sperre,
sondern eine Zumutung für den, der sie beachtet.

**Der Nachweis ist leer**, weil dieser Weg niemanden nach einem Code fragt.
Für ein Profil ohne zweiten Faktor heißt das „geht", für eines mit Faktor „nur
mit Freigabe" — dieselbe Zusage wie in
[0070](0070-rechte-rollen-und-sichtbarkeit.md), Abschnitt 5.

## 5. Was **nicht** mitkommt — und dass es gesagt wird

| | Warum |
| --- | --- |
| ein Platz auf eine **eigene Karte** | er zeigt auf eine Karte im Quellprofil, und die gibt es hier nicht |
| das **Format** | es ist eine Zeile im Quellprofil |

Beides mitzunehmen hieße, einen Verweis über die Profilgrenze zu legen — genau
das, was Werkstück 2 verhindert.

**Genannt wird beides.** Ein Deck, von dem vierzig Karten kommen und zehn
nicht, wird mit vierzig übernommen und zehn Befunden. Das ist dieselbe Haltung
wie bei einem halb lesbaren MSE-Paket (§8): 300 Karten und zwei benannte
Befunde, nicht eine Absage.

**Ein Deck ohne Format ist ungeprüft, nicht falsch** ([0061](0061-was-ein-deck-hier-ist.md)) —
deshalb ist die Formatabsage ein Befund und keine Ablehnung.

## 6. Keine Kette

**Gibt A an B und B an C, dann sagt die Herkunft bei C „von B"** — nicht „von
B, ursprünglich von A".

**Warum nicht mehr**: eine Kette müsste mit dem Deck **mitreisen**, und was
mitreist, hat der Gebende geschrieben. Was dieser Bestand aufschreibt, ist,
was er **selbst gesehen** hat. Eine übernommene Kette wäre eine Behauptung,
die wie eine Beobachtung aussieht.

**Aus demselben Grund wird der Zeitpunkt gesetzt und nicht mitgegeben.**

## 7. Der Name reist ab, die Kennung zeigt hin

`quelle` ist **abgeschrieben**: `"<Nutzer> · <Profil>"`, zum Zeitpunkt der
Übernahme. Wer sich später umbenennt, ändert damit nicht, woher etwas kam.

**Der Nutzer steht davor, weil „Mein Profil" der Vorgabename jedes ersten
Profils ist** — von einem Fremden gelesen sagt er nichts. Dieselbe Gestalt wie
bei `domain.Exemplar.Bezeichnung`, und sie wird in Go zusammengesetzt, nicht an
drei Flächen (§16).

`quell_profil` steht daneben und trägt **keinen Fremdschlüssel**:

* `CASCADE` nähme die Auskunft weg, während niemand hinsieht.
* `RESTRICT` verhinderte das Löschen eines Profils, weil jemand einmal ein Deck
  daraus übernommen hat.

Dieselbe Überlegung wie bei den Profilvorgaben
([0069](0069-eine-vorgabe-die-auf-eine-zeile-zeigt.md)) — und die **andere**
Richtung als bei den Freigaben ([0070](0070-rechte-rollen-und-sichtbarkeit.md)),
wo `CASCADE` richtig ist. Der Unterschied: eine Freigabe ohne Profil ist Müll,
eine Herkunft ohne Profil ist weiter eine Auskunft.

**Auf das eigene Ding aber schon**: `deck_id` und `exemplar_id` tragen
`CASCADE`. Eine Herkunft ohne das, wovon sie die Herkunft ist, würde wieder
gültig, sobald jemand dieselbe Kennung bekommt.

## 8. Die Herkunft reist **mit**, sie wird nicht nachgefragt

**Ein `LEFT JOIN` und keine eigene Bindung.** Ein zweiter Ruf je Datensatz wäre
eine Stelle, an die jede Fläche denken müsste — und wer sie vergäße, zeigte ein
übernommenes Deck als eigenes.

Sie steht deshalb in `DeckDTO` (beim **einzelnen** Deck; die Übersicht
beantwortet „welche Decks gibt es") und in `ExemplarDTO` (bei **jeder** Zeile,
weil die Sammlung als Liste gelesen wird).

## 9. Eine erhöhte Sammlungszeile bekommt **keine** Herkunft

**Weil sie schon vorher da war.** Wer zehn Stück hat und fünf einliest, hat
danach fünfzehn — und „diese Zeile kam aus sammlung.csv" wäre für zehn davon
falsch. Notiert wird nur an den Zeilen, die **neu entstehen**.

**Was das kostet, gehört gesagt**: bei einer gemischten Zeile steht danach gar
nichts. Das ist Absicht — Abschnitt 1.

## Was verworfen wurde

**Eine Bindung `Herkunft(art, kennung)`.** Sie wäre ein zweiter Ruf je
Datensatz und damit die Stelle, die man vergisst — Abschnitt 8.

**Eine Herkunft je Deckplatz.** „Diese vier Wald kamen von Ada, die anderen
nicht" ist eine Auskunft, die niemand braucht, und sie machte jeden Platz zu
einem Datensatz mit Vorgeschichte.

**Ein Austauschformat für Decks.** Es ist der Weg für den Tausch zwischen zwei
Rechnern — und der ist eine spätere Iteration. Es jetzt zu bauen hieße, ein
Format festzulegen, bevor jemand weiß, was hindurch soll (0033: was über
Fassungen hinweg zugesagt wird, wird zugesagt, wenn es steht).

**Eine Übernahme der Sammlung aus einem fremden Profil.** Für die Sammlung gibt
es den Weg von außen schon — `SammlungCSVEinlesen` —, und ein zweiter wäre
Umfang, den niemand verlangt hat. Er benutzte dieselbe Tabelle und dieselbe
Art; wer ihn will, baut ihn in einem Nachmittag.
