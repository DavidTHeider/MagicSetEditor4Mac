// **Ein Gestenvertrag für Maus, Finger und Tastatur** — GESTE
// (#2358),
// Befunde B-01, C-02, C-14; Artboards `Tisch09` und `Markttisch09`.
//
// # Was er festlegt
//
// **Eine Geste endet in einer Absicht** — Objekt, Quelle, Ziel, Ziel-Sitz,
// Profil und Epoche, dazu der Weg, auf dem sie kam. Ob eine Karte über den
// Tisch gezogen, mit dem Finger geschoben oder mit Leertaste, Pfeilen und ↩
// gelegt wird: was danach geschieht, sieht dieselbe Absicht. Der Tisch (P4,
// P5) und der Markttisch (P12) bauen darauf, und keiner von beiden schreibt
// eine zweite Tastenfolge.
//
// **Die Tasten** nach dem Artboard `Tisch09`: Leertaste greift, ←/→ und ↑/↓
// wandern zyklisch durch die Ziele, ↩ oder Leertaste legt ab, Escape bricht
// ab. Jeder Schritt hat eine Ansage — als Schlüssel und Werte; den Satz
// setzt, wer die Sprache kennt (`useGeste`).
//
// # Was er nicht weiß
//
// **Was eine Geste am Tisch bedeutet.** Ob eine Karte auf den Friedhof
// darf, entscheidet der Kern (Vermerk 0126, P3); hier gibt es Ziele, die der
// Aufrufer nennt, und eine Absicht, die er bekommt. **new** (§15): MSE2 hat
// keinen Spieltisch und keinen Handel. Gesucht über `src/` nach „drag"
// (Fassung f9d9356 vom 2021-07-06): gezogen wird dort an der Zeichenfläche
// des Symboleditors, beim Bildzuschnitt (`image_slice_window.cpp`), im
// Textfeld und beim Umordnen von Karten und Symbolteilen
// (`card_list.cpp`, Zeile 370; `symbol/part_list.cpp`) — keines davon
// bewegt ein Objekt zwischen Zonen oder zwischen zwei Seiten.

// Gestenziel ist ein Ort, an den ein Objekt gelegt werden kann.
export interface Gestenziel {
  kennung: string
  // name steht in der Ansage — aus dem Katalog des Aufrufers.
  name: string
  // sitz ist der Sitz, dem das Ziel gehört — am Markttisch gibt es keinen.
  sitz?: string
}

// Gestenbezug ist der Stand, aus dem eine Geste kommt: wer handelt, und in
// welcher Epoche des Bezugsobjekts (Partie, Handel). **Wechselt einer von
// beiden, gilt die Geste nicht mehr** — §7, Regel 1 und 2.
//
// **Die Epoche ist eine Zahl oder eine Fassung** — der Spieltisch zählt seinen
// Stand als Zeichenkette (`Partie.fassung`).
export interface Gestenbezug {
  profil: string
  epoche: number | string
}

export type Gestenweg = 'maus' | 'beruehrung' | 'tastatur'

// Gestenabsicht ist, was eine vollendete Geste meldet.
export interface Gestenabsicht {
  objekt: string
  quelle: string
  ziel: string
  zielSitz?: string
  profil: string
  epoche: number | string
  weg: Gestenweg
}

export type Gestenzustand =
  | { art: 'ruhend' }
  | {
      art: 'gegriffen'
      objekt: string
      quelle: string
      ziele: readonly Gestenziel[]
      stelle: number
      bezug: Gestenbezug
    }

// Gestenansage ist der Inhalt eines Satzes für die Live-Region.
export type Gestenansage =
  | {
      schluessel: 'gegriffen' | 'ziel'
      ziel: string
      stelle: number
      von: number
    }
  | { schluessel: 'abgelegt'; ziel: string }
  // **Zwei kommen nicht aus dem Vertrag, sondern aus dem Ausführen**
  // (`useGeste`): ein Ziel, das es nicht mehr gibt, und ein Stand, der
  // inzwischen ein anderer ist.
  | { schluessel: 'abgebrochen' | 'keinZiel' | 'zielFehlt' | 'veraltet' }

