// **Der Kampf in der Handlungsleiste** —
// #1623.
//
// **Eine eigene Datei**, seit die Leiste ihre Felder erst auf Zuruf zeigt
// (UX-08, #1680):
// jeder Lauf klappt vorher auf, und zusammen gingen beide über die
// Zeilengrenze aus §2.
//
// **Geprüft wird die Gestalt, in der der Kern liest** — ein Angriff als Paar,
// ein Block mit seinem Angreifer, und keine Zeile für eine leere Wahl.
import { screen, within } from '@testing-library/react'
import { describe, expect, test } from 'vitest'

import { einePartie, einZug } from '../seiten/partieprobe'
import { mitTypen } from '../test/planprobe'
import {
  eigene,
  eineKarte,
  oeffne,
  stelle,
  zeige,
} from '../test/handlungsprobe'

describe('der Kampf', () => {
  test('ein Angriff geht als Paar hinaus', async () => {
    const partie = einePartie({
      objekte: [eineKarte({ beherrscher: 'sitz1' })],
    })

    const { hand, befehlen } = zeige({ erlaubt: ['kampf.angreifer'], partie })

    await oeffne(hand, 'kampf.angreifer')
    await hand.selectOptions(screen.getByLabelText('Wald'), 'sitz2')
    await stelle(hand, 'kampf.angreifer')

    expect(befehlen).toHaveBeenCalledWith('kampf.angreifer', {
      angriffe: [{ kreatur: 'sitz1/k1/0', ziel: 'sitz2' }],
    })
  })

  test('ein Block nennt den Angreifer', async () => {
    const partie = einePartie({
      objekte: [
        eineKarte({ beherrscher: 'sitz1' }),
        eineKarte({
          kennung: 'sitz2/greif/0',
          name: 'Greif',
          beherrscher: 'sitz2',
        }),
      ],
      zug: einZug({
        angriffe: [{ kreatur: 'sitz2/greif/0', ziel: 'sitz1' }],
      }),
    })

    const { hand, befehlen } = zeige({ erlaubt: ['kampf.blocker'], partie })

    await oeffne(hand, 'kampf.blocker')

    // **Der Angreifer steht mit seinem Namen da** und nicht mit seiner
    // Kennung: wer blockt, sucht die Karte, die er vor sich sieht.
    expect(
      within(screen.getByLabelText('Wald')).getByRole('option', {
        name: 'Greif',
      }),
    ).toBeInTheDocument()

    await hand.selectOptions(screen.getByLabelText('Wald'), 'sitz2/greif/0')
    await stelle(hand, 'kampf.blocker')

    expect(befehlen).toHaveBeenCalledWith('kampf.blocker', {
      bloecke: [{ kreatur: 'sitz1/k1/0', angreifer: 'sitz2/greif/0' }],
    })
  })

  test('eine zurückgenommene Wahl verschwindet wieder', async () => {
    const partie = einePartie({
      objekte: [eineKarte({ beherrscher: 'sitz1' })],
    })

    const { hand, befehlen } = zeige({ erlaubt: ['kampf.angreifer'], partie })

    await oeffne(hand, 'kampf.angreifer')
    await hand.selectOptions(screen.getByLabelText('Wald'), 'sitz2')
    await hand.selectOptions(screen.getByLabelText('Wald'), '')
    await stelle(hand, 'kampf.angreifer')

    // **Eine leere Wahl ist keine Zeile** — CR 508.1a lässt „keine
    // Angreifer" ausdrücklich zu, und das ist etwas anderes als ein
    // Angriff auf niemanden.
    expect(befehlen).toHaveBeenCalledWith('kampf.angreifer', {
      angriffe: [],
    })
  })

  test('ein Angreifer ohne Karte steht mit seiner Kennung da', async () => {
    const { hand } = zeige({
      erlaubt: ['kampf.blocker'],
      partie: einePartie({
        objekte: [eineKarte({ beherrscher: 'sitz1' })],
        zug: einZug({
          angriffe: [{ kreatur: 'weg/greif/0', ziel: 'sitz1' }],
        }),
      }),
    })

    await oeffne(hand, 'kampf.blocker')

    // **Ein Befund und kein Fehler** (§8): der Kampfstand führt eine
    // Kennung, die der Tisch nicht mehr zeigt — etwa weil sie verdeckt
    // liegt. Die Kennung ist besser als eine leere Zeile.
    expect(
      within(screen.getByLabelText('Wald')).getByRole('option', {
        name: 'weg/greif/0',
      }),
    ).toBeInTheDocument()
  })

  test('ohne eigene Kreatur steht der Satz', async () => {
    const { hand } = zeige({
      erlaubt: ['kampf.angreifer'],
      partie: einePartie({ objekte: [] }),
    })

    await oeffne(hand, 'kampf.angreifer')

    expect(screen.getByText(eigene.kampfKeineKreatur)).toBeInTheDocument()
  })

  test('und ohne Angreifer steht der andere', async () => {
    const { hand } = zeige({
      erlaubt: ['kampf.blocker'],
      partie: einePartie({
        objekte: [eineKarte({ beherrscher: 'sitz1' })],
      }),
    })

    await oeffne(hand, 'kampf.blocker')

    expect(screen.getByText(eigene.kampfKeinAngreifer)).toBeInTheDocument()
  })

  // **Nur Kreaturen greifen an und blocken** — CR 508.1a und CR 509.1a
  // (Fassung 2026-09-25), #2390: hier standen alle eigenen Bleibenden,
  // und ein Land war die erste Zeile.
  test('ein Land steht nicht zur Wahl, eine Kreatur schon', async () => {
    const partie = einePartie({
      objekte: [
        eineKarte({
          kennung: 'sitz1/boden/0',
          name: 'Waldboden',
          beherrscher: 'sitz1',
          eigenschaften: mitTypen('Waldboden', ['land']).eigenschaften,
        }),
        eineKarte({
          kennung: 'sitz1/baer/0',
          name: 'Bär',
          beherrscher: 'sitz1',
          eigenschaften: mitTypen('Bär', ['creature']).eigenschaften,
        }),
      ],
    })

    const { hand } = zeige({ erlaubt: ['kampf.angreifer'], partie })
    const felder = await oeffne(hand, 'kampf.angreifer')

    expect(within(felder).getByLabelText('Bär')).toBeInTheDocument()
    expect(within(felder).queryByLabelText('Waldboden')).toBeNull()
  })
})
