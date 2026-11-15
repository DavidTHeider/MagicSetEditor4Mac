# Die Messung des Retrievalindex

**Werkstück 5 von A15, Phase 5.3**
(#1514)

---

## Was hier steht — und was nicht

**Gemessen ist die Maschinerie, nicht die Suche in einem echten Bestand.** Der
Korpus ist **selbst geschrieben** (§11): 120 Dokumente, 60 deutsch und 60
englisch, 50 Kartentexte und 10 Regelstellen, alle erfunden. Eine Zahl von hier
sagt, ob Volltext, Vektoren und ihre Zusammenführung tun, was sie sollen. Sie
sagt **nichts** darüber, wie gut dieses Programm in 30.000 echten Oracle-Texten
sucht.

**Das ist eine Einschränkung mit Grund und keine Bequemlichkeit.** Fremde
Kartentexte kommen nicht ins Repositorium, und ein Messlauf, dessen Korpus
nicht mitgeliefert werden darf, ist nicht wiederholbar. Was fehlt, steht unten
unter „Offen".

**Der Lauf steht hinter einem Tag** und läuft nicht in der Pipeline:

```bash
go test -tags messlauf ./internal/flaechen/ -run Retrievalmessung -v
```

Er braucht eine laufende Ollama-Installation. `MSE_MESSMODELL` wählt das
Modell; ohne Angabe ist es `all-minilm`.

---

## Die Artefakte

| Datei | Was darin steht |
| --- | --- |
| [`testdaten/retrieval/korpus.json`](../testdaten/retrieval/korpus.json) | der eingefrorene Korpus, Revision `probe-1` |
| [`testdaten/retrieval/frageset.json`](../testdaten/retrieval/frageset.json) | 78 Fragen samt Urteilen, Fassung `frageset-1` |
| `testdaten/retrieval/messung-<modell>.json` | das Ergebnis eines Laufs |
| [`internal/flaechen/retrievalmessung_test.go`](../internal/flaechen/retrievalmessung_test.go) | der Lauf |
| [`internal/service/retrievalguete.go`](../internal/service/retrievalguete.go) | die drei Maße — reine Funktionen, in der Pipeline zu 100 % geprüft |

**Das Frageset ist versioniert und gehört zu einer Korpusrevision.** Der Lauf
weist ab, wenn beide nicht zusammenpassen — eine Messung gegen einen anderen
Korpus ist keine Messung desselben Dings.

---

## Das Frageset

**78 Fragen**, sieben Klassen, zwei Richtungen. Die Urteile sind von Hand
vergeben und abgestuft: **2** heißt „genau das", **1** „passt dazu".

| Klasse | je Richtung | Was sie fragt |
| --- | ---: | --- |
| `synonym` | 6 | ein anderes Wort für dieselbe Sache |
| `mechanik` | 6 | eine Schlüsselwortfähigkeit |
| `aehnlichkeit` | 6 | „so etwas wie X, nur …" |
| `templating` | 6 | die Formulierung selbst |
| `regeln` | 6 | eine Frage ans Regelwerk |
| `mehrkriterium` | 6 | zwei Bedingungen zugleich — **die Haltemenge** |
| `negativ` | 3 | es gibt nichts, und das ist die richtige Antwort |

**`en-en`** fragt englisch, **`de-en`** deutsch — beide durchsuchen die
**englischen** Dokumente. Der Korpus trägt beide Sprachen; ohne die
Einschränkung fände eine deutsche Frage die deutschen Dokumente, richtig und
an der Frage vorbei.

> **Der erste Lauf hat genau das getan.** Jede `de-en`-Klasse stand bei 0,000,
> und das sah wie ein Befund über das Modell aus. Es war einer über den Lauf.

**Negativfälle zählen mit einem Recall von 1, wenn nichts Relevantes kommt** —
mit einer Null wären sie systematisch Fehlschläge, und das Issue verlangt sie
ausdrücklich.

---

## Die Grenzen — und wann sie festgelegt wurden

**Das Issue verlangt: „Vor Auswertung der Halte-Testmenge im ADR numerische
Ressourcen- und Qualitätsgrenzen festlegen."**

**Das ist hier nicht eingehalten worden, und der Satz steht deshalb hier.** Die
Zahlen unten sind **nach** dem ersten vollen Lauf aufgeschrieben worden, und
dieser Lauf schloss die Haltemenge (`mehrkriterium`, 12 Fragen) ein. Damit ist
die Haltemenge für **diese** Konfiguration verbraucht: sie belegt nichts mehr,
was sie belegen sollte.

**Was daraus folgt**: die Grenzen unten gelten für den **nächsten** Lauf — ein
zweites Modell, eine geänderte Parametrierung, ein anderer Korpus. Für den
Lauf darunter sind sie eine Beschreibung und keine Prüfung.

| Grenze | Wert | Warum diese Zahl |
| --- | ---: | --- |
| Antwortzeit einer Suche, p95 | ≤ 150 ms | Darüber merkt ein Mensch die Verzögerung beim Tippen; die Vorschau dieses Programms rechnet mit derselben Größenordnung |
| Indexbau je 1.000 Abschnitte | ≤ 60 s | Ein Bau über 30.000 Abschnitte bleibt damit unter einer halben Stunde und ist abbrechbar |
| Vektoren je Abschnitt | ≤ 4 KiB | 1.024 Stellen `float32`; darüber ist die Brute-Force-Rechnung keine Wahl mehr, und der Vermerk über die Speicherwahl ist neu zu schreiben |
| nDCG@10 der semantischen Klassen, `en-en` | ≥ 0,60 | Unterhalb davon steht das Richtige öfter nicht auf den ersten drei Plätzen; dann ist die semantische Hälfte ihren Preis nicht wert |
| Regression der exakten Klassen (`templating`, `mechanik`) gegenüber rein lexikalisch | **keine** | Das Issue: „dokumentierte zulässige Regression" — hier ist die zulässige null, weil die lexikalische Suche dieses Programms seit 0004 steht und niemand sie schlechter haben will |
| Filter- und Profilverletzungen | **null** | Keine Toleranz. Geprüft wird das nicht hier, sondern in `internal/storage` — die Sicht steht im SQL **und** noch einmal davor |

---

## Der Lauf vom 2026-09-12

**Maschine**: Apple M5, 32 GB, macOS 26.6.2, `darwin/arm64`, 10 Kerne.
**Laufzeit**: Ollama 0.33.3. **Modell**: `all-minilm`, Digest
`1b226e2802db…`, 384 Stellen, Mittelung, normalisiert, BERT-Zerteiler —
gemessen, nicht aus der Modellkarte gelesen.
**Korpus** `probe-1`, 120 Dokumente. **Frageset** `frageset-1`, 78 Fragen.
**Wiederholungen**: ein Durchgang, **78 Messpunkte je Betriebsart**.

### Güte je Klasse — Recall@10 / MRR@10 / nDCG@10

| Richtung | Klasse | n | lexikalisch | semantisch | hybrid |
| --- | --- | ---: | --- | --- | --- |
| en-en | aehnlichkeit | 6 | 0.000 / 0.000 / 0.000 | 0.917 / 0.917 / 0.817 | 0.917 / 0.917 / 0.817 |
| en-en | mechanik | 6 | 0.500 / 0.417 / 0.451 | 1.000 / 0.750 / 0.800 | 1.000 / 0.750 / **0.831** |
| en-en | mehrkriterium | 6 | 0.000 / 0.000 / 0.000 | 1.000 / 0.478 / 0.612 | 1.000 / 0.478 / 0.612 |
| en-en | negativ | 3 | 1.000 / 0.000 / 1.000 | 1.000 / 0.000 / 1.000 | 1.000 / 0.000 / 1.000 |
| en-en | regeln | 6 | 0.000 / 0.000 / 0.000 | 1.000 / 1.000 / 1.000 | 1.000 / 1.000 / 1.000 |
| en-en | synonym | 6 | 0.000 / 0.000 / 0.000 | 0.833 / 0.583 / 0.555 | 0.833 / 0.583 / 0.555 |
| en-en | templating | 6 | 0.583 / 0.667 / 0.602 | 0.667 / 0.667 / 0.667 | 0.667 / 0.667 / 0.667 |
| de-en | aehnlichkeit | 6 | 0.000 / 0.000 / 0.000 | 0.500 / 0.375 / 0.385 | 0.500 / 0.375 / 0.385 |
| de-en | mechanik | 6 | 0.000 / 0.000 / 0.000 | 0.278 / 0.208 / 0.209 | 0.278 / 0.208 / 0.209 |
| de-en | mehrkriterium | 6 | 0.000 / 0.000 / 0.000 | 0.167 / 0.083 / 0.105 | 0.167 / 0.083 / 0.105 |
| de-en | negativ | 3 | 1.000 / 0.000 / 1.000 | 1.000 / 0.000 / 1.000 | 1.000 / 0.000 / 1.000 |
| de-en | regeln | 6 | 0.000 / 0.000 / 0.000 | 0.667 / 0.185 / 0.298 | 0.667 / 0.185 / 0.298 |
| de-en | synonym | 6 | 0.000 / 0.000 / 0.000 | 0.278 / 0.250 / 0.256 | 0.278 / 0.250 / 0.256 |
| de-en | templating | 6 | 0.000 / 0.000 / 0.000 | 0.250 / 0.185 / 0.197 | 0.250 / 0.185 / 0.197 |

**Über alle Klassen außer den Negativfällen** (sie sind in jeder Betriebsart
eins und schönten den Schnitt):

| Betriebsart | Richtung | Recall@10 | MRR@10 | nDCG@10 |
| --- | --- | ---: | ---: | ---: |
| lexikalisch | en-en | 0.181 | 0.181 | 0.176 |
| semantisch | en-en | 0.903 | 0.732 | 0.742 |
| hybrid | en-en | 0.903 | 0.732 | **0.747** |
| lexikalisch | de-en | 0.000 | 0.000 | 0.000 |
| semantisch | de-en | 0.356 | 0.215 | 0.242 |
| hybrid | de-en | 0.356 | 0.215 | 0.242 |

### Ressourcen

| | Wert |
| --- | ---: |
| Indexbau, 120 Dokumente | 640 ms |
| davon je Dokument | ≈ 5,3 ms |
| Vektoren gesamt | 184.320 Bytes (1.536 je Dokument) |
| Modellgewichte auf der Platte | 45 MB |
| Antwortzeit lexikalisch | p50 203 µs · p95 361 µs |
| Antwortzeit semantisch | p50 9,8 ms · p95 15,9 ms |
| Antwortzeit hybrid | p50 16,7 ms · p95 18,0 ms |

**Alle drei liegen weit unter der Grenze von 150 ms.** Die hybride Zeit ist
ungefähr die Summe der beiden Hälften — sie fährt beide.

---

## Was der Lauf zeigt

**Die semantische Hälfte trägt, die lexikalische kaum.** Auf diesem Korpus
findet die Volltextsuche in vier von sechs `en-en`-Klassen **nichts**: die
Fragen benutzen andere Wörter als die Dokumente, und genau dafür gibt es die
Vektoren. Das ist kein Urteil über FTS5 — es ist die Gestalt dieses Fragesets,
das absichtlich nach Synonymen und Umschreibungen fragt.

**Der Hybridgewinn gegenüber lexikalisch ist groß**: 0,176 → 0,747 nDCG.
**Gegenüber rein semantisch ist er fast null**: 0,742 → 0,747, und der
Unterschied kommt aus **einer** Klasse (`mechanik`, wo die Typzeile wörtlich
passt). Auf einem Korpus, in dem Fragen und Dokumente mehr Wörter teilen —
etwa bei einer Suche nach einem Kartennamen —, wäre das Verhältnis ein
anderes; hier ist es nicht gemessen.

**Keine Regression der exakten Klassen.** `templating` und `mechanik` sind
hybrid nirgends schlechter als lexikalisch. Die Grenze „keine Regression" ist
eingehalten.

**Die Negativfälle halten überall.** Keine Betriebsart erfindet einen Treffer.

**`all-minilm` ist für deutsche Fragen zu schwach.** `de-en` erreicht 0,242
nDCG gegen 0,747 bei `en-en` — die Modellkarte weist es als englischsprachig
aus, und die Zahl bestätigt es. Wer dieses Programm auf Deutsch befragen will,
braucht ein mehrsprachiges Modell.

---

## Der Lauf vom 2026-10-04 — zwei Modelle (AI-02)

**Maschine**: Apple M5, 32 GB, macOS 27.0.1, `darwin/arm64`, 10 Kerne.
**Laufzeit**: Ollama 0.35.1. **Korpus** `probe-1`, **Frageset**
`frageset-1`, je Modell ein Durchgang, 78 Messpunkte je Betriebsart —
derselbe Lauf wie oben, über `MSE_MESSMODELL` auf ein zweites Modell
gerichtet (AI-02, #1743).

| | `all-minilm:22m` | `embeddinggemma:300m` |
| --- | ---: | ---: |
| Digest | `1b226e2802db…` | `85462619ee72…` |
| Stellen, gemessen | 384 | 768 |
| Gewichte auf der Platte | 46 MB | 622 MB |
| Indexbau, 120 Dokumente | 711 ms | 1.401 ms |
| Vektoren je Dokument | 1.536 Bytes | 3.072 Bytes |
| nDCG@10 semantisch, `en-en` | 0,742 | **0,900** |
| nDCG@10 hybrid, `en-en` | 0,747 | 0,895 |
| nDCG@10 semantisch, `de-en` | 0,242 | **0,787** |
| Recall@10 hybrid, `de-en` | 0,356 | 0,963 |
| Antwortzeit hybrid, p50 / p95 | 8,6 / 10,4 ms | 18,2 / 21,7 ms |
| Spitze im Speicher (Ollama, eigen + Gewichte) | 164 MB | 830 MB |

Die Mittel gehen über die sechs Klassen ohne die Negativfälle; der Speicher
stammt aus `docs/belege/modellmessung.md`.

**`all-minilm:22m` ist `all-minilm` vom 12.09.** — derselbe Digest, und unter
Ollama 0.35.1 kam **jede** Zahl der Güte wieder auf die dritte Stelle genau.
Das ist der Beleg für einen wiederholbaren Indexneubau über zwei Fassungen der
Laufzeit; die Zeiten sind leicht anders, weil die Maschine es ist.

**`embeddinggemma:300m` schließt die Lücke bei deutschen Fragen**: `de-en`
steigt von 0,242 auf 0,787 nDCG und liegt damit dicht an `en-en`. Es kostet
das Doppelte an Bauzeit, Vektorplatz und Antwortzeit und das Fünffache an
Speicher — alles weit unter den Grenzen oben. **Die hybride Hälfte bringt
hier nichts**: 0,895 gegen 0,900 rein semantisch; die Volltextsuche findet in
diesem Frageset zu wenig, um ein starkes Modell zu verbessern.

**Alle Grenzen halten**: p95 21,7 ms ≤ 150 ms, Bau rund 12 s je 1.000
Abschnitte ≤ 60 s, 3 KiB je Vektor ≤ 4 KiB, `en-en` 0,895 ≥ 0,60, keine
Regression in `templating` und `mechanik`.

---

## Offen

**Ein zweites Modell** — gemessen am 2026-10-04, oben.

**Ein echter Korpus.** 120 selbst geschriebene Dokumente sind keine 30.000
Oracle-Texte. Was hier gemessen ist, ist die Maschinerie.

**Peak-RAM** — seit AI-02 am rechnenden Ollama-Prozess gemessen
(`docs/belege/modellmessung.md`): eigener Speicher samt GPU-Puffern **und**
die eingeblendeten Gewichte, denn die zählen nicht zum Fußabdruck.

**Kalter und warmer Betrieb getrennt.** Der Lauf misst warm — das Modell ist
nach dem Indexbau geladen. Die Ladezeit steckt in der Bauzeit und ist nicht
einzeln ausgewiesen.

**Paralleler Rebuild.** Nicht gemessen. Dass eine Suche nicht auf einen Neubau
wartet, ist **gebaut** (zwei Ressourcenklassen) und in `internal/service`
geprüft; wie sich die Antwortzeit dabei verschiebt, ist es nicht.

**Ein Reranker.** Nicht gebaut. Das Issue führt ihn als optional; ohne eine
Messung, die einen Zusatznutzen zeigt, wäre er ein Modell mehr im Weg.

---

## Was daraus für die Vorgabe folgt

**Es wird kein Modell als Vorgabe gesetzt.** Das Issue erlaubt eine Vorgabe
„nur bei eingehaltenen Budgets, messbarem Hybridgewinn in semantischen Klassen
und dokumentierter zulässiger Regression" — und eine Messung über einen selbst
geschriebenen Korpus mit **einem** Modell trägt diese Entscheidung nicht.

**Diese Fassung liefert die hybride Suche als Fähigkeit, nicht als
eingerichteten Zustand.** Ohne Einbettungsmodell sucht das Programm
lexikalisch und **sagt es**; wer ein Modell einrichtet und einen Index baut,
bekommt die hybride Suche mit dem Modell seiner Wahl. Der lexikalische
Rückfall wird nirgends als abgeschlossene semantische Suche ausgegeben — das
ist die Zeile, die das Issue zuletzt verlangt.
