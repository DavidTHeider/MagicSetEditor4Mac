# 0175 — Ein Zusammenbruch im Treiber verbraucht die Verbindung, und nichts wird wiederholt

**Stand:** angenommen am 2026-10-07 · BK-03a, BK-03b
(#2127,
#2128).

## Die Frage

**`database/sql` fängt keinen `panic` des Treibers.** Bricht er in einer
Anweisung zusammen, gibt niemand die Verbindung an den Vorrat zurück. Der
Vorrat hat **eine** (`SetMaxOpenConns(1)`): fängt Wails oder
`domain.Auffangen` den Zusammenbruch, lebt der Prozess weiter, aber jeder
weitere Ruf wartet bis zu seiner Frist.

**Ob es einen realen Auslöser gibt**, ist gelesen und nicht herbeigeführt:
`ncruces/go-sqlite3` v0.35.4 wirft bei erschöpftem Speicher der WASM-Laufzeit
(`errutil.OOMErr`, etwa `stmt.go`, `conn.go`) und bei verletzten
Zusicherungen (`errutil.AssertErr`). Einen davon in einem Lauf auszulösen,
hieße, die Laufzeit an ihre Speichergrenze zu treiben; das ist offen.

## Die Messung

Am 2026-10-07 mit einer echten Datei, einer Verbindung im Vorrat und einer
Schicht zwischen Treiber und Hülle, die an einer benannten Stelle
zusammenbricht (`stolperschicht_intern_test.go`,
`treiberzusammenbruch_intern_test.go`). **Ohne Hülle** drang der
Zusammenbruch bei `Exec`, `Prepare`, `Query`, `Stmt.Exec`, `Stmt.Close`,
`BeginTx` und `Rows.Close` nach oben; bei BK-04 war das Ergebnis eine
belegte Verbindung und `database is locked` für jeden anderen Prozess.

## Die Entscheidung

**Eine Hülle um jede Verbindung** (`fangtreiber.go`) — Verbindung,
Anweisung, Zeilenfolge, Transaktion. Sie fängt den Zusammenbruch, gibt ihn
als `domain.ErrZusammengebrochen` zurück und **verbraucht** die Verbindung:
`IsValid` sagt danach `false`, und `database/sql` verwirft sie, statt sie
wieder auszugeben (`putConn` → `validateConnection`, Go 1.27.1). Die nächste
entsteht frisch über denselben Rückruf — Schlüssel, `synchronous`,
Volltextsuche.

| Stelle | Endstand | danach |
| --- | --- | --- |
| `Exec`, `Prepare`, `Query`, `Stmt.Exec`, `BeginTx`, `Tx.Commit`, `Tx.Rollback`, `Rows.Next`, `Rows.Close` | `ErrZusammengebrochen` | Vorrat frei, zweite Verbindung, der nächste Ruf gelingt |
| `Rows.Scan` | `sql: Scan error … ErrZusammengebrochen` | dito |
| `Stmt.Close` | **kein Fehler** — `database/sql` verwirft ihn beim Schließen einer Anweisung | dito; die Verbindung ist trotzdem verbraucht |
| `Connect`, Schlüssel und Pragmas | `ErrZusammengebrochen` | Vorrat zählt **null** Verbindungen, eine schon geöffnete ist geschlossen |

**Nachgetragen am 2026-10-07 nach dem Zweitblick über C und D** — zwei
Lücken, beide gemessen:

- **Das Verbinden lag außerhalb der Hülle.** `database/sql` zählt eine
  Verbindung **vor** `Connect` mit und nimmt sie nur bei einem
  zurückgegebenen Fehler wieder heraus (`sql.go:1434`, Go 1.27.1); ein
  `panic` beim Öffnen oder beim Setzen des Schlüssels hielt den einzigen
  Platz für immer, der nächste Ruf lief in seine Frist. Jetzt fängt
  `fangverbinder.Connect` beides.
- **Verworfen war nicht geschlossen.** Nach einem Zusammenbruch in
  `Stmt.Close` war die Anweisung nicht finalisiert, und SQLite schließt
  keine Verbindung, an der eine hängt (`sqlite3_close` gibt `BUSY`,
  ncruces `conn.go`, v0.35.4) — der Vorrat sah gesund aus, Datei und
  Laufzeitspeicher blieben. Jetzt finalisiert die Hülle nach dem
  Zusammenbruch ein zweites Mal; der Treiber sagt zu, dass das für eine
  schon geschlossene Anweisung harmlos ist (`stmt.go:21`). **Was dabei
  offen bleibt**: bricht auch der zweite Versuch zusammen, bleibt der
  Rohgriff — gemessen ist nur der einmalige Zusammenbruch. Der Lauf hält
  jetzt das Ergebnis jedes Schließens fest, statt nur den Vorrat zu zählen.

**Bewusst nicht `driver.ErrBadConn`.** Darauf wiederholt `database/sql` den
Ruf auf einer neuen Verbindung; ein Schreiben, das vor dem Zusammenbruch
schon wirkte, stünde zweimal im Bestand. Gemessen: ein `INSERT`, das nach
dem Ausführen zusammenbricht, steht genau einmal da.

**Eine Datei liest sich danach wie davor** — auch verschlüsselt: die neue
Verbindung trägt den Schlüssel. **Ein Bestand im Arbeitsspeicher**
(`:memory:`, nur in Prüfläufen) verlöre mit der Verbindung seinen Inhalt;
im Betrieb gibt es ihn nicht.

## Was es nicht ist

- **Keine Wiederholung und keine Rettung.** Der Ruf, der zusammenbrach,
  scheitert; wer ihn wollte, sieht den Fehler.
- **Kein Ersatz für BK-04** ([0173](0173-wer-eine-transaktion-beginnt-gibt-sie-zurueck.md)):
  bricht der **Aufrufer** einer Transaktion zusammen, ist das nicht der
  Treiber, und dort rollt das `defer` zurück.
- **Der Satz der Meldung** ist der von `ErrZusammengebrochen` — er spricht
  von einem Nebenlauf. Für die Oberfläche genügt der Schlüssel; der Grund
  steht im Protokoll.

## Verworfen

- **`recover` in jeder Ablage** — dreihundert Stellen statt einer, und keine
  davon wüsste, welche Verbindung verbraucht ist.
- **Den Vorrat neu öffnen** — verlöre einen Bestand im Arbeitsspeicher und
  hielte alle anderen Verbindungen an, um eine zu ersetzen.
