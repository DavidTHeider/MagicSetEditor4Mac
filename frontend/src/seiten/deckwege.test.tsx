// **Die Wege, die vom Deck wegführen** — UX-03.AK-01
// (#1675).
//
// # Was der Entwurf zeichnet
//
// Im Artboard *Decks* stehen drei Wege heraus: unten im Navigator „Karten
// holen aus" mit *Meiner Sammlung* und *Einem Set*, unten im Inspector der
// Hauptknopf *Mit diesem Deck spielen*
// (`docs/prototyp/project/Decks.dc.html`).
//
// **Eine eigene Datei**, damit `deckbauplatz.test.tsx` unter der
// Zeilengrenze aus §2 bleibt. Der Schnitt folgt der Frage: dort steht, dass
// man an einem Deck **arbeitet**, hier, wohin man es **verlässt**.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Anschlussraum } from '../anschlussraum'
import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { bauartprobe, deckprobe, formatprobe } from './deckprobe'
import { kataloge } from '../sprache/kataloge'

import type { Seitenschluessel } from '../navigation'

const texte = kataloge.de.seiten.decks

// Buehne stellt die Seite in den Raum — **mit einem `wechseln` für beide**,
// so wie `Anwendung` es tut: der Raum und die Seite schicken denselben Weg.
function Buehne({ gewechselt }: { gewechselt: (seite: string) => void }) {
  const [aktuell, setzeAktuell] = useState<Seitenschluessel>('decks')

  const wechseln = (seite: Seitenschluessel) => {
    gewechselt(seite)
    setzeAktuell(seite)
  }

  return (
    <Sprachraum sprache="de">
      <Anschlussraum aktuell={aktuell} wechseln={wechseln}>
        <p>{aktuell}</p>
        <Decks wechseln={wechseln} />
      </Anschlussraum>
    </Sprachraum>
  )
}

function zeige() {
  belegeApi({
    Decks: () => Promise.resolve({ decks: [deckprobe()], fehler: null }),
    Formate: () => Promise.resolve({ formate: [formatprobe()], fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
    Deck: () => Promise.resolve({ deck: deckprobe(), fehler: null }),
  })

  const gewechselt = vi.fn()

  render(<Buehne gewechselt={gewechselt} />)

  return gewechselt
}

// aufschlagen öffnet das Deck über den Navigator.
async function aufschlagen() {
  await userEvent.click(
    await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
  )
}

afterEach(() => {
  raeumeApi()
})

// wegeOeffnen klappt den Kasten der seltenen Wege auf (AP-6, #2363).
async function wegeOeffnen() {
  await userEvent.click(
    await screen.findByRole('button', { name: texte.wegeTitel }),
  )
}

// **Die seltenen Wege stehen in einem Kasten** — AP-6 (#2363), Entwurf
// `Bereiche09`: „Sekundäres im …“. Übernahme, Umschlag und „Karten holen
// aus“ standen einzeln am Fuß des Navigators und zählten vier zu den
// Bedienelementen im Bild (R2); das offene Deck lag mit 27 darüber.
describe('die seltenen Wege', () => {
  it('stehen zugeklappt in einem Kasten und gehen mit ihm auf', async () => {
    zeige()

    const kasten = await screen.findByRole('button', {
      name: texte.wegeTitel,
    })
    expect(kasten).toHaveAttribute('aria-expanded', 'false')
    for (const name of [
      texte.uebernahme,
      texte.umschlagTitel,
      texte.holenSammlung,
      texte.holenSet,
    ]) {
      expect(screen.queryByRole('button', { name })).toBeNull()
    }

    await userEvent.click(kasten)

    for (const name of [
      texte.uebernahme,
      texte.umschlagTitel,
      texte.holenSammlung,
      texte.holenSet,
    ]) {
      expect(screen.getByRole('button', { name })).toBeInTheDocument()
    }
  })
})

describe('Karten holen aus', () => {
  it('führt in die Sammlung', async () => {
    const gewechselt = zeige()

    await wegeOeffnen()
    await userEvent.click(
      await screen.findByRole('button', { name: texte.holenSammlung }),
    )

    expect(gewechselt).toHaveBeenCalledWith('sammlung')
  })

  it('führt zu den Sets', async () => {
    const gewechselt = zeige()

    await wegeOeffnen()
    await userEvent.click(
      await screen.findByRole('button', { name: texte.holenSet }),
    )

    expect(gewechselt).toHaveBeenCalledWith('sets')
  })
})

// **Am offenen Deck steht kein „Öffnen“** — AP-6 (#2363): es ist offen,
// und ein Knopf, der nichts mehr tut, zählte zu den Bedienelementen im Bild
// (R2). Die Marke sagt, welches es ist; der Rand der Marke stand schon da.
describe('das offene Deck in der Liste', () => {
  it('trägt die Marke „offen“ statt „Öffnen“', async () => {
    zeige()

    await aufschlagen()

    const liste = within(screen.getByRole('region', { name: texte.titel }))
    expect(await liste.findByText(texte.offenMarke)).toBeInTheDocument()
    expect(
      liste.queryByRole('button', { name: `${texte.waehlen}: Aggro` }),
    ).toBeNull()
  })
})

describe('Mit diesem Deck spielen', () => {
  it('nimmt das offene Deck mit auf die Spielenseite', async () => {
    const gewechselt = zeige()

    await aufschlagen()
    await userEvent.click(
      await screen.findByRole('button', { name: texte.mitspielen }),
    )

    expect(gewechselt).toHaveBeenCalledWith('spielen')
  })

  // **Ohne offenes Deck gibt es nichts mitzunehmen.** Der Inspector sagt
  // dann, dass keines offensteht — ein Knopf daneben führte ins Leere.
  it('steht ohne offenes Deck nicht da', async () => {
    zeige()

    await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` })

    expect(
      screen.queryByRole('button', { name: texte.mitspielen }),
    ).not.toBeInTheDocument()
  })
})
