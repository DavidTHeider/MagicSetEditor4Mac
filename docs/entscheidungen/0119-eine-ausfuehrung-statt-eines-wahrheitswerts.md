# 0119 — Eine Ausführung statt eines Wahrheitswerts

**Stand:** angenommen
**Datum:** 2026-09-18
**Werkstück:** `T2` von Phase 5.6
(#1467)
**Ergänzt:** [0033](0033-was-ueber-fassungen-hinweg-zugesagt-wird.md) — die
Regeln 5 bis 8 über die Flächen nach außen

## Lage

`domain.Exemplar` trug `Foil bool`. Ein Wahrheitswert kennt zwei Zustände, und
es gibt **drei**: `nonfoil`, `foil` und `etched`.

**Das ist nicht erschlossen, sondern lag schon da.** `domain.Finish` führt die
drei seit Phase 5.3 ([0093](0093-eine-preisreihe-hat-sieben-teile.md)) — die
**Preisdomäne** unterscheidet sie, der **Bestand** nicht. Was das kostet, ist
keine Vermutung: eine geätzte Karte trug `foil: true` und bekam damit den
Foilpreis. Bei *Abundant Harvest*, das bei Scryfall beide Preise führt, sind
das zwei verschiedene Beträge.

**Und der Handel braucht die Unterscheidung.** `domain.Handelsposition` nennt
seit `T1` ein `Finish` ([0117](0117-was-eine-handelsfassung-bindet.md)); eine
Zusage über ein Exemplar, das seine Ausführung nicht sagen kann, ist eine
Zusage über etwas anderes als das Stück Papier, das den Besitzer wechselt.

## Entscheidung

**Vier Sätze.**

### 1 — Die Ausführung ist eine Angabe, und leer heißt `nonfoil`

`domain.Exemplar.Finish` ist ein `domain.Finish`. **Der Nullwert wird
beantwortet und nicht durchgereicht**: `Exemplar.Ausfuehrung()` gibt für die
leere Angabe `nonfoil` — genau das, was `Foil bool` mit `false` meinte. Wer
daraus einen Befund machte, erklärte jeden Bestand von gestern für ungültig.

**`etched` wird nie geraten.** Aus `true` geht nicht hervor, welche der beiden
Veredelungen gemeint war; die Migration bildet deshalb genau auf `nonfoil` und
`foil` ab. Wer `etched` will, sagt es.

### 2 — Nach außen kommt `finish` hinzu, `foil` bleibt

Nach [0033](0033-was-ueber-fassungen-hinweg-zugesagt-wird.md) sind MCP, die
Kommandozeile und REST **außen**, und eine brechende Änderung dort braucht eine
Hauptfassung (Regel 7). 0.5.0 ist keine. Also ist die Änderung **erweiternd**
(Regel 6):

| | Lesend | Schreibend |
| --- | --- | --- |
| `finish` | die Ausführung als Wort | entscheidet, wenn es dasteht |
| `foil` | `true` genau für `foil` | entscheidet, solange `finish` fehlt |

**Die Regel steht an einer Stelle** — `app.AusfuehrungAus`. Drei Flächen, die
den Vorrang je selbst rechnen, rechnen ihn irgendwann verschieden (§16).

**`foil` ist abgekündigt** (Regel 8). Bei MCP steht die Abkündigung in der
`description` des Feldes, wie es das Protokoll hergibt; sonst in
`CHANGELOG.md`. Wegfallen kann es frühestens mit 1.0.0.

### 3 — Der Preis eines Exemplars: veredelt zu veredelt

`foil` und `etched` sind **zwei Veredelungen desselben Drucks**, und nicht
jeder Druck führt beide Preise. Ein veredeltes Exemplar nimmt deshalb einen
veredelten Preis — **seinen eigenen zuerst**, den anderen als Rückfall. Ein
gewöhnliches Exemplar nimmt **nur** einen gewöhnlichen.

**Beide Hälften stehen unter einem Lauf**, und beide haben ihre Gegenprobe
(§10): ohne die erste stünde eine Karte ohne Preis da, obwohl einer vorliegt;
ohne die zweite bekäme eine gewöhnliche Karte den Foilpreis.

### 4 — Das CSV trägt beide Spalten

Eine Ausfuhr ist eine **Austauschdatei**. Sie geht in ein fremdes Werkzeug und
in eine ältere Fassung dieses Programms zurück; beide lesen die Spalte, die sie
kennen. Geschrieben werden deshalb `finish` **und** `foil`, gelesen wird
`finish` mit `foil` als Rückfall.

## Der Preis, und er wird ausgesprochen

**Ein Gegenüber, das nur `foil` kennt und eine geätzte Zeile zurückschreibt,
macht sie zu `nonfoil`.** Das gilt für MCP, REST, die Kommandozeile und für
jedes Werkzeug, das ein CSV von hier liest und wieder schreibt.

**Es ist nicht zu verhindern, sondern nur zu verteilen**: der alte
Wahrheitswert hat zwei Zustände, und drei passen nicht hinein. Die Wahl ist
allein, ob eine geätzte Karte für ein altes Gegenüber als `foil` oder als
`nonfoil` erscheint — und beide Male verliert ein Rückschreiben die Angabe.

**Gewählt ist `false`**, weil es der Fachlage entspricht: `etched` ist eine
eigene Ausführung und nicht der Gleichstand eines Foils. Dieselbe Entscheidung
trägt die Preiswahl aus Satz 3.

## Verworfen

**Die gespeicherte Ausführung gegen ein leeres Feld verteidigen** — also
`etched` stehen lassen, wenn ein Schreibender nur `foil: false` schickt. Sie
wäre die verlustfreie Fassung und **hat einen Preis, der höher liegt**:
`service.Sammlungspflege.Aendern` müsste vor jedem Schreiben lesen, oder die
Regel stünde in einem `CASE` in `storage` — eine Fachregel in der Schicht, die
keine haben darf (§2). Und sie nähme einem alten Gegenüber die Möglichkeit,
ein Foil je wieder gewöhnlich zu machen, weil es beide Fälle gleich schickt.

**Nach außen ebenfalls ersetzen** und die Änderung als brechend führen. Das
verbietet 0033, Regel 7, und der Grund dort gilt hier unverändert: ein
Gegenüber braucht die Zusage, nicht die Ausrede, dass man sie noch nicht geben
musste.

**Ein zweites Feld im Bestand** — `foil` behalten und `finish` danebenlegen.
Das Issue nennt es ausdrücklich: „kein zweites paralleles Bestandsmodell". Zwei
Spalten, die dasselbe sagen, sagen es irgendwann verschieden.

## Folgen

- **Migration `0070_ausfuehrung.sql`** baut `sammlung` und
  `wiederherstellungsmengen` neu auf. **Ein Tabellenneubau und kein
  `ALTER`**: `foil` steht in beiden in einer `UNIQUE`-Bedingung, und die kann
  SQLite nicht ablegen. Übersetzt wird `CASE foil WHEN 1 THEN 'foil' ELSE
  'nonfoil' END` — Satz 1.
- **Die Sicht kennt eine dritte Wahl** (`domain.AusfuehrungEtched`), an jeder
  Fläche: `--etched` an der Kommandozeile, `ausfuehrung: "etched"` sonst.
- **Die Kommandozeile hat eine eigene Gestalt bekommen** (`cmd/mse`,
  `exemplarsicht`). Sie schüttete bis dahin die **Wails-DTO** aus — innen und
  außen über einen Typ, was 0033 Regel 5 trennt.
- **Zwei Stellen tun das weiter**, und sie stehen benannt: `mse --json
  ruecksprung-lage` und `GET /ruecksprung/lage` reichen
  `app.RuecksprunglageAntwort` durch. Solange das so ist, trägt `app.MengeDTO`
  den abgekündigten `foil` mit.
