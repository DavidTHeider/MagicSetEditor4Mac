// **Ein Zug je Entscheidung** — P13
// (#2381),
// Befund C-16, Vermerk [0200](../../docs/entscheidungen/0200-ein-zug-je-entscheidung.md).
//
// # Was ein Zug geht
//
// **Die Schritte seiner Gruppe, die der Vertrag als nächste nennt, bis das
// Gegenüber dran ist.** Zwei Gruppen: die Einigung (anbieten, zustimmen) und
// der Abschluss (vorbereiten, übergeben, verbuchen, abgleichen, die eigene
// Stimme zum Abschluss). Was als Nächstes kommt, rechnet dieses Fenster
// nicht aus — es liest nach jedem Schritt `naechster` und `lage` aus dem
// Dienst (0151 §2). Eine Gruppe endet, wo eine andere Entscheidung beginnt:
// wer einem Gegenangebot zustimmt, hat damit noch nichts übergeben.
//
// **Die Stimme des Gegenübers geht nie in einem Zug mit** — sie ist ein
// eigener Klick (`wer`), und die Zwei-Unterschriften-Regel aus T6 bleibt.
//
// # Wann in den Ausgang gelegt wird
//
// **Wenn ein Stand verlassen wird, und am Ende.** Drüben spielt die Leitung
// höchstens einen Schritt nach und keinen Abschlussschritt
// (`domain/handelsstandnahme.go`): die Übergabe muss deshalb im Ausgang
// liegen, **bevor** gebucht wird. Und eine Stimme reist mit ihrem Stand — eine
// zweite Nachricht derselben Art zur selben Revision hieße drüben
// „bekannt", und die Stimme darin ginge verloren.
//
// # Was ein Fehler unterwegs tut
//
// **Was gegangen ist, bleibt** (§7, Regel 3), und der Zug hält an der Stelle,
// die scheiterte. Nach einem gescheiterten **Schritt** legt er nichts mehr in
// den Ausgang — der Stand, den er hätte schicken können, trüge die Stimme
// nicht, um die es ging. Übertragen wird, was schon im Ausgang liegt. Ein
// zweiter Klick beginnt bei dem Schritt, den der Vertrag dann nennt.
import { api } from './fassade'
import { auspacken } from './umschlag'
import { gedeckt } from './handelszugdeckung'

import type { Handelsvorgang as Vorgangsstand } from './apihandelsleitung'

// Zugteil ist die Stelle, an der ein Zug scheitern kann.
export type Zugteil =
  'veraltet' | 'festhalten' | 'schritt' | 'lesen' | 'ausgang'

// Zugschritt ist ein Schritt, den ein Zug gehen kann.
export type Zugschritt =
  | 'anbieten'
  | 'zustimmen'
  | 'vorbereiten'
  | 'uebergeben'
  | 'verbuchen'
  | 'abgleichen'
  | 'bestaetigen'

export interface Zugauftrag {
  // stand ist der Stand, von dem der Zug ausgeht.
  stand: Vorgangsstand

  // schritt ist der erste — der, mit dem der Knopf beginnt.
  schritt: Zugschritt

  // bis ist der letzte, den der Knopf nennt — weiter geht ein Zug nie, auch
  // wenn der Vertrag inzwischen mehr erlaubte (Befund 1 des Zweitblicks über
  // Phase 8): „Übergabe bestätigen" verbucht nicht.
  bis: Zugschritt

  // gegenueber ist die Kennung, an die die Post geht.
  gegenueber: string

  // wer ist gesetzt, wo die Stimme des Gegenübers eingetragen wird — dann
  // geht der Zug genau diesen einen Schritt.
  wer?: string
}

export interface Zugausgang {
  // stand ist der zuletzt gelesene Stand.
  stand: Vorgangsstand

  // gegangen sind die Schritte, die geglückt sind.
  gegangen: Zugschritt[]

  // gescheitert nennt die Stelle — `ursache` ist `null`, wo nicht ein Ruf
  // scheiterte, sondern das Festhalten nicht eindeutig war.
  // Der Schritt ist leer, wo erst das letzte Legen in den Ausgang scheiterte.
  gescheitert?: { teil: Zugteil; schritt: Zugschritt | ''; ursache: unknown }

  // getragen ist, wie viele Nachrichten über die Leitung gingen.
  getragen: number

  // zustellung ist der Grund, aus dem das Übertragen scheiterte — der Stand
  // bleibt dann, und die Nachricht liegt im Ausgang.
  zustellung?: unknown
}

