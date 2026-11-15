# 0007 — Kein entfernter Lauf, und was das kostet

**Stand:** angenommen
**Datum:** 2026-08-08

## Lage

`AGENTS.md` §4 verlangte bis heute einen entfernten Lauf über GitHub Actions —
vor jedem Release, nach jeder Änderung an der Pipeline und bei allem
Plattformabhängigen. Der Grund war gut: er prüft auf einer **anderen**
Plattform, und genau dort zeigt sich, was lokal unsichtbar bleibt.

Beim ersten Versuch am 2026-08-08 kam er nicht zustande. Nicht am Code — GitHub
wies ihn ab, bevor ein einziger Schritt lief:

> The job was not started because recent account payments have failed or your
> spending limit needs to be increased.

Das ist eine Kontosache und nichts, was in diesem Repository zu beheben wäre.

## Entscheidung

**Es gibt keinen entfernten Lauf.** Die Pipeline läuft lokal, über
`make ci` und den `pre-push`-Hook, und das ist die einzige Prüfung.

Die Workflow-Datei `.github/workflows/ci.yml` bleibt liegen — sie löst
ausschließlich von Hand aus und wird nicht ausgelöst. Sie steht da als fertiger
Weg für den Fall, dass sich die Kontolage ändert, **nicht** als Zusage, dass
etwas geprüft würde.

## Verworfen

**Die Kontolage klären und den Lauf beibehalten.** Verworfen als Entscheidung,
nicht als Möglichkeit: es ist eine ausdrückliche Vorgabe, dass kein entfernter
Lauf stattfindet. Wer sie ändert, ändert zuerst diese Datei.

**Die Workflow-Datei löschen.** Verworfen — sie kostet nichts, solange
unmissverständlich dasteht, dass sie nicht läuft. Der Kopf der Datei sagt es,
und `AGENTS.md` §4 sagt es auch.

**So tun, als sei die lokale Pipeline gleichwertig.** Verworfen, und das ist der
eigentliche Inhalt dieses Vermerks. Sie ist es nicht.

## Folgen — und die gehören ausgesprochen

Was jetzt **nicht** geprüft wird, und zwar dauerhaft:

| Blinde Stelle | Warum sie lokal unsichtbar bleibt |
| --- | --- |
| **Alles Plattformabhängige jenseits der Übersetzung** | Die Stufe `types` baut für Windows und Linux, die Stufe `lint` prüft dort. Aber gebaut wird nicht gleich gelaufen — kein Test führt je auf einer anderen Plattform aus |
| **Der Stand aus der Sperrdatei** | Der entfernte Lauf hätte mit `npm ci` begonnen und `node_modules` neu gebaut; lokal wird genommen, was daliegt. Wer einmal `npm install <paket>` lief, prüft seither gegen andere Fassungen |
| **Eine Umgebung, die nicht diese ist** | Jede Werkzeugversion, jeder Pfad, jede Einstellung dieses Rechners geht in jedes Ergebnis ein. Eine grüne Pipeline sagt: „auf diesem Mac grün" |
| **Ein zweiter Rechner überhaupt** | Kein Lauf auf fremder Hardware. **Teilweise geschlossen:** ein frischer Klon in ein leeres Verzeichnis ist am 2026-08-08 vollständig durchgelaufen — `npm ci`, `make tools`, `make ci` grün. Das prüft die Vollständigkeit des Repositoriums, nicht eine zweite Plattform |

Diese Liste ist kein Vorwurf, sondern der Preis. Sie gehört in `docs/RELEASE.md`
und in die Release-Notiz jeder Fassung: **geprüft ist eine Maschine, nicht eine
Plattform.**

Wo es billig zu haben ist, wird die Lücke verkleinert:

- Die Stufe `types` übersetzt weiter für Windows und Linux (`docs/PORTIERUNG.md`).
- Die Stufe `lint` läuft über alle drei `GOOS`, damit Dateien hinter Build-Tags
  nicht ungeprüft bleiben.
- Ausgeliefert wird ohnehin nur macOS auf Apple Silicon
  ([0003](0003-keine-notarisierung.md)) — die Plattform, auf der auch gebaut wird.
