// **Der Fan ohne IT: erste Karte, erste Liste, Ausdruck** — G01
// (#2248),
// am nativen Fenster und **in einem Zug**.
//
// B03 hat die Teilaufgaben aus getrennten Drehbüchern zugeordnet und selbst
// gesagt, was das nicht belegt: einen durchgehenden Weg. Dieser geht ihn —
// ein frischer Bestand, die mitgelieferte Saat, die Einführung beim ersten
// Start — und zählt je Abschnitt die Handlungen und die Maschinenzeit
// (`huellenzaehler.mjs`, dieselbe Zählung wie `ausgangswerte`).
//
// Je Abschnitt steht da, **wo er beginnt, was zu entscheiden ist, welcher
// Fehler einem Menschen hier begegnet und woran das Ziel zu sehen ist** —
// das Ziel wird am Bestand gelesen, nicht am Bildschirm.
//
// **Netz**: die Liste wird bei Scryfall zugeordnet (drei Zeilen, wenige
// Anfragen; §17).
//
// **Was er nicht belegt**: einen Menschen. Die Probe kennt den Weg; Lesen,
// Suchen und Zögern misst sie nicht, und eine Erfolgsquote gibt es nur nach
// einer Studie.
import { existsSync, mkdtempSync, statSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

import { sichtbarerKnopf, zaehler } from '../huellenzaehler.mjs'

const feld = (bezeichnung) => `[...document.querySelectorAll('input, textarea')].find((e) => e.checkVisibility() && (e.getAttribute('aria-label') === ${JSON.stringify(bezeichnung)} || (e.labels?.[0]?.innerText ?? '').trim() === ${JSON.stringify(bezeichnung)}))`
const seite = "document.querySelector('main')?.dataset.seite"
const einfuehrungSteht = "!!document.querySelector('[data-einfuehrung]')?.checkVisibility()"

// abschnitt fährt einen Teil des Wegs und hält fest, was er kostet.
async function abschnitt(p, z, tafel, { name, beginn, entscheidungen, fehler, ziel }, weg) {
  p.notiz(`${name} — Beginn: ${beginn}`)
  z.beginne()
  const erreicht = await weg()
  const { handlungen, dauer } = z.ende(name)
  p.pruefe(`${name}: Ziel erreicht — ${ziel}`, erreicht === true, typeof erreicht === 'string' ? erreicht : '')
  tafel.push({ name, beginn, entscheidungen, fehler, ziel, handlungen, dauer })
}

export default async function g01einsteiger(p) {
  const ort = p.ort({ zuhause: true })
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  const ordner = mkdtempSync(join(tmpdir(), 'g01einsteiger-'))
  const bogen = join(ordner, 'meine-karten.pdf')
  const tafel = []

  await p.starte({ ort, einfuehrung: true })
  const z = zaehler(p)

  await abschnitt(p, z, tafel, {
    name: 'Erstkarte',
    beginn: 'der erste Start, die Einführung steht',
    entscheidungen: 'den Stil übernehmen oder wählen; wie Set und Karte heißen',
    fehler: 'keiner provoziert — ein leerer Setname wird am Feld abgewiesen (uxr04)',
    ziel: 'eine Karte mit Name und Text im Bestand',
  }, async () => {
    const einfuehrung = await p.warte('die Einführung oder die Startseite steht', () => p.seite(`return ${einfuehrungSteht} ? 'einfuehrung' : (${seite} === 'start' ? 'start' : false)`), 20000)
    p.zeige(`beim ersten Start: ${einfuehrung}`)
    if (einfuehrung === 'einfuehrung') await z.klick('Zu den Sets')
    else await z.taste('2', 'befehl')
    await p.warte('die Sets stehen', () => p.seite(`return ${seite} === 'sets'`), 20000)
    await z.klick('Neues Set')
    await p.warte('der Fokus steht im Namen', () => p.seite("return (document.activeElement?.labels?.[0]?.innerText ?? '').includes('Name des Sets')"), 20000)
    await z.tippe('Meine ersten Karten')
    await z.klick('Set anlegen')
    await p.warte('die Kartenseite steht', () => p.seite(`return ${seite} === 'karten'`), 20000)
    await p.bild('g01e-set-angelegt')
    await z.klick('Karte anlegen')
    await p.warte('das Namensfeld steht leer', () => p.seite(`return (${feld('name')})?.value === ''`), 20000)
    await z.klickAuf(feld('name'), 'das Feld »name«')
    await p.seite(`(${feld('name')}).focus(); return true`)
    await z.tippe('Feuerfuchs')
    await z.klickAuf(feld('text'), 'das Feld »text«')
    await p.seite(`(${feld('text')}).focus(); return true`)
    await z.tippe('Wenn der Feuerfuchs ins Spiel kommt, ziehe eine Karte.')
    await z.taste('Tab')
    await p.warte('die Karte steht im Bestand', () => p.mse(ort, 'karten', '1').includes('Feuerfuchs'), 20000)
    const beschriftungen = await p.seite("return [...document.querySelectorAll('main label, main [aria-label]')].filter((e) => e.checkVisibility()).map((e) => e.getAttribute('aria-label') ?? e.innerText.trim()).filter((t) => ['name', 'text', 'Name', 'Text'].includes(t))")
    p.zeige(`Feldbeschriftungen, wie die Saat sie nennt: ${JSON.stringify([...new Set(beschriftungen)])}`)
    await p.bild('g01e-erste-karte')
    // **Der Verlauf sagt, was geschah** — bis zum zweiten Lauf stand dort
    // »Karte angelegt oder gelöscht«.
    const verlauf = await p.warte('der Verlauf nennt die Karte', () => p.seite("const t = [...document.querySelectorAll('main table tr')].map((r) => r.innerText.replace(/\\s+/g, ' ')).find((z) => z.includes('Feuerfuchs') && z.includes('Karte')); return t ?? false"), 10000)
    p.pruefe('der Verlauf sagt »Karte angelegt«', verlauf.includes('Karte angelegt') && !verlauf.includes('oder gelöscht'), verlauf)

    return p.mse(ort, 'karten', '1').includes('Feuerfuchs')
  })

  await abschnitt(p, z, tafel, {
    name: 'Erste Liste',
    beginn: 'die leere Sammlung (⌘3)',
    entscheidungen: 'welcher Druck bei einem Namen ohne Set',
    fehler: 'eine Zeile, die es nicht gibt (»Gibtesnichtxyz«)',
    ziel: 'die Exemplare der Liste in der Sammlung',
  }, async () => {
    await z.taste('3', 'befehl')
    await p.warte('die Sammlung steht', () => p.seite(`return ${seite} === 'sammlung'`), 20000)
    await p.bild('g01e-leere-sammlung')
    await z.klick('Eine Liste einfügen')
    await p.warte('das Feld der Liste steht', () => p.seite(`return !!(${feld('Kartenliste')})`), 20000)
    await z.klickAuf(feld('Kartenliste'), 'das Feld »Kartenliste«')
    await p.seite(`(${feld('Kartenliste')}).focus(); return true`)
    // **Eine Liste wird eingefügt**, nicht getippt: eine Handlung. Die Probe
    // tippt sie Zeile für Zeile, weil sie die Ablage nicht anfasst.
    await z.tippe('4 Lightning Bolt (M11) 149')
    for (const zeile of ['1 Counterspell', '1 Gibtesnichtxyz']) {
      await p.taste('Enter')
      await p.tippeInsFeld(zeile)
    }
    await z.klick('Zuordnen')
    const gedeutet = await p.warte('die Zeilen sind zugeordnet', () => p.seite(`const t = document.body.innerText; return /brauchen eine Entscheidung|sind eindeutig/.test(t) ? t : false`), 60000)
    p.zeige(`nach dem Zuordnen: ${gedeutet.split('\n').filter((z2) => /Entscheidung|eindeutig|Druck|heißt|Gibtesnichtxyz/.test(z2)).join(' | ').slice(0, 600)}`)
    await p.bild('g01e-zugeordnet')
    const wahl = await p.seite("const s = [...document.querySelectorAll('select')].find((e) => e.checkVisibility() && (e.labels?.[0]?.innerText ?? '').includes('Welcher Druck')); if (!s) return 'keine Wahl'; const o = [...s.options].find((x) => x.value !== ''); Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(s, o.value); s.dispatchEvent(new Event('change', { bubbles: true })); return o.textContent")
    await z.wahl('return true')
    p.zeige(`gewählter Druck: ${wahl}`)
    await z.klick('Aufnehmen')
    const bericht = await p.warte('der Bericht steht', () => p.seite("const t = document.body.innerText; const m = /\\d+ neu, \\d+ erhöht, \\d+ Stück aufgenommen\\./.exec(t); return m ? m[0] : false"), 60000)
    p.zeige(`Bericht: ${bericht}`)
    await p.bild('g01e-aufgenommen')
    const csv = p.mse(ort, 'sammlung-ausgeben', 'csv').trim().split('\n')

    return csv.length >= 3 ? true : `Sammlung: ${csv.join(' / ')}`
  })

  await abschnitt(p, z, tafel, {
    name: 'Ausdruck',
    beginn: 'das offene Set',
    entscheidungen: 'Papier und Anschnitt — die Vorgaben bleiben',
    fehler: 'keiner provoziert — ein Ort ohne Schreibrecht nennt Ursache und Schritt (fix-f02)',
    ziel: 'ein Druckbogen als PDF',
  }, async () => {
    await z.menue('Ablage', 'Ausgeben', 'Druckbogen …')
    await p.warte('der Druckbogen steht', () => p.seite(`return !!${sichtbarerKnopf('Druckbogen legen …')}`), 20000)
    await p.bild('g01e-druckbogen')
    await z.dialog(bogen)
    await z.klick('Druckbogen legen …')
    await p.warte('der Bogen liegt da', () => Promise.resolve(existsSync(bogen) && statSync(bogen).size > 0), 120000)

    return existsSync(bogen) && statSync(bogen).size > 0
  })

  p.notiz('Die Tafel')
  for (const zeile of tafel) {
    p.zeige(`| ${zeile.name} | ${zeile.beginn} | ${zeile.entscheidungen} | ${zeile.fehler} | ${zeile.ziel} | ${String(zeile.handlungen)} | ${(zeile.dauer / 1000).toFixed(1).replace('.', ',')} s |`)
  }
  p.pruefe('⌘Q beendet', await p.beende())
}
