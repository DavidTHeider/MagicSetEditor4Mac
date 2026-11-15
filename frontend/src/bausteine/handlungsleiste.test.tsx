// **Die Handlungsleiste** —
// #1623.
//
// # Wogegen hier geprüft wird
//
// **Dass beim Klick das Richtige hinausgeht** — Art **und** Nutzlast, in der
// Gestalt, die der Kern liest: eine Zahl als Zahl, eine Liste als Liste, ein
// Paar als Paar. Eine `"1"` statt einer `1` fällt sonst erst an der Naht auf,
// wo die Meldung von etwas anderem spricht.
//
// **Und dass kein Knopf dasteht, den der Kern abwiese.** Was erlaubt ist,
// rechnet er aus; das Fenster zeigt genau das.
import { screen, within } from '@testing-library/react'
import { describe, expect, test } from 'vitest'

import { einePartie } from '../seiten/partieprobe'
import {
  eigene,
  eineKarte,
  oeffne,
  stelle,
  zeige,
} from '../test/handlungsprobe'

describe('was die Leiste zeigt', () => {
  test('ohne erlaubte Befehle steht der Satz statt der Knöpfe', () => {
    zeige({ erlaubt: [] })

    expect(screen.getByText(eigene.handlungenLeer)).toBeInTheDocument()
    expect(screen.queryByText(eigene.handlungenNurErlaubt)).toBeNull()
  })

  // **Warum ein Knopf fehlt, steht unter den Knöpfen** — Artboard *Tisch*:
  // angeboten wird, was der Kern gerade erlaubt, und der Rest steht gar
  // nicht da. Ohne den Satz sähe ein fehlender Knopf aus wie ein fehlender.
  test('mit erlaubten Befehlen sagt sie, dass nur sie dastehen', () => {
    zeige({ erlaubt: ['zone.mischen'] })

    expect(screen.getByText(eigene.handlungenNurErlaubt)).toBeInTheDocument()
  })

  test('ohne gewählte Karte steht der Satz statt der Kartenbefehle', () => {
    zeige({ erlaubt: ['objekt.tappen'] })

    expect(screen.getByText(eigene.handlungenKeineKarte)).toBeInTheDocument()

    expect(
      screen.queryByRole('button', {
        name: eigene.handlungsname['objekt.tappen'],
      }),
    ).toBeNull()
  })

  test('eine Gruppe ohne Befehle steht gar nicht da', () => {
    zeige({ erlaubt: ['zone.mischen'] })

    expect(screen.getByText(eigene.handlungenZone)).toBeInTheDocument()
    expect(screen.queryByText(eigene.handlungenKampf)).toBeNull()
  })

  test('eine getappte Karte bietet kein zweites Tappen', () => {
    zeige({
      erlaubt: ['objekt.tappen', 'objekt.enttappen'],
      objekt: eineKarte({ getappt: true }),
    })

    expect(
      screen.queryByRole('button', {
        name: eigene.handlungsname['objekt.tappen'],
      }),
    ).toBeNull()

    expect(
      screen.getByRole('button', {
        name: eigene.handlungsname['objekt.enttappen'],
      }),
    ).toBeInTheDocument()
  })

  test('und nur was hängt, lässt sich lösen', () => {
    zeige({ erlaubt: ['objekt.loesen'], objekt: eineKarte() })

    expect(
      screen.queryByRole('button', {
        name: eigene.handlungsname['objekt.loesen'],
      }),
    ).toBeNull()
  })
})

