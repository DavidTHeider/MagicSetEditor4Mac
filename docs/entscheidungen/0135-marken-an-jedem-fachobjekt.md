# 0135 — Marken an jedem Fachobjekt, gespeicherte Filter als Einstellung

**Datum:** 2026-09-26
**Anlass:** UX-12 (#1684) — „Benutzerdefinierte Tags und gespeicherte Filter in allen Fachsichten"
**Stand:** entschieden; gebaut in mehreren Zügen, der erste trägt Bestand, Anwendungsfall und die Flächen ohne Fenster

## Die Frage

Seit [0053](0053-marken-liegen-neben-dem-set.md) hängt an einer **Karte** eine
Marke — ein Wort, um sie wiederzufinden. UX-12 verlangt dasselbe für acht
weitere Arten: Sets, Sammlungsposten, Decks, Pakete, Symbole, Turniere,
Partien und Handelsvorgänge. Dazu mehrere Marken als Filter mit
dokumentierter UND-/ODER-Bedeutung und gespeicherte Filter, die einen
Neustart und die Ausfuhr der Einstellungen überleben.

## Was nachgelesen wurde

**MSE2 kennt keine Marke** — nachgelesen in 0053; die Erweiterung ist `new`.
Auch für die acht neuen Arten gibt es keine Vorlage: fünf davon (Sammlung,
Decks, Turniere, Partien, Handel) hat MSE2 gar nicht.

## Die Entscheidung

1. **Das Wort bleibt „Marke".** Der Entwurf sagt *Eigene Marken* und
   *+ Marke* (Artboard *Sammlung*); „Tag" stünde neben „Marke" für dasselbe.
2. **Eine Tabelle für alle neuen Arten** — `objektmarken(profil_id, art,
   kennung, marke)`. **Die Karte bleibt in `karten_marken`**: ihre Marken
   hängen über einen Fremdschlüssel an der Karte und gehen mit ihr, und sie
   umzuziehen wäre eine Datenmigration in einem eigenen Zug (§5). Der
   Anwendungsfall führt beide unter einer Frage zusammen.
3. **Eine Marke gehört dem Profil**, in dem sie gesetzt wird — auch an einem
   Paket, das dem Rechner gehört: dort ist sie eine **lokale Anmerkung** und
   ändert das Paket nicht (AK-01). Wer in ein anderes Profil wechselt, sieht
   dessen Marken und nicht die eigenen (AK-04).
4. **Eine Marke ist nie ein Recht** (AK-02). Keine Prüfung in
   `internal/service` liest eine Marke; setzen darf, wer im Profil schreiben
   darf (`lokalSchreiben`), lesen, wer lesen darf.
5. **Kollisionen**: Umbenennen auf einen Namen, den es im Profil schon gibt,
   wird **abgewiesen** — es sei denn, es heißt ausdrücklich
   *zusammenführen*. Dann tragen alle Objekte beider Marken danach die eine.
   Entfernen nimmt die Marke von **allen** Objekten des Profils.
6. **Filter**: `und` heißt *jede genannte Marke*, `oder` *wenigstens eine*.
   Keine Marke heißt kein Filter. Groß- und Kleinschreibung zählen nicht —
   die Normalform aus 0053 gilt überall.
7. **Gespeicherte Filter sind eine Einstellung** der Ebene *Profil*,
   `filter.gespeichert`: eine Liste aus Name, Sicht, Marken und Verknüpfung.
   Damit überleben sie den Neustart und reisen mit der Ausfuhr der
   Einstellungen, ohne dass es dafür einen zweiten Weg gibt.
8. **Wenn ein Objekt geht, gehen seine Marken** — je Art ein Auslöser in
   `docs/INVARIANTEN.md`. Ein Paket hat seinen Eintrag in `packages`; ein
   Handelsvorgang seinen in `handelsfolge`.

   **Und eine Zahlkennung hat genau eine Form** (`domain.Markenart.Kennung`,
   nachgetragen nach dem Zweitblick): Karte, Set, Exemplar, Deck und Turnier
   nehmen eine ganze Zahl ab eins und legen sie ohne führende Null ab.
   SQLite vergleicht `011` mit einer Zahlspalte als 11 — ohne diese Regel
   bestand `011` die Frage nach Deck 11, lag danach als `011` in der
   Textspalte und wurde weder von der Abfrage mit `11` noch vom Auslöser
   (`CAST(OLD.id AS TEXT)`) gefunden. Partie, Handel und Paket haben
   Textkennungen und bleiben, wie sie kommen.

   **Damit ändert sich eine Antwort, und das steht hier** (Nachlauf über
   UX-12): eine Kennung, die keine ganze Zahl ab eins ist — `abc`, `11.0`,
   `0` —, ist jetzt **ungültig** und nicht mehr **unbekannt**. Am Dienst
   heißt das `400`/`validierung` statt `404`/`nichtGefunden`; am Fenster, an
   der Kommandozeile und am Werkzeugserver dieselbe Fehlerart, weil alle vier
   durch denselben Dienst gehen. Veröffentlicht war die alte Antwort nie —
   UX-12 steht unter *Unveröffentlicht*.
9. **Symbole fehlen im ersten Zug, mit Grund**: ein Symbol ist heute kein
   eigenes Objekt, sondern ein Feld eines Sets. Es bekommt seine Marken mit
   DOM-02 (#1697),
   das es zur Ressource macht.

## Verworfen

- **Die Karte in die neue Tabelle ziehen** — eine Datenmigration ohne
  Gewinn für den Nutzer, und der Fremdschlüssel ginge verloren.
- **Gespeicherte Filter in einer eigenen Tabelle** — dann bräuchte die
  Ausfuhr der Einstellungen einen zweiten Weg für genau diese Liste.
- **Marken in die `.mse-set`** — aus demselben Grund wie in 0053.
