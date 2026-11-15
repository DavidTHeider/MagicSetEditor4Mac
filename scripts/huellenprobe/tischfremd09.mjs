// **Ein offener Tisch liest nach, was eine andere Fläche schreibt, am
// nativen Fenster** — #2388.
//
// Gefunden bei der Ausgangsmessung P0 (#2336): die Probe legte über
// `mse partie-befehl` zwei Kreaturen aufs Spielfeld, das Fenster zeigte nach
// 15 s noch die alte Fassung, und der nächste Klick prallte ab. Hier steht der
// Tisch offen, eine Karte geht **über die Kommandozeile** von der Hand aufs
// Spielfeld, und gemessen wird, wann das Fenster sie zeigt — **ohne Klick**.
// Die Abnahme verlangt zwei Abfragezyklen von `useBestandsstand` (je 2 s).
const zaehleIm = (p, zone, sitz = 'ada') => p.seite(`
  const feld = document.querySelector('[data-zonenfeld="${zone}"][data-sitz="${sitz}"]')?.getBoundingClientRect()
  if (!feld) return -1
  let anzahl = 0
  for (const li of document.querySelectorAll('[role=application] li')) {
    const r = li.getBoundingClientRect()
    const x = r.left + r.width / 2, y = r.top + r.height / 2
    if (x >= feld.left && x <= feld.right && y >= feld.top && y <= feld.bottom) anzahl += 1
  }
  return anzahl`)

export default async function tischfremd09(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'deck', 'neu', 'Tisch-Deck', 'probe')
  for (const karte of ['1', '2', '3', '4', '5']) p.mse(ort, 'deck', 'legen', '1', `--karte=${karte}`, '8')
  p.mse(ort, 'partie-anlegen', 'probe', 'ada:Ada:1', 'bo:Bo:1')
  const partie = JSON.parse(p.mse(ort, 'partien', '--json')).partien[0].id
  for (const sitz of ['ada', 'bo']) {
    p.mse(ort, 'partie-befehl', partie, sitz, `start-${sitz}`, 'spieler.starthand')
    p.mse(ort, 'partie-befehl', partie, sitz, `behalten-${sitz}`, 'spieler.behalten')
  }

  await p.starte({ ort })
  await p.taste('4', 'befehl')
  await p.wartetext('Partie anfangen', 20000)
  await p.klickeText('Partie anfangen')
  const weiter = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? '') === 'Weiterspielen: Ada · Bo')"
  await p.warte('»Weiterspielen« steht', () => p.seite(`return !!${weiter}`), 20000)
  await p.klickeAuf(weiter, '»Weiterspielen: Ada · Bo«')
  await p.klickeText('Weiterspielen als Ada')
  await p.warte('der Tisch trägt seine Zonenmarken', () => p.seite("return !!document.querySelector('[data-zonenfeld=\"tisch\"][data-sitz=\"ada\"]')"), 30000)
  await p.ruhe()

  const vorher = await zaehleIm(p, 'tisch')
  p.zeige(`auf dem Spielfeld vorher: ${String(vorher)}`)
  await p.bild('tisch-vorher')

  p.notiz('Eine Karte über die Kommandozeile aufs Spielfeld')
  const stand = JSON.parse(p.mse(ort, 'partie-zeigen', partie, 'ada', '--json')).partie
  const karte = stand.objekte.find((o) => o.eigner === 'ada' && o.zone === 'hand')
  p.pruefe('Ada hat eine Karte auf der Hand', !!karte)
  const los = Date.now()
  p.mse(ort, 'partie-befehl', partie, 'ada', 'fremd-1', 'objekt.schieben', JSON.stringify({ objekt: karte.kennung, zone: 'tisch' }))

  const gezeigt = await p.warte('das Fenster zeigt sie, ohne Klick', async () => (await zaehleIm(p, 'tisch')) > vorher, 15000).then(() => true, () => false)
  const dauer = Date.now() - los
  p.pruefe('das Fenster zeigt die neue Fassung ohne Klick', gezeigt, `nach ${String(dauer)} ms`)
  p.pruefe('innerhalb von zwei Abfragezyklen (4 s) und dem Lesen', gezeigt && dauer <= 5000, `${String(dauer)} ms`)
  await p.bild('tisch-nachher')

  p.notiz('Der nächste Befehl prallt nicht ab')
  const meldung = await p.seite("return [...document.querySelectorAll('[role=alert], [role=status]')].map((e) => e.innerText).join(' | ')")
  p.pruefe('keine Meldung über eine veraltete Fassung', !/rechnet gegen die Fassung/u.test(meldung ?? ''), meldung)
  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
}
