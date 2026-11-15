// **Zwei Spalten nach Seite, mit Preisen und Differenz** — TRADE-02b
// (#1951).
//
// # Wogegen hier geprüft wird
//
// **AK-01**: eine Zeile steht in ihrer Spalte und wechselt die Seite nicht
// über eine Auswahl; **AK-02**: je Zeile ein Preis oder „kein Preis", je
// Seite eine Summe, die Differenz mit Richtung und Herkunft der Preise.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Anschlussraum } from '../anschlussraum'
import { Handelspositionen } from './Handelspositionen'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'

import type { Handelsposition, Preisanhalt } from '../apihandelsleitung'
import { kennungsfeld } from '../seiten/tauschprobe'
import type { Seitenschluessel } from '../navigation'

const texte = kataloge.de.seiten.handeln
const eigene = texte.assistent

afterEach(() => {
  raeumeApi()
})

function zeile(von: string, printing: string): Handelsposition {
  return {
    von,
    printing,
    druckId: 0,
    menge: 1,
    guete: 'nearMint',
    sprache: 'de',
    finish: 'nonfoil',
    abbild: '',
    bewertung: 0,
  }
}

const erhoben = '2026-09-20T08:00:00Z'

// einAnhalt: ich gebe 19,90 €, das Gegenüber 17,95 € und eine Karte ohne
// Preis.
function einAnhalt(): Preisanhalt {
  const eur = (wert: number) => ({
    waehrung: 'EUR',
    wert,
    quelle: 'scryfall',
    erhoben,
  })

  // **Jede Zeile nennt ihre Karte** (P10, #2378) — die Kennung steht nicht
  // mehr im Hauptweg.
  const druck = (name: string) => ({
    name,
    setcode: 'prb',
    nummer: '1',
    bild: '',
    eigen: false,
  })

  return {
    zeilen: [
      { preise: [eur(1990)], druck: druck('Der Eine Ring') },
      { preise: [eur(1795)], druck: druck('Tarmogoyf') },
      { preise: [], druck: druck('Leere') },
    ],
    seiten: [
      { von: 'ich-1', summen: [eur(1990)], ohnePreis: 0 },
      // **Die Summe des Gegenübers ist älter** — der Satz darunter nennt
      // den ältesten Preis.
      {
        von: 'du-1',
        summen: [{ ...eur(1795), erhoben: '2026-09-18T08:00:00Z' }],
        ohnePreis: 1,
      },
    ],
    differenzen: [{ waehrung: 'EUR', wert: -195, vollstaendig: false }],
    hauptwaehrung: 'EUR',
  }
}

function Buehne({
  anfang,
  bearbeitbar,
  gesetzt,
  ich = 'ich-1',
}: {
  anfang: Handelsposition[]
  bearbeitbar: boolean
  gesetzt?: (positionen: Handelsposition[]) => void
  ich?: string
}) {
  const [aktuell, setzeAktuell] = useState<Seitenschluessel>('handeln')
  const [positionen, setzePositionen] = useState(anfang)

  return (
    <Sprachraum sprache="de">
      <Anschlussraum aktuell={aktuell} wechseln={setzeAktuell}>
        <Handelspositionen
          ich={ich}
          gegenueber="du-1"
          positionen={positionen}
          bearbeitbar={bearbeitbar}
          setzePositionen={(neu) => {
            gesetzt?.(neu)
            setzePositionen(neu)
          }}
        />
      </Anschlussraum>
    </Sprachraum>
  )
}

