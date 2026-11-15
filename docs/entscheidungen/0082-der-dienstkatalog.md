# 0082 — Der Dienstkatalog: der Vertrag im Quelltext, die Bereitschaft gerechnet

**Stand:** angenommen
**Datum:** 2026-09-09
**Anlass:** Werkstück A1 von Phase 5.3
(#1440) —
die vorhandene Dienst-/Frage-Architektur soll einen **gemeinsamen, belegten
Katalog** liefern: je Fähigkeit den Providervertrag, den Status und den
Nachweis, an allen vier Flächen gleich.

## Lage

**Was am 2026-09-09 galt, ist nachgelesen** (Stand `bb535e26`):

- Vier Dienste mit Schalter (`internal/domain/dienste.go`), eine Frage mit
  Wahl (`fragen.go`: nur die Suche), und die Wahl stand in einer eigenen
  Tabelle `fragenzeilen` neben den Schaltern.
- Die Bedingungen der Dienste standen an **vier Orten**: Vermerk 0056 (die
  drei neueren, Tabelle), Vermerk 0052 (Scryfalls Preise), und je Paket im
  Kommentar an `Adresse` und `Abstand`. Keiner davon war eine Auskunft, die
  eine Fläche geben konnte.
- Das Fenster führte in `Dienstschalter.tsx` eine feste Liste »geprüfter«
  Dienste — zwei Namen, von Hand, und einmal falsch gewesen (#1321,
  Werkstück 8). Ein Wächter hielt die Liste gegen die Dateien hinter dem
  Bautag `netzlauf`; **wann** ein Lauf ging und **gegen welche Fassung**,
  stand nirgends.
- Ob eine Fähigkeit *jetzt* ginge — Schalter an, Zugang da, Akteur darf nach
  draußen, Dienst für die Frage gewählt —, erfuhr man erst am Fehler der
  ersten Anfrage.

## Entscheidung

### 1 · Der Vertrag steht im Quelltext, je Dienst und Frage

`domain.Dienstvertrag` führt zu jeder **gebauten** Fähigkeit — ein Dienst,
eine Frage — Basisadresse, Endpunkte, Zugang, Tarif, Grenze, Frist,
Bedingungen, Abrufdatum und Beleg. Elf Verträge: Scryfall acht (Suche,
Kartendaten, Bild, Regeln, Manakosten, Setangaben, Preise, Sammeldaten), die
drei neueren je die Suche.

**Nur, was gebaut ist.** Ein Anbieter mag mehr anbieten; ein Vertrag steht
hier, weil es einen Klienten gibt, der den Endpunkt ruft, und einen
Verbraucher in `service`. Ein Katalog, der Fähigkeiten nennt, die niemand
rufen kann, wäre der Universalanbieter mit Methoden, die »nicht unterstützt«
melden — genau das, was #1440 ausschließt.

**Und ein Wächter hält ihn gegen die Pakete**
(`internal/werkzeuge/dienstvertraege_test.go`): die Basisadresse ist Zeichen
für Zeichen die `Adresse` des Pakets, jeder Endpunkt steht in dessen
Quelltext, jede Adresse ist `https://`, jedes Datum ist eines. Was er nicht
hält: ob Tarif und Grenze noch stimmen — dafür steht das Abrufdatum daneben.

### 2 · Die Wahl kommt aus dem Katalog

`fragenzeilen` ist weg. Eine Frage hat eine Wahl, sobald **zwei** Dienste
einen Vertrag dafür haben; die Dienste zur Wahl sind die mit Vertrag, in
Katalogreihenfolge, der erste ist die Vorgabe. Heute ist das weiter nur die
Suche — 0063, Entscheidung 1, unverändert, aus einer Quelle statt aus zwei.

### 3 · Die Bereitschaft ist gerechnet, nicht behauptet

`domain.Dienstlage` nimmt die Tatsachen — Schalter, fehlende Zugangsangabe,
Störung, Sperre des Riegels, Wahl — und gibt **das erste Hindernis**: aus,
zugangFehlt, gestoert, gesperrt, nichtGewaehlt, sonst bereit. Die Reihenfolge
ist die, in der man sie beheben würde: ein abgeschalteter Dienst braucht
keinen Schlüssel, ein fehlender Schlüssel keinen Riegel.

**Die Tatsachen sammelt `app`** (`dienstkatalog.go`), **und es schickt
nichts**: kein Klient wird gebaut, keine Adresse gerufen. Der Katalog sagt, ob
eine Anfrage ginge; er versucht es nicht. Ein Test mit vier Spionservern
zählt null Anfragen.

**Ein Hindernis bleibt bei seinem Dienst.** Ein Schlüsselbund, der für Scrydex
nicht antwortet, macht Scrydex gestört und sonst nichts. Den ganzen Katalog
scheitern lässt nur, was für alle gilt: die Einstellungen sind nicht lesbar.

### 4 · Der Nachweis ist ein Lauf mit Datum und Fassung — kein Gütesiegel

`domain.Livenachweis` nennt die Datei hinter `netzlauf`, den Tag des letzten
Laufs, die Fassung, mit der er lief, und die Endpunkte, die er wirklich rief.
Ein Vertrag trägt ihn nur, wenn der Lauf einen seiner Endpunkte ruft: Scryfalls
Bilder und Bulk-Dateien ruft kein Lauf, und ihre Verträge tragen keinen.

**Jede Fläche sagt dazu, ob die Fassung diese ist** (`giltFuerDieseFassung`).
Ein Lauf aus 0.3.0 steht als Lauf aus 0.3.0 da — er ist ein Nachweis über
0.3.0 und die Zusage, dass jemand ihn wiederholen kann. Die feste Liste im
Fenster ist weg; der Wächter `geprueftedienste_test.go` hält jetzt den
Katalog gegen die Läufe, in beide Richtungen, und dazu jede genannte Datei
gegen die Platte.

### 5 · Dieselbe Auskunft an allen vier Flächen

`Dienstkatalog` am Fenster, `lesen.dienstkatalog`, `mse dienstkatalog`,
`GET /dienstkatalog` — dieselbe Naht (`app.DienstkatalognahtUeber`), je Ruf
gebaut, damit ein umgelegter Schalter beim nächsten Ruf gilt. Ein Lauf in
`internal/flaechen` hält die drei kopflosen gegeneinander.

**Im Fenster steht der Katalog unter den Schaltern** — dort, wo man einen
Dienst einschaltet (0056, Entscheidung 4). Die fehlende Zugangsangabe kommt
als Schlüssel (`angabe`) und wird in der Sprache des Fensters benannt; der
Satz aus Go bleibt fürs Protokoll (§9).

## Folgen

- Die Externwirkung steht **neben** der Lesend-/Schreibend-Einordnung der
  Bindungen: `SCHREIBWEGE.md` sagt, ob eine Bindung schreibt; der Katalog
  sagt je Fähigkeit, dass sie nach draußen sendet (`domain.Wirkung`), und ob
  der Akteur das dürfte. Die erste Fähigkeit mit `externSchreiben` kommt mit
  Phase 5.6 und steht dann als Feld im Vertrag, nicht als `if` über den Dienst.
- Wer einen Dienst oder eine Frage aufnimmt, schreibt einen Vertrag — und
  bekommt Schalter, Wahl, Status, Wächter und die Auskunft an vier Flächen
  daraus. Eine zweite Liste daneben ist die Falle, die bei den Einstellungen
  zweimal zugeschnappt ist.
- Ein Nachweis veraltet **sichtbar**: mit jeder Fassung steht »nicht mit
  dieser Fassung« da, bis jemand den Lauf wiederholt und Datum und Fassung
  nachträgt. Das ist Absicht — die Alternative wäre ein Häkchen, das nie
  jemand entfernt.

## Verworfen

- **Den Vertrag nur im Vermerk lassen.** Ein Vermerk ist eine Begründung,
  keine Auskunft; keine Fläche kann ihn lesen, und niemand hält ihn gegen
  die Pakete.
- **Die Bereitschaft durch eine Probeanfrage feststellen.** Sie kostete eine
  Anfrage je Dienst bei jedem Aufschlagen der Seite — von Grenzen, die
  anderen gehören (§16, Regel 4) — und sagte bei einem 429 nichts, was der
  Schalter nicht schon sagt.
- **Ein Nachweis je Dienst statt je Endpunkt.** »Scryfall ist geprüft« hätte
  die Bulk-Dateien mitgemeint, die kein Lauf je gerufen hat. Die Reichweite
  einer Messung ist die Reichweite ihrer Aussage (§10).
- **Den Lauf beim Bauen automatisch fahren, damit der Nachweis nie
  veraltet.** Er befragt fremde Rechner und verbraucht fremde Grenzen; §3 hält
  ihn aus der Pipeline, und das bleibt so. Der Preis ist ein Nachweis, der
  von Hand nachgetragen wird — und ein Wächter, der wenigstens die Datei hält.
