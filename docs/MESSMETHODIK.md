# Messmethodik: der Abstand zu Magic Set Editor 2

**Wie die Zahlen zustande kommen, die dieses Projekt über seine Nähe zu MSE2
nennt.** Wer eine solche Zahl liest — im `README.md`, in der Roadmap, in einer
Release-Notiz —, soll hier nachlesen können, was gemessen wurde, womit, woran
und mit welchem Vorbehalt.

**Angelegt am 2026-09-07**, auf Anweisung des Auftraggebers. Bis dahin standen
Maß und Statistik als Kommentare am Testcode; wer keine Go-Datei aufschlägt,
fand sie nicht.

> **Der Grundsatz dahinter steht in [`AGENTS.md`](../AGENTS.md) §10:** die
> Reichweite einer Messung ist die Reichweite ihrer Aussage. Wer eine Zahl
> aufschreibt, schreibt dazu, worüber sie erhoben wurde.

---

## 1 Was verglichen wird

**Zwei Rasterungen derselben Datei.** Nicht zwei Karten, nicht zwei Vorlagen —
dieselbe `.mse-set`-Datei, einmal von MSE2 gerastert und einmal von diesem
Programm.

| | |
| --- | --- |
| **Eingabe** | die kanonischen Sets (Abschnitt 2) |
| **Referenz** | MSE2s eigene Ausgabe über `--export-images` |
| **Prüfling** | die Rasterung dieses Programms über dieselben Vorlagenpakete |

**Damit misst der Abstand nicht, ob die Karte richtig ist**, sondern wie weit
zwei Rasterungen derselben Beschreibung auseinanderliegen. Ob die Beschreibung
die Karte trifft, ist eine andere Frage — und die Referenz beantwortet sie
nicht (Abschnitt 2).

Der Weg, auf dem MSE2 die Referenz erzeugt, steht in
[`entscheidungen/0026`](entscheidungen/0026-mse2-rendert-die-referenz-selbst.md).

---

## 2 Woher die Sets stammen

