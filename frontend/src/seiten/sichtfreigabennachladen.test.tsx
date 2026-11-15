// **Widerrufen ist widerrufen** — §7, Regel 3
// (#1970).
//
// # Wogegen hier geprüft wird
//
// Go meldet nach dem Widerruf, dass die Liste danach nicht zu lesen war —
// `nachladefehler` neben einer leeren Liste. Der Kasten nimmt den Schritt
// als geschehen und liest selbst nach; was er dann nicht lesen kann, steht
// an der Liste. **Eine eigene Datei**, weil `sichtfreigaben.test.tsx` mit
// diesem Lauf über die Zeilengrenze aus §2 ginge.
import { screen } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { einKontakt, eineListe, zeige } from './peerkontaktprobe'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.technik

// **Hinter dem Aufklapper „Was dieses Gegenüber sieht"** (UXR-16, #1924):
// offen fragte jede bestätigte Zeile schon beim Zeichnen den Bestand; jetzt
// öffnet der Lauf den Kasten, wie ein Mensch es tut.
const aufklapper = eigene.freigabenTitel

afterEach(() => {
  raeumeApi()
})

it('nimmt einen Widerruf mit Nachladefehler als geschehen', async () => {
  const umfaenge = ['auswahl', 'dubletten', 'wuensche', 'sammlung', 'digitales']
  const lesen = vi
    .fn()
    .mockResolvedValueOnce({
      freigaben: [
        {
          empfaenger: 'bgegenueber',
          umfang: 'sammlung',
          zweck: 'handel',
          auswahl: [],
          behalten: 0,
          ablauf: '',
          erteilt: '2026-09-18T12:00:00Z',
        },
      ],
      umfaenge,
      fehler: null,
    })
    .mockResolvedValue({
      freigaben: [],
      umfaenge,
      fehler: { code: 'validierung', schluessel: '', meldung: 'klemmt' },
    })

  belegeApi({
    Peerkontakte: () =>
      Promise.resolve(
        eineListe([einKontakt({ stand: 'bestaetigt', rechte: ['handel'] })]),
      ),
    Sammlungsfreigaben: lesen,

    // **Die Wahl der Drucke liest die Sammlung** (UXR-F18, #2033) — ohne
    // Beleg stünde ein zweiter Alarm neben dem, um den es geht.
    Sammlung: () => Promise.resolve({ exemplare: [], fehler: null }),
    SammlungsfreigabeWiderrufen: () =>
      Promise.resolve({
        freigaben: [],
        umfaenge,
        nachladefehler: {
          code: 'validierung' as const,
          schluessel: '',
          meldung: 'klemmt',
        },
        fehler: null,
      }),
  })

  const hand = await zeige(aufklapper)

  await hand.click(
    await screen.findByRole('button', { name: eigene.freigabenWiderrufen }),
  )

  expect(await screen.findByRole('alert')).toHaveTextContent('klemmt')
  expect(lesen).toHaveBeenCalledTimes(2)
})
