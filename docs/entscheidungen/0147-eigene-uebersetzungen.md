# 0147 — Eigene Übersetzungen liegen über einer Basis und gehören dem Rechner

**Stand:** entschieden am 2026-09-30 · LOC-02.AK-01 bis AK-04
(#1688)

## Die Frage

LOC-02 verlangt, dass jemand „aus einer bestehenden Sprache eine neue
Übersetzung erstellen, bearbeiten, prüfen, importieren, exportieren und lokal
aktivieren" kann — mit einem Assistenten, der eine Basissprache kopiert
(AK-01), Prüfungen ohne ausgeführten Code (AK-02), einer Rückfallkette ohne
leere Beschriftung (AK-03) und einer Ausgabe mit Herkunft, Lizenz und Stand,
aber ohne Persönliches (AK-04). Offen war, **wo** eine solche Übersetzung
liegt, **wogegen** sie geprüft wird und **was „kopieren"** heißt.

## Was MSE2 tut

**Es lädt `.mse-locale`-Pakete und legt keine an.** Gesucht in
`src/data/locale.cpp` und `src/gui/` (Fassung f9d9356 vom 2021-07-06): ein
Sprachpaket ist eine Datei im Datenverzeichnis, und wer eine neue Sprache
will, schreibt sie von Hand. Eine Prüfung gegen eine Vorlage gibt es nicht.
Die Marke ist deshalb **new**; die dreizehn Pakete des Packs stehen im
Inventar ([`SPRACHINVENTAR.md`](../SPRACHINVENTAR.md)).

## Was gilt

1. **Kopieren heißt verweisen.** Eine neue Übersetzung nennt ihre Basis und
   trägt anfangs keinen Satz; was fehlt, zeigt die Basis. Das sieht aus wie
   eine Kopie — mit dem Unterschied, für den es die Abweichung gibt: eine
   Kopie machte einen unübersetzten Satz ununterscheidbar von einem
   übersetzten, und der Assistent könnte nicht mehr sagen, was fehlt.
2. **Geprüft wird gegen den deutschen Katalog**, flach eingebettet
   (`internal/app/sprachvorlage.json`, geschrieben und gehalten von
   `frontend/src/sprache/sprachvorlage.test.ts`). Deutsch, weil es die
   Ausgangssprache ist (§9). Jeder Satz trägt den deutschen, gegen den er
   übersetzt wurde; weicht der heutige ab, ist er **veraltet** — markiert,
   nicht verworfen. Ein Pfad, den diese Fassung nicht kennt, bleibt im Paket.
3. **Kein Code**: Platzhalter müssen dieselben sein wie im deutschen Satz,
   `**` und `` ` `` müssen schließen, und ein Tag, eine Skript- oder
   Datenadresse oder ein Steuerzeichen weist den Satz ab
   (`domain.Katalogvorlage.PruefeUebersetzung`). Das Fenster setzt Text,
   nie Markup.
4. **Ein Sprachpaket gehört dem Rechner**, wie ein installiertes MSE-Paket
   (Wanderung 0092, ohne Profil). **Welches gilt, ist eine Einstellung des
   Menschen** neben `sprache`: `sprache.eigene`. Gibt es das Paket nicht
   mehr, gilt wieder `sprache` — zurückgefallen, nicht leer.
5. **Die Rückfallkette**: eigene Sätze → Basis → (fehlt das Paket) die
   gewählte Sprache. Die Basis ist ein vollständiger Katalog; das hält
   `zusatzsprachen.test.ts`.
6. **Menüleiste und Dialoge sprechen die Basis.** Ihre Texte stehen in Go
   (`internal/app/sprachkatalog*.go`) und nicht in der Vorlage; eine eigene
   Übersetzung erreicht sie nicht. Das `lang`-Attribut trägt die Kennung der
   eigenen.
7. **Eine Änderung geschieht in einem Zug** (`Sprachpaketablage.
   SprachpaketAendern`, `_txlock=immediate`): die Sätze stehen in einer
   Zeile, und Fenster und MCP-Server sind zwei Prozesse über derselben Datei.
8. **Die Datei** heißt `mse4mac/sprachpaket/1` und hat genau diese Felder:
   Format, Kennung, Name, Basis, Stand, Lizenz, Herkunft, Programm,
   Zeitpunkt, Sätze. **Ohne Lizenz geht sie nicht hinaus.** Eingelesen wird
   sie streng — ein fremdes Feld weist sie ab —, und eine vorhandene ersetzt
   nur, wer es sagt.
9. **An allen vier Flächen dieselben Wege**: Fenster (Einstellungen →
   Allgemein), `mse sprachpaket`, sieben MCP-Werkzeuge, sieben REST-Wege.

## Was das kostet

- **Die dritte Zählform folgt der Sprache, die sie trägt** — seit dem
  Zweitblick über Phase 6.5 (Befund Z4): bis dahin fragte `zaehlwort` stets
  `Intl.PluralRules('ru')`, und eine eigene polnische Übersetzung zählte 21
  russisch. **Die Regel reist mit der Zählform**: `ueberlagere` gibt jeder
  Form, deren Wörter die Übersetzung setzt, deren Kennung mit — nicht
  aufzählbar, damit sie in keiner flachen Form als Satz erscheint. Die
  übrigen zählen nach der Basis, deren Wörter sie tragen; eine Sprache, die
  `Intl` nicht kennt, zählt russisch. Im Zweitblick stellte noch der
  Sprachraum die Regel in einem Zustand des Moduls, und der zuletzt
  gezeichnete galt für alle (Nachlauf, N5).
- **Der Editor zeigt fünfundzwanzig Sätze je Seite** und sucht in Pfad,
  deutschem Satz und Übersetzung. Kontext ist der Pfad; eine Vorschau an
  der Stelle, an der ein Satz erscheint, gibt es nicht.

## Verworfen

- **Eine Kopie der Basis** — aus dem Grund unter 1.
- **Pakete als Dateien neben den MSE-Paketen.** Sie gingen an der Sicherung
  des Bestands vorbei, und jede Fläche bräuchte einen eigenen Ort.
- **Pakete je Profil.** Eine Sprache ist keine Fachsache; zwei Profile
  derselben Person sollen dieselbe Übersetzung sehen.
