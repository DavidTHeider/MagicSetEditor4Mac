// **Ein Zweig des Vorlagenbaums** — ein Paket mit seinen Dateien, seinem
// Beiwerk und dem, was an ihm zu sagen ist (Werkstück U7 von Phase 5.1,
// #1418).
//
// **Eine eigene Datei seit UXR-18**
// (#1926):
// mit der Faltung — die Zeile vor dem Beiwerk und „Weitere zeigen" — ging
// `Vorlagenbaum.tsx` über die Zeilengrenze aus §2. Was gezeichnet wird,
// rechnet `vorlagenbaumfaltung.ts`; hier steht, wie es aussieht. Tabstelle
// und Pfeile bringt der Baum über `griffe` mit.
import { useTexte } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { Kopieangebot, Weiterknopf, knopfklassen } from './Vorlagenbaumteile'

import type { KeyboardEvent, MouseEvent } from 'react'
import type { Paketkopie } from '../haken/usePaketkopie'
import type { Dateiwahl } from './Vorlagenbaum'
import type { Zweigplan } from './vorlagenbaumfaltung'

// Zeilengriffe ist, was jeder Knopf des Baums vom Baum bekommt: die eine
// Tabstelle, den Halt beim Fokus, die Pfeile — und den Fokus nach „Weitere
// zeigen".
export interface Zeilengriffe {
  // **Ein Knoten des Baums** — die Pfeile laufen nur über ihn, nicht über
  // eine Kopieaktion daneben (WK02, #2103).
  'data-baumknoten': true
  tabIndex: 0 | -1
  ref: ((knopf: HTMLButtonElement | null) => void) | undefined
  onFocus: () => void
  onKeyDown: (ereignis: KeyboardEvent<HTMLButtonElement>) => void
}

interface Zweigeigenschaften {
  plan: Zweigplan
  gewaehlt: Dateiwahl | null
  geaendert: boolean
  kopie: Pick<Paketkopie, 'kopieren' | 'laeuft' | 'fehler'> | undefined
  griffe: (schluessel: string) => Zeilengriffe
  waehlen: (paket: string, datei: string, beiwerk?: boolean) => void

  // klappen nimmt den Schlüssel der Zeile und ihren Stand vor dem Klick.
  klappen: (schluessel: string, offen: boolean) => void

  // weiter öffnet das nächste Stück einer Liste; erster ist der Schlüssel
  // des ersten neuen Eintrags, der den Fokus bekommt.
  weiter: (liste: string, gezeigt: number, erster: string) => void

  // dateianstoss nimmt Rechtsklick und ⇧F10 an einer Datei entgegen — AP-5
  // (#2362); die Werkstatt wählt sie und öffnet dann ihre Handgriffe.
  dateianstoss: Dateianstoss
}

export type Dateianstoss = (
  ereignis: MouseEvent<HTMLElement> | KeyboardEvent<HTMLElement>,
  paket: string,
  datei: string,
  beiwerk: boolean,
) => void

// dateischluessel benennt eine Datei im Baum — Text wie Beiwerk.
export function dateischluessel(paket: string, datei: string): string {
  return `d:${paket}/${datei}`
}

