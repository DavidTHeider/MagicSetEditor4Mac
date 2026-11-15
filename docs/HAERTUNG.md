# Härtung — was geprüft ist, und womit

**Werkstück O11 von Phase 5.2**
(#1435).

Dieses Dokument beantwortet **eine** Frage: wenn jemand sagt „das ist geprüft",
**womit** ist es geprüft? Die Abnahme von O11 verlangt genau diese Trennung —
statische Prüfung, lokale Pipeline, Fuzzen mit Zeit, Lauf gegen einen echten
fremden Dienst und Probe an der laufenden Anwendung sind **fünf** verschiedene
Aussagen, und eine grüne Pipeline ist nur die zweite davon.

**Was hier steht, ist eine Buchführung und kein Wächter.** Wer eine Ebene
ergänzt, trägt sie ein; dass die Zahlen darin stimmen, halten die Wächter, die
in jeder Zeile genannt sind.

---

## 1 · Statische Prüfung — ohne die Anwendung zu starten

Läuft in `make ci`, Stufen `format`, `types`, `lint` und `security`.

| Was | Womit |
| --- | --- |
| Übersetzbarkeit für **drei** Ziele | `go build` für darwin, windows, linux; dazu die Tags `fremdpakete` und `netzlauf` |
| Stil und Fehlerklassen | `golangci-lint` (darunter `gosec`, `govet`, `errcheck`, `prealloc`), `gofumpt`, `eslint`, `tsc --noEmit` |
| Bekannte Schwachstellen | `govulncheck` (Go), `npm audit` (Frontend) — **und die beiden beißen verschieden weit**: `govulncheck` macht **jeden** Fund rot, meldet aber nur, was von diesem Code aus wirklich **erreichbar** ist; `npm audit` läuft mit `--audit-level=high` und lässt damit `moderate` und darunter **stehen**. Nachgemessen am 2026-09-21 an [GHSA-82fw-gwwq-j7x9](https://github.com/advisories/GHSA-82fw-gwwq-j7x9): drei mittelschwere Funde in `vitest`, `@vitest/coverage-v8` und `@vitest/mocker` sind durch die grüne Stufe gelaufen. Behoben durch den Hub auf **4.1.11** (ENG-04, #1729) — **die Schwelle selbst bleibt, wo sie ist**, und das ist eine Entscheidung und kein Versehen: jeder mittelschwere Fund in einer fremden Entwicklungsabhängigkeit hielte sonst die Pipeline an, auch wo es keine Fassung ohne ihn gibt. Wer sie senkt, senkt sie in `scripts/sicherheitslage.sh` und sagt dazu, was mit einem Fund ohne Fassung geschieht.<br>**Seit ENG-04 hat jeder Lauf einen von vier Zuständen** — `ohne-befund`, `befund`, `nicht-erreichbar`, `gescheitert` (`scripts/sicherheitslage.sh`). Ein Scanner, der nicht antwortete, wird **übergangen** und nicht grün, und ein übergangener Lauf stempelt nicht (§4). **Der Freigabebeleg nennt die Lage** samt Kennungen und Zahlen je Schwere — die mittelschweren Funde unter der Schwelle stehen damit wenigstens dort, wo jemand nachliest ([`ABHAENGIGKEITEN.md`](ABHAENGIGKEITEN.md)) |
| Schichtenregel (§2) | `internal/schichten` — Importrichtung, Composition Roots, Netzgrenze |
| **Wer hereinlässt** | `internal/mcp/keinlauscher_test.go` — weder `internal/mcp` noch `cmd/mse-mcp` enthält `net.Listen`, `http.ListenAndServe` oder `http.Server{`. Die Netzgrenze daneben führt, wer **hinaus** darf; dies ist die andere Richtung ([0106](entscheidungen/0106-drei-vertrauensgrenzen-und-nur-eine-wird-gebaut.md)). **Gelesen wird der Quelltext und nicht das Verhalten**: ein Lauscher hinter einem Schalter zeigt in diesem Lauf keines |
| Zeilengrenze (§2) | `scripts/ci-lint.sh` über `internal`, `scripts`, `frontend/src`, `saat` |
| Zusagen der Dokumentation | die Wächter in `internal/werkzeuge`: Fähigkeiten, Autorisierungsmatrix, Schreibwege, Feldprüfung, Kontraste, **verdünnte Farben**, Fuzzziele, Flächenzahlen, Vermerke, Pinnung |

**Was diese Ebene nicht sagt**: ob der Weg durch die Schichten trägt. Sie liest
Quelltext und Dokumente gegeneinander — kein einziger Aufruf geht dabei durch
eine Naht.

## 2 · Lokale Pipeline — die Anwendung im Testprozess

Läuft in `make ci`, Stufe `test`.

| Was | Womit |
| --- | --- |
| Fachkern, Dienste, Bestand, Naht | `go test -race ./...` mit den Schwellen aus §3 (100 % in elf Paketen, 90 % im Ganzen) |
| Oberfläche | `vitest --coverage`, 100 % in Anweisungen, Zweigen, Funktionen und Zeilen; axe-core gegen WCAG 2.2 AA |
| **Echte Prozesse** | `internal/flaechen` — die Kommandozeile, der MCP-Server und der Dienst werden **gestartet** und über ihre eigenen Ströme gefahren; dieselbe Fähigkeit muss an jeder Fläche dieselbe Antwort und **denselben fachlichen Fehler** geben (§16) |
| **Eine laufende Sitzung** | `internal/flaechen/mcpprozess_test.go` — **ein** `mse-mcp` über mehrere Rufe, und der Bestand wird **zwischen** ihnen angefasst. Die Zeile darüber startet je Ruf einen Prozess; was ein Ruf hinterlässt, sieht dort keiner. Damit prüfbar: dass die Ausstattung je Ruf wirklich je Ruf entsteht (A14, #1453) |
| **Was ohne Netz noch geht** | `internal/flaechen/nullnetz_test.go` — unter »nur lokal« mit **eingerichteter** Cloud geht über **vier** Wege (Katalog, Messung, Vorschlag, Einschätzung) **null** Anfragen hinaus, gezählt an der Gegenstelle; das ist zugleich die Antwort auf „wird eine abgewiesene Anfrage wiederholt". Und der Bestand bleibt benutzbar: Lesen und Schreiben gehen weiter. **Assets stehen nicht darin** — ein Bildvorschlag lebt im Prozess und gibt es nur am Fenster ([0102](entscheidungen/0102-wo-ein-erzeugtes-bild-liegt.md)) |
| **Ein wirklich gebauter Index** | `internal/flaechen/indexlauf_test.go` — die Probelaufzeit **bettet ein** (`/api/embed`, `/api/show` mit `model_info`), und eine selbst geschriebene Sammeldatei liegt daneben. Damit läuft `mse indexaufbau` in der Pipeline, und der **Fortsetzungszeiger** und der **laufende Aufbau** sind prüfbar — beides stand vorher hinter `netzlauf` (A14, #1453). **Was er nicht prüft, ist die Güte**: die Vektoren kommen aus einer Rechnung des Laufs; wie gut echte Treffer sind, steht in [`RETRIEVALMESSUNG.md`](RETRIEVALMESSUNG.md) |
| **Zwei Prozesse über einen echten Draht** | `internal/flaechen/handelsdraht_test.go` — zwei `mse` mit **getrennten Beständen**, ein TCP-Anschluss mit TLS 1.3 und eine Handelsnachricht, die wirklich ankommt (#1634). Die Zeilen darüber starten echte Prozesse, aber jeder spricht mit **sich**; jede Prüfung eines Transports endete davor bei einem Doppel oder bei `internal/leitung.Schleife` im selben Prozess. **Was er sofort gefunden hat**: `Handelspost` schlug den Stand nach und endete bei „gibt es nicht" — und ein Vorgang, den der Empfänger noch nicht kennt, ist **jeder erste**. Ein Handel kam damit bei dem, dem er angeboten wurde, **nie** an, bei grüner Pipeline und 100 % Abdeckung. **Und `handelsabnahme_test.go` geht hin und zurück** (`T12`, #1477): beide Seiten stimmen zu, bereiten vor und übergeben, bis sie auf `uebergabeOffen` stehen. **Auch das fand sofort einen Fehler** — eine **Zustimmung konnte nicht reisen**, weil `Nimmt` die nächste Revision verlangt und eine Stimme zu derselben hinzutritt. Das Protokoll und seine Grenzen stehen in `ABNAHME-T12.md` (Entwicklungsarchiv) |
| Saatkorpus der Fuzzziele | derselbe Lauf: `go test` fährt jedes `FuzzX` mit seinen Seeds als gewöhnlichen Test — die Regressionsprobe, nicht das Fuzzen |

**Was diese Ebene nicht sagt**: nichts über eine zweite Maschine, nichts über
einen frischen Klon, nichts über ein echtes Fenster
([0007](entscheidungen/0007-kein-entfernter-lauf.md)).

**Und vier Läufe dieser Ebene hängen an einer Erlaubnis dieses Rechners.**
Drei in `internal/entdeckung` und einer in `internal/app` fahren mDNS über
`lo0` mit dem echten Netzstapel. Unter macOS ab 15 braucht der laufende
Prozess dafür die Erlaubnis für das **lokale Netzwerk**
(*Systemeinstellungen → Datenschutz & Sicherheit → Lokales Netzwerk*); fehlt
sie, weist der Kernel **jede** Multicastsendung ab — nachgemessen am
2026-09-20 für IPv4 wie IPv6, auf jeder Schnittstelle.

**Seitdem steht ein Tor davor**, und es überspringt **laut**: die Meldung
nennt die Ursache und den Weg dorthin, damit niemand ein grünes Protokoll für
eine geprüfte Aussage hält. Was dieses Tor kostet, gehört dazu: an einem Tag
ohne die Erlaubnis ist die Entdeckung im lokalen Netz **nicht** geprüft —
nicht von dieser Ebene, und von keiner anderen.

**Und das gilt auch für den Draht.** Beide Prozesse laufen auf **dieser**
Maschine, über `127.0.0.1`; ungeprüft bleibt, was ein echtes Netz
dazwischenlegt — Namensauflösung, MTU, eine Fritz!Box, ein DERP-Relais. Das
ist `T12` (#1477),
und es bleibt eine Handprüfung.

## 3 · Fuzzen mit Zeit — außerhalb der Pipeline

**Sechzig Ziele in siebenundzwanzig Paketen**; die Aufrufe, die
Ausführungszahlen je Lauf und jeder gefundene Fehler stehen in
[`RELEASE.md`](RELEASE.md), **welche Fläche jedes liest und welche ohne
Ziel sind**, im Register [`FUZZZIELE.md`](FUZZZIELE.md) (QA-06). Ein Wächter hält beide Richtungen: kein Ziel ohne
Aufruf, kein Aufruf ohne Ziel.

**Zehn davon kamen mit A19**
(#1550,
2026-09-13) und schließen eine Lücke, die A14 angestrichen hatte: **keines**
der neun Blätter, die A1–A14 gebaut haben, hatte ein Ziel — und sie lesen
fremde Antworten, die schneller und öfter kommen als eine Datei von der
Platte.

**Vier davon nehmen Bytes, fünf eine `httptest`-Gegenstelle** (sie haben
keinen exportierten Zerteiler), und der zehnte ist der **gzip-Weg** von
`mtgjsonpreise`: eine Packbombe und ein abgeschnittener Strom sind der
Missbrauchsfall dieses Blatts.

**Was sie fanden**: eine **Klasse über vier Blätter** — ein Modelleintrag ohne
Kennung gab einen Befund, der sich nicht benennen konnte —, und ein **Angebot
ohne Blueprint** in `cardtrader`, das als vollwertiges Angebot durchkam.

Sie lesen, was von außen kommt — Dateien von 2007, Archive aus dem Netz, eine
eingebettete Sprache, fremde Schemata, einen Rückruf auf der Schleifenadresse,
ein CSV aus einem Marktplatz und den Dienst als Ganzes.

**Die Kampagne von C02** (#2227):
zehn Minuten je Ziel, in zwei Hälften zu je dreißig, weil ein Ziel alle Kerne
belegt und eine Hälfte damit rund fünf Stunden braucht. **Hälfte A** steht in
[`belege/fuzzkampagne-c02a.md`](belege/fuzzkampagne-c02a.md) (Stand
`61df27a1e`, 2026-10-10): neunundzwanzig ohne Fund, **ein Fund** in
`FuzzLiesInhalt` — Prüfung und Leser von MTG JSON sahen einen Schlüssel in
anderer Schreibung verschieden. Behoben mit `b2613e3a7`, die Eingabe liegt als
Saat unter `internal/mtgjson/testdata/fuzz/` und läuft seither in jedem
`go test` mit. **Hälfte B steht aus**; C02 bleibt offen, bis sie da ist. Was
eine Hälfte nicht findet, ist damit nicht ausgeschlossen.

**Ein langsames Ziel braucht eine kurze Minimierzeit.** `go test -fuzz`
verkleinert jede neue Eingabe bis zu 60 Sekunden, bevor sie zählt; wer je
Durchlauf eine Datei schreibt, verbringt damit die ganze Laufzeit. Mit der
Vorgabe fand `FuzzAuspacken` in einer Minute **keine** neue Eingabe, mit
`-fuzzminimizetime 5s` in zwei Minuten 165 — gemessen am 2026-09-23 (FMT-06,
#1695); die Aufrufe in `RELEASE.md` tragen den Schalter.

**Und ob die Tests einen echten Fehler sehen, fragt der Mutationslauf**
(`make mutationslauf`, QA-05): zwölf Wachen an den Invarianten aus
[`RISIKOKATALOG.md`](RISIKOKATALOG.md) — Rolle, Nonce, Rahmenlänge,
Pfadflucht, Kennwort, Migration, Doppelzustellung, halbe Verbuchung,
Schlangentiefe, Fassung, Sitzungsfreigabe — werden einzeln abgeschaltet, und
jedes Mal muss ein benannter Lauf rot werden. Am 2026-10-01: **12 von 12**
([`belege/mutationen.md`](belege/mutationen.md)).

**Was diese Ebene nicht sagt**: ob ein Ziel etwas taugt. Was ein Fund ist,
steht im Kopf jedes Ziels, festgelegt **vor** dem ersten Lauf — und gelesen
wird das von einem Menschen.

## 4 · Lauf gegen einen echten fremden Dienst

**In `make ci` geht kein einziger Aufruf hinaus** — `fm serve` (unten) rechnet auf diesem Rechner und öffnet keinen Port. Was das Netz braucht, steht
hinter dem Build-Tag `netzlauf`, läuft nicht in der Pipeline und zählt nicht
zur Abdeckung (§16).

**Und die Kartendienste ohne Konto haben eine Frist** (F04,
#2246):
Scryfall, magicthegathering.io, MTGJSON (Karten) und MTGJSON (Preise) stehen
zusätzlich hinter `dienstdrift`, und `make dienstdrift` legt ihren Beleg unter
`docs/belege/dienstdrift.json` ab. **`make belege` wird rot, wenn er älter als
30 Tage ist** oder ihr Klient sich seither geändert hat
(`docs/belege/dienstdrift.plan.json`). Was dort nicht steht, hat einen Grund:
die Aktualisierung braucht ein GitHub-Token, der Paketbezug ist auf einen
Stand gepinnt und lädt 316 MB, GraphQL, CardTrader, Scrydex und die
Cloudmodelle brauchen Konten. **Der erste Lauf am 2026-10-10** an `497d7d3`:
9 von 9 Fällen bestanden, Scryfall samt Zeittest in 1,4 s
(`docs/belege/dienstdrift.json`). Die Daten der übrigen Zeilen darunter
bleiben von Hand geführt.

| Dienst | Stand |
| --- | --- |
| Scryfall | gegen das Original gelaufen; **zuletzt am 2026-10-10** (`make dienstdrift`, `docs/belege/dienstdrift.json`, 9 von 9 Fällen des Laufs). Das Datum des allerersten Laufs steht hier nicht |
| magicthegathering.io | gegen das Original gelaufen ([0056](entscheidungen/0056-drei-weitere-kartendienste.md)); **zuletzt am 2026-10-10** (`docs/belege/dienstdrift.json`, derselbe Lauf). Das Datum des allerersten Laufs steht hier nicht |
| **Paketbezug** | **gelaufen am 2026-10-04** gegen `codeload.github.com` (`internal/service/paketbezug_netzlauf_test.go`, #2049): über den Weg der Paketseite `Basic-M15-Magic-Pack` am gepinnten Stand `df1bf72` — 13 s, 5.858 Dateien und 13 Stile, 278,5 MB ausgepackt, Prüfsumme `sha256:0bd0a06a…`, `Magic - Fonts` dabei, zehn ausführbare Dateien übergangen. **Gefunden**: die Liste nannte fünfmal „build.bat" ohne Pfad — sie nennt jetzt den Pfad im Paket. **Und ein Zweig ist auch am Original kein Stand**: `main` schlägt GitHub als Marke nach und findet keine. Das Archiv lag in einem Wegwerfverzeichnis und ist gelöscht |
| **Aktualisierung** | **gelaufen am 2026-10-04** gegen `api.github.com` (`internal/aktualisierung/netzlauf_test.go`, #2049): **ohne Marke** antwortet das private Repositorium mit 404, und der Klient macht daraus das fachliche Nein; **mit Marke** die Freigabe 0.5.0 vom 2026-09-20, Abbild 302.714.048 Bytes, und die Prüfsummenliste mit **geprüfter Signatur** gegen den eingebetteten Schlüssel. **Kein Abbild geladen** — über den Weg sagt die signierte Liste dasselbe. **Was daraus folgt**: wer keinen Lesezugriff hat, erfährt von keiner neuen Fassung — **bis zum 2026-11-15**; seither ist das Repositorium öffentlich, und ohne Marke antwortet es mit `v1.0.0` (Nachtrag zu 0190) |
| Anmeldeanbieter (OAuth) | **nie** gegen das Original gelaufen und **BLOCKED** (#2049): ein echter Fluss braucht je Anbieter eine eigene registrierte Anwendung (`client_id`) und eine Anmeldung durch einen Menschen. Geprüft sind Fluss und Rückruf gegen einen eigenen Server im Test |
| Fernleitung (tailcat) | **Auf einem Rechner über echtes DERP belegt** (SEC-21, #2162): `internal/fernleitung/fernweg_netzlauf_test.go` hält einen Fernstrom offen, gewöhnlich und mit `TS_DEBUG_ALWAYS_USE_DERP=1` (`docs/belege/sec21-fernweg.md`). **Zwei Rechner in getrennten Netzen bleiben außerhalb des Auftrags** (#2049, #1456) — ein direkter Weg durch ein NAT ist damit nicht belegt |
| **MTGJSON (Preise)** | **gelaufen am 2026-09-14** gegen `https://mtgjson.com/api/v5` (`internal/mtgjsonpreise/netzlauf_test.go`, Build-Tag `netzlauf`). `AllPricesToday.json.gz` kommt an, ist gzip und wird **strömend** gelesen; mit leerem Crosswalk kommen null Stände zurück — das ist die Zusage „überspringt, was der Bestand nicht kennt, **ohne es zu lesen**". Mit einer echten Kennung kamen **drei Stände** (`mtgjson:cardmarket · paper · retail · EUR · nonfoil`, beobachtet 2026-09-12), und alle sieben Teile einer Preisreihe waren gefüllt ([0093](entscheidungen/0093-eine-preisreihe-hat-sieben-teile.md)). **Die Kennung wird geholt, nicht eingetragen** — aus `LEA.json` desselben Anbieters am selben Tag; eine eingetragene wäre eine Behauptung über fremde Daten, die rot wird, ohne dass sich hier etwas ändert. **Nicht gelaufen ist `AllPrices.json`**: über ein Gigabyte, und über den Vertrag sagt sie nichts mehr als die kleine Datei |
| **MTGJSON (Karten)** | **gelaufen am 2026-09-14** gegen `api/v5/LEA.json` (`internal/mtgjson/netzlauf_test.go`). **295 Karten aus „Limited Edition Alpha"**, keine Schema- und keine Fassungsbefunde, und **keine einzige Karte ohne bekanntes Feld** — das wäre die Drift. Dazu **32 Befunde über Schlüssel, die die Zuordnung nicht kennt** (`uuid`, `legalities`, `printings`, `manaValue`, `purchaseUrls`, `skuIds`, `sourceProducts` …); sie sind der Entwurf und kein Mangel: unbekannte Schlüssel werden roh mitgeführt, damit ein Rundlauf sie nicht verliert (§8). **Eine Setdatei und nicht `AllPrintings`** — über die Gestalt sagt sie dasselbe und kostet ein Tausendstel |
| MTGJSON (GraphQL), CardTrader, Scrydex | **nie** gegen das Original gelaufen — **der Lauf steht seit #2049 bereit** (`internal/{mtggraphql,cardtrader,scrydex}/netzlauf_test.go`, Zugang aus `MSE_NETZLAUF_MTGJSON`, `MSE_NETZLAUF_CARDTRADER`, `MSE_NETZLAUF_SCRYDEX_SCHLUESSEL` und `…_TEAM`) und meldet ohne Zugang **BLOCKED**. Es liegt an den Zugangsdaten, nicht an der Größe: GraphQL gibt ohne Marke HTTP 200 mit `errors[0] „Access denied"` (das **Schema** ist am 2026-09-02 von Hand geprüft), CardTrader antwortet ohne Marke 401, Scrydex verlangt Schlüssel **und** Team. Geprüft ist die Zuordnung gegen eingecheckte, selbst geschriebene Beispiele (0056, A4 #1443) |
| **Ollama** (lokale KI-Laufzeit) | **gelaufen am 2026-09-12** gegen Ollama 0.33.3 (`internal/ollama/netzlauf_test.go`, Build-Tag `netzlauf`). Fassung, Modelle, feste Form und die **Vermessung eines Einbetters** kommen an: `gemma3:270m` antwortete `{"name": "Blitzschlag"}` gegen das echte Schema, `all-minilm` ergab Dimension 384, Pooling `mittel`, normalisiert, Zerteiler `bert`, und ein zu langer Text wurde wirklich abgewiesen. **Hier stand bis dahin „nie gelaufen"** — Ollama war auf diesem Rechner nicht installiert; A15 (#1514) braucht einen echten lokalen Einbetter, und die Installation wurde dafür freigegeben. **Nicht in der Pipeline**: der Lauf braucht eine Installation, die kein zweiter Rechner haben muss ([0098](entscheidungen/0098-was-ollama-belegt-und-was-nicht.md)) |
| **Anthropic** (Cloudmodelle) | **nie** gegen den echten Dienst gelaufen (A8, #1447), und hier liegt es nicht am Rechner: ein Lauf brauchte ein Konto **mit Guthaben**, und jeder Ruf kostete Geld. **Seit #2049 steht ein Lauf bereit**, der nur die Modellliste fragt — sie kostet nichts — und ohne `MSE_NETZLAUF_ANTHROPIC` **BLOCKED** meldet (`internal/anthropic/netzlauf_test.go`); ein Lauf, der rechnen lässt, bleibt einer eigenen Freigabe vorbehalten. Geprüft ist alles gegen selbst geschriebene Antworten: Auth und Anfrageform, minimale, erweiterte und kaputte Antwort, Refusal, unbekannter Grund, Abbruch, 429 mit und ohne `retry-after`, der 400 der selbst gesetzten Ausgabengrenze. Der Dienstvertrag trägt `Nachweis: nil`, und die Einstellungsseite sagt es ([0099](entscheidungen/0099-was-anthropic-belegt.md)) |
| **OpenAI** (Cloudmodelle) | **nie** gegen den echten Dienst gelaufen (A9, #1448) — derselbe Grund wie eine Zeile darüber: ein Lauf brauchte ein Konto **mit Guthaben**. **Seit #2049 steht ein Lauf bereit**, der nur die Modellliste fragt — sie kostet nichts — und ohne `MSE_NETZLAUF_OPENAI` **BLOCKED** meldet (`internal/openai/netzlauf_test.go`); ein Lauf, der rechnen lässt, bleibt einer eigenen Freigabe vorbehalten. Geprüft ist alles gegen selbst geschriebene Antworten: Auth und Anfrageform (samt `store: false`), die vier Ausgänge über ihre **drei** Fundstellen, eine Ablehnung neben `status: completed`, ein 429 in seinen **drei** Bedeutungen, die Kopfzeilen als **Dauer** statt als Zeitpunkt. **Auch die Messung selbst ist nie gegen einen echten Anbieter gelaufen** — ihr Weg über alle vier Flächen ist geprüft, ihr Ergebnis an einem Doppel ([0100](entscheidungen/0100-was-openai-nicht-sagt.md)) |
| **xAI** (Cloudmodelle) | **nie** gegen den echten Dienst gelaufen (A10, #1449) — derselbe Grund wie die zwei Zeilen darüber: ein Lauf brauchte ein Konto **mit Guthaben**. **Seit #2049 steht ein Lauf bereit**, der nur die Modellliste fragt — sie kostet nichts — und ohne `MSE_NETZLAUF_XAI` **BLOCKED** meldet (`internal/xai/netzlauf_test.go`); ein Lauf, der rechnen lässt, bleibt einer eigenen Freigabe vorbehalten. Geprüft ist alles gegen selbst geschriebene Antworten: die Modellliste mit ihren Modalitäten, die **vier** Ausgänge samt `end_turn`, die Bildform **dieses** Weges und die Deutung jedes dokumentierten Status. **Was dabei ungeprüft bleibt, ist die Gestalt des Fehlerrumpfs** — sie steht in der Dokumentation nirgends, deshalb liest dieses Paket tolerant und deutet aus dem Status |
| **Apples Gerätemodell** (`fm serve`) | **gelaufen am 2026-10-04** hinter dem eigenen Tag `fmlauf` (`internal/applefm/fmlauf_test.go`; `docs/belege/fmlauf.md`): Name und Flavortext auf Deutsch und Englisch, eine Ablehnung durch die Schutzregeln, Kaltstart 0,06 s, Antworten 0,3–1,1 s, **kein TCP-Port** am Prozess. Dazu die Güte am Testsatz (`docs/belege/fmguete.md`) und die Hüllenprobe `ai05` am nativen Fenster. **Seit AI-02 (#1743) auch in `make ci`, aber nur dort, wo `/usr/bin/fm` liegt**: `scripts/ci-geraetemodell.sh` startet `fm serve` und prüft, dass es antwortet und auf keinem TCP-Port lauscht; ohne `fm` wird der Schritt gelb übersprungen. Die übrigen Testläufe sperrt der Riegel weiter für `fm` und ihre Kindprozesse ([0166](entscheidungen/0166-was-fm-serve-belegt.md)) |

**Und der Schlüsselbund gegen ein fremdes Programm desselben Nutzers**
(SEC-05, #2146),
**gemessen am 2026-10-07** an einem künstlichen Bund in einem Lume-Gast
([`belege/sec05-schluesselbund.md`](belege/sec05-schluesselbund.md)): die
Einträge dieser Anwendung vertrauen `/usr/bin/security`, und **jedes Programm,
das unter demselben Nutzer läuft, kann sie über `security` lesen, ohne dass
macOS fragt** — Zugänge fremder Dienste, Anmeldegeheimnisse, den zweiten
Faktor und den Schlüssel der Peeridentität. Eine Bindung an die vier
Programme hält das ab, überlebt aber als Ad-hoc-Signatur keine Fassung
([0181](entscheidungen/0181-der-schluesselbund-bleibt-ungebunden-solange-ad-hoc-signiert-wird.md));
**mit einem eigenen Herausgeber überlebt sie** — gemessen am 2026-10-11 im
Gast ([`belege/i01-bindung.md`](belege/i01-bindung.md)) —, **gebaut wird sie
nicht**: fremder Code, über `DYLD_INSERT_LIBRARIES` in ein gebundenes, nicht
gehärtetes Programm geladen, liest den Eintrag ebenso still (E5, derselbe
Beleg), und ein Schlüssel, der ohne Menschen signiert, signiert auch für ein
fremdes Programm
([0185](entscheidungen/0185-ein-eigener-herausgeber-truege-die-bindung.md),
[0187](entscheidungen/0187-kein-eigener-herausgeber.md)).
Ein gesperrter Anmeldebund schützt auch diese Einträge.

**Und die Biometrie am echten Sensor** (SEC-19,
#2153),
hinter dem eigenen Tag `sensorlauf` (`internal/biometrie/sensorlauf_test.go`),
**gelaufen am 2026-10-07**: geladen sind genau `tmc/apple` v0.6.18 und
`purego` v0.11.0-alpha.6, Touch ID ist da, eine **Frist** endet nach 4,01 s
mit einem Fehler und ein **Abbruch** mitten im Dialog nach 2,01 s — beide
ohne späten Erfolg. **Nicht gelaufen**: die Ablehnung durch den Menschen am
Dialog und die Zustimmung. Den Dialog zeigt `coreautha`, und den bedient kein
Werkzeug ohne Bedienungshilfen-Zugriff — eine Sicherheitseinstellung, die
nur der Eigentümer des Rechners setzt. Ohne Ablehnung endet
`TestEineAblehnungIstKeinErfolg` an seiner Frist **rot**, nicht grün.

**Und ein Lauf ohne fremden Dienst, der trotzdem hier steht**: „Zwei Macs
auf einem" (`internal/app/zweimac_netzlauf_intern_test.go`,
`internal/flaechen/zweimac_netzlauf_test.go`) — zwei Anwendungen über echtes
mDNS, TCP und TLS, verzögert um 10 bis 100 ms. **Hinter `netzlauf`, weil er
Multicast ins Netz schickt** und auf der Netzwerkberechtigung dieses Rechners
hängt. Gelaufen am 2026-10-01; er fand zwei Fehler, die die Pipeline nicht
sah, und seine Zahlen stehen in
[`NETZEIGENSCHAFTEN.md`](NETZEIGENSCHAFTEN.md#zwei-macs-auf-einem--gemessen-am-2026-10-01).

**Was diese Ebene nicht sagt**: ob ein Anbieter sein Schema morgen ändert.
Dagegen steht die Regel „Pflichtfelder streng, Zusatzfelder tolerant" und ein
Befund statt eines stillen Werts (§16).

## 5 · Probe an der laufenden Anwendung

| Was | Womit |
| --- | --- |
| Aussehen, Abstände, beide Erscheinungsbilder | Sichtprüfung über `make dev` und `http://localhost:34115`, Abzüge über CDP |
| **Farbkontraste** | am gerenderten Element gemessen, über einen Canvas und nicht über ein Textmuster; das Ergebnis steht in [`KONTRASTE.md`](KONTRASTE.md) |
| Arbeitswege und Ladezeiten | `scripts/arbeitswegmass.mjs` gegen einen Messbestand mit 10.000 Exemplaren |
| Die Bindungen am echten Fenster | `scripts/nahtlauf.mjs` — er baut die Anwendung und ruft sie über die Brücke |
| **Last und Überlast an Dienst und Werkzeugserver** | `scripts/lastlauf.mjs` (`make lastlauf`, QA-08 #1739) — baut `mse-rest` und `mse-mcp`, startet sie je Profil frisch über einem **eigenen** Bestand und fährt Lesen, Schreiben, Rendern, Einfuhr, lange Aufträge, MCP über stdio und über HTTP, gemischte Klienten, zwei Identitäten, eine Überlast mit Erholung und langsame Klienten — **mit und ohne** Kontingente. Durchsatz, p50/p95/p99, Fehlerrate, Anteil der Drosselung, Speicher; die Zahlen und was sie nicht sagen stehen in [`LASTMESSUNG.md`](LASTMESSUNG.md). **Was er nicht ist**: eine zweite Maschine — Klient und Dienst teilen sich die Kerne |
| Was der Spieltisch **rechnet** | `scripts/tischmass.mjs` — Median, p95, p99 und Maximum über 200 Läufe; die Zahlen stehen in [`TISCHLEISTUNG.md`](TISCHLEISTUNG.md) |
| Was der Spieltisch **zeichnet** | `scripts/tischframezeit.mjs` — 240 Bilder an echter GPU, Bildabstand **und** Zeichenzeit getrennt; der Renderer wird ausgelesen und genannt |
| **Was WebKit tut** — Tastatur, Maus, Vorschlagslisten, WebGL, Vorleser | `scripts/webkitprobe.sh` (#1828) — ein Drehbuch mit **echten AppKit-Ereignissen** in einem WKWebView, eingerichtet wie in Wails 3 (`wk2`), und WebCores **Zugänglichkeitsbaum** samt Fokusansage im eigenen Prozess (`wk1`). **Keine Freigabe nötig**: Ereignisse gehen nur ans eigene Fenster. Die Selbstprobe (`make webkitprobe`) hat ein rotes Drehbuch, das rot sein muss; **die Abnahmen** (`make webkitabnahme`) bauen die Anwendung, legen einen frischen Prüfbestand an und fahren jedes Drehbuch unter `scripts/webkitprobe/abnahme/` — der Beleg steht in [`belege/webkitprobe/`](belege/webkitprobe/uebersicht.md). **Was sie nicht ist**: das Wails-Fenster selbst — IPC, Menüleiste, Rahmen, Dateidialoge und Ablegen gehören nicht dazu — und nicht VoiceOvers eigene Satzbildung |
| **Gezielte Sicherheitsprüfung** — Rechte-/Profilwechsel, Object-ID, Tokenentzug, Origin/Rebinding, Pfad-/Archivangriff, SSRF, Pairing-Replay, Datenoffenlegung | je Prüffall ein ausführbarer Lauf gegen die echte Fläche, nicht ein Scanner — [`PENTEST.md`](PENTEST.md) führt Scope, Prüffälle (mit Datei und Testname), Befunde mit Einstufung und Nachtest (QA-09, #1740). **Ein grüner Scanner wiegt kein offenes hohes Risiko auf** |
| **Was das native Fenster tut** — Leiste, Kürzel, Dateidialoge, Finder, Ziehen, der Tisch im Wails-Fenster | `scripts/huellenprobe.sh` (#1727) — das Bündel mit der **Abnahmeleitung** (Bauart `abnahme`), gestartet über LaunchServices; die Leitung liest die eigene Leiste, löst Einträge aus, legt Tastendrücke und Klicks in die eigene Ereignisschlange, öffnet und bricht Dateidialoge ab und spielt ein Ablegen nach, Wails' MCP liest die Seite. **Keine Freigabe nötig**: alles geschieht im eigenen Prozess. `make huellenprobe` fährt jedes Drehbuch unter `scripts/huellenprobe/`, der Beleg steht in [`belege/huellenprobe/`](belege/huellenprobe/). **Was sie nicht ist**: die Hand vor AppKit — Tastatur, Maus, Finder —, die Wahl in einem Dateidialog (sie kommt als vorbereitete Antwort) und die Zwischenablage des Rechners (kopiert wird in eine Ersatzablage). **Und ein Bild bei schlafendem Bildschirm**: dann gibt es keinen Fensterabzug, und der Beleg nennt jede Lücke — in ihrer Zeile und gezählt im Kopf — statt abzubrechen (`scripts/huellenbeleg.mjs`, seit DOM-04). **Aber eine Webansicht, die niemand sieht, zeichnet auch nicht**: am 2026-10-10 lief ein voller Lauf bei ausgeschaltetem Bildschirm, und zwanzig Drehbücher wurden an Fristen rot, die wie Befunde aussahen. Seitdem hält `huellenprobe.sh` den Bildschirm wach (`caffeinate -dimsu`), und schaltet ihn trotzdem jemand aus, liest der Lauf das Energieprotokoll und vermerkt die Ausphase im Kopf jedes betroffenen Belegs und im Manifest — **das Urteil bleibt**, denn wie die Anwendung bei hellem Bildschirm bestanden hätte, weiß niemand (C03, `scripts/bildschirmzeit.mjs`). **Seit 0.7.0 gilt zudem**: jedes Drehbuch läuft in einem eigenen Zuhause, und ein Abdruck der echten Orte dieses Rechners steht in jedem Beleg (T01, `scripts/huellenorte.mjs`); jede Anfrage hat eine Frist und nennt ihre Phase (T02); ein Bündel, das nicht aus dem Stand der Baueingänge im Baum gebaut ist, wird abgewiesen, `mse` vor dem Lauf frisch gebaut, und jeder Lauf hinterlässt unter `.huellenprobe/laeufe/<kennung>/` Bilder, Belege und ein Manifest mit HEAD, offenen Dateien, Bauabdruck, Programm- und Oberflächenabdrücken und Umwelt — die jüngsten zehn bleiben (T06, `scripts/baustand.mjs`) |
| **Jede Bindung über die echte Naht** — Vertragsrundgang, Reichweite, Zustände | `scripts/huellenprobe/vertrag.mjs` (B01, #2220) ruft im Abnahmebündel **jede** gebundene Methode über die Wails-IPC — mit Gruppen, die ihre Vorbedingungen anlegen und `gelungen` erwarten, und mit Ersatzargumenten für den Rest; das Rufbuch der Hülle (`internal/fenster/rufbuch.go`) zählt je Drehbuch Name, Ausgang und Dauer, `scripts/bindungsreichweite.mjs` bilanziert **erreicht** und **gelungen** gegen den ganzen Vertrag, je Klasse, und nennt die Quote der Bedienwege daneben. **Rufstörungen** (`internal/fenster/rufstoerung.go`) verzögern Bindungen oder lassen sie fachlich scheitern; `zustaende.mjs` (B03, #2222) zeigt damit jede der 25 Seiten in sechs Zuständen. **Was das nicht ist**: ein Bedienweg — ein Ruf am Vertrag sagt nicht, dass ein Mensch die Bindung über die Oberfläche erreicht — und keine Bewertung, ob ein Leerzustand **gut** erklärt; das sagen die Bilder |
| **Was die Schutzrichtlinie abweist** | jede Hüllenprobe (SEC-17, #2164) — `scripts/huellenschutz.mjs` liest nach jedem Drehbuch das Protokoll jedes seiner Orte und streicht jeden Verstoß gegen die CSP an, den das Drehbuch nicht ausdrücklich erwartet; `sec17` belegt, dass WebKit sie über Wails' eigenes Schema durchsetzt ([0179](entscheidungen/0179-die-webview-laeuft-unter-einer-gemessenen-schutzrichtlinie.md)). **Was das nicht belegt**: Wege, die kein Drehbuch geht |
| **Was es neben MSE2 kostet** — Zeit, CPU, Speicher, Dauerbetrieb, Lastgrenze | `node scripts/leistungsvergleich.mjs` (QA-04, #1735) — **die gebauten** `mse`, `mse-mcp` und `mse-rest` von heute und von 0.5.0 gegen das gebaute MSE2, dieselben Sets, dieselben Pakete, jedes Programm mit eigenem Zuhause. Was er findet, steht in [`LEISTUNGSVERGLEICH.md`](LEISTUNGSVERGLEICH.md), der Beleg in [`belege/leistungsvergleich.json`](belege/leistungsvergleich.json). **Was er nicht ist**: eine Messung am Fenster — das zeichnet mit einem Zeichenwerk für alle Karten; die kopflosen Flächen halten seit #2005 eines **je Bestand und Stand** im Prozess vor (`internal/app/zeichenvorrat.go`) — `mse-mcp` und `mse-rest` bauen es beim ersten Ruf und behalten es, ein einzelner `mse`-Aufruf ist ein eigener Prozess und baut es für sich. **Und ein Vergleich misst nicht die Drosselung**: gegen heute läuft er mit abgeschaltetem Guthaben (`grenzen.kontingente aus`, einmal je Lauf und danach wieder im vorgefundenen Zustand, auch nach einem Abbruch); nur der Teil `last` fährt mit ihm, denn dort ist die Drossel der Gegenstand. **Ein Fall mit Befund — zu wenige Bilder, ein beschädigtes Bild, eine andere Größe als bei MSE2 — fällt aus dem Vergleich der Zeiten**, und die Bilanz am Ende des Laufs nennt ihn |

**Und eine Grenze ist mit G3 dazugekommen** (#1460):
**die Browservorschau unter `wails dev` verliert den WebGL-Kontext wirklich.**
Gemessen am 2026-09-16 — PixiJS meldete „Could not retrieve shader source", und
ohne abgeschalteten Ticker wurden daraus 4.700 Konsolenmeldungen, bis die Seite
stehenblieb. Sie ist damit die **bessere** Probe für den Rückfall ohne Renderer
und die **schlechtere** für den Renderer selbst. Gemessen wird der deshalb in
einem eigenen Chrome mit `--use-angle=metal`, der den Kontext hält und sich
„ANGLE Metal Renderer: Apple M5" nennt — **der Name wird ausgelesen**, weil
headless Chrome sonst SwiftShader nimmt und die CPU ein zweites Mal misst.

**Was diese Ebene nicht sagt**, und das ist die wichtigste Zeile dieses
Dokuments: sie misst **die Entwicklungsfassung** — bis ENG-02 `wails dev`,
seitdem `make dev-browser`. Dort läuft React unter `StrictMode` (jeder
Ladeeffekt zweimal); Go lief unter `wails dev` mit `all=-N -l` (ohne
Optimierung und ohne Inlining) und wird seit ENG-02 gewöhnlich übersetzt —
**Zahlen von vorher und nachher sind deshalb nicht dieselbe Messung**. Eine Dauer von dort ist keine Zusage an Nutzende — die Herleitung
steht in [0092](entscheidungen/0092-eine-sicht-rechnet-dort-wo-die-liste-liegt.md).

## 6 · Was auf keiner Ebene steht

- **Eine zweite Maschine.** Kein entfernter Lauf, kein frischer Klon, keine
  Plattformabnahme für Windows oder Linux — nur die Übersetzung
  ([0007](entscheidungen/0007-kein-entfernter-lauf.md), §12).
- **Eine notarisierte Anwendung.** Es gibt keinen Developer-Account, und das
  ist eine Entscheidung ([0003](entscheidungen/0003-keine-notarisierung.md)).
- **Peer-zu-Peer und die KI-Formate.** Sie entstehen in den Phasen 5.3 und
  5.4; O11 nennt sie in seinem Umfang, und es gab sie am 2026-09-11 noch
  nicht. Ein Fuzzziel darüber prüfte eine leere Menge und sähe aus wie eines.

  **Seit dem 2026-09-14 gibt es die Peerseite** (N1, N2), und was von ihr auf
  keiner Ebene steht, ist genauer zu benennen:

  | Was | Warum es hier steht |
  | --- | --- |
  | **Zwei Rechner** | Der Lauf über einen echten Port geht über `127.0.0.1` (`internal/app/peertcp_intern_test.go`). Er fährt TCP, TLS 1.3, das Festnageln, die Rahmung und den Handschlag — aber beide Enden stehen auf **derselben** Maschine |
  | **Ein zweiter Rechner im selben Netz** | **Seit dem 2026-09-21 geht ein Lauf über die echte Schnittstelle**: `internal/entdeckung/zweiinstanzen_netzlauf_test.go` kündigt eine Anwendung an und sucht sie mit einer zweiten — über jede taugliche Schnittstelle, also über echtes Multicast. Gemessen auf diesem Rechner: gefunden über `en0` unter `192.168.178.74:47110`. **Beide Enden stehen weiterhin auf derselben Maschine**; was ein zweiter Rechner sieht, bleibt eine Abnahme von Hand. Er steht hinter `netzlauf`, weil er ein aktives Netz und auf macOS die Erlaubnis *Lokales Netzwerk* braucht |
  | **Ein zweiter Rechner über das Internet** | Der Weg über tailcat (N3) ist über eine **Naht** geprüft: was dort einen `net.Conn` liefert, ist ein `net.Pipe`. **Seit SEC-21 läuft zusätzlich ein Netzlauf über echtes DERP — auf einem Rechner** (`fernweg_netzlauf_test.go`, `docs/belege/sec21-fernweg.md`): zwei Knoten finden sich über das **DERP-Rendezvous** und sprechen den gesicherten Weg, gewöhnlich und mit erzwungenem Relais (`TS_DEBUG_ALWAYS_USE_DERP=1`). Belegt ist damit die Fassung der Einladung, der gesicherte Weg darüber, der Lebenszyklus und das Rendezvous selbst. **Nicht belegt bleibt, was zwei Rechner braucht**: ein Weg durch ein NAT, ein direkter Weg zwischen zwei Maschinen und der **selbsttätige** Rückfall aufs Relais, wenn der direkte scheitert. #1456 verlangt genau diese Ehrlichkeit: „Produktivfreigabe bleibt offen, bis eine reale Direkt- und Relaisprobe zwischen zwei Rechnern vorliegt" |
  | **Ein Drahtformatbruch von tailcat** | Die Einladung trägt eine Fassung und wird vor dem Wählen geprüft — das ist gemessen. **Ungemessen bleibt der Fall, für den es sie gibt**: ein tailcat, das sich bei gleicher Zahl anders verhält. Dafür bräuchte es zwei Fassungen der Bibliothek nebeneinander |
  | **WLAN mit Client-Isolation** | Dort kommt kein Multicast durch, und die Entdeckung fände niemanden. Dass sie dann einen benannten Fehler gibt statt eines Spinners, ist an der Frist festgemacht und nicht an einem solchen Netz gemessen |
  | **Die lokale Netzwerkberechtigung von macOS** | Sie wird an der **gebauten, aus dem Finder gestarteten** `.app` vergeben, und ein Terminallauf bekommt sie anders zugeteilt. Die Abnahme von #1455 sagt es ausdrücklich. **Ihr Fehlen ist am 2026-09-14 zum ersten Mal beobachtet worden** — die Sichtprüfung unter `wails dev` bekam `write udp4 0.0.0.0:52316->224.0.0.251:5353: sendto: no route to host`, und die Oberfläche meldete „Es ist ein unerwarteter Fehler aufgetreten". Seither gibt es dafür `domain.ErrOhneNetzerlaubnis` mit dem Weg in die Systemeinstellungen. **Seit dem 2026-10-01 halb gemessen** (#1871, #1996, Abnahme-VM unter Lume, macOS 26.4): die gebaute `.app` **bekommt** die Rückfrage, und seit #1996 sagt sie darin, wofür. **Und über die Grenze zweier Systeme gefunden**: ein Ankündiger in der VM wurde vom Wirt gefunden und umgekehrt (`192.168.64.2` ↔ `192.168.64.1`, über `bridge100`, null verworfen). **Was das nicht ist**: zwei Geräte im WLAN — die VM hängt am virtuellen Segment des Wirts |
  | **Firewall und isolierte Gäste-WLANs** | Sie lassen einen Port zu oder nicht, und ein Lauf auf der Rückschleife sieht davon nichts. **Am 2026-10-01 einmal über eine Systemgrenze**: der Wirt erreichte einen Horcher in der Abnahme-VM per TCP (`:41234`), noch während die Rückfrage offen stand — die Firewall des Gasts war aus |
  | **Schlafen, Aufwachen, Netzwechsel** | Eine TCP-Verbindung überlebt einen Adresswechsel nicht; die Wiederaufnahme ist fachlich zu bauen ([0113](entscheidungen/0113-kein-libp2p-im-lan.md)) |
  | **Eine wiederaufgenommene TLS-Sitzung** | Der Klient führt keinen Sitzungsspeicher, also kommt sie nicht vor. Dass `VerifyConnection` sie abdeckte, ist damit **abgeleitet** und nicht gemessen |
  | **Ein Fuzzziel über die Hülle** | Sie ist gebaut und hat eines noch nicht |
- **Eine `verbindlich`e Fähigkeit.** Die Matrix zählt heute null davon; sie
  kommen mit Phase 5.6.

---

## Was O11 hinzugefügt hat

**Neun Fuzzziele** — Einstellungsaustausch, Kompatibilitätsmanifest,
OAuth-Rückruf, Kartenliste, Sammlungs-CSV, der Dienst als Ganzes, das
vorgezeigte **Token**, die **Sicherung** und die portable **Referenz** — und
**zwei Funde**, beide behoben und als Testfall erhalten
([`RELEASE.md`](RELEASE.md)).

**Die letzten drei kamen im Nachgang** (Befund 3 des Zweitblicks): das Issue
nennt sie ausdrücklich, und die erste Fassung dieses Werkstücks hatte sie
übergangen — ohne es zu sagen. Zwei von ihnen sagen mehr zu als „abgewiesen“:
bei `FuzzTokenPruefen` muss **jede** Abweisung außer der am Scope gleich
aussehen, und bei `FuzzEntpacken` darf **keine** Datei außerhalb des genannten
Verzeichnisses landen.

**Und zwei von ihnen hielten ihren Kopf nicht** — Befunde 1 und 2 des
Nachlaufs zum Zweitblick über Phase 5.2, beide am 2026-09-11 behoben:

| Ziel | Was der Kopf versprach | Was der Rumpf tat | Wie es jetzt steht |
| --- | --- | --- | --- |
| `FuzzEntpacken` | keine Datei außerhalb des Verzeichnisses | Die Angriffssaaten brachten **kein Manifest** mit — und die eine, die eins hatte, nannte es `manifest.json` statt `sicherung.json`. `Entpacken` endete beim Manifestlesen, **bevor** je ein Dateiname benutzt wurde. Beanstandet wurde außerdem nur, was in `daneben` landete; ein `heim/ausbruch.txt` lief durch die Schleife und blieb unbemerkt | Jede Saat bringt ihr Manifest mit und **nennt den bösartigen Namen darin** — ausgepackt wird nur, was das Manifest führt. Der Ausbruch wird am **Eingabewert** geprüft (`ausbrecher`), nicht an der Wirkung: ein absoluter Name schriebe woanders hin, und dort sieht niemand nach. Unerwartete Einträge sind jetzt auch im Heim ein Fehler |
| `FuzzLoeseAuf` | „eine Kennung, die nicht trifft, und trotzdem ein Rückfall über einen unvollständigen Namen“ | Geprüft wurden ausschließlich Ergebnisart und Trefferzahl. Eine falsche eindeutige Auflösung über einen halben Namen wäre grün geblieben | `pruefeRueckfall` hält zwei Aussagen: ohne treffende Kennung und ohne **volle** Angabe muss „unbekannt“ herauskommen — und **trifft** eine Kennung, ist das Ergebnis genau ihre Menge. Dazu zwei Saaten, die es vorher nicht gab: eine nicht treffende Kennung neben einem halben Namen, und eine treffende neben einem Namen, der etwas **anderes** trifft |

**Die Gegenproben zu beiden sind gemessen**, nicht behauptet: bei
`FuzzEntpacken` werden vier Saaten rot, sobald `sichererWeg` nicht mehr abweist
— eine davon schreibt wirklich ein `ausbruch.txt` ins Heim. Bei `FuzzLoeseAuf`
wird je eine Saat rot, wenn der Rückfall auch halbe Angaben vergleicht und wenn
die Kennung hinter den Rückfall zurücktritt.

**Zwei Läufe über die Durchsetzung der Autorisierungsmatrix**
(`internal/flaechen/autorisierungsmatrix_test.go`), und sie sagen **nicht
dasselbe**:

| Lauf | Was er hält | Gegenprobe |
| --- | --- | --- |
| drei Sammlungswege und eine Preisablehnung | **403 mit `erlaubnis`**, der Bestand bleibt danach unverändert, und der Ruf scheitert an der Erlaubnis und nicht am Dienstschalter — wer nichts darf, soll nicht erfahren, wie dieser Rechner eingestellt ist | **wird rot**, wenn die Reichweite in `domain.Akteur.Darf` übersprungen wird |
| **alle 125 schreibenden Wege**, aus der Wegetafel des Dienstes gelesen | **kein 2xx**, und **vier Auskünfte lauten danach wie vorher** (Sets, Sammlung, Profile, Token) — über keinen dieser Wege geschieht etwas | **wird rot**, wenn **beide** Tokenprüfungen umgangen werden: `DELETE /token/{tokenID}` antwortet dann mit 200 |

**Die zweite Gegenprobe braucht beide Stellen**, und das ist eine Aussage über
die Anwendung: die Reichweite eines Tokens wird an **zwei** Stellen geprüft —
in `domain.Akteur.Darf` und im Scopevergleich von
`service.Tokenverwaltung.Pruefe` —, und eine allein genügt. Wer nur eine
entschärft, sieht nichts.

**Hier stand, der 125-Wege-Lauf bleibe grün.** Das war falsch und ist vom
Nachlauf zum Zweitblick über Phase 5.2 nachgemessen worden (Befund 5). Was
richtig bleibt: er belegt nicht, dass die **Erlaubnis** der Grund war — die
meisten dieser Wege scheitern schon an Rumpf oder Kennung. Das hält der Lauf
darüber, an vier benannten Wegen.

**Das war die Hälfte, die der statische Wächter offen ließ** — er sagt in
seinem eigenen Kopf, dass er nicht hält, „ob die Durchsetzung dort sitzt, wo
die Zeile es sagt".
