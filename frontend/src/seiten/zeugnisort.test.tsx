// **Das eigene Zeugnis der Dienste steht bei den Verbindungen** — UXR-19
// (#1927).
//
// # Der Anlass
//
// **Gefunden beim Zählen für R3**: *Dienste* zeichnete mit allen
// Einstellungen 17 Felder, und vier davon waren `dienst.tls.*` — Zeugnis,
// Schlüssel, Kette und Anker des REST-Dienstes und des Werkzeugservers
// (NET-03, #1707). `einstellungsgruppen.ts` stellt sie zu *Verbindungen*, in
// den Kasten für Fachleute; gezeichnet wurden sie unter **Fremde
// Kartendienste**, zwischen Scryfall und CardTrader.
//
// **Der Weg dorthin**: die Seite bekommt alles, was `dienst.` heißt, als
// Schalter der Dienste, und schnitt daraus nur die KI heraus. Was nicht KI
// war, galt als Kartendienst — auch das Zeugnis.
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { eintraege, zeigeEinstellungen } from '../test/einstellungsseite'
import { kataloge } from '../sprache/kataloge'

import type { EinstellungDTO } from '../api'

const eigene = kataloge.de.seiten.einstellungen

function zeile(schluessel: string, werte: string[]): EinstellungDTO {
  return {
    schluessel,
    wert: werte[0] ?? '',
    wirksam: werte[0] ?? '',
    werte,
    bereich: 'dienste',
  }
}

const scryfall = zeile('dienst.scryfall', ['aus', 'an'])
const zeugnis = zeile('dienst.tls.zertifikat', [])

const gezeichnet = (schluessel: string) =>
  within(screen.getByRole('tabpanel')).queryAllByText(
    (_, element) => element?.getAttribute('data-einstellung') === schluessel,
  )

describe('das Zeugnis der Dienste', () => {
  it('steht nicht bei den Kartendiensten, sondern im Kasten der Verbindungen', async () => {
    zeigeEinstellungen({
      alleEinstellungen: [...eintraege, scryfall, zeugnis],
      dienste: [scryfall, zeugnis],
    })

    await userEvent.click(
      screen.getByRole('tab', { name: eigene.gruppen.dienste }),
    )

    // **Erst der Beleg, dass die Gruppe zeichnet** — sonst sagte das Nein
    // darunter nur, dass nichts dasteht (§3).
    expect(gezeichnet('dienst.scryfall')).toHaveLength(1)
    expect(gezeichnet('dienst.tls.zertifikat')).toHaveLength(0)

    await userEvent.click(
      screen.getByRole('tab', { name: eigene.gruppen.verbindungen }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: eigene.fachleute }),
    )

    expect(gezeichnet('dienst.tls.zertifikat')).toHaveLength(1)
  })
})
