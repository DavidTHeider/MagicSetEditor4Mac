// **Ein Eingabefeld für einen Nutzlastschlüssel** —
// #1623.
//
// # Warum ein Bauteil und nicht achtunddreißig
//
// **Weil die Unterschiede in der Nutzlast liegen** — der Grund steht
// ausführlich im Kopf von `spiel/befehlsformen.ts`.
//
// **Die zusammengesetzten stehen daneben**: der Spielstein bringt vier
// Angaben mit (`Spielsteinfeld`), Angriffe und Blöcke sind Paare
// (`Kampffeld`). Beide beantworten eine eigene Frage, und diese Datei stand
// mit ihnen über der Zeilengrenze aus §2.
import { useTexte } from '../sprache/kontext'
import { alsListe, alsText, kartenwort } from '../spiel/feldwerte'
import { handlungswort, zonenwort } from '../seiten/partiewoerter'

import type { Befehlsfeld as Feld } from '../spiel/befehlsformen'
import type { Katalog } from '../sprache/kataloge'
import type { Partie } from '../apipartie'

// zonen ist die Reihenfolge, in der ein Mensch sie liest.
const zonen = [
  'hand',
  'tisch',
  'stapel',
  'bibliothek',
  'friedhof',
  'exil',
  'kommando',
] as const

// stellen sind die drei Plätze in einer geordneten Zone.
const stellen = ['oben', 'unten', 'zufaellig'] as const

// manaarten sind die sechs aus CR 106.1a — fünf Farben und farblos.
const manaarten = ['W', 'U', 'B', 'R', 'G', 'C'] as const

// **Der Fokusring steht in einer Größe und nicht in einer Vorlage.**
// `internal/werkzeuge/fokusring_test.go` löst `className={name}` auf und eine
// Zeichenkettenvorlage nicht — ein Ring, den er nicht sieht, zählt als
// keiner, und ein Wächter mit falschem Alarm ist schlechter als keiner (§10).
// Die **Breite** steht deshalb am Etikett darum.
const eingabeklassen =
  'w-full rounded-sm border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

// **Eine Zahl braucht keine Textbreite.** „Wie viele: 1" stand beim ersten
// Anlauf in einem Feld, in das ein Satz passte — gesehen bei der
// Sichtprüfung am 2026-09-17.
const breiten: Record<string, string> = {
  zahl: 'flex flex-col gap-1 text-sm w-20',
  text: 'flex flex-col gap-1 text-sm w-36',
  zielwahl: 'flex flex-col gap-1 text-sm h-24 w-44',
}

// etikettklassen ist die Breite eines Feldes samt seinem Aufbau.
//
// **Die ganze Zeichenkette und keine Vorlage**: `keintextimjsx` weist eine
// Zeichenkette mit Platzhalter im Baum ab — sie ist die Gestalt, in der ein
// Oberflächentext durchrutscht. Eine Klassenangabe ist keiner, und trotzdem
// gilt die Regel: der Wächter kann beides nicht unterscheiden, und eine
// Ausnahme für ihn wäre eine Lücke für jeden.
function etikettklassen(art: string): string {
  return breiten[art] ?? 'flex flex-col gap-1 text-sm'
}

export function Befehlsfeld({
  feld,
  wert,
  setze,
  partie,
}: {
  feld: Feld
  wert: string | string[]
  setze: (wert: string | string[]) => void
  partie: Partie
}) {
  const texte = useTexte()
  const eigene = texte.seiten.handlungen

  return (
    <label className={etikettklassen(feld.art)}>
      <span className="font-medium">
        {handlungswort(eigene.handlungsfeld, feld.name)}
      </span>

      {feld.art === 'zahl' || feld.art === 'text' ? (
        <input
          type={feld.art === 'zahl' ? 'number' : 'text'}
          value={alsText(wert)}
          onChange={(ereignis) => {
            setze(ereignis.target.value)
          }}
          className={eingabeklassen}
        />
      ) : (
        <select
          multiple={feld.art === 'zielwahl'}
          value={feld.art === 'zielwahl' ? alsListe(wert) : alsText(wert)}
          onChange={(ereignis) => {
            setze(
              feld.art === 'zielwahl'
                ? gewaehlte(ereignis.target)
                : ereignis.target.value,
            )
          }}
          className={eingabeklassen}
        >
          {wahlen(feld, partie, texte).map((eine) => (
            <option key={eine.wert} value={eine.wert}>
              {eine.text}
            </option>
          ))}
        </select>
      )}
    </label>
  )
}

