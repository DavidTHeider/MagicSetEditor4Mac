# Windows und Linux

**Diese Plattformen werden beschrieben, nicht gebaut.** Gebaut, geprüft und
ausgeliefert wird ausschließlich macOS auf Apple Silicon (`AGENTS.md` §12).

Diese Datei sagt, was zu tun wäre, wenn sich das ändert — und wo die Fallen
liegen. Sie ist eine Beschreibung, keine Anleitung, die jemand erprobt hat. Wo
etwas ungeprüft ist, steht es dabei.

---

## Was schon heute für beide Plattformen geschieht

Die Pipeline **übersetzt** für Windows und Linux mit, ohne dort zu laufen:

```bash
GOOS=windows GOARCH=amd64 go build ./internal/... .
GOOS=windows GOARCH=arm64 go build ./internal/... .
GOOS=linux   GOARCH=amd64 go build ./internal/... .
```

Und `golangci-lint` läuft über alle drei Ziele, damit Dateien hinter Build-Tags
nicht ungelintet bleiben. Der Grund: `golangci-lint` sieht nur, was für das
aktuelle `GOOS` übersetzt wird. Ohne diesen Schritt entzieht sich jede
`_windows.go` der Prüfung, und niemand merkt es, weil lokal alles grün ist.

**Ein Bau ist keine Erprobung.** Er sagt, dass der Code dort entsteht — nicht,
dass er dort etwas Richtiges tut.

Möglich ist das nur, weil der SQLite-Treiber (`modernc.org/sqlite`) ohne cgo
auskommt. Mit cgo bräuchte jede Zielplattform eine eigene Werkzeugkette, und
dieser Abschnitt gäbe es nicht.

---

## Windows

### Was zu tun wäre

| Schritt | Bemerkung |
| --- | --- |
| WebView2 | Wails 3 setzt unter Windows auf Microsoft Edge WebView2 (`go-webview2`, ohne cgo). Ab Windows 11 vorinstalliert; für Windows 10 gehört der Bootstrapper in den Installer |
| Bau | `GOOS=windows GOARCH=amd64 go build -tags production -ldflags "-H windowsgui" .` bzw. `arm64` — ohne cgo, deshalb übersetzt die Pipeline es schon heute mit. **Ein Bündelskript wie `scripts/bauen.sh` gibt es dafür nicht** |
| Icon und Manifest | als Ressource ins Programm (`.syso`, etwa über `go-winres`); unter Wails 2 legte `wails build` das aus `build/windows/` an, und diesen Weg gibt es seit ENG-02 nicht mehr |
| Installer | NSIS, wie es die Taskfile-Kette von Wails 3 anbietet — die dieses Projekt nicht benutzt (0132). **Ungeprüft**; in einem Schwesterprojekt scheiterte der NSIS-Lauf reproduzierbar mit `std::bad_alloc`, und ausgeliefert wurden die `.exe` einzeln |
| Signatur | Ohne Code-Signing-Zertifikat zeigt SmartScreen eine Warnung. Es gibt kein Zertifikat und es ist keines vorgesehen |

### Wo es klemmt

- **Dateirechte.** `0600` und `0700` bewirken unter Windows nichts. Alles, was
  Zugriffsschutz zusagt, gehört hinter `//go:build windows` mit einer eigenen
  Fassung über ACLs — und die Tests dazu gehören dorthin mit, statt zu fehlen.
  **Umgekehrt gilt es auch**: die Prüfung des Dienstzeugnisses weist einen
  Schlüssel ab, der für andere lesbar ist (`internal/leitung/eigeneszeugnis.go`,
  NET-03). Unter Windows gäbe `Mode().Perm()` dafür keine brauchbare Auskunft;
  die Prüfung gehörte dort über die ACL der Datei.
