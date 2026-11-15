# Umsetzungsleitlinien — MagicSetEditor4Mac

Diese Datei ist verbindlich für jede Änderung an diesem Repository — **von wem
auch immer**, ob Mensch oder Werkzeug. Sie präzisiert, *wie* gearbeitet wird;
*was* gebaut wird, steht in `docs/ROADMAP.md` und in den Issues. **Ein Verweis
`#N` aus der Zeit bis 1.0.0** nennt ein Issue oder einen PR des privaten
Entwicklungsarchivs, aus dem dieses Repositorium neu angelegt wurde
([0205](docs/entscheidungen/0205-1-0-0-ist-ein-neuanfang.md)).

`CLAUDE.md` ist **gleichlautend**. Nicht als Verweis, sondern Zeichen für
Zeichen dieselbe Datei — Werkzeuge lesen das eine oder das andere, und keines
soll dabei eine andere Regel bekommen. Dass beide gleich bleiben, ist nicht der
Umsicht überlassen: die Stufe `format` vergleicht sie und endet rot, sobald sie
auseinanderlaufen. Wer eine Regel ändert, ändert sie einmal und ruft
`make leitlinien` — das schreibt die zweite Datei aus der ersten.

---

## 1 Umgebung

Vor jeder Arbeit an Go-Code:

```bash
source scripts/activate.sh
```

Damit landen Modul-Cache, Build-Cache und Werkzeug-Binärdateien unter `.go/` im
Projekt statt in `~/go`. Verlassen mit `go_deactivate`.

**Niemals** Werkzeuge global installieren. Alle Werkzeuge sind gepinnt:

