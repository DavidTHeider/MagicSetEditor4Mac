# Administration

**Wo die Daten liegen, wie man sie sichert, und was bei einem
Fassungswechsel geschieht.**

Für die Bedienung steht [`HANDBUCH.md`](HANDBUCH.md), für Einrichtung und Bau
[`README.md`](../README.md).

---

## Wo alles liegt

Alles, was die Anwendung selbst führt, liegt an **einer** Stelle:

```
~/Library/Application Support/MagicSetEditor4Mac/
├── bestand.db      die Datenbank — Sets, Karten, Felder, Verlauf
│                   (verschlüsselt, wenn eingeschaltet — siehe unten)
├── bestand.db-wal  } SQLite im WAL-Modus: zwei Beiwerksdateien, die
├── bestand.db-shm  } zum Bestand gehören
├── bestand.db.sperre  die Nachbardatei, über die sich zwei Prozesse verständigen
├── bestand.db.wanderung  die Nachbardatei, unter der gesichert und gewandert wird
├── bestand.db.eintritt  die Schranke davor: eine wartende Wanderung lässt keinen neuen Ruf vorbei
├── pakete/         die bezogenen Vorlagenpakete
├── beiwerk/        eingesetzte Bilder
├── mse.log         das Protokoll
├── mse.log.1       seine Vorfassung — sie entsteht bei 4 MB
├── mse.log.sperre  } zwei leere Nachbardateien: wer das Protokoll hält,
├── mse.log.drehen  } und unter welcher Sperre es gedreht wird (PERF03)
└── Sitzungen/      die Startmarke des laufenden Fensters samt ihrer Sperre
                    und höchstens zehn Absturzspuren (F04, #2343)
```

