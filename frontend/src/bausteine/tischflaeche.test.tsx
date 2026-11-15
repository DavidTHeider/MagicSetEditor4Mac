// **Der gezeichnete Tisch** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Wogegen hier geprüft wird
//
// **Dass jede Karte ein Bedienelement ist und nicht nur ein Bild.** Unter
// jsdom gibt es kein WebGL — was hier steht, ist deshalb genau das, was ein
// Rechner ohne Grafikkarte sieht, und der Lauf belegt, dass er bedienbar
// bleibt.
//
// **Und dass derselbe Weg mit Maler dieselben Knöpfe trägt.** Sonst wären es
// zwei Oberflächen, von denen eine ungeprüft bliebe.
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { einObjekt, einePartie } from '../seiten/partieprobe'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Tischflaeche } from './Tischflaeche'
import { tischplan } from './../tisch/tischplan'

import type { Tischmalerei } from '../tisch/maler'
import type { Klick } from '../spiel/kartengriff'
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

const zeige = (
  plan: ReturnType<typeof tischplan>,
  malerei: () => Promise<Tischmalerei>,
  aufKarte?: (kennung: string, klick: Klick) => void,
  gewaehlt?: string,
) =>
  render(
    <Sprachraum sprache="de">
      <Tischflaeche
        plan={plan}
        bild={() => undefined}
        malerei={malerei}
        {...(aufKarte ? { aufKarte } : {})}
        {...(gewaehlt === undefined ? {} : { gewaehlt })}
      />
    </Sprachraum>,
  )

const ohneGrafik = () =>
  Promise.resolve<Tischmalerei>({ art: 'ohne', grund: 'kein-webgl' })

