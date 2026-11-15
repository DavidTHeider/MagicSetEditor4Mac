// **Der Tisch am nativen Fenster** — UXR-14
// (#1922).
//
// `uirevue` misst den Tisch nicht — er ist keine Seite, sondern entsteht aus
// einer Partie. Dieses Drehbuch legt eine Partie zu zweit an, setzt sich als
// Ada an den Tisch und misst bei 1280 × 800 und 1024 × 700:
//
// - **die Spalte neben der Leinwand** — Bedienelemente, Hauptknöpfe, Felder,
//   ob sie rollt (R2 nimmt den Tisch selbst aus, die Spalte nicht),
// - **was doppelt steht** — ein Name, der auf der Leinwand (`role=application`)
//   und in der Spalte als Knopf vorkommt,
// - **die ganze Seite** mit derselben Messung wie `uirevue`.
//
// Gemessen wird ohne gewählte Karte und mit einer gewählten Karte aus der
// eigenen Hand. **Was es nicht tut**: urteilen, ob die Anordnung dem Artboard
// *Tisch* gleicht. Das sieht ein Mensch an den Bildern.
import { messung } from './uirevue.mjs'

const groessen = [
  { name: 'normal', breite: 1280, hoehe: 800 },
  { name: 'eng', breite: 1024, hoehe: 700 },
]

const spalte = `
  const sichtbar = (e) => e.checkVisibility()
  const s = document.querySelector('[class*="@4xl:w-80"]')
  if (!s) return null
  const tisch = document.querySelector('[role=application]')
  const name = (b) => (b.getAttribute('aria-label') || b.innerText || '').trim()
  const knoepfe = [...s.querySelectorAll('button, a[href], [role=switch], [role=tab]')].filter(sichtbar)
  const felder = [...s.querySelectorAll('input, select, textarea')].filter((e) => sichtbar(e) && e.type !== 'hidden')
  const aufTisch = new Set(tisch ? [...tisch.querySelectorAll('button')].filter(sichtbar).map(name) : [])
  const rand = s.getBoundingClientRect()
  return {
    bedien: knoepfe.length + felder.length,
    bedienImBild: [...knoepfe, ...felder].filter((e) => { const r = e.getBoundingClientRect(); return r.bottom > 0 && r.top < innerHeight }).length,
    haupt: knoepfe.filter((b) => b.classList.contains('bg-marke')).length,
    felder: felder.length,
    rollt: s.scrollHeight > s.clientHeight + 1,
    hoehe: Math.round(rand.height),
    doppelt: knoepfe.map(name).filter((n) => n !== '' && aufTisch.has(n)),
    namen: knoepfe.map(name).slice(0, 40),
  }
`

const zeile = (groesse, ziel, z) =>
  `| ${groesse} | ${ziel} | ${String(z.bedien)} (${String(z.bedienImBild)}) | ${String(z.haupt)} | ${String(z.felder)} | ${z.rollt ? 'ja' : 'nein'} | ${z.doppelt.length === 0 ? '—' : z.doppelt.join(', ')} |`

