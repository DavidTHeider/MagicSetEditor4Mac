// **Eine Handelszeile nennt ihre Karte** — P10
// (#2378),
// Befund C-13: am Fenster stand „1 × counterspell-7ed".
//
// # Wogegen hier geprüft wird
//
// Name, Set, Nummer und Zustand je Zeile, das Bild, wo Scryfall es schon
// einmal beantwortet hat — und **ein eigener Druck ohne Bild** bekommt einen
// eigenen Platzhalter statt des Bildes einer anderen Karte. Die Kennung steht
// nur noch da, wo sonst nichts bekannt ist.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it } from 'vitest'

import { Anschlussraum } from '../anschlussraum'
import { Handelspositionen } from './Handelspositionen'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'

import type {
  Handelsdruck,
  Handelsposition,
  Preisanhalt,
  Preisanhaltantwort,
} from '../apihandelsleitung'

const texte = kataloge.de.seiten.handeln
const sammlung = kataloge.de.seiten.sammlung

afterEach(() => {
  raeumeApi()
})

const blitz = '00000000-0000-4000-8000-0000000000a1'
const sturm = '00000000-0000-4000-8000-0000000000b1'
const fremd = '00000000-0000-4000-8000-0000000000c1'

function zeile(von: string, printing: string, druckId = 0): Handelsposition {
  return {
    von,
    printing,
    druckId,
    menge: 2,
    guete: 'nearMint',
    sprache: 'de',
    finish: 'nonfoil',
    abbild: '',
    bewertung: 0,
  }
}

const leer: Handelsdruck = {
  name: '',
  setcode: '',
  nummer: '',
  bild: '',
  eigen: false,
}

function einAnhalt(drucke: Handelsdruck[]): Preisanhalt {
  return {
    zeilen: drucke.map((druck) => ({ preise: [], druck })),
    seiten: [],
    differenzen: [],
    hauptwaehrung: '',
  }
}