**Eine Startmarke, die beim nächsten Start noch liegt, ist ein Abbruch.**
Das Fenster legt `Sitzungen/fenster-<pid>.json` beim Start und nimmt sie beim
sauberen Schluss wieder weg; daneben hält es `fenster-<pid>.sperre` gesperrt
(`flock`) — das Betriebssystem gibt die Sperre mit dem Prozess frei, auch
nach SIGKILL —, und eine Datei `absturz-<zeit>-<pid>.txt`, in die Go bei
einem Absturz seinen Stapel schreibt. Findet der nächste Start eine Marke,
deren Sperre niemand hält, sagt das Fenster es in einem Band — **gleich, ob
inzwischen ein anderer Prozess dieselbe Nummer trägt** (Zweitblick über
Phase 1 von 0.9.0, Befund 11). **Lesen, Prüfen, Räumen und Legen geschehen
unter einer Sperre über den ganzen Ordner**, `Sitzungen/sitzungen.sperre`
(Nachlauf darüber, Befund 9): zwei Fenster, die zugleich starten oder enden,
räumen einander keine Marke weg. Diese Datei bleibt liegen. Gemeldet wird mit
Spur, wenn die Datei Inhalt hat, sonst als Ende von außen. **Die Spur geht
nirgendwohin von selbst**: sie steht im Protokollteil des Diagnosepakets,
geschwärzt und erst nach Zustimmung — **ohne den Wert, den der Code mit
`panic` warf**: er ist freier Text, und ein Kennwort darin erkennt keine
Schwärzung. Mit geht, was das Laufzeitsystem schreibt — Stapel, Signale,
Laufzeitfehler —, **und nur in der Gestalt, in der es sie schreibt**: eine
Zeile in anderer Gestalt wird geschwärzt, ebenso ein Wert, der nur wie ein
Laufzeitfehler beginnt, und die Marken am Kopf einer Goroutine (Nachlauf über
Phase 1, Befund 5). Die vollständige Spur bleibt auf diesem Rechner. **Das
gilt für die Spur eines Absturzes und nicht für das Protokoll**: eine
gefangene Panik nennt dort ihren Wert als `grund`, und für ihn gilt die
Schwärzung nach Gestalt wie für jede Fehlermeldung
(#2399). Leere Spuren
räumt der Schluss weg, von den übrigen bleiben die zehn neuesten.

**Und daneben, was sich wegwerfen lässt** — seit OPS-08
(#1725)
unter `Caches`, damit niemand es für Daten hält:

```
~/Library/Caches/MagicSetEditor4Mac/
├── scryfall/          Antworten von Scryfall — bis 0.6.0 lag er neben
│   ├── bilder/        dem Bestand und zieht beim ersten Start einmal um
│   └── sammeldaten/
└── bildvorschlaege/   Bilder, die die KI vorschlägt, bis jemand wählt
```

Was jeder Speicher hält, wie groß er werden darf und wie er leer wird, steht
in [`ZWISCHENSPEICHER.md`](ZWISCHENSPEICHER.md).

**Nicht in `~/Documents`** — dort liegen die Sets der Nutzenden, und das ist
etwas anderes ([0029](entscheidungen/0029-bilder-liegen-neben-dem-set.md)).

### `bestand.db.sperre` darf man löschen

Sie ist leer und trägt nichts; was zählt, ist die **Sperre des Betriebssystems**
darauf (#1163).

| Wer                                                       | Wann               | Wie lange                          |
| --------------------------------------------------------- | ------------------ | ---------------------------------- |
| jeder Verbraucher — Fenster, `mse`, `mse-mcp`, `mse-rest` | beim **Öffnen**    | bis die Datenbank offen ist        |
| das Zurückspielen                                         | vor dem **Tausch** | bis beide Umbenennungen durch sind |

**`bestand.db.wanderung` darf man ebenso löschen**, und sie gilt für etwas
anderes (M02, #2346):
**jeder Starter** hält sie ausschließlich, während er in die Buchführung sieht,
vor der Wanderung sichert und wandert. Ein zweiter — das Fenster und ein
Dienst nach einer Aktualisierung — wartet bis zu einer Minute und findet dann
nichts mehr zu tun; bis dahin sicherten beide, und die zweite „Sicherung vor"
hielt den Stand **nach** der Wanderung. Wer in der Minute nicht drankommt,
startet nicht, sondern hört „belegt" und versucht es gleich noch einmal.
**Dienst und Werkzeugserver halten sie geteilt**, je Anfrage, solange die
läuft — eine Wanderung wartet, bis sie fertig ist (unten). **Eine wartende Wanderung hat Vorrang** (#2406): sie nimmt
zuerst `bestand.db.eintritt` allein, und ein neuer Ruf kommt an dieser
Schranke nicht vorbei; die laufenden enden mit ihrer Antwort, und dann
wandert sie. Bis dahin konnten überlappende Rufe ohne Lücke sie die ganze
Minute aushungern. `bestand.db.eintritt` darf man löschen wie die beiden
anderen. **Lässt sich eine von beiden nicht anlegen** — etwa weil unter
ihrem Namen ein Ordner liegt —, wandert niemand und antwortet kein Dienst:
„gesperrt", mit dem Rat, den Ort freizugeben.

**Warum das Öffnen sie nur kurz hält**: wer die Datenbank schon offen **hat**,
wird von der SQLite-Probe gesehen — seine Verbindung hält Sperren auf `-wal`
und `-shm`. Was die Probe nicht sehen kann, ist ein Prozess, der genau
**jetzt** öffnet, und genau dieses Fenster deckt die Nachbardatei ab. Sie über
die ganze Laufzeit zu halten wäre ein Griff je Bestand, der nie zurückgegeben
wird — bei einer Grenze von 256 offenen Dateien scheiterte der 250.

**Sie überlebt keinen Absturz als Sperre**, und das ist richtig: das
Betriebssystem gibt sie mit dem Prozess frei. Eine Sperre, die einen Absturz
überlebte, machte den Bestand für immer unbenutzbar.

**Was sie nicht sieht**: ein `sqlite3` im Terminal nimmt sie nicht. Dafür fragt
das Zurückspielen zusätzlich SQLite selbst — die eine Prüfung fragt die
Teilnehmer, die andere die Datei.

**Nichts davon liegt im Programmbündel.** Im `.app` stehen nur Programm,
Symbole und die mitgelieferte Saat. Ein Austausch der Anwendung lässt die Daten
also unberührt.

### Was noch dazugehört

| Ort                                            | Was                                        |
| ---------------------------------------------- | ------------------------------------------ |
| `~/Downloads`                                  | wohin eine geladene Fassung käme (Phase Z) |
| Unified Log, Prozess `logger` — **ohne Subsystem**, und die Marke von `logger -t` übernimmt macOS nicht | dieselben Zeilen wie `mse.log` |

**Gelesen wird das Unified Log über den Prozess**, nicht über ein Subsystem —
das setzt nur `os_log`, und die Marke von `logger -t MagicSetEditor4Mac` lässt
macOS fallen (gemessen am 2026-10-15 auf macOS 27.0.1):

```bash
/usr/bin/log show --last 1h --style compact --predicate 'process == "logger"'
```

**`/usr/bin/log` und nicht `log`**: in der zsh ist `log` ein eingebauter
Befehl, der etwas anderes tut.

---

## Anmeldung über einen fremden Anbieter

**Diese Anwendung liefert keine Zugangsdaten mit**, und das ist eine
Entscheidung: eine mitgelieferte `client_id` wäre für alle dieselbe, und wer
sie missbraucht, bringt sie für alle zu Fall.

Wer sich über GitHub, Google, Microsoft oder X anmelden will, legt **beim
Anbieter selbst** eine OAuth-Anwendung an und trägt deren `client_id` unter
*Einstellungen → Anwendung → Anmeldung über einen Anbieter* ein.

**Als Rückrufadresse** gibt man dort die Adresse an, die beim Anmelden im
Browser steht. Sie zeigt auf die Schleifenadresse dieses Rechners und gilt nur,
solange die Anmeldung läuft.

### Das Geheimnis kommt in den Schlüsselbund

**Je Anbieter verschieden, und je Anbieter nachgelesen** (2026-09-05):

| Anbieter | `client_secret` |
| --- | --- |
| **GitHub** | **Pflicht** — auch mit PKCE |
| Google | freigestellt für Desktop-Klienten |
| Microsoft | **verboten** für öffentliche Klienten |
| X | nicht verlangt |

**Es kommt nicht in den Bestand.** Den gibt man weiter — als Sicherung, an
einen zweiten Rechner, in eine Fehlermeldung; den Schlüsselbund nicht.

Anzulegen in der **Schlüsselbundverwaltung** als *Passwort für Programme*:

| Feld | Wert |
| --- | --- |
| Name (`service`) | `MagicSetEditor4Mac-Anmeldung` |
| Account | der Anbietername — `github`, `google`, `microsoft` oder `x` |
| Passwort | das `client_secret` |

Auf der Kommandozeile dasselbe:

```
security add-generic-password -s MagicSetEditor4Mac-Anmeldung -a github -w
```

> **Am Fenster geht es auch ohne die Schlüsselbundverwaltung.** Seit
> #1401 steht
> unter *Einstellungen → Anwendung → Anmeldung über einen Anbieter* ein Feld
> je Anbieter. Der Weg hier bleibt für alle, die kein Fenster haben — und weil
> ein Eintrag, den man von Hand angelegt hat, dort erkannt wird.
>
> **Ohne Fenster gibt es kein Feld**, und der Grund ist der Schlüsselbund:
> macOS fragt jeden Prozess einmal um Erlaubnis, und ein Befehl, der öffnet,
> arbeitet und endet, stellte diese Frage bei jedem Aufruf.

## Sperren, Abmelden und ein verlorener Zugang

**Alles bleibt auf diesem Rechner** — ID-02
(#1701).
Ein lokales Konto braucht keine zentrale Registrierung und keine Bestätigung
per E-Mail; eine Adresse ist hier eine Kennung und kein nachgewiesener Kanal
(Vermerk [0066](entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md)).

| Zustand | Was das Fenster zeigt | Was Go tut |
| --- | --- | --- |
| **angemeldet** | die Anwendung | jede Bindung bekommt das Profil |
| **gesperrt** | nur den Sperrbildschirm | **jede gebundene Methode hält an, bevor sie etwas tut** — außer den zehn, die Sperranzeige und Anmeldung brauchen (`Sitzungsstand`, `Entsperren`, `EntsperrenMitSensor`, `Abmelden`, `LokalAnmelden`, `KennwortWiederherstellen`, `Nutzer`, `Einstellungen`, `SpracheSetzen`, `Startbefund`). Dienste, Horchposten und Peersitzungen, die das Fenster gestartet hat, werden beendet |
| **abgemeldet** | die Wahl, wer arbeitet | dito |

**Wie das Tor hält** (`internal/app/sitzungstor.go`): eine gebundene Methode
kommt an die Anwendung nur über `Bindungen.app()`, und dort steht das Tor;
gesperrt hält sie an, und Wails gibt das als Aufruffehler zurück. **Auch die
Rechnungen ohne Bestand halten** (Symmetrie, Zuschnitt, Teilliste) — über
`Bindungen.tor()` —, denn gesperrt ist gesperrt. `SpracheSetzen` geht durch:
es stellt nur die Sprache der Menüleiste, und die Sperranzeige spricht die
des Fensters. `Startbefund` auch: die Wurzel des Fensters fragt ihn vor dem
Tor, und gesperrt sagt er nur „offen“. **Bis zum Nachlauf über ID-02 stand das Tor in `App.kontext`**,
und am 2026-09-27 gingen 32 von 458 Bindungen daran vorbei — darunter `KIVorschlagen` und
`IndexAbbrechen`. Zwei Wächter halten es jetzt: einer ruft **jede** Bindung
gesperrt auf (`TestGesperrtHaeltJedeGebundeneMethode`), einer liest, dass
jede das Tor in ihrer ersten Anweisung ruft
(`TestJedeBindungRuftZuerstDasTor`). **Und nicht die Kommandozeile**: sie
läuft unter dem Konto von macOS. **Ein Lesefehler beim Start schließt**: lässt sich
der festgehaltene Stand nicht lesen, beginnt die Sitzung abgemeldet und nicht
offen.

**Der Zustand steht im Bestand** (Eintrag `sitzung` in der Tafel `bestand`):
wer gesperrt hat und die Anwendung beendet, findet sie gesperrt wieder. **Was
die Sperre nicht ist**: eine Verschlüsselung. Wer die Datei hat, liest sie —
dafür gibt es den verschlüsselten Bestand weiter unten. Die Sperre schützt
ein unbeaufsichtigtes Fenster, und die Kommandozeile läuft unter dem Konto von
macOS; sie **nennt** den Stand (`mse sitzung`) und hält sich beim
Kennwortwechsel an ihn.

**Gesperrt wird nur, wer zurückkommt** — mit einem Kennwort oder mit
eingerichteter **lokaler Entsperrung** (Touch ID). Die lokale Entsperrung ist
**kein Passkey**: sie meldet bei keinem Dienst an, und es gibt keinen
WebAuthn-Dienst in dieser Anwendung. Anmeldungen über einen Anbieter bleiben
davon getrennt.

### Ohne Netz

**Nichts davon braucht eines.** Kennwort, Zettel und Touch ID werden auf
diesem Rechner geprüft; nur die Anmeldung über einen fremden Anbieter
(oben) geht hinaus.

### Wenn das Kennwort vergessen ist

1. **Mit einem Code vom Zettel** — den zehn Wiederherstellungscodes, die es
   beim Einrichten des zweiten Faktors gab. Am Sperrbildschirm unter
   *Kennwort vergessen?*, an der Kommandozeile mit
   `mse kennwort-wiederherstellen <nutzer>` (Code und neues Kennwort je eine
   Zeile auf der Standardeingabe). Jeder Code gilt einmal, und er wird **im
   selben Geschäft** verbraucht, in dem das Kennwort gesetzt und — am
   Sperrbildschirm — die Sperre gelöst wird: scheitert eines, bleibt alles,
   wie es war. **Gesetzt wird nur das Kennwort**: keine
   Rolle, keine Freigabe, kein Gegenüber ändert sich, und ein neuer Schlüssel
   eines Gegenübers verlangt weiterhin seine eigene Bestätigung (AK-04).
   **Am Sperrbildschirm** ist danach entsperrt; **nach dem Abmelden** bleibt
   es abgemeldet, und angemeldet wird mit dem neuen Kennwort und dem zweiten
   Faktor.
2. **Ein Einmalkennwort aus der App genügt nicht** — es beweist das Telefon,
   nicht das Konto.
3. **Ohne Zettel bleibt die Sicherung** (unten, *Zurückspielen*). Eine
   Wiederherstellung über eine E-Mail-Adresse gibt es nicht und soll es nicht
   geben: die Anwendung verschickt nichts, und eine Adresse allein ist kein
   Nachweis.

### Wenn der Schlüsselbund nicht mitspielt

**Kennwort und Zettel liegen gehütet im Bestand** (Argon2id), nicht im
Schlüsselbund — sie gehen auch auf einem zweiten Mac. **Das Geheimnis des
Einmalkennworts liegt im Schlüsselbund**: kennt ihn dieser Rechner nicht, hilft
der Zettel beim zweiten Faktor. **Touch ID gilt nur auf dem Rechner, auf dem
sie eingerichtet ist.** Und der Schlüssel eines verschlüsselten Bestands
(unten) lässt sich nicht wiederherstellen — dafür gibt es die Sicherung.

## Verschlüsselung

**Der Bestand liegt voreingestellt im Klartext**, und das ist seit Werkstück 14
von Phase 4.12 eine Wahl und keine Feststellung
([0072](entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md)).

Eingeschaltet wird unter *Einstellungen → Bestand* oder so:

```bash
mse bestand-verschluesselung an
```

Nachgesehen wird an jeder der vier Flächen — `mse bestand-verschluesselung`
ohne Wort, das MCP-Werkzeug `lesen.bestand.verschluesselung`, oder:

```bash
curl -s -H "Authorization: Bearer $MSE_MARKE" http://127.0.0.1:7777/bestand/verschluesselung
```

### Der Schlüssel

**32 zufällige Bytes im Schlüsselbund dieses Rechners**, unter dem Dienstnamen
`MagicSetEditor4Mac-Bestand`. Kein Kennwort, das jemand sich ausdenkt — und
deshalb auch keines, das jemand sich merken kann.

> **Er lässt sich nicht wiederherstellen.** Ist der Schlüsselbundeintrag fort,
> sind die Daten fort. Es gibt keinen Zweitschlüssel, keine Hintertür und keine
> Wiederherstellungsfrage. **Vor dem Einschalten sichern.**

Wer mehrere Bestände führt, gibt ihnen über `MSE_SCHLUESSELBUND` je einen
eigenen Bund — dieselbe Form wie `MSE_BESTAND`.

### Was sie leistet und was nicht

| | |
| --- | --- |
| Die Datei ist ohne Schlüssel nicht zu öffnen | **ja** |
| Schutz gegen Fälschung oder Manipulation | **nein** — Adiantum sagt es selbst zu nicht |
| Ein Angreifer mit **mehreren** Abzügen lernt nichts | **nein** — die Chiffre ist deterministisch: er sieht, welche Blöcke sich geändert haben |
| Die Sicherung ist verschlüsselt | **nein**, und das ist Absicht |

**Der letzte Punkt ist der, den man beim Sichern wissen muss.** Eine
`.msesicherung` trägt den Bestand im **Klartext** — sonst ließe sie sich auf
keinem anderen Rechner zurückspielen, und eine Sicherung, die nur dort aufgeht,
wo sie entstand, ist keine. Wer sie weglegt, legt eine unverschlüsselte Kopie
weg; dafür gibt es verschlüsselte Disk-Images.

### Wenn der Bestand nicht mehr aufgeht

Meldet die Anwendung, die Datei sei verschlüsselt und der Schlüssel liege nicht
im Schlüsselbund, dann **ist die Datei in Ordnung**. Sie zu ersetzen macht den
Verlust endgültig. Erst nachsehen, ob der Eintrag noch da ist:

```bash
security find-generic-password -s MagicSetEditor4Mac-Bestand
```

## Ordner — wo Dialoge aufgehen, und was der Dienst anfassen darf

**Vier Einstellungen, alle auf dem Gerät** (OPS-02,
#1719),
unter *Einstellungen › Daten und Sicherung* oder so:

```bash
mse einstellungen setzen vorgabe.ausfuhrpfad ~/Karten/Ausfuhr
```

| Einstellung | Wo er gilt |
| --- | --- |
| `vorgabe.hochladepfad` | jeder Dialog zum **Öffnen** — Bilder, Sets, MTG JSON |
| `vorgabe.ausfuhrpfad` | jeder Dialog zum **Sichern**, der keinen eigenen Ort hat |
| `vorgabe.sicherungspfad` | die Sicherung des Bestands |
| `vorgabe.diagnosepfad` | Diagnosetext und Diagnosepaket |

**Die Angabe ist vollständig oder beginnt mit `~/`** — sonst nichts. Ein
relativer Pfad wäre relativ zu einem Arbeitsverzeichnis, und das ist beim
Fenster ein anderes als an der Kommandozeile; `~anna`, `$HOME` und `..`
werden abgewiesen, statt still anders gedehnt zu werden, als eine Shell es
täte. Leer heißt: wo das Betriebssystem den Dialog öffnet.

**Beim Setzen fragt die Anwendung die Platte** — auf allen vier Flächen
dieselbe Prüfung: gibt es den Ordner, ist er einer, darf sie hineinschreiben,
und sind dort noch 16 MB frei. Jede Antwort hat ihren eigenen Satz.

**Beim Sichern fragt sie noch einmal.** Fehlt der Ordner dann — ein
abgezogenes Laufwerk, eine getrennte Netzfreigabe —, hält der Vorgang mit
einer Meldung an, **statt die Datei an einem anderen Ort abzulegen**. Beim
Öffnen nicht: dort wird nichts gespeichert, und der Dialog geht auf, wo das
System ihn öffnet.

### Einen Ordner ändern verschiebt nichts

**Die vier Ordner sagen, wo ein Dialog aufgeht — nicht, wo etwas liegt.** Wer
einen ändert, findet seine früheren Dateien am alten Ort. Was die Anwendung
selbst führt — Bestand, `pakete/`, `beiwerk/` —, liegt beisammen neben dem
Bestand (oben) und zieht **als Ganzes** um: sichern, am neuen Ort mit
`MSE_BESTAND` starten, zurückspielen. Dieser Weg hat, was ein Umzug braucht —
**eine Vorschau** (`mse bestand-schauen`), **einen Rückweg**
(`bestand-vor-rueckspielung.db`, `beiwerk-vorher`) und **eine Prüfung** der
Sicherung, bevor getauscht wird (unten, *Zurückspielen*). Einen zweiten Weg
daneben, der Bilder oder Pakete einzeln woandershin legt, gibt es mit
Absicht nicht: ein Bestand, dessen Bilder anderswo liegen, geht beim nächsten
Zurückspielen halb verloren.

### Was ein offener Server anfassen darf

**`mse-rest` und `mse-mcp --adresse` nehmen Pfade — und nicht jeden.** Ein
Pfad an einen offenen Server liegt im Zuhause **ohne seine verborgenen
Winkel** (`~/.ssh`, `~/.zshrc`, `~/Library`), in den **eigenen Orten** der
Anwendung — dem Ordner des Bestands und `pakete/`, auch wenn sie unter
`~/Library` liegen —, in einem der vier Ordner oben oder im
Zwischenablageordner des Prozesses (`$TMPDIR`), über den Werkzeuge Dateien
reichen. Alles andere wird abgewiesen, bevor ein Endpunkt es sieht — **mit
Token**; ohne erfährt niemand, welche Orte es gibt. **Ein verborgener Ordner
ist überall ausgeschlossen**, auch in einem erlaubten: ein `.git` auf dem
Ausfuhrlaufwerk ist so wenig gemeint wie `~/.ssh`.

**Die vier Ordner zählen nur im Zuhause oder auf einem Laufwerk** unter
`/Volumes`: ein am Fenster eingestelltes `/srv/…` nimmt ein Server nicht an —
dafür gibt es `MSE_DIENSTORDNER` (unten). Über Dienst und Werkzeugserver
lässt sich ein solcher Ordner gar nicht erst einstellen, auch über stdio.

**Gefragt wird, wo die Datei wirklich liegt**: der aufgelöste Pfad gegen die
aufgelösten Ordner. Ein `~/Karten`, das nach `~/.ssh` zeigt, liegt in
`~/.ssh`, und `/Volumes/Macintosh HD` ist auf jedem Mac eine Verknüpfung auf
`/` — auch die Ordnerregel löst deshalb auf (Nachlauf über Phase 6.4). Die
Regel steht in `internal/domain/dienstpfad.go`, geprüft wird vor jedem Weg an
den Namen, unter denen ein Pfad reist; ein Wächter hält diese Namen gegen die
Felder beider Server.

**Weitere Ordner nennt `MSE_DIENSTORDNER`** — getrennt wie `PATH`, für einen
Betrieb, der `mse-rest` etwa auf `/srv/karten` ansetzt:

```bash
MSE_DIENSTORDNER=/srv/karten:/Volumes/Ablage mse-rest
```

**`mse-mcp` ohne `--adresse`, also über stdio, und die Kommandozeile prüfen
Pfade nicht** — dort spricht nur, wer den Prozess selbst gestartet hat, mit
den Rechten, die er ohnehin hat. Die Regel für eingestellte Ordner gilt über
stdio trotzdem (oben).

## Platz

**Was diese Anwendung ablegt**, alles unter
`~/Library/Application Support/MagicSetEditor4Mac/`:

| Posten | Was | Räumbar |
| --- | --- | --- |
| `bestand.db` (`-wal`, `-shm`) | die Daten | nein |
| `beiwerk/` | eingesetzte Bilder | nein |
| `pakete/` | bezogene Vorlagenpakete | über die Seite *Pakete* |
| `bestand-vor-*.db` | Sicherungen vor Migrationen | **ja**, bis auf den Weg zurück |
| `mse.log`, `mse.log.1` … `mse.log.9` | das Protokoll samt seinen Vorfassungen (`protokoll.aufbewahrung`) | **ja**, bis auf die laufende Datei |
| alles Übrige im Ordner | *Sonstiges* — darunter `Leistungsprofile/` (höchstens `messung.profile` Dateien, Vorgabe 3), `Sitzungen/` (Startmarken und höchstens zehn Absturzspuren) und `.sperre`-Dateien | nein — was die Anwendung nicht kennt, löscht sie nicht; die Leistungsprofile hält ihre eigene Zahl klein |

**Der letzte Posten ist der wichtige.** Er zählt, was kein anderer
beansprucht — beiseitegelegte Bestände, Reste eines Zurückspielens, alles, was
jemand dort abgelegt hat. **Am 2026-09-07 gemessen**: der Ordner wog 13 GB, und
die vier Posten von damals nannten davon 1,5. Eine Aufzählung ist eine Zusage,
und diese hielt sie nicht; ein Rest hält sie von selbst.

**Die letzten wachsen von selbst.** Vor jeder Migration, die den Bestand
umbaut, entsteht eine — und weil ihr Name den Stand trägt, ist es bei jedem
neuen Stand eine weitere. **Am 2026-09-07 gemessen**: dreißig Dateien, über
zwei Gigabyte.

**Nachsehen und wegräumen:**

```
mse platz
mse platz freigeben vorsicherungen
```

Am Fenster steht dasselbe unter *Einstellungen → Bestand*, über MCP als
`lesen.platz` und `schreiben.platz.freigeben`, über den Dienst als
`GET /platz` und `DELETE /platz/{posten}`.

> **Der Weg zurück bleibt stehen.** Wer nach einer Migration merkt, dass etwas
> fehlt, spielt ihn zurück; wer ihn mitnähme, nähme genau die Sicherung, für
> die es sie gibt.

**Es bleiben zwei, und die Gründe sind verschieden** (seit 0.4.0): die jüngste
Datei überhaupt, und die jüngste, die **diese** Fassung öffnen kann. Meistens
ist das dieselbe. Nach zwei Aktualisierungen ist es das nicht:
`bestand-vor-0.5.0.db` hält den Stand von 0.4.0 und geht auf 0.3.0 nicht auf —
wer nur „die jüngste" stehen ließe, räumte genau den Rückweg weg.

**`-wal` und `-shm` zählen dabei nicht als eigene Sicherung.** Sie gehören zu
der Datei, neben der sie liegen, und gehen mit ihr — bleibt sie, bleiben sie.
Das ist keine Feinheit: das Nachsehen, **welche** Sicherung aufgeht, legt selbst
ein `-shm` an (SQLite braucht es, um einen WAL zu lesen, auch nur lesend), und
das ist dann die jüngste Datei im Ordner. Ohne diese Unterscheidung belegte ein
Begleiter den Platz einer echten Sicherung.

## Was sich von selbst meldet

**Vier Lagen melden sich auf der Startseite**, oben in „Was Aufmerksamkeit
braucht" (OPS02, #2348).
Das Fenster fragt alle 30 Sekunden, solange es sichtbar ist:

| Lage | Wann | Wohin der Hinweis führt |
| --- | --- | --- |
| Platz | weniger als 2 GiB **oder** 5 % frei auf dem Volumen, auf dem der Bestand liegt; **wieder gut erst über 2,5 GiB und 6 %** | *Daten und Sicherung* |
| Sicherung | die letzte zurückgelesene Sicherung ist älter als 14 Tage. **Noch nie gesichert meldet sich hier nicht** — das steht unter *Letzte Sicherung* als „unbekannt" | *Daten und Sicherung* |
| Dienst beim Anmelden | seine plist liegt da, launchd nennt keinen laufenden Prozess, und der letzte Lauf endete nicht mit 0 (ein Signal zählt als 128 plus seine Nummer). **Startet launchd ihn immer wieder neu, bleibt der Hinweis**, bis er zwischen zwei Abrufen durchlief | *Diagnose* |
| Webprozess | WebKit hat den Prozess der Webansicht beendet, und Wails hat die Seite neu geladen — gezählt seit dem Start, mit einer Zeile im Protokoll | *Diagnose* |

**Was sich nicht fragen ließ, meldet sich nicht**: ein `launchctl`, das nicht
antwortet, ist kein beendeter Dienst. Der Speicher von Anwendung und
Webprozess steht unter *Diagnose › Laufzeiten und Verbrauch*, als physischer
Fußabdruck wie in der Aktivitätsanzeige; lässt sich der Webprozess nicht
zuordnen, steht dort, warum.

## Sicherung

**In der Anwendung**: _Einstellungen_ → _Bestand_ → **Sichern**. Es entsteht
eine einzelne Datei `<name>.msesicherung`, aus der heraus die Anwendung wieder
aufsetzen kann.

**Und ohne Fenster** seit 0.3.0
(#1169) —
das ist der Weg für einen Zeitplan oder ein Skript:

```bash
mse bestand-sichern ~/Sicherungen/$(date +%F).msesicherung
```

Dieselbe Fähigkeit steht über MCP (`schreiben.bestand.sichern`) und über den
Dienst (`POST /bestand/sicherung`).

### Was drin steht

Sie ist ein **ZIP-Archiv** mit drei Arten von Einträgen
(#1162):

| Eintrag          | Was                                                                                     |
| ---------------- | --------------------------------------------------------------------------------------- |
| `bestand.db`     | die Datenbank, geschrieben über `VACUUM INTO` — also **ohne** `-wal` und `-shm` daneben |
| `beiwerk/…`      | die eingesetzten Kartenbilder, mit ihrer Ordnerstruktur                                 |
| `sicherung.json` | Fassung, Zeitpunkt, und je Eintrag Größe und SHA-256                                    |

**Bis 0.2.0 war es eine nackte `.db`** — und damit unvollständig: die Bilder
liegen daneben unter `beiwerk/`, und
[0029](entscheidungen/0029-bilder-liegen-neben-dem-set.md) sagt selbst, dass
ein Set ohne sie unvollständig ist. **Auf demselben Rechner fällt das nie
auf**, weil das alte Verzeichnis noch daliegt; sichtbar wird es erst nach einem
Rechnerwechsel — also genau dann, wenn die Sicherung das Einzige ist, was
bleibt.

**Was sie sonst trägt, und was nicht**, steht als Matrix in
[`SICHERUNG.md`](SICHERUNG.md) (OPS-06): **eigene Pakete** — angelegt, von
Hand hineingelegt — reisen unter `pakete/` mit; **bezogene** stehen nur als
Verweis mit Herkunft und Prüfsumme im Manifest, denn sie sind wiederbeziehbar
(_Pakete_ → _Beziehen_) und wögen über ein Gigabyte. Die Vorschau nennt vor
dem Zurückspielen, welche davon fehlen. Protokoll, Zwischenspeicher und die
Geheimnisse im Schlüsselbund bleiben draußen.

**Vor jeder Migration entsteht zusätzlich eine automatisch**, unter
`bestand-vor-<Fassung>.db`
(#939). **Sie
ist eine nackte Datenbank** und trägt keine Bilder — sie ist der Rückweg einer
Migration und nicht die Sicherung vor einem Rechnerwechsel.

**Von Hand**, bei geschlossener Anwendung — dann aber **beides**:

```bash
sqlite3 ~/Library/Application\ Support/MagicSetEditor4Mac/bestand.db "VACUUM INTO '$HOME/Desktop/mse-sicherung.db'"
cp -R ~/Library/Application\ Support/MagicSetEditor4Mac/beiwerk ~/Desktop/beiwerk
```

Das Ziel darf nicht dastehen — weder die Anwendung noch SQLite überschreibt
etwas.

### Die Sicherung mit Kennwort

**Seit OPS-06 kann eine Sicherung ein Kennwort tragen**
(#1723) —
am Fenster zwei Felder vor dem Speichern-Dialog, an der Kommandozeile die
Umgebungsvariable:

```bash
MSE_SICHERUNGSKENNWORT='ein Satz aus vier Wörtern' mse bestand-sichern ~/Sicherungen/$(date +%F).msesicherung
```

Dann ist sie mit [age](https://age-encryption.org) verschlüsselt und geht auf
jedem Rechner auf — mit `age -d` sogar ohne diese Anwendung. **Das Kennwort
liegt nirgends**, auch nicht im Schlüsselbund: wer es vergisst, hat die
Sicherung verloren. Wie das gebaut ist und warum, steht in
[`SICHERUNG.md`](SICHERUNG.md).

**Ohne Kennwort ist sie Klartext, gleich wie der Rechner eingestellt ist.**
FileVault schützt das Laufwerk _dieses_ Rechners; eine Kopie auf einem
USB-Stick oder in einer Wolke ist davon nicht berührt.

**Ein verschlüsselter Bestand gibt keinen Klartext her, ohne dass es jemand
will** ([0072](entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md),
oben unter „Verschlüsselung"): eine Sicherung ohne Kennwort verlangt dann das
ausdrückliche „ohne Kennwort sichern" — am Fenster ein Haken, an der
Kommandozeile `--klartext`. Bis OPS-06 entstand dort still ein Archiv im
Klartext, und nur dieser Abschnitt sagte es.

---

## Zurückspielen

**Auf der Kommandozeile** seit 0.3.0. Erst hineinsehen, dann übernehmen — und
der Zeitpunkt aus dem Hineinsehen bestätigt:

```bash
mse bestand-schauen ~/Sicherungen/2026-08-28.msesicherung
```

```bash
mse bestand-zurueckspielen ~/Sicherungen/2026-08-28.msesicherung 2026-08-27T23:07:13.375Z
```

**Nur dort und im Fenster**, und beide Ablehnungen sind begründet: nicht über
MCP, weil der ganze Bestand nichts ist, was ein Sprachmodell still ersetzen
darf; nicht über den Dienst, weil er eine Verbindung über seine Laufzeit hält
und alle seine Nähte sie sich beim Aufbau abgeholt haben. Die Begründungen
stehen ausgeschrieben in [`FAEHIGKEITEN.md`](FAEHIGKEITEN.md).

**In der Anwendung**: _Einstellungen_ → _Bestand_ → **Sicherung zurückspielen
…** Der Weg hat drei Schritte, und die Reihenfolge ist die Zusage:

1. Der native Dateidialog fragt nach der Sicherung.
2. Sie wird **gelesen, nicht übernommen** — ausgepackt an einen Zwischenplatz
   und nur dort geöffnet, damit die Prüfung ihren Gegenstand nicht verändert.
   Für die Vorschau **allein der Bestand**: die Bilder werden erst geholt,
   wenn jemand die Rückfrage bejaht hat.
3. Erst wenn sie taugt, wird getauscht: **zuerst die Datenbank, dann die
   Bilder.** Scheitert der zweite Schritt, geht auch der erste zurück.

**Ein Absturz dazwischen wird beim nächsten Start nachgeholt.** Zwei
Umbenennungen in zwei Verzeichnissen kann kein Dateisystem zusammen ausführen;
was geht, ist Zuendeführen. Vor dem ersten Schritt entsteht neben dem Bestand
eine Marke `rueckspiel-laeuft.json`, die den zweiten vollständig beschreibt;
nach dem zweiten verschwindet sie. Findet ein Start sie vor, führt er den
Bildertausch aus und räumt sie weg — **ohne Rückfrage**, denn was zu tun ist,
haben die Nutzenden bereits bestätigt. Es steht im Protokoll.

**Ein Fehler dabei hält den Start nicht auf.** Die Anwendung ist benutzbar; was
fehlt, sind Bilder — und ein Programm, das deswegen nicht startet, nimmt auch
den Weg, es von Hand zu richten. Beide ersetzten Stände liegen dann noch da:
`bestand-vor-rueckspielung.db` und `beiwerk-vorher`.

**Eine Sicherung aus einer neueren Fassung wird abgewiesen**, nicht
stillschweigend geöffnet — und zwar an **zwei** Stellen:

| Woran                                        | Was es fängt                                                                                                               |
| -------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| die **Fassung** im Verzeichnis der Sicherung | eine künftige Fassung, deren Schema gleich blieb und deren Archivgestalt sich änderte                                      |
| die **Schemaschritte** im Bestand            | eine Datenbank, die Schritte führt, die diese Fassung nicht kennt — auch aus einer nackten `.db`, die kein Verzeichnis hat |

Eine Sicherung **ohne** Fassungsangabe oder mit einer unlesbaren kommt durch:
aus „weiß nichts darüber" wird kein „stammt aus der Zukunft". Was den Bestand
angeht, sagt die zweite Prüfung ohnehin mehr.

### Was beim Auspacken begrenzt ist

Eine Sicherung sagt über sich selbst, wie groß ihre Teile sind — und das ist
eine Angabe aus der Datei, nicht über sie. Deshalb stehen daneben Grenzen:

| Grenze                           | Wert                                                      |
| -------------------------------- | --------------------------------------------------------- |
| das Verzeichnis `sicherung.json` | 64 MiB                                                    |
| ein einzelner Eintrag            | 16 GiB                                                    |
| die Zahl der Einträge            | 1.048.576 — gezählt am ZIP-Verzeichnis, nicht am Manifest |
| alle Teile zusammen              | 64 GiB                                                    |

Ausgepackt wird **strömend** in eine Nachbardatei; an seinen Ort kommt ein Teil
erst, wenn seine SHA-256 stimmt. Eine fremde oder beschädigte Datei kostet
damit Platte, nicht Arbeitsspeicher.

**Was diese Grenzen nicht abdecken**: `archive/zip` liest das Inhaltsverzeichnis
eines Archivs **vollständig**, bevor dieses Programm es zu Gesicht bekommt. Was
dabei anfällt, begrenzt Go und nicht diese Zahlen.

### Und die Sicherung entsteht daneben

Geschrieben wird in eine Nachbardatei; ihren endgültigen Namen bekommt sie erst
am Ende, über einen harten Verweis. Damit steht unter dem Namen einer Sicherung
nie eine halbe.

**Auf FAT und exFAT geht das nicht** — beide kennen keine harten Verweise, und
macOS richtet externe Datenträger so ein, wenn sie auch an Windows sollen. Dort
wird stattdessen kopiert: die Sicherung entsteht, aber ein Abbruch **während**
dieses letzten Schritts kann eine halbe Datei unter dem endgültigen Namen
hinterlassen. Sie wird beim Zurückspielen benannt abgewiesen.

### Beide Formate

| Gewählt wird                        | Was geschieht mit den Bildern                                                                                                                                                      |
| ----------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `.msesicherung` **mit** `beiwerk/`  | das Bilderverzeichnis wird **ersetzt**                                                                                                                                             |
| `.msesicherung` **ohne** `beiwerk/` | es wird **geleert** — die Sicherung sagt, dass es keine gab                                                                                                                        |
| nackte `.db` aus 0.1.0 oder 0.2.0   | es **bleibt stehen** — sie sagt über Bilder nichts, und aus „weiß nichts darüber" ein „da war nichts" zu machen wäre der teuerste Datenverlust, den dieses Programm anrichten kann |

Unterschieden wird an den **ersten vier Bytes**, nicht an der Endung: wer eine
Datei umbenennt, ändert nicht, was drin steht.

**Der ersetzte Bilderstand liegt beiseite**, unter `beiwerk-vorher/` neben dem
Bestand — dieselbe Zusage wie für die Datenbank unter
`bestand-vor-rueckspielung.db`. Er bleibt bis zum nächsten Zurückspielen liegen.

---

## Zurückgehen — vier Arten, und zwei kann diese Anwendung nicht selbst

**Seit Werkstück O7 der Iteration 0.5.0**
(#1431).

Bis dahin gab es **einen** Knopf: *Sicherung zurückspielen*. Er tauscht die
Daten — und wer ihn drückt, weil eine neue Fassung nicht tut, was sie soll, hat
die Fassung damit nicht gewechselt. Vier Dinge können zurück:

| Art | Was zurückgeht | Wer es tut |
| --- | --- | --- |
| **Appfassung** | das Programm selbst | ein Mensch, nach Prozessende (`nachProzessende`) |
| **Schema** | die verbuchten Wanderungen | **niemand** — es gibt keine Rückwärtswanderung (`nurVorwaerts`) |
| **Datenstand** | `bestand.db` und `beiwerk/` | diese Anwendung, über das Zurückspielen oben |
| **Paketstand** | die Vorlagenpakete | diese Anwendung, über den Bezug an einem Stand |

```bash
mse ruecksprung-arten
```

**Warum es keine Rückwärtswanderung gibt.** Eine rückwärts laufende Migration
wäre ein zweiter Satz Wahrheit über dasselbe Schema — und der zweite ist der,
den niemand je ausführt, bis es darauf ankommt. Der Weg zurück ist eine
**Sicherung**; das Issue verlangt es ausdrücklich: Vorwärtsmigration plus
Sicherungswiederherstellung, keine Pflicht zu Down-Migrationen.

### Die Appfassung tauschen — der Ablauf

**Ein Programm, das sich selbst ersetzt, während es läuft**, bräuchte auf macOS
einen Helferprozess, der das Bündel nach dem Ende überschreibt — ein zweites,
unsigniertes Stück Software mit Schreibrecht auf `/Programme`. Das gibt es hier
nicht und soll es nicht geben
([0003](entscheidungen/0003-keine-notarisierung.md): weder Developer-ID noch
Notarisierung, und **kein `root`**).

Was es stattdessen gibt, ist ein Ablauf mit vier Haltepunkten. **Jeder ist
überprüfbar**, und ein Abbruch dazwischen ist beim nächsten Start lesbar:

| Schritt | Was geschieht | Woran man sieht, dass er geglückt ist |
| --- | --- | --- |
| **1 Bereitlegen** | die gewünschte Fassung als `.dmg` herunterladen und ihre Prüfsumme gegen die **signierte** `pruefsummen.txt` halten | `mse vertraeglichkeit` gegen den Bestand — sie sagt, ob die Fassung ihn öffnen kann |
| **2 Sichern** | *Wechsel vorbereiten* im Fenster: Platz, laufende Arbeit, und eine Sicherung, die **zurückgelesen** wird | die Vorprüfung nennt kein Hindernis mehr |
| **3 Tauschen** | Anwendung beenden, das alte Bündel **beiseitelegen** (nicht löschen), das neue nach `/Programme` ziehen | im Finder liegen beide da — der letzte funktionierende Stand ist noch vorhanden |
| **4 Anlaufen** | starten. Der Start migriert vorwärts und beantwortet danach die Gesundheitsfrage | `mse gesundheit` gibt 0, oder das Fenster zeigt „bereit" |

**Der letzte funktionierende Stand wird erst weggeräumt, wenn Schritt 4
geglückt ist.** Das ist der ganze Trick: solange das alte Bündel daneben liegt,
ist der Rückweg ein Ziehen im Finder und keine Wiederherstellung.

**Geht Schritt 4 schief**, führt der Weg zurück über dasselbe alte Bündel — und
der Bestand über die Sicherung aus Schritt 2. Eine Wanderung, die
durchgelaufen ist, wird dabei **nicht** rückgängig gemacht; sie ist einfach
nicht mehr da, weil der Bestand der von vorher ist.

### Und was mit Dienst und Werkzeugserver dabei geschieht

**Ein Prozess, der den Bestand offen hält, überlebt den Tausch des Bündels**
und liefe danach mit dem Code von vorher weiter (M02,
#2346,
Befund A-11). Deshalb:

| Wer | Was geschieht | Woran man es sieht |
| --- | --- | --- |
| **ein anderes Programm** — ein `mse-rest` von Hand, ein `mse-mcp` eines Sprachmodells | die Vorprüfung in Schritt 2 nennt es mit Name und Prozessnummer, gefragt bei `lsof`; Schritt 2 ist erst durch, wenn es beendet ist | das Hindernis „hält den Bestand offen" mit `mse-rest (4711)` |
| **der eingerichtete Dienst** | wer wandern wird — Fenster, `mse`, `mse-mcp` oder `mse-rest` —, hält ihn vorher an (`launchctl kickstart -k`), **wenn er diesen Bestand hält** (gefragt bei `lsof`, nicht an der plist): unter der Sperre über die Wanderung und **vor** der Sicherung, und **nachweislich** — gewandert wird erst, wenn der alte Prozess fort ist; dauert das samt Ruf länger als zehn Sekunden, oder lässt sich `launchd` nicht befragen, ist der Bestand belegt, und es wird weder gesichert noch gewandert. launchd startet ihn mit dem Programm neu, auf das die plist zeigt; er wartet an `bestand.db.wanderung`. Die Vorprüfung nennt ihn deshalb nicht | im Protokoll „der Dienst wurde vor der Wanderung neu gestartet" — oder „… ließ sich … nicht anhalten" samt Grund |
| **ein Dienst oder Werkzeugserver ab 0.9.0** | **jede** Anfrage — lesend wie schreibend, auch die Wechsel von Profil und Nutzer und der Lauf eines gelösten Auftrags — hält die Sperre über die Wanderung **geteilt**, solange sie läuft, und fragt darunter, ob dieser Prozess den Bestand noch kennt: zwischen Frage und Schreiben wandert niemand. Hat eine neuere Fassung gewandert, bekommt die Anfrage **503 `belegt`** (am Werkzeugserver denselben Fehler), es wurde nichts geschrieben, und der Prozess **endet von selbst**. `mse-rest` endet dabei mit 1, und launchd startet den eingerichteten Dienst mit dem Programm im Bündel neu (`SuccessfulExit` ist `false`) — **frühestens zehn Sekunden** nach dem letzten Start (`ThrottleInterval`, ein Mindestabstand und keine Zusage); ein von Hand gestarteter Prozess bleibt aus | der Satz „Dieser Prozess stammt aus einer älteren Fassung …", in `dienst.log` der Grund des Endes |
| **ein Dienst ab 0.9.0**, den launchd über einem Bestand einer **neueren** Fassung startet | er weist den Bestand beim Start ab und endet **mit 0** — launchd wirft ihn dann nicht wieder an, statt ihn alle zehn Sekunden zu starten. **Ein Dienst aus 0.8.0 tut das nicht**: er endet mit 1, und launchd startet ihn im Abstand von zehn Sekunden wieder, bis die plist auf das neue Programm zeigt | in `dienst.log` „… stammt aus einer neueren Fassung …" (eine plist aus 0.8.0 schreibt nach `dienst.txt`) |

**Zurückgespielt wird erst, wenn niemand den Bestand mehr hält.** Hält ihn der
eingerichtete Dienst, sagt das Zurückspielen genau das und nennt den Weg, ihn
anzuhalten (*Einstellungen › Diagnose › Kommandozeile und Werkzeugserver
einrichten*) — **nur, wenn er es ist**, gefragt bei `lsof`; hält ein anderes
Programm den Bestand, bleibt die allgemeine Meldung. `mse
bestand-zurueckspielen` und `mse bestand-schauen` wandern einen Bestand einer
**neueren** Fassung vorher **nicht** — sie scheiterten daran bis M02 — und
fragen dann auch nicht nach Profil und Rolle, deren Schema diese Fassung nicht
kennt. Einen leeren oder älteren wandern sie wie jeder Befehl; auf einem
frischen Rechner spielt `mse bestand-zurueckspielen` so die Sicherung vom
alten zurück.

**Ein Abbruch mitten in der Wanderung** — auch `kill -9` — lässt den Stand
von vorher stehen: die Wanderung ist eine Transaktion. Der nächste Start
wandert zu Ende und nimmt **dieselbe** Sicherung von vorher, statt eine
zweite anzulegen. **Eine volle Platte** hält den Start vor der Wanderung an,
mit „auf dem Datenträger ist kein Platz mehr"; der Bestand bleibt bitgleich,
und eine halbe Sicherung bleibt nicht liegen. Beides belegt die Hüllenprobe
`wanderung09`.

### Und was eine Rückkehr unklar lässt

**Ein historisches Backup kann bereits abgegebene Karten enthalten**
([0088](entscheidungen/0088-die-epoche-einer-rueckkehr.md)). Vor der
Bestätigung stehen deshalb zwei Zahlenreihen nebeneinander:

- **was seit der Sicherung dazukam** und beim Zurückspielen verschwindet, und
- **was die Sicherung mehr hat** — diese Stücke stehen danach wieder in der
  Sammlung, ohne dass jemand weiß, ob sie noch jemandem gehören.

```bash
mse ruecksprung-lage ~/Sicherungen/bestand-2026-09-01.msesicherung
```

Der Rückgabewert ist **1**, wenn etwas unklar bliebe oder von Hand nachzusehen
ist — ein Skript, das danach weiterläuft, soll hier anhalten.

**Was verloren geht, lässt sich vorher ausfahren**: ein Set über *Datei →
Speichern unter* im Fenster oder als MTG JSON über `mse mtgjson-schreiben
<set> <spiel> <ziel.json>`, die Sammlung über `mse sammlung-ausgeben`.

Nach dem Zurückspielen steht die Rückkehr im Bestand:

```bash
mse rueckkehren
mse rueckkehr-quittieren 7
```

**Die Quittung bucht nichts ab.** Sie sagt: „ich habe die betroffenen Stücke
von Hand geprüft." Bis dahin sind sie nicht handelbar — und sie werden auch
nicht selbsttätig abgezogen, denn die Differenz kann ebenso daher rühren, dass
jemand die Sammlung aufgeräumt hat. **Eine Epoche verjährt nicht**, und über
MCP lässt sie sich nicht quittieren: nachsehen heißt in die Kiste greifen und
zählen.

---

## Aufräumen — ein Befund ist kein Auftrag

**Seit Werkstück O8 der Iteration 0.5.0**
(#1432, Vermerk
[0089](entscheidungen/0089-ein-befund-ist-kein-auftrag.md)).

```bash
mse pflege-vorschau
```

**Rein lesend**, und deshalb an jeder Fläche. Der Rückgabewert ist **1**, wenn
etwas dasteht — ein Skript soll den Fall erkennen, ohne den Text zu lesen.

Zehn Stellen, drei Klassen:

| Klasse | Heißt |
| --- | --- |
| `sicher` | die Zeile zeigt auf nichts und hält nichts fest — vorausgewählt |
| `entscheidung` | sieht nach einem Rest aus und **kann** ein Nachweis sein — nicht vorausgewählt |
| `nurInformieren` | wird **nicht** angefasst; es gibt dafür keine Löschabfrage |

**Was nach Dubletten aussieht, sind keine.** Zustand, Sprache, Ausführung,
Erwerb und Lagerort unterscheiden zwei Sammlungszeilen mit derselben
Kartennummer — die Sammlung wird von der Pflege **überhaupt nicht** angefasst.
**Und eine offene Rückkehr wird nie weggeräumt**: sie ist der Nachweis darüber,
dass etwas unklar ist ([0088](entscheidungen/0088-die-epoche-einer-rueckkehr.md)).

### Ausgeführt wird gegen eine Revision

```bash
mse pflege 2026-09-10T12:00:00Z kartenwerteOhneKarte:7324443718f9…
```

Der Zeitpunkt ist der aus der Vorschau — er trägt ihre **Frist** (eine
Viertelstunde). Die Revision ist die Prüfsumme über die betroffenen Zeilen;
hat sich seither etwas geändert, **stoppt die betroffene Maßnahme als Konflikt
— und nur sie**.

**Geprüft wird zweimal**: gegen die Vorschau und noch einmal **in derselben
Transaktion** wie das Löschen. Das sind zwei Fragen — „hat sich seit dem
Anschauen etwas geändert?" und „ändert sich gerade etwas?".

**Vorher wird gesichert**, nach `bestand-vor-pflege-<zeitpunkt>.db` neben dem
Bestand; **danach wird noch einmal nachgesehen**, und die Bilanz nennt beide
Stände. Ist nichts zu tun, wird auch nicht gesichert.

**Nicht über MCP.** `entscheidung` heißt „ein Mensch entscheidet", und die
Vorschau ist die Zustimmung. Nachsehen darf ein Modell überall.

---

## Was ein Auftrag draußen kosten darf

**Seit Werkstück O12 der Iteration 0.5.0**
(#1439, Vermerk
[0091](entscheidungen/0091-geld-ueberlebt-den-prozess.md)).

```bash
mse kosten anbieter
mse kostenbudget anbieter monat 5000 EUR
```

**Der Betrag steht in der kleinsten Einheit**: `5000 EUR` sind fünfzig Euro.
Ganzzahlig, weil zwischen einer Grenze und einer Summe keine Rundung stehen
darf.

### Vier Zahlen, und die vierte ist der Grund für das Ganze

| Zahl | Heißt |
| --- | --- |
| `reserviert` | vor dem Ruf gebucht, Ausgang offen |
| `verbraucht` | abgerechnet, der Betrag steht |
| `unklar` | die Frist lief **nach** der möglichen Annahme ab |
| `verfuegbar` | was von der Grenze übrig bleibt |

**`unklar` zählt wie verbraucht**, bis ein Mensch nachgesehen hat. Ein Auftrag,
über den nichts feststeht, als kostenlos zu führen, ist die Buchung, die eine
Rechnung überrascht. Auf der Kommandozeile trägt der Rückgabewert dieselbe
Aussage: **1**, wenn etwas ungeklärt ist.

### Ohne Freigabe geht nichts hinaus

**Null ist die Vorgabe**, und sie sperrt: für ein Konto ohne Budget wird kein
kostenpflichtiger Auftrag angemeldet. Wer eine Freigabe wegnehmen will, setzt
sie auf null.

**Jeder einzelne Auftrag verlangt zusätzlich einen
[Nachweis](entscheidungen/0078-der-wirksame-akteur.md)** — die Grenze sagt, wie
viel im Zeitraum insgesamt ausgegeben werden darf; der Nachweis sagt, dass ein
Mensch **diesen einen** wollte. Er wird einmal eingelöst.

### Zwei Prozesse, ein Topf

**Die Grenze gilt über alle lokalen Prozesse.** Fenster, Kommandozeile, MCP und
Dienst buchen in dasselbe Journal, und die Reservierung ist eine Transaktion:
wer als Zweiter kommt, sieht den Ersten.

**Und ein Abbruch bucht nicht zweimal.** Die Auftragsidentität kommt vom
Aufrufer; wer nach einem Absturz mit derselben wiederkommt, bekommt den
vorhandenen Posten zurück.

### Was das Journal nicht weiß

**Fremde Nutzung desselben Zugangs.** Wer denselben API-Schlüssel auf einem
anderen Rechner oder in einem anderen Programm benutzt, verbraucht Kontingent,
von dem hier nichts steht. Eine harte Zusage über das ganze Anbieterkonto
verlangte eine Sperre beim Anbieter; diese Grenze ist eine **lokale**.

**Und es gibt heute keinen zahlenden Anbieter.** Scryfall, GitHub und die drei
Kartendienste kosten nichts; das Journal steht für die Aufträge, die Phase 5.6
bringt.

**Nicht über MCP gesetzt.** Der Stand ist lesbar — ein Modell soll wissen, wie
viel übrig ist —, die Grenze nicht: wer sie selbst höhersetzen könnte, hätte
keine.

---

## Zurücksetzen — sechs Arten, und jede benennt, was sie anfasst

**Seit Werkstück O4 der Iteration 0.5.0**
(#1428, Vermerk
[0090](entscheidungen/0090-zuruecksetzen-ist-eine-familie.md)).

```bash
mse ruecksetzungsarten
mse ruecksetzung-vorschau profil
```

**„Zurücksetzen" ist kein Vorgang, sondern eine Familie.** Das Fenster wieder
aufräumen und einen Menschen samt allem löschen haben nichts miteinander zu tun
außer dem Wort; ein Knopf für beides ist die Gestalt, in der jemand das eine
drückt und das andere bekommt.

| Art | Nimmt weg | Sicherung angeboten |
| --- | --- | --- |
| `layout` | vier Einstellungen — Erscheinungsbild, Schriftgröße, Schriftart, Seitenleiste. **Keine Karte** | nein |
| `profileinstellungen` | die Wahl dieses Profils; danach wird wieder **geerbt** | nein |
| `nutzereinstellungen` | die Wahl dieses Menschen | nein |
| `profil` | das Profil samt Sets, Karten, Decks, Exemplaren, Token und Freigaben | ja |
| `nutzer` | den Menschen samt allen seinen Profilen, seinem Bild, seinem Tisch- und Fensterbild, seinen Abzeichen und seinem TOTP-Geheimnis | ja |
| `anwendung` | legt den ganzen Bestand **beiseite** und öffnet einen frischen | ja |

**Zurückgesetzt heißt gelöscht — und damit geerbt.** Eine Profileinstellung,
die es nicht gibt, erbt von der Nutzereinstellung; ein geschriebenes „erben"
wäre eine Wahl, die wie keine aussieht.

### Die Vorschau zählt und benennt

**Rein lesend**, und deshalb an jeder Fläche. „Dieses Profil und alles, was
dazugehört" ist keine Auskunft, mit der jemand entscheidet — „drei Sets, 107
Karten, zwei Token, ein Abzeichen" ist eine: gezählt wird über sieben Stellen,
benannt werden bis zu fünf je Gruppe.

**Was im Weg steht, steht daneben.** Fünf Hindernisse:

| Marke | Heißt |
| --- | --- |
| `offeneRueckkehr` | über Mengen dieses Profils steht nach einer Rückkehr nichts fest ([0088](entscheidungen/0088-die-epoche-einer-rueckkehr.md)). Der Weg heraus ist die **Quittung** |
| `fremdesZiel` | das Profil oder der Mensch gehört dem Rufenden nicht — und dann kommt die Vorschau **ohne Zahlen** zurück |
| `letzterNutzer` | der letzte Mensch lässt sich nicht löschen |
| `laufenderAuftrag` | in diesem Fenster läuft noch etwas. **„Geordnet stoppen" heißt hier: nicht mitten hineinlöschen** — es gibt keinen Abbruch, und einem laufenden Schreiber den Bestand wegzunehmen wäre schlimmer als zu warten. Was in einem **zweiten** Fenster läuft, sieht dieses nicht: das Auftragsbuch steht im Arbeitsspeicher je Prozess |
| `ohneTausch` | nur am Dienst, nur für die ganze Anwendung (G09, #2273): `mse-rest` hält seinen Bestand, solange er läuft, und ein Neuanfang ließe ihn an einer geschlossenen Verbindung zurück. Neu angefangen wird am Fenster oder mit der Kommandozeile |

### Die ganze Anwendung verlangt eine Freigabe

```bash
mse ruecksetzung-vorschau anwendung          # nennt die Revision
mse zuruecksetzen anwendung "" <nachweis> --revision <revision>
```

**Eine Rolle reicht dafür nicht**: sie nähme **anderen Eigentümern** ihre Daten
weg. Verlangt wird ein [Nachweis](entscheidungen/0078-der-wirksame-akteur.md) —
ausgestellt am Fenster, von einem Menschen, für diesen einen Zweck, mit Frist
und **einmal** einlösbar. Eine Maschine bekommt keinen; damit gibt es diese Art
an der Kommandozeile nur mit einem, der am Fenster entstanden ist, und am
Dienst gar nicht (`ohneTausch`, oben).

**Ausgestellt wird er unter *Einstellungen › Daten und Sicherung***, Abschnitt
*Zurücksetzen*: „Die ganze Anwendung" wählen, *Nachsehen, was betroffen
wäre*, dann *Freigabe erteilen*. Sie gilt fünf Minuten; darunter steht die Kennung, die die
Kommandozeile als
`<nachweis>` nimmt. **Bis zum 2026-10-10 stellte ihn keine Fläche aus** (G12,
#2282) —
die Seite fragte nach einer Kennung, die nirgends stand.

**Der Neuanfang ist ein Tausch und keine Löschung.** Der bisherige Bestand wird
nach `bestand-vor-rueckspielung.db` beiseitegelegt, und der nächste Start legt
einen frischen an; die Bilanz nennt den Ort. Eine Schleife über dreißig
Löschanweisungen wäre unumkehrbar **und** unvollständig — sie müsste jede
Tabelle aufzählen, und die nächste Wanderung fügte eine hinzu, die niemand
nachträgt.

**Was anderen Programmen gehört, bleibt liegen.** Der Schlüsselbund wird
**nicht** mit ausgeräumt: die Einträge gehören zum beiseitegelegten Bestand und
werden gebraucht, falls jemand ihn zurückholt. Wer ein Geheimnis loswerden
will, löscht **seinen Nutzer** — dann geht es mit, und nur seines.

**Nicht über MCP.** Dieselbe Klasse wie beim Aufräumen: die Ausführung löscht,
und die Vorschau ist die Zustimmung. Nachsehen darf ein Modell überall.

---

## Was über Fassungen hinweg gilt

Die Regeln stehen vollständig in
[0033](entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md). Was
beim Betrieb zählt:

| Fall                              | Was geschieht                                                                      |
| --------------------------------- | ---------------------------------------------------------------------------------- |
| Bestand aus **0.7.0**             | öffnet sich **ohne einen Schritt** — er steht auf der Grundlinie                   |
| Bestand von **vor 0.7.0**         | wird **mit Rat abgewiesen** und nicht angefasst: einmal mit 0.7.0 öffnen           |
| Bestand nach der Grundlinie       | wird beim Start migriert — **vorwärts, immer**                                     |
| Bestand einer **neueren** Fassung | wird **abgewiesen**. Eine verbuchte, aber nicht vorhandene Migration ist der Beleg |
| Migration bricht ab               | der **alte** Stand bleibt: alle Schritte laufen in **einer** Transaktion           |

**101 Migrationsschritte** sind es heute, und die ersten 97 stehen seit 0.8.0 als **eine
Grundlinie** da (`internal/storage/grundlinie/`,
[0189](entscheidungen/0189-nur-die-grundlinie.md)): ein neuer Bestand entsteht
in einem Zug, und seine Buchführung nennt alle 97 mit ihrer Prüfsumme — wie
bei einem, der sie einzeln gegangen ist. Was danach kommt — als erstes
`0098_sicherungsvermerke.sql` (M01), dann `0099_setstand.sql` und `0100_setstand_genauer.sql` (PERF02), `0101_gegenbelegbefunde.sql` (#2500) —, liegt wieder als
Datei unter `internal/storage/migrationen/` und läuft in einer Transaktion
mit der Prüfung darin.

**Ein Bestand von vor 0.7.0 wird nicht mehr nachgezogen.** Die Anwendung
weist ihn ab, **bevor** sie etwas an ihm ändert, und sagt, was zu tun ist:
einmal mit MagicSetEditor4Mac 0.7.0 öffnen — dort steigt er auf —, danach
öffnet ihn auch diese Fassung. Für eine Sicherung gilt dasselbe: mit 0.7.0
zurückspielen, dann hier öffnen.

**Der 22. Schritt räumte auf, was ein alter Bestand angesammelt hatte.** Das
Schema trägt `ON DELETE CASCADE` seit dem zweiten Schritt — **SQLite beachtet
Fremdschlüssel aber nur, wenn `foreign_keys` an ist**, und das kam erst am
2026-08-15 dazu. Bis dahin ließ jede Löschung ihre Werte stehen
(#1220).

**`mse diagnose` zählt sie seither.** Die Zeile *Verwaiste Zeilen* sagt
„keine" oder nennt die Zahl je Tabelle; steht dort etwas anderes als „keine",
ist das ein Befund und kein Betriebszustand.

**Es gibt keinen Weg zurück.** Wer eine ältere Fassung auf einem neueren
Bestand öffnet, bekommt eine Abweisung — deshalb die automatische Sicherung
davor.

**Und die Abweisung nennt sie beim Namen** (seit 0.4.0). Sie sieht neben dem
Bestand nach, **öffnet** jede `bestand-vor-*.db` nur lesend und nennt die
jüngste, die aufgeht — mit ihrem Schemastand und dem, was darin liegt:

```
Zurück kommst du über "bestand-vor-0.4.0.db" neben dem Bestand —
Schemastand 0035_platz.sql, 12 Set(s) mit 840 Karten
```

Am Fenster steht derselbe Satz auf der Wiederherstellungsseite unter *Der Weg
zurück*; die Kommandozeile und die beiden Dienste hängen ihn an ihre Meldung.
**Liegt keine brauchbare da, steht dort nichts** — ein Hinweis auf eine Datei,
die es nicht gibt, ist schlimmer als keiner.

**Und jede gelaufene Migration wird mit einer Prüfsumme verbucht.** Ändert sich
ihre Datei danach, meldet der nächste Start „das Schema dieser Datenbank stimmt
mit keiner Fassung überein" und öffnet **gar nichts**. Das ist die Zusage und
kein Fehler: zwei Rechner sollen unter derselben Nummer nicht zwei verschiedene
Schemata bauen.

**Beim Entwickeln trifft das den eigenen Bestand**, und die Meldung nennt die
Ursache nicht. Die Summe geht über die **ganze Datei** — ein geänderter
Kommentar genügt. Wer eine Migration nach ihrem ersten lokalen Lauf noch
anfassen muss, belegt zuerst, dass nur Kommentare sich geändert haben (alte
Fassung rekonstruieren, `shasum -a 256` gegen `SELECT pruefsumme FROM
migrationen WHERE name = …` halten), sichert `bestand.db` und zieht die Summe
dort mit einem `UPDATE` nach. **Für ausgelieferte Fassungen gilt das nicht**:
dort ist eine gelaufene Migration unveränderlich.

---

## Der Abbruch

**Was geschieht, wenn die Anwendung mitten in einem Schreibvorgang endet?** Drei
Fälle, und sie sind nicht dasselbe.

### ⌘Q

**Die Anwendung schließt den Bestand — in höchstens fünf Sekunden.** Der Weg
unter Wails 3 (`v3.0.0-beta.27`, nachgelesen am 2026-10-15 in
`internal/fenster/wails.go`; in beta.30 unverändert, verglichen am 2026-10-20): macOS fragt `applicationShouldTerminate:`, Wails
ruft den Rückruf `OnShutdown`, und den setzt diese Anwendung
(`schlussFahren`). Darin geht **das Tor zu** — jede gebundene Methode weist
danach mit „die Anwendung endet" ab —, die Fenstergröße wird gemerkt
(höchstens eine Sekunde, `Bestenfallsfrist`), **die eigenen Läufe werden
beendet und abgewartet**, dann geht der Bestand zu; `fm serve` endet im
selben Rückruf. **Alles zusammen in der `Schlussfrist` von fünf Sekunden**
(`internal/app/lebenszyklus.go`); was danach noch läuft, findet den Bestand
geschlossen. **`Run` kehrt dabei nicht zurück** — das `defer` in `main.go` ist
nicht der Weg, auf dem der Bestand zugeht.

Hier stand bis 0.9.0 der Weg unter Wails 2 — `wails.Run` kehrt zurück,
`OnShutdown` läuft nicht. Befund A-22 des Audits 0.9.0 hat es angestrichen
(OPS04, #2350).

**Ein Schreibvorgang, der in der Frist nicht fertig wird, wird nicht zu Ende
geführt.** Was festgeschrieben war, bleibt; was nicht, rollt SQLite zurück.
Eine **Massenänderung** wird Zeile für Zeile angewandt — ein ⌘Q bei Zeile 250
von 500, die in fünf Sekunden nicht fertig wird, lässt die geschriebenen
geändert. Das ist Absicht — eine Massenänderung, die wegen einer einzigen
Zeile ganz zurückrollte, wäre die schlechtere Antwort —, und der Bericht
darüber ist dann fort.

### Absturz, Abschuss, Stromausfall

**Die Sperrdatei stirbt mit dem Prozess**, und das ist richtig: eine Sperre, die
einen Absturz überlebte, machte den Bestand für immer unbenutzbar.

**Was quittiert war, bleibt.** `journal_mode = wal` und `synchronous = FULL`,
beides unter einem Wächter auf **jeder** Verbindung: eine Änderung ist auf der
Platte, bevor die Antwort geht. Das gilt seit 0.4.0 auch bei Stromausfall —
davor stand `synchronous` bei NORMAL, und welche Verbindung ein Schreibvorgang
aus dem Vorrat bekam, entschied über seine Dauerhaftigkeit.

**Und die Dateien neben dem Bestand ebenso.** Bilder, Pakete, `.mse-set` und
die Rückspielmarke werden über eine Nachbardatei geschrieben und umbenannt —
ein Abbruch lässt dann eine sichtbare Datei mit führendem Punkt liegen, nicht
ein halbes PNG unter dem richtigen Namen. Ein Wächter hält es
(`internal/werkzeuge/atomarewege_test.go`).

**Und beim nächsten Start geht sie** — OPS04
(#2350),
Befund A-27. Geräumt wird **vor dem Öffnen und nur, wenn kein anderer
Prozess den Bestand hält**; im Bestandsordner und in `beiwerk/`, oben und eine
Ebene tiefer; was **länger als eine Stunde** unverändert liegt und heißt, wie
diese Anwendung vorläufig schreibt — `.<name>.<zufall>` oder
`.<name>-<zufall>` (`internal/service/restepflege.go`). Eine fremde Datei,
ein Ordner ohne Punkt und alles, was jünger ist, bleibt; ein Verweis wird
genommen, sein Ziel nicht. **Eine Sicherung ohne Kennwort entsteht neben
ihrem Ziel** — dorthin greift das Räumen nicht.

### Ruhezustand mit offener Datenbank

**Der Bestand liegt lokal**, und SQLite hält über Untätigkeit keine Sperre. Was
hier steht, ist **nicht** in einem Lauf hergestellt — ein Ruhezustand lässt sich
in einem Testlauf nicht nachstellen; an seine Stelle treten die gemessenen
Einstellungen oben.

**Auf einem Netzlaufwerk gilt nichts davon**, und das ist keine Einschränkung
dieser Anwendung, sondern von SQLite: dessen Sperren sind über NFS und SMB
nicht verlässlich. Der Bestand gehört auf die eigene Platte.

---

## Protokoll

**Drei Kanäle, unabhängig wählbar** (OPS-04,
#1721):
die Datei `mse.log`, das **Unified Log** von macOS und die
Standardfehlerausgabe. Vorgabe ist, was es immer gab — Datei neben dem Bestand
und Unified Log, Stufe *Ereignisse*. Unter *Einstellungen › Diagnose* oder so:

```bash
mse einstellungen setzen protokoll.stufe ausfuehrlich
```

| Einstellung | Was | Vorgabe |
| --- | --- | --- |
| `protokoll.stufe` | `fehler`, `warnung`, `info`, `ausfuehrlich` | `info` |
| `protokoll.datei`, `protokoll.system`, `protokoll.ausgabe` | je `an` oder `aus` | an, an, aus |
| `protokoll.format` | `text` zum Lesen, `json` für Werkzeuge | `text` |
| `protokoll.grenze` | Größe einer Datei in MB, 1 bis 64 | 4 |
| `protokoll.aufbewahrung` | ältere Dateien, 1 bis 9 | 1 |
| `protokoll.komponenten` | bei `ausfuehrlich` nur diese Pakete, etwa `app` — heute schreibt nur `app` ausführliche Zeilen, die Rufe des Fensters | alle |
| `protokoll.ort` | Ordner der Datei, mit `/` oder `~/` vorn | neben dem Bestand |

**Alle gelten ab dem nächsten Start — und für das Fenster.** Das Protokoll
entsteht vor dem Bestand — was beim Öffnen schiefgeht, soll darin stehen —,
und seine Einstellungen liegen im Bestand; bis er offen ist, gilt die Vorgabe.
**Dienst und Werkzeugserver** schreiben auf ihre Standardfehlerausgabe,
geschwärzt wie unten und mit der Vorgangskennung des Rufs; wohin das geht,
entscheidet, wer sie startet. Bis zum Zweitblick über Phase 6.4 schrieben sie
ungeschwärzt und ohne Kennung, während hier stand, jede Zeile trage sie.

**Ausführlich schreibt eine Zeile je Ruf des Fensters** — die Bindung und
die Vorgangskennung, unter der Komponente `app`; wer `protokoll.komponenten`
setzt und `app` weglässt, sieht sie nicht. Bis zum Zweitblick über Phase 6.4
war die Stufe wählbar, und es gab keine Zeile, die sie zeigte.

**Ein Kanal, der ausfällt, nimmt die anderen nicht mit.** Lässt sich die
Datei nicht anlegen, schreiben Unified Log und Standardfehlerausgabe weiter,
und die Diagnose nennt den Grund („Kanäle: Datei fehlt (…), Unified Log an,
…"). Bis 0.6.0 verstummte in diesem Fall das ganze Protokoll.

**Ausführlich heißt: kurz aufbewahrt.** Höchstens 1 MB je Datei und eine
ältere Datei, gleich was eingestellt ist; die Diagnose sagt, dass es an ist.
**Was darin steht, bleibt geschwärzt** wie auf jeder Stufe (unten).

**Wie lange es bleibt**: bis zur Grenze. Dann wird es zu `mse.log.1`, die
ältere zu `mse.log.2` und so fort bis zur eingestellten Aufbewahrung, und ein
neues beginnt. **Gemessen am 2026-09-07**: 84 kB am Tag, also mit der Vorgabe
rund sieben Wochen je Datei und gut drei Monate mit der Vorfassung.

**Gedreht wird beim Start und während der Sitzung** (PERF03,
#2353).
Bis dahin galt die Grenze nur beim Start, und ein Fenster, das eine Woche
offen stand, schrieb eine Datei ohne Ende. Jetzt sieht jeder Schreiber vor
jeder Zeile nach, ob am Ort noch seine Datei steht und ob sie über der
Grenze ist; dann dreht er unter `mse.log.drehen` und öffnet neu, und ein
anderer Prozess, der dieselbe Datei schreibt, folgt ihm mit der nächsten
Zeile. **Über die Grenze geht eine Datei damit höchstens um eine Zeile je
Schreiber.** Mit der Vorgabe sind es zwei Dateien zu 4 MB, auf *ausführlich*
zwei zu 1 MB; wer `protokoll.grenze` und `protokoll.aufbewahrung` höher
stellt, bekommt, was er einstellt — bis 64 MB und zehn Dateien. Gemessen:
das Nachsehen und die Zeile je Ruf kosten auf *ausführlich* im p95 rund 2 µs
je Ruf (`ausfuehrlichkosten_intern_test.go`).

**Das Unified Log hält niemanden auf.** Ein laufender `logger` liest die
Zeilen von seiner Standardeingabe; wer schreibt, legt die Zeile ab und
geht. Nimmt `logger` nichts an, warten höchstens 256 Zeilen, und was darüber
kommt, wird verworfen und gezählt — die Diagnose nennt das Unified Log dann
nicht „an". Ein gestorbener `logger` startet mit der nächsten Zeile neu, und
beim Beenden gehen die wartenden Zeilen noch hinaus, höchstens fünf Sekunden
lang. Bis PERF03 startete jede Zeile einen eigenen `logger` und wartete auf
ihn.

**Was eigens gedreht wird**: die Ausgabe des eingerichteten Dienstes. Seine
Ansage samt Marke ersetzt `mse-rest -ansage` je Start in `dienst.txt` (0600,
Ordner 0700) — es steht immer nur die gültige Marke darin, und ein Widerruf
nimmt die Datei mit. Fehler und Ende gehen nach `dienst.log`; über 1 MiB
beginnt die Datei beim nächsten Start neu, mit einer Zeile, die das sagt
(SEC02, #2355,
[0192](entscheidungen/0192-die-dienstmarke-lebt-so-lange-wie-ihr-prozess.md)).
**Bis 0.9.0** hängte launchd beides an `dienst.txt` an, und jede Marke seit der
Einrichtung stand darin; eine plist aus dieser Zeit nennt die Seite
*Schnittstellen* **veraltet**, bis sie neu eingerichtet ist.

**Was darin steht**: nur angemeldete Merkmale. Der Pfad des
Benutzerverzeichnisses wird zur Tilde. **Und die Gestalt eines Nachweises
wird zu `«geschwärzt»`** — alles hinter `Authorization:` oder `Bearer` und
jede Folge aus mindestens 32 Buchstaben **und** Ziffern, und seit SEC02
jede aus mindestens 32 Hexziffern allein —, in jedem Kanal:
Datei, Unified Log, Standardfehlerausgabe und die Diagnose (SEC-P5,
#2001).
Bis dahin tat das nur das Diagnosepaket. **Der Preis**: eine Prüfsumme in
einer Protokollzeile ist danach nicht mehr lesbar.

**Der Wert einer gefangenen Panik steht darin ganz** — unter `panik`, zur
Diagnose am eigenen Rechner; **in die Diagnose und das Diagnosepaket** geht er
nur in der Gestalt, die das Laufzeitsystem schreibt (`runtime error: …`),
sonst `«geschwärzt»` — dieselbe Zusage wie für die Absturzspur (#2399,
[0193](entscheidungen/0193-der-paniktext-bleibt-am-rechner.md)).

**Was darin nicht steht**: Kartentexte, Feldwerte, Setinhalte.

Das Protokoll ist das, was jemand einem Fehlerbericht beilegen kann.

**Und dafür braucht es kein Terminal.** Unter _Einstellungen → Technik_ steht
**Diagnose anzeigen**: Fassung, Rechnerart, der Ort des Bestands und die
letzten 200 Protokollzeilen, als Text zum Lesen — mit _Kopieren_ und _Als Datei
sichern_ daneben. Der Text steht da, **bevor** jemand ihn weitergibt; die
Anwendung schickt nichts von selbst hinaus
(#1166).

Bis dahin stand hier nur die Zeile unten, und die Zielgruppe dieser Anwendung
kann sie nicht: sie mag Magic und sammelt Karten. Ein Weg, den nur bedienen
kann, wer ohnehin ein Terminal hat, ist für einen Fehlerbericht der falsche.

Wer eines hat, hat es weiterhin leichter:

```bash
tail -n 200 ~/Library/Application\ Support/MagicSetEditor4Mac/mse.log
```

---

## Messen

**Was jeder Prozess über sich sagt** (OPS-05,
#1722) —
unter _Einstellungen › Diagnose_ für das Fenster, über `GET /metriken` für den
Dienst und `lesen.metriken` für den Werkzeugserver. **Jeder Prozess zählt für
sich**, im Arbeitsspeicher und ab seinem Start; nichts davon geht von selbst
hinaus, und es gibt keinen Versand an einen fremden Dienst.

| Zahl | Einheit | Woher |
| --- | --- | --- |
| Rufe je Fläche und Art, abgewiesen, gescheitert | Anzahl | Dienst und Werkzeugserver, je Ruf |
| Verteilung der Laufzeiten | Rufe je Eimer bis 1, 5, 25, 100, 500, 2500 ms und darüber | ebenda |
| Verbindungen zum Bestand: offen, in Benutzung, gewartet | Anzahl, Wartezeit in ms | `database/sql` |
| Zwischenspeicher fremder Dienste | Treffer, geholt, veraltet ausgeliefert, fehlgeschlagen | `internal/netz` |
| Arbeitsspeicher, nebenläufige Abläufe, Rechenzeit | Bytes, Anzahl, ms | Go und `getrusage` |
| Erhebungszeitraum und Aufwand | Zeitpunkte, µs | die Antwort selbst |

**Was nicht gemessen ist, steht in der Antwort** (`ungemessen`): die
Bindungen des Fensters — sie haben keine gemeinsame Stelle, an der ein Ruf
endet — und die Kommandozeile, die einen Befehl lang lebt.

### Die Ablaufspur

**Eine Zeile je Ruf** mit Vorgangskennung, Fläche, Weg, Dauer und Ausgang —
**der Weg als Muster** (`GET /sets/{setID}/karten`) oder Werkzeugname, nie
der rohe Pfad, und nie ein Rumpf oder Parameter. **Sie ist aus**, bis jemand
sie einschaltet:

```bash
mse einstellungen setzen messung.spur an
```

| Einstellung | Was | Vorgabe |
| --- | --- | --- |
| `messung.spur` | `an` oder `aus` | `aus` |
| `messung.abtastung` | Anteil der Rufe in Prozent, 1 bis 100 — genau, jeder n-te | 100 |
| `messung.spurgrenze` | Einträge, 10 bis 10000; der älteste geht | 1000 |
| `messung.aufbewahrung` | Minuten, 1 bis 1440 | 60 |
| `messung.profile` | Leistungsprofile, die liegen bleiben, 1 bis 20 | 3 |

**Die Spur gilt ab dem nächsten Start** des Prozesses, der misst, und lebt
so lange wie er — sie ist ein Ring im Arbeitsspeicher, keine Datei. **Am
Fenster bleibt sie leer**: seine Bindungen sind ungemessen (oben), und die
Diagnose zeigt dort nur, ob sie an ist. Wer die Spur des Dienstes behalten
will, sichert die Antwort von `GET /metriken`; das Diagnosepaket trägt die
Metriken des **Fensters**.

**Eine Kennung von außen**: wer den Dienst ruft, darf `X-Vorgang` mitgeben —
acht bis zweiunddreißig Kleinbuchstaben aus `0-9a-f`. Sie steht dann in der
Antwort, in der Spur und in jeder Protokollzeile, die der Ruf schreibt; eine
andere Gestalt wird nicht übernommen, der Dienst vergibt dann seine eigene.
**Der Werkzeugserver über HTTP nimmt sie genauso**; über stdio gibt es keinen
Kopf, und er vergibt sie selbst.

### Das Leistungsprofil

**Ein CPU-Profil des Fensters** über 1 bis 60 Sekunden, unter
_Einstellungen › Diagnose_ — im Format von `go tool pprof`, im Ordner
`Leistungsprofile` neben dem Bestand. Die Abtastrate ist die von Go, hundert
Mal je Sekunde. **Nur am Fenster**: ein Profil misst den Prozess, in dem es
läuft, und Dienst und Werkzeugserver ein Minutenprofil über eine Netzfläche
abzuverlangen, gäbe jedem Token einen Weg, sie zu bremsen.

```bash
go tool pprof -top ~/Library/Application\ Support/MagicSetEditor4Mac/Leistungsprofile/mse-*.pprof
```

---

## Netz

**Die Anwendung arbeitet ohne Netz.** Was hinausgeht, geht nur auf
ausdrückliche Handlung:

| Wohin              | Wofür                                          | Wann                                                                                                                                                                                                                                                         |
| ------------------ | ---------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| die Paketquelle    | Vorlagenpakete beziehen                        | wenn jemand auf _Pakete_ etwas holt                                                                                                                                                                                                                          |
| `api.scryfall.com` | Kartendaten und Kartenbilder                   | wenn jemand einen Scryfall-Weg benutzt                                                                                                                                                                                                                       |
| `api.github.com`   | nach einer neueren Fassung sehen und sie holen | im **Fenster** nur, wenn _Nach neuen Fassungen sehen_ nicht auf `aus` steht. Ohne Fenster auf Aufruf: `mse aktualisierung`, `GET /aktualisierung`, `POST /aktualisierung/laden` — dort **ohne** Rücksicht auf die Einstellung, denn ein Aufruf ist die Frage |

Ein Ausfall hält das Bearbeiten nicht an: der letzte gute Stand bleibt, und er
sagt, wie alt er ist.

Wer das nachprüfen will: `internal/schichten/netz_test.go` führt auf, welche
Pakete überhaupt ins Netz greifen dürfen. Ein neuer Eintrag dort ist eine
Entscheidung, kein Nebeneffekt.

**Und wohin** (SEC-P4,
#2000):
kein Weg verbindet sich mit `0.0.0.0`, `169.254.0.0/16` — dort stehen
Metadatendienste — oder einer Gruppenadresse, geprüft nach der Auflösung und
bei jeder Weiterleitung. **Ein Paket kommt nur von einer öffentlich
gerichteten Adresse** — nicht von diesem Rechner und nicht aus dem eigenen
Netz, denn die Adresse nennt, wer `POST /pakete` ruft, und das darf ein Token.
Ein Git-Server im eigenen Netz ist damit keine Paketquelle. **Die Adresse der
KI und den Proxy setzt nur ein Mensch** — am Fenster, an der Kommandozeile
oder mit der Prozessmarke —, nie ein Token; ein Ollama auf einem zweiten
Rechner im eigenen Netz bleibt erreichbar. Was im Einzelnen gilt, steht in
[`NETZEIGENSCHAFTEN.md`](NETZEIGENSCHAFTEN.md).

**`MSE_NETZ_SCHLEIFE=an` öffnet die Schleife für den Paketbezug** — für
Prüfläufe, deren Gegenstelle auf `127.0.0.1` steht, und für nichts sonst.
Das eigene Netz öffnet sie nicht.

---

## Betriebszustand

**Seit Werkstück O9 der Iteration 0.5.0**
(#1433)
beantwortet jeder Prozess zwei Fragen — und sie sind **nicht** dieselbe:

| Frage | Antwort |
| --- | --- |
| **Lebt er?** | Wer die Antwort bekommt, hat den Beweis in der Hand |
| **Kann er arbeiten?** | Das hängt am **Bestand** — und nur an ihm |

**Ein fremder Dienst macht diesen Prozess nicht unbereit.** Wenn Scryfall
nicht antwortet, steht das in der Auskunft — die Bereitschaft nimmt es nicht.
Das ist dieselbe Zusage wie oben unter *Netz*: ein Ausfall hält das Bearbeiten
nicht an.

**Drei Gründe, nicht bereit zu sein**, und jeder ist eine Marke, kein Satz:

| Marke | Heißt |
| --- | --- |
| `bestandZu` | es ist kein Bestand geöffnet — der Normalfall beim Start |
| `wanderungLaeuft` | die Migrationen laufen gerade |
| `bestandUnlesbar` | der Bestand ist da und antwortet nicht |

**An allen vier Flächen**, und jede in ihrer Form:

```bash
mse gesundheit          # Rückgabewert 0 bereit, 1 nicht bereit
```

```bash
curl -s -H "Authorization: Bearer $MSE_MARKE" http://127.0.0.1:7777/gesundheit
curl -s -o /dev/null -w '%{http_code}\n' \
  -H "Authorization: Bearer $MSE_MARKE" http://127.0.0.1:7777/bereit   # 200 oder 503
```

Über MCP heißt das Werkzeug `lesen.gesundheit`; im Fenster steht der Zustand
unter *Einstellungen → Technik*, über der Diagnose.

**Beide Wege des Dienstes stehen hinter der Marke.** Er hört **vorgegeben** auf
`127.0.0.1` — die Schleife ist eine Vorgabe und keine Mauer
([0115](entscheidungen/0115-die-schleife-ist-eine-vorgabe.md)): über die
Lauschwahl `netz.lauschen` (G08, #2272)
lauscht er auch auf weiteren Adressen, und **sobald eine die Schleife
verlässt, gelten TLS und Marke ohne Ausnahme**. Er weist ab, was nach einem
Browser aussieht; einen
Orchestrator, der eine Sonde ohne Nachweis bräuchte, gibt es hier nicht — und
„wandert gerade" ist eine Auskunft über diese Datei.

## Wer antwortet hier

**Jede Fläche ohne Fenster sagt, welches Programm sie ist** — Fassung,
Plattform, Go-Fassung und die Fläche, über die der Ruf kam:

```bash
mse meta
```

```bash
curl -s -H "Authorization: Bearer $MSE_MARKE" http://127.0.0.1:7777/meta
```

Über MCP heißt das Werkzeug `lesen.meta`. **Im Fenster steht dasselbe im
Diagnosetext** unter *Einstellungen → Technik*.

**Ohne Bestand, und das ist der Punkt**: „welche Fassung bist du" muss auch
dann eine Antwort haben, wenn der Bestand nicht aufgeht.

**Was eine Fläche kann, steht nicht darin.** Das sagt sie selbst — `tools/list`
über MCP, `mse hilfe` an der Kommandozeile, `GET /grenzen` und
`GET /dienstkatalog` am Dienst.

---

## Einen Fehler nachsehen — die Vorgangskennung

**Jeder Ruf in diese Anwendung bekommt an seiner Grenze eine Kennung**: acht
Hexziffern, gewürfelt am Fenster, an der Kommandozeile, am MCP-Server oder am
Dienst ([0087](entscheidungen/0087-eine-kennung-je-vorgang.md)). Sie steht in
**jeder** Protokollzeile dieses Rufs — und sie kommt mit seiner Antwort zurück.

Das ist der Weg zu dem, was die Oberfläche absichtlich **nicht** zeigt: ein
Fehler nennt einen Satz für Menschen, der technische Grund steht im Protokoll
(§8).

| Fläche | Wo die Kennung steht | Wo die Protokollzeilen stehen |
| --- | --- | --- |
| Fenster | am Fehler selbst — *Einstellungen › Diagnose* zeigt das Protokoll dazu | `mse.log` und das Unified Log, wie eingestellt |
| Kommandozeile | als letzte Zeile `Vorgang: …`, wenn ein Befehl fachlich scheitert | auf ihrer Standardfehlerausgabe, geschwärzt |
| MCP | am Fehlertext des Werkzeugs, als `(Vorgang: …)` | auf der Standardfehlerausgabe des Servers, geschwärzt — **vom Fenster gestartet in `mse.log`** |
| Dienst | als Kopfzeile `X-Vorgang` an **jeder** Antwort, auch der geglückten | dito |

**Bis zum Nachlauf über Phase 6.4** gingen die Zeilen eines vom Fenster
gestarteten Servers ins Nullgerät, und die Kommandozeile schrieb ungeschwärzt
und ohne Kennung. Nachgesehen wird im Protokoll des Fensters:

```bash
grep "a1b2c3d4" ~/Library/Application\ Support/MagicSetEditor4Mac/mse.log
```

— ohne `kennung=` davor: die Zeile eines Servers steht dort als Wert eines
Merkmals `zeile`, und seine Kennung darin.

**Das Protokoll sieht nur, wer Verwalter des gezeigten Profils ist.** Es gehört
dem **Prozess** und nicht dem Profil; wer als Gast in einem fremden arbeitet,
läse darin die Spuren seines Gastgebers. Der Rest der Diagnose — Fassung,
Rechnerart, Ort des Bestands, Schema- und Paketstände — steht jedem offen: er
ist eine Auskunft über die Installation und über niemanden.

---

## Ein Diagnosepaket weitergeben

**Für einen Fehlerbericht, der mehr braucht als einen Text** — Werkstück O10
von Phase 5.2
(#1434).
Unter *Einstellungen → Technik* stellt *Zusammenstellen* ein Paket zusammen,
zeigt es und schreibt es erst auf *Paket sichern*; *Paket löschen* nimmt es
wieder weg.

**Nur im Fenster.** Wer eine Kommandozeile hat, hat `tail`, `zip` und
`shasum` — und seit O9 `mse meta`, `mse gesundheit` und `GET /metriken`, also
jeden Teil einzeln. Und die Vorschau ist eine **Zustimmung**: zustimmen kann
nur jemand, der da ist.

**Was darin steht** — fünf Teile, dazu die Übersicht und ein Manifest mit
Größe und SHA-256 je Eintrag:

| Datei | Inhalt |
| --- | --- |
| `uebersicht.json` | dieselbe Vorschau, die vor dem Schreiben dastand |
| `fassung.txt` | Fassung, Plattform, Go-Fassung |
| `bestand.txt` | Ort, Schemastand, verwaiste Zeilen — **kein Inhalt** |
| `pakete.txt` | die eingebundenen Vorlagenpakete mit Stand |
| `metriken.json` | der Zählstand dieses Prozesses (O9) |
| `protokoll.txt` | die letzten Protokollzeilen, geschwärzt |
| `diagnose.json` | das Manifest |

**Was ausdrücklich nicht darin steht**: Karten, Feldwerte, Bilder, die
Sammlung, private Notizen, Token, Kennwörter, TOTP-Geheimnisse, Antworten
fremder Anbieter. Die Absicherung ist die **Auswahl** — das Paket liest diese
Stellen gar nicht —, und die Schwärzung der Protokollzeilen ist die zweite
Reihe: ein Nachweis in einer Kopfzeile und eine lange Folge aus Buchstaben und
Ziffern fallen heraus. **Der Preis steht dazu**: eine Prüfsumme in einer
Protokollzeile fällt mit, denn sie sieht aus wie ein Geheimnis.

**Der Protokollteil hängt an der Rolle** (O9): wer als Gast in einem fremden
Profil arbeitet, bekommt ihn nicht — und statt seiner steht der Grund im
Paket, damit man ihn vermisst.

**Nichts geht von selbst hinaus.** Es gibt keinen Weg, der ein Paket
verschickt; es entsteht dort, wo jemand hinzeigt. Es **nimmt kein bestehendes
weg**, und ein Abbruch lässt keine halbe Datei unter dem richtigen Namen
zurück: geschrieben wird daneben, benannt erst am Ende.

---

## Was gewirkt wurde — die Metriken

**Gezählt wird jede geprüfte Wirkung**, nach drei Merkmalen: **Fläche**
(Fenster, Kommandozeile, MCP, Dienst), **Wirkungsklasse** (`lokalLesen` bis
`verbindlich`, siehe [`AUTORISIERUNG.md`](AUTORISIERUNG.md)) und **Ausgang**
(gewährt oder abgewiesen).

**Alle drei sind Aufzählungen, und das ist der Punkt**: es gibt höchstens
**vierzig** Reihen, gleich wie viele Karten, Sets oder Profile im Bestand
liegen. Eine Kennung als Merkmal gibt es nicht — sonst wüchse die Tabelle mit
dem Bestand und wäre nach einem Tag Arbeit unlesbar.

```bash
curl -s -H "Authorization: Bearer $MSE_MARKE" http://127.0.0.1:7777/metriken
```

Über MCP heißt das Werkzeug `lesen.metriken`; im Fenster steht die Tabelle
unter *Einstellungen → Technik*, unter dem Betriebszustand.

**Nicht auf der Kommandozeile**, und der Grund ist technisch: dort lebt ein
Prozess genau einen Befehl lang, und ein Zähler über seine Lebenszeit wäre
immer leer.

**Was nicht gezählt wird**: reines Lesen. Es geht an keinem Riegel vorbei, und
diese Zahlen beantworten „was hat dieser Prozess **gewirkt**", nicht „wie viele
Rufe kamen an".

**Und ein Ruf, den der Zwischenspeicher beantwortet, zählt nicht.** Der Riegel
hängt am Netzklienten, der Zwischenspeicher liegt darüber: wer zweimal dieselbe
Karte sucht, sendet einmal — und die Zahl sagt das.

---

## Grenzen

**Seit Werkstück O5 der Iteration 0.5.0**
(#1429)
hält jeder Prozess dieser Anwendung Grenzen, die Fenster, Kommandozeile,
MCP-Server und Dienst teilen. Sie stehen **an einer Stelle** —
`domain.Vorgabegrenzen` — und jede Fläche nennt sie: `mse grenzen`,
`lesen.grenzen`, `GET /grenzen`, und in der Hilfe unter *Was hinausgeht und
hereinkommt*.

| Grenze | Wert | Was sie tut |
| --- | --- | --- |
| **Kosten je Ruf** | Lesen 1, Schreiben 2, Bild 3, Rastern 5, Ausfuhr 8, Einfuhr 8, Masse 10, Abbruch 0 | Einheiten, die ein Ruf vom Guthaben nimmt — Verhältnisse, keine Messungen. **Welche Fähigkeit teuer ist, sagt eine Tafel** (`kostentafel` in jeder Auskunft): Karte zeichnen, Bogen, MTG JSON, Sammlung ein- und ausgeben, Paket einlesen, Massenänderungen, Bilder — über REST, MCP und Kommandozeile dieselben Kosten (#2003). **Auch die lesenden `POST`-Wege** — Vorschauen und Deutungen, die einen Rumpf mitnehmen — stehen darin, mit der Klasse Lesen: `POST /einstellungen/vorschau` kostet wie `lesen.einstellungen.vorschau` 1 (#2020). Was nicht darin steht, kostet, was Lesen oder Schreiben kostet. **Ein Abbruch kostet nichts** und wird nie gedrosselt |
| **Guthaben je Principal** | 120 Einheiten Decke, 120 je Minute Zulauf | je Zugriffstoken beziehungsweise je Prozessmarke; je MCP-Sitzung eines. 60 Schreiben oder 120 Lesen in der Minute — die Parität zur Drossel von Phase 4 |
| **Guthaben global** | 600 Einheiten, 600 je Minute, **Reserve 120** | alle Maschinenzugänge zusammen; unter die Reserve kommt nur das Fenster |
| **Plätze je Ressourcenklasse** | Rastern 2, Einfuhr 1, Ausfuhr 1, Bild 4, Masse 1 | so viele Arbeiten einer Klasse laufen zugleich — gleich, welche Fläche sie anstößt. Wer keinen Platz bekommt, wartet, bis einer frei wird oder sein Ruf endet. **Rastern belegt jede Rasterung einer Karte** — Kartenbild, Bogen (je Karte), Vorschau, Tisch — und nicht das Rastern eines einzelnen Symbols (#2020) |
| **Rumpf und Antwort** | 32 MiB | REST weist einen größeren Rumpf ab und bricht eine größere Antwort ab, statt sie zu kürzen |
| **Verbindungen je Dienst** | 16 | `mse-rest` und `mse-mcp` über HTTP nehmen so viele Verbindungen zugleich an; eine weitere wartet, und eine ruhende gibt ihren Platz her, sobald jemand wartet (`verbindungen`) |
| **Aufträge des Dienstes** | 4 laufend, 16 behalten, 1 Stunde | was `mse-rest` mit `Prefer: respond-async` nebenher fährt; ein fünfter bekommt `429` (`auftraege`) |
| **Sperre gegen das Raten** | drei freie Versuche, dann 1 s, verdoppelt bis 5 min | je **vorgezeigtem Nachweis**, nicht je Rechner: wer `mse_falsch` rät, sperrt `mse_falsch` — der gültige Client merkt nichts. Dazu bleibt die Drossel für alle Ratenden zusammen: 60 in der Minute |

**Abschalten lassen sich nur die beiden Guthaben** — für eine kontrollierte
Umgebung, in der niemand einem anderen etwas wegnimmt: ein Lastlauf, ein
Stapel über Nacht
([0161](entscheidungen/0161-die-kontingente-lassen-sich-abschalten.md),
QA-08.AK-03):

```bash
mse einstellungen setzen grenzen.kontingente aus   # gilt ab dem nächsten Start eines Dienstes
mse grenzen                                         # nennt »kontingente aus«
```

Im Fenster steht derselbe Schalter unter *Einstellungen › Verbindungen*,
im eingeklappten Kasten *Was Fachleute hier sonst noch einstellen*. **Alles andere in der Tafel bleibt**:
Plätze, Rumpf und Antwort, die Sperre gegen das Raten samt der Drossel für
alle Ratenden, dazu Zugang, TLS jenseits der Schleife, die Verbindungsgrenze
und die Fristen. Dienst und Werkzeugserver lesen den Schalter beim Start, und
jede Auskunft über die Grenzen sagt, was gilt: `"kontingente": "aus"`.

**Was ein 429 sagt**: `Retry-After` ist die echte Restzeit in ganzen Sekunden,
nie null. **Was begrenzt bleibt**: die Tabellen der Principals (256) und der
Nachweise (1.024) verdrängen den am längsten nicht Gesehenen — wer Zufall
vorzeigt, füllt keinen Speicher. **Was es nicht gibt**: kein CAPTCHA, kein
Proof-of-Work — und ins Netz geht der Dienst nur über die Lauschwahl, mit TLS und Marke ([0115](entscheidungen/0115-die-schleife-ist-eine-vorgabe.md)).

## Passt die neue Fassung zu diesem Bestand?

**Seit Werkstück O6 der Iteration 0.5.0**
(#1430)
sagt jede Freigabe, **was sie öffnen kann** — und die Anwendung prüft es,
bevor irgendetwas getauscht wird.

**Verglichen wird nicht eine Schemanummer.** Zwei Fassungen können dieselbe
nennen und **andere** Wanderungen verbucht haben; eine nachträglich geänderte
Migration trägt dieselbe Nummer und einen anderen Inhalt. Geprüft werden
deshalb **Name und Prüfsumme jeder Wanderung** gegen den echten Ledger dieses
Bestands.

**Und seit 0.8.0 die Grundlinie** ([0189](entscheidungen/0189-nur-die-grundlinie.md)):
das Manifest nennt, bis zu welchem Schritt eine Fassung den Bestand nur noch
als Ganzes kennt (`"grundlinie"`). Fehlt dem Bestand ein Schritt bis dorthin,
öffnet sie ihn nicht, auch wenn sie jeden seiner Schritte kennt — das
Hindernis heißt `vorDerGrundlinie`. **Eine Fassung vor 0.8.0 liest das Feld
nicht**: sie hält die Freigabe 0.8.0 für einen Bestand aus 0.6.0 für
verträglich, und erst der Start von 0.8.0 weist ihn ab, ohne ihn anzufassen.

**Drei Ausgänge, und der dritte ist kein Nein:**

| Urteil | Heißt |
| --- | --- |
| `vertraeglich` | die Fassung kann diesen Bestand öffnen |
| `unvertraeglich` | sie kann es nicht — die Hindernisse sagen, woran es liegt |
| `ungeklaert` | die Freigabe sagt nichts über sich; sie ist älter als diese Prüfung |

```bash
mse vertraeglichkeit          # 0 ja, 1 nein, 2 ungeklärt
```

```bash
curl -s -H "Authorization: Bearer $MSE_MARKE" http://127.0.0.1:7777/vertraeglichkeit
```

Über MCP heißt das Werkzeug `lesen.vertraeglichkeit`; im Fenster steht das
Urteil **im Band**, das die neue Fassung meldet — vor dem Knopf, der sie lädt.

**Das Manifest hängt an der signierten Prüfsummenliste.** `manifest.json`
liegt der Freigabe bei, seine SHA-256 steht in `pruefsummen.txt`, und die ist
mit OpenPGP signiert — unverändert seit 0.4.0. Ein manipuliertes Manifest
fällt damit auf, ohne einen zweiten Schlüssel und ein zweites Format.

**Mit Apple-Notarisierung hat das nichts zu tun**
([0003](entscheidungen/0003-keine-notarisierung.md)): jene entscheidet, ob
Gatekeeper das Bündel startet, diese, ob die Fassung diesen Bestand öffnen
kann.

**Und „vorbereitet" heißt gesichert und zurückgelesen.** Der Knopf *Wechsel
vorbereiten* im Fenster prüft zusätzlich Platz und laufende Arbeit, schreibt
eine Sicherung und **liest sie zurück**; erst dann steht „bereit" da. Eine
Sicherung, die niemand gelesen hat, ist eine Behauptung. Das tut nur das
Fenster: es hält den Bestand dafür an, und eine Kommandozeile, die das auf
Zuruf täte, nähme ihn jemandem weg, der gerade arbeitet.

---

## Aktualisierung

**Die Vorgabe ist `aus`, und dann geht nichts hinaus.** Unter _Einstellungen →
Nach neuen Fassungen sehen_ stehen drei Stufen
(#1100):

| Stufe             | Was geschieht                                                              |
| ----------------- | -------------------------------------------------------------------------- |
| **aus** (Vorgabe) | nichts. Kein Netzverkehr, keine Frage                                      |
| **melden**        | beim Start ein Band, wenn eine neuere Fassung vorliegt — einmal je Fassung |
| **herunterladen** | dazu ein Knopf, der die Fassung ins Standard-Downloadverzeichnis holt      |

**Installiert wird nichts von selbst** ([0050](entscheidungen/0050-die-anwendung-ersetzt-sich-nicht-selbst.md)):
die geladene Datei wird im Finder geöffnet und die Anwendung in den
Programme-Ordner gezogen. Die Daten bleiben, weil sie außerhalb des Bündels
liegen; ein Wächter hält das fest (`internal/app/orteueberleben_test.go`).

### Was geprüft wird, bevor eine Datei liegt

**Zwei Dinge, in dieser Reihenfolge:**

1. Die **Signatur** über `pruefsummen.txt` — mit dem Schlüssel des
   Herausgebers, der im Programm liegt (`97721629…44C59428DE1F766E`, seit
   1.0.0). Sie sagt,
   **von wem** die Liste kommt.
2. Die **Prüfsumme** des Abbilds gegen die Zeile in dieser Liste. Sie sagt, ob
   das Geladene das Gemeinte ist.

Ohne Signatur wird gar nicht geladen. Passt eine der beiden nicht, bleibt
**nichts** liegen: geschrieben wird in eine Nachbardatei, und erst die geprüfte
bekommt ihren Namen.

### Die Zugangsmarke

Das Repositorium ist **heute privat**. Ohne Lesezugriff antwortet GitHub, es
gebe keine Freigabe — die Anwendung sagt das und arbeitet unverändert weiter.
**Es wird öffentlich**, bevor eine Fassung an Anwendende geht
([0190](entscheidungen/0190-das-repositorium-wird-oeffentlich.md)); danach
braucht die Prüfung keine Marke mehr, und die Marke bleibt der Weg für den,
der eine Vorabfassung aus einem privaten Stand testet.

Die Marke kommt aus der **Umgebung**: `GH_TOKEN`, sonst `GITHUB_TOKEN` —
dieselben zwei Namen und dieselbe Vorrangfolge wie bei `gh`. Nötig ist die
Berechtigung **`Contents` in der Stufe `read`**.

**Nicht aus den Einstellungen**, und das ist Absicht: die liegen im Bestand,
und der ist eine SQLite-Datei im Klartext — eine Marke dort stünde in jeder
Sicherung, auch in denen, die jemand weitergibt.

**Wer die Anwendung aus dem Finder startet**, sieht die Umgebung der
Anmeldesitzung und nicht die einer Shell: eine Marke, die nur in `~/.zshrc`
steht, ist im Fenster nicht da. Auf der Kommandozeile ist sie es.

### Dieselbe Auskunft ohne Fenster

```bash
mse aktualisierung            # beide Fassungen nennen
mse aktualisierung --laden    # holen, mit Fortschritt auf dem Fehlerstrom
```

Im Dienst sind es `GET /aktualisierung` und `POST /aktualisierung/laden`.
Beide fragen **ohne** Rücksicht auf die Einstellung — sie verhindert
ungefragtes Verhalten, und ein Aufruf ist die Frage. Über MCP gibt es die
Fähigkeit bewusst nicht; der Grund steht in
[`FAEHIGKEITEN.md`](FAEHIGKEITEN.md).

---

## Außerbetriebnahme

**Neu am 2026-09-07**, angestrichen bei der Sichtprüfung über die ganze
Anwendung: hier stand, **wo** alles liegt, und nirgends, wie man es wieder
loswird. Eine Anwendung, die sich nicht sauber entfernen lässt, ist nicht
fertig beschrieben.

### Beenden

`⌘Q` oder *MagicSetEditor4Mac → Beenden*. Was dabei geschieht, steht oben unter
[„Der Abbruch"](#der-abbruch) — der Bestand wird geschlossen, das Protokoll
geschrieben, die Sperrdatei freigegeben.

### Zuerst zurücknehmen, was die Einrichtung anlegte

**Neu mit OPS03** (#2349,
Befund A-10 des Audits 0.9.0): hier standen nur Programm, Anwendungsordner und
Schlüsselbund. Was *Werkzeuge ohne Fenster* außerhalb davon anlegt, blieb
liegen — ein LaunchAgent, der bei jeder Anmeldung ein fehlendes Programm
starten will, Verweise ins Leere und Einträge bei Codex und Claude, die auf
ein fehlendes `mse-mcp` zeigen.

**Das geht vor dem Löschen des Programms**, denn danach gibt es keinen der
beiden Wege mehr:

- am Fenster: *Einstellungen › Verbindungen › Werkzeuge ohne Fenster* →
  **Alles widerrufen**,
- an der Kommandozeile: `mse schnittstellen widerrufen` — oder ein Ziel allein,
  `mse schnittstellen widerrufen codex`.

**Genommen wird nur, was diese Anwendung selbst eintrug.** Fremde Einträge in
denselben Dateien bleiben bytegleich, ein fremder Eintrag unter demselben
Namen bleibt ganz und wird genannt. **Ein zweiter Widerruf ändert nichts** und
ist kein Fehler. **Und es geht auch, wenn das Programm schon verschoben oder
erneuert ist**: ein Eintrag, der in ein Bündel dieser Anwendung zeigt, das es
nicht mehr gibt, gilt als eigener und wird genommen.

| Ziel | Ort | Was genommen wird |
| --- | --- | --- |
| `befehle` | `~/.local/bin/mse`, `~/.local/bin/mse-mcp` | die zwei Verknüpfungen, wenn sie in ein Bündel `MagicSetEditor4Mac*.app` zeigen, dessen `Info.plist` die Kennung dieser Anwendung trägt — oder das es nicht mehr gibt; eine fremde Datei unter demselben Namen und eine Verknüpfung in ein fremdes Bündel bleiben |
| `befehlszeile` | `~/.zprofile` | der markierte Block, der `~/.local/bin` in den `PATH` nimmt; eine Datei, die dabei leer wird, geht |
| `codex` | `~/.codex/config.toml` | der markierte Block `[mcp_servers.magicseteditor4mac]` |
| `claudedesktop` | `~/Library/Application Support/Claude/claude_desktop_config.json` | der Eintrag `magicseteditor4mac` unter `mcpServers` — bytegleich zurück, wenn die Datei seither unverändert ist; das Gedächtnis dafür liegt unter `~/Library/Application Support/MagicSetEditor4Mac/schnittstellen/claudedesktop.json` |
| `claudecode` | `~/.claude.json` | über `claude mcp remove -s user magicseteditor4mac` — die Datei gehört `claude`, und diese Anwendung schreibt sie nicht selbst. **Ohne `claude` im `PATH` der Anmeldeshell bleibt der Eintrag** und wird genannt; ohne `claude` liest ihn niemand |
| `dienst` | `~/Library/LaunchAgents/io.github.davidtheider.magicseteditor4mac.dienst.plist` — **und, wenn er noch dasteht, der Dienst von vor 0.9.0** unter `~/Library/LaunchAgents/com.wails.MagicSetEditor4Mac.dienst.plist` | `launchctl bootout`, dann die plist und die Ansage — beim alten Etikett nur, wenn die plist `mse-rest` nennt; seine Ansage liegt unter `~/Library/Application Support/MagicSetEditor4Mac/schnittstellen/dienst.txt`, seine Fehlerausgabe daneben in `dienst.log` |

**Ist das Programm schon fort**, nimmt man dieselben Einträge von Hand — nur
die eigenen, also nur, was in ein Bündel dieser Anwendung zeigt:

```bash
ls -l ~/.local/bin/mse ~/.local/bin/mse-mcp
launchctl bootout gui/$(id -u)/io.github.davidtheider.magicseteditor4mac.dienst
```

### Dann die Daten — eine eigene Entscheidung

**Was hier steht, löscht Daten.** Kein Knopf dieser Anwendung tut das für
dich, auch „Alles widerrufen" nicht: es nimmt Einträge, keine Daten.

| Was | Wo | Bleibt sonst |
| --- | --- | --- |
| **Das Programm** | `/Programme/MagicSetEditor4Mac.app` | in den Papierkorb |
| **Alles, was die Anwendung führt** | `~/Library/Application Support/MagicSetEditor4Mac/` | Bestand, Pakete, Beiwerk, Protokoll, Profile, das Gedächtnis der Schnittstellen |
| **Was sich wegwerfen lässt** | `~/Library/Caches/MagicSetEditor4Mac/`, `~/Library/Caches/io.github.davidtheider.magicseteditor4mac/` | Antworten von Scryfall, Bildvorschläge; die Sperre des Umzugs der Vorgaben und, nach einem Abbruch, sein Arbeitsordner `umzug-…` |
| **Die Vorgaben** | `~/Library/Preferences/io.github.davidtheider.magicseteditor4mac.plist` | Fensterlage und Einstellungen von AppKit — über `defaults delete io.github.davidtheider.magicseteditor4mac`, nicht mit `rm`: `cfprefsd` hält sie sonst weiter |
| **Der Speicher der Webansicht** | `~/Library/WebKit/io.github.davidtheider.magicseteditor4mac/`, `~/Library/HTTPStorages/io.github.davidtheider.magicseteditor4mac/` | was die Oberfläche im Browser ablegt |
| **Was unter der alten Kennung liegt** | `~/Library/Preferences/com.wails.MagicSetEditor4Mac.plist`, `~/Library/WebKit/com.wails.MagicSetEditor4Mac/`, `~/Library/HTTPStorages/com.wails.MagicSetEditor4Mac/` — auch in Kleinschreibung | der Umzug auf die neue Kennung (R02, [0191](entscheidungen/0191-die-kennung-ist-io-github-davidtheider.md)) kopiert und lässt das Alte stehen |
| **Was der Umzug beiseitelegte** | `~/Library/WebKit/io.github.davidtheider.magicseteditor4mac.vor-umzug-…/`, `~/Library/HTTPStorages/io.github.davidtheider.magicseteditor4mac.vor-umzug-…/` | ein Speicher der Webansicht, den die neue Fassung vor dem Umzug schon angelegt hatte; der Umzug legte ihn beiseite, statt ihn zu überschreiben |
| **Was ein abgebrochener Umzug liegen ließ** | `~/Library/WebKit/io.github.davidtheider.magicseteditor4mac.umzug/`, `~/Library/HTTPStorages/io.github.davidtheider.magicseteditor4mac.umzug/` | der Zwischenort, in den der Umzug kopiert, bevor er umbenennt; brach er ab, bleibt er stehen, und der nächste Start fängt neu an |
| **Vier Einträge im Schlüsselbund** | siehe unten | **auch nach dem Löschen des Ordners** |

```bash
rm -rf ~/Library/Application\ Support/MagicSetEditor4Mac ~/Library/Caches/MagicSetEditor4Mac \
  ~/Library/Caches/io.github.davidtheider.magicseteditor4mac
defaults delete io.github.davidtheider.magicseteditor4mac
```

**Die eigenen Sets bleiben.** Sie liegen dort, wo sie angelegt wurden — in der
Regel `~/Documents`, und ausdrücklich **nicht** im Anwendungsordner
([0029](entscheidungen/0029-bilder-liegen-neben-dem-set.md)). Sie zu löschen
ist eine eigene Entscheidung.

**Ein Wächter hält beide Tafeln gegen den Quelltext**
(`internal/app/ausserbetriebnahmeorte_intern_test.go`): jedes Ziel der
Schnittstellen mit Datei, Gedächtnis und Begleitordner, die Webordner und die
Vorgaben unter beiden Kennungen, der Zwischenspeicher — ein Ort im Code ohne
Zeile hier macht die Pipeline rot. **Was er nicht hält**: dass die Befehle so
wirken, wie sie dastehen; das sieht nur ein Mensch.

### Der Schlüsselbund

**Er überlebt das Löschen des Ordners**, und das ist kein Versehen: dorthin
gehört, was man nicht weitergibt (§16). Vier Dienste tragen Einträge:

| Dienst | Was darin liegt |
| --- | --- |
| `MagicSetEditor4Mac` | Zugangsdaten fremder Kartendienste |
| `MagicSetEditor4Mac-Anmeldung` | die `client_secret` der Anmeldeanbieter |
| `MagicSetEditor4Mac-Faktor` | die TOTP-Geheimnisse der zweiten Faktoren |
| `MagicSetEditor4Mac-Bestand` | der Schlüssel des verschlüsselten Bestands |

```bash
for dienst in MagicSetEditor4Mac MagicSetEditor4Mac-Anmeldung \
              MagicSetEditor4Mac-Faktor MagicSetEditor4Mac-Bestand; do
  while security delete-generic-password -s "$dienst" >/dev/null 2>&1; do :; done
done
```

**Die Schleife ist Absicht**: je Dienst liegen mehrere Einträge — einer je
Anbieter, je Nutzer, je Bestand —, und `security` löscht einen davon je Aufruf.

> **Den Bestandsschlüssel zuletzt.** Wer ihn vor der Datenbank löscht, hat eine
> verschlüsselte Datei ohne Schlüssel — sie ist dann nicht mehr zu öffnen, auch
> nicht zum Sichern. **Diese Reihenfolge zählt nur, wenn die Verschlüsselung
> eingeschaltet war.**

**Ein Wächter hält diese Liste gegen den Quelltext**
(`internal/werkzeuge/schluesselbunddienste_test.go`): ein fünfter Dienst im
Code ohne Zeile hier macht die Pipeline rot. Was er **nicht** hält, ist der
Löschbefehl daneben — dass `security` so heißt und so wirkt, sieht nur ein
Mensch.

### Was danach noch daliegen kann

**Eine heruntergeladene Fassung.** Steht die Aktualisierung auf *herunterladen*,
liegt das `.dmg` im Downloadverzeichnis und wird nicht mit weggeräumt.

**Ein eingehängtes Disk-Image.** Wer gerade aktualisiert hat, wirft es über den
Finder aus.

---

## Gatekeeper

Die Anwendung ist **ad hoc signiert und nicht notarisiert**. Beim ersten Start
weist macOS sie ab; der Weg steht in
[`README.md`](../README.md#erster-start-nach-dem-download).

Das ist kein Mangel, den man verschweigt, sondern eine Eigenschaft, die man
erklärt — es gibt keinen Apple-Developer-Account, und es soll keinen geben.

## `mse-mcp` mit einem Token starten

**Seit Werkstück O2 von Phase 5.2** ([0080](entscheidungen/0080-maschinenzugaenge-und-widerrufbare-sitzungen.md))
kann sich der MCP-Prozess ausweisen. Ohne Token läuft er wie bisher: über
stdio, auf demselben Rechner, im gezeigten Profil, ohne Wache — das ist der
**lokale Altbetrieb**, und er bleibt.

Mit Token:

1. Ein Token anlegen — am Fenster unter *Nutzer und Profile*, an der Zeile mit
   `mse token-anlegen <profilID>` (JSON mit `bezeichnung`, `principal`,
   `scope`, `ablauf` auf der Standardeingabe). Der Klartext erscheint genau
   einmal.
2. Den Klartext in eine Datei schreiben, die nur der eigene Nutzer lesen darf:

   ```bash
   (umask 077 && printf '%s\n' 'mse_…' > ~/.mse/mcp-token)
   ```

   Eine Datei mit weiteren Rechten weist `mse-mcp` **ab** — ein Geheimnis,
   das schon für jeden lesbar war, ist keines mehr.
3. `mse-mcp` mit `MSE_TOKEN_DATEI` starten, etwa in der Konfiguration eines
   Clients:

   ```json
   { "mcpServers": { "mse": {
     "command": "/Applications/MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp",
     "env": { "MSE_TOKEN_DATEI": "/Users/ich/.mse/mcp-token" } } } }
   ```

   `MSE_TOKEN` mit dem Klartext geht auch; ein **Argument** gibt es nicht,
   denn Argumente stehen in der Prozessliste.

**Was dann gilt**: jeder Werkzeugruf prüft das Token — Widerruf, Ablauf und
Umfang (`lesen` trägt keine `schreiben.*`-Werkzeuge) wirken beim nächsten
Ruf, mitten in der Sitzung. Das Profil des Tokens ist das Profil der Sitzung.
Eine Rotation (`mse token-rotieren <id>`) gibt ein neues Geheimnis; das alte
gilt noch eine Stunde, damit der Client in Ruhe umgestellt werden kann.
