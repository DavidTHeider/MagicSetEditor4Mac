# Architektur

Verbindlich ist `AGENTS.md` §2. Diese Datei führt aus, *warum* der Schnitt so
liegt, und beschreibt das Datenmodell.

---

## Die Schichten

```
frontend/            React, TypeScript, Vite, Tailwind
    ↕ Wails 3        internal/fenster/ und frontend/src/anbindung.ts
internal/app/        Bindings-Fassade: DTOs, Fehlerübersetzung, Kontext
internal/service/    Anwendungsfälle, Transaktionsgrenzen, fachliche Regeln
internal/storage/    SQLite-Repositories, die Grundlinie des Schemas und Wanderungen danach
internal/domain/     Entitäten, Wertetypen, Validierung
```

**Abhängigkeiten zeigen ausschließlich nach innen.** `domain` kennt niemanden,
`app` kennt alle.

### Die Hülle

**Seit ENG-02 kennt `internal/app` Wails nicht mehr**
([0132](entscheidungen/0132-wails-3-als-huelle.md)). Die Anwendung beschreibt,
was das Fenster tun soll — eine Menüleiste als eigenes Modell
(`menuebaum.go`), Nähte für Meldung, Dialog und Zwischenablage —, und
`internal/fenster` setzt es auf Wails 3 um. Nur `main.go` kennt die Hülle, und
in ihr greift allein `wails.go` ins laufende Fenster.

**Warum eine eigene Schicht und nicht Wails in `app`**: unter Wails 2 standen
die Rufe in sechs Dateien von `internal/app` (0124), und jeder Lauf über die
Menüleiste las Wails' Verbünde. Wails 3 gibt einen Menüeintrag nicht mehr so
her — und baut unter Linux nur mit cgo; eine Anwendungsschicht, die es
importierte, fiele aus dem Kreuzbau der Pipeline. Auf der anderen Seite der
Naht stellt `frontend/src/anbindung.ts` dieselbe Gestalt her, die Wails 2
hinterließ (`window.go.app.Bindungen`, `window.runtime`), und Fassade,
Fenster und Läufe bleiben, wie sie waren.

### Zwei Composition Roots

Seit Phase 9 gibt es **zwei** Orte, an denen verdrahtet wird — und beide stehen
unter derselben Regel, geprüft von `internal/schichten`:

| Programm | Composition Root | Darf kennen |
| --- | --- | --- |
| Das Fenster (`main.go`) | `internal/app/app.go` | `app` und `fenster`; darin `storage` **nur** in `app.go` — vom Wächter erzwungen |
| Der MCP-Server (`cmd/mse-mcp/`) | `cmd/mse-mcp/server.go` | `app`, `mcp`, `storage` — für das **ganze** Paket |

**Die beiden sind nicht gleich streng bewacht, und das gehört gesagt.** Für
`internal/app` verengt `nurInDatei` die Kante nach `storage` auf `app.go`; für
`cmd/mse-mcp` gilt die Erlaubnis dem ganzen Paket. Dass der Composition Root
dort in `server.go` steht, ist Ordnung und keine geprüfte Zusage — das Paket
besteht heute aus zwei Dateien, und `main.go` enthält nur den Prozesseinstieg.

Der zweite ist ein eigener Prozess über stdio
([0010](entscheidungen/0010-mcp-als-eigener-prozess.md)). **Er ist kein
Netzwerkdienst**: kein Port, keine Adresse, nichts, was wartet.

**Der Wächter kennt `cmd/` seit dem Zweitblick über Phase 9.** Vorher las er nur
`internal/` und `main.go` — ein zweiter Composition Root, den niemand prüft, ist
genau die Stelle, an der eine Schicht still umgangen wird.

Repository-**Schnittstellen** stehen in `internal/service`, **implementiert**
werden sie in `internal/storage`, ohne dass `storage` `service` importiert. Der
Nutzen ist, dass `service` sagt, was es braucht, statt zu nehmen, was da ist.

**Das trägt nur, solange die Typen in den Signaturen beiden gehören.** Gos
strukturelle Typisierung gilt für **Schnittstellen** — nicht für benannte Typen
in Parameterpositionen. `[]storage.Rohblock` erfüllt eine Zusage über
`[]service.Rohblock` nie, auch wenn beide Verbunde Feld für Feld gleich sind.

Hier stand das Gegenteil, und es hat eine Schnittstelle gekostet, die niemand
erfüllen konnte — gefunden erst vom Integrationstest im Composition Root, weil
beide Fachtestebenen gegen ihren je eigenen Typ liefen und grün waren. Was über
die Naht geht, gehört deshalb nach `internal/domain`.

## Die Fachbibliotheken

Neben den Schichten stehen Pakete, die keine Schicht sind. Sie sind **Blätter**
— sie kennen `domain` und sonst nichts aus diesem Projekt:

