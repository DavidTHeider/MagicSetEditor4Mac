// **Der Inspector eines Pakets** — DOM-01.AK-01
// (#1696):
// „Herkunft, Version, Prüfsumme, Lizenz, Abhängigkeiten und lokale
// Änderungen sind am Objekt sichtbar."
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { PaketDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Paketinspector } from './Paketinspector'

afterEach(cleanup)

const t = kataloge.de.seiten.pakete.plan

function paket(teil: Partial<PaketDTO>): PaketDTO {
  return {
    art: 'mse-game',
    name: 'probe',
    anzeigename: 'Probe',
    version: '1.4.0',
    pfad: '/pakete/probe.mse-game',
    herkunft: '',
    pruefsumme: '',
    braucht: [],
    gruppe: '',
    rang: 0,
    artrang: 20,
    herkunftsart: 'eigen',
    hauptdatei: 'style',
    geaendert: false,
    gebrauchtVon: [],
    benutzt: 0,
    ...teil,
  }
}

function zeige(
  eines: PaketDTO | undefined,
  meldung?: { text: string; fehler: boolean },
) {
  const loeschen = vi.fn()
  const zurueck = vi.fn()
  const weitergabe = {
    weitergeben: vi.fn(() => Promise.resolve()),
    laeuft: () => false,
    meldung: () => meldung,
  }
  render(
    <Paketinspector
      paket={eines}
      laeuft={false}
      aufLoeschen={loeschen}
      aufZurueck={zurueck}
      weitergabe={weitergabe}
    />,
  )

  return { loeschen, zurueck, weitergabe }
}

// **Als Installer weitergeben** — FMT-02.AK-04
// (#1691):
// am Paket, und was dabei herauskam, steht darunter.
describe('Paketinspector: weitergeben', () => {
  it('gibt das Paket weiter, das er zeigt', async () => {
    const eines = paket({})
    const { weitergabe } = zeige(eines)

    await userEvent.click(screen.getByRole('button', { name: t.weitergeben }))

    expect(weitergabe.weitergeben).toHaveBeenCalledWith(eines)
  })

  // **Die Meldung rollt sich in den Blick** — gemessen am nativen Fenster
  // (Hüllenprobe `fmt02e`): der Inspector ist eine eigene Rollfläche, und
  // die Meldung stand unter ihrem sichtbaren Rand; wer klickte, sah nichts.
  it('rollt die Meldung in den Blick', () => {
    const rollen = vi.fn()
    Object.defineProperty(HTMLElement.prototype, 'scrollIntoView', {
      configurable: true,
      value: rollen,
    })

    zeige(paket({}), { text: 'Als Installer gesichert: /x', fehler: false })

    expect(rollen).toHaveBeenCalledWith({ block: 'nearest' })
    Reflect.deleteProperty(HTMLElement.prototype, 'scrollIntoView')
  })

  it('sagt, wohin es ging, und nennt einen Fehler als solchen', () => {
    zeige(paket({}), { text: 'Als Installer gesichert: /x', fehler: false })
    expect(screen.getByRole('status')).toHaveTextContent('/x')

    cleanup()

    zeige(paket({}), { text: 'Das ging nicht.', fehler: true })
    expect(screen.getByRole('alert')).toHaveTextContent('Das ging nicht.')
  })
})

describe('Paketinspector', () => {
  it('erklärt ohne Wahl, was hier stünde', () => {
    zeige(undefined)

    expect(screen.getByText(t.inspectorLeer)).toBeInTheDocument()
  })

  it('zeigt Nutzung, Herkunft, Lizenz und beide Richtungen der Abhängigkeit', async () => {
    const { loeschen, zurueck } = zeige(
      paket({
        benutzt: 3,
        gebrauchtVon: ['probe-schlicht.mse-style'],
        pruefsumme: 'sha256:abc',
        herkunftsart: 'bezogen',
        herkunft: 'https://example.invalid/pack',
      }),
    )

    expect(screen.getByRole('heading', { name: 'probe' })).toBeInTheDocument()
    expect(screen.getByText('mse-game · Fassung 1.4.0')).toBeInTheDocument()
    expect(screen.getByText('3 Sets benutzen es.')).toBeInTheDocument()
    expect(
      screen.getByText('Bezogen von https://example.invalid/pack.'),
    ).toBeInTheDocument()
    expect(screen.getByText('sha256:abc')).toBeInTheDocument()
    expect(screen.getByText(t.lizenzUnbekannt)).toBeInTheDocument()
    expect(screen.getByText('probe-schlicht.mse-style')).toBeInTheDocument()

    // **„Hängt ab von: nichts"** — eine Aussage, keine leere Liste.
    expect(screen.getByText(t.nichts)).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: /Löschen/ }))
    expect(loeschen).toHaveBeenCalledWith(
      expect.objectContaining({ name: 'probe' }),
    )

    await userEvent.click(screen.getByRole('button', { name: t.zurueck }))
    expect(zurueck).toHaveBeenCalledWith(
      expect.objectContaining({ name: 'probe' }),
    )
  })

  it.each([
    [paket({ benutzt: 0 }), t.keinSet],
    [paket({ benutzt: 1 }), t.einSet],
    [paket({ art: 'mse-symbol-font' }), t.setsUnbekannt],
  ])('zählt die Sets nur, wo ein Kopf sie nennt', (eines, satz) => {
    zeige(eines)

    expect(screen.getByText(satz)).toBeInTheDocument()
  })

  it.each([
    ['mitgeliefert' as const, '', t.woherMitgeliefert],
    ['uebernommen' as const, '/mse2/data', 'Übernommen aus /mse2/data.'],
    ['eigen' as const, '', t.woherEigen],
  ])('sagt, woher es stammt: %s', (herkunftsart, herkunft, satz) => {
    zeige(paket({ herkunftsart, herkunft }))

    expect(screen.getByText(satz)).toBeInTheDocument()
  })

  it('nennt eine lokale Änderung, eine fehlende Fassung und den Bedarf', () => {
    zeige(
      paket({
        art: 'mse-style',
        version: '',
        geaendert: true,
        braucht: ['probe.mse-game 1.0'],
      }),
    )

    expect(screen.getByText(t.geaendert)).toBeInTheDocument()
    expect(screen.getByText('mse-style · ohne Fassung')).toBeInTheDocument()
    expect(screen.getByText('probe.mse-game 1.0')).toBeInTheDocument()
  })

  // **Nach Häufigkeit, das Löschen zuletzt** — UXR-17 (#1925): das Issue
  // fragt, ob der Inspector seine Handlungen so ordnet. Bis dahin stand
  // *Löschen …* vorn, die seltenste und die einzige, die etwas wegnimmt;
  // das Artboard `Pakete` stellt die gefährliche Handlung ans Ende.
  it('ordnet die Handlungen nach Häufigkeit und stellt das Löschen ans Ende', () => {
    zeige(paket({}))

    // **Über das Ende der Aufschrift** — der Löschknopf trägt ein Zeichen
    // davor (`Aktionsknopf`).
    const knoepfe = screen.getAllByRole('button')
    const stelle = (name: string) =>
      knoepfe.findIndex((knopf) => knopf.textContent.endsWith(name))

    expect([t.weitergeben, t.zurueck, t.loeschen].map(stelle)).toEqual([
      0, 1, 2,
    ])
  })
})
