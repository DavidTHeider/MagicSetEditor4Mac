// **Die Zeilen der eigenen Übersetzungen** — LOC-02
// (#1688).
//
// **Auf Modulebene und nicht im Rumpf von `Eigenuebersetzungen`**: ein
// Bauteil, das dort entstünde, wäre bei jedem Zeichnen ein neuer Typ, und
// die Rückfrage vor dem Löschen und die Felder der Neuanlage verlören ihren
// Stand.
import { useId, useState } from 'react'

import { einsetzen } from '../sprache/einsetzen'
import { sprachen } from '../sprache/kataloge'
import { useTexte } from '../sprache/kontext'
import { Aktionsknopf } from './Aktionsknopf'
import { Knopf } from './Knopf'

import type { SprachpaketDTO } from '../apisprache'

const feld =
  'rounded-md border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

export function Paketzeile({
  paket,
  angehalten: halt,
  bearbeiten,
  ausgeben,
  loeschen,
}: {
  paket: SprachpaketDTO
  angehalten: boolean
  bearbeiten: () => void
  ausgeben: () => void
  loeschen: () => void
}) {
  const eigene = useTexte().seiten.einstellungen.eigenuebersetzung
  const [fragt, setzeFragt] = useState(false)
  const stande = eigene.staende as Record<string, string>

  return (
    <li className="flex flex-col gap-2 py-3">
      <div className="flex flex-wrap items-baseline gap-2">
        <span className="font-medium">{paket.name}</span>
        <span className="text-sm text-leise">
          {`${paket.kennung} · ${paket.basis} · ${stande[paket.stand] ?? paket.stand}`}
        </span>
      </div>
      <span className="text-sm text-leise">
        {einsetzen(eigene.zahlen, {
          uebersetzt: String(paket.uebersetzt),
          gesamt: String(paket.gesamt),
          veraltet: String(paket.veraltet),
        })}
      </span>
      {fragt ? (
        <div className="flex flex-wrap items-center gap-2">
          <span className="text-sm">
            {einsetzen(eigene.loeschenFrage, { name: paket.name })}
          </span>
          <Aktionsknopf aktion="loeschen" angehalten={halt} aufKlick={loeschen}>
            {eigene.loeschenJa}
          </Aktionsknopf>
          <Knopf
            aufKlick={() => {
              setzeFragt(false)
            }}
          >
            {eigene.abbrechen}
          </Knopf>
        </div>
      ) : (
        <div className="flex flex-wrap gap-2">
          <Aktionsknopf
            aktion="bearbeiten"
            angehalten={halt}
            aufKlick={bearbeiten}
          >
            {eigene.bearbeiten}
          </Aktionsknopf>
          <Knopf angehalten={halt} aufKlick={ausgeben}>
            {eigene.ausgeben}
          </Knopf>
          <Aktionsknopf
            aktion="loeschen"
            gewicht="neben"
            angehalten={halt}
            aufKlick={() => {
              setzeFragt(true)
            }}
          >
            {eigene.loeschen}
          </Aktionsknopf>
        </div>
      )}
    </li>
  )
}

export function Neuanlage({
  angehalten: halt,
  feldbefund,
  anlegen,
  einlesen,
}: {
  angehalten: boolean
  feldbefund: (name: string) => string | undefined
  anlegen: (kennung: string, name: string, basis: string) => void
  einlesen: (ersetzen: boolean) => void
}) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen.eigenuebersetzung
  const [kennung, setzeKennung] = useState('')
  const [name, setzeName] = useState('')
  const [basis, setzeBasis] = useState('en')
  const [ersetzen, setzeErsetzen] = useState(false)
  const hinweisID = useId()
  const befundKennung = feldbefund('kennung')
  const befundName = feldbefund('name')
  const befundBasis = feldbefund('basis')

  return (
    <div className="flex flex-col gap-3 border-t border-kante pt-4">
      <h4 className="text-sm font-medium">{eigene.neu}</h4>
      <div className="flex flex-wrap items-end gap-3">
        <label className="flex flex-col gap-1 text-sm">
          {eigene.kennung}
          <input
            value={kennung}
            aria-describedby={hinweisID}
            aria-invalid={befundKennung !== undefined}
            onChange={(ereignis) => {
              setzeKennung(ereignis.target.value)
            }}
            className={feld}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          {eigene.name}
          <input
            value={name}
            aria-invalid={befundName !== undefined}
            onChange={(ereignis) => {
              setzeName(ereignis.target.value)
            }}
            className={feld}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          {eigene.basis}
          <select
            value={basis}
            onChange={(ereignis) => {
              setzeBasis(ereignis.target.value)
            }}
            className={feld}
          >
            {sprachen.map((sprache) => (
              <option key={sprache} value={sprache}>
                {texte.allgemein.sprache[sprache]}
              </option>
            ))}
          </select>
        </label>
        <Aktionsknopf
          aktion="anlegen"
          angehalten={halt || kennung.trim() === '' || name.trim() === ''}
          aufKlick={() => {
            anlegen(kennung.trim(), name, basis)
          }}
        >
          {eigene.anlegen}
        </Aktionsknopf>
      </div>
      <p id={hinweisID} className="text-sm text-leise">
        {befundKennung ?? befundName ?? befundBasis ?? eigene.kennungHinweis}
      </p>
      <div className="flex flex-wrap items-center gap-3">
        <Knopf
          angehalten={halt}
          aufKlick={() => {
            einlesen(ersetzen)
          }}
        >
          {eigene.einlesen}
        </Knopf>
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={ersetzen}
            onChange={(ereignis) => {
              setzeErsetzen(ereignis.target.checked)
            }}
            className="size-4 accent-marke focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          />
          {eigene.ersetzen}
        </label>
      </div>
    </div>
  )
}
