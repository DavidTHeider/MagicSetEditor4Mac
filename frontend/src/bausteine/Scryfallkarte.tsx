// **Eine Karte aus einem Scryfall-Namen anlegen**
// (#1032).
//
// # Warum das neben „Karte anlegen" steht
//
// Bis dahin arbeitete der Abgleich nur an einer **vorhandenen** Karte: erst
// eine leere anlegen, dann abgleichen — zwei Schritte für einen Vorgang, und
// den ersten muss man kennen. Hier stehen beide Wege nebeneinander, und die
// Wahl ist die zwischen „leer" und „aus Scryfall".
import { useId, useState } from 'react'

import { Aktionsknopf } from './Aktionsknopf'
import { useTexte } from '../sprache/kontext'

interface ScryfallkarteEigenschaften {
  laeuft: boolean

  // anlegen bekommt den Namen und die Bildart — leer heißt: kein Bild.
  anlegen: (name: string, art: string) => void

  // name ist der vorbelegte Kartenname, wenn ihn jemand von außen setzt
  // (#1167).
  //
  // **Wahlfrei, und der Kasten bleibt tippbar.** Die Suche daneben füllt ihn
  // mit dem Namen, wie Scryfall ihn führt — wer ihn danach ändern will, kann
  // es. Ein Kasten, der sich nicht mehr ändern ließe, machte aus der Hilfe
  // eine Bevormundung.
  //
  // **Als Schlüssel und nicht als Wert**: der Aufrufer hängt ihn an den `key`,
  // damit ein neuer Vorschlag den Kasten neu aufbaut. Ein `value`, das den
  // Zustand überschriebe, nähme jedem Tastendruck seine Wirkung.
  name?: string
}

// Scryfallkarte fragt nach dem Kartennamen und legt die Karte an.
export function Scryfallkarte({
  laeuft,
  anlegen,
  name: vorbelegt,
}: ScryfallkarteEigenschaften): React.JSX.Element {
  const texte = useTexte().seiten.karten.scryfallkarte

  const [name, setzeName] = useState(vorbelegt ?? '')
  const [mitBild, setzeMitBild] = useState(true)
  const namensfeld = useId()

  return (
    <div className="flex flex-wrap items-end gap-2">
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
          placeholder={texte.platzhalter}
          type="text"
          value={name}
        />
      </div>

      {/* **Das Bild ist wahlweise dabei.** Ohne den Haken wird keines geholt —
          ein Bild ist ein eigener Abruf mit einer eigenen Bedingung. */}
      <label className="flex items-center gap-2 py-1 text-xs">
        <input
          checked={mitBild}
          className="focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          onChange={(ereignis) => {
            setzeMitBild(ereignis.target.checked)
          }}
          type="checkbox"
        />
        {texte.mitBild}
      </label>

      <Aktionsknopf
        aktion="anlegen"
        angehalten={laeuft || name.trim() === ''}
        aufKlick={() => {
          anlegen(name.trim(), mitBild ? 'ausschnitt' : '')
        }}
      >
        {texte.anlegen}
      </Aktionsknopf>
    </div>
  )
}
