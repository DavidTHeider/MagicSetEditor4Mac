// **Abschnitte in einer Gruppe der Einstellungen** — UXR-19
// (#1927).
//
// **Der Schnitt folgt der Frage** (§2): `formularwand.test.tsx` zählt, ob
// eine Gruppe die Grenze hält; hier steht, was ein Abschnitt **tut** — er ist
// zu, er geht auf, die Suche hält ihn offen, und kein Schlüssel steht in
// zweien.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { eintraege, zeigeEinstellungen } from '../test/einstellungsseite'
import { abschnitteJeGruppe } from './einstellungsabschnitte'
import { Einstellungsgruppe } from './einstellungsgebiete'
import { Sprachraum } from '../sprache/kontext'
import { fachschluessel, gruppeFuer } from './einstellungsgruppen'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { sprachpaketprobe } from '../test/sprachprobe'

import type { EinstellungDTO } from '../api'

const eigene = kataloge.de.seiten.einstellungen

afterEach(() => {
  raeumeApi()
})

function zeile(schluessel: string, wert = ''): EinstellungDTO {
  return {
    schluessel,
    wert,
    wirksam: wert,
    werte: [],
    bereich: 'anwendung',
    ebene: 'nutzer',
    herkunft: 'nutzer',
    vorgabe: '',
    wirkung: 'sofort',
  }
}

function zeige() {
  belegeApi({
    Sprachpakete: () =>
      Promise.resolve({ pakete: [sprachpaketprobe()], fehler: null }),
  })

  return zeigeEinstellungen({
    alleEinstellungen: [
      ...eintraege,
      zeile('kuerzel.ziffern', '3=karten'),
      zeile('sprache.eigene'),
    ],
  })
}

const kuerzelKnopf = () =>
  screen.getByRole('button', { name: eigene.kuerzelUeberschrift })

const uebersetzungsKnopf = () =>
  screen.getByRole('button', { name: eigene.eigenuebersetzung.titel })

describe('Allgemein', () => {
  it('zeigt die Sprache offen und Übersetzungen und Kürzel eingeklappt', () => {
    zeige()

    // **Zwei Werte sind zwei Segmente** (UXR-F2, #2008) — die Vorrichtung
    // führt `de` und `en`.
    expect(
      screen.getByRole('radiogroup', { name: eigene.namen.sprache }),
    ).toBeInTheDocument()

    expect(kuerzelKnopf()).toHaveAttribute('aria-expanded', 'false')
    expect(uebersetzungsKnopf()).toHaveAttribute('aria-expanded', 'false')

    // **Zu heißt: nicht gezeichnet** — weder die Tafel noch die zehn Wahlen
    // der Belegung, noch die Neuanlage einer Übersetzung.
    expect(screen.queryByText('⌘K')).toBeNull()
    expect(screen.queryByRole('combobox', { name: 'Seite für ⌘3' })).toBeNull()
    expect(
      screen.queryByRole('combobox', {
        name: eigene.eigenuebersetzung.verwenden,
      }),
    ).toBeNull()
  })

  it('öffnet die Kürzel auf einen Klick, mit Tafel und Belegung', async () => {
    zeige()

    await userEvent.click(kuerzelKnopf())

    expect(kuerzelKnopf()).toHaveAttribute('aria-expanded', 'true')
    expect(screen.getByText('⌘K')).toBeVisible()
    expect(screen.getByRole('combobox', { name: 'Seite für ⌘3' })).toHaveValue(
      'karten',
    )

    // **Der Name steht einmal**: am Knopf, der zugleich die Überschrift des
    // Kastens ist — eine zweite gleichlautende darunter wäre einer zu viel.
    expect(
      screen.queryByRole('heading', { name: eigene.kuerzelUeberschrift }),
    ).toBeNull()
  })

  it('öffnet die eigenen Übersetzungen auf einen Klick, ohne zweiten Namen', async () => {
    zeige()

    await userEvent.click(uebersetzungsKnopf())

    expect(
      await screen.findByRole('combobox', {
        name: eigene.eigenuebersetzung.verwenden,
      }),
    ).toBeVisible()
    expect(
      screen.queryByRole('heading', { name: eigene.eigenuebersetzung.titel }),
    ).toBeNull()
  })

  // **Die Suche hält den Kasten offen** — sonst führte sie zu einem Feld, das
  // gar nicht gezeichnet ist (UX-04.AK-03).
  it('hält offen, wo die Suche einen Treffer hat', async () => {
    zeige()

    await userEvent.type(
      screen.getByRole('searchbox', { name: eigene.suche }),
      'Belegung',
    )

    expect(kuerzelKnopf()).toHaveAttribute('aria-expanded', 'true')
    expect(screen.getByRole('combobox', { name: 'Seite für ⌘3' })).toBeVisible()
    expect(uebersetzungsKnopf()).toHaveAttribute('aria-expanded', 'false')
  })
})

