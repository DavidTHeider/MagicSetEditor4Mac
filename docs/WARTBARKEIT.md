# Wartbarkeit — Brennpunkte und was an ihnen gemessen wird

**Stand: 2026-09-25** — ENG-03
(#1728),
die Listen der Flächen #1825.

**Größe allein misst nichts**, sagt das Issue, und dieses Projekt hat dafür
einen Beleg: die Zeilengrenze aus §2 hat drei Einfuhrzyklen **erzeugt** — jeder
entstand, als eine Datei an der Grenze geteilt wurde und die Hälften einander
weiter brauchten. Diese Datei führt deshalb, **woran** eine Stelle teuer wird,
mit einer Zahl, und jeder Umbau hat ein Ziel, das man nachmessen kann (AK-01).

## Die fünf Maße

| Maß | Wie erhoben | Worüber |
| --- | --- | --- |
| **Änderungshäufigkeit** | `git log --since=2026-08-24 --name-only`, ohne Tests | 510 Commits auf `main` in einem Monat |
| **Komplexität** | `gocognit` aus golangci-lint v2.13.2, ohne Tests | die Pakete der Pipeline: `./internal/... ./cmd/... .` |
| **Zyklische Abhängigkeit** | Go: der Übersetzer lässt keine zu. Oberfläche: `frontend/src/einfuhrzyklen.test.ts` über jede Einfuhr, die zur Laufzeit bleibt | 845 Dateien unter `frontend/src`, ohne Tests |
| **Doppelte Logik** | `dupl` aus golangci-lint v2.13.2, ab 150 Token, ohne Tests | dieselben Pakete |
| **Testbarkeit** | Laufzeit je Paket unter `-race` aus `make ci`; Abdeckung aus `scripts/ci-abdeckung.sh` | drei volle Läufe vom 2026-09-24 |

**Die Reichweite gehört zur Zahl** (§10): `gocognit` und `dupl` sehen Go und
keine Oberfläche; die Zyklen sehen die Oberfläche und kein Go, weil Go sie
verbietet.

## Brennpunkte

| Brennpunkt | Maß | Vorher | Ziel | Stand |
| --- | --- | --- | --- | --- |
| **Racelauf von `internal/app`** | Wandzeit in der Teststufe | **1647 s, 1566 s und 1489 s** in drei Läufen — bei einer Frist von 1800 s; kein anderes Paket brauchte mehr als 517 s | **≤ 600 s** | **umgebaut**: `scripts/teilrennen.sh` fährt dasselbe Testprogramm in sechs Prozessen — **511 s** allein gemessen (mit vier: 610 s), **829 s** in der Pipeline neben allen anderen Paketen, dieselben 19 ungelaufenen Anweisungen |
| **Einfuhrzyklen der Oberfläche** | stark zusammenhängende Teile mit mehr als einer Datei | **3** — sechs, zwei und zwei Dateien | **0**, gehalten von einem Wächter | **umgebaut** — 0 |
| **Doppelte Logik** | Paare ab 150 Token | **3** | nur, was keine Entscheidung trägt | **umgebaut** — 1 (unten) |
| **Die Listen der Flächen** | Änderungen in einem Monat; Stellen, die eine neue Bindung verlangt | `fassade.ts` 90, `cmd/mse/hilfe.go` 80, `internal/app/anwendung.go` 70, `cmd/mse-rest/dienst.go` 61, `api.ts` 59, `internal/rest/wegetafel.go` 57, `internal/mcp/werkzeugliste.go` 52 — und **drei** Namenslisten für dieselbe Tatsache | eine neue Fähigkeit berührt **eine** Liste je Fläche und keine Aufzählung daneben | **erster Schnitt** (#1825): die Namen der Bindungen stehen nur noch im Vertrag — unten, mit der Messung |
| **Komplexe Funktionen** | `gocognit` | **3** über 30 (`mcp.ablauf` 35, `service.ohneAuszeichnungen` 35, `storage.schrittGehen` 31), **18** über 20 | keine neue über 35 | beobachtet — geteilt wird, wer angefasst wird |
| **Verlaufsprosa im Quelltext** | Absätze „Hier stand …" | Go 262 in 221 Dateien, Oberfläche 73, Skripte 16 — ohne Tests | jeder erklärt eine **heutige** Zusage oder steht in einem Vermerk | **offen** — AK-03, eigener Durchgang |

## Was die Umbauten gefunden haben

**Doppelte Logik ist kein Stilproblem, sondern ein Fehler, der zweimal
dasteht.** Zwei der drei Paare trugen einen:

- **`nutzer-bild` und `profil-badge`** (`cmd/mse/nutzerbild.go`) riefen bei
  `--keines` **erst das Setzen** mit diesem Pfad und dann das Wegnehmen. Lag im
  Arbeitsverzeichnis eine Datei dieses Namens, wurde sie gesetzt, die Fassung
  zählte hoch, und das Wegnehmen scheiterte an ihr — ein Konflikt, den der
  Befehl selbst erzeugt hatte. **Zweimal, in beiden Kopien.** Jetzt steht die
  Wahl einmal (`bildwechsel`), und `TestKeinesNimmt…` hält sie für beide.
- **Die Blöcke der Karten und der Schlüsselwörter** (`internal/service`)
  hatten denselben Fehler **nacheinander**: erst zählte der Kartenweg mit,
  statt nach der Stelle zuzuordnen (#621), dann der Schlüsselwortweg
  (Zweitblick über Phase W). Jetzt gehen beide durch `bloeckeAngleichen`.
- **Das dritte Paar bleibt**: die zwei MCP-Werkzeuge am Profilbild und am
  Abzeichen (`internal/mcp/nutzerbild.go`). Sie sind **Beschreibungen** —
  Name, Erklärung, Schema — und ihre Rufe reichen je vier Werte an eine
  Pflege weiter, ohne etwas zu entscheiden. Sie zusammenzulegen machte den
  Vertrag, den ein Sprachmodell liest, schwerer lesbar und gewönne nichts.

**Und der geteilte Lauf fand zwei Wächter, die die Maschine maßen.** Mit
sechs Teilen ist die Teststufe voll, und die Last wechselt, während Teile
fertig werden. `TestDerZerteilerWaechstNichtQuadratisch` wurde im ersten
Lauf rot — 0,50 s gegen 3,77 s, das 7,5-Fache bei einer Schranke von 6 —,
ohne dass der Zerteiler sich geändert hatte; der Abgleich der Tauschkandidaten
war am 2026-09-22 schon einmal genauso rot. **Beide messen jetzt anders**, und
jede der drei Änderungen hat ihren eigenen Grund: **Rechenzeit** (`getrusage`)
statt Wanduhr — sie zählt nicht, wie lange der Prozess auf einen Kern wartet;
**abwechselnd**, kurz und lang im Wechsel über fünf Runden — sonst erwischen
beide Größen verschiedene Maschinen; und **ohne Speicherbereinigung** — der
große Tauschlauf löste Sammelzyklen aus, der kleine keinen, und in Rechenzeit
kam ohne jede Last das 8,9-Fache heraus. Danach: 4,4-fach allein, 4,5-fach
unter Race-Detektor und wechselnder Last; ein eingebauter quadratischer
Aufwand macht beide Wächter weiter rot (14,4- und 15,4-fach).

**Ein dritter maß weiter mit der Wanduhr** — die fünfzigtausend Beziehungen
neben dem Abgleich, in derselben Datei, dreimal hintereinander und mit
Speicherbereinigung; dieser Absatz sagte trotzdem „beide messen jetzt anders"
und die Pipeline-Doku „zwei Wächter". Angestrichen vom Zweitblick über Phase
6.2. **Seitdem steht die Messregel in einer Funktion** (`abwechselnd` in
`tauschmessung_test.go`), die beide Tauschproben rufen, und gesucht ist nach
jeder Zeitmessung in Tests ohne Build-Tag: die übrigen prüfen Fristen, keine
setzt zwei Zeiten ins Verhältnis. **Gemessen am 2026-09-24**, allein: 4,7- bis
4,8-fach, unter Race-Detektor und Abdeckung 4,7- bis 4,9-fach; ein
eingebauter quadratischer Aufwand macht ihn rot (14,3-fach, unter Race
14,2-fach).

**Und die Zyklen** entstanden alle an der Zeilengrenze:
`tischgeometrie.ts` las `Stapelbreite` aus `tischplan.ts`, das
`tischgeometrie.ts` liest (jetzt bei den Maßen in `masse.ts`); drei Seiten
lasen `bildwunsch` über eine Wiederausfuhr in `Partietisch.tsx`, das sie
einbindet (jetzt direkt aus `tischbild.ts`); `quelltextfarben.ts` las einen Typ
mit `import { type … }`, und das bleibt unter `verbatimModuleSyntax` eine
Einfuhr (jetzt `import type`).

## Welche Listen der Flächen dasselbe sagen

**Gemessen am 2026-09-25** für #1825.
Die Frage war nicht „welche Dateien ändern sich oft", sondern „welche sagen
dieselbe Tatsache" — nur die darf zu einer werden, ohne dass eine Fläche etwas
verliert (§16).

| Liste | Was sie sagt | Dasselbe wie | Entscheidung |
| --- | --- | --- | --- |
| `frontend/src/fassade.ts` und vier Gruppendateien (`partie-`, `deck-`, `peer-`, `dateimethoden.ts`) | welche Bindungen das Fenster ruft — 428 Namen | `bindungsvertrag.json` (428), `gebundeneMethoden` (428); **kein Name nur in einer** | **gefallen**: die Fassade liest den Vertrag, `tsc` hält die Gleichheit mit `Api` in beide Richtungen |
| `gebundeneMethoden` in `internal/app/bindungen_intern_test.go` | welche Methoden Wails binden darf | dieselben 428 | **gefallen**: der Wächter liest den Vertrag |
| `frontend/src/bindungsvertrag.json` | Name **und Abdruck** jeder Signatur, quittiert | — sagt mehr | **bleibt**: die eine Liste, von Hand quittiert |
| `frontend/src/api.ts` samt Teilschnittstellen | die Typen jeder Bindung | — sagt mehr | bleibt |
| `cmd/mse/hilfe*.go` | Aufruf, Argumente, Satz je Befehl | — andere Gestalt | bleibt |
| `internal/rest/wegetafel.go` | Methode und Pfad je Weg | — andere Gestalt | bleibt |
| `internal/mcp/werkzeugliste.go` | Name, Beschreibung, Schema je Werkzeug | — andere Gestalt | bleibt |
| `internal/app/anwendung.go`, `cmd/mse-rest/dienst.go` | welche Naht woran hängt | — Verdrahtung, keine Aufzählung | bleibt |
| `docs/FAEHIGKEITEN.md`, `SCHREIBWEGE.md`, `AUTORISIERUNG.md` | Zuordnung: Fähigkeit, schreibend oder lesend, Klasse und Rolle | — sagen mehr | bleiben, jede hält ihr Wächter |

**Die drei letzten Gruppen sind keine Doppelungen**, und das ist das
Ergebnis der Analyse: eine REST-Route, ein MCP-Schema und ein Hilfesatz sind
drei **Gestalten** derselben Fähigkeit, keine drei Abschriften eines Namens.
Wer sie aus einer Beschreibung erzeugte, müsste in diese Beschreibung
hineinlegen, was jede Fläche anders braucht — Pfadmuster, JSON-Schema,
Beispielaufruf — und hätte die drei Listen in eine vierte verlegt. Das ist
der pauschale Umbau, vor dem §15 Satz 11 warnt, und er ändert vier
öffentliche Flächen zugleich.

### Die Messung

**Vorher** — gemessen an der letzten neuen Bindung vor diesem Schnitt,
`TischabzugSpeichern` am 2026-09-22: sieben Wächter wurden rot, und mit
`api.ts` verlangte die Bindung **acht** Stellen außer ihrer eigenen Methode —
`api.ts`, `gebundeneMethoden`, die Fassadenliste (dort `partiemethoden.ts`),
`FAEHIGKEITEN.md`, `AUTORISIERUNG.md`, `SCHREIBWEGE.md`, `README.md` und
einen Kopfkommentar in `scripts/nahtlauf.mjs`, der die Zahl der Bindungen
nannte. **Seither** kam mit ENG-02 der Vertrag dazu (`bindungsvertrag.json`),
und der Kopfkommentar fiel weg (#1823) — heute vor diesem Schnitt also
wieder **acht**.

**Nachher** — gemessen am 2026-09-25 an einer **Probebindung** auf diesem
Stand: eine Methode `Probebindung() string` an `Bindungen`, dann so lange
Wächter gefahren (`internal/app`, `internal/werkzeuge`, `tsc`), bis keiner
mehr rot war. Sie meldeten sich in drei Wellen und verlangten **sechs**
Stellen außer der Methode und ihrem Aufrufer im Fenster:

| Welle | Wächter | Stelle |
| --- | --- | --- |
| 1 | `bindungsvertrag_test`, `bindungsliste_intern_test` | `bindungsvertrag.json` — die Vorlage druckt der Wächter |
| 1 | `bindungsstelligkeit_test` | eine `api…`-Datei, als `Name: (…) =>` |
| 1 | `faehigkeiten_test`, `flaechengroesse_test`, `flaechenzahlen_test` | `FAEHIGKEITEN.md` — Zeile und vier Zahlen |
| 2 | `schreibwege_test` | `SCHREIBWEGE.md` — Liste und zwei Zahlen |
| 2 | `autorisierung_test` | `AUTORISIERUNG.md` — Zeile und drei Zahlen |
| 3 | `readmezahlen_test` | `README.md` — vier Zahlen |

**Entfallen sind die zwei Namenslisten**; die Fassade bekam den Namen
ohne Zutun, `tsc` war nach dem Vertrag und `api.ts` sofort grün. **Die
Probe ist wieder entfernt** — sie war eine Messung und keine Fähigkeit.

**Was die Messung nicht sagt**: ob sechs die richtige Zahl ist. Vier der
sechs Stellen sind Zahlen in Dokumenten, und jede ist eine Zusage an einen
Leser, die ein Wächter hält (#1287). Sie aus den Dokumenten zu nehmen hieße,
den Lesern die Zahl zu nehmen — das ist eine Entscheidung über die
Dokumente und keine über die Listen.

## Warum die Listen der Flächen ein eigenes Issue sind

**Sie sind der teuerste Brennpunkt und der gefährlichste Umbau.** Jede neue
Fähigkeit berührt heute die Fassade der Oberfläche, die Hilfe der
Kommandozeile, die Wegetafel von REST und die Werkzeugliste von MCP — dazu die
Wächter, die das zählen. Das ist gewollt (§16: dieselbe Fähigkeit an jeder
Fläche), aber die Zahl der Stellen je Fähigkeit ließe sich senken, etwa durch
**eine** Beschreibung je Fähigkeit, aus der jede Fläche ihre Zeile nimmt.
**Das ändert vier öffentliche Flächen zugleich** und gehört nach §16
eingeordnet; ein Nebenbei in diesem Paket wäre der pauschale Umbau, den §15
verbietet.

## Was diese Datei nicht sagt

**Ob eine Stelle verständlich ist.** Keines der fünf Maße misst das; sie
messen, was es kostet, eine Stelle zu ändern. Und die Zahlen gelten für den
Tag, an dem sie erhoben sind — wer sie fortschreibt, schreibt das Datum mit.
