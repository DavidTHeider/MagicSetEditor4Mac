# 0132 — Wails 3 als Hülle, gepinnt und mit Rückweg

**Stand:** entschieden am 2026-09-24
(#1727, ENG-02) —
die Bestandsmatrix steht in [`WAILS3.md`](../WAILS3.md).

## Was entschieden ist

**Die Anwendung läuft auf Wails `v3.0.0-beta.23`** statt auf Wails 2.13.0 —
die Fassung, die [0124](0124-wails-3-als-gepinnte-vorabfassung.md) als
Planungsziel festhält —, und das Frontend spricht mit ihr über
`@wailsio/runtime` in **derselben** Fassung. Beides ist auf die Stelle
gepinnt (§1); ein Wächter hält die zwei Zahlen gleich.

**Welche Fassung, entscheidet 0124; wie die Hülle gebaut ist, dieser
Vermerk.** Die Entwicklung lief zuerst gegen beta.25 und ist auf beta.23
zurückgegangen, als der Grund dafür nicht trug: in allem, was diese Anwendung
benutzt — `pkg/application`, der Assetserver, der Quelltext von
`@wailsio/runtime` — unterscheiden sich die zwei Fassungen **nicht**, gemessen
am 2026-09-24 mit `diff -rq` über den Modulquelltext. Verschieden sind nur
`pkg/updater`, den diese Anwendung nicht benutzt, und die Fassungsnummer.
Ohne Unterschied gilt, was ENG-02.AK-01 sagt: keine Änderung des Zielstands.

## Was der Vorabstatus kostet

**Wails 3 ist eine Beta** (0124). Deshalb:

- **Gehoben wird von Hand und mit Grund** — Renovate schlägt vor
  ([0129](0129-renovate-schlaegt-vor.md)), und ein Heben ist ein eigener PR mit
  der Liste der nativen Abnahmen aus `WAILS3.md`.
- **Die Hülle steht an zwei Stellen**: `internal/fenster` (Go) und
  `frontend/src/anbindung.ts`. `internal/app` kennt Wails nicht mehr — die
  Menüleiste ist ein eigenes Modell (`menuebaum.go`), und die Nähte zum Fenster
  setzt die Hülle. Ein Rückweg oder ein nächster Wechsel betrifft diese zwei
  Dateigruppen und nicht die Anwendung.

## Was dabei bewusst so gebaut ist

- **Die Bindungen werden nicht erzeugt.** Wails 3 kann Bindungen als
  JavaScript-Module schreiben (`wails3 generate bindings`); die Fassade dieses
  Projekts ist aber seit jeher handgeschrieben (`api.ts`), und `fassade.ts`
  liest sie an **einer** Stelle. `anbindung.ts` legt dort einen Ruf über
  `Call.ByName` für jede Methode ab — dieselbe Gestalt, die Wails 2 hinterließ,
  und damit dieselbe, die sechsundzwanzig Läufe als Vorrichtung benutzen.
- **Kein `wails3` im Bauweg** — hier weicht die Umsetzung von 0124 ab, das
  das Kommandozeilenwerkzeug in einem eigenen Modul erwartete. Das Bündel legt
  `scripts/bauen.sh` in fünf sichtbaren Schritten, und `make dev` startet Vite
  und das Programm selbst (`scripts/entwickeln.sh`); die Taskfile-Kette von
  Wails 3 brächte einen zweiten Projektaufbau und ein Werkzeug mit eigener
  Abhängigkeitsliste. `tools/wails/` entfällt damit ganz.
- **`wails.json` bleibt, aber nur als Ort der Fassung.** Kein Werkzeug von
  Wails liest sie mehr; `scripts/releaseversion.sh`, `scripts/bauen.sh` und
  der Wächter über `app.Fassung` lesen `productVersion` daraus. Alles andere
  darin wäre eine zweite Beschreibung dessen, was jetzt fest in
  `build/darwin/Info.plist` steht, und ist gestrichen. Der Name ist
  geschichtlich; ihn zu ändern berührte elf Stellen und ist kein Teil dieses
  Wechsels (§5, drei Züge).
- **Die Ereigniswerte gehen immer als Liste.** Wails 3 reicht keinen, einen
  oder mehrere Werte verschieden durch; eine Liste ist eindeutig
  (`fenster.nutzlast`).
- **Das Ablegen findet sein Feld über die Stelle.** Unter Wails 3 nimmt eine
  eigene Fläche über der Webansicht die Dateien entgegen, und die Webansicht
  sieht kein `dragover` mehr (`webview_window_darwin_drag.m`, nachgelesen).
  Das Feld ergibt sich aus der Stelle, die Go meldet (`data-bildfeld`).

## Was der Wechsel mitzieht

**`ProtonMail/go-crypto` steigt von 1.1.6 auf 1.3.0**, ohne dass diese
Anwendung es verlangt: Wails 3 nennt 1.3.0 in seinem `go.mod`, und Go nimmt
die höchste verlangte Fassung. Die Bibliothek prüft die Signatur einer
Freigabe ([0051](0051-die-signatur-wird-geprueft.md)). Die Läufe in
`internal/aktualisierung`, die selbst signieren und fremde, fehlende und
unlesbare Signaturen abweisen, sind mit 1.3.0 grün. **Was nicht gelaufen
ist**: der Lauf gegen eine echte Signatur von `gpg`
(`TestDieEchteFreigabesignaturPasstZumEingebettetenSchluessel`) — er braucht
eine gebaute Freigabe unter `build/freigabe/` und überspringt ohne sie. Er
läuft mit dem nächsten Freigabelauf, vor dem Veröffentlichen. **Wer Wails
hebt, prüft auch diese Zeile.**

Sonst bindet das Fenster **53** Module statt 56 (`go list -deps`, gemessen
am 2026-09-24): fort sind Wails 2 und fünf seiner Helfer, neu `adrg/xdg` und
`mattn/go-isatty`. Die Kommandozeilenwerkzeuge binden Wails gar nicht mehr —
das ist die Folge davon, dass `internal/app` es nicht mehr kennt.

## Die Zugänge, die kein Freigabebau hat (ENG-02.AK-05)

Wails 3 kennt drei Bauschalter, die einen Zugang öffnen:

| Schalter | Was er öffnet | Im Freigabebau |
| --- | --- | --- |
| ohne `production` | die Entwicklerwerkzeuge der Webansicht, das Kontextmenü „Element untersuchen" | **nein** — `scripts/bauen.sh` baut mit `production` |
| `mcp` | einen MCP-Server von Wails selbst, der das Fenster steuert | **nein** |
| `server` | einen HTTP-Server statt eines Fensters — Bindungen über das Netz | **nein** |

`scripts/release.sh` liest die Schalter am fertigen Programm (`go version -m`)
und bricht ab, wenn einer außer `production` darin steht. **Der fachliche
MCP-Server dieser Anwendung (`mse-mcp`) ist ein eigenes Programm** und hat
mit dem Schalter `mcp` nichts zu tun.

**`server` ist ein Werkzeug der Prüfung**: im Browser läuft dieselbe Anwendung
mit echten Bindungen — das, was unter Wails 2 nur `wails dev` im eigenen
Fenster konnte (`docs/SICHTPRUEFUNG.md`). Er ist kein Weg für Anwendende.

## Der Kreuzbau

**Unter Linux baut Wails 3 nur mit cgo** — gemessen am 2026-09-23:
`webview_window_linux.go` braucht C-Typen. Die Pipeline übersetzt für Linux
deshalb alles **außer** `main` und `internal/fenster`; für Windows baut sie
weiterhin alles. Was das kostet: ein Bruch in der Hülle unter Linux fiele
erst beim Bau auf einer Linux-Maschine auf — die es nach §12 nicht gibt.

## Die Ausgangsmessung und der Rückweg

**ENG-02 verlangt die nativen Kernwege vor und nach dem Wechsel** (AK-04), und
der Zweitblick über Phase 6.0 hatte vermerkt, die Ausgangsmessung gehöre
**vor** den Wechsel. Gemessen hat sie noch niemand — sie braucht einen
Menschen am Fenster.

**Abweichung, mit Grund:** der Wechsel geschieht trotzdem, und die Messung
bleibt möglich. Der letzte Stand mit Wails 2 trägt den Tag **`basis/wails2`**;
er baut mit den gepinnten Werkzeugen wieder dasselbe Programm
(`git checkout basis/wails2 && make bau`). Gemessen wird dort in einem eigenen
Bestand (`MSE_BESTAND`), wie UX-01.AK-02 es verlangt — die „Ausgangslage" ist
der gepinnte Bau und nicht der Zweig, auf dem er entstand.

**Der Rückweg (AK-06)**: der Wechsel ist **ein** PR. Scheitert eine native
Pflichtfunktion, wird er als Ganzes zurückgenommen; `internal/app` und das
Frontend tragen danach wieder Wails 2 wie vorher.

## Was verworfen wurde

| Verworfen | Warum |
| --- | --- |
| **Auf Wails 3.0 stabil warten** | Es gibt kein Datum. Die Planung setzt Wails 3, und je länger die Oberfläche auf Wails 2 wächst, desto mehr wandert mit |
| **Die erzeugten Bindungen** | Sie ersetzten die handgeschriebene Fassade durch eine zweite Beschreibung derselben Methoden — und die Vorrichtung der Läufe durch eine neue |
| **`/wails/runtime.js` statt des Pakets** | Er käme immer in der Fassung des Go-Moduls — aber ohne Typen, und `vite` müsste einen absoluten Import durchreichen. Das Paket in derselben Fassung und ein Wächter über die Gleichheit tragen dasselbe |
| **`wails3 dev` für die Entwicklung** | Es verlangt den Projektaufbau mit Taskfile; `make dev` startet Vite und das Programm selbst |
