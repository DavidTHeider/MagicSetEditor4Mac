// **Das Band nach einem Abbruch** — F04
// (#2343).
//
// Endete die letzte Sitzung nicht sauber, sagt es das Fenster beim nächsten
// Start — wann sie begann, ob eine Absturzspur da ist, und der Weg in die
// Diagnose. Ohne Abbruch steht nichts da.
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Anschlussraum, useAnschluss } from '../anschlussraum'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Wiederanlaufband } from './Wiederanlaufband'

import type { AbbruchDTO } from '../apiwiederanlauf'
import type { FehlerDTO } from '../formen'

const de = kataloge.de
const eigene = de.allgemein.wiederanlauf

afterEach(() => {
  raeumeApi()
})

// Unterwegs zeigt, was der Anschlussraum gerade trägt.
function Unterwegs() {
  const offen = useAnschluss().offen

  return (
    <output data-testid="unterwegs">
      {offen?.art === 'einstellungsgruppe' ? offen.gruppe : ''}
    </output>
  )
}

function zeige(abbrueche: AbbruchDTO[], fehler: FehlerDTO | null = null) {
  belegeApi({
    Wiederanlauf: () => Promise.resolve({ abbrueche, fehler }),
  })
  const wechseln = vi.fn()
  render(
    <Sprachraum sprache="de">
      <Anschlussraum aktuell="start" wechseln={wechseln}>
        <Wiederanlaufband />
        <Unterwegs />
      </Anschlussraum>
    </Sprachraum>,
  )

  return wechseln
}

// beantwortet wartet, bis die Antwort eingetragen ist — **die Wirkung und
// nicht den Ruf** (§3): vorher ist das Band auch dann leer, wenn es gleich
// etwas zeigen wird.
async function beantwortet() {
  await waitFor(() => {
    expect(
      screen.getByRole('status', { name: eigene.bereich }),
    ).toHaveAttribute('aria-busy', 'false')
  })
}

const mitSpur: AbbruchDTO = {
  gestartet: '2026-10-14T07:30:00Z',
  fassung: '0.9.0',
  mitSpur: true,
}

describe('Das Band nach einem Abbruch', () => {
  it('nennt den Beginn, sagt, dass eine Spur da ist, und führt in die Diagnose', async () => {
    const wechseln = zeige([mitSpur])

    expect(await screen.findByText(eigene.mitSpur)).toBeInTheDocument()
    expect(
      screen.getByRole('status', { name: eigene.bereich }),
    ).toHaveTextContent(/2026/)

    fireEvent.click(
      screen.getByRole('button', {
        name: de.bausteine.diagnoseverweis.oeffnen,
      }),
    )
    expect(wechseln).toHaveBeenCalledWith('einstellungen')
    expect(screen.getByTestId('unterwegs')).toHaveTextContent('diagnose')
  })

  it('sagt ohne Spur, dass es wohl ein Ende von außen war, und lässt sich ausblenden', async () => {
    zeige([{ ...mitSpur, mitSpur: false }])

    expect(await screen.findByText(eigene.ohneSpur)).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: eigene.ausblenden }))
    expect(screen.queryByText(eigene.ohneSpur)).toBeNull()
  })

  it('nennt den jüngsten Beginn und eine Spur, wenn irgendeiner sie hat', async () => {
    zeige([
      { ...mitSpur, gestartet: '2026-10-13T09:00:00Z' },
      { ...mitSpur, gestartet: '2026-10-14T09:00:00Z', mitSpur: false },
      { ...mitSpur, gestartet: '2026-10-12T09:00:00Z', mitSpur: false },
    ])

    const bereich = await screen.findByRole('status', { name: eigene.bereich })
    expect(await screen.findByText(eigene.mitSpur)).toBeInTheDocument()
    expect(bereich).toHaveTextContent(/14/)
    expect(bereich).not.toHaveTextContent(/13\./)
  })

  it('sagt den Satz ohne Zeit, wenn die Marke unlesbar war', async () => {
    zeige([{ gestartet: '0001-01-01T00:00:00Z', fassung: '', mitSpur: false }])

    expect(await screen.findByText(eigene.satzOhneZeit)).toBeInTheDocument()
  })

  it('zeigt nichts ohne Abbruch', async () => {
    zeige([])

    await beantwortet()
    expect(
      screen.getByRole('status', { name: eigene.bereich }),
    ).toBeEmptyDOMElement()
  })

  it('zeigt nichts, wenn die Frage scheitert', async () => {
    zeige([mitSpur], {
      code: 'abgebrochen',
      schluessel: 'abgebrochen',
      meldung: 'abgebrochen',
      werte: {},
    } as unknown as FehlerDTO)

    await beantwortet()
    expect(
      screen.getByRole('status', { name: eigene.bereich }),
    ).toBeEmptyDOMElement()
  })
})