export default async function uxr14(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'deck', 'neu', 'Probe-Deck', 'probe')
  for (const karte of ['1', '2', '3', '4', '5']) p.mse(ort, 'deck', 'legen', '1', `--karte=${karte}`, '8')
  p.mse(ort, 'partie-anlegen', 'probe', 'ada:Ada:1', 'bo:Bo:1')

  const zeilen = []
  const seiten = []
  for (const groesse of groessen) {
    p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
    p.mse(ort, 'einstellungen', 'setzen', 'fensterbreite', String(groesse.breite))
    p.mse(ort, 'einstellungen', 'setzen', 'fensterhoehe', String(groesse.hoehe))
    await p.starte({ ort })

    await p.taste('4', 'befehl')
    await p.wartetext('Partie anfangen', 20000)
    await p.klickeText('Partie anfangen')
    await p.wartetext('Weiterspielen')
    await p.klickeText('Weiterspielen')
    await p.klickeText('Weiterspielen als Ada')
    await p.warte('die Leinwand zeichnet', () => p.seite("return !!document.querySelector('canvas') && document.querySelector('canvas').checkVisibility()"), 20000)
    await new Promise((los) => setTimeout(los, 1500))

    const ohne = await p.seite(spalte)
    p.zeige(`Spalte ohne Wahl (${groesse.name}): ${JSON.stringify(ohne)}`)
    if (ohne) zeilen.push(zeile(groesse.name, 'Tisch, keine Karte gewählt', ohne))
    seiten.push({ groesse: groesse.name, ziel: 'Tisch (ganze Seite)', ...(await p.seite(messung)) })
    await p.bild(`tisch-${groesse.name}`)

    // **Erst die Starthand** — vor dem Beginn liegt keine Karte auf der
    // Hand (wie `beginnFuer` in `fix-f04`).
    if (await p.seite("return [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Starthand ziehen')")) {
      await p.klickeText('Starthand ziehen')
      await p.wartetext('Deine Starthand hat')
      await p.klickeText('Hand behalten')
      await p.warte('behalten', () => p.seite("return !document.body.innerText.includes('Deine Starthand hat')"))
      await new Promise((los) => setTimeout(los, 1000))
    }

    // **Eine Karte aus der eigenen Hand** — die Zonen liegen als Knöpfe auf
    // der Leinwand, die eigenen unten (wie in `fix-f04`). **Gewählt mit
    // ⌘-Klick**: seit P4 (#2372) spielt ein Klick sie aus.
    const gewaehlt = await p.seite("const b = [...document.querySelectorAll('[role=application] li button')].filter((x) => x.checkVisibility() && (x.getAttribute('aria-label') || x.textContent).includes('Hand')); if (b.length === 0) return false; return true")
    if (gewaehlt) {
      await p.klickeAuf("(() => { const b = [...document.querySelectorAll('[role=application] li button')].filter((x) => x.checkVisibility() && (x.getAttribute('aria-label') || x.textContent).includes('Hand')); return b.reduce((a, c) => (c.getBoundingClientRect().top > a.getBoundingClientRect().top ? c : a)) })()", 'eine Karte der eigenen Hand', { mit: ['befehl'] })
      await new Promise((los) => setTimeout(los, 1000))
      const mit = await p.seite(spalte)
      p.zeige(`Spalte mit Wahl (${groesse.name}): ${JSON.stringify(mit)}`)
      if (mit) zeilen.push(zeile(groesse.name, 'Tisch, eine Karte der Hand gewählt', mit))
      await p.bild(`tisch-wahl-${groesse.name}`)
    } else {
      p.zeige(`keine Karte auf der eigenen Hand (${groesse.name})`)
    }

    p.pruefe(`⌘Q beendet die Anwendung (${groesse.name})`, await p.beende())
  }

  p.notiz('Die Spalte neben der Leinwand')
  p.zeige('| Größe | Zustand | Bedienelemente (im Bild) | Hauptknöpfe | Felder | rollt | doppelt auf der Leinwand |')
  p.zeige('| --- | --- | ---: | ---: | ---: | --- | --- |')
  for (const z of zeilen) p.zeige(z)

  p.notiz('Die ganze Seite')
  p.zeige('| Größe | Ziel | Bedienelemente (im Bild) | Hauptknöpfe (je Region höchstens) | Felder (im Bild) | Überschriften | Rolltiefe | waagerecht |')
  p.zeige('| --- | --- | ---: | ---: | ---: | ---: | ---: | --- |')
  for (const z of seiten) {
    p.zeige(`| ${z.groesse} | ${z.ziel} | ${String(z.bedien)} (${String(z.bedienImBild)}) | ${String(z.haupt)} (${String(z.hauptJeRegion)}) | ${String(z.felder)} (${String(z.felderImBild)}) | ${String(z.ueberschriften)} | ${String(z.rolltiefe)} | ${z.waagerecht ? 'ja' : 'nein'} |`)
  }
  p.pruefe('R1: höchstens ein Hauptknopf je Region am Tisch', seiten.every((z) => z.hauptJeRegion <= 1), JSON.stringify(seiten.map((z) => z.hauptJeRegion)))
  p.pruefe('R5: der Tisch rollt nicht waagerecht', seiten.every((z) => !z.waagerecht))
}
