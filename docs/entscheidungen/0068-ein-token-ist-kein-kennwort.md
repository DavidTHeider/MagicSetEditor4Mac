# 0068 — Ein Token ist kein Kennwort

**Entschieden am 2026-09-05**, Werkstück 10 von Phase 4.12
(#1351),
Punkte 12 und 13 der Kommentierung.

## Der Anlass

Die Kommentierung verlangt technische Profile mit Scope und Ablauf und dazu
Zugriffstoken. Beim Bauen standen vier Fragen an, deren Antwort nicht aus der
Aufgabe folgt — und die eine davon sieht aus wie eine Sicherheitslücke, wenn
man ihre Begründung nicht kennt.

## 1. Gehütet wird mit SHA-256 und nicht mit Argon2id

**Obwohl das Kennwort (Werkstück 6) Argon2id benutzt.** Der Unterschied ist
nicht Nachlässigkeit, sondern die Frage, wogegen gehütet wird.

**Argon2id ist teuer, damit ein geratenes Kennwort teuer wird.** Menschen
wählen kurze Kennwörter, und ein Angreifer mit der gehüteten Form probiert
Wörterbücher; jede Millisekunde Rechenzeit je Versuch multipliziert sich mit
seiner Versuchszahl. Das ist der Zweck der Streckung, und für Kennwörter ist
sie unverzichtbar.

**Ein Token hier ist ein Zufallswert.** `crypto/rand.Text()` gibt Base32 nach
RFC 4648 mit mindestens 128 Bit Zufall — nachgelesen in der Dokumentation der
Standardbibliothek von Go 1.26.6, nicht erinnert (§1). Gegen so viele
Möglichkeiten hilft keine Streckung: sie ändert nichts an der Zahl der
Versuche, sie kostet nur den, der prüft. Und der prüft bei jedem einzelnen
Aufruf am REST-Zugang.

**Verglichen wird zeitkonstant** (`crypto/subtle`). Ein Vergleich, der beim
ersten falschen Byte abbricht, verrät Länge und Stelle.

**Was diese Entscheidung nicht ist**: eine Aussage über Kennwörter. Sie gilt
für **zufällige** Geheimnisse und für nichts sonst; das Kennwort bleibt, wo es
ist.

## 2. Die Kennung reist im Token mit

**Gestalt: `mse_<kennung>_<geheimnis>`.** Ohne die Kennung müsste jede Prüfung
**jede** gehütete Form durchprobieren — mit ihr ist es ein Nachschlagen und
ein Vergleich. Dieselbe Bauform benutzt GitHub für seine Zugriffstoken.

**Die Kennung ist kein Geheimnis**; sie steht in jeder Liste. Was schützt, ist
der hintere Teil.

**Und der Trenner kommt im Geheimnis nicht vor** — das ist die Wahl des
Alphabets und keine Hoffnung: Base32 kennt `A`–`Z` und `2`–`7`. Der erste
Anlauf nahm `base64.RawURLEncoding`, dessen Alphabet `_` enthält; das Trennen
zerlegte das Geheimnis mit, **aber nur manchmal** — je nachdem, welche Bytes
gezogen wurden. Ein Fehlschlag in etwa jedem vierten Fall ist einer, den man
für Flackern hält.

## 3. Die Erlaubnis hat zwei Stufen und kommt aus der HTTP-Methode

**Zwei und nicht hundert**: `lesen` und `schreiben`, und `schreiben` schließt
Lesen ein. `internal/mcp` trennt seit jeher genauso (`Lesend`/`Schreibend`).

**Welche Stufe eine Anfrage verlangt, sagt ihre Methode** — `GET` und `HEAD`
verlangen `lesen`, alles andere `schreiben`. Jeder Weg dieses Dienstes ist mit
seiner Methode eingetragen (`"GET /sets"`); eine Liste daneben wäre eine
zweite Wahrheit, die still veraltet.

**Was das kostet, gehört dazu**: ein paar `POST`-Wege arbeiten, ohne etwas
abzulegen — `/paketdateien/pruefung`, `/symbol/{rechnung}`,
`/sammlung/liste/deuten`. Ein Lesetoken kommt dort **nicht** durch, obwohl
nichts geschrieben würde. Die Abweichung geht in die sichere Richtung, und die
Alternative — eine gepflegte Liste der harmlosen `POST`s — hat die
Eigenschaft, still falsch zu werden.

## 4. Geprüft wird nur am REST-Zugang

**Er ist die einzige Fläche, die auf einem Port lauscht.** Das Fenster weiß,
wer davorsitzt; Kommandozeile und MCP laufen als der Mensch, der sie startet,
mit seinen Dateirechten. Ein Token dort zu verlangen, prüfte etwas, das schon
feststeht.

**Verwaltet** werden Token dagegen an allen vier Flächen (§16): anlegen,
auflisten, ändern, zurückziehen ist eine Verwaltungsaufgabe wie jede andere.

**Und ein vorgezeigtes Token stellt die Anfrage in sein Profil.** Es
überschreibt das gezeigte — genau darum gibt es technische Profile: was der
Automat anlegt, soll nicht im Profil des Menschen liegen. Belegt ist das im
Flächenlauf gegen den **gebauten** Dienst, nicht in einem Test einer Schicht:
dort ist die Wache eine Schnittstelle und der Klartext eine Zeichenkette, die
der Lauf selbst erfunden hat.

## Was verworfen wurde

**Ein Rechtebaum je Endpunkt.** Er wäre eine Aufzählung, die bei jedem neuen
Weg gepflegt sein will — und ein ungepflegter Eintrag in einer Rechteliste ist
gefährlicher als keiner. Rechte und Rollen **über Nutzer hinweg** sind
Werkstück 11 und eine andere Frage: nicht „was darf dieses Token", sondern
„was darf dieser Mensch bei den Sachen eines anderen".

**Ein Ablauf, der löscht.** Die Kommentierung sagt „löschbar", nicht
„gelöscht". Ein Bestand, der hinter dem Rücken Zeilen entfernt, nimmt jemandem
seine Daten, während er nicht hinsieht. Was hier steht, ist ein Zeitpunkt, ab
dem etwas nicht mehr gilt.

**Ein zweites Geheimnis beim Ändern.** Wer eine Bezeichnung ändert, soll
niemandem den Zugang entziehen. Wer ein neues Geheimnis will, legt ein neues
Token an und zieht das alte zurück — und das ist ein sichtbarer Schritt.
