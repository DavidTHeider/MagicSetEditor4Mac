// **Ein Turnier als Datei und als Bericht** — DOM-04.AK-03
// (#1699),
// Vermerk 0140.
//
// **Einlesen im Kopf der Seite**; Ausgeben und Bericht stehen seit AP-6
// (#2363) im Menü des offenen Turniers (`haken/useTurnierhandlungen.ts`).
// **Ein weggeklickter Dialog sagt nichts**: er ist eine Entscheidung und
// kein Fehler. Ein Nein steht über der Seite, wie jedes andere dort.
import { api } from '../api'
import { auspacken } from '../umschlag'
import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'
import type { Turnierstand } from '../haken/useTurniere'

// Turniereinlesen legt ein Turnier aus einer Datei an und schlägt es auf —
// **ein Schreiben**: danach wird nachgeladen (§7, Regel 3).
//
// **Nur der Knopf** — seit UXR-12
// (#1920)
// steht er im Kopf der Seite neben *Turnier anlegen …*, und was das Einlesen
// fand, steht darunter (`Einlesebericht`). Im Kopf bräche ein Bericht die
// Reihe der Knöpfe um.
export function Turniereinlesen({
  stand,
  eingelesen,
}: {
  stand: Turnierstand

  // eingelesen bekommt die Befunde eines Einlesens — `null` nach einem
  // weggeklickten Dialog: er ist eine Entscheidung und kein Ergebnis.
  eingelesen: (befunde: string[] | null) => void
}) {
  const eigene = useTexte().seiten.turniere

  const marke = 'turnier:einlesen'

  const einlesen = async () => {
    // **Ein Kasten und keine lose Variable**: was die Arbeit hineinlegt,
    // liest der Aufrufer danach — und das weiß die Typprüfung nur so.
    const ergebnis = { kennung: '' }

    await stand.fuehre(marke, async () => {
      const antwort = await api().TurnierEinlesenOeffnen()
      auspacken(antwort, () => undefined)

      ergebnis.kennung = antwort.kennung
      eingelesen(antwort.abgebrochen ? null : antwort.befunde)
    })

    if (ergebnis.kennung !== '') {
      await stand.waehle(ergebnis.kennung)
    }
  }

  return (
    <Knopf
      schreibt
      angehalten={stand.vorgang.laeuft(marke)}
      aufKlick={() => {
        void einlesen()
      }}
    >
      {eigene.einlesen}
    </Knopf>
  )
}

// Einlesebericht sagt, dass eingelesen wurde — und was dabei auffiel.
export function Einlesebericht({ befunde }: { befunde: string[] | null }) {
  const eigene = useTexte().seiten.turniere

  if (befunde === null) {
    return null
  }

  return (
    <div role="status" className="text-sm">
      <p>{eigene.eingelesen}</p>

      {befunde.length > 0 && (
        <>
          <p>{eigene.einlesebefunde}</p>
          <ul className="list-disc pl-5">
            {befunde.map((befund) => (
              <li key={befund}>{befund}</li>
            ))}
          </ul>
        </>
      )}
    </div>
  )
}
