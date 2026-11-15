// **Nach dem Anfangen steht der Tisch im Blick** — Befund F09 der nativen
// Prüfung von 0.5.0
// (#1675,
// #1680).
//
// # Wogegen hier geprüft wird
//
// **Gegen den Zustand, den die Prüfung aufgenommen hat**: „Partie anfangen"
// gedrückt, und der erste Zustand danach zeigte den unteren Bereich der
// Seite — der Tisch erschien erst nach dem Zurückscrollen. Erwartet war
// „Neue Partie mit Tisch und aktuellem Zug sichtbar; Fokus an neuem
// Arbeitskontext". Den Fokus prüft dieser Lauf; wohin das Fenster scrollt,
// folgt ihm — ein fokussiertes Element holt WebKit in den Blick.
import { screen } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { antwortAbwarten, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import { element, eineAntwort } from './partieprobe'
import { mitDecks, zeige } from './spielenprobe'

const eigene = kataloge.de.seiten.spielen

afterEach(() => {
  raeumeApi()
})

it('gibt nach dem Anfangen dem Tisch den Fokus', async () => {
  mitDecks({ PartieAnlegen: () => Promise.resolve(eineAntwort()) })

  const hand = zeige()

  await hand.type(await screen.findByLabelText(`${eigene.sitzName} 1`), 'Anna')
  await hand.type(screen.getByLabelText(`${eigene.sitzName} 2`), 'Bert')

  const felder = screen.getAllByLabelText(eigene.sitzDeck)

  await hand.selectOptions(element(felder[0]), '1')
  await hand.selectOptions(element(felder[1]), '2')

  await hand.click(screen.getByRole('button', { name: eigene.anfangen }))

  const kopf = await screen.findByRole('heading', {
    level: 1,
    name: eigene.tischTitel,
  })

  await expect.poll(() => kopf === document.activeElement).toBe(true)

  // **Ohne Ring** — UXR-24 (#1932): die Überschrift nimmt den Fokus, damit ein
  // Vorleseprogramm dort beginnt; bedienen lässt sie sich nicht, und ein Ring
  // um sie sah am Fenster aus wie ein Knopf. Dasselbe Muster wie am Kopf von
  // „Neues Set" und an `main` (`focus:outline-none`).
  expect(kopf.className).toContain('focus:outline-none')
  expect(kopf.className).not.toContain('focus-visible:outline-2')

  // **Und die Seite hat dann keine zweite Überschrift darüber** — wo man
  // ist, sagt die Brotkrume.
  expect(
    screen.queryByRole('heading', { name: eigene.titel }),
  ).not.toBeInTheDocument()

  await antwortAbwarten()
})
