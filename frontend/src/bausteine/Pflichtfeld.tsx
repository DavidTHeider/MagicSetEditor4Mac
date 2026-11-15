// **Ein Textfeld, das sagt, dass es eines braucht** — Werkstück F von Phase
// 4.10 (#1325).
//
// # Was vorher war
//
// Sechs Formulare haben ein Pflichtfeld — Deck, Format, Turnier, Teilnehmer,
// Autoersetzungsregel, Set. Bleibt es leer, **ruht der Knopf**: die
// Rückmeldung ist da, nur nicht **am Feld**, wie §6 es verlangt. Wer den
// ausgegrauten Knopf sieht, weiß nicht, welches der drei Felder fehlt.
//
// # Erst nach der Berührung
//
// **Ein leeres Formular ist kein Fehler, sondern ein leeres Formular.** Ein
// Hinweis, der beim Aufschlagen schon dasteht, ist eine Anklage für etwas,
// das niemand getan hat — und er stumpft ab, bis ihn keiner mehr liest.
//
// Gemeldet wird deshalb, wer das Feld **anfasst und leer verlässt**. Das ist
// dieselbe Regel wie bei [Zahlenfeld] nebenan: ein leeres Feld ist dort kein
// Fehler, weil man beim Ändern zuerst löscht.
import { useState, type RefObject } from 'react'

import { Feld } from './Feld'

import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

interface PflichtfeldEigenschaften {
  beschriftung: string
  wert: string
  angehalten?: boolean

  // griff reicht bis ans Feld durch.
  //
  // **Weil ein Menübefehl den Fokus hineinsetzt**: ⌘N und „Neues Set" tun das
  // beim Setnamen (`SetNeu.tsx`). Ohne ihn griffe der Befehl ins Leere.
  griff?: RefObject<HTMLInputElement | null>

  // autofokus nennt das Feld einer Schublade, in das der Fokus beim Öffnen
  // geht (`data-autofocus`, `Schublade`) — der Name des Sets in der
  // Setanlage (AP-10). Außerhalb einer Schublade bewirkt es nichts.
  autofokus?: boolean

  // befund ist ein Fehler, den die Anwendung **zu diesem Feld** gemeldet hat.
  //
  // **Er steht neben `fehlt` und nicht darin**: „leer" merkt das Fenster
  // selbst, „taugt nicht als Verzeichnisname" weiß nur der Dienst. Beide
  // gehören ans verursachende Feld (§6) — angestrichen vom Nachlauf über
  // Phase 4.10, der fand, dass die Paketanlage ihren Feldbezug wegwarf und
  // die Meldung als anonymen Hinweis über dem Formular zeigte.
  befund?: string | undefined

  // hinweis erklärt das Feld, **bevor** ein Fehler kommt (§6).
  //
  // Er steht unter dem Feld und verschwindet, sobald eine Meldung da ist —
  // zwei Sätze übereinander, von denen einer nicht mehr gilt, sind schlimmer
  // als einer.
  hinweis?: string | undefined

  // kennung sagt, dass der Wert **kein Wort** ist — ein Verzeichnisname,
  // eine Anmeldekennung. Dann schweigt die Rechtschreibkorrektur: am nativen
  // Fenster machte macOS beim Verlassen des Felds aus dem Kurznamen „namen“
  // ein „Namen“, und das Paket hieß anders als getippt (Hüllenprobe
  // `fmt02d`, FMT-02).
  kennung?: boolean

  // beschriebenVon nennt ein Element, das das Feld **zusätzlich** zu Hinweis
  // und Befund beschreibt — etwa den Grund, aus dem es einem Leser gesperrt
  // ist (H01, #2255).
  beschriebenVon?: string | undefined

  geaendert: (wert: string) => void
}

export function Pflichtfeld({
  beschriftung,
  wert,
  angehalten,
  griff,
  autofokus,
  befund,
  hinweis,
  kennung,
  beschriebenVon,
  geaendert,
}: PflichtfeldEigenschaften) {
  const texte = useTexte()

  const [beruehrt, setzeBeruehrt] = useState(false)

  const fehlt = beruehrt && wert.trim() === ''

  // **Der Befund der Anwendung schlägt das eigene Urteil.** Wer etwas
  // eingegeben hat, das der Dienst abweist, soll das lesen und nicht „darf
  // nicht leer sein".
  const meldung =
    befund ??
    (fehlt
      ? einsetzen(texte.bausteine.pflichtfeldLeer, { feld: beschriftung })
      : undefined)

  // **Beschriftung, Hinweis und Befund kommen aus `Feld`** — seit Werkstück
  // U2 von Phase 5.1 (#1413). Was hier bleibt, ist das Urteil über ein leeres
  // Pflichtfeld; wie es am Feld steht, ist die Zusage des Bausteins darunter.
  return (
    <Feld
      beschriftung={beschriftung}
      platzFuerBefund
      {...(hinweis === undefined ? {} : { hinweis })}
      {...(meldung === undefined ? {} : { befund: meldung })}
    >
      {(verbindung) => (
        <input
          {...verbindung}
          ref={griff}
          data-autofocus={autofokus === true ? true : undefined}
          type="text"
          value={wert}
          {...(kennung === true ? { spellCheck: false } : {})}
          disabled={angehalten}
          aria-describedby={
            [verbindung['aria-describedby'], beschriebenVon]
              .filter((teil) => teil !== undefined && teil !== '')
              .join(' ') || undefined
          }
          className="rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60 aria-[invalid=true]:border-fehler"
          onChange={(ereignis) => {
            geaendert(ereignis.target.value)
          }}
          // **Gemerkt wird, wer das Feld leer verlässt** — und nicht, dass
          // es einmal verlassen war (UX-08.NAT-02, S08.NAT-01). Wer einen
          // Namen tippt und „Anlegen" drückt, verlässt das Feld dabei; das
          // Formular leert es nach dem Erfolg, und darunter stand rot, der
          // Name brauche einen Wert.
          onBlur={() => {
            setzeBeruehrt(wert.trim() === '')
          }}
        />
      )}
    </Feld>
  )
}
