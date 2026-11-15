// **Ein Modell schlägt Felder vor** — Werkstück A7 von Phase 5.3
// (#1446).
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import {
  aufklappen,
  ausfuellenUndVorschlagen,
  einKatalog,
  einVorschlag,
  vorschlagstexte as eigene,
  zeigeVorschlagskasten as zeige,
} from '../test/kivorschlag'

import type { KIUebernahmeAntwort } from '../antworten'

afterEach(() => {
  raeumeApi()
})

// **Der Satz steht da, bevor irgendetwas geschieht.**
//
// **§6**: die Oberfläche erklärt den nächsten Schritt, statt ihn vorauszusetzen.
it('sagt, dass ein Vorschlag die Karte nicht ändert', async () => {
  belegeApi({ KIModelle: () => Promise.resolve(einKatalog()) })
  zeige()

  await aufklappen(userEvent.setup())

  expect(screen.getByText(eigene.kiVorschlagAendertNichts)).toBeInTheDocument()
})

// **Ohne Modell und Auftrag geht der Knopf nicht.**
it('lässt ohne Modell und Auftrag nicht vorschlagen', async () => {
  belegeApi({ KIModelle: () => Promise.resolve(einKatalog()) })
  zeige()

  await aufklappen(userEvent.setup())

  expect(
    screen.getByRole('button', { name: eigene.kiVorschlagen }),
  ).toBeDisabled()
})

// **Der Vorschlag kommt an — und schreibt nichts.**
it('zeigt die vorgeschlagenen Felder, ohne zu schreiben', async () => {
  const uebernahme = vi.fn()

  belegeApi({
    KIModelle: () =>
      Promise.resolve(einKatalog({ anbieter: 'ollama', alias: 'gemma4:8b' })),
    KIVorschlagen: () => Promise.resolve(einVorschlag()),
    KIVorschlagUebernehmen: uebernahme,
  })

  const geaendert = zeige()

  await ausfuellenUndVorschlagen()

  expect(await screen.findByText('name')).toBeInTheDocument()
  expect(screen.getByText('Blitzschlag')).toBeInTheDocument()

  // **Die Gegenprobe zum Satz darüber**: nichts wurde übernommen, und der
  // Editor hat nichts nachzuladen.
  expect(uebernahme).not.toHaveBeenCalled()
  expect(geaendert).not.toHaveBeenCalled()
})

// **Erst die Übernahme schreibt — und meldet es.**
it('übernimmt auf Verlangen und meldet die Zahl', async () => {
  belegeApi({
    KIModelle: () =>
      Promise.resolve(einKatalog({ anbieter: 'ollama', alias: 'gemma4:8b' })),
    KIVorschlagen: () => Promise.resolve(einVorschlag()),
    KIVorschlagUebernehmen: () => Promise.resolve({ felder: 1, fehler: null }),
  })

  const geaendert = zeige()

  const hand = await ausfuellenUndVorschlagen()

  await hand.click(
    await screen.findByRole('button', { name: eigene.kiVorschlagUebernehmen }),
  )

  expect(
    await screen.findByText('1 Felder wurden geschrieben.'),
  ).toBeInTheDocument()

  // **Geschrieben ist geschrieben** (§7, Regel 3): der Editor lädt nach.
  await waitFor(() => {
    expect(geaendert).toHaveBeenCalled()
  })
})

// **Ein Fehler steht am Ort und nimmt den Vorschlag weg.**
it('zeigt einen Fehler und behält keinen halben Stand', async () => {
  belegeApi({
    KIModelle: () =>
      Promise.resolve(einKatalog({ anbieter: 'ollama', alias: 'gemma4:8b' })),
    KIVorschlagen: () =>
      Promise.resolve(
        einVorschlag({
          fehler: { code: 'validierung', meldung: 'keine Laufzeit' },
        }),
      ),
  })

  zeige()

  await ausfuellenUndVorschlagen()

  expect(await screen.findByRole('alert')).toHaveTextContent('keine Laufzeit')
  // **Mit dem Weg zur Einrichtung** (AI-01.AK-04, #1742).
  expect(
    screen.getByRole('button', {
      name: kataloge.de.bausteine.kieinrichtung.oeffnen,
    }),
  ).toBeInTheDocument()
  expect(screen.queryByText('Blitzschlag')).not.toBeInTheDocument()
})

// **Was das Schema nicht durchgelassen hat, steht daneben.**
it('nennt die Befunde des Schemas', async () => {
  belegeApi({
    KIModelle: () =>
      Promise.resolve(einKatalog({ anbieter: 'ollama', alias: 'gemma4:8b' })),
    KIVorschlagen: () =>
      Promise.resolve(
        einVorschlag({ befunde: ['rarity: »episch« ist keine Wahl'] }),
      ),
  })

  zeige()

  await ausfuellenUndVorschlagen()

  expect(
    await screen.findByText('rarity: »episch« ist keine Wahl'),
  ).toBeInTheDocument()
})

// **Ein Vorschlag ohne brauchbares Feld lässt sich nicht übernehmen.**
it('lässt einen leeren Vorschlag nicht übernehmen', async () => {
  belegeApi({
    KIModelle: () =>
      Promise.resolve(einKatalog({ anbieter: 'ollama', alias: 'gemma4:8b' })),
    KIVorschlagen: () => Promise.resolve(einVorschlag({ felder: [] })),
  })

  zeige()

  await ausfuellenUndVorschlagen()

  expect(await screen.findByText(eigene.kiVorschlagNichts)).toBeInTheDocument()
  expect(
    screen.getByRole('button', { name: eigene.kiVorschlagUebernehmen }),
  ).toBeDisabled()
})

