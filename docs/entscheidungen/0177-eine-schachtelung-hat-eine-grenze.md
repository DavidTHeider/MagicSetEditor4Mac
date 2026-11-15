# 0177 — Eine Schachtelung hat eine Grenze, und sie greift beim Zerteilen

**Stand:** angenommen am 2026-10-07 · BK-08
(#2134).

## Die Frage

**Der Zerteiler von MSE-Script stieg rekursiv hinab, ohne zu zählen.** Das
Budget der Auswertung (`SchluesselZuTief`) setzt erst nach dem Zerteilen an;
davor brachten ein paar hundert Kilobyte Klammern aus einem fremden Paket
den Stapel zum Überlaufen — ein `fatal error`, den kein `recover` fängt, und
mit ihm endete das Programm samt jedem offenen Tisch. **Und nicht nur
Klammern**: Vorzeichenketten, Listen, Aufrufe, Blöcke, Bedingungen,
Zuweisungen und eingebettete Textskripte — die bekamen je Zeichenkette einen
**eigenen** Zerteiler und hätten eine Grenze je Zerteiler umgangen.

MSE2 zählt ebenso wenig (`src/script/parser.cpp`, Zeilen 444–454 und
718–719, Fassung f9d9356 vom 2021-07-06). Die Grenze ist deshalb **eine neue
Absicherung dieser Portierung** und keine Regel aus dem Altcode.

## Die Messung

Am 2026-10-07 über die 1.052 Textdateien der Pakete in `full-magic-pack`:
die tiefste Schachtelung von `(`, `[` und `{` ist **13** — in
`magic.mse-game/script`; danach `magic-blends.mse-include/new-blends` mit 12.

## Die Entscheidung

**`HoechsteSchachtelung = 256` Ebenen**, gezählt an den zwei Stellen, an
denen geschachtelt wird: jeder Ausdruck (`ausdruck`) und jedes weitere
Vorzeichen einer Kette. **Die rechte Seite eines Operators zählt nicht** —
sie liest eine Stufe höher und steigt damit höchstens um die Zahl der
Vorrangstufen; gemessen an `1 or else 1 or else …`. **Ein eingebettetes
Textskript erbt die Tiefe**, in der es steht.

Über der Grenze endet das Zerteilen mit `SchluesselZuTiefGeschachtelt` an der
Stelle, an der die Ebene beginnt — bei einem Textskript an der Stelle der
Zeichenkette. **Zu tief ist kein Text**: ein eingebetteter Ausdruck, der sich
nicht zerteilen lässt, fällt sonst auf Text zurück (#345); dieser nicht.

Eine Million Klammern endet mit diesem Fehler und nicht mit dem Prozess; ein
Drittel der Grenze geht in jeder Klasse durch.

**Zwei Eingänge, nicht einer** — nachgetragen nach dem Zweitblick über E
(Befund 1): neben `Zerteile` zerteilt `AuswertenTextskript` im
Zeichenkettenmodus und verwarf den Fehler, weil er vor der Grenze keinen
haben konnte. Aus dem verworfenen Befund wurde ein leerer Knoten und ein
`panic` — bei Erinnerungstexten, Parameterhilfen und Stilbildern. Jetzt
reicht er den Befund durch; `TestEineZuTiefeSchachtelungImTextskriptIstEinFehler`
fährt jede Rekursionsklasse auch durch diesen Eingang.

## Was es nicht ist

- **Keine Grenze der Dateigröße.** Eine Folge gleichrangiger Operatoren
  (`1 + 1 + …`) läuft in einer Schleife und bleibt erlaubt, wie lang sie
  auch ist.
- **Keine Änderung der Sprache**: alles, was ein Paket bis 256 Ebenen
  schreibt, liest sich wie vorher — und die tiefsten echten haben 13.
