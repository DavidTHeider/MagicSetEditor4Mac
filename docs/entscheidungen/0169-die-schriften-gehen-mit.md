# 0169 — Die Schriften der Vorlagen gehen mit

**Stand:** angenommen am 2026-10-05 · auf Anweisung des Auftraggebers ·
#2071.
**Dreht um:** `DRITTANBIETER.md` §5 („Ausgeliefert wird keine Schrift, die
nicht frei lizenziert ist"). **Ergänzt:**
[0025](0025-die-vorlagen-duerfen-mitgeliefert-werden.md),
[0037](0037-die-vorlagen-gehen-ins-buendel.md).

## Die Frage

**Eine frisch installierte Anwendung zeichnete Karten ohne Text.** Gefunden am
Freigabeabbild 0.6.0: `mse zeichnen` aus dem eingehängten `.dmg`, frisches
Zuhause — Rahmen und Bild standen, Name, Regeltext und Zahlen fehlten; die
Befunde nannten `Beleren Bold` und `MPlantin`. Die acht mitgelieferten Pakete
([0037](0037-die-vorlagen-gehen-ins-buendel.md)) bringen ihre Schriften nicht
mit; sie liegen im Pack **neben** `data/`, unter `Magic - Fonts/`. Die
Erststart-Notiz versprach trotzdem „sofort eine Karte". So ging schon 0.5.0
hinaus.

## Die Entscheidung

**`Magic - Fonts/` geht ins Bündel**, nach `Contents/Resources/vorlagen/` — 26
MB, 16 Einträge samt dem Unterordner der vorlagenspezifischen Schriften.
`scripts/mitliefern.sh` holt sie aus dem bezogenen Pack und bricht ab, wenn
sie fehlen. Die Anwendung findet sie ohne eine Zeile Go: sie sucht Schriften
über den ganzen Baum jeder Vorlagenwurzel, und `Resources/vorlagen` ist eine
(`internal/schriften`, `MitPaketen` — gemessen an der Liste „gesucht in" der
Meldung).

**Die Systemschriften und die selbst installierten bleiben vorn**: die Suche
geht erst die Orte von macOS durch, dann die Pakete (#140). Wer eine Schrift
gleichen Namens installiert hat, behält sie.

## Was das kostet, und wer es trägt

**Die Schriften haben je eigene Lizenzen, und die meisten schließen eine
Weitergabe aus.** Beleren gehört Wizards of the Coast, MPlantin geht auf
Monotypes Plantin zurück; das Pack führt keine Lizenz. Die Einwilligung vom
2026-08-17 ([0025](0025-die-vorlagen-duerfen-mitgeliefert-werden.md)) gilt dem
vollen Pack samt `Magic - Fonts/`, **kann aber keine Rechte übertragen, die
der Einwilligende selbst nicht hat.** Auf diesen Widerspruch hingewiesen, hat
der Auftraggeber am 2026-10-05 entschieden, die Schriften trotzdem
mitzuliefern, **auf seine Verantwortung für die Weitergabe**.

**Bis dahin galt das Gegenteil**, mit Grund: die Schriften wurden **bezogen** —
auf dem Rechner der Nutzenden von der Quelle selbst —, damit dieses Projekt
sie nicht weitergibt. Dieser Weg bleibt: wer das Pack bezieht, bekommt die
Schriften auch von dort.

## Verworfen

- **Nur ein Hinweis**, dass Kartentexte den Bezug des Packs brauchen. Er hätte
  die Weitergabe vermieden und die Freigabe mit einer Karte ohne Text
  ausgeliefert — der Auftraggeber hat das ausdrücklich nicht gewollt.
- **Nur die Schriften, die die acht Pakete nennen.** Kleiner, aber eine zweite
  Liste neben der des Packs, die beim nächsten Paket veraltet; und die
  Vorlagen, die jemand danach bezieht, nennen weitere.