| Paket | Aufgabe | Kennt |
| --- | --- | --- |
| `internal/bezug` | Vorlagenpakete über HTTP beziehen und sicher auspacken | `domain` |
| `internal/dateisperre` | Die beratende Sperre auf einer Datei — `flock`, eine Nachbardatei, auf Windows nichts. Zwei Verbraucher: der Bestand über einen Tausch, das Hostbudget über Prozessgrenzen | **nichts** |
| `internal/prozessspeicher` | Der physische Fußabdruck dieses Programms **und seiner WebKit-Prozesse** — gezählt wird, wofür macOS diesen Prozess als **verantwortlich** führt, wie die Aktivitätsanzeige. Fehlt die nicht öffentliche Zuordnung, heißt der Webspeicher „nicht gemessen" (OPS02, #2348) | **nichts** |
| `internal/msepaket` | MSE-Paketformate lesen und schreiben | `domain` |
| `internal/laufbeleg` | Belege für Messläufe — liest den Strom von `go test -json`, trennt **Ausführung** und **Ergebnis** und schreibt `docs/belege/<lauf>.json` samt Stand, Werkzeugen, Eingabehashes und Artefakten. **Kein Teil der Anwendung**: ein Werkzeug der Entwicklung, und kein ausgeliefertes Programm kennt es (QA-01, #1733) | **nichts** |
| `internal/msescript` | Zerteiler und Auswerter der eingebetteten Sprache | **nichts** |
| `internal/mcp` | Das Model Context Protocol über stdio samt seinen Werkzeugen | `domain` |
| `internal/mcpnaehte` | Die Nähte des MCP-Servers: 42 dünne Adapter von `mcp.Naehte` auf `app` — seit U12 ein Paket, damit der Server **und** die Kommandozentrale im Fenster dieselben Werkzeuge füllen (#1423) | `app`, `domain`, `mcp`, `service`, `storage` |
| `internal/mtgjson` | MTG JSON lesen und schreiben, samt der Zuordnung auf Kartenfelder | `domain` |
| `internal/cardtrader` | **CardTrader, lesend** — Spiele, Erweiterungen, Blueprints mit ihren Fremdkennungen und die Angebote des Marktplatzes (A4, #1443). Die Positivliste führt **vier** Adressen, alle mit `GET`; Kontofunktionen sind Sache von T11 ([0095](entscheidungen/0095-was-cardtrader-sagt-und-was-davon-gilt.md)) | `domain`, `netz` |
| `internal/mtgjsonpreise` | die **Preise** des Dienstes mtgjson.com — Bulkbezug, Grenzen, Zuordnung auf die neutrale Preisreihe (A3, #1442). **Nicht dasselbe wie `internal/mtgjson`**: dort ein Dateiformat, hier ein Dienst mit eigenem Vertrag und eigenen Grenzen | `domain`, `netz` |
| `internal/ollama` | die **lokale KI-Laufzeit** — installierte Modelle, ihre belegten Fähigkeiten, Unterhaltung mit fester Form und Einbettung (A7, #1446). Die Positivliste führt **fünf** Adressen; es gibt kein `/api/pull`, kein `/api/create`, kein `/api/delete`, und ein Wächter liest das aus dem Quelltext nach. Was es **nicht** belegt, steht in [0098](entscheidungen/0098-was-ollama-belegt-und-was-nicht.md): dass auf 11434 wirklich Ollama liegt | `domain`, `netz` |
| `internal/anthropic` | **Anthropic** — Modelle mit ihren **belegten** Fähigkeiten, Unterhaltung in fester Form und Bilder (A8, #1447). `GET /v1/models` nennt `structured_outputs` und `image_input` **selbst**; eine abgeschriebene Modellliste wäre eine Behauptung mit Verfallsdatum. Sieben Abschlussgründe werden zu **vier** Klassen — fertig, unvollständig, abgelehnt, unbekannt ([0099](entscheidungen/0099-was-anthropic-belegt.md)) | `domain`, `netz` |
| `internal/xai` | **xAI** — Unterhaltung in fester Form über **`/v1/chat/completions`** (A10, #1449). **`GET /v1/language-models` belegt zwei von drei**: `input_modalities` und `output_modalities` stehen da — Text und Bildlesen —, die **feste Form** nicht; die kommt aus der Messung von A9 ([0101](entscheidungen/0101-was-xai-belegt.md)). Es kennt einen **vierten Ausgang** (`end_turn`), nennt den Preis jedes Rufs in **Ticks** (`cost_in_usd_ticks`), und derselbe Anbieter hat an seinen zwei Endpunkten **zwei verschiedene** Bildformen. `/v1/responses` bleibt draußen, weil es Antworten **30 Tage** hält |
| `internal/applefm` | **Apples Gerätemodell** über `fm serve` (AI-05, #2042) — **nur über einen eigenen Unix-Socket** in einem `0700`-Ordner, nie über einen Port: mit einem Socketpfad über 104 Bytes lauscht `fm` still auf allen Schnittstellen, deshalb wird der Pfad geprüft. Immer `stream: false`, Schema über `response_format`; **eine Ablehnung ist ein eigener Ausgang** in zwei Gestalten (HTTP 500 mit der Meldung zu den Schutzregeln, und ein bekannter Satz im Inhalt). „Aktiviert" heißt drei Exit-Codes; **die Lizenz nimmt ein Mensch an** ([0166](entscheidungen/0166-was-fm-serve-belegt.md)) |
| `internal/openai` | **OpenAI** — Unterhaltung in fester Form über die **Responses-API** (A9, #1448). **`GET /v1/models` belegt nichts**: ein Modellobjekt trägt `id`, `object`, `created` und `owned_by` — kein `capabilities` wie bei Anthropic. Daraus entsteht deshalb **kein** Katalogeintrag, sondern je Modell ein Befund ([0100](entscheidungen/0100-was-openai-nicht-sagt.md)). Der Ausgang steht an **drei** Stellen — `status`, `incomplete_details.reason` und ein Inhaltsblock `refusal` —, ein **429 heißt dreierlei**, und `store` steht ausdrücklich auf `false`, weil die Vorgabe der Schnittstelle `true` ist |
| `internal/netz` | Der Transport zu fremden Diensten — drei Fristen, Kopfzeilen, ausgehende Drossel, 429 mit `Retry-After` | `domain` |
| `internal/pdf` | PDF aus Bildern und Linien schreiben — der Druckbogen und sonst nichts | **nichts** |
| `internal/render` | Rasterung einer Karte aus einem Zeichenplan | `domain` |
| `internal/rest` | Der Dienst auf der Rückschleife — Marke, Herkunft, Drossel, Endpunkte | `domain` |
| `internal/sammelliste` | Eine eingefügte Kartenliste in Zeilen zerteilen — Menge, Name, Set, Nummer, Ausführung, die **Überschrift**, unter der eine Zeile steht, und je unlesbarer Zeile ein Befund mit ihrer Nummer. Die **Zuordnung** auf einen Druck braucht Scryfall und steht in `service`; was aus einer Überschrift wird, entscheidet `domain` — für die Sammlung nichts, fürs Deck der Bereich (T8, #1473) |
| `internal/mtgapi` | Karten von magicthegathering.io lesen — das fremde DTO und die Zuordnung; kein Zugang nötig, und als einziger der drei neuen Dienste gegen sein Original geprüft ([0056](entscheidungen/0056-drei-weitere-kartendienste.md)) | `domain`, `netz` |
| `internal/mtggraphql` | Karten über die GraphQL-Schnittstelle von MTGJSON — Marke im Kopf, `errors` im Rumpf als Fehler; **nie gegen den echten Dienst gelaufen**, weil es hier keine Patreon-Marke gibt ([0056](entscheidungen/0056-drei-weitere-kartendienste.md)) | `domain`, `netz` |
| `internal/scrydex` | Karten bei Scrydex lesen — **zwei** Kopfzeilen (`X-Api-Key` und `X-Team-ID`), nur Magic von sechs Spielen; **nie gegen den echten Dienst gelaufen**, weil es hier kein Konto gibt ([0056](entscheidungen/0056-drei-weitere-kartendienste.md)) | `domain`, `netz` |
| `internal/anmeldung` | Anmeldung über einen fremden Anbieter — OAuth 2.0 mit PKCE und `state`, OIDC mit `nonce`; vier Anbieter in **zwei** Gestalten, und der Horcher auf der Schleifenadresse, den Google für Desktop-Anwendungen verlangt ([0066](entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md)). **Kennungen liefert dieses Projekt nicht mit**: wer sich anmelden will, registriert seine eigene Anwendung | `domain` |
| `internal/kennwort` | Kennwörter hüten und wiedererkennen — **Argon2id** mit m=19456, t=2, p=1, dem Mindestmaß aus dem OWASP Password Storage Cheat Sheet (gelesen am 2026-09-05). Die gehütete Form trägt ihre eigenen Zahlen, damit ein späteres Anheben die alten Kennwörter nicht wertlos macht. **Kein Pepper**, und das ist eine Entscheidung: er läge im Schlüsselbund und sperrte den Eigentümer aus, sobald er seinen Bestand auf einem neuen Mac zurückspielt. Gegen einen mitgelesenen Bestand hilft die Verschlüsselung ([0066](entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md)). **Die gelesene Form hat Schranken**, und die sind kein Formalismus: `argon2.IDKey` nimmt die genannte Menge Speicher wirklich, und bei `t=0` oder `p=0` **stürzt es ab** — eine von Hand geänderte Bestandszeile genügte (nachgestellt am 2026-09-05) | `domain` |
| `internal/vcard` | **vCards lesen** — ID-04 (#1703), [0137](entscheidungen/0137-kontakte-kommen-als-vcard.md). Name, Kennung, E-Mail, Telefon und ein Bild, nach RFC 6350 (4.0) und RFC 2426 (3.0, die Fassung von Apple Kontakte): Faltung, Maskierung, Base64 und `data:`. **Ein Bild als Adresse wird nicht geholt**, sondern genannt; Quoted-Printable aus 2.1 ebenso. Eine kaputte Karte hält die anderen nicht auf | — |
| `internal/totp` | Zeitbasierte Einmalkennwörter nach **RFC 6238** — reines Go über `crypto/hmac`. `T = Unix-Zeit / 30 s`, dynamische Kürzung nach RFC 4226 §5.3, sechs Ziffern; SHA-1, SHA-256 und SHA-512. **Das Prüffenster ist eine Entscheidung dieses Prüfers** und keine Vorgabe der RFC (§5.2 lässt es offen): **ein** Schritt in jede Richtung. Geprüft gegen alle **achtzehn** Vektoren aus Anhang B — eine selbst gerechnete Erwartung prüfte nur, dass dieselbe Rechnung zweimal dasselbe gibt. **Wo ein Geheimnis liegt, steht hier nicht**: es gehört in den Schlüsselbund | `domain` |
| `internal/peer` | Die Identität dieses Programms gegenüber anderen — Werkstück N1 von Phase 5.4 (#1454). Ein **Ed25519**-Paar; seine Kennung ist die eines libp2p-Peers, Byte für Byte nach `libp2p/specs`, `peer-ids/peer-ids.md` (nachgelesen am 2026-09-14): Protobuf `PublicKey{Type: Ed25519, Data}`, **identity**-Multihash (unter 42 Byte wird nicht gehasht), CIDv1 mit `libp2p-key` (0x72), base32 ohne Auffüllung, Multibase `b`. **Die Kennung trägt damit den ganzen Schlüssel** — zwei Kennungen zu vergleichen heißt, zwei Schlüssel zu vergleichen, und der Fingerabdruck ist eine **Lesehilfe** über 128 Bit und kein zweiter Nachweis. **Wo das Geheimnis liegt, steht hier nicht**: es gehört in den Schlüsselbund. **Ungeprüft bleibt**, dass eine echte libp2p-Fassung dieselbe Kennung errechnet. Hier stand „das gehört zu N2"; seit [0113](entscheidungen/0113-kein-libp2p-im-lan.md) vom 2026-09-14 kommt libp2p nicht mehr vor, und damit hat dieser Abgleich **keinen geplanten Ort mehr**. Für die Richtigkeit im Betrieb ist er auch nicht nötig: beide Seiten dieses Programms rechnen dieselbe Gestalt, und die ist aus der Spezifikation nachgelesen. Er wäre nötig, wenn libp2p je dazukäme — dann zuerst | `domain` |
| `internal/biometrie` | Touch ID und Face ID über `LocalAuthentication` — der **zweite** Weg zum zweiten Faktor. Die Richtlinie ist `…AuthenticationWithBiometrics` **ohne** Rückfall auf das Anmeldekennwort des Rechners: wer am Rechner angemeldet ist, kennt es, und ein zweiter Faktor, den der Angemeldete ohnehin hat, ist keiner. **Der Weg geht über [`tmc/apple`](https://github.com/tmc/apple) und `purego`** — cgo-frei, MIT, aber **v0.6.18** und darunter `purego v0.11.0-**alpha.6**` (gemessen am 2026-09-05; seit dem 2026-10-11 `purego` v0.11.1, eine Endfassung vor 1.0 — #2305). Zwei unfertige Schichten unter einem Sicherheitsmerkmal: deshalb steht jeder Ruf hinter einem Riegel, der einen `panic` in eine Antwort verwandelt, und deshalb ist **TOTP** der Weg, der immer geht. **Gemessen** ist die Verfügbarkeit (`canEvaluatePolicy` gibt `true`, `BiometryType` meldet `TouchID`); **nicht gemessen** ist, dass eine Abfrage einen Finger erkennt — dafür müsste jemand den Sensor berühren | `domain` |
| `internal/schluesselbund` | Zugangsdaten fremder Dienste dort ablegen, wo macOS sie hütet — **nicht** im Bestand, denn den gibt man weiter ([0056](entscheidungen/0056-drei-weitere-kartendienste.md)) | `domain` |
| `internal/schriften` | Finden von Schriftdateien auf dem Rechner | `render` |
| `internal/scryfall` | Kartendaten von Scryfall lesen — das fremde DTO und die Zuordnung; Pflichtfelder streng, Zusatzfelder tolerant | `domain`, `netz` |
| `internal/aktualisierung` | fragt GitHub nach der neuesten Freigabe — dasselbe Muster, dieselben drei Schichten. Die Vorgabe ist **aus**: ohne Einstellung fragt nichts (#1100) | `domain`, `netz` |
| `internal/sicherung` | Das Sicherungsarchiv schreiben und lesen — ZIP samt Manifest aus Fassung, Zeitpunkt, Größen und SHA-256; sicheres Auspacken ohne Pfadausbruch | `domain` |
| `internal/leitung` | Nachrichten über einen Strom tragen — Rahmung mit Längenpräfix, Größengrenze, und eine Schleife für zwei Anwendungen im selben Prozess. **Dazu der Weg über TCP**: `Waehle` und `Horcher`, beide mit TLS 1.3 und einem selbst unterschriebenen Zeugnis, dessen Schlüssel die Peeridentität **ist**. Geprüft wird gegen die erwartete Kennung und nicht gegen eine Zertifizierungsstelle; die Kanalbindung des Handschlags kommt aus `ExportKeyingMaterial` (RFC 5705) und ist damit geheim statt nur eindeutig. **Der Grund steht in [0113](entscheidungen/0113-kein-libp2p-im-lan.md)**: `internal/peer` unterschreibt und verschlüsselt nicht, und ohne libp2p gingen Kartenbestände im Klartext hinaus | `domain`, `peer` |
| `internal/austausch` | Der **portable Umschlag** (`.msepackage`) — Karten, Sets, Vorlagen, Decks und Sammlungssichten in einem ZIP mit **zwei** Manifesten: das äußere sagt, ob der Behälter heil ist, das innere, was fachlich darin steht — und es ist **unterschrieben**. Die **Quarantäne steht vor dem Auspacken**: Größen, Zahl, Tiefe, Namen, Abhängigkeiten. **Signiert heißt bekannte Herkunft, nicht sicherer Inhalt** — deshalb führt ein empfangenes Skript nichts aus. Es kennt `internal/sicherung` — die **dritte** benannte Ausnahme dieser Art, aus demselben Grund wie `svg` zu `render` | `domain`, `peer`, `sicherung` |
| `internal/fernleitung` | Die Verbindung über **fremde Netze** — tailcat, also Tailscales Datenebene ohne deren Kontrollebene: WireGuard, magicsock, DERP als Rendezvous und Rückfall, **ohne Konto**. **Nur auf Aufruf**, wie der Horchposten. Es kennt `internal/leitung` — die **zweite** benannte Ausnahme dieser Art, aus demselben Grund wie `svg` zu `render`: über dem Tunnel liegt derselbe gesicherte Weg wie im eigenen Netz. **Die Einladung trägt eine Fassung**, denn tailcats Drahtformat darf sich ändern ([0114](entscheidungen/0114-tailcat-gepinnt.md)) | `domain`, `leitung`, `peer` |
| `internal/entdeckung` | Gegenüber im selben Netz finden — mDNS über `hashicorp/mdns`, Dienstart `_mse4mac._tcp`. **Nur auf Aufruf**: ohne `Melde` kündigt sich nichts an, ohne `Suche` wird nicht gefragt. Rechnername und Adressen setzt das Paket **selbst** — sonst nähme die Bibliothek `os.Hostname()` und schlüge ihn im DNS nach. Eine Antwort ohne gültige Peerkennung wird verworfen und **gezählt** | `domain`, `peer` |
| `internal/svg` | Eine Karte als SVG aus demselben Zeichenplan schreiben | `render` |
| `internal/vorschauweg` | Gerasterte Kartenbilder an das eigene Fenster ausliefern — über die Assetleitung, die Wails ohnehin betreibt, statt als Base64 durch die JSON-Naht (#1270) | **nichts** |

`pdf` kennt aus demselben Grund nichts: es setzt Bilder und Linien auf Seiten,
und was eine Karte ist, weiß es nicht. Die Millimeter einer Kartengröße rechnet
der Bogenleger in `service` um ([0013](entscheidungen/0013-eigener-pdf-schreiber.md)).

`vorschauweg` kennt **nicht einmal `domain`**, und das ist die knappste Stelle
der Tabelle: es nimmt Bytes entgegen, gibt eine Adresse zurück und liefert sie
wieder aus. Ein fachlicher Fehler kommt dabei nicht vor — ein Bild ist da oder
verdrängt, und das sagt ein 404.

**Es ist das einzige Blatt, das `net/http` einführt**, und es steht deshalb in
der Ausnahmeliste von `internal/schichten/netz_test.go`. Eine Verbindung macht
es nicht auf: kein Klient, kein Lauscher, keine Adresse — nachgeprüft von
`internal/vorschauweg/keineverbindung_test.go`, nicht behauptet. Dass der Weg
ein eigenes Paket ist statt einer Datei in `internal/app`, hat genau diesen
Grund: so hat die Bindungsfassade den Weg nach draußen **nicht**.

`svg` kennt `render` — **dieselbe Richtung wie `schriften`**, und aus verwandtem
Grund: es liest einen fertigen Zeichenplan und ist damit dessen **zweiter**
Verbraucher. Genau das ist die Probe auf die Trennung eine Zeile weiter unten —
trägt sie auch, wenn jemand anderes als der Rasterer den Plan liest? Umgekehrt
ginge nicht: ein Rasterer, der von SVG wüsste, wäre die Vermischung, die die
Regel verhindert.

**Und `service` kennt `svg`**, so wie es `pdf` kennt: SVG ist ein
**Ausgabeformat der Ausfuhr**, und `Bildausfuhr` entscheidet, welches gewünscht
ist. Ohne diese Kante stand das Paket mit Tests und voller Abdeckung da, und
niemand konnte eine SVG erzeugen
(#852).

`msescript` kennt bewusst auch `domain` nicht. Es ist ein Sprachwerkzeug: Eingabe
ist Quelltext und eine Umgebung aus Werten, Ausgabe ist ein Wert. Eine
Kartenentität darin machte den Interpreter unprüfbar — man könnte ihn nicht mehr
gegen eine Tabelle aus Ausdruck und Ergebnis fahren, sondern bräuchte für jeden
Fall eine Karte.

## Die Naht, die am leichtesten verrutscht

MSE-Vorlagen sind **skriptgesteuert**. Ein Stil sagt nicht „das Symbol sitzt bei
x=30", sondern „x ist, was dieser Ausdruck ergibt". Daraus folgt eine Reihenfolge,
und sie ist die wichtigste Entscheidung dieses Projekts:

```
msepaket   liest den Stil            →  Felder mit Skripten als Text
msescript  wertet die Skripte aus    →  Werte
service    setzt beides zusammen     →  Zeichenplan
render     rastert den Plan          →  Bild
```

**`render` sieht kein Skript.** Wer eine Skriptauswertung im Renderer findet, hat
die fehlende Naht gefunden. Der Grund ist prüftechnisch: ein Renderer, der
auswerten kann, lässt sich nur noch über ganze Vorlagen prüfen — und ein
Goldbildvergleich, der bei jeder Skriptänderung anders ausfällt, sagt nichts
mehr darüber, ob das Rastern stimmt.

Der Zeichenplan ist damit auch die Stelle, an der man einen Fehler verortet: sieht
der Plan richtig aus und das Bild falsch, liegt es an `render`; sieht der Plan
falsch aus, an `msescript` oder an der Vorlage.

## Warum SQLite und nicht die Datei

MSE2 hält ein Set im Arbeitsspeicher und schreibt es als ZIP. Das ist einfach und
hat drei Grenzen, die man erst bei großen Sets merkt: keine Suche über
Kartentexte, keine Massenbearbeitung mit Rücknahme, und ein Rückgängig, das mit
dem Fenster endet.

Hier ist SQLite das **Arbeitsformat**, die MSE-Datei das **Austauschformat**:

```
.mse-set  ──import──►  SQLite  ──export──►  .mse-set
```

Import beim Öffnen, Export beim Speichern. Was dazwischen liegt — Suche,
Massenänderung, Verlauf — ist Datenbankarbeit.

Der Preis ist eine **Rundlaufzusage**: was hineingeht, muss unverändert wieder
herauskommen, auch wenn wir ein Feld nicht verstehen.

Sie zerfällt in zwei Teile mit verschiedenen Zeitpunkten
([0002](entscheidungen/0002-sqlite-als-arbeitsformat.md)): die **Daten**
verlustfrei mitzuführen ist Phase 2 und nicht nachrüstbar; der **Rundlauftest**
über echte Pakete ist Phase 6, weil es vorher keinen Schreiber gibt, gegen den
er laufen könnte.

Seit Phase 2 merkt sich jeder Knoten seine **Rohzeile** samt Zeilenende, und
`raw_blocks` bekommt sie wörtlich statt aus Schlüssel und Wert rekonstruiert.

Noch nicht mitgeführt ist, was **vor** diesem Mechanismus verloren geht: ein
Schlüssel, den der Leser kennt, schließt seinen Bereich vom Rohweg aus, auch
wenn der Leser dieses Vorkommen gar nicht verwertet — ein zweites `game:`, ein
zweites `set info:`, Kommentare auf der Wurzelebene, der skalare Rest eines
`card: irgendwas`. Vier Fälle, aufgezählt in
#88.

Der Treiber ist `modernc.org/sqlite` — SQLite ohne cgo. Das hält den Bau frei von
einem C-Compiler und die Kreuzkompilierung nach Windows und Linux (§12) möglich,
ohne dass jemand dort etwas einrichtet.

### Pragmas stehen in der Datenquelle, nicht in einer Anweisung danach

**Ein SQLite-Pragma gilt je Verbindung.** `database/sql` führt einen Vorrat und
öffnet nach, wann es will; wer ein Pragma über `db.ExecContext` setzt, trifft
**eine** Verbindung — die erste, die gerade frei war. Jede spätere kommt ohne.

Deshalb stehen `busy_timeout` und `foreign_keys` als `_pragma=` in der
Datenquelle: `modernc.org/sqlite` führt sie aus, sobald es **irgendeine**
Verbindung öffnet.

**Das war drei Phasen lang falsch, und kein Test hat es gesehen.** Alle liefen
gegen eine Datenbank im Arbeitsspeicher, und die hält `Oeffnen` bei genau einer
Verbindung — sonst wäre ihr Inhalt mit der ersten geschlossenen fort. Eine
Zusage, die nur auf einer Verbindung gilt, sieht kein Test, der eine benutzt.

`foreign_keys` ist dabei das schwerste Stück: `Kartenschreiber.FeldSetzen`
schließt aus der verletzten Fremdschlüsselbedingung auf `ErrNotFound`. Ohne das
Pragma meldet dieselbe Anweisung **Erfolg und schreibt nichts**.

Die Ausnahme ist `journal_mode`: es ist eine Eigenschaft der **Datei** und
bleibt, einmal gesetzt. Es steht deshalb als Anweisung nach dem Öffnen.

Gehalten wird die Regel von `TestDiePragmasGeltenAufJederVerbindung` — an einer
echten Datei, mit vier gleichzeitig offenen Verbindungen
(#157).

## Datenmodell

Der Kern, wie er in Phase 2 entsteht. Namen englisch (`AGENTS.md` §9), weil sie
an den MSE-Bezeichnern hängen.

| Tabelle | Trägt |
| --- | --- |
| `sets` | ein geöffnetes Set: Name, Spiel, Stil, Pfad der Herkunftsdatei |
| `cards` | eine Karte, Reihenfolge im Set, Zeitstempel |
| `card_values` | Feldwerte je Karte — Feldname und Wert als Text, weil MSE-Felder frei definiert sind |
| `set_values` | Feldwerte des Sets selbst (Setname, Symbol, Copyright …) |
| `styling_values` | die Styling-Optionen eines Stils — gefüllt seit #41; welche es gibt, sagt das Stilpaket (`styling field:`), nicht wir. **Sie liegen über dem Block `styling:` der Datei** (in `set_values`) — für die Optionsliste wie für die Rasterung, seit dem 2026-10-11 (#2247, `internal/service/stilwirkung.go`); bis dahin las die Rasterung nur den Block und die Liste nur diese Tabelle |
| `packages` | installierte Pakete: Art, Name, Version, Pfad und **Herkunft** — gefüllt seit #54 durch den Abgleich beim Öffnen der Paketseite |
| `package_deps` | Abhängigkeiten zwischen Paketen — aus `depends on:`, mitgeschrieben vom selben Abgleich |
| `raw_blocks` | roh mitgeführte Bereiche, die wir nicht deuten — für die Rundlaufzusage |
| `history` | Änderungsverlauf für Rückgängig über Sitzungen hinweg |
| `wert_herkunft` | woher der **aktuelle** Wert eines Feldes stammt — die Kartenseite bei Scryfall, wenn er von dort kam. Gepflegt in **derselben Transaktion** wie der Wert: von außen legt an, von hier löscht (#1025) |

**`wert_herkunft` und `history.herkunft` beantworten zwei verschiedene
Fragen.** Der Verlauf ist ein **Protokoll**: er sagt, dass irgendwann ein Wert
von außen kam. Er sagt nicht, ob er noch dasteht — wer die Nennung eines Bildes
daraus läse, behauptete über ein darübergeschriebenes eigenes Foto weiter, es
sei ein Bildausschnitt von Scryfall. Eine falsche Nennung ist schlimmer als
keine.

`card_values` ist bewusst **schmal und generisch**: ein MSE-Spiel definiert seine
Felder selbst, und eine Spalte je Feld hieße, für jedes fremde Spiel zu
migrieren. Der Preis ist, dass Typisierung nicht in der Datenbank steht, sondern
im Spielpaket — dort steht sie ohnehin.

Volltextsuche über `cards`/`card_values` läuft über eine FTS5-Tabelle, die per
Trigger nachgeführt wird. **Ihre Zeilen heißen wie die Wertzeilen**
(`0049_volltext_zeilennummer.sql`): der Auslöser löscht über die `rowid`, nicht
über `card_id` und `feld` — die Kartenkennung ist im Index unindiziert, und
eine Löschung darüber las bei 481.236 Werten den ganzen Index, 1,2 Sekunden je
geändertem Feld (gemessen am 2026-09-09, U5).

**Was im Bestand immer gilt, steht in [`INVARIANTEN.md`](INVARIANTEN.md)** —
je Regel der Ort, der sie hält (Domäne, Ablage, Schema), und der Lauf, der
es belegt; die Wächter dort befragen das Schema selbst (ENG-05,
#1730).
Die leseintensiven Listen sind in [`ABFRAGEPLAENE.md`](ABFRAGEPLAENE.md)
**gemessen**, über SQLites Zähler für Vollscans an den echten Abfragen.

**Die erste Wanderung nach der Grundlinie** ist `0098_sicherungsvermerke.sql`
(M01, #2345):
die Tafel `sicherungsvermerke`, eine Zeile je geschriebener und geprüfter
Sicherung — Zeitpunkt, Anlass, Fläche, **Dateiname ohne Pfad**, Größe,
SHA-256. Wer vermerkt, entscheidet der Aufrufer (OPS01, #2347); bis dahin ist
sie leer, und **leer heißt unbekannt**, nicht „nie gesichert". Der Aufstieg von
0.8.0 mit 25.000 Karten ist am nativen Fenster belegt
(`docs/belege/huellenprobe/aufstieg09.md`).

### Einstellungen liegen in zwei Tafeln und stehen in vier Ebenen

**Die Tafeln sind zwei**: `settings` je Profil und `nutzereinstellungen` je
Mensch. **Die Ebenen sind vier** — `geraet`, `nutzer`, `profil`,
`arbeitsplatz` —, und sie sagen etwas anderes: nicht, **wo** ein Wert liegt,
sondern **wem die Frage gehört** und wie weit sie reist
([0079](entscheidungen/0079-einstellungsebenen-und-portabler-import.md)).

**Beides auseinanderzuhalten ist die ganze Aussage.** `sprache` steht in
beiden Tafeln und ist trotzdem **eine** Frage auf der Ebene `nutzer`; der
Eintrag im Profil ist ihre Übersteuerung. Und `vorgabe.hochladepfad` liegt in
`settings` wie jede Profilvorgabe, gehört aber der Ebene `geraet` — es ist ein
Pfad auf **dieser** Platte und reist nie mit.

**Ein Wächter hält, dass jeder Schlüssel beider Reihen in genau einer Ebene
steht** (`internal/domain/einstellungsebene_test.go`), und ein zweiter, dass
nichts mitreist, dessen Name nach einem Geheimnis, einem Pfad oder einer
lokalen Kennung klingt (`internal/werkzeuge/austauschsperre_test.go`).

### Zehn Wurzeln gehören einem Profil

**Seit `0035_profiltrennung.sql`** (#1351,
Vermerk [0066](entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md))
tragen `sets`, `decks`, `turniere`, `formate`, `sammlung`, `settings`,
`export_options`, `spiel_listenspalten`, `spiel_sortierung`,
`spiel_autoersetzungen` und `spiel_autoersetzung_an` eine `profil_id`.

**Die Abkömmlinge erben sie über ihren Fremdschlüssel** und bekommen keine
eigene Spalte: eine `profil_id` an `card_values` wäre eine zweite Wahrheit
neben `card_id`, und zwei Wahrheiten laufen auseinander, sobald jemand eine
Karte verschiebt. Was das kostet, gehört gesagt — ein Abkömmling ist nur so
sicher wie die Abfrage, die seine Kennung geliefert hat.

**`profilvorgaben` bekommt ebenfalls keine, und zwar aus demselben Grund**: die
Tabelle **ist** die Zuordnung zu einem Profil (`0043_profilvorgaben.sql`,
Werkstück 15). Sie hält keine Fremdschlüssel auf das, worauf sie zeigt — zwei
ihrer fünf Arten nennen ein **Paket** auf der Platte, und ein `CASCADE` nähme
eine Vorgabe still weg, wo sie **gemeldet** gehört ([0069](entscheidungen/0069-eine-vorgabe-die-auf-eine-zeile-zeigt.md)).

**Und `zugriffstoken` bekommt keine, obwohl es zu einem Profil gehört** — es
trägt eine `profil_id` als **Fremdschlüssel** und nicht als Trennung
(`0042_zugriffstoken.sql`, Werkstück 10; seit `0048_tokenrotation.sql` mit
Principal, letzter Nutzung und der gehüteten Form des **vorigen** Geheimnisses
samt seiner Frist — Werkstück O2 von Phase 5.2, [0080](entscheidungen/0080-maschinenzugaenge-und-widerrufbare-sitzungen.md)).
Der Unterschied ist die Richtung: die
zehn Wurzeln oben werden **innerhalb** eines Profils gelesen, ein Token wird
**vorgezeigt** und sagt damit erst, in welchem Profil gearbeitet wird. Eine
Trennung, die das Profil voraussetzt, käme hier zu spät.

**Und `profilfreigaben` auch nicht, aus dem Grund des Tokens** — sie trägt eine
`profil_id` als **Fremdschlüssel** und sagt, wer außer dem Eigner an diesem
Profil etwas darf (`0044_profilfreigaben.sql`, Werkstück 11). Sie wird gelesen,
**bevor** feststeht, in welchem Profil gearbeitet wird; eine Trennung, die das
Profil voraussetzt, käme auch hier zu spät. **`ON DELETE CASCADE` ist hier
richtig** — anders als bei den Vorgaben: eine Freigabe ohne Profil oder ohne
Nutzer ist keine Auskunft, sondern Müll, und sie wieder gültig werden zu lassen,
sobald jemand dieselbe Kennung bekommt, wäre eine Rechteerweiterung durch Zufall
([0070](entscheidungen/0070-rechte-rollen-und-sichtbarkeit.md)).

**Und `deck_herkunft` und `sammlung_herkunft` bekommen keine, weil sie an
ihrem Ding hängen** (`0045_herkunft.sql`, Werkstück 12): sie tragen die Kennung
des Decks beziehungsweise des Exemplars als **Primärschlüssel** mit
`ON DELETE CASCADE` und erben die Trennung darüber — dieselbe Überlegung wie
bei den Abkömmlingen oben. Ihr `quell_profil` ist dagegen **kein**
Fremdschlüssel: eine Herkunft ohne das Profil, aus dem sie kam, ist weiter eine
Auskunft ([0071](entscheidungen/0071-woher-ein-deck-kommt.md)).

**Und `freigabenachweise` bekommt keine, aus dem Grund des Tokens** — ein
Nachweis trägt `profil_id` als **Fremdschlüssel** mit `ON DELETE CASCADE`
(`0047_freigabenachweise.sql`, Werkstück O1 von Phase 5.2): er wird
**vorgezeigt** und bindet dabei Nutzer, Profil, Zweck, Ziel und Revision; eine
Erlaubnis für ein Profil, das es nicht mehr gibt, wäre eine Erlaubnis für ein
Nichts ([0078](entscheidungen/0078-der-wirksame-akteur.md)). Der Nutzer ist
**kein** Fremdschlüssel: ein gelöschter Nutzer nimmt seine Profile mit, und
damit gehen die Nachweise ohnehin.

**Und `freigabenrevision` gehört dem Bestand, nicht einem Profil** — eine
Zeile, ein Zähler, hochgezählt von Auslösern an `bestand` und
`profilfreigaben` in derselben Transaktion, in der sich ändert, wer wo was
darf (dieselbe Wanderung). Ein Nachweis bindet die Revision, unter der er
ausgestellt wurde, und das Fenster liest daran ab, ob sein Kontext noch der
gezeigte ist.

**Die Eindeutigkeit gehört zur Trennung, und sie hat gefehlt.**
`0035_profiltrennung.sql` hat `sammlung` neu gebaut und deren `UNIQUE`
mitgezogen; die drei Indizes, die als eigenes `CREATE UNIQUE INDEX`
danebenstanden, blieben stehen. `0046_eindeutig_je_profil.sql` zieht sie nach:
`decks`, `turniere` und `formate` über `(profil_id, spiel, name)`, `sets` über
`(profil_id, herkunft)`. **Bemerkt hat es die Deckübernahme** — ein Fehler, den
man nur mit **zwei** benutzten Profilen sieht.

**Drei Tabellen bekommen ausdrücklich keine**: `packages` und `package_deps`
führen den Paketkatalog des **Rechners** (eine Vorlage liegt einmal da, gleich
wer sie benutzt), `preisstaende` sind fremde Daten und kein Besitz, und
`bestand` ist die Buchführung über das Schema selbst.

### Was ein Gegenüber sieht, ist erteilt und nicht abgeleitet

**Seit `T3`** (#1468)
gibt es `sammlungsfreigaben` — je Empfänger, Zweck und Umfang eine Zeile, mit
`sammlungsfreigabeauswahl` für die einzeln gewählten Drucke. **Ohne Freigabe
sieht niemand etwas**: der Standard ist geschlossen, und `domain.FreigabenFuer`
fragt zuerst den Peerkontakt — ein gesperrter bekommt nichts, gleich welche
Freigaben danebenliegen.

**Der Empfänger ist die Peerkennung und nicht die Zeile des Kontakts.** Sie
**ist** der öffentliche Schlüssel; eine Kontaktzeile ist die lokale Notiz
darüber. Eine Freigabe an der Zeile wäre nach einem Löschen und
Wiederkennenlernen wieder da, ohne dass jemand sie erneut erteilt hätte. Der
Preis ist, dass es keinen Fremdschlüssel gibt: eine Freigabe ohne Kontakt
bleibt stehen und zeigt nichts.

**Was hinausgeht, ist eine eigene Gestalt.** `domain.Sichtkarte` trägt sechs
Felder — öffentliche Druckreferenz, Menge, Zustand, Sprache, Ausführung,
lesbare Bezeichnung — und eine **sitzungsgebundene** Referenz statt der
Bestandskennung. Weder Erwerbspreis noch Notiz, Lagerort, Herkunft oder
Fassung; geprüft wird das an den **serialisierten Bytes** und gegen die Werte,
nicht gegen die Feldnamen.

### Ein Weg in den Besitz, und er fragt die Zusagen

**Seit `T2`** (#1467,
Vermerk [0120](entscheidungen/0120-eine-zusage-bindet-den-schreibweg.md))
halten `reservierungen` Stücke eines Exemplars für einen Handelsvorgang fest.
Damit das etwas wert ist, fragen `Sammlungsspeicher.Aendern` und `…Loeschen`
die offenen Zusagen — **in derselben Transaktion, die schreibt**, denn zwischen
Lesen und Schreiben passt eine zweite Zusage.

**Es gibt genau einen Weg dorthin.** Gezählt am 2026-09-18: drei SQL-Anweisungen
schreiben Zeilen der Sammlung, alle drei in `internal/storage/sammlung.go`;
Massenpflege und Listeneinfuhr haben keinen eigenen. Ein Wächter
(`internal/werkzeuge/bestandsschreibwege_test.go`) hält das — eine vierte
Anweisung erzwingt die Entscheidung, statt still an der Zusage vorbeizugehen.

### Die Ausführung ist eine Angabe, kein Wahrheitswert

**Seit `0070_ausfuehrung.sql`** (#1467,
Vermerk [0119](entscheidungen/0119-eine-ausfuehrung-statt-eines-wahrheitswerts.md))
tragen `sammlung` und `wiederherstellungsmengen` eine Spalte `finish` statt
`foil`: `nonfoil`, `foil` oder `etched`. Die Preisdomäne kennt die drei seit
Phase 5.3; der Bestand kannte zwei, und eine geätzte Karte bekam damit den
**Foilpreis**.

**Ein Tabellenneubau und kein `ALTER TABLE`.** `foil` steht in beiden Tabellen
innerhalb der `UNIQUE`-Bedingung, und die kann SQLite nicht ablegen — der
Schritt baut sie deshalb neu auf, so wie `0035_profiltrennung.sql` es mit
`sammlung` schon getan hat. Übersetzt wird `CASE foil WHEN 1 THEN 'foil' ELSE
'nonfoil' END`: **`etched` wird nie geraten**, denn aus `true` geht nicht
hervor, welche der beiden Veredelungen gemeint war.

**Nach außen kommt `finish` hinzu, `foil` bleibt** — MCP, REST und die
Kommandozeile sind eine Fläche mit fremdem Gegenüber (Vermerk
[0033](entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md), Regel
6 und 8). Im Bestand steht **ein** Feld; die zweite Spalte gibt es nur in der
Austauschdatei.

### Der Druck als Zeile, seine Kennungen mit Bedeutung

**Seit `0051_drucke.sql`** (#1441,
Vermerk [0086](entscheidungen/0086-sieben-begriffe-fuer-eine-karte.md)) ist
ein Druck eine Zeile in `drucke`, und `druckkennungen` sagt, wie andere
Systeme ihn nennen — mit **Semantik**: `druck` (eindeutig je System),
`konzept` (die Oracle-ID, die dreißig Drucke teilen), `produkt` (Händler,
mehrere je Druck) und `kunstwerk`. Eindeutig ist **nur** `druck`, deshalb ist
der eindeutige Index ein teilweiser. `sammlung`, `karten_verweis`,
`deckplaetze`, `preisstaende` und `turnierdeckplaetze` tragen `druck_id`
**neben** `printing`: der Alias bleibt, was jede Fläche kennt; die Zeile ist,
was ein Druck ohne Scryfall hat. `druckZeile` in `internal/storage` hält
beides zusammen — wer einen Alias schreibt, bekommt die Zeile dazu, angelegt
oder wiedergefunden, im selben Zug.

**`deckplaetze` wird dabei neu gebaut**: 0028 verlangte genau eines von
`karte_id` und `printing`, und SQLite ändert keine CHECK-Bedingung. Auf die
Tabelle zeigt niemand, die Kennungen bleiben.

**Die lokale Nummer verlässt den Bestand nie.** Eine Ausfuhr trägt die
portable Referenz — Fassung, System, Kennung, Rückfall —, und die Einfuhr
löst sie über `service.Druckregister` auf: eindeutig, Konflikt oder neu.

**Das Profil reist im Kontext**, nicht in der Signatur —
[`storage.ProfilAus`](../internal/storage/profil.go) liest es und gibt einen
**Fehler**, wenn keines gesetzt ist. Drei Wächter halten die Zusage; welcher
was hält und was nicht, steht im Vermerk.

## Was im Bestand liegt und was auf der Platte

**Entschieden in [0029](entscheidungen/0029-bilder-liegen-neben-dem-set.md)**,
und beide Male an Zahlen und nicht an Gefühl
(#247).

| Gegenstand | Wo | Warum |
| --- | --- | --- |
| Karten und Feldwerte | **Bestand** | `card_values`, schmal und generisch |
| Die Hauptdatei eines Sets | **Bestand**, Byte für Byte | die Rundlaufzusage aus #48 ([0011](entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)) |
| Vorlagenpakete und ihre Grafiken | **Platte** | über die zehn kanonischen Sets gemessen: 1.996,0 MB Bilder gegen 22,9 MB Hauptdateien — **98,9 %** des Gewichts |
| Bilder, die Nutzende einsetzen | **neben das Set**, nicht hinein | dieselbe Rechnung im Kleinen ([0029](entscheidungen/0029-bilder-liegen-neben-dem-set.md)) |
| Profilbilder und Profil-Abzeichen | **`beiwerk/nutzer/`**, neben den Kartenbildern und nicht darunter | ein Kartenbild gehört einem Set, ein Profilbild einem **Menschen** — dort abgelegt verlöre man es beim Löschen des Sets (#1351, Werkstück 8). Der Bestand führt nur den Dateinamen; er trägt die Kennung, weil ein Anzeigename sich ändert, **und seit dem Zweitblick über Phase 6.1 einen gewürfelten Teil** (`nutzer-3-<12 Hexziffern>.png`): der Bestand wechselt den Verweis mit der Fassung, erst danach geht das alte Bild — ein abgewiesenes Schreiben lässt das Bild, wie es war, und zwei gleichzeitige Wechsel teilen sich keine Datei (Nachlauf, Befund 1). **Ein Abzeichen gehört einem Profil** und heißt deshalb `profil-7-….png` neben `nutzer-3-….png` (Werkstück 9). Das Tischbild bleibt beim festen Namen: auf es zeigt kein Verweis |
| Der Paketkatalog | **Bestand, abgeleitet** | er wird bei jedem Öffnen abgeglichen |

**MTG JSON ist ein Austauschformat und kein Modell** — nachgemessen an
`Multicolor.mse-set`: von 62 Feldnamen kennt `internal/mtgjson/zuordnung.go`
**elf**, und was es an **jeder** der 3.008 Karten nicht kennt, ist `stylesheet`,
`stylesheet version`, `has styling` und `styling data`. Ein Modell, das nicht
weiß, womit eine Karte gezeichnet wird, kann für einen Karteneditor keines sein.

## Fehler über die Schichten

`domain` definiert `ErrValidation`, `ErrNotFound`, `ErrConflict`. `storage`
bildet SQLite-Fehler darauf ab, `app` übersetzt sie in eine für das Frontend
stabile Struktur aus Code, Meldung und optionalem Feldbezug. Rohe
SQLite-Meldungen erreichen die Oberfläche nie.

**Einen liest nur der REST-Zugang**: `ErrErlaubnis` heißt „dein Nachweis
stimmt, er darf das nur nicht" und wird dort zu einem `403`. `internal/app` hat
für ihn **absichtlich** keinen Zweig — das Fenster weist sich nicht mit einem
Token aus, und ein Zweig ohne Erzeuger wäre eine Zusage ohne Rumpf (§10,
Werkstück 10 von Phase 4.12).

Für fremde Daten kommt eine zweite Art hinzu: ein **Befund** ist kein Fehler,
sondern ein benannter Vorbehalt. Ein Set, von dem 300 Karten lesbar sind und zwei
nicht, wird mit 300 Karten und zwei Befunden geöffnet. Ein Befund trägt immer
seine Stelle — Datei, Zeile, Feld.

## Was die Grenze überschreitet

**Drei Stellen, und keine läuft von selbst**
(#678).

| Stelle | Richtung | Wann | Was |
| --- | --- | --- | --- |
| `internal/bezug` | hinaus | **auf Knopfdruck** | HTTP an einen Vorlagenkatalog, an einem **Stand** und mit Prüfsumme, mit sicherem Auspacken (§2, `internal/bezug`) |
| `cmd/mse-mcp` | herein | **wenn jemand es startet** | stdio, kein Port. Fünf Werkzeuge, davon eines schreibend ([0010](entscheidungen/0010-mcp-als-eigener-prozess.md)) |
| Dateien | beides | auf Knopfdruck | Sets, Pakete, geschriebene Bilder und PDF; der Bestand im Benutzerordner |

**Sonst nichts.** Keine Telemetrie, kein Konto, keine Anmeldung, keine
Aktualisierungsprüfung im Hintergrund.

### Die Frage, die am häufigsten kommt

**Wie schaltet man den MCP-Server ab?** Gar nicht — er ist nicht an. Er ist ein
eigenes Programm, das ein Aufrufer als Unterprozess startet; ohne diesen Aufruf
existiert er nicht. Ein Schalter in der Anwendung wäre eine Behauptung über
etwas, das sie nicht führt.

**Und was er ändert, sieht das Fenster erst nach einem Neuladen.** Der Preis
steht im Vermerk.

### Wo das noch steht

Dieselbe Sicht steht **in der Anwendung** unter *Hilfe* — dort, wo jemand sie
sucht, der den Quelltext nicht liest. Wer hier etwas ändert, ändert es dort mit.

---

## Was hier bewusst fehlt

- **Kein Netzwerkdienst.** Die Anwendung läuft lokal. Der Paketverwalter (Phase 7)
  lädt auf ausdrückliche Anweisung, sonst geht nichts hinaus.

  **Phase M hat diesen Satz zur Frage gestellt und ihn bestätigt**
  (#591): ein
  REST-Dienst auf `127.0.0.1` wäre einer, und
  [0039](entscheidungen/0039-kein-dienst-auf-der-rueckschleife.md) hat am
  2026-08-22 dagegen entschieden — kein benannter Abnehmer, und die Lücke liegt
  im **Umfang** der MCP-Werkzeuge (fünf, davon eines schreibend) und nicht im
  Protokoll. Der Satz bleibt, wie er ist.
- **Kein Plugin-System für Go-Code.** Erweiterbar ist die Anwendung über
  MSE-Pakete — das ist der Weg, den die Daten ohnehin vorgeben.
- **Keine eigene Skriptsprache.** MSE-Script wird nachgebaut, nicht ersetzt.
  Jede Erweiterung, die vorhandene Vorlagen brechen könnte, ist keine.