function zeige(positionen: Handelsposition[], drucke: Handelsdruck[]) {
  belegeApi({
    HandelPreisanhalt: () =>
      Promise.resolve({ anhalt: einAnhalt(drucke), fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Anschlussraum
        aktuell="handeln"
        wechseln={() => {
          /* hier wird nicht gewechselt */
        }}
      >
        <Handelspositionen
          ich="ich-1"
          gegenueber="du-1"
          positionen={positionen}
          bearbeitbar={false}
          setzePositionen={() => {
            /* nur lesen */
          }}
        />
      </Anschlussraum>
    </Sprachraum>,
  )
}

describe('eine Handelszeile', () => {
  it('nennt Name, Set, Nummer und Zustand und zeigt das Bild statt der Kennung', async () => {
    zeige(
      [zeile('ich-1', blitz), zeile('du-1', sturm)],
      [
        { ...leer, name: 'Blitzschlag', setcode: 'm10', nummer: '146' },
        {
          ...leer,
          name: 'Sturmrufer',
          setcode: 'zwe',
          nummer: '3',
          bild: 'https://cards.scryfall.io/small/sturmrufer.jpg',
        },
      ],
    )

    const meine = screen.getByRole('region', { name: texte.positionVonIch })
    const seine = screen.getByRole('region', {
      name: texte.positionVonGegenueber,
    })

    // **Bis der Anhalt da ist, steht keine Kennung** — ein Platzhalter.
    expect(within(meine).getByText(`2 × ${texte.druck.laedt}`)).toBeVisible()

    expect(await within(meine).findByText('2 × Blitzschlag')).toBeVisible()
    expect(
      within(meine).getByText(
        `M10 #146 · ${sammlung.zustaende.nearMint} · de · ${sammlung.nonfoil}`,
      ),
    ).toBeVisible()
    expect(within(seine).getByText('2 × Sturmrufer')).toBeVisible()

    // **Das Bild steht, wo es eines gibt** — schmückend, denn der Name
    // steht daneben (WCAG 1.1.1).
    const bild = seine.querySelector('img')
    expect(bild).toHaveAttribute(
      'src',
      'https://cards.scryfall.io/small/sturmrufer.jpg',
    )
    expect(bild).toHaveAttribute('alt', '')

    // **Keine Kennung im Hauptweg.**
    expect(screen.queryByText(new RegExp(blitz))).toBeNull()
    expect(screen.queryByText(new RegExp(sturm))).toBeNull()
  })

  it('nennt das Set allein, wenn die Nummer fehlt', async () => {
    zeige(
      [zeile('ich-1', blitz)],
      [{ ...leer, name: 'Blitzschlag', setcode: 'm10' }],
    )

    expect(
      await screen.findByText(
        `M10 · ${sammlung.zustaende.nearMint} · de · ${sammlung.nonfoil}`,
      ),
    ).toBeVisible()
  })

  it('zeigt einen eigenen Druck ohne Bild und nennt einen unbekannten', async () => {
    zeige(
      [zeile('ich-1', '', 3), zeile('du-1', '', 3), zeile('du-1', fremd)],
      [
        {
          ...leer,
          name: 'Eigener Wächter',
          setcode: 'set',
          nummer: '007',
          eigen: true,
        },
        { ...leer, eigen: true },
        leer,
      ],
    )

    const meine = screen.getByRole('region', { name: texte.positionVonIch })
    const seine = screen.getByRole('region', {
      name: texte.positionVonGegenueber,
    })

    expect(await within(meine).findByText('2 × Eigener Wächter')).toBeVisible()
    expect(
      within(meine).getByText(
        `${texte.druck.eigen} · SET #007 · ${sammlung.zustaende.nearMint} · de · ${sammlung.nonfoil}`,
      ),
    ).toBeVisible()

    // **Nie das Bild einer anderen Karte** — kein Bild, ein Platzhalter.
    expect(screen.queryByRole('img')).toBeNull()
    expect(document.querySelectorAll('img')).toHaveLength(0)

    const [eigenerFremder, unbekannter] = within(seine).getAllByRole('listitem')
    expect(eigenerFremder).toHaveTextContent(`2 × ${texte.druck.ohneNamen}`)
    expect(eigenerFremder).toHaveTextContent(texte.druck.eigen)

    // **Wo nichts bekannt ist, bleibt die Kennung** — eine Auskunft, und
    // keine erfundene.
    expect(unbekannter).toHaveTextContent(`2 × ${texte.druck.unbekannt}`)
    expect(unbekannter).toHaveTextContent(fremd)
  })

  it('behält die Namen, während der Anhalt neu fragt, und nennt nach einer Absage die Kennung', async () => {
    const antworten: ((wert: Preisanhaltantwort) => void)[] = []
    belegeApi({
      HandelPreisanhalt: () =>
        new Promise((aufloesen) => {
          antworten.push(aufloesen)
        }),
    })

    function Wechsel() {
      const [positionen, setzePositionen] = useState([zeile('ich-1', blitz)])

      return (
        <Sprachraum sprache="de">
          <Anschlussraum
            aktuell="handeln"
            wechseln={() => {
              /* hier wird nicht gewechselt */
            }}
          >
            <button
              type="button"
              onClick={() => {
                setzePositionen([zeile('du-1', sturm), zeile('ich-1', blitz)])
              }}
            >
              wechseln
            </button>
            <Handelspositionen
              ich="ich-1"
              gegenueber="du-1"
              positionen={positionen}
              bearbeitbar={false}
              setzePositionen={setzePositionen}
            />
          </Anschlussraum>
        </Sprachraum>
      )
    }

    render(<Wechsel />)
    antworten[0]?.({
      anhalt: einAnhalt([
        { ...leer, name: 'Blitzschlag', setcode: 'm10', nummer: '146' },
      ]),
      fehler: null,
    })
    expect(await screen.findByText('2 × Blitzschlag')).toBeVisible()

    // **Eine Zeile kommt davor** — die Stelle der alten verschiebt sich, ihr
    // Name bleibt, und die neue wartet.
    await userEvent.click(screen.getByRole('button', { name: 'wechseln' }))
    expect(screen.getByText('2 × Blitzschlag')).toBeVisible()
    expect(screen.getByText(`2 × ${texte.druck.laedt}`)).toBeVisible()

    // **Eine Absage lässt keinen Platzhalter stehen**: die neue Zeile nennt
    // ihre Kennung.
    antworten[1]?.({
      anhalt: { zeilen: [], seiten: [], differenzen: [], hauptwaehrung: '' },
      fehler: {
        code: 'unbekannt',
        meldung: 'kaputt',
        schluessel: '',
        feld: '',
      },
    })
    expect(
      await screen.findByText(`2 × ${texte.druck.unbekannt}`),
    ).toBeVisible()
    expect(screen.getByText(new RegExp(sturm))).toBeVisible()
    expect(screen.getByText('2 × Blitzschlag')).toBeVisible()
  })
})
