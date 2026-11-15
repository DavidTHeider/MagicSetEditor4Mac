# 0086 — Sieben Begriffe für eine Karte: die quellneutrale Druckidentität

**Stand:** angenommen
**Datum:** 2026-09-09
**Anlass:** Werkstück A2 von Phase 5.3
(#1441) — Scryfall-Verweise, Sammlung, Decks,
Preise, Spielen und Handeln sollen eine **auflösbare** Identität teilen, ohne
Oracle, Druck, Produkt und Exemplar gleichzusetzen.

## Lage

**Was am 2026-09-09 galt, ist nachgelesen** (Stand `e9ceb821`):

- `printing TEXT` — die Scryfall-UUID eines Drucks — ist die Identität an
  **fünf** Stellen des Bestands: `karten_verweis` (0020), `sammlung` (0018,
  0026, 0035, 0046; je Profil, `UNIQUE` mit Zustand, Sprache, Foil, Wunsch),
  `preisstaende` (0025), `deckplaetze` (0028, `karte_id` **oder** `printing`),
  `turnierdecklisten` (0031, Text neben der Bezeichnung).
- In `domain` heißt sie fünfmal `Printing` (`Kartenverweis`, `Exemplar`,
  `Deckplatz`, `Preisstand`, `Abzugsplatz`) und wird an **einer** Stelle
  geprüft (`deck.go`: eine UUID) und an einer anderen nur getrimmt.
- Vierzehn Dienste in `service` und fünf Ablagen in `storage` tragen sie
  durch; drei REST-Wege nennen `{printing}` in der Adresse, neun
  MCP-Werkzeuge nehmen `printing` als Argument, die Sammlungsausfuhr schreibt
  sie in CSV und JSON, die Listeneinfuhr liest sie.
- Es gibt **keinen** Kennungstyp und **keine** zweite Quelle: was kein
  Scryfall-Druck ist, hat keine Identität — eine eigene Karte im Deck hängt
  an `karte_id`, ein Turnier-Deck an einem Namen.

**Was die Anbieter sagen** (nachgelesen am 2026-09-09):

- CardTrader (`/blueprints/export`): ein Blueprint trägt `scryfall_id`
  (string), `card_market_ids` (**array**) und `tcg_player_id` (string). Ein
  Anbieterprodukt kann also **mehrere** Cardmarket-Kennungen tragen.
- MTGJSON (`identifiers`): `scryfallOracleId` ist „consistent across
  reprints"; `scryfallId` und `mtgjsonV4Id` nennen einen Druck;
  `mtgjsonFoilVersionId` / `mtgjsonNonFoilVersionId`, `mcmId`,
  `tcgplayerProductId` (+ Etched, Alternative Foil), `cardKingdomFoilId` … sind
  **je Ausführung oder je Händler** verschieden; `scryfallIllustrationId`
  nennt das Kunstwerk.

## Entscheidung

### 1 · Sieben Begriffe, und keiner ist der andere

| Begriff | Was er ist | Wo er heute steht |
| --- | --- | --- |
| **Kartenkonzept** | die Regelidentität — „Lightning Bolt" mit seinem Oracle-Text; n Drucke | nirgends (Oracle-ID wird nicht geführt) |
| **Druck** | eine Ausgabe des Konzepts in einem Set mit Nummer; die Sprache, eine Variante (Showcase, Extended) und die angebotenen Ausführungen (Finish) sind **Achsen** des Drucks | `printing` — die Scryfall-UUID |
| **Anbieterprodukt** | was ein Händler verkauft — ein CardTrader-Blueprint, eine Cardmarket-ID; kann einen Druck je Finish oder je Sprache teilen oder mehrere Kennungen tragen | nirgends |
| **Physisches Exemplar** | eine Zeile der Sammlung: Druck × Zustand × Sprache × Foil × Absicht, mit Menge | `sammlung` |
| **Eigenes Kartendesign** | eine Karte in einem Set dieses Programms, ohne Druck | `cards` |
| **Schnappschuss** | was zu einem Zeitpunkt von draußen kam: ein Preisstand, der Name/Setcode/Nummer neben einem Verweis | `preisstaende`, `karten_verweis.name…` |
| **Portable Referenz** | die Form, in der ein Druck einen Bestand verlässt — versioniert, mit System und Wert, nie mit einer lokalen Zeilennummer | nirgends (heute reist die UUID roh) |

### 2 · Eine Tabelle `drucke`, eine Tabelle `druckkennungen`

`drucke(id, setcode, nummer, sprache, name, vorlaeufig)` ist die lokale
Zeile, **nur intern**. `druckkennungen(druck_id, system, semantik, wert)`
nennt, wie andere den Druck nennen:

| Semantik | Bedeutung | Eindeutig? | Beispiele |
| --- | --- | --- | --- |
| `druck` | der Wert nennt genau diesen Druck | **ja**, je System: `UNIQUE(system, wert) WHERE semantik = 'druck'` | `scryfall` id, `mtgjson` uuid |
| `konzept` | der Wert nennt das Konzept, das mehrere Drucke teilen | nein | `scryfall` oracle_id |
| `produkt` | der Wert nennt ein Anbieterprodukt; mehrere je Druck, je Finish oder Händler | nein — aber ein Wert zeigt auf **einen** Druck, sonst Konflikt | `cardmarket` id, `cardtrader` blueprint, `tcgplayer` product |
| `kunstwerk` | die Illustration, über Drucke hinweg | nein | `scryfall` illustration_id |

**Kein `UNIQUE(system, wert)` über alle Semantiken** — das war der Fehler,
den das Issue benennt: viele Drucke teilen eine Oracle-ID, und ein Blueprint
trägt mehrere Cardmarket-IDs. Zeigt ein `druck`-Wert oder ein `produkt`-Wert
auf zwei Drucke, ist das ein **Auflösungskonflikt** — ein Befund, der gezeigt
wird, kein Automerge und **kein** Zusammenführen über den Namen.

### 3 · Die fünf Stellen bekommen `druck_id` — und `printing` bleibt als Alias

Wanderung 0051 füllt `drucke` aus jedem heutigen `printing`-Wert (DISTINCT
über die fünf Tabellen) und `druckkennungen(scryfall, druck, wert)`; die
fünf Tabellen bekommen `druck_id` mit Fremdschlüssel, `printing` bleibt und
ist ab jetzt **„die Scryfall-Kennung, falls es eine gibt"**. Ein Druck ohne
Scryfall-Kennung hat `printing = ''` und `vorlaeufig = 1` — sichtbar, nicht
still. Die Wanderung ist wiederanlauffähig (`INSERT … WHERE NOT EXISTS`) und
läuft hinter der bestehenden Sicherung; keine ältere Wanderung wird
angefasst (§ Migration nach dem Lauf bearbeitet).

### 4 · Die Flächen bleiben, wie sie sind — und bekommen eine Spalte dazu

`printing` als Parameter und Feld bleibt an REST, MCP, CLI und Fenster und
wird als `scryfall/druck` aufgelöst. Dazu kommt `druckID` (Zeichenkette, wie
jede Kennung an der Naht — kein `int64` im JavaScript) und `kennungen`
(System, Semantik, Wert). Die Ausfuhr trägt eine **portable Referenz**
`{fassung: 1, system, semantik, wert, name, setcode, nummer, sprache}`; die
Einfuhr löst über Kennungen auf und nimmt **nie** eine lokale Kennung des
Senders als eigenes Ziel — aus Bestand A nach Bestand B trifft dieselbe
Referenz denselben fachlichen Druck.

### 5 · Vorläufig ist ein Zustand, kein Fehler

Eine eigene Karte braucht keinen Scryfall-Zwang; ein Exemplar aus einer
Liste ohne Kennung bekommt einen **vorläufigen** Druck (Name, Setcode,
Nummer) und die Sichten sagen es. Eine spätere Auflösung ersetzt die
vorläufige Zeile nicht, sondern hängt die Kennung an — die Sammlung zählt
weiter dieselbe Zeile.

## Folgen

- Preise (A3) hängen künftig an `druck_id`, und ihre Quelle darf MTGJSON
  oder Scryfall sein; CardTrader (A4) hängt Produkte an Drucke.
- Die Sammlungs-Finishmigration bleibt beim T-Owner (Abgrenzung im Issue).
- Der Zweitblick prüft: zwei Drucke mit einer Oracle-ID, mehrsprachige und
  zweiseitige Varianten, mehrere Markt-IDs, derselbe Textwert in zwei
  Systemen, eine unbekannte Zusatzkennung, eine Kollision; ein 0.4.0-Bestand
  (Fixture) mit Anzahl, Mengen, fremden IDs, Verweisen, Preisen und
  `.mse-set`-Rundlauf; Abbruch mitten in der Wanderung und Wiederöffnen.

## Verworfen

- **Die Scryfall-UUID bleibt die Identität, andere Systeme werden daran
  gehängt.** Bequem — und falsch für alles, was Scryfall nicht kennt (eigene
  Karten, Drucke ohne Scryfall-Eintrag) und für Produkte, die einen Druck
  je Finish teilen.
- **Namensabgleich als Rückfall.** „Lightning Bolt" gibt es in dreißig Sets;
  ein Automerge über den Namen erzeugt genau die stillen Fehler, die 0063
  bei den Diensten ausschließt.
- **Alte Wanderungen umschreiben.** Ein Bestand, der 0018 schon gelaufen hat,
  merkte nichts davon; die Prüfsumme sperrte ihn aus.

## Nachtrag vom 2026-09-09 — beim Bau der Nähte

- **Der Deckplatz wird in 0051 neu gebaut.** 0028 verlangte genau eines von
  `karte_id` und `printing`, und ein Druck ohne Alias hat nur `druck_id`;
  SQLite ändert keine CHECK-Bedingung. Die Kennungen der Plätze bleiben, und
  auf die Tabelle zeigt niemand.
- **Der Alias bleibt so nachsichtig, wie er war.** Die Ablage legt zu jedem
  `printing` die Druckzeile an, ohne UUID-Zwang — ein Altbestand mit einem
  Alias, der keine UUID ist, wandert und liest sich zurück. Geprüft wird eine
  **ausdrückliche** Referenz in der CSV (`kennungssystem`, `kennung`), und
  eine, die den Alias nur wiederholt, gilt als Alias: sonst käme die eigene
  Ausfuhr strenger zurück, als sie hinausging.
- **Vorläufig heißt an den Nähten: ohne Scryfall-Alias.** Die Flächen kennen
  heute genau diese Kennung; ein Druck, den nur MTGJSON eindeutig nennt, zeigt
  sich dort ebenso als vorläufig, bis A4 die Kennungen an die Flächen bringt.
  [Druck.Vorlaeufig] fragt den Druck selbst.
