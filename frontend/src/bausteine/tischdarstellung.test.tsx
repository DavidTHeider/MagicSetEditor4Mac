// **Im schmalen Fenster bleibt die Tischleiste eine Zeile** — #2526
// (Issue (#2526)),
// Entscheidung des Auftraggebers vom 2026-10-20.
//
// # Wogegen
//
// Bei 1024 × 700 brach die Leiste in zwei Zeilen um, und weil sie im Tischort
// liegt, fraß die zweite Zeile die Höhe, die dem Zug darunter fehlte: er lag
// 95 Punkte unter der Falz (F09, `play06-tisch`). Jetzt stehen dort, wo Tisch
// und Zug untereinander stehen, nur die vier Knöpfe des Blicks in der Zeile;
// Ziel, Hintergrund, eigenes Bild und die Ausgabe des Tischbilds stehen hinter
// **„Darstellung"**. **Das Ziel auch**: gemessen bei 1024 × 700 rutschte
// „Darstellung" sonst allein in die zweite Zeile. Daneben, wo die Spalte neben
// dem Tisch steht, bleibt die Leiste, wie sie war.
//
// # Was jsdom nicht sieht
//
// **Die Container-Abfrage selbst** — jsdom rechnet kein CSS, und beide Lagen
// stehen hier im Baum. Geprüft wird, dass die Lagen an **derselben** Schwelle
// umschalten wie die Spalten (`@4xl`, `Partietisch.tsx`); dass die Leiste bei
// 1024 Punkten eine Zeile bleibt, misst `play06-tisch` am nativen Fenster.
import { fireEvent, render, screen, within } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import { einePartie } from '../seiten/partieprobe'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Tischflaeche } from './Tischflaeche'

import type { Tischmalerei } from '../tisch/maler'
import { tischplan } from '../tisch/tischplan'
import { einMaler } from '../test/malerprobe'

const eigene = kataloge.de.seiten.spielen

function zeige(teile: { hintergrund?: string } = {}) {
  const aufTischbild = vi.fn()
  const eigenesBild = vi.fn()
  const eigenesBildWeg = vi.fn()
  const maler = einMaler({ breite: 800, hoehe: 600 })

  render(
    <Sprachraum sprache="de">
      <Tischflaeche
        plan={tischplan(einePartie({ objekte: [] }))}
        bild={() => undefined}
        malerei={() => Promise.resolve<Tischmalerei>({ art: 'da', maler })}
        tischbild="holz"
        aufTischbild={aufTischbild}
        eigenesBild={eigenesBild}
        eigenesBildWeg={eigenesBildWeg}
        {...teile}
      />
    </Sprachraum>,
  )

  return {
    aufTischbild,
    eigenesBild,
    eigenesBildWeg,
    hand: nutzerereignis.setup(),
  }
}

// **Dieselbe Schwelle wie die Spalten** — schaltete die Leiste an einer
// eigenen um, stünde sie neben dem Tisch hinter dem Knopf oder darunter
// wieder in zwei Zeilen.
it('schaltet an derselben Schwelle um wie Tisch und Spalte', async () => {
  zeige()

  const knopf = await screen.findByRole('button', {
    name: eigene.tischDarstellung,
  })
  const schmal = knopf.closest('[data-tischleiste="schmal"]')
  const breit = document.querySelector('[data-tischleiste="breit"]')
  const ziel = screen.getByRole('group', { name: eigene.tischZiel })

  expect(schmal).toHaveClass('@4xl:hidden')
  expect(breit).toHaveClass('hidden', '@4xl:flex')
  expect(ziel.closest('[data-tischleiste="ziel"]')).toHaveClass(
    'hidden',
    '@4xl:flex',
  )
  expect(
    within(breit as HTMLElement).getByRole('button', {
      name: eigene.abzugTitel,
    }),
  ).toBeInTheDocument()
})

it('öffnet Ziel, Hintergrund, eigenes Bild und die Ausgabe hinter „Darstellung"', async () => {
  const { aufTischbild, eigenesBild, hand } = zeige()

  const knopf = await screen.findByRole('button', {
    name: eigene.tischDarstellung,
  })
  expect(knopf).toHaveAttribute('aria-expanded', 'false')

  await hand.click(knopf)

  expect(knopf).toHaveAttribute('aria-expanded', 'true')
  const kasten = screen.getByRole('group', { name: eigene.tischDarstellung })

  // **Das Ziel steht mit darin**, als eigene benannte Gruppe.
  expect(
    within(
      within(kasten).getByRole('group', { name: eigene.tischZiel }),
    ).getAllByRole('button'),
  ).toHaveLength(2)

  fireEvent.change(within(kasten).getByLabelText(eigene.tischbild), {
    target: { value: 'taverne' },
  })
  expect(aufTischbild).toHaveBeenCalledExactlyOnceWith('taverne')

  await hand.click(
    within(kasten).getByRole('button', { name: eigene.tischEigenes }),
  )
  expect(eigenesBild).toHaveBeenCalledOnce()
})

// **Wegnehmen nur, wo etwas wegzunehmen ist** — dieselbe Regel wie in der
// Zeile.
it('bietet „Eigenes Bild entfernen" nur mit eigenem Bild an', async () => {
  const { eigenesBildWeg, hand } = zeige({ hintergrund: 'tisch://eigen' })

  await hand.click(
    await screen.findByRole('button', { name: eigene.tischDarstellung }),
  )
  const kasten = screen.getByRole('group', { name: eigene.tischDarstellung })

  await hand.click(
    within(kasten).getByRole('button', { name: eigene.tischEigenesWeg }),
  )
  expect(eigenesBildWeg).toHaveBeenCalledOnce()
})

// **Die Ausgabe klappt unter der Leiste auf, und das Popover geht zu** — es
// stünde sonst über dem Feld, das es gerade geöffnet hat.
it('öffnet die Ausgabe aus „Darstellung" und schließt das Popover', async () => {
  const { hand } = zeige()

  const knopf = await screen.findByRole('button', {
    name: eigene.tischDarstellung,
  })
  await hand.click(knopf)
  const kasten = screen.getByRole('group', { name: eigene.tischDarstellung })

  await hand.click(
    within(kasten).getByRole('button', { name: eigene.abzugTitel }),
  )

  expect(
    screen.queryByRole('group', { name: eigene.tischDarstellung }),
  ).toBeNull()
  expect(knopf).toHaveAttribute('aria-expanded', 'false')
  expect(
    await screen.findByRole('button', { name: eigene.abzugErzeugen }),
  ).toBeInTheDocument()
})

// **Escape schließt „Darstellung"** — und der Fokus steht wieder am Knopf;
// am nativen Fenster misst es `play06-tisch`.
it('schließt „Darstellung" mit Escape', async () => {
  const { hand } = zeige()

  const knopf = await screen.findByRole('button', {
    name: eigene.tischDarstellung,
  })
  await hand.click(knopf)
  const kasten = screen.getByRole('group', { name: eigene.tischDarstellung })

  fireEvent.keyDown(within(kasten).getByRole('combobox'), { key: 'Escape' })

  expect(
    screen.queryByRole('group', { name: eigene.tischDarstellung }),
  ).toBeNull()
  expect(knopf).toHaveAttribute('aria-expanded', 'false')
})
