# 0102 — Wo ein erzeugtes Bild liegt, und warum dieses Programm nur Bytes nimmt

**Stand:** angenommen · **Datum:** 2026-09-12 · **Werkstück:** A12 von Phase 5.3
(#1451)

## Worum es geht

A12 stellt seinen Entscheidungspunkt voran:

> „Unsichere externe Fähigkeit, Infrastruktur oder Unterstützungsgrenze zuerst
> dokumentiert klären. Ein negatives Ergebnis ist keine Implementierung."

Die unsichere Fähigkeit ist die Bilderzeugung, und die Frage dahinter ist
kleiner, als sie klingt: **wo liegt das Bild, wenn der Dienst geantwortet
hat?** Von ihr hängt alles Weitere ab — ob es einen Downloadweg gibt, ob eine
Adresse ablaufen kann, ob ein Umleiter eine Rolle spielt, ob ein Schlüssel an
ein CDN gerät.

Drei Anbieter stehen in diesem Programm (A7, A9, A10). Sie geben **drei
verschiedene Antworten**, und eine davon ist „gar nicht".

Alles Folgende ist am **2026-09-12** nachgelesen.

## 1 OpenAI: immer Base64 — und `url` gibt es für diese Familie nicht

`POST /v1/images/generations` und `POST /v1/images/edits`. Die Antwort trägt
`created`, `data`, `usage`, dazu `background`, `output_format`, `quality` und
`size`; jeder Eintrag in `data` kann `b64_json`, `url` und `revised_prompt`
haben.

**Für die `gpt-image`-Familie ist `url` nicht unterstützt**, und
`response_format` wird nicht angenommen — die Dokumentation sagt, diese
Modelle geben „always base64-encoded images". Eine Adresse bekommt nur, wer
`dall-e-2` oder `dall-e-3` ruft; **sie gilt 60 Minuten**.

Nachgelesen unter
<https://developers.openai.com/api/docs/api-reference/images/create>.

Die Modellnamen, die `/v1/images/edits` annimmt, nennt die Referenz
ausdrücklich: `gpt-image-1`, `gpt-image-1-mini`, `gpt-image-1.5`,
`gpt-image-2`, `gpt-image-2-2026-04-21`, `gpt-image-2.5-sunburst`,
`gpt-image-2.5-flare` (beide auch mit Datumsstand `2026-09-08`) und
`chatgpt-image-latest`.

## 2 xAI: eine Adresse

`POST /v1/images/edits`, `application/json`, und was zurückkommt, ist ein
`url`. Das steht seit A10 in
[0101 §8](0101-was-xai-belegt.md) und wird hier nicht wiederholt, sondern
gebraucht: **derselbe Weg, ein anderer Rumpf, eine andere Antwortform.**

## 3 Ollama: über HTTP gar nicht — der Negativbefund

Der Blogbeitrag vom Anbieter kündigt Bilderzeugung als **Befehl** an:

```
ollama run x/z-image-turbo "…"
```

Die Bilder landen im aktuellen Verzeichnis. Die **Bildbearbeitung** steht dort
unter „What's next", also ausdrücklich noch nicht.

Und die Dokumentation kennt dazu **keinen Weg**: der Seitenindex unter
<https://docs.ollama.com/llms.txt> führt die Fähigkeiten einzeln — `streaming`,
`thinking`, `structured-outputs`, `vision`, `embeddings`, `tool-calling`,
`web-search` — und **keine Bilderzeugung**. `vision` ist das Lesen eines
Bildes, nicht sein Entstehen. Es gibt damit weder einen Endpunkt noch eine
Fähigkeitszeichenkette, an der `/api/show` sie belegen könnte.

### Nachgemessen am 2026-09-12, an der laufenden Fassung

**Seit A15 liegt Ollama auf diesem Rechner** (0.33.3, über Homebrew), und
damit ließ sich die Auskunft von oben **messen** statt lesen:

| Weg | Antwort |
| --- | --- |
| `POST /api/generate` mit echtem Modell | **200**, vollständige Antwort |
| `POST /api/images` | `404 page not found` |
| `POST /v1/images/generations` | `404 page not found` |

**Der Unterschied zwischen den zwei 404 ist der Beleg.** Ein unbekanntes
**Modell** beantwortet Ollama mit einem JSON-Fehlerrumpf; eine unbekannte
**Route** mit dem nackten Text aus Gos `http.NotFound`. Der erste Anlauf
dieser Messung fragte `/api/generate` mit leerem Rumpf, bekam ebenfalls 404
und hätte fast „den Weg gibt es nicht" ergeben — er war nur „das Modell gibt
es nicht" (§10: die Reichweite einer Messung ist die Reichweite ihrer
Aussage).

