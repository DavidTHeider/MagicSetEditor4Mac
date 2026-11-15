// **Ein Wechsel räumt den offenen Kontext** — §7, Regel 1
// (#1278,
// Befunde 4 und 6).
//
// # Warum zwei Fälle in einer Datei
//
// Weil es eine Regel ist und zwei Gestalten: der eine Zustand hängt an der
// **Karte** (der Ausklapper „Arbeitsstand"), der andere am **Set** (der
// Markenfilter).
//
// # Was hier **nicht** steht
//
// Der Zweitblick über die Nacharbeit hat beide als Befund erhoben — sie würden
// ihren Wechsel überleben. **Sie tun es nicht**, im laufenden Fenster
// nachgestellt: der Filter stand auf „alle" (17 Zeilen statt 1), der
// Ausklapper war zu und ein anderer DOM-Knoten.
//
// Der Grund ist, dass `Karten` die Listenspalte gar nicht zeichnet, solange
// die neue Liste unterwegs ist — §7 Regel 1 wird **durch das Abbauen**
// eingelöst. Das ist tragend und unsichtbar, und deshalb gibt es diese beiden
// Läufe: mit einem Zustand, der das Abbauen überlebt, werden sie rot.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api, KartenlisteAntwort } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { ausKarten } from '../test/kartenproben'
import { Karten } from './Karten'

const texte = kataloge.de.seiten.karten

afterEach(() => {
  raeumeApi()
})

const karten = [
  { id: '1', nummer: 0, werte: [{ feld: 'name', wert: 'Feuerball' }] },
  { id: '2', nummer: 1, werte: [{ feld: 'name', wert: 'Eisball' }] },
]

// aufbau gibt eine Liste, in der beide Karten die Marke `fertig` tragen.
function aufbau(): KartenlisteAntwort {
  return {
    spalten: [],
    zeilen: karten.map((karte) => ({
      karteID: karte.id,
      nummer: karte.nummer + 1,
      werte: { name: karte.werte[0]?.wert ?? '' },
      marken: karte.id === '1' ? ['fertig'] : [],
    })),
    sortierfeld: '',
    aufsteigend: true,
    ohneVorlage: false,
    fehler: null,
  }
}

function belege(teile: Partial<Api> = {}) {
  belegeApi({
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Vorschau: () => Promise.resolve({ vorschau: null, fehler: null }),
    Kartenbilder: () => Promise.resolve({ bilder: {}, fehler: null }),
    ...ausKarten(() => Promise.resolve({ karten })),
    Karten: () => Promise.resolve({ karten }),
    Kartenlistenaufbau: () => Promise.resolve(aufbau()),
    ...teile,
  })
}

function zeige(setID: string, karteID: string | undefined) {
  return render(
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID={setID}
        karteID={karteID}
        verlaufsstand={0}
        verlaufsfehler={undefined}
        bestandGeaendert={vi.fn()}
      />
    </Sprachraum>,
  )
}

describe('der Arbeitsstand beim Kartenwechsel', () => {
  it('klappt wieder zu', async () => {
    belege()

    const { rerender } = zeige('1', '1')

    const ausklapper = (await screen.findByText(texte.arbeitsstand)).closest(
      'details',
    )

    expect(ausklapper).not.toBeNull()

    await userEvent.click(screen.getByText(texte.arbeitsstand))
    expect(ausklapper).toHaveAttribute('open')

    // **Die andere Karte.** `Karten` bleibt stehen, nur die Kennung wechselt —
    // genau das, was ein Klick in der Liste tut.
    rerender(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="1"
          karteID="2"
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={vi.fn()}
        />
      </Sprachraum>,
    )

    const danach = (await screen.findByText(texte.arbeitsstand)).closest(
      'details',
    )

    expect(danach).not.toHaveAttribute('open')
  })
})

describe('der Markenfilter beim Setwechsel', () => {
  it('steht wieder auf „alle"', async () => {
    belege()

    const { rerender } = zeige('1', '1')

    // **Der Filter ist eine Reihe von Optionsfeldern**, keine Auswahlliste:
    // wer eine Marke anklickt, sieht die Liste im selben Zeichenschritt.
    const gewaehlt = await screen.findByRole('radio', { name: 'fertig' })

    await userEvent.click(gewaehlt)
    expect(gewaehlt).toBeChecked()

    // **Ein anderes Set, das dieselbe Marke kennt.** Genau der Fall, den
    // `wirksam` nicht fängt: die Marke gibt es, also blieb die Wahl stehen —
    // und das neue Set stand gefiltert da, ohne dass jemand gefiltert hätte.
    rerender(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="2"
          karteID="1"
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(
      await screen.findByRole('radio', { name: texte.marken.alle }),
    ).toBeChecked()
  })
})
