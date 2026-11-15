# 0179 — Die WebView läuft unter einer gemessenen Schutzrichtlinie

**Stand:** angenommen am 2026-10-07 · SEC-17
(#2164).

## Die Frage

**Die ausgelieferte Seite trug keine Content Security Policy.** Weder
`frontend/index.html` noch die Hülle setzte eine; ein eingeschleustes Skript
hätte alles gedurft, was die Seite darf — jede Bindung rufen, jedes Bild des
Bildwegs lesen. **Eine belegte Lücke, durch die es käme, gibt es nicht.** Die
Frage war, ob eine zweite Schicht greifen kann, ohne dem Fenster etwas
wegzunehmen — Tisch, Kartenbild, Dialoge, Erscheinungsbilder.

## Die Entscheidung

**Eine Richtlinie als Kopfzeile der Assetleitung**, nur im Fensterbau. Was
sie erlaubt, steht in der Hülle (`internal/fenster/schutznaht.go`); die Weiche,
die sie an jede Antwort setzt und Berichte annimmt, steht in
`internal/vorschauweg/schutzweiche.go` — ein Handler braucht `net/http`, und
dieses Blatt ist das eine, das es dafür kennen darf
(`internal/schichten/netz_test.go`, `keineverbindung_test.go`):

| Anweisung | Quellen | Warum |
| --- | --- | --- |
| `default-src` | `'self'` | alles Übrige nur aus dem Bündel |
| `script-src` | `'self'` | **kein Inline, kein `eval`** — der Tisch lädt dafür `pixi.js/unsafe-eval`: PixiJS 8.20.1 baut seine Uniform- und Puffer-Abgleiche mit `new Function` und bricht ohne `unsafe-eval` ab (`AbstractRenderer._unsafeEvalCheck`); das Modul aus demselben Paket ersetzt sie durch Polyfills (`lib/unsafe-eval/init.mjs`). Gemessen: ohne die Zeile zeichnete der Tisch unter der Richtlinie nicht |
| `style-src` | `'self' 'unsafe-inline'` | Mantine und `react-style-singleton` setzen `<style>` zur Laufzeit (gemessen: `style-src-elem inline`) |
| `img-src` | `'self' data:` und drei fremde Ursprünge | PixiJS prüft Bildformate mit eingebetteten AVIF- und WebP-Bildern (gemessen: `img-src data`). **Dazu die Vorschaubilder der Kartensuche**, deren Adresse der Dienst nennt: `https://cards.scryfall.io`, `https://gatherer.wizards.com` und — Ziel seiner Weiterleitung — `https://gatherer-static.wizards.com` (unten) |
| `worker-src` | `'self' blob:` | PixiJS lädt Texturen in einem Worker aus `blob:` (gemessen: `worker-src blob`) |
| `connect-src`, `font-src`, `media-src` | `'self'` | |
| `object-src`, `base-uri`, `form-action` | `'none'` | |
| `report-uri` | `/csp-bericht` | ein Verstoß landet im Protokoll — mit Anweisung, Ursprung und Stelle, **ohne Pfad**. Jedes Feld schreibt die Seite: aus der Anweisung kommt nur ein Name aus CSP Level 3, Abschnitt 6, aus einem abgewiesenen Wert nur ein Schlüsselwort der CSP oder ein Ursprung, aus der Quelle nur der Name einer Datei, die **im ausgelieferten Bündel liegt** — ohne Abfrage und Anker —, sonst `«Datei»` oder der Ursprung einer fremden |

**Gemessen, nicht vermutet.** Die Richtlinie begann streng — nur `'self'` —,
und die Hüllenprobe `sec17` fuhr sie zuerst **beobachtend**
(`Content-Security-Policy-Report-Only`, `MSE_CSP_BEOBACHTEN=1`, nur im
Abnahmebau). Jede Quelle außer `'self'` steht in der Tafel, weil ein
beobachtender Lauf sie angestrichen hat.

**Durchgesetzt** weist WebKit über Wails' eigenes Schema ein eingeschleustes
Inline-Skript, `eval` und ein fremdes Skript ab und meldet alle drei; Tisch,
Kartenbild, hell und dunkel arbeiten (`docs/belege/huellenprobe/sec17.md`).

**Und jede Hüllenprobe prüft sie mit**: `scripts/huellenschutz.mjs` liest
nach jedem Drehbuch **jedes** Protokoll unter seinen Orten — auch an einem
eingestellten Ort, auch gedreht, in Text- wie JSON-Form — und streicht jeden
Verstoß an, den es nicht ausdrücklich erwartet, und jeden Bericht, der sich
nicht lesen ließ. Ein Weg, der eine Ressource braucht, die die Richtlinie
nicht erlaubt, wird damit in dem Drehbuch rot, das ihn geht.

## Nachtrag vom selben Tag: die Bilder der Kartensuche

**Der Zweitblick über Phase H fand einen Weg, den kein Drehbuch ging**: die
Kartensuche zeigt je Treffer ein kleines Bild, dessen Adresse der Dienst
nennt, und die WebView holt es selbst. Die erste Fassung nahm ihr diese
Bilder — am nativen Fenster gemessen: beide 0 px breit, das Protokoll nannte
`img-src` gegen `https://cards.scryfall.io` und `https://gatherer.wizards.com`
(`sec17bilder`, erster Lauf).

**Gemessen an den echten Diensten am 2026-10-07**, je eine Suche: Scryfall
nennt 24 von 24 Drucken auf `https://cards.scryfall.io`; magicthegathering.io
nennt `http://gatherer.wizards.com/…`, und das antwortet mit 301 auf sein
`https`-Gegenstück, das wiederum mit 308 auf
`https://gatherer-static.wizards.com` weiterleitet. Von MTGJSON und Scrydex
übernimmt die Suche kein Bild. **Erlaubt sind genau diese drei Ursprünge**,
nicht `https:` und nicht `*.wizards.com`; `mtgapi` nennt Gatherer seither über
`https` — Klartext ließe die Richtlinie nicht durch, und App Transport
Security auch nicht. Der Datenschutzabschnitt der Hilfe nennt die drei.

**Verworfen** wurde, die Bilder über Go zu holen und über den Bildweg zu
reichen: das wäre ein Weg, der fremde Adressen auf Zuruf der Seite abruft —
genau die Fähigkeit, die eine Richtlinie einer Seite nehmen soll.

## Was bewusst nicht entschieden ist

- **`style-src 'unsafe-inline'`** ist ein Preis. Ein eingeschleustes Stück
  CSS kann die Seite verändern, und es erreicht nach draußen **die drei
  Bildursprünge** — als Bildanfrage mit einem Pfad seiner Wahl, und damit
  kann es dorthin tragen, was es auf der Seite liest. Schriften und
  Verbindungen bleiben auf `'self'`. Hier stand bis zum Nachlauf über
  Phase H, es erreiche nach draußen nichts; das galt nur, solange Bilder auf
  `'self'` standen. Der strengere Weg
  ist eine **Nonce je Dokument**: die Hülle setzte sie in `index.html` und in
  die Kopfzeile, Mantine nähme sie über `getStyleNonce`,
  `react-style-singleton` über `__webpack_nonce__`. Das ist ein eigener
  Schritt.
- **Der Servermodus** (`make dev-browser`, `-tags server`) trägt keine
  Richtlinie: er liefert dieselbe Seite an einen Browser, und dort
  entscheidet die Wache, wer überhaupt fragen darf.
- **Der Abnahmebau** erlaubt zusätzlich den Rückkanal der Hüllenprobe
  (`connect-src http://127.0.0.1:<WAILS_MCP_PORT>`) — sonst erreichte keine
  Antwort die Probe. Er misst damit dieselbe Richtlinie bis auf diese eine
  Quelle.

## Verworfen

- **Ein `<meta http-equiv>` in `index.html`**: es kann nicht beobachten
  (`Report-Only` gibt es nicht als Meta), und `vite dev` setzt sein Vorspiel
  inline — der Entwicklungsweg läge lahm.
- **`'unsafe-eval'` für den Tisch**: PixiJS bringt die Polyfills im selben
  gepinnten Paket mit (`pixi.js/unsafe-eval`); eine Erlaubnis für jedes
  `eval` der Seite wäre der Preis für eine Bequemlichkeit gewesen.
