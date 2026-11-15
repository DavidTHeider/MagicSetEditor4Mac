// **Die Anfasser und die Zahlenfelder der Symmetrie**
// (#1092).
//
// **Eine eigene Datei**, weil `symbolsymmetrie.test.tsx` über die Zeilengrenze
// aus §2 wuchs. Der Schnitt folgt der Frage: dort, **dass** es die Symmetrie
// gibt; hier, **wie** man sie bedient.
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { leererSetkopf } from '../test/setkopfprobe'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { mitSymmetrie, zweiKopien } from '../test/symbolproben'
import { Symbol } from './Symbol'
import { kurzform } from '../sprache/kurzform'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.symbol

// zeigeSymmetrie stellt den Editor auf ein geladenes Symbol mit Symmetrie.
function zeigeSymmetrie(weitere: Record<string, unknown> = {}) {
  belegeApi({
    Setangaben: vi.fn().mockResolvedValue({
      felder: [],
      werte: [{ feld: 'symbol', wert: 'probe.mse-symbol' }],
      kopf: leererSetkopf,
      fehler: null,
    }),
    Symbol: vi.fn().mockResolvedValue(mitSymmetrie()),
    Symmetriekopien: vi.fn().mockResolvedValue(zweiKopien()),
    ...weitere,
  })

  render(
    <Sprachraum sprache="de">
      <Symbol setID="1" />
    </Sprachraum>,
  )
}

describe('Die vier Zahlenfelder', () => {
  it('setzen Mitte und Griff — jedes für sich', async () => {
    const SymmetrieSetzen = vi.fn().mockResolvedValue(mitSymmetrie())

    zeigeSymmetrie({ SymmetrieSetzen })

    // **Alle vier**, und nicht eines stellvertretend: jedes Feld trägt seine
    // eigene Zuordnung, und ein vertauschtes X und Y fiele nur an dem einen
    // auf, das es betrifft.
    for (const [beschriftung, erwartet] of [
      [texte.mitteX, { mitteX: 0.3, mitteSetzen: true }],
      [texte.mitteY, { mitteY: 0.3, mitteSetzen: true }],
      [texte.griffX, { griffX: 0.3, griffSetzen: true }],
      [texte.griffY, { griffY: 0.3, griffSetzen: true }],
    ] as const) {
      fireEvent.change(await screen.findByLabelText(beschriftung), {
        target: { value: '0.3' },
      })

      await waitFor(() => {
        expect(SymmetrieSetzen).toHaveBeenCalledWith(
          expect.anything(),
          expect.objectContaining(erwartet),
        )
      })
    }
  })
})

describe('Ein Anfasser', () => {
  it('lässt jede andere Taste durch', async () => {
    const SymmetrieSetzen = vi.fn().mockResolvedValue(mitSymmetrie())

    zeigeSymmetrie({ SymmetrieSetzen })

    const mitte = await screen.findByRole('button', { name: texte.mitte })
    mitte.focus()

    // **Ein Buchstabe bewegt nichts.** Ohne diesen Fall stünde nur fest, dass
    // die Pfeiltasten etwas tun — nicht, dass alles andere in Ruhe gelassen
    // wird. Und eine Verneinung braucht einen Beleg, dass jemand zuhört (§3):
    // der Pfeil danach ist er.
    await userEvent.keyboard('x')
    expect(SymmetrieSetzen).not.toHaveBeenCalled()

    await userEvent.keyboard('{ArrowUp}')
    await waitFor(() => {
      expect(SymmetrieSetzen).toHaveBeenCalled()
    })
  })
})

describe('Eine Spiegelung', () => {
  it('sagt, warum eine ungerade Kopienzahl abrundet', async () => {
    zeigeSymmetrie({
      Symbol: vi.fn().mockResolvedValue({
        ...mitSymmetrie(),
        teile: mitSymmetrie().teile.map((teil) => ({
          ...teil,
          spiegelart: 'reflection',
        })),
      }),
    })

    // **Der Hinweis steht nur bei einer Spiegelung** — eine Drehung braucht
    // keine Paare, und ein Satz, der immer dasteht, wird nicht gelesen.
    expect(
      await screen.findByText(kurzform(texte.kopienHinweis)),
    ).toBeInTheDocument()
  })

  it('und eine Drehung sagt es nicht', async () => {
    zeigeSymmetrie()

    await screen.findByRole('button', { name: texte.drehung })

    expect(
      screen.queryByText(kurzform(texte.kopienHinweis)),
    ).not.toBeInTheDocument()
  })
})

