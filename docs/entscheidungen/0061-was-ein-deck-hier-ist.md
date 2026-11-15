# 0061 — Was ein Deck hier ist

**Datum:** 2026-09-01
**Status:** angenommen
**Betrifft:** Phase 4.6 (#1293), Werkstück 1

---

## Der Anlass

Phase 4.6 baut Decks und eine Formatprüfung. Bevor eine Zeile davon entsteht,
muss dastehen, **was ein Deck hier ist** — denn anders als bei fast allem
anderen in diesem Programm gibt es dafür keine Vorlage.

**Nachgesehen am 2026-09-01**, in allen drei Quellen aus §11:

| gesucht | `mse2-quelle` (`f9d9356`, 2.1.2) | `*.mse-game`, `*.mse-style` | `magic-set-generator` |
| --- | --- | --- | --- |
| `deck` | 0 in `src/` — der einzige Treffer ist ein deutscher Kommentar in `get_member.cpp`, der „deckt" enthält | in **Statistikbeschreibungen** und in **Erinnerungstexten** („Contraption deck") | **9** — `json-to-mse.py` (8) und `src/lib.rs` (1, ein `//TODO`) |
| `sideboard` | 0 | **0** | **1** — `json-to-mse.py`, Zeile 1151 |
| `legality`, `legal` | 0 | 0 | **1** — `json-to-mse.py`, Zeile 417 |
| `tournament` | 0 | **2** — beides `name: tournament pack`, eine **Packart** | 0 |

**Die Reichweite dieser Zahlen ist genannt** (§10): gesucht wurde in
`src/` von MSE2, in den `.mse-game`- und `.mse-style`-Paketen des Packs und —
**seit der Berichtigung unten** — im **ganzen** Baum von
`magic-set-generator`, in `*.py`, `*.rs` und `*.toml`, **ohne Rücksicht auf
Groß- und Kleinschreibung**. Das letzte ist keine Kleinigkeit: `Sideboard:`
steht dort mit großem S, und eine zeichenempfindliche Suche zählte es nicht. **Die Wörterbücher des Packs sind nicht mitgezählt** —
`en_US.dic` und seine Geschwister führen „sideboard" als Wort der englischen
Sprache, und das ist keine Aussage über MSE.

**Der erste Anlauf dieser Tabelle sagte „sideboard: 0" ohne diese Einschränkung
und war damit falsch.**

**Und der zweite ließ eine Zeile stehen.** Als der Zweitblick die dritte Spalte
auf den ganzen Baum ausweitete, wurde die Zeile `deck` berichtigt — die Zeile
`legality` darunter blieb auf `0`, obwohl dieselbe Ausweitung auch sie betraf.
Angestrichen von Befund 5 des Nachlaufs. Der Fehler ist derselbe wie beim
ersten Mal, nur eine Zeile tiefer: **behoben wurde der genannte Fall, nicht die
Klasse** (§10). Die Suche ist deshalb jetzt für **jede** Zeile über denselben
Baum gelaufen, und das Ergebnis steht in derselben Tabelle.

### Was der eine `legality`-Treffer ist

```python
        # check for legality
        if card_info.layout == 'token':
            raise ValueError('Token cards are not supported')
```

`json-to-mse.py`, Zeile 417. **Das ist keine Formatprüfung**, sondern die
Auskunft, dass dieses Werkzeug Tokens und Un-Cards nicht schreibt — dieselbe
Art Prüfung wie „diese Datei ist kein ZIP". Der Schluss unten bleibt: eine
Vorlage für ein Deckmodell gibt es in keiner der drei Quellen.

**Der Unterschied zählt trotzdem.** Eine Null, die eine Eins ist, macht jede
andere Zahl der Tabelle fraglich — und die Tabelle ist der ganze Grund, warum
der Vermerk sagen darf „hier gibt es keine Vorlage".

### Und die dritte Spalte war zu eng gesucht

**Angestrichen von Befund 10 des Zweitblicks über Phase 4.6.** Die Suche in
`magic-set-generator` lief nur über `src/` — das Werkzeug hat aber ein
Einstiegsskript daneben, und **das kennt Decklisten**:

```python
card_names |= {
    line.split(' ', 1)[1].replace('’', "'")
    for line in response.text.splitlines()
    if line not in ('', 'Sideboard:')
}
```

`json-to-mse.py`, Zeilen 1148–1152, nachgelesen am 2026-09-01.

**Was das ändert und was nicht.** Es **sammelt Kartennamen** aus einer
Deckliste im Netz und wirft dabei die Zeile `Sideboard:` weg — es ist ein
Namensfilter und kein Deckmodell: keine Mengen, keine Bereiche, keine
Zuordnung, keine Prüfung. Die Aussage dieses Vermerks bleibt damit stehen.

