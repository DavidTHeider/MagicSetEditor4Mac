// **Ein Profil für Vorrichtungen** — Werkstück 9 von Phase 4.12
// (#1351).
//
// # Warum es diesen Helfer gibt
//
// **Weil ein `ProfilDTO` seit Werkstück 10 zehn Felder hat.** Vierzehn Läufe
// bauten es von Hand, und jedes neue Feld hieß vierzehn Änderungen — von
// denen jede vergessen werden kann. **Der Helfer ist die eine Stelle**: wer
// ein Feld ergänzt, ergänzt es hier.
//
// **Und die Vorgabe ist der leere Wert**, nicht ein ausgedachter: ein Doppel,
// das großzügiger ist als der Betrieb, prüft die Vorrichtung und nicht die
// Anwendung (§3).
import type { ProfilDTO } from '../api'

export function probeprofil(teile: Partial<ProfilDTO> = {}): ProfilDTO {
  return {
    id: '10',
    nutzerID: '1',
    name: 'Mein Profil',
    subslogan: '',
    kurzebeschreibung: '',
    langebeschreibung: '',
    badge: '',
    technisch: false,
    ablauf: '',
    fassung: '1',
    ...teile,
  }
}
