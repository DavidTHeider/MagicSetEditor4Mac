// **Das Kennwort einer Sicherung** — OPS-06.AK-02
// (#1723).
//
// Zwei Bausteine, und sie beantworten zwei Fragen: **wie gesichert wird**
// (mit Kennwort, oder ausdrücklich ohne) und **wie eine verschlüsselte
// Sicherung aufgeht**. Beide stehen im Seitenbereich und nicht in einem
// Dialog — §6, keine Rückfrage über dem nativen Dateidialog.
import { useId, useState } from 'react'

import { Aktionsknopf } from './Aktionsknopf'
import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'

const feldklassen =
  'rounded-md border border-kante px-2 py-1 focus-visible:outline ' +
  'focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring'

interface Sicherungsformulareigenschaften {
  sichern: (kennwort: string, klartext: boolean) => void
  angehalten: boolean

  // verschluesselt sagt, ob der Bestand selbst verschlüsselt ist — dann
  // braucht eine Sicherung ohne Kennwort das ausdrückliche „ohne".
  verschluesselt: boolean
}

// Sicherungsformular fragt nach einem Kennwort, bevor der Speichern-Dialog
// aufgeht.
//
// **Zweimal eingegeben**, weil ein Tippfehler hier die Sicherung kostet: es
// gibt keinen Weg zurück zu einem vergessenen Kennwort.
export function Sicherungsformular({
  sichern,
  angehalten,
  verschluesselt,
}: Sicherungsformulareigenschaften) {
  const worte = useTexte().seiten.bestandspflege
  const kennwortID = useId()
  const wiederholungID = useId()
  const hinweisID = useId()
  const verschiedenID = useId()

  const [kennwort, setzeKennwort] = useState('')
  const [wiederholung, setzeWiederholung] = useState('')
  const [klartext, setzeKlartext] = useState(false)

  const verschieden = wiederholung !== '' && kennwort !== wiederholung
  const ohneKennwort = kennwort === ''
  const fehltKlartext = verschluesselt && ohneKennwort && !klartext

  return (
    <div className="mt-4 flex flex-col gap-2">
      <div className="flex flex-wrap gap-3">
        <label htmlFor={kennwortID} className="flex flex-col gap-1 text-sm">
          {worte.kennwort}
          <input
            id={kennwortID}
            type="password"
            autoComplete="new-password"
            value={kennwort}
            aria-describedby={hinweisID}
            onChange={(ereignis) => {
              setzeKennwort(ereignis.target.value)
            }}
            className={feldklassen}
          />
        </label>

        <label htmlFor={wiederholungID} className="flex flex-col gap-1 text-sm">
          {worte.kennwortWiederholen}
          <input
            id={wiederholungID}
            type="password"
            autoComplete="new-password"
            value={wiederholung}
            aria-invalid={verschieden}
            aria-describedby={verschieden ? verschiedenID : undefined}
            onChange={(ereignis) => {
              setzeWiederholung(ereignis.target.value)
            }}
            className={feldklassen}
          />
        </label>
      </div>

      {verschieden && (
        <p id={verschiedenID} className="text-sm text-fehler">
          {worte.kennwortVerschieden}
        </p>
      )}

      <p id={hinweisID} className="max-w-prose text-sm text-leise">
        {worte.kennwortHinweis}
      </p>

      {/* **Das „ohne" steht nur da, wo es etwas bedeutet**: bei einem
          unverschlüsselten Bestand ist eine Sicherung ohne Kennwort
          derselbe Klartext, den es schon gibt. */}
      {verschluesselt && ohneKennwort && (
        <label className="flex max-w-prose items-start gap-2 text-sm">
          <input
            type="checkbox"
            checked={klartext}
            onChange={(ereignis) => {
              setzeKlartext(ereignis.target.checked)
            }}
          />
          <span>
            {worte.klartext}
            <span className="block text-leise">{worte.klartextHinweis}</span>
          </span>
        </label>
      )}

      <div>
        <Aktionsknopf
          aktion="sichern"
          klein
          angehalten={angehalten || verschieden || fehltKlartext}
          aufKlick={() => {
            sichern(kennwort, klartext && ohneKennwort)
          }}
        >
          {worte.sichern}
        </Aktionsknopf>
      </div>
    </div>
  )
}

interface Kennwortabfrageeigenschaften {
  pfad: string
  entsperren: (kennwort: string) => void
  abbrechen: () => void
  angehalten: boolean

  // fehler ist die Meldung zum letzten Versuch — **am Feld** (§6).
  fehler: string | undefined
}

// Kennwortabfrage öffnet eine verschlüsselte Sicherung.
export function Kennwortabfrage({
  pfad,
  entsperren,
  abbrechen,
  angehalten,
  fehler,
}: Kennwortabfrageeigenschaften) {
  const worte = useTexte().seiten.bestandspflege
  const feldID = useId()
  const fehlerID = useId()
  const [kennwort, setzeKennwort] = useState('')

  return (
    <form
      className="mt-4 flex flex-col gap-2 rounded-md border border-kante p-4"
      onSubmit={(ereignis) => {
        ereignis.preventDefault()
        entsperren(kennwort)
      }}
    >
      <p className="font-medium">{worte.gesperrt}</p>
      <p className="text-sm break-all text-leise">{pfad}</p>

      <label htmlFor={feldID} className="flex flex-col gap-1 text-sm">
        {worte.gesperrtKennwort}
        <input
          id={feldID}
          type="password"
          autoComplete="current-password"
          value={kennwort}
          aria-invalid={fehler !== undefined}
          aria-describedby={fehler === undefined ? undefined : fehlerID}
          onChange={(ereignis) => {
            setzeKennwort(ereignis.target.value)
          }}
          className={feldklassen}
        />
      </label>

      {fehler !== undefined && (
        <div role="alert">
          <p id={fehlerID} className="text-sm text-fehler">
            {fehler}
          </p>
        </div>
      )}

      <div className="flex gap-3">
        <Knopf
          klein
          angehalten={angehalten || kennwort === ''}
          aufKlick={() => {
            entsperren(kennwort)
          }}
        >
          {worte.entsperren}
        </Knopf>
        <Knopf klein angehalten={angehalten} aufKlick={abbrechen}>
          {worte.abbrechen}
        </Knopf>
      </div>
    </form>
  )
}
