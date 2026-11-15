// **Steht das Formular still, wenn eine Karte aufgeht?** — eine Diagnose aus
// E06 (#2278).
//
// Der Zwei-Stunden-Lauf vom 2026-10-11 brach in seiner ersten Runde ab: der
// Klick auf das Textfeld traf »Vorschlagen« unter dem Namen. Die Probe hatte
// das Feld gemessen und still gesehen; danach rückte es noch. Dieses Drehbuch
// öffnet die Karte wie der Dauerlauf und hält vier Sekunden lang alle 50 ms
// fest, wo das Textfeld steht, wie weit sein Bereich gerollt ist und welche
// Teile der Seite sich dabei ändern.
//
// **Was es nicht belegt**: dass ein Mensch es sieht — nur, ob und wann sich
// die Stelle ändert, **sobald das Formular sichtbar ist**. Was es vorher
// unsichtbar tut, zählt nicht.
export const nurEinzeln = true

const feld = "[...document.querySelectorAll('input, textarea, [contenteditable=true]')].find((e) => e.checkVisibility() && e.getAttribute('aria-label') === 'text')"

const mitschrift = `return new Promise((fertig) => {
  const name = (k) => { let e = k.nodeType === 1 ? k : k.parentElement; while (e && e !== document.body) { const a = e.getAttribute('aria-label') || e.getAttribute('role'); if (a) return a.slice(0, 30); e = e.parentElement } return 'body' }
  let jetzt = {}
  const beob = new MutationObserver((liste) => { for (const m of liste) { const n = name(m.target); jetzt[n] = (jetzt[n] ?? 0) + 1 } })
  beob.observe(document.body, { subtree: true, childList: true, attributes: true, characterData: true })
  const proben = []
  const beginn = performance.now()
  // bild gibt die Überschriften und Beschriftungen des Inspektors mit ihrer Höhe
  // — vorher und nachher, damit zu sehen ist, was dazukam.
  const bild = () => { const i = document.querySelector('[aria-label="Felder und Prüfung"]'); if (!i) return 'kein Inspektor'; return [...i.querySelectorAll('h2, h3, legend, label, [role=status], [role=alert], section[aria-label], button')].filter((e) => e.checkVisibility()).map((e) => Math.round(e.getBoundingClientRect().top) + ':' + (e.innerText || e.getAttribute('aria-label') || e.tagName).trim().replace(/\\s+/g, ' ').slice(0, 24)).slice(0, 18).join(' | ') }
  const zuerst = bild()
  const takt = () => {
    const e = (${feld})
    let rolle = null
    for (let k = e?.parentElement; k; k = k.parentElement) { if (k.scrollHeight > k.clientHeight + 4 && /auto|scroll/.test(getComputedStyle(k).overflowY)) { rolle = k; break } }
    const wo = []
    for (let k = e?.parentElement; k && wo.length < 1; k = k.parentElement) { const a = k.getAttribute('aria-label'); if (a) wo.push(a.slice(0, 24)) }
    const alle = [...document.querySelectorAll('[aria-label="text"]')].length
    const name = [...document.querySelectorAll('input')].find((x) => x.getAttribute('aria-label') === 'name')
    wo.push('Name oben ' + (name ? String(Math.round(name.getBoundingClientRect().top)) : '–') + ', Feld hoch ' + (e ? String(Math.round(e.getBoundingClientRect().height)) : '–') + ', Feld links ' + (e ? String(Math.round(e.getBoundingClientRect().left)) : '–'))
    let inspektor = e
    for (let k = 0; inspektor && k < 6; k += 1) inspektor = inspektor.parentElement
    const kinder = []
    const gehe = (k, tiefe) => { for (const c of k.children) { const r = c.getBoundingClientRect(); if (r.height > 0 && tiefe < 3 && r.top < 700) { kinder.push(String(tiefe) + c.tagName.toLowerCase() + '@' + String(Math.round(r.top)) + '+' + String(Math.round(r.height)) + ((c.getAttribute('aria-label') || '').slice(0, 12))) ; gehe(c, tiefe + 1) } } }
    if (inspektor) gehe(inspektor, 0)
    proben.push({ kinder: kinder.slice(0, 40).join(' '), ms: Math.round(performance.now() - beginn), wo: wo.join(' < ') + ' (' + String(alle) + ' im Dokument)', oben: e ? Math.round(e.getBoundingClientRect().top) : null, gerollt: rolle ? Math.round(rolle.scrollTop) : null, fokus: document.activeElement?.getAttribute('aria-label') ?? document.activeElement?.tagName ?? null, aenderungen: Object.entries(jetzt).sort((a, b) => b[1] - a[1]).slice(0, 4).map(([k, v]) => k + ' ' + v).join(', ') })
    jetzt = {}
    if (performance.now() - beginn < 4000) setTimeout(takt, 50)
    else { beob.disconnect(); fertig({ proben, zuerst, zuletzt: bild() }) }
  }
  takt()
})`

export default async function kartenruhe(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  await p.starte({ ort })

  for (const runde of [1, 2]) {
    await p.taste('2', 'befehl')
    await p.wartetext('Werkstatt-Testkarten', 20000)
    await p.klickeText('Werkstatt-Testkarten')
    // **Gemessen wird, was zu sehen ist**: ein Formular mit `aria-busy` steht
    // unsichtbar da, bis die Vorschau seine Felder ordnet (#2301).
    await p.warte('das Textfeld steht sichtbar da', () => p.seite(`const e = (${feld}); return !!e && !e.closest('[aria-busy=true]')`), 20000)
    const { proben, zuerst, zuletzt } = await p.seite(mitschrift)
    p.zeilen.push(`- zuerst: ${zuerst}`, `- zuletzt: ${zuletzt}`)
    const wechsel = proben.filter((x, i) => i === 0 || x.oben !== proben[i - 1].oben || x.gerollt !== proben[i - 1].gerollt || x.fokus !== proben[i - 1].fokus)
    p.notiz(`Runde ${String(runde)}: ${String(wechsel.length)} Stände in ${String(proben.length)} Proben`)
    for (const x of wechsel) p.zeilen.push(`- ${String(x.ms)} ms: Textfeld oben ${String(x.oben)} in ${x.wo}, gerollt ${String(x.gerollt)}, Fokus ${String(x.fokus)}${x.aenderungen === '' ? '' : ` — ${x.aenderungen}`}`, `  - ${x.kinder}`)
    p.pruefe(`Runde ${String(runde)}: das Textfeld steht still, sobald es da ist`, wechsel.length <= 1, `${String(wechsel.length)} Stände`)
    await p.taste('1', 'befehl')
  }

  p.pruefe('⌘Q beendet', await p.beende())
}
