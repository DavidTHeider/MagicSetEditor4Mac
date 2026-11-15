# 0155 — Teilen ist eine Kopie

**Stand:** angenommen · **Datum:** 2026-10-01
**Issue:** NET-04 (#1708)
**Berührt:** [0150](0150-einladen-und-vergleichen.md) (Einladungen),
`docs/FAEHIGKEITEN.md`

## Die Frage

NET-04 verlangt, dass Einladungen, Karten, Sets, Sammlungen und Decks über
Mail, Nachrichten und AirDrop weitergegeben werden können — mit einer
Vorschau auf Inhalt, Umfang und Verluste (AK-02), mit Alternativen, wo ein
Ziel fehlt (AK-03), und ohne dass Teilen Besitz überträgt oder eine
Einladung annimmt (AK-04).

## Was MSE2 tut

**Es teilt nicht.** Gesucht in `/Users/david/mse2-quelle/src` (Fassung
f9d9356 vom 2021-07-06) nach `NSSharing`, `sharing`, `mailto`, `wxEmail` und
dem Wort `share`: einzig ein Kommentar in `gui/set/console_panel.cpp`, der
vom Verlauf der Konsole spricht. Marke **new**.

## Was gilt

1. **Drei Schritte**: vorbereiten, anbieten, verwerfen — `TeilenVorbereiten`,
   `TeilenAnbieten`, `TeilenVerwerfen`.
2. **Vorbereiten legt eine Kopie ab**, je Teilung ein eigener Ordner unter
   `~/Library/Caches/MagicSetEditor4Mac/teilen/<Kennung>` (`0700`, Dateien
   `0600`, atomar geschrieben). Die Kennung ist 32 Zeichen Hex; eine andere
   erreicht die Platte nicht, und ein Dateiname wird nie zum Weg.
3. **Je Art der Weg, den ihre Ausfuhr schon geht**:
   - **Einladung**: der Code als Text — Nachrichten und Mail setzen ihn ein;
     dazu „Code kopieren".
   - **Kartenbild**: PNG über die Bildausfuhr.
   - **Set**: `.mse-set` als **Archiv** über `Setschreiber.Abzug` — **ohne
     den Arbeitsort zu verschieben**. `Schreibe` wechselt ihn (#226); ein
     Set, dessen nächstes ⌘S in einen Ordner für Mailanhänge ginge, wäre der
     Fehler, den dieser Satz verhindert.
   - **Deck**: die Liste, die `DecklisteVorschau` wieder liest — und die es
     dafür an allen vier Flächen als eigene Fähigkeit gibt
     (`DecklisteAusgeben`, `lesen.deck.liste`, `mse deck liste`,
     `GET /decks/{id}/liste`); dazu „Liste kopieren".
   - **Sammlung**: CSV, die verlustfreie Form.
4. **Was mitgeht und was fehlt, sagt die Domäne** (`domain.Teilart`), als
   Schlüssel, die das Fenster in seine Sprache setzt. **Eine Sammlung trägt
   Privates** — Erwerbspreise, Lagerorte, Notizen, Wünsche —, und das steht
   da, bevor sie hinausgeht.
5. **Angeboten wird über `NSSharingServicePicker`** am Fenster — nativ und
   nicht nachgebaut (§6). Welche Ziele es gibt, weiß das System; nicht
   eingerichtete zeigt es nicht. Empfänger und Versand bestätigt der Mensch
   im gewählten Dienst — dieses Programm schickt nichts selbst.
6. **Über 20 MB** empfiehlt die Vorschau AirDrop oder Mail Drop.
7. **Verworfen wird auf „Fertig"**, beim Verlassen des Bausteins und beim
   nächsten Start für alles, was älter ist als ein Tag. Nicht sofort nach
   dem Anbieten: Mail legt den Anhang beim Verfassen an, AirDrop beim
   Senden.
8. **Nur am Fenster.** Ein Teilen-Dialog gehört zu einem Fenster wie ein
   Dateidialog; die Dateien selbst geben die Flächen ohne Fenster über ihre
   Ausfuhrwege aus.

## Was es nicht ist

**Keine Eigentumsübertragung und keine angenommene Einladung** (AK-04): wer
eine Einladung teilt, gibt einen Code weiter; angenommen wird sie drüben, mit
Vergleich ([0150](0150-einladen-und-vergleichen.md)). Wer ein Deck teilt,
gibt eine Liste weiter und keine Exemplare.

## Verworfen

- **Eine eigene Liste von Zielen** (Mail-Knopf, AirDrop-Knopf): sie zeigte
  Ziele, die auf diesem Rechner nicht eingerichtet sind, und wüsste von
  neuen nichts.
- **`mailto:` mit Anhang**: `mailto` trägt keine Anhänge, und ein Text im
  Link landete im Protokoll des Browsers.
- **Die Kopie sofort nach dem Anbieten löschen**: der Dienst liest sie
  später.
- **Der Dialog am Knopf verankert**: er zeigt heute von der Oberkante des
  Fensters herab. Am Knopf hieße, dass die Seite seine Lage durch die Naht
  reicht — in CSS-Pixeln einer WebView, deren Ursprung oben liegt, während
  der des Fensterinhalts unten liegt. Das ist eine Umrechnung mit eigener
  Fehlerklasse für einen Pfeil; **vertagt**, nicht verworfen.

## Gemessen

- Domäne, Dienst, App und Baustein mit ihren Läufen; die Gegenprobe ohne das
  Aufräumen nach einem Fehlschlag ist rot.
- Der Set-Abzug: ein Archiv mit der Hauptdatei, Byte für Byte, und der
  Arbeitsort bleibt, wo er war.
- Die Deckliste: der Rundlauf durch `mse deck-liste` am gebauten Programm.
- **Der Dialog am nativen Fenster**: die Hüllenprobe `net04`
  (`docs/belege/huellenprobe/net04.md`) — Vorschau, eine Kopie mit `0600`
  unter dem eigenen Zuhause der Probe, der Teilen-Dialog von macOS mit Mail,
  Nachrichten und den Erweiterungen dieses Rechners, und „Fertig" räumt die
  Kopie weg. **Die Abzüge bleiben außerhalb des Baums**: der Dialog zeigt
  die Kontaktvorschläge des Rechners.
