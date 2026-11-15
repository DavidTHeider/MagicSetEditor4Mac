// **Was sich aus den Karten eines Decks rechnen lässt** — Werkstück I von
// Phase 4.10
// (#1325).
//
// # Eigene Datei, weil es reine Rechnung ist
//
// Kurve und Farbverteilung hängen an nichts als den Feldwerten und den Mengen.
// Als Funktionen sind sie tabellengetrieben prüfbar; in einer Komponente wären
// sie es nur über den DOM.
//
// # Die Kostenschreibweise
//
// MSE schreibt Manakosten als Folge von Symbolen: `3WW`, `R`, `UU`, `2/W`,
// `X`. Die Zahl vorn ist der farblose Anteil, jeder Buchstabe ein farbiger.
// Nachgelesen in MSE2: `src/data/field/text.cpp` kennt `casting cost` nicht
// besonders — die Deutung steckt in den Spielpaketen, und `magic.mse-game`
// schreibt sie so (Fassung f9d9356 vom 2021-07-06).
//
// **Was diese Rechnung nicht kann**, und was daneben steht statt in einem
// Kopf: hybride Symbole (`2/W`) zählt sie als das, was zwischen den Schrägen
// steht, und `X` als null. Beides ist die Auslegung, die auch Scryfall in
// seinem `cmc` nimmt — nachgelesen am 2026-09-03 in ihrer Feldbeschreibung —,
// und beides ist selten genug, dass eine eigene Regel mehr verspräche, als
// sie hielte.

// Farbe ist eines der fünf Manasymbole.
export type Farbe = 'weiss' | 'blau' | 'schwarz' | 'rot' | 'gruen'

// buchstaben ordnet MSEs Symbole den Farben zu.
const buchstaben: Record<string, Farbe> = {
  W: 'weiss',
  U: 'blau',
  B: 'schwarz',
  R: 'rot',
  G: 'gruen',
}

// Kartenposten ist eine Karte mit ihrer Menge im Deck.
export interface Kartenposten {
  kosten: string
  menge: number

  // manawert ist die **nachgelesene** Zahl (Scryfalls `cmc`) — `undefined`,
  // wo es keine gibt: bei einer eigenen Karte rechnet dann [wertVon] aus den
  // Kosten.
  manawert?: number
}

// wertVon gibt den Manawert eines Postens — nachgelesen schlägt gerechnet.
//
// **Weil eine zweiseitige Karte leere Kosten hat und trotzdem etwas kostet.**
// Scryfall führt `mana_cost` bei ihr leer und den Wert in `cmc`; wer aus den
// Kosten rechnete, bekäme null (gesehen am 2026-09-09 in der Sichtprüfung zu
// U10).
export function wertVon(eintrag: Kartenposten): number {
  return eintrag.manawert ?? manawert(eintrag.kosten)
}

// wertBekannt sagt, ob der Manawert dieses Postens bekannt ist.
//
// **Unbekannt ist nicht null**, und das ist dieselbe Regel wie „ungeprüft ist
// nicht unzulässig" ([0061](../../../docs/entscheidungen/0061-was-ein-deck-hier-ist.md)).
// Zwei Fälle treffen sie: ein Platz auf einen Druck, den Scryfall nicht kennt
// — er hat weder Kosten noch Wert —, und eine **eigene** Karte, deren
// Kostenfeld niemand ausgefüllt hat. Beide als Nullkostenkarte zu zählen
// behauptete etwas, was niemand weiß; im Messdeck zog der erste Fall den
// Schnitt von 1,33 auf 0,92 (gesehen am 2026-09-09 in der Sichtprüfung).
//
// **Eine zweiseitige Karte fällt nicht darunter**: sie hat bei Scryfall leere
// Kosten und trotzdem ein `cmc`, und das steht in `manawert`.
export function wertBekannt(eintrag: Kartenposten): boolean {
  return eintrag.manawert !== undefined || eintrag.kosten.trim() !== ''
}

