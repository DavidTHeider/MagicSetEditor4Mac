# 0173 — Wer eine Transaktion beginnt, gibt sie zurück — auch im Zusammenbruch

**Stand:** angenommen am 2026-10-07 · BK-04
(#2129).

## Die Frage

**Sechs Schreibwege rollten nur zurück, wenn ihr benannter Fehler gesetzt
war** — `defer func() { if fehler != nil { … verwerfen(tx) } }()`. Ein
`panic` setzt ihn nicht. Gemessen am 2026-10-07 je Weg mit einer echten
Datei und einem Zusammenbruch vor einer Anweisung **in** der Transaktion
(`panikzuruecknahme_intern_test.go`), vor der Änderung:

| Wirkung | vorher |
| --- | --- |
| belegte Verbindungen des Pools (`SetMaxOpenConns(1)`) | 1 |
| zweite Verbindung, `BEGIN IMMEDIATE` mit 500 ms Frist | `database is locked` |
| derselbe Weg danach, 2 s Frist | `context deadline exceeded` |

Mit `_txlock=immediate` hält eine offene Transaktion die Schreibsperre der
**Datei**: nicht nur dieses Fenster wartete, sondern auch `mse`, MCP und
REST — bis das Programm endete. Die sechs: `imEinenZugAufStand` und
`inTransaktion` (Kartenfelder), `SetwertSetzenMitVorwert`,
`GruppeSetzenMitHerkunft`, `EinstellungenUebernehmen` und
`Kontaktablage.inTransaktion`. **Dieselbe Klasse an zwei weiteren Stellen**,
gefunden beim Durchsehen aller `BeginTx`: `imRetrievalzug` rollte nur im
Fehlerzweig zurück, `SprachpaketAendern` nur bei gesetztem Fehler — beide
rufen fremde Arbeit in ihrer Transaktion.

## Die Entscheidung

**Zurückgerollt wird immer, im `defer`, gleich nach dem Beginn**:
`defer func() { fehler = zurueckgenommen(tx, fehler) }()`.

| Ausgang | Zurückrollen | Fehler der Funktion |
| --- | --- | --- |
| `Commit` geglückt | `ErrTxDone`, wirkungslos | keiner |
| gewöhnlicher Fehler | läuft | der Fehler, ein Fehler des Zurückrollens angehängt (`errors.Join`) |
| `Commit` gescheitert | `ErrTxDone`, wirkungslos | der Fehler des `Commit` |
| `panic` | läuft | — der Zusammenbruch geht weiter nach oben, wo ihn `domain.Auffangen` oder Wails fängt ([0171](0171-ein-nebenlauf-faengt-seinen-zusammenbruch.md)) |

**Ein Fehler beim Zurückrollen ohne vorherigen Fehler wird nicht gemeldet**:
dann gab es nichts zurückzunehmen — `Commit` lief oder die Transaktion war
schon zu.

## Was es nicht ist

**Kein Wiederholen und keine Rettung.** Ein Zusammenbruch bleibt einer; die
Änderung ist verworfen, und wer sie wollte, sieht den Fehler. Was ein
Zusammenbruch **im Treiber** mit der Verbindung macht, ist BK-03a
(#2127) —
hier bricht der Lauf vor dem Treiber ab.

## Verworfen

- **Ein gemeinsamer Helfer `imGeschaeft` für alle acht.** Die Wege tragen
  benannte Ergebnisse (Vorwert, Stelle) und eigene Sätze für Beginn und
  Abschluss; sie in einen Rückruf zu zwängen hieße, acht Signaturen
  umzubauen, um ein `defer` zu ändern.
- **`recover` im Speicher.** Er fing, was nicht ihm gehört: ob ein
  Zusammenbruch weitergeht, entscheidet der Aufrufer, nicht die Ablage.
