# 0031 — Sieben Phasen bis zur Freigabe 0.1, und die Parität wird nicht abgewartet

**Stand:** angenommen
**Datum:** 2026-08-20
**Ergänzt:** [0020](0020-acht-phasen-fuer-die-paritaet.md) — dort wurden acht
Phasen für die Parität geschnitten; diese sieben schneiden den Weg zur ersten
Fassung
**Berührt:** [0010](0010-mcp-als-eigener-prozess.md) — Phase M kann ihn
fortschreiben oder bestätigen; entschieden wird das **dort**

## Lage

Zehn Phasen sind seit [0020](0020-acht-phasen-fuer-die-paritaet.md)
abgeschlossen — A bis J. Der Abstand zu MSE2 ist von 26,714 auf **21,581**
gefallen (573 Karten aus vier Sätzen, 2026-08-19), die Oberfläche ist benutzbar,
und die Anwendung öffnet, bearbeitet und schreibt Sets.

**Und sie ist nicht freigegeben.** `docs/ROADMAP.md` führt dazu seit Langem
denselben Satz: die Freigabe *findet noch nicht statt*, weil eine Karte aus
einem bezogenen Paket beim Dreiunddreißigfachen der Toleranz neben MSE2 liegt.

Damit steht die Freigabe an einer Bedingung, die **niemand datiert hat**. Der
Abstand fällt seit Phase A monoton, aber er fällt in immer kleineren Schritten:
Phase J hat 16,1 % gebracht, und der größte verbliebene Einzelposten — der Satz
— trägt 50,94 % von dem, was übrig ist. Es gibt keinen Stand, an dem „Parität"
eintritt und jemand es merkt.

**Zugleich fehlt der Anwendung Grundlegendes**, das mit dem Abstand nichts zu
tun hat: eine Karte lässt sich nicht löschen, nicht kopieren, nicht umordnen
(#562 hat das
Löschen ausdrücklich vertagt); die Hilfe führt **einen** Eintrag bei elf Seiten;
protokolliert wird an **einer** Stelle; SVG gibt es nicht.

Das ist die Lehre aus
#248, und §10
hat sie seit dem 2026-08-16 als **erste** Reviewdimension: ein Review, das nur
das Handwerk prüft, bestätigt eine Anwendung, die niemand benutzen kann.

## Entscheidung

**Sieben Phasen, K bis Q**, geführt in [`../ROADMAP.md`](../ROADMAP.md),
Abschnitt „Ziel: Freigabe 0.1 — die Phasen K bis Q". Buchstaben weiter, aus
demselben Grund wie in 0020.

| Phase | Sammelissue |
| --- | --- |
| K — Offen bleibt nur, was besser werden soll | #589 |
| L — Die Karte, ganz | #590 |
| M — Schnittstellen und ihre Absicherung | #591 |
| N — Was der Bestand aushalten muss | #592 |
| O — Zugang, Sprache, Hilfe | #593 |
| P — Aus- und Einfuhr | #594 |
| Q — Protokoll, Doku, Freigabe | #595 |

**Das Ziel ist ein Zustand, keine Zahl.** Nicht mehr „der Abstand unterschreitet
*x*", sondern: ein sauberer darwin/arm64-Bau, ad hoc signiert, als **0.1**
veröffentlicht — und eine Liste, in der nur noch steht, was besser werden soll,
nicht was fehlt.

**Die Parität bleibt das Ziel des Projekts.** Sie ist nur nicht mehr die
Bedingung der ersten Fassung. Der Abstand wird weiter gemessen und in der
Roadmap geführt; 0.1 sagt, dass die Anwendung benutzbar und weitergebbar ist,
nicht dass sie MSE2 trifft.

## Verworfen

**Erst Parität, dann freigeben.** Das war der Stand bis heute, und es ist die
naheliegende Lesart von
#280. Dagegen
steht, dass die Bedingung kein Eintrittsdatum hat: „Parität" ist erst dann
erfüllt, wenn eine Toleranz benannt und unterschritten ist, und **welche
Toleranz das wäre, ist nie festgelegt worden**. Eine Freigabe hinter einer
unbenannten Schwelle findet nicht statt.

**Eine Zwischenfassung je Phase.** [§12](../../AGENTS.md) sagt seit Phase 1,
warum nicht: eine Fassung, die nichts kann, ist keine Fassung, und eine
Versionsnummer je Phase erzeugt eine Historie, die niemand braucht. Daran
ändert sich nichts — 0.1 kommt in **einer** Phase, der letzten.

**Die drei Fragen aus M und N vorab entscheiden**, statt sie in ihre Phase zu
legen. Verworfen aus dem Grund, der schon in
#289 stand:
solche Entscheidungen altern schlecht, wenn man sie trifft, bevor man weiß, was
die Anwendung tut. Sie stehen deshalb als **Einwand** im Roadmap-Abschnitt —
sichtbar, undatiert, unentschieden — und werden in ihrer Phase als Vermerk
entschieden.

**Phase K weglassen und gleich bauen.** Vier der zwölf offenen Issues verlangen
eine Messung, deren Ergebnis den Zuschnitt von L bis Q ändern kann; zwei davon
(#458,
#446) betreffen
die **Referenz selbst**. Wer gegen eine fehlerhafte Referenz angleicht, macht
die Ausgabe schlechter und die Zahl besser.

## Folgen

- **Die Freigabe bekommt ein Datum, das an Arbeit hängt statt an einer
  Schwelle** — Phase Q, nach K bis P.
- **Der Abstand hört nicht auf, gemessen zu werden.** Die Zeile „Der Maßstab"
  in der Roadmap bleibt, und Phase K schärft sie um das, was die Referenz
  selbst beiträgt.
- **Phase M kann entfallen.** Ihr erster Schritt ist ein Vermerk über einen
  Dienst, den „Was ausdrücklich nicht kommt" heute ausschließt; fällt er gegen
  den Dienst, ist die Phase fertig. Das ist ein gültiger Ausgang und braucht
  dieselbe Deckung wie ein Ja (§10).
- **Nach Q steht das Gesamtreview** (§10) — funktionale Abdeckung zuerst, dann
  Deckung durch die drei Quellen, Nähte, Zusagen, Fehlerwege, Nebenläufigkeit,
  Dokumentation. Es ist kein achter Zweitblick, sondern der eine, den §10 ans
  Ende **aller** Phasen stellt.
- **Was 0.1 nicht behauptet**, gehört in die Release-Notiz und steht schon in
  [0007](0007-kein-entfernter-lauf.md): geprüft ist eine Maschine, nicht eine
  Plattform.