export function Vorlagenzweig({
  plan,
  gewaehlt,
  geaendert,
  kopie,
  griffe,
  waehlen,
  klappen,
  weiter,
  dateianstoss,
}: Zweigeigenschaften) {
  const texte = useTexte().seiten.vorlagen
  const { zweig, offen, naechsteDatei } = plan
  const { paket } = zweig
  const ist = (datei: string) =>
    gewaehlt?.paket === paket && gewaehlt.datei === datei

  // dateigriffe sind die Griffe des Baums, dazu Rechtsklick und ⇧F10.
  const dateigriffe = (datei: string, beiwerk: boolean) => {
    const baum = griffe(dateischluessel(paket, datei))

    return {
      ...baum,
      onContextMenu: (ereignis: MouseEvent<HTMLButtonElement>) => {
        dateianstoss(ereignis, paket, datei, beiwerk)
      },
      onKeyDown: (ereignis: KeyboardEvent<HTMLButtonElement>) => {
        dateianstoss(ereignis, paket, datei, beiwerk)
        if (!ereignis.defaultPrevented) {
          baum.onKeyDown(ereignis)
        }
      },
    }
  }

  const beiwerkknoepfe = plan.beiwerk.map((datei) => (
    <li key={datei}>
      <button
        type="button"
        aria-current={ist(datei) ? 'true' : undefined}
        aria-label={`${datei} — ${texte.beiwerk}`}
        {...dateigriffe(datei, true)}
        onClick={() => {
          waehlen(paket, datei, true)
        }}
        className={[
          knopfklassen,
          'font-mono text-xs text-leise',
          ist(datei) ? 'bg-betont' : '',
        ].join(' ')}
      >
        <span aria-hidden="true" className="mr-1">
          {'▣'}
        </span>
        {datei}
      </button>
    </li>
  ))
  const { naechstesBild } = plan
  const mehrBeiwerk = naechstesBild !== undefined && (
    <Weiterknopf
      gezeigt={plan.beiwerk.length}
      gesamt={zweig.beiwerk.length}
      griffe={griffe(`w:b:${paket}`)}
      aufKlick={() => {
        weiter(
          `b:${paket}`,
          plan.beiwerk.length,
          dateischluessel(paket, naechstesBild),
        )
      }}
    />
  )

  return (
    <li className="flex flex-col gap-0.5">
      <button
        type="button"
        aria-expanded={offen}
        {...griffe(`p:${paket}`)}
        onClick={() => {
          klappen(`p:${paket}`, offen)
        }}
        className={[knopfklassen, 'font-medium'].join(' ')}
      >
        <span aria-hidden="true" className="mr-1 inline-block w-3 text-leise">
          {offen ? '▾' : '▸'}
        </span>
        {zweig.name}
      </button>

      {offen && zweig.fehler !== undefined && (
        <p role="alert" className="pl-6 text-sm text-fehler">
          {zweig.fehler}
        </p>
      )}

      {offen &&
        zweig.nurLesbar &&
        (kopie === undefined ? (
          <p className="pl-6 text-sm text-leise">{texte.nurLesbar}</p>
        ) : (
          <Kopieangebot paket={paket} kopie={kopie} />
        ))}

      {offen &&
        zweig.fehler === undefined &&
        !zweig.nurLesbar &&
        zweig.dateien.length + zweig.beiwerk.length === 0 && (
          <p className="pl-6 text-sm text-leise">{texte.keineDateien}</p>
        )}

      {offen && zweig.dateien.length + zweig.beiwerk.length > 0 && (
        <ul className="flex flex-col gap-0.5 pl-4">
          {plan.dateien.map((datei) => (
            <li key={datei}>
              <button
                type="button"
                aria-current={ist(datei) ? 'true' : undefined}
                {...dateigriffe(datei, false)}
                aria-label={
                  ist(datei) && geaendert
                    ? `${datei} — ${texte.ungespeichert}`
                    : undefined
                }
                onClick={() => {
                  waehlen(paket, datei)
                }}
                className={[
                  knopfklassen,
                  'font-mono text-xs',
                  ist(datei) ? 'bg-betont' : '',
                ].join(' ')}
              >
                {datei}
                {ist(datei) && geaendert && (
                  <span aria-hidden="true" className="ml-1 text-leise">
                    •
                  </span>
                )}
              </button>
            </li>
          ))}
          {naechsteDatei !== undefined && (
            <Weiterknopf
              gezeigt={plan.dateien.length}
              gesamt={zweig.dateien.length}
              griffe={griffe(`w:d:${paket}`)}
              aufKlick={() => {
                weiter(
                  `d:${paket}`,
                  plan.dateien.length,
                  dateischluessel(paket, naechsteDatei),
                )
              }}
            />
          )}

          {/* **Das Beiwerk eines großen Baums steht hinter einer Zeile** —
              in einem kleinen wie bisher in der Reihe (UXR-18). */}
          {plan.beiwerkZeile ? (
            <li className="flex flex-col gap-0.5">
              <button
                type="button"
                aria-expanded={plan.beiwerkOffen}
                {...griffe(`b:${paket}`)}
                onClick={() => {
                  klappen(`b:${paket}`, plan.beiwerkOffen)
                }}
                className={[knopfklassen, 'text-xs text-leise'].join(' ')}
              >
                <span
                  aria-hidden="true"
                  className="mr-1 inline-block w-3 text-leise"
                >
                  {plan.beiwerkOffen ? '▾' : '▸'}
                </span>
                {einsetzen(texte.beiwerkZeile, {
                  anzahl: String(zweig.beiwerk.length),
                })}
              </button>
              {plan.beiwerkOffen && (
                <ul className="flex flex-col gap-0.5 pl-4">
                  {beiwerkknoepfe}
                  {mehrBeiwerk}
                </ul>
              )}
            </li>
          ) : (
            <>
              {beiwerkknoepfe}
              {mehrBeiwerk}
            </>
          )}
        </ul>
      )}
    </li>
  )
}
