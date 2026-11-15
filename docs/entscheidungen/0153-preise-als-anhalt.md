# 0153 — Preise als Anhalt: zwei Spalten, eine Regel, keine Umrechnung

**Stand:** entschieden am 2026-09-30 · TRADE-02b
(#1951)

## Die Frage

Das Artboard *Handelsassistent* zeigt in *Was tauscht ihr* zwei Spalten,
*Von dir* und *Von Turfer*. Jede Zeile trägt einen Preis, jede Seite eine
Summe, und darunter steht die Differenz: „1,95 € zu deinen Gunsten“. Gebaut
war eine Liste mit der Auswahl „Von wem“ ([0151](0151-der-handel-in-vier-phasen.md)).

Die Sichtprüfung zu TRADE-03 fand dazu, dass ein gewählter Handel seine
Zeilen **gar nicht** zeigte. Nach der Einigung wartete ein leeres Formular,
und ein Gegenangebot begann mit null Zeilen.

## Was MSE2 tut

**Es handelt nicht** ([0149](0149-was-eine-handelsnachricht-behaupten-darf.md)).
Marke **new**.

## Was gilt

1. **Zwei Spalten nach Seite** (AK-01). Jede Spalte hat ihren eigenen Knopf
   „+ Zeile“, der das Formular **in dieser Spalte** öffnet. Die Seite ist
   damit keine Auswahl mehr. Ohne Ausweis bleibt die eigene Spalte zu, sonst
   ginge eine Zeile mit leerer Seite hinaus.
2. **Die Zeilen des Standes stehen da.** Nach der Einigung stehen sie nur
   zum Lesen da, ohne Formular und ohne Entfernen. Ein Gegenangebot beginnt
   beim letzten Angebot. Wer schon geändert hat, behält seine Zeilen, bis
   sein Schreiben glückt; danach steht der neue Stand da.
3. **Je Zeile Preis mal Menge, je Seite eine Summe, die Differenz mit
   Richtung** (AK-02). Die Differenz ist *erhalten weniger gegeben* aus
   eigener Sicht und wird **in Worten** genannt („zu deinen Gunsten“,
   „zugunsten deines Gegenübers“, „ausgeglichen“), nicht mit Vorzeichen.
   Eine Zeile ohne Preis heißt „kein Preis“ und nicht 0. Fehlt einer Zeile
   der Preis, sagt die Differenz, dass sie nur über einen Teil rechnet.
4. **Dieselbe Regel wie das Portfolio**:
   - genannt, Papier, Verkauf;
   - kein `tix`;
   - veredelt zu veredelt (`jeWaehrungEiner`).

   Eine zweite Regel für dieselbe Frage liefe beim ersten neuen Markt von
   der ersten weg.
5. **Keine Umrechnung.** Summen und Differenz stehen je Währung. Zuerst
   gezeigt wird die **Vorgabe** (`vorgabe.waehrung`), wo eine Zeile sie
   nennt, sonst die Währung mit den meisten Zeilen.
6. **Ein Anhalt und keine Vorgabe** (AK-03). Die Preise werden bei jedem
   Lesen aus dem eigenen Preisbestand gerechnet und nie gespeichert. Sie
   gehen nicht in den Vertrag, nicht in den Abdruck und nicht über die
   Leitung. Die Rechnung steht deshalb in `service/preisanhalt.go` und
   nicht bei den `handels*.go`, denn deren Wächter verbietet dem Vertrag
   jedes Wort über Preise. Der Satz unter der Differenz nennt Datum und
   Quelle des ältesten Preises.
7. **An allen vier Flächen** (AK-04), rein lesend:
   - Fenster: `HandelPreisanhalt`, mit den Zeilen des Entwurfs;
   - MCP: `lesen.handelsanhalt`, mit einem Vorgang;
   - CLI: `mse handel anhalt`, mit einem Vorgang, Beträge in der kleinsten
     Einheit wie bei `markt-konto`;
   - REST: `GET /handel/{vorgang}/anhalt`, mit einem Vorgang.

   Die Änderung ist erweiternd.

## Gemessen (NAT-01)

Hüllenprobe `trade02`, derselbe Szenariosatz wie in 0151, jetzt mit einem
Preis aus einer Gegenstelle auf 127.0.0.1:

| | 0151 | jetzt |
| --- | ---: | ---: |
| sichtbare Knöpfe | 35 | 36 |
| sichtbare Eingaben | 12 | 0 |
| Überschriften | 3 | 3 |
| Schrittknöpfe, die hier nicht gehen | 0 | 0 |

**Ein Knopf mehr** ist der zweite „+ Zeile“. **Zwölf Eingaben weniger**,
weil das Formular erst auf Wunsch aufgeht und nach der Einigung gar nicht.

## Abweichung vom Artboard

- **Die Richtung**: das Artboard nennt 19,90 € gegeben gegen 17,95 €
  erhalten „zu deinen Gunsten“. Gerechnet wird hier erhalten weniger
  gegeben, und der Satz sagt es so.
- **„+ Aus der Sammlung“** heißt hier „+ Zeile“. Was aus der Sammlung
  kommt, reist über den Anschluss (UX-03) und öffnet das Formular in der
  eigenen Spalte schon ausgefüllt.

## Verworfen

- **Ein Kurs, um Euro und Dollar zu einer Summe zu machen**: das wäre eine
  zweite Quelle mit ungelesenen Bedingungen (§16, Satz 1).
- **Die Preise im Vertrag mitführen**: dann läge eine Bewertung auf dem
  Draht, und die regelt TRADE-03 ausdrücklich anders.
