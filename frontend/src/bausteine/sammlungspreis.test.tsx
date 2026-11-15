// **Ein Preis in der Sprache des Fensters** — LOC-03.AK-01 und AK-03
// (#1689):
// „lokalisierte Eingaben müssen verlustfrei in kanonische interne Werte
// übergehen", und „mehrdeutige Eingaben werden erklärt".
//
// # Warum das Formular übersetzt und nicht Go
//
// **Weil nur das Fenster die Sprache dessen kennt, der tippt.** Über die
// Naht geht die kanonische Form (`docs/LOKALISIERUNG.md`); Kommandozeile,
// REST und MCP nehmen sie unmittelbar. Vorher ging »19,99« unverändert
// hinüber, und Go wies es ab — richtig erklärt, aber für jemanden, der
// Deutsch schreibt, eine Abweisung dessen, was er richtig geschrieben hat.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Sammlungsformular } from './Sammlungsformular'
import { Sprachraum } from '../sprache/kontext'
import { raeumeApi } from '../test/api'
import { druckfeld } from '../test/druckfeld'
import { kataloge } from '../sprache/kataloge'

import type { Sprache } from '../sprache/kataloge'

afterEach(raeumeApi)

// aufnehmen tippt Druck, Preis und Währung und drückt den Knopf.
async function aufnehmen(
  sprache: Sprache,
  preis: string,
  waehrung: string,
  aufgenommen = vi.fn(() => Promise.resolve(true)),
) {
  const texte = kataloge[sprache].seiten.sammlung

  render(
    <Sprachraum sprache={sprache}>
      <Sammlungsformular aufAufnehmen={aufgenommen} laeuft={false} />
    </Sprachraum>,
  )

  await userEvent.type(await druckfeld(document.body, texte), 'abc')

  await userEvent.type(screen.getByLabelText(texte.erwerbspreis), preis)
  await userEvent.type(screen.getByLabelText(texte.erwerbswaehrung), waehrung)
  await userEvent.click(screen.getByRole('button', { name: texte.aufnehmen }))

  return aufgenommen
}

// **Was die Sprache selbst so schriebe, geht kanonisch hinüber.**
it.each([
  ['de', '19,99', 'EUR', '19.99'],
  ['de', '1.234,50', 'EUR', '1234.50'],
  ['en', '1,234.50', 'USD', '1234.50'],

  // **Beim Yen ist ein Punkt im Deutschen immer eine Gruppe** — der Yen
  // hat keine Nachkommastellen, also gibt es nichts zu raten.
  ['de', '1.234', 'JPY', '1234'],

  // **Die kanonische Form bleibt, wie sie ist** — in jeder Sprache.
  ['de', '4.17', 'EUR', '4.17'],
] as const)(
  '%s: »%s« %s geht als %s hinüber',
  async (sprache, preis, waehrung, kanonisch) => {
    const aufgenommen = await aufnehmen(sprache, preis, waehrung)

    expect(aufgenommen).toHaveBeenCalledWith(
      expect.objectContaining({ erwerbspreis: kanonisch }),
    )
  },
)

// **Was keine Lesart hat, geht unverändert** — Go nennt dann, was dasteht,
// und rät nicht.
it('reicht einen Text ohne Lesart unverändert weiter', async () => {
  const aufgenommen = await aufnehmen('en', '19,99', 'USD')

  expect(aufgenommen).toHaveBeenCalledWith(
    expect.objectContaining({ erwerbspreis: '19,99' }),
  )
})

// **Zwei Werte werden erklärt und nicht geraten** (AK-03): »1.234« Dinar
// ist im Deutschen tausendzweihundertvierunddreißig — und in kanonischer
// Form eins Komma zwei drei vier.
it('erklärt am Feld, was zwei Werte haben kann', async () => {
  const texte = kataloge.de.seiten.sammlung
  const aufgenommen = await aufnehmen('de', '1.234', 'BHD')

  // **Beide Lesarten in deutscher Schreibweise** — und am Feld, nicht als
  // Hinweis irgendwo (§6).
  const befund = await screen.findByRole('alert')
  expect(befund).toHaveTextContent('1.234,000')
  expect(befund).toHaveTextContent('1,234')
  expect(screen.getByLabelText(texte.erwerbspreis)).toHaveAttribute(
    'aria-invalid',
    'true',
  )

  // **Und nichts ist hinübergegangen** — der Befund oben belegt, dass der
  // Knopf gehört wurde.
  expect(aufgenommen).not.toHaveBeenCalled()

  // **Wer den Preis ändert, räumt den Befund** — er gehörte zum alten Text.
  await userEvent.type(screen.getByLabelText(texte.erwerbspreis), ',5')
  expect(screen.queryByRole('alert')).toBeNull()

  // **Und eindeutig geschrieben geht er hinüber.**
  await userEvent.click(screen.getByRole('button', { name: texte.aufnehmen }))
  expect(aufgenommen).toHaveBeenCalledWith(
    expect.objectContaining({ erwerbspreis: '1234.5' }),
  )
})

// **Wer die Währung ändert, räumt ihn auch** — die Stellen, an denen die
// zweite Lesart hing, kamen aus ihr.
it('räumt den Befund, wenn die Währung wechselt', async () => {
  const texte = kataloge.de.seiten.sammlung

  await aufnehmen('de', '1.234', 'BHD')
  await screen.findByRole('alert')

  await userEvent.clear(screen.getByLabelText(texte.erwerbswaehrung))

  expect(screen.queryByRole('alert')).toBeNull()
})