| | |
| ---: | --- |
| **Bezugsort** | <https://archive.org/details/mse-canon-cards-2022-10-24> |
| **Angekündigt** | MSE-Forum, [Thread 2128 „Canon Magic Cards"](https://magicseteditor.boards.net/thread/2128/canon-magic-cards) |
| **Erstellt von** | Forumsnutzer *GenevensiS*, Beitrag vom 2022-10-25 |
| **Datenquelle** | **Scryfall**, Stand **2022-10-24** — Karten **und** Bilder |
| **Umfang** | zehn Dateien, 1,8 GB, **25.148** Karten |
| **Geladen** | 2026-08-17 um 21:24 über Safari (Quarantänemarke der Dateien) |

**Zwei Einschränkungen nennt der Ersteller selbst**, und beide gehören zu jeder
Zahl, die daran gemessen wird:

> „There is bound to be some inaccuracies and maybe missing cards."

> „Silver bordered cards will likely be very off, I don't plan on fixing those
> though."

**Die Referenz ist damit kein Kanon im strengen Sinn.** Sie ist der beste
verfügbare Maßstab: dieselben Sets, die MSE2 selbst öffnet und rastert.

---

## 3 Das Maß

Für zwei Bilder gleicher Größe mit `P` Bildpunkten und `C = 3` Farbkanälen:

```
        ┌─────────────────────────────────┐
        │      1     P    C               │
  d  =  │  ───────  Σ    Σ  (u   − r  )²  │
        │   P · C   p=1  c=1    pc    pc  │
        └─────────────────────────────────┘
                    ↑ Wurzel über allem
```

In einer Zeile: **d = √( Σ (u−r)² / (P·C) )** — die Wurzel aus der mittleren
quadrierten Kanaldifferenz. `u` ist unsere Rasterung, `r` die von MSE2.

| | |
| --- | --- |
| **Skala** | 0 bis 255; **0** heißt punktgleich |
| **Toleranz** | 1,0 — der Wert, bis zu dem die Goldbilder dieses Projekts gleich heißen |
| **Kanäle** | Rot, Grün, Blau — **ohne Alpha** |

**Warum ohne Alpha**: eine MSE2-Ausgabe ist ein deckendes PNG. Ein Alphakanal,
der überall 255 ist, brächte einen vierten Summanden mit Differenz null und
verdünnte den Abstand um ein Viertel.

**Warum quadratisch und nicht als Betrag**: ein quadratisches Mittel gewichtet
große Abweichungen stärker, und genau die entscheiden, ob ein Mensch zwei
Karten unterscheidet. Das Projekt hat **zwei** Bildmaße, und sie tragen
absichtlich verschiedene Namen — `render.Vergleiche` mittelt Beträge, der
Kanonlauf mittelt Quadrate. Zwei Maße unter einem Namen wären genau der Fehler,
den §10 anstreicht.

**Quelle**: `render.Teilabstand.Abstand`, `internal/render/zerlegung.go`.

### Die Zerlegung

Derselbe Abstand, aufgeteilt nach der Ebenenart, die den Bildpunkt **zuletzt
gemalt** hat — Bild, Text, Fläche. Sie beantwortet die Frage, **wo** der
Abstand sitzt:

```
  Anteil(Art) = Quadratsumme(Art) / Quadratsumme(gesamt)
```

**Die Quadratsummen addieren sich, die Abstände nicht** — ein Abstand ist eine
Wurzel. Wer Teilabstände addierte, bekäme eine Zahl ohne Bedeutung.

### Die Versatzprobe

Sie prüft, ob ein Abstand nur daher rührt, dass etwas um ein bis zwei
Bildpunkte verschoben sitzt: für jede Verschiebung `(dx, dy)` mit
`|dx|, |dy| ≤ 2` wird `d` neu gerechnet und das Minimum genommen.

**Karten am Rand des Suchbereichs werden gezählt und genannt** — dort ist der
gewonnene Abstand eine **Untergrenze**, denn der wahre Versatz kann größer
sein.

---

## 4 Die Stichprobe

**Jedes Set ist eine Schicht**, und das ist nicht willkürlich: die Sets sind
nach Farbe und Sonderform getrennt, und daran hängt der Zeichenweg —
Goldrahmen und Blendmasken bei `Multicolor`, Artefaktmasken bei `Colorless`,
Überformate bei `Planechase` und `Vanguard`.

### Zwei Kohorten, zwei Fragen

**Seit QA-02 (#1734)
misst der Kanonlauf zwei Stichproben, und sie beantworten zwei verschiedene
Fragen:**

| | Die feste Kohorte | Die Zufallsstichprobe |
| --- | --- | --- |
| **Fragt** | hat sich die Zahl **bewegt**? | wie groß ist sie, und wie sicher ist das? |
| **Gezogen** | systematisch, gleichmäßig über die Kartenliste | zufällig, einmal, aus einer Saat |
| **Größe je Satz** | die Rechnung vom 2026-09-07, unverändert | exakt hypergeometrisch |
| **Trägt** | einen Vergleich über die Zeit | eine Entdeckungszusage und ein Vertrauensband |
| **Im Beleg** | `mse2.*` | `zufall.*` |
| **Lauf** | `TestDerAbstandZuMSE2AmKanon` | `TestDieZufallsstichprobe` |

**Die feste Kohorte bleibt, wie sie war** (AK-01): dieselben Karten in jedem
Lauf, sonst verglichen zwei Läufe zwei Mengen. `TestDieFesteKohorteBleibtWieSieWar`
hält ihre Größen fest.

### Wie groß die feste Kohorte ist — und warum das keine Zusage mehr ist

```
  n∞  =  ln(1 − c) / ln(1 − q)          c = 0,999   q = 0,01   →  n∞ ≈ 687,3

  n_h =  n∞ / (1 + (n∞ − 1) / N_h)      Endlichkeitskorrektur
```

Das zieht 687 bei `N=107` auf **93** herunter und bei `N=3788` auf **582**.
**Hier stand „580"** — das ist die Zahl für Green und Blue mit rund 3.690
Karten; gezogen hat jeder Lauf für Black 582 (gefunden am 2026-09-22 von dem
Lauf, der die feste Kohorte festnagelt).

**Die zweite Zeile ist die Endlichkeitskorrektur für die Varianz einer
Schätzung — keine Wahrscheinlichkeit, etwas zu finden** (AK-02). Hier stand,
sie sichere „mit 99,9 % mindestens einen Fehler, der 1 % der Karten betrifft".
**Gerechnet für Vanguard mit 107 Karten sind es 98,3953 %, nicht 99,9**: bei
zwei betroffenen Karten entgehen 93 gezogene beiden mit
`(14 · 13) / (107 · 106)`. Die Größe bleibt, weil die Kohorte bleiben soll —
**die Zusage ist gestrichen.**

> **Bis zum 2026-09-07 wurde `n∞` für alle 25.148 Karten gerechnet und dann
> durch zehn geteilt** — rund 68 je Set. Damit trug die Zusage die
> **Vereinigung** und nicht das einzelne Set: 68 aus 3.788 treffen einen Fehler
> in einem Prozent nur etwa jedes zweite Mal. Der Befundlauf
> (`TestDieKanonischenSetsZeichnen`) zieht so bis heute und ist damit
> ebenfalls eine feste Kohorte.

### Wie groß die Zufallsstichprobe ist

**Exakt, ohne Zurücklegen** (AK-02): bei `N` Karten, von denen `K = ⌈q · N⌉`
betroffen sind, entgehen `n` gezogene allen mit

```
  P(keine)  =  C(N−K, n) / C(N, n)  =  ∏_{i<K} (N−n−i) / (N−i)

  n_h       =  min { n : 1 − P(keine) ≥ c }
```

| Satz | Karten | betroffen bei 1 % | Zufall `n_h` | feste Kohorte |
| --- | ---: | ---: | ---: | ---: |
| Archenemy | 70 | 1 | **70** — alle | 64 |
| Planechase | 86 | 1 | **86** — alle | 77 |
| Vanguard | 107 | 2 | **104** | 93 |
| Multicolor | 3.008 | 31 | 598 | 560 |
| Colorless | 3.216 | 33 | 605 | 567 |
| Green | 3.689 | 37 | 626 | 580 |
| Blue | 3.693 | 37 | 626 | 580 |
| Red | 3.730 | 38 | 617 | 581 |
| White | 3.761 | 38 | 623 | 582 |
| Black | 3.788 | 38 | 627 | 582 |
| **zusammen** | **25.148** | | **4.582** | **4.266** |

**Die feste Kohorte zählt seit QA-03 eine Karte mehr**: *Phoebe, Head of
S.N.E.A.K.* stand immer in ihr, fand aber ihr Referenzbild nicht — MSE reinigt
`{card.name}.png` als Ganzes, und hier wurde bis dahin nur der Name gereinigt.

**Bei den kleinen Sätzen kommt jede Karte dran**: ist schon **eine** Karte
mehr als ein Prozent, muss die Stichprobe fast alle enthalten, um sie mit
99,9 % zu treffen. Die Zahlen hält `TestDieNoetigeStichprobeIstDieKleinsteDieReicht`
fest — samt der Gegenprobe, dass eine Karte weniger nicht reichte.

### Je Satz und über alle zugleich

**Zwei Zusagen, und sie zu verwechseln ist der zweithäufigste Fehler mit einer
Stichprobe** (AK-03). **Je Satz**: 99,9 % — jeder Satz für sich. **Über alle
zehn zugleich**: das Produkt, weil die Sätze unabhängig gezogen werden —
**99,25 %** für diesen Plan (die zwei ganz gezogenen tragen 1, Vanguard
0,99947). Wer „in keinem Satz" sagen will, sagt 99,25 und nicht 99,9.

### Welche Karten gezogen werden

**Die feste Kohorte**: gleichmäßig verteilt über die Kartenliste —
`i · (len−1) / (anzahl−1)` trifft die erste und die letzte Karte und verteilt
die übrigen dazwischen. **Warum nicht zufällig**: ein Lauf, der bei jedem
Aufruf andere Karten misst, gibt bei jedem Aufruf eine andere Zahl — und ein
Verlaufsmaß, das schwankt, ohne dass sich etwas geändert hat, ist keins. Der
Preis: eine systematische Auswahl kann eine periodische Struktur der Liste
treffen, und **eine Wahrscheinlichkeitsaussage trägt sie nicht**.

**Die Zufallsstichprobe**: einmal gezogen, mit einer Saat aus dem Zufall des
Systems, und **gespeichert** — `docs/belege/kanonstichprobe.json` nennt die
Saat, je Satz die Kartenzahl, einen Hash über die Kartenliste und die
gezogenen Stellen. **Das hebt den Einwand gegen den Zufall auf**: jeder Lauf
misst dieselben Karten. Und nachrechnen lässt sich der Plan: aus Saat und
Kartenzahlen folgt jede Stelle (`ziehePlan`, PCG aus `math/rand/v2`). Passt
die Kartenliste eines Satzes nicht mehr zum Hash, bricht der Lauf ab — **ein
neuer Plan ist eine neue Kohorte** und wird mit Absicht gezogen:

```bash
MSE_STICHPROBENPLAN="$PWD/docs/belege/kanonstichprobe.json" \
MSE_FREMDPAKETE=… MSE_KANON=… \
go test -tags fremdpakete -run 'TestZieheDenStichprobenplan$' ./internal/app/
```

**Gezogen am 2026-09-22.**

---

## 5 Die Statistik

### Je Schicht

```
  x̄_h  =  (1/n_h) · Σ x_hi

  s²_h =  (1/(n_h − 1)) · Σ (x_hi − x̄_h)²        ← Stichprobenvarianz
```

**Nenner `n−1`, nicht `n`**: die gemessenen Karten sind eine Stichprobe dessen,
was die Vorlagen zeichnen können. Wer durch `n` teilt, unterschätzt die
Streuung der Grundgesamtheit systematisch.

**Quantile nach der Nächste-Rang-Regel, ohne Interpolation**:

```
  Q(p)  =  x_(k)     mit  k = ⌈p · n⌉        (1-basierter Rang)
```

Eine interpolierte Zahl sähe genauer aus, als die Messung ist.

### Über den vollen Kanon

**Nicht durch Zusammenschütten.** Die Sets werden sehr verschieden stark
gezogen — Archenemy zu 91 %, Black zu 15 % —, und im Topf vertritt jede Karte
gleich viel. Der Mittelwert darüber schätzt den Mittelwert einer
Grundgesamtheit, die es nicht gibt.

Mit `N_h` Karten je Schicht, `N = Σ N_h` und `W_h = N_h / N`:

```
  x̄_str        =  Σ  W_h · x̄_h                        ← Kanonmittel

  Var(x̄_str)   =  Σ  W_h² · (s²_h / n_h) · (1 − n_h/N_h)
                                            └─ Endlichkeitskorrektur

  SE           =  √Var(x̄_str)

  95-%-Band    =  x̄_str  ±  1,96 · SE                 (normalapproximiert)
```

**Das Band setzt eine Zufallsauswahl in jeder Schicht voraus** — die
Varianzformel ist die der geschichteten Zufallsstichprobe. Es gilt deshalb für
die Zufallsstichprobe (`zufall.band.*`) und **nicht** für die feste Kohorte
(QA-02.AK-03). **Beschreibend** — Mittel, Median, Quantile, Streuung — sind
beide Kohorten; **modellabhängig** — Standardfehler und Band — ist nur die
gezogene, und nur sie schreibt beides in Protokoll und Beleg
(`TestDieFesteKohorteTraegtKeinBand`).

**Hier stand bis zum Zweitblick über Phase 6.0**, das Band der festen Kohorte
stehe im Protokoll, „weil dieselbe Rechnung läuft". Es stand auch im Beleg,
als `mse2.band.unten` und `mse2.band.oben` — ein Band unter einer Annahme, die
die Ziehung nicht erfüllt, maschinenlesbar neben dem gültigen. **Der Beleg vom
2026-09-23 trägt es noch**; die nächste Messung nicht mehr.

**Die Endlichkeitskorrektur gehört dazu**: eine Schicht, die **ganz** gezogen
ist, trägt keinen Fehler bei — ihr Mittel ist kein Schätzer mehr, sondern die
Wahrheit. Ohne sie stünde dort ein Fehlerband, das es nicht gibt.

**Für Median, Quartile und Perzentile eine gewichtete empirische
Verteilungsfunktion.** Jede gezogene Karte trägt `w_hi = N_h / n_h` — so viele
Karten des Kanons vertritt sie:

```
  Q(p)  =  min { x_(k) : Σ_{i≤k} w_(i)  ≥  p · Σ w_i }
```

Dieselbe Nächste-Rang-Regel wie oben, nur zählt jede Karte so oft, wie sie
vertritt. Bei gleichen Ziehungsanteilen fällt beides zusammen.

**Die gewichtete Streuung der Karten** — nicht die des Schätzers:

```
  s²_w  =  Σ w_i (x_i − x̄_w)²  /  (Σ w_i − 1)
```

> **Zwei Zahlen, zwei Fragen, und sie zu verwechseln ist der häufigste Fehler
> mit einer Stichprobe.** `s²_w` sagt, wie weit die Karten des Kanons
> auseinanderliegen. `Var(x̄_str)` sagt, wie genau ihr **Mittel** geschätzt
> ist. Die erste ist eine Eigenschaft der Karten, die zweite eine der Messung.

### Ein Beispiel, das den Unterschied zeigt

Zwei Schichten, je zehn Karten gezogen:

| Schicht | N | Wert | n |
| --- | ---: | ---: | ---: |
| A | 100 | 10 | 10 |
| B | 1.000 | 20 | 10 |

| | Topf | geschichtet |
| --- | ---: | ---: |
| **Mittel** | 15,00 | **19,09** |
| **Median** | 15 | **20** |

Über vier Punkte im Mittel und fünf im Median, allein durch die Gewichtung.
**Dieser Fall steht als Test** (`kanonstratum_intern_test.go`), damit die
Rechnung nicht wieder wegläuft.

---

## 5a Das Ergebnis vom 2026-09-07

**Der erste Lauf mit der Stichprobe je Schicht und dem geschichteten
Schätzer.** Vor der Freigabe von 0.4.0 erhoben, Laufzeit 24 Minuten.

> **Am 2026-09-20 wiederholt, vor der Freigabe 0.5.0** — Laufzeit 22 Minuten,
> mit einer **neu erzeugten** MSE2-Ausgabe (die alte lag auf diesem Rechner
> nicht mehr; `scripts/mse2-referenz.sh` hat die zehn Sätze in gut vierzehn
> Minuten wieder gerastert, 25.148 Bilder). **Jede Zahl dieses Abschnitts kam
> wieder heraus**, bis auf die vierte Nachkommastelle — Kanonmittel,
> Standardfehler, Band, Perzentile, Zerlegung, und auch die drei Zahlen des
> Topfes daneben.
>
> **Das ist die Probe auf Abschnitt 4**: die Stichprobe wird **ohne Zufall**
> gezogen, und deshalb ist eine Wiederholung eine Wiederholung und kein
> zweiter Wurf. Wäre eine Zahl gewandert, ohne dass jemand die Rasterung
> angefasst hat, wäre das der Fund gewesen.

### Der volle Kanon

| | |
| --- | ---: |
| Grundgesamtheit `N` | **25.148** Karten |
| Stichprobe `n` | **4.265** Karten (16,96 %) |
| Schichten | 10 Sets |
| **Kanonmittel** | **20,0564** |
| Standardfehler | 0,0576 |
| **95-%-Band** | **[19,9435 · 20,1693]** |

| | | | |
| --- | ---: | --- | ---: |
| Median | 19,3579 | Standardabweichung `s_w` | 4,2394 |
| Minimum | 9,0382 | Maximum | 47,0802 |
| Q1 (25 %) | 17,7823 | Q3 (75 %) | 21,3030 |
| IQA | 3,5208 | Streuung `s²_w` | 17,9729 |
| P90 | 23,8173 | P95 | 26,7568 |
| **P99** | **39,3881** | | |

### Was die Schichtung ändert

| | Topf | geschichtet |
| --- | ---: | ---: |
| Mittel | 19,942 | **20,0564** |
| Median | 19,291 | **19,3579** |
| P99 | 39,500 | **39,3881** |
| Standardabweichung | 4,436 | **4,2394** |

**Der Unterschied ist kleiner als das Beispiel in Abschnitt 5 befürchten
lässt** — die sieben großen Sets sind ähnlich groß und ähnlich stark gezogen.
**Die Richtung stimmt trotzdem und ist erklärbar**: die kleinen Sets sind die
**besten** (Planechase 13,1; Archenemy 14,2) und waren im Topf um das Fünf- bis
Sechsfache überrepräsentiert. Ihr Wegfallen hebt das Mittel.

**Das ist keine Beschönigung des Verfahrens.** Wäre eines der kleinen Sets das
schlechteste, ginge die Zahl in die andere Richtung — die Schichtung schätzt,
sie schönt nicht.

### Je Set

| Set | `N_h` | `n_h` | Mittel | Median | Min | Max | Q1 | Q3 | σ |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Planechase | 86 | 77 | **13,119** | 12,841 | 10,232 | 16,370 | 12,144 | 14,351 | 1,592 |
| Archenemy | 70 | 64 | 14,197 | 13,066 | 10,676 | 21,745 | 12,671 | 13,934 | 2,797 |
| Black | 3.788 | 582 | 18,879 | 18,565 | 9,498 | 42,527 | 17,243 | 20,115 | 3,028 |
| Green | 3.689 | 580 | 19,491 | 19,148 | 10,658 | 46,583 | 17,701 | 20,722 | 3,214 |
| Red | 3.730 | 581 | 19,562 | 19,235 | 10,776 | 47,080 | 17,754 | 20,850 | 3,343 |
| Blue | 3.693 | 580 | 19,784 | 19,383 | 11,823 | 45,928 | 17,788 | 21,230 | 3,481 |
| White | 3.761 | 582 | 20,470 | 20,027 | 13,341 | 42,300 | 18,602 | 21,697 | 3,337 |
| Vanguard | 107 | 93 | 20,720 | 18,202 | 12,735 | 40,502 | 16,820 | 21,680 | **6,140** |
| Multicolor | 3.008 | 559 | 21,050 | 20,438 | 11,896 | 43,536 | 18,447 | 22,902 | 3,985 |
| Colorless | 3.216 | 567 | **21,856** | 19,420 | 9,038 | 46,417 | 17,465 | 22,761 | **7,287** |

**Zwei Sets streuen doppelt so weit wie die übrigen** — `Colorless` mit 7,29
und `Vanguard` mit 6,14. Beide sind **gemischte Mengen**: Colorless bringt
Artefakte **und** Länder mit Artefakt-Blendmaske, Vanguard ein Überformat mit
eigenem Stil. Bei beiden klaffen Mittel und Median weit auseinander (2,4 und
2,5 gegen 0,3 bei den einfarbigen) — der Fingerabdruck zweier Zeichenwege in
einem Topf.

**Eine Standardabweichung ist hier also ein Befund über die Menge**, nicht nur
eine Streuungsangabe. Genau dafür steht sie da.

### Wo der Abstand sitzt

| Ebenenart | Anteil am Abstand | eigener Abstand | Flächenanteil |
| --- | ---: | ---: | ---: |
| **Text** | **55,65 %** | 75,174 | **4,05 %** |
| Bild | 40,85 % | 13,419 | 93,40 % |
| Symbolzug | 3,16 % | 64,915 | 0,31 % |
| Fläche | 0,35 % | 8,044 | 2,24 % |

**Der Text trägt 56 Prozent des Abstands auf 4 Prozent der Fläche.** Der
Symbolzug ist noch dichter: 3,2 Prozent auf 0,31 Prozent. Wer dieses Programm
näher an MSE2 bringen will, arbeitet an der **Schriftrasterung**, nicht an den
Bildern.

### Und es ist kein Versatz

| | |
| --- | --- |
| ganze Karte, ±2 Bildpunkte | **0 von 4.265** Karten gewinnen |
| nur der Text, ±2 | 468 Karten gewinnen, im Mittel 6,48 |

**Die ganze Karte sitzt richtig.** Der Text sitzt bei 11 Prozent der Karten um
ein bis zwei Punkte anders — bei 49 davon liegt das Optimum am Rand des
Suchbereichs, dort ist der Gewinn eine **Untergrenze**.

### Je Vorlage

| Vorlage | Karten | Mittel | schlechteste |
| --- | ---: | ---: | ---: |
| m15-aftermath | 6 | 27,614 | 29,627 |
| m15-mainframe-planeswalker | 42 | 24,429 | 38,503 |
| m15-saga | 15 | 22,845 | 27,083 |
| m15-mainframe-dfc | 47 | 21,687 | 37,067 |
| m15-vanguard | 93 | 20,720 | 40,502 |
| **m15-altered** | **3.899** | **20,064** | 47,080 |
| m15-adventure | 7 | 17,540 | 19,973 |
| m15-split-fusable | 12 | 17,074 | 20,112 |
| archenemy | 64 | 14,197 | 21,745 |
| m15-mainframe-planes | 77 | 13,119 | 16,370 |
| (seltenere zusammen) | 3 | 32,104 | — |

**`m15-altered` trägt 91 Prozent der Stichprobe** — das ist die gewöhnliche
M15-Karte, an der sich entscheidet, ob dieses Programm eine echte Magic-Karte
zeichnet. Ihr Mittel von 20,06 ist praktisch das Kanonmittel.

**Die Zeilen mit sechs oder sieben Karten sagen wenig.** Sie stehen hier, weil
eine fehlende Zeile schlimmer wäre als eine dünne — aber ein Mittel über sechs
Karten ist kein Maß, sondern ein Hinweis.

## 5b Das Ergebnis vom 2026-09-23 — nach QA-03

**Die Rasterung hat sich geändert**, zum ersten Mal seit 0.4.0: Kerning mit
Klassenpaaren, MSEs vierfacher Textweg, Ligaturen und die nachgebildete
Glättung von Quartz (#1656,
#1805,
#1806). Die
Zahlen stammen aus dem Beleg am Stand `7972c05` — PASS, 77 von
86 Fällen gemessen —, gegen die Referenz, die MSE2 am 2026-09-23 neu gezeichnet
hat. **Abschnitt 5a bleibt stehen**: er ist der Stand, an dem 0.4.0 und 0.5.0
gemessen wurden. Die Schritte dazwischen stehen in
[`SCHRIFTUNTERSUCHUNG.md`](SCHRIFTUNTERSUCHUNG.md).

### Der volle Kanon

| | feste Kohorte | Zufallsstichprobe |
| --- | ---: | ---: |
| Karten | **4.266** | **4.582** |
| **Kanonmittel** | **13,7922** | **13,7611** |
| Standardfehler | — | 0,0710 |
| **95-%-Band** | — | **[13,6219 · 13,9003]** |

| | | | |
| --- | ---: | --- | ---: |
| Median | 12,4389 | Standardabweichung `s_w` | 5,4245 |
| Minimum | 2,5578 | Maximum | 49,1265 |
| Q1 (25 %) | 10,4660 | Q3 (75 %) | 15,4356 |
| IQA | 4,9696 | Streuung `s²_w` | 29,4252 |
| P90 | 18,5887 | P95 | 23,0396 |
| **P99** | **39,4291** | | |

**Das Band gehört zur Zufallsstichprobe** (QA-02), die feste Kohorte ist der
Verlauf. Beide Mittel liegen 0,03 auseinander.

### Was die Schichtung ändert

| | Topf | geschichtet |
| --- | ---: | ---: |
| Mittel | 13,758 | **13,7922** |
| Median | 12,353 | **12,4389** |
| P99 | 39,516 | **39,4291** |
| Standardabweichung | 5,612 | **5,4245** |

### Je Set

| Set | `N_h` | `n_h` | Mittel | Median | Min | Max | Q1 | Q3 | σ |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Planechase | 86 | 77 | 10,312 | 10,178 | 8,237 | 12,643 | 9,734 | 10,957 | 0,911 |
| Vanguard | 107 | 93 | 11,846 | 6,977 | 2,558 | 39,725 | 4,164 | 20,082 | 10,598 |
| Archenemy | 70 | 64 | 12,099 | 10,318 | 8,888 | 22,443 | 9,932 | 11,156 | 3,879 |
| Black | 3.788 | 582 | 12,677 | 11,761 | 7,359 | 44,181 | 10,186 | 14,224 | 3,978 |
| Green | 3.689 | 580 | 12,933 | 12,032 | 7,635 | 48,417 | 10,476 | 14,438 | 4,206 |
| Red | 3.730 | 581 | 13,253 | 12,353 | 7,476 | 49,126 | 10,408 | 14,960 | 4,366 |
| Blue | 3.693 | 580 | 13,285 | 12,293 | 7,443 | 46,415 | 10,328 | 14,887 | 4,511 |
| White | 3.761 | 582 | 13,303 | 12,241 | 7,836 | 43,000 | 10,579 | 15,033 | 4,312 |
| Multicolor | 3.008 | 560 | 14,747 | 13,921 | 7,829 | 44,943 | 10,873 | 16,658 | 5,348 |
| Colorless | 3.216 | 567 | 17,172 | 14,319 | 5,280 | 46,907 | 11,395 | 18,306 | 8,803 |

**Vanguard streut am weitesten** — 10,6, bei einem Median von 7,0: die
Überformatkarten liegen nahe an MSE2, die Karten mit viel Text weit weg.

### Wo der Abstand sitzt

| Ebenenart | Anteil am Abstand | eigener Abstand | Flächenanteil |
| --- | ---: | ---: | ---: |
| Bild | 51,58 % | 11,02 | 93,01 % |
| Text | 41,98 % | 45,46 | 4,45 % |
| Symbolzug | 5,87 % | 64,53 | 0,31 % |
| Fläche | 0,58 % | 7,51 | 2,23 % |

**Der Text trägt 42 Prozent des Abstands auf 4,5 Prozent der Fläche** — vor
QA-03 waren es 56 Prozent auf 4,05. Die Fläche wuchs, weil geglätteter Text
mehr Bildpunkte deckt.

### Und es ist kein Versatz

| | |
| --- | --- |
| ganze Karte, ±2 Bildpunkte | **0 von 4.266** Karten gewinnen |
| nur der Text, ±2 | 93 Karten gewinnen, im Mittel 10,64 |

**Weniger Karten als vor QA-03** (468), aber wo der Text daneben sitzt, kostet
er mehr — kräftigerer Text an falscher Stelle. Bei 47 liegt das Optimum am
Rand des Suchbereichs; dort ist der Gewinn eine Untergrenze.

### Je Vorlage

| Vorlage | Karten | Mittel | schlechteste |
| --- | ---: | ---: | ---: |
| m15-aftermath | 6 | 24,462 | 26,556 |
| m15-mainframe-planeswalker | 42 | 21,337 | 38,370 |
| m15-mainframe-dfc | 47 | 16,170 | 34,393 |
| m15-altered | 3.900 | 13,767 | 49,126 |
| m15-saga | 15 | 13,442 | 22,042 |
| m15-adventure | 7 | 12,800 | 15,984 |
| archenemy | 64 | 12,099 | 22,443 |
| m15-vanguard | 93 | 11,846 | 39,725 |
| m15-split-fusable | 12 | 11,805 | 14,859 |
| m15-mainframe-planes | 77 | 10,312 | 12,643 |
| (seltenere zusammen) | 3 | 30,993 | — |

---

## 5c Der Stand der Freigabe 0.6.0 — 2026-10-04 und 2026-10-05

**Gemessen am 2026-10-04 am Stand `3bae5a6`** — PASS, 79 von 88 Fällen,
sauberer Baum — **und am 2026-10-05 am Freigabestand `8868908` nachgemessen**,
weil `make belege` den ersten Beleg als „anderer Stand" meldete: der Lauf misst
mehr als die Rasterung, und `internal/app` hatte sich bewegt. **Jede Kennzahl
kam gleich heraus.** Gegen dieselbe Referenz wie 5b; an der Rasterung hat sich
zwischen beiden nichts geändert.

**Der erste Anlauf am selben Abend war rot, und er hatte recht.** Das
Kanonmittel stand bei **28,5** über dem Deckel von 20,4, zu 84 bis 88
Prozent im Bild: seit #2040
schrieb eine leere Mehrfachwahl ihren Leernamen als Wert — `colorless` an
`card color` —, bevor die Vorgabeskripte liefen, die nur leere Felder
rechnen. Jede Karte des Kanons trug damit den farblosen Rahmen. MSE2 schreibt
den Namen so, dass der Wert Vorgabe bleibt (`multiple_choice.cpp`, Fassung
`f9d9356` vom 2021-07-06); `3bae5a6` tut es seitdem auch. **Kein anderer Lauf
der Pipeline hat das gesehen** — die Goldbilder prüfen gegen sich selbst.

| | feste Kohorte | Zufallsstichprobe |
| --- | ---: | ---: |
| Karten | **4.266** | **4.582** |
| **Kanonmittel** | **13,7923** | **13,7554** |
| Standardfehler | — | 0,0709 |
| **95-%-Band** | — | **[13,6165 · 13,8943]** |
| Median | 12,4389 | 12,4558 |
| Q1 · Q3 | 10,4660 · 15,4419 | 10,4496 · 15,4178 |
| Minimum · Maximum | 2,5578 · 49,1265 | 2,5578 · 73,1301 |
| P99 | 39,4291 | 38,9366 |
| Befunde `Portierung` · `Vorlage` · `Umgebung` | 0 · 0 · 0 (Befundlauf, 680 Karten) | 0 · 3 · 0 |

**Die Zerlegung nach Ebenenart und die Versatzprobe kamen gleich heraus** wie
in 5b — Bild 51,58 %, Text 41,98 %, Symbolzug 5,87 %, Fläche 0,58 %; 0 von
4.266 Karten gewinnen durch einen Versatz der ganzen Karte, 93 durch einen des
Textes. Nachgelesen im Rohprotokoll des Belegs
(`.laufbeleg/fremdpakete.jsonl`, dieselbe Prüfsumme wie im Beleg). Die Tafeln
je Set und je Vorlage im `README.md` stammen aus diesem Lauf.

## 6 Was die Zahlen nicht sagen

- **Nichts über Wahrnehmung.** Der Abstand ist ein Pixelmaß. Ob ein Mensch zwei
  Karten unterscheidet, beantwortet er nicht — das tut der Blick.
- **Nichts über Vorlagen, die in keiner Schicht vorkommen.** Zehn Sets sind
  zehn Sets.
- **Kein Vertrauensband um den Median.** Für Quantile aus einer geschichteten
  Stichprobe gibt es keinen geschlossenen Ausdruck ohne weitere Annahmen; ein
  Bootstrap wäre der Weg, und er ist nicht gebaut.
- **Nichts über die Richtigkeit der Referenz.** Siehe Abschnitt 2.
- **Und nichts über eine zweite Maschine.** Gemessen wird lokal
  ([0007](entscheidungen/0007-kein-entfernter-lauf.md)).
- **Nichts über Geschwindigkeit.** Wie lange ein Programm für dieselbe Aufgabe
  braucht, misst [`LEISTUNGSVERGLEICH.md`](LEISTUNGSVERGLEICH.md) (QA-04) —
  die Laufzeit dieses Laufs ist dafür kein Maß, denn er rastert, vergleicht
  und zerlegt.

---

## 7 Wie der Lauf gefahren wird

```bash
source scripts/activate.sh

MSE_FREMDPAKETE=/…/full-magic-pack/data \
MSE_SCHRIFTWURZEL=/…/full-magic-pack \
MSE_KANON="$HOME/Library/Application Support/MagicSetEditor4Mac/kanonische-sets" \
MSE_MSE2AUSGABE="$HOME/Library/Application Support/MagicSetEditor4Mac/mse2-ausgabe" \
go test -tags fremdpakete -timeout 7200s ./internal/app/ \
  -run TestDerAbstandZuMSE2AmKanon -v
```

**Alle vier Angaben sind nötig.** `TestMain` gibt dem Lauf ein eigenes
`HOME` — sonst schriebe ein Test in den echten Bestand —, und die Vorgaben
suchen dort. Fehlt eine, **überspringt** der Lauf und meldet `PASS`:

```
    keine MSE2-Ausgabe — scripts/mse2-referenz.sh Archenemy White
--- SKIP: TestDerAbstandZuMSE2AmKanon
```

**Ein `SKIP` ist grün.** Wer nur auf den Rückgabewert sieht, meldet eine
Messung, die nie stattgefunden hat. Am 2026-09-07 ist genau das passiert, und
aufgefallen ist es nur, weil die Ausgabe leer war statt voller Zahlen.

**Seit QA-01 (#1733)
ist der gewöhnliche Weg `make fremdpakete`.** Er setzt alle neun Angaben —
die vier hier, dazu `MSE_SETPAKET`, `MSE_REFERENZBILDER`, `MSE2_QUELLE`,
`MSE2_BINAER` und `MSE2_DATEN` für die Läufe daneben —, und **fehlt eine, wird nicht gemessen**: der Beleg
`docs/belege/fremdpakete.json` sagt dann BLOCKED und nennt, welche. Die
Kennzahlen dieses Laufs stehen im Beleg als `mse2.kanonmittel`,
`mse2.median` und so fort — ein Band nur unter `zufall.` —; wer die Zahlen im
README ändert, ändert sie aus dem Beleg. **Bis dahin setzte `make fremdpakete` zwei der vier nicht**, und
dieser Lauf sprang über das Ziel in jedem Lauf — die Zahlen im README kamen aus
Läufen von Hand ([`PIPELINE.md`](PIPELINE.md)).

`MSE_STICHPROBE` hebt die Schranke auf einen festen Wert oder auf `alle` — der
Lauf misst dann jede Karte und dauert Stunden. **Er trägt dann keine Aussage
über 99,9 Prozent mehr, sondern die volle Wahrheit über die Sets**, und die
Endlichkeitskorrektur macht jeden Standardfehler zu null.

**Und der Beleg hält fest, womit er gemessen wurde**: `MSE_STICHPROBE` steht
im Plan als **Schalter**, sein Wert — leer, wo er nicht gesetzt war — im
Beleg, und `make belege` nennt einen Beleg mit einem anderen Wert »anderer
Stand«. Bis zum Nachlauf des Zweitblicks über Phase 6.0 hieß ein Lauf mit
`alle` und eine Prüfung ohne den Schalter »aktuell«. Dasselbe gilt für die
Go-Fassung und die Plattform.

### Karte für Karte gegen einen früheren Lauf

**Mittel, Median und P99 sagen nicht, ob eine Karte schlechter wurde** —
tausend gebesserte und zehn verschlechterte senken alle drei. Seit F02
(#2244)
schreibt der Lauf deshalb auf Wunsch jede gezogene Karte mit Set, Stelle in
der Stichprobe, Name und Abstand, und hält sie gegen einen früheren Lauf:

```bash
MSE_KANON_JEKARTE=vorher.tsv  … go test … -run TestDerAbstandZuMSE2AmKanon   # Stand A
MSE_KANON_JEKARTE=nachher.tsv MSE_KANON_VORHER=vorher.tsv  … go test …     # Stand B
```

Eine Karte, die um mehr als **0,01** schlechter wird, macht den zweiten Lauf
**rot** und steht mit beiden Werten im Protokoll, die zwanzig schlimmsten
zuerst. **Die Toleranz ist klein, weil der Lauf sich wiederholt** (§5a).
**Und der Vergleich verlangt dieselbe Stichprobe**: eine Karte, die nur einer
der beiden Läufe kennt — ein anderer Stichprobenplan oder ein Ausfall —,
macht ihn ebenso rot wie keine einzige vergleichbare Karte, und das
Protokoll nennt die Prüfsumme der früheren Datei. Bis zum Zweitblick über
E/F wurden solche Karten nur gezählt; eine leere frühere Datei bestand
(Befund 2).

`MSE_KANON_VORHER` steht im Plan als **Schalter** (er kann den Lauf rot
machen), `MSE_KANON_JEKARTE` als Ablageort ohne Wirkung. **Was er nicht
hält**: ob die frühere Datei vom selben Plan, derselben MSE2-Ausgabe und
denselben Schriften stammt — das steht im Beleg des früheren Laufs, nicht in
der Datei.

Die Referenzbilder erzeugt `scripts/mse2-referenz.sh`; sie brauchen ein
gebautes MSE2 und die Schriften des Packs unter `~/Library/Fonts/mse2-referenz`.

---

## 8 Wo die Rechnung steht

| Was | Wo |
| --- | --- |
| Das Maß und die Zerlegung | `internal/render/zerlegung.go` |
| Die Stichprobengrößen | `internal/app/stichprobe_test.go` |
| Verteilung je Schicht | `internal/app/kanonverteilung_intern_test.go` |
| Der geschichtete Schätzer | `internal/app/kanonstratum_intern_test.go` |
| Der Lauf selbst | `internal/app/mse2kanon_intern_test.go` |

**Jede Rechnung wird geprüft**, und zwar ohne Build-Tag, damit sie in der
Pipeline läuft: die Verteilung an einer Reihe von 1 bis 10, deren Kennzahlen
von Hand nachzurechnen sind, und der geschichtete Schätzer an dem Beispiel aus
Abschnitt 5.
