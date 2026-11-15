# 0203 — Gespeichert wird, was gerechnet ist

**Datum:** 2026-10-20
**Stand:** angenommen — Weg (a), **brechend mit 0.9.0**, entschieden vom
Auftraggeber im Issue
**Anlass:** #2396
(Folge aus E2315, Phase 1 von 0.9.0)

## Der Befund

MSE2 rechnet die Wertskripte einer Karte auf zwei Wegen:

- **Beim Laden** rechnet `SetScriptManager::updateAll` je Karte jedes Feld
  einmal, in der Reihenfolge der Spielfelder, und zieht nichts nach
  (`src/script/script_manager.cpp`, Zeilen 265–300; `src/data/card.cpp`,
  Zeile 37).
- **Beim Bearbeiten** rechnet `updateValue` das geänderte Feld und zieht nach,
  wer es liest (ebd. Zeilen 249–262 und 309–399); **gesichert wird danach,
  was im Speicher steht**: der gerechnete Wert.

Fassung f9d9356 vom 2021-07-06.

Diese Anwendung speicherte, was eingegeben wurde, und rechnete beim Zeichnen
wie nach einer Bearbeitung. Mit der Ladereihenfolge hätte auf jedem hier
bearbeiteten Planeswalker der ungefilterte Text gestanden (`9fbbf1065`,
zurückgenommen); ohne sie zeichnete eine Karte aus einer fremden Datei anders
als in MSE2 — im Kanon zwölf Karten (Messung vom 2026-10-14).

## Die Entscheidung

1. **Der Schreibweg rechnet** (`Feldweg`, `service/feldwegnachzug.go`,
   `wertnachzug.go`): das gesetzte Feld durch sein Wertskript, dann wer es
   liest, jedes höchstens einmal, Leser nur nach einer Änderung — und
   gespeichert wird das gesetzte Feld mit seinem gerechneten Wert und jedes
   nachgezogene mit `save value`. Das gilt für ein Feld, für mehrere Felder
   einer Karte (Übernahme aus Scryfall, `felder-setzen`), für die Übernahme
   eines KI-Vorschlags und für eine KI-Gruppe über mehrere Karten.
2. **Das Zeichnen rechnet wie MSE2 beim Laden** (`mitWertskripten`):
   Reihenfolge des Spiels, einmal, ohne Nachziehen.
3. **Ein Aufruf ist ein Schritt im Verlauf** (`Kartenschreiber.inEinemSchritt`).
   MSE nimmt eine Bearbeitung als eine Aktion zurück
   (`SetScriptManager::onAction`, `src/script/script_manager.cpp`,
   Zeilen 141–147). Bis dahin trug hier jedes Feld einen eigenen Schritt;
   mit dem Nachzug hätte schon eine einzige Bearbeitung mehrere ⌘Z
   gebraucht, und dazwischen stünde ein halber Stand. **Den Schritt vertritt
   sein erster Eintrag**, und sein Umfang zählt Karten
   (`domain.Verlaufsschritte`): bis dahin der jüngste und die Einträge, und
   die Liste hätte das zuletzt nachgezogene Feld und „2 Karten“ genannt.
4. **An allen Flächen gleich**: Fenster, Kommandozeile, MCP und Dienst setzen
   über denselben Feldweg mit demselben Zeichenwerk.

**Gemessen am 2026-10-20**, Karte für Karte über die feste Kohorte (4.266
Karten) gegen den Stand davor (`f892cfd6b`): **zwölf näher an MSE2, keine
weiter weg** — zehn Planeswalker, Warlock Class und Cleric Class —, Kanonmittel
9,319 statt 9,328. Das sind genau die zwölf aus der Messung vom 2026-10-14.

## Einordnung nach §16 (0033)

