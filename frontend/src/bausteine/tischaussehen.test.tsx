// **Wie der Tisch aussieht** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Eine dritte Datei zum Tisch
//
// **Der Schnitt folgt der Frage** (§2): `tischflaeche.test.tsx` prüft, *was*
// er zeigt, `tischziehen.test.tsx`, *wie* er bedient wird — hier steht, *wie
// er aussieht*: welcher Hintergrund gilt und ob sich etwas bewegt.
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, onTestFinished, vi } from 'vitest'

import { bildwort } from './Tischsteuerung'
import { einObjekt, einePartie } from '../seiten/partieprobe'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Tischflaeche } from './Tischflaeche'
import { tischbilder } from '../tisch/tischbilder'
import { tischplan } from '../tisch/tischplan'

import type { Tischmalerei } from '../tisch/maler'
import { einMaler as einMalerOhne } from '../test/malerprobe'

const eigene = kataloge.de.seiten.spielen

const planMit = (...teile: Parameters<typeof einObjekt>[0][]) =>
  tischplan(
    einePartie({
      objekte: teile.map((eines) =>
        einObjekt({
          eigner: 'sitz1',
          beherrscher: 'sitz1',
          geheim: false,
          ...eines,
        }),
      ),
    }),
  )

const ohneGrafik = () =>
  Promise.resolve<Tischmalerei>({ art: 'ohne', grund: 'kein-webgl' })

const zeige = (
  plan: ReturnType<typeof tischplan>,
  malerei: () => Promise<Tischmalerei>,
) =>
  render(
    <Sprachraum sprache="de">
      <Tischflaeche plan={plan} bild={() => undefined} malerei={malerei} />
    </Sprachraum>,
  )

describe('die Tischbildwahl', () => {
  const mitWahl = (tischbild?: string) => {
    const aufTischbild = vi.fn()
    const maler = einMalerOhne({ breite: 800, hoehe: 600 })

    render(
      <Sprachraum sprache="de">
        <Tischflaeche
          plan={planMit()}
          bild={() => undefined}
          malerei={() => Promise.resolve<Tischmalerei>({ art: 'da', maler })}
          aufTischbild={aufTischbild}
          {...(tischbild === undefined ? {} : { tischbild })}
        />
      </Sprachraum>,
    )

    return aufTischbild
  }

  it('steht auf Holz, wo noch nichts gewählt ist', async () => {
    mitWahl()

    expect(
      await screen.findByRole('combobox', { name: eigene.tischbild }),
    ).toHaveValue('holz')
  })

  it('zeigt die gespeicherte Wahl', async () => {
    mitWahl('taverne')

    expect(
      await screen.findByRole('combobox', { name: eigene.tischbild }),
    ).toHaveValue('taverne')
  })

  it('meldet eine neue Wahl', async () => {
    const aufTischbild = mitWahl('holz')

    fireEvent.change(
      await screen.findByRole('combobox', { name: eigene.tischbild }),
      { target: { value: 'dungeon' } },
    )

    expect(aufTischbild).toHaveBeenCalledWith('dungeon')
  })

  it('zeigt gar keine Wahl, wo niemand auf sie hört', async () => {
    zeige(planMit(), ohneGrafik)

    await screen.findByRole('application')

    expect(
      screen.queryByRole('combobox', { name: eigene.tischbild }),
    ).not.toBeInTheDocument()
  })
})

describe('bildwort', () => {
  it('übersetzt jeden der sechs Namen', () => {
    for (const eines of tischbilder)
      expect(bildwort(eigene, eines.name)).not.toBe(eines.name)
  })

  it('setzt einen Bindestrichnamen zusammen', () => {
    expect(bildwort(eigene, 'dunkles-holz')).toBe(eigene.tischbildDunklesHolz)
  })

  // **Ein Name ohne Eintrag lässt die Wahl sonst leer dastehen** — dieselbe
  // Zusage wie bei `zonenwort` (§8).
  it('gibt die Kennung her, wo kein Wort dasteht', () => {
    expect(bildwort(eigene, 'gibtesnicht')).toBe('gibtesnicht')
  })
})

