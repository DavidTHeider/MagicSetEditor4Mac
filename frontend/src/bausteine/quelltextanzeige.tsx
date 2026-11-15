// **Die Anzeige unter der Eingabe** — abgetrennt am 2026-08-26, weil
// `Quelltextfeld.tsx` über die Zeilengrenze aus §2 ging.
//
// # Der Schnitt folgt der Frage
//
// Hier steht, **wie ein Text aussieht**: Zeilen, gefärbte Abschnitte, die
// Fehlermarke und der hervorgehobene Fund. Daneben in `Quelltextfeld.tsx`
// steht, **was das Feld tut** — Farben holen, die Suche aufschlagen, den
// Fokus führen.
import { farbe } from './farben'

import type { Klammerpaar } from './klammerpaare'
import type { Merkmalstelle, Skriptstelle } from '../api'

// Bereich ist ein Stück Text von einer Stelle bis zu einer anderen.
export interface Bereich {
  von: number
  bis: number
}

// abschnitte zerlegt den Text in gefärbte Stücke.
//
// **Über die Stellen und nicht über eine zweite Zerlegung.** Was zwischen zwei
// Merkmalen steht — Leerzeichen, Umbrüche, Unlesbares —, bleibt ungefärbt
// stehen; ohne diesen Rest läge jedes Zeichen hinter dem ersten Merkmal neben
// seiner Farbe.
export function abschnitte(
  wert: string,
  merkmale: Merkmalstelle[],
  stelle: Skriptstelle | null,
  treffer: Bereich | null,
  paar: Klammerpaar | null,
) {
  const zeilen = wert.split('\n')

  // **Der Fund wird in Zeile und Spalte umgerechnet**, weil die Anzeige nach
  // Zeilen aufgebaut ist. Ein Bereich über mehrere Zeilen kommt nicht vor: das
  // Suchfeld ist einzeilig.
  const fund = trefferstelle(wert, treffer)

  return zeilen.map((zeile, nummer) => (
    // Auch hier ist der Index die Stelle selbst.
    //
    // **`relative` je Zeile**, damit die Fehlermarke ihren Bezugspunkt in
    // **ihrer** Zeile hat. Ohne das säße sie am oberen Rand des ganzen Felds
    // und zeigte immer auf die erste.
    <div key={nummer} className="relative">
      {zeilenteile(zeile, nummer + 1, merkmale, stelle, fund, paar)}
    </div>
  ))
}

// zeilenteile färbt eine Zeile.
function zeilenteile(
  zeile: string,
  nummer: number,
  merkmale: Merkmalstelle[],
  stelle: Skriptstelle | null,
  fund: Fund | null,
  paar: Klammerpaar | null,
) {
  const teile = []
  let bei = 0

  for (const merkmal of merkmale.filter(
    (eintrag) => eintrag.zeile === nummer,
  )) {
    const anfang = merkmal.spalte - 1

    if (anfang < bei || anfang >= zeile.length) {
      continue
    }

    teile.push(zeile.slice(bei, anfang))
    teile.push(
      <span
        key={String(nummer) + '/' + String(merkmal.spalte)}
        className={farbe(merkmal.art)}
      >
        {zeile.slice(anfang, anfang + merkmal.breite)}
      </span>,
    )
    bei = anfang + merkmal.breite
  }

  teile.push(zeile.slice(bei))

  // **Eine leere Zeile braucht ein Zeichen**, sonst fällt sie in der Anzeige
  // zusammen und jede Nummer darunter zeigt auf die falsche Zeile.
  if (zeile === '') {
    teile.push(' ')
  }

  if (fund !== null && fund.zeile === nummer) {
    teile.push(fundmarke(zeile, fund))
  }

  teile.push(...klammermarken(zeile, nummer, paar))

  return stelle !== null && stelle.zeile === nummer
    ? [...teile, marke(zeile, stelle)]
    : teile
}

// marke streicht die gemeldete Spalte an.
//
// **Unter dem Zeichen und nicht daneben**: der Zeiger liegt in einer eigenen
// Ebene über der Zeile, damit er den Text nicht verschiebt.
function marke(zeile: string, stelle: Skriptstelle) {
  const davor = zeile
    .slice(0, Math.max(0, stelle.spalte - 1))
    .replace(/[^\t]/g, ' ')

  return (
    <span
      key="marke"
      data-testid="fehlermarke"
      className="absolute left-0 rounded-sm bg-fehler/25"
      style={{ marginLeft: `${String(davor.length)}ch` }}
    >
      {/* **Ein geschütztes Leerzeichen und nicht nichts.** Die Stelle
          hinter dem letzten Zeichen ist der häufigste Fall — „hier fehlt
          ein Wert" zeigt genau dorthin. Ein Zeiger ohne Inhalt wäre null
          Pixel breit. */}
      {zeile.slice(stelle.spalte - 1, stelle.spalte) || ' '}
    </span>
  )
}

