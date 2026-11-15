// **Der Treffer wird markiert, nicht kopiert** — UX-04.AK-03
// (#1676).
//
// # Der Anlass
//
// **Die Suche zeichnete eine zweite Liste über dem Gebiet** — mit einer
// zweiten Kopie desselben Bedienelements. Wer in der Gruppe suchte, in der
// der Treffer liegt, sah ihn **zweimal**; gemessen am laufenden Fenster am
// 2026-09-21, nachdem der Proxy überhaupt erst auffindbar wurde. Das ist
// dieselbe Klasse wie
// #1768, nur
// eine Ebene höher.
//
// **Das Artboard löst es anders**: keine Trefferliste, sondern der Kasten am
// Platz mit der Marke *gesucht* und „1 Treffer für „proxy"" im Kopf der
// Fläche.
//
// # Was er nicht prüft
//
// **Wie es aussieht.** Ob die Marke zu erkennen ist und ihr Ring 3:1 hält,
// sieht nur ein Mensch — gemessen steht es in `docs/KONTRASTE.md`.
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import {
  eintraege,
  zeigeEinstellungen as zeige,
} from '../test/einstellungsseite'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { EinstellungDTO } from '../api'

const eigene = kataloge.de.seiten.einstellungen

const proxy: EinstellungDTO = {
  schluessel: 'netz.proxy.http',
  wert: '',
  wirksam: '',
  werte: [],
  bereich: 'netz',
  ebene: 'geraet',
  herkunft: 'vorgabe',
  vorgabe: '',
  wirkung: 'sofort',
}

const alleEinstellungen = [...eintraege, proxy]

async function suche(frage: string) {
  await userEvent.type(screen.getByLabelText(eigene.suche), frage)
}

function feld() {
  return screen.getByRole('tabpanel')
}

describe('die Trefferanzeige', () => {
  // **Die Gruppe kommt mit.** Wer „proxy" tippt, während *Allgemein* offen
  // ist, soll nicht raten müssen, wo der Treffer liegt.
  it('öffnet die erste Gruppe mit einem Treffer', async () => {
    zeige({ alleEinstellungen })

    await suche('proxy')

    expect(
      // **Mit Regex, denn die Überschrift trägt die Trefferzahl mit** —
      // „Verbindungen 1 Treffer für „proxy"".
      within(feld()).getByRole('heading', {
        name: new RegExp(eigene.gruppen.verbindungen),
      }),
    ).toBeInTheDocument()
  })

  it('nennt die Trefferzahl im Kopf der Fläche', async () => {
    zeige({ alleEinstellungen })

    await suche('proxy')

    expect(
      within(feld()).getByText(
        einsetzen(eigene.trefferInGruppe, { zahl: '1', frage: 'proxy' }),
      ),
    ).toBeInTheDocument()
  })

  // **Und zwar genau einmal.** Das war der Befund: derselbe Schlüssel stand
  // im Trefferkasten und darunter noch einmal in seiner Gruppe.
  it('zeichnet den Treffer nicht doppelt', async () => {
    zeige({ alleEinstellungen })

    await suche('proxy')

    expect(
      feld().querySelectorAll('[data-einstellung="netz.proxy.http"]'),
    ).toHaveLength(1)
  })

  it('markiert das getroffene Feld', async () => {
    zeige({ alleEinstellungen })

    await suche('proxy')

    expect(within(feld()).getByText(eigene.gesuchtMarke)).toBeInTheDocument()
  })

  // **Die Spalte sagt, wo sonst noch etwas liegt.** Ohne sie wäre ein
  // Treffer in einer anderen Gruppe unsichtbar — der Knopf „Im Gebiet X
  // zeigen" hat das vorher getan.
  it('zählt die Treffer an jeder Gruppe', async () => {
    zeige({ alleEinstellungen })

    await suche('sprache')

    const zeile = screen.getByRole('tab', { name: eigene.gruppen.allgemein })

    expect(within(zeile).getByText('1')).toBeInTheDocument()
  })

  it('zählt nichts, wo nichts liegt', async () => {
    zeige({ alleEinstellungen })

    await suche('proxy')

    const zeile = screen.getByRole('tab', { name: eigene.gruppen.daten })

    expect(within(zeile).queryByText('1')).not.toBeInTheDocument()
  })

  // **Kein Treffer nirgends bleibt ein Satz** und keine leere Fläche.
  it('sagt es, wenn nichts passt', async () => {
    zeige({ alleEinstellungen })

    await suche('nilpferd')

    expect(
      within(feld()).getByText(
        einsetzen(eigene.sucheKeine, { frage: 'nilpferd' }),
      ),
    ).toBeInTheDocument()

    // **Mit dem Weg zurück** (G02, #2249).
    await userEvent.click(
      within(feld()).getByRole('button', {
        name: kataloge.de.bausteine.listenansicht.sucheLeeren,
      }),
    )

    expect(screen.getByLabelText(eigene.suche)).toHaveValue('')
    expect(
      within(feld()).queryByText(
        einsetzen(eigene.sucheKeine, { frage: 'nilpferd' }),
      ),
    ).toBeNull()
  })

  // **Die Eingabetaste springt zum Feld** — „führt direkt zum Feld", ohne
  // bei jedem Anschlag den Fokus zu stehlen.
  it('springt mit der Eingabetaste auf das erste getroffene Feld', async () => {
    zeige({ alleEinstellungen })

    await suche('proxy{Enter}')

    expect(
      feld().querySelector('[data-einstellung="netz.proxy.http"]'),
    ).toHaveFocus()
  })

  // **Eine Segmentreihe nimmt den Fokus nicht selbst** — UXR-F2
  // (#2008):
  // ihr Schlüssel steht an der Gruppe, und eine Gruppe ist keine Tabstelle.
  // Ohne den Weg über das Segment führte „führt direkt zum Feld" bei jeder
  // Einstellung mit drei oder vier Werten ins Leere.
  it('springt in einer Segmentreihe auf das gewählte Segment', async () => {
    zeige({ alleEinstellungen })

    await suche(`${eigene.namen.erscheinungsbild}{Enter}`)

    expect(
      await screen.findByRole('radio', {
        name: eigene.werte['erscheinungsbild.system'],
      }),
    ).toHaveFocus()
  })

  // **Wo keines gewählt ist, das erste** — das ist die eine Tabstelle einer
  // Reihe ohne Wahl. Ein Wert, den die Liste nicht kennt, kann aus einem
  // Bestand von morgen kommen.
  it('springt auf das erste Segment, wo keines gewählt ist', async () => {
    zeige({
      alleEinstellungen: [
        ...eintraege.map((eintrag) =>
          eintrag.schluessel === 'erscheinungsbild'
            ? { ...eintrag, wert: 'unbekannt', wirksam: 'unbekannt' }
            : eintrag,
        ),
        proxy,
      ],
    })

    await suche(`${eigene.namen.erscheinungsbild}{Enter}`)

    expect(
      await screen.findByRole('radio', {
        name: eigene.werte['erscheinungsbild.system'],
      }),
    ).toHaveFocus()
  })

  // **Der Steuerlauf** (§10): ohne ihn sagte der Lauf darüber nicht, ob der
  // Fokus vorher überhaupt woanders war.
  it('lässt den Fokus beim Tippen im Suchfeld', async () => {
    zeige({ alleEinstellungen })

    await suche('proxy')

    expect(screen.getByLabelText(eigene.suche)).toHaveFocus()
  })
})