// **Das Bauteil selbst**, mit einem Abschnitt, den es nur hier gibt: einer
// mit Einstellungen und ohne Beiwerk, und einer, der leer bliebe.
describe('ein Abschnitt in der Gruppe', () => {
  function zeigeGruppe() {
    render(
      <Sprachraum sprache="de">
        <Einstellungsgruppe
          name={eigene.gruppen.allgemein}
          eintraege={[zeile('sprache', 'de'), zeile('zeitzone', 'UTC')]}
          laeuft={() => false}
          setze={vi.fn()}
          abschnitte={[
            { name: 'kuerzel', schluessel: ['zeitzone'] },
            { name: 'eigeneUebersetzungen', schluessel: ['sprache.eigene'] },
          ]}
        />
      </Sprachraum>,
    )
  }

  it('zeichnet seine Einstellungen erst, wenn er aufgeht', async () => {
    zeigeGruppe()

    expect(
      screen.getByRole('textbox', { name: eigene.namen.sprache }),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('combobox', { name: eigene.namen.zeitzone }),
    ).toBeNull()

    await userEvent.click(kuerzelKnopf())

    // **Die Zeitzone ist seit UXR-F1 (#2007) eine Auswahl mit Suche**, kein
    // Textfeld mehr.
    expect(
      screen.getByRole('combobox', { name: eigene.namen.zeitzone }),
    ).toBeVisible()
  })

  // **Ein Knopf, der nichts aufklappt, verspräche etwas, was nicht kommt.**
  it('steht nicht da, wenn er nichts trüge', () => {
    zeigeGruppe()

    expect(
      screen.queryByRole('button', { name: eigene.eigenuebersetzung.titel }),
    ).toBeNull()
  })
})

describe('die Tafel der Abschnitte', () => {
  const plaene = Object.entries(abschnitteJeGruppe).flatMap(([gruppe, liste]) =>
    liste.map((plan) => ({ gruppe, plan })),
  )

  it('führt nur Schlüssel der Gruppe, in der der Abschnitt steht', () => {
    const fremde = plaene.flatMap(({ gruppe, plan }) =>
      plan.schluessel.filter((schluessel) => gruppeFuer(schluessel) !== gruppe),
    )

    expect(fremde).toEqual([])
  })

  it('führt keinen Schlüssel zweimal und keinen aus dem Kasten für Fachleute', () => {
    const alle = plaene.flatMap(({ plan }) => plan.schluessel)

    expect(new Set(alle).size).toBe(alle.length)
    expect(alle.filter((schluessel) => fachschluessel.has(schluessel))).toEqual(
      [],
    )
  })
})

// **Die Schnittstellen stehen unter „Verbindungen“** — AP-10 (#2367),
// Befund B-19: die Ein-Klick-Einrichtung von `mse`, `mse-mcp` und
// `mse-rest` stand hinter Protokollen und Messung unter „Diagnose“, und
// „Verbindungen“, wo der Dienst gestartet wird, nannte sie nicht.
describe('die Werkzeuge ohne Fenster', () => {
  it('stehen als erster Abschnitt unter „Verbindungen“ und nicht unter „Diagnose“', () => {
    const namen = (gruppe: keyof typeof abschnitteJeGruppe) =>
      (abschnitteJeGruppe[gruppe] ?? []).map((plan) => plan.name)

    expect(namen('verbindungen')[0]).toBe('werkzeuge')
    expect(namen('diagnose')).not.toContain('werkzeuge')
    expect(abschnitteJeGruppe.verbindungen?.[0]?.oben).toBe(true)
  })

  // **Oben heißt vor dem Beiwerk** — sonst stünde der Abschnitt unter
  // Identität, Gegenübern und Horchposten, und „oben unter Verbindungen“
  // (B-19) wäre nur in der Tafel wahr.
  it('steht ein Abschnitt mit „oben“ vor dem Beiwerk der Gruppe', () => {
    render(
      <Sprachraum sprache="de">
        <Einstellungsgruppe
          name={eigene.gruppen.verbindungen}
          eintraege={[]}
          laeuft={() => false}
          setze={vi.fn()}
          abschnitte={[
            {
              name: 'werkzeuge',
              schluessel: [],
              oben: true,
              beiwerk: <p>Werkzeuge darin</p>,
            },
          ]}
        >
          <p>Beiwerk der Gruppe</p>
        </Einstellungsgruppe>
      </Sprachraum>,
    )

    const knopf = screen.getByRole('button', {
      name: kataloge.de.seiten.technik.ueberschrift,
    })
    const beiwerk = screen.getByText('Beiwerk der Gruppe')

    expect(
      knopf.compareDocumentPosition(beiwerk) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy()
  })
})