| Werkzeug | Version | Ort der Pinnung |
| --- | --- | --- |
| Go | 1.27.4 | `go.mod` (`go`-Direktive) **und** `GOTOOLCHAIN` in `scripts/ci.sh` und `scripts/activate.sh`. **Gehoben am 2026-10-09 als Sicherheitsaktualisierung** (#2274): `govulncheck` meldete elf erreichte Schwachstellen in `net/http`, HTTP/2, `net/textproto`, `crypto/tls` und `os`. **Auf 1.27.3 am 2026-10-21**, wieder als Sicherheitsaktualisierung: GO-2026-6809 in `crypto/x509` (kompromittierte Zwischenzertifikate), erreicht über `leitung.ketteBilden`. **Auf 1.27.4 am 2026-11-15** für 1.0.0, wieder als Sicherheitsaktualisierung: GO-2026-6859 in `net/http` (ein HTTP/1-Server stürzt über einen Wettlauf im `MaxBytesReader` ab), erreicht über `bezug.Hole` und den REST-Dienst |
| Wails | v3.0.0-beta.30 | `go.mod` — die Hülle um die Anwendung ([0124](docs/entscheidungen/0124-wails-3-als-gepinnte-vorabfassung.md), [0132](docs/entscheidungen/0132-wails-3-als-huelle.md)). **Eine Vorabfassung, auf die Stelle gepinnt**, und `@wailsio/runtime` in `frontend/package-lock.json` trägt dieselbe Zahl — ein Wächter hält beide gleich. **Kein Kommandozeilenwerkzeug**: das Bündel legt `scripts/bauen.sh`. **Gehoben am 2026-10-08** (L01, #2189): zwischen beta.23 und beta.27 ändert sich am eingebundenen Code nur der `WKUIDelegate` für Kamera und Mikrofon (`webview_window_darwin.m`), das Laufzeitpaket ist bis auf die Fassungsnummer bytegleich — verglichen am Quelltext beider Fassungen. **Gehoben auf beta.30 am 2026-10-20** (#1784): `InvokeSync` ruft `wg.Done()` jetzt im `defer`, und ein Wartender hängt nach einer Panik nicht mehr (#2398) — die Umgehung, die die Anwendung dann beendete, ist fort; dazu ein Schutz gegen doppeltes Schließen, zwei Fensterereignisse weniger, die hier niemand hört, und Tastenkürzel nach der Tastaturbelegung; das Laufzeitpaket wieder bis auf die Fassungsnummer bytegleich |
| golang.org/x/image | v0.45.0 | `go.mod` — Glyphen ([0008](docs/entscheidungen/0008-schriftrasterung.md)). **Gehoben am 2026-09-23 als Sicherheitsaktualisierung** (#1779): geändert haben sich der WebP-Dekodierer `vp8l` und Assembler **für amd64** in `vector`; `font/sfnt`, `font/opentype` und `math/fixed` sind unverändert, auf arm64 rastert dieselbe Rechnung |
| golang.org/x/text | v0.42.0 | `go.mod` — die Ordnung einer Sammlung nach Sprache (#1500). **Mitgehoben am 2026-10-07 mit tailcat v0.7.0** (K01, #2187), das sie über `tailscale.com` verlangt |
| golang.org/x/net | v0.60.0 | `go.mod` — `HTTP_PROXY`, `HTTPS_PROXY` und `NO_PROXY` (#1533). **Lag schon als indirekte da**; direkt wurde sie, weil `http.ProxyFromEnvironment` die Umgebung beim ersten Ruf zwischenspeichert — gemessen am 2026-09-12. **Mitgehoben am 2026-10-07 mit tailcat v0.7.0** (K01). **Auf v0.60.0 gehoben am 2026-10-09** mit Go 1.27.2 — dieselben HTTP/2-Fehler (#2274) |
| tailcat | v0.7.0 | `go.mod` — die Verbindung über fremde Netze (#1456). **Gepinnt, weil es nichts zusagt**: „the Go API, the CLI flags and output, and the wire format may all change". Deshalb trägt die Einladung eine Fassung ([0114](docs/entscheidungen/0114-tailcat-gepinnt.md)). **Gehoben am 2026-10-07** (K01, #2187): seit v0.7.0 nennt `Server.Status()` die Gegenstellen, und der Fernweg damit seinen Weg; `wire.go` ist unverändert, `tailscale.com` folgt der Fassung, die tailcat verlangt |
| hashicorp/mdns | v1.0.7 | `go.mod` — die Entdeckung im selben Netz (#1455). **Nicht `grandcat/zeroconf`**: das ist bei v1.0.0 vom Januar 2020 stehengeblieben, und sein `miekg/dns` stammt von 2019 ([0113](docs/entscheidungen/0113-kein-libp2p-im-lan.md)) |
| filippo.io/age | v1.3.2 | `go.mod` — Sicherungen mit Kennwort (#1723). **Ein Format und kein eigenes**: `age -d` öffnet eine Sicherung auch ohne diese Anwendung, und eine zweite Stelle, an der sich Kryptographie falsch zusammensetzen ließe, gibt es nicht. BSD-3-Clause, von Filippo Valsorda gepflegt; bringt `filippo.io/hpke` v0.4.0 mit. **Gemessen am 2026-09-26**: rund 0,2 MB mehr im Binärprogramm `mse` |
| golangci-lint | v2.15.0 | `tools/go.mod` — **gehoben am 2026-10-20** (#1784): errcheck 1.30 meldet seither einen unbeachteten Fehler aus einem Kanal; sieben Testläufe warteten auf einen `chan error`, ohne ihn anzusehen, und drei davon verschwiegen dabei einen erwarteten Fehler |
| gosec | v2.29.0 | `tools/go.mod` — **mitgehoben** mit golangci-lint v2.14.0, das sie verlangt (#1784) |
| govulncheck | v1.9.0 | `tools/go.mod` — **gehoben am 2026-10-20**: am Quelltext bytegleich mit v1.8.0, nur die Abhängigkeiten sind neuer |
| gofumpt | v0.12.0 | `tools/go.mod` — **seit v0.12.0 mit gofmt einig** über die Rückgabe mehrerer Literale; zehn Dateien in `internal/mcp` hatte gofmt bis dahin anders eingerückt (#1784) |
| cyclonedx-gomod | v1.12.0 | `tools/go.mod` |
| Node | 24.15.0 | `.github/workflows/` (`node-version`) |
| npm | 11.12.1 | `frontend/package.json` (`packageManager`) |
| TypeScript | 6.0.3 | `frontend/package-lock.json` |
| ESLint | 9.39.5 | `frontend/package-lock.json` |
| PixiJS | 8.22.0 | `frontend/package-lock.json` — der gezeichnete Spieltisch (#1460). **Auf die Stelle genau gepinnt** (`--save-exact`) und nicht über einen Bereich: eine Grafikbibliothek, die sich zwischen zwei Installationen ändert, ändert das Bild. Gelesen wird wie bei TypeScript und ESLint aus der **Sperrdatei** — sie sagt, was installiert wird, `package.json` nur, was gewünscht war. **Gehoben auf 8.22.0 am 2026-10-08** (L02, #2190): zwei Verhaltensänderungen nachgelesen — 8.21 zerlegt gespiegelte Matrizen anders (der Tisch nutzt das nicht), 8.22 aktiviert beim ersten Tab wieder seine Barrierefreiheitsschicht; der Tisch schaltet das ab (`activateOnTab: false`), damit die eigene Tastaturbedienung bleibt, wie sie war |
| Mantine | 9.7.3 | `frontend/package-lock.json` — die Bauteilbibliothek (#1679, [0125](docs/entscheidungen/0125-mantine-als-bauteilbibliothek.md)). **Auf die Stelle genau gepinnt** (`--save-exact`), und `@mantine/hooks` trägt dieselbe Zahl: sein Peerbereich verlangt sie **ohne Spielraum**. Gemessen am 2026-09-20: 16 neue Pakete, 26 MB, MIT. **Auf 9.6.3 gehoben am 2026-10-05** (#1784), ihre fünf Abhängigkeiten unverändert. **Auf 9.7.1 gehoben am 2026-10-08** (L02): neue Bauteile, die hier nicht benutzt werden, und Korrekturen an ARIA (Select, Combobox, Tree). **Auf 9.7.3 gehoben am 2026-10-20** (#1784) und nicht auf die 9.7.2 des Dashboards: erst 9.7.3 dichtet die Fokusfalle unter macOS ohne Tastaturnavigation ab, und an ihr hängen `Modal` und `Drawer` — die Tab-Reihenfolge darin rechnet jetzt Radiogruppen, positive `tabindex` und verschachtelte Fallen; 10.351 Tests grün, die Zählung unverändert |
| Vitest | 5.0.5 | `frontend/package-lock.json` — der Testläufer der Oberfläche (ENG-04, #1729). **Auf die Stelle genau gepinnt** (`--save-exact`), und `@vitest/coverage-v8` trägt dieselbe Zahl: sein Peerbereich verlangt sie **ohne Spielraum** — dieselbe Kopplung wie bei Mantine. Wer eines von beiden allein hebt, bekommt keine Auflösung. **Gehoben von 4.1.11 auf die Hauptfassung 5 am 2026-10-08** (L03, #2191) — erst an einer Kopie gemessen: 1.065 Dateien, 8.929 Tests grün, 100 % in allen vier Quoten, 365 s statt 378 s; **das Maß verschob sich um fünf Anweisungen** (19.748 statt 19.753 gezählt), die Quote nicht. 4.1.11 war gewählt, weil `@vitest/mocker` bis 4.1.10 [GHSA-82fw-gwwq-j7x9](https://github.com/advisories/GHSA-82fw-gwwq-j7x9) trug. **Gehoben auf 5.0.5 am 2026-10-20** (#1784): 1.293 Dateien, 10.351 Tests grün, 100 % in allen vier Quoten; **das Maß verschob sich wieder** — 22.891 statt 22.886 Anweisungen, 13.422 statt 13.418 Zweige, 7.579 statt 7.574 Funktionen, 21.715 statt 21.710 Zeilen, mit `ast-v8-to-istanbul` 1.0.7 statt 1.0.5. Die drei Meldungen vom 2026-10-16 treffen hier nichts: `@vitest/ui` und `@vitest/browser` sind nicht installiert, und `@vitest/mocker` 5 liegt außerhalb ihrer Bereiche |
| GitHub Actions | je Commit-Hash | `.github/workflows/`, lesbarer Hinweis als Kommentar daneben |

Reproduzierbar ist der Frontend-Stand über `npm ci`, nicht über `npm install`.
Die TypeScript- und ESLint-Zeilen nennen die Version aus der **Sperrdatei**,
nicht aus `package.json` — dort dürfen Bereiche stehen, und eine Prüfung gegen
einen Bereich prüfte die Absicht statt dessen, was installiert wird.

**Die `go`-Direktive allein genügt auch nicht** — und aus demselben Grund.
Sie nennt eine **Untergrenze**: ist die Toolchain des Rechners neuer, nimmt
Go sie. Am 2026-09-01 hat Homebrew mitten in einer Sitzung von 1.26.6 auf
1.27.0 gehoben, und die Stufe `lint` wurde rot, weil golangci-lint v2.12.2
`math/rand/v2` von 1.27 nicht lesen konnte — bei unverändertem Quelltext.
Wirksam wird die Pinnung durch `GOTOOLCHAIN`, und die Zeile steht in
**`scripts/ci.sh`**: der `pre-push`-Hook liest `activate.sh` nicht, und ein
Lauf von Hand, der grün ist, während der beim Pushen rot ist, sind zwei
Pipelines (§4).

Gehoben auf 1.27.1 am 2026-09-15 — der Anlass steht in
[`docs/PIPELINE.md`](docs/PIPELINE.md) —, auf 1.27.2 am 2026-10-09 (#2274), auf
1.27.3 am 2026-10-21 (GO-2026-6809) und auf 1.27.4 am 2026-11-15 (GO-2026-6859), je
als Sicherheitsaktualisierung.

**`packageManager` allein genügt nicht.** Ohne Corepack liest `npm ci` das Feld
gar nicht; eine falsche Angabe liefe kommentarlos durch. Es ist die Deklaration
— wirksam wird sie dadurch, dass die Stufe `format` `npm --version` dagegen
hält. Wer das eine ohne das andere ändert, hat wieder eine Behauptung.

**TypeScript und ESLint bleiben bewusst eine Hauptversion zurück**, beides an
einem Peer-Bereich und beides am 2026-08-08 nachgemessen:

| Warum nicht die neueste | Nachgemessen |
| --- | --- |
| **TypeScript 7** | `typescript-eslint` gibt `typescript@>=4.8.4 <6.1.0` an und weist 7 ab. Zwischen 5 und 7 liegt die 6 — wer eine Hauptversion prüft, prüft **jede** dazwischen |
| **ESLint 10** | `eslint-plugin-jsx-a11y@6.10.2` gibt `^3 \|\| … \|\| ^9` an. Der Wechsel kostete für jeden Installationsweg `--legacy-peer-deps`, und das schaltet die Auflösung von Peer-Abhängigkeiten für den **ganzen** Lauf ab — ein Dauerpreis für einen Nutzen von null |

`@eslint/js` ist mit ESLint gleichauf versioniert und folgt derselben Sperre.

**Actions hängen an einem Commit, nicht an einem Tag.** `uses: …@v5` bezieht den
Code, den GitHub *heute* unter `v5` führt; ein unveränderter Commit dieses
Repositoriums führte morgen anderen aus. Der Kommentar dahinter nennt die
Fassung, denn ein Hash allein sagt niemandem, welche es ist.

Neues Werkzeug: `cd tools && go get -tool <pfad>@<version> && go install tool`, dann Tabelle und `README.md`. **Jede neue Abhängigkeit begründet Größe, Lizenz
und Pflege** und bekommt ihre Zeile in `docs/ABHAENGIGKEITEN.md` (ENG-04).

### Doku-Nachschlag

Für Bibliotheks- und Frameworkfragen — Wails, React, Vite, SQLite-Treiber,
Tailwind — ist **Context7** die Quelle, nicht das Gedächtnis. Der Server ist in
`.mcp.json` eingetragen. Die Regel dahinter ist dieselbe wie in §10: eine
Auskunft aus der Erinnerung ist kein Prüfergebnis. Wer eine API-Signatur, ein
Konfigurationsfeld oder ein Standardverhalten behauptet, hat es vorher
nachgeschlagen.

---

## 2 Architektur

Fünf Schichten. **Abhängigkeiten zeigen ausschließlich nach innen.**

```
frontend/            React, TypeScript, Vite, Tailwind
    ↕ Wails 3        internal/fenster/ und frontend/src/anbindung.ts
internal/app/        Bindings-Fassade: DTOs, Fehlerübersetzung, Kontext
internal/service/    Anwendungsfälle, Transaktionsgrenzen, fachliche Regeln
internal/storage/    SQLite-Repositories, eingebettete Migrationen
internal/domain/     Entitäten, Wertetypen, Validierung
```

Daneben dreißig **Fachbibliotheken**. Sie sind Blätter: sie kennen `internal/domain` und sonst nichts aus diesem Projekt — kein `service`, kein `storage`, kein `app`. **Die eine Kante, die alle Klienten fremder Dienste teilen**, ist `internal/netz` — der Transport mit Fristen und Proxy (#1533); hier stand sie nicht, gefunden vom Zweitblick über Phase 6.10 (Z5). Jede Kante hält `internal/schichten` (`regeln_test.go`, `regelnnetzblaetter_test.go`).

| Paket | Aufgabe |
| --- | --- |
| `internal/bezug` | Beziehen von Vorlagenpaketen über HTTP: Stand statt Zweig, Prüfsumme, sicheres Auspacken |
| `internal/msepaket` | Lesen und Schreiben der MSE-Paketformate (`.mse-set`, `.mse-game`, `.mse-style`, `.mse-symbol-font`, `.mse-include`, `.mse-locale`, `.mse-export-template`) |
| `internal/msescript` | Zerteiler und Auswerter der eingebetteten Sprache MSE-Script samt Standardbibliothek |
| `internal/mcp` | Das Model Context Protocol über stdio — Werkzeuge, mit denen ein Sprachmodell Karten liest und ändert ([0010](docs/entscheidungen/0010-mcp-als-eigener-prozess.md)) |
| `internal/pdf` | Schreiben von PDF aus Bildern und Linien — der Druckbogen und sonst nichts ([0013](docs/entscheidungen/0013-eigener-pdf-schreiber.md)) |
| `internal/mtgjson` | Lesen und Schreiben von **MTG JSON** — dem Kartenformat, das `magic-set-generator` liest ([0009](docs/entscheidungen/0009-phase-9-fremdformate-und-schnittstelle.md)) |
| `internal/render` | Rasterung einer Karte aus einem fertigen Zeichenplan — Bildschirm wie Druckauflösung |
| `internal/mtgapi` | Lesen von Karten bei **magicthegathering.io** — kein Zugang nötig, und als einziger der drei neuen Dienste gegen sein Original geprüft ([0056](docs/entscheidungen/0056-drei-weitere-kartendienste.md)) |
| `internal/mtggraphql` | Lesen von Karten über die **GraphQL-Schnittstelle von MTGJSON** — Marke im Kopf, `errors` im Rumpf als Fehler; nie gegen das Original gelaufen ([0056](docs/entscheidungen/0056-drei-weitere-kartendienste.md)) |
| `internal/scrydex` | Lesen von Karten bei **Scrydex** — **zwei** Kopfzeilen, nur Magic von sechs Spielen; nie gegen das Original gelaufen ([0056](docs/entscheidungen/0056-drei-weitere-kartendienste.md)) |
| `internal/mtgjsonpreise` | Die **Preise** von MTGJSON — eine Datei mit allem, strömend gelesen und nach dem Bestand gefiltert. **Nicht dasselbe wie `internal/mtgjson`**: dort ein Dateiformat, hier ein Dienst mit eigenen Grenzen (A4 baut darauf auf, [0093](docs/entscheidungen/0093-eine-preisreihe-hat-sieben-teile.md)) |
| `internal/cardtrader` | **CardTrader, lesend** — Spiele, Erweiterungen, Blueprints mit ihren Fremdkennungen und die Angebote des Marktplatzes. Die Positivliste führt **vier** Adressen, alle mit `GET`; ein Konto berührt es nicht ([0095](docs/entscheidungen/0095-was-cardtrader-sagt-und-was-davon-gilt.md)) |
| `internal/schluesselbund` | Zugangsdaten fremder Dienste im Schlüsselbund von macOS — **nicht** im Bestand, denn den gibt man weiter ([0056](docs/entscheidungen/0056-drei-weitere-kartendienste.md)) |
| `internal/schriften` | Finden von Schriftdateien auf dem Rechner — die Plattformseite zu `render.Schriftquelle` |
| `internal/svg` | Schreiben einer Karte als SVG aus einem fertigen Zeichenplan — der **zweite** Verbraucher desselben Plans |
| `internal/anmeldung` | Der Anmeldefluss über einen fremden Anbieter — OAuth 2.0 mit PKCE, OIDC mit `nonce`, Rückruf auf der Schleifenadresse. Er kennt keinen Nutzer: er gibt eine [domain.Fremdkennung] zurück, und **wem sie gehört**, entscheidet `service` (Werkstück 4 von Phase 4.12, #1351) |
| `internal/kennwort` | Kennwörter hüten und wiedererkennen — **Argon2id**, die gehütete Form trägt ihre eigenen Zahlen. Es weiß, was ein Hash ist, und nichts über Nutzer (Werkstück 6 desselben Issues) |
| `internal/vcard` | vCards lesen, wie Apple Kontakte sie ausgibt — nach RFC 6350 und 2426, **ohne Netz**: ein Bild als Adresse wird genannt und nicht geholt. Es kennt nichts aus dem Projekt; was aus einer Karte wird, entscheidet `service` (ID-04, #1703, [0137](docs/entscheidungen/0137-kontakte-kommen-als-vcard.md)) |
| `internal/totp` | Zeitbasierte Einmalkennwörter nach **RFC 6238** — reines Go, geprüft gegen alle achtzehn Vektoren aus Anhang B. Das Prüffenster ist eine Entscheidung dieses Prüfers und keine Vorgabe der RFC (Werkstück 7 desselben Issues) |
| `internal/peer` | Die Identität dieses Programms gegenüber anderen — ein **Ed25519**-Paar, dessen Kennung die eines libp2p-Peers ist (Gestalt nachgelesen am 2026-09-14). Die Kennung **trägt den Schlüssel**; der Fingerabdruck ist eine Lesehilfe und kein zweiter Nachweis. Es kennt weder Schlüsselbund noch libp2p (Werkstück N1 von Phase 5.4, #1454) |
| `internal/biometrie` | Touch ID und Face ID über `LocalAuthentication` — cgo-frei über [`tmc/apple`](https://github.com/tmc/apple) und `purego`. **Zwei unfertige Schichten** (v0.6.18 über `purego` v0.11.1 — beide vor 1.0, purego seit dem 2026-10-11 als Endfassung, #2305), deshalb steht jeder Ruf hinter einem Riegel, der einen `panic` in eine Antwort verwandelt (Werkstück 7) |
| `internal/prozessspeicher` | Der Speicher dieses Programms **und seiner WebKit-Prozesse**, die außerhalb liegen — der physische Fußabdruck aus `proc_pid_rusage`, zugeordnet über den **verantwortlichen Prozess**, den macOS jedem mitgibt. Diese Zuordnung ist **nicht öffentlich**; fehlt sie, heißt der Webspeicher „nicht gemessen" und nicht null. Über `purego` und hinter einem Riegel wie `biometrie` (OPS02, #2348) |
| `internal/ollama` | Die **lokale** KI-Laufzeit — fünf Wege, und `/api/pull` ist keiner davon. Was ein Modell kann, kommt aus `capabilities`; der **Einbettungsvertrag wird gemessen** statt geraten (A7, [0098](docs/entscheidungen/0098-was-ollama-belegt-und-was-nicht.md)) |
| `internal/anthropic` | **Claude** — zwei Wege, beide lesend, und der erste Anbieter, der nicht auf diesem Rechner rechnet. Fähigkeiten kommen aus `capabilities`, ein **429 sind zwei verschiedene Dinge** (A8, [0099](docs/entscheidungen/0099-was-anthropic-belegt.md)) |
| `internal/openai` | **OpenAI** — feste Form über die Responses-API. Seine Modellliste **belegt nichts**; deshalb gibt es die **Messung**, und `store` steht ausdrücklich auf `false` (A9, [0100](docs/entscheidungen/0100-was-openai-nicht-sagt.md)) |
| `internal/xai` | **Grok** — feste Form über `/v1/chat/completions`. Die Modellliste belegt **zwei von drei** Fähigkeiten, der Ausgang kennt ein viertes Wort, und **derselbe Anbieter hat an seinen zwei Endpunkten zwei Bildformen** (A10, [0101](docs/entscheidungen/0101-was-xai-belegt.md)) |
| `internal/applefm` | **Apples Gerätemodell** über `fm serve` — **nur über einen eigenen Unix-Socket**, nie über einen Port: mit einem zu langen Socketpfad lauscht `fm` still auf allen Schnittstellen, deshalb wird der Pfad geprüft. **Die Lizenz nimmt ein Mensch an**, dieses Programm ruft nie `sudo`, und Testläufe starten `/usr/bin/fm` nie (AI-05, #2042, [0166](docs/entscheidungen/0166-was-fm-serve-belegt.md)) |
| `internal/leitung` | Nachrichten über einen **Strom** — ein Rahmen mit Längenpräfix, damit eine Nachricht eine bleibt. **Die behauptete Länge wird geprüft, bevor sie zugeteilt wird**: sonst bestimmt ein Fremder mit vier Bytes den Arbeitsspeicher. Über `net.Conn` und damit über drei Wege zugleich — `net.Pipe` für zwei Anwendungen im selben Prozess, TCP für „Adresse von Hand", tailcat fürs Internet. **Und über TCP mit TLS 1.3**, denn `peer` unterschreibt und verschlüsselt nicht: selbst unterschriebenes Zeugnis auf der Peeridentität, **festgenagelt** auf die erwartete Kennung statt auf eine Zertifizierungsstelle, Kanalbindung aus `ExportKeyingMaterial` statt aus dem Vierertupel ([0113](docs/entscheidungen/0113-kein-libp2p-im-lan.md)) (N2, #1455) |
| `internal/fernleitung` | Die Verbindung über **fremde Netze** — tailcat, also Tailscales Datenebene ohne deren Kontrollebene: WireGuard zur Verschlüsselung, DERP als Rendezvous und Rückfall, **kein Konto**. Es kennt `internal/leitung` — die **zweite** benannte Ausnahme dieser Art, aus demselben Grund wie `svg` zu `render`: über dem Tunnel liegt derselbe gesicherte Weg wie im eigenen Netz. **Die Einladung trägt eine Fassung**, denn tailcats Drahtformat darf sich ändern — und ein Bruch soll **vor** dem Wählen erkannt werden statt als Zeitüberschreitung ([0114](docs/entscheidungen/0114-tailcat-gepinnt.md)) |
| `internal/entdeckung` | Wer im selben Netz da ist — **mDNS**, und nur auf Aufruf. `Melde` kündigt dieses Profil an, `Suche` fragt; von selbst geschieht keines von beiden, auch nach einem Neustart nicht. **Was hinausgeht, ist benannt**: die Peerkennung und der selbstgewählte Name — der **Rechnername ausdrücklich nicht**, denn die Bibliothek nähme ihn ohne Zutun, und sie schlüge dafür auch noch einen Namen im DNS nach. **Die Frist ist die Wanduhr und nicht ein Teilschritt**, und beides ist gemessen: der Kontext bricht das Warten nicht ab, und ab zwei Sekunden Frist kommt eine hinzu (N2, #1455) |

Harte Regeln:

- `internal/domain` importiert **nichts** aus dem Projekt und nichts aus
  `database/sql`. Wer dort einen SQL-Import sieht, hat einen Fehler gefunden.
- `internal/mtgjson` kennt `internal/domain` und sonst nichts aus dem Projekt.
  Es liest ein fremdes Kartenformat; die Zuordnung von MTG-JSON-Feldern auf
  Kartenfelder steht dort an **einer** Stelle je Spiel und nicht über den Code
  verteilt.
- `internal/pdf` importiert **nichts** aus dem Projekt. Es setzt Bilder und
  Linien auf Seiten; was eine Karte ist, weiß es nicht. Der Bogenleger in
  `service` rechnet Millimeter in Stellen um — der Schreiber nimmt sie entgegen.
- `internal/msescript` importiert **nichts** aus dem Projekt, auch nicht
  `domain`. Es ist ein Sprachwerkzeug; eine Kartenentität darin wäre eine
  Vermischung, die den Interpreter unprüfbar machte.
- `internal/schriften` implementiert eine Schnittstelle aus `internal/render` und
  darf sie deshalb kennen — **dieselbe Richtung wie `storage` zu `service`**: die
  Schnittstelle steht dort, wo sie gebraucht wird, die Plattformkenntnis dort, wo
  sie hingehört. Ein Rasterer, der Schriftverzeichnisse kennt, ließe sich nur auf
  einem Rechner prüfen.
- `internal/render` bekommt einen **Zeichenplan**, keine Skripte. Wer eine
  Skriptauswertung im Renderer findet, hat die Naht gefunden, die fehlt: die
  Auswertung gehört nach `service`, das Ergebnis ist der Plan.
- `internal/svg` kennt `internal/render` — **dieselbe Richtung wie
  `schriften`**, und aus verwandtem Grund: es schreibt einen fertigen Plan als
  SVG und ist damit der **zweite** Verbraucher desselben Plans. Genau das ist
  die Probe auf die Zeile darüber — trägt die Trennung auch, wenn jemand
  anderes als der Rasterer den Plan liest? Umgekehrt geht nicht: ein Rasterer,
  der von SVG wüsste, wäre die Vermischung, die diese Regel verhindert.

  **Und `internal/service` kennt es seit dem 2026-08-22**, so wie es `pdf`
  kennt: SVG ist ein **Ausgabeformat der Ausfuhr**, und `Bildausfuhr`
  entscheidet, welches gewünscht ist. Ohne diese Kante stand das Paket mit
  Tests und voller Abdeckung da, und **niemand konnte eine SVG erzeugen** —
  angestrichen vom Zweitblick über Phase P
  (#852).
  Ein Paket ohne Aufrufer ist keine Anwendungsfunktion; das ist die erste
  Dimension aus §10, und sie steht dort zuerst, weil sie schon einmal gefehlt
  hat.
- Repository-**Schnittstellen** werden in `internal/service` definiert,
  **implementiert** in `internal/storage`. Nie umgekehrt. Das ist eine fachliche
  Zuordnung, keine Importrichtung — wegen Gos struktureller Typisierung erfüllt
  `storage` sie ohne Import von `service`.
- `internal/app` enthält keine fachlichen Entscheidungen. Es übersetzt zwischen
  Oberfläche und Services. Ein `if` über einen Domänenzustand gehört dorthin nicht.
  Die benannte Ausnahme sind die **Composition Roots**: `app.go` erzeugt die
  Repositories für das Fenster, `mcpnaht.go` die für den MCP-Server, und beide
  reichen sie den Services. **Zwei, nicht einer** — es waren immer zwei, sie
  standen bis zum 2026-08-22 nur in derselben Datei; getrennt hat sie die
  Zeilengrenze weiter unten. Die Aufzählung bleibt eine Aufzählung: was nicht
  geht, ist die Kante auf `internal/app` als Ganzes zu öffnen.
- **`internal/fenster` ist die Hülle** (ENG-02, [0132](docs/entscheidungen/0132-wails-3-als-huelle.md)):
  es kennt `app` und Wails und setzt die Nähte der Anwendung auf das Fenster;
  **`internal/app` kennt Wails nicht**, und nur `main.go` kennt die Hülle. Eine
  Entscheidung in ihr gehört eine Schicht tiefer, wie ein `if` in `app`.
- Fachliche Regeln leben in `internal/domain` (eine Entität) oder in
  `internal/service` (mehrere).
- **Keine Quelltextdatei über 400 Zeilen.** Keine Näherung, keine Ausnahme für
  Tests — sie sind der Anlass der Regel: eine Testdatei mit 1.700 Zeilen ist die
  eine Stelle, an der jede Änderung stattfindet, und niemand sieht mehr, welche
  Verhaltensklasse ungeprüft ist. Die Stufe `lint` zählt nach, in `internal`,
  `scripts`, `frontend/src` und `saat`.
  Ausgenommen wäre allein, was niemand von Hand schreibt — seit Wails 3 gibt
  es das nicht mehr (unter Wails 2 `frontend/wailsjs/`).

  **Für Fließtext gilt sie nicht.** Hier stand „keine Datei", und der Wächter
  prüfte nie etwas anderes als Quelltext — `AGENTS.md` selbst, `CLAUDE.md` und
  `docs/FORMATE.md` liegen längst darüber. Der Zweitblick über Phase 3 hat den
  Widerspruch angestrichen. Der Grund für die Grenze ist, dass eine Datei mit
  vielen Zuständigkeiten unprüfbar wird; ein Dokument hat diese Eigenschaft
  nicht, und `FORMATE.md` künstlich zu zerlegen machte den Wächter darüber
  aufwendiger, ohne etwas zu gewinnen.
- Der Schnitt folgt der Frage, die eine Datei beantworten soll, nicht der
  Zeilenzahl. Bei einer Seite heißt das: *was sie tut* steht in einem Haken,
  *wie sie aussieht* in der Komponente.

---

## 3 Testgetriebenes Vorgehen

**Verbindlich, ohne Ausnahme.** Der Zyklus lautet:

1. **Rot** — einen Test schreiben, der das gewünschte Verhalten beschreibt und
   fehlschlägt. Den Fehlschlag tatsächlich **sehen**, nicht annehmen.
2. **Grün** — die minimale Implementierung, die den Test bestehen lässt.
3. **Aufräumen** — Struktur verbessern, Tests bleiben grün.

Kein Produktivcode ohne vorher fehlgeschlagenen Test. Wer eine Zeile
Implementierung schreibt, für die kein roter Test existiert, arbeitet an dieser
Leitlinie vorbei. Das gilt für Subagenten wie für alle anderen.

### Abdeckung

| Bereich | Schwelle | Durchsetzung |
| --- | --- | --- |
| `domain`, `service`, `storage`, `app`, `bezug`, `mcp`, `msepaket`, `msescript`, `mtgjson`, `pdf`, `render` | 100 % | `scripts/ci.sh test`, harte Schwelle |
| Gesamtprojekt Go | 90 % | dito |
| Frontend | 100 % in Anweisungen, Zweigen, Funktionen und Zeilen | `vitest --coverage`, harte Schwelle |

Ausnahmen stehen in `.covignore` und brauchen dort eine Begründung. Ausgenommen
wird nur, was sich ohne echtes Fenster, echten Prozess oder echte Plattform
nicht ausführen lässt. Eine unbequeme Verzweigung ist kein Grund.

### Testebenen

| Schicht | Vorgehen |
| --- | --- |
| `domain` | reine Einheitstests, keine Vorrichtungen |
| `msescript` | tabellengetrieben über Quelltext → erwarteter Wert; dazu Fehlerfälle mit Position |
| `msepaket` | gegen **echte** Paketdateien aus `testdaten/`, nicht gegen nachgebaute Zeichenketten |
| `render` | Goldbildvergleich mit benannter Toleranz; Abweichungen werden als Differenzbild abgelegt |
| `storage` | echte SQLite-Datenbank im Arbeitsspeicher, je Test frisch migriert |
| `service` | gegen In-Memory-Doubles der Repository-Schnittstellen; zusätzlich Integrationstests gegen echtes Storage für Transaktionsverhalten |
| `app` | Fehlerübersetzung und DTO-Abbildung |
| `frontend` | Vitest und Testing Library. Die Fassade wird über `src/test/api.ts` teilweise belegt — was ein Test nicht belegt, scheitert mit Nennung der Methode, statt `undefined` zu liefern |

Tabellengetriebene Tests sind der Normalfall. Jeder Fall bekommt einen
sprechenden Namen — `t.Run("lehnt leere Bezeichnung ab", …)`, nicht
`t.Run("fall3", …)`.

### Testdaten aus fremden Paketen

`testdaten/` enthält **keine** fremden Vorlagen (§11). Wo ein Test ein echtes
Paket braucht, wird es aus einer minimalen, selbst geschriebenen Fassung
gebaut — dieselbe Syntax, eigener Inhalt. Wer gegen ein bezogenes Paket prüfen
will, tut das in einem Test hinter dem Build-Tag `fremdpakete`; er läuft nicht
in der Pipeline und zählt nicht zur Abdeckung.

### Was die Tests hier nicht sehen werden

Diese Grenzen gelten von Anfang an, weil sie aus einem Schwesterprojekt und
aus diesem teuer bezahlt sind. **Die Regel steht hier, ihr Anlass samt
Messung in [`docs/TESTGRENZEN.md`](docs/TESTGRENZEN.md).**

- **Zwischen den Prüfebenen.** Frontend-Tests laufen nie durch die
  JSON-Serialisierung, Go-Tests sehen die Struct-Marken nicht. Vor jeder
  Freigabe wird die **gebaute** Anwendung gestartet.
- **Eine Vorrichtung, die großzügiger oder schneller ist als der Betrieb**,
  prüft sich selbst. Die Zusage gehört an die Naht; `test/api.ts` antwortet
  eine Makroaufgabe später, wie die Wails-IPC.
- **Farbkontraste** meldet axe unter jsdom als *unentschieden* — die
  Sichtprüfung findet sie.
- **Die rohe Eingabe gehört in den Test**, nicht der schon umgerechnete Wert.
- **Gewartet wird auf die Wirkung**, nicht auf den Ruf und nicht auf eine
  Zusicherung, die schon vor der Antwort zutraf. Ein Wächter zählt die
  Gestalt nach (`frontend/src/wartemuster.test.ts`).
- **Vor eine Verneinung gehört ein Beleg, dass jemand zuhört** — ein zweites
  Ereignis durch denselben Weg, dessen Wirkung man sieht.
- **Vor ein Fensterereignis auf einen asynchron gebauten Baum gehört
  `effekteNachholen()`** (`frontend/src/test/effekte.ts`): `findBy…` sagt zu,
  dass ein Knoten dasteht, nicht dass ein Effekt gelaufen ist.
- **Ein Test, der nach draußen greift**, bekommt `t.Setenv("HOME",
  t.TempDir())` **und** eine umgelenkte Adresse — beides einzeln reicht nicht.
- **Der Anmeldebund dieses Rechners ist für Testläufe gesperrt**
  (`internal/schluesselbund`, erkannt an `os.Args[0]`); Kindprozesse bekommen
  ihre Umgebung an **einer** Stelle, `internal/flaechen`.

### Maschinelle Prüfung auf Barrierefreiheit

`frontend/src/barrierefreiheit.test.tsx` prüft jede Seite mit axe-core gegen
WCAG 2.2 AA. Die Seitenliste kommt aus `navigation.ts`, nicht aus einer Kopie
daneben: eine zu kurze Aufzählung ist grün, und niemand sieht es.

### Sichtprüfung

Was weder ein DOM-Test noch axe zeigt — Abstände, Ausrichtung,
**Farbkontraste**, helles und dunkles Erscheinungsbild und bei diesem
Projekt die **Kartenausgabe** —, wird angesehen. **Wie, und woran das schon
gescheitert ist, steht in [`docs/SICHTPRUEFUNG.md`](docs/SICHTPRUEFUNG.md).**

- **Zwei Wege, die nicht dasselbe können**: `npm --prefix frontend run
  vorschau` zeigt Aufbau und Farben, aber kein Datum und keine Karte;
  `make dev-browser` zeigt unter `http://localhost:34115` dieselbe Anwendung
  mit echten Daten, `make dev` sie im Fenster.
- **Wer dabei etwas ablegt, nimmt `wails-sichtpruefung`** aus
  `.claude/launch.json` — eigener Bestand samt Bildern, eigener Bund. Ein
  Wächter hält Schalter und Vorgabeort (`werkzeuge/bestandswurzeln_test.go`).
- **Kontraste werden am gerenderten Element gemessen** und stehen in
  `docs/KONTRASTE.md`; Trennlinien halten 3:1 in beiden Erscheinungsbildern.

**Und eine Änderung an der Oberfläche braucht drei Belege** (ENG-01.AK-02,
#1726):

1. **den Weg vorher und nachher** — in Schritten, vom Einstieg bis zum Ziel;
   ein Bildschirm ohne seinen Weg belegt keine Bedienung,
2. **einen Nachweis am nativen Fenster** — ein Browser zeigt dieselbe Seite,
   aber nicht die Menüleiste, den Fensterrahmen und die WebView,
3. **vollständigen Funktionszugang** — was vorher erreichbar war, ist es
   danach auch; das hält `docs/FAEHIGKEITEN.md`.

**Und ein Test belegt sich mit seinem Lauf**, nicht mit seinem Ergebnis:
Commit, Eingaben, ausgeführte Fälle. Ein grünes Häkchen ohne Protokoll ist
eine Behauptung (§4).

---

## 4 Pipeline

Ein Einstiegspunkt, lokal wie in GitHub Actions:

```bash
make ci                 # alle Stufen
./scripts/ci.sh lint    # eine Stufe
```

Der GitHub-Workflow ruft **dasselbe Skript** auf. Wer eine Prüfung ergänzt,
ergänzt sie in `scripts/ci.sh` — nicht in der Workflow-Datei.

Stufen: `format` → `types` → `lint` → `test` → `security`. **Die Teststufe
fährt zwei Läufe**: mit Race-Detektor über alles, was Goroutinen hat, und
ohne über `internal/storage` und `internal/werkzeuge`
([0094](docs/entscheidungen/0094-der-detektor-misst-nicht-wazero.md)); ein
Wächter hält die Liste ehrlich.

**Die Regeln, und warum sie so lauten, stehen samt jeder Messung in
[`docs/PIPELINE.md`](docs/PIPELINE.md):**

- **Die Pipeline läuft vor dem Push, nicht darin**: `make ci && git push`.
  Der `pre-push`-Hook (`make hooks`) **prüft** nur, ob sie für genau diesen
  Stand grün war — Kennung des Commits und sauberer Baum. `--no-verify` ist
  für Notfälle und wird im PR begründet.
- **Davor der Riegel**: `make vorpruefung` findet in Minuten, was die
  Pipeline in zwanzig findet. Was er nicht fährt, zählt sein eigener Kopf auf;
  er ersetzt `make ci` nicht.
- **Ein `make ci` je Push, nicht je Werkstück.** Mehrere Werkstücke einer
  Phase gehen auf einen Zweig; dazwischen laufen **gezielte** Läufe der
  berührten Pakete und Dateien, der Riegel einmal davor. Vorher den Platz
  prüfen (`df -h /`); `make platz` räumt Build- und Lint-Cache.
- **Ein Lauf hinterlässt einen Beweis** (`.ci-protokoll.log`, und die
  Schlusszeile nennt die rote Stufe) **und läuft allein** — eine Sperre weist
  den zweiten ab; `MSE_CI_TROTZDEM=1` hebt sie auf.
- **Was eine Stufe im roten Fall zeigt, benennt den Fund** — eine Bilanz ohne
  Namen kostet einen zweiten vollen Lauf.
- **Eine übersprungene Stufe ist kein Beleg** — der Lauf stempelt nicht —,
  und ein Messlauf hinterlässt einen: `make fremdpakete` → `docs/belege/`.

### Es gibt nur den lokalen Lauf

**Ein entfernter Lauf findet nicht statt** ([0007](docs/entscheidungen/0007-kein-entfernter-lauf.md)).
Die Workflow-Datei `.github/workflows/ci.yml` liegt als fertiger Weg da und wird
nicht ausgelöst.

Das ist eine Entscheidung mit Preis, und der gehört ausgesprochen: **geprüft ist
eine Maschine, nicht eine Plattform.** Dauerhaft unsichtbar bleiben damit

- alles Plattformabhängige jenseits der bloßen Übersetzung — gebaut ist nicht
  gelaufen,
- der Stand aus der Sperrdatei, weil lokal genommen wird, was daliegt, statt
  `npm ci` zu fahren,
- jede Eigenheit dieses Rechners, die in jedes Ergebnis eingeht,
- ein frischer Klon auf einer zweiten Maschine.

Wer eine grüne Pipeline als Beleg anführt, führt genau so viel an. Die Liste
steht ausführlich in `docs/entscheidungen/0007` und gehört in jede
Release-Notiz.

---

## 5 Git und GitHub

### Ablauf je Unteraufgabe

1. Issue anlegen — wo es schon steht, das bestehende —, Phase und Board zuordnen.
2. Branch von `main`: `feat/<issue-nr>-<kurzbezeichnung>` — **je Phase oder
   zusammenhängender Gruppe von Paketen einer**, nicht je Paket.
3. Testgetriebene Commits, **je Paket mindestens einer**, der sein Issue nennt.
4. PR mit `Closes #<nr>` für **jedes** erledigte Paket im Rumpf.
5. **Inline-Review** nach §10 durch den arbeitenden Zweig selbst.
6. Squash-Merge bei einem Paket, **Merge-Commit bei mehreren** — sonst geht
   die Zuordnung Commit → Paket im Squash verloren.

Der **Zweitblick durch Codex kommt am Phasenende**, nicht je PR — §10.

### Drei Migrationen, drei Züge

**Ein Frameworkwechsel, eine Datenmigration und eine Neugestaltung der
Oberfläche werden getrennt geplant, getrennt gezogen und getrennt geprüft**
(ENG-01.AK-03,
#1726).
Wails 2 → 3 ist ein PR, eine Schemaänderung ein anderer, ein neues Artboard
ein dritter. Wer zwei davon zusammenlegt, weiß beim ersten Fehler nicht,
welcher ihn gebracht hat — und kein Rückweg trennt sie wieder.

**Kein pauschaler Umbau**: was groß ist, wird in Schritte geschnitten, von
denen jeder für sich grün ist (§15, Satz 11).

### Branch-Präfixe

`feat/` · `fix/` · `test/` · `docs/` · `refactor/` · `chore/` · `ci/`

### Commits

Format: `<typ>(<bereich>): <beschreibung>`

```
feat(msepaket): Paketkopf mit Abhängigkeiten lesen
test(msescript): Zerteiler gegen unvollständigen Ausdruck absichern
fix(render): Maske vor dem Verlauf anwenden
```

- Typen: `feat` `fix` `test` `docs` `refactor` `chore` `ci`
- Beschreibung auf Deutsch, Kleinschreibung nach dem Doppelpunkt außer bei
  Eigennamen, kein Punkt am Ende, Imperativ oder Nominalphrase.
- **Alle Commits sind GPG-signiert** mit Schlüssel `1503F87D4C128638`, seit
  1.0.0 und nur mit der Noreply-Adresse von GitHub als Kennung;
  `user.signingkey` steht in der Einstellung **dieses Repositoriums**.
  `commit.gpgsign` ist global aktiv; nicht deaktivieren. **Freigaben signiert
  ein anderer**, `44C59428DE1F766E` aus `internal/aktualisierung/signatur.go`
  ([0205](docs/entscheidungen/0205-1-0-0-ist-ein-neuanfang.md)).
- Ein Commit ist ein abgeschlossener Gedanke. Roter Test und zugehörige
  Implementierung dürfen zusammen in einen Commit.

### Pull Requests

Rumpf enthält: was geändert wurde, warum, wie geprüft, `Closes #<nr>`.

### Schutz von main

**Seit das Repositorium öffentlich ist** (2026-11-15) schützt ein Ruleset
`main` auf dem Server: Änderungen nur über einen PR, kein Force-Push, kein
Löschen; die Admin-Rolle darf es für Notfälle umgehen. **Eine grüne Pipeline
erzwingt es nicht** — einen entfernten Lauf gibt es nicht (§4). Das hält
weiter der `pre-push`-Hook: ein direkter Push auf `main` wird abgewiesen.
Umgehung nur über `MSE_ERLAUBE_MAIN_PUSH=1` und nur für den ersten Push oder
echte Notfälle.

---

## 6 Oberfläche

Leitlinien für `frontend/`:

- Durchgängige Tastaturbedienbarkeit, sichtbarer Fokusring. Jede Aktion, die per Maus erreichbar ist, ist auch per Tastatur erreichbar.
- Command-Palette über `⌘K` für Navigation und Schnellanlage.
- Optimistische Aktualisierung mit Rücknahme bei Fehlern.
- Helles und dunkles Erscheinungsbild nach Systemeinstellung; **Bewegung höchstens
  200 ms und eine Schattenstufe nur an Überlagerungen** (Vermerk 0065, Festlegungen 10 und 11).
- Keine verschachtelten Dialoge. Bearbeitung im Seitenbereich, nicht in einem
  Modal über einem Modal.
- Leere Zustände erklären den nächsten Schritt, statt „Keine Daten" zu zeigen.
- Fehler erscheinen am verursachenden Feld, nicht als anonymer Hinweis.
- **Die Karte ist das Hauptobjekt.** Eine Änderung an einem Feld ist in der
  Vorschau zu sehen, ohne dass jemand etwas bestätigt.
- macOS-Eigenheiten werden bedient, nicht nachgeahmt: Menüleiste, `⌘,` für
  Einstellungen, `⌘W`/`⌘Q` im gewohnten Sinn, Volltextsuche über `⌘F`.

### Die Bauteilregeln sind das Prüfmaß

Die Entwürfe der Zieloberflächen 0.6.0 und 0.9.0 sind gebaut und liegen seit
1.0.0 im Entwicklungsarchiv
([0205](docs/entscheidungen/0205-1-0-0-ist-ein-neuanfang.md)); maßgeblich ist
der gebaute Stand. **Was eine Fläche einhält, sagt das Prüfmaß** in
[`docs/UIREVIEW.md`](docs/UIREVIEW.md): R1 bis R6 misst das Drehbuch `uirevue`
am nativen Fenster, R7 zählt `wahlkatalog.test.tsx`. Wer eine Fläche neu
entwirft, entwirft sie gegen diese Regeln. **Abweichen ist erlaubt und wird
dort eingetragen** — stillschweigend abzuweichen ist ein Befund, derselbe wie in
§10: eine Zusage, die niemand mehr prüft.

---

## 7 Nebenläufige Änderungen

Vier Regeln, für jede Seite, die schreibt — beim Bauen zu prüfen, nicht erst im
Review:

1. **Ein Wechsel des Bezugsobjekts räumt den offenen Kontext.** Wer Set oder
   Karte wechselt, schließt Formular, Rückfrage und geladene Nebenlisten mit.
2. **Eine Antwort gehört zu dem Stand, aus dem sie stammt.** Wechsel werden in
   einer Referenz mitgezählt; wer eine Antwort einträgt, prüft zuvor den Stand.
   Gilt für Erfolg, Fehler und Rücknahme gleichermaßen.
3. **Geschrieben ist geschrieben.** Nachladen nach einer geglückten Änderung ist
   Bestenfalls-Arbeit: ein Fehler dabei wird gemeldet, macht die Änderung aber
   nicht rückgängig. Nach einem *gescheiterten* Schreiben wird gar nicht
   nachgeladen — sonst überschreibt ein Lesefehler die Meldung, die sagt, was zu
   tun ist.
4. **Ein Vorgang je Datensatz.** Ein zweiter Aufruf, während der erste läuft,
   geht gegen einen Stand, den das Backend noch nicht kennt.

Eine Rücknahme stellt **nur den betroffenen Datensatz** wieder her, nie den
ganzen Stand von vorhin.

**Wo sie gelten, steht in `docs/SCHREIBWEGE.md`** — jede der 125 Stellen, an
denen das Fenster schreibt, mit der Gestalt, in der sie jede Regel einlöst. Ein
Wächter hält die Liste vollständig: eine neue Bindung erzwingt die Entscheidung
schreibend oder lesend, statt still in einer Gruppe zu verschwinden.

**Regel 1 und 2 haben dort drei Gestalten, und das ist der Satz, für den es die
Datei gibt**: ein `key`, der das Bauteil beim Wechsel abräumt, löst **beide**
zugleich ein — wer ihn entfernt, nimmt auch die unsichtbare weg.

Regel 2 und 4 stehen als Haken bereit (`useLauf`, `useVorgang`), dazu `useLaden`
für den Wächter jedes Ladeeffekts. Regel 1 und 3 lassen sich nicht fassen: *was*
zum offenen Kontext gehört, weiß nur die Seite, und „geschrieben ist
geschrieben" ist eine Aussage über die Reihenfolge im Aufrufer.

**Und einer eigens für dieses Projekt:** eine Kartenvorschau wird **abgebrochen**,
nicht abgewartet. Rendern kostet; wer während des Tippens vier Läufe anstößt,
zeigt am Ende den drittletzten Stand, wenn er den Lauf nicht an einen `context`
bindet.

---

## 8 Fehlerbehandlung

`internal/domain` definiert `ErrValidation`, `ErrNotFound`, `ErrConflict`.
`internal/storage` bildet SQLite-Fehler darauf ab. `internal/app` übersetzt sie
in eine für das Frontend stabile Struktur aus Code, Meldung und optionalem
Feldbezug.

Rohe SQLite-Meldungen erreichen die Oberfläche nie. Fehler werden mit `%w`
umschlossen und mit `errors.Is`/`errors.As` geprüft — nie per
Zeichenkettenvergleich. Wer einen Grund mitführen will, ohne ihn zu zeigen,
nimmt `domain.MitGrund`: der Satz ist für Menschen, der Grund fürs Protokoll
([0075](docs/entscheidungen/0075-der-satz-und-sein-grund.md)).

**Zwei Stellen zeigen ihn trotzdem, und sie stehen benannt** — das native
Startmodal und die Wiederherstellungsseite. Beide zeigen ihn **unter** dem
Satz, nicht an seiner Stelle, und beide haben denselben Grund: dort erreicht
niemand eine zweite Spur. Wer eine dritte braucht, trägt sie in
[0075](docs/entscheidungen/0075-der-satz-und-sein-grund.md) ein — sonst ist sie
ein Befund.

Eine unbedingte Zusage mit stillen Ausnahmen ist schlimmer als eine engere
Zusage (§10). **Der Anlass** steht in der Chronik, „Anlässe der Leitlinien" (WART-01, #2172).

**Fremde Daten sind der Normalfall, nicht der Ausnahmefall.** Ein MSE-Paket kann
von 2007 stammen, halb geschrieben, in einer Kodierung, die niemand mehr
benutzt. `msepaket` gibt für jeden Fehler die **Stelle** her — Datei, Zeile,
Feld — und liest so weit weiter, wie es verantwortbar ist. Ein Set, von dem 300
Karten lesbar sind und zwei nicht, wird mit 300 Karten und zwei benannten
Befunden geöffnet, nicht abgewiesen.

---

## 9 Dokumentation

Wird mit der Änderung aktualisiert, nicht danach:

| Datei | Aktualisieren wenn |
| --- | --- |
| `README.md` | Einrichtung, Kommandos oder Architektur ändern sich |
| `CHANGELOG.md` | sich für **Anwendende** sichtbares Verhalten ändert. Nicht bei innerem Umbau — dafür gibt es `docs/ROADMAP.md`. Der Punkt entsteht **jetzt**, unter `## [Unveröffentlicht]`, und bekommt seine datierte Überschrift beim Release |
| `docs/ROADMAP.md` | bei jedem Merge — Phase, Fortschritt, Entscheidung. Führt **nur Offenes** |
| `docs/CHRONIK.md` | wenn etwas fertig ist. Offene Arbeit zwischen erledigter zu suchen, lässt Erledigtes als offen führen |
| `docs/ARCHITEKTUR.md` | Schichten, Datenmodell oder Entscheidungen ändern sich — **eine Regel über gespeicherte Daten** (Bedingung, Auslöser, Fremdschlüssel, Wurzel, Index) steht zudem in `docs/INVARIANTEN.md` mit Sitz und Beleg, und jeder Auslöser dort mit seiner Invariante — ein Lauf zählt nach; eine leseintensive Liste wird in `docs/ABFRAGEPLAENE.md` **gemessen** |
| `docs/FORMATE.md` | eine Erkenntnis über ein MSE-Dateiformat kommt hinzu |
| `docs/MESSMETHODIK.md` | eine Zahl über den Abstand zu MSE2 kommt hinzu oder ihr Weg ändert sich — **das Maß, die Stichprobe und der geschichtete Schätzer stehen dort mit ihren Gleichungen**, damit eine Zahl nachlesbar ist statt behauptet |
| `docs/KONTRASTE.md` | eine Vordergrundfarbe kommt hinzu oder ändert sich — sie wird **gemessen**, nicht geschätzt, und ein Wächter hält die Datei gegen den Quelltext |
| `docs/LOKALISIERUNG.md` | eine Stelle kommt hinzu, an der **Sprache** eine Rolle spielt — sie wird dort einer von **drei Klassen** zugeordnet: kanonische Form (keine Sprache), Menschentext (die des Fensters) oder technischer Schlüssel (keine Sprache, aber eine **stabile** Ordnung). Ein Wächter hält die Liste vollständig; **welche Klasse die richtige ist**, sieht nur ein Mensch — und eine verwechselte gibt eine Liste, die auf zwei Rechnern verschieden aussieht |
| `docs/STILOPTIONEN.md` | eine `styling field:`-Art bekommt ein Bedienelement oder verliert eines — die Zahlen sind **gezählt**, nicht geschätzt, und ein Wächter hält die Namen gegen den Quelltext |
| `docs/SCHLUESSELWOERTER.md` | eine Schlüsselwortfähigkeit aus CR 702 kommt hinzu oder fällt weg — sie wird dort **gezählt** statt gefühlt, über die zehn kanonischen Sets, und der Lauf dazu steht hinter `fremdpakete`. **Zwei Zahlen gehören dazu**: wie oft ein Text das Wort **nennt** und auf wie vielen Karten es **steht** — der Unterschied war bei `flying` 45 Prozent |
| `docs/FAEHIGKEITEN.md` | eine Bindung oder ein MCP-Werkzeug kommt hinzu oder fällt weg — sie wird dort **einer Fähigkeit zugeordnet**, und ein Wächter hält beide Richtungen gegen den Quelltext |
| `docs/FLAECHENPARITAET.md` | eine **schreibende** Bindung kommt hinzu oder fällt weg, oder ein Lauf über die Flächen entsteht — sie bekommt dort ihre **Objektklasse**, und die Klasse nennt ihre Läufe für **Ergebnis, Fehlerklasse und Rolle**. Ein Wächter hält die Zuordnung, die Läufe und die Bilanz gegen den Quelltext; **ob ein Lauf vergleicht, was seine Spalte sagt**, hält er nicht (C04, #2229) |
| `docs/SCHREIBWEGE.md` | eine Bindung kommt hinzu oder fällt weg — sie wird dort als **schreibend oder lesend** eingetragen, und ein Wächter hält beide Listen gegen den Quelltext |
| `docs/FELDPRUEFUNG.md` | eine **freie Eingabe** kommt hinzu oder fällt weg — sie wird dort als **geprüft oder ausdrücklich ungeprüft** eingetragen, mit Grund, und ein Wächter hält die Zahl je Datei gegen den Quelltext. **Hier stand sie bis zum 2026-09-22 nicht**, obwohl sie dieselbe Bauform hat wie die Zeile darüber und derselbe Wächtertyp sie hält — gefunden, als er eine neue Eingabe anstrich (PLAY-03.AK-05, #1711) |
| `docs/EINSTELLUNGEN.md` | eine **Einstellung** kommt hinzu, fällt weg oder ändert Ebene, Werte, Vorgabe oder Wirkung — ihre Zeile im Schema wird **erzeugt**, nicht geschrieben: ein Wächter vergleicht beide Tafeln mit `internal/domain` und nennt im roten Fall die Zeile, wie sie lauten soll. **Die Vorrangfolge darüber hält er nicht**; wer an ihr etwas ändert, ändert den Absatz (OPS-01, #1718) |
| `docs/REGELABDECKUNG.md` | ein Kapitel der **Comprehensive Rules** wird umgesetzt, oder die CR bekommt eine neue Fassung — je Kapitel ein Stand, bei jedem umgesetzten die Zeile, an der es geschieht. **Sie löst das „Starter Set“ ab**, von dem das Regelwerk spricht: nicht eine Auswahl, die wir für genug halten, sondern das ganze Regelwerk mit einem Stand je Kapitel. Ein Wächter zählt die Zeilen und hält `README.md` dagegen |
| `docs/RETRIEVALMESSUNG.md` | eine Zahl über die Güte des Suchindex kommt hinzu — **Korpus, Frageset, Modell und Maschine stehen dabei**, und die Grenzen gelten für den **nächsten** Lauf: was nach der Auswertung aufgeschrieben wird, prüft nichts mehr |
| `docs/MSE2-OBERFLAECHE.md` | eine Einstellfläche von MSE2 kommt hinzu oder wird hier gebaut — sie ist **nachgelesen**, nicht aus dem Feldnamen erschlossen |
| `docs/PORTIERUNGSKARTE.md` | eine Einheit wird aus MSE2 portiert — Zuordnung, Kennzeichnung, Tests, Abweichungen (§15) |
| `docs/MSE2BESTAND.md` | ein Menübefehl von MSE2 kommt hinzu oder bekommt eine Entsprechung — sie ist die **Obermenge**, gegen die „echtes Subset" überhaupt zu prüfen ist. Ein Wächter hinter `fremdpakete` hält sie gegen MSEs Locale; was er **nicht** hält, ist die Einordnung |
| `docs/PARITAET.md` | eine Einheit von MSE2 wird gebaut, vertagt oder abgelehnt — sie bekommt dort eine der **sechs** Marken. Sie beantwortet, was `PORTIERUNGSKARTE.md` offen lässt: **was ist unvermessen?** Zwei Wächter halten sie, und der zweite liegt hinter `fremdpakete`, weil er den MSE2-Quelltext braucht |
| `docs/PORTIERUNG.md` | sich etwas an Windows oder Linux ändert — beschrieben, nicht gebaut (§12) |
| `docs/FASSUNGEN.md` | eine Fassung kommt hinzu oder ändert sich — eine Konstante, deren Name auf »fassung« endet, **oder** eine Kennung der Gestalt `mse4mac/<wort>/<zahl>`. Sie steht dort mit Wert, Aufstieg, Abstieg und Beleg; ein Wächter hält beide Richtungen gegen den Quelltext (#1731) |
| `docs/NETZEIGENSCHAFTEN.md` | eine Eigenschaft des Transports oder des Dienstes ändert sich — **Proxy, Protokoll, Fristen, Adresse und Port**, je Stelle und **gemessen**. Ein Wächter hält die Aufzählung der HTTP-Klienten (`internal/werkzeuge/httpklienten_test.go`): ein zweiter Transport erzwingt die Entscheidung, statt still andere Eigenschaften zu bekommen (#1533) |
| `docs/DATEIOPERATIONEN.md` | eine Dateioperation kommt hinzu oder ändert ihren Vertrag — **Eingabe, Rolle, Grenze, Ergebnis, Fehler und Nebenwirkung** je Operation und Fläche. Ein Wächter hält jeden genannten Weg gegen den Dienst und jeden Auftragsweg gegen die Seite (`internal/werkzeuge/dateioperationen_test.go`); **ob die Sätze stimmen, hält er nicht** (FMT-03, #1692) |
| `docs/HAERTUNG.md` | eine Prüfebene kommt hinzu oder ändert sich — sie trennt **statische Prüfung, lokale Pipeline, Fuzzen mit Zeit, Lauf gegen einen echten fremden Dienst und Probe an der laufenden Anwendung**. Eine grüne Pipeline ist nur die zweite davon |
| `docs/RELEASE.md` | der Freigabevorgang ändert sich |
| `DRITTANBIETER.md` **und** `docs/ABHAENGIGKEITEN.md` | eine Abhängigkeit oder ein fremder Inhalt kommt hinzu, fällt weg oder wird gehoben — im Inventar mit Fassung, Lizenz und der Stelle, die sie verantwortet; ein Wächter hält es gegen Pinnungen und Importe |
| `docs/entscheidungen/` | eine Entscheidung fällt, die man später hinterfragen wird |
| `AGENTS.md` **und** `CLAUDE.md` | eine Leitlinie ändert sich — beide, über `make leitlinien` |

Ein PR, der Verhalten ändert, aber keine Dokumentation berührt, ist
begründungspflichtig.

### Sprache

Dokumentation, Kommentare, Commit-Beschreibungen und Fehlermeldungen sind
deutsch.

**Oberflächentexte sind zweisprachig, und Deutsch ist die Ausgangssprache.** Sie
stehen unter `frontend/src/sprache/de/` und `en/`, in **fünf Katalogen**:
`allgemein`, `bausteine`, `seiten`, `hilfe`, `fehler`.

**Fünf Kataloge, nicht fünf Dateien.** Hier stand bis zum 2026-08-20 das
zweite, und `seiten.ts` ist darüber auf 405 Zeilen gewachsen — an der
Zeilengrenze aus §2 vorbei wäre es nur mit einer Ausnahme gegangen, die
niemandem nützt. Ein Katalog darf sich auf mehrere Dateien verteilen, solange
er als **einer** ankommt; die Kartenseite steht deshalb in `seitenkarten.ts`
und wird in `seiten.ts` eingesetzt. Was gleich bleibt, ist die Zahl der
Kataloge — an ihr hängt der Vertragstest gegen die Go-Seite.

**Ein Text im Quelltext ist ein Befund.** `keintextimjsx.test.ts` liest jede
Komponente und weist ab: Textknoten, beschriftende Eigenschaften — `aria-label`
wie `titel` —, Zeichenketten mit Platzhaltern und Zeichenketten in Ausdrücken.

Die Meldungen aus Go tragen einen **Schlüssel** neben dem deutschen Satz. Ein
Vertragstest hält beide Seiten zusammen: was Go setzt, steht in beiden
Katalogen, und was im Katalog steht, setzt Go.

**Nicht übersetzt wird, was aus einem MSE-Paket kommt.** Feldnamen, Spielnamen,
Stilbezeichnungen und Skriptbezeichner stehen so in der Datei; eine Übersetzung
machte jedes vorhandene Paket unlesbar. Was MSE selbst in seinen `.mse-locale`
mitbringt, wird von dort gelesen, nicht von uns nachgebaut.

**Für Bezeichner entscheidet das Vokabular, nicht die Sichtbarkeit**
([0015](docs/entscheidungen/0015-bezeichner-folgen-dem-vokabular.md)):

| Der Bezeichner benennt | Sprache | Beispiele |
| --- | --- | --- |
| einen **fremden** Begriff — MSE, MTG JSON, Go, Wails | englisch | `Card`, `StyleField`, `FindByID`, `SearchQuery` |
| einen **eigenen** Begriff dieses Projekts | deutsch | `Planbauer`, `Bogenleger`, `Zerteile`, `Bildquelle` |
| Paketinternes und Lokales | deutsch | `zeileLesen`, `umschliessen`, `fehler` |
| Testfunktionen und Testfallnamen | deutsch | `TestCardValidateBefunde`, `"leerer Kartenname"` |

**Weil MSE-Bezeichner englisch sind**: ein Feld heißt in jeder Vorlage `casting
cost`, und ein deutscher Typname daneben verschleiert, dass es dasselbe ist.
Deshalb heißt die Entität `Card` und nicht `Karte`.

Hier stand bis #115
„exportiert heißt englisch, weil die Schicht an Go und Wails grenzt". Die Regel
war in **keinem** Paket befolgt — am wenigsten in `app`, also in den
Wails-Bindungen selbst (72 deutsche gegen 57 englische Namen). Eine mechanische
Regel, an die sich niemand hält, ist keine Regel, sondern ein Vorwurf, den man
später gegen jeden erheben kann. Was sie kostet, steht im Vermerk: „fremd oder
eigen?" ist eine Frage ans Review und nicht an einen Wächter.

Kommentare erklären **warum**, nicht **was**. Ein Kommentar, der die Zeile
darunter wiederholt, wird gelöscht.

---

## 10 Review

Zwei Ebenen, und sie sind nicht dasselbe.

### Inline, während der Phase

Jeder PR wird vor dem Merge im selben Zug gelesen: Testabdeckung, Schichtenregel,
Zeilengrenze, Doku, die vier Regeln aus §7. Das ist Handwerk, kein zweites
Werkzeug.

### Zweitblick durch Codex, am Phasenende

Ist eine Phase fertig, liest **Codex** ihren gesamten Beitrag — nicht den
einzelnen PR, sondern die Naht zwischen ihnen. Zwei Aufgaben:

- **Review** der Phase vor ihrem Abschluss.
- **Eröffnen von Issues** für Befunde, die den Rahmen der Phase sprengen.

**Kleine Unterphasen einer Iteration teilen sich einen Zweitblick**, wo sie
eine Naht bilden. Er läuft im Hintergrund, während die nächste beginnt.

**Rückmeldungen werden geprüft, nicht blind umgesetzt.** Ist ein Befund falsch,
wird das begründet statt stillschweigend übergangen.

#### Die Dimensionen, nach denen gefragt wird

Der Auftrag nennt sie ausdrücklich. Ohne sie liest der Zweitblick, was ihm
auffällt — und das ist verlässlich das Handwerk.

| Dimension | Die Frage |
| --- | --- |
| **Funktionale Abdeckung** | Löst dieser Beitrag ein, wofür es die Anwendung gibt — und **woran ist das zu sehen**? |
| **Deckung durch die Quellen** | Ist jede Aussage über fremdes Verhalten **nachgelesen** — oder aus Nutzungsbeispielen erschlossen und als Spezifikation ausgegeben? |
| Nähte | Trägt der Weg **durch** die Schichten, nicht nur jede für sich? |
| Zusagen | Hält jeder Wächter, was er behauptet, und steht daneben, was er nicht prüft? |
| Fehlerwege | Ist jeder betreten worden, und bleibt kein halber Stand zurück? |
| Nebenläufigkeit | Die vier Regeln aus §7 |
| Dokumentation | Sagt sie, was heute gilt? |

**Die erste steht zuerst, weil sie am 2026-08-16 gefehlt hat.** Der Zweitblick
über Phase 7 fand sechs Befunde — Nähte, Fehlerpfade, überholte Läufe — und
**keinen** darüber, dass ein bezogenes Paket eine leere Karte zeichnet. Das war
kein Versäumnis von ihm: er wurde nicht danach gefragt. Ein Review, das nur das
Handwerk prüft, bestätigt eine Anwendung, die niemand benutzen kann
(#248).

#### Die drei Quellen gehören in den Auftrag

**Der Zweitblick liest gegen die Quellen aus §11, nicht nur gegen das eigene
Repositorium.** Ohne diesen Satz tut er es nicht — am 2026-08-16 erreichte er
von den dreien genau **eine**: die Datenpakete. MSE2 lag auf dem Rechner gar
nicht, und `magic-set-generator` stand im Auftrag nicht drin.

Was das gekostet hat, hat er selbst angestrichen: „die Packfundstellen belegen
Aufruf, Richtung und Parameternamen, **nicht die Pixel- oder Alphaformel**".
Zwei Blendformeln waren in der Richtung richtig und im Detail falsch — je
Farbkanal statt über die Helligkeit, ohne Alpha statt mit, Fehler statt
Klemmen. Nachgelesen wurde das erst, als `blend_image.cpp` wirklich dalag.

**Wo sie liegen, sagt der Auftrag.** Sie kommen **nicht** ins Repositorium
(§11); geklont wird daneben, und der Pfad steht im Auftrag an den Zweitblick.

**Genau zwei Durchgänge je Zweitblick**: ein Zweitblick, und nach der Korrekturrunde
**höchstens ein** Nachlauf. Danach ist die Phase abgeschlossen, gleich was der
Nachlauf noch findet — was materiell ist, wird behoben oder als Issue eröffnet.

Der Grund für die Begrenzung ist nicht, dass ein dritter Durchgang nichts fände.
Er fände etwas: eine Korrektur, die den **genannten** Fall behebt und die Klasse
dahinter stehen lässt, sieht wie eine Lösung aus und ist keine — und ab dem
zweiten Durchgang prüft der Zweitblick zunehmend die Antwort auf den
Zweitblick. Das ist nicht wertlos, aber es ist etwas anderes als die Prüfung der
Änderung, und es hat kein natürliches Ende.

**Was das kostet, gehört ausgesprochen:** ein Befund, der erst im dritten
Durchgang aufgetaucht wäre, bleibt bis zum Abschlussreview liegen.

**Am Ende aller Phasen steht ein eigenständiges Gesamtreview** — nicht über
eine Phase, sondern über das fertige Werk und die Nähte zwischen den Phasen.
Dort kommt an, was die Zwei-Durchgänge-Regel unterwegs durchgelassen hat.

Beim Beheben nicht den Fall behandeln, sondern die Klasse — die Frage lautet
nicht „stimmt es jetzt für dieses Beispiel?", sondern „welche anderen Beispiele
hätten denselben Weg genommen?".

### Sechs Sätze, die für beide Ebenen gelten

- **Ein Kopf, der mehr verspricht als sein Rumpf hält, ist schlimmer als gar
  keiner.** Eine Liste dessen, was ein Fund ist, liest sich wie eine Zusage —
  und niemand prüft nach, ob der Rumpf sie einlöst. **Zum Ergebnis gehört dabei
  auch das Nein**: eine Ablehnung ist eine Antwort, und wer nur den Erfolgsweg
  prüft, hat die Hälfte ausgelassen.

  **Der Anlass** steht in der Chronik, „Anlässe der Leitlinien" (WART-01, #2172).

- **Ein Wächter, der die halbe Zusage hält, macht die andere unsichtbar.** Wer
  einen Wächter schreibt, schreibt neben ihn, was er *nicht* prüft — oder besser:
  er prüft die Zusage dort, wo sie eingelöst wird. Die Gegenprobe — Eintrag
  entfernen, roten Lauf sehen — ist der einzige Beleg, dass ein Wächter greift.
- **Und eine Gegenprobe, die nichts ändert, ist keine.** Wer eine Absicherung
  entschärft und einen grünen Lauf sieht, hat zwei Lesarten vor sich: der Test
  enthält seinen Fall nicht — oder der Eingriff war wirkungslos. Erst wenn
  belegt ist, dass er das Verhalten wirklich verschiebt, sagt das Grün etwas
  über den Test.

  **Der Anlass** steht in der Chronik, „Anlässe der Leitlinien" (WART-01, #2172).
- **Auch eine Ablehnung braucht Deckung.** Wer etwas nicht baut und den Grund
  erfindet, hat keine Entscheidung getroffen, sondern eine Ausrede aufgeschrieben
  — und eine Ausrede prüft niemand nach, weil sie wie eine Begründung aussieht.
- **Die Erinnerung eines Werkzeugs ist kein Prüfergebnis.** Wer einen gemeldeten
  Befund übernimmt, ohne die Fundstelle am **heutigen** Stand nachzulesen,
  arbeitet an einem Repository, das es nicht mehr gibt. Dasselbe gilt für
  Bibliotheks-APIs — dafür steht Context7 bereit (§1).
- **Die Reichweite einer Messung ist die Reichweite ihrer Aussage.** Ein `find`
  über *ein* installiertes Paket belegt nichts über „das Pack", ein Zähler, der
  nur den Planbau durchläuft, nichts über die Rasterung. Wer eine Zahl
  aufschreibt, schreibt dazu, **worüber** sie erhoben wurde — und wer aus ihr
  eine Ablehnung begründet, prüft zuerst, ob sie so weit reicht.

  **Der Anlass** steht in der Chronik, „Anlässe der Leitlinien" (WART-01, #2172).

---

## 11 Recht, Lizenz und fremde Inhalte

Dieses Projekt steht unter **GPL-2.0-or-later** (`LICENSE`). Damit gelten drei
Regeln, und keine davon ist Formsache.

### Aus MSE2 darf übernommen werden — und die Herkunft wird genannt

Bis zum 2026-08-16 stand hier das Gegenteil: MIT, und deshalb kein Blick in den
GPL-lizenzierten Quelltext von MSE2. Diese Entscheidung ist umgedreht worden
([0019](docs/entscheidungen/0019-gpl-2-0-statt-mit.md)), weil sie das Ziel
gekostet hätte — von 5.950 Feldstilen aus `M15-Magic-Pack` deuteten 310, und die
Ursachen lagen in Sprachdetails, die man einem laufenden Programm nur sehr teuer
abliest.

**Der C++-Quelltext von MSE2 darf gelesen werden, und es darf daraus portiert
werden** — übersetzt, umgeschrieben, übernommen. Was dabei ohne Ausnahme gilt:

- **Jede übernommene Regel nennt ihre Herkunft an Ort und Stelle** — Datei und
  Fassung, im Kommentar neben dem Code. **Die Fassung ist der Commit-Hash mit
  seinem Datum** (`Fassung f9d9356 vom 2021-07-06`), und ein Wächter macht
  einen Verweis ohne sie rot
  ([0028](docs/entscheidungen/0028-die-fassung-in-einer-herkunftsangabe.md)). Nicht aus Höflichkeit: die GPL verlangt,
  dass eine Bearbeitung als solche kenntlich ist (§2 a), und ein späterer Leser
  muss unterscheiden können, was erschlossen und was übernommen wurde.
- **Kein wörtliches Einfügen von C++.** Was hier entsteht, ist Go — die
  Übernahme betrifft *Verhalten*, nicht Zeilen. Ein Block, der aussieht wie
  übersetzter C++, ist schwerer zu prüfen als einer, der die Regel beschreibt
  und sie in diesem Projekt umsetzt.
- **`docs/FORMATE.md` bleibt**, und ihre Angabe `Quelle:` unterscheidet ab jetzt
  zusätzlich, ob eine Erkenntnis aus einem **Datenpaket** oder aus **MSE2**
  stammt. Eine Erkenntnis ohne Quelle bleibt wertlos.

**Seit dem 2026-08-17 ist der Quelltext der erste Weg und nicht der letzte**
([0024](docs/entscheidungen/0024-portierung-statt-neuerfindung.md), §15). Die
Datenpakete sagen, was *in* einer Vorlage steht — Feldnamen, Zählungen,
Häufigkeiten; der Quelltext sagt, was das Programm damit **tut**. Wer die zweite
Frage aus der ersten Quelle beantwortet, errät sie.

Hier stand bis dahin das Gegenteil: „der fremde Quelltext ist die Antwort auf
eine Frage, nicht der Einstieg in eine." Was dieser Satz gekostet hat, steht im
Vermerk — sieben still ausfallende Fehler an einem Tag.

### Vier Quellen, und jede hat ihre eigene Bedingung

Seit dem 2026-08-16 sind es drei, seit dem 2026-09-15 vier. Sie stehen hier
zusammen, weil sie **nicht** dieselben Regeln haben und eine Verwechslung
teuer wäre.

| Quelle | Lizenz | Was gilt |
| --- | --- | --- |
| **Die Datenpakete selbst** — `.mse-game`, `.mse-style`, `.mse-include` | keine, fremdes Material | Gelesen wird, nichts wird übernommen. Nichts davon kommt ins Repository (unten) |
| **MSE2**, der C++-Quelltext | GPL-2.0 | Portieren erlaubt, seit [0019](docs/entscheidungen/0019-gpl-2-0-statt-mit.md). Herkunft an Ort und Stelle, kein wörtliches C++ |
| **[`fenhl/magic-set-generator`](https://github.com/fenhl/magic-set-generator)**, Rust | **MIT** | Portieren erlaubt, und die MIT-Lizenz erlaubt es großzügiger als die GPL. Trotzdem gilt hier dieselbe Herkunftsregel |
| **Die offiziellen Regelwerke von Wizards of the Coast** — die **Comprehensive Rules** und die **Magic Tournament Rules** | urheberrechtlich geschützt, keine Weitergabelizenz | **Nachgelesen wird, nichts wird übernommen.** Zitiert wird die **Regelnummer** samt Fassung, nie der Wortlaut; die Dokumente kommen **nicht** ins Repositorium. Sie sind die Quelle der Wahrheit für **Spielregeln** — und zwar die einzige (unten) |

**Warum das dritte Repositorium zählt:** es *schreibt* `.mse-set`-Dateien. Damit
ist es die einzige belegte Auskunft darüber, was in einem echten Set steht —
und ein echtes Set liegt hier nicht vor. Der erste Ertrag war sofort ein
Fehler: `stylesheet:` trägt die **Kurzform**, das Paket heißt
`<spiel>-<stylesheet>.mse-style`, und bis dahin suchte dieses Programm den
rohen Wert. Jedes von diesem Werkzeug geschriebene Set öffnete damit ohne
seinen Stil.

**Die Pflicht ändert sich nicht, nur ihr Grund.** Bei magic-set-generator
verlangt sie, dass Lizenztext und Urhebervermerk bei jeder Weitergabe
mitgehen — das steht in `DRITTANBIETER.md`. Die Herkunftsangabe im Kommentar
steht davon unabhängig da: sie ist für den Leser, nicht für den Juristen.

**Und sie verlangt zusätzlich eine Prüfung, die bei MSE2 entfällt:** MIT-Code in
ein GPL-Projekt zu übernehmen geht, umgekehrt nicht. Wer etwas aus diesem
Projekt dorthin zurückgeben will, kann das nicht ohne Weiteres.

### Und für das Spiel selbst gibt es keine Altimplementierung

**MSE2 spielt nicht.** Es ist ein Karteneditor; einen Spielkern hat es nie
gehabt — nachgemessen am 2026-09-15 für Werkstück G1
(#1458) über
`/Users/david/mse2-quelle/src`: `game_state`, `battlefield`, `graveyard`,
`mulligan`, `zone`, `tapped` je **null** Dateien.

**Damit greift §15 ins Leere, und genau das ist gefährlich.** Die Regel dort
lautet „rekonstruiere aus dem vorhandenen Quelltext"; wo es keinen gibt, bliebe
„plausibel ergänzen" — der Fehler, den §15 verhindern soll. Für Phase 5.5 und
alles, was Spielabläufe berührt, gilt deshalb:

> **Der Ablauf eines Spiels wird in den offiziellen Regelwerken nachgelesen,
> nicht erschlossen.** Eine Fachregel ohne Regelnummer ist eine Annahme.

**Zwei Dokumente, beide von Wizards of the Coast**, beide mit Datum
abgerufen am 2026-09-15:

| Dokument | Wofür | Stand beim Abruf |
| --- | --- | --- |
| **Comprehensive Rules** (CR), `magic.wizards.com/en/rules` | Zonen, Zugstruktur, Stapel und Priorität, Combat, zustandsbasierte Aktionen, Marken, Anhängsel — alles, was am Tisch gilt | Textfassung vom **2026-09-25** (`MagicCompRules 20260925.txt`), **je Kapitel abgeglichen** mit der vom 2026-08-19 (#1842, `docs/REGELABDECKUNG.md` §1) |
| **Magic Tournament Rules** (MTR), `wpn.wizards.com/en/rules-documents` | Decklisten, Sideboarding, Best-of, Zeitregeln, Rundenablauf — alles, was ein Turnier daraus macht | **2026-02-27** (`MTG_MTR_2026_Feb27_EN.pdf`) |

**Wie zitiert wird**, und es ist dieselbe Form wie bei MSE2 (§11): die
**Nummer** und die **Fassung**, im Kommentar neben dem Code.

```go
// **Eine Bibliothek ist eine geordnete Zone** — CR 401.1 (Fassung 2026-09-25).
```

Der Bestand kennt diese Form schon: CR 903.5a/b/e für die Commandergrößen,
CR 205.4a für die Supertypen, CR 305.6 für die fünf Landgrundtypen, CR 712.8a
für die Vorderseite einer zweiseitigen Karte, die MTR für die vier Tiebreaker.
**Neu ist, dass sie verbindlich wird** — und dass die Fassung mitgeht.

**Wie eine falsche Nummer aussieht, und wie oft das schon geschah**, steht
in [`docs/REGELWERKE.md`](docs/REGELWERKE.md). Der Wächter
[`internal/werkzeuge/regelwortlaut_test.go`](internal/werkzeuge/regelwortlaut_test.go)
hält den **Wortlaut** fern; ob eine Nummer das Richtige trifft, hält er
nicht — das bleibt Handarbeit.

**Drei Dinge, die dabei nicht geschehen:**

- **Kein Wortlaut im Repositorium.** Die CR ist urheberrechtlich geschützt;
  eine Regel wird **beschrieben** und mit ihrer Nummer belegt, nicht
  abgeschrieben. Dieselbe Trennung wie bei MSE2, nur aus einem anderen Grund:
  dort ist es die GPL, hier gibt es gar keine Lizenz.

  **Deshalb ist die Vorgabe jetzt sicher, statt wiederholt zu werden** (§10,
  dieselbe Wendung wie beim Anmeldebund in §3):
  [`internal/werkzeuge/regelwortlaut_test.go`](internal/werkzeuge/regelwortlaut_test.go)
  weist jedes **englische** Zitat von mehr als vier Wörtern ab, das innerhalb
  von 150 Zeichen auf einen Verweis auf eines der beiden Regelwerke folgt.
  **Was er durchlässt, steht in seinem Kopf**: kurze Begriffe — ein Supertyp,
  ein Kartenname — und die **deutsche Wiedergabe**, denn genau die verlangt
  diese Regel.
- **Kein Dokument im Baum.** Die CR wächst mit jedem Set; eine eingecheckte
  Kopie wäre am Tag der nächsten Freigabe falsch. Sie wird abgerufen, die
  Fassung wird notiert.
- **Keine Regelnummer ohne Nachlesen.** Eine erfundene Nummer ist schlimmer
  als keine: sie sieht aus wie ein Beleg (§10, „auch eine Ablehnung braucht
  Deckung").

**Und was ausdrücklich nicht daraus folgt**: dass jede Regel gebaut wird. Was
dieses Programm automatisch durchsetzt, entscheidet die Regelmatrix in G4
(#1461) —
und ein Sandkasten setzt gar nichts durch. Die Regel hier sagt nur: **was
gebaut wird, ist nachgelesen.**

### Keine fremden Vorlagen im Repository

`Full-Magic-Pack` und vergleichbare Sammlungen werden **nicht** eingecheckt:
sie wiegen über ein Gigabyte, und sechs der zehn kanonischen Sets liegen je über
227 MB — GitHub weist eine Datei über 100 MB ab, auch privat.

**Die Einwilligung zur Mitlieferung liegt seit dem 2026-08-17 vor**
([0025](docs/entscheidungen/0025-die-vorlagen-duerfen-mitgeliefert-werden.md)),
und sie ändert an dieser Regel nichts: der rechtliche Einwand fällt, die Größe
bleibt. Bezogen wird über `internal/bezug` an einem **Stand**, und die
Paketseite bietet Adresse und Stand vorausgefüllt an.

Die kanonischen MSE2-Sets liegen unter
`~/Library/Application Support/MagicSetEditor4Mac/kanonische-sets/` — bei den
bezogenen Paketen und den Referenzbildern, nicht im Baum. **Woher sie stammen,
steht in [0026](docs/entscheidungen/0026-mse2-rendert-die-referenz-selbst.md)**:
`archive.org/details/mse-canon-cards-2022-10-24`, Daten aus Scryfall mit Stand
2022-10-24, samt der Einschränkungen, die ihr Ersteller selbst nennt.

Stattdessen ein **Paketverwalter**, der Vorlagen auf dem Rechner der Nutzenden
bezieht und einbindet, dazu eine kleine, selbst erstellte Saat aus rechtlich
unbedenklichen Vorlagen unter `saat/`.

### Was die Lizenz nicht abdeckt

Die GPL-2.0 gilt für den **Code** dieses Repositoriums. Sie gilt
ausdrücklich **nicht** für Inhalte Dritter — Kartenrahmen, Mana-Symbole,
Schriftarten, Regeltexte, Namen und Marken von Wizards of the Coast oder anderen
Rechteinhabern. Das steht so in `LICENSE`, in `README.md` und in
`DRITTANBIETER.md`, und es steht in der Anwendung selbst unter „Über".

---

## 12 Plattformen und Freigabe

### Gebaut wird für macOS

Ziel ist **darwin/arm64**, ab macOS 15.2. Die Untergrenze steht in
`build/darwin/Info.plist` als `LSMinimumSystemVersion`, und ein Wächter hält sie
gegen diesen Satz — eine Zusage in zwei Dateien läuft sonst auseinander, und die
im Bündel ist die, die gilt.

**Zugesagt wird, was ein Lauf belegt** ([0184](docs/entscheidungen/0184-die-untergrenze-ist-macos-15-2-weil-nur-sie-belegt-ist.md));
gesenkt wird sie, wenn ein Prüfbündel mit tieferer Grenze dort durchläuft.

Der Bau läuft über `scripts/release.sh` und erzeugt ein `.app` und ein `.dmg`.

**Freigegeben wird erst, wenn alle neun Phasen durch sind.** Nicht am Ende jeder
Phase: eine Fassung, die nichts kann, ist keine Fassung, und eine Versionsnummer
je Phase erzeugt eine Historie, die niemand braucht. Der Weg ist ab Phase 1
gebaut und erprobt; benutzt wird er in Phase 8 — der **letzten** Phase in der
Bauabfolge, auch wenn Phase 9 die höhere Nummer trägt (§13).

Windows und Linux werden in `docs/PORTIERUNG.md` **beschrieben** — Bauwege,
Abhängigkeiten, offene Punkte — und **nicht gebaut, nicht geprüft, nicht
ausgeliefert**. Die Pipeline übersetzt trotzdem für beide Ziele (`GOOS=windows`,
`GOOS=linux`), damit plattformabhängiger Code nicht unbemerkt bricht. Ein Bau
ist keine Erprobung; er sagt nur, dass der Code dort überhaupt entsteht.

### Keine Notarisierung

Es gibt **keinen Apple-Developer-Account** und es soll keinen geben. Damit:

- Die Anwendung wird **ad-hoc signiert** (`codesign --sign -`). Auf arm64 ist das
  keine Wahl, sondern Voraussetzung — ein unsigniertes Binärprogramm startet dort
  nicht.
- Sie wird **nicht notarisiert** und trägt kein Developer-ID-Zertifikat.
- Beim ersten Start weist Gatekeeper sie ab. Nutzende geben sie **einmalig** in
  den Systemeinstellungen frei: *Datenschutz & Sicherheit* → Abschnitt
  *Sicherheit* → **„Dennoch öffnen"**. Ab macOS 15 ist das der einzige Weg; das
  frühere Öffnen über das Kontextmenü gibt es nicht mehr.
- `README.md` und `docs/RELEASE.md` beschreiben diesen Weg mit Bild. Er ist kein
  Mangel, den man verschweigt, sondern eine Eigenschaft, die man erklärt.

Wer diese Entscheidung ändern will, ändert zuerst
`docs/entscheidungen/0003-keine-notarisierung.md`.

---

## 13 Phasen und Issues

Die Arbeit ist in **neun Phasen** geschnitten; `docs/ROADMAP.md` führt sie.

**Acht davon standen von Anfang an, die neunte kam am 2026-08-15 dazu**
([0009](docs/entscheidungen/0009-phase-9-fremdformate-und-schnittstelle.md)).
Ihre Nummer sagt, wann sie geschnitten wurde, nicht wann sie gebaut wird: Phase
9 wird vor Phase 8 gebaut, denn Phase 8 ist die Freigabe und in ihr entsteht
nichts Neues mehr. Wer eine Phase hinzufügt, tut es so — mit Vermerk, mit
sichtbarer Begründung in der Roadmap, und nicht, indem er eine bestehende größer
macht.

**Die Issues für alle Phasen bestehen von Anfang an.** Das ist Absicht: ein
Vorhaben, dessen spätere Hälfte nur in einer Überschrift steht, verschiebt seine
Entscheidungen nach hinten, wo sie teurer sind.

**Sie werden am Ende jeder Phase nachgeschärft.** Verbindlich, nicht bei
Gelegenheit — der Phasenabschluss ist erst vollständig, wenn er stattgefunden
hat:

1. Zweitblick durch Codex über die abgeschlossene Phase (§10).
2. Befunde entweder beheben oder als Issue eröffnen.
3. **Die Issues durchgehen, die von der Phase abhängen** — was sie gezeigt
   hat, steht danach darin: neue Unteraufgaben, geänderter Zuschnitt,
   geschlossene Punkte, die sich erledigt haben. Die übrigen bleiben liegen.
4. `docs/ROADMAP.md` und `docs/CHRONIK.md` fortschreiben, `CHANGELOG.md` datieren.

Was in Schritt 3 **nicht** geschieht: eine Phase wird nicht stillschweigend
größer. Kommt Umfang hinzu, bekommt er ein eigenes Issue und eine sichtbare
Begründung in der Roadmap.

---

## 14 Subagenten

**Die Unteraufgaben einer Phase werden inline abgewickelt**, nicht von
Subagenten: jeder Bericht muss nachgelesen werden, und das kostet mehr, als
er spart. Ein Subagent kommt nur dazu, wo er Zeit spart — **höchstens zwei
zugleich, nur Sonnet**. Für ihn gilt alles oben, und vier Dinge zusätzlich:

- **Ein Subagent bekommt eine Unteraufgabe, nicht eine Phase.** Der Zuschnitt ist
  das Issue.
- **Testgetrieben heißt auch für ihn: den roten Lauf sehen.** Ein Bericht „Tests
  geschrieben und implementiert" ohne den roten Zwischenstand ist kein Beleg.
- **Der Bericht eines Subagenten ist eine Behauptung, bis sie nachgelesen ist.**
  Wer sein Ergebnis übernimmt, führt `make ci` selbst aus und liest den Diff.
- **Und er portiert, statt zu entwerfen** — die zehn Regeln dafür stehen in
  §15.

---

## 15 Portierung statt Neuerfindung

**Der C++-Quelltext von MSE2 ist die fachliche und technische
Referenzimplementierung dieses Projekts** ([0024](docs/entscheidungen/0024-portierung-statt-neuerfindung.md)).
Er liegt unter `/Users/david/mse2-quelle`, geklont **neben** dem Repository
(§11). Die Regel dazu ist ein Satz:

> **Erfinde keine bestehende Anwendungslogik neu. Rekonstruiere sie aus dem
> vorhandenen C++-Quelltext und portiere sie nachvollziehbar.**

Wo Verhalten, Algorithmen, Zustandsübergänge, Datenformate, Validierungen,
Berechnungen oder Randfälle dort schon implementiert sind, wird **dort
nachgelesen**. Nicht aus Namen abgeleitet, nicht aus einem laufenden Programm
erschlossen, nicht plausibel ergänzt.

**Warum das eine eigene Regel ist:** an einem einzigen Tag sind sieben Fehler
behoben worden, die alle dieselbe Gestalt hatten — erschlossen statt
nachgelesen, plausibel und falsch, und **jeder fiel still aus**. Die Liste steht
im Vermerk.

### Zwölf Sätze, die für jede Änderung gelten

1. **Zielarchitektur ist Go und TypeScript für macOS** (§2, §12).
2. **Der C++-Quelltext ist die Quelle der Wahrheit** für vorhandene Fachlogik.
3. **Keine Erfindung vorhandener Fachlogik.**
4. **Keine stillschweigende Änderung bestehender Semantik.** Wer sie ändert,
   sagt es und begründet es.
5. **Die Migration erfolgt inkrementell und nachvollziehbar** — ein Werkstück,
   ein Issue, ein PR (§5).
6. **Vor der Umsetzung wird die Altimplementierung lokalisiert und verstanden**
   — Datei, Klasse, Funktion, Aufrufer.
7. **Jede Abweichung vom Altverhalten wird dokumentiert und begründet.**
8. **Tests entstehen gegen das Verhalten der Altimplementierung**, soweit es
   sich festhalten lässt (§3).
9. **Windows-Abhängigkeiten werden ausdrücklich erkannt** und durch
   plattformneutrale oder macOS-eigene Mittel ersetzt — nach ihrem **Zweck**,
   nicht API gegen API.
10. **Architekturverbesserungen sind erwünscht, Verhaltensänderungen nicht** —
    *Semantik konservativ, Architektur modern.*
11. **Kein großflächiger Umbau ohne vorherige Analyse.**
12. **Was funktioniert, wird nicht ohne Grund neu geschrieben.**

### Die Reihenfolge je Werkstück

1. Die Stelle im Altcode finden.
2. Aufrufer und Abhängigkeiten ansehen.
3. Datenstrukturen und Zustände bestimmen.
4. Ein- und Ausgaben festhalten.
5. Sonderfälle und Fehlerwege erfassen.
6. Vorhandene Tests des Altprojekts lesen.
7. Das Verhalten durch **eigene** Tests festhalten — den roten Lauf sehen (§3).
8. Erst dann in Go schreiben.
9. Gegen das Altverhalten vergleichen.
10. Die Portierungskarte fortschreiben.

**Ist das Verhalten unklar, wird weiter gelesen** — Aufrufketten verfolgen,
Tests des Altprojekts ansehen, Beispieldaten prüfen. Ein plausibles Verhalten zu
erfinden ist der Fehler, den diese Regel verhindert.

### Und es gibt ein Orakel

**MSE2 beantwortet Fragen über MSE-Script selbst:**

```bash
magicseteditor --cli <SET>.mse-set -q --raw
```

Es liest Ausdrücke von der Standardeingabe und gibt das Ergebnis zurück. Wer
wissen will, was `sort_text(order:"in_place((WUBRG))", "U/W/x")` ergibt, fragt —
statt es aus `spec_sort.cpp` zu erschließen. Gebaut wird MSE2 nach
[0026](docs/entscheidungen/0026-mse2-rendert-die-referenz-selbst.md).

**Der erste Anlauf von #415
erwartete `W/G`, wo MSE `G/W` gibt** — geraten, obwohl der Quelltext dalag. Das
Orakel entschied es in einer Minute. Eine erwartete Zahl, die man erfragen kann,
wird erfragt.

### Wie C++ nach Go abgebildet wird

Verhalten wird portiert, **nicht Syntax**. Eine Zeile-für-Zeile-Übersetzung ist
weder verlangt noch erwünscht (§11 verbietet wörtliches C++ ohnehin).

| C++ | Go |
| --- | --- |
| Klasse, die Daten hält | `struct` |
| abstrakte Schnittstelle | `interface` — **dort, wo sie gebraucht wird** (§2) |
| RAII | `defer` oder ausdrücklicher Lebenszyklus |
| Ausnahmen | `error` als Rückgabewert (§8) |
| STL-Container | Slice und Map |
| Threads | Goroutinen **nur**, wenn es fachlich nötig ist |
| Mutex | `sync` |
| Enum | typisierte Konstanten |
| Namensraum | Paket |

### Vier Kennzeichnungen, und eine davon ist gefährlich

| Marke | Heißt |
| --- | --- |
| **ported** | Verhalten aus dem Altcode übernommen |
| **adapted** | Verhalten für die neue Plattform angepasst — begründet |
| **new** | tatsächlich neue Funktionalität, für die es keine Entsprechung gibt |
| **uncertain** | das Verhalten des Originals ist noch nicht geklärt |

**`new` darf nicht benutzt werden, um eine unterlassene Suche zu verdecken.**
Wer sie setzt, sagt dazu, **wo** er gesucht hat.

### Die Portierungskarte

`docs/PORTIERUNGSKARTE.md` führt die Zuordnung: neue Einheit, C++-Quelle,
Kennzeichnung, Tests, Abweichungen. Sie wird **mit** der Änderung
fortgeschrieben, nicht danach (§9).

**Sie ist keine Vollständigkeitsliste.** Neun Phasen sind gebaut worden, bevor
diese Regel galt; die Karte füllt sich von vorn nach hinten, angefangen bei dem,
was portiert **wird**.

### Portierungsbasis, Kompatibilitätsreste, Produktentwicklung

**Drei Arten von Arbeit, und nicht jede Regel dieses Abschnitts gilt für
alle drei** (ENG-01.AK-01,
#1726):

| Art | Was sie ist | Wo sie steht | Was gilt |
| --- | --- | --- | --- |
| **Portierungsbasis** | was aus MSE2 übernommen und gegen es geprüft ist | `docs/PORTIERUNGSKARTE.md`, Marke `ported` oder `adapted` | eine Änderung folgt den zwölf Sätzen oben |
| **Kompatibilitätsreste** | was MSE2 kann und hier fehlt, falsch oder ungeklärt ist | `docs/PARITAET.md`, #1656 | ein Befund bleibt, bis ein Beleg ihn schließt — **eine Umbenennung erledigt keinen** |
| **Produktentwicklung** | was es in MSE2 nicht gibt: Tisch, Handel, Netz, KI, Sammlung — **und die Härtung eigener Wege** (Abbruch, Grenzen, Zugang) | Marke `new`, samt dem Ort, an dem gesucht wurde; Härtung ohne Eintrag | die Spezifikation im Issue — für Spielregeln die CR (§11). **Die Reihenfolge je Werkstück entfällt**, wo MSE2 die Naht gar nicht hat |

**Die Unterscheidung schützt in beide Richtungen.** Wer einen
Kompatibilitätsrest als neue Funktion ausgibt, erfindet, was MSE2 schon
beantwortet hat; wer eine neue Funktion an MSE2 misst, sucht eine Quelle,
die es nicht gibt, und erfindet dann eine.

### Für Subagenten

Zusätzlich zu §14 gilt für jeden Agenten und Subagenten:

1. **Vor der Umsetzung den C++-Originalcode untersuchen.**
2. **Quelldateien, Klassen, Funktionen und Datenstrukturen benennen** — im
   Kommentar am Code, mit Fassung (§11).
3. **Unterscheiden** zwischen belegtem Altverhalten, technischer
   Portierungsentscheidung und notwendiger Annahme.
4. **Annahmen auf das Nötigste beschränken.**
5. **Keine Fachlogik aus Funktionsnamen, Oberflächentexten oder Vermutungen
   ableiten.**
6. **Änderungen klein und prüfbar halten.**
7. **Tests für portiertes Verhalten schreiben oder erweitern.**
8. **Nach der Änderung Tests, Linter und Bau ausführen** — `make ci` (§4).
9. **Die Gesamtarchitektur nicht eigenmächtig ändern.**
10. **Bei Widersprüchen zuerst den Altcode und seine Tests ansehen.**

**Ein Subagent entwirft keine Fachlogik.** Er portiert sie.

---

## 16 Fremde Dienste und was über Fassungen hinweg gilt

**Neu am 2026-08-23**, mit dem Ziel 0.2.0
([0043](docs/entscheidungen/0043-neun-phasen-bis-0-2-0.md)). Bis dahin gab es
genau **einen** Netzverbraucher — `internal/bezug` — und genau **eine** Fläche
nach außen neben dem Fenster. Beides ändert sich, und beides braucht eine Regel,
bevor die zweite entsteht.

### Ein fremder Dienst wird gekapselt, nicht eingewoben

Ein Klient für einen fremden Dienst ist ein **Blatt** (§2): er kennt
`internal/domain` und sonst nichts aus diesem Projekt. Er hat drei Schichten,
und sie werden nicht vermischt:

| Schicht | Was sie tut |
| --- | --- |
| **Transport** | HTTP, Fristen, Wiederholung, Drossel, Zwischenspeicher |
| **Fremdes DTO** | die Gestalt, die der Dienst schickt — mit seinen Feldnamen |
| **Zuordnung** | die Übersetzung auf `domain`, an **einer** Stelle je Spiel |

**Die dritte Schicht ist die, die man weglässt.** Wer den fremden Verbund
direkt in `service` reicht, hat die Struktur eines fremden Anbieters in seiner
Fachlogik — und ändert sie der Anbieter, ändert sich hier alles. Das Muster
steht schon da: `internal/mtgjson` hält seine Zuordnung an einer Stelle je
Spiel, und §2 sagt es dort ausdrücklich.

**Klient, Uhr und Zwischenspeicher werden eingesetzt, nicht erzeugt.** Sonst
braucht jeder Test das Netz, und was das wert ist, sagt §3.

### Sechs Sätze, ohne die kein Dienst dazukommt

1. **Die Bedingungen sind nachgelesen, nicht erinnert** (§1, §10). Grenzen,
   Pflichtkopfzeilen, Zwischenspeicherfristen und Nutzungsbedingungen stehen im
   Vermerk oder im Issue, mit Abrufdatum. Eine Auskunft aus dem Gedächtnis ist
   kein Prüfergebnis.
2. **Die Anwendung nennt sich.** Ein `User-Agent`, der dieses Programm benennt,
   ist bei manchen Diensten Pflicht und überall die Mindesthöflichkeit.
3. **Jede Anfrage hat eine Frist und einen Abbruch.** `http.DefaultClient` hat
   keine — wer ihn nimmt, hat keine.
4. **Eine Grenze wird eingehalten, bevor sie greift.** Gedrosselt wird
   **unter** der genannten Zahl; ein 429 wird gelesen und nicht durchbrochen.
   Fachliche 4xx werden **nicht** wiederholt.
5. **Ohne Netz arbeitet die Anwendung weiter.** Ein Dienst, dessen Ausfall das
   Bearbeiten anhält, ist falsch eingebaut. Der letzte gute Stand bleibt, und
   er sagt, wie alt er ist.
6. **Nichts wird still übernommen.** Was von außen in ein Kartenfeld will,
   zeigt vorher, was sich ändert, und trägt danach seine Herkunft.

### Drift wird erkannt, nicht erlitten

**Pflichtfelder streng, Zusatzfelder tolerant.** Ein neues Feld beim Anbieter
bricht nichts; ein **verschwundenes Pflichtfeld** bricht laut. Ein unbekannter
Aufzählungswert wird als **Befund** geführt und nicht verschluckt — das ist
dieselbe Regel wie in §8 für fremde Paketdaten.

Geprüft wird gegen **eingecheckte, selbst geschriebene** Beispiele; ein Lauf
gegen den echten Dienst steht hinter einem Build-Tag, läuft nicht in der
Pipeline und zählt nicht zur Abdeckung — wie `fremdpakete` (§3).

**Und die Netzgrenze bleibt eine Aufzählung.** `internal/schichten/netz_test.go`
führt, wer das Netz erreichen darf. Ein neuer Eintrag ist eine Entscheidung und
kein Nebeneffekt: er wird begründet, und wo er einen Vermerk berührt, wird der
**fortgeschrieben und nicht umgangen**.

### Was über Fassungen hinweg zugesagt wird

**Die Regeln stehen in
[0033](docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md),
alle zehn** — vier für Schema und Datei, sechs für die Flächen. Sie stehen dort
und nicht hier, weil sie eine **Entscheidung** sind und keine Arbeitsweise: sie
haben ein Datum, einen Anlass und ein Verworfen.

Was hier steht, ist der Satz, der beim Arbeiten greift:

> **Wer etwas an MCP, an der Kommandozeile oder an REST ändert, ordnet die
> Änderung ein** — verträglich, erweiternd oder brechend — **und schreibt das
> in den PR.** Wer es nicht einordnen kann, hat die Fläche nicht verstanden.

Die Wails-Naht ist davon ausgenommen und zwar ausdrücklich: sie ist eine innere
Naht zwischen zwei Hälften **derselben** Auslieferung; sie einzufrieren hieße,
die Oberfläche einzufrieren.

**Wie viele Bindungen es sind, steht hier nicht.** Es stand hier, und es war
falsch — „64", als es 142 waren, und niemand hat es gesehen, weil eine Zahl in
einer Leitlinie unter keinem Wächter steht. Die Zahl steht in
[`docs/FAEHIGKEITEN.md`](docs/FAEHIGKEITEN.md), wird dort gezählt, und ein
Wächter hält jede Flächenzahl im ganzen Repositorium dagegen
(`internal/werkzeuge/flaechenzahlen_test.go`, #1287): **sie stimmt, oder ihr
Absatz nennt ein Datum.**

### Dieselbe Fähigkeit, überall dieselbe Antwort

**Eine neue Fähigkeit landet nicht an einer Fläche.** Was das Fenster kann,
können Kommandozeile und MCP — oder es steht begründet, warum nicht (etwa ein
Dateiwahldialog, den es ohne Fenster nicht gibt).

Ein Wächter zählt es. **Was er nicht hält**, steht neben ihm: er sieht, dass
eine Fähigkeit vorhanden ist, nicht dass sie dasselbe tut. Dafür gibt es je
Fähigkeit einen Lauf, der **denselben** Arbeitsablauf über jede Fläche fährt und
dieselben fachlichen Fehler erwartet.

---

## 17 Freigaben für die Arbeit an 0.8.0

**Erteilt vom Auftraggeber am 2026-10-09**, um 0.8.0 über die Phasen des
Dach-Issues #2204
zu verwirklichen: **Shell-Befehle, VM-Steuerung, Computer Use, Netzwerk**,
die offene **Schlüsselbundfreigabe** und alles, was die **Ein-Klick-Einrichtung**
von `mse`, `mse-mcp` und `mse-rest` aus dem Fenster braucht — samt der Wahl,
auf welchen Schnittstellen ein Dienst lauscht. Wofür jede gilt und **was keine
davon deckt** (echte Bestände und Konten, Systemeinstellungen, Kennwörter,
unangekündigte große Downloads), steht in [`docs/FREIGABEN.md`](docs/FREIGABEN.md).

**Sie ändern keine Regel aus §1 bis §16.** Sie nehmen die Rückfrage weg,
**ob** ein Werkzeug benutzt werden darf — nicht die Frage, **ob es das
Richtige misst**.

<!-- graft:start -->
## Graft — repo context graph

This repo is indexed in `graft/`: small linked markdown nodes that explain each
system and carry exact file:line spans, kept in sync with the code through git.

For ANY task here — understanding how something works, finding where code lives,
or scoping a change — get context from the graph before grepping or opening
source files. Re-ask freely (it's cheap) and reuse literal identifiers you
already have (symbol, error string, file name) as the query. New to this repo?
Run `graft map` first — a token-budgeted orientation (dir clusters, hubs,
hotspots), no LLM, no key.

- Run `graft ask "<your question>" --source` → ranked nodes with the relevant
  code spans inlined (each hit's ≤8-line crux by default; `--full` for whole
  definitions when the crux isn't enough). Match the tool to the task shape:
  for understanding or editing, the top node IS the answer — cite its
  `covers:` file:line spans and edit straight from `--source`. For
  exhaustive tasks ("every occurrence / every caller of this pattern"), ranked
  results are top-N, not complete — run `graft grep "<literal>"` instead
  (exhaustive over indexed files, grouped by enclosing symbol), falling back
  to raw `grep -rn` only for unindexed files.
- `graft skeleton <file>` → every definition's signature + span, ~10× cheaper
  than reading the file; use it to skim an API surface.
- `graft callers <symbol>` gives precomputed, exact edges — who calls this.
  Add `--direction out` for what it calls, or `--depth N` to walk
  transitively for the full blast radius. For structural questions, skip
  ranking and use this directly.
- Or browse: `graft/INDEX.md` lists every node; follow the links.
- Monorepos and folders of multiple repos rank fairly across sub-projects —
  hits carry `[scope/]` labels naming which one they're from. Narrow with
  `graft ask "<task>" --in <scope>/` once you know where you're working.

If a returned span is truncated ("+N more lines"), open the file at that exact
range before finalizing. Only open source files when a node genuinely lacks a
needed detail, and then at the exact file:line the node points to — never
re-read whole files.

After big code changes, refresh the graph with `graft build` (deterministic,
no API key, $0).
<!-- graft:end -->
