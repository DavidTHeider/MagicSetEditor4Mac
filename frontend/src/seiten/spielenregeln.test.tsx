// **Was der Startassistent über die Regeln sagt** — Werkstücke G4 und G8 von
// Phase 5.5.
//
// **Neben `spielen.test.tsx` und nicht darin**: dort geht es um die Sitze und
// ihre Decks, hier um das, was **vor** dem Anfangen dasteht — Regelschau,
// Spielart, Format. Die Datei stand mit beidem bei 403 Zeilen (§2).
import { screen, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { antwortAbwarten, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { einDeck, eineAntwort, element } from './partieprobe'
import { einNutzer } from '../test/anmeldebelegung'
import { probeprofil } from '../test/profil'
import { mitDecks, zeige } from './spielenprobe'

const eigene = kataloge.de.seiten.spielen

afterEach(() => {
  raeumeApi()
})

// **Vor dem Anfangen steht, was durchgesetzt wird.**
//
// # Der Abnahmepunkt, für den es diesen Lauf gibt
//
// **„Spieler sehen vor Start, welche Regeln automatisch durchgesetzt
// werden"** (#1461). Danach wäre es zu spät — dann steht die Partie, und die
// Frage war, ob sie so stehen soll.
//
// **Und gefragt wird erst, wenn es ein Spiel gibt.** Solange kein Deck
// gewählt ist, gäbe eine Anfrage eine Absage auf eine Frage, die niemand
// gestellt hat — genau daran hing schon einmal ein Fehler im Inspector eines
// frisch angelegten Decks.
it('zeigt vor dem Anfangen, was durchgesetzt wird', async () => {
  const gefragt: unknown[][] = []

  mitDecks({
    Regelfreigabe: (
      spiel: string,
      strenge: string,
      sitze: number,
      modus: string,
      abgewaehlt: string[],
    ) => {
      gefragt.push([spiel, strenge, sitze, modus, abgewaehlt])

      return Promise.resolve({
        freigabe: {
          erlaubt: true,
          regelwerk: 'sandkasten',
          fassung: 1,
          wirksam: '2026-08-19',
          hoechste: 'begleitet',
          matrix: [
            {
              bereich: 'zonen',
              grad: 'automatisch',
              quelle: 'CR 400 (Fassung 2026-08-19)',
              grenzen: ['keine Zusatzzonen über die sieben hinaus'],
            },
          ],
          luecken: ['combat'],
          hindernisse: [],

          // **Modus und Baum kamen mit PLAY-02** (#1710) — ohne den Baum
          // zeigte die Schau nichts, denn sie zeichnet ihn und nicht die
          // Matrix.
          modus: 'anfaenger',
          baum: [{ kennung: 'ablauf', bereiche: ['zonen'] }],
          widersprueche: [],
        },
        fehler: null,
      })
    },
  })

  const hand = zeige()

  // **Die Sitze sind vorbelegt** (P6, #2374) — mit dem Deck steht das Spiel
  // fest, und die Auskunft kommt, ohne dass jemand wählt.

  // **Die Einzelheiten stehen eingeklappt in der Regeltafel** — seit dem
  // 2026-09-24, das Artboard *Eine Partie anfangen* (Befund F13).
  await hand.click(
    await screen.findByRole('button', { name: eigene.regelgrenzenTitel }),
  )

  expect(await screen.findByText(/CR 400/)).toBeInTheDocument()

  expect(gefragt).toEqual([['magic', 'sandkasten', 2, 'anfaenger', []]])

  // **Die Grenze steht dabei** — eine Zusage ohne Grenze behauptete
  // Vollständigkeit.
  expect(
    screen.getByText('keine Zusatzzonen über die sieben hinaus'),
  ).toBeInTheDocument()
})

// **Und was bei der Regelfreigabe klagt, steht als Warnung da.**
//
// **Nicht verschluckt**: wer nicht erfährt, was durchgesetzt wird, fängt eine
// Partie unter einer Annahme an. Das Issue verlangt die Auskunft **vor** dem
// Start — und ihr Ausbleiben ist selbst eine.
it('meldet, was bei der Regelfreigabe schiefging', async () => {
  mitDecks({
    Regelfreigabe: () => Promise.reject(new Error('Das Regelwerk fehlt.')),
  })

  const hand = zeige()

  const wahlen = await screen.findAllByRole('combobox')

  await hand.selectOptions(element(wahlen[0]), '1')

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Das Regelwerk fehlt.',
  )
})

// **Der Assistent sagt, was fehlt — und nicht nur, dass etwas fehlt.**
//
// #1460 verlangt „verständliche Sperren": ein Knopf, der grau ist und
// schweigt, lässt raten.
// **Seit P6 sind die Decks vorbelegt** (#2374); ohne Namen im Profil fehlt
// zuerst der Name am ersten Sitz.
it('nennt den nächsten offenen Schritt, solange nichts losgeht', async () => {
  mitDecks()

  zeige()

  expect(await screen.findByText(eigene.fehltName)).toBeInTheDocument()
  expect(screen.getByRole('button', { name: eigene.anfangen })).toBeDisabled()
})

// **Die Spielart folgt aus der Sitzzahl und steht da.**
//
// **CR 100.1** (Fassung 2026-08-19): die Regeln gelten für zwei oder mehr
// Spieler. Ein Goldfish hat einen — ein Aufbau und kein Spiel, und genau so
// heißt es.
it('nennt die Spielart, sobald ein Deck gewählt ist', async () => {
  mitDecks()

  const hand = zeige()

  const felder = await screen.findAllByLabelText(eigene.sitzDeck)
  await hand.selectOptions(element(felder[0]), '1')

  expect(
    await screen.findByText(new RegExp(eigene.spielartSelfplay)),
  ).toBeInTheDocument()

  // **Und bei einem Sitz heißt sie anders.**
  await hand.click(
    element(screen.getAllByRole('button', { name: eigene.sitzWeg })[1]),
  )

  expect(
    await screen.findByText(new RegExp(eigene.spielartGoldfish)),
  ).toBeInTheDocument()
})

// **Das Format kommt aus dem Deck** und steht neben dem Spiel — wie dieses
// wird es nicht gewählt: es gehört zum Deck, und ein Feld daneben ließe sich
// auf etwas anderes stellen.
it('nennt das Format, wo das Deck eines trägt', async () => {
  mitDecks({
    Decks: () =>
      Promise.resolve({
        decks: [
          { ...einDeck('1', 'Aggro'), formatID: 'commander' },
          einDeck('2', 'Kontrolle'),
        ],
        fehler: null,
      }),
  })

  const hand = zeige()

  const felder = await screen.findAllByLabelText(eigene.sitzDeck)
  await hand.selectOptions(element(felder[0]), '1')

  expect(
    await screen.findByText(new RegExp(`${eigene.format}: commander`)),
  ).toBeInTheDocument()
})

// **Ein Sitz lässt sich an ein Profil binden** — Werkstück G8 von Phase 5.5
// (#1465).
//
// # Wogegen hier geprüft wird
//
//   - Das Feld steht **erst da, wenn es zwei Profile gibt**. Eines wäre ein
//     Feld mit einer Wahl (§6).
//   - Die Vorgabe ist **kein Profil** — die Bindung ist freiwillig.
//   - Was gewählt wird, geht **so** hinaus.
//
// **Und die Verneinung wartet die Antwort ab** (§3): „da steht kein Feld"
// trifft auch dann zu, wenn die Profile noch unterwegs sind. Der Beleg, dass
// jemand zuhört, ist der Lauf darunter — mit **zwei** Profilen steht das Feld
// da, mit derselben Vorrichtung und demselben Weg.
it('bietet die Profile erst ab zweien an', async () => {
  const gefragt = vi.fn(() =>
    Promise.resolve({
      nutzer: [einNutzerMit([probeprofil({ id: '10' })])],
      fehler: null,
    }),
  )

  mitDecks({
    Nutzer: gefragt,
    WerArbeitet: () => Promise.resolve({ nutzerID: '1', fehler: null }),
  })

  zeige()

  await screen.findByLabelText(`${eigene.sitzName} 1`)

  await antwortAbwarten()

  expect(gefragt).toHaveBeenCalled()

  expect(screen.queryByLabelText(eigene.sitzprofil)).not.toBeInTheDocument()
})

it('schickt das gewählte Profil mit', async () => {
  const angelegt = vi.fn(() => Promise.resolve(eineAntwort()))

  mitDecks({
    PartieAnlegen: angelegt,
    Nutzer: () =>
      Promise.resolve({
        nutzer: [
          einNutzerMit([
            probeprofil({ id: '10', name: 'Anna' }),
            probeprofil({ id: '11', name: 'Bert' }),
          ]),
        ],
        fehler: null,
      }),
    WerArbeitet: () => Promise.resolve({ nutzerID: '1', fehler: null }),
  })

  const hand = zeige()

  await hand.type(await screen.findByLabelText(`${eigene.sitzName} 1`), 'Anna')
  await hand.type(screen.getByLabelText(`${eigene.sitzName} 2`), 'Bert')

  const decks = screen.getAllByLabelText(eigene.sitzDeck)
  await hand.selectOptions(element(decks[0]), '1')
  await hand.selectOptions(element(decks[1]), '2')

  // **Die Vorgabe ist kein Profil** — erst die Wahl bindet.
  const profile = await screen.findAllByLabelText(eigene.sitzprofil)

  expect(element(profile[0])).toHaveValue('')

  await hand.selectOptions(element(profile[0]), '10')
  await hand.selectOptions(element(profile[1]), '11')

  await hand.click(screen.getByRole('button', { name: eigene.anfangen }))

  await waitFor(() => {
    expect(angelegt).toHaveBeenCalledWith(
      expect.objectContaining({
        sitze: [
          expect.objectContaining({ kennung: 'sitz1', profilID: '10' }),
          expect.objectContaining({ kennung: 'sitz2', profilID: '11' }),
        ],
      }),
    )
  })
})

// einNutzerMit gibt einen Nutzer mit diesen Profilen.
//
// **Über [einNutzer] und nicht von Hand**: ein Verbund, dem Felder fehlen, ist
// eine Vorrichtung, die großzügiger ist als der Betrieb (§3) — und `mitDecks`
// nimmt `unknown` entgegen, sodass `tsc` sie nicht anhielte.
function einNutzerMit(profile: ReturnType<typeof probeprofil>[]) {
  return { ...einNutzer(), profile }
}
