// **Merkmale statt fester Grautöne, am gerenderten Element gemessen** — N13
// aus dem Nachlauf über Phase 6.3 (#1664), Phase 6.10 (#1671).
//
// Der Hinweis „nur lesen" stand als festes Paar `text-amber-800` auf
// `bg-slate-100`; jetzt als `text-warnung` auf `bg-betont`, der gewählte
// Eintrag im Vorlagenbaum als `bg-betont`. Gemessen wird, was WebKit für
// diese Klassen rechnet — im Fenster der Anwendung, mit ihrem Stilblatt —,
// hell und dunkel, und das Verhältnis nach WCAG 2.2.
//
// **Was es nicht tut**: Farbwelten und eigene Farben durchgehen. Die
// misst `kontrast.ts` für jedes Merkmal selbst.
const paare = [
  ['text-warnung', 'bg-betont'],
  ['text-text', 'bg-betont'],
  ['text-leise', 'bg-betont'],
]

const messen = `
  const lum = (farbe) => { const [r, g, b] = farbe.match(/[\\d.]+/g).slice(0, 3).map((w) => { const c = Number(w) / 255; return c <= 0.04045 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4 }); return 0.2126 * r + 0.7152 * g + 0.0722 * b }
  const main = document.querySelector('main')
  return ${JSON.stringify(paare)}.map(([vorn, grund]) => {
    const p = document.createElement('p'); p.className = vorn + ' ' + grund; p.textContent = 'Probe'; main.prepend(p)
    const stil = getComputedStyle(p); const a = lum(stil.color); const b = lum(stil.backgroundColor); p.remove()
    return { vorn, grund, farbe: stil.color, hinten: stil.backgroundColor, verhaeltnis: Math.round(((Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05)) * 100) / 100 }
  })
`

export default async function n13(p) {
  const ort = p.ort()
  for (const bild of ['hell', 'dunkel']) {
    p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
    await p.starte({ ort })
    await p.warte('die Seite steht', () => p.seite("return !!document.querySelector('main')"), 20000)
    for (const w of await p.seite(messen)) {
      p.zeige(`${bild}: ${w.vorn} auf ${w.grund} — ${w.farbe} auf ${w.hinten}: ${String(w.verhaeltnis)}`)
      p.pruefe(`${w.vorn} auf ${w.grund} hält 4,5 : 1 (${bild})`, w.verhaeltnis >= 4.5, String(w.verhaeltnis))
    }
    p.pruefe(`⌘Q beendet die Anwendung (${bild})`, await p.beende())
  }
}
