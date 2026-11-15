// **Die Vorrichtung der Kontaktläufe** — Werkstück N1 von Phase 5.4
// (#1454).
//
// **Eine eigene Datei**, weil zwei Läufe sie brauchen und beide sonst über die
// Zeilengrenze aus §2 gingen.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { Peerkontakte } from './Peerkontakte'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

import type { Peerkontakt, Peerkontakteantwort } from '../apipeerkontakt'

const eigene = kataloge.de.seiten.technik

// dieZwecke sind die vier aus `domain`.
export const dieZwecke = ['paarung', 'teilen', 'handel', 'partie']

// einKontakt baut ein Gegenüber.
export function einKontakt(anders: Partial<Peerkontakt> = {}): Peerkontakt {
  return {
    kennung: 'bgegenueber',
    fingerabdruck: '1111-2222-3333-4444-5555-6666-7777-8888',
    selbstgewaehlterName: 'Fremder Rechner',
    eigenerName: '',
    stand: 'unbestaetigt',
    gesehen: '2026-09-14T03:00:00Z',
    rechte: [],
    adresse: '',
    erstmals: '2026-09-14T03:00:00Z',
    notiz: '',
    lage:
      anders.stand === 'gesperrt'
        ? 'widerrufen'
        : (anders.stand ?? 'unbestaetigt'),
    vergleichscode: '',
    ...anders,
  }
}

// eineListe baut eine Antwort.
export function eineListe(
  kontakte: Peerkontakt[] = [einKontakt()],
  fehler: Peerkontakteantwort['fehler'] = null,
): Peerkontakteantwort {
  return { kontakte, zwecke: dieZwecke, fehler }
}

// zeige stellt den Kasten auf und schlägt ihn auf — und, wo einer genannt
// ist, den Aufklapper in der offenen Zeile.
//
// **Seit UXR-16** (#1924)
// stehen Adresse und Notiz und die erteilten Sichten hinter je einem
// Aufklapper; wer sie prüft, öffnet ihn wie ein Mensch.
export async function zeige(
  aufklapper?: string,
): Promise<ReturnType<typeof userEvent.setup>> {
  render(
    <Sprachraum sprache="de">
      <Peerkontakte />
    </Sprachraum>,
  )

  const hand = userEvent.setup()

  await hand.click(screen.getByRole('button', { name: eigene.kontakteZeigen }))

  if (aufklapper !== undefined) {
    await hand.click(await screen.findByRole('button', { name: aufklapper }))
  }

  return hand
}
