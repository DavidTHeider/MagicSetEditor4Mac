// **Keine Formularwand auf der Einstellungsseite** — UXR-19
// (#1927),
// Regel R3 aus `docs/UIREVIEW.md`: höchstens **zwölf Felder** gezeichnet auf
// einer Fläche, sonst gruppiert und eingeklappt bis auf eine Gruppe.
//
// # Der Anlass
//
// **Am nativen Fenster gemessen, 2026-10-02**: *Allgemein* zeichnete 21
// Felder. Gemessen war nur die Gruppe, die beim Öffnen offen steht — die
// übrigen zeichnen ihre Felder ebenso, nur sieht sie das Drehbuch nicht.
//
// # Was er hält
//
// **Jede Gruppe, gezeichnet mit jeder Einstellung, die sie führt** — die
// Liste kommt aus dem Katalog (`test/einstellungskatalog.ts`), mit den Werten,
// die Go ihr gibt: seit UXR-F2 (#2008)
// entscheidet die Zahl der Werte über das Bedienelement, und eine Gruppe mit
// lauter An/Aus zeichnete anderes als die wirkliche. Gezählt wird wie im
// Drehbuch `uirevue`: Eingaben, Auswahlen, Textfelder, Schalter; dazu das
// Suchfeld, das in derselben Fläche steht. **Was eingeklappt ist, zählt
// nicht** — ein `Aufklappkasten` zeichnet erst beim Öffnen.
//
// **Eine Segmentreihe zählt einmal.** Ihre Optionsfelder sind die Werte einer
// Frage, und die Frage ist es, die R3 zählt: vier Segmente für das Thema sind
// ein Feld, nicht vier. Optionsfelder außerhalb einer Segmentreihe — eines je
// Stil, eines je Karte — zählen weiter einzeln; deshalb zählt der Lauf die
// Gruppe (`role="radiogroup"`) und nicht jedes Optionsfeld.
//
// # Was er nicht hält
//
// **Die Rolltiefe (R4).** jsdom rechnet kein Layout; sie misst das Drehbuch am
// Fenster. **Und Felder, die ein Beiwerk erst nach einer Antwort zeichnet**,
// die dieser Lauf nicht belegt — eine Liste eigener Übersetzungen etwa. Er
// zählt, was ohne Antwort dasteht, und das ist die Untergrenze.
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { zeigeEinstellungen } from '../test/einstellungsseite'
import { profilEintraege } from '../test/einstellungskatalog'
import { einstellungsgruppen } from './einstellungsgruppen'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'

import type { Einstellungsgruppe } from './einstellungsgruppen'

afterEach(() => {
  raeumeApi()
})

// grenze ist R3 — das Suchfeld eingerechnet, denn es steht in derselben Fläche.
const grenze = 12

// alle sind jede Einstellung des Katalogs, mit ihren Werten.
const alle = profilEintraege()

const mitVorsatz = (vorsatz: string) =>
  alle.filter((eintrag) => eintrag.schluessel.startsWith(vorsatz))

// felder zählt, was in der offenen Gruppe gezeichnet ist — und das Suchfeld.
//
// **Die Optionsfelder einer Segmentreihe nicht einzeln**: die Reihe zählt, als
// `[role=radiogroup]`.
function felder(): number {
  const gezeichnet = new Set(
    [
      ...screen
        .getByRole('tabpanel')
        .querySelectorAll(
          'input:not([type=hidden]), select, textarea, [role=combobox], [role=switch], [role=radiogroup]',
        ),
    ].filter(
      (feld) =>
        !(feld.matches('input') && feld.closest('[role=radiogroup]') !== null),
    ),
  )

  return gezeichnet.size + 1
}

async function zaehle(gruppe: Einstellungsgruppe): Promise<number> {
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

  return felder()
}

// **Vor UXR-19 rissen fünf von zehn die Grenze** — gezählt mit diesem Lauf
// am 2026-10-02: *Allgemein* 21, *Erscheinungsbild* 19, *Dienste* 18 (vier
// davon gehörten zu den Verbindungen, `zeugnisort.test.tsx`), *Daten und
// Sicherung* 14, *Diagnose* 23. Sie fielen eine nach der anderen aus einer
// Liste der offenen; die Liste ist leer und deshalb fort.
describe('R3 auf der Einstellungsseite', () => {
  it.each(einstellungsgruppen)(
    '%s zeichnet höchstens zwölf Felder, solange niemand aufklappt',
    async (gruppe) => {
      expect(await zaehle(gruppe)).toBeLessThanOrEqual(grenze)
    },
  )
})

// **Eine Gegenprobe, dass der Lauf Segmentreihen sieht und sie einmal zählt**:
// *Erscheinungsbild* trägt vier davon (Thema, Farbwelt, Schriftbereich,
// Fensterbild). Ohne sie wäre die Zählweise oben ungeprüft.
describe('die Zählweise der Segmentreihen', () => {
  it('zählt eine Reihe als ein Feld und nicht als ihre Optionsfelder', async () => {
    const gezaehlt = await zaehle('erscheinungsbild')
    const flaeche = screen.getByRole('tabpanel')

    // Jedes Optionsfeld einzeln gezählt — so zählte `uirevue` Optionsfelder
    // außerhalb einer Reihe.
    const einzeln =
      new Set(
        flaeche.querySelectorAll(
          'input:not([type=hidden]), select, textarea, [role=combobox], [role=switch]',
        ),
      ).size + 1

    expect(
      flaeche.querySelectorAll('[role=radiogroup]').length,
    ).toBeGreaterThanOrEqual(4)
    expect(einzeln).toBeGreaterThan(gezaehlt)
  })
})
