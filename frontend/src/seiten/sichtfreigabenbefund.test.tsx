// **Ein Fehler an einer Sicht steht am Feld** — UXR-F18
// (#2033).
//
// # Was hier gemessen wird
//
// **§6 und §8**: die fehlende Auswahl an der Wahl der Drucke, eine
// abgewiesene Behaltemenge an ihrem Zahlenfeld — und nicht als anonymer Satz
// über dem Kasten. Die Domäne nennt das Feld (`auswahl`, `behalten`); das
// Fenster hängt die Meldung dorthin.
//
// **Eine eigene Datei**, weil `sichtfreigabenumfang.test.tsx` mit diesen
// Läufen über die Zeilengrenze aus §2 ginge.
import { screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { antwortAbwarten, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import {
  aufstellen,
  erteilen,
  freigabenliste,
  umfangWaehlen,
  worte,
} from '../test/sichtfreigabenprobe'

afterEach(() => {
  raeumeApi()
})

describe('Ein Fehler steht am Feld', () => {
  it('sagt an der Wahl, dass eine Auswahl ohne Drucke nichts freigibt — und ruft nichts', async () => {
    const erteilt = vi.fn(() => Promise.resolve(freigabenliste()))
    const hand = await aufstellen({ SammlungFreigeben: erteilt })

    await screen.findByRole('group', { name: worte.drucke })
    await hand.click(erteilen())
    await antwortAbwarten()

    // **Erst dass niemand zuhört, dann die Verneinung** (§3).
    expect(erteilt).not.toHaveBeenCalled()

    const gruppe = screen.getByRole('group', { name: worte.drucke })
    expect(within(gruppe).getByRole('alert')).toHaveTextContent(worte.fehlt)
    expect(screen.getAllByRole('alert')).toHaveLength(1)
  })

  it('nimmt den Befund zurück, sobald ein Druck gewählt ist', async () => {
    const hand = await aufstellen()

    await screen.findByRole('group', { name: worte.drucke })
    await hand.click(erteilen())
    expect(await screen.findByText(worte.fehlt)).toBeInTheDocument()

    await hand.click(screen.getByRole('checkbox', { name: /Blitzschlag/ }))

    expect(screen.queryByText(worte.fehlt)).toBeNull()
  })

  it('nimmt den Befund zurück, wenn der Umfang wechselt', async () => {
    const hand = await aufstellen()

    await screen.findByRole('group', { name: worte.drucke })
    await hand.click(erteilen())
    expect(await screen.findByText(worte.fehlt)).toBeInTheDocument()

    await umfangWaehlen(hand, 'sammlung')
    await umfangWaehlen(hand, 'auswahl')

    expect(screen.queryByText(worte.fehlt)).toBeNull()
  })

  // **Ein Schlüssel, den kein Katalog kennt** — dann steht die Meldung aus
  // Go da, und der Lauf misst, dass sie ankommt und wo.
  it('zeigt den Feldbefund des Dienstes an der Wahl und nicht über dem Kasten', async () => {
    const hand = await aufstellen({
      SammlungFreigeben: () =>
        Promise.resolve(
          freigabenliste([], {
            code: 'validierung' as const,
            schluessel: 'gibtesnicht',
            meldung: 'so nicht',
            feld: 'auswahl',
          }),
        ),
    })

    await hand.click(
      await screen.findByRole('checkbox', { name: /Blitzschlag/ }),
    )
    await hand.click(erteilen())

    const gruppe = screen.getByRole('group', { name: worte.drucke })
    expect(await within(gruppe).findByRole('alert')).toHaveTextContent(
      'so nicht',
    )
    expect(screen.getAllByRole('alert')).toHaveLength(1)
  })

  it('zeigt den Feldbefund des Dienstes an der Behaltemenge', async () => {
    const hand = await aufstellen({
      SammlungFreigeben: () =>
        Promise.resolve(
          freigabenliste([], {
            code: 'validierung' as const,
            schluessel: 'gibtesnicht',
            meldung: 'so nicht',
            feld: 'behalten',
          }),
        ),
    })

    await umfangWaehlen(hand, 'dubletten')
    await hand.click(erteilen())

    const feld = await screen.findByLabelText(worte.behalten)
    await screen.findByText('so nicht')

    expect(feld).toBeInvalid()
    expect(feld).toHaveAccessibleDescription(
      expect.stringContaining('so nicht'),
    )
    expect(screen.getAllByRole('alert')).toHaveLength(1)
  })

  it('hält „Sicht erteilen" an, solange die Behaltemenge nicht gilt', async () => {
    const erteilt = vi.fn(() => Promise.resolve(freigabenliste()))
    const hand = await aufstellen({ SammlungFreigeben: erteilt })

    await umfangWaehlen(hand, 'dubletten')

    const feld = screen.getByLabelText(worte.behalten)
    await hand.clear(feld)
    await hand.type(feld, '0')

    // **Der Satz am Feld und der Knopf angehalten** — nicht der letzte
    // gültige Wert im Stillen. **Und schon 0 gilt nicht**: die Domäne liest
    // 0 als die Vorgabe 1 (`MitVorgaben`), und ein Feld, das 0 annähme,
    // zeigte etwas anderes, als erteilt wird.
    expect(
      await screen.findByText(
        kataloge.de.allgemein.ganzeZahlAb.replace('{kleinstens}', '1'),
      ),
    ).toBeInTheDocument()
    expect(erteilen()).toBeDisabled()

    await hand.clear(feld)
    expect(erteilen()).toBeDisabled()

    await hand.type(feld, '2')
    expect(erteilen()).toBeEnabled()
  })

  // **Das Zahlenfeld steht beim Zurückkommen mit dem letzten gültigen Wert
  // da** — ein halb getippter Text überlebt den Wechsel nicht, und der Knopf
  // dürfte nicht an ihm hängenbleiben.
  it('lässt „Sicht erteilen" wieder zu, wenn der Umfang wechselt', async () => {
    const hand = await aufstellen()

    await umfangWaehlen(hand, 'dubletten')

    const feld = screen.getByLabelText(worte.behalten)
    await hand.clear(feld)
    await hand.type(feld, '-2')
    expect(erteilen()).toBeDisabled()

    await umfangWaehlen(hand, 'sammlung')
    expect(erteilen()).toBeEnabled()

    await umfangWaehlen(hand, 'dubletten')
    expect(screen.getByLabelText(worte.behalten)).toHaveValue(1)
    expect(erteilen()).toBeEnabled()
  })
})
