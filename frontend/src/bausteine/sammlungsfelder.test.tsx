// **Die Felder des Sammlungsformulars sagen, was sie erwarten** — UX-06
// (#1678).
//
// Menge, Sprache und Währung: eine ganze Zahl ab eins wie in
// `domain.Exemplar.Validate`, Scryfalls Sprachcodes als Vorschläge ohne
// Grenze, und ein Währungscode, der klein geschrieben sein darf, weil
// `Normalisieren` ihn anhebt.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Sammlungsformular } from './Sammlungsformular'
import { einsetzen } from '../sprache/einsetzen'
import { Sprachraum } from '../sprache/kontext'
import { raeumeApi } from '../test/api'
import { druckfeld } from '../test/druckfeld'
import { kataloge } from '../sprache/kataloge'

import type { ExemplarDTO } from '../api'

const texte = kataloge.de.seiten.sammlung

afterEach(raeumeApi)

function zeige() {
  const aufgenommen = vi.fn<(exemplar: ExemplarDTO) => Promise<boolean>>(() =>
    Promise.resolve(true),
  )

  render(
    <Sprachraum sprache="de">
      <Sammlungsformular aufAufnehmen={aufgenommen} laeuft={false} />
    </Sprachraum>,
  )

  return aufgenommen
}

it('hält eine Menge von null am Feld fest, statt die letzte gültige zu schicken', async () => {
  const aufgenommen = zeige()

  await userEvent.type(await druckfeld(), 'abc')
  await userEvent.clear(screen.getByLabelText(texte.menge))
  await userEvent.type(screen.getByLabelText(texte.menge), '0')

  expect(screen.getByRole('alert')).toHaveTextContent(
    einsetzen(kataloge.de.allgemein.ganzeZahlAb, { kleinstens: '1' }),
  )
  expect(screen.getByRole('button', { name: texte.aufnehmen })).toBeDisabled()

  await userEvent.clear(screen.getByLabelText(texte.menge))
  await userEvent.type(screen.getByLabelText(texte.menge), '3')
  await userEvent.click(screen.getByRole('button', { name: texte.aufnehmen }))

  expect(aufgenommen.mock.calls[0]?.[0].menge).toBe(3)
})

it('schlägt Scryfalls Sprachcodes vor und nimmt auch einen anderen', async () => {
  const aufgenommen = zeige()

  const sprache = screen.getByRole('combobox', { name: texte.sprache })
  const liste = document.getElementById(sprache.getAttribute('list') ?? '')

  expect(
    liste?.querySelector('option[value="ph"]')?.getAttribute('label'),
  ).toBe('Phyrexianisch')

  await userEvent.type(await druckfeld(), 'abc')
  await userEvent.type(sprache, 'xx')
  await userEvent.click(screen.getByRole('button', { name: texte.aufnehmen }))

  expect(aufgenommen.mock.calls[0]?.[0].sprache).toBe('xx')
})

it('nimmt eine klein geschriebene Währung und hält eine ohne Gestalt an', async () => {
  zeige()

  await userEvent.type(await druckfeld(), 'abc')

  // **Die Währung steht seit UXR-10 (#1918) unter „Weitere Angaben"** —
  // aufgeklappt hat sie schon das Feld der Kennung (AP-10).
  const waehrung = screen.getByRole('combobox', { name: texte.erwerbswaehrung })

  await userEvent.clear(waehrung)
  await userEvent.type(waehrung, 'eur')
  await userEvent.tab()

  expect(screen.queryByRole('alert')).toBeNull()
  expect(screen.getByRole('button', { name: texte.aufnehmen })).toBeEnabled()

  await userEvent.clear(waehrung)
  await userEvent.type(waehrung, '€')
  await userEvent.tab()

  expect(screen.getByRole('alert')).toHaveTextContent(
    einsetzen(kataloge.de.bausteine.waehrung.keineGestalt, { wert: '€' }),
  )
  expect(screen.getByRole('button', { name: texte.aufnehmen })).toBeDisabled()
})