describe('Ein Anfasser mit Umschalt', () => {
  it('springt um ein Zehntel statt um ein Hundertstel', async () => {
    const SymmetrieSetzen = vi.fn().mockResolvedValue(mitSymmetrie())

    zeigeSymmetrie({ SymmetrieSetzen })

    const mitte = await screen.findByRole('button', { name: texte.mitte })
    mitte.focus()

    // **Dieselbe Feinheit wie bei einem Punkt** (§6): fein genug, um etwas
    // auszurichten, grob genug, um in wenigen Anschlägen über die Fläche zu
    // kommen.
    await userEvent.keyboard('{Shift>}{ArrowRight}{/Shift}')

    await waitFor(() => {
      expect(SymmetrieSetzen).toHaveBeenCalledWith(
        expect.anything(),
        expect.objectContaining({ mitteX: 0.6 }),
      )
    })
  })
})

describe('Eine Kopie ohne Punkte', () => {
  it('zeichnet nichts, statt zu stolpern', async () => {
    const Symmetriekopien = vi.fn().mockResolvedValue({
      kopien: [[{ punkte: [] }], [{ punkte: [] }]],
      fehler: null,
    })

    zeigeSymmetrie({ Symmetriekopien })

    // **Gewartet wird auf den Spion, nicht auf den Anfasser** (§3).
    //
    // Der Anfasser steht da, sobald das gewählte Teil eine Symmetrie ist —
    // also **bevor** die Kopien ankommen. Eine Zusicherung darauf träfe schon
    // vor der Antwort zu, und der Zweig, um den es geht, bliebe unbetreten.
    await waitFor(() => {
      expect(Symmetriekopien).toHaveBeenCalled()
    })

    // **Sie kommt vor**: eine Symmetrie über einem Teil, dessen Punkte jemand
    // einzeln gelöscht hat. Der Editor zeigt dann die Anfasser und sonst
    // nichts — kein leerer Pfad, kein Absturz.
    const mitte = await screen.findByRole('button', { name: texte.mitte })
    expect(mitte).toBeInTheDocument()

    await waitFor(() => {
      const pfade = [...(mitte.closest('svg')?.querySelectorAll('path') ?? [])]

      expect(pfade.every((pfad) => pfad.getAttribute('d') === '')).toBe(true)
    })
  })
})

describe('Eine überholte Kopienanfrage', () => {
  it('trägt ihren Fehler nicht mehr ein', async () => {
    // **Der Fall aus §7, Regel 2.** Wer den Griff zieht, schickt eine Anfrage
    // je Bewegung; sie kommen über dieselbe Naht zurück, aber nicht zwingend
    // in der Reihenfolge, in der sie hinausgingen. Eine **gescheiterte**
    // überholte Antwort trüge sonst eine Meldung ein, die zu einem Stand
    // gehört, den niemand mehr sieht.
    let ersteAbweisen: (grund: unknown) => void = () => undefined
    let rufe = 0

    const Symmetriekopien = vi.fn(() => {
      rufe += 1

      if (rufe === 1) {
        return new Promise((_erfuellen, abweisen) => {
          ersteAbweisen = abweisen
        })
      }

      return Promise.resolve(zweiKopien())
    })

    zeigeSymmetrie({
      Symmetriekopien,
      SymmetrieSetzen: vi.fn().mockResolvedValue({
        ...mitSymmetrie(),
        teile: mitSymmetrie().teile.map((teil) => ({ ...teil, kopien: 4 })),
      }),
    })

    // **Die zweite Anfrage geht hinaus, bevor die erste antwortet.**
    fireEvent.change(await screen.findByLabelText(texte.kopien), {
      target: { value: '4' },
    })

    await waitFor(() => {
      expect(rufe).toBeGreaterThan(1)
    })

    ersteAbweisen(new Error('die überholte Anfrage'))

    // **Und ihre Meldung erscheint nicht.** Der Beleg, dass jemand zuhört,
    // steht daneben: das Feld trägt die 4 aus der zweiten Antwort.
    await waitFor(() => {
      expect(screen.getByLabelText(texte.kopien)).toHaveValue(4)
    })

    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })
})
