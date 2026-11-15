// **Die Oberfläche der KI am nativen Fenster** — UXR-21
// (#1929).
//
// Misst mit **derselben** Messung wie `uirevue`, bei 1280 × 800 und
// 1024 × 700, hell und dunkel:
//
// - *Einstellungen › KI* — Anbieter, Laufzeit, Kontingent, Modelle,
// - *Karten* mit der ersten Karte — und darin **was nach dem Formular
//   steht**: der KI-Teil des Inspectors, gezählt für sich, samt der Tiefe,
//   die der Inspector selbst rollt (die Messung sieht nur `main` und seine
//   Vorfahren).
//
// Dazu **ein Weg**, den UXR-21.AK-04 verlangt: an einer KI-Meldung „KI
// einrichten" — kommt man damit in die Gruppe KI der Einstellungen?
//
// **Was es nicht tut**: urteilen, ob die Anordnung dem Artboard gleicht.
// Das sieht ein Mensch an den Bildern.
import { messung, oeffnen, zurSeite } from './uirevue.mjs'

const groessen = [
  { name: 'normal', breite: 1280, hoehe: 800 },
  { name: 'eng', breite: 1024, hoehe: 700 },
]

// mitText findet einen Knopf oder eine Lasche an der ersten Zeile seines Texts
// — die Gruppenliste trägt unter jedem Namen einen Untertitel.
const mitText = (text) => `[...document.querySelectorAll('button, [role=tab]')].find((b) => b.checkVisibility() && b.innerText.trim().split('\\n')[0] === ${JSON.stringify(text)})`

// nachDemFormular zählt, was im Inspector **hinter** dem Kartenformular steht
// — der KI-Teil — und wie tief der Inspector rollt.
const nachDemFormular = `
  const sichtbar = (e) => e.checkVisibility({ checkOpacity: true, checkVisibilityCSS: true })
  const formular = document.querySelector('main form')
  if (!formular) return null
  const danach = (e) => !formular.contains(e) && (formular.compareDocumentPosition(e) & Node.DOCUMENT_POSITION_FOLLOWING) !== 0
  const bedien = [...document.querySelectorAll('main button, main summary, main a[href], main [role=tab]')].filter(sichtbar).filter(danach)
  const felder = [...document.querySelectorAll('main input:not([type=hidden]), main select, main textarea, main [role=switch]')].filter(sichtbar).filter(danach)
  const imBild = (e) => { const r = e.getBoundingClientRect(); return r.bottom > 0 && r.top < innerHeight }
  let roller = formular.parentElement
  while (roller && !(roller.scrollHeight > roller.clientHeight + 1 && /(auto|scroll)/.test(getComputedStyle(roller).overflowY))) roller = roller.parentElement
  const ersteZeile = (e) => (e.getAttribute('aria-label') || e.innerText || '').trim().split('\\n')[0].slice(0, 40)
  return {
    bedien: bedien.length,
    bedienImBild: bedien.filter(imBild).length,
    felder: felder.length,
    namen: bedien.map(ersteZeile),
    inspectorRolltiefe: roller ? Math.round((roller.scrollHeight / roller.clientHeight) * 10) / 10 : 1,
  }
`

const imBild = `
  const main = document.querySelector('main')
  return [...main.querySelectorAll('button, summary, a[href], input, select, textarea, [role=switch]')]
    .filter((e) => { if (!e.checkVisibility({ checkOpacity: true }) || e.type === 'hidden') return false; const r = e.getBoundingClientRect(); return r.bottom > 0 && r.top < innerHeight })
    .map((e) => (e.getAttribute('aria-label') || e.labels?.[0]?.innerText || e.innerText || e.getAttribute('name') || e.tagName).trim().split('\\n')[0].slice(0, 40))
`

const zeile = (groesse, bild, ziel, z) =>
  `| ${groesse} | ${bild} | ${ziel} | ${String(z.bedien)} (${String(z.bedienImBild)}) | ${String(z.haupt)} (${String(z.hauptJeRegion)}) | ${String(z.felder)} (${String(z.felderImBild)}) | ${String(z.ueberschriften)} | ${String(z.rolltiefe)} | ${z.waagerecht ? 'ja' : 'nein'} |`

