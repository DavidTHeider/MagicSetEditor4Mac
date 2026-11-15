# 0062 — Was ein Turnier hier ist

**Datum:** 2026-09-01
**Status:** angenommen
**Betrifft:** Phase 4.7 (#1294), Werkstück 1

---

## Der Anlass

Phase 4.7 baut Turniere. Wie beim Deck gibt es dafür **keine Vorlage in den
drei Quellen aus §11** — und anders als beim Deck gibt es eine vierte, die
fast alles beantwortet.

**Nachgesehen am 2026-09-01:**

| gesucht | `mse2-quelle` (`f9d9356`) | `*.mse-game`, `*.mse-style` | `magic-set-generator` |
| --- | --- | --- | --- |
| `tournament` | **0** | **6** — zweimal `name: tournament pack` (eine **Packart**, in `magic.mse-game/packs` und `space.mse-game/packs`), viermal die Kartennamen „Spike, Tournament Grinder" und „Tournament Grounds", jeder einmal mit und einmal ohne Leerzeichen in `magic-namecheck.mse-export-template/namelist` | **0** |
| `pairing` | **0** | **0** | **0** |
| `standings` | **0** | **1** — der Kartenname „StandingStones" | **0** |
| `swiss` | **9 — und keiner davon ist gemeint**: alle neun sind `wxFONTFAMILY_SWISS`, eine **Schriftfamilie** | **0** | **0** |
| `tiebreak` | **2 — und auch die nicht**: ein `int tiebreaker` in `data/pack.cpp:180`, der **Packstücke** sortiert | **0** | **0** |
| `round` | **197 — keiner davon eine Turnierrunde**: `background`, `around`, `DrawRoundedRectangle`, `rounding` | — | — |

**Die Reichweite dieser Zahlen** (§10): gesucht wurde in `src/` von MSE2, im
**ganzen** `data/`-Verzeichnis des Packs und im **ganzen** Baum von
`magic-set-generator` (`*.py`, `*.rs`, `*.toml`), ohne Rücksicht auf Groß- und
Kleinschreibung.

**Gezählt sind Vorkommen, nicht Zeilen** — `grep -roi … | wc -l`, nicht
`grep -ric`. Der Unterschied ist keiner der Genauigkeit: eine Zeile kann
vier Treffer tragen, und die Zeile in `namelist` tut es. **Die mittlere
Spalte nannte zuerst 4**, weil sie Zeilen zählte und die Einheit nicht
dazusagte; nachgezählt sind es 6 — angestrichen von Befund 5 des Nachlaufs
über Phase 4.7. Der Handgriff steht hier, damit ihn jemand wiederholen kann:

```bash
grep -roi --exclude="*.dic" --exclude="*.aff" "tournament" data | wc -l
```

**Die Wörterbücher des Packs sind nicht mitgezählt** — dieselbe Einschränkung
wie in [0061](0061-was-ein-deck-hier-ist.md). `en_US.dic` und seine
Geschwister führen „tournament", „swiss" und „pairing" als Wörter der
englischen Sprache; über MSE sagt das nichts. **Mitgezählt sind sie in der
Spalte trotzdem gewesen, bis nachgesehen wurde**: 41 der 41 `swiss`-Treffer im
Pack stehen dort, und ohne die Wörterbücher bleibt null.

**Die mittlere Spalte nannte zuerst nur `*.mse-game` und `*.mse-style` als
Bereich** und zählte trotzdem Treffer aus einer `.mse-export-template` mit —
angestrichen von Befund 4 des Zweitblicks über Phase 4.7. Der Bereich heißt
jetzt, was wirklich durchsucht wurde: das ganze `data/`. **Eine Messung, deren
Aussage weiter reicht als ihr genannter Bereich, ist nicht nachvollziehbar,
auch wenn ihre Zahlen stimmen.**

**Der erste Anlauf dieser Tabelle schrieb bei `swiss`, `tiebreak` und in der
mittleren Spalte Nullen — und keine davon stimmte.** Sie stand da, weil sie plausibel war und weil der
Schluss ohnehin feststand. Genau so ist Befund 5 des Nachlaufs über Phase 4.6
entstanden. Nachgezählt sind es neun und zwei, beide Male ein **Homonym** —
dieselbe Sorte Treffer wie der deutsche Kommentar mit „deckt" in
[0061](0061-was-ein-deck-hier-ist.md). Der Schluss trägt trotzdem: **eine
Zahl, die man nicht nachgesehen hat, trägt nichts, auch wenn sie stimmt.**

**Und die mittlere Spalte war zweimal falsch gesucht.** Die Pakete des Packs
sind **Verzeichnisse** (`magic.mse-game/`), keine Dateien mit dieser Endung —
ein `--include="*.mse-game"` findet darin **nichts** und meldet null. Dieselbe
Falle wie das große `S` in `Sideboard:` beim Deckvermerk: der Filter war
falsch, nicht die Quelle.

MSE2 zeichnet Karten. Ein Turnier ist keine Karte, und das ist keine Lücke,
sondern die Grenze des Vorbilds.

## Die vierte Quelle, und sie ist maßgeblich

**Die *Magic Tournament Rules*** — die Regeln, nach denen wirklich gespielt
wird. Sie sind keine Vorlage für ein Datenmodell, aber sie beantworten **jede**
Frage nach den Zahlen, und sie tun es so genau, dass jedes Raten daneben ein
Fehler wäre.

**Abgerufen am 2026-09-01:**
`https://media.wizards.com/ContentResources/WPN/MTG_MTR_2026_Feb27_EN.pdf` —
Fassung **vom 27. Februar 2026**, verlinkt von der Seite
<https://wpn.wizards.com/en/rules-documents>.

**Sie wird nicht mitgeliefert** (§11): sie ist fremdes Material, und ihr
Urheber ist Wizards of the Coast. Was hier steht, ist die **Zuordnung** auf
dieses Programm, mit Fundstelle.

### §3.1 — die vier Tiebreaker, in dieser Reihenfolge

1. Match points
2. Opponents' match-win percentage
3. Game-win percentage
4. Opponents' game-win percentage

Und der Satz, der eine eigene Entscheidung verlangt: „Not all of these
tiebreakers may be used in formats with single-game matches."

### Appendix C — die Formeln

| Größe | Regel |
| --- | --- |
| Match points | 3 je Sieg, 1 je Unentschieden, 0 je Niederlage |
| Game points | 3 je gewonnenes Spiel, 1 je unentschiedenes, 0 je verlorenes. **Unfertige Spiele zählen als unentschieden, ungespielte als 0** |
| Match-win % | Match points ÷ (3 × gespielte Runden), **Untergrenze 0,33** |
| Game-win % | Game points ÷ (3 × gespielte Spiele), **Untergrenze 0,33** |
| Opponents' match-win % | Mittel der Match-win % **jedes** Gegners; die eigenen Byes bleiben unberücksichtigt |
| Opponents' game-win % | dasselbe für Game-win %, jeder Gegner mit derselben Untergrenze |
| Bye | gilt als **2–0**: 3 Match points, 6 Game points |

Die Untergrenze steht dort mit Beispiel: ein Spieler mit 1–3–0, der aussteigt,
hat 3/(4·3) = 0,25 — **gerechnet wird mit 0,33**. Der Grund steht daneben, und
er ist der Grund, warum die Grenze nicht weggelassen werden darf: sie
„limits the effect low performances have when calculating and comparing
opponents' match-win percentage".

### Appendix E — die empfohlene Rundenzahl

| Spieler | Swiss-Runden | Playoff |
| --- | --- | --- |
| 4 (nur Team/2HG) | **„2 Single-Elimination Rounds (No Swiss)"** | „None (Run Single Elimination)" |
| 5–8 | **„3 Single-Elimination Rounds (No Swiss)"** | „None (Run Single Elimination)" |
| 9–16 | 4 (Limited mit Draft im Playoff) / 5 (sonst) | Top 8 / Top 4 |
| 17–32 | 5 | Top 8 |
| 33–64 | 6 | Top 8 |
| 65–128 | 7 | Top 8 |
| 129–226 | 8 | Top 8 |
| 227–409 | 9 | Top 8 |
| 410+ | 10 | Top 8 |

**Die ersten beiden Zeilen nennen keine Swiss-Runde.** Sie stehen in der
Spalte „Swiss Rounds", und dort steht wörtlich, dass es **keine** ist. Das
sind K.-o.-Runden, und dieses Programm baut kein K.-o.-System (Verworfen,
unten). Für bis zu acht Teilnehmer hat Appendix E deshalb **keine**
Empfehlung, die hierher passt — [domain.EmpfohleneRunden] gibt dort null.

**Der erste Anlauf las die 3 als drei Swiss-Runden**, weil sie in dieser
Spalte steht — und die Abnahme zu
#1294
lautete darauf „acht Teilnehmer, drei Runden nach Appendix E". Die Zahl stand
richtig da und meinte etwas anderes; angestrichen von Befund 4 des Nachlaufs
über Phase 4.7. **Eine Quelle wird gelesen und nicht überflogen** — auch
dann, wenn die Zelle eine Zahl enthält, die zur erwarteten Antwort passt.

### §10.4 — das Paarungssystem

„Unless otherwise announced, tournaments are assumed to follow the Swiss
pairing algorithm."

---

## Die Entscheidung

### 1. Ein Turnier ist ein Verlauf, kein Zustand

**Es wird nicht gerechnet und abgelegt, sondern geführt.** Ein Turnier hält
seine Runden, jede Runde ihre Paarungen, jede Paarung ihr Ergebnis. Die
Rangliste ist **abgeleitet** und wird nirgends gespeichert.

**Warum das die wichtigste Zeile dieses Vermerks ist:** eine abgelegte Tabelle
und ein Verlauf, aus dem sie folgt, laufen auseinander, sobald jemand ein
Ergebnis berichtigt. Und Ergebnisse werden berichtigt — das ist der Normalfall
am Turniertisch, nicht der Ausnahmefall.

**Was es kostet:** die Rangliste wird bei jeder Anzeige neu gerechnet. Bei
zehn Runden und vierhundert Spielern sind das vierhunderttausend Multiplikationen,
und das ist nichts.

### 2. Ein Teilnehmer ist ein Name, kein Nutzerkonto

**Wer an einem Küchentisch ein Turnier führt, tippt Namen.** Ein Teilnehmer
trägt einen Namen und **wahlweise** ein Deck aus dem Bestand.

**Kein Verweis auf ein Nutzerkonto** — die gibt es erst in Phase 4.12, und ein
Turnier, das darauf wartete, käme nie.

### 3. Die Deckliste eines Teilnehmers ist ein Abzug, kein Verweis

**Der Preis beider Wege gehört ausgesprochen**, denn das ist die Frage, die
das Issue offen ließ.

| | Verweis auf das Deck | **Abzug beim Anmelden** |
| --- | --- | --- |
| Ein Tippfehler im Deck wird berichtigt | wirkt rückwirkend auf die Runde 1 | wirkt nicht |
| Jemand baut sein Deck während des Turniers um | das Turnier sieht das neue | das Turnier sieht das angemeldete |
| Das Deck wird gelöscht | das Turnier verliert seine Liste | die Liste bleibt |

**Gewählt ist der Abzug**, und zwar aus dem mittleren Grund. Eine Deckliste im
Turnier ist ein **eingereichtes Dokument**; sie nachträglich zu ändern, ist
genau das, was die MTR unter „Deck Problem" führt. Ein Datenmodell, in dem das
unbemerkt geht, bildet die Sache falsch ab.

**Der Verweis geht trotzdem mit** — als Herkunft, damit man das Deck
aufschlagen kann. Er ist die Auskunft „von hier kam es", nicht die Quelle der
Liste.

### 4. Die Rundenzahl ist ein Vorschlag, kein Zwang

Appendix E wird **angeboten**: wer neun Teilnehmer einträgt, bekommt fünf
Runden vorgeschlagen. Übersteuern geht. Bei bis zu acht kommt **keine**
Empfehlung, weil Appendix E dort keine Swiss-Runde nennt (oben).

**Und ab der ersten gelegten Runde steht sie.** MTR §10.2 verbietet, die
Rundenzahl zu ändern, sobald sie angekündigt ist. Vorher folgt die
Empfehlung der Anmeldung, nachher nicht mehr — wer aussteigt, ändert die
angekündigte Zahl nicht. Der erste Anlauf rechnete auch mitten im Turnier
neu, und die Anzeige sprang, sobald jemand ausstieg.

**Warum nicht zwingend:** Appendix E gilt für *Premier tournaments* und ist
sonst ausdrücklich „at the Tournament Organizer's discretion". Ein Programm,
das die Empfehlung erzwänge, behauptete eine Regel, die es nicht gibt.

### 5. Die vier Tiebreaker, alle vier, in der Reihenfolge aus §3.1

**Keine Auswahl, keine Einstellung.** Sie stehen fest, und die Reihenfolge
auch.

**Die eine Ausnahme steht in §3.1 selbst**: bei Formaten mit
Einzelspiel-Matches sind Game-win % und Opponents' game-win % ohne Aussage —
jedes Match hat dann genau ein Spiel, und die Zahl ist eine Kopie der
Match-Zahl. Ein Turnier, das als **Einzelspiel** geführt wird, zeigt sie
deshalb nicht. Es rechnet sie trotzdem: eine Spalte auszublenden ist eine
Anzeigefrage, und wer später doch danach fragt, soll keine Lücke in den Daten
finden.

### 6. Gerechnet wird ungerundet, gerundet wird beim Anzeigen

**Das ist eine eigene Entscheidung** und steht so **nicht** in der MTR. Sie
zeigt Prozentwerte mit drei Nachkommastellen (`0.667`), sagt aber nirgends,
wann gerundet wird.

**Gewählt:** die Zwischenwerte bleiben ungerundet, gerundet wird erst die
Anzeige, auf **vier** Nachkommastellen.

**Warum vier und nicht drei:** die Reihenfolge entscheidet sich an den Zahlen,
und zwei Spieler, die sich in der vierten Stelle unterscheiden, sind nicht
gleichauf. Wer auf drei rundete, machte sie es — und die Rangfolge hinge dann
an der Sortierung.

**Warum überhaupt gerundet wird:** ohne Rundung stünde `0.6666666666666666` in
der Tabelle, und das ist keine Auskunft, sondern eine Fließkommazahl.

### 7. Ein Ergebnis wird als Spielstand eingetragen, nicht als Sieger

**Zwei Zahlen und ein Unentschieden**: Spiele für A, Spiele für B, unentschiedene
Spiele. Wer gewonnen hat, folgt daraus.

**Warum nicht „A gewinnt":** weil Appendix C **Game points** braucht, und die
folgen aus dem Spielstand. Ein Modell, das nur den Sieger hält, könnte
Tiebreaker 3 und 4 nicht rechnen — und das merkte niemand, bis zwei Spieler
gleichauf sind.

### 8. Ein Bye ist ein Ergebnis, kein fehlender Gegner

**Er wird als Paarung ohne Gegner geführt** und zählt als 2–0. Damit steht er
im Verlauf, ist sichtbar und lässt sich zurücknehmen.

**Wer ihn bekommt, sagt die MTR nicht.** Gewählt: **der Letzte im Stand, der
noch keinen hatte** — und wenn alle einen hatten, der Letzte. **Das ist eine
eigene Regel**, und sie steht hier, damit niemand sie später für nachgelesen
hält.

**Hier stand zuerst „bei Gleichstand der zuerst eingetragene", und das war
falsch** — angestrichen von Befund 3 des Zweitblicks über Phase 4.7. Der Satz
widersprach sich selbst: der letzte Tiebreaker der Rangliste **ist** die
Anmeldereihenfolge, und „der Letzte im Stand" heißt bei völligem Gleichstand
deshalb „der zuletzt Angemeldete". Zwei Sätze, die dasselbe meinten und
Gegenteiliges sagten.

**Und der zuletzt Angemeldete ist der richtige.** Ein Bye gibt zwar Punkte,
aber kein Spiel — und wer zuletzt dazukam, sitzt eher aus als wer zuerst da
war. Für einen Küchentisch ist das die Fairness, die zählt.

### 9. Aufgeben heißt aussteigen, nicht verschwinden

Ein Teilnehmer, der aufgibt, bleibt in der Rangliste und in den Paarungen der
vergangenen Runden. Er wird nur nicht mehr gepaart.

**Das ist keine Freundlichkeit, sondern Appendix C**: seine Match-win % geht
weiter in die Opponents' match-win % **jedes** Gegners ein, den er hatte — mit
der Untergrenze 0,33. Wer ihn löschte, veränderte die Rangliste aller anderen.

### 10. Unterbrechen ist kein eigener Zustand

**Ein Turnier ist immer unterbrechbar**, weil nichts im Arbeitsspeicher steht,
was nicht auch im Bestand steht. Es gibt kein „Turnier läuft"-Flag, das ein
Absturz falsch hinterließe.

**Die Zusage aus dem Issue** — „ein Absturz mitten in der Ergebniserfassung
hinterlässt keine halbe Runde" — löst deshalb nicht ein Zustandsfeld ein,
sondern eine **Transaktionsgrenze**: die Paarungen einer Runde entstehen in
**einem** Zug, und ein Ergebnis ist ein Schreibvorgang.

---

## Verworfen

**Ein K.-o.-System.** Es ist einfacher zu bauen und schickt die Hälfte der
Teilnehmer nach einer Runde nach Hause. §10.4 nennt Swiss als Vorgabe, und für
einen Küchentisch ist das auch der richtige Weg.

**Eine gespeicherte Rangliste.** Siehe Entscheidung 1 — sie liefe auseinander,
sobald jemand ein Ergebnis berichtigt.

**Ein eigenes Punktesystem.** Es gäbe keinen Grund dafür, und es machte jede
Auskunft dieses Programms unvergleichbar mit jedem anderen.

**Die Tiebreaker abschaltbar machen.** Eine Einstellung, die niemand versteht
und die jede Rangliste unvergleichbar macht. Die MTR nennt vier, in einer
Reihenfolge; wer eine andere will, rechnet von Hand.

**Die Deckliste als Verweis.** Siehe Entscheidung 3 — sie machte ein
eingereichtes Dokument nachträglich änderbar.

---

## Abnahme dieses Vermerks

Die drei Beispielspieler aus Appendix C kommen als Testfälle vor und ergeben
genau die Werte, die dort stehen:

| Turnierergebnis | Match points | Runden | Match-win % |
| --- | --- | --- | --- |
| 5-2-1 | 16 | 8 | 16/(8·3) = **0,667** |
| 1-3-0, dann Ausstieg | 3 | 4 | 3/(4·3) = 0,25 → **0,33** |
| 3-2-0, mit Erstrundenbye, dann Ausstieg | 9 | 5 | 9/(5·3) = **0,60** |

Und die beiden aus dem Abschnitt *Game-win percentage*:

| Spielverlauf | Game points | Spiele | Game-win % |
| --- | --- | --- | --- |
| 2-0 / 2-1 / 1-2 / 2-0 | 21 | 10 | 21/(3·10) = **0,70** |
| 1-2 / 1-2 / 0-2 / 1-2 | 9 | 11 | 9/(3·11) = 0,27 → **0,33** |
