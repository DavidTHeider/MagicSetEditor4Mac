# Schlüsselwortfähigkeiten — welche zwölf gelten, und warum gerade die

**Stand: 2026-09-18.** Diese Datei beantwortet die Frage, die das Issue zu
`M14` (#1616)
stellt: **CR 702 trägt 771 Regeln** — ein Viertel des Regelwerks —, und „nach
Häufigkeit geschnitten" ist die richtige Idee und noch kein Schnitt.

Hier steht der Schnitt, die **Zählung**, aus der er folgt, und die Reichweite
dieser Zählung.

---

## 1 Wie gezählt wurde

**Die Stichprobe sind die zehn kanonischen Sets** unter
`~/Library/Application Support/MagicSetEditor4Mac/kanonische-sets/` —
`White`, `Blue`, `Black`, `Red`, `Green`, `Multicolor`, `Colorless`,
`Archenemy`, `Planechase`, `Vanguard`. Woher sie stammen, steht in
[0026](entscheidungen/0026-mse2-rendert-die-referenz-selbst.md):
`archive.org/details/mse-canon-cards-2022-10-24`, Daten aus Scryfall mit
Stand **2022-10-24**.

**Der Lauf steht im Baum** und heißt
[`internal/werkzeuge/schluesselwoerterzaehlen_test.go`](../internal/werkzeuge/schluesselwoerterzaehlen_test.go).
Er liegt hinter dem Build-Tag `fremdpakete`: die Sets kommen nicht ins
Repositorium (§11), und ein Lauf, der sie braucht, gehört in keine Pipeline
(§3).

```bash
go test -tags=fremdpakete ./internal/werkzeuge/ -run WieOftEinSchluesselwort -v
```

**Gezählt wird, auf wie vielen Karten ein Schlüsselwort *steht***, nicht wie
oft ein Text es nennt. Der Unterschied ist der Kern dieser Datei.

| | |
| --- | ---: |
| Sets | **10** |
| Karten | **25.148** |
| davon mit Regeltext | **24.422** |

---

## 2 Zwei Zahlen, und nur eine beantwortet die Frage

**„Kommt das Wort vor" ist nicht „hat die Karte die Fähigkeit".** „Kann nicht
von Kreaturen ohne Flugfähigkeit geblockt werden" nennt Flugfähigkeit und gibt
sie niemandem.

**`M6` (#1608)
las als Teilzeichenkette**, und der Preis stand als Grenze in der Regelmatrix:
„ein Kartentext, der von Erstschlag **spricht**, ohne ihn zu haben, würde hier
mitgelesen." Was der Satz nicht sagte, war **wie viel**.

**Jetzt ist es gemessen**, an derselben Stichprobe:

| Schlüsselwort | Text **nennt** es | Karte **hat** es | zu viel |
| --- | ---: | ---: | ---: |
| `flying` | 3.349 | **2.315** | +45 % |
| `trample` | 1.049 | **622** | +69 % |
| `haste` | 1.042 | **432** | +141 % |
| `first strike` | 604 | **288** | +110 % |
| `deathtouch` | 332 | **185** | +79 % |

**Seit `M14` wird die Zeile gelesen** — CR 702.1: ein Objekt nennt nur den
**Namen** der Fähigkeit, und ein Erinnerungstext fasst die Regel zusammen.
Sie steht deshalb auf einer eigenen Zeile, mehrere durch Kommas getrennt;
CR 207.2a beschreibt den Erinnerungstext in Klammern.

> **Hier stand CR 702.1b**, und die Nummer war falsch: sie handelt von
> **variablen Werten** in einer verliehenen Schlüsselwortfähigkeit.
> Angestrichen vom Zweitblick über Phase 5.5 (§11). Eine Zeile,
die **nur** aus bekannten Wörtern besteht, gibt sie — jede andere nicht.

**Was das kostet, gehört dazu** (§10): eine Zeile mit einem Schlüsselwort, das
dieses Programm nicht führt — „Flying, ward {2}" — gibt **gar nichts**. Das
ist die richtige Richtung des Irrtums: lieber eine Fähigkeit zu wenig als
eine, die niemand hat.

---

## 3 Die Zählung

**57 Namen aus CR 702**, über die gezählt wurde; aufgeführt ist, auf wie
vielen der 24.422 Karten mit Regeltext die Fähigkeit **steht**.

| # | Schlüsselwort | Karten | Anteil | gebaut |
| ---: | --- | ---: | ---: | :---: |
| 1 | `flying` (CR 702.9) | 2.315 | 9,5 % | ✅ |
| 2 | `trample` (CR 702.19) | 622 | 2,5 % | ✅ |
| 3 | `haste` (CR 702.10) | 432 | 1,8 % | ✅ |
| 4 | `flash` (CR 702.8) | 376 | 1,5 % | — |
| 5 | `vigilance` (CR 702.20) | 358 | 1,5 % | ✅ |
| 6 | `first strike` (CR 702.7) | 288 | 1,2 % | ✅ |
| 7 | `defender` (CR 702.3) | 268 | 1,1 % | ✅ |
| 8 | `reach` (CR 702.17) | 207 | 0,8 % | ✅ |
| 9 | `lifelink` (CR 702.15) | 198 | 0,8 % | ✅ |
| 10 | `deathtouch` (CR 702.2) | 185 | 0,8 % | ✅ |
| 11 | `menace` (CR 702.111) | 165 | 0,7 % | ✅ |
| 12 | `devoid` (CR 702.114) | 98 | 0,4 % | — |
| 13 | `double strike` (CR 702.4) | 75 | 0,3 % | ✅ |
| 14 | `indestructible` (CR 702.12) | 75 | 0,3 % | ✅ |
| 15 | `hexproof` (CR 702.11) | 62 | 0,3 % | — |
| 16 | `convoke` (CR 702.51) | 52 | 0,2 % | — |
| 17 | `changeling` (CR 702.73) | 47 | 0,2 % | — |
| 18 | `infect` (CR 702.90) | 45 | 0,2 % | — |
| 19 | `prowess` (CR 702.108) | 44 | 0,2 % | — |
| 20 | `swampwalk` (CR 702.14) | 40 | 0,2 % | — |

**Danach wird es dünn**: `protection` kommt über alle fünf Farben auf 127,
`shadow` auf 35, `cascade` auf 34, `exalted` auf 33, `shroud` auf 33 — und ab
`islandwalk` (31) liegt jedes einzelne unter **einem Promille** der Karten.
Die vollständige Liste gibt der Lauf aus.

---

## 4 Der Schnitt, und was ihn begründet

**Zwölf werden gebaut.** Elf davon sind die häufigsten, die dieses Programm
auch **durchsetzen** kann; der zwölfte ist `indestructible`.

| Schlüsselwort | Was es hier tut | Regel |
| --- | --- | --- |
| Flugfähigkeit | geblockt nur von Flugfähigkeit oder Reichweite | CR 702.9b |
| Reichweite | blockt Flugfähiges | CR 702.17b |
| Bedrohlich | geblockt nur von zweien oder mehr | CR 702.111b |
| Verteidiger | greift nicht an | CR 702.3b |
| Wachsamkeit | das Angreifen tappt nicht | CR 702.20b |
| Erstschlag | teilt im ersten Schadensschritt aus | CR 702.7b, CR 510.4 |
| Doppelschlag | teilt in **beiden** aus | CR 702.4b |
| Eile | hebt die Einsatzverzögerung auf | CR 702.10b, CR 302.6 |
| Trampelschaden | der Rest geht am Blocker vorbei | CR 702.19b |
| Lebensverknüpfung | Schaden gibt ebenso viel Leben | CR 702.15b |
| Todesberührung | jeder Schaden zerstört | CR 702.2b, CR 704.5h |
| Unzerstörbarkeit | wird nicht zerstört | CR 702.12b, CR 704.5g |

**Warum `indestructible` trotz Platz 14 dabei ist**: es hebt eine
**zustandsbasierte Aktion** auf (CR 704.5g), und die läuft seit `M5`. Eine
Regel, die für ihn nicht gälte, wäre **falsch** statt unvollständig — eine
unzerstörbare Kreatur stürbe am Kampfschaden, und niemand sähe warum.

**Warum `flash` (Platz 4) nicht dabei ist**: es ändert, **wann** ein Zauber
gewirkt werden darf (CR 307.1), und Zeitregeln setzt dieses Programm nicht
durch — `objekt.wirken` fragt nach Priorität und nach dem Abschnitt, nicht
nach dem Kartentyp. Ein `flash`, das nichts erlaubt, was ohnehin erlaubt ist,
wäre eine Zusage ohne Wirkung.

**Warum `devoid` (Platz 12) nicht dabei ist**: es setzt die Farbe eines
Objekts (CR 702.114a), und eine Farbe führt dieses Programm nicht — CR 105
steht in `docs/REGELABDECKUNG.md` als offen.

**Warum `protection` (127 über fünf Farben) nicht dabei ist**: CR 702.16s
nennt **vier** Wirkungen — nicht verzaubert, nicht geblockt, nicht als Ziel,
kein Schaden — und jede davon fragt nach der Farbe. Es ist das Schlüsselwort
mit dem schlechtesten Verhältnis von Häufigkeit zu Aufwand.

**Und `phasing` (CR 702.26) steht ausdrücklich dabei, obwohl es in dieser
Stichprobe gar nicht vorkommt**: das Issue verlangt es namentlich. Es ist
**nicht gebaut**, und die Grenze „kein Phasing im Enttappen-Schritt (CR 502.1)"
steht deshalb weiter in der Regelmatrix — eine Grenze, die in einer gezählten
Teilmenge still verschwindet, wäre eine Zusage, die niemand zurücknimmt (§10).

---

## 5 Was der Tisch meldet

**Eine nicht umgesetzte Fähigkeit auf einer Karte ist ein Befund am Tisch und
kein stiller Ausfall** — die dritte Abnahme des Issues.

**Fünfundvierzig der 57 gezählten Namen** stehen in
[`internal/domain/schluesselbefund.go`](../internal/domain/schluesselbefund.go).
Liegt eine Karte mit einem davon auf dem Schlachtfeld, sagt die Partie es —
an allen vier Flächen, über `PartieAntwort.Befunde`.

**Was das nicht fängt**, und zwar benannt:

- **die 137 weiteren Schlüsselwörter aus CR 702**, die in den kanonischen Sets
  nicht vorkamen. Sie stehen in keiner Liste und werden deshalb auch nicht
  gemeldet;
- **jede Fähigkeit, die kein Schlüsselwort ist** — „Wann immer diese Kreatur
  angreift, ziehe eine Karte" ist ein Satz und braucht `M8`
  (#1610);
- **einen Effekt, der ein Schlüsselwort verleiht** (CR 613) — gelesen wird der
  **gedruckte** Text;
- **ein deutsches Paket**: gelesen wird englisch, so wie es in der Datei steht
  (§9).

---

## 6 Die Reichweite dieser Zahlen

**Die Reichweite einer Messung ist die Reichweite ihrer Aussage** (§10).

- **Die Stichprobe ist nach Farben geschnitten**, nicht nach Erscheinungsjahr
  oder Format. Was nach dem **2022-10-24** gedruckt wurde, fehlt ganz — `toxic`
  (6) und `backup` (1) stehen deshalb so weit unten.
- **Gezählt sind Karten, nicht gespielte Karten.** Eine Fähigkeit auf tausend
  Karten, die niemand spielt, steht hier oben.
- **Die Zeilenlesart zählt konservativ**: eine Zeile mit einem unbekannten
  Wort daneben gibt nichts. Die wahren Zahlen liegen also **über** den
  genannten, nicht darunter.
- **`landwalk` ist ein Sammelname** (CR 702.14) und zählt neben den fünf
  einzelnen Ländern mit.