async function einstellungenKI(p, groesse, bild, zeilen) {
  await zurSeite(p, 'Einstellungen')
  await p.warte('die Gruppe KI steht', () => p.seite(`return !!${mitText('KI')}`), 20000)
  await p.klickeAuf(mitText('KI'), 'die Gruppe KI')
  await new Promise((los) => setTimeout(los, 1500))
  const wert = await p.seite(messung)
  zeilen.push(zeile(groesse, bild, 'Einstellungen › KI', wert))
  p.zeige(`im Bild unter Einstellungen › KI (${groesse}, ${bild}): ${(await p.seite(imBild)).join(' · ')}`)
  p.pruefe(`R1 unter Einstellungen › KI (${groesse}, ${bild})`, wert.hauptJeRegion <= 1, JSON.stringify(wert))
  p.pruefe(`R5 unter Einstellungen › KI (${groesse}, ${bild})`, !wert.waagerecht)
  await p.bild(`ki-einstellungen-${groesse}-${bild}`)
}

async function kartenKI(p, groesse, bild, zeilen) {
  await zurSeite(p, 'Karten')
  await p.warte('das Formular steht', () => p.seite("return !!document.querySelector('main form')"), 20000)
  await new Promise((los) => setTimeout(los, 1000))
  const wert = await p.seite(messung)
  zeilen.push(zeile(groesse, bild, 'Karten', wert))
  const ki = await p.seite(nachDemFormular)
  p.zeige(`nach dem Formular (${groesse}, ${bild}): ${String(ki.bedien)} Bedienelemente (${String(ki.bedienImBild)} im Bild), ${String(ki.felder)} Felder, der Inspector rollt ${String(ki.inspectorRolltiefe)} — ${ki.namen.join(' · ')}`)
  // **Das Bild zeigt den KI-Teil**, nicht den Kopf des Inspectors.
  await p.seite("const f = document.querySelector('main form'); let e = f.nextElementSibling ?? f; e.scrollIntoView({ block: 'start', behavior: 'instant' }); return true")
  await p.bild(`ki-karten-${groesse}-${bild}`)
  // **Seit UXR-21 ein Kasten** — gemessen wird auch, was er aufgeschlagen
  // trägt; vorher gab es ihn nicht, und die Zeile bleibt leer.
  if (await p.seite(`return !!${kikasten}`)) {
    await p.klickeAuf(kikasten, '»KI zu dieser Karte«')
    await new Promise((los) => setTimeout(los, 500))
    const offen = await p.seite(nachDemFormular)
    p.zeige(`nach dem Formular, Kasten offen (${groesse}, ${bild}): ${String(offen.bedien)} Bedienelemente (${String(offen.bedienImBild)} im Bild), ${String(offen.felder)} Felder, der Inspector rollt ${String(offen.inspectorRolltiefe)} — ${offen.namen.join(' · ')}`)
    await p.bild(`ki-karten-offen-${groesse}-${bild}`)
    return { ...ki, offen }
  }
  return ki
}

// kikasten ist die Zusammenfassung des KI-Kastens am Kartenformular.
const kikasten = "[...document.querySelectorAll('main summary')].find((s) => s.checkVisibility() && s.innerText.trim() === 'KI zu dieser Karte')"

// kiEinrichten fährt den Weg von einer KI-Meldung in die Einstellungen —
// über die Einschätzung. **Ohne Laufzeit und ohne Gerätemodell steht dort
// kein Modell bereit**, und das ist der Fall, den AI-01.AK-04 meint.
async function kiEinrichten(p) {
  p.notiz('Der Weg: an einer KI-Meldung „KI einrichten"')
  await zurSeite(p, 'Karten')
  const zusammenfassung = (text) => `[...document.querySelectorAll('main summary')].find((s) => s.checkVisibility() && s.innerText.trim() === ${JSON.stringify(text)})`
  // **Seit UXR-21 hinter dem KI-Kasten** — erst er, dann die Einschätzung.
  await p.warte('das Formular steht', () => p.seite("return !!document.querySelector('main form')"), 20000)
  if (await p.seite(`return !!${kikasten}`)) await p.klickeAuf(kikasten, '»KI zu dieser Karte«')
  await p.warte('die Einschätzung steht', () => p.seite(`return !!${zusammenfassung('Beurteilen lassen')}`), 20000)
  await p.seite(`${zusammenfassung('Beurteilen lassen')}.scrollIntoView({ block: 'center', behavior: 'instant' }); return true`)
  await p.klickeAuf(zusammenfassung('Beurteilen lassen'), '»Beurteilen lassen«')
  const kasten = `${zusammenfassung('Beurteilen lassen')}.parentElement`
  const knopf = `[...${kasten}.querySelectorAll('button')].find((b) => b.innerText.trim() === 'KI einrichten')`
  // **Gewartet wird auf eine Auskunft** — Knopf, Satz oder Meldung —, nicht
  // nur auf den Knopf: sonst stünde vorher nach 30 s „nicht da" statt dessen,
  // was das Fenster sagt.
  const lage = await p.warte('die Einschätzung sagt, was fehlt', () => p.seite(`const k = ${kasten}; if (${knopf}) return 'Knopf'; const m = [...k.querySelectorAll('[role=alert], p')].map((e) => e.innerText.trim()).find((t) => /kein Modell|antwortet|Laufzeit/.test(t)); return m ? 'Satz: ' + m.slice(0, 120) : ''`), 30000).catch(() => 'nichts')
  p.zeige(`an der Einschätzung: ${lage}`)
  await p.seite(`${kasten}.scrollIntoView({ block: 'center', behavior: 'instant' }); return true`)
  await p.bild('ki-einrichten-meldung')
  p.pruefe('am Satz, dass kein Modell bereitsteht, steht »KI einrichten«', lage === 'Knopf', lage)
  if (lage !== 'Knopf') return
  const anfang = Date.now()
  await p.klickeAuf(knopf, '»KI einrichten«')
  const ziel = await p.warte('die Einstellungen stehen in der Gruppe KI', () => p.seite("const m = document.querySelector('main'); if (m?.dataset.seite !== 'einstellungen') return ''; return [...document.querySelectorAll('[aria-selected=true]')].map((e) => e.innerText.trim().split('\\n')[0]).join(' / ')"), 20000)
  p.zeige(`danach gewählt: ${ziel} — nach ${String(Date.now() - anfang)} ms`)
  p.pruefe('die Gruppe KI ist aufgeschlagen', ziel.split(' / ').includes('KI'), ziel)
  await p.bild('ki-einrichten-ziel')
}