describe('Bewegung', () => {
  const zeigeMit = (bewegt?: boolean) => {
    const maler = einMalerOhne({ breite: 800, hoehe: 600 })

    render(
      <Sprachraum sprache="de">
        <Tischflaeche
          plan={planMit({ kennung: 'a', zone: 'tisch', name: 'Wald' })}
          bild={() => undefined}
          malerei={() => Promise.resolve<Tischmalerei>({ art: 'da', maler })}
          {...(bewegt === undefined ? {} : { bewegt })}
        />
      </Sprachraum>,
    )

    return maler
  }

  it('lässt Karten fliegen, wo nichts dagegen spricht', async () => {
    const maler = zeigeMit()

    await waitFor(() => {
      expect(maler.zeichne).toHaveBeenCalled()
    })

    expect(maler.zeichne).toHaveBeenLastCalledWith(
      expect.objectContaining({ bewegt: true }),
    )
  })

  it('lässt sie stehen, wo die Anwendung es abschaltet', async () => {
    const maler = zeigeMit(false)

    await waitFor(() => {
      expect(maler.zeichne).toHaveBeenCalled()
    })

    expect(maler.zeichne).toHaveBeenLastCalledWith(
      expect.objectContaining({ bewegt: false }),
    )
  })

  // **`prefers-reduced-motion` hat die stärkere Stimme** — wer sie setzt, hat
  // einen Grund, und die Anwendung darf ihn nicht übergehen.
  it('folgt der Systemeinstellung, auch gegen die eigene Vorgabe', async () => {
    // **jsdom bringt `matchMedia` gar nicht mit** — deshalb wird hier
    // gesetzt und danach wieder gelöscht statt zurückgelegt. Das ist zugleich
    // der Beleg dafür, dass der Rückfall in `bewegungErwuenscht` greift: ohne
    // diesen Block antwortet niemand, und die Anwendung entscheidet allein.
    window.matchMedia = ((frage: string) => ({
      matches: frage.includes('reduced-motion'),
      media: frage,
      addEventListener: () => undefined,
      removeEventListener: () => undefined,
    })) as unknown as typeof window.matchMedia

    const maler = zeigeMit(true)

    await waitFor(() => {
      expect(maler.zeichne).toHaveBeenCalled()
    })

    expect(maler.zeichne).toHaveBeenLastCalledWith(
      expect.objectContaining({ bewegt: false }),
    )

    Reflect.deleteProperty(window, 'matchMedia')
  })

  // **Und der Einstellung der Anwendung** — UX-10.AK-03 (#1682): „Immer
  // wenig" steht am Wurzelelement, und die Karten bleiben stehen.
  it('folgt der Einstellung „Immer wenig"', async () => {
    document.documentElement.dataset.bewegung = 'reduziert'
    onTestFinished(() => {
      delete document.documentElement.dataset.bewegung
    })

    const maler = zeigeMit(true)

    await waitFor(() => {
      expect(maler.zeichne).toHaveBeenCalled()
    })

    expect(maler.zeichne).toHaveBeenLastCalledWith(
      expect.objectContaining({ bewegt: false }),
    )
  })
})

describe('das eigene Tischbild', () => {
  const mitEigenem = (hintergrund?: string) => {
    const eigenesBild = vi.fn()
    const eigenesBildWeg = vi.fn()
    const maler = einMalerOhne({ breite: 800, hoehe: 600 })

    render(
      <Sprachraum sprache="de">
        <Tischflaeche
          plan={planMit()}
          bild={() => undefined}
          malerei={() => Promise.resolve<Tischmalerei>({ art: 'da', maler })}
          eigenesBild={eigenesBild}
          eigenesBildWeg={eigenesBildWeg}
          {...(hintergrund === undefined ? {} : { hintergrund })}
        />
      </Sprachraum>,
    )

    return { eigenesBild, eigenesBildWeg, maler }
  }

  it('bietet an, ein eigenes Bild zu wählen', async () => {
    const { eigenesBild } = mitEigenem()

    fireEvent.click(
      await screen.findByRole('button', { name: eigene.tischEigenes }),
    )

    expect(eigenesBild).toHaveBeenCalled()
  })

  // **Ein Knopf, der nichts tut, sagt nicht, dass nichts da ist** — er sagt,
  // dass etwas nicht ging.
  it('bietet das Wegnehmen erst an, wo etwas wegzunehmen ist', async () => {
    mitEigenem()

    await screen.findByRole('button', { name: eigene.tischEigenes })

    expect(
      screen.queryByRole('button', { name: eigene.tischEigenesWeg }),
    ).not.toBeInTheDocument()
  })

  it('nimmt es weg, wo eines liegt', async () => {
    const { eigenesBildWeg } = mitEigenem('/tischbild/1.png')

    fireEvent.click(
      await screen.findByRole('button', { name: eigene.tischEigenesWeg }),
    )

    expect(eigenesBildWeg).toHaveBeenCalled()
  })

  it('reicht die Adresse an den Maler weiter', async () => {
    const { maler } = mitEigenem('/tischbild/1.png')

    await waitFor(() => {
      expect(maler.zeichne).toHaveBeenCalled()
    })

    expect(maler.zeichne).toHaveBeenLastCalledWith(
      expect.objectContaining({ hintergrund: '/tischbild/1.png' }),
    )
  })

  it('zeigt gar keine Wahl, wo niemand auf sie hört', async () => {
    zeige(planMit(), ohneGrafik)

    await screen.findByRole('application')

    expect(
      screen.queryByRole('button', { name: eigene.tischEigenes }),
    ).not.toBeInTheDocument()
  })
})

