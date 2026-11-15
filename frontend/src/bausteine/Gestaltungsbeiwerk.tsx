// **Was unter den Einstellungen des Erscheinungsbilds steht** — UX-09
// (#1681).
//
// **Das Fensterbild mit Vorschau** (AK-01): wie es mit der gewählten
// Skalierung und unter dem Schleier aussieht, bevor jemand die ganze
// Oberfläche danach absucht — und daneben der Weg, es sofort wieder
// wegzunehmen. **Die eigenen Farben mit ihrem Befund** (AK-03, AK-04): eine
// Zeile, die die Form verfehlt, wird an ihrer Stelle genannt und nicht
// geschrieben; eine lesbare Form, die nicht zu lesen wäre, wird geschrieben
// und gilt nicht — und hier steht, warum.
//
// **Seit #2465 in zwei Teilen**: das Fensterbild steht offen unter den
// Einstellungen des Erscheinungsbilds (`Gestaltungsbeiwerk`), die eigenen
// Farben in einem Abschnitt darunter (`Eigenfarben`) — am Fenster rollte die
// Gruppe eng 2,7 Bildschirme.
import { useId, useState } from 'react'

import { Aktionsknopf } from './Aktionsknopf'
import { Einstellungsherkunft } from './Einstellungsherkunft'
import { Knopf } from './Knopf'
import { Satz } from './Satz'
import { eigeneMerkmale, merkmale } from '../gestaltung'
import { useGestaltung } from '../gestaltungsraum'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useTexte } from '../sprache/kontext'

import type { EinstellungDTO } from '../api'
import type { Eigenbildstand } from '../haken/useEigenesBild'

const feldklassen =
  'rounded-md border border-kante bg-flaeche px-2 py-1 font-mono text-sm ' +
  'focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'

// groessen ordnet jeder Stellung die Größe des Bildes zu — dieselbe wie in
// `themen.css`, damit die Vorschau zeigt, was das Fenster zeigen wird.
const groessen = {
  fuellen: ['cover', 'no-repeat'],
  einpassen: ['contain', 'no-repeat'],
  kacheln: ['auto', 'repeat'],
} as const

// groesseFuer nimmt für „aus" und alles Unbekannte die Vorgabe **Füllen** —
// die Vorschau zeigt dann, wie das Bild aussähe, wenn es gälte.
function groesseFuer(modus: string): readonly [string, string] {
  return modus in groessen
    ? groessen[modus as keyof typeof groessen]
    : groessen.fuellen
}

export function Gestaltungsbeiwerk({ hintergrund }: { hintergrund: string }) {
  const stand = useGestaltung()
  const worte = useTexte().seiten.einstellungen.gestaltung

  if (stand === undefined) {
    return null
  }

  return (
    <div className="flex flex-col gap-6">
      <Fensterbildwahl fensterbild={stand.fensterbild} modus={hintergrund} />
      {stand.schmuckNurOben && (
        <p className="max-w-prose text-leise">{worte.schmuckNurOben}</p>
      )}
      {stand.ruecksetzfehler !== undefined && (
        <p role="alert" className="max-w-prose text-fehler">
          {worte.zuruecksetzenGescheitert} {stand.ruecksetzfehler}
        </p>
      )}
    </div>
  )
}

// Eigenfarben ist der Abschnitt „Eigene Farben" — das Feld der Merkmale mit
// seinem Befund (#2465).
export function Eigenfarben({
  gestaltung,
  angehalten,
  setze,
}: {
  gestaltung: EinstellungDTO | undefined
  angehalten: boolean
  setze: (wert: string) => void
}) {
  const stand = useGestaltung()

  if (stand === undefined || gestaltung === undefined) {
    return null
  }

  return (
    <Merkmalfeld
      key={gestaltung.wert}
      eintrag={gestaltung}
      gespeichert={gestaltung.wert}
      befunde={stand.befunde}
      ruht={stand.kontrastHoch}
      angehalten={angehalten}
      setze={setze}
    />
  )
}

function Fensterbildwahl({
  fensterbild,
  modus,
}: {
  fensterbild: Eigenbildstand
  modus: string
}) {
  const worte = useTexte().seiten.einstellungen.gestaltung
  const [groesse, wiederholung] = groesseFuer(modus)
  const adresse =
    fensterbild.adresse === ''
      ? ''
      : `${fensterbild.adresse}?stand=${String(fensterbild.stand)}`

  return (
    <section className="flex flex-col gap-3">
      <h3 className="font-medium">{worte.fensterbild}</h3>
      {adresse === '' ? (
        <p className="max-w-prose text-leise">{worte.keinBild}</p>
      ) : (
        <>
          {/* **Unter demselben Schleier wie im Fenster** — `bg-grund` und
              `color-mix` wie in `themen.css`, damit die Vorschau nicht
              schöner ist als das, was dann dasteht. */}
          <div
            role="img"
            aria-label={worte.vorschau}
            data-vorschau="fensterbild"
            className="h-32 w-56 rounded-md border border-kante bg-grund"
            style={{
              backgroundImage: `linear-gradient(color-mix(in srgb, var(--mse-grund) var(--mse-schleier), transparent), color-mix(in srgb, var(--mse-grund) var(--mse-schleier), transparent)), url("${adresse}")`,
              backgroundSize: `auto, ${groesse}`,
              backgroundRepeat: wiederholung,
              backgroundPosition: 'center',
            }}
          />
          {modus === 'aus' && (
            <p className="max-w-prose text-leise">
              <Satz text={worte.aus} />
            </p>
          )}
        </>
      )}
      <div className="flex flex-wrap gap-2">
        <Knopf
          angehalten={fensterbild.laeuft || !fensterbild.bereit}
          aufKlick={fensterbild.waehlen}
        >
          {worte.waehlen}
        </Knopf>
        {/* **Ohne Rückfrage, wie beim Tischbild** — das Bild lässt sich
            neu wählen (`test/aktionstafel.ts`). */}
        {adresse !== '' && (
          <Aktionsknopf
            aktion="entfernen"
            angehalten={fensterbild.laeuft}
            aufKlick={fensterbild.entfernen}
          >
            {worte.entfernen}
          </Aktionsknopf>
        )}
      </div>
      {fensterbild.fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {fensterbild.fehler}
        </p>
      )}
    </section>
  )
}

