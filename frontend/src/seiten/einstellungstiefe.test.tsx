// **Was eine Gruppe der Einstellungen eingeklappt trägt** — #2465
// (Issue (#2465)),
// Folge von AP-11.
//
// **Der Anlass ist gemessen**: `einstellungshoehen09` zerlegt jede Gruppe
// über R4 am nativen Fenster bis zum Bauteil. In *Dienste* trug der Katalog
// der Dienste 2.314 Punkte und die Zugangsdaten 890, in *Daten und
// Sicherung* das Zurückgehen, Aufräumen und Zurücksetzen, die Kosten und das
// Mitnehmen zusammen 1.466; in *KI* die Modelle, in *Erscheinungsbild* die
// eigenen Farben, in *Diagnose* die Zahlen seit dem Start.
//
// **Hier steht, dass sie eingeklappt stehen** — zu, nicht gezeichnet, ohne
// Ruf an Go, und der Name einmal: am Knopf. **Ob die Gruppen damit R4
// halten**, misst `ap11ruhe09` am Fenster; jsdom kennt keine Höhen.
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { abschnitteJeGruppe } from './einstellungsabschnitte'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { zeigeEinstellungen, zumGebiet } from '../test/einstellungsseite'

import type { Einstellungsgruppe as Gruppe } from './einstellungsgruppen'

const texte = kataloge.de
const eigene = texte.seiten.einstellungen

afterEach(() => {
  raeumeApi()
})

const knopf = (name: string) => screen.getByRole('button', { name })

// Fall ist ein Abschnitt, den #2465 einklappt.
interface Fall {
  gruppe: Gruppe
  titel: string

  // fragt ist die Bindung, die sein Inhalt beim Zeichnen ruft — erst, wenn
  // jemand ihn aufklappt.
  fragt?:
    'Dienstkatalog' | 'Zugaenge' | 'Kostenstand' | 'KIModelle' | 'Metriken'

  // darin ist ein Bedienelement, das nur sein Inhalt trägt.
  darin?: () => HTMLElement
}

const faelle: Fall[] = [
  { gruppe: 'dienste', titel: eigene.katalog, fragt: 'Dienstkatalog' },
  { gruppe: 'dienste', titel: eigene.zugaenge, fragt: 'Zugaenge' },
  {
    gruppe: 'daten',
    titel: eigene.abschnitte.eingriffe,
    darin: () =>
      screen.getByRole('heading', { name: texte.seiten.ruecksprung.titel }),
  },
  { gruppe: 'daten', titel: texte.seiten.kosten.titel, fragt: 'Kostenstand' },
  {
    gruppe: 'daten',
    titel: eigene.abschnitte.mitnehmen,
    darin: () =>
      screen.getByRole('checkbox', { name: eigene.austausch.mitArbeitsplatz }),
  },
  { gruppe: 'ki', titel: eigene.kimodelleTitel, fragt: 'KIModelle' },
  { gruppe: 'erscheinungsbild', titel: eigene.gestaltung.merkmale },
  {
    gruppe: 'diagnose',
    titel: texte.seiten.technik.metriken,
    fragt: 'Metriken',
  },
]

describe('die Abschnitte aus #2465', () => {
  it.each(faelle)(
    '„$titel“ steht in $gruppe zu und geht auf einen Klick auf',
    async ({ gruppe, titel, fragt, darin }) => {
      // **Eine Antwort, die nie kommt** — gefragt wird, ob gerufen wird,
      // nicht was dann dasteht.
      const ruf = vi.fn(() => new Promise<never>(() => undefined))
      belegeApi(fragt === undefined ? {} : { [fragt]: ruf })
      zeigeEinstellungen()
      await zumGebiet(eigene.gruppen[gruppe])

      expect(knopf(titel)).toHaveAttribute('aria-expanded', 'false')
      expect(ruf).not.toHaveBeenCalled()

      await userEvent.click(knopf(titel))

      expect(knopf(titel)).toHaveAttribute('aria-expanded', 'true')
      if (fragt !== undefined) expect(ruf).toHaveBeenCalledTimes(1)
      if (darin !== undefined) expect(darin()).toBeVisible()

      // **Der Name steht einmal** — am Knopf, der zugleich die Überschrift
      // des Kastens ist.
      expect(screen.queryByRole('heading', { name: titel })).toBeNull()
    },
  )

  // **Die Suche hält den Kasten offen** — der Schlüssel `gestaltung` steht
  // im Abschnitt, auch wenn ihn kein gewöhnliches Feld zeichnet.
  it('öffnet die eigenen Farben, wenn die Suche sie trifft', async () => {
    zeigeEinstellungen({
      alleEinstellungen: [
        {
          schluessel: 'gestaltung',
          wert: '',
          wirksam: '',
          werte: [],
          bereich: 'anwendung',
        },
      ],
    })
    await zumGebiet(eigene.gruppen.erscheinungsbild)

    await userEvent.type(
      screen.getByRole('searchbox', { name: eigene.suche }),
      eigene.gestaltung.merkmale,
    )

    expect(knopf(eigene.gestaltung.merkmale)).toHaveAttribute(
      'aria-expanded',
      'true',
    )
  })

  it('lässt offen, wer gefragt wird, und was den Bestand sichert', async () => {
    zeigeEinstellungen()

    await zumGebiet(eigene.gruppen.dienste)
    expect(screen.getByRole('heading', { name: eigene.dienste })).toBeVisible()

    await zumGebiet(eigene.gruppen.daten)
    expect(
      screen.getByRole('heading', {
        name: texte.seiten.bestandspflege.platzUeberschrift,
      }),
    ).toBeVisible()
  })

  it('stellt sie je Gruppe in die Reihenfolge des Zeichnens', () => {
    const namen = (gruppe: Gruppe) =>
      (abschnitteJeGruppe[gruppe] ?? []).map((plan) => plan.name)

    expect(namen('dienste')).toEqual(['dienstkatalog', 'zugaenge', 'handel'])
    expect(namen('daten')).toEqual([
      'eingriffe',
      'kosten',
      'mitnehmen',
      'ordner',
    ])
    expect(namen('ki')).toEqual(['kimodelle', 'adressen'])
    expect(namen('erscheinungsbild')).toEqual([
      'eigeneFarben',
      'fenster',
      'spalten',
    ])
    expect(namen('diagnose')).toEqual([
      'wirkung',
      'protokoll',
      'messung',
      'speicher',
    ])
  })
})
