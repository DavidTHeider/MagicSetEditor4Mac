// **Farbwelt, Fensterbild und eigene Merkmale am Wurzelelement** — UX-09
// (#1681).
//
// **Neben `erscheinungsbild.ts` und nicht darin**, weil die Frage eine andere
// ist: dort, wie hell und wie groß; hier, welche Farben, welches Bild und was
// davon lesbar bleibt.
//
// # Hoher Kontrast sticht alles hier
//
// **Wer mehr Kontrast verlangt — in der Anwendung oder im System —, bekommt
// die gemessene Tinte** (`barrierefreiheit.css`) und nichts von dem, was hier
// gesetzt wird: keine Farbwelt, kein Bild, keine eigene Farbe. Dieselbe
// Richtung wie bei jeder Übersteuerung: sie macht strenger, nie lockerer.
import {
  lesbarkeitsbefunde,
  type Lesbarkeitsbefund,
  type Messlage,
} from './kontrast'

// farbwelten ist die Reihe aus `domain.EinstellungFarbwelt`.
export const farbwelten = ['pergament', 'klar', 'schiefer', 'moos'] as const

// merkmale sind die zehn, die sich ändern lassen — dieselben wie
// `domain.Gestaltungsmerkmale`.
export const merkmale = [
  '--mse-grund',
  '--mse-flaeche',
  '--mse-text',
  '--mse-leise',
  '--mse-kante',
  '--mse-marke',
  '--mse-betont',
  '--mse-ring',
  '--mse-marketext',
  '--mse-warnung',
] as const

// setzeFarbwelt schreibt die Wahl ans Wurzelelement.
//
// **Pergament entfernt das Attribut**: es ist, was `stil.css` ohnehin sagt.
// Eine unbekannte Wahl ebenso — sie kann nur aus einem von Hand geänderten
// Bestand kommen, und `domain.PruefeEinstellung` lässt sie nicht durch.
export function setzeFarbwelt(
  wurzel: HTMLElement,
  wahl: string,
  kontrastHoch: boolean,
): void {
  if (!kontrastHoch && wahl !== 'pergament' && istFarbwelt(wahl)) {
    wurzel.setAttribute('data-farbwelt', wahl)

    return
  }

  wurzel.removeAttribute('data-farbwelt')
}

function istFarbwelt(wahl: string): boolean {
  return (farbwelten as readonly string[]).includes(wahl)
}

// fensterbildadresse ist die einzige Gestalt, die als Bild ans Fenster darf:
// der eigene Bildweg, eine Kennung, ein Stand.
//
// **Keine fremde Adresse** (AK-04, „keine unbemerkten Netzbezüge"): was die
// Bindung zurückgibt, wird hier noch einmal gegen diese Form gehalten, bevor
// es in ein `url()` kommt.
const fensterbildadresse = /^\/fensterbild\/\d+\.png(\?stand=\d+)?$/

// setzeHintergrund legt das Fensterbild unter den Grund — oder nimmt es weg.
export function setzeHintergrund(
  wurzel: HTMLElement,
  modus: string,
  adresse: string,
  kontrastHoch: boolean,
): void {
  const zeigen =
    !kontrastHoch &&
    ['fuellen', 'einpassen', 'kacheln'].includes(modus) &&
    fensterbildadresse.test(adresse)

  if (!zeigen) {
    wurzel.removeAttribute('data-hintergrund')
    wurzel.style.removeProperty('--mse-fensterbild')

    return
  }

  wurzel.style.setProperty('--mse-fensterbild', `url("${adresse}")`)
  wurzel.setAttribute('data-hintergrund', modus)
}

// Merkmalslesung ist das Ergebnis des Lesens: die Angaben — oder die erste
// Zeile, die die Form verfehlt.
export type Merkmalslesung =
  | { angaben: [string, string][]; zeile?: undefined }
  | { angaben?: undefined; zeile: number }

