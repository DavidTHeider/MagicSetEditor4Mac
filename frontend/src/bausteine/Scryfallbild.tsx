// **Ein Kartenbild von Scryfall holen**
// (#1005).
//
// # Warum die Nennung hier steht und nicht in einer Fußnote
//
// Scryfall verlangt es, nachgelesen am 2026-08-23 unter
// `https://scryfall.com/docs/api`:
//
// > When using the art_crop, list the artist name and copyright elsewhere in
// > the same interface presenting the art crop […] Users should be able to
// > identify the artist and source of the image somehow.
//
// „Same interface" ist **diese** Fläche. Der Satz steht deshalb neben dem
// Zuschnittfenster, solange das Bild dort liegt — und der Künstler geht
// zusätzlich als `illustrator` in die Karte, damit auch die Ausfuhr ihn trägt.
import { useState } from 'react'

import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

interface ScryfallbildEigenschaften {
  // feld ist das Bildfeld, in das das Bild kommt.
  feld: string

  // vorschlag ist der Kartenname, den das Formular schon kennt.
  //
  // **Vorbelegt und nicht festgelegt**: eine eigene Karte heißt oft anders als
  // die, deren Bild man will.
  vorschlag: string

  laeuft: boolean

  holen: (name: string, art: string) => void
}

// Scryfallbild fragt nach Name und Bildart und holt das Bild.
export function Scryfallbild({
  feld,
  vorschlag,
  laeuft,
  holen,
}: ScryfallbildEigenschaften): React.JSX.Element {
  const texte = useTexte().seiten.karten.scryfall

  const [name, setzeName] = useState(vorschlag)
  const [art, setzeArt] = useState('normal')

  const namensfeld = `scryfall-name-${feld}`
  const artfeld = `scryfall-art-${feld}`

  return (
    <section className="mt-4 border-t border-kante pt-4">
      <h3 className="text-sm font-semibold">{texte.ueberschrift}</h3>
      <Erklaerungsabsatz
        text={texte.erklaerung}
        klasse="mt-1 text-xs text-leise"
      />

      <div className="mt-2 flex flex-wrap items-end gap-2">
        <div className="flex flex-col gap-1">
          <label className="text-xs" htmlFor={namensfeld}>
            {texte.name}
          </label>
          <input
            className="rounded-sm border border-kante px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            id={namensfeld}
            onChange={(ereignis) => {
              setzeName(ereignis.target.value)
            }}
            type="text"
            value={name}
          />
        </div>

        <div className="flex flex-col gap-1">
          <label className="text-xs" htmlFor={artfeld}>
            {texte.art}
          </label>
          <select
            className="rounded-sm border border-kante px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            id={artfeld}
            onChange={(ereignis) => {
              setzeArt(ereignis.target.value)
            }}
            value={art}
          >
            <option value="klein">{texte.artKlein}</option>
            <option value="normal">{texte.artNormal}</option>
            <option value="ausschnitt">{texte.artAusschnitt}</option>
          </select>
        </div>

        <Knopf
          angehalten={laeuft || name.trim() === ''}
          aufKlick={() => {
            holen(name.trim(), art)
          }}
        >
          {texte.holen}
        </Knopf>
      </div>

      {art === 'ausschnitt' && (
        <Erklaerungsabsatz
          text={texte.ausschnittHinweis}
          klasse="mt-2 text-xs text-leise"
        />
      )}
    </section>
  )
}
