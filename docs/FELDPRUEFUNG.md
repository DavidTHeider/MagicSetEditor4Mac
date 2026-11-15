# Feldprüfung — was jede freie Eingabe erwartet

**Werkstück F von Phase 4.10**
(#1325).

`AGENTS.md` §6 verlangt: *„Fehler erscheinen am verursachenden Feld, nicht als
anonymer Hinweis."* Gemessen am **2026-09-03**: **58 freie Eingaben** in
**43 Dateien**, und **zwei** Dateien setzen `aria-invalid`.

**Neu gemessen am 2026-09-06**: **61 freie Eingaben** in **47 Dateien**, und
**sechs** Dateien setzen `aria-invalid`. Die eine kam mit
#1401 dazu —
das Geheimnisfeld eines Anmeldeanbieters, mit derselben Begründung wie das
Zugangsfeld daneben.

**Neu gemessen am 2026-09-08** — Werkstück U3 von Phase 5.1
(#1414):
**59 freie Eingaben** in **47 Dateien**. Zwei weniger als am 2026-09-06, und
**keine ist verschwunden**: die vier Zahlen eines Formats und die behaltene
Stückzahl einer Sammlung liegen jetzt in
[`Zahlenfeld`](../frontend/src/bausteine/Zahlenfeld.tsx), das Muster einer
Autoersetzungsregel und der Name eines Formats in
[`Pflichtfeld`](../frontend/src/bausteine/Pflichtfeld.tsx).
`bausteine/Sammlungsordnung.tsx` fällt damit aus der Tabelle — dieselbe
Bewegung wie bei `Bogeneinstellungen` am 2026-09-03.

**Neu gemessen am 2026-09-09** — Werkstück U7 von Phase 5.1
(#1418):
**64 freie Eingaben** in **51 Dateien**. Eine dazu, die Suche über Pakete und
Dateien im Baum der Vorlagenwerkstatt — ungeprüft aus demselben Grund wie der
Schnellfilter: ein Suchfeld nimmt jede Zeichenfolge.

**Neu gemessen am 2026-10-02** — UXR-08
(#1916):
**eine dazu**, die Suche über der Wahl „Worüber“ der Statistik
([`Statistikwahl`](../frontend/src/seiten/Statistikwahl.tsx)). Sie steht
erst ab fünfzehn Einträgen da — an einem Set aus `magic.mse-game` sind es
fünfzig Dimensionen und gut vierzig Felder — und ist **ungeprüft** wie jede
Suche. Die zwei Auswahlfelder „Feld“ und „Dimension“, die vorher dort
standen, sind eine Wahl geworden; sie zählten hier nie mit.

**Neu gemessen am 2026-10-03** — UXR-F18
(#2033):
**eine dazu**, die Suche über die Drucke der eigenen Sammlung
([`Druckwahl`](../frontend/src/bausteine/Druckwahl.tsx)) — ungeprüft wie
jede Suche. Die Behaltemenge, die dort neben der Wahl steht, ist **kein**
neues Feld: sie liegt in `Zahlenfeld`
([`Behaltemenge`](../frontend/src/bausteine/Behaltemenge.tsx)).

**Neu gemessen am 2026-10-21** — #2538
(Issue (#2538)):
**eine dazu**, die Suche über einem großen Fach des Markttischs
([`Fachliste`](../frontend/src/bausteine/Marktfach.tsx)) — ungeprüft wie
jede Suche.

**Neu gemessen am 2026-10-17** — AP-10
(#2367):
**eine dazu**, die Suche nach einem Druck über dem Formular für ein neues
Exemplar ([`Drucksuche`](../frontend/src/bausteine/Drucksuche.tsx)) —
ungeprüft wie jede Suche.

**Neu gemessen am 2026-09-11** — Werkstück A7 von Phase 5.3
(#1446):
**zwei dazu** — Modell und Auftrag eines KI-Vorschlags, beide ausdrücklich
ungeprüft. **Die Adresse der KI-Laufzeit zählt nicht mit**: sie steht in
[`Wahlfreitext`](../frontend/src/seiten/Wahlfreitext.tsx), wie jede andere
Einstellung ohne Werteliste — dieselbe Bewegung wie bei `Sammlungsordnung` am 2026-09-08.
Geprüft ist sie trotzdem, und zwar dort, wo die Regel steht:
`domain.PruefeKIAdresse`.

**Neu gemessen am 2026-09-17** — die **Handlungsleiste**
(#1623):
**zwei dazu** in zwei Dateien — ein Feld je Spielbefehl und eines für die
vier Angaben eines Spielsteins. **Alle ausdrücklich ungeprüft**, und der Grund steht in
beiden Zeilen: was ein Spielbefehl annimmt, entscheidet der Spielkern (§2),
und was ein Spielstein ist, sagt der, der ihn erzeugt (CR 701.7a).

**Neu gemessen am 2026-09-10** — Werkstück O12 von Phase 5.2
(#1439):
**drei dazu** — Konto, Betrag und Währung des Kostenjournals. Der Betrag ist
**im Fenster** geprüft, und zwar auf eine ganze Zahl ab null: er steht in der
kleinsten Einheit, und ein Komma wäre dort ein Missverständnis.

**Neu gemessen am 2026-09-10** — Werkstück O4 von Phase 5.2
(#1428):
**zwei dazu**, der Name zur Bestätigung einer Rücksetzung und der Nachweis für
die ganze Anwendung. Beide sind **geprüft**, und zwar dort, wo die Antwort
liegt: im Bestand.

**Neu gemessen am 2026-09-09** — Werkstück U5 von Phase 5.1
(#1416):
**60 freie Eingaben** in **48 Dateien** — eine dazu, der Schnellfilter über
der Kartenliste, ausdrücklich ungeprüft.

**Neu gemessen am 2026-09-22** — PLAY-03.AK-05
(#1711):
**eine dazu**, die Suche über die eigenen Partien. **Ungeprüft**, aus
demselben Grund wie die Einstellungssuche und der Schnellfilter: ein Suchfeld
nimmt jede Zeichenfolge, und kein Treffer ist ein **Leerzustand**, kein
Fehler — er steht als eigener Satz da und nennt den Weg zurück. **Die beiden
Wahlfelder daneben zählen nicht mit**: Stand und Ordnung kennen nur, was
angeboten wird, und angeboten wird beim Stand nur, was wirklich vorkommt.

**Neu gemessen am 2026-09-12** —
#1539:
**eine dazu**, der Höchstbetrag einer KI-Freigabe. Sie ist **geprüft**, und
zwar im Fenster: eine ganze Zahl größer als null. **Die Währung ist keine
Eingabe** — anders als beim Kostenbudget daneben kommt sie aus dem Konto,
denn eine Freigabe, die in einer anderen Währung lautet als das Budget,
trägt keinen Auftrag (`domain.Kostengrenze`).

**Neu gemessen am 2026-09-13** — Werkstück A13.4 von Phase 5.3
(#1452):
**eine dazu**, die Belegfrage eines KI-Vorschlags. Sie ist **ungeprüft** und
**leer erlaubt** — dann wird der Suchindex gar nicht gefragt. **Die Freigabe
des eigenen Bestands zählt nicht mit**: sie ist ein Kästchen und keine freie
Eingabe, und was sie freigibt, prüft `domain` (eine Freigabe ohne Wirkung
wird abgewiesen).

**Neu gemessen am 2026-09-13** — Werkstück A13.2b von Phase 5.3
(#1452):
**eine dazu**, das Format einer Balanceanalyse. Sie ist **ungeprüft**, und der
Grund ist ein inhaltlicher: dieses Programm kennt kein Turnierformat, und eine
Werteliste über „Commander, Pauper, …" wäre eine Behauptung über fremde
Regelwerke, die niemand hier pflegt. **Leer ist erlaubt** — und dann steht in
der Antwort, dass der Vergleich für kein Format besonders gilt. Ein Feld, das
eine Einschränkung erzeugt statt eines Fehlers, gehört nicht an die Feldwache.

**Neu gemessen am 2026-09-13** — Werkstück A13.1b von Phase 5.3
(#1452):
**eine dazu**, der Auftrag eines Batches. Sie ist **ungeprüft**, aus
demselben Grund wie beim einzelnen Vorschlag: ein Auftrag an ein Sprachmodell
ist Fließtext. **Die Kartenauswahl daneben zählt nicht mit** — Ankreuzfelder
sind gebunden, und eine Wahl kennt nur, was dasteht.

**Neu gemessen am 2026-09-14** — Werkstück N1 von Phase 5.4
(#1454):
**drei dazu**, alle bei den Peerkontakten. Zwei davon sind **geprüft**, und
zwar an der einzigen Stelle, an der es zählt: die Kennung und der verglichene
Fingerabdruck werden in Go gegen den Schlüssel gehalten, und ein Irrtum kommt
als Meldung zurück. Die dritte — wie man ein Gegenüber selbst nennt — ist
**ungeprüft**: ein Name ist Fließtext, und dieser hier ist ausdrücklich kein
Nachweis.

**Neu gemessen am 2026-09-25** — UX-05 S07.NAT-02
(#1677):
**eine dazu**, die Suche über Stile — beim neuen Set und auf der Stilseite
eines offenen Sets, dieselbe `Stilliste`. **Ungeprüft**, aus demselben Grund
wie die übrigen Suchen: jede Zeichenfolge ist gültig, und kein Treffer ist ein
Satz, kein Fehler. **Die Stile darunter sind Optionsfelder** und keine freie
Eingabe; sie kennen nur, was installiert ist. **Aus der Tabelle fällt nichts**:
die Auswahllisten, die sie ersetzen, zählten nie mit.

**Und noch einmal am 2026-09-25** — UX-08.AK-04
(#1680):
**vier Felder „Welcher Handel“ werden eine Wahl.** Vorgang und Buchung haben
keine freie Eingabe mehr, Zusagen und Ausgang je eine weniger; dazu kommt
**eine**, die Kennung für Fachleute in `Handelswahl`. Die Besitzzeile samt
Fassung und die Nachrichtenkennung stehen seither unter „Für Fachleute“ —
gewählt wird aus der Sammlung und aus der Liste dessen, was aussteht.

**Neu gemessen am 2026-10-10** — F03
(#2245):
**sechs freie Eingaben weniger als vor F03**, und keine ist verschwunden —
sie liegen in zwei Bauteilen, die eine Zahl wirklich lesen. **Der Anlass war
eine Klasse, kein Feld**: vier Stellen lasen eine Zahl mit `parseInt` und
gaben weiter, was niemand getippt hatte — ein anderes Exemplar, einen
anderen Spielstand, eine andere Leihdauer, einen anderen Port. Die drei
Zahlen einer Zusage und der Spielstand einer Paarung liegen jetzt in
[`Ganzzahlfeld`](../frontend/src/bausteine/Ganzzahlfeld.tsx), die Stunden
einer Deckleihe in `Zahlenfeld`. **Und dieselbe Klasse eine Stufe leiser**:
sechs Stellen mit `type="number"` gaben `Number(...)` an ein `int` in Go — ein leeres
Feld wurde null, „2.5“ kam als Satz über JSON zurück statt am Feld, und die
Zuordnungswahl machte aus „0“ still eine Eins. Die Menge einer
Handelsposition liegt jetzt in `Ganzzahlfeld`; Zuordnung, Deckkarte,
Inspector und die zwei Budgetzahlen einer Kontextfreigabe in `Zahlenfeld`,
und Deckausset liest streng. `bausteine/Handelszusagen.tsx` und
`bausteine/Zuordnungswahl.tsx` fallen aus der Tabelle.

**Neu gemessen am 2026-10-02** — UXR-07
(#1915):
**eine dazu**, und keine neue Frage. Die Stilseite eines offenen Sets nimmt
statt der `Stilliste` die
[`Stilauswahl`](../frontend/src/bausteine/Stilauswahl.tsx) — **dieselbe
Suche**, aber über einem Auswahlfeld statt über einem Optionsfeld je Stil:
das waren in der Schublade *Eigenschaften* gemessen 382 Felder
(`docs/UIREVIEW.md`, R3). **Ungeprüft** wie jede Suche. Die `Stilliste`
bleibt beim neuen Set und behält ihre Zeile.

**Und noch einmal am 2026-10-02** — UXR-04
(#1912):
**eine weniger**, und keine Frage verschwindet. Auch das neue Set nimmt die
`Stilauswahl` — mit dem Full-Magic-Pack zeichnete die `Stilliste` dort 355
Optionsfelder. Sie hat damit keinen Aufrufer mehr und fällt samt ihrer Zeile
weg; ihre Suche ist dieselbe, die die `Stilauswahl` schon führt.

**Und am 2026-10-03** — UXR-F1
(#2007):
**keine mehr und keine weniger**, nur ein anderer Ort. Die `Stilauswahl` hatte
ein Suchfeld über einem Auswahlfeld; sie nimmt jetzt die
[`Suchauswahl`](../frontend/src/bausteine/Suchauswahl.tsx), in der das Feld
selbst sucht, und ihre Zeile geht dorthin. Dieselbe Auswahl nehmen Testkarte,
Spielwahl der Paketanlage und Zeitzone — die zwei letzteren hatten ein
natives Auswahlfeld beziehungsweise einen Freitext in
[`Wahl`](../frontend/src/seiten/Wahl.tsx); der bleibt für jede andere
Einstellung ohne Werteliste.

**Und weiter am 2026-10-03** — UXR-F1 (#2007), die übrigen R7-Stellen:
**eine weniger**. Die [`Statistikwahl`](../frontend/src/seiten/Statistikwahl.tsx)
trägt ihr Suchfeld nicht mehr; ab fünfzehn Einträgen sucht die `Suchauswahl`
selbst, deren Zeile oben steht, und darunter bleibt es ein natives Feld. Das
[`Auswahlfeld`](../frontend/src/bausteine/Auswahlfeld.tsx) und die Kartenfelder
mit fünfzehn und mehr Wahlen bringen **keine neue Eingabe**: sie rufen die
`Suchauswahl` auf, die in dieser Tabelle schon steht.

**Acht Dateien nennen `aria-invalid` selbst**, gezählt am 2026-09-08 über
`grep -rl` in `frontend/src` ohne Tests. **Diese Zahl misst seit Werkstück U2
weniger als früher**: [`Feld`](../frontend/src/bausteine/Feld.tsx) trägt das
Attribut jetzt für alle, die auf ihm aufbauen — `Pflichtfeld`, `Zahlenfeld`,
`Schluesselwortfeld` und das Samenfeld des Zufallspacks kommen darin nicht mehr
vor, obwohl sie es setzen. Wer die Zusage prüfen will, liest die Marke in der
Tabelle und die Tests daneben, nicht diesen Zähler (§10: die Reichweite einer
Messung ist die Reichweite ihrer Aussage).

**Und am 2026-09-05**: **60 freie Eingaben** in **46 Dateien**, und
**sechs** Dateien setzen `aria-invalid`. Zwei kamen in Phase 4.12 dazu
(#1351):
[`Kennwortfeld`](../frontend/src/bausteine/Kennwortfeld.tsx) aus Werkstück 6
und das **Textfeld in [`Wahlfreitext`](../frontend/src/seiten/Wahlfreitext.tsx)** aus
Werkstück 9 — es steht dort, wo eine Einstellung **keine** Werteliste hat, und
ein Verzeichnis lässt sich nicht aufzählen.

> **Der Satz darüber war beim Aufschreiben um eins zu hoch.** Mit derselben
> Zählung (`freieEingabenZaehlen` aus dem Wächter) und **ohne** das neue
> Bauteil kommen heute **57 in 43** heraus, nicht 58. Die Dateizahl stimmte.
> **Ein Wächter hält die Zeilen der Tabelle, diesen Satz hält keiner** — er
> ist Fließtext, und dafür gibt es hier keinen (§10: die Reichweite einer
> Messung ist die Reichweite ihrer Aussage).
>
> **Und die vier neuen `aria-invalid` sind nicht neu**, sondern nie gezählt
> worden: `Feldeingabe`, `Massenaenderung`, `Bildereinstellungen` und
> `Zahlenfeld` setzen es längst. Gezählt am 2026-09-05 über `grep -rl` in
> `frontend/src` ohne Tests.

**Diese Liste ist die Entscheidung, nicht die Beobachtung.** Je Datei steht
hier, was ihre Eingaben erwarten und ob hier geprüft wird — **oder
ausdrücklich, dass nicht geprüft wird, und warum**. Ein Suchfeld nimmt jede
Zeichenfolge; das ist eine Antwort und keine Lücke.

**Zwei der Zahlen waren beim ersten Aufschreiben falsch, und der Wächter hat
es gesagt.** Er las rohen Text und zählte `<input>` und `<textarea>` **in
Kommentaren** mit — in `Quelltextfeld.tsx` und `Wortlistenwahl.tsx` stehen
sie dort als Erklärung. Seit dem 2026-09-03 blendet er Kommentare aus; die
Zahlen oben sind die berichtigten.

**Ein Wächter hält sie**: `internal/werkzeuge/feldpruefung_test.go` zählt die
freien Eingaben je Datei und hält sie gegen diese Tabelle. Eine neue Eingabe
erzwingt damit die Entscheidung, statt still in der Menge zu verschwinden —
dieselbe Bauform wie bei [`SCHREIBWEGE.md`](SCHREIBWEGE.md).

---

## Kennungsfelder

**Felder, deren Wert genau so ankommen muss, wie er getippt wurde** — #2389.
Gefunden bei der Ausgangsmessung P0 (#2336): im Feld „Druck" schrieb macOS aus
`p0-blitz` ein `P0-blitz`, und das eigene Exemplar passte im Handel nicht
mehr. Jedes dieser Felder trägt `autocapitalize`, `autocorrect` und
`spellcheck` aus — über [`kennungseingabe`](../frontend/src/bausteine/kennungseingabe.ts),
und `frontend/src/kennungsfelder.test.ts` hält diese Tafel in beiden
Richtungen: jede Datei hier trägt so viele Felder, wie die Zahl sagt, und
jede Datei mit einem Kennungsfeld steht hier.

| Datei | Felder | Was |
| --- | ---: | --- |
| `bausteine/Handelspositionsform.tsx` | 1 | der **Druck** einer Position |
| `bausteine/Vorschlagsfeld.tsx` | 1 | jeder Code mit Vorschlägen — **Sprache**, **Währung**, **Spiel** |
| `bausteine/Kontaktaufnahme.tsx` | 1 | die **Kennung** eines Gegenübers; der Name daneben ist frei |
| `bausteine/Abdruckvergleich.tsx` | 1 | der verglichene **Fingerabdruck** |
| `bausteine/Einladungsannahme.tsx` | 1 | der **Einladungscode** |
| `bausteine/Kontaktadresse.tsx` | 1 | die **Adresse** eines Gegenübers |
| `bausteine/Handelswahl.tsx` | 1 | die **Kennung eines Handels** |
| `bausteine/Handelsausgang.tsx` | 1 | die **Kennung einer Nachricht** |
| `bausteine/Horchposten.tsx` | 1 | die **Adresse**, unter der gehorcht wird; der Name ist frei |
| `bausteine/Dienstposten.tsx` | 2 | **Adresse** und **Port** des Dienstes |
| `seiten/Paketbezug.tsx` | 1 | **Adresse, Stand, Verzeichnis, Name** eines Bezugs — ein Feld, viermal gebaut |
| `seiten/Ruecksetzung.tsx` | 1 | der **Name**, der zur Bestätigung abgetippt wird |

**Am nativen Fenster gemessen, am 2026-10-16**
([`druckfeld09`](belege/huellenprobe/druckfeld09.md)): `p0-blitz` und
`p0-schock` mit Tasten getippt, angeboten, und `mse handel liste --json` trägt
sie so. **Die Gegenprobe ist dasselbe Drehbuch an einem Bündel, dessen Feld
„Druck" die drei Eigenschaften nicht trägt** — dort stand `P0-blitz` und
`P0-schock` im Handel. WebKit unter macOS beachtet sie also; welche der drei
es braucht, ist nicht getrennt gemessen, und gesetzt sind alle.

**Das Feld selbst zeigte dabei auch in der Gegenprobe den getippten Wert.**
macOS schreibt den Anfang erst groß, wenn das Feld verlassen wird — wer den
Wert des Felds nach dem Tippen liest, misst nichts. Gemessen wird deshalb am
Handel.

## Die drei Klassen

| Marke | Heißt |
| --- | --- |
| **geprüft** | Die Regel ist **hier** bekannt, und das Feld sagt es am Feld — mit `aria-invalid` und `aria-describedby` |
| **offen** | Die Regel ist hier bekannt und wird **noch nicht** geprüft. Ein Befund mit Adresse |
| **ungeprüft** | Ausdrücklich nichts zu prüfen — mit Grund |

**Warum es „ungeprüft" gibt.** Das Issue sagt es: *„Und nicht prüfen, wo die
Regel woanders steht. Was in einem Paket gilt, weiß das Paket; eine hier
erfundene Regel machte ein gültiges Paket unbrauchbar (§11)."*

---

## Die Dateien

**84**, gezählt am 2026-09-26. Hier stand bis dahin „Die dreiundfünfzig
Dateien" über 83 Zeilen: eine Zahl in einer Überschrift steht unter keinem
Wächter — der hält die Eingaben **je Datei**, nicht ihre Summe.

| Datei | Eingaben | Arten | Marke | Was sie erwartet |
| --- | ---: | --- | --- | --- |
| `bausteine/Kontaktaufnahme.tsx` | 2 | text | **geprüft** | Die **Kennung** ist Pflicht — der Knopf bleibt ohne sie angehalten —, und ob sie eine ist, entscheidet `internal/service`: eine Zeichenkette, aus der kein Schlüssel wird, wird abgewiesen, statt als Kontakt mit leerem Abdruck dazustehen. Der **Name daneben ist frei** und belegt nichts (N1, #1454) |
| `bausteine/Abdruckvergleich.tsx` | 1 | text | **geprüft** | Der **verglichene Fingerabdruck**, aus `Peerkontaktzeile` herausgelöst (NET-01, #1705). Er wird nicht im Fenster geprüft, sondern gegen die Kennung gehalten — dort, wo der Schlüssel liegt; stimmt er nicht, kommt es als Meldung zurück. **Das ist der Punkt der Eingabe**: ein Häkchen „ja, stimmt" belegte nichts (N1, #1454) |
| `bausteine/Einladungsausgabe.tsx` | 2 | text, Auswahl | **geprüft** | Der **Name** in der eigenen Einladung — ein bis 64 Zeichen, **ohne unsichtbare Zeichen** (`domain.Kontakteinladung`, ein U+202E kehrte die Schreibrichtung um); der Knopf steht angehalten, solange er leer ist. Der **Zweck** ist eine Auswahl aus den Zwecken, die Go schickt (NET-01, #1705) |
| `bausteine/Einladungsannahme.tsx` | 1 | Textbereich | **geprüft** | Der **Einladungscode** des Gegenübers. Geprüft wird im Dienst, und zwar zuerst die **Unterschrift** über die Bytes, die ankamen, dann Fassung, Ablauf und Zweck; Leerraum aus einem Messenger zählt nicht. Der Fehler steht am Feld (`aria-invalid`) (NET-01, #1705) |
| `bausteine/Einladungsvergleich.tsx` | 2 | text | **geprüft** | Der **eigene Name** für das Gegenüber — höchstens 64 Zeichen (`service.aufnahmeAus`); leer heißt: es heißt, wie es sich nennt. Er geht nie hinaus (NET-01, #1705). **Und der Name, der hinausgeht** — nur bei einer einlösbaren Einladung gefragt, leer vorgegeben, ein bis 64 lesbare Zeichen **ohne unsichtbare** (`service.aufnahmeAus` über `domain.BrauchbarerName`); leer heißt: ohne Namen (NET-01b, #1954) |
| `bausteine/Kontaktadresse.tsx` | 1 | text | **geprüft** | Die **Adresse**, unter der ein Gegenüber erreichbar ist. Geprüft wird sie in `internal/domain` (`PruefeAdresse`) und nicht hier: Wirt und Port, mit eckigen Klammern für IPv6. **Ein leerer Wert ist gültig** und löscht die eingetragene — „ich kenne keine mehr" ist eine Aussage, und deshalb hält der Knopf nur an, solange sich **nichts geändert** hat (N2, #1455) |
| `bausteine/Kontaktnotiz.tsx` | 1 | text | **ungeprüft** | **Und das ist entschieden.** Die **Notiz**, die dieses Profil sich zu einem Gegenüber macht — Freitext für den, der ihn schreibt. Es gibt nichts, wogegen sie zu halten wäre: sie geht über keine Leitung, eine Sichtkarte trägt sie nicht, und eine Freigabe erst recht nicht. **Ein leerer Wert ist gültig** und löscht die vorhandene — dieselbe Form wie bei der Adresse, und der Knopf hält nur an, solange sich nichts geändert hat (T3, #1468) |
| `bausteine/Gestaltungsbeiwerk.tsx` | 1 | textarea | **geprüft** | Die **eigenen Farben** (UX-09.AK-04, #1681): eine Zeile, die die Form verfehlt, nennt ihren Befund am Feld mit der Zeilennummer (`aria-invalid`) und wird nicht geschrieben; geschrieben wird erst, was die Form hat, und Go prüft dieselbe Form noch einmal (`domain.PruefeGestaltung`). **Eine lesbare Form, die nicht zu lesen wäre**, wird geschrieben und gilt nicht — die Seite nennt das Paar |
| `seiten/Diagnose.tsx` | 1 | text | **ungeprüft** | **Und das ist entschieden.** Der **Filter** über die Diagnose — eine Zeichenfolge, die nur auswählt, welche Zeilen dastehen; er wird nirgends geschrieben und geht nirgends hin. „Diagnose öffnen“ setzt ihn auf eine Vorgangskennung, und wer ihn leert, sieht wieder alles (NET-02, S13.NAT-04, #1706) |
| `bausteine/Marktfach.tsx` | 1 | search | **ungeprüft** | **Und das ist entschieden**, wie bei jeder Suche: sie steht erst über einem Fach mit mehr als 200 Karten (#2538), filtert nach dem Namen, was ohnehin geladen ist, und schreibt nichts; kein Treffer sagt „Keine Karte in diesem Fach heißt so.“ |
| `bausteine/Listenansicht.tsx` | 1 | search | **ungeprüft** | **Und das ist entschieden**, aus demselben Grund wie beim Schnellfilter: ein **Suchfeld** über Sets und Decks nimmt jede Zeichenfolge, und was nichts findet, sagt die Liste — „Nichts passt zu …“ samt „Suche leeren“ (UX-11.AK-04, #1683) |
| `bausteine/Sicherungskennwort.tsx` | 3 | password | **geprüft** | Das Kennwort einer Sicherung, seine Wiederholung und das Kennwort beim Öffnen (#1723). **Die Länge prüft Go** (`domain.PruefeSicherungskennwort`, zwölf Zeichen) und meldet es am Feld; **die Gleichheit prüft das Fenster**, bevor es den Knopf freigibt — ein Tippfehler kostete sonst die Sicherung. Beim Öffnen prüft age selbst: ein falsches Kennwort ist ein Befund am Feld |
| `bausteine/Vorlagenumschlag.tsx` | 1 | text | **geprüft** | Der **Name des Pakets**, das entstehen soll. Geprüft wird er an zwei Stellen: hier auf **leer** — ohne Namen gibt es kein Paket —, und in `app` auf **Gestalt**: er darf kein Verzeichnis nennen, denn ein `../` darin legte ein Paket irgendwohin. **Die Endung erzwingt dieses Feld nicht**: was sie bedeutet, entscheidet `msepaket.ArtAusName`, und die Absage dort nennt die sieben, die es gibt — eine zweite Prüfung hier liefe mit ihr auseinander (#1639) |
| `bausteine/Ganzzahlfeld.tsx` | 1 | text | **geprüft** | **Eine ganze Zahl ab null als Text** (F03, #2245) — die **Besitzzeile**, die **Menge** und die **gesehene Fassung** einer Zusage ([`Handelszusagen`](../frontend/src/bausteine/Handelszusagen.tsx), seit UX-08.AK-04 unter „Für Fachleute“, #1680) und der **Spielstand** einer Paarung ([`Turniermatch`](../frontend/src/bausteine/Turniermatch.tsx)). **Streng gelesen** ([`ganzzahl.ts`](../frontend/src/ganzzahl.ts)): nur Ziffern; was keine ganze Zahl ist, steht am Feld, sobald jemand absendet, und geht nicht hinaus. **Bis dahin las `parseInt`**: „12a“ hielt Exemplar 12 fest, „2.5“ Stück wurden zwei, und „viele“ ging als Spielstand null in ein Turnier. Ob eine Zusage hineinpasst, entscheidet weiter der **Bestand** in einer Transaktion ([domain.ReservierungPasst]), ob ein Stand gilt, `domain.Ergebnis.Pruefe` — eine null geht deshalb hinaus, und Go urteilt (#1643) |
| `bausteine/Handelsausgang.tsx` | 1 | text | **geprüft** | Die **Nachricht**, die weggeräumt wird — seit UX-08.AK-04 (#1680) **unter „Für Fachleute“**: was nach einem Abriss aussteht, hat seinen eigenen Knopf zum Wegräumen. Leer bekommt das Feld seinen eigenen Satz, am Kasten und nicht als anonymer Hinweis (§6). **Keine Form**: die Kennung kommt von außen (#1637) |
| `bausteine/Handelspositionsform.tsx` | 1 | text | **ungeprüft** | **Eine freie Eingabe, drei gewählte** — seit UX-06.AK-01, **seit TRADE-02b (#1951) in einem eigenen Formular je Spalte** und ohne die Auswahl „Von wem“ — die Spalte sagt die Seite. Seit UX-06.AK-01 (#1678). Frei ist der **Druck**; die **Stückzahl** liegt seit F03 (#2245) in [`Ganzzahlfeld`](../frontend/src/bausteine/Ganzzahlfeld.tsx), ab eins, und eine Zeile mit „viele“ oder „2.5“ kommt nicht mehr in den Handel; **Güte und Ausführung werden gewählt**, aus den sieben Zuständen und drei Ausführungen der Sammlung, und die **Sprache** steht in [`Vorschlagsfeld`](../frontend/src/bausteine/Vorschlagsfeld.tsx). **Der Anlass war ein Befund**: die Güte stand als Freitext mit der Vorgabe `NM` da, und verbucht wird gegen `domain.IstZustand` — eine Position mit der Vorgabe scheiterte erst am Ende des Handels. **Was eine Position ist, entscheidet weiter der Vertrag** (`domain.Handelsposition.pruefe`); eine Stückzahl, die keine Zahl ist, wird **null** und nicht `NaN` (#1637) |
| `bausteine/Handelswahl.tsx` | 1 | text | **geprüft** | Die **Kennung eines Handels**, der oben noch nicht steht — der Weg für Fachleute, mit dem Satz, was sie ist (UX-08.AK-04, #1680). **Auf leer geprüft**: „Öffnen“ bleibt ohne sie angehalten. **Ob es den Handel gibt, sagt der Bestand** beim ersten Ruf eines Kastens darunter. **Keine Form**: eine Kennung kann von der Gegenseite kommen. **Sie ersetzt vier Felder** — in Vorgang, Zusagen, Ausgang und Buchung stand dieselbe Frage |
| `bausteine/Tauschabgleich.tsx` | 1 | text | **geprüft** | Der **freigegebene Stand eines Gegenübers**, eingesetzt als JSON — so, wie `Freigabesicht` ihn hergibt. Geprüft wird er in `Kontakttausch.tsx`, **bevor** er hinausgeht: was sich nicht lesen lässt oder weder `angebot` noch `wuensche` nennt, wird abgewiesen. **Der Grund ist die Unterscheidung**: „dazu passt nichts" und „das war kein Stand" sind zwei Auskünfte, und ein ungeprüfter Text bekäme für beides dieselbe (T4, #1469) |
| `bausteine/Horchposten.tsx` | 2 | text | **geprüft** | Zwei Felder für zwei Entscheidungen. Die **Adresse, unter der gehorcht wird**: dieselbe Prüfung wie bei der Adresse eines Kontakts, mit einem Unterschied — **leer ist hier kein gültiger Wert**, und der Knopf bleibt angehalten, solange das Feld leer ist. Was wirklich gebunden wurde, sagt die Antwort und nicht die Eingabe: `:0` bekommt einen Port vom Betriebssystem. Und der **Name, unter dem angekündigt wird**: er ist **frei und belegt nichts** — ein leerer geht durch, und das Gegenüber sieht dann nur den Fingerabdruck (N2, #1455) |
| `bausteine/Dienstposten.tsx` | 2 | text | **ungeprüft** | **Und das ist entschieden.** Zwei Felder für eine Bindung: **Adresse** und **Port**. **Der Port wird seit F03 streng gelesen** ([`ganzzahl.ts`](../frontend/src/ganzzahl.ts)): `parseInt` machte aus „8080x“ den Port 8080 und aus „1e3“ den Port 1; **seit #2269 steht, was kein Port sein kann, am Feld** — nicht nur Ziffern, über 65535 oder (nach dem Klick) leer: `aria-invalid`, der Satz „Eine ganze Zahl zwischen 0 und 65535“ als Befund unter dem Feld, und es geht nichts hinaus. Die Null bleibt ein Port („einen freien suchen“). Das ist die **Gestalt** einer Zahl, nicht die Lage; was der Dienst sachlich abweist, bleibt die allgemeine Meldung. Sie bleiben **ungeprüft in der Sache**, weil die Prüfung, auf die es ankommt, **nicht im Fenster stattfinden darf**: ob eine Bindung den Rechner verlässt, entscheidet [domain.PruefeBindung] — und daran hängt, ob TLS und Zugang Pflicht sind ([0115](entscheidungen/0115-die-schleife-ist-eine-vorgabe.md)). Eine zweite Prüfung hier wäre eine zweite Antwort auf eine **Sicherheitsfrage**, und die im Fenster wäre die ungeprüfte. **Was das Feld abweist, sagt die Antwort**: ein Name wie `localhost` endet mit „ist keine Adresse — ein Name wird aufgelöst"; ein Port, der der Gestalt nach einer ist und trotzdem nicht klappt — belegt, nicht erlaubt —, kommt als allgemeine Meldung zurück. **Und ein leerer Port ist keiner**: nach dem Klick steht der Befund am Feld, und keine stille Null geht hinaus (N6, #1572; bis #2269 ging ein `-1` hinaus) |
| `bausteine/Beschriftetesfeld.tsx` | 1 | text | **ungeprüft** | Ein Bauteil ohne eigene Bedeutung — was das Feld erwartet, weiß sein Aufrufer |
| `bausteine/Tischgespraech.tsx` | 1 | text | **geprüft** | Ein Satz am Tisch: das Feld hält `maxlength` auf 500 Zeichen, und der Knopf bleibt angehalten, solange nichts als Leerzeichen dasteht. **Verlassen wird sich darauf nicht** — `domain.Chat.Pruefe` weist einen leeren oder zu langen Satz ein zweites Mal ab, denn ein Gast schickt nicht nur über dieses Fenster (G5, #1462) |
| `seiten/Dateiaktionen.tsx` | 1 | text | **geprüft** | Der **Name einer Datei im Paket** — neu angelegt oder umbenannt (FMT-02, #1691). Hier auf **leer**: ohne Namen bleibt der Knopf gesperrt. **Was gültig ist, entscheidet Go** — `msepaket.paketpfad` weist einen Ausbruch ab, `service.Anlegen` alles, was keine Vorlage ist, und ein belegter Name ist ein Konflikt; die Meldung steht in der Zeile unter dem Feld |
| `seiten/Paketanlage.tsx` | 1 | text | **geprüft** | Der Vollname ist frei — er darf leer sein, dann gilt der Kurzname. **Der Kurzname ist Pflicht** und steht in [`Pflichtfeld`](../frontend/src/bausteine/Pflichtfeld.tsx); ob er ein Verzeichnisname sein kann, entscheidet `internal/service` und meldet es am Feld (#1350) |
| `bausteine/Deckanlage.tsx` | 1 | text | **geprüft** | Nur die Notiz ist frei. **Der Name ist Pflicht** und steht seit dem 2026-09-03 in [`Pflichtfeld`](../frontend/src/bausteine/Pflichtfeld.tsx) — es meldet, wer das Feld anfasst und leer verlässt. **Das Spiel war bis zum 2026-09-10 ebenfalls frei** und ist jetzt eine Wahl über die Spiele der Sets dieses Profils: eine Eingabe weniger, und drei Handlungen weniger im Kernweg `deck` (Befund 6 des Zweitblicks über Phase 5.1) |
| `bausteine/Deckausset.tsx` | 1 | number | **geprüft** | Die Menge je Karte (FMT-04, #1693). **Geprüft im Dienst**: eine Menge unter eins weist `domain.Deckplatz.Validate` ab, und das Nein steht am Feld (`aria-invalid`, `feldbefund`). Das Set ist eine Wahl über die Sets mit Karten, keine Eingabe |
| `bausteine/Eigenuebersetzungzeilen.tsx` | 2 | text | **geprüft** | Kennung und Name einer neuen Übersetzung (LOC-02, #1688). **Geprüft im Dienst**: die Kennung gegen BCP 47 in kanonischer Schreibweise und gegen die eingebauten Sprachen (`domain.PruefeSprachkennung`), der Name auf eine Zeile mit höchstens 60 Zeichen (`domain.Sprachpaket.Pruefe`); das Nein steht am Feld (`aria-invalid`) und im Hinweis darunter. Die Basis ist eine Wahl |
| `bausteine/Uebersetzungsangaben.tsx` | 3 | text | **geprüft** | Name, Lizenz und Herkunft einer eigenen Übersetzung (LOC-02, #1688). **Geprüft im Dienst** (`domain.Sprachpaket.Pruefe`): eine Zeile, höchstens 60 beziehungsweise 200 Zeichen; die Lizenz ist **zum Ausgeben** Pflicht. Der Stand ist eine Wahl |
| `bausteine/Uebersetzungszeile.tsx` | 1 | textarea | **geprüft** | Ein übersetzter Satz (LOC-02, #1688). **Geprüft im Dienst gegen den deutschen Satz** (`domain.Katalogvorlage.PruefeUebersetzung`): dieselben Platzhalter, schließende Auszeichnung, kein Tag, keine Skript- oder Datenadresse, kein Steuerzeichen, höchstens 4.000 Zeichen; das Nein steht an der Zeile |
| `bausteine/Ausgabeauswahl.tsx` | 1 | search | **ungeprüft** | Das Feld, das die **Teilmenge eingrenzt** — ab fünfzehn Karten, nach dem Namen (UXR-09, #1917). Es wählt nur aus, welche Kästchen dastehen, und schreibt nichts; was gewählt ist, bleibt gewählt, auch wenn es ausgeblendet wird. Kein Treffer sagt „Nichts passt zu …“. **Die Kästchen daneben sind gebunden** und stehen deshalb nicht hier |
| `bausteine/Uebersetzungseditor.tsx` | 1 | search | **ungeprüft** | Die Suche über Pfad, deutschen Satz und Übersetzung (LOC-02, #1688) — sie filtert nur, was ohnehin gezeigt wird, und schreibt nichts |
| `bausteine/Handelsliste.tsx` | 1 | search | **ungeprüft** | Die Suche über die eigenen Handel (TRADE-01, #1715) — jede Zeichenfolge ist gültig, und kein Treffer ist ein leerer Zustand mit „Auswahl zurücksetzen“. **Die Länge prüft der Dienst** (200 Zeichen, `Handelsfilter.Pruefe`), denn Kommandozeile, Dienst und MCP haben kein Feld davor |
| `bausteine/Deckuebernahme.tsx` | 2 | text | **ungeprüft** | Beide Felder sind **Kennungen** aus einem anderen Profil, und was dort gültig ist, weiß der Bestand — nicht diese Seite. Eine hier erfundene Regel („nur Ziffern") wäre eine Behauptung über einen Schlüssel, dessen Gestalt `internal/app` vergibt. **Der Knopf bleibt angehalten, solange eines leer ist** — das ist kein Prüfen, sondern die Regel, dass ein Ruf ohne beide Kennungen keinen Sinn hat. Was falsch ist, sagt die Antwort: „das Profil führt der Bestand nicht" oder „dieser Nachweis reicht nicht so weit" (§8) |
| `bausteine/Deckkartenwahl.tsx` | 2 | text, search | **geprüft** | Werkstück I von Phase 4.10: die Karte wird **gewählt** und nicht getippt — zwei Auswahlfelder statt zweier Kennungsfelder. Frei bleibt die Scryfall-Kennung eines Drucks — **die Menge liegt seit F03 in `Zahlenfeld`**, ganzzahlig ab eins, und hält den Knopf an, solange sie keine ist; der Knopf ist angehalten, solange keines von beiden gewählt ist. **Die Suche über die Karten eines Sets** (UXR-11, #1919) steht erst ab fünfzehn Karten da und nimmt jede Zeichenfolge — sie grenzt nur das Auswahlfeld ein, und die gewählte Karte bleibt angeboten; geprüft ist sie damit so wenig wie jedes andere Suchfeld, und sie muss es nicht sein |
| `bausteine/Formatverwaltung.tsx` | 1 | text | **geprüft** | Der Scryfall-Name ist frei — leer heißt dort „Scryfall kennt es nicht", und das ist gültig. **Das Spiel** steht seit UX-06.AK-01 (#1678) im [`Vorschlagsfeld`](../frontend/src/bausteine/Vorschlagsfeld.tsx), mit den installierten Spielen als Vorschlag und Freitext daneben (`useSpielnamen`). **Der Name ist Pflicht** und liegt in `Pflichtfeld` |
| `bausteine/Drucksuche.tsx` | 1 | search | **ungeprüft** | die Suche nach einem Druck über dem Formular für ein neues Exemplar (AP-10, #2367). Sie nimmt jede Zeichenfolge und reicht sie als Zeile `1 <Name>` an dieselbe Deutung wie die Listeneinfuhr (`ListeDeuten`); leer sucht sie nicht. Ein Treffer füllt Druck, Name, Set und Nummer, kein Treffer sagt ein Satz am Feld. **Die Prüfung des Drucks steht anderswo**: das Feld »Druck« unter »Weitere Angaben« bleibt, und `Exemplar.Validate` weist eine leere Kennung ab |
| `bausteine/Druckwahl.tsx` | 1 | search | **ungeprüft** | die Suche über die Drucke der eigenen Sammlung bei einer Sichtfreigabe „Ausgewählte Drucke“ (UXR-F18, #2033). Sie steht erst ab fünfzehn Drucken da (R7) und nimmt jede Zeichenfolge — sie grenzt nur die Liste ein (höchstens die Wahl und zwölf Treffer), die Wahl bleibt angeboten, und passt nichts, sagt ein Satz es. **Die Prüfung der Wahl steht anderswo**: eine Auswahl ohne Drucke weist die Domäne ab (`Sammlungsfreigabe.pruefeUmfangsangaben`), das Fenster sagt es an der Gruppe, bevor es ruft. Die Behaltemenge liegt in `Zahlenfeld` (ab 1, ganzzahlig, ohne Obergrenze — die Domäne liest 0 als die Vorgabe 1, `Sammlungsfreigabe.MitVorgaben`) |
| `bausteine/Formatzeile.tsx` | 1 | text | **geprüft** | Grenzen eines Formats. **Die vier Zahlen liegen seit dem 2026-09-08 in [`Zahlenfeld`](../frontend/src/bausteine/Zahlenfeld.tsx)** — ab null, ganzzahlig und **ohne Obergrenze**: genau das, was `domain.Format.zahlenPruefen` sagt (`< 0` ist ein Befund, mehr steht dort nicht). **Null bleibt gültig** und heißt keine Grenze, kein Sideboard, beliebig oft. Der Name liegt in `Pflichtfeld` (`Format.Validate`, Feld `name`). Der Befund über **zwei** Felder — die Obergrenze unter der Untergrenze — lässt sich am einzelnen Feld nicht entscheiden und kommt als Feldfehler aus Go; er steht über `verursachendesFeld` an dem Feld, das Go nennt. Frei bleibt der Scryfall-Name: was ein Schlüssel aus `legalities` ist, sagt `Format.Validate`, und eine zweite Regex hier liefe damit auseinander |
| `bausteine/Deckeinfuhr.tsx` | 1 | textarea | **ungeprüft** | Eine Deckliste, Zeile für Zeile — derselbe Grund wie eine Zeile tiefer: was gültig ist, entscheidet die **Zuordnung**, und sie meldet jede Zeile einzeln. **Die Vorschau ist hier die Prüfung**: sie schreibt nichts und sagt je Zeile, was daraus würde (T8, #1473) |
| `bausteine/Listeneinfuhr.tsx` | 1 | textarea | **ungeprüft** | Eine Kartenliste, Zeile für Zeile. Was gültig ist, entscheidet die **Zuordnung** — sie meldet jede Zeile einzeln, und eine Vorprüfung hier wiese Zeilen ab, die der Dienst versteht |
| `bausteine/Massenvorgaenge.tsx` | 1 | text | **geprüft** | Der neue Wert einer Massenänderung — `Massenaenderung.tsx` daneben trägt die Prüfung |
| `bausteine/Quelltextfeld.tsx` | 1 | textarea | **geprüft** | MSE-Script. Die Prüfung ist der **Zerteiler**, und die Stelle steht am Zeichen — genauer als jedes `aria-invalid` |
| `seiten/Wahlfreitext.tsx` | 1 | text | **ungeprüft** | **Ein Ordner, in dem ein Dateidialog aufgeht** — Werkstück 9 von Phase 4.12 (#1351), Punkt 8. Das Feld steht nur da, wo eine Einstellung **keine** Werteliste hat; jede andere ist ein Schalter, Segmente oder eine Auswahlliste (UXR-F2, #2008). **Aus `Wahl.tsx` herausgelöst**, als die Wahl mehr Bedienelemente bekam und über die Zeilengrenze ging. **Ob das Verzeichnis existiert, weiß der Rechner und nicht das Fenster** — und wer einen Pfad tippt, den es noch nicht gibt, hat vielleicht recht. Geprüft wird das eine, was die Domäne beurteilen kann: ein `\x00` beendet in jedem Systemaufruf die Zeichenkette |
| `bausteine/Zahlenfeld.tsx` | 1 | number | **geprüft** | **Das Bauteil, das die Prüfung trägt** — es kennt seinen Bereich, setzt `aria-invalid` und nennt ihn am Feld. **Wer ein Feld darauf umstellt, verliert seine eigene Zeile**: `Bogeneinstellungen.tsx` und `seiten/SetNeu.tsx` stehen seit dem 2026-09-03 nicht mehr in dieser Tabelle, weil sie kein eigenes `<input>` mehr haben. Der Wächter hat es angestrichen, und das ist richtig so — die Prüfung ist umgezogen, nicht verschwunden |
| `bausteine/Pflichtfeld.tsx` | 1 | text | **geprüft** | **Das Bauteil, das die Pflicht trägt** — es meldet, wer ein Feld anfasst und leer verlässt, nennt das Feld in seinem Satz und reicht einen `ref` durch. **Erst nach der Berührung**: ein leeres Formular ist kein Fehler, sondern ein leeres Formular |
| `bausteine/Kennwortfeld.tsx` | 1 | password/text | **geprüft** | **Das Bauteil, das das Kennwort trägt** — Werkstück 6 von Phase 4.12 (#1351). Es setzt `aria-invalid` und nennt den Befund am Feld; **was ein taugliches Kennwort ist, entscheidet `internal/kennwort`** und meldet es als Feldfehler zurück. Die Mindestlänge steht als Hinweis am Feld, **bevor** jemand drückt. **Der Typ wechselt zwischen `password` und `text`**: wer sein Kennwort nicht sehen kann, vertippt sich und erfährt es erst bei der Anmeldung |
| `bausteine/Nachweisfeld.tsx` | 1 | text | **geprüft** | **Der zweite Faktor** — Werkstück 7 von Phase 4.12 (#1351). **Die Prüfung steht im Dienst und nicht am Feld**, und das ist eine Entscheidung: sechs Ziffern aus der App und zwölf Zeichen vom Zettel sind für den, der sie eingibt, dasselbe — ein Feld, das die Form vorschriebe, wiese den Rückweg ab. Der Knopf ruht, solange nichts dasteht; der Befund kommt von `service.Zweiterfaktor` und erscheint über dem Formular, weil kein Feld daran schuld ist |
| `bausteine/Sammlungsformular.tsx` | 1 | text | **geprüft** | Der **Erwerbspreis** — das eine Feld dieses Formulars, das sein `input` selbst hält, weil es einen Befund am Feld braucht (§6). **Seit UX-06 stehen drei weitere in eigenen Bausteinen** (#1678): die Menge in `Zahlenfeld` (ganz, ab eins — wie `Exemplar.Validate`), die Sprache in `Vorschlagsfeld`, die Währung in `Waehrungsfeld` (klein geschrieben erlaubt, `Normalisieren` hebt es an). Was die Sprache des Fensters selbst so schriebe, wird kanonisch (`lokal.ts`); was **zwei Werte** haben kann — »1.234« Dinar im Deutschen —, wird am Feld erklärt und geht nicht hinüber; was keine Lesart hat, geht unverändert, und `domain.GeldAus` sagt, was nicht stimmt (LOC-03, #1689) |
| `bausteine/Sammlungsinspector.tsx` | 1 | text | **geprüft** | Der **Erwerbspreis beim Bearbeiten** — dieselbe Lesart wie beim Anlegen (`preiseingabe.ts`): im Feld steht der Preis in der Schreibweise des Fensters, hinüber geht die kanonische Form, **zwei Werte** werden am Feld erklärt. Bis zum Zweitblick über Phase 6.3 ging der Text roh an Go, und »1.234« Dinar wurde still zu 1,234 (S6) |
| `bausteine/Sammlungsausfuhr.tsx` | 1 | textarea | **ungeprüft** | Die **Ausgabe** einer Ausfuhr, nicht die Eingabe |
| `bausteine/Sammlungssicht.tsx` | 1 | search | **ungeprüft** | Ein Suchfeld nimmt jede Zeichenfolge |
| `bausteine/Scryfallabgleich.tsx` | 1 | text | **ungeprüft** | Ein Kartenname für einen fremden Dienst. **Die Regel steht dort** — was Scryfall findet, weiß Scryfall (§16) |
| `bausteine/Scryfallbild.tsx` | 1 | text | **ungeprüft** | dito |
| `bausteine/Scryfallkarte.tsx` | 2 | text | **ungeprüft** | dito |
| `bausteine/KIBild.tsx` | 1 | textarea | **ungeprüft** | Die Beschreibung des gewünschten Bildes (A12, #1451) — Fließtext, und jede Zeichenfolge ist gültig. Eine leere Beschreibung sperrt den Knopf, statt einen Fehler zu zeigen. **Die Länge prüft die Domäne** ([domain.KIBildauftrag]): zweitausend Zeichen, und darüber kommt eine Meldung statt eines Auftrags, der bezahlt wird. Alles andere in diesem Kasten ist eine **Wahl** — Modell, Feld, Format, Güte —, und eine Wahl kennt nur, was angeboten wird |
| `bausteine/KIBatchkasten.tsx` | 1 | textarea | **ungeprüft** | Der Auftrag an ein Sprachmodell für mehrere Karten (A13.1b, #1452) — Fließtext wie beim einzelnen Vorschlag daneben, und jede Zeichenfolge ist gültig. Ein leerer Auftrag sperrt den Knopf, statt einen Fehler zu zeigen; dasselbe tut eine leere Auswahl. **Die Auswahl selbst ist keine freie Eingabe** — es sind Ankreuzfelder, und die kennen nur, was das Set hat |
| `bausteine/KIEinschaetzung.tsx` | 2 | text, textarea | **ungeprüft** | Die Frage an ein Sprachmodell (A13.2a, #1452) — Fließtext, und jede Zeichenfolge ist gültig. **Sie geht auch an den Suchindex**, und auch dort ist sie freier Text: eine Volltextfrage prüft nichts, sie findet weniger. Eine leere Frage sperrt den Knopf, statt einen Fehler zu zeigen. **Die Art wird gewählt statt getippt** — zwei Werte, und eine Wahl kennt nur, was es gibt. **Dazu das Format** (A13.2b): es benennt, worauf sich ein Vergleich bezieht — „Commander", „Pauper", der Name einer Hausrunde. Eine Werteliste gäbe es dafür nicht, denn dieses Programm kennt kein Turnierformat; leer ist es erlaubt und steht dann **als Einschränkung in der Antwort**, nicht als Fehler am Feld
| `bausteine/KIVorschlag.tsx` | 2 | text, textarea | **ungeprüft** | Der Auftrag an ein Sprachmodell (A7, #1446) — Fließtext, und jede Zeichenfolge ist gültig. Ein leerer Auftrag sperrt den Knopf, statt einen Fehler zu zeigen. **Seit A8 (#1447) ist es nur noch eines**: Anbieter und Modell werden aus dem Katalog **gewählt** statt getippt, und eine Wahl kennt nur, was bereitsteht | **Und seit A13.4 (#1452) sind es zwei**: dazu die **Belegfrage**, die an den Suchindex geht — ungeprüft aus demselben Grund wie die Frage der Einschätzung darüber, und **leer erlaubt**: dann wird gar nicht gesucht, und der Vorschlag bleibt der gewöhnliche ohne Belege
| `seiten/Leistungsprofil.tsx` | 1 | number | **geprüft** | Die **Dauer** eines Leistungsprofils, eine ganze Zahl von 1 bis 60 (OPS-05, #1722): außerhalb steht `aria-invalid` am Feld, die Grenze als Beschreibung darunter, und der Knopf hält an. Go prüft dieselbe Grenze noch einmal |
| `bausteine/Einstellungssuche.tsx` | 1 | search | **ungeprüft** | die Suche über alle Einstellungen (#1422) — jede Zeichenfolge ist gültig, und kein Treffer ist ein leerer Zustand, kein Fehler |
| `bausteine/Partienauswahlleiste.tsx` | 1 | search | **ungeprüft** | die Suche über die eigenen Partien (PLAY-03.AK-05, #1711) — jede Zeichenfolge ist gültig, und kein Treffer ist ein **Leerzustand** mit eigenem Satz und einem Knopf, der die Auswahl zurücknimmt. **Der Stand und die Ordnung daneben sind Wahlfelder** und stehen deshalb nicht hier |
| `bausteine/Suchauswahl.tsx` | 1 | text | **ungeprüft** | die Eingabe einer Auswahl mit Suche (UXR-F1, #2007) — Stil (UXR-07 und UXR-04), Testkarte, Spiel beim Anlegen eines Pakets und Zeitzone. Was getippt wird, **filtert** nur und ist nie die Wahl: erst eine Zeile ist es. Jede Zeichenfolge ist gültig, passt nichts, sagt ein Satz es, und Escape stellt den Wert her. Was gewählt werden kann, bestimmt die Liste — bei der Zeitzone hält `internal/werkzeuge/zeitzonenliste_test.go`, dass Go jede Zeile annimmt |
| `seiten/Hilfesuche.tsx` | 1 | search | **ungeprüft** | die Suche in der Hilfe (UXR-22, #1930) — jede Zeichenfolge ist gültig, und kein Treffer ist ein leerer Zustand mit dem Weg zurück, kein Fehler |
| `bausteine/Textsuche.tsx` | 2 | search, text | **ungeprüft** | Suchen und Ersetzen — jede Zeichenfolge ist gültig |
| `bausteine/Turnieranlage.tsx` | 1 | text | **geprüft** | Die Notiz ist frei; **das Spiel** steht im `Vorschlagsfeld` mit den installierten Spielen als Vorschlag — ein eingelesenes Turnier darf eines nennen, das hier fehlt (UX-06.AK-01); **der Name ist Pflicht** und liegt in `Pflichtfeld` |
| `bausteine/Turnierpaarung.tsx` | 1 | text | **geprüft** | Ein Ergebnis, seit dem 2026-09-08 gegen `domain.Ergebnis.Pruefe` gelesen: **ganze Zahlen ab null**, die **Summe größer als null** (wer nicht gespielt hat, hat noch keines), und die Summe höchstens so groß, wie ein Match hier ist — `service.hoechstensSpiele` gibt bei einem Einzelspielturnier eins und sonst **null**, und null heißt keine Grenze. Die drei Kästchen halten seither die **rohe** Zeichenkette: `Number.isFinite(…) ? … : 0` machte aus jedem Vertipper eine Null, und drei Nullen sind ein Stand, den Go abweist. Der Befund über die Summe hängt an allen drei — welches zu viel hat, sagt die Regel nicht. Zurücknehmen bleibt frei: es schickt `null` und liest die Kästchen nicht |
| `bausteine/Zugaenge.tsx` | 1 | password | **ungeprüft** | Ein Zugangsschlüssel eines fremden Dienstes. **Ob er gilt, weiß nur der Dienst** — eine Formprüfung hier wiese einen gültigen Schlüssel ab, sobald der Anbieter sein Format ändert |
| `bausteine/Anbietergeheimnisse.tsx` | 1 | password | **ungeprüft** | Das `client_secret` einer OAuth-Anwendung. **Dieselbe Begründung wie eine Zeile darüber**: ob es gilt, weiß nur der Anbieter, und eine Formprüfung wiese ein gültiges Geheimnis ab, sobald er sein Format ändert |
| `bausteine/Befehlsfeld.tsx` | 1 | text, number | **ungeprüft** | Die Felder eines Spielbefehls (#1623). **Und das ist eine Entscheidung und kein Versäumnis**: was ein Befehl annimmt, entscheidet der Spielkern, und er sagt es in derselben Sprache wie über jede andere Fläche — „die Kennung %q ist schon vergeben", „den Sitz %q gibt es nicht". Eine zweite Prüfung im Fenster wäre dieselbe Regel ein zweites Mal (§2) und liefe beim ersten neuen Befehl auseinander. **Was das Fenster prüft, ist die Gestalt**: eine Zahl steht in einem `number`, eine Zone und ein Sitz in einer Auswahlliste — dort ist ein Vertipper gar nicht möglich. **Der Befund steht am Kopf des Tisches**, wo die Meldung des letzten Befehls steht |
| `bausteine/Spielsteinfeld.tsx` | 1 | text | **ungeprüft** | Name, Typzeile, Stärke und Widerstandskraft eines Spielsteins — **vier Eingaben aus einem Feld im Quelltext**, weil sie über dieselbe Zeile laufen — CR 701.7a lässt den Effekt sie nennen, und **jede Zeichenfolge ist eine gültige Eigenschaft**: ein Spielstein darf `*/*` tragen (CR 208.2). Ein leeres Feld fällt heraus statt als leere Eigenschaft mitzugehen; dass ein Spielstein **ohne** Eigenschaften keiner ist, weist die Domäne ab |
| `rahmen/Palette.tsx` | 1 | text | **ungeprüft** | Die Befehlspalette sucht; jede Eingabe ist gültig |
| `rahmen/Befehlsformular.tsx` | 2 | text, textarea | **geprüft** | die Felder eines Befehls aus seinem Schema (#1423): Pflicht, Zahl und JSON prüft `argumenteAus` vor dem Ruf, der Befund steht am Feld; alles andere sind Daten für das Werkzeug, das in Go am Schema prüft |
| `seiten/Autoersetzungen.tsx` | 2 | text | **geprüft** | **Das Muster ist Pflicht** und liegt seit dem 2026-09-08 in [`Pflichtfeld`](../frontend/src/bausteine/Pflichtfeld.tsx) — `domain.PruefeAutoersetzung` weist ein leeres ab (eine Regel ohne Muster greift nie). Bis dahin sagte das nur der ausgegraute Knopf, und wer ihn sieht, weiß nicht, **welches** der beiden Felder fehlt. **Die Anordnung war der offene Punkt** und ist entschieden: die Zeile steht auf `items-end`, der Satz wächst nach unten statt die Nachbarn zu heben. **Der Ersatz bleibt frei** — `PruefeAutoersetzung` prüft ihn nicht, und eine Regel, die einen Text entfernt, ist gültig; ebenso der Ersatz in der Regelzeile |
| `bausteine/KIKostenfreigabe.tsx` | 1 | text | **geprüft** | Der Höchstbetrag einer KI-Freigabe (#1539). **Dieselbe Prüfung wie bei `Kosten.tsx`** und aus demselben Grund: **nur Ziffern** in der kleinsten Einheit ([`ziffernbetrag.ts`](../frontend/src/ziffernbetrag.ts)), hier **größer als null** — eine Freigabe über nichts ist keine, und der Knopf bleibt sonst aus. **»1.000« war bis zum Nachlauf über Phase 6.3 ein Cent** (N5); jetzt sagt das Feld, dass es nur Ziffern nimmt. **Die Währung steht nicht als Feld da**: sie kommt aus dem Budget des Kontos, und wer sie tippen dürfte, bekäme beim ersten Vertipper die Absage „die Grenze gilt in EUR“ und suchte sie am falschen Ort |
| `seiten/Kosten.tsx` | 2 | text | **geprüft** | Konto und Betrag (O12, #1439). **Der Betrag ist im Fenster geprüft**: er ist eine **Folge von Ziffern** ([`ziffernbetrag.ts`](../frontend/src/ziffernbetrag.ts)) — der Knopf bleibt sonst aus, und das Feld sagt warum, denn Beträge stehen in der kleinsten Einheit, und „12,50" wäre dort keine Zahl, sondern ein Missverständnis. **„1.000" las `Number` bis zum Nachlauf über Phase 6.3 still als eins** (N5). **Die Währung steht seit UX-06 in [`Waehrungsfeld`](../frontend/src/bausteine/Waehrungsfeld.tsx)** (#1678): Vorschläge aus `Intl`, die **Gestalt** eines Codes am Feld — klein geschrieben wird hier erklärt, weil `Betrag.Pruefe` es abweist —, und ohne Gestalt bleibt der Knopf aus. **Das Konto prüft die Go-Seite**: welches es gibt, weiß der Bestand |
| `bausteine/Vorschlagsfeld.tsx` | 1 | text | **geprüft** | **Ein Freitext mit Vorschlägen** (UX-06.AK-01, #1678) — `datalist` unter einem `input`. Es trägt den Befund seines Aufrufers am Feld; **die Regel steht beim Aufrufer**: `Waehrungsfeld` prüft die Gestalt eines Codes wie `domain.istWaehrungscode`, die Sprache eines Drucks im Sammlungsformular bleibt Freitext — Scryfalls zwanzig Codes sind Vorschläge und keine Grenze, denn der Bestand nimmt jeden an |
| `seiten/Ruecksetzung.tsx` | 1 | text | **geprüft** | Der **Name** zur Bestätigung (O4, #1428). **Die Go-Seite prüft ihn und nicht das Fenster, und das ist die Entscheidung**: er wird gegen den **gespeicherten** gehalten — was hier zu prüfen wäre, wüsste das Fenster nur aus einer zweiten Kopie —, und eine Formprüfung hier wiese einen richtigen Namen ab, sobald jemand sein Profil umbenennt. **Der Befund steht am Feld**: die Absage kommt als Fehler zurück und erscheint über dem Knopf, der sie ausgelöst hat. **Das zweite Feld, der Nachweis, ist fort** (G12, #2282): er war eine Kennung, die niemand sah; jetzt holt ein Knopf die Freigabe |
| `seiten/Bildereinstellungen.tsx` | 1 | text | **ungeprüft** | Die Namensvorlage nimmt jede Zeichenfolge; `{nummer}` und `{name}` werden eingesetzt, alles andere bleibt stehen. **Die Breite steht seit dem 2026-09-03 nicht mehr hier**, sondern in [`Zahlenfeld`](../frontend/src/bausteine/Zahlenfeld.tsx) — 1 bis 10 000 Bildpunkte, im laufenden Fenster nachgesehen |
| `seiten/Ersetzen.tsx` | 2 | search, text | **ungeprüft** | Suchen und Ersetzen über ein Set |
| `seiten/Feldeingabe.tsx` | 3 | text, textarea | **geprüft** | Ein Kartenfeld. **Was gültig ist, sagt das Vorlagenpaket** — geprüft wird, was es erklärt, und nicht mehr |
| `seiten/Kartenmarken.tsx` | 1 | text | **ungeprüft** | Eine Marke ist ein freies Wort |
| `bausteine/Objektmarken.tsx` | 1 | text | **geprüft** | Eine Marke an jedem Fachobjekt (UX-12, #1684): **Go prüft** — leer und über 40 Zeichen weist `domain.PruefeMarke` ab, die Meldung steht am Feld |
| `bausteine/Markenfilter.tsx` | 1 | text | **geprüft** | Der Name eines gespeicherten Filters: leer hält den Knopf an, und `domain.PruefeGespeicherteFilter` weist doppelte Namen einer Sicht ab |
| `bausteine/Markenpflege.tsx` | 2 | text | **geprüft** | Die Marke für die gewählten und der neue Name einer Marke — **Go prüft** beide wie jede Marke; eine Kollision ist ein Konflikt am Feld |
| `seiten/Kartennotizen.tsx` | 1 | textarea | **ungeprüft** | Eine Notiz ist freier Text |
| `seiten/Listenspalten.tsx` | 1 | search | **ungeprüft** | Sucht in der Spaltenliste |
| `seiten/Massenfelder.tsx` | 1 | text | **geprüft** | Der neue Wert — gegen dieselbe Regel wie das Einzelfeld — seit U6 (#1417) in `Massenfelder.tsx`, die Vorschau daneben zeigt vor dem Schreiben, was der Wert je Karte hieße |
| `seiten/Namensliste.tsx` | 1 | search | **ungeprüft** | Sucht in den Skriptnamen |
| `seiten/Paketbezug.tsx` | 1 | text | **ungeprüft** | Adresse, Stand, Unterverzeichnis und Name eines Bezugs — Kennungen, eine Eingabe im Quelltext für vier Felder. **Seit UXR-17 in der Schublade *Paket beziehen*** (#1925); bis dahin stand hier `seiten/Pakete.tsx` mit „Sucht in den installierten Paketen“, und das war nie wahr. Ob Adresse und Stand dastehen, sagt der Knopf *Herkunft anzeigen …* unter dem Formular; ob es ein Stand ist, `internal/bezug` |
| `seiten/Paketfilter.tsx` | 1 | search | **ungeprüft** | Sucht in den installierten Paketen |
| `seiten/Schluesselwortfeld.tsx` | 1 | text | **geprüft** | Vier Felder, **eine** Regel: ob ein Muster greift, entscheidet `service.musterGreift` — portiert aus `Keyword::prepare` (`src/data/keyword.cpp`, Zeilen 205–209, Fassung f9d9356 vom 2021-07-06; `valid = !match_re.matches(_(""))`). Der Satz stand bis zum 2026-09-08 als namenloser `<span>` **unter** der Zeile; seither hängt er über [`Feld`](../frontend/src/bausteine/Feld.tsx) am **Muster**, mit `aria-invalid` und `aria-describedby`. Name, Erinnerungstext und Anmerkung bleiben frei — `Schluesselwortpflege.Setzen` weist an ihnen nichts ab |
| `seiten/ScryfallSatz.tsx` | 1 | text | **ungeprüft** | Eine Scryfall-Abfrage — die Regel steht bei Scryfall |
| `seiten/Scryfallsuche.tsx` | 1 | text | **ungeprüft** | dito |
| `seiten/Schnellfilter.tsx` | 1 | search | **ungeprüft** | Der Schnellfilter der Kartenliste — Werkstück U5 von Phase 5.1 (#1416). Die Sprache aus MSE2 (`src/data/filter.cpp`) kennt keinen ungültigen Ausdruck: eine offene Anführung reicht bis zum Ende, ein Feld ohne Wert fällt weg, und was nichts trifft, zeigt eine leere Liste mit Trefferzahl |
| `seiten/Vorlagenbaum.tsx` | 1 | search | **ungeprüft** | Die Suche über Pakete und Dateien der Vorlagenwerkstatt — Werkstück U7 von Phase 5.1 (#1418). Ein Suchfeld nimmt jede Zeichenfolge; was nichts trifft, zeigt „Kein Paket und keine Datei passt zur Suche.“, und `Escape` leert es |
| `seiten/Setloeschen.tsx` | 1 | text | **geprüft** | Der Name muss **wörtlich** stimmen — die Rückfrage vor dem Löschen |
| `seiten/Stiloption.tsx` | 1 | text | **ungeprüft** | Eine Stiloption. **Was gültig ist, sagt das Stilpaket** (§11) |
| `seiten/Suche.tsx` | 1 | search | **ungeprüft** | Die Volltextsuche über alle Sets |
| `seiten/Wissensrecherche.tsx` | 1 | search | **ungeprüft** | Die Frage an den Wissensbestand. **Was gültig ist, entscheidet Go**: `domain.Rechercheanfrage.Pruefe` weist eine leere Frage ab und begrenzt sie auf 2.000 Zeichen, und die Meldung sagt, welches von beidem. Dieselbe Lage wie bei `Suchindex.tsx` — eine Regel hier wäre ihre zweite Fassung |
| `seiten/Kontextfreigabe.tsx` | 2 | text | **ungeprüft** | Aufgabe und Modell; **die beiden Budgetzahlen liegen seit F03 in `Zahlenfeld`** (ganzzahlig ab null — ein leeres Feld wurde vorher null, „8192.5“ ging an ein `int`); **der Anbieter wird seit UX-06.AK-01 gewählt** (vier, dieselben wie `internal/app/kiwerkzeug.go`). **Darüber hinaus tragen sie keine Grenze**: was ein zulässiges Budget ist, weiß nur das Zielmodell, und `service.pruefeKontextwunsch` sagt es als Satz — „das Kontextfenster ist 0 und die Ausgabereserve 0, dann bleibt für den Kontext nichts". Eine Grenze hier müsste je Modell eine andere sein und wäre am Tag darauf falsch |
| `seiten/Suchindex.tsx` | 1 | text | **ungeprüft** | Die Suchfrage. **Was gültig ist, entscheidet Go**: `service.pruefeSuchwunsch` weist eine leere Frage und eine über 2.000 Zeichen ab, und die Meldung sagt, welche von beidem. Eine Längengrenze hier wäre eine zweite Fassung derselben Regel; eine Pflichtmarke wäre falsch, weil die leere Frage einen **Satz** bekommt und keinen roten Rahmen |
| `seiten/Suchindexaufbau.tsx` | 1 | text | **ungeprüft** | **Ein Feld im Quelltext, drei auf dem Bildschirm**: Modell, Sammeldatei und Quellrevision teilen sich ein Bauteil, und der Wächter zählt `<input>`-Stellen. **Alle drei sind Namen aus einer fremden Welt**: welche Modelle installiert sind, weiß die Laufzeit; ob die Datei da ist, weiß das Dateisystem; wie eine Scryfall-Revision aussieht, sagt Scryfall. Eine hier erfundene Regel wiese einen gültigen Namen ab (§11). Was falsch ist, sagt die Antwort — „das Modell taugt hier nicht als Einbetter" oder „die Sammeldatei ist kein gzip" (§8) |
| `seiten/Symbolbibliothek.tsx` | 2 | text, search | **geprüft** | Der Name eines aufgenommenen Symbols und die Suche in der Bibliothek (DOM-02, #1697). **Den Namen prüft Go**: `domain.PruefeSymbolname` weist leer, über 80 Zeichen und einen Zeilenumbruch ab, und ein Name, den es ohne Rücksicht auf Groß und Klein schon gibt, kommt als `doppelterName` **am Feld** an. **Die Suche ist ungeprüft**: sie schränkt ein und weist nichts ab |
| `seiten/Symbolbibliothekseintrag.tsx` | 1 | text | **geprüft** | Der neue Name beim Umbenennen — dieselbe Regel wie beim Aufnehmen, und der Befund steht an diesem Feld (DOM-02, #1697) |
| `seiten/Symmetriefeld.tsx` | 2 | number | **ungeprüft** | **Die Kopien liegen seit dem 2026-09-03 in `Zahlenfeld`** — ab 2 und nach oben begrenzt. **Mitte und Griff haben keine Grenze, und das ist nachgelesen**: `app.symmetrieAendern` setzt beide ohne jeden Vergleich, und MSE2 tut dasselbe — `SymmetryMoveAction::move` (`src/data/action/symbol_part.cpp`, Zeilen 461–476, Fassung f9d9356 vom 2021-07-06) rastet auf ein Gitter und auf Vielfache von 2π/24, **begrenzt aber nichts**. Eine Mitte außerhalb des Einheitsquadrats ist ein zulässiger Drehpunkt; ein `min`/`max` hier wäre eine erfundene Regel (§11). **Ungeprüft heißt nicht ungefiltert**: `Number("")` ist `0`, und wer das Feld leert, um eine Zahl zu tippen, schob damit die Mitte auf null — seit dem 2026-09-08 löst nur eine endliche Zahl einen Ruf aus |
| `seiten/Vorlagenwahl.tsx` | 1 | search | **ungeprüft** | Das Feld, das die **Exportvorlagen eingrenzt** — ab fünfzehn, über Namen und Paketkennung (UXR-09, #1917; R7 aus `docs/UIREVIEW.md`, gezählt: achtzehn für Magic in `full-magic-pack`). Es grenzt nur ein, was die Auswahl daneben anbietet; die gewählte Vorlage bleibt angeboten, und kein Treffer sagt „Nichts passt zu …“ |
| `seiten/Wahlfeld.tsx` | 1 | text | **ungeprüft** | Eine Wahl aus dem Paket |
| `seiten/Zufallspack.tsx` | 1 | text | **geprüft** | Der Same einer Ziehung. **Hier stand, jede Zeichenfolge sei gültig, und das war falsch**: `app.samenAus` liest ihn mit `strconv.ParseUint(text, 10, 64)` — Ziffern, höchstens 2⁶⁴−1, und ein **leerer** Text heißt null und ist gültig. Seit dem 2026-09-08 sagt das Feld es selbst, und der Feldfehler aus Go (`NewFieldError("same", …)`) steht über `verursachendesFeld` an ihm statt als anonymer Satz über der Lasche. Der Knopf für einen neuen Samen zählt über `BigInt` hoch — `Number(same) + 1` verlor oberhalb von 2⁵³ genau die Stellen, deretwegen der Same als Text reist. **Eine Zahl der Packs gibt es nicht**: `PackZiehen(setID, art, same)` zieht ein Pack, und die Zeile, die eine Zahl ab eins nannte, meinte ein Feld, das dieses Programm nicht hat |
| `seiten/Zuschnittfenster.tsx` | 1 | number | **ungeprüft** | Der Zuschnitt in Bildpunkten. **Die Regel steht in Go und weist nicht ab, sondern berichtigt**: `render.Zuschnittwahl.Begrenze` (portiert aus `ImageSlice::constrain`) schiebt jede Fläche in die Quelle zurück, und der Rahmen zeigt danach, was gilt — jede Eingabe geht durch `Zuschnittgrenze`. Eine zweite Fassung im Fenster wäre eine zweite Antwort auf dieselbe Frage (§15). **Ungefiltert ist es trotzdem nicht**: seit dem 2026-09-08 löst ein leeres Feld keinen Zug mehr aus — `Number("")` war `0`, und der Rahmen sprang, während jemand eine neue Zahl tippte |

---

## Was am 2026-09-03 geschah

**`bausteine/Zahlenfeld.tsx` ist gebaut und an zwei Stellen eingesetzt** —
Bildbreite und Anschnitt. Es meldet nur, was gilt: ein Wert außerhalb des
Bereichs erreicht den Aufrufer nicht, und ein leeres Feld ist kein Fehler.
Nachgesehen im laufenden Fenster.

**Vier weitere Zahlenfelder tragen nur `min` und `max`** — zwei in der
Symmetrie, der Zuschnitt und „Behalten". Sie standen bis heute als **geprüft**
in dieser Tabelle, und das war falsch: `min`/`max` ohne Formular sind eine
Zusage, die niemand einlöst. Die Marken sind berichtigt; die Kopienzahl der
Symmetrie ist umgestellt, die vier übrigen sind der nächste Schritt.

**Die Klasse „Pflichtfeld leer" ist zu** — `bausteine/Pflichtfeld.tsx` trägt
sie, und **fünf** Formulare benutzen es: Deck, Format, Turnier, Teilnehmer
und Set. Es meldet, wer ein Feld anfasst und **leer verlässt** — nicht
schon beim Aufschlagen: ein leeres Formular ist kein Fehler, sondern ein
leeres Formular, und ein Hinweis, der immer dasteht, stumpft ab.

**Eine sechste blieb offen**: das Muster einer Autoersetzungsregel
(`seiten/Autoersetzungen.tsx`). Es liegt in einer Zeile mit dem Ersatz und
dem Anlegeknopf, und ein Hinweis darunter verschöbe die ganze Zeile —
dort war erst die Anordnung zu entscheiden. **Sie ist es seit dem
2026-09-08** (unten).

---

## Was am 2026-09-08 geschah

**Werkstück U3 von Phase 5.1**
(#1414):
die acht Zeilen mit der Marke **offen** sind entschieden. Es gibt keine mehr.

**Fünf sind geprüft, und jede Regel ist nachgelesen** — in
`internal/domain`, `internal/service` oder `internal/app`, nicht aus dem
Feldnamen erschlossen:

| Datei | Woher die Regel kommt |
| --- | --- |
| `bausteine/Formatzeile.tsx` | `domain.Format.zahlenPruefen` und `Format.Validate` |
| `bausteine/Turnierpaarung.tsx` | `domain.Ergebnis.Pruefe` und `service.hoechstensSpiele` |
| `seiten/Autoersetzungen.tsx` | `domain.PruefeAutoersetzung` |
| `seiten/Schluesselwortfeld.tsx` | `service.musterGreift`, portiert aus `Keyword::prepare` |
| `seiten/Zufallspack.tsx` | `app.samenAus` — `strconv.ParseUint(text, 10, 64)` |

**Zwei sind ungeprüft, und der Grund steht dabei**: die Mitte und der Griff
einer Symmetrie haben in `app.symmetrieAendern` **und in MSE2** keine Grenze,
und der Zuschnitt wird in Go **berichtigt** statt abgewiesen. Beides ist eine
Antwort und keine Lücke — eine hier erfundene Grenze wiese einen gültigen Wert
ab (§11).

**Eine ist umgezogen**: `bausteine/Sammlungsordnung.tsx` benutzt
`Zahlenfeld` und hat kein eigenes `<input>` mehr. Sie steht deshalb nicht mehr
in der Tabelle — wie `Bogeneinstellungen.tsx` am 2026-09-03.

### Was der Anhang falsch sagte, und wie es auffiel

**„Die Zahl der Packs ist eine Zahl ab eins."** Es gibt keine Zahl der Packs.
`App.PackZiehen(setID, art, same)` zieht **ein** Pack; ein Feld dafür hat
dieses Programm nicht. Das freie Feld auf der Packlasche ist der **Same**, und
über den sagte diese Tabelle „jede Zeichenfolge ist gültig" — auch das falsch:
`samenAus` liest ihn mit `ParseUint(…, 10, 64)` und weist alles andere als
Feldfehler ab.

**Beides fiel nur auf, weil die Quelle gelesen wurde.** Aus dem Feldnamen
„Same" folgt keine dieser beiden Auskünfte, und aus einer Zeile im Anhang
folgt sie erst recht nicht (§15: keine Fachlogik aus Feldnamen oder
Vermutungen ableiten).

### Und drei Eingaben waren aufgeräumt, bevor jemand sie sah

Drei Stellen machten aus einer unbrauchbaren Eingabe stillschweigend eine
**Null** — und null bedeutet an allen dreien etwas:

| Stelle | Was `0` dort heißt |
| --- | --- |
| `Formatzeile` (`mitZahl`) | keine Grenze, kein Sideboard, beliebig oft |
| `Turnierpaarung` | ein Ergebnis ohne ein einziges Spiel — von Go abgewiesen |
| `Symmetriefeld`, `Zuschnittfenster` | die Mitte bzw. der Ausschnitt springt an den Rand |

Der Fehler entstand jedes Mal **im Fenster** und wurde erst am anderen Ende
sichtbar. Die Tests dazu tippen deshalb die **rohe** Zeichenkette (§3).

---

## Geschlossen oder offen — die Durchsicht aller Eingaben

**UX-06.AK-01** (#1678):
*geschlossene Wertemengen werden gewählt, fachlich offene behalten Freitext
mit Vorschlägen, und keine Umstellung schließt zulässige Kartendaten aus.*
Durchgesehen am 2026-09-25, **jede Zeile der Tabelle oben**, nach der einen
Frage: gibt es eine **feste Menge** gültiger Werte, die dieses Programm
kennt?

| Gruppe | Zeilen der Tabelle | Wertemenge | Wie sie dasteht |
| --- | --- | --- | --- |
| **Namen, Notizen, Freitext** | Kontaktnotiz, Deckanlage, Turnieranlage (Name, Notiz), Formatverwaltung (Name), Paketanlage, Vorlagenumschlag, Tischgespräch, Kartennotizen, Kartenmarken, KI-Aufträge, Spielsteinfeld | offen | Freitext — jede Zeichenfolge ist ein gültiger Name |
| **Suchen** | Sammlungssicht, Einstellungssuche, Partienauswahl, Textsuche, Palette, Ersetzen, Listenspalten, Namensliste, Paketfilter, Schnellfilter, Vorlagenbaum, Suche, Wissensrecherche, Suchindex | offen | Freitext — eine Suche findet weniger, sie irrt nicht |
| **Kennungen und Geheimnisse** | Paketbezug, Kontaktaufnahme, Abdruckvergleich, Einladungsannahme, Handelsvorgang, Handelszusagen, Handelsausgang, Handelsbuchung, Deckuebernahme, Zugaenge, Anbietergeheimnisse, Kennwortfeld, Nachweisfeld | offen, aber mit Gestalt | Freitext, die Gestalt prüft der Dienst (`PruefeAdresse`, Kennung, Nachweis) |
| **Adressen und Ports** | Kontaktadresse, Horchposten, Dienstposten | offen, mit Gestalt | Freitext, geprüft in `domain.PruefeAdresse` und `PruefeBindung` (UX-06.AK-02) |
| **Zahlen** | Zahlenfeld, Zuordnungswahl, Formatzeile, Symmetriefeld, Zuschnittfenster, Turnierpaarung, Kosten, KIKostenfreigabe, Zufallspack | offen, typisiert | `type="number"` oder `Zahlenfeld`, Grenzen wo es sie gibt |
| **Was ein Paket erklärt** | Feldeingabe, Stiloption, Wahlfeld, Massenfelder, Schluesselwortfeld, Quelltextfeld | **vom Paket** bestimmt | was das Paket als Wahl erklärt, wird gewählt (`Wahlfeld`); alles andere bleibt frei — **eine Auswahl hier schlösse zulässige Kartendaten aus**, und genau das verbietet AK-01 |
| **Fremde Dienste** | Scryfallabgleich, Scryfallbild, Scryfallkarte, ScryfallSatz, Scryfallsuche | offen | Freitext — was Scryfall findet, weiß Scryfall (§16) |
| **Geschlossen, und jetzt gewählt** | Handelspositionen (Güte, Ausführung), Kontextfreigabe (Anbieter) | **geschlossen** | Auswahl — **beide standen bis zum 2026-09-25 als Freitext da**, und die Güte ergab mit ihrer Vorgabe `NM` eine Position, die sich nicht verbuchen ließ |
| **Geschlossen, und schon gewählt** | Sammlungsformular (Zustand, Ausführung, Währung, Sprache), SetNeu (Spiel als Optionsfelder seit UX-05 — #1677 —, der Stil als Auswahl mit Suche seit UXR-04 — #1912), Deckkartenwahl, KIVorschlag (Anbieter, Modell) | geschlossen | Auswahl oder Vorschlagsfeld seit UX-06 (#1827) und davor |
| **Offen, mit Vorschlägen** | Sprache (Sammlung, Handel), Währung, Vorschlagsfeld, **Spiel in Format und Turnier** | offen, bekannte Werte | `datalist` — Scryfalls zwanzig Sprachcodes, `Intl` für Währungen, **die installierten Spiele** (`useSpielnamen`, seit dem 2026-09-25); jeder andere Wert bleibt möglich |

**Was offen bleibt, und warum es offen bleibt**: das **Modell** in der
Kontextfreigabe und im Aufbau des Suchindex. Die Namen, die eine Laufzeit
kennt, ändern sich schneller als diese Anwendung; eine Auswahl aus dem
heutigen Katalog wiese morgen ein gültiges Modell ab. Die Wahl mit Katalog
gibt es dort, wo die Anwendung den Katalog schon fragt (`KIModellwahl`).

## Was diese Tabelle nicht sagt

- **Ob die Prüfung stimmt.** Sie hält Zahlen und Marken gegen den Quelltext.
  Eine Datei, die `aria-invalid` setzt und die falsche Regel prüft, käme hier
  durch — dafür stehen die Tests neben den Bauteilen.
- **Ob der Satz am Feld gut ist.** §6 verlangt, dass er sagt, **was falsch ist
  und was zu tun ist**. Das prüft ein Mensch.
- **Die Felder eines Formulars, die kein `<input>` sind.** Auswahlfelder und
  Ankreuzflächen stehen hier nicht: sie können nur gültige Werte annehmen.
