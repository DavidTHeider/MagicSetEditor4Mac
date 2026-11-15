// **Suchen, Ordnen und Filtern in der Partienliste** — PLAY-03.AK-05
// (#1711).
//
// **Die Rechnung ohne einen Baum**; der Lauf daneben
// (`partienliste.test.tsx`) prüft, was davon zu sehen ist.
import { describe, expect, it } from 'vitest'

import { partienauswahl, partienname, staendeIn } from './partienauswahl'

import type { Partiezeile } from '../apipartie'

// eineZeile gibt eine Partie, wie die Liste sie bekommt.
function eineZeile(abweichend: Partial<Partiezeile> = {}): Partiezeile {
  return {
    id: 'p1',
    fassung: '3',
    spiel: 'magic',
    stand: 'laeuft',
    sitze: [
      { kennung: 'sitz-1', name: 'David' },
      { kennung: 'sitz-2', name: 'Gegenüber' },
    ],
    objekte: 12,
    zug: 0,
    eigenerSitz: '',
    dran: false,
    ...abweichend,
  }
}

const alles = { frage: '', stand: '', ordnung: 'zuletzt' } as const

// kennungen nennt die Partien einer Auswahl beim Namen ihrer Kennung.
const kennungen = (zeilen: readonly Partiezeile[]) =>
  zeilen.map((zeile) => zeile.id)

describe('partienname', () => {
  it('nennt die Partie an ihren Sitznamen', () => {
    expect(partienname(eineZeile())).toBe('David · Gegenüber')
  })

  // **Eine Partie ohne Namen steht mit ihrer Kennung da** — sie ist immer
  // noch eine, und eine leere Zeile wäre keine Auskunft (§8).
  it('fällt auf die Kennung zurück', () => {
    expect(partienname(eineZeile({ sitze: [] }))).toBe('p1')
  })
})

describe('die Suche', () => {
  const drei = [
    eineZeile({ id: 'p1' }),
    eineZeile({
      id: 'p2',
      spiel: 'vs',
      sitze: [{ kennung: 's', name: 'Bert' }],
    }),
    eineZeile({ id: 'p3', sitze: [] }),
  ]

  // **Eine leere Frage filtert nicht.**
  //
  // **Und das ist der Unterschied zur Einstellungssuche** (U11, #1422): dort
  // ist die Suche der Einstieg, und eine leere Frage trifft nichts. Hier ist
  // die Liste der Inhalt der Fläche — wer nichts eingibt, will alles sehen.
  it('gibt alles, solange niemand fragt', () => {
    expect(kennungen(partienauswahl(drei, alles, 'de'))).toEqual([
      'p1',
      'p2',
      'p3',
    ])
  })

  it('findet einen Sitznamen, gleich wie geschrieben', () => {
    expect(
      kennungen(partienauswahl(drei, { ...alles, frage: 'bert' }, 'de')),
    ).toEqual(['p2'])
  })

  it('findet das Spiel', () => {
    expect(
      kennungen(partienauswahl(drei, { ...alles, frage: 'vs' }, 'de')),
    ).toEqual(['p2'])
  })

  // **Eine namenlose Partie muss auffindbar bleiben** — ihre Kennung ist das
  // Einzige, was sie hat, und sie steht deshalb auch da.
  it('findet eine namenlose Partie an ihrer Kennung', () => {
    expect(
      kennungen(partienauswahl(drei, { ...alles, frage: 'p3' }, 'de')),
    ).toEqual(['p3'])
  })

  // **Gesucht wird, was dasteht.** Eine Sitzkennung steht nur da, wo der
  // Name fehlt — sonst wäre sie ein unsichtbarer Treffer, und niemand sähe,
  // warum die Zeile bleibt.
  it('findet keine Sitzkennung, die nirgends steht', () => {
    expect(partienauswahl(drei, { ...alles, frage: 'sitz-2' }, 'de')).toEqual(
      [],
    )
  })

  it('gibt nichts für eine Frage, die niemand trifft', () => {
    expect(partienauswahl(drei, { ...alles, frage: 'zzz' }, 'de')).toEqual([])
  })
})

describe('der Standfilter', () => {
  const drei = [
    eineZeile({ id: 'p1', stand: 'laeuft' }),
    eineZeile({ id: 'p2', stand: 'pausiert' }),
    eineZeile({ id: 'p3', stand: 'beendet' }),
  ]

  it('lässt ohne Wahl alles stehen', () => {
    expect(kennungen(partienauswahl(drei, alles, 'de'))).toEqual([
      'p1',
      'p2',
      'p3',
    ])
  })

  it('zeigt nur den gewählten Stand', () => {
    expect(
      kennungen(partienauswahl(drei, { ...alles, stand: 'pausiert' }, 'de')),
    ).toEqual(['p2'])
  })

  // **Filter und Suche gelten zusammen** — ohne diese Zeile wäre der Lauf
  // auch grün, wenn der zweite den ersten überschriebe (§10).
  it('gilt zusammen mit der Suche', () => {
    const viele = [
      ...drei,
      eineZeile({
        id: 'p4',
        stand: 'pausiert',
        sitze: [{ kennung: 's', name: 'Bert' }],
      }),
    ]

    expect(
      kennungen(
        partienauswahl(
          viele,
          { ...alles, stand: 'pausiert', frage: 'bert' },
          'de',
        ),
      ),
    ).toEqual(['p4'])
  })
})

