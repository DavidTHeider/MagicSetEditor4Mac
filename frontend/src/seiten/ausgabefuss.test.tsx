// **Der Knopf, der schreibt, steht in einer Fußleiste, die stehen bleibt** —
// AP-10 (#2367),
// Befund B-20: bei 1.440 × 838 lag er unter der Falz.
//
// **Ob er stehen bleibt, sieht jsdom nicht** — es rechnet keine Lage. Hier
// steht, dass jedes Ziel seinen Hauptknopf in der Leiste trägt und der
// Druckbogen daneben seine Zahl; dass sie bei 1.440 × 838 und bei der
// Mindestgröße sichtbar ist, zeigt der Lauf am nativen Fenster.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { stillerBogenprobelauf } from '../test/apiprobelauf'
import { Ausgeben } from './Ausgeben'
import { SetVorlagenausfuhr } from './SetVorlagenausfuhr'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten

const satz = {
  id: '4',
  name: 'Mein Probeset',
  spiel: 'magic',
  stil: 'm15',
  herkunft: '/pfad/probeset.mse-set',
  karten: 10,
  herkunftFehlt: false,
}

function belege(belegung: Record<string, unknown> = {}) {
  belegeApi({
    // **Der Assistent merkt sich das Ziel** (AP-10, #2367).
    EinstellungSetzen: () => Promise.resolve({}),
    Sets: () => Promise.resolve({ sets: [satz] }),
    Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
    BogenProbelauf: () => Promise.resolve(stillerBogenprobelauf),
    Exportvorlagen: () =>
      Promise.resolve({
        vorlagen: [
          {
            paket: 'magic-spoiler.mse-export-template',
            name: 'Spoiler',
            dateityp: '*.html|*.html',
            optionen: [],
          },
        ],
        befunde: [],
      }),
    ...belegung,
  })
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Ausgeben setID="4" seite="ausgeben" wechseln={vi.fn()} />
    </Sprachraum>,
  )
}

async function waehle(ziel: string) {
  await userEvent.click(await screen.findByRole('radio', { name: ziel }))
}

// fussVon gibt die Fußleiste, in der ein Knopf steht.
function fussVon(knopf: HTMLElement) {
  return knopf.closest('[data-ausgabefuss]')
}

describe('Die Fußleiste der Ausgabe', () => {
  it.each([
    ['den Bildern', texte.bilder.titel, texte.bilder.ausgeben],
    ['dem Druckbogen', texte.druckbogen.titel, texte.druckbogen.legen],
    [
      'MTG JSON',
      texte.sets.mtgjsonUeberschrift,
      `${texte.sets.mtgjsonSichern} — ${satz.name}`,
    ],
  ])('trägt bei %s den Hauptknopf', async (_, ziel, knopf) => {
    belege()
    zeige()

    await waehle(ziel)

    const flaeche = await screen.findByRole('region', { name: ziel })
    const fuss = fussVon(
      await within(flaeche).findByRole('button', { name: knopf }),
    )

    expect(fuss).toHaveAttribute('data-ausgabefuss', 'fest')
  })

  it('trägt bei der Exportvorlage den Hauptknopf', async () => {
    belege()
    zeige()

    await waehle(texte.sets.vorlageWaehlen)
    expect(
      fussVon(
        await screen.findByRole('button', {
          name: texte.sets.vorlageAusfuehren,
        }),
      ),
    ).toHaveAttribute('data-ausgabefuss', 'fest')
  })

  // **Neben dem Knopf des Druckbogens steht, was er legt** — und nur, solange
  // die Zahl zur Wahl gehört.
  it('nennt neben dem Druckbogen, was er legt, und schweigt beim Rechnen', async () => {
    const BogenProbelauf = vi
      .fn()
      .mockResolvedValueOnce(stillerBogenprobelauf)
      .mockReturnValue(new Promise(() => undefined))

    belege({ BogenProbelauf })
    zeige()

    await waehle(texte.druckbogen.titel)

    const flaeche = await screen.findByRole('region', {
      name: texte.druckbogen.titel,
    })
    const fuss = fussVon(
      await within(flaeche).findByRole('button', {
        name: texte.druckbogen.legen,
      }),
    )

    expect(
      await within(fuss as HTMLElement).findByText('9 je Seite · 2 Seiten'),
    ).toBeVisible()

    await userEvent.selectOptions(
      within(flaeche).getByRole('combobox', { name: texte.druckbogen.papier }),
      'letter',
    )

    expect(
      await within(flaeche).findByText(texte.ausgeben.probelauf.rechnet),
    ).toBeVisible()
    expect(
      within(fuss as HTMLElement).queryByText('9 je Seite · 2 Seiten'),
    ).not.toBeInTheDocument()
  })

  // **Außerhalb des Assistenten steht die Exportvorlage an einer Setkarte**,
  // in einem Kasten — dort bliebe eine Leiste über der ganzen Seite stehen.
  it('steht an der Setkarte nicht fest', async () => {
    belege()

    render(
      <Sprachraum sprache="de">
        <SetVorlagenausfuhr satz={satz} schliessen={vi.fn()} />
      </Sprachraum>,
    )

    expect(
      fussVon(
        await screen.findByRole('button', {
          name: texte.sets.vorlageAusfuehren,
        }),
      ),
    ).toHaveAttribute('data-ausgabefuss', 'frei')
  })
})
