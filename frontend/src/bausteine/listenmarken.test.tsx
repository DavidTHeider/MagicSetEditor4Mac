// **Marken und Filter in der gemeinsamen Liste** — UX-12
// (#1684):
// Sets und Decks bekommen sie über `Listenansicht`.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Listenansicht } from './Listenansicht'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.bausteine

afterEach(() => {
  raeumeApi()
})

interface Eintrag {
  id: string
  name: string
}

it('zeigt die Marken am Eintrag und filtert über Go', async () => {
  const filtern = vi.fn().mockResolvedValue({ kennungen: ['a'], fehler: null })
  belegeApi({
    MarkenJeObjekt: vi
      .fn()
      .mockResolvedValue({ marken: { a: ['rot'] }, fehler: null }),
    MarkenFiltern: filtern,
    Einstellungen: vi
      .fn()
      .mockResolvedValue({ einstellungen: [], fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Listenansicht<Eintrag>
        bezeichnung="Decks"
        eintraege={[
          { id: 'a', name: 'Rot' },
          { id: 'b', name: 'Blau' },
        ]}
        schluessel={(eintrag) => eintrag.id}
        suchtext={(eintrag) => eintrag.name}
        sortierungen={[]}
        form="liste"
        setzeForm={vi.fn()}
        leer={null}
        marken={{ art: 'deck', name: (eintrag) => eintrag.name }}
        eintrag={(eintrag) => <span>{eintrag.name}</span>}
      />
    </Sprachraum>,
  )

  expect(await screen.findByText('rot')).toBeInTheDocument()

  await userEvent.click(
    screen.getByRole('button', {
      name: texte.objektmarken.werkzeugeUeberschrift,
    }),
  )
  await userEvent.click(
    screen.getByRole('button', { name: 'Nach »rot« filtern' }),
  )

  await waitFor(() => {
    expect(screen.queryByText('Blau')).toBeNull()
  })
  expect(filtern).toHaveBeenCalledWith('deck', ['rot'], 'und')

  // **Was die Suche leert, leert auch die Marken** — sonst bliebe die Liste
  // leer, ohne dass jemand sähe, warum.
  await userEvent.type(screen.getByLabelText(texte.listenansicht.suche), 'zzz')
  await userEvent.click(
    await screen.findByRole('button', {
      name: texte.listenansicht.sucheLeeren,
    }),
  )
  expect(await screen.findByText('Blau')).toBeInTheDocument()
  await antwortAbwarten()
})

// **In der Liste ist der Knopf leise** — UXR-24 (#1932): die Zeile trägt
// `group/eintrag`, und „Marken von …" erscheint erst unter dem Zeiger oder
// mit dem Fokus darin.
it('legt den Knopf der Marken in der Liste leise', async () => {
  belegeApi({
    MarkenJeObjekt: vi.fn().mockResolvedValue({ marken: {}, fehler: null }),
    Einstellungen: vi
      .fn()
      .mockResolvedValue({ einstellungen: [], fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Listenansicht<Eintrag>
        bezeichnung="Sets"
        eintraege={[{ id: 'a', name: 'Rot' }]}
        schluessel={(eintrag) => eintrag.id}
        suchtext={(eintrag) => eintrag.name}
        sortierungen={[]}
        form="liste"
        setzeForm={vi.fn()}
        leer={null}
        marken={{ art: 'set', name: (eintrag) => eintrag.name }}
        eintrag={(eintrag) => <span>{eintrag.name}</span>}
      />
    </Sprachraum>,
  )

  const knopf = await screen.findByRole('button', { name: 'Marken von Rot' })
  expect(knopf.closest('li')?.className).toContain('group/eintrag')
  expect(knopf.className).toContain('[@media(hover:hover)]:opacity-0')
  await antwortAbwarten()
})

// **Ohne Marken kein Filter nach Marken** — AP-6 (#2363): er filtert nichts,
// solange nichts eine trägt, und zählte im offenen Deck als einundzwanzigstes
// Bedienelement im Bild (R2). Dieselbe Regel wie für „Öffnen“ am offenen
// Deck: eine Handlung, die nichts tun kann, steht nicht da. Sobald ein
// Eintrag eine Marke trägt, ist er da (der erste Test oben).
it('zeigt den Filter nach Marken erst, wenn ein Eintrag eine trägt', async () => {
  const marken = vi.fn().mockResolvedValue({ marken: {}, fehler: null })
  belegeApi({
    MarkenJeObjekt: marken,
    Einstellungen: vi
      .fn()
      .mockResolvedValue({ einstellungen: [], fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Listenansicht<Eintrag>
        bezeichnung="Decks"
        eintraege={[{ id: 'a', name: 'Rot' }]}
        schluessel={(eintrag) => eintrag.id}
        suchtext={(eintrag) => eintrag.name}
        sortierungen={[]}
        form="liste"
        setzeForm={vi.fn()}
        leer={null}
        marken={{ art: 'deck', name: (eintrag) => eintrag.name }}
        eintrag={(eintrag) => <span>{eintrag.name}</span>}
      />
    </Sprachraum>,
  )

  await waitFor(() => {
    expect(marken).toHaveBeenCalled()
  })
  await antwortAbwarten()

  expect(
    screen.queryByRole('button', {
      name: texte.objektmarken.werkzeugeUeberschrift,
    }),
  ).toBeNull()
})
