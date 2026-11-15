// Das Speichern eines Sets als `.mse-set`.
//
// **Zwei Wege und ein Unterschied**: ⌘S geht dorthin zurück, wo das Set
// herkommt, „Sichern unter" fragt. Was die Oberfläche davon zu leisten hat, ist
// nicht der Ort — den kennt der Dienst —, sondern die **Wahl** zwischen beiden.
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { SetDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { useSetSpeichern } from '../haken/useSetSpeichern'
import { belegeApi, raeumeApi } from '../test/api'
import { waehleSethandlung } from '../test/sethandlung'
import { Sets } from './Sets'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function probeSet(teile: Partial<SetDTO> = {}): SetDTO {
  return {
    id: '4',
    name: 'Mein Probeset',
    spiel: 'magic',
    stil: 'm15',
    herkunft: '/pfad/probeset.mse-set',
    karten: 2,
    herkunftFehlt: false,
    ...teile,
  }
}

function speicherung(teile: Record<string, unknown> = {}) {
  return Promise.resolve({
    pfad: '/pfad/probeset.mse-set',
    befunde: [],
    abgebrochen: false,
    ...teile,
  })
}

// **Mit dem echten Haken.** Seit der Zweitblick über Phase 6 die zweite
// Vorgangsliste angestrichen hat, bekommt `Sets` den Speicherlauf von außen —
// und dieser Test prüft das Speichern selbst, nicht das Durchreichen. Also
// baut er ihn dort, wo `App` ihn baut: eine Ebene darüber.
function MitEchtemLauf() {
  const speichern = useSetSpeichern()

  return <Sets wechseln={vi.fn()} waehlen={vi.fn()} speichern={speichern} />
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <MitEchtemLauf />
    </Sprachraum>,
  )
}

// meldung wartet auf die Rückmeldung des Sicherns. **Über die Rolle und den
// Satz zusammen**: der Pfad steht auch in der Spalte „Herkunft" derselben
// Zeile, und seit UX-11 (#1683) zählt die Liste ebenfalls in einem `status`
// mit — die Rolle allein fände zwei.
function meldung(satz: string) {
  return screen.findByText(satz, { selector: '[role=status] > p' })
}

// druecke wählt im Menü der Setkarte — dort stehen Sichern und Sichern unter
// seit UXR-04 (#1912), vorher als zwei Knöpfe an jeder Karte.
async function druecke(eintrag: string) {
  await waehleSethandlung(texte, 'Mein Probeset', eintrag)
}

describe('Ein Set sichern', () => {
  it('schreibt das gewählte Set an seinen Herkunftsort', async () => {
    const SetSpeichern = vi.fn(() => speicherung())

    belegeApi({
      Sets: () => Promise.resolve({ sets: [probeSet()] }),
      SetSpeichern,
    })

    zeige()

    await druecke(texte.seiten.sets.speichern)

    expect(SetSpeichern).toHaveBeenCalledWith('4')

    expect(
      await meldung(`${texte.seiten.sets.gespeichert} /pfad/probeset.mse-set`),
    ).toBeVisible()
  })

  it('fragt bei „Sichern unter" nach einem Ort', async () => {
    const SetSpeichernUnter = vi.fn(() =>
      speicherung({ pfad: '/woanders/probeset.mse-set' }),
    )

    belegeApi({
      Sets: () => Promise.resolve({ sets: [probeSet()] }),
      SetSpeichernUnter,
    })

    zeige()

    await druecke(texte.seiten.sets.speichernUnter)

    expect(SetSpeichernUnter).toHaveBeenCalledWith('4')
    expect(
      await meldung(
        `${texte.seiten.sets.gespeichert} /woanders/probeset.mse-set`,
      ),
    ).toBeVisible()
  })

  // **Was nicht mitkam, steht beim Erfolg** (§8): die Setdatei ist geschrieben,
  // und dass ein Bild fehlt, erführe man sonst erst beim Öffnen.
  it('nennt, was nicht mitgekommen ist', async () => {
    belegeApi({
      Sets: () => Promise.resolve({ sets: [probeSet()] }),
      SetSpeichern: () =>
        speicherung({ befunde: ['probeset.mse-set → image1: nicht lesbar'] }),
    })

    zeige()

    await druecke(texte.seiten.sets.speichern)

    expect(
      await screen.findByText(texte.seiten.sets.beimSpeichernBefunde),
    ).toBeInTheDocument()
    expect(
      screen.getByText('probeset.mse-set → image1: nicht lesbar'),
    ).toBeInTheDocument()
  })

  it('sagt nichts, wenn niemand einen Ort gewählt hat', async () => {
    belegeApi({
      Sets: () => Promise.resolve({ sets: [probeSet()] }),
      SetSpeichernUnter: () => speicherung({ abgebrochen: true, pfad: '' }),
    })

    zeige()

    await druecke(texte.seiten.sets.speichernUnter)

    expect(
      screen.queryByText(texte.seiten.sets.gespeichert, { exact: false }),
    ).not.toBeInTheDocument()
  })

  it('meldet einen Fehler beim Sichern', async () => {
    belegeApi({
      Sets: () => Promise.resolve({ sets: [probeSet()] }),
      SetSpeichern: () =>
        Promise.resolve({
          pfad: '',
          befunde: [],
          abgebrochen: false,
          fehler: {
            code: 'nichtGefunden',
            schluessel: 'fehler.nichtGefunden',
            meldung: 'das Set führt keine Quelldatei',
            feld: '',
          },
        }),
    })

    zeige()

    await druecke(texte.seiten.sets.speichern)

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'das Set führt keine Quelldatei',
    )
  })
})
