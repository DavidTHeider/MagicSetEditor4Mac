// **Die Handgriffe am Dateibaum** — FMT-02.AK-01
// (#1691).
//
// **Eine Zeile über dem Editor**, nicht je Datei im Baum: sie wirkt auf die
// gewählte Datei und ihr Paket, und dort liegt auch der Stand, den Umbenennen
// und Löschen einer Vorlage brauchen. Eingabe und Rückfrage stehen **in der
// Zeile selbst** — kein Dialog über dem Dialog (§6).
//
// **Hinter einem Menüknopf, nicht als Reihe** — UXR-18
// (#1926).
// Vier gleichrangige Knöpfe standen über dem Editor für Handgriffe, die man
// selten braucht; das Artboard `Werkstatt` zeigt dort gar keine. Ein Menü
// nimmt sie in eine Tabstelle, und ein ruhender Eintrag bleibt sichtbar
// statt zu verschwinden.
//
// **Die Hauptdatei bietet weder Umbenennen noch Löschen an** — ohne sie ist
// das Paket keines mehr, und der Dienst weist es ohnehin ab. **Ungesichertes
// sperrt beides**: ein Umbenennen trüge den getippten Text nicht mit.
import { useState } from 'react'

import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Knopf } from '../bausteine/Knopf'
import { Menue } from '../bausteine/Menue'
import { useTexte } from '../sprache/kontext'

import type { KeyboardEvent } from 'react'
import type { Dateiaktionen as Handgriffe } from '../haken/useDateiaktionen'
import { dateieintraege } from './dateieintraege'

interface DateiaktionenEigenschaften {
  aktionen: Handgriffe
  datei: string

  // hauptdatei ist wahr, wenn die gewählte Datei die des Pakets ist.
  hauptdatei: boolean

  // gesperrt ist wahr, solange die gewählte Datei ungesichert ist.
  gesperrt: boolean
}

export function Dateiaktionen({
  aktionen,
  datei,
  hauptdatei,
  gesperrt,
}: DateiaktionenEigenschaften) {
  const texte = useTexte().seiten.vorlagen
  const { modus, setzeModus } = aktionen
  const ruhe = () => {
    setzeModus('ruhe')
  }

  const eintraege = dateieintraege(aktionen, hauptdatei, gesperrt, texte)

  return (
    <div
      role="group"
      aria-label={texte.dateiaktionen}
      className="flex flex-col gap-2"
    >
      {modus === 'ruhe' && (
        <div className="flex flex-wrap items-center gap-2">
          <Menue bezeichnung={texte.datei} eintraege={eintraege} />
          {!hauptdatei && gesperrt && (
            <p className="text-xs text-leise">{texte.dateiErstSichern}</p>
          )}
        </div>
      )}

      {modus === 'anlegen' && (
        <Namensfeld
          bezeichnung={texte.dateiName}
          anfang=""
          bestaetigen={texte.dateiAnlegen}
          laeuft={aktionen.laeuft}
          aufBestaetigen={(name) => void aktionen.anlegen(name)}
          aufAbbrechen={ruhe}
        />
      )}

      {modus === 'umbenennen' && (
        <Namensfeld
          bezeichnung={texte.dateiNeuerName}
          anfang={datei}
          bestaetigen={texte.dateiUmbenennenJa}
          laeuft={aktionen.laeuft}
          aufBestaetigen={(neu) => void aktionen.umbenennen(neu)}
          aufAbbrechen={ruhe}
        />
      )}

      {modus === 'entfernen' && (
        <div className="flex flex-wrap items-center gap-2">
          <p className="text-sm">{texte.dateiLoeschenFrage}</p>
          <Aktionsknopf
            aktion="loeschen"
            gewicht="haupt"
            angehalten={aktionen.laeuft}
            aufKlick={() => void aktionen.entfernen()}
          >
            {texte.dateiLoeschenJa}
          </Aktionsknopf>
          <Knopf gewicht="neben" aufKlick={ruhe}>
            {texte.dateiAbbrechen}
          </Knopf>
        </div>
      )}

      {aktionen.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {aktionen.fehler}
        </p>
      )}
    </div>
  )
}

interface NamensfeldEigenschaften {
  bezeichnung: string
  anfang: string
  bestaetigen: string
  laeuft: boolean
  aufBestaetigen: (name: string) => void
  aufAbbrechen: () => void
}

// Namensfeld nimmt einen Dateinamen entgegen — ↵ bestätigt, Esc bricht ab.
function Namensfeld({
  bezeichnung,
  anfang,
  bestaetigen,
  laeuft,
  aufBestaetigen,
  aufAbbrechen,
}: NamensfeldEigenschaften) {
  const texte = useTexte().seiten.vorlagen
  const [name, setzeName] = useState(anfang)
  const leer = name.trim() === ''

  const aufTaste = (ereignis: KeyboardEvent<HTMLInputElement>) => {
    if (ereignis.key === 'Enter' && !leer) {
      ereignis.preventDefault()
      aufBestaetigen(name.trim())
    } else if (ereignis.key === 'Escape') {
      ereignis.preventDefault()
      aufAbbrechen()
    }
  }

  return (
    <div className="flex flex-wrap items-center gap-2">
      <label className="flex items-center gap-2 text-sm">
        <span>{bezeichnung}</span>
        <input
          // eslint-disable-next-line jsx-a11y/no-autofocus -- das Feld erscheint auf einen Knopfdruck hin, und dorthin gehört der Fokus
          autoFocus
          // Ein Dateiname ist kein Wort — keine Rechtschreibkorrektur.
          spellCheck={false}
          value={name}
          onChange={(ereignis) => {
            setzeName(ereignis.target.value)
          }}
          onKeyDown={aufTaste}
          className="min-w-0 rounded-md border border-kante bg-flaeche px-2 py-1 font-mono text-xs focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        />
      </label>
      <Knopf
        gewicht="haupt"
        angehalten={leer || laeuft}
        aufKlick={() => {
          aufBestaetigen(name.trim())
        }}
      >
        {bestaetigen}
      </Knopf>
      <Knopf gewicht="neben" aufKlick={aufAbbrechen}>
        {texte.dateiAbbrechen}
      </Knopf>
    </div>
  )
}
