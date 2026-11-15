# 0075 — Der Satz für Menschen und der Grund fürs Protokoll sind zweierlei

**Datum:** 2026-09-06
**Stand:** angenommen
**Anlass:** Werkstück 1 von Phase 4.13
(#1352)

## Die Lage

`fmt.Errorf("…: %w", ursache)` tut **zwei Dinge auf einmal**: es hängt die
Ursache in die Fehlerkette und ihren Text in den Satz. Der Satz ist genau das,
was `internal/app.Uebersetze` als `Fehler.Meldung` weitergibt.

**Am 2026-09-06 gezählt**, über den ganzen Baum `internal/`:

| | Zahl |
| --- | --- |
| `fmt.Errorf` mit einem der neun Wahrzeichen | **500** |
| davon mit eingesetztem Fremdtext | **323** |
| davon mit einem **vollen Pfad** (nicht `filepath.Base`) | **25** |
| davon mit einem **zweiten `%w`** — einem fremden Go-Fehler | **42** |
| Vereinigung der beiden letzten (13 überschneiden sich) | **54** |

Die schärfste Form stand in `internal/storage/fehler.go`:

```go
return fmt.Errorf("%w: %w", domain.ErrBelegt, fehler)
```

— der **rohe SQLite-Text**, eingewickelt und weitergereicht. §8 sagt: „Rohe
SQLite-Meldungen erreichen die Oberfläche nie."

## Warum das nicht nur eine Schönheitsfrage ist

**Das Fenster übersetzt über den Schlüssel** und sieht den Go-Satz nie; ein
Vertragstest hält beide Katalogseiten gegen Go, und der Satz ist dort nur der
Rückfall für einen Schlüssel, den es nach diesem Test nicht geben kann.

**Kommandozeile, MCP und REST haben keinen Katalog.** Sie zeigen `Meldung`
immer. Für sie waren alle 54 Stellen offen — und das sind drei von vier
Flächen (§16).

## Die Entscheidung

**Eine eigene Fehlerform**, `domain.MitGrund(art, grund, satz, werte…)`:

- `Error()` gibt **das Wahrzeichen und den eigenen Satz** — und nicht den Grund.
- `Unwrap() []error` gibt **beide**, damit `errors.Is` weiterhin das Wahrzeichen
  *und* den Grund findet.
- `domain.GrundVon(fehler)` holt den Grund heraus — **für das Protokoll und nur
  dafür**.
- `app.Uebersetze` schreibt ihn auf, bevor die Meldung hinausgeht.

**Der Grund verschwindet also nicht.** Er wechselt den Kanal: aus der Meldung
ins Protokoll. Wer einen Fehler sucht, findet ihn dort; wer die Anwendung
bedient, liest einen Satz, der sagt, was zu tun ist.

## Was verworfen wurde

**Die Meldung an der Oberfläche abschneiden.** `Error()` ist eine flache
Zeichenkette; „alles hinter dem letzten Doppelpunkt weg" trifft mal den
Treibertext und mal den halben Satz. Eine Regel, die sich nur an einem Beispiel
prüfen lässt, ist keine.

**Alles ins Protokoll und gar nichts in die Meldung.** Dann stünde an drei
Flächen nur noch „Es ist ein unerwarteter Fehler aufgetreten" — und der Grund,
aus dem #1055
den eigenen Satz überhaupt eingeführt hat, fiele wieder weg.

## Was **nicht** umgestellt wurde, und warum

**Nicht jeder eingewickelte Fehler ist ein fremder.** Zwanzig Stellen wickeln
einen Fehler aus **diesem** Projekt ein, und dessen Satz ist die Auskunft:

- **Skriptfehler** tragen Zeile und Spalte — der Inhalt der Skriptkonsole.
- **Paketfehler** benennen die Datei im Paket, die fehlt.
- **`unbekannteAbweisen`** nennt den Namen der Migration, die diese Fassung
  nicht kennt — ohne ihn bliebe nur, dass etwas nicht stimmt.
- **Paketorte** bleiben ganz: ein Paket liegt dort, wo jemand es hingestellt
  hat, und `msepaket/bildquelle_test.go` unterscheidet **genau daran**, ob eine
  leere Wurzel durchsucht wurde. Gekürzt wäre der Wächter dort blind.

Sie stehen als benannte Ausnahmen im Wächter
(`internal/werkzeuge/meldungswege_test.go`), jede mit ihrem Grund.

## Was der Wächter nicht prüft

**Ob ein Pfad einer ist, den jemand selbst gewählt hat.** Eine Maschine
unterscheidet das nicht; die Ausnahmen tun es.

**Und ob der Satz gut ist.** Er sieht, dass nichts Fremdes darin steht, nicht
dass darin steht, was zu tun ist. Das ist Werkstück 3 derselben Phase.

## Zwei Stellen zeigen den Grund trotzdem — benannt

**Nachgetragen am 2026-09-07**, Befund 3 des Zweitblicks über Phase 4.13. Der
Satz oben sagt „für das Protokoll und nur dafür", und zwei Stellen halten ihn
nicht. Das war bis dahin **nicht aufgeschrieben** — und eine Zusage mit stillen
Ausnahmen ist schlimmer als eine engere Zusage (§10).

| Stelle | Was sie zeigt | Warum |
| --- | --- | --- |
| `app.startsatz` — das native Modal | `fehler.Error()` unter dem Ratschlag | **Die Anwendung ist danach weg.** `main` zeigt den Dialog und beendet sich; es gibt kein Fenster, in dem jemand nachsehen könnte. Der Text verweist auf die Protokolldatei, und der Grund steht darunter für die Rückfrage |
| `app.Startbefund.Grund` — die Wiederherstellungsseite | derselbe Text unter „Was gemeldet wurde" | **Die Diagnoseseite braucht einen offenen Bestand**, und genau der fehlt. Das Protokoll gibt es — aber wer hier steht, käme nur heran, indem er die Datei von Hand sucht. Der Satz aus dem Katalog steht darüber und sagt, was zu tun ist; der Grund darunter, für den, der einen Fehler sucht |

**Hier stand zuerst „es gibt kein Protokoll", und das war falsch**: `main` legt
es **vor** dem Bestand an, und der Startfehler wird darin vermerkt. Angestrichen
vom Nachlauf zum Zweitblick über Phase 4.13. Der Grund für beide Ausnahmen ist
nicht, dass es keine zweite Spur gäbe — sondern dass sie in diesem Augenblick
niemand erreicht.

**Beide zeigen ihn unter einer eigenen Überschrift**, nicht anstelle des
Satzes. Das ist der Unterschied zu dem, was diese Entscheidung abstellt: dort
stand der Grund **statt** einer Meldung, hier steht er **unter** ihr.

**Und beide sind die Ausnahme, nicht die Regel.** Wer eine dritte Stelle
braucht, trägt sie hier ein — sonst ist sie ein Befund.
