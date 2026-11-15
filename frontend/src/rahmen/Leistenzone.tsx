// **Eine Zone der Seitenleiste** — UX-02
// (#1674),
// gezeichnet im Artboard `Main` des Entwurfs.
//
// # Warum ein eigenes Bauteil
//
// **Weil `Seitenleiste.tsx` mit ihm über die Zeilengrenze ginge** (§2). Der
// Schnitt folgt der Frage: dort steht, **was die Leiste ist** — Kopf,
// Kontextleiste, Fuß mit der Befehlspalette; hier, **wie eine Zone darin
// aussieht**.
//
// # Und warum die Überschrift kein Ziel ist
//
// **Weil sie keine Seite benennt, sondern eine Gruppe.** *Bereiche* und
// *Programm* sind Aufschriften über einer Liste; was man anklickt, sind die
// Einträge darunter. Eine klickbare Überschrift wäre ein Ziel, das die
// rovierende Zeile mitzählen müsste, ohne dass jemand es erwartet.
import { bereichFuer } from '../bereiche'
import { beschriftung } from '../navigation'
import { leistentaste } from './leistentaste'
import { useTexte } from '../sprache/kontext'

import type { Leisteneintrag } from '../leistenordnung'
import type { RovierendeZeile } from '../haken/useRovierendeZeile'
import type { Seitenschluessel } from '../navigation'

// **Die Klassen stehen benannt und nicht als Ausdruck im JSX** — dieselbe
// Regel wie in `Seitenleiste.tsx`: `keintextimjsx.test.ts` kann einem
// Klassennamen nicht ansehen, dass er keine Aufschrift ist.
const zonenkopf =
  'px-3 pt-4 pb-1 text-[0.625rem] font-medium uppercase tracking-[0.09em] text-leise'

const trennlinie = 'mx-3 mt-3 border-t border-kante'

// **Oben steht dieselbe Zone quer** — UX-14.AK-02
// (#1686).
// Die Überschrift bleibt für den Vorleser da und verschwindet fürs Auge, die
// Trennlinie steht senkrecht, und die Liste läuft als Zeile.
const zonenkopfQuer = 'sr-only'
const trennlinieQuer = 'mx-2 my-2 border-l border-kante'
const listeSenkrecht = 'flex flex-col gap-0.5'
const listeQuer = 'flex flex-row flex-wrap items-center gap-0.5'

// **Der aufgeschlagene Eintrag trägt drei Zeichen, nicht eines** — und das
// dritte ist das einzige, das die Prüfung besteht.
//
// # Warum Grund und Schrift nicht genügen
//
// `bg-betont` und `font-medium` allein: **Gemessen am 2026-09-21 an der
// laufenden Anwendung** steht der hervorgehobene Grund im dunklen
// Erscheinungsbild mit **1,13 : 1** gegen die Fläche der Leiste — WCAG 1.4.11
// verlangt 3 : 1 für den Zustand eines Bedienelements. Der Unterschied ist
// da, und man sieht ihn nicht.
//
// **Der Entwurf löst es mit einem Balken**: `box-shadow: inset 2px 0 0
// var(--marke)` (`prototyp/project/mse.css`, Zeile 94). Hier steht ein
// **Rand** statt eines Schattens — Vermerk 0065, Festlegung 3 —, und weil er
// bei jedem Eintrag durchsichtig danebensteht, rückt nichts, wenn er
// erscheint. `border-marke` auf `bg-flaeche` ist gemessen: 6,67 hell und 6,14
// dunkel ([`KONTRASTE.md`](../../../docs/KONTRASTE.md)).
//
// **Zweiunddreißig Punkte hoch, wie im Artboard** (`.nav-eintrag`: 7 Punkte
// Rand um eine Zeile von 13) — UXR-02 (#1910). Mit `py-2` um eine Zeile von 24
// waren es 40, und die Leiste rollte im Mindestfenster: *Hilfe* stand unter
// ihrem Rand (`leistenmass.test.tsx`). **Eine feste Höhe und kein Rand**: mit
// `py-1.5` maß das native Fenster 37 statt 32 (`uxr02`, 2026-10-02): die
// Zeile ist dort höher, als `text-sm` verspricht — vermutlich, weil die
// Sinnbilder aus einer anderen Schrift kommen; nachgemessen ist nur die Höhe.
const eintragsklassen =
  'flex h-8 w-full items-center gap-2 rounded-md border-l-2 border-transparent px-3 text-left text-sm hover:bg-betont focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring aria-[current=page]:border-marke aria-[current=page]:bg-betont aria-[current=page]:font-medium'

// **Das Sinnbild trägt die Marke** — so zeichnet es der Entwurf
// (`.nav-eintrag .sym { color: var(--marke) }`, Zeile 95). `text-marke` auf
// `bg-flaeche` ist gemessen: 6,67 hell und 6,14 dunkel.
// **Oben trägt die aufgeschlagene Seite den Balken unten** statt links —
// derselbe Rand, dieselbe gemessene Farbe, nur an der Kante, die zur Zeile
// passt.
const eintragsklassenQuer =
  'flex items-center gap-2 whitespace-nowrap rounded-md border-b-2 border-transparent px-3 py-2 text-left hover:bg-betont focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring aria-[current=page]:border-marke aria-[current=page]:bg-betont aria-[current=page]:font-medium'

const sinnbildklassen = 'w-4 shrink-0 text-center text-marke'

const kuerzelklassen = 'shrink-0 text-xs tabular-nums text-leise'

const markenklassen =
  'shrink-0 rounded-full border border-marke px-1.5 text-xs text-marke'

