// **Die eigene Kopie in der Werkstatt** — FMT-02.AK-04 und S10.NAT-01
// (#1691).
//
// Ein mitgeliefertes oder fremdverwaltetes Paket wird über eine eigene Kopie
// angepasst. **Zwei Stellen bieten sie an**: der Baum unter einem Paket, das
// sich nur lesen lässt — die Saat liegt außerhalb der Werkstatt, ihre Dateien
// lassen sich nicht aufzählen —, und der Hinweis über dem Editor eines
// bezogenen oder übernommenen. Danach steht die Werkstatt in der Kopie und
// sagt, dass das Original bleibt.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { kurzform } from '../sprache/kurzform'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, raeumeApi } from '../test/api'
import { paketDTO } from '../test/vorlagenprobe'
import { belegeWerkstatt } from '../test/werkstattprobe'
import { Vorlagen } from './Vorlagen'

import type { Api } from '../api'
import type { Paketherkunftsart } from '../pakete'

const texte = kataloge.de.seiten.vorlagen

// kurz ist, was über dem Editor von einem Hinweis zur Kopie zu lesen ist.
// **Seit UXR-18 (#1926) der erste Satz** — die Begründung steht hinter dem
// Fragezeichen, damit Menü, Warnung und Kopie in eine Zeile passen
// (`dateikopf.test.tsx`).
const kurz = kurzform

// paket ist ein Stil der Probe mit Herkunftsart.
function paket(name: string, herkunftsart: Paketherkunftsart, ort = '/saat/') {
  return {
    ...paketDTO(name + '.mse-style'),
    art: 'mse-style',
    name,
    anzeigename: name,
    pfad: ort + name + '.mse-style',
    herkunftsart,
  }
}

const kopie = paket('probe-schlicht-kopie', 'eigen', '/pakete/')

const ausserhalb = {
  code: 'validierung' as const,
  schluessel: '',
  meldung: 'liegt nicht im Paketverzeichnis',
  werte: {},
}

function zeige(original: ReturnType<typeof paket>, eigene: Partial<Api> = {}) {
  let kopiert = false
  const kopieren = vi.fn(() => {
    kopiert = true

    return Promise.resolve({ paket: kopie, fehler: null, abgebrochen: false })
  })
  const lesen = vi.fn((_paket: string, datei: string) =>
    Promise.resolve({
      inhalt: 'mse version: 2.0.0\n',
      stand: datei,
      fehler: null,
    }),
  )

  belegeWerkstatt({
    Pakete: () =>
      Promise.resolve({
        pakete: kopiert ? [original, kopie] : [original],
        befunde: [],
        fehler: null,
      }),
    // **Die Saat lässt ihre Dateien nicht aufzählen** — wie im Betrieb:
    // sie liegt außerhalb des Paketverzeichnisses der Werkstatt.
    Paketdateien: (pfad: string) =>
      Promise.resolve(
        pfad.startsWith('/saat/')
          ? { dateien: [], beiwerk: [], fehler: ausserhalb }
          : { dateien: ['style'], beiwerk: [], fehler: null },
      ),
    PaketdateiLesen: lesen,
    PaketKopieren: kopieren,
    ...eigene,
  })

  render(
    <Sprachraum sprache="de">
      <Vorlagen />
    </Sprachraum>,
  )

  return { kopieren, lesen }
}

// zweig gibt den Eintrag eines Pakets im Baum.
async function zweig(name: string): Promise<HTMLElement> {
  const eintrag = (
    await screen.findByRole('button', { name: name + '.mse-style' })
  ).closest('li')

  if (eintrag === null) {
    throw new Error(`kein Zweig für ${name}`)
  }

  return eintrag
}

// oeffneStil öffnet die Datei `style` eines Pakets im Baum.
async function oeffneStil(name: string) {
  await userEvent.click(
    await within(await zweig(name)).findByRole('button', { name: 'style' }),
  )
  await screen.findByLabelText(texte.inhalt)
}

const originalBleibt = einsetzen(texte.kopieOriginal, {
  original: 'probe-schlicht',
})

const vergeben = {
  paket: null,
  abgebrochen: false,
  fehler: {
    code: 'konflikt' as const,
    schluessel: '',
    meldung: 'Der Name ist vergeben.',
    werte: {},
  },
}

describe('Die eigene Kopie im Baum', () => {
  afterEach(raeumeApi)

  it('steht unter einem mitgelieferten Paket und öffnet die Hauptdatei der Kopie', async () => {
    const { kopieren, lesen } = zeige(paket('probe-schlicht', 'mitgeliefert'))

    const saat = await zweig('probe-schlicht')
    expect(await within(saat).findByText(texte.nurLesbar)).toBeInTheDocument()

    await userEvent.click(
      within(saat).getByRole('button', { name: texte.kopieren }),
    )

    expect(await screen.findByText(originalBleibt)).toBeInTheDocument()
    expect(kopieren).toHaveBeenCalledWith('mse-style', 'probe-schlicht', '')
    await waitFor(() => {
      expect(lesen).toHaveBeenCalledWith(kopie.pfad, kopie.hauptdatei)
    })
  })

  it('nennt einen Fehler der Kopie an ihrem Paket', async () => {
    const { lesen } = zeige(paket('probe-schlicht', 'mitgeliefert'), {
      PaketKopieren: () => Promise.resolve(vergeben),
    })

    const saat = await zweig('probe-schlicht')
    await userEvent.click(
      await within(saat).findByRole('button', { name: texte.kopieren }),
    )

    expect(await within(saat).findByRole('alert')).toHaveTextContent(
      'Der Name ist vergeben.',
    )
    expect(lesen).not.toHaveBeenCalled()
  })
})

