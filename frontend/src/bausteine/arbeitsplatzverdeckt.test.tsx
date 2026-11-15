// **Ein Fokus unter der Spalte schließt sie** — N12 aus dem Nachlauf über
// Phase 6.3 (#1664), verlagert in Phase 6.11 (#1672).
//
// Aus der Spalte über dem Zentrum führte Tab in Bedienelemente, die unter
// ihr lagen: der Fokus stand dann auf etwas, das niemand sah (WCAG 2.4.11),
// und Escape schloss die Spalte nicht mehr, weil er nicht in ihr stand.
// **Entschieden**: sie schließt, wenn der Fokus auf etwas geht, das sie
// verdeckt — und nur dann. Ein sichtbares Feld daneben lässt sie offen, und
// Escape in ihm gehört weiter dem Feld (M1, `arbeitsplatzescape.test.tsx`).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, it } from 'vitest'

import { Arbeitsplatz } from './Arbeitsplatz'
import { effekteNachholen } from '../test/effekte'
import { gibJsdomEineBreite } from '../test/sichtbreite'
import { Sprachraum } from '../sprache/kontext'

let raeumeBreite: () => void = () => undefined

beforeEach(() => {
  raeumeBreite = gibJsdomEineBreite(1000)
})

afterEach(() => {
  raeumeBreite()
})

// liegtBei gibt einem Element ein Rechteck — jsdom misst jedes mit null.
function liegtBei(element: Element, links: number, rechts: number) {
  Object.defineProperty(element, 'getBoundingClientRect', {
    configurable: true,
    value: () => new DOMRect(links, 0, rechts - links, 40),
  })
}

async function offen() {
  render(
    <Sprachraum sprache="de">
      <Arbeitsplatz
        bezeichnung="Kartenarbeitsplatz"
        navigator={{
          bezeichnung: 'Karten',
          inhalt: <button type="button">Eine Karte</button>,
        }}
        zentrum={
          <>
            <input aria-label="Unter der Spalte" />
            <input aria-label="Daneben" />
          </>
        }
        inspector={{ bezeichnung: 'Felder', inhalt: <p>Das Formular</p> }}
      />
    </Sprachraum>,
  )
  await effekteNachholen()
  await userEvent.click(
    screen.getByRole('button', { name: 'Navigator einblenden' }),
  )

  const spalte = screen.getByRole('complementary', { name: 'Karten' })
  liegtBei(spalte, 0, 240)
  liegtBei(screen.getByRole('textbox', { name: 'Unter der Spalte' }), 20, 200)
  liegtBei(screen.getByRole('textbox', { name: 'Daneben' }), 400, 600)

  return spalte
}

it('schließt die Spalte, wenn der Fokus unter sie geht, und lässt ihn dort', async () => {
  await offen()
  const darunter = screen.getByRole('textbox', { name: 'Unter der Spalte' })

  darunter.focus()
  await effekteNachholen()

  expect(screen.queryByRole('complementary', { name: 'Karten' })).toBeNull()
  expect(darunter).toHaveFocus()
})

it('lässt sie offen, wenn der Fokus auf ein sichtbares Feld daneben geht', async () => {
  const spalte = await offen()
  const daneben = screen.getByRole('textbox', { name: 'Daneben' })

  daneben.focus()
  await effekteNachholen()

  expect(spalte).toBeVisible()
  expect(daneben).toHaveFocus()
})

it('lässt sie offen, wenn der Fokus in ihr bleibt', async () => {
  const spalte = await offen()

  screen.getByRole('button', { name: 'Eine Karte' }).focus()
  await effekteNachholen()

  expect(spalte).toBeVisible()
})
