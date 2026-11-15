// **Was ein Bildauftrag kostet, wird vorher freigegeben** —
// #1539.
//
// # Warum eine eigene Datei
//
// **Weil `kibild.test.tsx` an der Zeilengrenze aus §2 stand.** Der Schnitt
// folgt der Frage: dort steht, **was** ein Bildauftrag tut, hier, **wer ihn
// bezahlt**.
import { screen } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import {
  bildtexte as eigene,
  einKatalog,
  einVorschlag,
  mitFreigabe,
  ohneNachweis,
  zeigeBildkasten as zeige,
} from '../test/kibild'

afterEach(() => {
  raeumeApi()
})

// **Ein Modell auf diesem Rechner wird nicht freigegeben.**
//
// **Weil eine Freigabe, die überall verlangt wird, keine Auskunft ist**: wer
// sie für ein kostenloses Modell geben müsste, gewöhnte sich das Wegklicken
// an — und dann gälte die Frage auch dort nichts mehr, wo sie zählt.
it('erzeugt ohne Freigabe, wo nichts kostet', async () => {
  const vorschlagen = vi.fn(() => Promise.resolve(einVorschlag()))
  const erteilen = vi.fn()

  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog('lokal')),
      KIBildHerkunft: () => Promise.resolve(ohneNachweis()),
      KIBildVorschlagen: vorschlagen,
      KIFreigabeErteilen: erteilen,
    }),
  )

  const { hand } = zeige()

  await hand.click(screen.getByText(eigene.kiBildTitel))

  await hand.selectOptions(
    await screen.findByLabelText(eigene.kiVorschlagAnbieter),
    'openai/gpt-image-2.5-flare',
  )

  await hand.type(screen.getByLabelText(eigene.kiBildAufgabe), 'ein Blitz')
  await hand.click(screen.getByRole('button', { name: eigene.kiBildErzeugen }))

  // **Erst der Beleg, dass erzeugt wurde** — die Verneinung darunter träfe
  // sonst auch zu, wenn der Knopf nichts ausgelöst hätte (§3).
  await screen.findByRole('figure')

  expect(erteilen).not.toHaveBeenCalled()
})

// **Und wer nicht freigibt, bekommt kein Bild.**
it('erzeugt nichts, wenn niemand freigibt', async () => {
  const vorschlagen = vi.fn()

  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: () => Promise.resolve(ohneNachweis()),
      KIBildVorschlagen: vorschlagen,
    }),
  )

  const { hand } = zeige()

  await hand.click(screen.getByText(eigene.kiBildTitel))

  await hand.selectOptions(
    await screen.findByLabelText(eigene.kiVorschlagAnbieter),
    'openai/gpt-image-2.5-flare',
  )

  await hand.type(screen.getByLabelText(eigene.kiBildAufgabe), 'ein Blitz')
  await hand.click(screen.getByRole('button', { name: eigene.kiBildErzeugen }))

  await hand.click(
    await screen.findByRole('button', { name: eigene.kiFreigabeVerzichten }),
  )

  await vi.waitFor(() => {
    expect(screen.queryByText(eigene.kiFreigabeTitel)).not.toBeInTheDocument()
  })

  expect(vorschlagen).not.toHaveBeenCalled()
})
