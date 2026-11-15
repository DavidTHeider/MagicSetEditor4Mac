// **Die Vorrichtung der Einstellungsseite** — abgetrennt am 2026-09-05,
// als `einstellungen.test.tsx` mit dem Feld `wirksam` aus Werkstück 9 von
// Phase 4.12
// (#1351)
// über die Zeilengrenze aus §2 ging.
//
// **Der Schnitt folgt der Frage**: hier steht, **womit** geprüft wird;
// nebenan, **was** geprüft wird. Zwei Läufe teilen sich diese Einträge, und
// eine zweite Abschrift liefe beim nächsten Feld zurück.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'

import type { ComponentType, ReactNode } from 'react'

import { Einstellungen } from '../seiten/Einstellungen'
import { Sprachraum } from '../sprache/kontext'
import {
  stillePflege,
  stilleDatenpflege,
  stillerRuecksprung,
} from './bestandspflegeprobe'

import type { EinstellungDTO } from '../api'

// **Die Aktualisierung steht in dieser Liste nicht** — ihre Zeile kommt aus
// dem Bestand wie jede andere; wo ein Test sie meint, gibt er sie mit.
export const aktualisierung: EinstellungDTO = {
  schluessel: 'aktualisierung',
  wert: 'aus',
  wirksam: 'aus',
  werte: ['aus', 'melden', 'holen'],
  bereich: 'anwendung',
}

export const eintraege: EinstellungDTO[] = [
  {
    schluessel: 'sprache',
    wert: 'de',
    wirksam: 'de',
    werte: ['de', 'en'],
    bereich: 'anwendung',
  },
  {
    schluessel: 'erscheinungsbild',
    wert: 'system',
    wirksam: 'system',
    werte: ['system', 'hell', 'dunkel'],
    bereich: 'anwendung',
  },
  {
    schluessel: 'schriftgroesse',
    wert: 'normal',
    wirksam: 'normal',
    werte: ['klein', 'normal', 'gross', 'sehrgross'],
    bereich: 'anwendung',
  },
]

// **Alle vier**, denn der Fehler betraf jeden von ihnen gleich.
export const alleDienste: EinstellungDTO[] = (
  ['scryfall', 'mtgapi', 'mtgjson', 'scrydex'] as const
).map((name) => ({
  schluessel: `dienst.${name}`,
  wert: 'aus',
  wirksam: 'aus',
  werte: ['aus', 'an'],
  bereich: 'dienste',
}))

// zeige zeichnet die Seite und gibt den Spion auf `setze`.
export function zeigeEinstellungen(
  teile: Partial<Parameters<typeof Einstellungen>[0]> = {},

  // huelle umgibt die Seite — etwa mit einer Bühne für eine Übergabe aus
  // der Befehlspalette (UX-02.AK-02).
  huelle?: ComponentType<{ children: ReactNode }>,
) {
  const setze = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Einstellungen
        eintraege={eintraege}
        dienste={[]}
        fragen={[]}
        ki={[]}
        vorgaben={[]}
        anmeldung={[]}
        ursache={undefined}
        meldefehler={undefined}
        setze={setze}
        laeuft={() => false}
        bestandspflege={stillePflege()}
        ruecksprung={stillerRuecksprung()}
        pflege={stilleDatenpflege()}
        setID={undefined}
        {...teile}
      />
    </Sprachraum>,
    { wrapper: huelle },
  )

  return setze
}

// **Seit Werkstück I von Phase 4.10 gliedern Laschen die Seite**
// (#1325).
// Die Tests gehen damit denselben Weg wie ein Anwender.
export async function zumGebiet(name: string) {
  await userEvent.click(await screen.findByRole('tab', { name }))
}
