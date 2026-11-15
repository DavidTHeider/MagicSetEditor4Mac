// **Was der Markttisch mit einer gezogenen Karte tut** — P12
// (#2380),
// Befunde C-14, C-18; Entwurf `Markttisch09`.
//
// # Zwei Fächer, zwei Hälften
//
// **Links das eigene Fach** — Sammlung und Wunschliste —, **rechts das
// freigegebene Fach des Gegenübers**, in der Mitte „Ich gebe" und „Ich
// bekomme". Ein eigenes Stück geht nur nach „Ich gebe"; ein fremdes und ein
// eigener Wunsch nur nach „Ich bekomme". **Eine gemischte Wahl hat kein Ziel**
// und wird nicht halb gelegt.
//
// # Was eine Zeile wird
//
// **Ein Druck und keine Besitzzeile** (wie überall im Handel): welches
// Exemplar weggeht, sagt „Zusagen". Dieselbe Karte ein zweites Mal zählt die
// Zeile hoch — bis zu der Menge, die im Fach liegt: **über alle Fachzeilen
// derselben Variante** (#2497), denn ein Wunsch und ein Angebot des
// Gegenübers mit je einem Stück sind zwei.
import type { ExemplarDTO } from '../api'
import type { Handelsposition } from '../apihandelsleitung'
import type { Sichtkarte } from '../apisichtfreigabe'

export type Fachquelle = 'mein' | 'wunsch' | 'sein'

export type Ablage = 'gebe' | 'bekomme'

// Fachkarte ist eine Karte in einem Fach.
export interface Fachkarte {
  kennung: string
  quelle: Fachquelle
  printing: string
  druckId: number
  name: string
  zustand: string
  sprache: string
  finish: string
  menge: number
}

export function ausExemplar(exemplar: ExemplarDTO): Fachkarte {
  const quelle: Fachquelle = exemplar.wunsch ? 'wunsch' : 'mein'

  return {
    kennung: `${quelle}:${String(exemplar.id)}`,
    quelle,
    printing: exemplar.printing,
    druckId: Number(exemplar.druckID) || 0,
    name: exemplar.bezeichnung,
    zustand: exemplar.zustand,
    sprache: exemplar.sprache,
    finish: exemplar.finish,
    menge: exemplar.menge,
  }
}

export function ausSichtkarte(karte: Sichtkarte): Fachkarte {
  return {
    kennung: `sein:${karte.referenz}`,
    quelle: 'sein',
    printing: karte.printing,
    druckId: 0,
    name: karte.bezeichnung,
    zustand: karte.zustand,
    sprache: karte.sprache,
    finish: karte.finish,
    menge: karte.menge,
  }
}

// zieleFuer sagt, wohin eine Wahl darf — leer, wenn sie gemischt ist.
export function zieleFuer(karten: readonly Fachkarte[]): Ablage[] {
  const ziele = new Set(
    karten.map((karte) => (karte.quelle === 'mein' ? 'gebe' : 'bekomme')),
  )

  return ziele.size === 1 ? [...ziele] : []
}

// legen trägt eine Wahl in die Zeilen ein — **unverändert**, wenn sie das
// Ziel nicht hat. `fach` ist alles, was in den Fächern liegt; daran misst
// sich, wie weit eine Zeile wachsen darf.
export function legen(
  positionen: Handelsposition[],
  karten: readonly Fachkarte[],
  ziel: Ablage,
  ich: string,
  gegenueber: string,
  fach: readonly Fachkarte[] = karten,
): Handelsposition[] {
  if (!zieleFuer(karten).includes(ziel)) {
    return positionen
  }

  const von = ziel === 'gebe' ? ich : gegenueber
  let neu = positionen

  for (const karte of karten) {
    const stelle = neu.findIndex((zeile) => liegtIn(zeile, karte, von))
    // **Was im Fach von dieser Variante liegt**, auf dieser Seite — die
    // gelegte Karte mindestens, auch wenn ein Aufrufer sie nicht nennt.
    const vorrat = Math.max(
      karte.menge,
      fach
        .filter(
          (eine) =>
            zieleFuer([eine]).includes(ziel) && gleicheVariante(eine, karte),
        )
        .reduce((summe, eine) => summe + eine.menge, 0),
    )

    neu =
      stelle < 0
        ? [...neu, zeileAus(karte, von, druckIdVon(karte))]
        : neu.map((zeile, hier) =>
            hier === stelle
              ? // **Hoch bis zum Vorrat, nie herunter**: eine Menge, die
                // ein Mensch schon höher gesetzt hat, bleibt.
                {
                  ...zeile,
                  menge: Math.max(
                    zeile.menge,
                    Math.min(zeile.menge + 1, vorrat),
                  ),
                }
              : zeile,
          )
  }

  return neu
}

// druckIdVon nennt die Druckzeile einer Karte — **ein eigener Druck ohne
// Scryfall-Kennung geht über seine Zeile**, nur auf der eigenen Seite, denn
// nur dort ist sie eine.
function druckIdVon(karte: Fachkarte): number {
  return karte.printing === '' ? karte.druckId : 0
}

// liegtIn sagt, ob eine Karte in dieser Zeile liegt. **Eine Zeile ist eine
// Variante** — Druck, Zustand, Sprache und Ausführung (Befund 3 des
// Zweitblicks über Phase 8): ein englisches Foil desselben Drucks ist eine
// andere Vereinbarung als ein deutsches.
function liegtIn(zeile: Handelsposition, karte: Fachkarte, von: string) {
  return (
    zeile.von === von &&
    zeile.printing === karte.printing &&
    zeile.druckId === druckIdVon(karte) &&
    zeile.guete === karte.zustand &&
    zeile.sprache === karte.sprache &&
    zeile.finish === karte.finish
  )
}

function gleicheVariante(eine: Fachkarte, andere: Fachkarte) {
  return (
    eine.printing === andere.printing &&
    druckIdVon(eine) === druckIdVon(andere) &&
    eine.zustand === andere.zustand &&
    eine.sprache === andere.sprache &&
    eine.finish === andere.finish
  )
}

function zeileAus(
  karte: Fachkarte,
  von: string,
  druckId: number,
): Handelsposition {
  return {
    von,
    printing: karte.printing,
    druckId,
    menge: 1,
    guete: karte.zustand,
    sprache: karte.sprache,
    finish: karte.finish,
    abbild: '',
    bewertung: 0,
  }
}