describe('was beim Klick hinausgeht', () => {
  test('ein Befehl ohne Felder schickt nur die Kennung', async () => {
    const { hand, befehlen } = zeige({
      erlaubt: ['objekt.umdrehen'],
      objekt: eineKarte(),
    })

    await hand.click(
      screen.getByRole('button', {
        name: eigene.handlungsname['objekt.umdrehen'],
      }),
    )

    expect(befehlen).toHaveBeenCalledWith('objekt.umdrehen', {
      objekt: 'sitz1/k1/0',
    })
  })

  test('eine Zahl geht als Zahl hinaus', async () => {
    const { hand, befehlen } = zeige({ erlaubt: ['spieler.leben'] })

    await stelle(hand, 'spieler.leben')

    expect(befehlen).toHaveBeenCalledWith('spieler.leben', { um: -1 })
  })

  test('und eine geänderte Zahl ebenso', async () => {
    const { hand, befehlen } = zeige({ erlaubt: ['spieler.vergiftung'] })

    const feld = within(
      await oeffne(hand, 'spieler.vergiftung'),
    ).getByLabelText(eigene.handlungsfeld.um)

    await hand.clear(feld)
    await hand.type(feld, '3')
    await stelle(hand, 'spieler.vergiftung')

    expect(befehlen).toHaveBeenCalledWith('spieler.vergiftung', { um: 3 })
  })

  test('ein leeres Feld fällt heraus', async () => {
    const { hand, befehlen } = zeige({ erlaubt: ['spieler.muenzwurf'] })

    await stelle(hand, 'spieler.muenzwurf')

    expect(befehlen).toHaveBeenCalledWith('spieler.muenzwurf', {})
  })

  test('und ein gewählter Ruf geht mit', async () => {
    const { hand, befehlen } = zeige({ erlaubt: ['spieler.muenzwurf'] })

    await hand.selectOptions(
      within(await oeffne(hand, 'spieler.muenzwurf')).getByLabelText(
        eigene.handlungsfeld.ruf,
      ),
      'kopf',
    )

    await stelle(hand, 'spieler.muenzwurf')

    expect(befehlen).toHaveBeenCalledWith('spieler.muenzwurf', { ruf: 'kopf' })
  })

  test('eine Sitzwahl nennt die Kennung und nicht den Namen', async () => {
    const { hand, befehlen } = zeige({
      erlaubt: ['objekt.beherrschen'],
      objekt: eineKarte(),
    })

    await hand.selectOptions(
      within(await oeffne(hand, 'objekt.beherrschen')).getByLabelText(
        eigene.handlungsfeld.wer,
      ),
      'sitz2',
    )

    await stelle(hand, 'objekt.beherrschen')

    expect(befehlen).toHaveBeenCalledWith('objekt.beherrschen', {
      objekt: 'sitz1/k1/0',
      wer: 'sitz2',
    })
  })

  test('eine Zielwahl geht als Liste hinaus', async () => {
    const partie = einePartie({ objekte: [eineKarte()] })

    const { hand, befehlen } = zeige({
      erlaubt: ['objekt.wirken'],
      objekt: eineKarte(),
      partie,
    })

    await hand.selectOptions(
      within(await oeffne(hand, 'objekt.wirken')).getByLabelText(
        eigene.handlungsfeld.ziele,
      ),
      ['sitz2'],
    )

    await stelle(hand, 'objekt.wirken')

    expect(befehlen).toHaveBeenCalledWith('objekt.wirken', {
      objekt: 'sitz1/k1/0',
      ziele: ['sitz2'],
    })
  })

  test('ein Spielstein bringt seine Eigenschaften mit', async () => {
    const { hand, befehlen } = zeige({ erlaubt: ['spieler.spielstein'] })

    const reihe = await oeffne(hand, 'spieler.spielstein')

    await hand.type(
      within(reihe).getByLabelText(eigene.spielsteinName),
      'Soldat',
    )
    await hand.type(
      within(reihe).getByLabelText(eigene.spielsteinTyp),
      'Creature',
    )
    await hand.type(within(reihe).getByLabelText(eigene.spielsteinStaerke), '1')

    await stelle(hand, 'spieler.spielstein')

    // **Ohne Kennung** (#2480) — sie vergibt der Kern.
    expect(befehlen).toHaveBeenCalledWith('spieler.spielstein', {
      felder: { name: 'Soldat', 'super type': 'Creature', power: '1' },
    })
  })

  // **Die drei Feldarten hier standen bis zum 2026-09-24 nur nebenbei
  // da**: jede Form zeigte ihre Felder, und ein Lauf über irgendeinen Knopf
  // zeichnete sie mit. Seit sie erst auf Zuruf aufklappen (UX-08), fand die
  // Abdeckung sie ungeprüft — sie bekommen je einen Lauf.
  test('eine Manaart geht als ihr Kürzel hinaus', async () => {
    const { hand, befehlen } = zeige({ erlaubt: ['spieler.mana'] })

    await hand.selectOptions(
      within(await oeffne(hand, 'spieler.mana')).getByLabelText(
        eigene.handlungsfeld.art,
      ),
      'G',
    )
    await stelle(hand, 'spieler.mana')

    expect(befehlen).toHaveBeenCalledWith('spieler.mana', { art: 'G', um: 1 })
  })

  test('eine Reihung nennt die wartende Fähigkeit beim Namen', async () => {
    const partie = einePartie({
      ausgeloest: [
        { kennung: 'a7', art: 'strahlung', beherrscher: 'sitz1', wen: '' },
      ],
    })

    const { hand, befehlen } = zeige({ erlaubt: ['spieler.reihung'], partie })

    const feld = within(await oeffne(hand, 'spieler.reihung')).getByLabelText(
      eigene.handlungsfeld.kennung,
    )

    expect(
      within(feld).getByRole('option', {
        name: eigene.ausloesername.strahlung,
      }),
    ).toBeInTheDocument()

    await hand.selectOptions(feld, 'a7')
    await stelle(hand, 'spieler.reihung')

    expect(befehlen).toHaveBeenCalledWith('spieler.reihung', { kennung: 'a7' })
  })

  test('eine Objektwahl bietet, was offen auf dem Tisch liegt', async () => {
    const partie = einePartie({
      objekte: [
        eineKarte(),
        eineKarte({ kennung: 'sitz1/k2/0', name: 'Berg' }),
      ],
    })

    const { hand, befehlen } = zeige({
      erlaubt: ['objekt.anhaengen'],
      objekt: eineKarte(),
      partie,
    })

    const feld = within(await oeffne(hand, 'objekt.anhaengen')).getByLabelText(
      eigene.handlungsfeld.an,
    )

    // **„Nichts" steht vorn**: ohne Wahl geht kein `an` hinaus, und der Kern
    // sagt, was fehlt.
    expect(
      within(feld).getByRole('option', { name: eigene.handlungKeineWahl }),
    ).toBeInTheDocument()

    await hand.selectOptions(feld, 'sitz1/k2/0')
    await stelle(hand, 'objekt.anhaengen')

    expect(befehlen).toHaveBeenCalledWith('objekt.anhaengen', {
      objekt: 'sitz1/k1/0',
      an: 'sitz1/k2/0',
    })
  })

  test('ein Emblem geht ohne Kennung hinaus — sie vergibt der Kern', async () => {
    const { hand, befehlen } = zeige({ erlaubt: ['spieler.emblem'] })

    await hand.click(
      screen.getByRole('button', {
        name: eigene.handlungsname['spieler.emblem'],
      }),
    )

    expect(befehlen).toHaveBeenCalledWith('spieler.emblem', {})
  })
})