**Falsch war die Zahl, nicht der Schluss** — und der Unterschied ist der
Grund für §10: eine zu eng gesuchte Null sieht aus wie eine Messung.

**Damit ist die Lage eindeutig**: Deck, Sideboard, Formatprüfung und Turnier
sind **keine Portierung**. Was das Pack führt, sind Wörter in Regeltexten und
eine Packart namens „tournament pack" — kein Datenmodell und keine Regel. Sie sind neue Produktentscheidungen und gehören
hierher, nicht in die Paritätsmatrix (§15: `new` darf nicht benutzt werden, um
eine unterlassene Suche zu verdecken — hier steht, wo gesucht wurde).

**Ein Anker ist trotzdem da**, und er kommt aus dem Paket: `magic.mse-game`
rechnet in seinen Statistiken schon heute damit, dass ein Set **ein Deck sein
kann** — `average opening hand`, `land drop hit % (On The Play)` und
`(On The Draw)`. Seit Phase 4.11 rechnen alle drei. Das ist keine
Datenstruktur, aber es ist eine Aussage über die Absicht: **ein Set kann ein
Deck sein.**

---

## Die Entscheidung

### 1. Ein Deck ist ein eigener Datensatz, kein Set

**Ein Set ist ein Entwurf, ein Deck ist eine Zusammenstellung.** Sie beantworten
verschiedene Fragen:

| | Set | Deck |
| --- | --- | --- |
| Was steht darin | Karten, die jemand **entwirft** | Karten, die jemand **spielt** |
| Wie oft dieselbe | einmal — eine Karte ist ein Entwurf | mehrfach — vier Wald sind vier Wald |
| Wo es hingehört | eine `.mse-set`-Datei, weitergebbar | der Bestand dieses Rechners |
| Woher die Karten kommen | aus diesem Set | aus **jedem** Set und aus Scryfall |

**Der letzte Punkt entscheidet es.** Ein Deck aus eigenen Karten zieht aus
mehreren Sets zugleich; ein Deck aus echten Karten enthält gar keine eigene
Karte. Wäre das Deck ein Set, müsste es Karten **kopieren** — und eine
berichtigte Karte im Ursprungsset erreichte das Deck nie. Dieselbe Überlegung
wie bei den Autoersetzungen
([0060](0060-eigene-autoersetzungen-liegen-im-bestand.md)): **verweisen statt
kopieren.**

### 2. Ein Deckplatz verweist — und zwar auf eines von beidem

Ein Eintrag im Deck nennt **entweder** eine eigene Karte **oder** einen echten
Druck:

    Deckplatz
      karteID   — eine Karte im Bestand (ein Entwurf)
        oder
      printing  — ein Druck bei Scryfall (eine UUID)
      menge     — wie oft
      bereich   — Hauptdeck oder Sideboard

**Niemals beides und niemals keines.** Ein Platz ohne Bezug ist kein Platz; ein
Platz mit beidem ließe offen, was gilt.

**`printing` ist derselbe Bezug wie in der Sammlung** (`sammlung.printing`) und
im Kartenverweis (`domain.Kartenverweis.Printing`). Das ist der Grund, warum
„habe ich das?" später eine Verknüpfung ist und keine Namenssuche: dieselbe
Kennung steht auf beiden Seiten.

**Und eine eigene Karte mit gesetztem Kartenverweis ist beides zugleich** — sie
ist ein Entwurf **und** zeigt auf einen Druck. Für die Formatprüfung zählt dann
der Druck; das ist die einzige Stelle, an der der Verweis mehr ist als Anzeige.

### 3. Hauptdeck und Sideboard sind ein Feld, keine zwei Listen

Zwei Listen hießen zwei Wege für jede Handlung — anlegen, verschieben,
zählen —, und jede Regel müsste zweimal geschrieben werden. **Ein `bereich`**
je Platz hält sie zusammen.

**Warum überhaupt ein Sideboard**: es ist der einzige Teil eines Decks, der
außerhalb einer Partie existiert, und jedes Format hat eigene Zahlen dafür. Es
wegzulassen hieße, die Formatprüfung später zu zerreißen.

### 4. Ein Deck gehört zu einem Spiel, nicht zu einem Set

`magic`, `space`, `vs` — ein Format ist immer das eines Spiels. **Ohne diese
Angabe ließe sich kein Format zuordnen**, und ein Deck aus Karten zweier Spiele
wäre keines.

### 5. Fassungen: der Bestandsverlauf, kein eigener

