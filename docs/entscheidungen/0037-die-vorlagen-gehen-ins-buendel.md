# 0037 — Die Vorlagen gehen ins Bündel, nicht ins Repositorium

**Stand:** entschieden am 2026-08-21 · **ergänzt** durch [0169](0169-die-schriften-gehen-mit.md): seit 0.6.0 gehen auch die Schriften der Vorlagen mit
**Betrifft:** [0025](0025-die-vorlagen-duerfen-mitgeliefert-werden.md), AGENTS.md §11, §12
**Werkstück:** #702

## Was sich geändert hat

[0025](0025-die-vorlagen-duerfen-mitgeliefert-werden.md) hielt fest, dass die
Vorlagen **mitgeliefert werden dürfen**, und entschied trotzdem dagegen — mit
zwei Sätzen, von denen nur noch einer trägt:

> **Sie kommen nicht ins Repositorium**, und zwar aus einem Grund, der mit Recht
> nichts zu tun hat: GitHub weist eine Datei über 100 MB ab.
>
> **Sie kommen nicht ins `.app`-Bündel.** Eine Anwendung von einem Gigabyte, die
> nach dem Start noch nichts kann, ist keine Erleichterung.

Der erste Satz gilt weiter. Der zweite stand auf einer Zahl, die niemand
nachgemessen hatte: **ein Gigabyte**. Nachgemessen sind es 167 MB.

## Die Zahlen

| Was | Größe |
| --- | --- |
| Alle zehn kanonischen Sets | 1,8 GB |
| Die sechs Farbsets | je 217–287 MB |
| Vanguard (107 Karten) | 8,6 MB |
| Ganzes `M15-Magic-Pack` | 805 MB |
| davon `magic-modules.mse-include` | **158 MB**, 3.300 Dateien |
| **Der Schluss, den eine M15-Karte braucht** | **167 MB** |
| Das gebaute `.app` mit Beigaben | **207 MB** |

Und was ein Renderlauf wirklich anfasst — gemessen über eine Sonde im
Paketleser, die danach wieder entfernt wurde:

| Lauf | gelesene Dateien | Menge |
| --- | --- | --- |
| Kanon, 120 Karten aus allen zehn Sets | 405 | 12 MB |
| **Vanguard ganz, 107 Karten, 0 Befunde** | **60** | **1 MB** |

## Die Entscheidung

**Ins Bündel, nicht ins Repositorium.** `scripts/mitliefern.sh` kopiert die
Pakete aus dem **bezogenen** Pack dieses Rechners und ein Beispielset aus den
kanonischen Sets ins `.app`. Wer die Anwendung lädt, hat beides; wer den
Quelltext klont, lädt 167 MB Binärkunst **nicht** mit.

**Nur im Freigabeweg.** Nicht als `postBuildHook` neben `scripts/saat.sh`: 167
MB bei jedem `wails dev` wären unbenutzbar.

**Ganz und nicht ausgeschnitten.** Der Ausschnitt wäre hundertmal kleiner — und
seine Vollständigkeit ließe sich nicht belegen: die Stilskripte setzen ihre
Pfade zusammen (`"cards/375 m15 simple/" + …`). Ein fehlendes Rahmenbild fiele
erst bei jemandem auf, der eine Karte in einen Rahmen setzt, den die Stichprobe
nicht hatte: sichtbar für ihn, unsichtbar für uns (AGENTS.md §10).

**Beispielset ist Vanguard.** 107 Karten, 8,6 MB, und der einzige der zehn, der
über **alle** seine Karten befundfrei zeichnet. Archenemy verlangt ein
`archenemy.mse-style`, das im Pack nicht liegt; die sechs Farbsets wiegen je
über 217 MB.

## Was das kostet

**Das `.dmg` wiegt statt weniger MB rund 200.** Für einen Karteneditor ist das
üblich — MSE2 selbst kommt mit seinen Paketen auf ein Vielfaches —, und es ist
der Preis dafür, dass nach dem ersten Start **sofort** eine Karte dasteht.

**Der Freigabeweg braucht das bezogene Pack.** `scripts/mitliefern.sh` bricht
ab, wenn es fehlt: ein Bündel ohne Vorlagen sähe aus wie eines mit und zeichnete
nichts.

**Die mitgelieferten Pakete veralten mit der Fassung.** Wer sie aktualisieren
will, bezieht neu und gibt neu frei. Der Paketbezug bleibt daneben stehen und
ist weiterhin der Weg zu allem, was **nicht** mitkommt — 319 der 327 Pakete des
Packs.

**Nachgetragen am 2026-08-28**: die beiden Zahlen sind der Stand vom
2026-08-24. `M15-Magic-Pack` führt heute **324** Pakete, also 316 über den
Bezug; das vollständige `Full-Magic-Pack` führt **526**. Was sich **nicht**
bewegt hat, ist die Zahl, die dieses Repositorium bestimmt: **acht** Pakete
gehen ins Bündel, und sie stehen in `scripts/mitliefern.sh`. Der Vermerk
behält seine Zahlen, weil er ein Datum hat — die Sätze, die *heute* gelten,
stehen in `README.md` und `docs/RELEASE.md`.

**Die Saat bleibt.** Sie ist der Stand für die Tests, für `wails dev` und für
den Fall, dass jemand ohne Freigabeweg baut.

## Wie geprüft

`TestNurAusDemBuendel` (hinter `fremdpakete`) startet die Anwendung mit einem
**leeren** Paketverzeichnis und allein den mitgelieferten Vorlagen: 107 Karten,
die erste gezeichnet mit 375×523 und **null** Befunden.
