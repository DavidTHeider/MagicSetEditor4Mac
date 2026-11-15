// **Eine Liste, zwei Gestalten** — UX-11
// (#1683):
// Sets und Decks bekommen, was die Sammlung schon hatte — umschaltbare
// Ansichten mit Suche und Sortierung.
//
// # Warum ein Bauteil für beide
//
// **Damit keine Galerie ihre eigene Fachlogik hat** (AK-02). Was ein Eintrag
// kann — öffnen, umbenennen, löschen —, zeichnet **derselbe** Aufruf in beiden
// Gestalten (`eintrag(e, gestalt)`); die Gestalt ändert die Anordnung, nicht
// die Handlung und nicht die Rechte.
//
// # Was ein Ansichtswechsel behält (AK-01)
//
// **Suche, Sortierung und die gezeigten Seiten** stehen hier und nicht in der
// Gestalt — ein Wechsel lässt sie stehen. **Und den Eintrag, an dem man war**:
// wer in einem Eintrag den Fokus hatte, hat ihn nach dem Wechsel an demselben,
// und er steht im Blick.
//
// # Große Bestände (AK-03)
//
// **Seitenweise**: sechzig Einträge, dann ein Knopf für die nächsten. Eine
// Liste von tausend Sets zeichnet sonst tausend Kacheln mit je fünf Knöpfen,
// bevor man die erste sieht.
import { useEffect, useId, useRef, useState, type ReactNode } from 'react'

import type { Markenart } from '../apiobjektmarken'
import { useMarkenfilter } from '../haken/useMarkenfilter'
import { Knopf } from './Knopf'
import { Markenfilterknopf } from './Markenfilterkasten'
import { Markenzeile } from './Markenzeile'
import { Segmente } from './Segmente'
import { einsetzen } from '../sprache/einsetzen'
import { useSprache, useTexte } from '../sprache/kontext'
import { falte } from '../sprache/vergleich'

// Gestalt ist die Anordnung eines Eintrags.
export type Gestalt = 'zeile' | 'kachel'

// Listenform ist die gemerkte Ansicht eines Bereichs.
export type Listenform = 'liste' | 'kacheln'

export interface Sortierung<T> {
  name: string
  bezeichnung: string
  vergleich: (eines: T, anderes: T) => number
}

export const seitengroesse = 60