- **Pfade.** MSE-Pakete tragen Pfade mit `/` in ihren Verweisen
  (`/magic-modules.mse-include/cards/…`). Das sind **Paketpfade**, keine
  Dateipfade — sie dürfen nie durch `filepath.Join` laufen, sondern durch
  `path.Join`. Unter macOS fällt der Unterschied nicht auf; unter Windows sofort.
- **Schriften.** Die Kartenausgabe sucht Schriften im System und in den
  bezogenen Paketen (#140). Der Suchweg unter
  Windows ist ein anderer, und die Namen sind es teils auch.
- **Festschreiben** (BK-06b, Vermerk 0174). Unter macOS und Linux gibt es
  zwei Stufen — `fsync` für Ausgaben, `F_FULLFSYNC` (`(*os.File).Sync`) für
  Marken des Wiederanlaufs. Unter Windows gibt es nur `FlushFileBuffers`;
  `internal/service/festschreiben_andere.go` nimmt es für beide Stufen.
  **Und ein Verzeichnis lässt sich dort nicht öffnen, um es festzuschreiben**
  — `verzeichnisFestschreiben` scheiterte bei jeder Ausgabe. Vor einem
  Windows-Bau gehört es dort abgeschaltet, mit einem eigenen Test.
- **Zeilenenden.** MSE-Dateien tragen gemischte Zeilenenden. Der Zerteiler
  normalisiert beim Lesen und schreibt zurück, was ursprünglich dastand — sonst
  bricht die Rundlaufzusage aus `ARCHITEKTUR.md`.

---

## Linux

### Was zu tun wäre

| Schritt | Bemerkung |
| --- | --- |
| Abhängigkeiten | Wails 3 setzt auf GTK 4 und WebKitGTK 6.0 (`pkg-config: gtk4 webkitgtk-6.0`); mit dem Bauschalter `gtk3` auf GTK 3 und WebKit2GTK 4.1 (`gtk+-3.0 webkit2gtk-4.1`) — nachgelesen in `linux_cgo.go` und `linux_cgo_gtk3.go`, v3.0.0-beta.23 |
| Bau | `CGO_ENABLED=1 go build -tags production .` auf einer Linux-Maschine — **nur mit cgo**, und deshalb übersetzt die Pipeline Programm und Hülle für Linux nicht mit (0132) |
| Paketierung | `.deb` über ein eigenes Skript, `.AppImage` als Alternative. Beides ungeprüft |
| Desktop-Eintrag | `.desktop`-Datei und Icon nach Freedesktop-Konvention |

### Wo es klemmt

- **GTK- und WebKitGTK-Version.** Der Wechsel von WebKit2GTK `4.0` auf `4.1`
  hat schon Distributionsstände auseinandergerissen, und Wails 3 geht einen
  Schritt weiter: **ohne Schalter GTK 4 und WebKitGTK 6.0**, mit `gtk3` das
  Paar, das Wails 2 mit `webkit2_41` wählte.

  **Und der Bau muss es wissen.** Wails wählt die ABI über einen Build-Tag,
  nicht über das, was installiert ist. Eine Distribution ohne WebKitGTK 6.0
  braucht `-tags gtk3`, sonst fehlt die Bibliothek schon beim Bau. Unter
  Wails 2 stand hier derselbe Fehler mit `webkit2_41` — gefunden vom
  Zweitblick über Phase 8.
- **Schriften.** Über fontconfig, wieder ein anderer Weg als unter macOS.
- **Dunkles Erscheinungsbild.** Wird über die GTK-Einstellung gemeldet, nicht wie
  unter macOS über `NSApplication`.

---

## Die Schriftsuche ist gekapselt

Seit Phase 4 hat der Renderer dafür eine Naht: `render.Schriftquelle` mit einer
einzigen Methode, `Suche(name) (Schrift, error)`. **Plattformabhängig ist die
Umsetzung, nicht der Aufrufer** — `internal/render` weiß nicht, wo Schriften
liegen, und soll es nicht wissen.

| Plattform | Weg |
| --- | --- |
| **macOS** | `~/Library/Fonts`, `/Library/Fonts`, `/System/Library/Fonts`. Core Text kennt mehr, aber der Weg dorthin führt über cgo, und cgo beendet die Kreuzkompilierung (§12) |
| **Windows** | `%WINDIR%\Fonts` und die Registrierung unter `Software\Microsoft\Windows NT\CurrentVersion\Fonts`; dazu Schriften je Benutzer seit Windows 10 |
| **Linux** | fontconfig. Die Verzeichnisse allein reichen nicht, weil Verweise und Ersetzungsregeln dort einen Teil der Antwort ausmachen |

**Eine fehlende Schrift wird benannt, nicht ersetzt.** Das gilt auf jeder
Plattform und ist keine Frage der Umsetzung: eine Karte mit falscher Schrift
sieht richtig aus und ist es nicht. Wer sie mit `Helvetica` statt `MPlantin`
gedruckt bekommt, merkt es an dem Tag, an dem die Karten aus der Druckerei
kommen.

### Was seit Phase 6 gebaut ist

`internal/schriften` ist die Plattformseite dieser Naht. Sie geht die
Verzeichnisse durch und vergleicht den **Dateinamen ohne Endung**, ohne
Rücksicht auf Groß- und Kleinschreibung. Gelesen werden `.ttf`, `.otf` und
`.ttc`; aus einer Sammlung kommt die **erste** Schnittform.

Die Orte stehen in `orte_darwin.go` und `orte_andere.go` — getrennt über
Build-Tags, damit der Bau für Windows und Linux nicht bricht. **Beschrieben,
nicht erprobt** (§12): die Liste für die anderen Plattformen ist keine Zusage,
dass die Suche dort trägt.

### Seit #140: die Schriften bezogener Pakete

Ein bezogenes Paket bringt seine Schriften **mit** — `M15-Magic-Pack` führt sie
unter `Magic - Fonts/`, also *neben* den Paketen und nicht in ihnen. Der
Suchweg bekommt deshalb die Paketwurzel dazu
(`~/Library/Application Support/MagicSetEditor4Mac/pakete`), und darin wird der
**ganze Baum** durchsucht: ein fester Unterverzeichnisname wäre eine Annahme
über fremde Sammlungen.

**Die Reihenfolge ist System vor Paket.** Was jemand selbst installiert hat, ist
die Schrift, die er meint; eine Paketschrift gleichen Namens ersetzte sie sonst,
ohne dass jemand es angeordnet hätte. Die Reihenfolge steht nicht in einem
Kommentar, sondern in der Bauweise — und `TestDasSystemGehtVorDemPaket` hält sie
fest.

**Die Schriften bleiben, wo sie liegen.** Sie ins System zu kopieren wäre eine
Installation, die niemand angeordnet hat, und sie überlebte die Deinstallation
der Anwendung.

**Für Windows und Linux ändert das nichts an der Lage:** die Paketwurzel liegt
dort unter dem jeweiligen Konfigurationsverzeichnis (`os.UserConfigDir`), und
der Rest der Tabelle oben — Registrierung, fontconfig — bleibt ungebaut.

### Vier benannte Lücken

| Lücke | Was sie kostet |
| --- | --- |
| **`.dfont`** wird nicht gelesen | ein Teil des macOS-Grundbestands liegt so; ein Ressourcenbündel aus der Zeit vor Mac OS X braucht einen eigenen Weg |
| **Der Dateiname ist der Schriftname** | eine Schrift, deren Datei anders heißt als sie selbst, wird nicht gefunden. Der Weg dahin führt über die Namenstabelle der Schriftdatei |
| **Nur die erste Schnittform einer Sammlung** | `Helvetica Neue Bold` liegt in derselben `.ttc` und ist von hier aus nicht erreichbar |
| **Manche Sammlungen gar nicht** | `x/image` weist `HelveticaNeue.ttc` mit „unsupported cmap encodings" ab — nachgesehen, nicht vermutet, und als Testfall festgehalten |

Auf **Windows** und **Linux** kommt hinzu, was oben in der Tabelle steht: die
Registrierung beziehungsweise fontconfig. Beides ist nicht gebaut.

Das **Rastern** der Glyphen ist entschieden und gebaut:
`golang.org/x/image` ([0008](entscheidungen/0008-schriftrasterung.md)). Der Weg
über Core Text wäre auf macOS der bessere und beendete die Kreuzkompilierung —
deshalb steht er im Vermerk als das, was verworfen wurde.

---

## Was die Phasen 5 bis 9 dazugelegt haben

Nachgetragen zum Abschluss von Phase 8
(#63). Bis dahin
führte diese Datei nur die Schriftsuche, weil sie die einzige Naht mit
Build-Tags ist. Plattformabhängig ist aber auch, was **ohne** Build-Tag
übersetzt und dann anderswo etwas anderes tut — und das ist die gefährlichere
Sorte: `go build GOOS=linux` bleibt grün.

### Die Saat kommt nur im macOS-Bündel an

Seit Phase 7 liegt eine kleine, selbst erstellte Saat aus Vorlagen **im
Programmbündel** (#58).
Der ganze Weg dorthin ist macOS:

| Stück | Warum es nur macOS ist |
| --- | --- |
| `scripts/bauen.sh` | Legt ein `.app` und ruft `saat.sh` darin. Für Windows und Linux gibt es kein Bauskript (ENG-02; unter Wails 2 war es ein `postBuildHook` nur für `darwin/*`) |
| `scripts/saat.sh` | Legt nach `Contents/Resources/saat` und **signiert das Bündel neu** (`codesign`). Beides gibt es außerhalb von macOS nicht |
| `app.Saatverzeichnis` | Rechnet `…/Contents/MacOS/Programm` → `…/Contents/Resources/saat` |

**Die Folge ist kein Baufehler, sondern eine stille Lücke:** außerhalb eines
`.app` findet `Saatverzeichnis` nichts und gibt einen leeren Pfad zurück. Die
Anwendung startet, und es ist einfach keine Vorlage da. Wer portiert, braucht
je Plattform einen Ort neben dem Programm (Windows: das Verzeichnis der `.exe`;
Linux: `/usr/share/magicseteditor4mac` oder das AppImage-Wurzelverzeichnis),
einen eigenen Bauweg, der die Saat dorthin legt — und **kein** `codesign`.

### Groß- und Kleinschreibung im Dateisystem

`msepaket` liest eine Paketdatei über ihren **genauen** Namen
(`internal/msepaket/bildquelle.go`, `paketdatei`). Ein Skript in einer Vorlage
schreibt `image: "mana_w.png"`, in der Datei steht `Mana_W.png` — auf einem
Mac in der Vorgabe (APFS, ohne Beachtung der Schreibweise) fällt das nie auf,
unter Linux ist es ein fehlendes Bild.

**Erschlossen aus dem Code, nicht unter Linux nachgemessen** (§12). Es steht
hier trotzdem, weil es die teuerste Sorte Befund ist: die Vorlage ist in
Ordnung, das Programm ist in Ordnung, und die Karte bleibt leer.

Es trifft im Übrigen auch macOS — auf einem Volume, das mit Beachtung der
Schreibweise formatiert ist. Wer das beheben will, behebt es an **einer**
Stelle: in `paketdatei`, nicht bei jedem Aufrufer.

### Wo der Bestand liegt

`app.Datenbankpfad` und `app.Paketverzeichnis` gehen über `os.UserConfigDir()`.
Das ist auf jeder Plattform übersetzbar und meint überall etwas anderes:

| Plattform | `os.UserConfigDir()` | Anmerkung |
| --- | --- | --- |
| **macOS** | `~/Library/Application Support` | richtig für Bestand |
| **Windows** | `%AppData%` (roaming) | Ein Set-Bestand und ein Paketbaum gehören eher nach `%LocalAppData%` — sonst wandern sie durch ein Roaming-Profil |
| **Linux** | `$XDG_CONFIG_HOME`, sonst `~/.config` | **Falsch nach XDG.** Bestand gehört nach `~/.local/share`; `~/.config` ist für Konfiguration |

Beide Anmerkungen sind **ungeprüft** — beschrieben aus der Festlegung der
jeweiligen Plattform, nicht aus einem Lauf.

### Wo MSE2 vermutet wird

`app.MSE2Orte` rät fünf Orte. Drei davon sind macOS (CrossOver-Bottles,
`/Applications`), zwei sind Wine unter dem Heimatverzeichnis. Unter **Windows**
läge MSE2 schlicht in `%ProgramFiles(x86)%\Magic Set Editor 2\data` — der Pfad
steht nicht in der Liste. Das ist kein Fehler: die Liste sind Vermutungen, und
daneben steht immer „Verzeichnis wählen …". Wer portiert, ergänzt sie.

### Die Tastenkürzel sind ⌘, nicht Strg

`frontend/src/rahmen/Rahmen.tsx` liest `ereignis.metaKey`. Unter Windows und
Linux ist die gewohnte Taste `Ctrl` (`ctrlKey`), und ⌘ gibt es dort nicht. Die
Kürzel wären also nicht bloß anders beschriftet, sondern gar nicht erreichbar.
Dass sie an **einer** Stelle gelesen werden, ist die gute Nachricht: eine
Zuordnung je Plattform genügt.

Dasselbe gilt für die native Menüleiste und für `⌘,`/`⌘W`/`⌘Q` (§6): sie
bedienen macOS-Gewohnheiten absichtlich, und auf einer anderen Plattform wären
es andere.

### Apples Gerätemodell gibt es nur auf macOS 27

**`/usr/bin/fm` ist ein Werkzeug von macOS** (AI-05,
#2042). Auf
Windows und Linux gibt es es nicht; die Erkennung meldet dort „fehlt", der
Vorrang überspringt das Gerät, und kleine Textaufgaben gehen an einen lokalen
Anbieter wie Ollama. **Nichts muss dafür gebaut werden** — der Unix-Socket aus
`internal/netz` übersetzt auf beiden, und die Stelle, die `fm` startet, prüft
die Datei, bevor sie etwas ruft.

### Was **nicht** plattformabhängig ist

Damit die Liste oben nicht länger aussieht, als sie ist:

- `internal/pdf` schreibt Bytes und kennt keine Plattform. **Drucken** ist eine
  andere Frage — der Bogen wird als Datei abgelegt, und wer ihn druckt, benutzt
  das Programm seiner Plattform.
- `internal/msescript`, `internal/domain`, `internal/mtgjson` sind reine
  Sprach- und Datenarbeit.
- `internal/bezug` benutzt `net/http` und `archive/zip`; beides trägt überall.
  Die Ausbruchsprüfung beim Auspacken rechnet mit `filepath` und ist damit auf
  jeder Plattform die richtige.
- SQLite läuft ohne cgo (`modernc.org/sqlite`) — der Grund, warum dieser
  Abschnitt überhaupt kurz bleiben kann.

---

## Die Dateisperre fehlt auf Windows

**Seit Phase 3.2** gibt es `internal/dateisperre`: eine beratende Sperre auf
einer Nachbardatei, über `flock`
(#1163).
Zwei Wege benutzen sie:

| Wer | Wofür |
| --- | --- |
| `internal/storage` | hält den Bestand über einen Tausch — sonst kann ein zweiter Prozess in dem Augenblick öffnen, in dem die Datei umbenannt wird |
| `internal/netz` | teilt das Hostbudget einer Gegenstelle über Prozessgrenzen — sonst hat jeder Prozess das volle Budget |

**Auf Linux gilt sie unverändert**: `flock` ist dort dasselbe wie auf macOS.

**Auf Windows gibt es sie nicht.** `dateisperre_andere.go` tut nichts und
gelingt. Damit steht dort der Stand vor Phase 3.2:

- Das Fenster zwischen der Alleinzugriffsprüfung und dem Tausch bleibt offen.
- Fenster, MCP-Server und Dienst drosseln je für sich.

**Was zu tun wäre**: `LockFileEx` mit `LOCKFILE_EXCLUSIVE_LOCK` und
`LOCKFILE_FAIL_IMMEDIATELY` über `golang.org/x/sys/windows`. Das ist eine neue
Abhängigkeit, und sie käme für eine Plattform, die dieses Projekt beschreibt
und nicht baut — deshalb steht sie hier und nicht in `go.mod`.

**Ungeprüft**: ob `LockFileEx` auf einem Netzlaufwerk dasselbe zusagt wie
`flock`. Wer es baut, misst es nach.

---

## Das Protokoll — und warum es auf Linux anders aussähe

**Beschrieben, nicht gebaut** (§12) —
#848.

Auf macOS geht das Protokoll in zwei Senken: eine **Datei** neben dem Bestand
und **Apples Unified Log**. Der zweite Weg führt über `/usr/bin/logger`, weil
`os_log` eine C-Schnittstelle ist und cgo hier ausgeschlossen bleibt
([0002](entscheidungen/0002-sqlite-als-arbeitsformat.md)).

**Auf Linux gibt es das Unified Log nicht.** Der entsprechende Weg wäre
`systemd-cat` oder ein Schreiben nach `/dev/log`; beides ist dieselbe Bauform
wie hier — ein `io.Writer` hinter `systemprotokoll()`, gewählt über einen
Build-Tag.

**Seit OPS-04 ist der Kanal abschaltbar** (`protokoll.system`), und auf
einer Plattform ohne Weg dorthin schreibt er nichts — die Diagnose nennt ihn
dann abgewählt oder fehlend. **journald** wäre auf Linux der Kanal an dieser
Stelle, etwa über `systemd-cat` oder das Journal-Protokoll; beschrieben,
nicht gebaut (§12).

**Heute gibt `systemprotokoll()` dort `io.Discard`**, und die Datei schreibt
weiter. Das ist die richtige Antwort für eine Plattform, die nicht gebaut,
nicht geprüft und nicht ausgeliefert wird: keine halbe Anbindung, die
aussieht, als sei sie erprobt.

**Was zu tun wäre**, wenn jemand Linux baute:

1. `internal/app/protokoll_linux.go` mit `systemprotokoll()` über `systemd-cat`
   — dieselbe Gestalt wie `protokoll_darwin.go`.
2. Ein Test, der den Satz **zurückliest** — auf macOS tut das
   `protokoll_darwin_test.go` über `log show`; auf Linux wäre es `journalctl`.
   Ohne den Rückweg ist die Anbindung eine Behauptung.
3. Die Kosten messen. Auf macOS sind es **2,7 ms je Zeile** für einen Prozess;
   `systemd-cat` dürfte ähnlich liegen, und ob das reicht, entscheidet, wie oft
   protokolliert wird.

## Was gälte, wenn eine dieser Plattformen dazukäme

Nicht „einmal bauen und hochladen". Es käme hinzu:

1. Ein **Läufer** in der Pipeline, der dort wirklich testet — nicht nur
   übersetzt. Ohne das ist eine ausgelieferte Binärdatei eine Behauptung.
2. Ein **Rauchtest**: startet die Anwendung, legt sie ihre Datenbank an, endet
   sie sauber?
3. Ein Abschnitt in `docs/RELEASE.md` und ein Ziel in `scripts/release.sh`.
4. Ein Eintrag in `README.md` — und zwar mit dem Hinweis, was auf dieser
   Plattform **nicht** geprüft ist.

Bis dahin gilt dieser Satz: Windows und Linux sind hier beschrieben, damit
niemand bei null anfängt. Sie sind nicht unterstützt.