// manawert rechnet die Kosten in eine Zahl.
//
// **Ziffernfolgen zählen als Zahl, Buchstaben als eins.** `3WW` gibt 5. Ein
// `X` gibt null — es steht für einen Wert, den erst das Spiel bestimmt.
export function manawert(kosten: string): number {
  let wert = 0
  let zahl = ''

  for (const zeichen of kosten.toUpperCase()) {
    if (zeichen >= '0' && zeichen <= '9') {
      zahl += zeichen

      continue
    }

    if (zahl !== '') {
      wert += Number(zahl)
      zahl = ''
    }

    if (zeichen in buchstaben) {
      wert += 1
    }
  }

  return wert + (zahl === '' ? 0 : Number(zahl))
}

// farbenVon gibt die Farben, die in diesen Kosten vorkommen — jede einmal.
//
// **Jede einmal, auch wenn das Symbol zweimal steht.** `UU` ist eine blaue
// Karte, keine zwei; die Verteilung zählt Karten nach Farbe, nicht Symbole.
export function farbenVon(kosten: string): Farbe[] {
  const gefunden = new Set<Farbe>()

  for (const zeichen of kosten.toUpperCase()) {
    const farbe = buchstaben[zeichen]

    if (farbe !== undefined) {
      gefunden.add(farbe)
    }
  }

  return [...gefunden]
}

// kurve zählt die Karten je Manawert — aufsteigend, ohne Lücken dazwischen.
//
// **Ohne Lücken**, damit ein Balkenbild nicht lügt: fehlt die 3, steht dort
// eine Null und nicht der nächste Wert daneben.
export function kurve(posten: readonly Kartenposten[]): [number, number][] {
  const nach = new Map<number, number>()

  for (const eintrag of posten) {
    // **Die Kurve zählt nur, was sie kennt.** Hier stand die Auswahl beim
    // Aufrufer (`kosten !== ''`), und sie war zu grob: sie warf die
    // zweiseitige Karte — leere Kosten, **bekannter** Wert — mit dem
    // unbekannten Druck in einen Topf.
    //
    // **Länder hält der Aufrufer heraus**, denn ob eine Karte eines ist,
    // steht in ihrer Typzeile, und die kennt diese Rechnung nicht.
    if (!wertBekannt(eintrag)) {
      continue
    }

    const wert = wertVon(eintrag)

    nach.set(wert, (nach.get(wert) ?? 0) + eintrag.menge)
  }

  if (nach.size === 0) {
    return []
  }

  const hoechster = Math.max(...nach.keys())
  const zeilen: [number, number][] = []

  for (let wert = 0; wert <= hoechster; wert += 1) {
    zeilen.push([wert, nach.get(wert) ?? 0])
  }

  return zeilen
}

// farbverteilung zählt die Karten je Farbe.
//
// **Eine mehrfarbige Karte zählt in jeder ihrer Farben.** Die Summe ist damit
// größer als das Deck — und genau das will die Frage „wie viel Blau brauche
// ich" wissen.
export function farbverteilung(
  posten: readonly Kartenposten[],
): [Farbe, number][] {
  const reihenfolge: Farbe[] = ['weiss', 'blau', 'schwarz', 'rot', 'gruen']
  const nach = new Map<Farbe, number>()

  for (const eintrag of posten) {
    for (const farbe of farbenVon(eintrag.kosten)) {
      nach.set(farbe, (nach.get(farbe) ?? 0) + eintrag.menge)
    }
  }

  // **Ein `flatMap` und kein `filter` mit `get` dahinter.** Das Paar aus
  // `has` und `get` bräuchte einen Rückfall für einen Fall, den `has` gerade
  // ausgeschlossen hat — ein Zweig, den kein Aufruf erreicht (§3).
  return reihenfolge.flatMap((farbe): [Farbe, number][] => {
    const zahl = nach.get(farbe)

    return zahl === undefined ? [] : [[farbe, zahl]]
  })
}
