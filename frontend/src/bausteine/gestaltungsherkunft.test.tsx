// **Die eigenen Farben sagen, woher sie kommen, und lassen sich zurücksetzen**
// — Befund M15 des Zweitblicks über Phase 6.3
// (#1665):
// `gestaltung` zeichnete weder Ebene noch Herkunft noch Wirkzeitpunkt und
// hatte kein „Zurück zur Vorgabe".
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Eigenfarben } from './Gestaltungsbeiwerk'
import { Gestaltungsraum } from '../gestaltungsraum'
import { Nutzereinstellungsraum } from '../nutzereinstellungsraum'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { probeprofil } from '../test/profil'

const eigene = kataloge.de.seiten.einstellungen

afterEach(() => {
  raeumeApi()
})

function zeige(wert: string) {
  belegeApi({
    GezeigtesProfil: vi.fn(() =>
      Promise.resolve({
        profil: probeprofil({
          id: '10',
          nutzerID: '7',
          name: 'P',
          fassung: '1',
        }),
      }),
    ),
    Nutzereinstellungen: vi.fn(() => Promise.resolve({ einstellungen: [] })),
    Fensterbild: vi.fn(() => Promise.resolve({ adresse: '', fehler: null })),
  })
  const setze = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Nutzereinstellungsraum>
        <Gestaltungsraum einstellungen={{ wert: () => '' }} neuLesen={vi.fn()}>
          <Eigenfarben
            gestaltung={{
              schluessel: 'gestaltung',
              wert,
              wirksam: wert,
              werte: [],
              bereich: 'anwendung',
              ebene: 'profil',
              herkunft: wert === '' ? 'vorgabe' : 'profil',
              wirkung: 'sofort',
              vorgabe: '',
            }}
            angehalten={false}
            setze={setze}
          />
        </Gestaltungsraum>
      </Nutzereinstellungsraum>
    </Sprachraum>,
  )

  return setze
}

describe('die Herkunft der eigenen Farben', () => {
  it('nennt die Herkunft als Marke und setzt auf die Vorgabe zurück', async () => {
    const setze = zeige('marke=#112233')

    // **Am Feld und ohne den rohen Wert** — er steht im Feld darüber, und
    // mehrere Zeilen CSS in einem Satz liest niemand.
    const satz = await screen.findByText(
      'Gesetzt von diesem Profil, Ebene Profil.',
    )
    expect(satz.closest('section')).toHaveTextContent(
      eigene.gestaltung.merkmale,
    )
    expect(screen.queryByText(/Es gilt/)).toBeNull()
    expect(screen.getByText('von diesem Profil')).toBeInTheDocument()
    expect(screen.queryByText(/Wirkt sofort/)).toBeNull()

    await userEvent.click(
      screen.getByRole('button', { name: eigene.gestaltung.zurVorgabe }),
    )

    expect(setze).toHaveBeenCalledWith('')
  })

  // **Wo nichts gesetzt ist, gibt es nichts zurückzusetzen.**
  it('zeigt ohne eigene Farben keinen Weg zurück', async () => {
    zeige('')

    // **Gewartet wird auf die Gruppe**, nicht auf einen Satz, den es nicht
    // mehr gibt (AP-11, #2368).
    await screen.findByText(eigene.gestaltung.merkmale)
    expect(
      screen.queryByRole('button', { name: eigene.gestaltung.zurVorgabe }),
    ).toBeNull()
  })
})
