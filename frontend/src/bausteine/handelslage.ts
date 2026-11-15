// **Wo ein Handel für mich steht, in einem Satz** — TRADE-01
// (#1715).
//
// **Die Lage kommt vom Dienst und wird hier nicht berechnet** — sie hängt an
// Stimmen und Rollen, und eine zweite Rechnung im Fenster liefe beim
// nächsten Übergang still auseinander.
import { einsetzen } from '../sprache/einsetzen'
import { standsatz } from './handelsstaende'

import type { Handelsvorgangzeile } from '../apihandelsverzeichnis'
import type { Katalog } from '../sprache/kataloge'

// partnername nennt die Gegenüber einer Zeile — mit Namen, wo einer bekannt
// ist, sonst mit dem Anfang der Kennung.
export function partnername(zeile: Handelsvorgangzeile): string {
  return zeile.gegenueber
    .map((partner) =>
      partner.name === '' ? partner.kennung.slice(0, 12) : partner.name,
    )
    .join(', ')
}

// lagesatz gibt den Satz zur Lage einer Zeile.
export function lagesatz(
  zeile: Handelsvorgangzeile,
  handeln: Katalog['seiten']['handeln'],
): string {
  return lagesatzFuer(
    {
      lage: zeile.lage,
      naechster: zeile.naechster,
      zustand: zeile.vorgang.zustand,
    },
    partnername(zeile),
    handeln,
  )
}

// Lageangabe ist, was ein Satz über die Lage braucht — aus einer Zeile der
// Liste oder aus einem aufgeschlagenen Vorgang (TRADE-02, #1716).
export interface Lageangabe {
  lage: string
  naechster: string
  zustand: string
}

// lagesatzFuer sagt dasselbe für jeden, der eine Lage trägt.
export function lagesatzFuer(
  stand: Lageangabe,
  name: string,
  handeln: Katalog['seiten']['handeln'],
): string {
  const eigene = handeln.liste

  switch (stand.lage) {
    case 'dran': {
      const schritt = (eigene.schritte as Record<string, string | undefined>)[
        stand.naechster
      ]

      return schritt === undefined
        ? eigene.lageDranOhneSchritt
        : einsetzen(eigene.lageDran, { schritt })
    }

    case 'wartet':
      return einsetzen(eigene.lageWartet, { name })

    case 'abgeschlossen':
      return eigene.lageAbgeschlossen
  }

  return einsetzen(eigene.lageAbgebrochen, {
    zustand: standsatz(handeln, stand.zustand),
  })
}
