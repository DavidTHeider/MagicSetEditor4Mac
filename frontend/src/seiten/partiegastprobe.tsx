// **Vorrichtungen für den Gasttisch** — Werkstück G5 von Phase 5.5
// (#1462).
//
// **Eine eigene Datei**, weil zwei Läufe sie brauchen — der übers Hinsetzen
// und der übers Spielen. Zwei Kopien liefen beim ersten neuen Feld
// auseinander.
import { render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'

import { Gastbeitritt, Gastplatz, sitztAmFremdenTisch } from './Partiegast'
import { Sprachraum } from '../sprache/kontext'
import { usePartiegast } from '../haken/usePartiegast'
import { belegeApi } from '../test/api'
import { einKontakt, eineListe } from './peerkontaktprobe'

import type { Api } from '../api'

// derWirt ist ein Gegenüber, bei dem mitgespielt werden darf — bestätigt,
// und „Eine Partie spielen" ist ihm freigegeben (UX-08.AK-04, #1680).
export const derWirt = einKontakt({
  kennung: 'bwirt',
  eigenerName: 'Wirt',
  stand: 'bestaetigt',
  rechte: ['partie'],
})

// gastApi belegt die Fassade samt der Kontakte, unter denen der Wirt steht.
//
// **Die Wahl lädt sie bei jedem Aufstellen** — ohne Beleg stünde neben dem,
// worum es in einem Lauf geht, ein Ladefehler.
export function gastApi(teile: Partial<Api>): Api {
  return belegeApi({
    Peerkontakte: () => Promise.resolve(eineListe([derWirt])),
    ...teile,
  })
}

// waehleWirt wählt ihn, statt eine Kennung zu tippen.
export async function waehleWirt(
  hand: ReturnType<typeof nutzerereignis.setup>,
): Promise<void> {
  await hand.click(await screen.findByRole('radio', { name: /^Wirt · / }))
}

// laufzeitbelegen setzt eine Wails-Laufzeit, die Rückrufe festhält.
//
// **Ohne sie gibt es keinen Meldeweg** — und das ist im Betrieb unter `vite
// preview` genauso. Der Lauf darunter prüft, dass dann nichts bricht.
export function laufzeitbelegen(): Map<string, (...daten: unknown[]) => void> {
  const horcher = new Map<string, (...daten: unknown[]) => void>()

  ;(globalThis as { runtime?: unknown }).runtime = {
    EventsOn: (name: string, rueckruf: (...daten: unknown[]) => void) => {
      horcher.set(name, rueckruf)

      return () => {
        horcher.delete(name)
      }
    },
  }

  return horcher
}

// Gastprobe stellt Platz und Beitritt nebeneinander, wie die Seite es tut —
// **seit UXR-13 (#1921) an zwei Orten**: der Platz auf der Seite, der
// Beitritt in einer Schublade, die zugeht, sobald jemand sitzt. Die
// Schublade selbst prüft `spielenschubladen.test.tsx`; hier geht es um den
// Weg an den fremden Tisch.
function Gastprobe() {
  const lauf = usePartiegast()

  return (
    <>
      <Gastplatz lauf={lauf} />
      {!sitztAmFremdenTisch(lauf) && <Gastbeitritt lauf={lauf} />}
    </>
  )
}

// zeige zeichnet die Gastfläche und gibt die Hand.
export function zeigeGast() {
  render(
    <Sprachraum sprache="de">
      <Gastprobe />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}
