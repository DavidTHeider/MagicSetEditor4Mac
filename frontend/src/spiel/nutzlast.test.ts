// **Was aus eingesammelten Werten wird** —
// #1623.
//
// # Warum diese Läufe neben denen der Leiste stehen
//
// **Weil sie die Gestalt prüfen und nicht die Bedienung.** Ob eine `1` als
// Zahl hinausgeht, entscheidet sich hier; dass ein Klick sie auslöst, steht
// in `bausteine/handlungsleiste.test.tsx`. Ein Fall wie „dieselbe Kennung
// ist schon vergeben" ließe sich über die Oberfläche nur umständlich
// herstellen — und sagte dann etwas über den Aufbau statt über die Regel.
import { describe, expect, test } from 'vitest'

import { befehlsformen } from './befehlsformen'
import { formenDerGruppe } from './formenwahl'
import { einObjekt, einePartie } from '../seiten/partieprobe'
import { nutzlastAus, sichtbareFelder } from './nutzlast'

import type { Befehlsform } from './befehlsformen'

// formVon gibt die Form einer Art aus der Tabelle.
function formVon(art: string): Befehlsform {
  const form = befehlsformen.find((eine) => eine.art === art)

  if (form === undefined) {
    throw new Error(`keine Form zu ${art}`)
  }

  return form
}

describe('die Nutzlast', () => {
  // **Ein Sitzfeld beginnt beim eigenen Sitz** — P3 (#2371): das
  // Auswahlfeld zeigte den ersten Sitz, und ohne Wahl ging keiner hinaus.
  test.each(['objekt.beherrschen', 'objekt.ziehen'])(
    '%s schickt ohne Wahl den eigenen Sitz',
    (art) => {
      const nutzlast = nutzlastAus(
        formVon(art),
        {},
        einObjekt({ kennung: 'sitz1/k1/0' }),
        einePartie({ fuer: 'sitz2' }),
      )

      expect(Object.values(nutzlast)).toContain('sitz2')
    },
  )

  test('ein Kartenbefehl ohne gewähltes Objekt nennt keines', () => {
    expect(
      nutzlastAus(formVon('objekt.umdrehen'), {}, undefined, einePartie()),
    ).toEqual({})
  })

  test('eine Liste geht als Liste hinaus', () => {
    expect(
      nutzlastAus(
        formVon('objekt.wirken'),
        { ziele: ['sitz2', 'sitz1/k1/0'] },
        einObjekt({ kennung: 'sitz1/k1/0' }),
        einePartie(),
      ),
    ).toEqual({
      objekt: 'sitz1/k1/0',
      ziele: ['sitz2', 'sitz1/k1/0'],
    })
  })

  test('eine leere Liste ebenso — und nicht als nichts', () => {
    expect(
      nutzlastAus(
        formVon('kampf.angreifer'),
        { angriffe: [] },
        undefined,
        einePartie(),
      ),
    ).toEqual({ angriffe: [] })
  })

  test('ein Paar ohne zweite Hälfte trägt eine leere', () => {
    // **Der Fall kommt aus einer Zeile ohne Trennzeichen** — er entsteht in
    // der Oberfläche nicht, und der Kern weist ihn ab. Was hier geprüft
    // wird, ist, dass die Gestalt **stimmt**: ein Paar mit leerem Ziel ist
    // etwas anderes als ein `undefined` im JSON.
    expect(
      nutzlastAus(
        formVon('kampf.blocker'),
        { bloecke: ['nurkreatur'] },
        undefined,
        einePartie(),
      ),
    ).toEqual({ bloecke: [{ kreatur: 'nurkreatur', angreifer: '' }] })
  })

  test('ein Spielstein aus einer Liste wird wieder eine Zeile', () => {
    expect(
      nutzlastAus(
        formVon('spieler.spielstein'),
        { felder: ['Soldat', 'Creature'] },
        undefined,
        einePartie(),
      ),
    ).toEqual({
      // **Ohne Kennung** (#2480): sie vergibt der Kern, denn nur er kennt
      // die vergebenen.
      felder: { name: 'Soldat', 'super type': 'Creature' },
    })
  })

  test('ein Feld ohne Vorgabe und ohne Eingabe fällt heraus', () => {
    expect(
      nutzlastAus(
        formVon('objekt.anhaengen'),
        {},
        einObjekt({ kennung: 'sitz1/k1/0' }),
        einePartie(),
      ),
    ).toEqual({ objekt: 'sitz1/k1/0' })
  })

  test('und ein leeres Spielsteinfeld ebenso', () => {
    expect(
      nutzlastAus(
        formVon('spieler.spielstein'),
        { felder: '|Creature||' },
        undefined,
        einePartie(),
      ),
    ).toEqual({
      felder: { 'super type': 'Creature' },
    })
  })
})

