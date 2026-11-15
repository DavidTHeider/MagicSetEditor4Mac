// **Wie weit der Zug ist, und wer gegen wen kämpft** — Artboard *Tisch*:
// unter „Zug 7 · Kampf · Blocker" ein Balken, neben dem Stapel der Kampf in
// Zeilen.
//
// **Eine eigene Datei**, weil `zugleiste.test.tsx` fragt, wer dran ist; hier
// steht, was der Kasten über den Zug **zeigt**.
import { render, screen, within } from '@testing-library/react'
import { expect, it } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Zugleiste } from './Zugleiste'
import { einObjekt, einSitz, einZug, einePartie } from '../seiten/partieprobe'
import { kataloge } from '../sprache/kataloge'

import type { Kampfpaar } from '../apizug'

const eigene = kataloge.de.seiten.spielen
const nichts = () => undefined

// zeige zeichnet die Leiste über einer Partie im Kampf.
function zeige(
  zug: Parameters<typeof einZug>[0],
  angriffe: Kampfpaar[] = [],
  bloecke: Kampfpaar[] = [],
) {
  const kreatur = (kennung: string, name: string, wer: string) =>
    einObjekt({
      kennung,
      name,
      zone: 'tisch',
      eigner: wer,
      beherrscher: wer,
      geheim: false,
    })

  render(
    <Sprachraum sprache="de">
      <Zugleiste
        partie={einePartie({
          sitze: [
            einSitz({ kennung: 'sitz1', name: 'Anna' }),
            einSitz({ kennung: 'sitz2', name: 'Bert' }),
          ],
          objekte: [
            kreatur('a1', 'Aschenwächter', 'sitz1'),
            kreatur('a2', 'Irrlichtweber', 'sitz1'),
            kreatur('a3', 'Glutwind', 'sitz1'),
            kreatur('a4', 'Eisenmoos', 'sitz1'),
            kreatur('b1', 'Hallenwart', 'sitz2'),
          ],
          zug: einZug({
            nummer: 7,
            aktiv: 'sitz1',
            prioritaet: 'sitz1',
            ...zug,
            angriffe,
            bloecke,
          }),
        })}
        erlaubt={[]}
        laeuft={false}
        passen={nichts}
        abkuerzen={nichts}
        zugende={nichts}
      />
    </Sprachraum>,
  )
}

it('zeigt als Balken, wie weit der Zug ist — mit dem Abschnitt als Namen', () => {
  zeige({ phase: 'kampf', schritt: 'blocker' })

  const balken = screen.getByRole('progressbar', {
    name: 'Kampfphase · Blocker',
  })

  expect(balken).toHaveAttribute('aria-valuenow', '7')
  expect(balken).toHaveAttribute('aria-valuemax', '12')
  expect(balken.firstElementChild).toHaveStyle({ width: '58%' })
})

// **Ein Abschnitt, den das Fenster nicht kennt, bekommt keinen Balken** —
// die Zeile darüber nennt ihn trotzdem, mit seiner Kennung.
it('zeichnet keinen Balken für einen unbekannten Abschnitt', () => {
  zeige({ phase: 'kampf', schritt: 'nachspiel' })

  expect(screen.queryByRole('progressbar')).not.toBeInTheDocument()
})

it('nennt den Kampf in Zeilen — geblockte einzeln, viele ungeblockte zusammen', () => {
  zeige(
    { phase: 'kampf', schritt: 'blocker' },
    [
      { kreatur: 'a1', ziel: 'sitz2' },
      { kreatur: 'a2', ziel: 'sitz2' },
      { kreatur: 'a3', ziel: 'sitz2' },
      { kreatur: 'a4', ziel: 'sitz2' },
    ],
    [{ kreatur: 'b1', angreifer: 'a1' }],
  )

  const kampf = screen.getByRole('list', { name: eigene.kampfTitel })
  const zeilen = within(kampf)
    .getAllByRole('listitem')
    .map((zeile) => zeile.textContent)

  expect(zeilen).toEqual([
    'Aschenwächter → Hallenwart',
    '3 weitere ungeblockt → Bert',
  ])
})

// **Wenige ungeblockte stehen einzeln da** — eine Zahl lohnt sich erst, wenn
// die Namen die Spalte füllten.
it('nennt wenige ungeblockte beim Namen', () => {
  zeige({ phase: 'kampf', schritt: 'angreifer' }, [
    { kreatur: 'a2', ziel: 'sitz2' },
    { kreatur: 'a3', ziel: 'sitz2' },
  ])

  const kampf = screen.getByRole('list', { name: eigene.kampfTitel })

  expect(
    within(kampf)
      .getAllByRole('listitem')
      .map((zeile) => zeile.textContent),
  ).toEqual(['Irrlichtweber → Bert', 'Glutwind → Bert'])
})

it('fasst viele ungeblockte ohne geblockte davor ohne „weitere“ zusammen', () => {
  zeige({ phase: 'kampf', schritt: 'angreifer' }, [
    { kreatur: 'a2', ziel: 'sitz2' },
    { kreatur: 'a3', ziel: 'sitz2' },
    { kreatur: 'a4', ziel: 'sitz2' },
  ])

  expect(screen.getByRole('listitem')).toHaveTextContent('3 ungeblockt → Bert')
})

it('nennt mehrere Blocker an einem Angreifer', () => {
  zeige(
    { phase: 'kampf', schritt: 'blocker' },
    [{ kreatur: 'a1', ziel: 'sitz2' }],
    [
      { kreatur: 'b1', angreifer: 'a1' },
      { kreatur: 'a4', angreifer: 'a1' },
    ],
  )

  expect(screen.getByRole('listitem')).toHaveTextContent(
    'Aschenwächter → Hallenwart und Eisenmoos',
  )
})

it('schweigt über den Kampf, solange niemand angreift', () => {
  zeige({ phase: 'hauptVorKampf', schritt: '' })

  expect(
    screen.queryByRole('list', { name: eigene.kampfTitel }),
  ).not.toBeInTheDocument()
})
