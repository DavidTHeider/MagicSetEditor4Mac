# 0123 — `streng` zählt nur, was zur Variante gehört

**Datum:** 2026-09-19 ·
**Anlass:** #1626,
gefunden beim Phasenabschluss 5.5 (§13, Schritt 3) · **Gilt ab:**
`Grundprofilfassung` 22

## Der Befund

**`TraegtStrenge(StrengeStreng)` konnte für das Grundprofil nie wahr werden** —
gleich wie viel gebaut wird. Die Bedingung lautete „keine Lücke", und
`Luecken()` ging über **jeden** Bereich der Matrix. Einer davon ist
`commander`: die Variante aus CR 903, und im **Grundprofil** steht sie auf
`manuell`, weil eine Sandkastenpartie keine Commanderpartie ist.

**Damit hing eine Zusage an einer Bedingung, die sie ausschloss.** Gemessen am
2026-09-18: dreizehn Bereiche, zehn davon `automatisch`, `Luecken()` nennt
`spielsteine`, `typen` und `commander`. Die ersten beiden sind Arbeit. Der
dritte ist eine Kategorie: er gehört einer Variante, die diese Partie nicht
spielt.

## Die Entscheidung

**`Luecken()` und `TraegtStrenge` zählen nur die Bereiche, die zum Profil
gehören.** Ein Matrixeintrag trägt dafür ein Merkmal — [Matrixeintrag.NurVariante]
—, und ein Bereich mit diesem Merkmal zählt nur in einem Profil, das die
Variante **spielt**.

**Wie ein Profil das sagt**: es führt den Bereich auf `automatisch` oder
`begleitet`. Das Commanderprofil tut genau das (`Commanderfassung` 1, vier
Grenzen); das Grundprofil führt ihn auf `manuell` — und damit ist er dort
**keine Lücke**, sondern eine Nichtzuständigkeit.

## Was verworfen wurde

**„Jeder Bereich zählt, und `streng` ist eine Aussage über ein Format."** Das
war die zweite vertretbare Lesart, und sie hat einen Vorzug: sie braucht kein
neues Merkmal. Sie ist verworfen, weil sie die Auskunft **falsch** macht statt
nur streng: eine Sandkastenpartie, die `commander` als offene Lücke führt,
sagt einem Menschen, hier fehle Arbeit — und es fehlt keine. Eine Lücke, die
niemand schließen soll, ist keine.

**Und sie hätte die Zusage nicht gerettet**: mit ihr bliebe `streng` für das
Grundprofil dauerhaft unerreichbar, und eine Stufe, die nie erreicht werden
kann, ist keine Stufe.

## Was dabei nicht entschieden ist

**Ob `spielsteine` und `typen` je `automatisch` werden.** Sie bleiben
`begleitet`, und ihre Gründe stehen in der Matrix: neun Grenzen bei den
Spielsteinen (CR 111, CR 701.7a und die Kartengestalten aus CR 309, 714,
716–721), vier bei den Typen (CR 205.3b–m, CR 205.2b/c, CR 310 mit
vierundzwanzig ungenannten Regeln, und Schicht 4 aus CR 613). Das ist Arbeit
und keine Entscheidung — sie hat ihren Platz in der Regelabdeckung.

**Damit trägt das Grundprofil auch nach dieser Entscheidung kein `streng`** —
aber jetzt aus einem Grund, den jemand schließen kann.
