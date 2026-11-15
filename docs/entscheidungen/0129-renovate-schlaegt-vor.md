# 0129 — Renovate schlägt vor, entschieden wird von Hand

**Stand:** entschieden am 2026-09-23
(#1779)

## Was galt

**Eine `renovate.json` ohne Inhalt.** Der Einrichtungs-PR
(#1777) kam in
der Nacht zum 2026-09-22 nach `main`, als Merge-Commit und nicht als Squash
(§5). Damit galt Renovates Vorgabe — und die kennt keine der Pinnungen aus §1:
auf dem Dashboard (#1784)
standen TypeScript 7, ESLint 10, `@eslint/js` 10 und npm 12 zur Freigabe,
jeweils eine Stelle, die §1 aus **gemessenen** Gründen zurückhält.

Vier Sicherheits-PRs lagen offen (#1780–#1783), alle in den Werkzeugmodulen,
und sieben weitere standen im Stau, zwei davon in der ausgelieferten
Anwendung: `golang.org/x/crypto` und `golang.org/x/image`.

## Was gilt

**Renovate bleibt — als Vorschlag, nicht als Entscheidung.**
`renovate.json5` sagt es mit Gründen:

- **Jede gewöhnliche Aktualisierung braucht eine Freigabe** auf dem
  Dashboard (`:dependencyDashboardApproval`). Das Dashboard ist damit die
  Liste dessen, was es gäbe.
- **Sicherheitsaktualisierungen kommen ohne Freigabe** — die Pflicht gilt für
  sie nicht (nachgelesen über Context7 am 2026-09-23), und `osvVulnerabilityAlerts`
  ist an.
- **Was §1 zurückhält, wird nicht vorgeschlagen**: TypeScript unter 6.1,
  ESLint und `@eslint/js` unter 10, `@types/node` unter 25 (Node ist 24),
  Wails 2 unter 3 (Wails 3 ist ENG-02 und ein eigener Zug, §5). Die `go`- und
  `toolchain`-Zeilen gar nicht: Go hängt über `GOTOOLCHAIN` an drei Stellen,
  und eine Zeile allein wären zwei Pipelines.
- **Gruppen, wo §1 Kopplungen misst**: Mantine (`hooks` verlangt dieselbe
  Zahl) und Vitest (`coverage-v8` ebenso), beide auf die Stelle genau.
- **Actions bleiben an Hashes**, und der Kommentar mit der Fassung steht
  jetzt **hinter** dem Hash — nur so hält Renovate beides nach.
- **Gemergt wird nie von selbst** (`automerge: false`). Es gibt nur den
  lokalen Lauf ([0007](0007-kein-entfernter-lauf.md)); ein Renovate-PR ist
  erst gut, wenn `make ci` auf diesem Rechner für genau seinen Stand grün
  war — also wird er auf einen eigenen Zweig gezogen, geprüft und dort
  gemergt.

**Die elf Sicherheitsaktualisierungen sind in einem Zug gehoben**, nicht in
elf PRs — dieselben drei `go.mod` und elf Pipelines wären Arbeit ohne
Erkenntnis. `x/image` ist dabei nachgesehen: geändert haben sich der
WebP-Dekodierer und Assembler für amd64, nicht die Glyphenrasterung.

## Verworfen

- **Renovate wieder abschalten.** Es nähme die Sicherheitsvorschläge mit,
  und die sind der Teil, der sich lohnt: elf standen offen oder im Stau.
  `govulncheck` in der Stufe `security` nennt solche Funde zwar, wird aber
  nur rot, wenn der Code sie **aufruft** — und schreibt keine Änderung.
- **Die gepinnten Pakete ganz ausnehmen** (`ignoreDeps`). Dann käme auch
  keine Sicherheitsaktualisierung für sie, und ein TypeScript-Patch unter 6.1
  bliebe genauso ungesehen wie der Sprung auf 7.
- **Automatisch mergen, was grün ist.** „Grün" hieße hier: kein Lauf. Es
  gibt keinen entfernten Lauf, der einen Renovate-PR prüfen würde.
