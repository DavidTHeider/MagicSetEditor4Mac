# 0083 — Ressourcenbudgets ohne gegenseitige Aussperrung

**Stand:** angenommen
**Datum:** 2026-09-09
**Anlass:** Werkstück O5 von Phase 5.2
(#1429):
lokale Maschinenzugänge dürfen den Editor weder aushungern noch unbeschränkt
Ressourcen binden.

## Lage

**Was am 2026-09-09 galt, ist nachgelesen** (Stand `1e72164c`):

- Der Dienst hatte **zwei Fensterdrosseln**, je 60 in der Minute
  (`internal/rest/dienst.go`): eine für Fehlversuche, eine für Arbeit. Beide
  zählten **für alle Aufrufer zusammen**: sechzig falsche Marken von irgendwem,
  und der richtige Aufrufer war für eine Minute draußen. `Retry-After` war eine
  feste 60, gleich wie viel vom Fenster um war.
- MCP kannte **keine** Grenze; die Kommandozeile auch nicht. Wer über eine der
  beiden zehn Rasterungen zugleich anstieß, tat es. Kein Dienst hatte eine
  Parallelitätsgrenze.
- Ein Principal war seit O2 ([0080](0080-maschinenzugaenge-und-widerrufbare-sitzungen.md))
  bekannt — Token je technischem Profil, eine Sitzung je MCP-Prozess —, aber
  nichts rechnete je Principal.

## Entscheidung

### 1 · Guthaben statt Fenster, je Principal und global

Ein [Kontingent](../../internal/domain/haushalt.go) ist ein Eimer mit Decke
(`Burst`) und Zulauf (`Nachfuellen` je `Fenster`); jeder Ruf kostet nach seiner
**Kostenklasse** (Lesen 1, Schreiben 2, Bild 3, Rastern 5, Ausfuhr 8, Einfuhr
8, Masse 10 — Verhältnisse, keine Messungen). Ein Maschinenzugang bucht bei
**seinem** Guthaben und beim **globalen**; das Fenster hat keines, bucht aber
global mit — und **nur es darf in die Reserve**: was unter `Reserve` liegt,
bekommt kein Principal. So bleibt der Editor arbeitsfähig, wenn ein Skript das
Übrige aufbraucht.

Die Zahlen: 120 je Principal und Minute (die Parität zur alten Drossel — 60
Schreiben oder 120 Lesen), 600 global mit 120 Reserve. Die Wartezeit ist die
bis zur nächsten möglichen Einheit, aufgerundet auf Sekunden, nie null.

### 2 · Plätze je Ressourcenklasse, service-nah

`service.Haushalt` hält je Ressourcenklasse einen Semaphor (Rastern 2,
Einfuhr 1, Ausfuhr 1, Bild 4, Masse 1). **Er steht in den Diensten, die die
Arbeit tun** — Importeur, Exporteur, Bogenleger, Massenpflege, Preispflege,
Bildweg —, nicht an einer Fläche: dieselbe Grenze für Fenster, Kommandozeile,
MCP und Dienst, und keine Fläche kann daran vorbei. Ein Prozess hat **einen**
Maschinenhaushalt; ein Dienst nimmt über `MitHaushalt` einen eigenen, damit ein
Test nie auf einen fremden Platz wartet (§16). Ein Abbruch des Kontexts gibt
den Platz frei; Freigeben ist idempotent.

### 3 · Die Sperre hängt am Nachweis

`service.Zugangssperre` führt je **vorgezeigtem Nachweis** — als SHA-256, nie
im Klartext — eine `Sperrlage`: drei freie Versuche, dann eine Sekunde, jede
weitere verdoppelt bis fünf Minuten. Wer `mse_falsch` rät, sperrt
`mse_falsch`; wer `mse_richtig` hat, merkt nichts. Kein Schlaf je Anfrage:
die Antwort ist sofort da und nennt `blockedUntil` als Restzeit. Die Drossel
für alle Ratenden zusammen (60 in der Minute) bleibt dahinter — gegen ein
Wörterbuch, das je Versuch einen frischen Nachweis nimmt.

### 4 · Begrenzter Zustand

Die Tabellen der Principals (256) und der Nachweise (1.024) verdrängen den am
längsten nicht Gesehenen. Ein verdrängter Principal beginnt wieder voll — ein
Guthaben ist keinen Speicher wert, den ein Angreifer bestimmt. Rumpf und
Antwort des Dienstes sind auf 32 MiB begrenzt; eine größere Antwort wird
**abgebrochen, nicht gekürzt**, denn gekürztes JSON sähe aus wie ganzes.

### 5 · Die Grenzen sind eine Fähigkeit

`Grenzen`, `lesen.grenzen`, `mse grenzen`, `GET /grenzen` — eine Auskunft aus
dem Maschinenhaushalt, an allen vier Flächen dieselbe ([0033](0033-was-ueber-fassungen-hinweg-zugesagt-wird.md), Regel 9).

## Einordnung (§16)

**Erweiternd.** Neue Auskunft an allen Flächen; `Retry-After` bleibt eine
Kopfzeile, nur ihr Wert stimmt jetzt; Grenzen waren nie Vertragszusage. Nichts
Brechendes.

## Verworfen

- **Die Drossel je Principal vervielfachen.** Ein festes Fenster kann keinen
  Schub von einem Dauerlauf unterscheiden und keine ehrliche Wartezeit nennen.
- **Die Plätze an den Flächen.** Dann liefe die Kommandozeile daran vorbei —
  genau der Fall, den das Issue benennt.
- **CAPTCHA, Proof-of-Work, ein Dienst im Netz.** Der Dienst bleibt auf der
  Rückschleife; ein Peer-Netz ist ein eigener Modus (Phase 5.6).
- **Ein Guthaben fürs Fenster.** Die Vorschau rastert beim Tippen dutzendfach;
  ein Guthaben hielte den Menschen auf, den es schützen soll. Sein Schutz ist
  die Reserve und der Platz, nicht die Zahl.

## Folgen

- Ein Bestand, den mehrere Prozesse teilen, hat **je Prozess** Grenzen; die
  Datei ist die gemeinsame Grenze (0017), nicht der Haushalt.
- Wer eine Grenze ändert, ändert `domain.Vorgabegrenzen` und
  `docs/ADMINISTRATION.md` — und nichts sonst: die Flächen lesen.

## Nachtrag 2026-10-02 — eine Kostentafel, und die Grenzen der Dienste dabei

**Der Satz „die Flächen lesen" stimmte für die Zahlen und nicht für die
Zuordnung.** Welche Kostenklasse ein Ruf hat, entschied jede Fläche selbst:
REST nach einer kurzen Liste von Wegen, MCP nach Wörtern im Werkzeugnamen.
Der Lastlauf QA-08 fand dieselbe Fähigkeit zu verschiedenen Preisen —
Folgebefund F1, #2003.

- **Die Zuordnung steht jetzt neben den Zahlen**: `domain.Kostentafel` nennt
  je teurer Fähigkeit ihre Klasse und ihre Wege an REST, MCP und
  Kommandozeile; beide Dienste fragen `domain.KostenklasseFuer`. Was nicht
  darin steht, kostet nach seiner Art — ein Wort im Namen macht nichts
  teuer. Was teuer ist und warum, steht im Kopf von `kostentafel.go`.
- **Ein Abbruch ist eine Klasse und kostet nichts** (`KostenAbbruch`): er
  nimmt Last, und was nichts kostet, wird nie abgewiesen — auch nicht unter
  der Reserve.
- **Verbindungs- und Auftragsgrenze gehören zu den Grenzen**:
  `Haushaltsgrenzen.Verbindungen` (16, `mse-rest` und `mse-mcp` über HTTP)
  und `Haushaltsgrenzen.Auftraege` (4 laufend, 16 behalten, eine Stunde).
  Damit gilt die Folge oben auch für sie: wer sie ändert, ändert
  `domain.Vorgabegrenzen` und `docs/ADMINISTRATION.md`.

**Einordnung (§16)**: `grenzen` bekommt drei Felder und `kosten` die Klasse
`abbruch` — **erweiternd**. Andere Preise für dieselben Rufe — **verträglich**,
denn Grenzen sind keine Vertragszusage (oben); spürbar ist es trotzdem, und
deshalb steht es im Änderungsprotokoll.