describe('welche Formen eine Gruppe zeigt', () => {
  test('ohne Objekt zeigt sie auch, was an einen Zustand gebunden ist', () => {
    const arten = formenDerGruppe('karte', ['objekt.tappen']).map(
      (form) => form.art,
    )

    expect(arten).toEqual(['objekt.tappen'])
  })

  test('und was hängt, lässt sich lösen', () => {
    const arten = formenDerGruppe('karte', ['objekt.loesen'], {
      getappt: false,
      haengtAn: 'sitz1/k1/0',
      haufen: '',
    }).map((form) => form.art)

    expect(arten).toEqual(['objekt.loesen'])
  })

  // **Und was in einem Haufen liegt, lässt sich herausnehmen** — CR 400.5
  // (#1756).
  test('und was in einem Haufen liegt, lässt sich herausnehmen', () => {
    const arten = formenDerGruppe('karte', ['objekt.enthaeufen'], {
      getappt: false,
      haengtAn: '',
      haufen: 'sitz1/k2/0',
    }).map((form) => form.art)

    expect(arten).toEqual(['objekt.enthaeufen'])
  })

  // **Die Gegenprobe**: ohne Haufen steht er nicht da. Sonst sagte der Lauf
  // oben nur, dass der Befehl in der Liste steht.
  test('und wo keiner liegt, steht er nicht da', () => {
    const arten = formenDerGruppe('karte', ['objekt.enthaeufen'], {
      getappt: false,
      haengtAn: '',
      haufen: '',
    }).map((form) => form.art)

    expect(arten).toEqual([])
  })
})

// **Die Stelle gehört nur in eine geordnete Zone** — Befund **F04** der
// nativen Prüfung
// (#1748).
//
// Die ausgelieferte 0.5.0 antwortete auf „Karte schieben" mit *„in hand gibt
// es keine Stelle, an die sich etwas legen ließe"* — und zwar **im
// Auslieferungszustand des Formulars**: die Vorgabe der Zone ist `hand`, die
// Vorgabe der Stelle ist `oben`, und die Auswahlliste bot keinen neutralen
// Wert. Wer den Befehl unverändert stellte, bekam einen Fehler; wer ihn
// beheben wollte, fand nichts, womit.
describe('die Stelle beim Schieben', () => {
  test('eine ungeordnete Zielzone bekommt keine Stelle', () => {
    expect(
      nutzlastAus(
        formVon('objekt.schieben'),
        { zone: 'hand', stelle: 'oben' },
        einObjekt({ kennung: 'sitz1/k1/0' }),
        einePartie(),
      ),
    ).toEqual({ objekt: 'sitz1/k1/0', zone: 'hand' })
  })

  test('und die Vorgabe allein reicht dafür aus', () => {
    expect(
      nutzlastAus(
        formVon('objekt.schieben'),
        {},
        einObjekt({ kennung: 'sitz1/k1/0' }),
        einePartie(),
      ),
    ).toEqual({ objekt: 'sitz1/k1/0', zone: 'hand' })
  })

  test('eine geordnete Zielzone behält sie', () => {
    expect(
      nutzlastAus(
        formVon('objekt.schieben'),
        { zone: 'bibliothek', stelle: 'unten' },
        einObjekt({ kennung: 'sitz1/k1/0' }),
        einePartie(),
      ),
    ).toEqual({ objekt: 'sitz1/k1/0', zone: 'bibliothek', stelle: 'unten' })
  })
})

// **Was `sichtbareFelder` einem Prädikat antwortet** — die zweite Hälfte von
// F04. Sie steht hier und nicht in `bausteine/schiebeform.test.tsx`, weil sie
// eine Aussage über den Vertrag ist und nicht über eine Bedienung.
describe('der Nachschlag, den ein Feldprädikat bekommt', () => {
  // eineForm baut eine Form mit einem bedingten Feld.
  function eineForm(nurWenn: (wert: (name: string) => string) => boolean) {
    return {
      art: 'probe.form',
      gruppe: 'zug',
      felder: [
        { name: 'a', art: 'text', vorgabe: 'eins' },
        { name: 'b', art: 'text', vorgabe: 'zwei', nurWenn },
      ],
    } as Befehlsform
  }

  test('gibt die Vorgabe, wo niemand etwas eingegeben hat', () => {
    expect(
      sichtbareFelder(
        eineForm((wert) => wert('a') === 'eins'),
        {},
      ).map((feld) => feld.name),
    ).toEqual(['a', 'b'])
  })

  test('gibt den eingegebenen Wert, wo einer dasteht', () => {
    expect(
      sichtbareFelder(
        eineForm((wert) => wert('a') === 'eins'),
        {
          a: 'anders',
        },
      ).map((feld) => feld.name),
    ).toEqual(['a'])
  })

  // **Ein Name, den diese Form nicht hat, ist kein Absturz.** Die Antwort ist
  // „nichts gewählt" — dieselbe vorsichtige Form wie überall sonst hier.
  test('gibt leer, wo die Form den Namen gar nicht führt', () => {
    expect(
      sichtbareFelder(
        eineForm((wert) => wert('gibtsnicht') === ''),
        {},
      ).map((feld) => feld.name),
    ).toEqual(['a', 'b'])
  })
})