const einigung: readonly Zugschritt[] = ['anbieten', 'zustimmen']
const abschluss: readonly Zugschritt[] = [
  'vorbereiten',
  'uebergeben',
  'verbuchen',
  'abgleichen',
  'bestaetigen',
]

// gruppeVon sagt, zu welcher Entscheidung ein Schritt gehört.
const gruppeVon: Record<Zugschritt, readonly Zugschritt[]> = {
  anbieten: einigung,
  zustimmen: einigung,
  vorbereiten: abschluss,
  uebergeben: abschluss,
  verbuchen: abschluss,
  abgleichen: abschluss,
  bestaetigen: abschluss,
}

// folgt gibt den Schritt, der nach `vorher` im selben Zug kommt — den, den
// der Stand nennt, wenn er in derselben Gruppe später steht und nicht hinter
// `bis`.
function folgt(
  vorher: Zugschritt,
  stand: Vorgangsstand,
  bis: Zugschritt,
): Zugschritt | undefined {
  const gruppe = gruppeVon[vorher]
  const naechster = gruppe.find((einer) => einer === stand.naechster)

  return stand.lage === 'dran' &&
    naechster !== undefined &&
    gruppe.indexOf(naechster) > gruppe.indexOf(vorher) &&
    gruppe.indexOf(naechster) <= gruppe.indexOf(bis)
    ? naechster
    : undefined
}

// derselbe sagt, ob zwei Stände dasselbe zur Entscheidung vorlegen —
// Revision, Abdruck, Zustand, Stimmen, Lage und nächster Schritt.
function derselbe(a: Vorgangsstand, b: Vorgangsstand): boolean {
  const stimmen = (stand: Vorgangsstand) =>
    [...stand.zustimmungen].sort().join('|')

  return (
    a.revision === b.revision &&
    a.abdruck === b.abdruck &&
    a.zustand === b.zustand &&
    a.lage === b.lage &&
    a.naechster === b.naechster &&
    stimmen(a) === stimmen(b)
  )
}

// verlaesst sagt, ob `schritt` diesen Stand verlässt — jeder, der kein
// offener Beschluss ist, und ein Beschluss, wenn die Stimme die letzte ist.
// `beschluss`, `zustimmungen` und `parteien` kommen aus `domain`.
export function verlaesst(
  schritt: string,
  stand: Vorgangsstand,
  wer: string,
): boolean {
  const stimmen = new Set([...stand.zustimmungen, wer])

  return (
    schritt !== stand.beschluss ||
    stand.parteien.every((partei) => stimmen.has(partei))
  )
}

// freigeben holt die Zustimmung dieses Fensters für einen Abschlussschritt
// (T10): der Klick ist sie — für jeden Schritt, den sein Knopf nennt.
async function freigeben(
  vorgang: string,
  schritt: string,
  wer: string,
): Promise<string> {
  const erteilt = await api().HandelFreigabeErteilen(vorgang, schritt, wer)

  return auspacken(erteilt, () => erteilt.kennung)
}

// gehe geht einen Schritt und gibt den neuen Stand — oder `undefined`, wo
// die Antwort keinen trägt. **`gesehen` ist der Abdruck des Stands, den der
// Zug zuletzt kannte** (#2496): der Dienst prüft ihn gegen den Stand, auf
// den er den Schritt anwendet. Die Abschlussschritte brauchen ihn nicht —
// nach der Einigung ändert sich die Revision nicht mehr.
async function gehe(
  schritt: Zugschritt,
  vorgang: string,
  wer: string,
  gesehen: string,
): Promise<Vorgangsstand | undefined> {
  if (schritt === 'verbuchen') {
    const freigabe = await freigeben(vorgang, schritt, '')
    const antwort = await api().Handelsverbuchung(vorgang, freigabe)
    auspacken(antwort, () => antwort)

    return undefined
  }

  if (schritt === 'abgleichen') {
    const antwort = await api().Handelsabgleich(vorgang)
    auspacken(antwort, () => antwort)

    return undefined
  }

  if (schritt === 'bestaetigen') {
    const freigabe = await freigeben(vorgang, schritt, wer)
    const antwort = await api().Handelsbestaetigung(vorgang, wer, freigabe)
    auspacken(antwort, () => antwort)

    return undefined
  }

  const antwort = await api().HandelSchritt(vorgang, schritt, '', gesehen)

  return auspacken(antwort, () => antwort.vorgang)
}

