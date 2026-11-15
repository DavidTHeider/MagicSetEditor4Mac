// **Ob in diesem Profil geschrieben werden darf** — G10
// (#2276).
//
// B03 hat jede Seite mit einem Leser geöffnet, und auf sechzehn standen
// Schreibhandlungen bedienbar da. Die Regel kommt aus **einer** Stelle: der
// Raum fragt den Arbeitskontext je Profilstand, und jeder Knopf, der
// schreibt, liest ihn.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { ArbeitskontextAntwort } from '../apikontext'
import { Knopf } from '../bausteine/Knopf'
import { Nurlesehinweis } from '../bausteine/Nurlesehinweis'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { Schreibrecht } from './schreibrecht'

afterEach(raeumeApi)

const satz = kataloge.de.bausteine.nurlesen.knopf

function kontext(nurLesend: boolean): ArbeitskontextAntwort {
  return {
    nutzerID: '1',
    profilID: '2',
    eignerID: '3',
    rolle: nurLesend ? 'leser' : 'verwalter',
    nurLesend,
    modus: 'fenster',
    revision: 4,
    fehler: null,
  }
}

function baum(stand: number, geklickt: string[]) {
  return (
    <Sprachraum sprache="de">
      <Schreibrecht stand={stand}>
        <Nurlesehinweis />
        <Knopf
          schreibt
          aufKlick={() => {
            geklickt.push('anlegen')
          }}
        >
          Anlegen
        </Knopf>
        <Knopf
          aufKlick={() => {
            geklickt.push('ansehen')
          }}
        >
          Ansehen
        </Knopf>
      </Schreibrecht>
    </Sprachraum>
  )
}

describe('Das Schreibrecht', () => {
  it('sperrt für einen Leser jeden Knopf, der schreibt, und erklärt warum', async () => {
    const gefragt = vi.fn().mockResolvedValue(kontext(true))
    belegeApi({ Arbeitskontext: gefragt })
    const geklickt: string[] = []

    render(baum(0, geklickt))

    // **Der Streifen liest denselben Raum** — einmal gefragt, nicht je Stelle.
    expect(await screen.findByRole('status')).toHaveTextContent(
      /als Leser: du kannst alles ansehen und nichts ändern/,
    )

    const anlegen = screen.getByRole('button', { name: 'Anlegen' })
    expect(anlegen).toHaveAttribute('aria-disabled', 'true')
    expect(anlegen).toHaveAccessibleDescription(satz)
    expect(anlegen).toBeEnabled()

    // **Mit der Tastatur erreichbar und doch ohne Wirkung**: wer ihn mit Tab
    // erreicht, hört den Grund — `disabled` nähme ihn aus der Reihe.
    await userEvent.tab()
    expect(anlegen).toHaveFocus()
    await userEvent.keyboard('{Enter}')
    await userEvent.click(anlegen)

    // **Der Knopf daneben, der nicht schreibt, geht** — und ist der Beleg,
    // dass jemand zuhört (§3).
    const ansehen = screen.getByRole('button', { name: 'Ansehen' })
    expect(ansehen).not.toHaveAttribute('aria-disabled')
    await userEvent.click(ansehen)

    expect(geklickt).toEqual(['ansehen'])
    expect(gefragt).toHaveBeenCalledTimes(1)
  })

  it('fragt je Profilstand neu, ohne den Baum darunter abzubauen', async () => {
    const gefragt = vi
      .fn()
      .mockResolvedValueOnce(kontext(true))
      .mockResolvedValueOnce(kontext(false))
    belegeApi({ Arbeitskontext: gefragt })
    const geklickt: string[] = []

    const { rerender } = render(baum(0, geklickt))
    await screen.findByRole('status')
    const vorher = screen.getByRole('button', { name: 'Anlegen' })

    rerender(baum(1, geklickt))

    await waitFor(() => {
      expect(screen.queryByRole('status')).not.toBeInTheDocument()
    })

    const nachher = screen.getByRole('button', { name: 'Anlegen' })
    expect(nachher).toBe(vorher)
    expect(nachher).not.toHaveAttribute('aria-disabled')
    expect(nachher).not.toHaveAccessibleDescription(satz)

    await userEvent.click(nachher)
    expect(geklickt).toEqual(['anlegen'])
    expect(gefragt).toHaveBeenCalledTimes(2)
  })

  // **Eine Antwort gehört zu dem Stand, aus dem sie stammt** (§7, Regel 2):
  // die späte Antwort des alten Profils sperrte sonst das neue.
  it('lässt die späte Antwort eines verlassenen Profils nicht ein', async () => {
    let altesProfil: (antwort: ArbeitskontextAntwort) => void = () => {
      throw new Error('noch nicht gefragt')
    }
    const gefragt = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise<ArbeitskontextAntwort>((fertig) => {
            altesProfil = fertig
          }),
      )
      .mockResolvedValueOnce(kontext(false))
    belegeApi({ Arbeitskontext: gefragt })
    const geklickt: string[] = []

    const { rerender } = render(baum(0, geklickt))
    rerender(baum(1, geklickt))

    await waitFor(() => {
      expect(gefragt).toHaveBeenCalledTimes(2)
    })

    altesProfil(kontext(true))
    await new Promise((fertig) => {
      setTimeout(fertig, 0)
    })

    await userEvent.click(screen.getByRole('button', { name: 'Anlegen' }))
    expect(geklickt).toEqual(['anlegen'])
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  // **Ein Fehlschlag sperrt nichts**: er behauptete sonst eine Sperre, die es
  // vielleicht nicht gibt. Die verlässliche Auskunft bleibt die Absage am Tor.
  it('sperrt bei einem Lesefehler nichts', async () => {
    const gefragt = vi.fn().mockRejectedValue(new Error('weg'))
    belegeApi({ Arbeitskontext: gefragt })
    const geklickt: string[] = []

    render(baum(0, geklickt))

    await waitFor(() => {
      expect(gefragt).toHaveBeenCalled()
    })
    // **Auf die Absage und nicht auf den Ruf** — erst dann kann sie etwas
    // gesperrt haben, und erst dann sagt der Klick, dass sie es nicht tat.
    await antwortAbwarten()
    await userEvent.click(screen.getByRole('button', { name: 'Anlegen' }))

    expect(geklickt).toEqual(['anlegen'])
  })

  it('trägt nach dem Abbau nichts mehr ein', async () => {
    const gefragt = vi.fn().mockResolvedValue(kontext(true))
    belegeApi({ Arbeitskontext: gefragt })

    const { unmount } = render(baum(0, []))
    unmount()

    await new Promise((fertig) => {
      setTimeout(fertig, 0)
    })
    expect(gefragt).toHaveBeenCalledTimes(1)
  })
})
