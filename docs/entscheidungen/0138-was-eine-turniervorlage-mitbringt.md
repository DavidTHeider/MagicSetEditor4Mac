# 0138 — Was eine Turniervorlage mitbringt

**Stand:** entschieden am 2026-09-27 · DOM-04.AK-01 und AK-03
(#1699)

## Die Frage

DOM-04 verlangt **wiederverwendbare Vorlagen** (AK-01) und eine **Kopie
einer Turniervorlage**, die *„keine fremden Spielerzustimmungen oder alten
Transaktionskennungen"* übernimmt (AK-03). Was genau ist an einem Turnier
Vorlage, und was ist Verlauf?

## Die Entscheidung

**Eine Vorlage ist ein vorhandenes Turnier, und die Kopie ist ein neues.**
Es gibt keine eigene Art „Vorlage" neben dem Turnier: wer den Freitagabend
jede Woche spielt, nimmt den von letzter Woche.

| Kommt mit | Kommt **nicht** mit | Warum |
| --- | --- | --- |
| Spiel, Format, Einzelspiel, Notiz | — | das sind die Regeln des Abends |
| auf Wunsch die **Namen** der Teilnehmenden | ihre **Decks** und **eingereichten Listen** | eine Liste ist die Zusage eines Menschen für **ein** Turnier (0062, Entscheidung 3). Sie in ein neues zu tragen, meldete jemanden mit einer Liste an, die er dort nie eingereicht hat — das ist die *fremde Zustimmung* aus AK-03 |
| — | Runden, Paarungen, Ergebnisse, gebundene Partien | Verlauf (0062, Entscheidung 1); eine Paarung trägt Kennungen des alten Turniers und seiner Partien — die *alten Transaktionskennungen* aus AK-03 |
| — | wer ausgestiegen ist | ein Ausstieg ist ein Ereignis des alten Abends |

**Durch dasselbe Anlegen wie ein neues Turnier** (`Turnierleitung.Kopieren`
ruft `Anlegen`): die Kopie durchläuft dieselben Prüfungen, und ein Format,
das inzwischen zu einem anderen Spiel gehörte, käme nicht still mit. **Die
Namen entstehen im selben Zug wie das Turnier** — scheitert einer, gibt es
kein halbes Turnier.

**Ein Turnier trägt seitdem ein Format** (Wanderung 0088) — eine Angabe
und kein Zwang, wie am Deck. Dass es zum Spiel des Turniers gehört, hält
der Dienst (`Turnierleitung.formatPruefen`), denn die Tafel kennt die
Spiele nicht.

## Verworfen

| Verworfen | Weil |
| --- | --- |
| **Eine eigene Tafel `turniervorlagen`** | Sie hielte dieselben vier Felder wie `turniere` und brauchte eigene Wege an vier Flächen — für eine Unterscheidung, die niemand trifft: das Turnier von letzter Woche *ist* die Vorlage |
| **Die Decks mitkopieren, „weil es bequem ist"** | Siehe oben: eine Liste ist eingereicht. Wer mit demselben Deck wieder antritt, meldet es im neuen Turnier an — ein Klick, und der Abzug zeigt dann den **heutigen** Stand des Decks |
| **Der Assistent sammelt alles und legt am Ende an** | Teilnehmende und Runde sind Vorgänge an einem Turnier, das es gibt. Ein Assistent, der bis zum Schluss sammelte, schriebe am Ende vier Vorgänge hintereinander — und hinterließe bei einem Fehler im letzten einen halben Stand |
