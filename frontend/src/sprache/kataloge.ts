// Die Sprachkataloge, zusammengeführt — seit LOC-01 (#1687) acht.
//
// **Deutsch ist die Ausgangssprache** (AGENTS.md §9): der Typ `Katalog` wird aus
// dem deutschen Katalog abgeleitet, und der englische muss ihn erfüllen. Ein
// fehlender englischer Schlüssel ist damit ein **Typfehler**, kein Testbefund —
// er fällt auf, bevor irgendetwas läuft.
//
// `sprache.test.ts` prüft zusätzlich zur Laufzeit. Das ist keine Doppelung: der
// Typ hält die Struktur, der Test hält den Inhalt — eine englische Zeichenkette,
// die versehentlich den deutschen Satz trägt, ist typrichtig und trotzdem falsch.
import { allgemein as deAllgemein } from './de/allgemein'
import { bausteine as deBausteine } from './de/bausteine'
import { fehler as deFehler } from './de/fehler'
import { hilfe as deHilfe } from './de/hilfe'
import { seiten as deSeiten } from './de/seiten'

import { allgemein as enAllgemein } from './en/allgemein'
import { bausteine as enBausteine } from './en/bausteine'
import { fehler as enFehler } from './en/fehler'
import { hilfe as enHilfe } from './en/hilfe'
import { seiten as enSeiten } from './en/seiten'

import { allgemein as esAllgemein } from './es/allgemein'
import { bausteine as esBausteine } from './es/bausteine'
import { fehler as esFehler } from './es/fehler'
import { hilfe as esHilfe } from './es/hilfe'
import { seiten as esSeiten } from './es/seiten'

import { allgemein as frAllgemein } from './fr/allgemein'
import { bausteine as frBausteine } from './fr/bausteine'
import { fehler as frFehler } from './fr/fehler'
import { hilfe as frHilfe } from './fr/hilfe'
import { seiten as frSeiten } from './fr/seiten'

import { allgemein as itAllgemein } from './it/allgemein'
import { bausteine as itBausteine } from './it/bausteine'
import { fehler as itFehler } from './it/fehler'
import { hilfe as itHilfe } from './it/hilfe'
import { seiten as itSeiten } from './it/seiten'

import { allgemein as jaAllgemein } from './ja/allgemein'
import { bausteine as jaBausteine } from './ja/bausteine'
import { fehler as jaFehler } from './ja/fehler'
import { hilfe as jaHilfe } from './ja/hilfe'
import { seiten as jaSeiten } from './ja/seiten'

import { allgemein as ruAllgemein } from './ru/allgemein'
import { bausteine as ruBausteine } from './ru/bausteine'
import { fehler as ruFehler } from './ru/fehler'
import { hilfe as ruHilfe } from './ru/hilfe'
import { seiten as ruSeiten } from './ru/seiten'

import { allgemein as zhAllgemein } from './zh/allgemein'
import { bausteine as zhBausteine } from './zh/bausteine'
import { fehler as zhFehler } from './zh/fehler'
import { hilfe as zhHilfe } from './zh/hilfe'
import { seiten as zhSeiten } from './zh/seiten'

export const sprachen = [
  'de',
  'en',
  'es',
  'fr',
  'it',
  'ja',
  'ru',
  'zh',
] as const

export type Sprache = (typeof sprachen)[number]

// Sprachstand sagt, wie weit eine Sprache ist — LOC-01.AK-03
// (#1687).
//
// **Vier Stände, und sie sind nicht dasselbe**: ladbar heißt, jeder Schlüssel
// ist da (das prüft `zusatzsprachen.test.ts`); maschinell übersetzt heißt,
// kein Mensch hat gelesen; redaktionell geprüft heißt, ein Mensch hat; und
// abgenommen heißt, jemand hat die Anwendung darin benutzt. **Gleiche
// Schlüssel belegen keine Qualität** — deshalb steht der Stand hier und nicht
// in einer Zählung.
export type Sprachstand =
  'ausgangssprache' | 'maschinell' | 'geprueft' | 'abgenommen'

// sprachstaende ist das Register. **Englisch steht als maschinell übersetzt
// da**, und das ist keine Bescheidenheit: es ist zusammen mit dem Code
// entstanden, und kein Mensch hat es redigiert.
export const sprachstaende: Record<Sprache, Sprachstand> = {
  de: 'ausgangssprache',
  en: 'maschinell',
  es: 'maschinell',
  fr: 'maschinell',
  it: 'maschinell',
  ja: 'maschinell',
  ru: 'maschinell',
  zh: 'maschinell',
}

