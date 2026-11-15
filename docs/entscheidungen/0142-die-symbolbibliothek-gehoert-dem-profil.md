# 0142 — Die Symbolbibliothek gehört dem Profil, ein Set trägt eine Kopie

**Stand:** entschieden am 2026-09-27 · DOM-02.AK-01, AK-02 und AK-04
(#1697)

## Die Frage

DOM-02 verlangt eine Bibliothek mit „Vorschau, Suche, Tags, Kopie,
Import/Export sowie klare Unterscheidung zwischen Setsymbol und
Symbolschrift" (AK-01), Verwendungsnachweise und die Folgen **vor** dem
Löschen (AK-02) und Dateien, die begrenzt und verständlich abgewiesen werden,
statt die Anwendung zu blockieren (AK-04). Offen war, **wem** ein Symbol
gehört, **wie** ein Set es trägt und **woran** man sieht, welches Set welches
trägt.

## Was MSE2 tut

**Es hat keine Bibliothek.** Gesucht in `src/gui/symbol/` (Editor,
Werkzeugleisten, Teilliste) und `src/data/symbol.cpp` (Fassung f9d9356 vom
2021-07-06): ein Symbol ist die `.mse-symbol`-Datei **eines** Sets, der
Symboleditor öffnet und schreibt genau sie, und ein zweites Set bekommt sein
Symbol nur, indem jemand die Datei von Hand kopiert. Die Portierungsmarke der
Bibliothek ist deshalb **new**; die Grenzen und die Vorschau stützen sich
auf portiertes Verhalten (unten).

## Entschieden

1. **Die Bibliothek gehört dem Profil.** Eine eigene Tafel
   `symbolbibliothek` (Migration 0091) mit Profilkennung in jeder Frage; ein
   Symbol eines anderen Profils ist eines, das es nicht gibt. Ein Name steht
   **einmal** je Profil, ohne Rücksicht auf Groß und Klein, und ein doppelter
   kommt am Namensfeld an.

2. **Ein Set trägt eine Kopie, keinen Verweis.** Einsetzen legt die
   Zeichnung als neue `.mse-symbol` **neben das Set** — dieselbe Ablage wie
   ein gezeichnetes Symbol — und setzt sie als dessen Symbol. **Erst die
   Datei, dann der Wert**: ein Wert ohne Datei zeichnete das Set ohne Symbol,
   eine Datei ohne Wert ist nur eine Datei mehr. Ein Verweis in die
   Bibliothek wäre in einer weitergegebenen `.mse-set` ins Leere gegangen.

3. **Welches Set welches Symbol trägt, sagt der Abdruck.** Die Prüfsumme der
   Zeichnung **in ihrer geschriebenen Form** — zwei gleiche Zeichnungen sind
   dasselbe Symbol, gleich woher sie kamen. Die Liste zählt die Träger bei
   jedem Lesen über die Sets des Profils; ein Set, dessen Datei fehlt oder
   sich nicht lesen lässt, trägt keines. **Kein Verwendungszähler im
   Bestand**: er liefe mit jeder Handänderung an einer Datei auseinander.

4. **Löschen mit Folgen, und die Sets behalten ihre Kopie.** Ein getragenes
   Symbol geht nur bestätigt; die Absage (`symbolVerwendet`, ein Konflikt)
   nennt die Sets. Verloren geht der Eintrag der Bibliothek und mit ihm der
   Nachweis, dass diese Sets dasselbe Symbol tragen — **kein Symbol eines
   Sets**. Das Fenster fragt mit einer Entscheidung, die dieselben Sets
   nennt; die übrigen Flächen verlangen `bestaetigt`.

5. **Grenzen mit eigenen Schlüsseln** (AK-04): 1 MiB je Datei, 4.096 Teile,
   65.536 Punkte, 32 Ebenen — gemessen am installierten Pack und mit
   Spielraum gesetzt (`msepaket/symbolgrenzen.go`). Jede Grenze hat einen
   eigenen Schlüssel samt Zahl und Grenze, und **dieselben Grenzen gelten
   beim Ablegen einer gezeichneten Zeichnung**.

6. **Symbolschriften stehen daneben, nur zur Ansicht.** Sie sind Pakete,
   gehören dem Rechner und werden unter *Pakete* eingebunden; die Bibliothek
   führt sie nicht.

7. **Marken über die Art `symbol`** (UX-12, Vermerk 0135) — eine neunte
   Markenart; ein Auslöser nimmt die Marken eines gelöschten Symbols mit.

8. **An jeder Fläche, und Dateien nur, wo es eine Platte gibt.** Fenster und
   Kommandozeile lesen und schreiben `.mse-symbol`-Dateien; MCP und REST
   reichen den **Text** (Vermerk 0140). Das Vorschaubild — schwarz mit weißem
   Rand, Randbreite 0,05 wie MSEs `SymbolVariation` — steht nur am Fenster.

## Verworfen

- **Ein Verweis vom Set in die Bibliothek.** Er bräche jedes weitergegebene
  Set und jede Sicherung ohne Bestand.
- **Die Bibliothek als Paket** (`.mse-include` o. ä.). Pakete gehören dem
  Rechner und gehen über den Paketverwalter mit Plan und Fassung; ein Symbol
  ist ein Arbeitsstück eines Profils und wird wie ein Deck geführt.
- **Ein gespeicherter Verwendungszähler** — siehe Punkt 3.

## Der zweite Zug (2026-09-27)

9. **Eine Fassung zurück und nicht eine Geschichte.** Sichern legt die
   bisherige Zeichnung in `vorher`, Zurücknehmen tauscht beide. Eine längere
   Geschichte hätte eine eigene Tabelle gebraucht, und AK-02 verlangt die
   Rücknahme **einer** Änderung.

10. **Die Sets werden genannt und nicht still umgeschrieben.** Nach dem
    Sichern tragen sie die **vorige** Fassung — der Abdruck sagt, welche.
    Erst *In diese Sets übernehmen* schreibt ihre Kopien neu, und zwar
    **unter demselben Namen**. MSE2 legt beim Sichern eine **neue** Datei an
    und setzt den Wert um (`SymbolWindow::onFileStore`,
    `src/gui/symbol/window.cpp` Zeilen 249–258, Fassung f9d9356 vom
    2021-07-06); hier gilt, was schon das Ablegen tut (#890): der Wert bleibt,
    die Datei wird ersetzt, und die Rücknahme führt die Bibliothek.

11. **Neu gezeichnet wird, was das Symbol trägt, und nur das** (AK-03). Der
    Plan einer Karte trägt die Teile des Symbols selbst, und sein Abdruck
    geht in den Bildschlüssel ein — eine ersetzte Datei ändert genau die
    Schlüssel der Karten dieses Sets. Belegt in
    `service/symbolneuzeichnen_intern_test.go`, samt Gegenprobe.

12. **Die Fassungen zeigt nur das Fenster.** Drei Bilder je Fassung liest in
    einer Textantwort niemand; die Teile gibt jede Fläche.

13. **Ein Riegel vor jeder Datei** (0078). Einsetzen, Nachziehen und
    Ausgeben schreiben Dateien, bevor oder ohne dass sich ein Wert im
    Bestand ändert — die lesende Verbindung hielte sie nicht auf.

## Was offen bleibt

**Die Handabnahme am nativen Fenster** — sie steht mit den anderen in
#1871.
