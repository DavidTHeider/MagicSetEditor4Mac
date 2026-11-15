// **Ein Kernweg je Rolle, nur mit der Tastatur, und was ein Vorleser dabei
// bekäme** — H01 (#2255),
// am nativen Fenster.
//
// Je Rolle — **Verwalter**, **Mitarbeiter**, **Leser** (`domain.Rollen`) —
// derselbe Weg wie in `docs/HANDPRUEFUNGEN.md`: zu den Sets, das Set öffnen,
// eine Karte wählen, ein Feld ändern, zurück. Dazu einmal die **Zeitzone**
// (Einstellungen › Allgemein): tippen, Pfeil, Escape. Erreicht wird jede
// Stelle mit **Tab und Pfeilen** — was der Tabulator nicht erreicht, ist
// unerreichbar —, und an jeder steht in der Tafel, was dort steht: Rolle,
// Name, Zustand, Beschreibung, und wohin der Fokus danach geht. **Wo der
// Name feststeht, wird er verglichen** — die Seite, die Zeile, das Feld,
// die Zeitzone —, nicht nur auf »nicht leer« geprüft. 28 Stellen: acht je
// Rolle, vier an der Zeitzone.
//
// **Das Urteil ist eine Auswertung des DOM**, aus dem WebKit seinen Baum
// baut — mit einer Näherung der Namensberechnung, nicht der Baum selbst
// (Befund 5 des Nachlaufs über Phase I).
//
// **Was er nicht belegt** — und warum H01 trotzdem ohne Menschen schließt
// (`out-of-scope`, Weisung vom 2026-10-11): die **gesprochene** Ausgabe.
// Gelesen wird der DOM, aus dem WebKit den Zugänglichkeitsbaum baut, mit
// einer Näherung der Namensberechnung (`aria-labelledby`, `aria-label`,
// Beschriftung, Text). VoiceOvers eigene Satzbildung, seine Reihenfolge und
// seine Kürzel sieht er nicht; dafür bräuchte es VoiceOver an — eine
// Bedienungshilfe, die nur ein Mensch umlegt (§17) — und ein Ohr.
const schlafe = (ms) => new Promise((los) => setTimeout(los, ms))
const setname = 'Werkstatt-Testkarten'

// beschreibe legt in der Seite eine Näherung dessen ab, was ein Vorleser an
// einem Element findet — nach jedem Neuladen neu.
const beschreibe = `window.h01 = (e) => {
  if (!e || e === document.body || e === document.documentElement) return { rolle: 'body', name: '', zustand: '', beschreibung: '' }
  // **Ohne Verborgenes**, wie die Namensberechnung (accname): ein Zeichen mit
  // aria-hidden gehört nicht zum Namen (erster Lauf: »Zeitzone ▾«).
  const glatt = (x) => {
    if (!x) return ''
    const kopie = x.cloneNode(true)
    kopie.querySelectorAll?.('[aria-hidden=true], [hidden]').forEach((h) => h.remove())
    return (kopie.textContent ?? '').replace(/\\s+/g, ' ').trim()
  }
  const ids = (a) => (e.getAttribute(a) ?? '').split(/\\s+/).filter(Boolean).map((i) => document.getElementById(i)).filter(Boolean)
  const implizit = { BUTTON: 'button', A: 'link', TEXTAREA: 'textbox', SELECT: 'combobox', H1: 'heading', H2: 'heading', MAIN: 'main' }[e.tagName]
    ?? (e.tagName === 'INPUT' ? ({ checkbox: 'checkbox', radio: 'radio' }[e.type] ?? 'textbox') : e.tagName.toLowerCase())
  const rolle = e.getAttribute('role') || implizit
  const textrolle = ['button', 'link', 'row', 'option', 'tab', 'menuitem', 'heading', 'cell', 'gridcell', 'treeitem'].includes(rolle)
  const name = ids('aria-labelledby').map(glatt).join(' ') || e.getAttribute('aria-label') || glatt(e.labels?.[0]) || e.getAttribute('title') || (textrolle ? glatt(e) : '') || e.getAttribute('placeholder') || ''
  const zustand = ['selected', 'expanded', 'checked', 'invalid', 'disabled', 'readonly', 'current', 'pressed', 'rowindex', 'rowcount']
    .filter((a) => e.hasAttribute('aria-' + a)).map((a) => a + '=' + e.getAttribute('aria-' + a))
  if (e.disabled) zustand.push('disabled')
  if (e.readOnly) zustand.push('readonly')
  const aktiv = e.getAttribute('aria-activedescendant')
  const beschreibung = [...ids('aria-describedby').map(glatt), aktiv ? 'aktiv: ' + glatt(document.getElementById(aktiv)) : ''].filter(Boolean).join(' / ')
  return { rolle, name: name.slice(0, 90), zustand: zustand.join(', '), beschreibung: beschreibung.slice(0, 160) }
}; return true`
const fokus = 'return window.h01(document.activeElement)'