// Sprachherkunft sagt, woher die Texte einer Sprache kommen und ob ein Mensch
// sie redigiert hat — H04
// (#2258).
//
// **Ein leerer Tag ist keiner**: `geprueft` leer heißt, keine Redaktion fand
// statt, und das Fenster sagt es so. `docs/SPRACHINVENTAR.md` führt dasselbe,
// und `sprachinventar.test.ts` hält beide zusammen.
export interface Sprachherkunft {
  herkunft: 'ausgangssprache' | 'mitDemCode' | 'sprachmodell'

  // uebersetzt ist der Tag der Übersetzung in `JJJJ-MM-TT` — leer, wo es
  // keinen gibt.
  uebersetzt: string

  // geprueft ist der Tag einer redaktionellen Prüfung — leer, wo keine war.
  geprueft: string
}

// sprachherkunft ist das Register dazu. **Deutsch ist auch nicht redigiert**:
// es ist die Sprache, in der dieses Programm zuerst geschrieben wird, und
// gelesen hat es niemand eigens dafür.
export const sprachherkunft: Record<Sprache, Sprachherkunft> = {
  de: { herkunft: 'ausgangssprache', uebersetzt: '', geprueft: '' },
  en: { herkunft: 'mitDemCode', uebersetzt: '', geprueft: '' },
  es: { herkunft: 'sprachmodell', uebersetzt: '2026-09-29', geprueft: '' },
  fr: { herkunft: 'sprachmodell', uebersetzt: '2026-09-29', geprueft: '' },
  it: { herkunft: 'sprachmodell', uebersetzt: '2026-09-29', geprueft: '' },
  ja: { herkunft: 'sprachmodell', uebersetzt: '2026-09-29', geprueft: '' },
  ru: { herkunft: 'sprachmodell', uebersetzt: '2026-09-29', geprueft: '' },
  zh: { herkunft: 'sprachmodell', uebersetzt: '2026-09-29', geprueft: '' },
}

// bcp47 ist die Kennung einer Sprache nach BCP 47 — für `Intl` und für
// `lang`. **Chinesisch ist vereinfacht** (zh-Hans), entschieden in
// `docs/SPRACHINVENTAR.md`.
export const bcp47: Record<Sprache, string> = {
  de: 'de',
  en: 'en',
  es: 'es',
  fr: 'fr',
  it: 'it',
  ja: 'ja',
  ru: 'ru',
  zh: 'zh-Hans',
}

// Texte<T> übernimmt die **Struktur** eines Katalogs, nicht seine Werte.
//
// Ohne diese Abbildung wäre der deutsche Katalog wegen `as const` ein Verbund
// aus Literaltypen — `bezeichnung: 'Sprache'` —, und kein englischer Satz könnte
// ihn je erfüllen: `'Language'` ist nicht `'Sprache'`. Der Typ hielte dann nicht
// die Struktur, sondern die deutsche Fassung, und das ist etwas anderes.
type Texte<T> = {
  [Schluessel in keyof T]: T[Schluessel] extends string
    ? string
    : Texte<T[Schluessel]>
}

export interface Katalog {
  allgemein: Texte<typeof deAllgemein>
  bausteine: Texte<typeof deBausteine>
  fehler: Texte<typeof deFehler>
  hilfe: Texte<typeof deHilfe>
  seiten: Texte<typeof deSeiten>
}

export const kataloge: Record<Sprache, Katalog> = {
  de: {
    allgemein: deAllgemein,
    bausteine: deBausteine,
    fehler: deFehler,
    hilfe: deHilfe,
    seiten: deSeiten,
  },
  en: {
    allgemein: enAllgemein,
    bausteine: enBausteine,
    fehler: enFehler,
    hilfe: enHilfe,
    seiten: enSeiten,
  },
  es: {
    allgemein: esAllgemein,
    bausteine: esBausteine,
    fehler: esFehler,
    hilfe: esHilfe,
    seiten: esSeiten,
  },
  fr: {
    allgemein: frAllgemein,
    bausteine: frBausteine,
    fehler: frFehler,
    hilfe: frHilfe,
    seiten: frSeiten,
  },
  it: {
    allgemein: itAllgemein,
    bausteine: itBausteine,
    fehler: itFehler,
    hilfe: itHilfe,
    seiten: itSeiten,
  },
  ja: {
    allgemein: jaAllgemein,
    bausteine: jaBausteine,
    fehler: jaFehler,
    hilfe: jaHilfe,
    seiten: jaSeiten,
  },
  ru: {
    allgemein: ruAllgemein,
    bausteine: ruBausteine,
    fehler: ruFehler,
    hilfe: ruHilfe,
    seiten: ruSeiten,
  },
  zh: {
    allgemein: zhAllgemein,
    bausteine: zhBausteine,
    fehler: zhFehler,
    hilfe: zhHilfe,
    seiten: zhSeiten,
  },
}
