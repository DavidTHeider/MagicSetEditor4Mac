// **Die Adressen der Anbieter stehen eingeklappt** — UXR-21
// (#1929).
//
// Am nativen Fenster gemessen (`docs/belege/huellenprobe/uxr21.md`, vorher):
// *Einstellungen › KI* rollte **3,2 Bildschirme**, eng 4,2 — über R4 aus
// `docs/UIREVIEW.md`. Vier der zwölf Felder waren Adressen: je Anbieter eine
// und die der Laufzeit, gezeichnet unter jedem Schalter, auch wenn er aus
// ist. Eine Adresse stellt man einmal ein oder nie; wer einschaltet, fragt,
// **ob** gefragt wird, nicht wohin.
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it } from 'vitest'

import {
  eintraege,
  zeigeEinstellungen as zeige,
} from '../test/einstellungsseite'
import { kataloge } from '../sprache/kataloge'
import { raeumeApi } from '../test/api'

import type { EinstellungDTO } from '../api'

const eigene = kataloge.de.seiten.einstellungen

afterEach(() => {
  raeumeApi()
})

function zeile(schluessel: string, werte: string[] = []): EinstellungDTO {
  return {
    schluessel,
    wert: werte[0] ?? '',
    wirksam: werte[0] ?? '',
    werte,
    bereich: 'dienste',
  }
}

const dienste = [
  zeile('dienst.anthropic', ['aus', 'an']),
  zeile('dienst.anthropic.adresse'),
  zeile('dienst.openai', ['aus', 'an']),
  zeile('dienst.openai.adresse'),
]
const ki = [zeile('ki.laufzeit', ['aus', 'an']), zeile('ki.adresse')]

const adressen = [
  eigene.namen['dienst.anthropic.adresse'],
  eigene.namen['dienst.openai.adresse'],
  eigene.namen['ki.adresse'],
]

async function zurKI() {
  zeige({ eintraege, dienste, fragen: [], ki })
  await userEvent.click(screen.getByRole('tab', { name: eigene.gruppen.ki }))
}

const knopf = () =>
  screen.getByRole('button', { name: eigene.abschnitte.adressen })

it('zeigt die Schalter offen und die Adressen eingeklappt', async () => {
  await zurKI()

  expect(screen.getByRole('switch', { name: 'Anthropic' })).toBeInTheDocument()
  expect(knopf()).toHaveAttribute('aria-expanded', 'false')
  for (const name of adressen) {
    expect(screen.queryByLabelText(name)).not.toBeInTheDocument()
  }
})

it('öffnet die Adressen auf einen Klick — alle drei, jede einmal', async () => {
  await zurKI()

  await userEvent.click(knopf())

  for (const name of adressen) {
    expect(screen.getAllByLabelText(name)).toHaveLength(1)
  }
})

// **Die Suche hält den Kasten offen** — sonst führte „Adresse" zu einem
// Feld, das gar nicht gezeichnet ist (UX-04.AK-03).
it('hält offen, wo die Suche eine Adresse trifft', async () => {
  await zurKI()

  await userEvent.type(
    screen.getByRole('searchbox', { name: eigene.suche }),
    'Adresse von Anthropic',
  )

  expect(knopf()).toHaveAttribute('aria-expanded', 'true')
  // **Mit der Marke „gesucht" im Namen** — deshalb der Anfang.
  expect(
    screen.getByRole('textbox', { name: new RegExp(`^${adressen[0] ?? ''}`) }),
  ).toBeVisible()
})
