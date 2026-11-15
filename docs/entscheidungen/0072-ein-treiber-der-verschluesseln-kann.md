# 0072 — Ein Treiber, der verschlüsseln kann

**Entschieden am 2026-09-06**, Werkstück 14 von Phase 4.12
(#1351) ·
**Löst ab** [0042](0042-keine-verschluesselung-des-bestands.md) ·
**Ergänzt** [0002](0002-sqlite-als-arbeitsformat.md)

## Der Anlass

**Vor zwei Wochen ist das Gegenteil entschieden worden**, und 0042 verlangt von
jedem, der das umdreht, dass er zuerst ihn ändert. Das ist geschehen; der
Kasten oben in 0042 sagt, welche seiner beiden Hälften warum hinfällig ist.

Kurz: er wog **Kartensets** gegen **cgo**. Heute liegen im Bestand
Personendaten, und der Treiber, der verschlüsseln kann, braucht kein cgo —
**gemessen am 2026-09-06**: `GOOS=windows` und `GOOS=linux go build ./...`
glücken beide nach dem Tausch. Die Prüfstufe aus §12, deren Wegfall der Preis
gewesen wäre, steht unverändert.

Der Bestand hält seit Phase 4.12 Personendaten: Namen, Geburtsdatum, Anschrift,
Fotos, Anbieterkonten, TOTP-Geheimnisse. `docs/HILFE` sagt seit Werkstück 13,
**wo** das liegt — in einer Datei im Zuhause des Rechners, im Klartext.

Der Satz „im Klartext" ist die Entscheidung, die hier fällt.

## Portierungsmarke: **new** — und hier ist gesucht worden

§15 verlangt, dass eine `new`-Marke sagt, **wo** gesucht wurde. Am 2026-09-06
gemessen:

| Quelle | Treffer | Was es ist |
| --- | ---: | --- |
| `/Users/david/mse2-quelle/src` auf `encrypt\|crypt\|cipher` | **0** | — |
| `/Users/david/magic-set-generator/src` auf dasselbe | **0** | — |

**MSE2 speichert nichts über Menschen**, also verschlüsselt es auch nichts. Es
gibt kein Altverhalten, das zu portieren wäre.

## 1. Der Treiber wird getauscht, nicht ergänzt

`modernc.org/sqlite` **kann nicht verschlüsseln**, und es ist nicht so gebaut,
dass man es ihm beibringt: die Chiffre sitzt unter dem Dateizugriff, und den
führt es selbst.

`ncruces/go-sqlite3` legt genau dort eine Naht offen — ein VFS lässt sich
umhüllen —, und bringt zwei fertige Hüllen mit (`adiantum`, `xts`).

**Beide Treiber sind cgo-frei**, und das ist die Bedingung aus
[0002](0002-sqlite-als-arbeitsformat.md), die unverändert gilt: mit cgo bräuchte jede
Zielplattform eine eigene Werkzeugkette, und die Kreuzübersetzung nach Windows
und Linux in der Stufe `types` gäbe es nicht. `modernc` übersetzt SQLites C
nach Go; `ncruces` führt es als WebAssembly über `wazero` aus.

**Zwei Treiber nebeneinander wären die schlechtere Wahl gewesen** — dann gäbe
es zwei Verbindungsaufbauten, zwei Fehlerzuordnungen und die Frage, welcher
gerade gilt, an jeder Naht. Getauscht ist einmal teuer und danach eine Sache.

### Was der Tausch gekostet hat, ist gemessen

| | vorher | nachher |
| --- | ---: | ---: |
| Module in `go.sum` | 138 Zeilen | **100** |
| Direkte und mittelbare Module | — | **21 weg, 3 dazu** |
| Laufzeit `internal/storage` | 18,2 s | **18,8 s** |
| Bau (`go build ./...`) | glückt | glückt |

## 2. Vier Unterschiede, die kein Aufruf zeigt

Sie stehen hier, weil sie **still** ausfallen — die Pipeline war bei dreien von
ihnen grün, bis ein Lauf sie an einer ganz anderen Stelle traf.

### FTS5 ist eine Erweiterung, keine Voreinstellung

`ncruces` baut SQLite ohne FTS5; angemeldet wird es **je Verbindung**. Ohne
diesen Haken endet die dritte Wanderung mit „no such module: fts5".

**Gemessen**: ohne ihn fielen **509** Läufe der Speicherschicht, mit ihm elf.

### Fremdschlüssel sind voreingestellt **an**

SQLite selbst hat sie aus, `modernc` ließ sie aus, `ncruces` schaltet sie an.
Gemessen: dieselbe Adresse ohne jedes Pragma gibt dort `PRAGMA foreign_keys = 1`.

Am Produktivweg ändert das nichts — `Oeffnen` setzt das Pragma seit
#152 ohnehin.
An **drei Tests** ändert es alles: sie bauen den Bestand von vor dem 2026-08-15
nach, in dem Waisen entstehen konnten. Ohne ein ausdrückliches
`foreign_keys(0)` lässt sich der Schaden nicht mehr herstellen, den die
Migration wegräumt — die Läufe wären grün, weil sie nichts mehr prüfen.

### Derselbe Fehler kommt in drei Gestalten

`modernc` gab immer einen `*sqlite.Error`. `ncruces` gibt

    INSERT mit NULL            → *sqlite3.Error
    BeginTx auf Nichtdatenbank → sqlite3.ExtendedErrorCode, blank

Ein `errors.As` auf den **Zeiger** greift an der zweiten daneben. Was das
kostete, waren zwei Auskünfte, an denen ein Mensch etwas tut (§8), und sie
sagen Gegenteiliges: „die Datei ist keine Datenbank" kam als „von einem anderen
Programm geöffnet" an — „ersetze die Datei" statt „schließe das andere
Fenster" —, und „belegt, gleich nochmal" kam gar nicht mehr an.

**Gefragt wird jetzt nach dem Code und nicht nach der Hülle.** Der Treiber gibt
jeder Gestalt ein `As` auf `*ExtendedErrorCode`; ein einziges `errors.As` deckt
damit alle ab, ohne dass ein Typschalter dastünde, der bei der nächsten Gestalt
wieder danebengriffe.

### `…/embed` ist abgekündigt und druckt beim Start

Das Paket sagt in seinem `init()` eine Zeile auf die Standardausgabe. Sie hat
**fünfzehn** Flächenläufe rot gemacht: der REST-Dienst sagt seine Adresse auf
der ersten Zeile an, und die war dann das Banner.

**Der Import stammte aus dem Gedächtnis** und nicht aus der Doku — §1 in
Reinform. Die Paketdoku sagt in ihrer ersten Zeile: „Package embed is no longer
needed."

## 3. Was die Verschlüsselung zusagt — und was nicht

Gewählt ist **Adiantum**: XChaCha12 als Stromchiffre, AES als Blockchiffre, NH
und Poly1305 zum Hashen, Argon2id für die Ableitung aus Text. Verschlüsselt
werden 4-KiB-Blöcke, die Seitengröße von SQLite.

**Nachgelesen am 2026-09-06** in der Paketdoku, nicht erschlossen (§16, Regel 1):

| Zusage | Gilt |
| --- | --- |
| Der Bestand liegt nicht mehr im Klartext auf der Platte | **ja** — gemessen, siehe unten |
| Ohne Schlüssel ist er nicht zu öffnen | **ja** — gemessen |
| Schutz gegen Fälschung oder Manipulation | **nein**, ausdrücklich nicht |
| Ein Angreifer mit **mehreren** Abzügen lernt nichts | **nein** — die Chiffre ist deterministisch: er sieht, welche Blöcke sich geändert haben und welche zurückgenommen wurden |

**Der letzte Punkt gehört an die Oberfläche und nicht nur hierher.** Wer
Sicherungen aufhebt, hebt Momentaufnahmen auf, und genau die sind das
Angriffsbild. Die Hilfe sagt es deshalb im selben Absatz wie das Versprechen.

**Verworfen wurde `xts`** — es nimmt nur von NIST und FIPS 140 zugelassene
Bausteine, und das ist ein Vorzug, wenn jemand danach fragt. Hier fragt niemand
danach: der Bestand liegt auf **einem** Rechner, es gibt keine Zulassung
einzuhalten, und Adiantum ist auf Geräten ohne AES-Beschleunigung schneller.
Umstellen ließe sich später, denn beide Hüllen nehmen dieselbe Naht.

### Gemessen, nicht angenommen

Ein Probelauf am 2026-09-06 hat vier Dinge festgestellt:

    VACUUM INTO 'file:…?vfs=adiantum&hexkey=…'  → glückt
    das Klartextwort auf der Platte              → nicht auffindbar
    Wiederöffnen mit vfs + PRAGMA hexkey         → liest es zurück
    falscher oder fehlender Schlüssel            → SQLITE_NOTADB

**Die vierte Zeile ist ein Befund und keine Beobachtung.** `SQLITE_NOTADB`
heißt in diesem Programm seit
#886 „die
Datei ist kein Bestand" — und die Auskunft, die ein Mensch daraufhin bekommt,
lautet sinngemäß *ersetze sie*. Bei einem verschlüsselten Bestand ist das die
**falsche** Auskunft und im schlimmsten Fall ein Datenverlust: die Datei ist in
Ordnung, der Schlüssel fehlt.

## 4. Der Schlüssel liegt im Schlüsselbund, nicht in der Adresse

Die Paketdoku bietet drei Wege an: `key`, `hexkey` und `textkey` als
URI-Parameter — oder dasselbe als `PRAGMA` unmittelbar nach dem Öffnen. Sie
sagt selbst, warum das zweite besser ist: ein URI-Parameter ist über
`vfs.Filename.URIParameters` für jeden Teil der Anwendung lesbar.

**Genommen wird das `PRAGMA`**, und es passt in den Haken, den es seit dem
Treiberwechsel ohnehin gibt — denselben, der FTS5 anmeldet.

Der Schlüssel selbst sind **32 zufällige Bytes** im Schlüsselbund von macOS,
über `internal/schluesselbund` — dieselbe Ablage wie für die Zugangsdaten
fremder Dienste ([0056](0056-drei-weitere-kartendienste.md)), und aus demselben
Grund: er gehört nicht in den Bestand, denn den gibt man weiter.

**Kein Kennwort, das ein Mensch sich ausdenkt.** Ein `textkey` liefe über
Argon2id und wäre bequemer zu merken — und genau deshalb schwächer als 32
zufällige Bytes. Wer den Rechner entsperrt hat, kommt an den Schlüsselbund;
das ist dieselbe Schwelle, die auch die Anwendung schützt.

## 5. Dieselbe Falle zweimal: `VACUUM INTO` erbt das VFS

**Gefunden am 2026-09-06, und beim zweiten Mal gesucht statt gefunden** (§10:
nicht den Fall behandeln, sondern die Klasse).

SQLite öffnet das Ziel eines `VACUUM INTO` über das VFS der **Hauptdatenbank**,
solange die Adresse kein eigenes nennt. Bei einem verschlüsselten Bestand ist
das die Chiffre — und das trifft **jeden** Weg, der aus der offenen Datenbank
heraus eine zweite Datei schreibt. Davon gibt es genau drei:

| Weg | Was ohne `vfs=` geschah |
| --- | --- |
| `Verschluesseln` | richtig — hier **soll** die Chiffre gelten |
| `Entschluesseln` | „unable to open database file". Der Flächenlauf hat es gefunden: **einschalten glückte, ausschalten nicht** |
| `Sichern` | dasselbe. Ein verschlüsselter Bestand ließ sich **gar nicht sichern** |

**Der dritte war der gefährlichere**, und zwar nicht wegen des Fehlschlags:
wäre er geglückt, hätte er eine **verschlüsselte Sicherung** geschrieben. Die
geht nur auf dem Rechner auf, auf dem sie entstand — und damit wäre sie genau
das nicht mehr, wofür es sie gibt.

**Beide nennen jetzt `vfs=os` ausdrücklich**, und beide haben einen Lauf in der
Speicherschicht: einer prüft, dass das Entschlüsselte die Klartextmarke trägt,
der andere, dass die Sicherung sie trägt **und** sich ohne Schlüssel öffnen
lässt. Gegenprobe für beide gesehen.

**Was daraus folgt, wenn eine Sicherung zurückgespielt wird**: sie ist
Klartext, also ist der Bestand danach Klartext — auch wenn er vorher
verschlüsselt war. Das ist keine Überraschung, sondern die Lage: der Dateikopf
sagt, was gilt, und die Oberfläche liest ihn. Der alte Schlüssel bleibt im
Schlüsselbund liegen und stört dort niemanden; wer wieder verschlüsseln will,
schaltet erneut ein und bekommt einen neuen.
