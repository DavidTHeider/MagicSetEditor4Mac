// **Kontrast gerechnet** — UX-09.AK-03
// (#1681):
// *„Benutzerdefinierte Schriften und Farben werden auf Lesbarkeit, Kontrast
// und Beschriftungsüberläufe geprüft."*
//
// **Nach WCAG 2.2**: relative Leuchtdichte aus linearisiertem sRGB,
// Verhältnis `(L1 + 0,05) / (L2 + 0,05)`. Dieselbe Rechnung wie
// `scripts/pergamentmass.mjs`, nur für reine Farben — **eine Faser gibt es
// über eigenen Farben nicht**: `setzeEigeneMerkmale` nimmt sie weg. Bis zum
// Zweitblick über Phase 6.3 stand dieser Satz hier, und die Faser blieb auf
// Pergament liegen (S4).

// lesbarePaare sind die Paare, auf denen die Oberfläche steht, mit ihrer
// Schwelle: Text 4,5 : 1 (WCAG 1.4.3), Ränder und Fokus 3 : 1 (1.4.11).
//
// **Dieselben wie in `docs/KONTRASTE.md`** unter „Die Gestaltungssprache auf
// Pergament" — und `themen.test.ts` rechnet jede Farbwelt gegen sie.
export const lesbarePaare: readonly (readonly [string, string, number])[] = [
  ['--mse-text', '--mse-grund', 4.5],
  ['--mse-text', '--mse-flaeche', 4.5],
  ['--mse-text', '--mse-betont', 4.5],
  ['--mse-leise', '--mse-grund', 4.5],
  ['--mse-leise', '--mse-flaeche', 4.5],
  ['--mse-leise', '--mse-betont', 4.5],
  ['--mse-marke', '--mse-grund', 4.5],
  ['--mse-marke', '--mse-flaeche', 4.5],
  ['--mse-warnung', '--mse-grund', 4.5],
  ['--mse-warnung', '--mse-flaeche', 4.5],
  ['--mse-warnung', '--mse-betont', 4.5],
  ['--mse-marketext', '--mse-marke', 4.5],
  ['--mse-kante', '--mse-grund', 3],
  ['--mse-kante', '--mse-flaeche', 3],
  ['--mse-kante', '--mse-betont', 3],
  ['--mse-ring', '--mse-grund', 3],
  ['--mse-ring', '--mse-flaeche', 3],
  ['--mse-ring', '--mse-betont', 3],
]

// statusfarben sind die Farben der Meldungen, die **kein eigenes Merkmal**
// sind: `text-hinweis` für Warnungen und `text-erfolg` für Erfolg (AP-2,
// #2359). Sie folgen dem Erscheinungsbild, nicht den eigenen Farben, und
// deshalb misst die Sperre sie gegen deren Flächen mit: bis zum Zweitblick
// über Phase 6.3 stand bei dunklen eigenen Farben im hellen Bild jede
// Fehlermeldung bei 2,07 : 1 (S4).
//
// **Der Fehler steht hier nicht mehr.** Bis AP-2 trug er Tailwinds
// `red-800`/`red-300` und stand deshalb hier; seither ist er `text-fehler`
// aus `--mse-warnung`, also ein Merkmal, und die Paare oben messen ihn — mit
// der Farbe, die wirklich gilt, auch wenn jemand sie selbst gesetzt hat.
//
// **Die Werte sind Tailwinds OKLCH** aus `tailwindcss/theme.css` (4.x), in
// sRGB umgerechnet und auf den Farbraum geklemmt — dieselbe Rechnung wie in
// `docs/KONTRASTE.md`, und dieselben, die `stil.css` als `--mse-hinweis` und
// `--mse-erfolg` setzt.
export const statusfarben: Record<
  'hell' | 'dunkel',
  readonly (readonly [string, string])[]
> = {
  hell: [
    ['status:warnung', '#973c00'],
    ['status:erfolg', '#006045'],
  ],
  dunkel: [
    ['status:warnung', '#ffd230'],
    ['status:erfolg', '#a4f4cf'],
  ],
}

// Messlage ist, worauf gemessen wird: das Erscheinungsbild, und — wenn ein
// Fensterbild unter dem Grund liegt — der Anteil des Grunds im Schleier.
export interface Messlage {
  dunkel: boolean
  schleier?: number
}

// Lesbarkeitsbefund ist ein Paar unter seiner Schwelle.
export interface Lesbarkeitsbefund {
  vorn: string
  hinten: string
  verhaeltnis: number
  schwelle: number
}

// kanal linearisiert einen sRGB-Kanal.
function kanal(wert: number): number {
  const anteil = wert / 255

  return anteil <= 0.04045 ? anteil / 12.92 : ((anteil + 0.055) / 1.055) ** 2.4
}

