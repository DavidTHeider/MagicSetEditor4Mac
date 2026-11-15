# Abhängigkeiten — was diese Anwendung mitbringt, und wie es gehoben wird

**Stand: 2026-09-24** — ENG-04
(#1729); die Vorabfassungen und das
Dashboard nachgelesen am 2026-10-11 — J03
(#2267).

Das ist das **Inventar** aus ENG-04.AK-01: jede **direkte** Abhängigkeit mit
Fassung, Lizenz und der Stelle, die sie **verantwortet** — Go, npm, die
Werkzeuge der Pipeline, die Actions, die Bibliotheken des Systems und die
Modelle. Die **transitiven** stehen vollständig in den Stücklisten jeder
Freigabe (CycloneDX, `stueckliste-go.json` und `stueckliste-npm.json`); hier
steht ihre Bilanz.

**„Verantwortet" heißt: hier bricht es, wenn die Abhängigkeit sich ändert.**
Bei Go sind das genau die Pakete, die sie importieren, bei npm die Stellen,
unterhalb derer sie eingeführt wird. Wer eine Abhängigkeit hebt, liest dort
nach; wer eine Stelle umbaut, sieht hier, was an ihr hängt.

**Ein Wächter hält die Tabellen gegen die Quellen**
(`internal/werkzeuge/abhaengigkeiten_test.go`): jede Pinnung hat ihre Zeile
und jede Zeile ihre Pinnung, in derselben Fassung; die Spalte „Verantwortet"
wird an den Importen gemessen, und **eine unbenutzte Abhängigkeit wird rot**
(AK-04). Was er nicht hält, steht in seinem Kopf — vor allem die Lizenz einer
Go-Abhängigkeit, die nur der Modulzwischenspeicher kennt.

## Go — direkt

| Modul | Fassung | Lizenz | Verantwortet | Wofür |
| --- | --- | --- | --- | --- |
| `filippo.io/age` | v1.3.2 | BSD-3-Clause | `internal/sicherung` | Sicherungen mit Kennwort — scrypt und ChaCha20-Poly1305 im Strom, mit `age -d` auch ohne diese Anwendung zu öffnen (#1723); bringt `filippo.io/hpke` v0.4.0 (BSD-3-Clause) mit |
| `github.com/ProtonMail/go-crypto` | v1.3.0 | BSD-3-Clause | `internal/aktualisierung`, `internal/app` | die OpenPGP-Signatur einer Freigabe prüfen; die Fassung verlangt Wails 3 |
| `github.com/dlclark/regexp2` | v1.12.0 | MIT | `internal/msescript` | reguläre Ausdrücke in MSE-Script, deren Gestalt Gos `regexp` nicht kennt |
| `github.com/ebitengine/purego` | v0.11.1 | Apache-2.0 | `internal/fenster`, `internal/prozessspeicher` | **Direkt seit dem 2026-10-07, nur im Abnahmebau** (`abnahme_hand_darwin.go`): `class_replaceMethod` aus der Objective-C-Laufzeit, damit die Hüllenprobe den ersten Klick zählen lässt, ohne die Klasse der Webansicht zu tauschen — der Tausch brach die Beobachtung von AppKit, und eine Größenänderung beendete die Anwendung. **Keine neue Abhängigkeit**: dieselbe, die `tmc/apple` mitbringt; vorher stand sie als indirekte da. **Seit dem 2026-10-15 auch im ausgelieferten Bau**: `internal/prozessspeicher` misst den Speicher der WebKit-Prozesse (OPS02, #2348). **Seit dem 2026-10-11 die Endfassung v0.11.1** statt alpha.6 (#2305) — tmc/apple verlangt weiter alpha.6, geladen wird nach MVS die höhere |
| `github.com/hashicorp/mdns` | v1.0.7 | MIT | `internal/entdeckung` | wer im selben Netz ist ([0113](entscheidungen/0113-kein-libp2p-im-lan.md)) |
| `github.com/ncruces/go-sqlite3` | v0.35.4 | MIT | `internal/storage` | der Bestand — SQLite als WebAssembly, ohne cgo |
| `github.com/tailscale/tailcat` | v0.7.0 | BSD-3-Clause | `internal/fernleitung` | die Verbindung über fremde Netze ([0114](entscheidungen/0114-tailcat-gepinnt.md)); **gehoben am 2026-10-07** (K01, #2187), weil erst v0.7.0 die Gegenstellen am Server nennt |
| `github.com/tmc/apple` | v0.6.18 | MIT | `internal/biometrie`, `internal/fenster` | Touch ID und Face ID über `LocalAuthentication`, cgo-frei; in `fenster` im **Abnahmebau** — die Hüllenprobe liest dort Leiste, Dialoge und Fenster über die Objective-C-Laufzeit (#1727) — **und seit NET-04 für den Teilen-Dialog** (`teilen_darwin.go`, `NSSharingServicePicker`, #1708); keine neue Abhängigkeit, dieselbe Fassung |
| `github.com/wailsapp/wails/v3` | v3.0.0-beta.30 | MIT | `internal/fenster` | die Hülle: Fenster, Menüleiste, Dialoge, Brücke ([0132](entscheidungen/0132-wails-3-als-huelle.md)) |
| `golang.org/x/crypto` | v0.57.0 | BSD-3-Clause | `internal/kennwort` | Argon2id für Kennwörter; **mitgehoben am 2026-10-07 mit tailcat v0.7.0** (K01); gemessen: eine mit v0.56.0 gehütete Form erkennt v0.57.0 wieder, ein falsches Kennwort weist es ab |
| `golang.org/x/image` | v0.45.0 | BSD-3-Clause | `internal/render`, `internal/schriften`, `internal/service` | Glyphen rastern ([0008](entscheidungen/0008-schriftrasterung.md)), BMP und TIFF schreiben |
| `golang.org/x/net` | v0.60.0 | BSD-3-Clause | `internal/netz`, `internal/entdeckung` | der Proxy aus der Umgebung (#1533); **seit K02** (#2188) auch `ipv4.PacketConn.SetMulticastInterface` für die Probe beim Ankündigen — keine neue Abhängigkeit |
| `golang.org/x/text` | v0.42.0 | BSD-3-Clause | `internal/domain` | die Ordnung einer Sammlung nach Sprache (#1500) |
| `tailscale.com` | v1.103.0-pre.0.20260916030321-a2263542f260 | BSD-3-Clause | `internal/fernleitung` | **Direkt seit NET-02** (#1706) und nur an einer Stelle: zwei Haken aus `tailscale.com/feature`, damit der Weg zum DERP-Relais den eingestellten Proxy nimmt. **Keine neue Abhängigkeit** — dieselbe Fassung, die `tailcat` mitbringt; vorher stand sie als indirekte da |

## Go — Werkzeuge der Pipeline

In `tools/go.mod` hinter `tool`-Direktiven; sie gehen **nicht** ins Programm.
Verantwortet heißt hier: diese Skripte rufen sie.

| Modul | Fassung | Lizenz | Verantwortet | Wofür |
| --- | --- | --- | --- | --- |
| `github.com/golangci/golangci-lint/v2` | v2.15.0 | GPL-3.0 | `scripts/ci-lint.sh`, `scripts/vorpruefung.sh` | die Stufe `lint` |
| `github.com/securego/gosec/v2` | v2.29.0 | Apache-2.0 | `scripts/ci-sicherheit.sh`, `scripts/vorpruefung.sh` | Sicherheitsprüfung im Quelltext |
| `golang.org/x/vuln` | v1.9.0 | BSD-3-Clause | `scripts/ci-sicherheit.sh`, `scripts/sicherheitslage.sh` | `govulncheck` — bekannte Schwachstellen im erreichbaren Code |
| `mvdan.cc/gofumpt` | v0.12.0 | BSD-3-Clause | `scripts/ci.sh` | die Stufe `format` |
| `github.com/CycloneDX/cyclonedx-gomod` | v1.12.0 | Apache-2.0 | `scripts/releasestueckliste.sh` | die Stückliste einer Freigabe |

## npm — im Programm

Vite baut aus ihnen das Bündel; nichts anderes geht aus `frontend/` ins
Programm.

| Paket | Fassung | Lizenz | Verantwortet | Wofür |
| --- | --- | --- | --- | --- |
| `@mantine/core` | 9.7.3 | MIT | `frontend/src/bausteine/`, `frontend/src/mantine.ts`, `frontend/src/sprache/kontext.tsx`, `frontend/src/stil.css` | die Bauteilbibliothek ([0125](entscheidungen/0125-mantine-als-bauteilbibliothek.md)) |
| `@mantine/hooks` | 9.7.3 | MIT | `@mantine/core` | verlangt von `@mantine/core`, in derselben Fassung (§1) |
| `@wailsio/runtime` | 3.0.0-beta.30 | MIT | `frontend/src/main.tsx` | die Brücke zu Go, in der Fassung der Hülle |
| `pixi.js` | 8.22.0 | MIT | `frontend/src/tisch/` | der gezeichnete Spieltisch (#1460) |
| `react` | 19.3.1 | MIT | `frontend/src/` | die Oberfläche |
| `react-dom` | 19.3.1 | MIT | `frontend/src/main.tsx`, `frontend/src/seiten/Karteneditor.tsx`, `frontend/src/bausteine/Arbeitsplatzschalter.tsx` | die Oberfläche im Dokument — die Schalterzeile stellt sich über ein Portal an einen fremden Ort (P03) |

## npm — nur beim Entwickeln

| Paket | Fassung | Lizenz | Verantwortet | Wofür |
| --- | --- | --- | --- | --- |
| `@eslint/js` | 9.39.5 | MIT | `frontend/eslint.config.js` | die Grundregeln von ESLint — mit ihm gleichauf (§1) |
| `@tailwindcss/vite` | 4.3.3 | MIT | `frontend/vite.config.ts` | Tailwind im Bau |
| `@testing-library/dom` | 10.4.2 | MIT | `@testing-library/react` | verlangt als Peer |
| `@testing-library/jest-dom` | 7.0.1 | MIT | `frontend/src/test/einrichtung.ts` | Zusicherungen über Knoten |
| `@testing-library/react` | 16.3.3 | MIT | `frontend/src/` | Läufe über Bauteile |
| `@testing-library/user-event` | 14.6.8 | MIT | `frontend/src/` | Eingaben wie ein Mensch |
| `@types/node` | 24.19.2 | MIT | `frontend/tsconfig.json` | Typen für Konfiguration und Wächter — der Laufzeit Node 24 folgend (§1) |
| `@types/react` | 19.3.0 | MIT | `frontend/tsconfig.json` | Typen für JSX |
| `@types/react-dom` | 19.3.0 | MIT | `frontend/src/main.tsx` | Typen für `react-dom/client` |
| `@vitejs/plugin-react` | 6.1.2 | MIT | `frontend/vite.config.ts` | React im Bau |
| `@vitest/coverage-v8` | 5.0.5 | MIT | `frontend/vite.config.ts` | die Abdeckung, in der Fassung von Vitest (§1) |
| `axe-core` | 4.14.0 | MPL-2.0 | `frontend/src/test/axe.ts` | die Prüfung auf Barrierefreiheit (§3) |
| `eslint` | 9.39.5 | MIT | `frontend/package.json` | der Linter der Oberfläche — eine Hauptfassung zurück (§1) |
| `eslint-plugin-jsx-a11y` | 6.10.2 | MIT | `frontend/eslint.config.js` | Barrierefreiheit im JSX |
| `eslint-plugin-react-hooks` | 7.1.1 | MIT | `frontend/eslint.config.js` | die Regeln der Haken |
| `globals` | 17.13.0 | MIT | `frontend/eslint.config.js` | die Namen der Laufzeitumgebungen für ESLint |
| `jsdom` | 30.1.3 | MIT | `frontend/vite.config.ts` | das Dokument der Läufe |
| `prettier` | 3.9.10 | MIT | `frontend/package.json` | Formatierung |
| `tailwindcss` | 4.3.3 | MIT | `frontend/src/stil.css` | die Stilklassen |
| `typescript` | 6.0.3 | Apache-2.0 | `frontend/src/` | Typprüfung, dazu der Zerteiler der Wächter — eine Hauptfassung zurück (§1) |
| `typescript-eslint` | 8.71.2 | MIT | `frontend/eslint.config.js` | ESLint für TypeScript |
| `vite` | 8.3.4 | MIT | `frontend/vite.config.ts` | Bau und Entwicklungsserver |
| `vitest` | 5.0.5 | MIT | `frontend/src/`, `frontend/vite.config.ts` | der Testläufer |

**Neun davon führt kein Quelltext ein**, und das ist kein Befund: `eslint`
und `prettier` sind Kommandos in `package.json`, `jsdom` und
`@vitest/coverage-v8` stehen als Name in der Konfiguration, die drei `@types`
findet TypeScript selbst, und zwei sind Peers eines anderen Pakets. Der
Wächter führt jeden dieser Gründe einzeln und prüft ihn nach.

## Actions

Nur in `.github/workflows/ci.yml`, und die Datei wird nicht ausgelöst
([0007](entscheidungen/0007-kein-entfernter-lauf.md)). Jede hängt an einem
Commit, die Fassung steht als Kommentar dahinter (§1).

| Action | Fassung | Lizenz | Verantwortet | Wofür |
| --- | --- | --- | --- | --- |
| `actions/checkout` | v7.0.1 | MIT | `.github/workflows/ci.yml` | den Stand holen |
| `actions/setup-go` | v7.0.0 | MIT | `.github/workflows/ci.yml` | Go in der gepinnten Fassung |
| `actions/setup-node` | v7.1.0 | MIT | `.github/workflows/ci.yml` | Node in der gepinnten Fassung |

## Transitiv — die Bilanz

**Gezählt am 2026-09-24, beim Entwickeln nachgezählt am 2026-10-03** (ohne `fast-glob`, unten). Vollständig, mit Pfad und Lizenz, steht es in den
Stücklisten jeder Freigabe.

| Was | Anzahl | Lizenzen | Gezählt über |
| --- | --- | --- | --- |
| **npm im Programm** | 38 Pakete | MIT 33, BSD-3-Clause 2, ISC 1, 0BSD 1, (MIT OR CC0-1.0) 1 | Einträge der Sperrdatei ohne `dev` |
| npm beim Entwickeln | 433 Pakete | MIT 346, MPL-2.0 25, Apache-2.0 21, ISC 15, BSD-2-Clause 9, BSD-3-Clause 6, BlueOak-1.0.0 4, CC0-1.0 2, MIT-0 2, 0BSD 1, CC-BY-4.0 1, Python-2.0 1 | Einträge der Sperrdatei mit `dev` |
| **Go im Programm** | 51 Module | BSD-3-Clause 24, MIT 17, Apache-2.0 7, BSD-2-Clause 2, ISC 1 | `go list -deps -tags production` über `.`, `./internal/...` und `./cmd/...` für darwin/arm64; die Lizenz aus der Lizenzdatei im Modulzwischenspeicher |

**Was das für die Lizenz dieses Programms heißt** — es steht unter
GPL-2.0-or-later: alles, was ins Programm geht, ist MIT, BSD, ISC, 0BSD oder
Apache-2.0. **Apache-2.0 verträgt sich mit der GPL erst ab Fassung 3**; das
„or later" trägt es. **Die 25 MPL-2.0-Pakete liegen alle beim Entwickeln** —
`axe-core` und `lightningcss`, das Tailwind im Bau benutzt, samt seinen
Plattformbauten —, ebenso CC-BY-4.0 (die Browserdaten von `caniuse-lite`) und
Python-2.0 (`argparse`).

## Bibliotheken des Systems

Nichts davon wird mitgeliefert; alles gehört zu macOS.

| Was | Wer es braucht | Wie |
| --- | --- | --- |
| AppKit, Cocoa, Foundation, QuartzCore, Carbon, ServiceManagement, UniformTypeIdentifiers, WebKit | Wails 3 | cgo, gebunden beim Bau |
| LocalAuthentication | `internal/biometrie` | `purego`, zur Laufzeit geladen — cgo-frei |
| libproc und `responsibility_get_pid_responsible_for_pid` aus libSystem | `internal/prozessspeicher` | `purego`, zur Laufzeit gebunden — cgo-frei; die Zuordnung ist **nicht öffentlich**, fehlt sie, heißt der Webspeicher „nicht gemessen" (OPS02, #2348) |
| der Schlüsselbund | `internal/schluesselbund` | `/usr/bin/security` als Kindprozess — cgo-frei |
| **Apples Gerätemodell** (`/usr/bin/fm`, macOS 27) | `internal/applefm` | `fm serve --socket` als Kindprozess, gesprochen wird über einen Unix-Socket — cgo-frei, kein Swift ([0166](entscheidungen/0166-was-fm-serve-belegt.md), AI-05 #2042). **Die Bedingungen**: `sudo fm license` bindet an die Softwarelizenz von macOS; der Auftraggeber hat sie am 2026-10-03 gelesen — nichtkommerzielle Nutzung und die Einbindung der Schnittstelle sind erlaubt, Fassung *FM1 version 1.0*. **Neu zu entscheiden**, wenn das Programm verkauft werden soll oder `fm license --status` eine andere Fassung meldet. Ohne das Werkzeug fehlt die Fähigkeit, sonst nichts |
| SQLite | `internal/storage` | **kein** Systembaustein: WebAssembly in `ncruces/go-sqlite3` |
| AppKit, WebKit (auch `WebView`, seit macOS 10.14 als veraltet markiert) und `swiftc` aus Xcode | die **WebKit-Probe** (`scripts/webkitprobe/`, #1828) — ein Prüfwerkzeug, **nicht** Teil der Anwendung | übersetzt von `scripts/webkitprobe.sh` nach `.go/bin/`; kein Paket, keine Fassung zu pinnen — Swift und WebKit sind die des Rechners, und der Bericht nennt die WebKit-Fassung |

## Modelle

**Keines wird mitgeliefert, keines geladen.** Ein Profil nennt die Modelle,
die es bei Ollama, Anthropic, OpenAI oder xAI benutzt — und Apples
Gerätemodell `system` gehört zum Betriebssystem; ihre Lizenzen sind die
ihrer Anbieter, und `/api/pull` benutzt diese Anwendung ausdrücklich nicht
([0098](entscheidungen/0098-was-ollama-belegt-und-was-nicht.md)). Welche
kleinen Modelle geprüft sind, misst AI-02
(#1743).

## Vorabfassungen und ihre Ausstiegsfragen

**Nachgelesen am 2026-10-11** — J03
(#2267).
Elf Abhängigkeiten im Programm tragen eine Vorabkennung oder die
Hauptfassung 0, sagen also **nichts über ihre Schnittstelle zu**. Je eine
Zeile nennt, **wann** gewechselt und wann zurückgegangen wird — damit die
Entscheidung vorher fällt und nicht in dem Moment, in dem etwas bricht.

**Woher jede Spalte kommt.** „Neueste" aus den Releases auf GitHub und dem
Modulproxy (`proxy.golang.org/<modul>/@latest`), der Abstand aus dem
Vergleich der Tags (`compare/<alt>...<neu>`). **Der Anteil** aus
`go tool nm -size` am nicht gestrippten Abnahmebau von `dfaad91b9`
(2026-10-11), summiert über die Symbole mit dem Modulpfad als Präfix —
Daten und Typinformation zählt das nicht vollständig, und transitive Module
nur, wo es dasteht. **Die Drift** aus der Stufe `security` (`govulncheck`
mit `-show verbose`, `npm audit`) und den Sicherheitsmeldungen der
Repositorien. **Ein Wächter hält die Tafel vollständig**
(`internal/werkzeuge/vorabfassungen_test.go`): jede Vorabfassung im Programm
hat ihre Zeile in der gepinnten Fassung und einen Auslöser. Ob Pflege, Drift
und Auslöser **stimmen**, hält er nicht — das ist dieser Tag.

| Abhängigkeit | Fassung | Neueste, nachgelesen | Lizenz · Anteil | Pflege | Sicherheitsdrift | Wechsel und Rückweg |
| --- | --- | --- | --- | --- | --- | --- |
| `github.com/wailsapp/wails/v3` | `v3.0.0-beta.30` | beta.33 vom 2026-10-19; beta.31 bis beta.33 erschienen am 18. und 19. Oktober und sind nicht durchgesehen | MIT · 1,0 MB | nächtliche Vorabfassungen, fünf zwischen dem 2026-09-20 und dem 2026-10-05; 36.518 Sterne, letzter Push 2026-10-11 | keine Meldung im Repositorium, `govulncheck` ohne Befund | **Heben**, wenn eine Fassung am darwin-Code einen Fehler behebt, den diese Anwendung trifft, oder eine Sicherheitsmeldung trägt — beta.30 tat das erste (#2398, `InvokeSync`). **Ausstieg** aus der Vorabfassung ist 3.0.0. **Zurück** über die vorige Pinnung beider Zahlen, wenn Hüllenprobe oder Sichtprüfung rot werden ([0124](entscheidungen/0124-wails-3-als-gepinnte-vorabfassung.md)) |
| `@wailsio/runtime` | `3.0.0-beta.30` | 3.0.0-beta.33 — mit Wails | MIT · im Oberflächenbündel | mit Wails | `npm audit` ohne Befund | **Nur mit Wails**, nie allein: ein Wächter hält beide Zahlen gleich, Renovate führt sie in einer Gruppe |
| `github.com/tailscale/tailcat` | `v0.7.0` | v0.7.0 ist die neueste (2026-09-20). Auf `main` seither 17 Commits, darunter **eine brechende**: `AllowedClients` wird zum Haken `AllowClient` — diese Anwendung setzt keines von beiden (`tailcatnaht.go`) | BSD-3-Clause · 0,1 MB, mit `tailscale.com`, gVisor und WireGuard 4,6 MB | Tailscale; sagt **nichts** zu — Schnittstelle, Kommandozeile und Drahtformat „may all change" ([0114](entscheidungen/0114-tailcat-gepinnt.md)) | GHSA-w496-344h-g2vv (niedrig) betraf den Modus `recv` der Kommandozeile bis 0.4.0, nicht die Bibliothek und nicht diese Fassung | **Heben** nur, wenn eine Fassung etwas behebt oder bringt, das der Fernweg braucht — und dann mit einem Lauf zwischen zwei Anwendungen, denn die Einladung trägt die Fassung und ein Bruch fällt **vor** dem Wählen auf. **Zurück** über die vorige Pinnung, wenn eine Einladung zwischen zwei Fassungen nicht mehr trägt. **Ausstieg**, wenn tailcat ruht: der Fernweg fällt weg, das eigene Netz bleibt (`internal/leitung` kennt tailcat nicht) |
| `tailscale.com` | `v1.103.0-pre.0.20260916030321-a2263542f260` | v1.104.1 vom 2026-10-07 | BSD-3-Clause · 2,6 MB | Tailscale; die Pseudofassung ist **die, die tailcat v0.7.0 verlangt** | `govulncheck` ohne Befund im erreichbaren Code | **Folgt tailcat und wird nie allein gehoben** — Renovates Vorschlag v1.104.1 ist abgelehnt (unten): eine Datenebene, gegen die tailcat nicht gebaut ist, wäre ungeprüft. Allein nur bei einer Meldung von `govulncheck` im **erreichbaren** Code, als Sicherheitsaktualisierung |
| `github.com/ebitengine/purego` | `v0.11.1` | v0.11.1 vom 2026-09-21 ist die neueste — **gehoben am 2026-10-11** von alpha.6 (#2305). Dazwischen 19 Commits; nachgelesen: `objc.Send` und der variadische Stapelweg unverändert, `shouldBundleStackArgs` zählt Felder jetzt ohne `structs.HostLayout`, 64-Bit-Ganzzahlen kodiert `objc` als `q`/`Q` statt `l`/`L` | Apache-2.0 · 0,1 MB | Ebitengine, letzter Push 2026-10-09 | keine Meldung | **Heben** bei einer Meldung oder einem Fehler an `objc`, Blöcken oder der arm64-Aufrufkonvention — gemessen mit dem Sensorlauf (`sensorlauf`), der Hüllenprobe und dem Teilen-Dialog. **Zurück** auf alpha.6, die Fassung, die `tmc/apple` verlangt, wenn einer davon rot wird. **Ausstieg** aus 0.x ist 1.0 |
| `github.com/tmc/apple` | `v0.6.18` | v0.6.18 ist das neueste Tag; letzter Commit 2026-08-22 | MIT · 5,2 MB | **eine Person**, 31 Sterne, ein offenes Issue — das schwächste Glied dieser Tafel | keine Meldung | **Der Ausstieg ist vorbereitet**: jeder Ruf steht hinter einem Riegel, der einen `panic` in eine Antwort verwandelt (§2). Gewechselt wird, wenn eine macOS-Fassung einen Ruf bricht oder das Projekt ein halbes Jahr ruht — Ersatz ist purego direkt über `objc`, wie `abnahme_hand_darwin.go` es schon tut |
| `github.com/ncruces/go-sqlite3` | `v0.35.4` | v0.35.6 vom 2026-09-23 — v0.35.5 bringt einen neuen Speicherzuteiler, v0.35.6 Leistung; keine Korrektur | MIT · 3,7 MB | ncruces, letzter Push 2026-10-01; jede Fassung mit Artefaktbescheinigung | keine Meldung | **Heben** im nächsten gebündelten Go-PR **nach** 0.8.0: ein neuer Zuteiler unter dem Bestand gehört gemessen (`leistungsschnitte`, E02), und vor einer Freigabe misst man nicht neu. Sofort bei einer Meldung zu SQLite oder zum Laufzeitkern. **Zurück** über v0.35.4 — die Datei bleibt dieselbe |
| `golang.org/x/crypto` | `v0.57.0` | v0.58.0 vom 2026-10-09 | BSD-3-Clause · 0,2 MB | das Go-Team | **GO-2026-5932** auf Modulebene, **ohne Korrektur**: `x/crypto/openpgp` ist unbetreut und unsicher gebaut. Diese Anwendung importiert das Paket nicht — `govulncheck` zählt null im erreichbaren Code, null in importierten Paketen, eine im Modul. OpenPGP läuft über `github.com/ProtonMail/go-crypto` | **Heben** bei einer Meldung im erreichbaren Code als eigener PR, sonst gebündelt. Wer `x/crypto/openpgp` importiert, macht die Meldung erreichbar und die Stufe rot — das ist gewollt |
| `golang.org/x/image` | `v0.45.0` | v0.47.0 vom 2026-10-09 | BSD-3-Clause · 0,2 MB | das Go-Team | `govulncheck` ohne Befund | **Heben** gebündelt oder bei einer Meldung; **mit Goldbildlauf**, denn `font/sfnt` und `vector` ändern die Glyphen ([0008](entscheidungen/0008-schriftrasterung.md)) |
| `golang.org/x/net` | `v0.60.0` | v0.61.0 vom 2026-10-09 | BSD-3-Clause · 0,2 MB | das Go-Team | `govulncheck` ohne Befund — v0.60.0 war die Sicherheitsaktualisierung vom 2026-10-09 (#2274) | **Heben** gebündelt oder bei einer Meldung; der Proxy und das Ankündigen im eigenen Netz hängen daran |
| `golang.org/x/text` | `v0.42.0` | v0.43.0 vom 2026-10-09 | BSD-3-Clause · 1,5 MB | das Go-Team | `govulncheck` ohne Befund | **Heben** gebündelt oder bei einer Meldung; die Ordnung einer Sammlung hängt daran, und `domain` hält sie mit Tests |

**Die vier aus `golang.org/x` stehen hier, weil die Klasse es verlangt**,
nicht weil sie ein Risiko derselben Art wären: die Hauptfassung 0 ist dort
die Regel, gepflegt vom Go-Team neben der Toolchain.

## Wie gehoben wird (AK-02)

**Renovate schlägt vor, entschieden wird hier**
([0129](entscheidungen/0129-renovate-schlaegt-vor.md), `renovate.json5`).

| Art | Wie | Wer entscheidet |
| --- | --- | --- |
| **Sicherheitsaktualisierung** | ein eigener PR je Fund, **nie gebündelt** und ohne Freigabe — so Renovates Vorgabe für `vulnerabilityAlerts` (`groupName: null`, nachgelesen über Context7 am 2026-09-24) | sofort: `make ci` auf seinem Stand, dann Merge |
| **Patch und Minor** | nach Freigabe auf dem Dashboard, **gebündelt**: ein PR für die Oberfläche, einer für die Go-Module, einer für die Werkzeuge, einer für die Actions | gesammelt, wenn eine Pipeline ohnehin läuft |
| **Gekoppelt oder bildwirksam** | eigene Gruppen: Mantine, Vitest, Wails 3 (je zwei Pakete in einer Zahl, §1) und PixiJS (ändert das Bild) | mit Sichtprüfung — ein Tisch, der anders zeichnet, sieht kein Test |
| **Hauptfassung** | nie gebündelt, nach Freigabe | erst mit einem Issue, das den Grund nennt |
| **Sperrdatei** | ihre Pflege (`lockFileMaintenance`) als eigener PR, nach Freigabe | wie Patch und Minor |
| **Zurückgehalten** | TypeScript unter 6.1, ESLint und `@eslint/js` unter 10, `@types/node` unter 25, Go selbst — Gründe in §1 | nie ohne Änderung von §1 |

**Gebündelt, weil jede Aktualisierung einen ganzen Lauf kostet.** Es gibt nur
den lokalen Lauf (§4): ein PR je Paket hieße zwanzig Minuten je Paket.

**Der Diff der Sperrdatei ist der Beleg.** Er steht in jedem PR, und gemergt
wird nie von selbst (`automerge: false`): ein Renovate-PR ist erst gut, wenn
`make ci` für genau seinen Stand grün war. **Dass die Regeln aus §1 in
`renovate.json5` wirksam bleiben, hält ein Wächter**
(`internal/werkzeuge/renovate_test.go`) — gegen die Peerbereiche aus der
Sperrdatei, nicht gegen eine Abschrift.

### Das Dashboard am 2026-10-11

**Triagiert gegen den heutigen Stand** — J03
(#2267),
Dashboard #1784
in der Fassung vom 2026-10-11 13:33 UTC. Freigegeben ist **nichts**, und
jede Zeile sagt warum.

| Vorschlag | Entscheidung | Grund |
| --- | --- | --- |
| `@wailsio/runtime` und Wails 3 auf beta.28 | **nicht jetzt** | am darwin-Code nichts geändert (Tafel oben); eine Hebung kostet Hüllenprobe und Sichtprüfung für null Wirkung |
| `tailscale.com` auf v1.104.1 | **abgelehnt** | folgt tailcat; v0.7.0 verlangt die Pseudofassung vom 2026-09-16 |
| `npm` und `vite` (Gruppe „Oberfläche ohne Hauptfassung") | **nach 0.8.0** | `vite` 8.3.3 → 8.3.4 bringt Korrekturen am Bau und am Entwicklungsserver, **keine Sicherheitskorrektur** — die drei Meldungen vom 2026-10-06 (GHSA-rq7h-c2jc-7f22, GHSA-vfpm-58rq-9qcg, GHSA-9jrq-w75r-8gcw) sind in 8.3.3 behoben, der installierten Fassung. `npm` 11.12.1 → 11.21.0 ändert §1 und die Stufe `format` mit |
| `npm` auf 12 | **abgelehnt bis zu einem Issue** | eine Hauptfassung ohne Grund (Regel oben) |
| Pflege der Sperrdatei | **nach 0.8.0** | sie hebt Hunderte transitive Fassungen auf einmal; vor einer Freigabe ist das die Änderung mit der größten Fläche und dem kleinsten Anlass |
| Mantine 9.7.2 | **nach 0.8.0, mit Sichtprüfung** | Korrekturen an Tooltip (`aria-describedby` bei abgeschaltetem Tooltip), Transition und PinInput; nichts davon ist hier gemeldet. Gekoppelt und bildwirksam — eine eigene Gruppe |
| `golang.org/x/vuln` v1.9.0 | **nicht jetzt** | erschienen am 2026-10-11, dem Tag der Triage; ein Prüfwerkzeug am Tag seines Erscheinens zu heben, hieße seine Fehler mitzuprüfen |
| `actions/setup-node` v7.1.0 | **gebündelt mit dem nächsten Lauf** | die Workflow-Datei wird nicht ausgelöst (§4); die Pinnung kostet nichts außer dem Lauf |

**Was Renovate nicht zeigt**: purego hat seit dem 2026-09-01 eine
Endfassung und `ncruces/go-sqlite3` zwei Patchfassungen — keines von beiden
steht im Dashboard. Die Tafel oben ist deshalb die Quelle für die Drift, und
das Dashboard ist ein Vorschlag (0129).

### Das Dashboard am 2026-10-16

**Nachgelesen für SEC04**
(#2357),
Dashboard #1784
in der Fassung vom 2026-10-16. **Je Änderung drei Quellen**: die
Releasenotiz des Tags (`gh api repos/<repo>/releases/tags/<tag>`), die
Sicherheitsmeldungen **des Repositoriums** (`/repos/<repo>/security-advisories`)
und die der **globalen Datenbank** (`/advisories?affects=<paket>@<fassung>`).
Beide braucht es — die Meldung zu Prettier stand am 2026-10-16 nur in der
ersten (unten).

**Der isolierte Test je Änderung ist NOT_RUN.** Jede Hebung lädt Pakete, und
ein Download braucht die ausdrückliche Zustimmung im Gespräch; ohne sie ist
eine Hebung nicht geprüft und wird **nicht** übernommen. Freigegeben ist
deshalb wieder nichts.

| Vorschlag | Quelle, nachgelesen | Entscheidung | Grund |
| --- | --- | --- | --- |
| `prettier` 3.9.9 → 3.9.10 | Notiz vom 2026-10-10: nur `micromark-extension-gfm-table`. **GHSA-4cm8-f57v-mqm6** (mittel, 2026-10-11): die Rechenzeit beim Formatieren einer GFM-Tabelle wächst quadratisch, betroffen 3.9.0 bis 3.9.9 | **Sicherheitsaktualisierung, eigener PR — wartet auf die Zustimmung zum Download** | nur beim Entwickeln, und dort **nicht erreichbar**: Prettier formatiert `frontend/`, und dort liegt keine Markdown-Datei (`find frontend -name '*.md' -not -path '*/node_modules/*'`, null Treffer am 2026-10-16) — eine fremde Tabelle kommt nie an. Die Regel oben verlangt den eigenen PR trotzdem |
| `typescript-eslint` 8.71.1 → 8.71.2 | Notiz vom 2026-10-12: vier Korrekturen an `no-unused-vars` und `no-unnecessary-type-parameters` | gebündelt, nach Zustimmung | ein Patch innerhalb der Sperre aus §1 — `typescript <6.1.0` bleibt |
| `@types/node` 24.19.1 → 24.19.2 | nur Typen | gebündelt, nach Zustimmung | innerhalb der Sperre `<25` |
| `vite` 8.3.3 → 8.3.4, `npm` 11.12.1 → 11.21.0 | wie am 2026-10-11 | gebündelt, nach Zustimmung | keine Sicherheitskorrektur; `npm` ändert §1 und die Stufe `format` mit |
| Vitest 5.0.3 → 5.0.4, mit `@vitest/coverage-v8` | Notiz vom 2026-10-12: experimentelle Neuerungen (`CACHEDIR.TAG`, `test.ui.theme`, Trace-Ansicht), keine Korrektur | gekoppelte Gruppe, nach Zustimmung | keine Meldung. Beim Sprung auf 5 verschob sich das Maß um fünf Anweisungen (§1) — nach der Hebung die Zählung vergleichen |
| Wails 3 und `@wailsio/runtime` beta.28 | wie am 2026-10-11 | **nicht jetzt** | am darwin-Code nichts geändert |
| Mantine 9.7.2 | **GHSA-54v8-pg45-q335** (mittel, 2026-10-13) betrifft `@mantine/form` vor 9.7.2 | **nicht jetzt**, mit Sichtprüfung | `@mantine/form` ist **nicht installiert** (Sperrdatei, 2026-10-16); die Meldung trifft diese Anwendung nicht |
| Pflege der Sperrdatei, darin `react`/`react-dom` 19.3.1, `jsdom` 30.1.3, `@testing-library/user-event` 14.6.8 | React 19.3.1 (2026-10-14): nur React Server Components — **GHSA-xqc2-xf4c-x3h3** (kritisch) betrifft `react-server-dom-parcel`, `-turbopack` und `-webpack`, keines installiert. jsdom 30.1.3 (2026-10-15): Leistung und Zugangsdaten von `XMLHttpRequest` über Umleitungen | eigener PR, nach Zustimmung | im Fenster läuft React ohne Server Components; jsdom nur in Tests |
| `golang.org/x/vuln` v1.9.0 | Modulproxy: erschienen 2026-10-11 03:39 UTC, Commit `e672cbe8f` | **allein, nach Zustimmung** | der Grund vom 2026-10-11 — am Tag des Erscheinens — trägt nicht mehr; ein Prüfwerkzeug hebt man allein, damit sein eigener Unterschied sichtbar bleibt |
| `actions/setup-node` v7.1.0 | — | gebündelt mit dem nächsten Lauf | die Workflow-Datei wird nicht ausgelöst (§4). Die drei gepinnten Hashes stimmen mit ihren Tags überein, nachgelesen über `git/ref/tags` am 2026-10-16 |
| `tailscale.com` v1.104.1 | — | **abgelehnt** | folgt tailcat |
| `npm` 12 | — | **abgelehnt bis zu einem Issue** | eine Hauptfassung ohne Grund |

**Was Renovate hier nicht sieht, und warum**: es führt Prettier in der Gruppe
„ohne Hauptfassung" und nicht als Sicherheitsaktualisierung. Seine
`vulnerabilityAlerts` lesen die Meldungen von Dependabot, und **die sind für
dieses Repositorium abgeschaltet** (`gh api …/dependabot/alerts`: HTTP 403,
„Dependabot alerts are disabled for this repository", 2026-10-16). Die Regel
„ein eigener PR je Fund" kann so nie greifen. Sie anzuschalten ist eine
Einstellung des Repositoriums und gehört dem Menschen; bis dahin ist die
Quellenprüfung oben der Weg, auf dem eine Meldung ankommt.

**Angeschaltet am 2026-10-19**, auf Zustimmung des Auftraggebers
(`gh api -X PUT …/vulnerability-alerts`, HTTP 204). **Die erste Meldung kam
sofort**: [GHSA-2v4p-qf9q-27wj](https://github.com/advisories/GHSA-2v4p-qf9q-27wj)
(hoch) — ein xDS-Server von gRPC-Go stürzt bei einer Anfrage ohne
`:authority` und `Host`, betroffen 1.83.0 bis 1.83.1. Sie steht in
**`tools/go.mod`**: `gosec` → `autofix` → `google.golang.org/genai` →
`gax-go` → `grpc` (`go mod why`, 2026-10-20). **Erreichbar ist sie nicht** —
`gosec` startet keinen gRPC-Server, und in der Anwendung kommt `grpc` nicht
vor. Gehoben auf 1.83.2 (Notiz vom 2026-08-25: genau diese Korrektur) in einem
eigenen PR, wie die Regel oben verlangt.

### Gehoben am 2026-10-20

**Auf Zustimmung des Auftraggebers vom 2026-10-19**, je in einem eigenen PR,
je mit Quelldurchsicht und `make ci`:

| Werkzeug | Fassung | Was sich ändert |
| --- | --- | --- |
| `grpc` (in `tools/go.mod`) | 1.83.1 → 1.83.2 | die Meldung oben (#2514) |
| `golangci-lint` | 2.14.0 → 2.15.0 | `errcheck` prüft jetzt den Empfang aus einem `chan error`; acht Stellen in Tests, drei davon zeigten einen erwarteten Fehler, den kein Test las (#2515) |
| `govulncheck` | 1.8.0 → 1.9.0 | **am Quelltext nichts**: die beiden Fassungen sind außer `go.mod` und `go.sum` bytegleich, verglichen im Modulcache am 2026-10-20. Gehoben sind nur ihre Abhängigkeiten, und die hatte golangci-lint 2.15.0 schon mitgebracht — bis auf `x/telemetry`, dessen Code ebenso gleich blieb |
| `actions/setup-node` | v7.0.0 → v7.1.0 | `mise.toml`, absolute Pfade in `node-version-file`, Prüfung der Node-Fassung; der Hash gegen den Tag gelesen. Die Workflow-Datei läuft nicht (§4) |
| `prettier` | 3.9.9 → 3.9.10 | **Sicherheitsaktualisierung** (GHSA-4cm8-f57v-mqm6, hier nicht erreichbar, Absatz unten); die Formatprüfung bleibt ohne Änderung |
| Vitest mit `@vitest/coverage-v8` | 5.0.3 → 5.0.5 | 10.351 Tests grün; **das Maß verschob sich** um fünf Anweisungen, vier Zweige, fünf Funktionen und fünf Zeilen (`ast-v8-to-istanbul` 1.0.7) |
| Oberfläche ohne Hauptfassung | user-event 14.6.8, `@types/node` 24.19.2, jsdom 30.1.3, React 19.3.1, typescript-eslint 8.71.2, Vite 8.3.4 | keine Meldung vor oder nach der Hebung; Vite ändert den Vorlade-Helfer im gebauten Bündel |
| Mantine | 9.7.1 → **9.7.3** | nicht auf die 9.7.2 des Dashboards: erst 9.7.3 dichtet die Fokusfalle unter macOS ohne Tastaturnavigation ab, an der `Modal` und `Drawer` hängen |
| Wails mit `@wailsio/runtime` | beta.27 → beta.30 | `InvokeSync` gibt den Wartenden nach einer Panik frei (#2398); die Umgehung, die die Anwendung dann beendete, ist fort |

**Nicht gehoben, mit Grund:**

| Vorschlag | Grund |
| --- | --- |
| `npm` 11.12.1 → 11.21.0 | npm ist das Werkzeug dieses Rechners; es zu heben hieße, es global zu installieren (§1). `packageManager` und die Stufe `format` halten 11.12.1 |
| `npm` 12 | eine Hauptfassung ohne Grund — abgelehnt bis zu einem Issue |
| `tailscale.com` v1.104.1 | folgt tailcat; v0.7.0 verlangt die Pseudofassung vom 2026-09-16 |
| `typescript-eslint` 8.72.0 | eine Nebenfassung, erschienen nach dem durchgesehenen Vorschlag |
| Wails beta.31 bis beta.33 | erschienen am 18. und 19. Oktober, nicht durchgesehen |

## Sicherheitsbefunde (AK-03)

**Vier Zustände, und jeder heißt anders** (`scripts/sicherheitslage.sh`):

| Zustand | Heißt |
| --- | --- |
| `ohne-befund` | der Scanner hat geantwortet, nichts liegt über der Schwelle |
| `befund` | eine Schwachstelle — `govulncheck`: im erreichbaren Code; `npm audit`: ab `high` |
| `nicht-erreichbar` | die Datenbank oder die Registry hat nicht geantwortet — **nichts geprüft** |
| `gescheitert` | der Lauf ist aus einem anderen Grund gescheitert |

Die Stufe `security` zeigt sie, und **der Freigabebeleg nennt sie** samt den
Kennungen oder den Zahlen je Schwere — ein Scanner, der nicht antwortete,
steht dort als nicht erreichbar und nicht als grün. **Ein Befund und ein
gescheiterter Scanner halten die Freigabe an**, ein nicht erreichbarer auch —
es sei denn, jemand entscheidet es ausdrücklich (`MSE_FREIGABE_OHNE_SCANNER=1`),
und dann steht die Entscheidung im Beleg (`freigabeurteil`, Zweitblick über
Phase 6.2). Die Stücklisten gehen mit jeder Freigabe hinaus und stehen in den
signierten Prüfsummen.

**Am 2026-10-03 hielt `npm audit` die Pipeline an**: GHSA-vfj7-8cjw-p6xm
(`braces` bis 3.0.3, ein Stapelüberlauf über tief geschachtelte Muster) —
**ohne Korrektur**, herein über `fast-glob` → `micromatch` → `braces`, nur
beim Entwickeln. Statt den Befund zu dulden, ist die Abhängigkeit gegangen:
die zehn Wächter, die damit Quelldateien aufzählten, nehmen Nodes
`fs.globSync` (`frontend/src/test/dateien.ts`) — gemessen am selben Tag
fanden beide dieselben 1.562 Dateien. 16 Pakete weniger.

**Am 2026-10-06 meldete `npm audit` GHSA-68fv-2mgg-jv7q** (`source-map-js`
1.0.0 bis 1.2.1, behoben in 1.2.2 vom 2026-09-30) — QA-DEP01
(#2176).
Herein über `css-tree` (jsdom), `postcss`, `@tailwindcss/node` und
`magicast`, alle mit `^1.2.1` und alle **nur beim Entwickeln** (`dev: true`);
`frontend/src` importiert es nicht, und ins Bündel geht es nicht. Gehoben
wurde der eine Knoten der Sperrdatei mit `npm update source-map-js` —
derselbe Bereich der Eltern, kein Override, keine andere Zeile. Danach meldet
`npm audit` null Befunde. **Was das nicht belegt**: ob eine präparierte
Source-Map den Bau je erreicht hätte — untersucht ist die Fassung, nicht
jeder Eingang des Werkzeugwegs.

**Am 2026-10-16 wusste `npm audit` von einer Meldung nichts** — SEC04
(#2357).
GHSA-4cm8-f57v-mqm6 betrifft `prettier` 3.9.0 bis 3.9.9 (installiert ist
3.9.9) und stand an diesem Tag **nur** unter den Meldungen des Repositoriums
`prettier/prettier`: die globale Datenbank antwortete auf
`affects=prettier@3.9.9` mit null Einträgen, und die Stufe `security`
desselben Tages zeigte „mittel 0". Dass die Abfrage greift, belegt die
Gegenprobe `affects=source-map-js@1.2.1`, die GHSA-68fv-2mgg-jv7q nennt.
**Bewertet**: nur beim Entwickeln und nicht erreichbar (Tafel oben).
**Gehoben auf 3.9.10 am 2026-10-20**, auf Zustimmung des Auftraggebers vom
2026-10-19, in einem eigenen PR: die Notiz nennt nur
`micromark-extension-gfm-table`, die Formatprüfung über `frontend/src` bleibt
ohne Änderung.
**Was daraus folgt**: eine grüne Stufe `security` heißt, dass die globale
Datenbank nichts kennt — nicht, dass niemand etwas gemeldet hat. Vor einer
Freigabe gehören die Meldungen der Repositorien der direkten Abhängigkeiten
dazu (`/repos/<repo>/security-advisories`); der Freigabebeleg nennt, ob sie
gelesen wurden.

**`govulncheck -show verbose` am 2026-10-16**: null Funde im erreichbaren
Code und in importierten Paketen, einer auf Modulebene — GO-2026-5932 in
`golang.org/x/crypto` v0.57.0, ohne Korrektur, im nicht importierten
`x/crypto/openpgp` (Tafel der Vorabfassungen oben).

## Eine neue Abhängigkeit (AK-04)

**Wer eine Abhängigkeit aufnimmt, schreibt drei Dinge auf, bevor sie
kommt** — in den PR und in [`DRITTANBIETER.md`](../DRITTANBIETER.md):

1. **Größe, gemessen**: bei npm die Zahl der neuen Pakete und die Megabyte in
   `node_modules` (so bei Mantine: 16 Pakete, 26 MB), bei Go die neuen Module
   und der Zuwachs des Programms.
2. **Lizenz, nachgelesen** — in der Lizenzdatei, nicht im Paketnamen.
3. **Pflege**: wer sie hält, wann die letzte Fassung kam, was sie **nicht**
   zusagt (tailcat: „may all change", [0114](entscheidungen/0114-tailcat-gepinnt.md)).

Dann bekommt sie ihre Zeile hier, mit der Stelle, die sie verantwortet.
**Wer eine entfernt, streicht ihre Zeile** — und wer eine Stelle umbaut, bis
nichts mehr eine Abhängigkeit einführt, sieht den Wächter rot, bis sie aus
`go.mod` oder `package.json` verschwunden ist.
