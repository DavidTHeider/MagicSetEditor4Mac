# 0201 — Nach außen eine Sichtkennung statt der Kennung

**Datum:** 2026-10-19
**Stand:** angenommen
**Anlass:** #2480
(Zweitblick über Phase 7 von 0.9.0, Befund 1, S1), vor einer Netzfreigabe

## Der Befund

Die Kennung eines Spielobjekts trägt Sitz, Deckplatz und Exemplarnummer
(`bo/p0-12/0`, `partiedeckabzug.go`). Ein verborgenes Objekt verliert in der
Sicht Karte, Name und Ordnung, **behält aber die Kennung** (`objektsicht.go`).
Drei Folgen:

- **Geschwister:** Sobald ein Exemplar offen lag, sind die übrigen desselben
  Deckplatzes in Hand und Bibliothek zuzuordnen.
- **Wiederkehr:** Eine bekannte Karte, die in eine verborgene Zone
  zurückgeht, auch nach dem Mischen, bleibt wiedererkennbar.
- **Reihenfolge:** Die Sicht nennt die Objekte in der Reihenfolge des
  Speichers, und das ist die des Decks. Geschwister stehen also
  nebeneinander. Wer nur die Kennung ersetzte, ließe dieses Leck offen.

Gefunden hat das ein Zweitblick, nicht ein Lauf. Die Analyse der Wege, auf
denen eine Kennung den Wirt verlässt oder zu ihm zurückkommt, liegt dem PR bei.

## Die Entscheidung

**Nach außen geht eine Sichtkennung.** Sie wird aus der Kennung,
einem Geheimnis der Partie und einem Wechselzähler am Objekt gerechnet
(HMAC-SHA256, gekürzt wie `Sitzungsreferenz`). Innen bleibt die Kennung, was
sie ist. Rücknahme, Kette, Speicher und Regeln rechnen weiter mit ihr.

1. **Für alle Betrachter dieselbe.** Eine Sichtkennung je Betrachter schützte
   nur vor Absprachen zwischen Mitspielern, und die verhindert kein Programm.
   Eine gemeinsame hält Präsenz und Bildwünsche einfach, denn sie werden
   unverändert weitergereicht.