// **Abbrechen räumt den Vorschlag weg, ohne zu schreiben.**
it('verwirft einen Vorschlag auf Verlangen', async () => {
  const uebernahme = vi.fn()

  belegeApi({
    KIModelle: () =>
      Promise.resolve(einKatalog({ anbieter: 'ollama', alias: 'gemma4:8b' })),
    KIVorschlagen: () => Promise.resolve(einVorschlag()),
    KIVorschlagUebernehmen: uebernahme,
  })

  zeige()

  const hand = await ausfuellenUndVorschlagen()

  await hand.click(
    await screen.findByRole('button', { name: eigene.kiVorschlagAbbrechen }),
  )

  await waitFor(() => {
    expect(screen.queryByText('Blitzschlag')).not.toBeInTheDocument()
  })

  expect(uebernahme).not.toHaveBeenCalled()
})

// **Ein Fehler beim Übernehmen steht am Ort — und der Vorschlag bleibt.**
//
// **§7, Regel 3, von der anderen Seite**: nach einem **gescheiterten**
// Schreiben wird nicht nachgeladen, und was nicht geschrieben wurde, bleibt
// zum Ansehen stehen.
it('zeigt einen Fehler beim Übernehmen und behält den Vorschlag', async () => {
  belegeApi({
    KIModelle: () =>
      Promise.resolve(einKatalog({ anbieter: 'ollama', alias: 'gemma4:8b' })),
    KIVorschlagen: () => Promise.resolve(einVorschlag()),
    KIVorschlagUebernehmen: () =>
      Promise.resolve<KIUebernahmeAntwort>({
        felder: 0,
        fehler: {
          code: 'konflikt',
          schluessel: 'konflikt',
          meldung: 'die Karte hat sich geändert',
        },
      }),
  })

  const geaendert = zeige()

  const hand = await ausfuellenUndVorschlagen()

  await hand.click(
    await screen.findByRole('button', { name: eigene.kiVorschlagUebernehmen }),
  )

  // **Der Satz kommt aus dem Katalog und nicht aus Go** (§9): übersetzt wird
  // über den Schlüssel, und den kennt das Fenster.
  expect(await screen.findByRole('alert')).toHaveTextContent(
    kataloge.de.fehler.konflikt,
  )

  expect(screen.getByText('Blitzschlag')).toBeInTheDocument()
  expect(geaendert).not.toHaveBeenCalled()
})

// **Ein Kasten, der zugeht, fragt nicht noch einmal.**
//
// **Und der Zweig ist der, den `onToggle` beim Schließen nimmt**: `offen`
// fällt zurück, `gefragt` bleibt stehen. Ohne diesen Lauf wäre nur der halbe
// Umschalter betreten — und niemand sähe, dass der Inhalt verschwindet.
it('räumt seinen Inhalt ab, wenn er zugeht', async () => {
  const modelle = vi.fn(() =>
    Promise.resolve(einKatalog({ anbieter: 'ollama', alias: 'gemma4:8b' })),
  )

  belegeApi({ KIModelle: modelle })
  zeige()

  const hand = userEvent.setup()

  await aufklappen(hand)
  await screen.findByRole('option', { name: 'ollama/gemma4:8b' })

  await aufklappen(hand)

  expect(screen.queryByRole('combobox')).not.toBeInTheDocument()

  // **Der Beleg, dass jemand zuhört** (§3): ein zweites Aufklappen bringt
  // die Wahl zurück — ohne einen zweiten Ruf.
  await aufklappen(hand)

  expect(
    await screen.findByRole('option', { name: 'ollama/gemma4:8b' }),
  ).toBeInTheDocument()

  expect(modelle).toHaveBeenCalledTimes(1)
})

// **Ohne Modell steht da, warum** — und nicht eine leere Wahl.
it('sagt, dass kein Modell bereitsteht', async () => {
  belegeApi({ KIModelle: () => Promise.resolve(einKatalog()) })
  zeige()

  await aufklappen(userEvent.setup())

  expect(
    await screen.findByText(eigene.kiVorschlagOhneModell),
  ).toBeInTheDocument()
  // **Mit dem Weg dorthin** (AI-01.AK-04, #1742) — der Satz sagt, was fehlt.
  expect(
    screen.getByRole('button', {
      name: kataloge.de.bausteine.kieinrichtung.oeffnen,
    }),
  ).toBeInTheDocument()
})

// **Ein Fehler beim Lesen des Katalogs steht da und hält nichts auf.**
it('meldet einen Fehler beim Lesen des Katalogs', async () => {
  belegeApi({
    KIModelle: () => Promise.reject(new Error('die Laufzeit antwortet nicht')),
  })
  zeige()

  await aufklappen(userEvent.setup())

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'die Laufzeit antwortet nicht',
  )
  expect(
    screen.getByRole('button', {
      name: kataloge.de.bausteine.kieinrichtung.oeffnen,
    }),
  ).toBeInTheDocument()

  // **Und der Satz »kein Modell« steht dann nicht daneben**: zwei Auskünfte
  // über dieselbe Leere sind eine zu viel.
  expect(
    screen.queryByText(eigene.kiVorschlagOhneModell),
  ).not.toBeInTheDocument()
})
