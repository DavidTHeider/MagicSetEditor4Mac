// **Apples Gerätemodell für diese Sitzung** — AI-05.AK-04
// (#2042).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { KIGeraetSitzung } from './KIGeraetSitzung'
import { KIModelle } from './KIModelle'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { KIKatalogantwort } from '../antworten'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.einstellungen

function zeige(an: boolean, gemeldet = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <KIGeraetSitzung an={an} gemeldet={gemeldet} />
    </Sprachraum>,
  )

  return gemeldet
}

// **Ausschalten fragt die Bindung mit `false` und lädt den Katalog neu** —
// gewartet wird auf den Neuladeruf, nicht auf den Klick.
it('schaltet das Gerätemodell für die Sitzung ab', async () => {
  const geschaltet = vi.fn((an: boolean) =>
    Promise.resolve({ an, fehler: null }),
  )
  belegeApi({ KIGeraetemodellFuerSitzung: geschaltet })

  const gemeldet = zeige(true)
  const schalter = screen.getByRole('switch', { name: eigene.kiGeraetSitzung })
  expect(schalter).toBeChecked()

  await userEvent.click(schalter)

  await vi.waitFor(() => {
    expect(gemeldet).toHaveBeenCalledOnce()
  })
  expect(geschaltet).toHaveBeenCalledWith(false)
  expect(screen.getByText(eigene.kiGeraetSitzungHinweis)).toBeInTheDocument()
})

// **Und wieder an** — derselbe Weg in die andere Richtung.
it('schaltet es wieder an', async () => {
  const geschaltet = vi.fn((an: boolean) =>
    Promise.resolve({ an, fehler: null }),
  )
  belegeApi({ KIGeraetemodellFuerSitzung: geschaltet })

  const gemeldet = zeige(false)
  await userEvent.click(
    screen.getByRole('switch', { name: eigene.kiGeraetSitzung }),
  )

  await vi.waitFor(() => {
    expect(gemeldet).toHaveBeenCalledOnce()
  })
  expect(geschaltet).toHaveBeenCalledWith(true)
})

// **Ein Fehler steht am Schalter** (§6), und neu geladen wird nicht (§7,
// Regel 3: nach einem gescheiterten Schreiben nicht nachladen).
it('zeigt einen Fehler und lädt nicht nach', async () => {
  belegeApi({
    KIGeraetemodellFuerSitzung: () =>
      Promise.resolve({
        an: true,
        fehler: {
          code: 'erlaubnis',
          schluessel: 'erlaubnis',
          meldung: 'Nicht erlaubt.',
        },
      }),
  })

  const gemeldet = zeige(true)
  await userEvent.click(
    screen.getByRole('switch', { name: eigene.kiGeraetSitzung }),
  )

  expect(await screen.findByRole('alert')).toBeInTheDocument()
  expect(gemeldet).not.toHaveBeenCalled()
})

// katalog baut eine Antwort mit diesen Anbietern.
function katalog(anbieter: KIKatalogantwort['anbieter']): KIKatalogantwort {
  return {
    nurLokal: true,
    vorrang: [],
    vorrangbefunde: [],
    einrichtung: 'bereit',
    anbieter,
    cloud: { stand: 'ungewiss', beleg: 'keine Serverdatei' },
    befunde: [],
    modelle: [],
    fehler: null,
  }
}

// **Apples Gerätemodell bekommt einen Sitzungsschalter** (AI-05, #2042) —
// abgeschaltet für die Sitzung steht er aus, und abgeschaltet über die
// Einstellung steht er gar nicht da: dort täte er nichts.
it('zeigt den Sitzungsschalter nur, wo er etwas tut', async () => {
  const geraet = {
    name: 'apple',
    einrichtung: 'aus',
    laufzeit: '',
    adresse: '',
    ort: 'lokal',
    befunde: [],
    messbar: [],
    gemessen: [],
    sitzungAus: true,
  }
  belegeApi({
    KIModelle: () => Promise.resolve(katalog([geraet])),
  })

  const { unmount } = render(
    <Sprachraum sprache="de">
      <KIModelle />
    </Sprachraum>,
  )

  expect(
    await screen.findByRole('switch', { name: eigene.kiGeraetSitzung }),
  ).not.toBeChecked()

  // **Und keine Adresse, die es nicht gibt**: das Gerätemodell spricht über
  // einen eigenen Socket — „die Vorgabeadresse auf diesem Rechner" ist der
  // Satz für Ollama (gesehen an der Hüllenprobe zu AI-05).
  expect(screen.getByText(eigene.kiAdresseSocket)).toBeInTheDocument()
  unmount()

  belegeApi({
    KIModelle: () =>
      Promise.resolve(katalog([{ ...geraet, sitzungAus: false }])),
  })

  render(
    <Sprachraum sprache="de">
      <KIModelle />
    </Sprachraum>,
  )

  expect(await screen.findByText(/apple/)).toBeInTheDocument()
  expect(screen.queryByRole('switch')).toBeNull()
})