// neuLaden lädt die Webansicht neu und wartet, bis die Leiste wieder steht.
async function neuLaden(p) {
  await p.seite('setTimeout(() => location.reload(), 30); return true')
  await schlafe(1500)
  await p.warte('die Leiste steht wieder', () => p.seite("return document.readyState === 'complete' && document.querySelectorAll('nav button').length > 3"), 30000)
  await p.seite(beschreibe)
}

// tabBis drückt Tab, bis das fokussierte Element die Bedingung erfüllt —
// höchstens `hoechstens` Mal. Gibt die Zahl der Tabs oder -1; dann steht die
// **Tab-Folge** im Beleg, damit »unerreichbar« seinen Weg zeigt.
async function tabBis(p, bedingung, hoechstens = 80) {
  const folge = []
  for (let i = 0; i <= hoechstens; i++) {
    const [treffer, b] = await p.seite(`const e = document.activeElement; return [!!(e && (${bedingung})), window.h01(e)]`)
    if (treffer) return i
    if (folge.at(-1) !== `${b.rolle} »${b.name}«`) folge.push(`${b.rolle} »${b.name}«`)
    await p.taste('Tab')
    await schlafe(60)
  }
  p.zeige(`nicht erreicht — Tab-Folge: ${folge.slice(0, 60).join(' → ').slice(0, 1800)}`)
  return -1
}

// bedienbar ist, was ein Vorleser als Ziel ansagt — kein Behälter, dessen
// Text zufällig den gesuchten enthält (erster Lauf: `main` traf den Setnamen).
const bedienbar = "e.matches('button, a[href], input, textarea, select, [role=button], [role=link], [role=row], [role=option], [role=gridcell], [role=treeitem], [role=combobox], [role=tab]')"

// stelle trägt eine Pflichtstelle in die Tafel ein — mit ihrem Urteil.
// **Stumm** heißt ohne Namen, **falsch** eine andere Rolle oder ein anderer
// Zustand als erwartet, **unerreichbar** nicht mit der Tastatur erreicht.
function stelle(tafel, wer, was, b, { erwarteteRolle, erwarteterName, erwarteterZustand, unerreichbar = false, danach = '' } = {}) {
  const urteil = unerreichbar ? 'unerreichbar'
    : b.rolle === 'body' || !b.name ? 'stumm'
      : (erwarteteRolle && b.rolle !== erwarteteRolle) || (erwarteterName !== undefined && b.name !== erwarteterName)
        || (erwarteterZustand && !erwarteterZustand.test(`${b.zustand} ${b.beschreibung}`)) ? 'falsch' : 'ok'
  tafel.push({ wer, was, ...b, danach, urteil })
}

const seite = "return document.querySelector('main')?.dataset.seite ?? ''"

