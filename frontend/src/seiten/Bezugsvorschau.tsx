// **Was eine Aktualisierung täte** — G04
// (#2251):
// die berührten Pakete und Sets, die Gründe für eine Bestätigung, und dass
// der bisherige Stand aufgehoben bleibt. **Erst danach wird aktualisiert.**

import { useId, useState } from 'react'

import { Knopf } from '../bausteine/Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { BezugsvorschauAntwort } from '../paketbezuege'

interface Eigenschaften {
  vorschau: BezugsvorschauAntwort
  vorigerStand: string
  laeuft: boolean
  aktualisieren: (bestaetigt: boolean) => void
  verwerfen: () => void
}

const kurz = (stand: string) => stand.slice(0, 7)

export function Bezugsvorschau({
  vorschau,
  vorigerStand,
  laeuft,
  aktualisieren,
  verwerfen,
}: Eigenschaften) {
  const t = useTexte().seiten.pakete.bezuege
  const [bestaetigt, setzeBestaetigt] = useState(false)
  const titel = useId()
  const bereit = !vorschau.bestaetigungNoetig || bestaetigt

  return (
    <section
      aria-labelledby={titel}
      className="mt-3 rounded-md border border-kante bg-flaeche p-3"
    >
      <h4 id={titel} className="font-medium">
        {einsetzen(t.vorschauTitel, { name: vorschau.name })}
      </h4>

      <p className="text-sm text-leise">
        {einsetzen(t.vonZu, {
          von: kurz(vorschau.von),
          zu: kurz(vorschau.zu),
          vorsprung: String(vorschau.vorsprung),
          rueckstand: String(vorschau.rueckstand),
        })}
      </p>

      {vorschau.pakete.length === 0 ? (
        <p className="mt-2 text-sm">{t.keinePakete}</p>
      ) : (
        <ul className="mt-2 list-disc pl-5 text-sm">
          {vorschau.pakete.map((paket) => (
            <li key={paket.paket}>
              {einsetzen(t.paketzeile, {
                paket: paket.paket,
                neu: String(paket.neu),
                geaendert: String(paket.geaendert),
                entfernt: String(paket.entfernt),
              })}
            </li>
          ))}
        </ul>
      )}

      {vorschau.uebrige > 0 && (
        <p className="mt-1 text-sm text-leise">
          {einsetzen(t.uebrige, { zahl: String(vorschau.uebrige) })}
        </p>
      )}

      <p className="mt-2 text-sm">
        {vorschau.sets.length === 0
          ? t.keineSets
          : einsetzen(t.sets, { sets: vorschau.sets.join(', ') })}
      </p>

      <p className="mt-2 text-sm">
        {t.aufheben}
        {vorschau.vorigeErsetzt && (
          <> {einsetzen(t.vorigeWeicht, { stand: kurz(vorigerStand) })}</>
        )}
      </p>

      {vorschau.gruende.length > 0 && (
        <ul className="mt-2 list-disc pl-5 text-sm">
          {vorschau.gruende.map((grund) => (
            <li key={grund}>{t.gruende[grund]}</li>
          ))}
        </ul>
      )}

      {vorschau.bestaetigungNoetig && (
        <label className="mt-2 flex items-start gap-2 text-sm">
          <input
            type="checkbox"
            checked={bestaetigt}
            onChange={(ereignis) => {
              setzeBestaetigt(ereignis.target.checked)
            }}
          />
          <span>{t.bestaetigen}</span>
        </label>
      )}

      <div className="mt-3 flex flex-wrap gap-2">
        <Knopf
          gewicht="haupt"
          angehalten={laeuft || !bereit}
          aufKlick={() => {
            aktualisieren(bestaetigt)
          }}
        >
          {einsetzen(t.ausfuehren, { zu: kurz(vorschau.zu) })}
        </Knopf>
        <Knopf aufKlick={verwerfen}>{t.verwerfen}</Knopf>
      </div>
    </section>
  )
}