// **Der Zonenname kommt aus dem Katalog bis auf die Leinwand** —
// PLAY-03.AK-06
// (#1711).
//
// # Warum dieser Lauf über den Maler geht und nicht über den DOM
//
// **Weil es auf der Leinwand nichts zu greifen gibt.** Ein Zonenname ist unter
// PixiJS kein Element; was ein DOM-Lauf sehen kann, ist nur, **was dem Maler
// übergeben wurde**. Genau das ist hier die Zusage: die Fläche reicht eine
// Beschriftung durch, und sie ist die aus dem Sprachkatalog — nicht der rohe
// Zonenschlüssel und nicht ein erfundener Name.
describe('die Beschriftung der Zonen', () => {
  it('reicht die Zonennamen aus dem Katalog an den Maler', async () => {
    const maler = einMalerOhne({ breite: 800, hoehe: 600 })

    render(
      <Sprachraum sprache="de">
        <Tischflaeche
          plan={planMit()}
          bild={() => undefined}
          malerei={() => Promise.resolve<Tischmalerei>({ art: 'da', maler })}
        />
      </Sprachraum>,
    )

    // **Gewartet wird auf die Zusicherung selbst und nicht auf einen Ruf.**
    // Der Maler entsteht asynchron; bis er da ist, hat niemand gezeichnet.
    // Ein `toHaveBeenCalled` davor wäre das Warten auf den Ruf statt auf
    // seine Wirkung — und die Wirkung ist hier, **was** übergeben wurde.
    await waitFor(() => {
      const wunsch = maler.zeichne.mock.calls.at(-1)?.[0] as
        { beschriftung?: (zone: string) => string } | undefined

      expect(wunsch?.beschriftung?.('bibliothek')).toBe(eigene.zoneBibliothek)
    })
  })

  // **Und über einem Stapel die Kurzform** — seit dem 2026-09-24: über 63
  // Einheiten passt „Bibliothek" nicht (`zonenschilder.ts`).
  it('reicht über einem Stapel die Kurzform an den Maler', async () => {
    const maler = einMalerOhne({ breite: 800, hoehe: 600 })

    render(
      <Sprachraum sprache="de">
        <Tischflaeche
          plan={planMit()}
          bild={() => undefined}
          malerei={() => Promise.resolve<Tischmalerei>({ art: 'da', maler })}
        />
      </Sprachraum>,
    )

    await waitFor(() => {
      const wunsch = maler.zeichne.mock.calls.at(-1)?.[0] as
        { beschriftung?: (zone: string, kurz?: boolean) => string } | undefined

      expect(wunsch?.beschriftung?.('bibliothek', true)).toBe(
        eigene.zonekurzBibliothek,
      )
    })
  })
})

// **Die zwei neuen Wege stehen als Knöpfe da** — PLAY-05.AK-05
// (#1713):
// *„Auf die Auswahl zoomen" und „Auf den eigenen Sitz zoomen" sind unmittelbar
// erreichbar.* „Unmittelbar" heißt: ein Knopf, nicht ein Kürzel und nicht ein
// Menü.
describe('die Tischsteuerung', () => {
  const zeigeSteuerung = () => {
    const maler = einMalerOhne({ breite: 800, hoehe: 600 })

    render(
      <Sprachraum sprache="de">
        <Tischflaeche
          plan={planMit()}
          bild={() => undefined}
          malerei={() => Promise.resolve<Tischmalerei>({ art: 'da', maler })}
        />
      </Sprachraum>,
    )
  }

  it('bietet den Weg zur Auswahl an', async () => {
    zeigeSteuerung()

    expect(
      await screen.findByRole('button', { name: eigene.tischAuswahl }),
    ).toBeInTheDocument()
  })

  it('bietet den Weg zum eigenen Sitz an', async () => {
    zeigeSteuerung()

    expect(
      await screen.findByRole('button', { name: eigene.tischEigenerSitz }),
    ).toBeInTheDocument()
  })
})
