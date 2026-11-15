// **Jeder Katalog spricht die Lesenden in einer Form an — der vertrauten** —
// AP-11 (#2368),
// Befund B-27, Vermerk 0196.
//
// Die Kataloge mischten: „Hier arbeiten Sie gerade“ neben „du“, im
// Französischen 440 „vous“ neben 27 „tu“, im Russischen „вы“ neben „ты“,
// im Chinesischen „您“ neben „你“. Die Regel je Sprache steht im Vermerk;
// hier steht, woran ein Rückfall zu erkennen ist.
//
// # Was er nicht hält
//
// **Eine Höflichkeitsform, die keine Marke trägt.** Im Spanischen sind
// „puede“ und „su“ dritte Person **oder** Anrede — dort fängt er „usted“ und
// die förmlichen Imperative, die in einer Oberfläche vorkommen, nicht jede
// Verbform. Im Deutschen ist ein „Sie“ am Satzanfang auch „sie“ (die
// Mehrzahl, eine Sache); er prüft die großgeschriebenen Formen **im** Satz.
// Englisch hat keine zwei Formen, Italienisch und Japanisch tragen heute
// keine förmlichen — der Lauf hält sie so.
import { describe, expect, it } from 'vitest'

import { kataloge } from './kataloge'
import type { Sprache } from './kataloge'

// texte gibt jeden Satz eines Katalogs mit seinem Pfad.
function texte(wert: unknown, pfad = ''): [string, string][] {
  if (typeof wert === 'string') return [[pfad, wert]]
  if (wert === null || typeof wert !== 'object') return []

  return Object.entries(wert).flatMap(([schluessel, kind]) =>
    texte(kind, pfad === '' ? schluessel : `${pfad}.${schluessel}`),
  )
}

// imSatz sagt, ob ein großgeschriebenes Wort mitten im Satz steht — am
// Anfang kann „Sie“ auch die Mehrzahl sein.
function imSatz(satz: string, stelle: number): boolean {
  const davor = satz.slice(0, stelle).trimEnd()

  return davor !== '' && !/[.!?:„“"'(—–-]$/.test(davor)
}

const regeln: Partial<Record<Sprache, (satz: string) => string[]>> = {
  de: (satz) =>
    [...satz.matchAll(/\b(Sie|Ihnen|Ihr|Ihre|Ihren|Ihrem|Ihrer)\b/g)]
      .filter(
        (treffer) => treffer[1] === 'Ihnen' || imSatz(satz, treffer.index),
      )
      .map((treffer) => treffer[1] ?? ''),
  fr: (satz) => [...satz.matchAll(/\b(vous|votre|vos)\b/gi)].map((t) => t[0]),
  es: (satz) =>
    [
      ...satz.matchAll(
        /\b(usted|ustedes|Pulse|Haga|Elija|Seleccione|Introduzca|Escriba|Compruebe|Vuelva|Abra|Cierre|Pruebe|Utilice|Use|Indique|Espere)\b/g,
      ),
    ].map((t) => t[0]),
  it: (satz) =>
    [...satz.matchAll(/\b(Lei|Suo|Sua|Suoi|Sue)\b/g)]
      .filter((t) => imSatz(satz, t.index))
      .map((t) => t[0]),
  ru: (satz) =>
    [
      ...satz.matchAll(/(?<![а-яё])(вы|вас|вам|вами|ваш[а-яё]*)(?![а-яё])/giu),
    ].map((t) => t[0]),
  zh: (satz) => [...satz.matchAll(/您/g)].map((t) => t[0]),
  ja: (satz) =>
    [...satz.matchAll(/お客様|あなた様|いただ|くださいませ/g)].map((t) => t[0]),
}

// **Wo zwei zugleich angesprochen werden** — Spielende am Tisch, Handelnde
// beim Tausch, zwei Profile beim Ziffernvergleich. Das Deutsche sagt dort
// „ihr“, und die Mehrzahl ist im Französischen „vous“, im Russischen „вы“
// (Vermerk 0196, Satz 3). Jeder Pfad steht hier, weil der deutsche Satz an
// derselben Stelle „ihr“, „euch“ oder „euer“ sagt.
const mehrzahl = new Set([
  'seiten.handeln.assistent.abgleich',
  'seiten.handeln.assistent.phasen.angebot',
  'seiten.handeln.partnerErklaerung',
  'seiten.handeln.partnerLeer',
  'seiten.handeln.tauschTitel',
  'seiten.handeln.vorgangVereinbarung',
  'seiten.spielen.regelSelbst',
  'seiten.spielen.regelgradManuell',
  'seiten.technik.einladung.erklaerung',
  'seiten.technik.einladung.gemeldet',
  'seiten.technik.einladung.meinNameHinweis',
  'seiten.technik.einladung.nameHinweis',
  'seiten.technik.einladung.technikOhneAdresse',
  'seiten.technik.einladung.vergleichen',
  'seiten.technik.fernErklaerung',
])

// deutsch gibt den deutschen Satz an einem Pfad.
function deutsch(pfad: string): string {
  let wert: unknown = kataloge.de

  for (const schluessel of pfad.split('.')) {
    wert = (wert as Record<string, unknown>)[schluessel]
  }

  return typeof wert === 'string' ? wert : ''
}

describe('die Mehrzahl in der Ausnahmeliste', () => {
  it('steht nur, wo das Deutsche „ihr“, „euch“ oder „euer“ sagt', () => {
    expect(
      [...mehrzahl].filter(
        (pfad) => !/\b(ihr|euch|euer|eure[mnrs]?)\b/i.test(deutsch(pfad)),
      ),
    ).toEqual([])
  })
})

describe('die Anrede in den Katalogen', () => {
  for (const [sprache, regel] of Object.entries(regeln) as [
    Sprache,
    (satz: string) => string[],
  ][]) {
    it(`${sprache}: spricht vertraut an`, () => {
      const foermlich = texte(kataloge[sprache])
        .filter(([pfad]) => !mehrzahl.has(pfad))
        .flatMap(([pfad, satz]) =>
          regel(satz).map(
            (wort) => `${pfad}: „${wort}“ in „${satz.slice(0, 80)}“`,
          ),
        )

      expect(foermlich).toEqual([])
    })
  }
})
