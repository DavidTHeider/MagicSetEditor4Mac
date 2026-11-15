// **Das offene Turnier steht vor der Liste** — UXR-12
// (#1920).
//
// # Der Befund
//
// Die Werkbank stand **unter** der Liste aller Turniere. Eine Zeile misst
// mit ihren Marken rund 84 Punkte; bei acht Turnieren begann das offene erst
// unterhalb von 800 Punkten — wer *Öffnen* drückte, sah nichts geschehen, und
// die Liste wächst mit jedem Turnierabend um eine Zeile (R4 in
// `docs/UIREVIEW.md`, und die Seite hat keinen Navigator, Vermerk 0076).
//
// # Was jetzt gilt
//
// Ist ein Turnier offen, steht es unter dem Kopf der Seite und die Liste
// darunter. **Wer es aus der Liste öffnet, landet dort**: der Fokus geht auf
// seinen Namen, und das Fenster rollt mit — sonst stünde er auf einem Knopf,
// über dem gerade eine Werkbank erschienen ist.
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import {
  aufgeschlagen,
  zumTurniergebiet,
  alleTurniere,
} from './turnierwerkbank'
import { raeumeApi } from '../test/api'
import { abend, liste, rangliste } from '../test/turnierbestand'
import { kataloge } from '../sprache/kataloge'
import { effekteNachholen } from '../test/effekte'
import type { Api } from '../api'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

function amAbend(weitere: Partial<Api> = {}) {
  return aufgeschlagen(abend, {
    Turniere: () => Promise.resolve({ turniere: liste, fehler: null }),
    Turnier: (kennung: string) =>
      Promise.resolve({
        turnier: liste.find((turnier) => turnier.id === kennung) ?? null,
        fehler: null,
      }),
    TurnierRangliste: () => Promise.resolve({ rangliste, fehler: null }),
    ...weitere,
  })
}

describe('das offene Turnier und die Liste', () => {
  it('stellt das offene Turnier vor die Liste', async () => {
    const werkbank = await amAbend()
    await alleTurniere()
    const uebersicht = screen.getByRole('region', { name: texte.titel })

    expect(
      werkbank.compareDocumentPosition(uebersicht) &
        Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy()
  })

  it('führt den Fokus zum Turnier, das aus der Liste geöffnet wird', async () => {
    await amAbend()

    await alleTurniere()
    await userEvent.click(
      screen.getByRole('button', { name: `${texte.waehlen}: Märzturnier` }),
    )

    const werkbank = await screen.findByRole('region', { name: 'Märzturnier' })
    await waitFor(() => {
      expect(
        within(werkbank).getByRole('heading', { name: 'Märzturnier' }),
      ).toHaveFocus()
    })
  })

  // **Nur das Öffnen springt.** Nach einem Schreiben lädt die Seite das
  // Turnier nach und baut die Werkbank meistens neu auf; zöge sie dabei den
  // Fokus an den Namen, landete jedes Abmelden oben.
  it('springt nach einem Schreiben nicht noch einmal', async () => {
    const aussteigen = vi.fn<Api['TurnierAussteigen']>(() =>
      Promise.resolve({ fehler: null }),
    )
    const gelesen = vi.fn<Api['Turnier']>(() =>
      Promise.resolve({ turnier: abend, fehler: null }),
    )
    await amAbend({ TurnierAussteigen: aussteigen, Turnier: gelesen })

    // **Erst der Beleg, dass der Sprung greift**: aus der Liste geöffnet,
    // steht der Fokus am Namen.
    await alleTurniere()
    await userEvent.click(
      screen.getByRole('button', { name: `${texte.waehlen}: ${abend.name}` }),
    )
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: abend.name })).toHaveFocus()
    })
    const vorher = gelesen.mock.calls.length

    await zumTurniergebiet('teilnehmer')
    await userEvent.click(
      screen.getByRole('button', { name: `${texte.aussteigen}: Ada` }),
    )
    await waitFor(() => {
      expect(aussteigen).toHaveBeenCalledWith('7', 't1', true)
    })

    // **Gewartet wird auf das Nachladen** — und darauf, dass React danach
    // seine Effekte gezogen hat: dort säße der Sprung.
    await waitFor(() => {
      expect(gelesen.mock.calls.length).toBeGreaterThan(vorher)
    })
    await screen.findByRole('region', { name: abend.name })
    await effekteNachholen()

    expect(screen.getByRole('heading', { name: abend.name })).not.toHaveFocus()
  })
})