2. **Neu gewürfelt, wenn ein Objekt verborgen wird oder seine verborgene
   Lage sich ändert.** Das ist CR 400.7 („ein neues Objekt") dort, wo es auf
   das Geheimnis ankommt. Auslöser:
   - Ein Objekt kommt in eine nicht offene Zone (Bibliothek, Hand).
   - Es wird verdeckt.
   - Seine Ordnung ändert sich in einer verdeckten Zone (Mischen, Umordnen).

   **Ein Wechsel zwischen offenen Zonen würfelt nicht.** Eine Kreatur, die
   stirbt, bleibt am Tisch dasselbe Ding, und das Fenster kann sie vom
   Schlachtfeld auf den Friedhof ziehen sehen.
3. **Gerechnet wird an einer Stelle, nach jedem Befehl:** der Stand vorher
   gegen den Stand danach (`domain.Sichtwechsel`). Kein Schreiber einer Zone
   muss daran denken. Ein neuer Schreiber kann es also nicht vergessen.
4. **Die Sicht ordnet verborgene Objekte nach ihrer Sichtkennung.** Sichtbare
   behalten ihre Reihenfolge.
5. **Ein Ereignis trägt die Sichtkennungen seines Augenblicks**, außerhalb der
   Kette. Die Kette bleibt die der echten Kennungen, und nur der Wirt rechnet
   sie. `Fuer` setzt die Sichtkennungen ein. Der Sichtabdruck wird über dieselbe
   übersetzte Fassung gerechnet, die der Gast bekommt.
6. **Eingehend wird zurückgerechnet** (`Partieleitung.Befehl`). Wer eine
   Sichtkennung nennt, meint das Objekt, das sie jetzt trägt. **Ein Gast über
   die Leitung darf keine echte Kennung nennen.** Fenster, Kommandozeile, MCP
   und REST sind Werkzeuge des Wirts und dürfen es weiter, denn der Wirt hält
   den Bestand ohnehin.
7. **Eine Absage nennt keine echte Kennung.** Ihr Grund geht über die Leitung,
   und die Sätze der Domäne zitieren Kennungen.
8. **Ein Geheimnis je Partie**, erzeugt beim Anlegen und beim Einlesen.
   Eine Partie ohne Geheimnis (alle vor 0.9.0) bekommt es mit ihrem
   nächsten Befehl, im selben Schreiben. **Bis dahin** ist ihre
   Sichtkennung mit leerem Schlüssel nachzurechnen. Das betrifft nur Partien
   aus der Zeit vor dieser Fassung, und keine davon ging über eine
   freigegebene Leitung.
9. **Die Kennung eines neuen Objekts vergibt der Kern**, wo der Befehl
   keine nennt (Spielstein, Emblem). Das Fenster bildete sie aus seiner
   Sicht und schlüge nach diesem Wechsel immer dieselbe vor, denn es sieht
   nur Sichtkennungen. Ein Befehl, der eine nennt, behält sie.
10. **Ein Ereignis trägt seine Zuordnung als Ergänzung an `#sichtnamen`**,
    einen Empfänger, der kein Sitz ist. Ein eigenes Feld bräuchte eine
    Spalte, und eine Schemaänderung ist ein eigener Zug. Die Rücknahme fragt
    deshalb `Ereignis.Erfahren` und nicht mehr, ob es überhaupt eine
    Ergänzung gibt. **Ein Ereignis aus der Zeit davor** trägt keine
    Zuordnung und nennt in der Rückschau seine Kennungen weiter.

## Was nicht dazugehört

- **Der volle Zustand** (`ErlaubtVollenZustand`) bleibt, was er ist. Wer ihn
  bekommen darf, bekommt alles, auch Geheimnis und Kennungen.
- **Die Schlüsselwortbefunde des Tischs** (`tischbefunde`) sind ein eigener
  Seitenkanal für verdeckte Karten. Sie stehen nicht in #2480 und werden
  getrennt geprüft.
- **Kein Schema:** Partie und Objekte liegen als JSON in `partien.zustand`.
  Die zwei neuen Felder brauchen keine Wanderung.

## Einordnung (§16)

**Am Draht der Partie verträglich.** Gäste behandeln Kennungen als undurchsichtig,
denn kein Weg im Fenster zerlegt sie. Was sie jetzt bekommen, ist eine andere
Zeichenkette in denselben Feldern. **Für einen Gast einer älteren Fassung:**
Er sieht Sichtkennungen und schickt sie zurück. Das geht, denn er kennt keine
anderen. Ein Wirt einer älteren Fassung schickt echte Kennungen, und ein neuer
Gast arbeitet mit ihnen, wie bisher. **Ein Rückweg ist deshalb nicht nötig.**

**Kommandozeile, MCP und REST:** Ihre Sichten tragen Sichtkennungen, und ihre
Befehle nehmen beide. Die `kennung` eines Spielsteins oder Emblems ist
jetzt wahlfrei. **Erweiternd.**

## Verworfen

- **Die Kennung bei jedem Zonenwechsel neu vergeben** (CR 400.7 wörtlich).
  Das hätte Rücknahme, Commanderschaden (CR 903.10a verfolgt den Commander
  über Zonen hinweg), Ereignisverweise und gespeicherte Lagen gebrochen. Die
  Sichtkennung erreicht dasselbe für das Geheimnis, ohne die innere
  Identität zu berühren.
- **Einen Zähler an jedem Schreiber einer Zone hochsetzen.** Laut Analyse
  sind es sechs Stellen außerhalb von `NeuesObjekt`. Eine vergessene ließe
  das Leck offen, und kein Lauf sähe es.
