// **Ein Freitext mit Vorschlägen** — UX-06.AK-01
// (#1678).
//
// # Wofür
//
// **Für Werte, deren Menge offen ist, aber eine übliche Auswahl hat**: die
// Währung eines Preises, die Sprache eines Drucks. Das Kriterium sagt es
// wörtlich — „fachlich offene Werte behalten Freitext mit Vorschlägen" —, und
// die Umstellung darf nichts ausschließen, was der Bestand annimmt. Eine
// geschlossene Auswahl hieße hier: ein Druck in einer Sprache, die Scryfall
// nächstes Jahr aufnimmt, ließe sich nicht mehr erfassen.
//
// # Warum `datalist` und nicht Mantines `Autocomplete`
//
// **Weil es genau das Verhalten ist, das gebraucht wird, und keines mehr**:
// das Feld bleibt ein `input`, der Browser legt die Vorschläge darunter, und
// WebKit kennt es seit Safari 12.1. Mantine brächte eine zweite Liste mit
// eigener Tastaturführung und eigenem Stil — für einen Fall, in dem nichts
// davon fehlt (UX-07.AK-03: keine zweite Bibliothek neben der ersten).
import { useId } from 'react'

import { Feld } from './Feld'
import { kennungseingabe } from './kennungseingabe'

// vorschlagsklasse ist die Gestalt eines Eingabefelds — die von
// `Beschriftetesfeld` mit dem roten Rand eines Befunds, wie bei `Pflichtfeld`
// und `Zahlenfeld`. **Ausgeschrieben und nicht zusammengesetzt**: der Wächter
// über den Fokusring folgt einer Größe nur in derselben Datei.
// **Die Höhe eines gewöhnlichen Textfelds, ausdrücklich**: WebKit gibt
// einem Feld mit Vorschlagsliste Platz für ihren Pfeil — gemessen 35 statt
// 30 Punkte, und in einer Zeile mit anderen Feldern stand es höher als seine
// Nachbarn (Hüllenprobe `trade02`, TRADE-02).
const vorschlagsklasse =
  'h-[1.875rem] rounded-sm border border-kante bg-flaeche px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring aria-[invalid=true]:border-fehler'

// Vorschlag ist ein Wert, den das Feld anbietet — mit einem Namen, wenn der
// Wert allein nichts sagt (`EUR` — Euro).
export interface Vorschlag {
  wert: string
  bezeichnung?: string
}

interface VorschlagsfeldEigenschaften {
  beschriftung: string
  wert: string
  setzeWert: (wert: string) => void
  vorschlaege: readonly Vorschlag[]

  befund?: string | undefined

  // hinweis erklärt das Feld, bevor ein Befund kommt — etwa, warum es
  // Pflicht ist (DOM-04.AK-01, #1699).
  hinweis?: string | undefined

  // angehalten sperrt das Feld, wenn sein Wert schon feststeht — etwa aus
  // einer Vorlage.
  angehalten?: boolean | undefined

  // beispiel steht als Platzhalter im leeren Feld — ein Wert, wie er
  // aussieht (AK-02), und kein Hinweis, der das Formular höher macht.
  beispiel?: string | undefined

  // verlassen meldet, dass das Feld den Fokus abgibt — der Moment, zu dem
  // ein Urteil über den Wert fair ist (derselbe wie bei `Pflichtfeld`).
  verlassen?: () => void

  // breite ist die Spaltenzahl — ein Code braucht fünf, ein Name zwanzig.
  breite: number
}

export function Vorschlagsfeld({
  beschriftung,
  wert,
  setzeWert,
  vorschlaege,
  befund,
  hinweis,
  angehalten,
  beispiel,
  verlassen,
  breite,
}: VorschlagsfeldEigenschaften) {
  const listenID = useId()

  return (
    <Feld
      beschriftung={beschriftung}
      {...(befund === undefined ? {} : { befund })}
      {...(hinweis === undefined ? {} : { hinweis })}
    >
      {(verbindung) => (
        <>
          <input
            {...kennungseingabe}
            {...verbindung}
            type="text"
            list={listenID}
            value={wert}
            placeholder={beispiel}
            size={breite}
            disabled={angehalten}
            autoComplete="off"
            onChange={(ereignis) => {
              setzeWert(ereignis.target.value)
            }}
            onBlur={verlassen}
            className={vorschlagsklasse}
          />
          <datalist id={listenID}>
            {vorschlaege.map((vorschlag) => (
              <option
                key={vorschlag.wert}
                value={vorschlag.wert}
                label={vorschlag.bezeichnung}
              />
            ))}
          </datalist>
        </>
      )}
    </Feld>
  )
}