// Fund ist derselbe Bereich in Zeile und Spalte.
interface Fund {
  zeile: number
  spalte: number
  breite: number
}

// trefferstelle rechnet einen Bereich in Zeile und Spalte um.
//
// **Ohne Schleife und damit ohne toten Rückweg.** Der erste Anlauf lief über
// die Zeilen und brauchte danach ein `return` für den Fall, dass keine passt —
// den kein Lauf betreten kann, weil der Bereich aus `indexOf` über denselben
// Text kommt. Ein Zweig ohne Fall gehört weg (§3).
//
// `lastIndexOf` gibt −1, wenn vor der Stelle kein Umbruch steht; die Spalte
// ist dann `von + 1`, und das ist genau die erste Zeile.
function trefferstelle(wert: string, treffer: Bereich | null): Fund | null {
  if (treffer === null) {
    return null
  }

  const davor = wert.slice(0, treffer.von)

  return {
    zeile: davor.split('\n').length,
    spalte: treffer.von - davor.lastIndexOf('\n'),
    breite: treffer.bis - treffer.von,
  }
}

// fundmarke hebt den gefundenen Text hervor.
//
// **Über der Zeile und nicht darin**: eine Hervorhebung zwischen den
// Abschnitten verschöbe sie, und die Farbe stünde neben ihrem Zeichen.
function fundmarke(zeile: string, fund: Fund) {
  const davor = zeile
    .slice(0, Math.max(0, fund.spalte - 1))
    .replace(/[^\t]/g, ' ')

  return (
    <span
      key="fund"
      data-testid="fundmarke"
      className="absolute left-0 rounded-sm bg-marke/25/25"
      style={{ marginLeft: `${String(davor.length)}ch` }}
    >
      {zeile.slice(fund.spalte - 1, fund.spalte - 1 + fund.breite)}
    </span>
  )
}

// klammermarken hebt das Klammerpaar unter dem Schreibzeiger hervor.
//
// **Beide oder eine.** Steht der Partner in derselben Zeile, stehen zwei
// Marken darin; steht er in einer anderen, je eine dort. Fehlt er, bleibt es
// bei einer — und sie ist rot.
//
// **Rot und nicht bloß blass**: eine Klammer ohne Partner ist derselbe Befund,
// den der Zerteiler später meldet, nur früher. Wer sie in derselben Farbe
// zeigte wie ein gefundenes Paar, nähme der Anzeige ihren einzigen Nutzen.
function klammermarken(
  zeile: string,
  nummer: number,
  paar: Klammerpaar | null,
) {
  if (paar === null) {
    return []
  }

  const gefunden = paar.partner !== null

  // **`flatMap` und nicht `filter`.** TypeScript liest aus einem `filter`
  // keinen engeren Typ heraus — `partner` bliebe `Merkmalstelle | null`, und
  // die Marke bekäme ein `null`, von dem sie weiß, dass es keines ist.
  return [paar.hier, paar.partner].flatMap((klammer) =>
    klammer !== null && klammer.zeile === nummer
      ? [klammermarke(zeile, klammer, gefunden)]
      : [],
  )
}

// klammermarke streicht **eine** Klammer an.
//
// Dieselbe Bauart wie [marke] und [fundmarke]: eine eigene Ebene über der
// Zeile, damit die Hervorhebung den Text nicht verschiebt.
function klammermarke(
  zeile: string,
  klammer: Merkmalstelle,
  gefunden: boolean,
) {
  const davor = zeile
    .slice(0, Math.max(0, klammer.spalte - 1))
    .replace(/[^\t]/g, ' ')

  return (
    <span
      key={'klammer/' + String(klammer.zeile) + '/' + String(klammer.spalte)}
      data-testid={gefunden ? 'klammermarke' : 'klammerohnepartner'}
      className={
        gefunden
          ? 'absolute left-0 rounded-sm bg-erfolg/30'
          : 'absolute left-0 rounded-sm bg-fehler/40'
      }
      style={{ marginLeft: `${String(davor.length)}ch` }}
    >
      {zeile.slice(klammer.spalte - 1, klammer.spalte)}
    </span>
  )
}