export function Listenansicht<T>({
  bezeichnung,
  eintraege,
  schluessel,
  suchtext,
  sortierungen,
  form,
  setzeForm,
  eintrag,
  leer,
  marken,
  festgehalten,
}: {
  bezeichnung: string
  eintraege: readonly T[]
  schluessel: (eintrag: T) => string
  suchtext: (eintrag: T) => string
  sortierungen: readonly Sortierung<T>[]
  form: Listenform
  setzeForm: (form: Listenform) => void
  eintrag: (eintrag: T, gestalt: Gestalt) => ReactNode
  leer: ReactNode

  // marken gibt der Liste Marken und einen Filter nach ihnen — UX-12
  // (#1684). Der Schlüssel eines Eintrags ist seine Kennung.
  marken?: { art: Markenart; name: (eintrag: T) => string }

  // festgehalten hält Einträge in der Liste, auch wenn Suche, Marke oder Seite
  // sie nicht träfen — #2469: das offene Deck verschwand sonst, und mit ihm
  // seine Handlungen. Wie die gewählte Karte in der Druckwahl.
  festgehalten?: ((eintrag: T) => boolean) | undefined
}) {
  const texte = useTexte()
  const sprache = useSprache()
  const worte = texte.bausteine.listenansicht
  const markenstand = useMarkenfilter(marken?.art)
  const suchID = useId()
  const sortID = useId()
  const ansichtID = useId()
  const [suche, setzeSuche] = useState('')
  const [sortierung, setzeSortierung] = useState(sortierungen[0]?.name ?? '')
  const [seiten, setzeSeiten] = useState(1)
  const zuletzt = useRef<string | undefined>(undefined)
  const liste = useRef<HTMLUListElement>(null)

  // gewechselt sagt, dass die Gestalt **hier** gewechselt wurde — nur dann
  // holt der Effekt darunter den Fokus zurück.
  const gewechselt = useRef(false)

  // **Der Eintrag, an dem man war, bleibt es** — nach dem Wechsel der
  // Gestalt, wenn der Fokus vorher in ihm stand.
  //
  // **Und nur dann**: eine aufgehobene Seite (`<Activity>`) lässt ihre
  // Effekte beim Einblenden neu laufen, und ohne die Marke nahm die Liste
  // dabei den Fokus aus dem Inhalt — gefunden von der WebKit-Probe
  // `ux10-tastatur` am 2026-09-26, die Klasse aus G11 des Zweitblicks über
  // Phase 6.3.
  //
  // **Die Marke wird verbraucht, auch ohne Ziel** (Nachlauf über Phase 6.3,
  // N20): stand der Fokus beim Wechsel in keinem Eintrag, blieb sie sonst
  // liegen, und das nächste Einblenden nahm ihn doch.
  useEffect(() => {
    const ziel = zuletzt.current
    const warGewechselt = gewechselt.current
    gewechselt.current = false
    if (!warGewechselt || ziel === undefined) {
      return
    }

    const wurzel = liste.current
    /* v8 ignore next -- ein Wechsel holt den Fokus nur zurück, wenn er in der
       Liste stand (`gewechselt`), und dann steht sie da. */
    if (wurzel === null) return

    const element = [
      ...wurzel.querySelectorAll<HTMLElement>('[data-eintrag]'),
    ].find((kandidat) => kandidat.dataset.eintrag === ziel)
    element?.scrollIntoView({ block: 'nearest' })
    element
      ?.querySelector<HTMLElement>('button, a[href], input, select')
      ?.focus()
  }, [form])

  if (eintraege.length === 0) {
    return <>{leer}</>
  }

  const frage = falte(suche.trim(), sprache)
  const gewaehlt = sortierungen.find((art) => art.name === sortierung)
  const gehalten = (eintrag: T) => festgehalten?.(eintrag) === true
  const gefunden = eintraege
    .filter(
      (eintrag) =>
        gehalten(eintrag) ||
        ((frage === '' || falte(suchtext(eintrag), sprache).includes(frage)) &&
          (markenstand.getroffen === undefined ||
            markenstand.getroffen.has(schluessel(eintrag)))),
    )
    .sort(gewaehlt?.vergleich ?? (() => 0))
  const gezeigt = [
    ...gefunden.slice(0, seiten * seitengroesse),
    ...gefunden.slice(seiten * seitengroesse).filter(gehalten),
  ]

  return (
    <section aria-label={bezeichnung} className="flex flex-col gap-3">
      <div className="flex flex-wrap items-end gap-3">
        <div className="flex flex-col gap-1">
          <label htmlFor={suchID} className="text-sm font-medium">
            {worte.suche}
          </label>
          <input
            id={suchID}
            type="search"
            value={suche}
            placeholder={worte.suchePlatzhalter}
            onChange={(ereignis) => {
              setzeSuche(ereignis.target.value)
              setzeSeiten(1)
            }}
            className="rounded-md border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          />
        </div>
        {sortierungen.length > 1 && (
          <div className="flex flex-col gap-1">
            <label htmlFor={sortID} className="text-sm font-medium">
              {worte.sortierung}
            </label>
            <select
              id={sortID}
              value={sortierung}
              onChange={(ereignis) => {
                setzeSortierung(ereignis.target.value)
              }}
              className="rounded-md border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            >
              {sortierungen.map((art) => (
                <option key={art.name} value={art.name}>
                  {art.bezeichnung}
                </option>
              ))}
            </select>
          </div>
        )}
        {/* **Ein Segment und kein Paar Knöpfe** — AP-6 (#2363), R7: zwei
            Werte einer Frage sind eine Optionsgruppe mit einer Tabstelle.
            Das Paar zählte zweimal zu den Bedienelementen im Bild (R2). */}
        <div className="flex flex-col gap-1">
          <span id={ansichtID} className="text-sm font-medium">
            {worte.ansicht}
          </span>
          <Segmente
            beschriftetVon={ansichtID}
            optionen={[
              { wert: 'liste', name: worte.liste },
              { wert: 'kacheln', name: worte.kacheln },
            ]}
            wert={form}
            angehalten={false}
            fokusBleibt
            aendern={(wert) => {
              // **Nur wenn der Fokus noch in der Liste stand**, holt der
              // Wechsel ihn in den Eintrag zurück: ein Klick ins Segment
              // fokussiert in WebKit nicht, und der Eintrag wird neu
              // gezeichnet. Wählt jemand mit den Pfeilen, steht der Fokus im
              // Segment — und bleibt dort, sonst ginge der zweite Pfeil ins
              // Leere.
              gewechselt.current =
                liste.current?.contains(document.activeElement) ?? false
              setzeForm(wert === 'kacheln' ? 'kacheln' : 'liste')
            }}
          />
        </div>
        <p role="status" className="text-sm text-leise">
          {einsetzen(worte.zahl, {
            gezeigt: String(gefunden.length),
            gesamt: String(eintraege.length),
          })}
        </p>
      </div>

      {/* **Ohne Marken kein Filter nach Marken** — AP-6 (#2363): er filtert
          nichts, solange nichts eine trägt, und zählte im Bild mit (R2).
          Wirkt er noch, wenn die letzte Marke ging, passt nichts mehr — und
          „Suche leeren“ darunter löst ihn mit. */}
      {marken !== undefined &&
        Object.values(markenstand.je).some((eigene) => eigene.length > 0) && (
          <Markenfilterknopf art={marken.art} stand={markenstand} />
        )}

      {/* **Gefiltert leer ist nicht leer** (AK-04): hier gibt es Einträge,
          nur passt keiner — und der nächste Schritt ist, die Suche zu leeren,
          nicht etwas anzulegen. */}
      {gefunden.length === 0 ? (
        <div className="flex flex-wrap items-center gap-3">
          <p className="text-leise">
            {einsetzen(worte.nichtsPasst, { suche })}
          </p>
          <Knopf
            aufKlick={() => {
              setzeSuche('')
              markenstand.setzeFilter({ ...markenstand.filter, marken: [] })
            }}
          >
            {worte.sucheLeeren}
          </Knopf>
        </div>
      ) : (
        <ul
          ref={liste}
          data-form={form}
          onFocus={(ereignis) => {
            zuletzt.current =
              (ereignis.target as HTMLElement).closest<HTMLElement>(
                '[data-eintrag]',
              )?.dataset.eintrag ?? zuletzt.current
          }}
          // **Eine Kachel ist nie breiter als ihr Platz** (#2019): 16rem im
          // Navigator von 15rem ließ ihn waagerecht rollen — gemessen am
          // nativen Fenster (`huellenprobe/deckkacheln.mjs`).
          className={
            form === 'kacheln'
              ? 'grid grid-cols-[repeat(auto-fill,minmax(min(16rem,100%),1fr))] gap-3'
              : 'flex flex-col gap-2'
          }
        >
          {gezeigt.map((einer) => (
            <li
              key={schluessel(einer)}
              data-eintrag={schluessel(einer)}
              className="group/eintrag"
            >
              {eintrag(einer, form === 'kacheln' ? 'kachel' : 'zeile')}
              {marken !== undefined && (
                <Markenzeile
                  art={marken.art}
                  kennung={schluessel(einer)}
                  name={marken.name(einer)}
                  marken={markenstand.je[schluessel(einer)] ?? []}
                  nachladen={markenstand.nachladen}
                  leise
                />
              )}
            </li>
          ))}
        </ul>
      )}

      {gefunden.length > gezeigt.length && (
        <div>
          <Knopf
            aufKlick={() => {
              setzeSeiten(seiten + 1)
            }}
          >
            {einsetzen(worte.weitere, {
              anzahl: String(
                Math.min(seitengroesse, gefunden.length - gezeigt.length),
              ),
            })}
          </Knopf>
        </div>
      )}
    </section>
  )
}
