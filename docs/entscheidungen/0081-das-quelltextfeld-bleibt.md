# 0081 — Das Quelltextfeld bleibt; CodeMirror 6 kommt nicht

**Stand:** angenommen
**Datum:** 2026-09-09
**Anlass:** Werkstück U7 von Phase 5.1
(#1418) —
die Vorlagenwerkstatt braucht Klammern, Suche, Vervollständigung, Hover und
Sprung zur Definition; das Issue verlangt, CodeMirror 6 gegen das vorhandene
`Quelltextfeld` zu evaluieren, und setzt zwei Bedingungen: gemeinsam mit der
Skriptkonsole nutzbar, und **keine zweite MSE-Script-Semantik in TypeScript**.

## Lage

**Was am 2026-09-09 galt, ist nachgelesen** (Stand `7805e3e1`):

| Frage | Antwort |
| --- | --- |
| Was das Feld ist | ein `<textarea>` über einem `<pre>`, Zeichen für Zeichen deckungsgleich (`bausteine/Quelltextfeld.tsx`, seit #1164) |
| Woher die Farben kommen | aus dem Lexer in Go — `Skriptfarben` für MSE-Script, `Paketfarben` für Paketdateien (#1181); ein zweiter Lexer in TypeScript liefe an dem Tag auseinander, an dem einer eine Regel lernt |
| Was es schon kann | Zeilennummern, Fehlerstelle (`stelle`), Vervollständigung aus `namen` (die Skriptnamen einer Karte aus Go), zwei Sprachen |
| Was es nicht kann | Klammernpaar, Suche und Ersetzen im Puffer, Hover, Sprung zur Definition, Umbruch |
| Wer es benutzt | die Skriptkonsole, das Kartenskript und der Vorlagenbereich — **eine** Bedienung an drei Stellen |

**CodeMirror 6, nachgeschlagen am 2026-09-09 über `npm view`**: `@codemirror/view`
6.43.11, `state` 6.7.4, `language` 6.12.4, `autocomplete` 6.20.3, `search`
6.7.2, `commands` 6.11.0, `lint` 6.9.7, `@lezer/highlight` 1.2.3 — alle MIT,
ohne React-Peer, Abhängigkeiten nur `crelt`, `style-mod` und `w3c-keyname`.
Entpackt zusammen rund 2,8 MB (mit Quellkarten), gebündelt grob 300 bis 400 kB.
Die Eingabe ist ein `contenteditable` mit `role="textbox"`; die Tastatur ist
vollständig, VoiceOver liest zeilenweise.

## Entscheidung

**Das `Quelltextfeld` wird ausgebaut. CodeMirror 6 wird nicht aufgenommen.**

Was U7 braucht, bekommt das Feld — aus denselben Quellen, aus denen es heute
lebt:

| Fähigkeit | Woher |
| --- | --- |
| Klammernpaar | aus den Abschnitten des Go-Lexers, der Klammern kennt; das Feld hebt das Gegenstück hervor, wenn der Zeiger daneben steht |
| Suche und Ersetzen im Puffer | im Feld selbst, `⌘F` innerhalb des Feldes; die Volltextsuche des Sets bleibt, was sie ist |
| Vervollständigung | `namen` wie bisher — Skriptnamen, Feldnamen, und in einer Paketdatei die Schlüssel des Formats |
| Hover | die Herkunft des Namens **unter dem Schreibzeiger** — Art, Datei, Zeile aus dem Symbolindex in Go (`Paketsymbole`), im Inspector statt als Tooltip: über einer Textfläche gibt es kein Wort unter der Maus, und ein Satz neben dem Feld liest sich auch mit der Sprachausgabe. **Hier stand `Parameterhilfen`** — das sind die Menüs eines Schlüsselworts, nicht die Hilfe eines Skriptnamens; nachgelesen beim Bau |
| Sprung zur Definition | die Paketdatei, die den Namen definiert — aus demselben Index (`Paketsymbole`), dazu die Referenzen (`Symbolverwendungen`); die Werkstatt öffnet die Datei und setzt den Zeiger. **Hier stand „Suche im Text“** — das wäre die zweite Lesart des Formats in TypeScript gewesen, die das Issue verbietet |

## Warum

1. **Die Sprache bleibt in Go.** Eine Lezer-Grammatik oder ein
   `StreamLanguage` wäre die zweite Semantik, die das Issue verbietet. CodeMirror
   ohne eigene Sprache — Einfärbung als Decorations aus den Go-Abschnitten — ist
   möglich, aber dann leistet es nur Klammern, Suche und Falten; dafür 300 kB
   und eine zweite Bedienung neben dem Feld, das die Konsole schon hat.
2. **Die Abdeckung.** Fremdcode steht nicht unter der 100-%-Pflicht (§3), die
   Integrationsschicht schon — und jsdom kann CodeMirror nur ohne Layout: jeder
   Test dieser Schicht liefe gegen ein Doppel, nicht gegen das Feld. Ein
   `<textarea>` kennt jsdom vollständig; die heutigen Tests des Feldes greifen
   hinein.
3. **Eine Bedienung, nicht zwei.** Konsole, Kartenskript und Werkstatt teilen
   das Feld. Wer der Werkstatt einen anderen Editor gibt, hat zwei Sätze
   Tastenkürzel und zwei Arten Fehlerstelle in einer Anwendung.

## Verworfen

- **CodeMirror 6 mit Decorations aus Go** — technisch tragfähig; verworfen wegen
  1 und 2.
- **Monaco** — deutlich größer, dieselben Einwände, dazu ein eigener Worker.
- **Ein eigener Lexer in TypeScript**, damit CodeMirror etwas zu färben hat —
  genau das, was das Issue ausschließt (§10: eine zweite Quelle für dieselbe
  Regel läuft auseinander).

## Preis

Kein Falten, keine Mehrfachzeiger, keine Minikarte, kein Umbruch. Wer das
braucht, öffnet die Datei im Editor seiner Wahl — der Pfad steht in der
Werkstatt, und die Werkstatt bemerkt die fremde Änderung über den Stand der
Datei (`PaketdateiLesen`/`PaketdateiSchreiben`, §7 Regel 2).

## Folgen

- `bausteine/Quelltextfeld.tsx` bekommt den Sprung (`sprungziel`) und meldet
  den Zeiger (`aufZeiger`); Klammernpaar und Suche hatte es schon. Jede neue
  Fähigkeit gilt an allen drei Stellen, die das Feld benutzen.
- `DRITTANBIETER.md` bleibt, wie es ist — es kommt keine Abhängigkeit hinzu.
- Wer die Entscheidung umdrehen will, ändert zuerst diesen Vermerk und legt
  die Messung daneben: Bündelgröße, Abdeckung der Integrationsschicht und ein
  Lauf mit VoiceOver.
