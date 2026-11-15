// **Das Proxyfeld aus Sicht dessen, der es sucht** — S13.NAT-01 und
// S13.NAT-02 (#1706):
//
//   - „Suche nach Proxy, HTTP und Netz führt zu einem editierbaren passenden
//     Feld oder klarer schreibgeschützter Herkunft."
//   - „Eine gültige Proxyänderung ist im Status sichtbar; Abbruch ändert
//     nichts."
//
// # Der Befund
//
// **Einen Abbruch gab es nicht.** Wer im Feld tippte und es sich anders
// überlegte, hatte nur das Verlassen — und das schrieb. Jetzt verwirft Esc
// den Entwurf, wie in jedem Textfeld von macOS.
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, it } from 'vitest'

import {
  eintraege,
  zeigeEinstellungen as zeige,
} from '../test/einstellungsseite'
import { kataloge } from '../sprache/kataloge'
import { einsetzen } from '../sprache/einsetzen'

import type { EinstellungDTO } from '../api'

const eigene = kataloge.de.seiten.einstellungen

function netzeintrag(schluessel: string, wert = ''): EinstellungDTO {
  return {
    schluessel,
    wert,
    wirksam: wert,
    werte: [],
    bereich: 'netz',
    ebene: 'geraet',
    herkunft: wert === '' ? 'vorgabe' : 'geraet',
    vorgabe: '',
    wirkung: 'sofort',
  }
}

const netz = [
  netzeintrag('netz.proxy.http'),
  netzeintrag('netz.proxy.https'),
  netzeintrag('netz.schnittstelle'),
]

function feld(schluessel: string): HTMLInputElement {
  const gefunden = screen
    .getByRole('tabpanel')
    .querySelector<HTMLInputElement>(`input[data-einstellung="${schluessel}"]`)

  if (gefunden === null) {
    throw new Error(`kein Feld für ${schluessel}`)
  }

  return gefunden
}

it.each([
  ['Proxy', 'netz.proxy.http'],
  ['HTTP', 'netz.proxy.http'],
  ['Netz', 'netz.schnittstelle'],
])(
  'die Suche nach „%s" führt zu einem Feld, das sich ändern lässt',
  async (frage, schluessel) => {
    zeige({ alleEinstellungen: [...eintraege, ...netz] })

    await userEvent.type(screen.getByLabelText(eigene.suche), frage)

    expect(feld(schluessel)).toBeEnabled()
  },
)

it('verwirft mit Esc und schreibt nichts', async () => {
  const setze = zeige({ alleEinstellungen: [...eintraege, ...netz] })
  const hand = userEvent.setup()

  await hand.type(screen.getByLabelText(eigene.suche), 'proxy')
  await hand.type(feld('netz.proxy.http'), 'http://neu.firma:3128')
  await hand.keyboard('{Escape}')
  await hand.tab()

  expect(feld('netz.proxy.http')).toHaveValue('')
  expect(setze).not.toHaveBeenCalled()
})

// **Die Gegenprobe, dass jemand zuhört** (§3): ohne Esc schreibt dasselbe
// Feld.
it('übernimmt mit ⏎', async () => {
  const setze = zeige({ alleEinstellungen: [...eintraege, ...netz] })
  const hand = userEvent.setup()

  await hand.type(screen.getByLabelText(eigene.suche), 'proxy')
  await hand.type(feld('netz.proxy.http'), 'http://neu.firma:3128{Enter}')

  expect(setze).toHaveBeenCalledWith('netz.proxy.http', 'http://neu.firma:3128')
})

// **Und was gilt, steht daneben** — mit Herkunft und Ebene.
it('zeigt einen gesetzten Proxy im Status', async () => {
  zeige({
    alleEinstellungen: [
      ...eintraege,
      netzeintrag('netz.proxy.http', 'http://proxy.firma:3128'),
    ],
  })

  await userEvent.type(screen.getByLabelText(eigene.suche), 'proxy')

  // **Als Marke** (AP-11, #2368) — der Wert steht im Feld, der ganze Satz
  // für das Vorleseprogramm.
  const feld = within(screen.getByRole('tabpanel'))
  expect(
    feld.getByText(
      einsetzen(eigene.herkunftMarke, {
        herkunft: eigene.herkunftNamen.geraet,
      }),
    ),
  ).toBeInTheDocument()
  expect(
    feld.getByText(
      einsetzen(eigene.herkunftZeile, {
        herkunft: eigene.herkunftNamen.geraet,
        ebene: eigene.ebenenNamen.geraet,
      }),
    ),
  ).toBeInTheDocument()
})
