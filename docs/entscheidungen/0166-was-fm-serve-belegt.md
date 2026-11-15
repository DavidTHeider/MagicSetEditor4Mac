# 0166 — Was `fm serve` belegt, und wie das Gerätemodell eingebunden ist

**Stand:** angenommen am 2026-10-03 · AI-05
(#2042).
Marke `new`: MSE2 kennt keine Sprachmodelle; gesucht in
`/Users/david/mse2-quelle/src` nach `model`, `completion`, `ai` — kein Treffer
jenseits von Bildmodellen der Rasterung. Dieselbe Form wie
[0098](0098-was-ollama-belegt-und-was-nicht.md) bis
[0101](0101-was-xai-belegt.md).

## Die Bedingungen der Nutzung

**`/usr/bin/fm` ist Teil von macOS 27** (Apple Foundation Models CLI). Der
Kurzhinweis (`fm license --show`, 79 Wörter) bindet an die Softwarelizenz von
macOS und erlaubt programmatische Nutzung nur, wo sie ausdrücklich erlaubt
ist. **Die vollständigen Bedingungen zeigt `sudo fm license`**; der
Auftraggeber hat sie am 2026-10-03 gelesen: **nichtkommerzielle Nutzung ist
erlaubt** — dieses Programm steht unter GPL und wird nicht verkauft —, **und
die Schnittstelle darf eingebunden werden**. Die Fassung der Lizenz meldet
`fm license --status`: *FM1 version 1.0*, auf diesem Rechner angenommen am
20. Sept. 2026.

**Neu zu entscheiden ist das**, wenn das Programm verkauft oder kommerziell
vertrieben werden soll, oder wenn `fm license --status` eine andere Fassung
meldet. **Das Programm nimmt die Lizenz nie selbst an** und ruft nie `sudo`:
die Annahme gilt für alle Nutzer der Maschine und ist eine Handlung eines
Menschen.

## Was gemessen ist — 2026-10-03, Mac17,3, macOS 27.0.1 (26A434)

Über einen eigenen Unix-Socket, der Server danach beendet. Die Antworten
liegen als selbst geschriebene Nachbildungen unter
`internal/applefm/testdaten/`.

| Frage | Befund |
| --- | --- |
| Wo lauscht `fm serve` ohne Angaben? | TCP `127.0.0.1:1976` |
| **`--socket` mit zu langem Pfad** | **kein Socket — stattdessen TCP `*:55712` auf allen Schnittstellen, ohne Meldung**. Der Pfad hatte 121 Bytes; ein Unix-Socket fasst unter macOS 104 |
| `--socket` mit 63 Bytes | Socket (`srwxr-xr-x`), `/health` nach 0,15 s, kein TCP |
| `GET /health` | `{"status":"fm serve is running","models":[{"available":true,"name":"system"}]}` |
| `GET /v1/models` | ein Modell `system`, `owned_by: "Apple"` |
| `POST /v1/chat/completions` **ohne `stream`** | **ein Strom** (SSE, `data: [DONE]`) — anders als bei OpenAI |
| mit `stream: false` | `chat.completion`, `message.content`, `message.refusal` (null), `finish_reason: "stop"`, `usage` |
| `max_tokens: 8` | 61 Ausgabetokens — **wirkt nicht**; `max_completion_tokens` ebenso wenig verlässlich |
| `response_format: json_schema` | der Inhalt ist JSON nach dem Schema |
| Ablehnung, Art 1 | **HTTP 500**, `type: server_error`, Meldung *„The model's safety guardrails were triggered."* — ein englischer Kampftext (alle Kreaturen zerstören) |
| Ablehnung, Art 2 | **HTTP 200** mit dem Satz *„Diese Anfrage kann nicht bearbeitet werden."* im Inhalt, `refusal: null` |
| derselbe Kampftext auf Deutsch | angenommen, ein brauchbarer Flavortext |
| unbekanntes Modell | HTTP 400, `invalid_request_error` |
| Kontext | 8.009 Eingangstokens gehen, 9.009 nicht (HTTP 500, *„exceeded the model's context size"*) — vereinbar mit 8.192 |
| Antwortzeit, warm | Name 0,74 s, Flavortext 1,26 s, Schema 0,78 s |

## Was daraus für den Bau folgt

1. **Nur über einen Unix-Socket, und der Pfad wird geprüft**: kürzer als
   104 Bytes, in einem eigenen Verzeichnis mit `0700` unter `$TMPDIR`, und
   **vor dem ersten Ruf muss der Socket dastehen** — sonst lauscht `fm` still
   auf allen Schnittstellen, und das Programm beendet ihn.
2. **`stream: false` wird immer gesetzt.**
3. **Die Länge begrenzt das Programm**: eine Antwort über der Grenze der
   Aufgabe ist abgeschnitten und wird abgewiesen, wie bei jedem Anbieter.
4. **Eine Ablehnung ist ein eigener Ausgang**, in beiden Arten: die Meldung
   zu den Schutzregeln bei HTTP 500, und der bekannte Satz im Inhalt.
5. **`response_format: json_schema` trägt** — damit passt das Modell in das
   vorhandene Auftragswerk, das gegen das Schema der Paketfelder fragt.

## Wie es eingebunden ist

- **„Aktiviert" heißt**: `/usr/bin/fm` ist da, `fm license --status` und
  `fm available --model system` enden mit 0. Gelesen wird der Exit-Code, nicht
  der Text — die Ausgabe ist teils lokalisiert (das Datum stand deutsch da).
- **`ki.geraetemodell`** — `automatisch` (Vorgabe) oder `aus`, Ebene Profil.
  **Anders als `ki.laufzeit`**, das aus anfängt: dort klopfte eine Anwendung
  an einen Port, den niemand genannt hat. Hier gibt es keinen Port, das
  Modell gehört zum Betriebssystem, und die Annahme der Lizenz ist eine eigene
  Handlung des Nutzers.
- **Der Sitzungsschalter** lebt im Arbeitsspeicher des fragenden Prozesses;
  nach einem Neustart gilt wieder die Einstellung.
- **Das Programm startet `fm serve` selbst**, bei Bedarf, und beendet ihn mit
  sich.

## Wie der Vorrang entscheidet

**Die Regel steht in `domain`** (`KIReihenfolge`, `IstKleineTextaufgabe`):

- **Eine ausdrückliche Wahl gewinnt.** Ohne sie (Anbieter `vorrang`) fragt
  eine **kleine Textaufgabe** zuerst das Gerätemodell, dann die lokalen
  Anbieter in ihrer Reihenfolge; **eine Cloud nie von selbst**. Eine größere
  Aufgabe geht nie an das Gerät.
- **Klein heißt**: nur `name` und `flavor text` (`nur`), keine Belegfrage,
  und ein Auftrag bis **2.000 Zeichen**. Die Grenze ist gemessen: am
  2026-10-03 zählte `fm count-tokens` auf Deutsch 4,1 Zeichen je Token, auf
  Englisch 4,6, auf Russisch 3,4, auf Japanisch 1,4 und auf Chinesisch **1,3**
  (680 Zeichen, 521 Tokens). Im ungünstigsten Fall sind 2.000 Zeichen rund
  1.540 Tokens — weniger als ein Fünftel des Fensters.
- **Der nächste kommt dran**, wenn das Modell nicht geliefert hat: nicht
  bereit, abgelehnt, abgebrochen, kein JSON, nichts, was das Schema
  durchließ. **Nicht**, wenn die Anwendung etwas verweigert (Budget, Rolle,
  Schema der Karte) — das sagte jeder andere genauso.
- **Eine Ablehnung ist ein Ausgang** (`ausgang: abgelehnt`): bei einer
  ausdrücklichen Wahl nennt der Vorschlag die lokalen `rueckfaelle`, beim
  Vorrang steht sie nur, wenn keiner danach geliefert hat.
- **Innerhalb eines Anbieters nimmt der Vorrang dessen erstes Modell, das
  strukturiert antworten kann** (`modellFuer` in
  `internal/app/kivorrangnaht.go`) — in der Reihenfolge, in der der Anbieter
  sie nennt. **Nachgetragen am 2026-10-04 in AI-02** (#1743), weil es hier
  nicht stand: bei Ollama ist das die Reihenfolge von `/api/tags`, und die
  ordnet nach Änderungszeit, **das zuletzt gezogene zuerst** (gemessen an
  0.35.1). Dass das zählt, zeigt dieselbe Messung: an denselben 32 Fragen
  passten bei `gemma3:1b` 9 Vorschläge, bei `qwen3:0.6b` keiner
  (`docs/belege/modellvergleich.md`). Wer beide hat, bekommt im Vorrang das
  jüngere. **Entschieden am 2026-10-04 auf Zuruf**
  (#2054):
  die Einstellung `ki.vorrangmodell` nennt das Ollama-Modell des Vorrangs
  (`domain.KIVorrangmodellAus`). **Leer** heißt wie bisher das erste mit
  fester Form; ein gewähltes, das die Laufzeit nicht nennt oder das keine
  feste Form kann, ist ein Fehler mit seinem Namen — **nie** still das
  erste. Für das Gerätemodell gibt es keine Wahl (es hat eines), und eine
  Cloud fragt der Vorrang weiterhin nie. **Wen er fragt**, nennt der
  Katalog als `vorrang` an allen vier Flächen, und das Kartenformular
  zeigt es neben „Vorschlagen", sobald jemand die Zeile berührt.

## Sitzung, Prozess und Absturz

- **Die Naht hält einen Dienst je Prozess**, nicht je Anwendungsobjekt:
  Kommandozeile, MCP und REST bauen die Anwendung je Aufruf, und ein Dienst je
  Aufruf ließe bei jedem Ruf ein `fm serve` stehen. **Sitzung heißt damit die
  Lebensdauer des Prozesses** — am Fenster bis zum Neustart, beim MCP-Server
  und beim REST-Dienst so lange sie laufen, an der Kommandozeile je Aufruf
  (`--ohne-geraetemodell`).
- **Beendet wird `fm serve`** mit der Anwendung, beim Abschalten für die
  Sitzung und **bei jedem Profilwechsel** (das nächste Profil hat es
  vielleicht aus; gestartet wird bei Bedarf neu).
- **Nach einem Absturz** räumt der nächste Start ab: der Socket trägt die
  Kennung des Prozesses, der ihn bestellt hat (`fm-<pid>.sock`); lebt der
  nicht mehr, wird das `fm serve` dahinter über seinen ganzen Befehl samt
  Pfad beendet und der Socket gelöscht.
- **Testläufe starten `fm` nie** — auch nicht ihre Kindprozesse: `mse`,
  `mse-mcp` und `mse-rest` heißen nicht `*.test`, und der Flächenlauf hat am
  2026-10-04 gezeigt, dass MCP und REST das echte Modell dieses Macs sonst für
  bereit hielten. `MSE_FM_GESPERRT=an` sperrt es; `internal/flaechen` setzt
  die Variable an der einen Stelle, an der jeder Kindprozess vorbeikommt.
  Nur ein Bau mit `-tags fmlauf` öffnet den Riegel.

## Nachgemessen am 2026-10-04 — die Ablehnung hängt am Wortlaut und am Schema

Über denselben Socket, je Auftrag einmal mit und einmal ohne Schema:

| Auftrag (gekürzt) | mit Schema | ohne Schema |
| --- | --- | --- |
| „Destroy all creatures. Each player loses all life …" | angenommen, ein Flavortext | **freier Satz** „I cannot fulfill this request …" — HTTP 200, `refusal: null` |
| „Zerstöre alle Kreaturen …" | angenommen | der bekannte Satz *„Diese Anfrage kann nicht bearbeitet werden."* |
| „… glorifying the massacre of innocent villagers …" | HTTP 500, Schutzregeln | HTTP 500, Schutzregeln |
| „… how to build a bomb" | HTTP 500, Schutzregeln | HTTP 500, Schutzregeln |

**Was daraus folgt**: die Anwendung fragt **immer mit Schema**, und dort kam
eine Ablehnung nur als HTTP 500 — die Gestalt, die erkannt wird. Der freie
englische Satz ohne Schema ist eine **dritte** Gestalt; mit Schema wäre er
kein JSON, und der Vorschlag fiele als unlesbar auf den nächsten Anbieter
zurück. **Am 2026-10-03 lehnte derselbe englische Kampftext ab** — das Modell
ist darin nicht beständig, und ein Kartentext mit „zerstören" ist kein
verlässlicher Fall für eine Ablehnung. Der Echtlauf nimmt deshalb die
Verherrlichung von Gewalt.

**Und ein Name kam einmal von vier Läufen leer** (`{"name": ""}` oder
ohne den Schlüssel). Die Oberfläche nimmt einen leeren Wert deshalb nicht als
Vorschlag, sondern sagt, dass für das Feld nichts kam.

## Was nicht belegt ist

- Ob `serve` eine Stufe der Schutzregeln annimmt (`respond` kennt
  `--guardrails`) — nicht gefunden, nicht gebaut.
- Die Exit-Codes im **negativen** Fall (`fm license --status` ohne Annahme,
  `fm available` ohne Modell) — auf diesem Rechner nicht herstellbar; ein
  Code ungleich 0 heißt hier „nicht aktiviert".
- Ob das Modell auf einem anderen Mac dieselben Ablehnungen trifft.
