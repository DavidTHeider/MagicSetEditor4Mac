// **Zweck und Umfang einer Sicht in der Sprache des Fensters** — UXR-16
// (#1924).
//
// # Der Befund
//
// **Die Sichten standen mit den Schlüsseln aus Go da**: „handel · dubletten"
// in der Liste, „teilen", „partie", „auswahl", „digitales" als Wahlen. Der
// Zweck hat seinen Satz längst (`zwecksatz`, derselbe wie an den Schaltern
// darüber); der Umfang bekommt seinen jetzt — und einer, den dieses Fenster
// nicht kennt, steht weiter als Schlüssel da.
import { screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { einKontakt, eineListe, zeige } from './peerkontaktprobe'
import { kataloge } from '../sprache/kataloge'
import { umfaenge, zweiSichten } from '../test/gegenueberbestand'

const eigene = kataloge.de.seiten.technik
const namen = eigene.gegenueber.umfaenge

afterEach(() => {
  raeumeApi()
})

describe('eine erteilte Sicht', () => {
  it('nennt Zweck und Umfang beim Namen', async () => {
    belegeApi({
      Peerkontakte: () =>
        Promise.resolve(
          eineListe([einKontakt({ stand: 'bestaetigt', rechte: ['handel'] })]),
        ),
      Sammlungsfreigaben: () =>
        Promise.resolve({
          freigaben: zweiSichten('bgegenueber'),
          umfaenge: [...umfaenge, 'leihgaben'],
          fehler: null,
        }),
    })

    await zeige(eigene.freigabenTitel)

    // **Die Dubletten nennen, was jemand behält** (UXR-F18, #2033) — die
    // Zeile liest sich wie das, was erteilt ist.
    const behalten = kataloge.de.bausteine.druckwahl.detailBehalten.replace(
      '{behalten}',
      '4',
    )

    expect(
      await screen.findByText(
        `${eigene.zweckHandel} · ${namen.dubletten} · ${behalten}`,
      ),
    ).toBeInTheDocument()
    expect(
      screen.getByText(`${eigene.zweckHandel} · ${namen.wuensche}`),
    ).toBeInTheDocument()

    const umfang = screen.getByLabelText(eigene.freigabenUmfang)

    expect(
      within(umfang)
        .getAllByRole('option')
        .map((wahl) => wahl.textContent),
    ).toEqual([
      namen.auswahl,
      namen.dubletten,
      namen.wuensche,
      namen.sammlung,
      namen.digitales,
      'leihgaben',
    ])
  })

  // **`paarung` steht nicht zur Wahl** — das ist die Entscheidung.
  it('bietet paarung nicht als Zweck an', async () => {
    belegeApi({
      Peerkontakte: () =>
        Promise.resolve(
          eineListe([einKontakt({ stand: 'bestaetigt', rechte: ['handel'] })]),
        ),
      Sammlungsfreigaben: () =>
        Promise.resolve({ freigaben: [], umfaenge, fehler: null }),
    })

    await zeige(eigene.freigabenTitel)

    const wahl = await screen.findByLabelText(eigene.freigabenZweck)

    // **Beim Satz gesucht und nicht beim Schlüssel** — die Wahl nennt den
    // Zweck seit UXR-16 wie die Schalter darüber. (Aus `sichtfreigaben.test.tsx`
    // hierher gezogen, als die Namen dazukamen und die Datei an die
    // Zeilengrenze stieß.)
    expect(
      within(wahl).queryByRole('option', { name: eigene.zweckPaarung }),
    ).toBeNull()
    expect(
      within(wahl).getByRole('option', { name: eigene.zweckHandel }),
    ).toBeTruthy()

    // **Und `teilen` auch nicht** — dieses Gegenüber darf es nicht, und eine
    // Sicht dafür gälte nie.
    expect(
      within(wahl).queryByRole('option', { name: eigene.zweckTeilen }),
    ).toBeNull()
  })
})