describe('Ein zweiter Versuch', () => {
  afterEach(raeumeApi)

  it('räumt den Fehler des ersten, wenn er glückt', async () => {
    const kopieren = vi
      .fn()
      .mockResolvedValueOnce(vergeben)
      .mockResolvedValue({ paket: kopie, fehler: null, abgebrochen: false })
    zeige(paket('probe-schlicht', 'mitgeliefert'), {
      PaketKopieren: kopieren,
    })

    const saat = await zweig('probe-schlicht')
    const knopf = await within(saat).findByRole('button', {
      name: texte.kopieren,
    })

    await userEvent.click(knopf)
    expect(await within(saat).findByRole('alert')).toBeInTheDocument()

    await userEvent.click(knopf)
    expect(await screen.findByText(originalBleibt)).toBeInTheDocument()
    expect(within(saat).queryByRole('alert')).not.toBeInTheDocument()
  })
})

describe('Die eigene Kopie über dem Editor', () => {
  afterEach(raeumeApi)

  it('nennt bei einem bezogenen Paket die lokale Änderung und wechselt in dieselbe Datei', async () => {
    const { kopieren, lesen } = zeige(
      paket('probe-schlicht', 'bezogen', '/pakete/'),
      {
        Paketdateien: () =>
          Promise.resolve({
            dateien: ['card_fields', 'style'],
            beiwerk: [],
            fehler: null,
          }),
      },
    )

    await oeffneStil('probe-schlicht')
    expect(screen.getByText(kurz(texte.kopieFremd))).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: texte.kopieren }))

    expect(await screen.findByText(originalBleibt)).toBeInTheDocument()
    expect(kopieren).toHaveBeenCalledWith('mse-style', 'probe-schlicht', '')
    await waitFor(() => {
      expect(lesen).toHaveBeenCalledWith(kopie.pfad, 'style')
    })
    expect(screen.queryByText(kurz(texte.kopieFremd))).not.toBeInTheDocument()
  })

  it('nennt bei einem mitgelieferten Paket, das sich öffnen lässt, warum', async () => {
    zeige(paket('probe-schlicht', 'mitgeliefert'), {
      Paketdateien: () =>
        Promise.resolve({ dateien: ['style'], beiwerk: [], fehler: null }),
    })

    await oeffneStil('probe-schlicht')

    expect(screen.getByText(kurz(texte.kopieMitgeliefert))).toBeInTheDocument()
  })

  it('schweigt bei einem eigenen Paket', async () => {
    zeige(paket('probe-eigen', 'eigen', '/pakete/'))

    await oeffneStil('probe-eigen')
    await antwortAbwarten()

    expect(
      screen.queryByRole('button', { name: texte.kopieren }),
    ).not.toBeInTheDocument()
  })

  it('zeigt einen Fehler der Kopie und bleibt beim Original', async () => {
    const { lesen } = zeige(paket('probe-schlicht', 'bezogen', '/pakete/'), {
      PaketKopieren: () => Promise.resolve(vergeben),
    })

    await oeffneStil('probe-schlicht')
    await userEvent.click(screen.getByRole('button', { name: texte.kopieren }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Der Name ist vergeben.',
    )
    expect(lesen).not.toHaveBeenCalledWith(kopie.pfad, 'style')
  })

  // **Eine späte Antwort entführt die Wahl nicht** (§7, Regel 2): wer
  // inzwischen eine andere Datei geöffnet hat, bleibt dort — die Kopie liegt
  // trotzdem da und steht im Baum.
  it('wechselt nicht, wenn inzwischen eine andere Datei offen ist', async () => {
    let fertig: () => void = () => undefined
    const original = paket('probe-schlicht', 'bezogen', '/pakete/')
    const kopieren = vi.fn(
      () =>
        new Promise<{
          paket: typeof kopie
          fehler: null
          abgebrochen: boolean
        }>((erfuellen) => {
          fertig = () => {
            erfuellen({ paket: kopie, fehler: null, abgebrochen: false })
          }
        }),
    )
    const { lesen } = zeige(original, {
      Pakete: () =>
        Promise.resolve({
          pakete:
            kopieren.mock.calls.length > 0 ? [original, kopie] : [original],
          befunde: [],
          fehler: null,
        }),
      Paketdateien: () =>
        Promise.resolve({
          dateien: ['style', 'card_fields'],
          beiwerk: [],
          fehler: null,
        }),
      PaketKopieren: kopieren,
    })

    await oeffneStil('probe-schlicht')
    await userEvent.click(screen.getByRole('button', { name: texte.kopieren }))
    await userEvent.click(screen.getByRole('button', { name: 'card_fields' }))
    await waitFor(() => {
      expect(lesen).toHaveBeenCalledWith(original.pfad, 'card_fields')
    })

    fertig()

    expect(
      await screen.findByRole('button', {
        name: 'probe-schlicht-kopie.mse-style',
      }),
    ).toBeInTheDocument()
    expect(lesen).not.toHaveBeenCalledWith(kopie.pfad, 'card_fields')
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('bleibt, wenn die Antwort keine Kopie nennt', async () => {
    const { lesen } = zeige(paket('probe-schlicht', 'bezogen', '/pakete/'), {
      PaketKopieren: () =>
        Promise.resolve({ paket: null, fehler: null, abgebrochen: false }),
    })

    await oeffneStil('probe-schlicht')
    await userEvent.click(screen.getByRole('button', { name: texte.kopieren }))
    await antwortAbwarten()

    expect(lesen).not.toHaveBeenCalledWith(kopie.pfad, 'style')
    expect(screen.getByText(kurz(texte.kopieFremd))).toBeInTheDocument()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })
})
