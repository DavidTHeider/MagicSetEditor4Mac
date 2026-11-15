// **Die Knöpfe einer Rückfrage, an einer Stelle** — #2412.
//
// `Bestandspflege` und `Bestandschiffre` führten je eine eigene Kopie, und in
// der einen fehlte die Größe: „Zurückspielen" stieß an den Rand, stand größer
// als „Abbrechen" daneben, und der Knopf war niedriger als sein Nachbar
// (gefunden in PERF01 am nativen Fenster). **Wer eine Rückfrage mit zwei
// Knöpfen baut, nimmt beide von hier** — `knopfklassen.test.ts` hält, dass
// keine Datei einen eigenen Hauptknopf führt.
//
// **`text-marketext` auf `bg-marke`** steht gemessen in `docs/KONTRASTE.md`,
// wie beim Hauptgewicht von `Knopf`.
const groesse = 'px-3 py-1.5 text-sm font-medium'

const fokus =
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring disabled:opacity-60'

// **Die Rückmeldung bei Zeiger und Druck** — AP-4
// (#2361),
// Befund B-11, Entwurf `Zustaende09`. Hover hebt den Grund, Drücken senkt den
// Knopf um einen Punkt — nur Farbe, Deckkraft und Lage, in `--dauer-kurz`
// (Vermerk 0065, Festlegung 10), unter wenig Bewegung 0 ms. **Ein gesperrter
// Knopf antwortet nicht**: er sagt es mit seiner Deckkraft, und ein Hover-Ton
// darauf versprach einen Klick, der nichts tut. `Knopf` nimmt dieselben.
export const rueckmeldung =
  'transition-[color,background-color,border-color,opacity,translate] duration-(--dauer-kurz) ' +
  'enabled:not-aria-disabled:active:translate-y-px'

// **Der Hover-Ton je Gewicht**: der Hauptknopf behält seine Marke und wird um
// ein Zehntel durchscheinender, die übrigen bekommen den Grund `betont`. Die
// Paare stehen gemessen in `docs/KONTRASTE.md`.
export const hoverHaupt = 'enabled:not-aria-disabled:hover:bg-marke/90'
export const hoverNeben = 'enabled:not-aria-disabled:hover:bg-betont'

export const nebenknopf = `rounded-md border border-kante ${groesse} ${fokus} ${rueckmeldung} ${hoverNeben}`

export const hauptknopf = `rounded-md bg-marke ${groesse} text-marketext ${fokus} ${rueckmeldung} ${hoverHaupt}`
