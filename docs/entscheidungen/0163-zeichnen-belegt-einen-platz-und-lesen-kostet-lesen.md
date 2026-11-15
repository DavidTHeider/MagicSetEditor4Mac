# 0163 — Zeichnen belegt einen Platz an der Pforte, und Lesen kostet Lesen

**Stand:** angenommen am 2026-10-03 · Folgebefund F3 des Lastlaufs QA-08
(#2020),
Marke `new`. Gesucht wurde in `/Users/david/mse2-quelle/src`: MSE2 hat weder
Dienst noch Werkzeugserver, also auch keine Kosten je Aufruf; das Nächste
ist `src/gui/thumbnail_thread.cpp` (Fassung f9d9356 vom 2021-07-06), eine
Warteschlange, die **ein** Arbeiter (`ThumbnailThreadWorker`) für die
Vorschaubilder abarbeitet. Daraus folgt keine Zahl, nur dass dort schwere
Arbeit in einer Schlange läuft und nicht zugleich. Fortschreibung
von [0083](0083-ressourcenbudgets-ohne-aussperrung.md) (Plätze und Kosten) und
[0161](0161-die-kontingente-lassen-sich-abschalten.md) (was bleibt, wenn die
Guthaben gehen).

Zwei Fragen aus demselben Lesen von `docs/LASTMESSUNG.md`, und beide waren
Zusagen, die niemand eingelöst hatte.

## Die erste Frage: wer belegt einen Renderplatz?

`domain.RessourceRendern` stand in den Vorgabegrenzen, `GET /grenzen` nannte
zwei Plätze, und **außerhalb der Tests belegte nirgends jemand einen** — ein
`Belegen` oder `Versuchen` mit dieser Klasse fand sich weder in `internal/`
noch in `cmd/`. Acht Klienten, die zeichnen, bekamen deshalb nicht „zwei
Plätze", wie der Lastlauf es deutete, sondern acht gleichzeitige Rasterungen.

**Wer rastert** (nachgelesen am 2026-10-03, jeder Aufruf von `Zeichne` in
`internal/` und `cmd/`):

| Aufruf | Wer ihn nimmt |
| --- | --- |
| `service.Bildausfuhr.Schreibe` | Kartenbild: `mse zeichnen`, `schreiben.karte.zeichnen`, `POST /karten/{}/bild`, das Fenster |
| `service.Bogenleger.zeichne` | Druckbogen: `mse bogen`, `schreiben.set.bogen`, `POST /sets/{}/bogen`, das Fenster |
| `svg.schreiber.alsRaster` | die Rasterteile einer SVG, über `Bildausfuhr` |
| `service.Spieltexturen.zeichne` | Karten am Spieltisch (nur das Fenster) |
| `app.App.zeichneVorschau` | Vorschau, Probevorschau, Stilprobevorschau, Vorlagenvorschau (nur das Fenster) |
| `app.App.symbolbild`, `BibliothekssymbolFassungen` | Kacheln der Symbolpalette und der Fassungen (nur das Fenster) |

### Die Entscheidung

1. **Eine Pforte am Rasterer.** `render.Rasterer` bekommt über `MitPforte` eine
   `render.Pforte`, die er nach der Prüfung des Plans und **vor dem ersten
   Bildpunkt** fragt und nach dem Lauf — auch nach einem Abbruch — zurückgibt.
   `service.Rasterpforte` belegt dahinter einen Platz der Klasse `rendern` im
   Haushalt der Maschine; wer keinen bekommt, wartet, bis einer frei wird oder
   sein Kontext endet.
2. **Gehängt wird sie einmal, in `app.NeuesZeichenwerk`.** Das Fenster und
   jede Fläche ohne Fenster bauen ihr Zeichenwerk dort; beide Wurzeln
   (`zeichenwerkVerdrahten`, `Zeichennaht`) gehen hindurch, und `cmd/*` kennt
   `internal/service` weiterhin nicht.
3. **Ein Platz je Rasterung und nicht je Auftrag.** Ein Druckbogen hält für
   seine Dauer den Platz der Klasse `ausfuhr` und belegt für jede Karte einen
   der zwei Plätze `rendern`. Die Reihenfolge ist immer *erst `ausfuhr`, dann
   `rendern`*, und `rendern` wartet auf nichts weiter: es gibt keinen Kreis, in
   dem zwei Läufe aufeinander warten.
4. **Die Zahl bleibt zwei und steht, wo sie stand**: `parallel.rendern` in
   `mse grenzen`, `lesen.grenzen`, `GET /grenzen` und der Hilfe. Geändert hat
   sich, dass sie gilt.

### Was nicht durch die Pforte geht

Das Rastern **eines Symbols oder Bildbaums** (`SymbolAlsBild`, `BildAusBaum`
für die Wahlbilder einer Stiloption und `symbols_to_html`, die Miniaturen der
Symbolbibliothek über `render.RastereSymbol`): es ist klein und kommt innerhalb
anderer Arbeit vor. Eine Pforte davor hielte eine Rasterung auf, die selbst
schon einen Platz hat.

### Verworfen

- **Der Platz in jedem Dienst**, wie bei Einfuhr, Ausfuhr und Masse. Für das
  Zeichnen wären es sechs Stellen (Kartenbild, Bogen, SVG, Tisch, Vorschau,
  Palette), und jede neue, die rastert, müsste an die siebte denken. Alle gehen
  durch `Rasterer.Zeichne`; dort steht die Pforte, und kein Weg kommt daran
  vorbei.
- **Der Platz für die ganze Dauer eines Bogens oder einer Bildausfuhr.** Ein
  Bogen aus dreihundert Karten hielte einen der zwei Plätze, bis er fertig ist,
  und die Vorschau im Fenster wartete darauf — dasselbe, was die Zahl zwei
  verhindern soll („die Vorschau soll nicht auf den Druckbogen warten",
  `domain.Vorgabegrenzen`).
- **Die Pforte im Haushalt der Maschine voraussetzen.** Der Rasterer kennt
  keinen Haushalt (`internal/render` importiert nichts aus dem Projekt, §2);
  er kennt eine `Pforte`, und `internal/service` setzt sie.

### Was die Messung noch nicht sagt

**Der Lastlauf ist nicht wiederholt.** Die Zeile *REST rendern* in
`docs/LASTMESSUNG.md` (968 ms Median bei 7,2 Rasterungen je Sekunde, bis
2,7 GiB RSS) ist die von damals, als acht Rasterungen zugleich liefen; was
zwei statt acht an Durchsatz und Speicher ändern, ist **nicht gemessen**. Und
der Tisch des Fensters, der sechzig Texturen anfordert, teilt sich jetzt
zwei Plätze mit allem anderen, was im Fenster zeichnet — ob das spürbar ist,
sieht nur eine Probe am Fenster.

## Die zweite Frage: was kostet ein lesender Weg, der `POST` heißt?

Der Dienst las die Art eines Wegs an seiner Methode: alles außer `GET` und
`HEAD` buchte als Schreiben. MCP liest sie am Werkzeug. Eine Vorschau, die
einen Rumpf mitnimmt — ein Set, eine Liste, ein Ausdruck, eine Form —, ist im
Dienst ein `POST` und über MCP ein `lesen.*`: `POST /einstellungen/vorschau`
kostete 2, `lesen.einstellungen.vorschau` 1. **Einundzwanzig Paare** in
`docs/FAEHIGKEITEN.md` (neunzehn Fähigkeiten); der Wächter fand mit dem
Beschluss, jedes Paar der Tabelle gegen beide Flächen zu halten, **keines
darüber hinaus**. Dieselbe Methodenregel entscheidet am Dienst, ob ein
Zugriffstoken schreiben darf (`rest.schreibend`, `tokenwache.go`).

### Die Entscheidung

1. **Die Kosten kommen aus der Kostentafel, je Fähigkeit** — aus derselben
   Quelle wie bei MCP. Die einundzwanzig Wege stehen dort mit der Klasse
   Lesen (`internal/domain/kostenlesen.go`), je Fähigkeit eine Zeile mit ihren
   Wegen an REST und MCP. Beide Flächen fragen `domain.KostenklasseFuer`; die
   Methode entscheidet nur noch dort, wo die Tafel schweigt.
   **Acht der einundzwanzig Wege sind ein Muster**: der Dienst trägt die
   Rechnungen des Symboleditors als `POST /symbol/{rechnung}` ein, und
   `FAEHIGKEITEN.md` nennt jede mit ihrem Wert. Der Dienst fragt die Tafel
   deshalb mit dem Muster **und** mit dem Pfad der Anfrage
   (`domain.KostenklasseFuerSchreibweisen`, der erste Treffer gilt). Das
   Muster allein fand nur einen Teil — der Lauf `TestJederWegDerTafelBuchtSeineKlasse`
   war vor dieser Änderung bei genau diesen acht rot. Der Handler hinter dem
   Muster ruft ausschließlich `lesen.symbol.*`; eine Rechnung, die die Tafel
   nicht nennt, kostet nach ihrer Methode.
2. **Die Methodenregel für Token bekommt keine zweite Art.** `rest.schreibend`
   bleibt, was sie war: `GET` und `HEAD` lesen, alles andere schreibt. Ein
   Token, das nur lesen darf, kommt an diese Wege weiterhin nicht heran. Die
   Kosten fragen sie nur noch dort, wo die Tafel keine Zeile hat.
3. **Gehalten von einem Test über jedes Paar** (`internal/werkzeuge/
   kostenparitaet_test.go`). Er paart die Wege einer Fähigkeit mit ihren
   Werkzeugen nach der Stelle in der Zelle und fragt beide Seiten an den echten
   Funktionen — `rest.KostenklasseFuer` und `mcpnaehte.KontingentFuer`, mit
   der Art, die das Werkzeug wirklich hat (`schreiben.turnier.vorschlag` heißt
   schreibend und ist lesend). Eine Zeile, die sich nicht paaren lässt, hält
   eine schwächere Zusage: was MCP für die Fähigkeit berechnet, berechnet der
   Dienst für mindestens einen ihrer Wege. **Rot vorher**: genau die
   einundzwanzig Paare, kein anderes.

### Warum die Tafel, obwohl der Token-Vermerk eine Liste harmloser `POST`s ablehnt

`tokenwache.go` begründet die Methode für die Erlaubnis damit, dass eine
gepflegte Liste der harmlosen `POST`s die Eigenschaft hat, **still falsch zu
werden**. Das gilt dort, und es gilt für die Kosten nicht in derselben
Richtung:

- **Für die Erlaubnis** hieße ein falscher Eintrag, dass ein Token, das nur
  lesen darf, einen **schreibenden** Weg ruft. Das ist ein Fehler im Schutz.
- **Für den Preis** heißt ein falscher Eintrag, dass ein schreibender Weg 1
  statt 2 kostet. Das ist ein Fehler in einem Verhältnis, das ohnehin „gewählt,
  nicht gemessen" ist (`domain.Vorgabegrenzen`) — und ein **fehlender** Eintrag
  kostet 2 statt 1, also zu viel und nicht zu wenig.

Dazu hält ein Test die Liste, wo bei der Erlaubnis keiner sie hielte.

### Verworfen

- **Die Methodenregel für Token ändern**, damit ein lesender `POST` auch einem
  lesenden Token offensteht. Das ist eine Entscheidung über Schutz, keine über
  Preise, und sie gehörte in einen eigenen Vermerk. Hier nur der Hinweis, dass
  die Fläche damit **weiter** eine Abweichung hat: über MCP kann ein lesendes
  Profil die Vorschauen rufen, über REST ein lesendes Token nicht.
- **Die Art aus dem MCP-Katalog ableiten.** `internal/rest` kennt nur
  `internal/domain` (§2), und die Zuordnung Weg zu Werkzeug steht in einer
  Tabelle der Dokumentation, nicht in Daten, die der Dienst lesen könnte.
- **Ein Wort im Weg, das „Vorschau" heißt.** Genau das hat #2003 abgeschafft:
  ein Wort im Namen macht nichts billig und nichts teuer.
- **Einen Kopf oder ein Feld, mit dem der Aufrufer sagt, sein `POST` lese.**
  Eine Zusage, die der Rufer macht, ist keine.
- **Die Kommandozeile in die Zeilen aufnehmen.** `mse symbol-rechnen` bedient
  sechs der Fähigkeiten, und der Wächter, der verlangt, dass kein Weg in zwei
  Zeilen steht, ließe das nicht zu — und die Kommandozeile bucht nichts.

## Einordnung (§16)

- **Die Pforte** ist **verträglich**: kein Feld, kein Fehlercode, kein Weg ändert
  sich, und `grenzen` nennt dieselbe Zahl. Ein Ruf, der zeichnet, kann unter
  Last länger dauern, und endet sein Kontext beim Warten, endet er wie bei
  jedem anderen Abbruch.
- **Die Kostentafel in `grenzen`** wächst um neunzehn Zeilen — **erweiternd**.
- **Die einundzwanzig Wege kosten 1 statt 2** — **verträglich**: ein Rufer, der
  vorher durchkam, kommt jetzt auch durch, und mit gleichem Guthaben häufiger.
- **Die Erlaubnis eines Tokens** bleibt, wie sie war.

## Folgen

- Wer eine **lesende `POST`-Vorschau** hinzufügt, trägt sie in
  `internal/domain/kostenlesen.go` ein — der Wächter nennt das Paar.
- `render.Fassung` steigt auf **5**: der Wächter verlangt es bei jeder Änderung
  am Rasterer, auch wenn kein Punkt anders fällt (so schon bei OPS-08). Der
  Preis ist ein Neuzeichnen der Spielbilder je Sitzung.
- Wer eine **neue Art zu rastern** baut, geht durch `Rasterer.Zeichne` und
  bekommt die Pforte. Wer an `Rasterer` vorbei Bildpunkte in der Größe einer
  Karte anlegt, hat die Grenze umgangen, und das ist der Fehler, den dieser
  Vermerk beschreibt.
- `docs/LASTMESSUNG.md` nennt, was nicht wiederholt ist.
