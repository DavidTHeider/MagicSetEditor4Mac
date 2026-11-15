# Beleg: die Anwendung im eigenen Netz unter *Lokales Netzwerk* — K02

**Gefahren am 2026-10-07 abends** mit der Freigabe des Auftraggebers für die
VM-Gäste: zwei Lume-Gäste aus `macos-sequoia-vanilla:latest` (macOS 15.2),
überbrückt im Netz des Hauses — A `192.168.178.96`, B `192.168.178.95`.
**Geprüft wird die gebaute Anwendung** (`scripts/bauen.sh freigabe`, Kennung
`com.wails.MagicSetEditor4Mac`, ad hoc signiert), in A gestartet über
LaunchServices (`open -a`) und bedient über VNC wie von einem Menschen. Als
Gegenüber in B laufen über SSH `mse nachbarn` und der Bibliothekslauf hinter
`vmlauf` — **beide sind von der Erlaubnis ausgenommen** (TN3179:
„Command-line tools run from Terminal or over SSH"), sie messen also nur, was
von A ankommt.

## Warum ein eigener Beleg neben SEC-22

`belege/sec22-vmlauf.md` belegt die **Bibliothek**, gestartet über SSH — und
damit gerade nicht, was ein Mensch erlebt. Der Audit endete an einer
ungeklärten Vertrauensnaht: seine Probe-App bekam `no route` trotz Erlaubnis.

**Nachgelesen in TN3179** („Understanding local network privacy", abgerufen
am 2026-10-07): macOS erkennt ein Programm an seiner Signatur — „ad hoc signed
code" ist ausdrücklich eine Herausforderung — und an der UUID seines
Hauptprogramms: „If your main executable has no UUID, or shares a UUID with
other programs, local network privacy may behave weirdly."

**Gemessen am Bündel**: das Hauptprogramm hat eine UUID (`otool -l`),
`NSLocalNetworkUsageDescription` steht in der `Info.plist`, die festgelegte
Anforderung ist `cdhash H"…"` und nichts sonst — jede Fassung ist für macOS
ein anderes Programm.

## Die Fassungen

| | Herkunft | cdhash | UUID des Hauptprogramms |
| --- | --- | --- | --- |
| f1 | `main` (`3792ce97f`) | `53620e3c…` | `65B2627C-…` |
| f2 | Zweig nach K01 | `10f3a4d0…` | `7B41740A-…` |
| Kopie von f2, Kennung `…neinprobe` | f2, neu signiert | neu | **`7B41740A-…`, dieselbe** |
| f3 | Zweig mit der Korrektur | neu | `2E3707A2-…` |
| Kopie von f3, Kennung `…nein3` | f3, neu signiert | neu | **`2E3707A2-…`, dieselbe** |
| Kopie von f3, Kennung `…erstprobe` | f3 mit neuer `LC_UUID`, neu signiert | neu | `937B55C0-…`, neu |

## Was geschah

| Fall | Was macOS tat | Was die Anwendung zeigte | Was B sah |
| --- | --- | --- | --- |
| **f1, erster Start**, Ankündigen | fragte — „Allow "MagicSetEditor4Mac" to find devices on local networks?" mit unserem Erklärtext; **Allow** | „announcing" | `VM-A-App` mit A's Fingerabdruck — **und A fand B** (`VM-b`) |
| **f2 als Update** am selben Ort | **fragte nicht** | „announcing" | `VM-A-App2` — die Erlaubnis überstand das Update |
| **Kopie mit derselben UUID**, andere Kennung | **fragte nicht und ließ nichts hinaus** | vorher: **„announcing"** — eine Zusage ohne Wirkung; die Suche meldete richtig „did not get out" | nichts |
| f3, Update | fragte nicht | „announcing" | `VM-A-f3` — die Probe gibt keinen Fehlalarm |
| f3-Kopie mit derselben UUID | fragte nicht | **jetzt: „Nothing got out onto the local network …"**, „not announcing" | nichts |
| f3-Kopie mit neuer UUID, **erster Start** | fragte | beim ersten Klick schon die Meldung, weil die Probe vor der Antwort scheitert; nach **Allow** und einem zweiten Klick „announcing" | `VM-A-Erst` |

## Was daraus folgt

- **Die Annahme, ein Update verliere unter Ad-hoc-Signatur die Erlaubnis, ist
  widerlegt** — für macOS 15.2, ein Update am selben Ort und dieselbe Kennung.
- **Die stille Absage gibt es**, und zwar dann, wenn ein Programm die UUID
  eines anderen trägt: macOS fragt nicht und lässt nichts hinaus. Das ist die
  Klasse, vor der TN3179 warnt — und **plausibel**, nicht belegt, die Ursache
  des `no route` der Probe-App im Audit.
- **Die Anwendung sagte dabei die Unwahrheit**: der mDNS-Server bindet seine
  Sockets auch ohne Erlaubnis, gesendet wird erst als Antwort, und deren
  Fehler verschluckt `hashicorp/mdns`. **Behoben in K02**: `entdeckung.Melde`
  schickt nach dem Aufstellen eine Probe an die mDNS-Gruppe
  (`internal/entdeckung/hinaus.go`) und meldet ohne Erfolg
  `ErrOhneNetzerlaubnis` — so, wie Apple es selbst nahelegt („connect a UDP
  socket to a local network address").
- **Der Satz dazu** heißt jetzt „Ins lokale Netz ging nichts hinaus" — er galt
  bisher nur der Suche — und sagt, was nach der Rückfrage zu tun ist. Der
  englische behauptete „Usually the permission is missing", der deutsche nennt
  beide Lesarten; jetzt sagen alle acht dasselbe.
- **Und ein zweiter Befund am Fenster**: das Feld „Address to listen on"
  trug `:41234` nur als Platzhalter, und „Become reachable" blieb ohne
  Erklärung angehalten. **Behoben**: das Feld beginnt mit `:41234` als Wert;
  gesehen in f3 — erreichbar ohne zu tippen.

## Was dieser Beleg nicht sagt

- **„Don't Allow" von Hand.** Eine verweigerte Erlaubnis lässt sich unter
  macOS 15 nicht zurücksetzen (TN3179: die Knöpfe dafür kamen mit 27.2); die
  Absage ist am Fall „dieselbe UUID" gemessen, den macOS still verweigert.
- **macOS 26 und 27** — die Gäste laufen mit 15.2, dem Stand des Audits.
- **Eine gesicherte Verbindung aus der Anwendung** zu B: gefunden haben sich
  beide; das Wählen über TLS ist als Bibliothekslauf in `sec22-vmlauf.md`
  belegt, nicht als Bedienung am Fenster.
- **Die Systemeinstellungen** zeigen die Liste *Lokales Netzwerk* hier nicht
  an — beide Direktadressen führten nur bis *Datenschutz & Sicherheit*.
