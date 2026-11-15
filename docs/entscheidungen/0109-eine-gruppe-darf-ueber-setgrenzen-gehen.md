# 0109 — Eine Gruppenübernahme darf über Setgrenzen gehen

**Datum:** 2026-09-13 · **Stand:** entschieden ·
**Schreibt fort:** [0064](0064-ein-feld-auf-vielen-karten-darf-ueber-mcp.md) ·
**Anlass:** Werkstück A13.1b von Phase 5.3
(#1452)

## Die Frage, und was schon entschieden ist

**Darf `schreiben.ki.gruppe` mehrere vorhandene Karten in einem Ruf ändern?**

**Das ist entschieden, und zwar in
[0064](0064-ein-feld-auf-vielen-karten-darf-ueber-mcp.md)**: „ein Feld,
mehrere **benannte** Karten, ein Set, eine Transaktion." Der Maßstab dort ist
nicht die Zahl, sondern die Herkunft des Umfangs — *„Der Umfang dieses
Werkzeugs steht in seinen Argumenten."*

Die Gruppenübernahme löst das ein. Jede Kennung in ihrer `wahl` ist ein
**Auftrag**, und ein Auftrag trägt seine Zielkarte samt eingefrorener Revision
seit dem Augenblick, in dem ihn jemand angelegt hat. Sie kann keine Karte
treffen, die nicht vor dem Ruf feststand.

**Der unbedingte Satz aus [0044](0044-was-ein-werkzeug-schreiben-darf.md)** —
„kein Werkzeug ändert mehrere vorhandene Karten auf einmal" — **war schon vor
diesem Vermerk überholt**; 0064 hat ihn ersetzt. Wer ihn heute liest, liest
eine Zusage, die seit dem 2026-09-02 nicht mehr gilt.

## Was hier wirklich neu ist

**Zwei Dinge, und nur das zweite braucht eine Entscheidung.**

| | `schreiben.felder` (0064) | `schreiben.ki.gruppe` |
| --- | --- | --- |
| Felder je Karte | **eines**, für alle dasselbe | **mehrere**, je Karte andere |
| Sets | **eines** | **mehrere** möglich |

Das erste ist kein neuer Maßstab: der Umfang steht weiterhin in den
Argumenten, nur nennt jeder Eintrag zusätzlich seine Felder. Wer eine Karte
und zwei Felder nennt, ändert zwei Felder.

**Das zweite ist einer.** 0064 sagt „ein Set", und der Grund steht dort:

> eine aus einem **anderen** Set lässt den ganzen Aufruf scheitern — die
> Aufsicht fragt für jede einzeln

## Was gilt

**Eine Gruppenübernahme darf Karten aus mehreren Sets umfassen.**

Der Grund, aus dem 0064 auf ein Set beschränkt, trägt hier nicht: dort nennt
der **Aufrufer** eine Setkennung und eine Kartenliste, und eine Karte, die
nicht dazu passt, wäre ein Vertippen mit Wirkung. Hier nennt der Aufrufer
**Aufträge**, und zu welchem Set eine Zielkarte gehört, liest das Programm an
der Karte selbst ([app.kikarten.FelderGruppeSetzen]). Es gibt keine Angabe,
die falsch sein könnte.

**Die Aufsicht bleibt und fragt weiter für jede Karte einzeln** — vor dem
ersten Schreiben, nicht mittendrin. Was sie prüft, ist unverändert: gehört
diese Karte zu diesem Set.

## Was das kostet

**Eine Rücknahme trifft die Gruppe nicht als Gruppe.** `schreiben.zuruecknehmen`
bekommt eine **Setkennung** und sucht den jüngsten geltenden Schritt (0064,
`neuesterSchritt`). Eine Gruppe über drei Sets hat drei solche Schritte, und
wer sie zurücknehmen will, ruft dreimal. **Vorwärts ist sie atomar, rückwärts
nicht** — und das steht hier, weil es sonst niemand merkt.

Der schwerere Weg — ein Verlaufsschritt über Setgrenzen — ist nicht gebaut. Er
hinge an `neuesterSchritt`, und der ist die Rücknahme für **jeden** Schreibweg
dieses Programms; ihn für diesen einen Fall umzubauen wäre ein Eingriff an der
Stelle mit den meisten Aufrufern.

## Wie man das Gegenteil erkennt

**An einem Argument, das eine Menge beschreibt statt sie aufzuzählen** — ein
Filter, ein Suchausdruck, ein „alle mit". Sobald ein schreibendes Werkzeug so
etwas nimmt, fällt der Maßstab aus 0064, und zwar unabhängig von der Zahl der
getroffenen Karten.

## Verworfen

**„Dann eben nur ein Set je Gruppe."** Eine Grenze ohne Grund: die Setkennung
kommt hier nicht vom Aufrufer, also kann sie nicht falsch sein. Sie kostete
einen Ruf je Set und brächte keine Zusage ein.

**„Dann eben nicht über MCP."** Das wäre §16 zuwider und in der Sache
schlechter: das Modell übernähme weiter einzeln — drei Rufe, drei
Transaktionen — und genau der Teilbestand, den A13.1b verhindert, entstünde
wieder.
