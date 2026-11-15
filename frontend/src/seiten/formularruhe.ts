// **Wonach das Formular seine Felder ordnet** — die Flächen der Vorschau
// (#2301).
//
// Die Felder stehen in der Lage, in der der Stil sie zeichnet (#674), und
// die kennt erst die Vorschau. Bis zum 2026-10-11 ordnete das Formular vor
// ihr nach dem, was die Karte führt, und kam sie, kam ein Bildfeld dazu: am
// nativen Fenster sprang das Textfeld beim ersten Öffnen um 118 Punkte, und
// ein Klick in diesen 50 ms traf den falschen Knopf.
import { useEffect } from 'react'

import type { FeldflaecheDTO, VorschauDTO } from '../formen'
import { leseordnung } from './angebotenefelder'
import type { Ansichtswahl } from './vorschauansicht'

export interface Feldordnung {
  // flaechen sind die, nach denen das Formular ordnet: die des Stils, wie sie
  // zuerst gesehen wurden, ergänzt um die Felder, die diese Karte neu bringt.
  flaechen: readonly FeldflaecheDTO[]

  // reihenfolge ist die Ordnung dieser Felder, wie sie einmal stand — oder
  // `undefined`, solange der Stil noch keine Vorschau hatte.
  reihenfolge: readonly string[] | undefined

  // unruhig sagt, dass für diesen Stil noch keine Vorschau Flächen gemeldet
  // hat und die erste noch lädt: dann steht das Formular unsichtbar da.
  // **Scheitert sie, zeigt es sich** — ohne Flächen, aber bedienbar.
  unruhig: boolean
}

// useFeldordnung hält eine Ordnung **je Stil** fest, in der Ansicht über dem
// Editor, die den Kartenwechsel übersteht.
//
// **Gemessen, bevor entschieden** (#2309): im Kanon ändert sich die
// Leseordnung zweier aufeinanderfolgender Karten desselben Stils in 601 von
// 1.033 Schritten (`internal/app/feldordnungkanon_intern_test.go`), und die
// eigene Vorschau braucht beim Blättern im Median 129 ms
// (`docs/belege/huellenprobe/blaettern.md`). Die Flächen der vorigen Karte
// zu übernehmen ließ das Formular in zwei von drei Schritten umspringen; auf
// die eigenen zu warten ließ es bei jedem leer stehen. **Die Reihenfolge, in
// der die Felder eines Stils zuerst standen, bleibt**, auch beim Tippen; ein
// Feld, das eine Karte neu bringt, reiht sich nach seiner eigenen Lage ein.
//
// **Gelernt wird unter dem Stil, mit dem die Vorschau gezeichnet ist**
// (`gezeigt.stil`, Zweitblick über die Folgen vor 0.8.0, Befund 8): beim
// Stilwechsel einer Karte steht die vorige Vorschau noch da, und eine
// Probevorschau zeigt einen Stil, der nur getippt ist. **Geordnet wird nach
// dem geschriebenen** (`stil`).
//
// **Was das nicht hält**: dass die Ordnung zur Lage **dieser** Karte passt.
// Sie passt zu der, die der Stil zuerst zeigte.
export function useFeldordnung(
  [ansicht, setzeAnsicht]: Ansichtswahl,
  gezeigt:
    | (Pick<VorschauDTO, 'flaechen' | 'stil'> &
        Partial<Pick<VorschauDTO, 'gezeichnet'>>)
    | undefined,
  laedt: boolean,
  stil: string,
): Feldordnung {
  useEffect(() => {
    if (gezeigt === undefined) return

    const unter = paketVon(gezeigt)

    setzeAnsicht((vorher) => {
      const bekannt = vorher.flaechenJeStil?.[unter]
      const vereint = vereine(bekannt, gezeigt.flaechen)
      const zugeordnet = vorher.gezeichnetJeStil?.[gezeigt.stil] === unter
      if (vereint === bekannt && zugeordnet) return vorher

      return {
        ...vorher,
        gezeichnetJeStil: { ...vorher.gezeichnetJeStil, [gezeigt.stil]: unter },
        flaechenJeStil: { ...vorher.flaechenJeStil, [unter]: vereint },
        ordnungJeStil: {
          ...vorher.ordnungJeStil,
          [unter]: einreihen(
            vorher.ordnungJeStil?.[unter],
            leseordnung(vereint),
          ),
        },
      }
    })
  }, [gezeigt, setzeAnsicht])

  // **Unter welchem Paket der geschriebene Stil gerade steht** (#2314): die
  // Vorschau sagt es, wenn sie zu ihm gehört; sonst das, was sie zuletzt
  // dafür sagte.
  const unter =
    gezeigt?.stil === stil
      ? paketVon(gezeigt)
      : (ansicht.gezeichnetJeStil?.[stil] ?? stil)
  const bekannt = ansicht.flaechenJeStil?.[unter]
  const flaechen =
    gezeigt?.stil === stil ? vereine(bekannt, gezeigt.flaechen) : bekannt

  return {
    flaechen: flaechen ?? [],
    reihenfolge:
      flaechen === undefined
        ? undefined
        : einreihen(ansicht.ordnungJeStil?.[unter], leseordnung(flaechen)),
    unruhig: flaechen === undefined && laedt,
  }
}

// paketVon ist der Schlüssel, unter dem eine Vorschau lernt: das Paket, mit
// dem sie gezeichnet ist — und der geschriebene Stil, wo es keines nennt
// (ein Behelfsplan, oder eine Vorschau von vor #2314).
function paketVon(
  gezeigt: Pick<VorschauDTO, 'stil'> & Partial<Pick<VorschauDTO, 'gezeichnet'>>,
): string {
  return gezeigt.gezeichnet === undefined || gezeigt.gezeichnet === ''
    ? gezeigt.stil
    : gezeigt.gezeichnet
}

// einreihen setzt die Felder einer neuen Leseordnung in eine alte, **ohne
// die alte umzustellen**: ein bekanntes Feld nimmt den nächsten Platz der
// alten Folge, ein neues den, den die neue ihm gibt.
function einreihen(
  alt: readonly string[] | undefined,
  neu: readonly string[],
): readonly string[] {
  if (alt === undefined) return neu

  const bekannt = new Set(alt)
  const folge = alt.filter((feld) => neu.includes(feld))
  let naechstes = 0

  return neu.map((feld) => {
    if (!bekannt.has(feld)) return feld

    /* v8 ignore next -- `leseordnung` nennt jedes Feld einmal, also hat die
       alte Folge so viele bekannte wie die neue; TypeScript sieht es nicht. */
    return folge[naechstes++] ?? feld
  })
}
// vereine legt die Flächen eines Feldes, das der Stil noch nicht kannte, zu
// den bekannten — und gibt **dieselben** zurück, wenn nichts neu ist, damit
// die Ansicht nicht bei jeder Vorschau einen neuen Stand bekommt.
function vereine(
  bekannt: readonly FeldflaecheDTO[] | undefined,
  gezeigt: readonly FeldflaecheDTO[],
): readonly FeldflaecheDTO[] {
  if (bekannt === undefined) return gezeigt

  const felder = new Set(bekannt.map((flaeche) => flaeche.feld))
  const neu = gezeigt.filter((flaeche) => !felder.has(flaeche.feld))

  return neu.length === 0 ? bekannt : [...bekannt, ...neu]
}