**Und die Bauart steht dagegen, nicht nur die Fassung.** Bilderzeugende
Modelle sind Diffusionsmodelle: Z-Image Turbo hat 6 Mrd. Parameter, FLUX.2
Klein mehr; quantisiert sind das 4–12 GB. Die kleinsten destillierten
Varianten liegen bei 0,5–2 GB. Zwischen dem Einbettungsmodell dieses
Projekts (45 MB) und einem Bildmodell liegen zwei Größenordnungen. **Gezogen
wurde keines** — die Aussage stützt sich auf die Parameterzahlen der
Modellkarten, nicht auf einen Lauf.

**Das ist der dokumentierte Negativbefund, den der Entscheidungspunkt
verlangt** — und er hat eine Folge, die im Issue schon steht: A9 bleibt der
Referenz-Bildanbieter, denn der Wechsel auf A7 wäre ein Wechsel auf etwas, das
es über die Naht dieses Programms nicht gibt. `internal/ollama` bekommt
deshalb in diesem Werkstück **keinen** achten Weg.

### Nachgemessen am 2026-10-04, mit einem gezogenen Bildmodell

AI-02 (#1743)
hat gezogen, was oben ausdrücklich nicht gezogen war: `x/flux2-klein:4b`,
5.727 MB, Digest `50a0c0ab15ac…`, an **Ollama 0.35.1** aus Homebrew (mit
`mlx` 0.32.1 und `mlx-c` 0.6.0 als Abhängigkeiten). Zwei Dinge haben sich
seit oben geändert, eines nicht:

| Frage | 0.33.3 (2026-09-12) | 0.35.1 (2026-10-04) |
| --- | --- | --- |
| Belegt `/api/show` die Fähigkeit? | es gab keine Zeichenkette | **ja**: `capabilities: ["image"]`, dazu `requires: 0.14.0` |
| Gibt es einen Bildweg? | `/api/images`, `/v1/images/generations` → 404 | **unverändert** 404 |
| Erzeugt `/api/generate` ein Bild? | nicht gefragt — kein Modell | **nein**: `400 {"error":"image generation models are not currently supported"}` |

**Und der eigene Befehl der Laufzeit sagt dasselbe**: `ollama run
x/flux2-klein:4b "…"` endet mit derselben Meldung, und im Verzeichnis liegt
nichts. Die Laufzeit kennt das Modell, zieht es, nennt seine Fähigkeit — und
führt es nicht aus. Der Beleg samt Wortlaut steht in
[`docs/belege/modellmessung.md`](../belege/modellmessung.md).

**Die Entscheidung oben bleibt damit, und ihr Grund wird enger**: nicht
„Ollama kann keine Bilder", sondern „diese Laufzeit führt ein Bildmodell
nicht aus, obwohl sie es kennt". Ein achter Weg wäre ein Weg zu einer
Antwort, die heute 400 lautet.

## 4 Die Entscheidung: dieses Programm nimmt nur Bytes

Aus den drei Antworten folgt eine Regel, und sie ist enger als die drei
Verträge zusammen:

> **Ein Bild erreicht dieses Programm als Bytes in der Antwort oder gar
> nicht.** Ein `url` im Rumpf wird **nicht geholt**; er wird als Befund
> geführt.

Das Issue verlangt „keine frei nachladbaren Fremd-URLs, keine Secrets an
CDN/Redirect weiterreichen". Wer das als **Prüfung** baut, hat einen Zweig,
der etwas erlaubt und etwas verbietet — und die Abnahme muss dann zeigen, dass
der verbietende Zweig auch betreten wird. Wer es als **Form** baut, hat
keinen Zweig: es gibt keinen Holweg, den eine Adresse auslösen könnte, und die
Positivliste aus §16 führt keinen.

**Damit fallen drei Abnahmepunkte des Issues in sich zusammen** — abgelaufene
Download-Adresse, Umleiter, Schlüssel am CDN. Sie können nicht eintreten, weil
die Anwendung nach der Antwort nicht mehr ins Netz geht. Das ist dieselbe
Gestalt wie der Rohprompt in A11: **eine Zusage, die niemand einhalten muss,
weil es die Möglichkeit nicht gibt.**

Was sie kostet, gehört dazu: **xAI taugt damit als Bildanbieter nicht** —
nicht weil sein Vertrag schlecht wäre, sondern weil dieses Programm die Form
seiner Antwort nicht annimmt. Das ist eine Wegentscheidung und keine Wertung,
und sie ist umkehrbar: wer den Holweg baut, baut ihn samt Fristen, Umleitern,
Größengrenze und dem Beleg, dass der Schlüssel nicht mitreist.

## 5 Der Bearbeitungsweg hat zwei Rumpfformen, und beide heißen gleich

**Derselbe Pfad, zwei verschiedene Rümpfe.** Das ist der Fund aus
[0101 §8](0101-was-xai-belegt.md), und er ist jetzt auf beiden Seiten genau
gelesen:

| | OpenAI | xAI |
| --- | --- | --- |
| Weg | `POST /v1/images/edits` | `POST /v1/images/edits` |
| Rumpf | `application/json` **oder** `multipart/form-data` | nur `application/json` |
| Feld | `images` — eine **Liste** | `image` — ein **Objekt** |
| Eintrag | `{"image_url": "…"}` oder `{"file_id": "…"}` | `{"type": "image_url", "url": "…"}` |
| Höchstzahl | 16 Bilder | in der Dokumentation nicht genannt |
| Maske | `mask`, genau eines von `image_url` oder `file_id` | nicht genannt |
| Antwort | `data[].b64_json` | `url` |

`image_url` nimmt „a fully qualified URL or base64-encoded data URL". **Dieses
Programm schickt die Daten-URI** und nie eine Adresse: eine Adresse setzte
voraus, dass das Bild irgendwo öffentlich liegt, und die Bilder dieses
Programms liegen neben dem Set
([0029](0029-bilder-liegen-neben-dem-set.md)).

**Und `file_id` wird nicht benutzt.** Es setzte `/v1/files` voraus, also einen
Weg, der beim Anbieter **etwas ablegt**. Die Positivliste von A9 nennt genau
zwei Wege und sagt in ihrem eigenen Kommentar, dass es kein `/v1/files` gibt;
A12 fügt zwei hinzu, die beide **nichts hinterlassen**.

Die Positivliste von `internal/openai` wächst damit von zwei auf **vier**:

```
/v1/models   /v1/responses   /v1/images/generations   /v1/images/edits
```

Das ist eine **erweiternde** Änderung im Sinne von
[0033](0033-was-ueber-fassungen-hinweg-zugesagt-wird.md): kein vorhandener
Ruf ändert seine Bedeutung.

## 6 Was ein Bild wiegt, entscheidet nicht der Anbieter

`size` kennt `1024x1024`, `1536x1024`, `1024x1536` und freie Maße — Vielfache
von 16, Seitenverhältnis zwischen 1:3 und 3:1, keine Kante über 3840 Pixel,
insgesamt zwischen 655.360 und 8.294.400 Pixeln. `output_format` kennt `png`
(Vorgabe), `jpeg` und `webp`, `output_compression` 0–100 für die letzten
beiden. Für die Bearbeitung nennt die Dokumentation eine Dateigrenze von
**unter 50 MB** je Bild.

**Keine dieser Zahlen ist die Grenze dieses Programms.** Sie sagen, was der
Dienst annimmt; was hier ankommt, muss durch die Ablage passen und in einer
Karte landen. Die Schranken stehen deshalb auf dieser Seite, und sie stehen
**doppelt** — in Bytes und in Pixeln:

- eine komprimierte Antwort von 50 MB ist eine Bildbombe, gleich ob der Dienst
  sie annimmt,
- und eine 2 KB große Datei mit 40.000 × 40.000 Pixeln ist es auch. Sie wiegt
  erst beim Auspacken, und dann wiegt sie 6,4 GB.

Deshalb wird **vor** dem Auspacken die Bytezahl geprüft, dann der Kopf der
Datei gelesen — Signatur, nicht Endung und nicht `Content-Type` —, dann die
Maße aus dem Kopf, und **erst dann** das Bild wirklich dekodiert.

## Was das kostet

**Nichts davon ist gegen den echten Dienst gelaufen.** Ein Lauf brauchte einen
Zugang, und ein Bildauftrag kostet Geld — anders als eine Modellliste. Was
hier steht, ist die **Dokumentation** vom 2026-09-12; was das Programm tut,
ist gegen selbst geschriebene Beispiele geprüft (§16).

Das heißt konkret: die Form ist belegt, das **Verhalten** nicht. Wenn OpenAI
morgen für `gpt-image` doch eine Adresse zurückgibt, sieht dieses Programm
einen Befund und kein Bild — laut und nicht still (§16, Drift). Das ist der
Fall, für den die Regel aus Abschnitt 4 gebaut ist.

**Und die Wails-Probe zeigt die Naht, nicht den Anbieter.** Sie fährt gegen
einen Server, der den dokumentierten Vertrag spricht; damit ist der Weg durch
Fenster, Dienst, Vorrat, Ablage und Karte wirklich gelaufen — und über
`api.openai.com` sagt sie nichts.
