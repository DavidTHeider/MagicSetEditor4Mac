// **Ein Fund im Netz als Name mit Marke** — UXR-16
// (#1924).
//
// # Der Befund
//
// **Jeder Fund trug Fingerabdruck, Adresse und gemerkte Adresse offen**, der
// Name nur als „Nennt sich: …" — bei sechs Funden achtzehn Zeilen
// Kennziffern, und *wer ist da?* stand dazwischen. Das Artboard *Gegenüber*
// zeigt einen Fund als Namen mit Marke; Fingerabdruck und Adresse stehen
// unter „Technische Angaben".
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { Nachbarn } from './Nachbarn'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { zaehle } from '../test/bedienzaehlung'
import { fund, sechsFunde } from '../test/gegenueberbestand'

import type { Nachbar } from '../apipeerkontakt'

const eigene = kataloge.de.seiten.technik

afterEach(() => {
  raeumeApi()
})

async function suche(nachbarn: Nachbar[]) {
  belegeApi({
    Nachbarn: () => Promise.resolve({ nachbarn, verworfen: 0, fehler: null }),
  })

  const { container } = render(
    <Sprachraum sprache="de">
      <Nachbarn />
    </Sprachraum>,
  )
  const hand = userEvent.setup()

  await hand.click(screen.getByRole('button', { name: eigene.nachbarnSuchen }))
  await screen.findAllByRole('listitem')

  return { container, hand }
}

describe('ein Fund im Netz', () => {
  it('nennt den Fund beim Namen, mit seiner Marke', async () => {
    await suche([fund('Ida')])

    const zeile = screen
      .getByText('Idas Mac', { selector: 'li span' })
      .closest('li') as HTMLElement

    expect(
      within(zeile).getByText(eigene.nachbarnUnbekannt),
    ).toBeInTheDocument()
    expect(
      within(zeile).getByRole('button', { name: eigene.nachbarnAufnehmen }),
    ).toBeInTheDocument()
  })

  it('legt Fingerabdruck und Adresse unter „Technische Angaben"', async () => {
    const { hand } = await suche([fund('Ida')])

    expect(screen.queryByText(/9f4b-8c8e-9025/)).toBeNull()
    expect(screen.queryByText('192.168.1.50:41234')).toBeNull()

    await hand.click(
      screen.getByRole('button', { name: eigene.einladung.technik }),
    )

    expect(screen.getByText(/9f4b-8c8e-9025/)).toBeInTheDocument()
    expect(screen.getByText('192.168.1.50:41234')).toBeInTheDocument()
    expect(screen.getByText('idas-mac.local:41234')).toBeInTheDocument()
  })

  it('nennt ohne Namen den Fingerabdruck', async () => {
    await suche([fund('Jan', { name: '', fingerabdruck: 'abcd-ef01' })])

    expect(
      screen.getByText('abcd-ef01', { selector: 'li span' }),
    ).toBeInTheDocument()
  })

  it('zeichnet bei sechs Funden keine Kennziffer offen', async () => {
    const { container } = await suche(sechsFunde())

    // **Je Fund höchstens zwei Bedienelemente** — „Als Kontakt aufnehmen"
    // oder „Adresse übernehmen" und „Technische Angaben" —, dazu
    // „Nachsehen" und die Hilfe an der Erklärung.
    expect(zaehle(container).bedien).toBeLessThanOrEqual(6 * 2 + 3)
    expect(screen.queryByText(eigene.kontakteAbdruck)).toBeNull()
  })
})
