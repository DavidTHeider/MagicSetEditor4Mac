// **Jede Einstellung hat auf der Seite das Bedienelement ihrer Werte** —
// UXR-F2 (#2008).
//
// # Warum noch ein Lauf neben `wahlkatalog.test.tsx`
//
// **Der Lauf dort zeichnet `Wahl` allein.** Auf der Seite zeichnen aber auch
// die Kartendienste, die Fragen, die KI-Laufzeit und die Bereichsschalter —
// jedes auf seinem Weg, und jeder könnte an der Regel vorbei eine Auswahlliste
// bauen. Hier geht die ganze Seite durch: jede Gruppe, alles aufgeklappt, und
// jeder Knoten mit einem Schlüssel wird gegen die Regel gehalten.
//
// # Was er hält
//
// - **Jede Einstellung des Katalogs mit Werteliste steht irgendwo da** — und
//   jede mit dem Element, das ihre Werte verlangen (`erwartetesElement`).
// - **Keine steht zweimal da** — gezählt am Schlüssel, über alle Gruppen.
//
// # Was er nicht hält
//
// **Die Einstellungen, die ein Beiwerk zeichnet** (die Belegung der Ziffern,
// die eigenen Farben, die eigene Übersetzung): sie haben keine Werteliste.
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterAll, afterEach, describe, expect, it } from 'vitest'

import { zeigeEinstellungen } from '../test/einstellungsseite'
import {
  alsEintrag,
  einstellungskatalog,
  elementVon,
  erwartetesElement,
} from '../test/einstellungskatalog'
import { einstellungsgruppen } from './einstellungsgruppen'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'

afterEach(() => {
  raeumeApi()
})

// zeilen sind die Einstellungen der Seite — die Tafel des Profils und des
// Geräts.
const zeilen = einstellungskatalog().filter((zeile) => zeile.tafel === 'profil')

// **Die Dienste an**, damit die Frage nach der Suche alle vier zur Wahl
// stellt: sie bietet nur an, was eingeschaltet ist (`Dienstfragen`).
const alle = zeilen.map((zeile) =>
  alsEintrag(
    zeile,
    zeile.schluessel.startsWith('dienst.') && zeile.werte.length > 0
      ? 'an'
      : undefined,
  ),
)

const mitVorsatz = (vorsatz: string) =>
  alle.filter((eintrag) => eintrag.schluessel.startsWith(vorsatz))

// gesehen sammelt über alle Gruppen, welcher Schlüssel wie oft dastand.
const gesehen = new Map<string, number>()

describe('die Einstellungsseite, Gruppe für Gruppe', () => {
  it.each(einstellungsgruppen)(
    '%s zeichnet jede Einstellung mit dem Element ihrer Werte',
    async (gruppe) => {
      belegeApi({})
      zeigeEinstellungen({
        alleEinstellungen: alle,
        dienste: mitVorsatz('dienst.'),
        fragen: mitVorsatz('frage.'),
        ki: mitVorsatz('ki.'),
      })

      await userEvent.click(
        screen.getByRole('tab', {
          name: kataloge.de.seiten.einstellungen.gruppen[gruppe],
        }),
      )

      // **Alles aufgeklappt** — was hinter einem Abschnitt steht, steht dort
      // auch, und soll dieselbe Regel halten.
      for (const knopf of within(screen.getByRole('tabpanel')).queryAllByRole(
        'button',
        { expanded: false },
      )) {
        await userEvent.click(knopf)
      }

      for (const knoten of screen
        .getByRole('tabpanel')
        .querySelectorAll('[data-einstellung]')) {
        const schluessel = knoten.getAttribute('data-einstellung') ?? ''
        const zeile = zeilen.find(
          (kandidat) => kandidat.schluessel === schluessel,
        )

        gesehen.set(schluessel, (gesehen.get(schluessel) ?? 0) + 1)

        // Ein Beiwerk, das keine Werteliste kennt, ist nicht Sache dieses
        // Laufs.
        if (zeile?.art !== 'liste') {
          continue
        }

        expect(elementVon(knoten), `${schluessel} in ${gruppe}`).toBe(
          erwartetesElement(zeile.werte, zeile.schluessel),
        )
      }
    },
  )

  // **Zuletzt, weil er die Gruppen davor braucht**: Vitest führt die Läufe
  // einer Datei in der Reihe aus, in der sie stehen.
  it('zeigt jede Einstellung mit Werteliste, und keine zweimal', () => {
    const mitListe = zeilen.filter((zeile) => zeile.art === 'liste')

    // Eine Gegenprobe, dass die Läufe davor etwas gesehen haben.
    expect(gesehen.size).toBeGreaterThan(40)

    expect(
      mitListe
        .filter((zeile) => !gesehen.has(zeile.schluessel))
        .map((zeile) => zeile.schluessel),
      'diese Einstellungen zeichnete keine Gruppe',
    ).toEqual([])

    expect(
      [...gesehen]
        .filter(([, zahl]) => zahl > 1)
        .map(([schluessel]) => schluessel),
      'diese Einstellungen standen mehrfach da',
    ).toEqual([])
  })
})

afterAll(() => {
  gesehen.clear()
})