// hex liest eine Hexfarbe als drei Kanäle — oder gar nicht.
function hex(farbe: string): [number, number, number] | undefined {
  const roh = farbe.trim().toLowerCase()
  if (!/^#([0-9a-f]{3}|[0-9a-f]{6})$/.test(roh)) {
    return undefined
  }

  const lang =
    roh.length === 4
      ? roh
          .slice(1)
          .split('')
          .map((zeichen) => zeichen + zeichen)
          .join('')
      : roh.slice(1)

  return [0, 2, 4].map((stelle) =>
    parseInt(lang.slice(stelle, stelle + 2), 16),
  ) as [number, number, number]
}

// leuchtdichte gibt die relative Leuchtdichte einer Hexfarbe — `undefined`,
// wenn es keine ist.
function leuchtdichte(farbe: string): number | undefined {
  const kanaele = hex(farbe)
  if (kanaele === undefined) {
    return undefined
  }

  const [rot, gruen, blau] = kanaele.map(kanal) as [number, number, number]

  return 0.2126 * rot + 0.7152 * gruen + 0.0722 * blau
}

// kontrast gibt das Verhältnis zweier Hexfarben — `undefined`, wenn eine
// davon keine ist.
export function kontrast(vorn: string, hinten: string): number | undefined {
  const eins = leuchtdichte(vorn)
  const zwei = leuchtdichte(hinten)
  if (eins === undefined || zwei === undefined) {
    return undefined
  }

  return (Math.max(eins, zwei) + 0.05) / (Math.min(eins, zwei) + 0.05)
}

// geschleiert gibt die Farbe, die ein Grund mit diesem Anteil über einem
// schwarzen oder weißen Bildpunkt ergibt (`bildpunkt`: 0 oder 255 je Kanal) — `color-mix(in srgb, …)` über dem Bild, also in den
// kodierten Kanälen gemischt.
function geschleiert(grund: string, bildpunkt: number, anteil: number): string {
  const oben = hex(grund)
  if (oben === undefined) {
    return ''
  }

  return `#${oben
    .map((kanal) =>
      Math.round(anteil * kanal + (1 - anteil) * bildpunkt)
        .toString(16)
        .padStart(2, '0'),
    )
    .join('')}`
}

// lesbarkeitsbefunde gibt jedes Paar unter seiner Schwelle.
//
// **`lies` gibt die wirksame Farbe eines Merkmals** — im Fenster aus
// `getComputedStyle`, also nach Farbwelt, Erscheinungsbild und eigenen
// Merkmalen. **Ein Paar, dessen Farbe es nicht kennt, lässt es aus**: raten
// hieße, eine Zahl zu behaupten, die niemand gemessen hat.
//
// **Und drei Messungen jenseits der Merkmale** (Zweitblick über Phase 6.3,
// S4): die Farben der Meldungen im gewählten Erscheinungsbild gegen Grund,
// Fläche und betonte Fläche; und liegt ein Fensterbild unter dem Grund, jedes
// Paar auf dem Grund **und jede Meldungsfarbe** (N4) gegen den Schleier über
// einem schwarzen und einem weißen Bildpunkt — dem dunkelsten und dem
// hellsten, die das Bild haben kann.
export function lesbarkeitsbefunde(
  lies: (merkmal: string) => string,
  lage: Messlage = { dunkel: false },
): Lesbarkeitsbefund[] {
  const befunde: Lesbarkeitsbefund[] = []
  const miss = (
    vorn: string,
    vornfarbe: string,
    hinten: string,
    hintenfarbe: string,
    schwelle: number,
  ) => {
    const verhaeltnis = kontrast(vornfarbe, hintenfarbe)
    if (verhaeltnis !== undefined && verhaeltnis < schwelle) {
      befunde.push({ vorn, hinten, verhaeltnis, schwelle })
    }
  }

  for (const [vorn, hinten, schwelle] of lesbarePaare) {
    miss(vorn, lies(vorn), hinten, lies(hinten), schwelle)
  }

  for (const [vorn, farbe] of statusfarben[lage.dunkel ? 'dunkel' : 'hell']) {
    for (const hinten of ['--mse-grund', '--mse-flaeche', '--mse-betont']) {
      miss(vorn, farbe, hinten, lies(hinten), 4.5)
    }
  }

  if (lage.schleier !== undefined) {
    const grund = lies('--mse-grund')

    for (const [ueber, bildpunkt] of [
      ['fensterbild:dunkel', 0],
      ['fensterbild:hell', 255],
    ] as const) {
      const hinten = geschleiert(grund, bildpunkt, lage.schleier)

      for (const [vorn, auf, schwelle] of lesbarePaare) {
        if (auf === '--mse-grund') {
          miss(vorn, lies(vorn), ueber, hinten, schwelle)
        }
      }

      // **Und die Meldungen stehen auch auf dem Grund** — bis zum Nachlauf
      // über Phase 6.3 maß die Sperre sie dort nur ohne Schleier (N4).
      for (const [vorn, farbe] of statusfarben[
        lage.dunkel ? 'dunkel' : 'hell'
      ]) {
        miss(vorn, farbe, ueber, hinten, 4.5)
      }
    }
  }

  return befunde
}
