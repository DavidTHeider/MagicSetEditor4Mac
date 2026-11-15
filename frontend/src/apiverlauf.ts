// **Der Verlauf eines Sets — ansehen, zurücknehmen, wiederherstellen.**
//
// Herausgelöst aus `api.ts`, als die Datei mit den Metriken (Werkstück O9,
// #1433) über die Zeilengrenze aus §2 wuchs. **Der Schnitt folgt der Frage**:
// dort steht, was die Fassade insgesamt kann; hier, wie ein Set zu seiner
// eigenen Geschichte kommt.
import type { BewegungAntwort } from './formen'
import type { VerlaufAntwort } from './antwortenverlauf'

export interface Verlaufsapi {
  Verlauf: (setID: string) => Promise<VerlaufAntwort>

  // vorgang nimmt genau diesen Schritt (U6, #1417); leer heißt: den jüngsten.
  Zuruecknehmen: (setID: string, vorgang: string) => Promise<BewegungAntwort>
  Wiederherstellen: (setID: string, vorgang: string) => Promise<BewegungAntwort>
}