// Wahl ist ein Eintrag in einer Auswahlliste.
interface Wahl {
  wert: string
  text: string
}

// wahlen gibt die Einträge, unter denen ein Feld wählen lässt.
//
// **Aus der Partie und nicht aus einer Liste im Code** — wer sie fest
// hinschriebe, böte einen Sitz an, den es nicht gibt, und ein Objekt, das
// weggeräumt ist.
function wahlen(feld: Feld, partie: Partie, texte: Katalog): Wahl[] {
  switch (feld.art) {
    case 'zone':
      return zonen.map((zone) => ({
        wert: zone,
        text: zonenwort(texte.seiten.spielen, zone),
      }))

    case 'stelle':
      return stellen.map((stelle) => ({
        wert: stelle,
        text: texte.seiten.spielen[stellenwort(stelle)],
      }))

    case 'sitz':
      return partie.sitze.map((sitz) => ({
        wert: sitz.kennung,
        text: sitz.name,
      }))

    case 'manaart':
      return manaarten.map((art) => ({ wert: art, text: art }))

    case 'ruf':
      return [
        { wert: '', text: texte.seiten.handlungen.handlungRufOhne },
        { wert: 'kopf', text: texte.seiten.handlungen.handlungRufKopf },
        { wert: 'zahl', text: texte.seiten.handlungen.handlungRufZahl },
      ]

    case 'zielwahl':
      return [...tischwahlen(partie), ...sitzwahlen(partie)]

    // **Die wartenden Fähigkeiten** — CR 603.3b (M9, #1611). **Alle und
    // nicht nur die des Gefragten**: wer hier auswählt, hat die Leiste
    // offen, und der Kern weist ab, was ihm nicht gehört.
    case 'ausloeser':
      return partie.ausgeloest.map((einer) => ({
        wert: einer.kennung,
        text: handlungswort(texte.seiten.handlungen.ausloesername, einer.art),
      }))
  }

  // **Die Vorgabe ist die Objektwahl** — sie ist die häufigste, und ein
  // eigener Zweig für sie wäre einer, den kein Lauf von der Vorgabe
  // unterscheiden könnte (§10).
  return [
    { wert: '', text: texte.seiten.handlungen.handlungKeineWahl },
    ...tischwahlen(partie),
  ]
}

// tischwahlen sind die Objekte, die offen auf dem Tisch liegen.
function tischwahlen(partie: Partie): Wahl[] {
  return partie.objekte
    .filter((eines) => eines.zone === 'tisch' && !eines.geheim)
    .map((eines) => ({
      wert: eines.kennung,
      text: kartenwort(eines),
    }))
}

// sitzwahlen sind die Sitze — ein Zauber darf einen Spieler treffen
// (CR 115.1).
function sitzwahlen(partie: Partie): Wahl[] {
  return partie.sitze.map((sitz) => ({
    wert: sitz.kennung,
    text: sitz.name,
  }))
}

// stellenwort gibt den Katalogschlüssel zu einer Stelle.
function stellenwort(
  stelle: string,
): 'stelleOben' | 'stelleUnten' | 'stelleZufaellig' {
  if (stelle === 'oben') {
    return 'stelleOben'
  }

  if (stelle === 'unten') {
    return 'stelleUnten'
  }

  return 'stelleZufaellig'
}

// gewaehlte liest die gewählten Einträge einer Mehrfachauswahl.
function gewaehlte(feld: HTMLSelectElement): string[] {
  return [...feld.selectedOptions].map((eine) => eine.value)
}