export interface Leistenzoneneigenschaften {
  eintraege: readonly Leisteneintrag[]

  // ueberschrift steht über der Liste.
  ueberschrift: string

  // trenner sagt, ob eine Linie über der Überschrift steht.
  //
  // **Die Dienstzone bekommt sie** — UX-02.AK-04: sie ist kein fachlicher
  // Bereich, und ohne die Linie sähe sie wie einer aus.
  trenner: boolean

  // versatz ist die Stelle des ersten Eintrags in der ganzen Leiste.
  //
  // **Über die ganze und nicht je Zone**: ↑ und ↓ wandern quer durch die
  // Überschriften hindurch, weil eine Überschrift kein Ziel ist.
  versatz: number

  // gesamt ist die Zahl der Einträge in der ganzen Leiste.
  gesamt: number

  aktuell: Seitenschluessel
  wechseln: (seite: Seitenschluessel) => void
  halt: RovierendeZeile

  // waagerecht legt die Zone quer, für die Leiste oben (UX-14.AK-02).
  waagerecht?: boolean

  // reife ist je Schalter seine Reife — eine, die nicht *stabil* ist, steht
  // als Marke am Bereich (OPS-10.AK-04, #1908).
  reife?: Readonly<Record<string, string>>
}

export function Leistenzone({
  eintraege,
  ueberschrift,
  trenner,
  versatz,
  gesamt,
  aktuell,
  wechseln,
  halt,
  waagerecht = false,
  reife = {},
}: Leistenzoneneigenschaften) {
  const texte = useTexte()
  const namen = texte.allgemein.bereiche
  const name = (eintrag: Leisteneintrag) =>
    eintrag.alsBereich
      ? namen[eintrag.bereich.schluessel]
      : beschriftung(texte, eintrag.ziel)
  const reifen: Record<string, string | undefined> =
    texte.seiten.einstellungen.schalter.reife
  const marke = (eintrag: Leisteneintrag) => {
    const stufe = eintrag.alsBereich
      ? reife[eintrag.bereich.schluessel]
      : undefined

    return stufe === undefined || stufe === 'stabil'
      ? undefined
      : (reifen[stufe] ?? stufe)
  }

  return (
    <>
      {trenner && <div className={waagerecht ? trennlinieQuer : trennlinie} />}

      <div className={waagerecht ? zonenkopfQuer : zonenkopf}>
        {ueberschrift}
      </div>

      <ul className={waagerecht ? listeQuer : listeSenkrecht}>
        {eintraege.map((eintrag, lauf) => (
          <li key={eintrag.ziel}>
            <button
              type="button"
              title={
                waagerecht && eintrag.kuerzel !== ''
                  ? `⌘${eintrag.kuerzel}`
                  : undefined
              }
              aria-current={hier(eintrag, aktuell) ? 'page' : undefined}
              tabIndex={halt.annehmen(versatz + lauf) ? 0 : -1}
              onKeyDown={(ereignis) => {
                leistentaste(
                  ereignis,
                  versatz + lauf,
                  gesamt,
                  halt.rueckenAuf,
                  waagerecht,
                )
              }}
              onClick={() => {
                wechseln(eintrag.ziel)
              }}
              className={waagerecht ? eintragsklassenQuer : eintragsklassen}
            >
              <span aria-hidden="true" className={sinnbildklassen}>
                {eintrag.sinnbild}
              </span>

              {/* **Gekürzt wird die Zeichnung, nicht der Name** — H03
                  (#2257): „Nutzer und Profile" stand auf Französisch,
                  Japanisch und Russisch mit Auslassungszeichen da. Umbrechen
                  geht nicht, die Zeile ist 32 Punkte hoch; der Titel trägt
                  den vollen Namen. Oben bricht die Zeile um, und dort steht
                  im Titel das Kürzel. */}
              <span
                className="flex-1 truncate"
                title={waagerecht ? undefined : name(eintrag)}
              >
                {name(eintrag)}
              </span>

              {marke(eintrag) !== undefined && (
                <span className={markenklassen}>{marke(eintrag)}</span>
              )}

              {/* **Das Kürzel steht dabei, statt gelernt werden zu müssen.**
                  `aria-hidden`, weil es die Taste zeigt und nicht den Namen. */}
              {/* **Ohne Ziffer kein Zeichen** — seit die Belegung einstellbar
                  ist (UX-02.AK-03), hat nicht jeder Bereich eine. */}
              {/* **Oben steht es im Titel** (UX-14.AK-02): die Zeile braucht
                  ihre Breite für die Namen. */}
              {!waagerecht && eintrag.kuerzel !== '' && (
                <span aria-hidden="true" className={kuerzelklassen}>
                  ⌘{eintrag.kuerzel}
                </span>
              )}
            </button>
          </li>
        ))}
      </ul>
    </>
  )
}

// hier sagt, ob ein Eintrag die aufgeschlagene Fläche benennt.
//
// **Zwei Fragen und nicht eine.** Ein Bereichseintrag leuchtet, sobald
// *irgendeine* seiner Seiten offen ist — auch eine Lasche des Sets, die gar
// nicht in der Leiste steht. Ein Diensteintrag leuchtet nur für seine eigene
// Seite: alle drei teilen sich einen Bereich, und die Bereichsfrage ließe
// alle drei zugleich leuchten.
function hier(eintrag: Leisteneintrag, aktuell: Seitenschluessel): boolean {
  return eintrag.alsBereich
    ? bereichFuer(aktuell)?.schluessel === eintrag.bereich.schluessel
    : aktuell === eintrag.ziel
}
