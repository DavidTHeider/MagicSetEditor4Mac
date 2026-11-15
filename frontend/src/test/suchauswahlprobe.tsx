// **Die Probe der Auswahl mit Suche** — UXR-F1
// (#2007).
//
// Geteilt zwischen den zwei Dateien, die den Baustein prüfen, damit keine von
// ihnen die 400 Zeilen aus §2 reißt und beide dasselbe Feld meinen.
import { screen } from '@testing-library/react'
import { useState } from 'react'
import { vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Suchauswahl } from '../bausteine/Suchauswahl'
import type { Suchwahl } from '../bausteine/suchwahl'

export const zonen: Suchwahl[] = [
  { wert: '', name: 'Wie das System' },
  { wert: 'Africa/Cairo', name: 'Africa/Cairo' },
  { wert: 'America/New_York', name: 'America/New_York' },
  { wert: 'Europe/Berlin', name: 'Europe/Berlin' },
  { wert: 'Europe/Paris', name: 'Europe/Paris' },
]

export const stile: Suchwahl[] = [
  { wert: 'magic-m15', name: 'M15 Normal', zusatz: 'magic-m15' },
  {
    wert: 'magic-m15-altered',
    name: 'Altered Art',
    zusatz: 'magic-m15-altered',
  },
  {
    wert: 'probe-schlicht',
    name: 'Schlicht',
    zusatz: 'probe-schlicht · probe',
  },
]

interface ProbeEigenschaften {
  eintraege?: Suchwahl[]
  anfang?: string
  aendern?: (wert: string) => void
  platzhalter?: string
  angehalten?: boolean
}

// Probe hält den Wert wie ein Aufrufer — kontrolliert — und gibt dem Feld die
// Beschriftung, die der Aufrufer ihm geben muss.
export function Probe({
  eintraege = zonen,
  anfang = 'Europe/Berlin',
  aendern = vi.fn(),
  platzhalter,
  angehalten = false,
}: ProbeEigenschaften) {
  const [wert, setzeWert] = useState(anfang)

  return (
    <Sprachraum sprache="de">
      <label htmlFor="probe-auswahl">Zeitzone</label>
      <Suchauswahl
        id="probe-auswahl"
        bezeichnung="Zeitzone"
        eintraege={eintraege}
        wert={wert}
        aendern={(neu) => {
          aendern(neu)
          setzeWert(neu)
        }}
        {...(platzhalter === undefined ? {} : { platzhalter })}
        angehalten={angehalten}
      />
    </Sprachraum>
  )
}

export const feld = () => screen.getByRole('combobox', { name: 'Zeitzone' })

// angeboten nennt die Zeilen, die die geöffnete Liste gerade führt.
export const angeboten = () =>
  screen.queryAllByRole('option').map((zeile) => zeile.textContent)

// ansage ist die Live-Region der Auswahl — was Vorleseprogramme ansagen.
export const ansage = () => document.querySelector('[aria-live="polite"]')
