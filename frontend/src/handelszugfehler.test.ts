// **Jeder Teilschritt hat seinen eigenen Fehlerweg** — P13
// (#2381),
// Abnahmekriterium 2: was geglückt ist, bleibt; der Zug hält an der Stelle,
// die scheiterte, und nennt sie; ein zweiter Klick beginnt bei dem Schritt,
// den der Vertrag dann nennt, und geht keinen doppelt.
import { afterEach, describe, expect, it } from 'vitest'

import { ziehe } from './handelszug'
import { du, einDienst, einStand, ich } from './test/handelszugprobe'
import { belegeApi, raeumeApi } from './test/api'

import type { Zugschritt } from './handelszug'
import type { Handelsvorgang as Vorgangsstand } from './apihandelsleitung'

afterEach(() => {
  raeumeApi()
})

// zug belegt einen Dienst, der den gesehenen Stand bestätigt, und geht den
// Zug — so weit, wie ein Knopf der Buchung geht.
async function zug(
  folge: Record<string, Vorgangsstand>,
  scheitert: Partial<Record<string, string>>,
  stand: Vorgangsstand,
  schritt: Zugschritt,
  bis: Zugschritt = 'bestaetigen',
) {
  const dienst = einDienst(folge, scheitert, stand)
  belegeApi(dienst.api as never)
  const ausgang = await ziehe({ stand, schritt, bis, gegenueber: du })

  return { ausgang, rufe: dienst.rufe }
}

const offen = einStand({
  zustand: 'uebergabeOffen',
  naechster: 'verbuchen',
  phase: 'verbuchen',
})
const verbucht = einStand({
  zustand: 'lokalVerbucht',
  naechster: 'abgleichen',
  phase: 'verbuchen',
})
const abgleich = einStand({
  zustand: 'abgleichOffen',
  beschluss: 'bestaetigen',
  naechster: 'bestaetigen',
  phase: 'verbuchen',
})
const bestaetigt = einStand({
  ...abgleich,
  zustimmungen: [ich],
  naechster: '',
  lage: 'wartet',
})
const vorbereitet = einStand({
  zustand: 'vorbereitet',
  beschluss: 'uebergeben',
  zustimmungen: [du],
  naechster: 'uebergeben',
})

const folge = {
  uebergeben: offen,
  verbuchen: verbucht,
  abgleichen: abgleich,
  [`bestaetigen:${ich}`]: bestaetigt,
}

