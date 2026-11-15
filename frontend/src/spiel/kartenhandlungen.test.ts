// **Was an einer Karte sofort geht** — PLAY-04.AK-01
// (#1712).
import { describe, expect, it } from 'vitest'

import { formenDerGruppe } from './formenwahl'
import { sofortigeHandlungen } from './kartenhandlungen'

const offen = { getappt: false, haengtAn: '', haufen: '' }
const getappt = { getappt: true, haengtAn: '', haufen: '' }

describe('sofortigeHandlungen', () => {
  it('gibt nur, was ohne ein weiteres Feld auskommt', () => {
    expect(
      sofortigeHandlungen(
        ['objekt.tappen', 'objekt.schieben', 'objekt.umdrehen'],
        offen,
      ),
    ).toEqual(['objekt.tappen', 'objekt.umdrehen'])
  })

  // **Was der Kern nicht erlaubt, steht nicht da** — ein Eintrag, den er
  // abwiese, wäre eine Einladung, die er zurücknimmt.
  it('gibt nichts, was der Kern nicht erlaubt', () => {
    expect(sofortigeHandlungen([], offen)).toEqual([])
  })

  // **Und was am Zustand nicht passt, ebenso wenig** — „Tappen" an einer
  // getappten Karte ist ein Knopf, der immer abgewiesen wird.
  it('folgt dem Zustand des Objekts', () => {
    const beide = ['objekt.tappen', 'objekt.enttappen']

    expect(sofortigeHandlungen(beide, offen)).toEqual(['objekt.tappen'])
    expect(sofortigeHandlungen(beide, getappt)).toEqual(['objekt.enttappen'])
  })

  // **Keine Handlung, die es nur im Kontextmenü gibt** — das ist die Abnahme
  // von U2 (#1413), und sie gilt auch hier: alles, was das Menü anbietet,
  // steht auch in der Handlungsleiste.
  it('bietet nichts an, was die Handlungsleiste nicht auch zeigt', () => {
    const alles = [
      'objekt.tappen',
      'objekt.umdrehen',
      'objekt.schieben',
      'objekt.zaehler',
      'objekt.opfern',
    ]

    const inDerLeiste = formenDerGruppe('karte', alles, offen).map(
      (form) => form.art,
    )

    for (const art of sofortigeHandlungen(alles, offen)) {
      expect(inDerLeiste).toContain(art)
    }
  })
})