// Gestenschritt ist, was ein Greifen oder eine Taste bewirkt — **immer mit
// Ansage**: ein Schritt, den niemand hört, wäre einer, den niemand bemerkt.
export interface Gestenschritt {
  zustand: Gestenzustand
  absicht?: Gestenabsicht
  ansage: Gestenansage
}

// Griff ist, was der Aufrufer beim Greifen nennt.
export interface Griff {
  objekt: string
  quelle: string
  ziele: readonly Gestenziel[]
}

// absichtAus baut die Absicht — einmal, für jeden Weg.
function absichtAus(
  objekt: string,
  quelle: string,
  ziel: Gestenziel,
  bezug: Gestenbezug,
  weg: Gestenweg,
): Gestenabsicht {
  return {
    objekt,
    quelle,
    ziel: ziel.kennung,
    ...(ziel.sitz === undefined ? {} : { zielSitz: ziel.sitz }),
    profil: bezug.profil,
    epoche: bezug.epoche,
    weg,
  }
}

// greifen nimmt ein Objekt auf — **nur, wenn es ein Ziel gibt**: eine
// Geste ohne Ziel könnte nur abgebrochen werden, und das sagt die Ansage
// gleich.
export function greifen(griff: Griff, bezug: Gestenbezug): Gestenschritt {
  const erstes = griff.ziele[0]
  if (erstes === undefined) {
    return { zustand: { art: 'ruhend' }, ansage: { schluessel: 'keinZiel' } }
  }

  return {
    zustand: { art: 'gegriffen', ...griff, stelle: 0, bezug },
    ansage: {
      schluessel: 'gegriffen',
      ziel: erstes.name,
      stelle: 1,
      von: griff.ziele.length,
    },
  }
}

// gestentaste deutet eine Taste im Zustand der Geste — `undefined`, wenn sie
// die Geste nicht betrifft: eine andere Taste, oder keine Geste läuft.
export function gestentaste(
  zustand: Gestenzustand,
  taste: string,
): Gestenschritt | undefined {
  if (zustand.art === 'ruhend') {
    return undefined
  }

  const anzahl = zustand.ziele.length
  const wandern = (richtung: 1 | -1): Gestenschritt | undefined => {
    const stelle = (zustand.stelle + richtung + anzahl) % anzahl
    const ziel = zustand.ziele[stelle]
    if (ziel === undefined) {
      return undefined
    }

    return {
      zustand: { ...zustand, stelle },
      ansage: {
        schluessel: 'ziel',
        ziel: ziel.name,
        stelle: stelle + 1,
        von: anzahl,
      },
    }
  }

  switch (taste) {
    case 'ArrowRight':
    case 'ArrowDown':
      return wandern(1)
    case 'ArrowLeft':
    case 'ArrowUp':
      return wandern(-1)
    case 'Enter':
    case ' ': {
      const ziel = zustand.ziele[zustand.stelle]
      if (ziel === undefined) {
        return undefined
      }

      return {
        zustand: { art: 'ruhend' },
        absicht: absichtAus(
          zustand.objekt,
          zustand.quelle,
          ziel,
          zustand.bezug,
          'tastatur',
        ),
        ansage: { schluessel: 'abgelegt', ziel: ziel.name },
      }
    }
    case 'Escape':
      return {
        zustand: { art: 'ruhend' },
        ansage: { schluessel: 'abgebrochen' },
      }
    default:
      return undefined
  }
}

// Zug ist ein beendeter Zug mit Maus oder Finger — das Ziel hat die Fläche
// unter dem Zeiger gefunden (`tischziehen.ts`).
export interface Zug {
  objekt: string
  quelle: string
  ziel: Gestenziel
}

// absichtAusZug macht aus einem Zug dieselbe Absicht wie die Tastatur.
export function absichtAusZug(
  zug: Zug,
  bezug: Gestenbezug,
  weg: Exclude<Gestenweg, 'tastatur'>,
): Gestenabsicht {
  return absichtAus(zug.objekt, zug.quelle, zug.ziel, bezug, weg)
}