function Merkmalfeld({
  eintrag,
  gespeichert,
  befunde,
  ruht,
  angehalten,
  setze,
}: {
  eintrag: EinstellungDTO
  gespeichert: string
  ruht: boolean
  befunde: {
    vorn: string
    hinten: string
    verhaeltnis: number
    schwelle: number
  }[]
  angehalten: boolean
  setze: (wert: string) => void
}) {
  const worte = useTexte().seiten.einstellungen.gestaltung
  const feldID = useId()
  const hilfeID = useId()
  const befundID = useId()
  const [entwurf, setzeEntwurf] = useState(gespeichert)
  const [zeile, setzeZeile] = useState<number>()

  // **Die Zahl in der Gegend des Fensters** (LOC-03): 4,50 oder 4.50.
  const format = new Intl.NumberFormat(useDarstellung(), {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
  const zahl = (wert: number) => format.format(wert)

  return (
    <section className="flex flex-col gap-2">
      {/* **Für das Auge steht der Name am Knopf des Abschnitts** (#2465);
          das Feld behält ihn, damit es ohne Blick benannt ist. */}
      <label htmlFor={feldID} className="sr-only">
        {worte.merkmale}
      </label>
      <p id={hilfeID} className="max-w-prose text-sm text-leise">
        {einsetzen(worte.merkmaleHilfe, { merkmale: merkmale.join(', ') })}
      </p>
      <Einstellungsherkunft eintrag={eintrag} ohneWert />
      <textarea
        id={feldID}
        data-einstellung="gestaltung"
        rows={6}
        spellCheck={false}
        value={entwurf}
        aria-invalid={zeile !== undefined}
        aria-describedby={
          zeile === undefined ? hilfeID : `${hilfeID} ${befundID}`
        }
        onChange={(ereignis) => {
          setzeEntwurf(ereignis.target.value)
          setzeZeile(undefined)
        }}
        className={feldklassen}
      />
      {zeile !== undefined && (
        <div role="alert">
          <p id={befundID} className="max-w-prose text-fehler">
            {einsetzen(worte.zeilenfehler, { zeile: String(zeile) })}
          </p>
        </div>
      )}
      <div className="flex flex-wrap gap-2">
        <Knopf
          angehalten={angehalten}
          aufKlick={() => {
            // **Erst lesen, dann schreiben** (§6: der Fehler am Feld): eine
            // Zeile, die die Form verfehlt, wird genannt und nicht
            // geschrieben — Go wiese sie ohnehin ab, aber ohne ihre Stelle.
            const gelesen = eigeneMerkmale(entwurf)
            if (gelesen.zeile !== undefined) {
              setzeZeile(gelesen.zeile)

              return
            }

            setze(entwurf)
          }}
        >
          {worte.uebernehmen}
        </Knopf>
        {/* **Der Weg zurück zur Vorgabe** — Befund M15 (#1665): „keine
            eigenen Farben" ist die Vorgabe, und bis dahin hieß der Weg
            dorthin, das Feld von Hand zu leeren. */}
        {gespeichert !== '' && (
          <Knopf
            gewicht="neben"
            angehalten={angehalten}
            aufKlick={() => {
              setze('')
            }}
          >
            {worte.zurVorgabe}
          </Knopf>
        )}
      </div>
      {gespeichert.trim() !== '' &&
        (ruht ? (
          <p role="status" className="max-w-prose text-sm text-leise">
            {worte.ruhtBeiKontrast}
          </p>
        ) : befunde.length === 0 ? (
          <p role="status" className="text-sm text-leise">
            {worte.angewandt}
          </p>
        ) : (
          <div role="status" className="max-w-prose text-sm">
            <p className="text-fehler">{worte.nichtLesbar}</p>
            <ul className="list-disc pl-5">
              {befunde.map((befund) => (
                <li key={`${befund.vorn}/${befund.hinten}`}>
                  {einsetzen(worte.paar, {
                    vorn: worte.paarnamen[befund.vorn] ?? befund.vorn,
                    hinten: worte.paarnamen[befund.hinten] ?? befund.hinten,
                    verhaeltnis: zahl(befund.verhaeltnis),
                    schwelle: zahl(befund.schwelle),
                  })}
                </li>
              ))}
            </ul>
          </div>
        ))}
    </section>
  )
}
