# 0024 — Der C++-Quelltext ist die Quelle der Wahrheit, nicht ein Nachschlagewerk

**Stand:** gültig
**Datum:** 2026-08-17

## Lage

[0001](0001-neubau-statt-portierung.md) hat den **Neubau** festgelegt: kein
Blick in den fremden Quelltext, Kompatibilität allein über das Dateiformat.
[0019](0019-gpl-2-0-statt-mit.md) hat den Blick erlaubt, weil die Lizenz
gewechselt ist — aber die Arbeitsweise blieb, wie sie war: **erschließen aus
den Datenpaketen, im Quelltext nachschlagen, wenn es klemmt.**

Diese Reihenfolge ist umgedreht worden.

## Was der bisherige Weg gekostet hat

Am 2026-08-17 sind an **einem** Tag sieben Fehler behoben worden, die alle
dieselbe Gestalt hatten: eine Regel war **erschlossen** statt **nachgelesen**,
sie klang plausibel, und sie war falsch.

| Befund | Was erschlossen war | Was im Quelltext steht |
| --- | --- | --- |
| #382 | eine Schriftfarbe ist ein Farbname | 2.321 von 3.011 sind **Skripte** |
| #388 | ohne `alignment:` wird eingepasst | `ChoiceStyle::ChoiceStyle(…) : alignment(ALIGN_STRETCH)` |
| #390 | eine Schleife sammelt ihre Ergebnisse | „add all results from the body, start with nil" |
| #393 | ein Schriftname ist ein Name | 664 von 3.119 sind Skripte |
| #396 | ein Feldname steht so da, wie er heißt | `canonical_name_form`: `' '` → `'_'` |
| #398 | ein Feldwert steht hinter dem Doppelpunkt | er darf ein eingerückter Block sein |

**Jeder dieser Fehler fiel still aus.** Ein Textfeld ohne Wert wird gar nicht
gezeichnet; eine schwarze Schrift auf schwarzem Grund ist gezeichnet und
unsichtbar. Keiner erzeugte einen Befund, keiner machte einen Test rot — und
zusammen ließen sie eine Karte entstehen, die aussah wie eine Karte und keine
war.

**Jeder einzelne stand im Quelltext, und zwar in zwei bis zwanzig Zeilen.**

## Entscheidung

**Der C++-Quelltext von MSE2 ist die fachliche und technische
Referenzimplementierung.** Wo Verhalten, Algorithmen, Zustandsübergänge,
Datenformate, Validierungen, Berechnungen oder Randfälle dort schon
implementiert sind, wird **dort nachgelesen** — nicht erschlossen, nicht
geraten, nicht aus Namen abgeleitet.

Die Reihenfolge lautet ab jetzt:

1. Die Stelle im Altcode finden.
2. Ihr Verhalten verstehen, samt Aufrufern und Sonderfällen.
3. Es durch Tests festhalten.
4. Erst dann in Go schreiben.

Was das im Einzelnen heißt, steht in `AGENTS.md` §15 — verbindlich, gleichlautend
in `CLAUDE.md`.

## Was das **nicht** heißt

- **Keine Syntaxübersetzung.** Portiert wird Verhalten, nicht Zeilen; §11 hält
  das schon fest, und es bleibt.
- **Keine Architektur aus 2007.** Der Schnitt in fünf Schichten und acht
  Fachbibliotheken (§2) bleibt, ebenso Go-Idiome statt RAII und Ausnahmen.
  **Semantik konservativ, Architektur modern.**
- **Kein Verzicht auf die Datenpakete.** Sie bleiben die Quelle für alles, was
  *in* den Vorlagen steht — Feldnamen, Zählungen, Häufigkeiten. Der Quelltext
  sagt, was das Programm damit *tut*.

## Preis

**Es ist langsamer.** Wer vor jeder Regel den Altcode liest, braucht länger als
wer sie errät. Der Tausch ist bewusst: die sieben Fehler oben haben mehr Zeit
gekostet als das Nachlesen gekostet hätte — und sie waren nicht die ersten.

**Und es ist nicht überall möglich.** Wo im Altcode keine Entsprechung existiert
— die Oberfläche dieses Projekts, sein Bestand, seine Pipeline —, gibt es nichts
nachzulesen. Dort gilt weiter, was immer galt; die Kennzeichnung `new` in der
Portierungskarte sagt es ausdrücklich, und sie darf **nicht** benutzt werden, um
eine unterlassene Suche zu verdecken.
