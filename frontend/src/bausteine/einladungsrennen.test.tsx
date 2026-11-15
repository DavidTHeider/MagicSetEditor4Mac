// **Eine verspätete Annahme gehört zu ihrem Stand** — Befund Z6 des
// Zweitblicks und N6 des Nachlaufs über Phase 6.7 (NET-Teil), §7 Regel 2.
//
// Während die Annahme läuft — mit dem Melden beim Einladenden bis zu
// fünfzehn Sekunden —, bleibt das Codefeld offen. **Derselbe Code ein
// zweites Mal ist ein neuer Stand**: wer A, dann B, dann wieder A eingibt und
// neu prüft, hat einen neuen Vergleich vor sich, und die alte Antwort räumt
// ihn nicht — weder mit Erfolg noch mit Fehler. Die Liste lädt nach einer
// geglückten Annahme trotzdem nach (Regel 3).
import { screen } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'

import {
  alsAntwort,
  derCode,
  eigene,
  eineVorschau,
  pruefe,
  stelleAuf,
} from '../test/einladungsprobe'

import type { Einladungsapi } from '../apikontakteinladung'
import type { Peerkontakteantwort } from '../apipeerkontakt'

afterEach(() => {
  raeumeApi()
})

// eineHaengendeAnnahme gibt die Annahme und den Griff, der sie beendet.
function eineHaengendeAnnahme() {
  let beenden: (antwort: Peerkontakteantwort) => void = () => undefined
  const annehmen = vi.fn<Einladungsapi['KontakteinladungAnnehmen']>(
    () =>
      new Promise<Peerkontakteantwort>((erfuellen) => {
        beenden = erfuellen
      }),
  )

  return {
    annehmen,
    beenden: (antwort: Peerkontakteantwort) => {
      beenden(antwort)
    },
  }
}

// vonANachBUndZurueck nimmt an, gibt einen anderen Code ein, dann wieder den
// ersten, und prüft ihn neu — der Vergleich steht frisch da.
async function vonANachBUndZurueck(
  annehmen: Einladungsapi['KontakteinladungAnnehmen'],
) {
  belegeApi({
    KontakteinladungPruefen: () => alsAntwort(eineVorschau()),
    KontakteinladungAnnehmen: annehmen,
  })

  const { hand, nachAufnahme } = stelleAuf()
  await pruefe(hand)
  await hand.click(await screen.findByRole('button', { name: eigene.spaeter }))

  const feld = await screen.findByLabelText(eigene.eingabe)
  await hand.clear(feld)
  await hand.type(feld, 'ein anderer')
  await hand.clear(feld)
  await hand.click(feld)
  await hand.paste(derCode)
  await hand.click(screen.getByRole('button', { name: eigene.pruefen }))
  expect(
    await screen.findByRole('heading', { name: 'Anna bestätigen' }),
  ).toBeInTheDocument()

  return { feld, nachAufnahme }
}

it('räumt den neuen Vergleich nicht mit der alten Antwort ab', async () => {
  const { annehmen, beenden } = eineHaengendeAnnahme()
  const { feld, nachAufnahme } = await vonANachBUndZurueck(annehmen)

  beenden({ kontakte: [], zwecke: [], fehler: null })
  await antwortAbwarten()

  expect(nachAufnahme).toHaveBeenCalled()
  expect(feld).toHaveValue(derCode)
  expect(
    screen.getByRole('heading', { name: 'Anna bestätigen' }),
  ).toBeInTheDocument()
})

it('setzt den Fehler der alten Annahme nicht an den neuen Vergleich', async () => {
  const { annehmen, beenden } = eineHaengendeAnnahme()
  const { nachAufnahme } = await vonANachBUndZurueck(annehmen)

  beenden({
    kontakte: [],
    zwecke: [],
    fehler: { code: 'konflikt', schluessel: '', meldung: 'schon angenommen' },
  })
  await antwortAbwarten()

  // **Geprüft vor jeder weiteren Eingabe** — sonst räumte sie den Befund
  // selbst weg.
  expect(screen.queryByText('schon angenommen')).toBeNull()
  expect(nachAufnahme).not.toHaveBeenCalled()
  expect(annehmen).toHaveBeenCalledTimes(1)
})
