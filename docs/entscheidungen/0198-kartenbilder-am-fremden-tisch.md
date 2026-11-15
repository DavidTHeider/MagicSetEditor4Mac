# 0198 — Am fremden Tisch zeichnet das Bild, wem das Deck gehört — erkannt an einer Bildkennung, nie an einer Zeile

**Stand:** angenommen am 2026-10-18 · #2377 (P9, Befunde C-11, C-12). Umgesetzt in `internal/domain/sichtbauen.go`, `internal/domain/kontrollbild.go`, `internal/service/partiewirtkartenbild.go`, `internal/service/gastkartenbild.go`, `internal/app/fremdkartenbild.go`, `frontend/src/haken/useSpieltexturen.ts`.

## Lage

Eine Kartenkennung (`KarteID`) ist eine Zeile **eines** Bestands. Am Tisch
lagen sie ungeprüft nebeneinander:

- **Der Gast bekam die Zeilen des Wirts.** In seinem eigenen Bestand
  nachgeschlagen fand er unter derselben Zahl eine **andere** Karte — oder
  keine. Seit `nat04` zeigte der Gast deshalb gar keine Vorderseiten, nur
  Platzhalter und einen Satz.
- **Eine Leihe nannte die Zeilen des Verleihers** und wurde beim Entleiher
  in dessen Bestand nachgeschlagen — derselbe Fehler, still.
- **Die Zeile kam aus dem Deck des Beherrschers**, nicht des Eigners: eine
  Karte, die den Beherrscher wechselte, bekam eine Zeile aus einem fremden
  Deck mit derselben Platzkennung.

Das Abnahmekriterium von P9 verlangt: keine nackten Kartenkennungen des
Wirts am Gast, der Kollisionsfall gleiche Kennung/andere Karte geprüft, und
der Gast sieht die Bilder, die er sehen darf.

## Entscheidung

1. **Ein Abzug nennt seine Quelle** (`Partiedeck.Quelle`): die Peerkennung
   des Bestands, auf den seine Kartenkennungen zeigen — **leer heißt: der,
   in dem die Partie liegt**. Eine Leihe trägt die des Verleihers; ein
   Gastdeck die des Gastes (P9b).
2. **Eine Sicht gibt Kennungen nur aus dem Bestand her, für den sie
   gerechnet wird.** Der Wirt rechnet jedem Gast mit dessen Peerkennung
   (`MitBestandssicht`, im Kontext wie die Sitzbindung); am eigenen Fenster
   mit keiner. Die Zeile kommt aus dem Deck des **Eigners**.
3. **Daneben eine undurchsichtige Bildkennung** (`Sichtobjekt.Bild`): ein
   Hash über Partie, Sitz, Quelle, Abzug und Platz — an jedem Ende dieselbe,
   in keinem Bestand nachzuschlagen. **Sitz und Quelle seit dem Zweitblick
   über Phase 7** (Befund 3): der Abdruck fasst die Gestaltung nicht, und
   zwei gleich gebaute Decks an zwei Sitzen teilten sonst ein Bild. **Leer, wo nicht hineingesehen werden darf**,
   wie die Zeile; sonst erkennte sie eine Karte über einen Umweg.
4. **Wer ein Bild nicht selbst zeichnen kann, fragt den, dem das Deck
   gehört** — mit zwei Kontrollarten, `bildwunsch` und `bild`, nach einem
   **Objekt** des Tisches und nie nach einer Zeile. Der Wirt zeichnet nur,
   was der Gast **sehen darf** und was aus **seinem** Bestand stammt;
   sonst antwortet er mit einem Grund. Nichts geht hinaus, das die Leitung
   nicht trägt.
5. **Der Gast prüft, was ankommt**: ein PNG mit genau den Maßen, die die
   Antwort nennt — sonst landet nichts im Fach und nichts in der
   Webansicht.
6. **Am fremden Tisch zählt eine Zeile nur mit Bildkennung.** Ein Wirt vor
   0.9.0 schickt seine Zeilen ohne sie; das Fenster nimmt sie dann nicht und
   zeigt den Platzhalter — lieber kein Bild als ein falsches.

## Verworfen

- **Die Bilder dem Tischbild beilegen**: sechzig Karten je Zug, und jede
  Änderung schickte alle wieder. Gefragt wird, was gezeigt wird, und das Fach
  hält es.
- **Kartendaten statt Bildern schicken und am Gast zeichnen**: dem Gast
  fehlen Stil, Schriften und die Bilder der Felder — er baute eine Karte nach
  und keine Karte (dieselbe Begründung wie an `Abzugsseite.KarteID`).
- **Die Zeile übersetzen** (eine Zuordnung Wirtszeile → Gastzeile): es gibt
  keine — derselbe Karteninhalt hat in zwei Beständen nichts gemeinsam außer
  dem, was man sieht.

## Rückweg

Ein Wirt ohne diese Fassung schickt keine Bildkennungen; der Gast zeigt dann
Platzhalter wie vor 0.9.0 — **und fragt ihn nie**: gefragt wird nur zu einem
Objekt mit Bildkennung. Das ist mehr als Höflichkeit: eine unbekannte
Kontrollart ist bei jedem Ende ein Fehler aus `Angekommen`, und der beendet die
Sitzung (`Partienetz.Bediene`). Ein Gast ohne diese Fassung fragt ebenfalls
nicht und zeigt Platzhalter — die Zeilen des Wirts bekommt er nicht mehr. Die
Kontrollarten sind damit **erweiternd** für jedes Paar, in dem das neuere Ende
fragt, und kommen bei keinem älteren an.
