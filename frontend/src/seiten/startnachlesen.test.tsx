// **Der Überblick bleibt beim Nachlesen stehen** — U10
// (#2125).
//
// Der Stand des Bestands stand in den Abhängigkeiten der Ladefrage, und eine
// neue Frage räumt: bei jedem fremden oder eigenen Schreiben verschwand der
// Überblick, bis die Antwort kam — obwohl der Kommentar darüber sagt, er
// stehe immer.
import { render, screen, within } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Start } from './Start'

const texte = kataloge.de.seiten.start

afterEach(() => {
  raeumeApi()
})

it('lässt den Überblick stehen, während er nachliest', async () => {
  const seite = (stand: number) => (
    <Sprachraum sprache="de">
      <Start wechseln={vi.fn()} waehleSet={vi.fn()} bestandsstand={stand} />
    </Sprachraum>
  )

  const { rerender } = render(seite(0))

  // **Die Zahlen und nicht die Liste** — seit AP-4 steht die Liste auch
  // beim Laden da, mit Platzhaltern; „sie steht noch“ sagte damit nichts
  // mehr darüber, ob der Überblick beim Nachlesen stehen bleibt.
  const zeile = screen.getByRole('list', { name: texte.ueberblick.titel })
  const vorher = await within(zeile).findAllByRole('button')

  // **Das Nachlesen hängt** — so lange, wie ein Bestand gesperrt ist.
  belegeApi({ Sets: () => new Promise(() => undefined) })

  rerender(seite(1))

  expect(
    within(
      screen.getByRole('list', { name: texte.ueberblick.titel }),
    ).getAllByRole('button'),
  ).toHaveLength(vorher.length)
})