describe('die Ordnung', () => {
  // **`zuletzt` ist die Reihenfolge, in der die Liste ankommt.**
  //
  // **Und nicht eine gerechnete**: `Partiezeile` trägt kein Datum, und
  // `fassung` zählt Ereignisse und keine Zeit — eine Partie mit vielen
  // Zügen ist nicht die jüngere. Der Bestand ordnet nach Änderungszeit
  // ([app.Partien]), und diese Ordnung lässt sie dabei.
  const dreiDurcheinander = [
    eineZeile({ id: 'p1', zug: 4, sitze: [{ kennung: 's', name: 'Cora' }] }),
    eineZeile({ id: 'p2', zug: 9, sitze: [{ kennung: 's', name: 'Anna' }] }),
    eineZeile({ id: 'p3', zug: 4, sitze: [{ kennung: 's', name: 'Bert' }] }),
  ]

  it('lässt die gegebene Reihenfolge, wo nichts anderes gewählt ist', () => {
    expect(kennungen(partienauswahl(dreiDurcheinander, alles, 'de'))).toEqual([
      'p1',
      'p2',
      'p3',
    ])
  })

  it('ordnet nach dem Namen, der dasteht', () => {
    expect(
      kennungen(
        partienauswahl(dreiDurcheinander, { ...alles, ordnung: 'name' }, 'de'),
      ),
    ).toEqual(['p2', 'p3', 'p1'])
  })

  // **Der weiteste Zug zuerst** — wer nach dem Zug ordnet, sucht die Partie,
  // in der am meisten passiert ist.
  it('ordnet nach dem Zug, den weitesten zuerst', () => {
    expect(
      kennungen(
        partienauswahl(dreiDurcheinander, { ...alles, ordnung: 'zug' }, 'de'),
      ),
    ).toEqual(['p2', 'p1', 'p3'])
  })

  // **Bei gleichem Zug bleibt die gegebene Reihenfolge** — sonst
  // sprängen zwei gleich weite Partien bei jedem Zeichnen umeinander.
  it('lässt gleiche Züge in ihrer Reihenfolge', () => {
    const gleich = [
      eineZeile({ id: 'p1', zug: 4 }),
      eineZeile({ id: 'p2', zug: 4 }),
    ]

    expect(
      kennungen(partienauswahl(gleich, { ...alles, ordnung: 'zug' }, 'de')),
    ).toEqual(['p1', 'p2'])
  })

  // **Geordnet wird nicht an Ort und Stelle.** Die Liste kommt aus einem
  // Haken, und wer sie umsortierte, ordnete den Zustand eines anderen um.
  it('lässt die übergebene Liste unberührt', () => {
    const gegeben = [...dreiDurcheinander]

    partienauswahl(gegeben, { ...alles, ordnung: 'zug' }, 'de')

    expect(kennungen(gegeben)).toEqual(['p1', 'p2', 'p3'])
  })
})

describe('staendeIn', () => {
  // **Was nicht dasteht, wird nicht angeboten.** Ein Filter auf einen Stand,
  // den keine Partie hat, führte in eine leere Liste — und niemand sähe,
  // warum.
  //
  // **Geordnet nach dem Lebenslauf einer Partie und nicht nach dem
  // Alphabet**: „läuft, pausiert, beendet" ist die Reihenfolge, in der man
  // sie durchläuft. Alphabetisch stünde „beendet" vorn, und das wäre der
  // Stand, den am wenigsten jemand sucht.
  it('nennt jeden vorkommenden Stand einmal, im Lauf einer Partie', () => {
    expect(
      staendeIn([
        eineZeile({ stand: 'beendet' }),
        eineZeile({ stand: 'pausiert' }),
        eineZeile({ stand: 'laeuft' }),
        eineZeile({ stand: 'pausiert' }),
      ]),
    ).toEqual(['laeuft', 'pausiert', 'beendet'])
  })

  // **Ein unbekannter Stand fällt nicht weg** — fremde Daten sind der
  // Normalfall (§8), und ein Kern, der einen vierten einführte, ließe die
  // Liste sonst mit einem Filter dastehen, der ihn verschweigt.
  it('hängt an, was es noch nicht gibt', () => {
    expect(
      staendeIn([eineZeile({ stand: 'verschollen' }), eineZeile()]),
    ).toEqual(['laeuft', 'verschollen'])
  })

  it('nennt für eine leere Liste nichts', () => {
    expect(staendeIn([])).toEqual([])
  })
})