// kernweg fährt den Weg einer Rolle.
async function kernweg(p, ort, tafel, { rolle, nutzer, schreibt }) {
  await p.starte({ ort })
  await p.seite(beschreibe)
  for (const [bindung, wert] of [['NutzerWechseln', nutzer], ['ProfilWechseln', '1']]) {
    p.zeige(`${rolle} — ${bindung}: ${JSON.stringify(await p.seite(`try { const a = await window.go.app.Bindungen.${bindung}('${wert}', ''); return a?.fehler ?? 'gewechselt' } catch (e) { return 'abgewiesen: ' + String(e.message ?? e).slice(0, 160) }`))}`)
    await neuLaden(p)
  }
  const hinweis = await p.seite("return [...document.querySelectorAll('[role=status]')].map((e) => e.innerText.trim()).find((t) => /nur lesen|als Leser/i.test(t)) ?? ''")
  p.pruefe(`${rolle}: der Hinweis auf »nur lesen« steht ${schreibt ? 'nicht' : ''} da`, schreibt ? hinweis === '' : hinweis !== '', hinweis.slice(0, 120))

  // 1. Zu den Sets — ⌘2, und wohin der Fokus dann geht.
  await p.taste('2', 'befehl')
  await p.warte('die Sets stehen', async () => (await p.seite(seite)) === 'sets', 20000)
  await schlafe(400)
  stelle(tafel, rolle, 'nach ⌘2: Fokus', await p.seite(fokus), { erwarteteRolle: 'main', erwarteterName: 'Sets' })

  // 2. Das Set — mit Tab erreicht, mit Enter geöffnet.
  const zumSet = await tabBis(p, `${bedienbar} && window.h01(e).name.includes(${JSON.stringify(setname)})`)
  const amSet = await p.seite(fokus)
  await p.taste('Enter')
  const offen = await p.warte('das Set ist offen', async () => (await p.seite(seite)) === 'karten', 15000).catch(() => false)
  await schlafe(600)
  stelle(tafel, rolle, `das Set (${String(zumSet)} × Tab)`, amSet, { unerreichbar: zumSet < 0 || !offen, danach: JSON.stringify((await p.seite(fokus)).name) })

  // 3. Eine Karte — die Zeile mit der wandernden Tabstelle, dann ↓ und Enter.
  const zurZeile = await tabBis(p, "e.getAttribute('role') === 'row'")
  const ersteZeile = await p.seite(fokus)
  stelle(tafel, rolle, `die erste Kartenzeile (${String(zurZeile)} × Tab)`, ersteZeile, { erwarteteRolle: 'row', erwarteterZustand: /rowindex=\d+/, unerreichbar: zurZeile < 0 })
  const stelleVorher = Number(/rowindex=(\d+)/.exec(ersteZeile.zustand)?.[1] ?? 0)
  await p.taste('ArrowDown')
  await schlafe(300)
  await p.taste('Enter')
  await schlafe(600)
  // **Die nächste Zeile, und gewählt** — der Index steigt um eins.
  stelle(tafel, rolle, 'die nächste Kartenzeile nach ↓ und Enter', await p.seite(fokus), { erwarteteRolle: 'row', erwarteterZustand: new RegExp(`selected=true, rowindex=${String(stelleVorher + 1)}\\b`) })

  // 4. Das Feld »name« — erreicht, geändert, verlassen.
  const zumFeld = await tabBis(p, "e.getAttribute('aria-label') === 'name'", 120)
  const feld = await p.seite(fokus)
  // **Der Leser hört schon beim Ankommen, dass und warum er nur liest.**
  stelle(tafel, rolle, `das Feld »name« (${String(zumFeld)} × Tab)`, feld, { erwarteteRolle: 'textbox', erwarteterName: 'name', erwarteterZustand: schreibt ? /^(?!.*readonly)/s : /readonly.*Nur lesen/s, unerreichbar: zumFeld < 0 })
  if (zumFeld >= 0) {
    // **Ans Ende, nicht über den Inhalt**: Tab markiert ihn, und Tippen
    // ersetzte ihn (zweiter Lauf).
    await p.taste('ArrowRight')
    await p.tippe(' H')
    await p.taste('Tab')
    await schlafe(1500)
    // **Zurück mit der Tastatur** (Befund 5 des Nachlaufs): ⇧Tab, nicht ein
    // Fokus aus dem Skript.
    await p.taste('Tab', 'umschalt')
    await schlafe(300)
    const danach = await p.seite(fokus)
    const meldung = await p.seite("return [...document.querySelectorAll('[role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim()).join(' / ').slice(0, 200)")
    if (schreibt) {
      stelle(tafel, rolle, 'das Feld nach dem Schreiben, mit ⇧Tab', danach, { erwarteteRolle: 'textbox', erwarteterName: 'name', erwarteterZustand: /^(?!.*invalid=true)/ })
    } else {
      // **Der Leser hört, dass es gesperrt ist, und warum** — am Feld oder
      // als Meldung, und der Bestand bleibt, wie er war.
      stelle(tafel, rolle, 'das Feld nach dem Schreibversuch, mit ⇧Tab', { ...danach, beschreibung: `${danach.beschreibung}${meldung ? ` / Meldung: ${meldung}` : ''}`.slice(0, 200) },
        { erwarteteRolle: 'textbox', erwarteterName: 'name', erwarteterZustand: /readonly.*Nur lesen/s })
    }
  }

  // 5. Zurück — Escape, dann ⌘[ — und wohin der Fokus geht.
  await p.taste('Escape')
  await schlafe(400)
  stelle(tafel, rolle, 'nach Escape: Fokus', await p.seite(fokus))
  await p.taste('[', 'befehl')
  const zurueck = await p.warte('⌘[ führt zurück', async () => (await p.seite(seite)) !== 'karten', 10000).then(() => true, () => false)
  await schlafe(400)
  stelle(tafel, rolle, `nach ⌘[ (Seite ${await p.seite(seite)})`, await p.seite(fokus), { erwarteteRolle: 'main', erwarteterName: 'Sets', unerreichbar: !zurueck })
  p.pruefe(`${rolle}: ⌘Q beendet`, await p.beende())
}

