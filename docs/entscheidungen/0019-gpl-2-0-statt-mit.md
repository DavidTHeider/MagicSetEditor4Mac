# 0019 — GPL-2.0 statt MIT, und damit der Blick in MSE2

**Stand:** angenommen
**Datum:** 2026-08-16
**Ersetzt:** [0001](0001-neubau-statt-portierung.md) in seinem Lizenzteil

## Lage

[0001](0001-neubau-statt-portierung.md) hat vor acht Tagen das Gegenteil
entschieden, und die Begründung dort ist unverändert richtig: die GPL ist
ansteckend, MSE2 steht unter GPL-2.0, und wer daraus ableitet, kann sein
Ergebnis nicht unter MIT stellen. Was sich geändert hat, ist nicht die
Rechtslage, sondern die **Zielsetzung** — der Auftraggeber dieses Projekts hat
am 2026-08-16 entschieden, dass die Anwendung vollständig funktionieren soll und
dass MIT dafür weichen darf.

Der Anlass ist messbar. Der Stand am Tag der Entscheidung, über die 190 Stile von
`M15-Magic-Pack`:

| | |
| --- | ---: |
| Feldstile insgesamt | 5.950 |
| davon zu einer Bildebene gedeutet | 310 |
| Befunde | 9.883 |

Die Befunde häufen sich nicht an den Nähten dieses Programms, sondern an der
**Sprache**: 1.606-mal rechnet `font size` mit einem Wert, den MSE offenbar
anders auffasst als wir; 570-mal verlangt ein `replace` einen Vorausblick, den
RE2 nicht kennt; 327-mal vergleicht `>` zwei Werte, die bei uns beide `nichts`
sind; 229-mal steht eine Zeichenkette dort, wo wir einen Wahrheitswert
verlangen. Das sind keine Fehler im Sinne von Versehen. Das sind
**Semantikentscheidungen**, und jede einzelne aus Beobachtung zu erschließen
kostet mehr, als das ganze Projekt wert ist.

## Entscheidung

**Dieses Repository steht ab sofort unter GPL-2.0-or-later** (`LICENSE`).

Damit fällt die Regel aus [0001](0001-neubau-statt-portierung.md) und aus
`AGENTS.md` §11, den C++-Quelltext von MSE2 nicht zu lesen. Er darf gelesen
werden, und es darf daraus portiert werden — übersetzt, umgeschrieben,
übernommen.

**Was dafür gilt, ohne Ausnahme:**

- **Herkunft wird benannt.** Wo eine Verhaltensregel aus MSE2 stammt, steht das
  im Quelltext an der Stelle, samt Datei und Fassung. Nicht aus Höflichkeit: die
  GPL verlangt, dass die Bearbeitung eines Werks als solche kenntlich ist
  (§2 a), und ein späterer Leser muss unterscheiden können, was erschlossen und
  was übernommen wurde.
- **`docs/FORMATE.md` bleibt.** Ihr Zweck ändert sich: sie war der Beleg der
  Unabhängigkeit, sie ist jetzt die Beschreibung des Formats. Die Angabe
  `Quelle:` unterscheidet ab jetzt zusätzlich, ob eine Erkenntnis aus einem
  Datenpaket oder aus MSE2 stammt.
- **Fremde Inhalte bleiben ausgenommen.** Kartenrahmen, Symbole, Schriften,
  Regeltexte und Marken sind von dieser Lizenz nicht gedeckt und waren es nie
  ([0005](0005-keine-fremden-vorlagen.md)). Der Geltungsbereich steht
  unverändert am Ende von `LICENSE`.

## Was das kostet

Das gehört ausgesprochen, weil eine Lizenzentscheidung sich nicht
zurücknehmen lässt:

- **Der Code ist nicht mehr frei verwendbar.** Wer ihn einbaut, muss sein
  eigenes Werk unter GPL-2.0 stellen. Ein Unternehmen, das ihn in ein
  geschlossenes Produkt nehmen wollte, kann das nicht mehr.
- **Der Weg zurück ist versperrt.** Sobald eine Zeile aus MSE2 abgeleitet ist,
  lässt sich MIT nicht wiederherstellen, ohne genau diese Zeilen zu entfernen
  und nachzuweisen, dass nichts davon übrig ist. In der Praxis heißt das: nie.
- **Abhängigkeiten müssen vereinbar sein.** Bisher galt „keine GPL-Abhängigkeit"
  (`DRITTANBIETER.md`). Die Richtung dreht sich: eine Abhängigkeit unter Apache
  2.0 ist mit GPL-**2.0** nicht vereinbar. Weil `LICENSE` „version 2 or any
  later version" sagt, löst sich das über GPL-3.0 — aber es ist eine Prüfung,
  die vorher nicht nötig war.
- **Die vier Wochen Arbeit an `msescript` bleiben.** Sie werden nicht
  weggeworfen, sondern ergänzt. Was aus Beobachtung richtig erschlossen wurde,
  ist richtig geblieben.

## Verworfen

**Bei MIT bleiben und weiter erschließen.** Der ehrliche Preis stand oben:
5.640 von 5.950 Feldstilen deuten nicht, und die Ursachen liegen in
Sprachdetails, die man einem laufenden Programm nur über sehr viele gezielte
Versuche abliest. Das ist möglich und es ist absehbar teurer als das Ziel.

**MIT behalten und trotzdem lesen.** Verworfen aus demselben Grund, aus dem
[0001](0001-neubau-statt-portierung.md) „nur zur Orientierung" verworfen hat:
das wäre keine Entscheidung, sondern eine Lizenzverletzung mit Beschriftung.

**LGPL.** Sie erlaubte das Einbinden in geschlossene Werke, deckt aber die
Übernahme von GPL-2.0-Code nicht — MSE2 steht unter GPL, und eine Bearbeitung
davon darf nicht unter eine schwächere Copyleft-Lizenz.

## Folgen

- `LICENSE` trägt den GPL-2.0-Text samt Geltungsbereich.
- `AGENTS.md` und `CLAUDE.md` §11 sind umgeschrieben; die Regel „kein Blick in
  den MSE2-Quelltext" ist ersatzlos weg, an ihre Stelle tritt die Pflicht zur
  Herkunftsangabe.
- `README.md`, `DRITTANBIETER.md`, `wails.json` und der „Über"-Dialog in
  `main.go` nennen die neue Lizenz.
- `internal/werkzeuge/lizenz_test.go` hält die Angaben gegeneinander: eine
  Lizenz, die an sechs Stellen steht, läuft sonst auseinander, und die im Bündel
  ist die, die gilt.
