// **Ein Leser sieht die Schreibzeilen eines Menüs gesperrt** — #2288
// (#2288),
// am nativen Fenster:
//
//   1. als Leser im Profil das Saatset öffnen und „Diese Karte" aufklappen —
//      Duplizieren, Ausschneiden, Einfügen und Löschen tragen `aria-disabled`
//      und den Grund als Beschreibung, Kopieren nicht,
//   2. ein Klick auf „Karte löschen" öffnet keine Rückfrage,
//   3. das Kartenfeld »Name« ist schreibgeschützt, und Tippen ändert nichts
//      (#2290).
//
// **Der Weg vorher** steht im Issue: dieselben Zeilen standen bedienbar da,
// und erst die Bindung wies ab. Der Leserwechsel ist derselbe wie in
// `zustaende.mjs` (Zelle „nur lesen").
const zeile = (wort) => `[...document.querySelectorAll('[role=menuitem]')].find((e) => e.checkVisibility() && e.innerText.includes(${JSON.stringify(wort)}))`

const lage = (wort) => `const e = ${zeile(wort)}; if (!e) return null; const ids = (e.getAttribute('aria-describedby') ?? '').split(' ').filter(Boolean); return { gesperrt: e.getAttribute('aria-disabled') === 'true', grund: ids.map((id) => document.getElementById(id)?.textContent.trim() ?? '').join(' ') }`

async function neuLaden(p) {
  await p.seite('setTimeout(() => location.reload(), 30); return true')
  await new Promise((weiter) => { setTimeout(weiter, 1500) })
  await p.warte('die Leiste steht wieder', () => p.seite("return document.readyState === 'complete' && document.querySelectorAll('nav button').length > 3"), 30000)
}

export default async function g10menue(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'nutzer-neu', 'Zweite')
  p.mse(ort, 'freigabe-setzen', '1', '2', 'leser')
  await p.starte({ ort })

  p.notiz('1 — als Leser: „Diese Karte"')
  p.zeige(`Nutzerwechsel: ${JSON.stringify(await p.seite("try { const a = await window.go.app.Bindungen.NutzerWechseln('2', ''); return a.fehler ?? 'gewechselt' } catch (e) { return 'abgewiesen: ' + String(e.message ?? e).slice(0, 200) }"))}`)
  await neuLaden(p)
  p.zeige(`Profilwechsel: ${JSON.stringify(await p.seite("try { const a = await window.go.app.Bindungen.ProfilWechseln('1', ''); return a.fehler ?? 'gewechselt' } catch (e) { return 'abgewiesen: ' + String(e.message ?? e).slice(0, 200) }"))}`)
  await neuLaden(p)
  await p.taste('2', 'befehl')
  await p.wartetext('Werkstatt-Testkarten', 20000)
  await p.klickeText('Werkstatt-Testkarten')
  await p.warte('»Diese Karte« steht da', () => p.seite("return [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.includes('Diese Karte'))"), 20000)
  await p.klickeText('Diese Karte')
  await p.warte('das Menü ist offen', () => p.seite(`return !!(${zeile('Kopieren')})`), 10000)

  for (const wort of ['Karte duplizieren', 'Ausschneiden', 'Einfügen', 'Karte löschen']) {
    const l = await p.seite(lage(wort))
    p.pruefe(`»${wort}« ist gesperrt und erklärt`, l?.gesperrt === true && l.grund !== '', JSON.stringify(l))
  }
  const kopieren = await p.seite(lage('Kopieren'))
  p.pruefe('»Kopieren« bleibt offen', kopieren?.gesperrt === false, JSON.stringify(kopieren))
  await p.bild('g10menue-leser')

  p.notiz('2 — „Karte löschen" tut nichts')
  await p.klickeAuf(zeile('Karte löschen'), '»Karte löschen«')
  await new Promise((weiter) => { setTimeout(weiter, 800) })
  p.pruefe('keine Rückfrage geht auf', await p.seite("return ![...document.querySelectorAll('[role=dialog], [role=alertdialog]')].some((d) => d.checkVisibility())"))
  await p.bild('g10menue-nach-klick')

  p.notiz('3 — das Kartenfeld »Name« ist schreibgeschützt (#2290)')
  await p.taste('Escape')
  const name = "document.querySelector('main input[aria-label=\"name\"]')"
  await p.warte('das Feld »Name« steht da', () => p.seite(`return !!(${name})`), 10000)
  p.pruefe('»Name« ist schreibgeschützt', await p.seite(`return ${name}.readOnly === true`))
  const vorher = await p.seite(`return ${name}.value`)
  await p.klickeAuf(name, 'das Feld »Name«')
  await p.tippe('xyz')
  await new Promise((weiter) => { setTimeout(weiter, 500) })
  p.pruefe('Tippen ändert nichts', (await p.seite(`return ${name}.value`)) === vorher, vorher)
  await p.bild('g10menue-feld')

  p.pruefe('⌘Q beendet', await p.beende())
}
