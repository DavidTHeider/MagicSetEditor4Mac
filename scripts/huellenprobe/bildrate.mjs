// **Die Bildzeit im Wails-Fenster selbst** — PLAY-06.AK-01
// (#1714),
// Rest aus #1871.
//
// Gemessen war WebKit in der Probe (`belege/webkitprobe/`); das Fenster hat
// davor die Brücke und den Go-Prozess. Hier läuft dieselbe Anwendung im
// nativen Fenster, mit einer Partie am gezeichneten Tisch, und gemessen wird
// mit `requestAnimationFrame` im Fenster — **zweimal**: in Ruhe und während
// der Tisch verschoben wird (Pfeiltasten auf der Fläche, wie ein Mensch es
// mit der Tastatur tut).
//
// **Was das nicht ist**: eine Framezeitverteilung unter Last mit vier
// Spielern (PLAY-06.AK-02) — das misst die Probe. Und es ist **dieser** Mac;
// die Zahl gilt für ihn.

const messen = (p, bilder) => p.seite(`
  return new Promise((fertig) => {
    const zeiten = []
    let vorher = 0
    const schlag = (jetzt) => {
      if (vorher !== 0) zeiten.push(jetzt - vorher)
      vorher = jetzt
      if (zeiten.length < ${String(bilder)}) requestAnimationFrame(schlag)
      else {
        const sortiert = [...zeiten].sort((a, b) => a - b)
        const q = (anteil) => sortiert[Math.min(sortiert.length - 1, Math.floor(anteil * sortiert.length))]
        fertig({ n: zeiten.length, mittel: zeiten.reduce((s, z) => s + z, 0) / zeiten.length, median: q(0.5), p95: q(0.95), p99: q(0.99), max: sortiert[sortiert.length - 1] })
      }
    }
    requestAnimationFrame(schlag)
  })`)

const zahl = (wert) => wert.toFixed(1)

export default async function bildrate(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'deck', 'neu', 'Probe-Deck', 'probe')
  for (const karte of ['1', '2', '3', '4', '5']) p.mse(ort, 'deck', 'legen', '1', `--karte=${karte}`, '8')
  p.mse(ort, 'partie-anlegen', 'probe', 'ada:Ada:1', 'bo:Bo:1')

  await p.starte({ ort })
  await p.taste('4', 'befehl')
  await p.wartetext('Partie anfangen', 20000)
  await p.klickeText('Partie anfangen')
  // **Die Sitze stehen im Menü „Weiterspielen" der Zeile** — seit UXR-13
  // (#1921); vorher trug die Zeile je Sitz einen Knopf.
  await p.wartetext('Weiterspielen')
  await p.klickeText('Weiterspielen')
  await p.klickeText('Weiterspielen als Ada')
  await p.warte('der Tisch ist gezeichnet', () => p.seite("return !!document.querySelector('[role=application] canvas')"), 20000)
  await new Promise((los) => setTimeout(los, 2000))
  await p.bild('bildrate-tisch')

  p.notiz('In Ruhe, 300 Bilder')
  const ruhe = await messen(p, 300)
  p.zeige(`Ruhe: Mittel ${zahl(ruhe.mittel)} ms, Median ${zahl(ruhe.median)}, p95 ${zahl(ruhe.p95)}, p99 ${zahl(ruhe.p99)}, Höchstwert ${zahl(ruhe.max)} (n = ${String(ruhe.n)})`)
  p.pruefe('in Ruhe zeichnet das Fenster Bilder', ruhe.n === 300 && ruhe.mittel > 0)

  p.notiz('Während der Tisch mit den Pfeiltasten verschoben wird')
  await p.seite("document.querySelector('[role=application]')?.focus(); return true")
  const laufend = messen(p, 300)
  for (let i = 0; i < 40; i++) {
    await p.taste(i % 2 === 0 ? 'ArrowRight' : 'ArrowLeft')
    await new Promise((los) => setTimeout(los, 50))
  }
  const bewegt = await laufend
  p.zeige(`Bewegt: Mittel ${zahl(bewegt.mittel)} ms, Median ${zahl(bewegt.median)}, p95 ${zahl(bewegt.p95)}, p99 ${zahl(bewegt.p99)}, Höchstwert ${zahl(bewegt.max)} (n = ${String(bewegt.n)})`)
  p.pruefe('auch in Bewegung zeichnet es Bilder', bewegt.n === 300 && bewegt.mittel > 0)

  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
}
