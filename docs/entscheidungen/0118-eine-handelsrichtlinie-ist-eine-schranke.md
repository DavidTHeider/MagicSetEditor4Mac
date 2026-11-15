# 0118 — Eine Handelsrichtlinie ist eine Schranke, keine neue Gestalt

**Stand:** angenommen · **Datum:** 2026-09-18
**Berührt:** [0079](0079-einstellungsebenen-und-portabler-import.md),
[0117](0117-was-eine-handelsfassung-bindet.md)

## Warum es diesen Vermerk gibt

Weil Werkstück `T1` von Phase 5.6
(#1466) den
Begriff **Richtlinie** in seiner Begriffsliste nennt — und ihn nirgends
erklärt. Er kommt in keinem anderen Issue der Phase vor; gemessen am
2026-09-18 über alle dreizehn.

**Ein Begriff ohne Definition wird beim Bauen erfunden**, und zwar von dem,
der zuerst daran kommt. Dieser Vermerk erfindet ihn nicht, sondern findet ihn:
**es gibt ihn schon.**

## Die Definition, und sie stammt nicht von hier

`internal/domain/richtlinie.go`, Werkstück `O3` von Phase 5.2
(#1427):

> **Eine Richtlinie ist eine Schranke und keine Vorliebe.** Bei „welche
> Schriftgröße" hat der Mensch eine Vorgabe und das Profil das letzte Wort —
> es geht um Geschmack. Bei „darf dieses Programm ins Netz" wäre dieselbe
> Regel ein Loch: wer als Mensch „nein" gesagt hat, sagt es für **jedes**
> seiner Profile.
>
> **Deshalb wird nicht gestochen, sondern gefaltet**: der strengste Wert aller
> geltenden Ebenen gewinnt, und die Reihenfolge der Ebenen spielt keine Rolle.

**Damit ist die Frage beantwortet, und zwar ohne neuen Mechanismus**: eine
Handelsrichtlinie ist eine Schranke über dem Handel, gefaltet wie
`richtlinie.nurlokal`, eingetragen in **beide** Reihen, auf der Ebene des
**Menschen**.

## Warum sie sich nicht als Liste ausdrücken lässt

**Weil eine Schranke sich falten muss.** Die Faltung kennt zwei Arten:
`reihung` (aufzählbare Werte von streng nach locker) und `obergrenze` (die
kleinere Zahl ist strenger). Eine Liste „welche Vertragsarten sind erlaubt"
lässt sich **nicht** so ordnen: ist `{tausch}` strenger als `{teilen}`? Die
Frage hat keine Antwort, und eine erfundene wäre schlimmer als keine.

**Deshalb fünf Schlüssel und keine Liste:**

| Schlüssel | Art | streng ist … |
| --- | --- | --- |
| `richtlinie.handel` | Reihung | `aus` |
| `richtlinie.handel.markt` | Reihung | `aus` |
| `richtlinie.handel.menge` | Obergrenze je Position | die kleinere Zahl |
| `richtlinie.handel.wert` | Obergrenze je Angebot | die kleinere Zahl |
| `richtlinie.handel.frist` | Obergrenze in **Minuten** | die kleinere Zahl |

**Die Frist steht in Minuten und nicht als Dauer**: `kleinsteGrenze`
vergleicht Zahlen. `1h30m` ließe sich nicht der Größe nach ordnen, ohne es
vorher zu deuten — und eine Faltung, die ihre Werte erst deuten muss, ist
keine.

## Drei Fragen, drei Stellen — und das ist der eigentliche Punkt

**Die Schranke entscheidet ausdrücklich nicht, wer handeln darf.** Das war
schon entschieden, und zwar in Phase N:

| Frage | Wer sie beantwortet | Seit |
| --- | --- | --- |
| **Darf dieser Peer überhaupt handeln?** | `Peerkontakt.Darf(ZweckHandel)` — **bestätigter** Kontakt **und** erteiltes Recht; gesperrt bleibt gesperrt | `N1`, #1454 |
| **Was bekommt er zu sehen?** | eine Freigabe je Empfänger und Umfang | `T3`, #1468 |
| **Was lässt dieses Profil überhaupt zu?** | **die Richtlinie** — unabhängig davon, wer fragt | hier |

**Ohne die dritte wäre jede Grenze eine Eigenschaft des Gegenübers.** „Ich
handle grundsätzlich keine Stapel über vier Stück" ist aber keine Aussage über
einen Partner, sondern über einen selbst; wer sie an jedem Kontakt einzeln
hinterlegte, hätte sie beim nächsten vergessen.

**Und sie setzt auf derselben Infrastruktur auf wie das Spielen.** `ZweckHandel`
steht seit Phase 5.4 neben `ZweckPartie` in derselben Aufzählung, der
Handschlag ordnet beiden ihr Protokoll zu, und die Leitung darunter ist
dieselbe — TCP mit TLS 1.3 im eigenen Netz, tailcat darüber hinaus. **Der
Handel bekommt keine zweite Netzschicht, und er bekommt auch keine zweite
Schrankenschicht.**

## Die Vorgabe: handeln ja, Markt nein

**Der Markt greift auf einen fremden Anbieter und auf Geld.** Die Phase nennt
ihn ein „optionales externes Schreibpaket, kein stiller Checkout"
(#1410); eine
Vorgabe, die ihn einschlösse, wäre genau der stille Checkout.

**Die anderen drei Arten bleiben unter zwei Menschen** — Teilen gibt eine
Sicht frei, die Deckleihe reicht einen eingefrorenen Stand weiter, der Tausch
bewegt Papier von Hand. Sie ohne Not abzuschalten hieße, ein Programm
auszuliefern, das seine eigene Phase nicht kann.

## Was verworfen wurde

**Eine eigene Tabelle je Profil.** Sie hätte einen zweiten Weg zu allen vier
Flächen gebraucht — und die Einstellungen haben längst einen, geprüft in
`domain`, abgelegt in `storage`, gesetzt über `EinstellungSetzen`.

**Eine Liste erlaubter Vertragsarten.** Sie lässt sich nicht falten; der Grund
steht oben.

**Eine Schranke je Kontakt.** Das ist die erste Frage der Tabelle, und sie hat
ihre Antwort seit `N1`. Zwei Stellen für dieselbe Frage wären die
Verwechslung, die man später nicht mehr findet.
