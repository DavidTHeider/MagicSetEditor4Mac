// **Jede Gruppe der Einstellungen hält WCAG 2.2 AA — auch aufgeklappt** —
// UXR-19 (#1927).
//
// # Warum ein eigener Lauf
//
// **`barrierefreiheit.test.tsx` sieht die Seite, wie sie aufgeht** — die
// erste Gruppe, nichts aufgeklappt. Seit UXR-19 steht in jeder Gruppe ein
// Teil hinter einem Aufklapper, und *Bereiche und Funktionen* trägt Schalter
// statt Auswahllisten. Was dort gezeichnet wird, sah kein axe-Lauf.
//
// # Was er nicht hält
//
// **Farbkontraste** — axe meldet sie unter jsdom als unentschieden (§3);
// sie stehen gemessen in `docs/KONTRASTE.md`.
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { zeigeEinstellungen } from '../test/einstellungsseite'
import { profilEintraege } from '../test/einstellungskatalog'
import { einstellungsgruppen } from './einstellungsgruppen'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { verstoesse } from '../test/axe'

afterEach(() => {
  raeumeApi()
})

// alle sind jede Einstellung des Katalogs, mit ihren Werten — seit UXR-F2
// (#2008) zeichnet die Seite einen Schalter, Segmente oder eine Auswahlliste,
// je nachdem, und axe soll alle drei sehen.
const alle = profilEintraege()

describe('die Gruppen der Einstellungen, aufgeklappt', () => {
  it.each(einstellungsgruppen)('%s hält WCAG 2.2 AA', async (gruppe) => {
    belegeApi({})
    zeigeEinstellungen({ alleEinstellungen: alle })

    await userEvent.click(
      screen.getByRole('tab', {
        name: kataloge.de.seiten.einstellungen.gruppen[gruppe],
      }),
    )

    for (const knopf of within(screen.getByRole('tabpanel')).queryAllByRole(
      'button',
      { expanded: false },
    )) {
      await userEvent.click(knopf)
    }

    expect(await verstoesse(document.body)).toEqual([])
  })
})
