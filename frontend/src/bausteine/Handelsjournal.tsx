// **Das Journal eines Handels, im eigenen Kasten** — UXR-15
// (#1923).
//
// # Warum getrennt von der Buchung
//
// **Weil Nachsehen kein Schritt ist.** Bis hierher las die Buchung das
// Journal nach jedem ihrer drei Knöpfe und zeigte es darunter; wer nur
// wissen wollte, was sich bewegt hat, musste dafür einen Schritt gehen. Das
// Issue fragt ausdrücklich, ob das Journal vom laufenden Vorgang getrennt
// steht — jetzt steht es in „Verbuchen" neben Unterwegs und Verlauf.
//
// # Wann es fragt
//
// **Erst, wenn jemand den Kasten aufklappt** — er zeichnet seinen Inhalt
// beim ersten Öffnen —, **und nach jedem Schritt der Buchung wieder**
// (`stand`). Ein Lesefehler steht hier und nimmt keine Buchung zurück (§7,
// Regel 3): die Buchung hat ihre eigene Zeile.
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'
import { Journalliste } from './Journalliste'
import { Ladefehler, Wirdgeladen } from './Ladehinweise'

interface Handelsjournaleigenschaften {
  // vorgang ist der gewählte Handel — leer heißt: noch keiner.
  vorgang: string

  // stand zählt die Schritte der Buchung — jeder lädt nach.
  stand: number
}

export function Handelsjournal({
  vorgang,
  stand,
}: Handelsjournaleigenschaften) {
  const eigene = useTexte().seiten.handeln

  const journal = useLaden(async () => {
    if (vorgang === '') {
      return []
    }

    const antwort = await api().Handelsjournal(vorgang)

    return auspacken(antwort, () => antwort.zeilen)
  }, [vorgang, stand])

  if (journal.fehler !== undefined) {
    return (
      <Ladefehler fehler={journal.fehler} wiederholen={journal.nachladen} />
    )
  }

  if (journal.daten === undefined) {
    return <Wirdgeladen />
  }

  return (
    <Journalliste
      titel={eigene.buchungJournal}
      zeilen={journal.daten}
      leer={eigene.buchungJournalLeer}
    />
  )
}