// eigeneMerkmale liest die Zeilen — dieselbe Form wie
// `domain.PruefeGestaltung`.
//
// **Noch einmal geprüft, obwohl Go schon prüft**: was im Bestand steht,
// kann jemand von Hand hineingeschrieben haben, und dieser Weg setzt Werte
// ans Wurzelelement. Eine Zeile, die hier nicht passt, setzt **nichts** —
// auch die übrigen nicht. **Und die Zeile geht zurück**, damit das Feld den
// Befund an der Stelle nennt (§6, §8).
export function eigeneMerkmale(wert: string): Merkmalslesung {
  const angaben: [string, string][] = []

  for (const [stelle, roh] of wert.split('\n').entries()) {
    const zeile = roh.trim()
    if (zeile === '') {
      continue
    }

    const [, name, farbe] =
      /^(--mse-[a-z]+)\s*:\s*(#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6}))\s*;?$/.exec(
        zeile,
      ) ?? []
    if (
      name === undefined ||
      farbe === undefined ||
      !(merkmale as readonly string[]).includes(name) ||
      angaben.some(([schon]) => schon === name)
    ) {
      return { zeile: stelle + 1 }
    }

    angaben.push([name, farbe.toLowerCase()])
  }

  return { angaben }
}

// setzeEigeneMerkmale setzt die eigenen Farben — **und nimmt sie zurück, wenn
// sie nicht lesbar sind** (AK-03).
//
// # Die Sperre
//
// **Erst setzen, dann messen, dann entscheiden** — im selben Durchlauf, bevor
// das Fenster zeichnet (`useLayoutEffect` im Aufrufer). Gemessen wird am
// Wurzelelement, also nach Farbwelt und Erscheinungsbild: eine Farbe, die im
// hellen Bild trägt, kann im dunklen unlesbar sein, und dann gilt sie dort
// nicht. Die Befunde gehen zurück, damit die Einstellungsseite sagt, **warum**
// nichts geschah.
//
// **`lage` sagt, worauf gemessen wird** — helles oder dunkles Bild, und ob
// ein Fensterbild unter dem Grund liegt (`kontrast.ts`, `Messlage`).
//
// **Über eigenen Farben liegt keine Faser** (Zweitblick über Phase 6.3, S4):
// auf Pergament kostete sie rund 16 % des Verhältnisses, und die Sperre misst
// reine Farben.
export function setzeEigeneMerkmale(
  wurzel: HTMLElement,
  wert: string,
  kontrastHoch: boolean,
  lage: Messlage = { dunkel: false },
  lies: (merkmal: string) => string = (merkmal) =>
    getComputedStyle(wurzel).getPropertyValue(merkmal).trim(),
): Lesbarkeitsbefund[] {
  for (const merkmal of merkmale) {
    wurzel.style.removeProperty(merkmal)
  }
  wurzel.style.removeProperty('--mse-faser')

  const { angaben } = eigeneMerkmale(wert)
  if (kontrastHoch || angaben === undefined || angaben.length === 0) {
    return []
  }

  for (const [name, farbe] of angaben) {
    wurzel.style.setProperty(name, farbe)
  }

  const befunde = lesbarkeitsbefunde(lies, lage)
  if (befunde.length > 0) {
    for (const [name] of angaben) {
      wurzel.style.removeProperty(name)
    }

    return befunde
  }

  wurzel.style.setProperty('--mse-faser', 'none')

  return befunde
}

// schleierAnteil liest, wie viel Grund über dem Fensterbild liegt —
// `--mse-schleier` als Prozent (`themen.css`). **Ohne lesbaren Wert gilt der
// ganze Grund**: dann misst die Sperre, was ohne Bild gilt, statt eine Zahl
// zu raten.
export function schleierAnteil(wurzel: HTMLElement): number {
  const wert = Number.parseFloat(
    getComputedStyle(wurzel).getPropertyValue('--mse-schleier'),
  )

  return Number.isFinite(wert) ? wert / 100 : 1
}
