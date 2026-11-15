// **Anwendung und Webprozess, getrennt gemessen** — OPS02 (#2348),
// Abnahmekriterium 2: nicht der Go-Heap als Gesamtspeicher, und was fehlt,
// steht ausdrücklich da.
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Betriebszahlen } from './Betriebszahlen'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { volleMetriken } from '../test/metriken'

import type { ProzessDTO } from '../apimetriken'

const texte = kataloge.de.seiten.technik.betrieb
const mib = 1024 * 1024

function zeige(prozess: Partial<ProzessDTO>, ungemessen?: string[]) {
  const voll = volleMetriken()

  render(
    <Sprachraum sprache="de">
      <Betriebszahlen
        zahlen={{
          ...voll,
          prozess: { ...voll.prozess, ...prozess },
          ungemessen: ungemessen ?? voll.ungemessen,
        }}
      />
    </Sprachraum>,
  )
}

describe('Der Speicher in den Betriebszahlen', () => {
  it('nennt den Fußabdruck der Anwendung und den ihrer Webprozesse', () => {
    zeige({
      fussabdruckBytes: 412 * mib,
      web: [
        { art: 'webinhalt', fussabdruckBytes: 300 * mib },
        { art: 'gpu', fussabdruckBytes: 80 * mib },
      ],
    })

    expect(screen.getByText(/412,0 MB Fußabdruck/)).toBeInTheDocument()
    expect(screen.getByText(/380,0 MB zusammen/)).toHaveTextContent(
      /Webinhalt 300,0 MB.*Grafik 80,0 MB/,
    )
  })

  it('sagt, warum die Webprozesse nicht gemessen sind', () => {
    zeige({ fussabdruckBytes: mib, web: [], webUngemessen: 'keine Zuordnung' })

    expect(screen.getByText(/keine Zuordnung/)).toBeInTheDocument()
  })

  it('sagt ohne Webansicht, dass es keine gibt', () => {
    zeige({ fussabdruckBytes: mib, web: [], webUngemessen: '' })

    expect(screen.getByText(texte.ohneWeb)).toBeInTheDocument()
  })

  it('gibt ohne Fußabdruck nicht den Heap als Speicher aus', () => {
    zeige({ fussabdruckBytes: 0, web: [], webUngemessen: 'nichts' }, [
      'speicher',
    ])

    expect(screen.queryByText(/Fußabdruck —/)).toBeNull()
    expect(
      screen.getByText(
        new RegExp(texte.ungemessenNamen.speicher.replace(/[()]/g, '.')),
      ),
    ).toBeInTheDocument()
  })
})
