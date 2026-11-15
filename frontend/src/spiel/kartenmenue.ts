// **Das Menü einer Karte aus der Auskunft des Kerns** — P4
// (#2372).
// Herausgezogen aus `Zonenkarten.tsx`, weil der gezeichnete Tisch dasselbe
// Menü bekommt: zwei Menüs für dieselbe Karte liefen auseinander.
//
// # Woraus es besteht
//
// **Aus `handlungen` und nicht aus `erlaubt`** — der Kern sagt je Objekt,
// was an ihm gilt (P3); eine Handkarte bietet kein Tappen mehr an. **Nur,
// was ohne ein weiteres Feld geht**, steht direkt darin: ein Menüeintrag ist
// eine Handlung und kein Formular. **„Auf den Friedhof“ geht über die
// gedeutete Geste** (`objekt.ziehen`), denn was dabei geschieht, entscheidet
// das Profil der Partie. Alles andere führt der letzte Eintrag in die
// Handlungsleiste — **keine Aktion, die es nur im Kontextmenü gibt**
// (#1413),
// und ein Menü ohne Eintrag wäre eine Sackgasse.
import type { Objekthandlung, Spielobjekt } from '../apipartie'
import type { Menueeintrag } from '../bausteine/Menueliste'
import { handlungswort } from '../seiten/partiewoerter'
import { einsetzen } from '../sprache/einsetzen'
import { kartenwort } from './feldwerte'
import { ohneFeld } from './kartengriff'

export function kartenmenue(
  eigene: { kartenmenueWeitere: string; kartenmenueFriedhof: string },
  worte: { handlungsname: Record<string, unknown> },
  objekt: { kennung: string; eigner: string },
  handlung: Objekthandlung | undefined,
  befehlen: (art: string, nutzlast: Record<string, unknown>) => void,
  waehle: (kennung: string) => void,
): Menueeintrag[] {
  const arten = handlung?.handlungen ?? []

  // **Was schreibt, sagt es** (#2482) — für einen Leser stehen diese
  // Zeilen gesperrt und erklärt da, im Menü wie in der Menüleiste.
  const sofort = arten.filter(ohneFeld).map((art) => ({
    schluessel: art,
    name: handlungswort(worte.handlungsname, art),
    schreibt: true,
    aufWahl: () => {
      befehlen(art, { objekt: objekt.kennung })
    },
  }))

  const friedhof = arten.includes('objekt.ziehen')
    ? [
        {
          schluessel: 'friedhof',
          name: eigene.kartenmenueFriedhof,
          schreibt: true,
          aufWahl: () => {
            befehlen('objekt.ziehen', {
              objekt: objekt.kennung,
              zone: 'friedhof',
              sitz: objekt.eigner,
            })
          },
        },
      ]
    : []

  return [
    ...sofort,
    ...friedhof,
    {
      schluessel: 'weitere',
      name: eigene.kartenmenueWeitere,
      aufWahl: () => {
        waehle(objekt.kennung)
      },
    },
  ]
}

// kartenbeschriftung sagt, was auf einem Kartenknopf steht.
//
// **Ein Geheimnis bleibt eines.** Wo nichts hineingesehen werden darf, steht
// das Wort dafür — und nicht ein leerer Knopf, den niemand deuten kann.
//
// **Der Name über `kartenwort`** und nicht über `objekt.name`: der Kern
// schickt ihn bei einer Karte in ihren Eigenschaften, am Objekt nur bei einem
// Spielstein; sonst hieße eine Karte nach ihrer Kennung, etwa „p0-1".
export function kartenbeschriftung(
  eigene: { geheim: string; getappt: string },
  objekt: Spielobjekt,
): string {
  const name = objekt.geheim ? eigene.geheim : kartenwort(objekt)

  return objekt.getappt ? `${name} (${eigene.getappt})` : name
}

// menueZurKarte gibt Bezeichnung und Einträge des Menüs einer Karte — für
// die Liste und den gezeichneten Tisch dasselbe.
export function menueZurKarte(
  eigene: {
    kartenmenue: string
    kartenmenueWeitere: string
    kartenmenueFriedhof: string
    geheim: string
    getappt: string
  },
  worte: { handlungsname: Record<string, unknown> },
  objekt: Spielobjekt,
  handlung: Objekthandlung | undefined,
  befehlen: (art: string, nutzlast: Record<string, unknown>) => void,
  waehle: (kennung: string) => void,
): { bezeichnung: string; eintraege: Menueeintrag[] } {
  return {
    bezeichnung: einsetzen(eigene.kartenmenue, {
      name: kartenbeschriftung(eigene, objekt),
    }),
    eintraege: kartenmenue(eigene, worte, objekt, handlung, befehlen, waehle),
  }
}
