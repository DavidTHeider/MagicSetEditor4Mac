// **Die Dateien aller Pakete auf einmal** — für den Vorlagenbaum der
// Vorlagenwerkstatt (Werkstück U7 von Phase 5.1,
// #1418).
//
// # Warum alle, und nicht das gewählte
//
// Der Kernweg `paket` soll fünf Handlungen kosten (UX-ARCHITEKTUR): ⌘5,
// Datei, in den Text, tippen, Speichern. Ein Baum, der seine Dateien erst
// beim Aufklappen holt, kostete eine sechste — und der Filter über
// Dateinamen könnte nur suchen, was schon aufgeklappt war. Ein Ruf je Paket
// ist ein Verzeichnislauf; bei dreihundert Paketen sind das dreihundert
// kleine Rufe, nebeneinander.
//
// # Warum ein Haken mit einem `useLaden`
//
// Hooks laufen nicht in Schleifen, und ein Bauteil je Paket, das seine
// Dateien nach oben meldet, wäre abgeleiteter Zustand über Effekte. Ein
// einziger Ladevorgang über `Promise.all` gibt eine Liste, und ein
// Paketwechsel in der Liste räumt sie wie jeden anderen Lesestand (§7).
//
// **Ein Fehler eines Pakets bleibt an seinem Zweig.** `Promise.all` würde
// mit dem ersten Fehler abbrechen und die anderen Pakete mitreißen; deshalb
// fängt jeder Ruf seinen eigenen.
//
// **Ein mitgeliefertes Paket ist kein Fehler.** Die Werkstatt liest nur unter
// ihrer Wurzel (`service.Paketwerkstatt.ort`), und ein Paket aus dem Bündel
// liegt außerhalb — Änderungen dort verschwänden mit der nächsten Fassung.
// Die Werkstatt meldet das als Validierungsfehler; der Baum sagt es als
// Satz und nicht in Rot, denn niemand hat etwas falsch gemacht.
import { Bindungsfehler } from '../bindungsfehler'
import { api } from '../fassade'
import { meldungAus } from '../meldung'
import { auspacken } from '../umschlag'
import { useTexte } from '../sprache/kontext'
import { useLaden } from './useLaden'

// Zweig ist ein Paket mit seinen bearbeitbaren Dateien.
export interface Zweig {
  paket: string

  // name ist die Aufschrift — der letzte Teil des Pfads.
  name: string

  dateien: string[]

  // beiwerk sind Bilder und Schriften — im Baum, aber nicht im Editor
  // (FMT-02, #1691).
  beiwerk: string[]
  fehler: string | undefined

  // nurLesbar sagt, dass das Paket außerhalb der Werkstatt liegt.
  nurLesbar: boolean
}

export interface Vorlagenbaumstand {
  zweige: Zweig[]
  laedt: boolean
}

// ordner gibt den letzten Teil eines Pfads.
//
// **Ohne `?? ''`.** `lastIndexOf` gibt −1, wenn kein Schrägstrich dasteht,
// und `slice(0)` dann den ganzen Namen.
export function ordner(pfad: string): string {
  return pfad.slice(pfad.lastIndexOf('/') + 1)
}

// **`fassung` stößt das Nachladen an** — nach einem Handgriff am Dateibaum
// (FMT-02, #1691) steht eine Datei mehr oder weniger da, und die Liste der
// Pakete hat sich dabei nicht geändert.
export function useVorlagenbaum(
  pakete: readonly string[],
  fassung = 0,
): Vorlagenbaumstand {
  const texte = useTexte()
  const schluessel = pakete.join('\n')

  const stand = useLaden<Zweig[]>(
    () =>
      Promise.all(
        pakete.map(async (paket): Promise<Zweig> => {
          const zweig = {
            paket,
            name: ordner(paket),
            dateien: [] as string[],
            beiwerk: [] as string[],
          }

          try {
            const antwort = await api().Paketdateien(paket)

            const gelesen = auspacken(antwort, () => antwort)

            return {
              ...zweig,
              dateien: gelesen.dateien,
              beiwerk: gelesen.beiwerk,
              fehler: undefined,
              nurLesbar: false,
            }
          } catch (ursache: unknown) {
            const nurLesbar =
              ursache instanceof Bindungsfehler &&
              ursache.code === 'validierung'

            return {
              ...zweig,
              fehler: nurLesbar ? undefined : meldungAus(ursache, texte),
              nurLesbar,
            }
          }
        }),
      ),
    // **Über den Inhalt der Liste, nicht über ihre Identität** — der Aufrufer
    // baut sie bei jedem Zeichnen neu.
    [schluessel, texte, fassung],
  )

  return { zweige: stand.daten ?? [], laedt: stand.laedt }
}
