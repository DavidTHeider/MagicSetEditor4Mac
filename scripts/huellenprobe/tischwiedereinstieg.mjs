// **Der Tisch nach einer Navigation, wieder und wieder** — UI-N01
// (#2089).
//
// Der Minimalweg aus dem Audit für 0.7.0: Tisch → Spielen (⌘4) → »Partie
// anfangen« → »Weiterspielen als Ada« → Tisch. Bis zum 2026-10-06 blieb der
// **zweite** Einstieg stehen, und die Webansicht antwortete nicht mehr.
//
// Je Erscheinungsbild zehn Zyklen bei 1280 × 800. Jeder Zyklus belegt
// positiv: die Leinwand des Tischs steht sichtbar da, ihr WebGL-Kontext ist
// nicht verloren, sie ist **neu** und nicht die vom letzten Besuch, die Karten
// der Partie stehen wieder auf ihr, »Näher heran« rückt sie, und die Zahl der
// Leinwände im Dokument wächst nicht. Dazu die Zeit je Zyklus.
//
// **Was es nicht belegt**: die hundert Wechsel derselben Sitzung samt
// Prozesslast aus der Abnahme der Iteration (ABN-01) — und ob ein Mensch
// den Tisch danach bedienen kann, sieht ein Mensch.

const zyklen = Number(process.env.MSE_TISCH_ZYKLEN ?? 10)

// tischbefund liest, was am Tisch steht — Leinwand, Kontext, Karten. **Und ob
// die Leinwand neu ist**: sie wird markiert, und ein zweiter Blick auf dieselbe
// sähe die Marke. Ein Zurückschalten auf den alten Tisch belegte nichts.
const tischbefund = (p) => p.seite(`
  const tisch = document.querySelector('[role=application]')
  const leinwand = tisch?.querySelector('canvas')
  const gl = leinwand && (leinwand.getContext('webgl2') ?? leinwand.getContext('webgl'))
  const neu = !!leinwand && !leinwand.dataset.gesehen
  if (leinwand) leinwand.dataset.gesehen = 'ja'
  return {
    sichtbar: !!leinwand && leinwand.checkVisibility() && leinwand.width > 0,
    neu,
    verloren: gl ? gl.isContextLost() : null,
    karten: tisch ? tisch.querySelectorAll('li').length : 0,
    leinwaende: document.querySelectorAll('canvas').length,
  }`)

const knopfDa = (p, text) => p.seite(`return [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(text)})`)

// zumTisch geht über ⌘4 und »Partie anfangen« an den Tisch — **beim ersten
// Mal** über die Liste der Partien und »Weiterspielen als Ada«, danach steht
// die offene Partie gleich da.
async function zumTisch(p) {
  await p.taste('4', 'befehl')
  await p.wartetext('Partie anfangen', 20000)
  await p.klickeText('Partie anfangen')
  const weg = await p.warte('die Partienliste oder der Tisch steht da', async () => {
    if (await knopfDa(p, 'Weiterspielen')) return 'liste'

    return (await p.seite("return !!document.querySelector('[role=application] canvas')?.checkVisibility()")) ? 'tisch' : false
  }, 20000)
  if (weg === 'liste') {
    await p.klickeText('Weiterspielen')
    await p.klickeText('Weiterspielen als Ada')
  }

  return p.warte('der Tisch zeichnet', async () => {
    const befund = await tischbefund(p)

    return befund.sichtbar && befund.verloren === false && befund.karten > 0 ? befund : false
  }, 20000)
}

// bedient drückt »Näher heran« und wartet, bis die Karten über der Leinwand
// woanders stehen — die Kamera des neuen Malers folgt.
async function bedient(p) {
  const stelle = "return document.querySelector('[role=application] li')?.getAttribute('style') ?? ''"
  const vorher = await p.seite(stelle)
  await p.klickeText('Näher heran')

  return Boolean(await p.warte('»Näher heran« rückt die Karten', async () => (await p.seite(stelle)) !== vorher, 5000))
}

export default async function tischwiedereinstieg(p) {
  for (const bild of ['hell', 'dunkel']) {
    const ort = p.ort()
    p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
    p.mse(ort, 'deck', 'neu', 'Probe-Deck', 'probe')
    for (const karte of ['1', '2', '3', '4', '5']) p.mse(ort, 'deck', 'legen', '1', `--karte=${karte}`, '8')
    p.mse(ort, 'partie-anlegen', 'probe', 'ada:Ada:1', 'bo:Bo:1')
    p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)

    p.notiz(`Erscheinungsbild ${bild}, ${String(zyklen)} Zyklen bei 1280 × 800`)
    await p.starte({ ort })
    await p.groesse(1280, 800)
    const erster = await zumTisch(p)
    p.pruefe(`erster Einstieg (${bild}): Leinwand, Kontext, ${String(erster.karten)} Karten`, true)
    await p.bild(`tisch-erst-${bild}`)

    const zeiten = []
    let leinwaende = erster.leinwaende
    for (let zyklus = 1; zyklus <= zyklen; zyklus += 1) {
      const anfang = Date.now()
      const befund = await zumTisch(p)
      zeiten.push(Date.now() - anfang)
      p.pruefe(`Zyklus ${String(zyklus)} (${bild}): eine neue Leinwand mit denselben ${String(erster.karten)} Karten`, befund.neu && befund.karten === erster.karten, JSON.stringify(befund))
      p.pruefe(`Zyklus ${String(zyklus)} (${bild}): »Näher heran« wirkt`, await bedient(p))
      if (befund.leinwaende > leinwaende) p.pruefe(`Zyklus ${String(zyklus)} (${bild}): keine weitere Leinwand`, false, `${String(leinwaende)} → ${String(befund.leinwaende)}`)
      leinwaende = befund.leinwaende
    }
    await p.bild(`tisch-zuletzt-${bild}`)

    // **Ein verlorener Kontext, und der Weg zurück** (AK3): der Tisch sagt,
    // dass er nicht mehr zeichnet, und der nächste Einstieg zeichnet wieder —
    // auf einer neuen Leinwand.
    await p.seite("const c = document.querySelector('[role=application] canvas'); (c.getContext('webgl2') ?? c.getContext('webgl')).getExtension('WEBGL_lose_context').loseContext(); return true")
    p.pruefe(`ein verlorener Kontext wird gemeldet (${bild})`, Boolean(await p.warte('der Tisch meldet den Verlust', () => p.seite("return [...document.querySelectorAll('[role=alert]')].some((e) => e.checkVisibility() && e.innerText.trim() !== '')"), 10000)))
    const zurueck = await zumTisch(p)
    p.pruefe(`danach zeichnet der Tisch wieder (${bild})`, zurueck.karten === erster.karten && !(await p.seite("return [...document.querySelectorAll('[role=alert]')].some((e) => e.checkVisibility() && e.innerText.trim() !== '')")))
    p.zeige(`Zeit je Zyklus (${bild}): ${zeiten.map((z) => String(z)).join(', ')} ms; Leinwände im Dokument: ${String(erster.leinwaende)} → ${String(leinwaende)}`)
    p.pruefe(`⌘Q beendet die Anwendung (${bild})`, await p.beende())
  }
}