describe('die zwei Spalten eines Tauschs', () => {
  it('stellt jede Zeile in die Spalte ihrer Seite, mit Preis und Summe', async () => {
    const gefragt = vi.fn(() =>
      Promise.resolve({ anhalt: einAnhalt(), fehler: null }),
    )
    belegeApi({ HandelPreisanhalt: gefragt })

    render(
      <Buehne
        bearbeitbar={false}
        anfang={[
          zeile('ich-1', 'ring'),
          zeile('du-1', 'goyf'),
          zeile('du-1', 'leere'),
        ]}
      />,
    )

    const meine = screen.getByRole('region', { name: texte.positionVonIch })
    const seine = screen.getByRole('region', {
      name: texte.positionVonGegenueber,
    })

    expect(within(meine).getAllByRole('listitem')).toHaveLength(1)
    // **Güte und Ausführung als Wörter**, nicht als Schlüssel.
    const sammlung = kataloge.de.seiten.sammlung
    expect(
      within(meine).getByText(
        `${sammlung.zustaende.nearMint} · de · ${sammlung.nonfoil}`,
      ),
    ).toBeInTheDocument()
    expect(within(seine).getAllByRole('listitem')).toHaveLength(2)

    // **Je Zeile ein Preis oder „kein Preis"**, je Seite eine Summe.
    expect(await within(meine).findAllByText('19,90 €')).toHaveLength(2)
    // **Summe und Zeile sind hier gleich** — die zweite Zeile hat keinen Preis.
    expect(within(seine).getAllByText('17,95 €')).toHaveLength(2)
    expect(within(seine).getByText(eigene.keinPreis)).toBeInTheDocument()

    // **Die Differenz mit Richtung** — ich gebe mehr, also zugunsten des
    // Gegenübers; und unvollständig, weil eine Zeile keinen Preis hat.
    expect(screen.getByRole('status')).toHaveTextContent(
      '⇆ 1,95 € zugunsten deines Gegenübers',
    )
    expect(
      screen.getByText('1 ohne Preis — gerechnet ist nur, was einen hat.'),
    ).toBeInTheDocument()
    expect(
      screen.getByText(
        /^Preise vom 18\.09\.2026.*, aus scryfall\. Sie sind ein Anhalt/,
      ),
    ).toBeInTheDocument()

    // **Nur lesen**: kein Hinzufügen und kein Entfernen — Preise holen geht
    // auch nach der Einigung (P10, #2378), es schreibt nicht in den Handel.
    expect(
      screen.getAllByRole('button').map((knopf) => knopf.textContent),
    ).toEqual([texte.preise.holen])
    expect(gefragt).toHaveBeenCalledTimes(1)
  })

  it('legt eine Zeile in der Spalte an, deren Knopf gedrückt wurde', async () => {
    belegeApi({
      HandelPreisanhalt: () =>
        Promise.resolve({ anhalt: einAnhalt(), fehler: null }),
    })
    const gesetzt = vi.fn()
    const hand = userEvent.setup()

    render(<Buehne bearbeitbar anfang={[]} gesetzt={gesetzt} />)

    // **Keine Auswahl „Von wem"** — die Spalte sagt es.
    expect(screen.queryByLabelText(texte.positionVon)).toBeNull()
    expect(screen.getByText(texte.positionLeer)).toBeInTheDocument()

    await hand.click(
      screen.getByRole('button', {
        name: `Zeile hinzufügen: ${texte.positionVonGegenueber}`,
      }),
    )
    await hand.type(await kennungsfeld(hand), 'shivan')
    await hand.click(
      screen.getByRole('button', { name: texte.positionHinzufuegen }),
    )

    expect(gesetzt).toHaveBeenLastCalledWith([
      expect.objectContaining({ von: 'du-1', printing: 'shivan' }),
    ])

    // **Das Formular schließt**, und die Zeile steht in ihrer Spalte — mit
    // der Karte, die der Anhalt für die erste Zeile nennt.
    expect(screen.queryByLabelText(texte.positionPrinting)).toBeNull()
    expect(
      await within(
        screen.getByRole('region', { name: texte.positionVonGegenueber }),
      ).findByText('1 × Der Eine Ring'),
    ).toBeVisible()

    // **Abbrechen legt nichts an.**
    await hand.click(
      screen.getByRole('button', {
        name: `Zeile hinzufügen: ${texte.positionVonIch}`,
      }),
    )
    await hand.type(await kennungsfeld(hand), 'x')
    await hand.click(
      screen.getByRole('button', { name: eigene.zeileVerwerfen }),
    )
    expect(gesetzt).toHaveBeenCalledTimes(1)
  })

  it('lässt die eigene Spalte zu, solange der Ausweis fehlt', () => {
    render(<Buehne bearbeitbar anfang={[]} ich="" />)

    // **Sonst ginge eine Zeile mit leerer Seite hinaus** — der Befund vom
    // 2026-09-19, damals an der Auswahl „Von wem".
    expect(
      screen.getByRole('button', {
        name: `Zeile hinzufügen: ${texte.positionVonIch}`,
      }),
    ).toBeDisabled()
    expect(
      screen.getByRole('button', {
        name: `Zeile hinzufügen: ${texte.positionVonGegenueber}`,
      }),
    ).toBeEnabled()
  })

  it('entfernt eine Zeile aus ihrer Spalte', async () => {
    belegeApi({
      HandelPreisanhalt: () =>
        Promise.resolve({ anhalt: einAnhalt(), fehler: null }),
    })
    const gesetzt = vi.fn()
    const hand = userEvent.setup()

    render(
      <Buehne
        bearbeitbar
        anfang={[zeile('ich-1', 'ring'), zeile('du-1', 'goyf')]}
        gesetzt={gesetzt}
      />,
    )

    await hand.click(
      screen.getByRole('button', {
        name: texte.positionEntfernenName.replace('{nummer}', '2'),
      }),
    )

    expect(gesetzt).toHaveBeenLastCalledWith([
      expect.objectContaining({ printing: 'ring' }),
    ])
  })

  it('sagt es, wenn keine Karte einen Preis hat', async () => {
    belegeApi({
      HandelPreisanhalt: () =>
        Promise.resolve({
          anhalt: {
            zeilen: [
              {
                preise: [],
                druck: {
                  name: '',
                  setcode: '',
                  nummer: '',
                  bild: '',
                  eigen: false,
                },
              },
            ],
            seiten: [{ von: 'ich-1', summen: [], ohnePreis: 1 }],
            differenzen: [],
            hauptwaehrung: '',
          },
          fehler: null,
        }),
    })

    render(<Buehne bearbeitbar={false} anfang={[zeile('ich-1', 'ring')]} />)

    expect(await screen.findByText(eigene.keinePreise)).toBeInTheDocument()
  })

  it('sagt, wenn beide Seiten gleich viel geben, und rechnet zu deinen Gunsten', async () => {
    const anhalt = einAnhalt()
    anhalt.differenzen = [{ waehrung: 'EUR', wert: 0, vollstaendig: true }]
    // **Hier ist die zweite Summe die jüngere** — genannt bleibt die erste.
    anhalt.seiten = anhalt.seiten.map((seite, stelle) =>
      stelle === 1
        ? {
            ...seite,
            summen: seite.summen.map((summe) => ({
              ...summe,
              erhoben: '2026-09-21T08:00:00Z',
            })),
          }
        : seite,
    )
    belegeApi({
      HandelPreisanhalt: () => Promise.resolve({ anhalt, fehler: null }),
    })

    const { unmount } = render(
      <Buehne bearbeitbar={false} anfang={[zeile('ich-1', 'ring')]} />,
    )
    expect(await screen.findByRole('status')).toHaveTextContent(
      eigene.ausgeglichen,
    )
    unmount()

    anhalt.differenzen = [{ waehrung: 'EUR', wert: 250, vollstaendig: true }]
    render(<Buehne bearbeitbar={false} anfang={[zeile('ich-1', 'goyf')]} />)
    expect(await screen.findByRole('status')).toHaveTextContent(
      '⇆ 2,50 € zu deinen Gunsten',
    )
  })

  it('meldet, wenn die Preise sich nicht lesen lassen', async () => {
    belegeApi({
      HandelPreisanhalt: () => Promise.reject(new Error('Preise zu')),
    })

    render(<Buehne bearbeitbar={false} anfang={[zeile('ich-1', 'ring')]} />)

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Preise zu')
    })
  })
})
