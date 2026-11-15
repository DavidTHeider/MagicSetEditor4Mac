// Die Rückfrage vor dem Löschen eines Sets, samt Bestätigung durch seinen
// Namen (#1218).
//
// # Warum der Name abgetippt wird und kein „Ja" reicht
//
// **Weil es keinen Papierkorb gibt**
// ([0054](../../../docs/entscheidungen/0054-ein-set-laesst-sich-loeschen.md)).
// Eine Rückfrage mit „Ja" und „Nein" lässt sich wegklicken, ohne gelesen zu
// haben — und der zweite Klick sitzt an der Stelle, an der eben noch etwas
// Harmloses stand. Ein Name, den man abtippt, verlangt einen Blick auf das,
// was gemeint ist.
//
// Das ist derselbe Griff, den GitHub für das Löschen eines Repositoriums
// benutzt, und er ist hier aus demselben Grund richtig: der Vorgang ist
// selten, folgenreich und nicht rückgängig zu machen.
//
// **Kein Modal** (§6): die Frage steht unter der Karte, deren Menü sie
// gestellt hat.
//
// # Was diese Komponente **nicht** tut
//
// Sie ruft den Dienst nicht, und sie entscheidet nicht, wann sie dasteht.
// Gefragt wird über das Menü der Setkarte (UXR-04,
// #1912) —
// bis dahin trug sie ihren eigenen Knopf „Löschen …". Der Vorgangshaken liegt
// eine Ebene höher — einer je Set und nicht einer je Zeile (§7, Regel 4) —,
// und was nach dem Löschen mit der Liste geschieht, weiß nur die Seite (§7,
// Regel 1).
import { useEffect, useRef, useState } from 'react'

import { type SetDTO } from '../api'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Knopf } from '../bausteine/Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { zaehlwort } from '../sprache/zaehlwort'
import { useTexte } from '../sprache/kontext'

interface SetloeschenEigenschaften {
  satz: SetDTO
  laeuft: boolean

  // loeschen bekommt den **abgetippten** Namen, nicht den aus dem Satz.
  //
  // **Und das ist kein Umweg.** Der Bestand hält ihn gegen den *heutigen*
  // Namen; hat jemand anderes das Set inzwischen umbenannt, weist er ab (§7,
  // Regel 2). Wer hier `satz.name` durchreichte, gäbe einen Stand von vorhin
  // weiter und machte die Prüfung im Bestand wirkungslos.
  loeschen: (setID: string, name: string) => void

  // abbrechen nimmt die Frage zurück — die Karte hängt sie dann aus, und
  // mit ihr geht der getippte Name.
  abbrechen: () => void

  // befunde sind die Reste eines **geglückten** Löschens (§7, Regel 3).
  befunde: string[]

  fehler: string | undefined
}

export function Setloeschen({
  satz,
  laeuft,
  loeschen,
  abbrechen,
  befunde,
  fehler,
}: SetloeschenEigenschaften) {
  const texte = useTexte()

  const [getippt, setzeGetippt] = useState('')
  const feld = useRef<HTMLInputElement>(null)

  // **Der Fokus geht ins Feld, sobald gefragt ist.** Wer im Menü „Löschen …"
  // wählt, will als Nächstes den Namen tippen; ohne das stünde er wieder auf
  // dem Menüknopf und müsste die Frage erst suchen.
  useEffect(() => {
    feld.current?.focus()
  }, [])

  const feldID = `setloeschen-${satz.id}`
  const passt = getippt === satz.name

  if (befunde.length > 0) {
    return (
      <div role="status" className="text-hinweis">
        <p>{texte.seiten.sets.loeschenBefunde}</p>

        <ul className="mt-1 list-disc pl-5">
          {befunde.map((befund) => (
            <li key={befund}>{befund}</li>
          ))}
        </ul>
      </div>
    )
  }

  return (
    <div className="w-full rounded-lg border border-fehler p-3">
      <h4 className="font-semibold">{texte.seiten.sets.loeschenTitel}</h4>

      {/* **`role="alert"`**: wer die Seite mit einer Vorlesehilfe bedient,
          soll die Warnung hören, ohne sie zu suchen. */}
      <p role="alert" className="mt-1 text-text">
        {einsetzen(texte.seiten.sets.loeschenWarnung, {
          name: satz.name,
          karten: zaehlwort(satz.karten, texte.allgemein.zaehlwoerter.karten),
        })}
      </p>

      <p className="mt-1 text-sm text-leise">
        {texte.seiten.sets.loeschenSammlung}
      </p>

      <label htmlFor={feldID} className="mt-3 block font-medium">
        {texte.seiten.sets.loeschenFeld}
      </label>

      <input
        ref={feld}
        id={feldID}
        type="text"
        value={getippt}
        autoComplete="off"
        onChange={(ereignis) => {
          setzeGetippt(ereignis.target.value)
        }}
        className="mt-1 w-full rounded-sm border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
      />

      {!passt && (
        <p className="mt-1 text-sm text-leise">
          {einsetzen(texte.seiten.sets.loeschenPasstNicht, { name: satz.name })}
        </p>
      )}

      {fehler !== undefined && (
        <p role="alert" className="mt-1 text-fehler">
          {fehler}
        </p>
      )}

      <div className="mt-3 flex flex-wrap gap-2">
        <Aktionsknopf
          aktion="loeschen"
          gewicht="haupt"
          angehalten={!passt || laeuft}
          aufKlick={() => {
            loeschen(satz.id, getippt)
          }}
        >
          {texte.seiten.sets.loeschenJa}
        </Aktionsknopf>

        <Knopf aufKlick={abbrechen}>{texte.seiten.sets.loeschenNein}</Knopf>
      </div>
    </div>
  )
}
