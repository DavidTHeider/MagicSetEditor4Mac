# Beleg: Lieferkette und Erreichbarkeit — SEC04

**Gemessen am 2026-10-16** auf dem Zweig `feat/2357-sec04-lieferkette`,
gebaut auf `de77b1815`; Abhängigkeitsdateien (`go.mod`, `go.sum`,
`tools/go.mod`, `frontend/package.json`, `frontend/package-lock.json`,
`.github/workflows/`) unverändert seit `9986133e7`. Mac17,3, macOS 27.0.1,
Go 1.27.2, cyclonedx-gomod v1.12.0, govulncheck v1.8.0, npm 11.12.1.
**Eingaben**: der Quelltext dieses Stands; keine Bestände, keine Konten,
kein Netz außer den Abfragen unten. Issue
#2357,
Bericht `SEC04-2357.md` (Entwicklungsarchiv).

## Stücklisten

| Lauf | Ergebnis |
| --- | --- |
| `cyclonedx-gomod mod -licenses -output x.json` | **XML** (`<?xml version="1.0"…<bom xmlns="http://cyclonedx.org/schema/bom/1.6"`), 87.828 Byte — so groß wie `stueckliste-go.json` der Freigabe 0.8.0 (`gh release view v0.8.0`, Größe des Assets) |
| `cyclonedx-gomod mod -licenses -json -output x.json` | CycloneDX 1.6 in JSON, 68 Komponenten, jede mit Lizenzangabe |
| gegen `go list -deps -tags production` über `.`, `./internal/...`, `./cmd/...` für darwin/arm64 | **53 Module im Programm, alle in der Stückliste**; 15 Komponenten mehr, alle nur für andere Plattformen oder nur Werkzeug (`wintun`, `wireguard/windows`, `go-winio`, `wingoes`, `go-ole`, `netlink`, `rtnetlink`, `ethtool`, `godbus`, `memconn`, `go-colorable`, `go-isatty`, `go-cmp`, `x/mod`, `x/tools`) |
| Lizenzen der 53, aus der Stückliste | BSD-3-Clause 27, MIT 16, Apache-2.0 7, 0BSD 1, MIT-0 1, BSD-2-Clause 1 — die Bilanz in `ABHAENGIGKEITEN.md` (51 Module, gezählt am 2026-09-24 über die Lizenzdateien) weicht in der Zahl um die zwei seither aufgenommenen Module und in der Einordnung um den Erkennungsweg ab |
| `npm sbom --sbom-format cyclonedx` gegen `package-lock.json` | CycloneDX 1.5 in JSON, **405 Pakete, jedes in der Sperrdatei**; 50 nur in der Sperrdatei, alle `optional` und `dev`, 45 davon mit `os`/`cpu` einer fremden Plattform |
| npm im Programm (Sperrdatei ohne `dev`) | 38 — MIT 33, BSD-3-Clause 2, ISC 1, 0BSD 1, (MIT OR CC0-1.0) 1; wie in `ABHAENGIGKEITEN.md` |

## Actions und Kopplung

| Prüfung | Ergebnis |
| --- | --- |
| `actions/checkout` v7.0.1, `actions/setup-go` v7.0.0, `actions/setup-node` v7.0.0 gegen `gh api repos/<repo>/git/ref/tags/<tag>` | alle drei Hashes **gleich** dem Commit ihres Tags |
| Wails 3 in `go.mod` und `@wailsio/runtime` in der Sperrdatei | beide beta.27; der Wächter `frontendpinnung_test.go` hält es |

## Sicherheitsbefunde

| Quelle | Ergebnis |
| --- | --- |
| `govulncheck -show verbose ./...` | Symbole: keine. Pakete: keine. Module: **GO-2026-5932** in `golang.org/x/crypto` v0.57.0, ohne Korrektur — „Your code is affected by 0 vulnerabilities" |
| `npm audit` (Stufe `security` von `make ci` am selben Tag, Stand `df07f662e`) | kritisch 0, hoch 0, mittel 0, niedrig 0 |
| globale Datenbank, `gh api /advisories?ecosystem=npm&affects=<paket>@<fassung>` für 14 direkte Pakete in ihrer installierten Fassung | **null** je Paket. Gegenprobe `affects=source-map-js@1.2.1`: GHSA-68fv-2mgg-jv7q — die Abfrage greift |
| Meldungen der Repositorien seit 2026-09-01, `gh api /repos/<repo>/security-advisories` über vite, vitest, wails, mantine, pixijs, prettier, typescript-eslint, jsdom, react, tailwindcss | **GHSA-4cm8-f57v-mqm6** (prettier 3.9.0–3.9.9, mittel); GHSA-54v8-pg45-q335 (`@mantine/form` < 9.7.2, nicht installiert); GHSA-xqc2-xf4c-x3h3 (`react-server-dom-*`, nicht installiert); die drei zu vite vom 2026-10-06, in 8.3.3 behoben |
| Erreichbarkeit der Prettier-Meldung | `find frontend -name '*.md' -not -path '*/node_modules/*'` — **null**; Prettier formatiert nur `frontend/` |
| Dependabot | `gh api repos/…/dependabot/alerts` — HTTP 403, abgeschaltet |

## Erreichbarkeit

| Wächter | Gefahrene Fälle | Rot gesehen | Gegenprobe am Schaden |
| --- | --- | --- | --- |
| `hakenohneaufrufer.test.ts` | jede Datei unter `src/haken/` | ja — zwei Vorrichtungen am alten Ort | — |
| `nachrichtenarten_test.go` | elf Arten aus `domain/handelsumschlag.go` | ja — `hallo`, `sichtanfrage` | ohne den einzigen Sender von `uebergabe`: rot |
| `exportverbraucher_test.go` | jede exportierte Funktion ohne Empfänger in fünf Paketen, über 2.700 Quelldateien ohne Tests | ja — 92 bei leerer Liste | ohne den einzigen Ruf von `service.GriffZiehen`: rot. Ohne den Ruf von `domain.Sitzungsreferenz` in `freigabepflege.go`: **grün** — `ZeileZuReferenz` ruft sie, und die nehmen nur Tests (Grenze im Kopf) |
| `bindungsaufrufer_test.go`, `TestEineVorrichtungRuftNicht` | acht Pfade | ja — `src/test/` zählte als Rufer | `TestJedeBindungWirdGerufen` bleibt grün: keine Bindung hing nur an einer Vorrichtung |

## Grenzen

- **Isolierter Test der Renovate-Änderungen: NOT_RUN** — jede lädt Pakete, und eine Zustimmung zum Download liegt nicht vor.
- **Die Stückliste von 0.8.0 ist nicht geöffnet** — der Schluss auf XML stützt sich auf Bauweg und Größe.
- **Die globale Datenbank und die Repositorien** sind am 2026-10-16 abgefragt; eine Meldung danach steht hier nicht.
- **Methoden und transitiv toter Code** liegen außerhalb des Exportwächters.