// zeitzone fährt die Auswahl der Zeitzone mit der Tastatur.
async function zeitzone(p, ort, tafel) {
  await p.starte({ ort })
  await p.seite(beschreibe)
  await p.taste(',', 'befehl')
  await p.warte('die Einstellungen stehen', async () => (await p.seite(seite)) === 'einstellungen', 20000)
  await schlafe(500)
  const hin = await tabBis(p, "window.h01(e).name === 'Zeitzone'", 150)
  const feld = await p.seite(fokus)
  stelle(tafel, 'Verwalter', `Zeitzone (${String(hin)} × Tab)`, feld, { erwarteteRolle: 'combobox', erwarteterName: 'Zeitzone', unerreichbar: hin < 0 })
  if (hin < 0) return
  await p.tippe('berl')
  await schlafe(800)
  stelle(tafel, 'Verwalter', 'Zeitzone nach »berl«', await p.seite(fokus), { erwarteteRolle: 'combobox', erwarteterName: 'Zeitzone', erwarteterZustand: /expanded=true/ })
  await p.taste('ArrowDown')
  await schlafe(400)
  stelle(tafel, 'Verwalter', 'Zeitzone nach ↓', await p.seite(fokus), { erwarteteRolle: 'combobox', erwarteterName: 'Zeitzone', erwarteterZustand: /aktiv: .*Berlin/ })
  await p.taste('Escape')
  await schlafe(400)
  stelle(tafel, 'Verwalter', 'Zeitzone nach Escape', await p.seite(fokus), { erwarteteRolle: 'combobox', erwarteterName: 'Zeitzone', erwarteterZustand: /expanded=false/ })
  p.pruefe('Zeitzone: ⌘Q beendet', await p.beende())
}

export default async function h01rollen(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  // **Drei Menschen, eine Freigabe je Rolle** am Profil 1 des ersten.
  p.mse(ort, 'nutzer-neu', 'Lea')
  p.mse(ort, 'nutzer-neu', 'Mo')
  p.mse(ort, 'freigabe-setzen', '1', '2', 'leser')
  p.mse(ort, 'freigabe-setzen', '1', '3', 'mitarbeiter')
  p.zeige(`Freigaben: ${p.mse(ort, 'freigaben', '1').replace(/\s+/g, ' ').trim()}`)

  const tafel = []
  for (const lauf of [{ rolle: 'Verwalter', nutzer: '1', schreibt: true }, { rolle: 'Mitarbeiter', nutzer: '3', schreibt: true }, { rolle: 'Leser', nutzer: '2', schreibt: false }]) {
    p.notiz(lauf.rolle)
    const vorher = p.mse(ort, 'karten', '1')
    await kernweg(p, ort, tafel, lauf)
    const nachher = p.mse(ort, 'karten', '1')
    // **Die Wirkung am Bestand, nicht am Bildschirm**: wer schreibt, ändert
    // ihn; der Leser nicht.
    p.pruefe(lauf.schreibt ? `${lauf.rolle}: die Änderung steht im Bestand` : 'Leser: der Bestand bleibt, wie er war', lauf.schreibt ? nachher !== vorher : nachher === vorher, nachher.split('\n').slice(0, 2).join(' | '))
  }
  p.notiz('Zeitzone')
  await zeitzone(p, ort, tafel)

  p.notiz('Die Pflichtstellen')
  p.zeilen.push('| Rolle | Stelle | Rolle im Baum | Name | Zustand | Beschreibung | Fokus danach | Urteil |', '| --- | --- | --- | --- | --- | --- | --- | --- |',
    ...tafel.map((t) => `| ${t.wer} | ${t.was} | ${t.rolle} | ${t.name || '—'} | ${t.zustand || '—'} | ${(t.beschreibung || '—').replaceAll('|', '/')} | ${t.danach || '—'} | ${t.urteil} |`))
  const schlecht = tafel.filter((t) => t.urteil !== 'ok')
  p.pruefe(`null stumme, falsche oder unerreichbare Pflichtstellen (von ${String(tafel.length)})`, schlecht.length === 0, schlecht.map((t) => `${t.wer}: ${t.was} — ${t.urteil}`).join('; '))
}
