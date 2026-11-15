# 0170 — Der Rufkontext endet mit dem Ruf

**Stand:** angenommen am 2026-10-06 · BK-01,
#2109.
**Ersetzt** den Absatz „Was dieser Typ deshalb tut" in `internal/app/bindungen.go`
(Wails reiche keinen Kontext hinein, der Kontext komme aus dem Fenster).

## Die Frage

**Ein abgebrochener Ruf des Fensters wartete weiter.** Jede Bindung lief mit
dem Kontext aus `App.Bereit` — dem des Fensters, der erst mit der Anwendung
endet. Wails 3 bricht einen Ruf ab, wenn die Oberfläche es verlangt
(`CancellablePromise.cancel`) oder sich das Fenster schließt
(`CancelWindowCalls`, `messageprocessor.go` in v3.0.0-beta.23); der Kontext,
den es dafür abbricht, kam bei uns nie an.

Gemessen am 2026-10-06, `internal/app/rufabbruch_intern_test.go`, vor der
Änderung:

| Was hält | Wirkung des Abbruchs |
| --- | --- |
| die **eine** Verbindung des Pools (`SetMaxOpenConns(1)`) | keine — der Ruf kehrte nach 5 s Frist nicht zurück |
| ein **fremder Schreiber** (`BEGIN IMMEDIATE` aus einer zweiten Verbindung) | keine — zurück erst nach 4,8 s, als `busy_timeout(5000)` ablief; wäre die Sperre vorher gefallen, hätte er geschrieben |

## Die Entscheidung

**Jede exportierte Bindung nimmt den Kontext ihres Rufs als ersten
Parameter** — Wails füllt ihn (`needsContext`, `bindings.go`), und das
Fenster ruft unverändert mit denselben Argumenten. Er geht **durch das Tor**:
`Bindungen.app(ctx)` und `Bindungen.tor(ctx)` prüfen Sitzung und Schalter
mit ihm, `Bindungen.ruf(ctx)` gibt ihn weiter (`internal/app/rufkontext.go`):
**das Ende vom Ruf, die Werte zuerst vom Ruf und dann vom Fenster** — was
`App.Bereit` am Fenster ablegt, kommt so weiter an — und eine **neue**
Vorgangskennung je Ruf.

**Dabei fiel ein zweiter Fehler auf**: bis hierher trug jeder Ruf des Fensters
**dieselbe** Vorgangskennung. `MitNeuemVorgang` würfelt nur, wenn noch keine
da ist, und am Kontext aus `Bereit` war immer schon eine — die Kennung in
einer Fehlermeldung führte im Protokoll zu jedem Ruf zugleich.
`bindungenkontext_test.go` hält beides, die Gegenprobe mit der alten Regel
gab „drei Rufe, 1 Kennung".
`App.kontext()` gibt es nicht mehr; was ohne Ruf geschieht — Menüleiste,
nachgeholte Datei, Rückspielrest beim Start —, nimmt `App.fensterkontext()`.

**Ein abgebrochener Ruf ist kein klemmender Bestand**: `verbindungNachziehen`
schreibt dann keinen Fehler ins Protokoll.

## Was länger lebt als die Antwort

**Wails bricht den Kontext auch dann ab, wenn der Ruf regulär endet**
(`defer cancel()` in `messageprocessor_call.go`). Was nach der Antwort
weiterläuft, darf ihn deshalb nicht behalten. Gesucht wurde über jede
Goroutine in `internal/app`, `internal/service` und den Blättern, die eine
Bindung erreicht, und über jeden Kontext, der in einem Verbund landet:

| Lauf | Gestartet von | Gelöst über | Beendet durch |
| --- | --- | --- | --- |
| Partie als Gast | `PartieBeitreten` | `context.WithoutCancel` (`partiebeitritt.go`) | `PartieBeitrittBeenden`, Ende der Leitung, Anwendungsende |
| Horchposten | `HorchStarten` | `context.WithCancel(context.WithoutCancel(…))` (`peerhorchen.go`) | `HorchBeenden`, Anwendungsende |
| Fernhorchen | `FernHorchStarten` | dasselbe (`peerfern.go`) | `FernHorchBeenden`, Anwendungsende |
| REST- und MCP-Dienst | `RESTStarten`, `MCPStarten` | dasselbe (`dienstwerk.go`) | `RESTBeenden`, `MCPBeenden`, Schalter, Anwendungsende |
| Partie als Wirt | der Horchposten | dessen Kontext, nicht der eines Rufs (`partieaufnahme.go`) | Ende der Leitung |

**Alle anderen Wege sind kurz**: KI-Stapel, Ausfuhr, Dateiläufe, Anmeldung
und Bildvorschläge warten innerhalb des Rufs. Transaktionen leben nur als
Parameter innerhalb eines Rufs; keine steht in einem Verbund. Der Tischmelder
aus `gastAmTisch` lebt nur so lange wie `PartieWuenschen`.

**Was diese Tabelle nicht hält**: eine neue Goroutine, die den Rufkontext
behält. Kein Wächter sieht das; die Probe ist, dass ein Lauf nach seiner
Startantwort noch lebt (`scripts/huellenprobe/nat04.mjs` für die Partie).

## Dieselbe Fähigkeit an den anderen Flächen

| Fläche | Abbruch kommt von | Antwort | Beleg |
| --- | --- | --- | --- |
| Fenster | `cancel()`, Fenster geschlossen | `abgebrochen` (`CodeAbgebrochen`) | `internal/app/rufabbruch_intern_test.go` |
| REST | Verbindung des Klienten endet (`anfrage.Context()`) | HTTP 499, `abgebrochen` | `internal/rest/fehlerwege_test.go` |
| Kommandozeile | ⌃C (`signal.NotifyContext` in `cmd/mse/main.go`) | Abbruch des Laufs | `cmd/mse/abbruch_test.go` |
| MCP | Ende der Eingabe oder Signal (BK-07c) | Ende der Bedienung | `internal/mcp/bedienabbruch_test.go` |

**Die Laufverträge weichen ab, und das mit Grund**: REST und MCP kennen je
Anfrage einen Kontext von ihrem Server, die Kommandozeile einen je Prozess.
Das Fenster ist die einzige Fläche, deren Abbruch von der Oberfläche selbst
kommt.

## Verworfen

- **Ein Fensterkontext mit Frist.** Eine Einheitsfrist bräche lange Läufe ab,
  die zu Recht lange laufen (Bildvorschläge, Ausfuhr), und ließe kurze immer
  noch warten, bis sie abläuft. Der Abbruch gehört dem, der wartet.
- **Den Kontext nur in den „langsamen" Bindungen.** Welche das sind, sagt der
  Bestand, nicht die Bindung: jede liest über denselben Pool mit einer
  Verbindung.
