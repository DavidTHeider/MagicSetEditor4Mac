// **Der Stand eines Datei-Laufs, wo er angestoßen wurde** — FMT-03.AK-02
// (#1692).
import { act, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { Laufstand } from './Laufstand'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { laufzeitprobe, type Laufzeitprobe } from '../test/laufzeitprobe'
import { ereignisAuftrag } from '../fenster'
import type { AuftragDTO } from '../apiaufmerksamkeit'

const texte = kataloge.de.bausteine.laufstand

let probe: Laufzeitprobe | undefined

afterEach(() => {
  probe?.raeume()
  raeumeApi()
})

function auftrag(teile: Partial<AuftragDTO>): AuftragDTO {
  return {
    id: '4',
    art: 'bogen',
    was: 'Mein Bogen.pdf',
    fortschritt: '',
    stand: 'laeuft',
    grund: '',
    begonnen: '2026-09-29T10:00:00Z',
    sekunden: 1,
    erledigt: 0,
    gesamt: 0,
    einheit: 'karte',
    abbrechbar: true,
    ...teile,
  }
}

async function aufstellen(sprache: 'de' | 'en' = 'de') {
  probe = laufzeitprobe()
  const gesehen = render(
    <Sprachraum sprache={sprache}>
      <Laufstand art="bogen" />
    </Sprachraum>,
  )
  await effekteNachholen()

  return gesehen
}

// nichtsDa sagt, dass kein Stand zu sehen ist — Sprachraum legt selbst
// einen Knoten an, der Behälter ist deshalb nie leer.
function nichtsDa() {
  expect(screen.queryByRole('status')).not.toBeInTheDocument()
  expect(screen.queryByRole('progressbar')).not.toBeInTheDocument()
}

function melde(stand: AuftragDTO) {
  act(() => {
    probe?.melde(ereignisAuftrag, stand)
  })
}

describe('Laufstand', () => {
  it('zeigt nichts, bevor ein Lauf seiner Art beginnt — und hört nur auf seine', async () => {
    await aufstellen()
    nichtsDa()

    // **Ein Lauf einer anderen Art bleibt unsichtbar** — und gleich danach
    // zeigt ein eigener, dass jemand zuhört (§3).
    melde(auftrag({ art: 'set', erledigt: 1, gesamt: 2 }))
    nichtsDa()

    melde(auftrag({ erledigt: 3, gesamt: 9 }))
    expect(screen.getByRole('status')).toHaveTextContent('3 von 9 Karten')
    expect(screen.getByRole('progressbar')).toHaveAttribute(
      'aria-valuenow',
      '3',
    )
  })

  it('zählt ohne Gesamtzahl und ohne vorgetäuschten Stand', async () => {
    await aufstellen()

    melde(auftrag({ erledigt: 1, gesamt: 0, einheit: 'bild' }))
    expect(screen.getByRole('status')).toHaveTextContent('1 Bild')
    expect(screen.getByRole('progressbar')).not.toHaveAttribute('aria-valuenow')

    melde(auftrag({ erledigt: 0, gesamt: 0, einheit: '' }))
    expect(screen.getByRole('status')).toHaveTextContent(texte.laeuft)
  })

  it('hält den Lauf an und sagt danach, was geblieben ist', async () => {
    const AuftragAbbrechen = vi.fn(() =>
      Promise.resolve({ auftraege: [], laufend: 0, fehler: null }),
    )
    belegeApi({ AuftragAbbrechen })
    await aufstellen()

    melde(auftrag({ erledigt: 2, gesamt: 5, einheit: 'paket' }))
    expect(screen.getByRole('status')).toHaveTextContent('2 von 5 Paketen')

    fireEvent.click(screen.getByRole('button', { name: texte.abbrechen }))
    expect(AuftragAbbrechen).toHaveBeenCalledWith('4')
    expect(
      await screen.findByRole('button', { name: texte.wirdAngehalten }),
    ).toBeDisabled()

    melde(auftrag({ stand: 'abgebrochen', abbrechbar: false }))
    expect(screen.getByRole('status')).toHaveTextContent(texte.abgebrochen)

    // **Und was bis dahin fertig wurde** — hier drei Bilder einer Vorlage.
    melde(
      auftrag({
        stand: 'abgebrochen',
        erledigt: 3,
        gesamt: 0,
        einheit: 'bild',
      }),
    )
    expect(screen.getByRole('status')).toHaveTextContent('Bis dahin: 3 Bilder.')
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })

  it('meldet ein gescheitertes Anhalten am Knopf', async () => {
    belegeApi({
      AuftragAbbrechen: () =>
        Promise.resolve({
          auftraege: [],
          laufend: 0,
          fehler: {
            code: 'nichtGefunden',
            schluessel: '',
            meldung: 'der Auftrag läuft nicht mehr',
          },
        }),
    })
    await aufstellen()

    melde(auftrag({ erledigt: 1, gesamt: 3 }))
    fireEvent.click(screen.getByRole('button', { name: texte.abbrechen }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'der Auftrag läuft nicht mehr',
    )
    expect(
      screen.getByRole('button', { name: texte.abbrechen }),
    ).not.toBeDisabled()
  })

  it('verschwindet mit einem fertigen Lauf und übergeht Unlesbares', async () => {
    await aufstellen('en')

    act(() => {
      probe?.melde(ereignisAuftrag, 'kein Auftrag')
      probe?.melde(ereignisAuftrag)
    })
    nichtsDa()

    melde(auftrag({ erledigt: 1, gesamt: 1, abbrechbar: false }))
    expect(screen.getByRole('status')).toHaveTextContent('1 of 1 card')
    expect(screen.queryByRole('button')).not.toBeInTheDocument()

    melde(auftrag({ stand: 'fertig' }))
    nichtsDa()
  })

  it('meldet sich ab, wenn die Seite geht', async () => {
    const { unmount } = await aufstellen()
    expect(probe?.horchende(ereignisAuftrag)).toBe(1)

    unmount()
    expect(probe?.horchende(ereignisAuftrag)).toBe(0)
  })
})
