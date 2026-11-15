# Sicherung und Wiederherstellung

**Was eine Sicherung trägt, wie sie geschützt ist, wie sie zurückkommt und wer
das darf** — OPS-06
(#1723).
Die Bedienung steht in [`ADMINISTRATION.md`](ADMINISTRATION.md) unter
„Sicherung" und „Zurückspielen"; hier steht, **was gilt und warum**.

Ein Wächter hält die Matrix gegen den Quelltext
(`internal/werkzeuge/sicherungsumfang_test.go`): jeder Posten der
Platzübersicht und jeder Teil des Archivs steht darin. **Was er nicht hält**:
ob die Einordnung stimmt — das ist eine Entscheidung und steht mit Grund da.

---

## Der Umfang (AK-01)

| Was | Wo es liegt | In der Sicherung | Warum |
| --- | --- | --- | --- |
| **Der Bestand** — Sets, Karten, Sammlung, Decks, Turniere, Partien, Handel | `bestand.db` | **ja**, als `bestand.db` (über `VACUUM INTO`, ohne `-wal` und `-shm`) | das ist, was sich nirgends sonst wiederbeschaffen lässt |
| **Nutzer, Profile, Rollen, Freigaben** | im Bestand | **ja** | sie stehen in derselben Datei; **alle** Profile reisen mit, weil ein Zurückspielen den ganzen Bestand tauscht |
| **Einstellungen aller vier Ebenen** | im Bestand | **ja** — auch die des Geräts | eine Sicherung gibt den Stand **dieses** Rechners wieder. Auf einem anderen prüft der Pfadvertrag jeden eingestellten Ort beim Gebrauch; der Konfigurationsexport daneben lässt das Gerät dagegen weg (`EINSTELLUNGEN.md`) |
| **Metadaten der Sicherung** | — | **ja**, als `sicherung.json`: Fassung, Zeitpunkt, je Teil Größe und SHA-256, die Pakete | ohne sie fiele eine beschädigte Sicherung erst auf, wenn jemand sie braucht |
| **Eingesetzte Bilder** | `beiwerk/` | **ja**, als `beiwerk/…` | ein Set ist ohne sie unvollständig ([0029](entscheidungen/0029-bilder-liegen-neben-dem-set.md)) |
| **Eigene Pakete** — angelegt, von Hand hineingelegt, aus der Saat | `pakete/`, **ohne** Prüfsumme | **ja**, als `pakete/<name>/…` | es gibt sie nirgends sonst |
| **Bezogene Pakete** | `pakete/`, **mit** Prüfsumme | **nein** — nur als Verweis im Manifest: Name, Herkunft, Prüfsumme | wiederbeziehbar, und zusammen über ein Gigabyte. Die Vorschau nennt vor dem Zurückspielen, welche fehlen |
| **Pakete aus einer MSE2-Installation** | außerhalb von `pakete/` | **nein** | sie gehören nicht zum Bestand; mitgenommen trügen sie fremde Dateien auf einen anderen Rechner |
| **Sicherungen vor Migrationen** | `bestand-vor-*.db` | **nein** | sie sind der Rückweg **dieses** Rechners zu einer älteren Fassung, keine Nutzerdaten |
| **Das Protokoll** | `mse.log*` | **nein** | es beschreibt keinen Stand |
| **Zwischenspeicher** | `~/Library/Caches/MagicSetEditor4Mac` | **nein** | füllt sich beim nächsten Gebrauch ([`ZWISCHENSPEICHER.md`](ZWISCHENSPEICHER.md)) |
| **Leistungsprofile** | `Leistungsprofile/` | **nein** | Messungen dieses Prozesses |
| **Geheimnisse** — Zugänge fremder Dienste, der Schlüssel des verschlüsselten Bestands, Anbietergeheimnisse | Schlüsselbund | **nein, ausdrücklich** | den Bestand gibt man weiter, den Schlüsselbund nicht ([0056](entscheidungen/0056-drei-weitere-kartendienste.md)); auf einem anderen Rechner werden Zugänge neu eingetragen |
| **Was sonst im Anwendungsordner liegt** | Posten „sonstiges" | **nein** | was die Anwendung nicht versteht, packt sie nicht ein |

---

## Das Kennwort (AK-02)

**Eine Sicherung kann ein Kennwort tragen.** Dann ist sie mit
[age](https://age-encryption.org) verschlüsselt: aus dem Kennwort wird mit
scrypt (Arbeitsfaktor 2^18, rund eine Sekunde) der Schlüssel, der Inhalt ist
mit ChaCha20-Poly1305 **authentifiziert** — eine veränderte Datei geht nicht
auf, statt falsch.

- **Portabel.** Das Kennwort liegt **nirgends**, auch nicht im Schlüsselbund:
  die Sicherung geht auf jedem Rechner auf, und mit `age -d` sogar ohne diese
  Anwendung.
- **Kein Rückweg.** Wer das Kennwort vergisst, hat die Sicherung verloren.
  Das sagt die Oberfläche **vor** dem Sichern, und deshalb wird es zweimal
  eingegeben.
- **Mindestens zwölf Zeichen**, ohne Vorschriften über Zeichenklassen —
  Länge hilft gegen einen Angriff auf eine Datei, der niemand zählt, und ein
  Satz aus vier Wörtern ist leichter zu behalten (`domain.PruefeSicherungskennwort`).
- **Kein Klartext auf dem Weg.** Das Archiv entsteht im Strom durch die
  Verschlüsselung; der Zwischenstand der Datenbank liegt neben dem Bestand und
  **nicht** am Ziel — ein USB-Stick behielte eine gelöschte Datei lesbar.

**Ein verschlüsselter Bestand gibt keinen Klartext her, ohne dass es jemand
will.** Ist der Bestand verschlüsselt ([0072](entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md)),
verlangt eine Sicherung ohne Kennwort das ausdrückliche „ohne Kennwort sichern"
— am Fenster ein Haken, an der Kommandozeile `--klartext`, am Dienst
`"klartext": true`, über MCP `klartext`. Bis OPS-06 entstand dort still ein
Archiv im Klartext.

| Fläche | Kennwort beim Sichern | Kennwort beim Lesen |
| --- | --- | --- |
| Fenster | zwei Felder vor dem Speichern-Dialog | ein Feld, sobald eine verschlüsselte Sicherung gewählt ist |
| Kommandozeile | `MSE_SICHERUNGSKENNWORT` | dieselbe Umgebungsvariable |
| Dienst | `"kennwort"` im Rumpf von `POST /bestand/sicherung` | Kopf `X-Sicherungskennwort` — **nie** in der Adresse |
| Werkzeugserver | **keines** | **keines** |

**Nicht als Argument und nicht in der Adresse**: beides steht in
Prozesslisten, Verläufen und Protokollen. **Und nicht über MCP**: ein
Kennwort, das durch ein Sprachmodell geht, steht in dessen Kontext und ist
keines mehr. Eine verschlüsselte Sicherung lässt sich dort deshalb weder
anlegen noch ansehen — die Meldung sagt, dass sie ein Kennwort braucht.

---

## Die Prüfung nach dem Schreiben (F05)

**Eine Sicherung bekommt ihren Namen erst, wenn sie ihre Prüfung besteht.**
`VACUUM INTO` schreibt in eine Nachbardatei; die wird festgeschrieben, nur
lesend geöffnet und mit `PRAGMA quick_check` gelesen. Besteht sie nicht, wird
sie **verworfen** — unter dem gewählten Namen liegt dann nichts, und das
Fenster sagt „die Sicherung bestand ihre Prüfung nicht und wurde verworfen".
Der Bestand selbst bleibt unberührt. **Belegt** durch
`TestEineSicherungDieIhrePruefungNichtBestehtWirdVerworfen` (`internal/storage`):
eine Datenseite der Nachbardatei wird zwischen Schreiben und Prüfen
verbogen.

`quick_check` und nicht `integrity_check`: es lässt den Abgleich der Indizes
mit ihren Tabellen aus und ist dafür um ein Vielfaches schneller; eine
zerschriebene Seite findet es trotzdem.

---

## Die Wiederherstellung (AK-03)

**Erst sehen, dann tauschen.** Die Vorschau nennt Zeitpunkt, Schemastand,
Sets, Karten, was mit den Bildern geschieht, ob die Sicherung verschlüsselt
ist, **welche eigenen Pakete mitkommen und welche bezogenen neu zu beziehen
sind**.

**Geprüft wird, was ankommt, und nicht, was gewählt wurde**:

1. **Integrität** — jede Datei gegen Größe und SHA-256 aus dem Manifest, eine
   verschlüsselte zusätzlich über die Authentifizierung von age; die
   Datenbank wird nur lesend geöffnet, **mit `PRAGMA quick_check` Seite für
   Seite gelesen** und auf ihren Schemastand geprüft. Der Seitenlauf kam mit
   F05 (#2344):
   bis dahin ging eine Sicherung mit einer zerschriebenen Datenseite durch
   das Öffnen und durch die Buchführung. **Scheitert die Prüfung selbst, gilt
   die Sicherung nicht als bestanden.**
2. **Vollständigkeit** — jeder Teil, den das Manifest nennt, muss da sein;
   was fehlt, weist die Sicherung ab.
3. **Ohne Netz** — nichts davon fragt einen Dienst. Bezogene Pakete, die
   fehlen, werden **genannt** und nicht still geholt.

**Der alte Stand bleibt wiederherstellbar**: der ersetzte Bestand liegt als
`bestand-vor-rueckspielung.db` daneben, die Bilder als `beiwerk-vorher`;
scheitert ein Schritt, geht der vorige zurück, und ein Absturz dazwischen wird
beim nächsten Start zu Ende geführt. **Eigene Pakete werden ergänzt und nicht
überschrieben** — was hier inzwischen bearbeitet wurde, bleibt.

**Belegt** durch `TestEineVerschluesselteSicherungKommtAufEinemLeerenBestandGanzZurueck`
(`internal/app`): gesichert mit Kennwort und zwei Profilen, zurückgespielt auf
einen **leeren** Bestand an einem anderen Ort, danach **Tabelle für Tabelle**
verglichen — jede Zeile jeder Tabelle, mit dem Vermerk über die Rückkehr als
einziger Ausnahme. Und durch `TestEigenePaketeReisenMitUndBezogeneAlsVerweis`
(`internal/service`).

---

## Das Alter und der freiwillige Plan (OPS01)

**Wann zuletzt gesichert wurde, steht im Bestand** — OPS01
(#2347),
Befund A-06 des Audits 0.9.0, in der Tafel `sicherungsvermerke` aus M01.

- **Gesetzt wird die Zeit erst nach zurückgelesener und geprüfter Datei.** Die
  Sicherung entsteht in einer Nachbardatei, wird festgeschrieben, **ganz
  zurückgelesen** — jeder Teil gegen Größe und SHA-256 aus seinem Manifest,
  eine verschlüsselte durch age hindurch — und erst dann unter ihrem Namen
  veröffentlicht (`internal/sicherung/zuruecklesen.go`). Danach steht der
  Vermerk: Zeit, Anlass (von Hand oder nach Plan), Fläche, Dateiname,
  Größe und Prüfsumme der Datei. **Eine Sicherung, die scheitert, ändert die
  Zeit nicht.** Scheitert erst der Vermerk, bleibt die geprüfte Datei liegen,
  und der Fehler sagt beides (`sicherungUnvermerkt`). Andersherum —
  erst vermerken, dann veröffentlichen — nennte ein Absturz dazwischen eine
  Sicherung, die es nicht gibt.
- **Das Alter** sehen alle vier Flächen gleich: am Fenster unter
  *Einstellungen › Daten und Sicherung*, `mse bestand-sicherungsstand`,
  `lesen.bestand.sicherungsstand` und `GET /bestand/sicherungsstand`. **Über
  14 Tage** heißt es `alt` und ist hervorgehoben. **Unbekannt heißt
  unbekannt**: ein Bestand, der über diese Fassung nie gesichert wurde, nennt
  keinen Zeitpunkt. **Die Sicherung vor einer Wanderung zählt nicht** — sie
  liegt neben dem Bestand und ohne Bilder. **Die Probe vor einer
  Aktualisierung wird gar nicht vermerkt** (`Bestandsweg.Probesichern`): sie
  wird geschrieben, zurückgelesen und gleich wieder gelöscht. Liegt der Vermerk nach der Uhr,
  heißt es `unstimmig`.
- **Der Plan ist freiwillig und beginnt erst nach einer Ordnerwahl**
  (`sicherung.planordner`, Ebene `geraet`): einmal die Woche eine Sicherung
  `bestand-plan-<Weltzeit>.msesicherung` in diesen Ordner, **sieben eigene
  bleiben**. Eigen ist eine Datei, die ein Vermerk wiedererkennt — Name des
  Plans, eine gewöhnliche Datei und kein Verweis, Größe und SHA-256 wie
  vermerkt. **Alles andere bleibt liegen**, auch eine Datei unter dem Namen
  des Plans, die niemand vermerkt hat, und eine, die sich seither geändert
  hat. **Weggeräumt wird erst nach einer geglückten, geprüften und vermerkten
  Sicherung**: eine volle Platte, ein Abbruch oder ein fehlender Stick nehmen
  nichts weg. Ein fehlender Ordner wird **nicht angelegt** — sonst läge die
  Sicherung unter `/Volumes` auf der eigenen Platte.
- **Der Plan läuft im Fenster**, kurz nach dem Start, danach stündlich gefragt
  und gleich nach einer Ordnerwahl. **Zwei Prozesse** halten eine Sperre
  neben dem Bestand (`bestand.db.plan`); wer sie nicht bekommt, sichert nicht.
  Ein Tausch des Bestands beendet ihn vorher und startet ihn danach neu.
- **Was er nicht kann**: sichern, solange das Fenster zu ist — dafür gibt es
  `mse bestand-sichern` aus einem eigenen Zeitplan. **Nach einem
  Zurückspielen** kennt der Bestand die jüngeren Vermerke nicht mehr; ihre
  Dateien bleiben dann liegen, bis jemand sie wegräumt. Der Ordner ist eine
  Einstellung **dieses Rechners** (Ebene `geraet`) und reist mit keinem
  Profil; gesichert wird der ganze Bestand.

---

## Aufbewahrung, Platz, Rückweg, Rollen (AK-04)

**Aufbewahrung.** Eine Sicherung ist eine Datei am Ort, den jemand wählt; die
Anwendung räumt sie nie weg — **bis auf die eigenen des freiwilligen Plans**,
von denen sieben bleiben (oben). **Sicherungen vor Migrationen** entstehen von
selbst (`bestand-vor-<Fassung>.db`) und lassen sich unter *Einstellungen ›
Daten und Sicherung* räumen — **bis auf den Weg zurück**: es bleiben die
jüngste Datei überhaupt und die jüngste, die **diese** Fassung öffnen kann.

**Platz.** Die Platzübersicht nennt Bestand, Bilder, Pakete, Vorsicherungen und
Protokoll mit Gewicht; eine Sicherung überschreibt nie eine vorhandene Datei.

**Rückweg zur kompatiblen Fassung.** Eine Sicherung aus einer **neueren**
Fassung wird abgewiesen (die Fassung steht im Manifest); zurück zu einer
älteren Fassung geht es über die Sicherung vor der Migration, die diese Fassung
angelegt hat. **Und zwar durch das Fenster der älteren Fassung**: es öffnet
den neueren Bestand nicht, zeigt die Wiederherstellung, nennt die Sicherung,
die es öffnen kann, und spielt sie zurück — belegt am 2026-10-14 mit 0.8.0 und
25.000 Karten (`aufstieg09`, M01). **Die Kommandozeile von 0.8.0
kann es nicht**: `bestand-zurueckspielen` wandert den Bestand, bevor es
zurückspielt, und scheitert an der unbekannten Wanderung (gemessen). **Ab
0.9.0 wandern `bestand-schauen` und `bestand-zurueckspielen` vorher nicht**
(M02, #2346): über dem Bestand einer neueren Fassung lässt sich die Sicherung
ansehen und zurückspielen, und gewandert wird erst der zurückgespielte —
`TestZurueckGehtAuchUeberEinemNeuerenBestand`.

| Handlung | Träger | Wer darf | Welche Bestätigung |
| --- | --- | --- | --- |
| Sichern | Bestand | wer den Bestand öffnen kann; am Dienst und über MCP die **Schreibwache** — ein Leser nicht | der Speichern-Dialog am Fenster; ein Kennwort zweimal |
| Ansehen | Bestand | wer den Bestand öffnen kann | — es ändert nichts |
| Zurückspielen | Bestand | **nur am Fenster und an der Kommandozeile** — nicht über MCP (Umkehrbarkeit, [0044](entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)), nicht am Dienst | am Fenster die Rückfrage mit den Zahlen der Sicherung; an der Kommandozeile der **Zeitpunkt** aus `bestand-schauen` |
| Vorsicherungen und Protokolle räumen | Bestand | wer den Bestand öffnen kann; am Dienst und über MCP die Schreibwache | am Fenster eine **Rückfrage** je Posten — bis OPS-06 löschte ein Klick sofort |
| Ausführen (Set, Sammlung, Bilder) | Profil | Rolle **mitarbeiter** — der Riegel in der Ausfuhr ([0078](entscheidungen/0078-der-wirksame-akteur.md)) | der Dateidialog |

**Warum der Träger der Bestand ist und nicht ein Profil**: eine Sicherung trägt
**alle** Profile, und wer sie zurückspielt, tauscht sie alle. Der Bestand
gehört dem Rechner — Profile und Rollen trennen Arbeit, nicht Menschen, die
dieselbe Anmeldung am Mac teilen ([`AUTORISIERUNG.md`](AUTORISIERUNG.md),
Bedrohungsmodell). Wer eine Sicherung **ohne** Kennwort bekommt, bekommt den
ganzen Bestand; mit Kennwort nur, wer es kennt.
