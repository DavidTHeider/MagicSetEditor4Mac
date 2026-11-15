# 0067 — Ein Profil erbt vom Menschen

**Entschieden am 2026-09-05**, Werkstück 9 von Phase 4.12
(#1351),
Punkt 9 der Kommentierung.

## Der Anlass

Sprache, Erscheinungsbild, Schriftgröße und Anzeigeschrift hingen seit
Wanderung 0035 am **Profil**. Wer zwei Profile führt und beide auf Englisch
haben will, sagt das zweimal — und beim dritten noch einmal.

Die Kommentierung nennt es Punkt 9: *„Override-Einstellungen für das jeweilige
Profil."* Das Wort **Override** sagt die Richtung: es gibt etwas darunter, das
übersteuert wird.

## Was entschieden ist

**Der Mensch hält die Vorgabe, das Profil sticht sie.** Drei Stufen, und die
Reihenfolge ist die Aussage:

1. was das **Profil** gewählt hat,
2. sonst was der **Mensch** gewählt hat,
3. sonst die Vorgabe.

**`erben` ist die Vorgabe des Profils**, nicht ein Wert unter anderen. Ein
frisches Profil soll nichts über den Menschen behaupten; stünde dort `system`,
hätte es eine Wahl getroffen, die niemand getroffen hat.

**Vier Schlüssel, und nur diese vier**: `sprache`, `erscheinungsbild`,
`schriftgroesse`, `schriftart`. Sie beschreiben, wie die Anwendung **aussieht
und spricht**, und das ist eine Eigenschaft des Menschen. `aktualisierung`
betrifft den **Rechner** (geht etwas ins Netz?), `bildkonflikt` einen
**Ausgabelauf**, `ersteschritte` ein Fenster, das einmal aufgeht — keine wird
dadurch besser, dass sie zwei Ebenen hat.

## Was das kostet

**Jede Antwort trägt jetzt zwei Werte.** `wert` ist die Wahl, `wirksam` ist,
was gilt. Wer nur eines von beiden hätte, zeigte entweder eine Wahl, die
niemand getroffen hat, oder eine Anwendung, die anders aussieht, als die Liste
sagt. Das ist ein Feld mehr an vier Flächen — **erweiternd** (§16), denn es
kommt hinzu und nimmt nichts weg.

**Und ein Lesen mehr je Antwort.** `Einstellungsverwalter.Stand` liest die
Profilzeile **und** die Zeile des Menschen. Zwei Abfragen statt einer; gemessen
wurde nicht, und deshalb steht hier keine Zahl.

## Was ein bestehender Bestand merkt

**Nichts.** Seine Profile tragen **gesetzte** Werte, und gesetzt sticht. Die
Vorgabe `erben` gilt nur, wo nie jemand gewählt hat — und dort galt vorher
dieselbe Vorgabe, nur eine Ebene tiefer. Der Lauf, der das hält, ist
`TestOhneWahlGeltenDieVorgabewerte` in `internal/service`: er stand vor dieser
Änderung da, prüft seither denselben Satz **durch die Verschmelzung hindurch**
und ist unverändert grün.

## Was verworfen wurde

**Die Verschmelzung im Fenster.** Sie wäre billiger gewesen — kein zweites
Feld, kein zweites Lesen in Go. Aber dann hielte sie kein Go-Test, und
Kommandozeile, MCP und Dienst bekämen eine andere Antwort als das Fenster. Das
ist derselbe Grund, aus dem schon die **Vorgabewerte** in `internal/domain`
stehen und nicht in `App.tsx` (#245).

**Ein `erben` auch beim Menschen.** Er ist die unterste Ebene; wer dort „wie
der Nutzer" wählen könnte, zeigte auf sich selbst, und die Frage hätte keine
Antwort mehr. Ein Wächter hält das fest.
