# 0143 — Zwei Phasen nach der Planung: Schnittstellen und Bereichsschalter, Entschlackung der Oberfläche

**Stand:** entschieden am 2026-09-29 · **Anlass**: drei Anforderungen, die
nach der Planung von 0.6.0 hinzugekommen sind, ausdrücklich für dieses Release

## Worum es geht

Die Iteration 0.6.0 hatte zwölf Phasen (#1660).
Drei Anforderungen kamen während der Umsetzung hinzu:

1. **Reviews zur Entschlackung und Verbesserung der grafischen Oberfläche** —
   je Fläche eines, gegen das Artboard und seine Prinzipien, damit keine
   Fläche zu viele Funktionen in seltsamer Anordnung trägt.
2. **CLI, REST und MCP per Knopf einrichten und widerrufen**, samt der
   Einstellungen, die sie auf einem Mac brauchen.
3. **Bereiche und Funktionen abschaltbar**, vorgabemäßig alle an — um Neues
   kontrolliert einzuführen und Altes kontrolliert stillzulegen.

## Entscheidung

**Zwei neue Phasen und kein Zuwachs an bestehenden** (AGENTS.md §13: eine
Phase wird nicht stillschweigend größer):

| Phase | Issue | Inhalt |
| --- | --- | --- |
| **6.12** — Schnittstellen einrichten und Bereiche schalten | #1905 | OPS-09 (#1907), OPS-10 (#1908) |
| **6.13** — Entschlackung der Oberfläche | #1906 | UXR-01 bis UXR-24 (#1909 bis #1932) |

**Die Reihenfolge**: beide nach den Phasen, die noch laufen (6.5, 6.7, 6.8,
6.10), und **vor der Freigabephase 6.11** — sie ändern, was ausgeliefert
wird, und 6.11 nimmt es ab. 6.12 vor 6.13, damit die Reviews der
Einstellungen deren neue Gruppen schon vorfinden.

**Der Schnitt von 6.13 folgt der Fläche, nicht dem Seitenschlüssel.** Von
den 25 Schlüsseln sind dreizehn seit UX-02 Muster innerhalb einer Fläche —
eine Schublade, ein Popover, ein Schritt eines Assistenten —; ein Review je
Schlüssel risse zusammen, was zusammen bedient wird. Vorn steht ein
Querschnitt (**UXR-01**: das Prüfmaß — was eine Fläche zu viel hat, gemessen
statt gefühlt), hinten ein Abschluss (**UXR-24**: dieselbe Messung nach dem
Umbau und die Inventurdifferenz), dazwischen zweiundzwanzig Flächen nach der
Zielstruktur des Prototyps.

**(2) setzt OPS-03 fort und nicht neu an.** Die Verknüpfungen nach
`~/.local/bin` bestehen; OPS-03.AK-02 sah markierte, umkehrbare
Profiländerungen vor, gebaut ist „das Shellprofil fasst sie nicht an". OPS-09
schließt diese Lücke und nimmt REST und MCP hinzu.

**(3) ist kein Sicherheitsriegel.** Ein abgeschalteter Bereich ist verborgen,
nicht gesichert; die Rollen aus `AUTORISIERUNG.md` gelten unverändert. Das
steht in OPS-10 als Abnahmekriterium, damit niemand einen Schalter später für
eine Berechtigung hält.

## Verworfen

- **Ein Review-Issue je Seitenschlüssel (25).** Siehe oben — der Schnitt
  träfe Muster statt Flächen.
- **Ein einziges großes Review-Issue.** Es ließe sich nicht abnehmen: eine
  Fläche, die fertig ist, wartete auf alle anderen.
- **Die Reviews in UX-15 nachtragen.** UX-15 ist abgeschlossen und hatte ein
  engeres Ziel (kurze Texte, Hilfesymbol); es nachträglich zu erweitern
  machte eine abgenommene Phase wieder offen.

## Folgen

- Die Iteration hat **vierzehn** Phasen; `ROADMAP.md`, das Projekt (Feld
  *Phase* mit 6.12 und 6.13) und das Iterationsissue nennen sie.
- `docs/UIREVIEW.md` entsteht mit UXR-01 und trägt je Fläche die Messwerte
  vorher und nachher.
- Die Zuordnung Issue → Artboard in `PROTOTYP-0.6.0.md` führt die neuen
  Issues; ihr Wächter hält sie.
