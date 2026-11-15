# Die Sammlung bei fünftausend Karten

Was diese Datei beantwortet: **was kostet eine Sammlung, die groß genug ist,
dass niemand sie mehr im Kopf hat** — und wo die Zahlen herkommen.

Sie gehört zu Phase 4.5
(#1292). Die
Abnahme dort verlangt ausdrücklich **eine benannte Schranke für Aufbau und
Suche, gemessen an einem Bestand dieser Größe, nicht an zwanzig Zeilen**.

---

## Woran gemessen wird

**Nicht am Bestand dieses Rechners.** Der stand am 2026-08-30 auf **null**
Exemplaren; `mse sammlung --json` gab `{"exemplare": []}`, und jeder Lauf
darüber kam in unter 10 ms zurück. Eine Zusage daraus wäre eine Aussage über
eine leere Tabelle gewesen — die Reichweite einer Messung ist die Reichweite
ihrer Aussage (§10).

Gemessen wird an **5.000 selbst gelegten Exemplaren**, in einem **eigenen
Zuhause**:

```bash
node scripts/sammlungssaat.mjs          # legt /tmp/mse45heim samt Bestand an
HOME=/tmp/mse45heim make dev-browser    # dieselbe Anwendung unter :34115
node scripts/sammlungsmass.mjs
```

**Seit Phase 5.1 legt die Saat auch ein Set mit fünf Karten** und nimmt
Zuhause und Anzahl als Argumente — `node scripts/sammlungssaat.mjs
/tmp/mse51heim 10000` ist der Messbestand aus
[`UX-ARCHITEKTUR.md`](UX-ARCHITEKTUR.md). Und jedes Exemplar trägt seit
Phase 4.12 ein Profil: die Saat liest die Kennung des ersten Profils aus dem
Bestand, statt sie zu raten, und weist ab, wenn weniger Zeilen liegen als
bestellt.

Verlegt wird `HOME` **und** der Bestand darunter. Hier stand, das Fenster
nehme `MSE_BESTAND` bewusst nicht entgegen; das gilt seit dem 2026-09-14 nicht
mehr, und seit QA-10.NAT-01
(#1741)
liegen auch Pakete und Bilder neben dem Bestand, der offen ist
(`internal/app/bestandsort.go`).

Der Go-Weg misst sich ohne Fenster:

```bash
go test -tags fremdpakete -run '^$' -bench BenchmarkWegSammlung ./internal/app/
```

---

## Die Zahlen

Gemessen am **2026-08-30** auf einem **Apple M5**, bei 1280 × 800, gegen 5.000
Exemplare **ohne einen einzigen Preis**. Je drei Läufe.

| Weg | gemessen | Schranke | wo sie steht |
| --- | ---: | ---: | --- |
| Die ganze Sammlung laden (Go) | 4,46–4,56 ms | 7 ms | `messsammlung_bench_intern_test.go` |
| Die Exemplare eines Drucks (Go) | 17,1–17,3 µs | 100 µs | dito |
| Die Sammlung zusammenrechnen (Go) | 4,26–4,29 ms | 7 ms | dito |
| Die Wails-Naht (1,5 MB JSON) | 167 ms | — | gemessen, nicht gehalten |
| Die Seite aufbauen (Fenster) | 321–358 ms | 600 ms | `scripts/sammlungsmass.mjs` |
| Suche, vier Tastendrücke (Fenster) | 67–69 ms | 120 ms | dito |

**Jede Schranke gilt für diese Maschine**
([0007](entscheidungen/0007-kein-entfernter-lauf.md)). Sie ist ein Wächter
gegen **Rückschritt**, keine Zusage an Nutzende.

---

## Zwei Befunde, die die Messung gebracht hat

### Das Portfolio fragte einmal je Exemplar

Der erste Lauf gab **89 ms** für das Zusammenrechnen gegen **5 ms** für das
Laden derselben Zeilen — und die Preistabelle war dabei **leer**. `Portfolio`
rief `Neueste` in der Schleife über die Exemplare: 5.000 Abfragen, von denen
keine etwas fand.

Behoben mit `Preisspeicher.NeuesteAlle` — eine Abfrage, dann eine Zuordnung je
Druck in Go. Nachgemessen 4,3 ms, **Faktor 20**.

**Gehalten wird die Zahl der Abfragen, nicht die Zeit.** Eine
Millisekundenschwelle beschreibt diese Maschine;
`service.TestPortfolioFragtNichtJeExemplar` gilt auf jeder. Er fährt **zwei**
Bestandsgrößen mit demselben Zähler — eine feste Zahl ließe sich mit einer
Konstante erfüllen, die zufällig passt.

### Die Seite zeichnete fünftausend Zeilen

Ohne Grenze: **680–791 ms** Aufbau und **145–150 ms** je Suche. Mit der Grenze
aus `frontend/src/seiten/sammlungssicht.ts`: **321–358 ms** und **67–69 ms**.

**Die Grenze liegt bei 200 Zeilen**, und sie ist sichtbar: unter der Liste
steht „200 von 500 gezeichnet" und daneben ein Knopf. Eine still gekappte Liste
sähe aus wie eine vollständige.

**Warum eine Grenze und keine Fensterung**: weil niemand fünftausend Zeilen
liest. Wer so viel besitzt, sucht; die Liste ist die Antwort auf eine Suche und
kein Katalog zum Durchblättern. Eine Fensterung gäbe dieselbe Geschwindigkeit
und behielte einen Bildlauf, den niemand benutzt — zum Preis von Zeilenhöhen,
die sich messen lassen müssten, obwohl eine Zeile sich aufklappt.

**Eine neue Sicht setzt die Grenze zurück** (§7, Regel 1). Sonst behielte eine
Suche mit viertausend Treffern die aufgeklappte Zahl der vorigen.

---

## Nach der ganzen Phase noch einmal gemessen

**Am 2026-08-31**, nachdem alle sechs Werkstücke gebaut waren — dieselbe
Maschine, derselbe Bestand aus 5.000 Exemplaren:

| Weg | vorher | nachher |
| --- | ---: | ---: |
| Die ganze Sammlung laden (Go) | 4,46–4,56 ms | 4,89 ms |
| Die Sammlung zusammenrechnen (Go) | 4,26–4,29 ms | 4,59 ms |
| Die Seite aufbauen (Fenster) | 321–358 ms | 281 ms |
| Suche, vier Tastendrücke (Fenster) | 67–69 ms | 66 ms |

**Die Seite ist schneller geworden, obwohl drei Kästen dazugekommen sind** —
sie laden nichts, bis jemand sie fragt. Der Go-Weg ist um 0,3 ms langsamer:
das sind die beiden Spalten aus [0057](entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md),
die bei jeder Zeile mitgelesen werden.

### Und die Naht dazu, durch die laufende Brücke

Alle elf neuen Bindungen antworten (`window.go.app.Bindungen.*`, 153
insgesamt). Nachgesehen wurde, was **zwischen** den Prüfebenen liegt:

  - jede Liste kommt als `[]` heraus und nicht als `null`,
  - das CSV trägt alle vierzehn Spalten samt `lagerort` und `wunsch`,
  - 5.002 Zeilen für 5.000 Exemplare (Kopfzeile und Abschluss),
  - 3.750 Dubletten bei `behalten = 1` — die Mengen laufen 1, 2, 3, 4, also
    hat drei Viertel des Bestands mehr als eines.

### Und nach dem Nachlauf noch einmal — mit der Streuung dazu

**Am 2026-08-31**, nachdem die elf Befunde des Nachlaufs behoben und die
Sammelanfrage aus #1348
gebaut waren. **Vier Läufe hintereinander**, dieselbe Maschine, derselbe
Bestand:

| Messung | Läufe |
| --- | --- |
| Aufbau der Seite | 329, 372, 199, 191 ms |
| Suche, vier Tastendrücke | 67, 67, 66, 84 ms |

**Die Streuung steht hier, weil eine einzelne Zahl mehr behauptet hätte, als
das Werkzeug hergibt.** Der Aufbau liegt zwischen 191 und 372 ms — der erste
Lauf nach dem Start ist der teuerste, danach fällt es. Die 281 ms der Tabelle
darüber sind **ein** Lauf und liegen mitten darin; sie sind kein Beleg dafür,
dass es schneller geworden wäre, und die Zeile darüber behauptet es auch
nicht. Beide Schranken (600 ms und 120 ms) hält jeder Lauf.

### Die Gegenprobe, die der Nachlauf verlangt hat

Befund 8 des Nachlaufs stand als **nicht geprüft** da: der Wächter liest seit
dem 2026-08-30 die tatsächliche Länge aus der laufenden Bindung, aber die
Gegenprobe dazu war nur behauptet. Am 2026-08-31 ist sie gefahren, und zwar
beide Hälften:

    node scripts/sammlungsmass.mjs http://localhost:34115 4999
    → im Bestand liegen 5000 Exemplare, gemessen wird für 4999

    node scripts/sammlungsmass.mjs http://localhost:34115 5000 199
    → 200 Zeilen im DOM, erwartet 199 — die Zeichengrenze greift nicht

**Beide Male rot, und zwar aus dem richtigen Grund.** Damit steht fest, dass
der Wächter misst, was er zu messen behauptet: die Zahl im Bestand **und** die
Zahl auf dem Bildschirm — und nicht bloß, dass irgendetwas dasteht.

---

## Was hier **nicht** steht

- **Ob 5.000 die richtige Zahl ist.** Sie ist die aus #1292. Wer 50.000
  besitzt, misst neu — und die 167 ms der Naht werden dann die teure Stelle.
- **Ob die Seite gut aussieht.** Dafür gibt es die Sichtprüfung (§3) und
  Phase 4.10 (#1325).
- **Was die Naht kostet, wenn jemand sie hält.** 167 ms für 1,5 MB stehen
  gemessen da und ohne Schranke: sie zu drücken hieße, die Antwort zu
  verkleinern (eine Seite je Ruf), und das ist eine Produktentscheidung und
  kein Rückschritt, den ein Wächter fangen soll.
- **Ein zweiter Rechner.** Gemessen ist eine Maschine, keine Plattform
  ([0007](entscheidungen/0007-kein-entfernter-lauf.md)).
