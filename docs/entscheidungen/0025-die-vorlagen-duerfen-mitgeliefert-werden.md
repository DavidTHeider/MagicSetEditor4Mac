# 0025 — Die Vorlagen dürfen mitgeliefert werden, und trotzdem werden sie bezogen

**Stand:** gültig **zur Hälfte** — den zweiten Satz der Entscheidung hat
[0037](0037-die-vorlagen-gehen-ins-buendel.md) am 2026-08-21 umgedreht: die
Vorlagen **gehen** ins `.app`-Bündel. Wer nur hier liest, liest das Gegenteil
dessen, was heute ausgeliefert wird (§10 — ein Kopf, der mehr verspricht als
sein Rumpf hält).
**Ergänzt:** [0005](0005-keine-fremden-vorlagen.md)
**Datum:** 2026-08-17

## Lage

[0005](0005-keine-fremden-vorlagen.md) hält fremde Vorlagensammlungen aus dem
Repository. Drei Gründe standen darin: keine Lizenz, über ein Gigabyte,
Windows-Binärdateien und geschützte Grafiken.

Am 2026-08-17 hat der Urheber dieses Projekts die **Einwilligung** erteilt: das
volle Stil-Pack (ohne die Windows-Binärdateien) und die kanonischen Sets dürfen
mitgeliefert werden.

Damit fällt der **rechtliche** Einwand. Die anderen bleiben.

## Was gemessen wurde

| | |
| ---: | --- |
| `MagicSetEditorPacks/Full-Magic-Pack`, Stand 622160f (2026-08-07) | 376 `.mse-style`-Pakete |
| davon `data/` | **971 MB** |
| `Magic - Fonts/` | 26 MB |
| Windows-Binärdateien (`mse.exe`, `magicseteditor.exe`, `unins000.exe`) | fallen weg |
| kanonische Sets (Scryfall-Stand 2022-10-24) | **1,8 GB** in zehn Dateien, sechs davon über 227 MB |

## Entscheidung

**Sie kommen nicht ins Repositorium**, und zwar aus einem Grund, der mit Recht
nichts zu tun hat: GitHub weist eine Datei über 100 MB ab — auch in einem
privaten Repositorium, auch mit Einwilligung. Sechs der zehn kanonischen Sets
liegen darüber. Für das Pack käme ein Gigabyte in die Historie, das dort nie
wieder herausgeht.

**Sie kommen nicht ins `.app`-Bündel.** Eine Anwendung von einem Gigabyte, die
nach dem Start noch nichts kann, ist keine Erleichterung.

> **Dieser Satz gilt seit dem 2026-08-21 nicht mehr**
> ([0037](0037-die-vorlagen-gehen-ins-buendel.md)). Er stand auf einer Zahl,
> die niemand nachgemessen hatte — es sind **167 MB** und kein Gigabyte. Acht
> Pakete gehen seitdem nach `Contents/Resources/vorlagen`, und der Zettel im
> `.dmg` sagt es; ein Wächter hält ihn dagegen
> (`internal/werkzeuge/beigabenzettel_test.go`).

**Sie werden bezogen — und das ist der Weg, der ohnehin gebaut ist.**
`internal/bezug` holt ein Repositorium an einem **Stand** (nicht an einem
Zweig), prüft die Summe und packt sicher aus. Neu ist nur, dass die
Paketseite die Adresse und den Stand **vorausgefüllt** anbietet: wer die
Anwendung zum ersten Mal öffnet, drückt einen Knopf statt eine URL abzutippen.

Die kanonischen Sets liegen unter
`~/Library/Application Support/MagicSetEditor4Mac/kanonische-sets/` — dort, wo
die bezogenen Pakete und die Referenzbilder schon liegen. Die Läufe hinter dem
Build-Tag `fremdpakete` benutzen sie von dort.

## Was das kostet

**Ohne Netz gibt es keine Vorlagen.** Die Saat unter `saat/` bleibt deshalb, was
sie ist: klein, selbst geschrieben, mitgeliefert — genug, damit die Anwendung
beim ersten Start etwas zeigt.

**Und der Stand veraltet.** Ein fester Commit ist der Preis dafür, dass eine
Prüfsumme etwas festhält; wer die Vorgabe aktualisiert, ändert eine Zeile und
schreibt hier eine dazu.

---

## Nachtrag vom 2026-08-21: die zweite Hälfte ist zurückgenommen

„Sie kommen nicht ins `.app`-Bündel" stand auf einer Zahl, die niemand
nachgemessen hatte — **ein Gigabyte**. Nachgemessen sind es 167 MB für den
Schluss, den eine M15-Karte braucht, und 207 MB für das gebaute Bündel samt
Beispielset.

Der Rest dieses Vermerks gilt weiter: ins **Repositorium** kommen sie nicht.

Was jetzt gilt, steht in
[0037](0037-die-vorlagen-gehen-ins-buendel.md).