Ein Deck ändert sich, und wer es ändert, will zurückkönnen. **Das kann der
Bestand schon** — jede schreibende Handlung geht in denselben Verlauf wie
Karten und Sets (⌘Z, `zuruecknehmen`).

**Kein eigenes Fassungswesen**, und das ist eine bewusste Auslassung: „Deck
2026-09 gegen Deck 2026-08 stellen" ist eine andere Fähigkeit — sie hat einen
Namen (Versionierung) und gehört in ein eigenes Issue, wenn jemand sie braucht.
Was hier steht, ist **eine** Fassung je Deck, mit Rücknahme.

### 6. Ein Deck wird nicht in eine `.mse-set` geschrieben

Es liegt im Bestand, wie die Sammlung
([0052](0052-die-sammlung-ist-nicht-das-set.md)). Ein Deck in eine Setdatei zu
schreiben machte die Datei unlesbar für MSE2 — und eine Datei, die MSE2 nicht
mehr öffnet, verletzt die Datenkompatibilität, die dieses Projekt zusagt.

**Ausfuhr ist etwas anderes und kommt später**: eine Deckliste als Text ist ein
Ausgabeformat, kein Speicherort.

---

## Was daraus für die Formatprüfung folgt

**Drei Antworten, und sie sind verschieden.** Das ist die eigentliche Zusage
dieses Vermerks:

| Die Karte ist | Die Antwort kommt von | Wie sie heißt |
| --- | --- | --- |
| ein **echter Druck** (`printing`) | **Scryfall**, `legalities` je Druck | `legal`, `not_legal`, `restricted`, `banned` — Scryfalls eigene Wörter |
| eine **eigene Karte ohne Verweis** | niemandem | **`ungeprüft`** |
| eine **eigene Karte mit Verweis** | Scryfall, über den Druck | wie oben, und die Herkunft steht dabei |

**`ungeprüft` ist kein `nicht zulässig`.** Eine selbst entworfene Karte ist
naturgemäß in keinem Format der Welt gelistet; sie deshalb rot zu färben hieße,
jedem eigenen Deck zu sagen, es sei kaputt. Das Programm sagt stattdessen, was
es weiß und was nicht — dieselbe Haltung wie bei den Vorbehalten eines
eingefahrenen Sets (§8).

**Und die Zulässigkeit wird nie erfunden.** Es gibt keine eingebaute Liste
verbotener Karten und keine geratene Regel. Was nicht aus Scryfall kommt, ist
`ungeprüft` — auch dann, wenn eine Karte „offensichtlich" verboten wäre.

---

## Was ein Format hier ist

**Daten, nicht Code.** Ein Format ist ein Datensatz aus Name, Spiel,
Deckgrößen, Sideboardgröße und der Zahl gleicher Karten:

    Format
      name              — "Standard", "Commander", "Mein Küchenformat"
      spiel             — magic
      mindestgroesse    — 60
      hoechstgroesse    — 0 heißt: keine Grenze
      sideboardgroesse  — 15
      hoechstzahlGleich — 4
      scryfallname      — "standard", leer für ein eigenes Format

**Warum als Daten**: wer ein eigenes Format spielt, soll es anlegen können,
ohne dass jemand das Programm ändert. Und `scryfallname` ist die Brücke: er
sagt, welches Feld in `legalities` gilt — leer heißt, dass es dort keines gibt
und jede Karte `ungeprüft` bleibt.

**Was ein Format hier nicht kann**: Karten einzeln erlauben oder verbieten,
Farbidentität prüfen (Commander), Rangfolgen zwischen Formaten. Das sind
eigene Werkstücke; dieser Vermerk sagt nur, dass das Modell sie nicht
ausschließt.

---

## Verworfen

**Ein Deck als Set mit einem Flag.** Es hätte die Statistiken sofort geliefert
(sie rechnen ohnehin über `set.cards`) und wäre in einer `.mse-set`
gespeichert. Verworfen aus zwei Gründen: eine Karte müsste kopiert werden statt
verwiesen, und mehrfach dieselbe Karte ginge gar nicht — ein Set kennt jede
Karte einmal.

**Eine Zulässigkeit, die das Programm selbst rechnet.** Verworfen, weil sie
täglich veraltet und weil jede Erfindung eine Aussage über fremdes Verhalten
wäre, die niemand nachlesen kann (§10).

**Ein Deck ohne Spiel.** Verworfen: dann gäbe es kein Format.

---

## Abnahme dieses Vermerks

Er ist eingelöst, wenn das Domänenmodell des nächsten Werkstücks genau diese
Gestalt hat — und wenn eine eigene Karte im Deck den Status **`ungeprüft`**
trägt und nicht `nicht zulässig`.
