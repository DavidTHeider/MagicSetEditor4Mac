// **Der Weg über den Karten nennt die aufgeschlagene** — UX-03.AK-02
// (#1675).
//
// # Wogegen hier geprüft wird
//
// **Gegen „Karten" als letzte Station.** Das Artboard *Karten* sagt
// *Gestalten › Werkstatt der Dämmerung › Aschenwächter*; gebaut stand dort
// bis zum 2026-09-24 der Name der Fläche, weil der Rahmen die gewählte Karte
// nicht kennt. Seit die Seite ihren Weg selbst sagt (`seitenweg.tsx`), steht
// dort die Karte — geprüft durch Rahmen und Seite zusammen.
import { screen, within } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { raeumeApi } from '../test/api'
import { belegeProbeset, zeigeSeite } from '../test/seitenvorrichtung'

const texte = kataloge.de

afterEach(() => {
  raeumeApi()
})

it('nennt die aufgeschlagene Karte als letzte Station', async () => {
  belegeProbeset()

  zeigeSeite('karten', 'de')

  const weg = screen.getByRole('navigation', {
    name: texte.allgemein.navigation.weg,
  })

  // **Die Karte steht da, sobald sie aufgeschlagen ist** — die erste der
  // Liste, wie beim Öffnen eines Sets.
  const letzte = await within(weg).findByText('Feuerball')

  expect(letzte).toHaveAttribute('aria-current', 'page')
  expect(within(weg).queryByText(texte.seiten.karten.titel)).toBeNull()
})
