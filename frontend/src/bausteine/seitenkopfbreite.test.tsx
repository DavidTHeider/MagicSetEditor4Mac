// **Die Knopfzeile des Seitenkopfs reicht nie über den Kopf hinaus** — QA01
// (#2382),
// gefunden in der WebKit-Abnahme `ux14-groessen`: bei sehr großer Schrift
// stand »Bei Bo mitspielen« auf der Decks-Seite halb außerhalb des Fensters,
// und die Seite rollte waagerecht.
//
// **jsdom misst nichts**, deshalb hält dieser Test die Klasse und nicht die
// Breite: `shrink-0` ließ die Zeile so breit wie alle Knöpfe nebeneinander,
// und ihr eigenes `flex-wrap` griff nie. Erst die Obergrenze lässt es
// greifen. Gemessen wird am gerenderten Element in `ux14-groessen`.
import { render, screen } from '@testing-library/react'
import { expect, it } from 'vitest'

import { Seitenkopf } from './Seitenkopf'
import { Sprachraum } from '../sprache/kontext'

it('begrenzt die Knopfzeile auf die Breite des Kopfs, damit sie umbricht', () => {
  render(
    <Sprachraum sprache="de">
      <Seitenkopf titel="Decks">
        <button type="button">Bei Bo mitspielen</button>
      </Seitenkopf>
    </Sprachraum>,
  )

  const zeile = screen.getByRole('button', {
    name: 'Bei Bo mitspielen',
  }).parentElement

  expect(zeile).toHaveClass('flex-wrap', 'max-w-full')
})
