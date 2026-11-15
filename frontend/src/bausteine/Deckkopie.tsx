// **Ein Deck kopieren** — DOM-03.AK-02
// (#1698).
//
// **Das Formular im Popover am Menü des Decks** — die Zeile bleibt, wo sie
// ist, und die Frage ist eine Zeile lang: der Name der Kopie. Den Lauf hält
// `useDeckkopie`; begonnen wird im Menü „…“ oder im Rechtsklick (AP-6,
// #2363).
import type { Deckkopierlauf } from '../haken/useDeckkopie'
import { useTexte } from '../sprache/kontext'
import { Aktionsknopf } from './Aktionsknopf'
import { Knopf } from './Knopf'
import { Pflichtfeld } from './Pflichtfeld'

export function Deckkopieformular({
  lauf,
  name,
}: {
  lauf: Deckkopierlauf
  name: string
}) {
  const eigene = useTexte().seiten.decks

  return (
    <>
      <Pflichtfeld
        beschriftung={eigene.kopieName}
        wert={name}
        befund={lauf.befund}
        geaendert={lauf.setzeName}
      />

      <div className="flex flex-wrap gap-2">
        <Aktionsknopf
          aktion="anlegen"
          gewicht="haupt"
          angehalten={name.trim() === '' || lauf.laeuft}
          aufKlick={() => {
            void lauf.kopieren(name.trim())
          }}
        >
          {eigene.kopieAnlegen}
        </Aktionsknopf>

        <Knopf aufKlick={lauf.abbrechen}>{eigene.kopieAbbrechen}</Knopf>
      </div>
    </>
  )
}
