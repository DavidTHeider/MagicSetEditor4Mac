// **Der Kampf in Zeilen** — das Artboard *Tisch* zeigt ihn neben dem Stapel:
// „Aschenwächter → Hallenwart", „drei weitere ungeblockt → Bo".
//
// # Wogegen hier geprüft wird
//
// **Gegen den Tisch vom 2026-09-24**: der Kampf stand im Kern (`M6`, CR 508.1
// und CR 509.1, Fassung 2026-09-25) und in den Feldern, mit denen man ihn
// erklärt — lesen ließ er sich nirgends. Wer blocken will, muss wissen, wer
// angreift; wer den Schaden verteilt, wer geblockt ist.
import { describe, expect, it } from 'vitest'

import { einObjekt, einePartie, einSitz, einZug } from '../seiten/partieprobe'
import { kampfzeilen } from './kampfzeilen'

import type { Kampfpaar } from '../apizug'

// imKampf gibt eine Partie mit vier Kreaturen Annas und zwei Berts.
function imKampf(angriffe: Kampfpaar[], bloecke: Kampfpaar[] = []) {
  const kreatur = (kennung: string, name: string, wer: string) =>
    einObjekt({
      kennung,
      name,
      zone: 'tisch',
      eigner: wer,
      beherrscher: wer,
      geheim: false,
    })

  return einePartie({
    sitze: [
      einSitz({ kennung: 'sitz1', name: 'Anna' }),
      einSitz({ kennung: 'sitz2', name: 'Bert' }),
    ],
    objekte: [
      kreatur('a1', 'Aschenwächter', 'sitz1'),
      kreatur('a2', 'Irrlichtweber', 'sitz1'),
      kreatur('a3', 'Glutwind', 'sitz1'),
      kreatur('a4', 'Eisenmoos', 'sitz1'),
      kreatur('b1', 'Hallenwart', 'sitz2'),
      kreatur('b2', 'Steinmoos', 'sitz2'),
    ],
    zug: einZug({
      nummer: 7,
      phase: 'kampf',
      schritt: 'blocker',
      angriffe,
      bloecke,
    }),
  })
}

describe('kampfzeilen', () => {
  it('ist leer, solange niemand angreift', () => {
    expect(kampfzeilen(imKampf([]))).toEqual([])
  })

  it('nennt einen geblockten Angreifer mit seinem Blocker', () => {
    expect(
      kampfzeilen(
        imKampf(
          [{ kreatur: 'a1', ziel: 'sitz2' }],
          [{ kreatur: 'b1', angreifer: 'a1' }],
        ),
      ),
    ).toEqual([
      { art: 'geblockt', angreifer: 'Aschenwächter', blocker: ['Hallenwart'] },
    ])
  })

  // **Mehrere Blocker an einem Angreifer** — CR 509.1a lässt es zu.
  it('und alle Blocker, wenn mehrere ihn blocken', () => {
    expect(
      kampfzeilen(
        imKampf(
          [{ kreatur: 'a1', ziel: 'sitz2' }],
          [
            { kreatur: 'b1', angreifer: 'a1' },
            { kreatur: 'b2', angreifer: 'a1' },
          ],
        ),
      ),
    ).toEqual([
      {
        art: 'geblockt',
        angreifer: 'Aschenwächter',
        blocker: ['Hallenwart', 'Steinmoos'],
      },
    ])
  })

  it('fasst die ungeblockten je Ziel zusammen, nach den geblockten', () => {
    expect(
      kampfzeilen(
        imKampf(
          [
            { kreatur: 'a1', ziel: 'sitz2' },
            { kreatur: 'a2', ziel: 'sitz2' },
            { kreatur: 'a3', ziel: 'sitz2' },
            { kreatur: 'a4', ziel: 'sitz2' },
          ],
          [{ kreatur: 'b1', angreifer: 'a1' }],
        ),
      ),
    ).toEqual([
      { art: 'geblockt', angreifer: 'Aschenwächter', blocker: ['Hallenwart'] },
      {
        art: 'ungeblockt',
        angreifer: ['Irrlichtweber', 'Glutwind', 'Eisenmoos'],
        ziel: 'Bert',
      },
    ])
  })

  // **Ein Angriff auf einen Planeswalker** trägt dessen Kennung als Ziel
  // (CR 508.1b) — genannt wird er beim Namen.
  it('nennt ein Ziel, das eine Karte ist, beim Namen der Karte', () => {
    expect(kampfzeilen(imKampf([{ kreatur: 'a1', ziel: 'b2' }]))[0]).toEqual({
      art: 'ungeblockt',
      angreifer: ['Aschenwächter'],
      ziel: 'Steinmoos',
    })
  })

  // **Fremde Daten sind der Normalfall** (§8): eine Kennung, die diese Sicht
  // nicht führt, steht als Kennung da und nicht als Lücke.
  it('nennt, was die Sicht nicht kennt, bei seiner Kennung', () => {
    expect(
      kampfzeilen(
        imKampf(
          [{ kreatur: 'x9', ziel: 'weg' }],
          [{ kreatur: 'y9', angreifer: 'x9' }],
        ),
      ),
    ).toEqual([{ art: 'geblockt', angreifer: 'x9', blocker: ['y9'] }])
  })

  // **Beschädigte Daten** (§8): der Kern schickt jeden Block mit seinem
  // Angreifer und jeden Angriff mit seinem Ziel (CR 508.1b, CR 509.1a) — eine
  // von Hand geänderte Partie muss es nicht.
  it('lässt einen Block ohne Angreifer niemanden blocken', () => {
    expect(
      kampfzeilen(
        imKampf([{ kreatur: 'a1', ziel: 'sitz2' }], [{ kreatur: 'b1' }]),
      ),
    ).toEqual([
      { art: 'ungeblockt', angreifer: ['Aschenwächter'], ziel: 'Bert' },
    ])
  })

  it('und führt einen Angriff ohne Ziel unter einem leeren', () => {
    expect(kampfzeilen(imKampf([{ kreatur: 'a1' }]))).toEqual([
      { art: 'ungeblockt', angreifer: ['Aschenwächter'], ziel: '' },
    ])
  })
})
