// **Die Vorrichtung der Kontaktseite** — ID-04
// (#1703):
// Kontakte, Karten, Belegung und Zeichnen, geteilt von `kontakte.test.tsx`
// und `kontaktwege.test.tsx`.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { Adressbuch } from '../seiten/Adressbuch'
import type { KartenvorschauDTO, KontaktDTO } from '../apikontakte'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi } from './api'

export const texte = kataloge.de.seiten.adressbuch

export function kontakt(teil: Partial<KontaktDTO> = {}): KontaktDTO {
  return {
    id: 0,
    fassung: 0,
    name: 'Ada Lovelace',
    wege: [
      { art: 'email', wert: 'ada@example.org' },
      { art: 'telefon', wert: '+49 170 0' },
    ],
    bild: null,
    bildart: '',
    ursprung: 'ada',
    uebernommen: '',
    ...teil,
  }
}

export const bo = kontakt({
  id: 3,
  fassung: 2,
  name: 'Bo',
  ursprung: 'bo',
  wege: [],
})

export function karten(): KartenvorschauDTO[] {
  return [
    { kontakt: kontakt(), abgleich: 'neu', vorhanden: null },
    {
      kontakt: kontakt({ name: 'Bo Neu', ursprung: 'bo', wege: [] }),
      abgleich: 'geaendert',
      vorhanden: bo,
    },
    {
      kontakt: kontakt({ name: 'Cy', ursprung: 'cy', wege: [] }),
      abgleich: 'gleich',
      vorhanden: kontakt({ id: 4, name: 'Cy' }),
    },
    {
      kontakt: kontakt({ name: 'Ada Lovelace', ursprung: '', wege: [] }),
      abgleich: 'vermutlich',
      vorhanden: kontakt({ id: 5 }),
    },
  ]
}

export function belege(teil: Parameters<typeof belegeApi>[0] = {}) {
  belegeApi({
    Kontakte: () => Promise.resolve({ kontakte: [bo], fehler: null }),
    KontakteEinlesen: () =>
      Promise.resolve({
        karten: karten(),
        befunde: ['Zeile 9 hat keinen Doppelpunkt — sie wird übergangen.'],
        abgebrochen: false,
        fehler: null,
      }),
    ...teil,
  })
}

export async function zeige() {
  render(
    <Sprachraum sprache="de">
      <Adressbuch />
    </Sprachraum>,
  )

  await userEvent.click(
    await screen.findByRole('button', { name: texte.titel }),
  )
}
