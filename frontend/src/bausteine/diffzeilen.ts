// **Der Zeilen-Diff vor dem Speichern** — Werkstück U7 von Phase 5.1
// (#1418).
//
// # Wofür
//
// Eine Vorlage gilt für jedes Set, das ihr Paket benutzt. Wer speichert,
// soll vorher sehen, **was** er speichert — nicht den ganzen Text, sondern
// die Zeilen, die anders sind als der gelesene Stand.
//
// # Warum eine längste gemeinsame Teilfolge
//
// Ein Vergleich Zeile für Zeile sähe nach einer eingefügten Zeile jede
// weitere als geändert. Die längste gemeinsame Teilfolge hält, was gleich
// geblieben ist, und zeigt nur, was dazwischen wich — dieselbe Idee wie bei
// `diff`, ohne dessen Heuristiken: die Texte sind hunderte, nicht
// hunderttausende Zeilen, und eine Tafel aus Zeilen mal Zeilen ist bezahlbar.
//
// # Warum hier und nicht in Go
//
// Weil es Darstellung ist, keine Fachlogik: das Feld hat beide Stände in der
// Hand, und eine Bindung dafür trüge nur Text hin und zurück. **new** — MSE2
// hat keinen Vorlageneditor (gesucht in `src/gui/`, Fassung f9d9356 vom
// 2021-07-06).

export type Diffart = 'gleich' | 'weg' | 'neu'

// Diffzeile ist eine Zeile mit ihrer Nummer im alten und im neuen Stand.
export interface Diffzeile {
  art: Diffart
  text: string
  alt?: number
  neu?: number
}

// zeilen zerlegt einen Text — ein abschließender Umbruch ist keine Zeile.
function zeilen(text: string): string[] {
  const teile = text.split('\n')

  if (teile.length > 1 && teile[teile.length - 1] === '') {
    teile.pop()
  }

  return teile
}

// zeilendiff vergleicht zwei Stände Zeile für Zeile.
export function zeilendiff(alt: string, neu: string): Diffzeile[] {
  const a = zeilen(alt)
  const b = zeilen(neu)
  const breite = b.length + 1

  // laenge[i * breite + j] ist die Länge der längsten gemeinsamen Teilfolge
  // von a[0..i) und b[0..j).
  const laenge = new Uint32Array((a.length + 1) * breite)

  // bei liest die Tafel — `Number`, weil ein Feld einer typisierten Liste
  // nie fehlt und ein `?? 0` einen Zweig anlegte, den kein Lauf betritt.
  const bei = (i: number, j: number): number => Number(laenge[i * breite + j])

  for (let i = 1; i <= a.length; i += 1) {
    for (let j = 1; j <= b.length; j += 1) {
      laenge[i * breite + j] =
        a[i - 1] === b[j - 1]
          ? bei(i - 1, j - 1) + 1
          : Math.max(bei(i - 1, j), bei(i, j - 1))
    }
  }

  // Rückwärts vom Ende: was gleich ist, bleibt; sonst geht die Seite mit der
  // längeren Teilfolge — bei Gleichstand zuerst die neue, damit vorwärts
  // gelesen das Entfernte vor dem Hinzugekommenen steht, wie bei `diff`.
  //
  // **`String(…)` statt `?? ''`** aus demselben Grund wie bei `bei`: die
  // Zeile existiert, sobald ihr Zähler über null steht.
  const rueckwaerts: Diffzeile[] = []
  let i = a.length
  let j = b.length

  while (i > 0 || j > 0) {
    if (i > 0 && j > 0 && a[i - 1] === b[j - 1]) {
      rueckwaerts.push({
        art: 'gleich',
        text: String(a[i - 1]),
        alt: i,
        neu: j,
      })
      i -= 1
      j -= 1
    } else if (j > 0 && (i === 0 || bei(i, j - 1) >= bei(i - 1, j))) {
      rueckwaerts.push({ art: 'neu', text: String(b[j - 1]), neu: j })
      j -= 1
    } else {
      rueckwaerts.push({ art: 'weg', text: String(a[i - 1]), alt: i })
      i -= 1
    }
  }

  return rueckwaerts.reverse()
}

// huenke gruppiert die geänderten Zeilen samt `kontext` Zeilen davor und
// danach — was weiter weg liegt, hat niemand geändert und muss niemand lesen.
export function huenke(diff: Diffzeile[], kontext: number): Diffzeile[][] {
  const gezeigt = new Array<boolean>(diff.length).fill(false)

  diff.forEach((zeile, stelle) => {
    if (zeile.art === 'gleich') {
      return
    }

    for (
      let nachbar = Math.max(0, stelle - kontext);
      nachbar <= Math.min(diff.length - 1, stelle + kontext);
      nachbar += 1
    ) {
      gezeigt[nachbar] = true
    }
  })

  const gruppen: Diffzeile[][] = []
  let offen: Diffzeile[] | null = null

  diff.forEach((zeile, stelle) => {
    if (!gezeigt[stelle]) {
      offen = null

      return
    }

    if (offen === null) {
      offen = []
      gruppen.push(offen)
    }

    offen.push(zeile)
  })

  return gruppen
}
