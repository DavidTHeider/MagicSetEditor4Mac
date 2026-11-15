// **Wie der Tisch aussieht und ob er selbst ordnet** — die drei
// Einstellungen des Spieltischs, herausgezogen aus `Partietisch.tsx`, als P4
// (#2372)
// dort die Bedienung der Karten brauchte und die Datei an der Zeilengrenze
// aus §2 stand. Der Schnitt folgt der Frage: dort, was der Tisch zeigt; hier,
// was der Mensch über ihn eingestellt hat.
import { useEinstellungen } from './useEinstellungen'
import { useNutzereinstellungen } from './useNutzereinstellungen'

// Autoanordnungschluessel ist der Name, unter dem die Anordnung liegt —
// PLAY-04.AK-03
// (#1712).
//
// **Derselbe Schlüssel wie in `domain.EinstellungAutoanordnung`.** Ein
// Vertragslauf hält beide Seiten zusammen; hier steht er, weil TypeScript
// die Go-Konstante nicht kennt.
export const Autoanordnungschluessel = 'tisch.autoanordnung'

// Tischbildschluessel ist der Name, unter dem die Wahl liegt.
//
// **Mit Bereich davor**, wie jede andere Nutzereinstellung: `tisch.bild`
// sagt, wozu sie gehört, und ein nacktes `bild` sagte es nicht.
export const Tischbildschluessel = 'tisch.bild'

// Beschriftungsschluessel ist der Name, unter dem steht, ob und wie groß der
// Tisch Zonen und Sitze beschriftet — `domain.EinstellungTischbeschriftung`,
// seit dem 2026-09-24.
export const Beschriftungsschluessel = 'tisch.beschriftung'

export interface Tischeinstellungen {
  bildname: string
  setzeBild: (name: string) => void
  ordnetSelbst: boolean
  beschriftungsstufe: string
}

export function useTischeinstellungen(): Tischeinstellungen {
  // **Das Tischbild gehört dem Menschen, nicht dem Profil** — wer seinen Tisch
  // als Dungeon sieht, sieht ihn so in jedem seiner Profile
  // (`useNutzereinstellungen`, Werkstück 8 von Phase 4.12).
  const einstellungen = useNutzereinstellungen()
  const profileinstellungen = useEinstellungen()
  const tischbildname = einstellungen.wert(Tischbildschluessel) || 'holz'

  // **Ob der Tisch selbst ordnet** — PLAY-04.AK-03
  // (#1712).
  // Die Einstellung gehört dem Menschen, nicht dem Profil: wer seine
  // Kreaturen nach Angriffsreihen legt, legt sie so mit jedem Deck.
  // **Aus den Profileinstellungen und nicht aus den Nutzereinstellungen**:
  // die Ebene sagt, **wem die Frage gehört** (`nutzer`), die Tafel, **wo der
  // Wert liegt** — und das ist `settings`, wie bei der Feldtiefe. Der
  // Kopf von `einstellungsebene.go` sagt es ausdrücklich.
  const ordnetSelbst =
    profileinstellungen.wert(Autoanordnungschluessel) !== 'aus'

  // **Und wie er beschriftet ist** — aus derselben Tafel, aus demselben
  // Grund. Ein leerer Wert ist die Vorgabe.
  const beschriftungsstufe =
    profileinstellungen.wert(Beschriftungsschluessel) || 'normal'

  return {
    bildname: tischbildname,
    setzeBild: (name) => {
      einstellungen.setze(Tischbildschluessel, name)
    },
    ordnetSelbst,
    beschriftungsstufe,
  }
}
