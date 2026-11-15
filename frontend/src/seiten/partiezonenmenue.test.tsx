// **Das Kontextmenü an einer Karte** — PLAY-04.AK-01
// (#1712).
//
// # Wogegen hier geprüft wird
//
// **Dass der dritte Weg wirklich da ist.** Ziehen prüft
// `tischziehenregeln.test.ts`, die Handlungsleiste prüfen die Läufe daneben
// — das Kontextmenü gab es am Tisch nicht.
import { render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import { Partiezonen } from './Partiezonen'
import { Sprachraum } from '../sprache/kontext'
import { einObjekt, einePartie, ruhenderLauf } from './partieprobe'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.spielen
const handlungen = kataloge.de.seiten.handlungen

// zeige zeichnet die Zonenliste mit einer Karte auf dem Tisch.
function zeige(
  befehlen: (art: string, nutzlast: Record<string, unknown>) => void = () =>
    undefined,
  waehle: (kennung: string) => void = () => undefined,
  erlaubt: string[] = ['objekt.tappen', 'objekt.schieben'],
) {
  render(
    <Sprachraum sprache="de">
      <Partiezonen
        partie={einePartie({
          objekte: [
            einObjekt({
              kennung: 'k1',
              zone: 'tisch',
              name: 'Wald',
              geheim: false,
            }),
          ],
        })}
        texturen={ruhenderLauf()}
        gewaehlt={undefined}
        waehle={waehle}
        wirkungen={new Map()}
        erlaubt={erlaubt}
        befehlen={befehlen}
        // **Was am Objekt gilt, sagt der Kern** (P3/P4, #2371, #2372) — das
        // Menü liest seine Auskunft, nicht `erlaubt`.
        handlungen={[
          {
            objekt: 'k1',
            handlungen: erlaubt,
            primaer: '',
            zweit: '',
          },
        ]}
        greife={() => undefined}
        tastenFuer={() => ''}
        taste={() => false}
      />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}

// **Die rechte Taste öffnet das Menü.**
it('stellt einen Befehl aus dem Kontextmenü', async () => {
  const befohlen = vi.fn()
  const hand = zeige(befohlen)

  await hand.pointer({
    keys: '[MouseRight]',
    target: screen.getByRole('button', { name: /Wald/ }),
  })

  await hand.click(
    screen.getByRole('menuitem', {
      name: handlungen.handlungsname['objekt.tappen'],
    }),
  )

  expect(befohlen).toHaveBeenCalledWith('objekt.tappen', { objekt: 'k1' })
})

// **Was ein Feld braucht, steht nicht als Befehl darin** — ein Menüeintrag
// ist eine Handlung und kein Formular. Stattdessen führt ein Eintrag zur
// Handlungsleiste, und der wählt die Karte.
it('führt für die übrigen Handlungen zur Leiste', async () => {
  const gewaehlt = vi.fn()
  const hand = zeige(() => undefined, gewaehlt)

  await hand.pointer({
    keys: '[MouseRight]',
    target: screen.getByRole('button', { name: /Wald/ }),
  })

  expect(
    screen.queryByRole('menuitem', {
      name: handlungen.handlungsname['objekt.schieben'],
    }),
  ).not.toBeInTheDocument()

  await hand.click(
    screen.getByRole('menuitem', { name: eigene.kartenmenueWeitere }),
  )

  expect(gewaehlt).toHaveBeenCalledWith('k1')
})

// **Was der Kern nicht erlaubt, steht nicht darin** — ein Eintrag, den er
// abwiese, wäre eine Einladung, die er zurücknimmt.
it('zeigt keinen Befehl, den der Kern nicht erlaubt', async () => {
  const hand = zeige(
    () => undefined,
    () => undefined,
    [],
  )

  await hand.pointer({
    keys: '[MouseRight]',
    target: screen.getByRole('button', { name: /Wald/ }),
  })

  expect(
    screen.queryByRole('menuitem', {
      name: handlungen.handlungsname['objekt.tappen'],
    }),
  ).not.toBeInTheDocument()

  // **Der Weg zur Leiste bleibt trotzdem** — dort steht, was gerade gilt,
  // und ein Menü ohne jeden Eintrag wäre eine Sackgasse.
  expect(
    screen.getByRole('menuitem', { name: eigene.kartenmenueWeitere }),
  ).toBeInTheDocument()
})
