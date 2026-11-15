// **Das Menü einer Karte am gezeichneten Tisch steht im Fenster** — P4
// (#2372),
// gefunden am nativen Fenster (`tischhand09`): das Menü steht
// `position: fixed`, und die Karte trug ihre Drehung am `li` darum. Ein
// Vorfahr mit `transform` macht sich selbst zum Bezug eines festen
// Elements — das Menü stand dann neben der gedrehten Karte, und die Fläche
// mit `overflow: hidden` schnitt es ab. Unter jsdom zeichnet niemand, also
// wird die Ursache geprüft: **zwischen Menü und Fenster keine Transformation**.
import { fireEvent, render, screen } from '@testing-library/react'
import { expect, it } from 'vitest'

import { einObjekt, einePartie } from '../seiten/partieprobe'
import { Sprachraum } from '../sprache/kontext'
import { tischplan } from '../tisch/tischplan'
import type { Tischmalerei } from '../tisch/maler'
import { einMaler } from '../test/malerprobe'
import { Tischflaeche } from './Tischflaeche'

it('dreht die Karte am Knopf und nicht um das Menü herum', async () => {
  const maler = einMaler({ breite: 800, hoehe: 600 })
  render(
    <Sprachraum sprache="de">
      <Tischflaeche
        plan={tischplan(
          einePartie({
            objekte: [
              einObjekt({
                kennung: 'a',
                zone: 'tisch',
                name: 'Wald',
                getappt: true,
                eigner: 'sitz1',
                beherrscher: 'sitz1',
                geheim: false,
              }),
            ],
          }),
        )}
        bild={() => undefined}
        malerei={() => Promise.resolve<Tischmalerei>({ art: 'da', maler })}
        menueFuer={() => ({
          bezeichnung: 'Handlungen an Wald',
          eintraege: [
            { schluessel: 'x', name: 'Tappen', aufWahl: () => undefined },
          ],
        })}
      />
    </Sprachraum>,
  )
  const knopf = await screen.findByRole('button', { name: /Wald/ })

  expect(knopf.style.transform).toMatch(/^rotate\(/)

  fireEvent.contextMenu(knopf, { clientX: 10, clientY: 10 })
  const menue = screen.getByRole('menu')
  const mitDrehung: string[] = []
  for (
    let vorfahr = menue.parentElement;
    vorfahr;
    vorfahr = vorfahr.parentElement
  ) {
    if (vorfahr.style.transform !== '') mitDrehung.push(vorfahr.tagName)
  }

  expect(mitDrehung).toEqual([])
})
