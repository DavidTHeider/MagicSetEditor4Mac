# 0172 — Der Schluss hat eine Reihenfolge und eine Frist

**Stand:** angenommen am 2026-10-06 · BK-07a
(#2115),
BK-02 (#2110).
**Ergänzt:** [0170](0170-der-rufkontext-endet-mit-dem-ruf.md),
[0171](0171-ein-nebenlauf-faengt-seinen-zusammenbruch.md).

## Die Frage

**Die Anwendung endete an drei Stellen verschieden.** `App.Beenden` beendete
Dienstposten und Gerätemodell und schloss den Bestand; was sonst lief —
Horchposten, Peersitzungen, der Tisch als Gast, die Sitzungen der Gäste beim
Wirt —, lief über einem geschlossenen Bestand weiter. Der Haken der Hülle bei
⌘Q merkte nur die Fenstergröße, **ohne Frist**, und `App.Beenden` lief dort
gar nicht, weil ⌘Q nicht aus `Run` zurückkehrt. Und das Gerätemodell hielt
beim Stoppen die Sperre seines Starts, bis zu zehn Sekunden.

Gemessen am 2026-10-06, vor der Änderung:

| Fall | Wirkung |
| --- | --- |
| `Beenden` beim Wirt, ein Gast sitzt (`schluss_intern_test.go`) | die Gastsitzung las danach den geschlossenen Bestand: „die Partie lesen: sql: database is closed" |
| Merken der Fenstergröße bei gehaltenem Pool **und** gehaltenem Profilschloss (`schlussfrist_intern_test.go`) | kehrt nicht zurück — das Fenster bliebe offen |
| Stop des Gerätemodells während eines Starts, der auf `/health` wartet (`applefm/dienststop_test.go`) | wartete 10,02 s auf die Startfrist |

## Die Entscheidung

**Ein Schluss, eine Reihenfolge, eine Gesamtfrist** (`App.Schliessen`,
`lebenszyklus.go`; `Beenden` ist dasselbe mit [Schlussfrist] = 5 s):

1. **Das Tor geht zu** — jede Bindung danach endet mit `abgebrochen`, bevor
   sie etwas anfasst.
2. **Bestenfalls** — was beim Schließen gemerkt werden soll (die
   Fenstergröße), in **höchstens einer Sekunde** der Frist. Ein Verlust
   kostet Komfort und wird im Protokoll genannt. Gemessen an der
   Hüllenprobe `bk07a`: mit der ganzen Frist nahm ein Merken, das an einem
   gesperrten Bestand hing, den Läufen danach ihre Zeit.
3. **Die eigenen Läufe enden** — der Tisch als Gast, dann dieselben Schritte
   wie beim Wechsel des Profils: KI-Auftragswerk, Gerätemodell,
   Peersitzungen, Horchen, Fernhorchen, Spielbilder, Vorschauen,
   Dienstposten, Teilungen.
4. **Gewartet wird** auf die gezählten Läufe: den Tisch des Gastes und jede
   Sitzung eines Gastes beim Wirt (`schlusslage.kinder`).
5. **Zuletzt der Bestand.**

**Läuft die Frist ab, geht der Bestand trotzdem zu.** Was dann noch läuft,
bekommt keine neue Verbindung mehr — `sql: database is closed` —, und
`Schliessen` gibt `ErrSchlussNachFrist` mit dem zurück, was nicht fertig
wurde. Ein zweiter Schluss findet nichts mehr vor.

**Was das nicht zusagt** (Zweitblick über A und B, Befund 3): eine
Transaktion, die beim Schließen schon offen ist, behält ihre Verbindung —
`database/sql` entzieht sie nicht — und kann noch `INSERT` und `COMMIT`.
Sie schreibt dann ganz oder gar nicht, wie jede Transaktion; einen halben
Stand gibt es nicht. Hier stand bis dahin „ein Zustand, in dem nichts mehr
geschrieben wird", und das war mehr, als `DB.Close` hält.

**Ein gescheiterter Start ist kein Schluss**: `abbrechen` räumt ab und lässt
das Tor offen, damit Startbefund und Rückweg erreichbar bleiben (#962).

**Die Hülle schließt mit derselben Methode**: `OnShutdown` ruft
`Schliessen` mit der Fenstergröße als Bestenfalls-Schritt; der Haken bei
`WindowClosing` merkt die Größe über `app.Bestenfalls` in derselben Frist.

**Und ein Stop bricht einen Start ab, statt auf ihn zu warten**
(`applefm.Dienst`): `Beende` und `Abschalten` brechen den Kontext eines
laufenden Starts ab, bevor sie die Sperre nehmen; der Start beendet seinen
Prozess selbst. **Das ändert eine Zusage** (§15, Satz 4): bis hierher lief ein
Start, in den „aus" fiel, zu Ende und gab einen Klienten zurück, dessen
Prozess gleich danach endete — jetzt bekommt er den Abbruch
(`dienstzustand_test.go`, Z2).

## Wem welcher Lauf gehört

| Lauf | Besitzer | Endet durch | Gezählt |
| --- | --- | --- | --- |
| Tisch als Gast (`amFremdenTisch`) | die Anwendung | `partiebeitrittBeenden`, Ende der Leitung | ja |
| Sitzung eines Gastes beim Wirt (`partieBedienen`) | der Horchposten, der sie annahm | Peersitzungen pausieren, Ende der Leitung | ja |
| Horchposten, Fernhorchen | die Anwendung | `horchenBeenden`, `fernhorchenBeenden` — sie warten selbst | über `fertig` |
| REST- und MCP-Dienst | die Anwendung | `diensteBeenden` — Kindprozess | über den Prozess |
| Gerätemodell | der Prozess | `geraetBeenden` | über den Prozess |
| Aufträge des Fensters, KI-Rufe | der Ruf | der Rufkontext (0170) | — |
| Aufträge von `mse-rest` | das Auftragsbuch | `Dienst.AuftraegeSchliessen` (BK-07b) | im Buch |

## Was dieser Vermerk nicht zusagt

- **Beim Wechsel des Profils wird nicht gewartet.** Er läuft unter
  `umschaltung.Lock`; ein Lauf, der zum Ende noch `imProfil` bräuchte,
  wartete auf dieselbe Sperre. Beendet wird dort wie bisher — abgebrochen
  und geschlossen, nicht abgewartet.
- **Ein Bestand wird im Lauf nicht ersetzt.** Das Zurückspielen einer
  Sicherung legt eine Marke, und getauscht wird beim nächsten Start, bevor
  ein Lauf beginnt (`rueckspielmarke.go`).

## Verworfen

- **Eine Frist um jedes SQL.** Sie begrenzt weder `sync.RWMutex` noch die
  Sperre des Gerätemodells — genau dort hing es.
- **Den Bestand erst schließen, wenn alles fertig ist.** Dann hielte ein
  einziger tauber Lauf das Fenster offen, das jemand schließen wollte.
