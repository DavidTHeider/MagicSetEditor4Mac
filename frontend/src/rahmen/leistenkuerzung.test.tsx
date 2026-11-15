// **Was die Seitenleiste kürzt, bleibt lesbar** — H03
// (#2257).
//
// # Wie das gefunden wurde
//
// Am nativen Fenster im Mindestmaß 1024 × 700, über alle acht Sprachen
// (`scripts/huellenprobe/h03sprachen.mjs`): **„Nutzer und Profile" stand auf
// Französisch, Japanisch und Russisch mit Auslassungszeichen da** — und der
// volle Name war für das Auge nirgends zu haben. Der Vorleser hörte ihn, denn
// `truncate` kürzt nur die Zeichnung.
//
// **Umbrechen geht nicht**: die Zeile ist 32 Punkte hoch, weil die Leiste sonst
// im Mindestfenster rollt (`leistenmass.test.tsx`). Deshalb trägt jede
// gekürzte Beschriftung ihren vollen Text als Titel.
//
// **Was jsdom nicht kann, ist messen** — ob ein Text wirklich zu breit ist.
// Geprüft wird deshalb die Klasse: jede Beschriftung, die gekürzt werden
// **kann**, trägt ihn. Ob der Titel am nativen Fenster erscheint, belegt die
// Hüllenprobe.
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Kontextleiste } from './Kontextleiste'
import { Rahmen } from './Rahmen'
import { leisteneintraege } from '../leistenordnung'
import { kataloge, sprachen } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { einNutzer } from '../test/anmeldebelegung'
import { belegeApi, raeumeApi } from '../test/api'
import { leisteneintrag } from '../test/leistenweg'

import type { Sprache } from '../sprache/kataloge'

afterEach(() => {
  raeumeApi()
})

function gekuerzte(wurzel: HTMLElement) {
  return [...wurzel.querySelectorAll<HTMLElement>('.truncate')].filter(
    (e) => !e.classList.contains('sr-only'),
  )
}

describe('die senkrechte Leiste in jeder Sprache', () => {
  it.each(sprachen)(
    '%s: jede gekürzte Beschriftung trägt ihren vollen Text als Titel',
    async (sprache: Sprache) => {
      belegeApi({})
      render(
        <Sprachraum sprache={sprache}>
          <Rahmen
            aktuell="start"
            wechseln={vi.fn()}
            seitenleiste={{
              offen: true,
              umschalten: vi.fn(),
              waagerecht: false,
            }}
          >
            <p>{kataloge[sprache].seiten.start.was}</p>
          </Rahmen>
        </Sprachraum>,
      )
      for (const eintrag of leisteneintraege) {
        await leisteneintrag(eintrag.ziel, sprache)
      }

      const leiste = document.getElementById('seitenleiste')
      if (leiste === null) {
        throw new Error('die Seitenleiste fehlt')
      }
      const kandidaten = gekuerzte(leiste)

      // **Vor der Verneinung der Beleg, dass gezählt wird** (§3): mindestens
      // jeder Eintrag der Leiste und die Palette.
      expect(kandidaten.length).toBeGreaterThan(leisteneintraege.length)
      for (const e of kandidaten) {
        expect(e, e.outerHTML).toHaveAttribute('title', e.textContent.trim())
      }
    },
  )
})

describe('die Kontextleiste mit langen Namen', () => {
  it('gibt Nutzer- und Profilnamen ihren vollen Text als Titel', async () => {
    const vorlage = einNutzer()
    belegeApi({
      Arbeitskontext: vi.fn().mockResolvedValue({
        nutzerID: '1',
        profilID: '20',
        eignerID: '2',
        rolle: 'leser',
        nurLesend: true,
        modus: 'fenster',
        revision: 1,
        fehler: null,
      }),
      Nutzer: vi.fn().mockResolvedValue({
        nutzer: [
          {
            ...vorlage,
            id: '1',
            name: 'Annabella Maximiliane von Langenstein',
          },
          {
            ...vorlage,
            id: '2',
            name: 'Bartholomäus Friedrich Wilhelmsen',
            profile: [
              {
                ...vorlage.profile[0],
                id: '20',
                nutzerID: '2',
                name: 'Turniersammlung für den Winter in Hamburg',
              },
            ],
          },
        ],
        fehler: null,
      }),
      Nutzerbilder: vi
        .fn()
        .mockResolvedValue({ bilder: {}, badges: {}, fehler: null }),
    })
    render(
      <Sprachraum sprache="de">
        <Kontextleiste wechseln={vi.fn()} waagerecht={false} />
      </Sprachraum>,
    )
    // **Eine Zeile seit AP-11** (#2368): Name und Profil zusammen, der
    // Eigner des fremden Profils klein darunter — beide gekürzt, beide mit
    // ihrem Text als Titel.
    await screen.findByText(
      'Annabella Maximiliane von Langenstein · Turniersammlung für den Winter in Hamburg',
    )

    const leiste = screen.getByRole('region', { name: 'Arbeitskontext' })
    const kandidaten = gekuerzte(leiste)

    expect(kandidaten).toHaveLength(2)
    for (const e of kandidaten) {
      expect(e).toHaveAttribute('title', e.textContent.trim())
    }
  })
})
