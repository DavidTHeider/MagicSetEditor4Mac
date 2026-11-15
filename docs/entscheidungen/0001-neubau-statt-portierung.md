# 0001 — Neubau statt Portierung, MIT statt GPL

**Stand:** in seinem Lizenzteil **überholt** durch
[0019](0019-gpl-2-0-statt-mit.md) (2026-08-16)
**Datum:** 2026-08-08

> **Was heute gilt:** dieses Projekt steht unter GPL-2.0-or-later, und der
> MSE2-Quelltext darf gelesen und daraus portiert werden. Der Vermerk hier
> bleibt unverändert stehen — er ist die Aufzeichnung seines Tages, und die
> Begründung darin ist nach wie vor richtig: unter MIT wäre der Blick in MSE2
> eine Lizenzverletzung gewesen. Geändert hat sich nicht die Rechtslage,
> sondern die Zielsetzung. Was der Wechsel kostet, steht in
> [0019](0019-gpl-2-0-statt-mit.md).

## Lage

Die Aufgabe lautete, Magic Set Editor 2 auf den Mac zu portieren und unter MIT
zu stellen. Beides zusammen geht nicht.

MSE2 steht unter **GPL-2.0** (`COPYING` im Repository twanvl/MagicSetEditor2).
Die GPL ist ansteckend: ein Werk, das GPL-Code enthält, muss selbst unter GPL
stehen. Das gilt für kopierten Code, für übersetzten Code und für Code, der als
Bearbeitung eines GPL-Werks entsteht. Eine MIT-Lizenz über einem solchen Werk
wäre nicht bloß riskant, sie wäre falsch.

Hinzu kommt, dass „Portierung" hier ohnehin kein technischer Vorgang wäre: MSE2
sind 2,4 MB C++ auf wxWidgets. Das Ziel ist Go, TypeScript und ein WebView. Es
gibt keine Zeile, die man übernehmen *könnte*, ohne sie neu zu schreiben.

## Entscheidung

**Neubau in Go und TypeScript, ohne Blick in den MSE2-Quelltext.**
Kompatibilität entsteht ausschließlich über das **Dateiformat**.

Der C++-Quelltext von MSE2 wird nicht gelesen — nicht zum Abschauen, nicht zum
Nachschlagen eines Sonderfalls. Erschlossen wird aus dem, was ein Anwender auch
sehen kann:

- den Datenpaketen selbst (`.mse-game`, `.mse-style`, …),
- dem Verhalten des laufenden Programms,
- öffentlicher Dokumentation.

Jede Formaterkenntnis wird in `docs/FORMATE.md` **mit Quelle** festgehalten:
welche Paketdatei, welche Stelle. Eine Erkenntnis ohne Quelle ist von einer
abgeschriebenen nicht zu unterscheiden.

Damit trägt MIT. Dateiformate und Schnittstellen sind kein schutzfähiges Werk;
ein aus Dateien erschlossenes Format ist es auch nicht.

## Verworfen

**Das Projekt unter GPL-2.0 stellen.** Wäre technisch der bequemere Weg — man
dürfte abschauen. Verworfen, weil MIT gefordert ist und weil die GPL jede
spätere Verwendung des Codes an dieselbe Lizenz bindet.

**MSE2 lesen, aber „nur zur Orientierung".** Verworfen, und zwar ausdrücklich.
Diese Grenze ist im Streitfall nicht belegbar: wer den Quelltext gelesen hat,
kann nicht nachweisen, dass das Ergebnis unabhängig entstanden ist. Der Wert der
Regel liegt darin, dass sie ohne Ausnahme gilt.

**Nur lesen, nicht schreiben.** Verworfen — ein Programm, das fremde Sets öffnet
und nicht zurückschreiben kann, ist eine Einbahnstraße und für die eigentliche
Arbeit unbrauchbar.

## Folgen

- Der Interpreter für MSE-Script muss vollständig neu entstehen (Phase 3). Das
  ist der teuerste Teil des Projekts und die Stelle, an der ein Blick in den
  fremden Quelltext am meisten sparte — genau deshalb steht die Regel hier.
- Verhaltensunterschiede zu MSE2 sind zu erwarten und werden **benannt**, nicht
  weggeräumt. Wo unser Ergebnis abweicht, gehört es in `docs/FORMATE.md`.
- `docs/FORMATE.md` ist nicht Dokumentation im üblichen Sinn, sondern der Beleg
  der Unabhängigkeit. Sie wird entsprechend geführt.
- `DRITTANBIETER.md` §1 hält dieselbe Regel für Außenstehende fest.
