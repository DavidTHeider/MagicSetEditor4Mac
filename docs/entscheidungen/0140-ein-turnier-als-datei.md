# 0140 — Ein Turnier als Datei und als Bericht

**Stand:** entschieden am 2026-09-27 · DOM-04.AK-03
(#1699)

## Die Frage

AK-03 verlangt, dass sich ein Turnier **ausgeben, einlesen, nach einer
Vorlage kopieren** lässt und dass es einen **Abschlussbericht** gibt — und
dass eine Kopie *„keine fremden Spielerzustimmungen oder alten
Transaktionskennungen"* übernimmt. Die Vorlage steht in Vermerk 0138. Offen
waren die Gestalt der Datei, was an ihr **nicht** reist, und wie ein Bericht
entsteht, der in zwei Sprachen gelesen wird.

Portierungsmarke **new** — MSE2 kennt keine Turniere (Vermerk 0138).

## Die Entscheidung

### Die Datei ist ein Umschlag mit Fassung

Dieselbe Gestalt wie die Partiedatei (PLAY-03.AK-05): `{"fassung": 1,
"turnier": {…}}`, eingerückt, mit Zeilenende, **höchstens vier Mebibyte und
zwölf Stufen tief**. Gemessen am 2026-09-27: ein Turnier mit 64
Teilnehmenden, je 75 Zeilen Liste und acht vollen Runden wiegt 575.139 Bytes;
das Format selbst ist sieben Stufen tief. Endung `.mseturnier`.

**Darin steht der Verlauf, wie ein Mensch ihn nachspielen könnte**: Name,
Spiel, Format **beim Namen**, Einzelspiel, Notiz, Abschlussdatum;
Teilnehmende mit Ausstieg und eingereichter Liste; Runden mit Paarungen und
Ergebnissen.

**Darin steht keine Kennung dieses Bestands** — nicht die des Turniers, der
Teilnehmenden, Runden, Paarungen, Decks oder lokalen Druckzeilen. Die
Paarungen zählen die Teilnehmenden **ab 1 in der Reihenfolge der Datei**.
`printing` reist mit: es ist die Kennung eines Drucks **bei seiner Quelle**
und gilt überall. Nicht darin steht auch das **Protokoll**: es erzählt, was in
diesem Bestand geschah; das eingelesene Turnier beginnt sein eigenes mit
*eingelesen*. Das ist die Antwort auf die Zusage aus AK-03 — eine Datei, die
nichts aus diesem Bestand trägt, kann drüben nichts wiederbeleben, was
jemandem hier gehörte.

### Einlesen prüft, was die Leitung beim Spielen verlangt

Ein eingelesener Verlauf nimmt keinen Weg, den die Leitung kennt; wer eine
Datei schreibt, kann eine zweite Runde ohne erste hinschreiben. Die Domäne
(`ZerlegeTurnier`) verlangt deshalb: Name und Spiel; jede teilnehmende Person
mit Namen, jede Listenzeile mit **Bezeichnung oder Druck** (die Leitung zieht
einen echten Druck ohne Bezeichnung ab — Befund 1 des Zweitblicks über Phase
6.6), Menge ab 1 und bekanntem Bereich; Runden **ab 1 ohne Lücke**, **jede mit
mindestens einer Paarung und höchstens einem Bye** (Befund 2); niemand zweimal
in einer Runde und keine Seite, die es nicht gibt; ein Bye **ohne** Ergebnis (er ist eines); **kein offener
Tisch vor einer weiteren Runde**; jedes Ergebnis nach denselben Grenzen wie
beim Eintragen; ein Abschlussdatum nur als RFC 3339 und nur, wenn das Turnier
sich abschließen ließe (Vermerk 0139).

**Angelegt wird in einem Zug** (`Turnierablage.Einlesen`) — scheitert ein
Teil, entsteht nichts. Ein vergebener Name im selben Spiel ist ein
**Konflikt**, kein stilles Umbenennen. Ein **Format, das es hier nicht
gibt**, ist ein **Befund**: das Turnier kommt ohne Format. Ein Nein wäre zu
streng — das Format ist eine Angabe und kein Zwang (0138) —, und wer eine Datei
einliest, will den Verlauf.

### Der Bericht ist Markdown in der Sprache des Fensters

Kopf mit Stand, Rangliste und jede Runde mit ihren Ergebnissen. **Markdown**,
weil ein Bericht gelesen und weitergegeben wird, nicht wieder eingelesen —
dafür gibt es die Datei. Die Zahlen stehen wie in der Rangliste des Fensters
(Siege samt Byes, Quoten auf vier Stellen), bei einem Einzelspielturnier ohne
Spiel- und Gegnerspielquote (0062/5). **Die Wörter kommen von außen**: die
Domäne schreibt den Bericht, `app` gibt ihr die Wörter in der Sprache der
Anwendung — im Fenster dessen, an Kommandozeile, MCP und REST Deutsch. Ein
laufendes Turnier bekommt denselben Bericht als Zwischenstand; der Kopf sagt
es.

### An allen vier Flächen, in zwei Gestalten

| | Fenster | Kommandozeile | MCP | REST |
| --- | --- | --- | --- | --- |
| Ausgeben | `TurnierAusgabeSpeichern` (Dialog) | `mse turnier ausgeben <id> [datei]` | `lesen.turnier.ausgabe` | `GET /turniere/{id}/ausgabe` |
| Einlesen | `TurnierEinlesenOeffnen` (Dialog) | `mse turnier einlesen <datei>` | `schreiben.turnier.einlesen` | `POST /turniere/einlesen` |
| Bericht | `TurnierBerichtSpeichern` (Dialog) | `mse turnier bericht <id> [datei]` | `lesen.turnier.bericht` | `GET /turniere/{id}/bericht` |

Fenster und Kommandozeile stehen an einer Platte und bekommen einen **Ort**;
MCP und REST stehen an keiner und bekommen den **Text** — dieselbe Teilung
wie bei der Partiedatei (§16).

## Verworfen

- **Die Datenbankkennungen mitschreiben und beim Einlesen umschreiben.** Sie
  wären in der Datei eine Auskunft über einen fremden Bestand, und ein
  Leser, der sie für stabil hält, verwiese ins Leere.
- **Ein unbekanntes Format als Nein.** Siehe oben — ein Format ist keine
  Bedingung für den Verlauf.
- **Den Bericht im Fenster bauen.** Dann hätten drei Flächen keinen, oder
  einen anderen; eine Fähigkeit, vier Antworten.
- **Den Bericht als PDF.** Er wäre schöner und ließe sich nicht in einen Chat
  kopieren; ein Druck aus der Markdown-Vorschau leistet dasselbe.

## Was es kostet

- **Ob damals jede spielbereite Person gepaart wurde, prüft das Einlesen
  nicht.** Dazu müsste die Datei sagen, **wann** jemand ausgestiegen ist; sie
  sagt nur, **dass**. Eine Datei, die jemanden in einer Runde auslässt,
  kommt also durch — angestrichen im Zweitblick über Phase 6.6 und bewusst
  offen gelassen, bis das Protokoll mitreist.

- **Die Reihenfolge der Teilnehmenden ist Teil des Formats.** Wer die Liste in
  der Datei umsortiert, muss die Paarungen mitsortieren — sonst spielt drüben
  jemand anderes gegen wen.
- **Der Bericht an MCP und REST ist Deutsch.** Eine Sprache je Ruf gibt es an
  diesen Flächen nicht; wer Englisch braucht, liest die Rangliste als Daten.