// ziehe geht einen Zug.
export async function ziehe(auftrag: Zugauftrag): Promise<Zugausgang> {
  const { gegenueber, wer } = auftrag
  const stimme = wer ?? auftrag.stand.ich

  let stand = auftrag.stand
  let schritt = auftrag.schritt

  // **Wer bucht, legt die Übergabe vorher in den Ausgang** — auch eine, die
  // nicht dieser Zug herstellte. Eine Stimme zählt drüben nur im selben Stand
  // (`NimmtStand`): wer die Übergabe zuerst bestätigt, dessen Stimme reist
  // mit dem Schritt nach „vorbereitet" und zählt dort nicht, und die Übergabe
  // schließt bei ihm erst, wenn die Stimme des Gegenübers ankommt — ohne
  // dass dieses es erfährt. Gefunden am nativen Fenster (`p13abschluss09`).
  // Kam sie schon an, ist eine zweite Nachricht wirkungslos.
  let offen = schritt === 'verbuchen'
  let gelegt = 0
  let festgehalten = wer !== undefined
  const gegangen: Zugschritt[] = []
  let gescheitert: Zugausgang['gescheitert']

  // **Nur von dem Stand aus, den der Mensch sah** (Befund 1 des Zweitblicks
  // über Phase 8): kam dazwischen ein Gegenangebot oder eine Stimme, hätte
  // der Knopf etwas anderes geheißen. Der Zug geht dann nichts und zeigt den
  // neuen Stand. **Und der Dienst prüft dasselbe noch einmal** (#2496):
  // jeder Schritt nennt den Abdruck, den er meint (`gehe`).
  try {
    const gelesen = await api().HandelStand(stand.vorgang)
    const frisch = auspacken(gelesen, () => gelesen.vorgang)

    if (!derselbe(frisch, stand)) {
      return {
        stand: frisch,
        gegangen: [],
        gescheitert: { teil: 'veraltet', schritt, ursache: null },
        getragen: 0,
      }
    }
  } catch (ursache) {
    return {
      stand,
      gegangen: [],
      gescheitert: { teil: 'lesen', schritt: '', ursache },
      getragen: 0,
    }
  }

  const lege = async () => {
    const antwort = await api().HandelSchicken(stand.vorgang, gegenueber)
    auspacken(antwort, () => antwort)
    gelegt += 1
    offen = false
  }

  for (;;) {
    if (!festgehalten && ['vorbereiten', 'uebergeben'].includes(schritt)) {
      try {
        if (!(await gedeckt(stand))) {
          gescheitert = { teil: 'festhalten', schritt, ursache: null }
          break
        }
      } catch (ursache) {
        gescheitert = { teil: 'festhalten', schritt, ursache }
        break
      }

      festgehalten = true
    }

    if (offen && verlaesst(schritt, stand, stimme)) {
      try {
        await lege()
      } catch (ursache) {
        gescheitert = { teil: 'ausgang', schritt, ursache }
        break
      }
    }

    let neu: Vorgangsstand | undefined

    try {
      neu = await gehe(schritt, stand.vorgang, stimme, stand.abdruck)
    } catch (ursache) {
      gescheitert = { teil: 'schritt', schritt, ursache }
      break
    }

    gegangen.push(schritt)
    offen = true

    try {
      if (neu === undefined) {
        const antwort = await api().HandelStand(stand.vorgang)
        neu = auspacken(antwort, () => antwort.vorgang)
      }
    } catch (ursache) {
      gescheitert = { teil: 'lesen', schritt, ursache }
      break
    }

    stand = neu

    const weiter =
      wer === undefined ? folgt(schritt, stand, auftrag.bis) : undefined

    if (weiter === undefined) {
      break
    }

    schritt = weiter
  }

  // **Was gegangen ist, geht hinaus** — nur nicht nach einem gescheiterten
  // Schritt (siehe oben) und nicht ein zweites Mal nach einem gescheiterten
  // Ausgang: wer es noch einmal will, drückt es.
  if (offen && (gescheitert === undefined || gescheitert.teil === 'lesen')) {
    try {
      await lege()
    } catch (ursache) {
      gescheitert ??= { teil: 'ausgang', schritt: '', ursache }
    }
  }

  let getragen = 0
  let zustellung: unknown

  if (gelegt > 0) {
    try {
      const bericht = await api().HandelZustellen(gegenueber)
      getragen = auspacken(bericht, () => bericht.getragen)
    } catch (ursache) {
      zustellung = ursache
    }
  }

  return {
    stand,
    gegangen,
    ...(gescheitert === undefined ? {} : { gescheitert }),
    getragen,
    ...(zustellung === undefined ? {} : { zustellung }),
  }
}
