// **Die Vorgabe des gezeigten Profils, für die Stellen, die sie benutzen** —
// Werkstück 15 von Phase 4.12
// (#1351),
// Punkte 4 bis 7.
//
// # Warum es diesen Haken gibt
//
// **Weil eine Vorgabe ohne Wirkung eine Zeile im Bestand ist.** Wer ein
// Vorgabespiel setzt, will es im Formular für ein neues Set vorgewählt sehen —
// sonst hat er etwas eingestellt, das nichts tut.
//
// # Warum zwei Rufe und nicht einer
//
// **Weil das gezeigte Profil und seine Vorgaben zwei Fragen sind.** Ein Ruf,
// der beides gäbe, wäre eine dritte Bindung neben zwei vorhandenen — und die
// Antwort auf „wer arbeitet gerade" hat mit „womit fängt ein neuer Datensatz
// an" nichts zu tun.
//
// **Ein verwaister Wert kommt trotzdem zurück.** Wer ihn benutzt, bekommt vom
// Bestand ein Nein mit der richtigen Meldung; ihn hier zu verschlucken machte
// aus einer falschen Vorgabe eine unsichtbare.
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden } from './useLaden'

// useVorgabe gibt die Werte **einer** Art — leer, solange nichts dasteht.
//
// **Eine Liste, auch bei den einwertigen Arten**: zwei Gestalten zwängen jeden
// Aufrufer zu wissen, welche gerade gilt. Wer einen Wert will, nimmt den
// ersten.
export function useVorgabe(art: string): readonly string[] {
  const geladen = useLaden(async () => {
    const gezeigt = await api().GezeigtesProfil()
    const profil = auspacken(gezeigt, () => gezeigt.profil)

    if (profil === null) {
      return []
    }

    const antwort = await api().Profilvorgaben(profil.id)
    const vorgaben = auspacken(antwort, () => antwort.vorgaben)

    return vorgaben.find((stand) => stand.art === art)?.werte ?? []
  }, [art])

  // **Ein Fehler bleibt hier stumm**, und das ist Absicht: eine Vorgabe ist
  // Bequemlichkeit. Wer ein Set anlegen will, soll das können, auch wenn die
  // Vorwahl nicht zu holen war — eine Meldung darüber stünde im Weg und sagte
  // nichts, was jemand täte.
  return geladen.daten ?? []
}
