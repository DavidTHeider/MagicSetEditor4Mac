// **Die Knöpfe der Rückfrage „Diese Sicherung zurückspielen?", am nativen
// Fenster** — #2412.
//
// # Was hier geschieht
//
// 1. **Hell**: *Einstellungen › Daten und Sicherung*, eine Sicherung gewählt,
//    die Rückfrage steht — „Zurückspielen" und „Abbrechen" sind gleich hoch
//    und haben denselben Innenabstand, gemessen am gerenderten Element.
// 2. **Abbrechen mit der Tastatur**: Fokus auf „Abbrechen", Eingabetaste —
//    die Rückfrage geht, nichts wurde zurückgespielt.
// 3. **Dunkel**: dieselbe Rückfrage.
//
// # Vorher
//
// **Das Bild beim Abbruch von `budgets09`**, Lauf
// `2026-10-15T01-16-05Z-35133` (PERF01, #2351): der Text von
// „Zurückspielen" stieß an den Rand, und der Knopf war niedriger als
// „Abbrechen".
//
// # Was es nicht belegt
//
// **VoiceOver** — die Beschriftungen sind gelesen, gesprochen hat sie hier
// niemand.
import { join } from 'node:path'

const knopf = (text) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.replace(/\\s+/g, ' ').trim() === ${JSON.stringify(text)})`

async function zurRueckfrage(p, sicherung) {
  await p.taste(',', 'befehl')
  const gruppe = "[...document.querySelectorAll('button, a, [role=tab]')].find((e) => e.checkVisibility() && e.innerText.replace(/\\s+/g, ' ').trim().startsWith('Daten'))"
  await p.warte('die Gruppe »Daten« steht zur Wahl', () => p.seite(`return !!(${gruppe})`), 20000)
  await p.klickeAuf(gruppe, 'die Gruppe »Daten«')
  await p.dialogantwort(sicherung)
  await p.warte('»Sicherung zurückspielen …« steht da', () => p.seite(`return !!(${knopf('Sicherung zurückspielen …')})`), 20000)
  await p.seite(`(${knopf('Sicherung zurückspielen …')}).scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(knopf('Sicherung zurückspielen …'), '»Sicherung zurückspielen …«')
  await p.wartetext('Diese Sicherung zurückspielen?', 20000)
  await p.seite(`(${knopf('Zurückspielen')}).scrollIntoView({ block: 'center' }); return true`)
}

// masse sind Höhe, Innenabstand und Schriftgröße beider Knöpfe.
function masse(p) {
  return p.seite(`
    const mass = (b) => { const s = getComputedStyle(b); return { hoehe: b.getBoundingClientRect().height, links: s.paddingLeft, oben: s.paddingTop, schrift: s.fontSize } }
    return { haupt: mass(${knopf('Zurückspielen')}), neben: mass(${knopf('Abbrechen')}) }`)
}

function pruefeMasse(p, gemessen, wann) {
  p.zeige(`${wann}: ${JSON.stringify(gemessen)}`)
  p.pruefe(`${wann}: beide Knöpfe gleich hoch`, Math.abs(gemessen.haupt.hoehe - gemessen.neben.hoehe) < 0.5, JSON.stringify(gemessen))
  p.pruefe(`${wann}: derselbe Innenabstand, 12 px seitlich`, gemessen.haupt.links === '12px' && gemessen.neben.links === '12px' && gemessen.haupt.oben === gemessen.neben.oben, JSON.stringify(gemessen))
  p.pruefe(`${wann}: dieselbe Schriftgröße`, gemessen.haupt.schrift === gemessen.neben.schrift, JSON.stringify(gemessen))
}

export default async function rueckfrage09(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  const sicherung = join(ort.ort, 'probe.msesicherung')
  p.mse(ort, 'bestand-sichern', sicherung)

  p.notiz('1 — hell: die Rückfrage, beide Knöpfe gemessen')
  await p.starte({ ort })
  await zurRueckfrage(p, sicherung)
  pruefeMasse(p, await masse(p), 'hell')
  await p.bild('rueckfrage-hell')

  p.notiz('2 — Abbrechen mit der Tastatur')
  await p.seite(`(${knopf('Abbrechen')}).focus(); return true`)
  await p.taste('Enter')
  await p.warte('die Rückfrage ist fort', async () => !(await p.seite("return document.body.innerText.includes('Diese Sicherung zurückspielen?')")), 10000)
  p.pruefe('⌘Q beendet', await p.beende())

  p.notiz('3 — dunkel')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'dunkel')
  await p.starte({ ort })
  await zurRueckfrage(p, sicherung)
  p.pruefe('Mantine steht dunkel', (await p.seite('return document.documentElement.dataset.mantineColorScheme')) === 'dark')
  pruefeMasse(p, await masse(p), 'dunkel')
  await p.bild('rueckfrage-dunkel')
  p.pruefe('⌘Q beendet', await p.beende())
}