export default async function uxr21(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')

  const zeilen = []
  const ki = []
  for (const bild of ['hell', 'dunkel']) {
    p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
    for (const groesse of groessen) {
      await oeffnen(p, ort, groesse)
      await einstellungenKI(p, groesse.name, bild, zeilen)
      ki.push({ groesse: groesse.name, bild, ...(await kartenKI(p, groesse.name, bild, zeilen)) })
      p.pruefe(`⌘Q beendet die Anwendung (${groesse.name}, ${bild})`, await p.beende())
    }
  }

  // **Der Weg zur Einrichtung braucht einen Rechner ohne Modell** — auf
  // diesem stünde das Gerätemodell im Katalog, und die Einschätzung hätte
  // eines. Es wird für diesen Lauf ausgeschaltet und danach wieder auf die
  // Vorgabe gestellt.
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'hell')
  p.mse(ort, 'einstellungen', 'setzen', 'ki.geraetemodell', 'aus')
  await oeffnen(p, ort, groessen[0])
  await kiEinrichten(p)
  p.pruefe('⌘Q beendet die Anwendung (ohne Modell)', await p.beende())
  p.mse(ort, 'einstellungen', 'setzen', 'ki.geraetemodell', 'automatisch')

  // **Und auf Englisch** — UXR-21.AK-03 verlangt DE und EN.
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'en')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'hell')
  await oeffnen(p, ort, groessen[0])
  await zurSeite(p, 'Settings', 'en')
  await p.warte('the AI group stands', () => p.seite(`return !!${mitText('AI')}`), 20000)
  await p.klickeAuf(mitText('AI'), 'the AI group')
  await new Promise((los) => setTimeout(los, 1500))
  p.zeige(`im Bild unter Settings › AI: ${(await p.seite(imBild)).join(' · ')}`)
  await p.bild('ki-einstellungen-englisch')
  p.pruefe('⌘Q beendet die Anwendung (englisch)', await p.beende())

  p.notiz('Die Tabelle')
  p.zeige('| Größe | Bild | Ziel | Bedienelemente (im Bild) | Hauptknöpfe (je Region höchstens) | Felder (im Bild) | Überschriften | Rolltiefe | waagerecht |')
  p.zeige('| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |')
  for (const z of zeilen) p.zeige(z)
  p.notiz('Nach dem Formular — der KI-Teil des Inspectors')
  p.zeige('| Größe | Bild | Bedienelemente (im Bild) | Felder | Inspector rollt | Kasten offen: Bedienelemente (im Bild) | Felder | Inspector rollt |')
  p.zeige('| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |')
  for (const z of ki) p.zeige(`| ${z.groesse} | ${z.bild} | ${String(z.bedien)} (${String(z.bedienImBild)}) | ${String(z.felder)} | ${String(z.inspectorRolltiefe)} | ${z.offen ? `${String(z.offen.bedien)} (${String(z.offen.bedienImBild)})` : '—'} | ${z.offen ? String(z.offen.felder) : '—'} | ${z.offen ? String(z.offen.inspectorRolltiefe) : '—'} |`)
}
