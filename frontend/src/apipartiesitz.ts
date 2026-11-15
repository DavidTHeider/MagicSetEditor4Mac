// **Ein Sitz an der Naht** — aus `apipartie.ts` herausgelöst, als sie mit
// P9b (#2377) an die Zeilengrenze aus §2 stieß. Eingeführt wird er weiter von
// dort; diese Datei hält nur, was ein Sitz ist.

// Sitz ist ein Platz am Tisch — ohne das, was an ihm geheim ist.
export interface Sitz {
  kennung: string
  name: string

  // deckabdruck sagt, dass zwei Seiten dasselbe Deck meinen — **ohne es zu
  // zeigen**.
  deckabdruck: string
  karten: number

  // wartet sagt, dass dieser Sitz auf das Deck seines Gastes wartet, und
  // eingeladen, auf wen — **nur am Fenster des Wirts** (P9b, #2377).
  wartet: boolean
  eingeladen: string

  // profilID ist das Profil, an das dieser Sitz gebunden ist (G8,
  // #1465)
  // — **leer heißt: keines**, und das ist der Normalfall.
  //
  // **Daran hängt der Sichtwechsel**: wer das gezeigte Profil wechselt,
  // bekommt die Sicht **dieses** Sitzes. Ein Sitz ohne Profil bleibt über die
  // Sichtwahl daneben erreichbar.
  profilID: string

  // leben und vergiftung machen den Sitz zum Spieler — CR 119.1 und
  // CR 122.1f (G9, #1592).
  //
  // **Für jeden sichtbar**: ein Lebensstand ist am Tisch öffentlich.
  leben: number
  vergiftung: number

  // manavorrat ist das ungenutzte Mana als Buchstabenfolge — `RRG`, **leer
  // heißt leer** (M3,
  // #1605).
  //
  // **Für jeden sichtbar**, aus demselben Grund wie der Lebensstand: CR
  // 106.4a und CR 106.4b verlangen, dass ein übriger Vorrat angesagt wird.
  manavorrat: string

  // **Die fünf, die bis zum 2026-09-21 in der Domäne lagen** — PLAY-03
  // (#1711).
  //
  // `domain.Sitz` führt zwölf Felder, und acht kamen hier an. Die übrigen
  // waren gerechnet, gespeichert und geprüft — und nie zu sehen.
  //
  // **Keines ist geheim**: CR 728.1 setzt Strahlungsmarken an den
  // **Spieler**, CR 400.2 macht die Kommandozone öffentlich, CR 903.8 zählt
  // Würfe, die alle gesehen haben, CR 903.10a Kampfschaden, und der Griff
  // in eine leere Bibliothek (CR 704.5b) geschieht am Tisch.

  // strahlung sind die Radioaktivitätsmarken — CR 728.1.
  strahlung: number

  // commander ist die Kartenkennung des Commanders — CR 903.3. **Leer
  // heißt: keiner**, und das ist der Normalfall.
  commander: string

  // commandersteuer zählt die Würfe aus der Kommandozone — CR 903.8.
  commandersteuer: number

  // commanderschaden ist der Kampfschaden **je Commander** — CR 903.10a.
  //
  // **Je Kennung und nicht als Summe**: zwei mit je zwanzig sind vierzig
  // und trotzdem keine einundzwanzig.
  commanderschaden: Record<string, number>

  // leergezogen vermerkt den Griff in eine leere Bibliothek — CR 704.5b.
  leergezogen: boolean

  // eroeffnung ist der Stand im Beginn der Partie — CR 103.5
  // (#1839).
  // **Leer bei einer Partie von vorher**: sie beginnt nicht noch einmal.
  eroeffnung: '' | 'ausstehend' | 'offen' | 'behalten'

  // mulligans zählt alle genommenen, auch den freien aus CR 103.5c;
  // unterlegen, wie viele Karten noch unter die Bibliothek müssen.
  mulligans: number
  unterlegen: number
}
