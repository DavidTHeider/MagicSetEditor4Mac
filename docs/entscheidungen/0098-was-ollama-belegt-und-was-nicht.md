# 0098 — Was Ollama belegt und was nicht

**Stand:** angenommen · **Datum:** 2026-09-11 · **Werkstück:** A7 von Phase 5.3
(#1446)

## Worum es geht

Mit A7 bekommt der Modellkatalog aus
[0097](0097-was-ueber-ein-ki-modell-feststeht.md) seinen ersten Anbieter. Die
schwierige Frage ist nicht, wie man `/api/tags` liest, sondern **was eine
Antwort von dort eigentlich belegt** — und was man ihr nicht entnehmen darf,
ohne etwas zu erfinden.

Sechs Entscheidungen halten das auseinander.

## 1 Die Positivliste ist geschlossen, und ein Wächter liest sie nach

`internal/ollama` kennt **fünf** Adressen: `/api/version`, `/api/tags`,
`/api/show`, `/api/chat`, `/api/embed`. Es gibt kein `/api/pull`, kein
`/api/create`, kein `/api/delete`.

Das Issue verlangt es wörtlich: „keine automatische
Installation/Downloads/Modellpulls" und „Kein impliziter Modellpull, kein
Cloudfallback, kein global installiertes Tool."

**Zwei Läufe halten es, und sie messen Verschiedenes.** Ein Spion sagt, welche
Wege die heutigen Rufe wirklich gehen; ein Leser über den **Quelltext** sagt,
dass es keine anderen gibt. Einer allein reichte nicht (§10, „die Reichweite
einer Messung"): der Spion sähe ein `/api/pull` nicht, das nur unter einer
Bedingung gerufen wird, und der Leser sähe nicht, ob ein Weg der Liste
überhaupt begangen wird.

**Die Gegenprobe ist gefahren**: ein eingefügtes `/api/pull` macht den Leser
rot.

## 2 Loopback beweist nichts, eine fremde Adresse beweist etwas

Das Issue sagt beide Hälften:

- „**Loopback allein beweist keine lokale Inferenz**" — auf `127.0.0.1` kann
  ein Stellvertreter sitzen, der weiterreicht.
- „Remote-Ollama nur ausdrückliche Konfiguration und **eindeutige
  Einstufung**."

Daraus folgt eine **einseitige** Regel
([`domain.KIOrtFuerAdresse`](../../internal/domain/kiadresse.go)): eine
Adresse, die nicht auf der Schleife liegt, stuft jedes Modell dahinter als
nicht-lokal ein. Umgekehrt heißt „auf der Schleife" **nicht** „lokal", sondern
nur: dieses Programm hat keinen Gegenbeweis.

**Und das Issue sagt auch, was dabei bleibt:** „Für beliebige fremde
Loopback-Proxies **keine absolute Netzwerkgarantie** behaupten." Genau deshalb
steht die Einstufung an der Adresse und nicht an der Antwort der Laufzeit —
die meldet ein Modell ohne `-cloud`-Suffix als lokal, gleich wo sie steht.

## 3 Der Cloudstand hat drei Werte, nicht zwei

Ollama nennt **zwei** dokumentierte Wege, die Auslagerung abzuschalten
(nachgelesen am 2026-09-11 unter <https://docs.ollama.com/cloud> und
<https://docs.ollama.com/faq>):

- `~/.ollama/server.json` mit `{"disable_ollama_cloud": true}` — **eine Datei,
  die zu lesen ist**;
- `OLLAMA_NO_CLOUD=1` — **die Umgebung eines fremden Prozesses**, und die ist
  es nicht.

Wer eine fehlende Datei als „Cloud an" ausgäbe, behauptete etwas über eine
Umgebung, die er nicht kennt. Deshalb: `aus`, `an`, **`ungewiss`** — und jeder
Stand trägt seinen Beleg.

## 4 Strukturierte Ausgabe ist eine Fähigkeit der Laufzeit

Ollama meldet unter `capabilities` das Wort `completion`. Dass eine Antwort in
eine **Form** gezwungen werden kann, steht dort **nicht** — es steht in der
Dokumentation: der Parameter `format` nimmt ein JSON-Schema, und die Laufzeit
hält es (nachgelesen am 2026-09-11 unter
<https://docs.ollama.com/capabilities/structured-outputs>).

Dieses Programm trägt deshalb neben `textgenerierung` auch
`strukturiert` ein — **und nur für lokale Modelle**. Ebenda: „Ollama's Cloud
currently does not support structured outputs." Eine Fähigkeit, die hier für
ein Cloudmodell behauptet würde, führte zu einer Antwort ohne Form; sie fiele
erst an der Schemaprüfung auf, und dann sähe es aus wie ein Fehler des Modells.

### Und ein Modell, das denkt, wird gebeten, es zu lassen

**Nachgetragen am 2026-10-04 in AI-02**
(#1743):
`qwen3:0.6b` meldet `thinking` unter `capabilities`, und **ohne das Feld
`think` schaltet Ollama 0.35.1 das Denken ein**. Gemessen über die 32 Fragen
des Testsatzes, gestellt wie dieses Paket sie stellt
(`docs/belege/modellmessung.md`): ohne das Feld dachte qwen3 vor jeder
Antwort — Median 267 Token und 1,8 s —, und eine Frage nach einem
Kartennamen hatte nach drei Minuten und über 7.400 Token Denktext noch keinen
Inhalt. Mit `"think": false`: Median 14 Token und 0,16 s, keine Frage über
eine halbe Sekunde. `gemma3:1b`, das nicht denken kann, nimmt das Feld ohne
Fehler an.

**`/api/chat` schickt deshalb immer `"think": false`.** Dieses Paket liest
`message.thinking` gar nicht; Denktext wäre in jedem Fall bezahlt und
weggeworfen. **Was das kostet**: ein Modell, das nur denkend gut antwortet,
antwortet hier schlechter, als es könnte — gemessen ist das nicht, und ein
Aufschalten wäre eine eigene Entscheidung mit eigener Messung.

**Und wie weit die Zusage reicht** — angestrichen vom Zweitblick über Phase
6.11 (Z6): gemessen ist `"think": false` an **zwei** Modellen, `qwen3:0.6b`
und `gemma3:1b`. Ollamas Dokumentation (*Thinking*, nachgelesen am
2026-10-04) sagt, `false` gelte, **wenn das Modell es erlaubt**; welche Werte
eines erlaubt, nennt `/api/show` unter `thinking.values` — Modelle mit
benannten Stufen wie `gpt-oss` führen dort `low`, `medium`, `high`. Was ein
solches Modell mit `false` tut, sagt die Dokumentation nicht, und gemessen
ist es hier nicht: die Modelle aus AI-02 sind gelöscht, ein weiteres zu
ziehen braucht eine Freigabe. **Für jedes andere Modell ist die Wirkung
deshalb eine Annahme**, kein Beleg.

## 5 Ein Einbettungsvertrag wird gemessen, nicht geraten

[0097](0097-was-ueber-ein-ki-modell-feststeht.md) verlangt von einem
Einbettungsmodell vier Angaben. Sie kommen aus **drei** Quellen, und keine ist
geraten:

| Teil | Woher |
| --- | --- |
| Zusammenfassung | `model_info`, Schlüssel auf `.pooling_type` |
| Zerteiler | `model_info["tokenizer.ggml.model"]` |
| Dimension | **gemessen** an einem echten Lauf |
| Normalisiert | **gemessen** am Betrag des Vektors |

Die beiden Schlüsselnamen sind am 2026-09-11 in `gguf-py/gguf/constants.py`
von llama.cpp nachgelesen: `Keys.LLM.POOLING_TYPE = "{arch}.pooling_type"` —
architekturabhängig, deshalb wird ein **Suffix** gesucht — und
`Keys.Tokenizer.MODEL = "tokenizer.ggml.model"`, fest. Ebenda zählt
`PoolingType` `NONE = 0`, `MEAN = 1`, `CLS = 2`, `LAST = 3`, `RANK = 4`; null
und vier tragen keinen Vertrag, denn bei `NONE` kommt ein Vektor je **Zeichen**
heraus und bei `RANK` eine Bewertung.

**Fehlt die Zusammenfassung, gibt es keinen Vertrag** — sie ist der eine Teil,
der sich nicht messen lässt. Ein geratener Wert ergäbe Vektoren, die
systematisch danebenliegen, **ohne dass etwas scheitert**: ein Index darauf
wäre still falsch. Das Modell bleibt im Katalog, mit null Fähigkeiten und einem
Befund, der den Grund nennt.

**Und nichts wird still gekürzt.** `/api/embed` hat `truncate: true` als
Vorgabe („If true, truncate inputs that exceed the context window. If false,
returns an error", nachgelesen am 2026-09-11); dieses Paket sagt ausdrücklich
`false` und nimmt den Fehler. Ein gekürzter Text gibt einen Vektor, der zu
einem **anderen** Text gehört.

## 6 Keine Mindestfassung — und das ist eine Entscheidung, keine Lücke

Das Issue nennt „Version inkompatibel" als einen der Zustände, die eine
Einrichtungsansicht unterscheiden soll. Dieses Programm prüft **keine**
Mindestfassung.

**Der Grund ist nachgelesen, nicht erfunden**: „Ollama's API isn't strictly
versioned, but the API is expected to be stable and backwards compatible"
(2026-09-11). Eine Zahl, die dieses Programm sich ausdächte, wiese morgen eine
Installation ab, die tut, was sie soll — §10, „auch eine Ablehnung braucht
Deckung".

**Die Fassung steht trotzdem in jeder Antwort.** Sie gehört zu jedem Beleg:
dieselbe Datei tut unter zwei Fassungen nicht dasselbe, und wer eine
Unverträglichkeit sieht, kann sie benennen.

## Was das alles nicht belegt

**Dass auf 11434 wirklich Ollama liegt.** Ein beliebiger Dienst kann dort
antworten und weiterleiten. Was dieses Paket hat, ist eine **Auskunft von dem,
was antwortet** — und die steht mit ihrer Herkunft da, nicht als Gewissheit.

**Hier stand bis zum 2026-09-12: „kein Lauf gegen eine echte
Installation."** Auf dem Rechner, auf dem A7 gebaut wurde, war Ollama nicht
installiert. **Er ist es seit A15** — der Retrievalindex
(#1514)
braucht einen echten lokalen Einbetter, und die Installation ist ausdrücklich
freigegeben worden.

**Gelaufen ist er damit, und zwar ganz** (`internal/ollama/netzlauf_test.go`,
Ollama 0.33.3, 2026-09-12):

| Was | Ergebnis |
| --- | --- |
| Fassung | `Ollama 0.33.3` |
| Modelle | `gemma3:270m` mit **zwei** belegten Fähigkeiten, `all-minilm:latest` mit **null** |
| Feste Form | `gemma3:270m` antwortete `{"name": "Blitzschlag"}` gegen das echte Schema |
| Einbettungsvertrag | **gemessen**: Dimension 384, Pooling `mittel`, normalisiert, Zerteiler `bert` |
| Zu langer Text | wirklich abgewiesen, nicht still gekürzt |

**Zwei Sätze dieses Vermerks sind damit belegt statt behauptet.** Abschnitt 5
sagt, der Einbettungsvertrag werde gemessen — jetzt ist er es. Und dass eine
**Meldung noch kein Vertrag** ist, steht als Befund im Lauf: Ollama meldet für
`all-minilm` die Einbettung, und die Fähigkeit gilt trotzdem erst nach der
Vermessung.

**Was weiter offen bleibt**: der Lauf läuft nicht in der Pipeline und zählt
nicht zur Abdeckung — er braucht eine Installation, die kein zweiter Rechner
haben muss. Er steht hinter dem Build-Tag `netzlauf`, und `docs/HAERTUNG.md`
führt ihn unter der fünften Prüfebene.