describe('Tischflaeche', () => {
  it('macht aus jeder Karte ein Bedienelement', async () => {
    zeige(planMit({ kennung: 'a', zone: 'tisch', name: 'Wald' }), ohneGrafik)

    expect(
      await screen.findByRole('button', { name: /Wald/ }),
    ).toBeInTheDocument()
  })

  it('nennt Zone und Zustand im Namen des Knopfes', async () => {
    zeige(
      planMit({
        kennung: 'a',
        zone: 'tisch',
        name: 'Bär',
        getappt: true,
      }),
      ohneGrafik,
    )

    const knopf = await screen.findByRole('button', { name: /Bär/ })

    expect(knopf).toHaveAccessibleName(
      new RegExp(`${eigene.zoneTisch}.*${eigene.tischGetappt}`),
    )
  })

  it('sagt verdeckt, wo kein Name dasteht', async () => {
    zeige(
      planMit({ kennung: 'a', zone: 'tisch', name: '', verdeckt: true }),
      ohneGrafik,
    )

    // **Der Kartenknopf und nicht irgendeiner** — seit UX-15
    // (#1771)
    // trägt der Hilfeknopf neben der Erläuterung des Tischabzugs den Satz
    // „… eine verdeckte Karte bleibt auch im Bild verdeckt" im Namen. Ein
    // Suchwort allein sagte damit etwas über die Fläche statt über die Karte
    // (§10). Der Name einer Karte **fängt** mit ihr an.
    expect(
      await screen.findByRole('button', {
        name: new RegExp(`^${eigene.tischVerdeckt}`),
      }),
    ).toBeInTheDocument()
  })

  it('sagt verdeckt genau einmal, wo kein Name dasteht', async () => {
    zeige(
      planMit({ kennung: 'a', zone: 'tisch', name: '', verdeckt: true }),
      ohneGrafik,
    )
    const knopf = await screen.findByRole('button', {
      name: new RegExp(`^${eigene.tischVerdeckt}`),
    })

    expect(
      (knopf.textContent.match(new RegExp(eigene.tischVerdeckt, 'g')) ?? [])
        .length,
    ).toBe(1)
  })

  it('nennt den Zustand neben dem Namen, wo beides bekannt ist', async () => {
    zeige(
      planMit({ kennung: 'a', zone: 'tisch', name: 'Wald', verdeckt: true }),
      ohneGrafik,
    )

    expect(
      await screen.findByRole('button', { name: /Wald/ }),
    ).toHaveAccessibleName(new RegExp(`Wald.*${eigene.tischVerdeckt}`))
  })

  // **CR 400.5** (Fassung 2026-08-19): was woran hängt, welche Marken
  // daraufliegen und ob etwas getappt ist, muss für **alle** klar bleiben —
  // und „alle" schließt ein, wer die Karte hört statt sieht. Eine Linie auf
  // der Leinwand erreicht ein Vorleseprogramm nicht.
  it('nennt die Marken einer Karte in ihrer Beschriftung', async () => {
    zeige(
      planMit({
        kennung: 'a',
        zone: 'tisch',
        name: 'Bär',
        zaehler: { plus: 3, aufladung: 1 },
      }),
      ohneGrafik,
    )

    expect(
      await screen.findByRole('button', { name: /Bär/ }),
    ).toHaveAccessibleName(/1 Marke aufladung.*3 Marken plus/)
  })

  it('nennt, woran ein Anhängsel hängt', async () => {
    zeige(
      planMit(
        { kennung: 'ziel', zone: 'tisch', name: 'Bär' },
        { kennung: 'aura', zone: 'tisch', name: 'Segen', haengtAn: 'ziel' },
      ),
      ohneGrafik,
    )

    expect(
      await screen.findByRole('button', { name: /Segen/ }),
    ).toHaveAccessibleName(new RegExp(`${eigene.tischHaengtAn} ziel`))
  })

  it('lässt eine Marke mit der Zahl null weg', () => {
    // **Null Marken sind keine Marken** (CR 122.1).
    zeige(
      planMit({
        kennung: 'a',
        zone: 'tisch',
        name: 'Bär',
        zaehler: { plus: 0 },
      }),
      ohneGrafik,
    )

    expect(
      screen.getByRole('button', { name: /Bär/ }),
    ).not.toHaveAccessibleName(/plus/)
  })

  it('hebt die gewählte Karte hervor und die anderen nicht', async () => {
    zeige(
      planMit(
        { kennung: 'a', zone: 'tisch', name: 'Wald' },
        { kennung: 'b', zone: 'tisch', name: 'Berg' },
      ),
      ohneGrafik,
      undefined,
      'a',
    )

    expect(await screen.findByRole('button', { name: /Wald/ })).toHaveAttribute(
      'aria-pressed',
      'true',
    )
    expect(screen.getByRole('button', { name: /Berg/ })).toHaveAttribute(
      'aria-pressed',
      'false',
    )
  })

  it('erklärt, warum nicht gezeichnet wird, und bleibt bedienbar', async () => {
    zeige(planMit({ kennung: 'a', zone: 'tisch', name: 'Wald' }), ohneGrafik)

    expect(await screen.findByRole('alert')).toHaveTextContent(
      eigene.tischOhneWebgl,
    )
    expect(screen.getByRole('button', { name: /Wald/ })).toBeInTheDocument()
  })

  it('nennt jeden der drei Gründe mit eigenen Worten', async () => {
    const saetze = [
      ['kein-renderer', eigene.tischOhneRenderer],
      ['verloren', eigene.tischVerloren],
    ] as const

    for (const [grund, satz] of saetze) {
      const stand = zeige(planMit(), () =>
        Promise.resolve<Tischmalerei>({ art: 'ohne', grund }),
      )

      expect(await screen.findByRole('alert')).toHaveTextContent(satz)
      stand.unmount()
    }
  })

  it('gibt einen Knopf für jede Kamerahandlung', async () => {
    const maler = einMalerOhne({ breite: 800, hoehe: 600 })
    zeige(planMit(), () => Promise.resolve<Tischmalerei>({ art: 'da', maler }))

    for (const titel of [
      eigene.tischEinpassen,
      eigene.tischNaeher,
      eigene.tischWeiter,
      eigene.tischDrehen,
    ]) {
      fireEvent.click(await screen.findByRole('button', { name: titel }))
    }

    expect(maler.zeichne).toHaveBeenCalled()
  })

  it('bedient die Kamera auch über die Tastatur', async () => {
    const maler = einMalerOhne({ breite: 800, hoehe: 600 })
    zeige(planMit(), () => Promise.resolve<Tischmalerei>({ art: 'da', maler }))
    const flaeche = await screen.findByRole('application')
    const vorher = maler.zeichne.mock.calls.length

    for (const taste of [
      'ArrowLeft',
      'ArrowRight',
      'ArrowUp',
      'ArrowDown',
      '+',
      '-',
      '0',
      'q',
    ]) {
      fireEvent.keyDown(flaeche, { key: taste })
    }

    expect(maler.zeichne.mock.calls.length).toBeGreaterThan(vorher)
  })

  // **Mit Anzahl und ⌘** — P4 (#2372): ein Klick löst die Primärhandlung
  // aus, ein Doppelklick die zweite, ⌘ wählt. Was davon gilt, entscheidet
  // der Aufrufer; die Fläche reicht durch, was geschah.
  it('meldet, worauf gedrückt wurde, und wie', async () => {
    const aufKarte = vi.fn()
    zeige(
      planMit({ kennung: 'a', zone: 'tisch', name: 'Wald' }),
      ohneGrafik,
      aufKarte,
    )
    const knopf = await screen.findByRole('button', { name: /Wald/ })

    fireEvent.click(knopf, { detail: 1 })
    fireEvent.click(knopf, { detail: 2, metaKey: true })

    expect(aufKarte.mock.calls).toEqual([
      ['a', { detail: 1, metaKey: false }],
      ['a', { detail: 2, metaKey: true }],
    ])
  })

  it('bleibt still, wo niemand zuhört', async () => {
    zeige(planMit({ kennung: 'a', zone: 'tisch', name: 'Wald' }), ohneGrafik)

    fireEvent.click(await screen.findByRole('button', { name: /Wald/ }))

    expect(screen.getByRole('button', { name: /Wald/ })).toBeInTheDocument()
  })

  // **Im gezeichneten Stapel ist nur die oberste ein Knopf** — #2472: die
  // übrigen lagen 0,4 Punkte darunter, und axe meldete in WebKit verdeckte
  // Ziele (WCAG 2.2, 2.5.8). **Ohne Zeichnung stehen alle da** — die Liste
  // ist der Weg zu jeder einzelnen.
  it('macht im gezeichneten Stapel nur die oberste Karte zum Knopf', async () => {
    const stapel = planMit(
      { kennung: 'b0', zone: 'bibliothek', name: '', verdeckt: true },
      { kennung: 'b1', zone: 'bibliothek', name: '', verdeckt: true },
      { kennung: 'b2', zone: 'bibliothek', name: '', verdeckt: true },
    )
    const maler = einMalerOhne({ breite: 800, hoehe: 600 })
    const gezeichnet = zeige(stapel, () =>
      Promise.resolve<Tischmalerei>({ art: 'da', maler }),
    )
    await screen.findByRole('button', { name: eigene.tischEinpassen })

    expect(
      screen.getAllByRole('button', {
        name: new RegExp(eigene.zoneBibliothek),
      }),
    ).toHaveLength(1)

    gezeichnet.unmount()
    zeige(stapel, ohneGrafik)

    // **Gewartet wird auf die Liste** — bis die Malerei absagt, steht die
    // Fläche als gezeichnete da.
    await waitFor(() => {
      expect(
        screen.getAllByRole('button', {
          name: new RegExp(eigene.zoneBibliothek),
        }),
      ).toHaveLength(3)
    })
  })

  it('nennt die Objekte, für die kein Platz gefunden wurde', async () => {
    zeige(
      planMit({ kennung: 'nirgends', zone: 'ante', name: 'Einsatz' }),
      ohneGrafik,
    )

    expect(
      await screen.findByText(new RegExp(eigene.tischHeimatlos)),
    ).toHaveTextContent('nirgends')
  })
})