**Brechend** an Kommandozeile, MCP und REST (Regel 7): wer ein Feld setzt,
findet danach den gerechneten Wert im Bestand und nicht den übergebenen. Der
Auftraggeber hat Weg (a) gewählt, **mit 0.9.0** und nicht erst mit 1.0 — 0.9.0
ist ohnehin die Fassung, in der die Gleichheit mit MSE2 abgeschlossen wird.
Steht das Spielpaket nicht zur Verfügung, bleibt es beim eingegebenen Wert
(#720).

## Verworfen

- **Weg (b), verträglich** — den Rohwert speichern, beim Schreiben der
  `.mse-set` den gerechneten. Der Bestand hielte einen anderen Wert als die
  Datei, und jede Fläche, die den Bestand liest, sähe den ungerechneten.
- **Nach jeder Eingabe alle Wertskripte rechnen und speichern** — das wäre
  MSEs Laden und nicht sein Bearbeiten: ein Feld, das niemand liest, nähme
  seinen gerechneten Wert an, obwohl MSE es stehen ließe.

## Was das kostet

- **Die Abhängigkeiten stehen im Quelltext und nicht in einem erhobenen
  Graphen** (adapted): `card.<feld>` in einem Wertskript. MSE erhebt
  `dependent_scripts` durch eine abstrakte Auswertung und folgt dabei auch
  Funktionen der Initskripte (`src/script/dependency.cpp`). Ein Feld, das ein
  anderes nur über eine solche Funktion liest, wird hier beim Schreiben nicht
  nachgezogen — gezeichnet wird es trotzdem richtig, weil das Zeichnen jedes
  Feld rechnet.
- **Die Massenänderung rechnet nicht** (new — MSE2 hat keine). Sie schreibt
  **einen** Wert auf viele Karten in einer Transaktion; ein gerechneter Wert
  kann je Karte anders ausfallen, und dafür bräuchte die Ablage einen Weg, der
  je Karte eigene Werte in einen Schritt schreibt. Berechnete Felder
  (`save value: false`) und nicht editierbare weist sie schon ab.
- **Eine neue Karte rechnet nicht** — weder eine angelegte noch eine aus
  einem Kartenskript. MSE rechnet beim Anlegen jeden Wert der neuen Karte
  (`AddCardAction`, ebd. Zeilen 168–178); hier steht, was das Skript liefert
  oder nichts, bis jemand ein Feld bearbeitet.
- **Eine Rücknahme rechnet nicht.** Sie stellt einen gespeicherten Stand her,
  und gespeichert ist seit dieser Entscheidung der gerechnete. MSE rechnet
  auch nach einer Rücknahme (`onAction` ruft `updateValue` unabhängig von
  `undone`); das fällt nur bei einem Skript auf, das auf seinem eigenen
  Ergebnis anders rechnet als auf der Eingabe.

**Gezeichnet wird in allen drei Fällen richtig**, denn das Zeichnen rechnet
jedes Feld. Abweichen kann nur, was im Bestand steht. Die Massenänderung und
die neue Karte stehen in
#2516.

Fassung f9d9356 vom 2021-07-06 für alle Stellen aus MSE2.

## Nachtrag vom 2026-10-20: auf dem Stand der Rechnung

**Das Gesamtreview über 0.9.0 fand zwei Lücken dieses Vermerks** (Befunde 1
und 2, `docs/audits/0.9.0/zweitblicke/9-gesamtreview.md`):

- **Die Bedingung des Fensters** (U02b, #2179) prüfte jedes geschriebene Feld
  gegen die Werte, die das Fenster vom **bearbeiteten** kennt — ein
  nachgezogenes Feld scheiterte als fremde Änderung. Ein nachgezogener Wert
  trägt jetzt `domain.Value.Nachgezogen` und steht nicht unter der
  Bedingung; ein gesetzter bleibt darunter.
- **Gerechnet wurde vor der Transaktion.** Jetzt schreiben Feld, Felder,
  Übernahme und Gruppe auf dem Kartenstand, aus dem gerechnet wurde; ist er
  überholt, kommt `ErrConflict` zurück, und nennt ein Auftrag einen anderen
  Stand als den gerechneten, ebenso. **Gewählt statt neu zu rechnen**: ein
  zweiter Rechenlauf müsste einen überholten Stand von einem fremden
  Konflikt unterscheiden und wäre ein zweiter Weg, der sich verirren kann.
- **Einordnung** nach 0033: Teil des brechenden Wechsels dieses Vermerks —
  an Kommandozeile, MCP und REST kann ein Schreiben jetzt mit einem Konflikt
  enden, wenn dieselbe Karte gleichzeitig geändert wird.

**Was es nicht hält**: Setangaben, die ein Wertskript liest, stehen nicht im
Kartenstand (#2533).
**Seit #2533 tun sie es**: der Stand einer Rechnung ist Karte und Setangaben
(`domain.Rechenstand`), und der Bestand hält beide in der Transaktion.

## Nachtrag vom 2026-10-20: Massenänderung und neue Karte rechnen

**Die zwei Wege aus „Was das kostet“ rechnen jetzt** (#2516):

- **Die Massenänderung** (new) rechnet den Wert auf jeder Karte und zieht
  nach, wer ihn liest (`Feldweg.MassenSetzen`); die Ablage schreibt je Karte
  eigene Werte in **einem** Schritt, mit der Erwartung aus der Vorschau (U6)
  und dem Stand je Karte (`Kartenschreiber.FelderGerechnetSetzen`). **Die
  Vorschau zeigt das Gerechnete**, denn eine, die den eingegebenen Wert zeigt,
  während der gerechnete gespeichert wird, wäre eine Behauptung. Ohne lesbare
  Vorlage bleibt der Rohwert (#720).
- **Eine neue Karte** rechnet jedes Feld mit Wertskript einmal, in der
  Reihenfolge des Spiels — `AddCardAction` in `SetScriptManager::onAction`,
  `src/script/script_manager.cpp`, Zeilen 168–178 —, angelegt wie aus einem
  Kartenskript, und geht mit diesen Werten in **einem** Schritt in den Bestand
  (`Kartenschreiber.KarteAnlegenMitWerten`). **Abweichend (adapted)**: ein
  leeres Ergebnis wird nicht gespeichert, denn der Bestand führt nur Zeilen,
  die etwas sagen; und die übrigen Karten rechnen nicht nach (MSE ruft danach
  `updateAllDependend` über die Kartenliste, Zeilen 180–188) — was davon
  abhängt, rechnet weiter das Zeichnen.
- **Einordnung** nach 0033: Teil des brechenden Wechsels dieses Vermerks.

Damit bleibt aus „Was das kostet“ nur die Rücknahme, die nicht rechnet.
