// **Wenn der Wirt wegbricht** — Werkstück G5 von Phase 5.5
// (#1462).
//
// **Aus `partiegast.test.tsx` herausgelöst**, als P5
// (#2373)
// sie über die Zeilengrenze aus §2 brachte. Der Schnitt folgt der Frage:
// dort, wie ein Gast an den Tisch kommt; hier, was er sieht, wenn die Leitung
// abreißt.
import { screen, waitFor } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { antwortAbwarten } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { einePartie } from './partieprobe'
import { ereignisBeitritt, ereignisTischbild } from '../haken/usePartiegast'
import { kataloge } from '../sprache/kataloge'
import {
  gastApi,
  laufzeitbelegen,
  waehleWirt,
  zeigeGast,
} from './partiegastprobe'

import type { PartiebeitrittAntwort } from '../apipartie'

const eigene = kataloge.de.seiten.spielen

// **Ein Abriss ist kein Aufstehen, und das Fenster sagt es.**
//
// # Warum das eine eigene Zusage ist
//
// **Weil ohne sie das leere Formular zurückkäme** — und niemand wüsste, ob er
// hinausgeworfen wurde oder ob die Leitung wegbrach. Der Platz bleibt, und
// fortsetzen kann nur derselbe Wirt (§6, und `App.tischVerloren`).
it('sagt, dass der Wirt weggebrochen ist', async () => {
  gastApi({
    PartieBeitrittStand: vi.fn().mockResolvedValue({
      laeuft: false,
      partie: '',
      gegenueber: '',
      pausiert: true,
      fehler: null,
    }),
  })

  const hand = zeigeGast()

  await hand.click(screen.getByRole('button', { name: eigene.gastNachsehen }))

  expect(await screen.findByText(eigene.gastPausiert)).toBeInTheDocument()
})

// **Und wer gar nicht erst saß, bekommt den Satz nicht zu sehen.**
//
// **Die Gegenprobe** (§10): ein Fenster, das ihn immer zeigte, sagte nichts.
it('schweigt über eine Pause, die es nicht gab', async () => {
  gastApi({
    PartieBeitrittStand: vi.fn().mockResolvedValue({
      laeuft: false,
      partie: '',
      gegenueber: '',
      pausiert: false,
      fehler: null,
    }),
  })

  const hand = zeigeGast()

  await hand.click(screen.getByRole('button', { name: eigene.gastNachsehen }))
  await antwortAbwarten()

  expect(screen.queryByText(eigene.gastPausiert)).not.toBeInTheDocument()
})

// **Lehnt der Wirt ab, liest der Gast seinen Grund** — P9c
// (#2377),
// gefunden am nativen Fenster (`netzpartie09`): bis dahin stand nur der Satz
// über eine abgerissene Leitung da, und auch der erst nach „Nachsehen“.
//
// **Go meldet `partie:beitritt`**, und das Fenster fragt den Stand von
// selbst — ohne Klick.
it('zeigt die Absage des Wirts, sobald Go sie meldet', async () => {
  const horcher = laufzeitbelegen()
  const stand = vi.fn().mockResolvedValue({
    laeuft: false,
    partie: '',
    gegenueber: '',
    pausiert: true,
    absage: 'dieser Tisch wartet auf dein Deck',
    fehler: null,
  })
  gastApi({ PartieBeitrittStand: stand })

  zeigeGast()
  await effekteNachholen()

  await waitFor(() => {
    horcher.get(ereignisBeitritt)?.()
    expect(stand).toHaveBeenCalled()
  })

  expect(await screen.findByText(eigene.gastAbgesagt)).toBeInTheDocument()
  expect(
    screen.getByText('dieser Tisch wartet auf dein Deck'),
  ).toBeInTheDocument()

  // **Eine Absage ist kein Abriss** — der Satz darüber träfe nicht zu.
  expect(screen.queryByText(eigene.gastPausiert)).not.toBeInTheDocument()
})

// **Und der Tisch des Wirts geht mit** — gefunden am nativen Fenster
// (`netzpartie09`): der Wirt grüßt mit seinem Bild, bevor er ablehnt, und
// der Abgewiesene sah danach weiter einen Tisch samt „Aufgeben“, an dem er
// nie saß (§7, Regel 1). Für einen Abriss gilt dasselbe — der Satz darunter
// sagt es seit G5.
it('räumt den Tisch, wenn der Platz verloren ist', async () => {
  const horcher = laufzeitbelegen()
  gastApi({
    PartieBeitreten: vi.fn().mockResolvedValue({
      laeuft: true,
      partie: 'p1',
      gegenueber: 'abcd',
      pausiert: false,
      absage: '',
      fehler: null,
    }),
    PartieBeitrittStand: vi.fn().mockResolvedValue({
      laeuft: false,
      partie: '',
      gegenueber: '',
      pausiert: true,
      absage: 'dieser Tisch wartet auf dein Deck',
      fehler: null,
    }),
  })

  const hand = zeigeGast()
  await waehleWirt(hand)
  await hand.click(screen.getByRole('button', { name: eigene.gastBeitreten }))
  await screen.findByText(eigene.gastSitzt, { exact: false })

  await waitFor(() => {
    horcher.get(ereignisTischbild)?.('p1', {
      partie: einePartie(),
      erlaubt: [],
      handlungen: [],
    })
  })
  expect(
    await screen.findByRole('heading', { name: eigene.tischTitel }),
  ).toBeInTheDocument()

  horcher.get(ereignisBeitritt)?.()

  expect(await screen.findByText(eigene.gastAbgesagt)).toBeInTheDocument()
  expect(
    screen.queryByRole('heading', { name: eigene.tischTitel }),
  ).not.toBeInTheDocument()
})

// **Eine ältere Antwort trägt den verlorenen Platz nicht wieder ein** —
// Nachlauf zum Zweitblick über Phase 7, R6. Der Stand nach der Meldung wird
// am Vorgang vorbei gelesen; kam er mit der Absage an, bevor die Antwort auf
// das Beitreten eintraf, setzte diese den Gast wieder an den Tisch (§7,
// Regel 2).
it('lässt eine verspätete Antwort den verlorenen Platz nicht zurückbringen', async () => {
  const horcher = laufzeitbelegen()
  let beigetreten: (antwort: PartiebeitrittAntwort) => void = () => undefined
  gastApi({
    PartieBeitreten: () =>
      new Promise((fertig) => {
        beigetreten = fertig
      }),
    PartieBeitrittStand: vi.fn().mockResolvedValue({
      laeuft: false,
      partie: '',
      gegenueber: '',
      pausiert: true,
      absage: 'dieser Tisch wartet auf dein Deck',
      fehler: null,
    }),
  })

  const hand = zeigeGast()
  await waehleWirt(hand)
  await hand.click(screen.getByRole('button', { name: eigene.gastBeitreten }))

  horcher.get(ereignisBeitritt)?.()
  expect(await screen.findByText(eigene.gastAbgesagt)).toBeInTheDocument()

  beigetreten({
    laeuft: true,
    partie: 'p1',
    gegenueber: 'abcd',
    pausiert: false,
    absage: '',
    fehler: null,
  })
  await antwortAbwarten()

  expect(screen.getByText(eigene.gastAbgesagt)).toBeInTheDocument()
  expect(
    screen.queryByText(eigene.gastSitzt, { exact: false }),
  ).not.toBeInTheDocument()
})
