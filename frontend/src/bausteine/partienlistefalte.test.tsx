// **Die Partienliste mit echten Mengen** — UXR-13
// (#1921).
//
// # Wogegen hier geprüft wird
//
// **Gegen die Zeile, wie sie bis UXR-13 aussah**: je Sitz ein Knopf, dazu
// Duplizieren, Ausführen, Löschen und „+ Marke" — bei neunzehn Partien
// (`test/spielbestand.ts`) 122 Bedienelemente neben dem Anfang einer neuen
// Partie, gezählt nach den Wählern der Hüllenprobe `uirevue` (R2 und R6 in
// `docs/UIREVIEW.md`). Das Artboard *Eine Partie anfangen* zeigt je Partie
// eine Zeile.
//
//   - **Je Zeile zwei Bedienelemente**, gleich wie viele Sitze: das Menü
//     „Weiterspielen" mit den Sitzen und dem Zuschauen, das Menü der Partie
//     mit Duplizieren, Ausführen, Löschen und Marken.
//   - **Vorn vier, der Rest stückweise** — mit der Zahl daneben, damit eine
//     gekappte Liste nicht wie eine vollständige aussieht.
//   - **Der Markenfilter erst, wenn eine Partie eine Marke trägt** — vorher
//     kann er nur wegnehmen, wie die Suchleiste über einer einzigen Partie.
import {
  cleanup,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import { Partienliste } from './Partienliste'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { element } from '../seiten/partieprobe'
import { partienbestand } from '../test/spielbestand'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { stilleVerwaltung } from '../test/verwaltung'

import type { Partiezeile } from '../apipartie'

const eigene = kataloge.de.seiten.spielen
const liste = kataloge.de.bausteine.listenansicht

// zeige zeichnet die Liste und gibt die Hand.
function zeige(partien: Partiezeile[]) {
  render(
    <Sprachraum sprache="de">
      <Partienliste
        geladen
        wiederholen={() => undefined}
        partien={partien}
        fehler={undefined}
        einsteigen={() => undefined}
        verwaltung={stilleVerwaltung()}
      />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}

// zeilen sind die Partien, die dastehen — je eine mit ihrem Menü
// „Weiterspielen".
function zeilen(): HTMLElement[] {
  return screen.getAllByRole('listitem').filter(
    (zeile) =>
      within(zeile).queryByRole('button', {
        name: /^Weiterspielen: /,
      }) !== null,
  )
}

it('zeigt vorn vier Partien und den Rest stückweise', async () => {
  const hand = zeige(partienbestand)

  expect(zeilen()).toHaveLength(4)
  expect(
    screen.getByText(einsetzen(liste.zahl, { gezeigt: '4', gesamt: '19' })),
  ).toBeInTheDocument()

  await hand.click(
    screen.getByRole('button', {
      name: einsetzen(liste.weitere, { anzahl: '10' }),
    }),
  )
  expect(zeilen()).toHaveLength(14)

  await hand.click(
    screen.getByRole('button', {
      name: einsetzen(liste.weitere, { anzahl: '5' }),
    }),
  )
  expect(zeilen()).toHaveLength(19)
})

it('bietet den Markenfilter erst an, wenn eine Partie eine Marke trägt', async () => {
  const filter = kataloge.de.bausteine.objektmarken.werkzeugeUeberschrift

  // **Erst der Beleg, dass die Liste zuhört** (§3): mit einer Marke steht
  // der Filter da.
  belegeApi({
    MarkenJeObjekt: () =>
      Promise.resolve({ marken: { p3: ['Turnier'] }, fehler: null }),
  })
  zeige(partienbestand)

  expect(
    await screen.findByRole('button', { name: filter }),
  ).toBeInTheDocument()

  cleanup()

  const gefragt = vi.fn(() => Promise.resolve({ marken: {}, fehler: null }))
  belegeApi({ MarkenJeObjekt: gefragt })
  zeige(partienbestand)

  await waitFor(() => {
    expect(gefragt).toHaveBeenCalled()
  })
  await antwortAbwarten()

  expect(screen.queryByRole('button', { name: filter })).not.toBeInTheDocument()

  raeumeApi()
})

// **Und er bleibt, solange eine Marke gewählt ist** — sonst ließe sich ein
// Filter auf eine Marke, die inzwischen fort ist, nicht mehr zurücknehmen,
// und die Liste bliebe leer.
it('lässt den Markenfilter stehen, solange eine Marke gewählt ist', async () => {
  const marken = kataloge.de.bausteine.objektmarken
  let vergeben: Record<string, string[]> = { p1: ['Turnier'] }

  belegeApi({
    MarkenJeObjekt: () => Promise.resolve({ marken: vergeben, fehler: null }),
    MarkenFiltern: () => Promise.resolve({ kennungen: ['p1'], fehler: null }),
    ObjektmarkeEntfernen: () => {
      vergeben = {}

      return Promise.resolve({ fehler: null })
    },
  })
  const hand = zeige(partienbestand.slice(0, 1))

  await hand.click(
    await screen.findByRole('button', { name: marken.werkzeugeUeberschrift }),
  )
  await hand.click(
    await screen.findByRole('button', {
      name: einsetzen(marken.filterWaehlen, { marke: 'Turnier' }),
    }),
  )

  await hand.click(
    screen.getByRole('button', {
      name: einsetzen(eigene.partienHandlungen, { name: 'Ada · Bo' }),
    }),
  )
  await hand.click(screen.getByRole('menuitem', { name: marken.ueberschrift }))
  await hand.click(
    screen.getByRole('button', {
      name: einsetzen(marken.entfernen, { marke: 'Turnier' }),
    }),
  )

  // **Die Marke ist fort** — und der Filter steht noch da.
  await waitFor(() => {
    expect(screen.queryByText(marken.keine)).toBeInTheDocument()
  })
  expect(
    screen.getByRole('button', { name: marken.werkzeugeUeberschrift }),
  ).toBeInTheDocument()

  raeumeApi()
})

it('trägt je Zeile zwei Bedienelemente, gleich wie viele Sitze', () => {
  zeige(partienbestand)

  for (const zeile of zeilen()) {
    expect(within(zeile).getAllByRole('button')).toHaveLength(2)
  }

  // **Auch die Partie zu viert** — vorher trug sie sieben.
  const viere = zeilen().find((zeile) =>
    zeile.textContent.includes('Ada · Cleo · Dario · Bo'),
  )
  expect(within(element(viere)).getAllByRole('button')).toHaveLength(2)
})

it('bietet im Menü „Weiterspielen" jeden Sitz und das Zuschauen an', async () => {
  const hand = zeige(partienbestand.slice(1, 2))

  await hand.click(
    screen.getByRole('button', {
      name: einsetzen(eigene.partienWeiterspielenIn, {
        name: 'Ada · Cleo · Dario · Bo',
      }),
    }),
  )

  expect(
    screen.getAllByRole('menuitem').map((eintrag) => eintrag.textContent),
  ).toEqual([
    ...['Ada', 'Cleo', 'Dario', 'Bo'].map((name) =>
      einsetzen(eigene.partienEinsteigenAls, { name }),
    ),
    eigene.partienZuschauen,
  ])
})

it('öffnet die Marken einer Partie über ihr Menü', async () => {
  const hand = zeige(partienbestand.slice(0, 1))

  expect(
    screen.queryByLabelText(kataloge.de.bausteine.objektmarken.setzenFuer),
  ).not.toBeInTheDocument()

  await hand.click(
    screen.getByRole('button', {
      name: einsetzen(eigene.partienHandlungen, { name: 'Ada · Bo' }),
    }),
  )
  await hand.click(
    screen.getByRole('menuitem', {
      name: kataloge.de.bausteine.objektmarken.ueberschrift,
    }),
  )

  expect(
    screen.getByLabelText(kataloge.de.bausteine.objektmarken.setzenFuer),
  ).toBeInTheDocument()
})