describe('ein Zug, der unterwegs scheitert', () => {
  it('behält die Übergabe, wenn das Verbuchen scheitert, und schickt sie hinaus', async () => {
    const { ausgang, rufe } = await zug(
      folge,
      {
        verbuchen: 'Der Bestand reicht nicht.',
      },
      vorbereitet,
      'uebergeben',
    )

    expect(ausgang.gegangen).toEqual(['uebergeben'])
    expect(ausgang.gescheitert?.teil).toBe('schritt')
    expect(ausgang.gescheitert?.schritt).toBe('verbuchen')
    expect(String(ausgang.gescheitert?.ursache)).toContain(
      'Der Bestand reicht nicht.',
    )
    expect(rufe).toEqual([
      'stand',
      'uebergeben',
      'ausgang',
      'freigabe:verbuchen:',
      'verbuchen',
      'zustellen',
    ])
  })

  it('verbucht beim zweiten Klick kein zweites Mal — er beginnt beim Abgleich', async () => {
    const { ausgang: halb } = await zug(
      folge,
      { abgleichen: 'Gerade nicht.' },
      offen,
      'verbuchen',
    )

    expect(halb.gegangen).toEqual(['verbuchen'])
    expect(halb.gescheitert?.schritt).toBe('abgleichen')
    raeumeApi()

    // **Der Stand nach dem ersten Klick** sagt, was als Nächstes kommt.
    const { ausgang: ganz, rufe } = await zug(folge, {}, verbucht, 'abgleichen')

    expect(ganz.gegangen).toEqual(['abgleichen', 'bestaetigen'])
    expect(rufe).not.toContain('verbuchen')
    expect(rufe).not.toContain('freigabe:verbuchen:')
  })

  it('schickt nichts, wenn schon der erste Schritt scheitert', async () => {
    const { ausgang, rufe } = await zug(
      folge,
      {
        uebergeben: 'Nicht deine Stimme.',
      },
      vorbereitet,
      'uebergeben',
    )

    expect(ausgang.gegangen).toEqual([])
    expect(rufe).toEqual(['stand', 'uebergeben'])
  })

  it('schickt einen Stand ohne seine Stimme nicht hinaus', async () => {
    // **Sonst hieße die Nachricht mit der Stimme drüben „bekannt"** — eine
    // zweite derselben Art zur selben Revision.
    const { ausgang, rufe } = await zug(
      {
        vorbereiten: einStand({
          zustand: 'vorbereitet',
          beschluss: 'uebergeben',
          naechster: 'uebergeben',
        }),
      },
      { uebergeben: 'Gerade nicht.' },
      einStand({ zustand: 'angenommen', naechster: 'vorbereiten' }),
      'vorbereiten',
    )

    expect(ausgang.gegangen).toEqual(['vorbereiten'])
    expect(rufe).toEqual(['stand', 'vorbereiten', 'uebergeben'])
  })

  it('geht nicht weiter, wenn die Freigabe fehlt', async () => {
    const { ausgang, rufe } = await zug(
      folge,
      {
        'freigabe:verbuchen:': 'Das geht nur am Fenster.',
      },
      offen,
      'verbuchen',
    )

    // **Die Übergabe liegt schon im Ausgang und geht hinaus** — sie gilt,
    // auch wenn hier nicht gebucht wird.
    expect(ausgang.gescheitert?.teil).toBe('schritt')
    expect(rufe).toEqual([
      'stand',
      'ausgang',
      'freigabe:verbuchen:',
      'zustellen',
    ])
  })

  it('schickt hinaus, was gegangen ist, auch wenn der Stand danach nicht zu lesen ist', async () => {
    const { ausgang, rufe } = await zug(
      folge,
      { 'stand-danach': 'Der Bestand ist belegt.' },
      offen,
      'verbuchen',
    )

    expect(ausgang.gegangen).toEqual(['verbuchen'])
    expect(ausgang.gescheitert?.teil).toBe('lesen')
    expect(rufe).toEqual([
      'stand',
      'ausgang',
      'freigabe:verbuchen:',
      'verbuchen',
      'stand',
      'ausgang',
      'zustellen',
    ])
  })

  it('bucht nicht, solange die Übergabe nicht im Ausgang liegt', async () => {
    const { ausgang, rufe } = await zug(
      folge,
      {
        ausgang: 'Der Schlüsselbund hat abgelehnt.',
      },
      vorbereitet,
      'uebergeben',
    )

    expect(ausgang.gegangen).toEqual(['uebergeben'])
    expect(ausgang.gescheitert?.teil).toBe('ausgang')
    expect(ausgang.gescheitert?.schritt).toBe('verbuchen')
    expect(rufe).toEqual(['stand', 'uebergeben', 'ausgang'])
  })

  it('nennt ein Scheitern im Ausgang auch am Ende des Zugs', async () => {
    const { ausgang, rufe } = await zug(
      {
        zustimmen: einStand({ zustand: 'angenommen', lage: 'wartet' }),
      },
      { ausgang: 'Der Schlüsselbund hat abgelehnt.' },
      einStand({
        zustand: 'angeboten',
        beschluss: 'zustimmen',
        zustimmungen: [du],
        naechster: 'zustimmen',
      }),
      'zustimmen',
      'zustimmen',
    )

    expect(ausgang.gegangen).toEqual(['zustimmen'])
    expect(ausgang.gescheitert).toEqual(
      expect.objectContaining({ teil: 'ausgang', schritt: '' }),
    )
    expect(rufe).toEqual(['stand', 'zustimmen', 'ausgang'])
  })

  it('behält alles, wenn nur das Übertragen scheitert, und sagt es eigens', async () => {
    const { ausgang } = await zug(
      folge,
      {
        zustellen: 'Bert ist nicht zu erreichen.',
      },
      vorbereitet,
      'uebergeben',
    )

    expect(ausgang.gegangen).toEqual([
      'uebergeben',
      'verbuchen',
      'abgleichen',
      'bestaetigen',
    ])
    expect(ausgang.gescheitert).toBeUndefined()
    expect(ausgang.getragen).toBe(0)
    expect(String(ausgang.zustellung)).toContain('Bert ist nicht zu erreichen.')
  })
})
