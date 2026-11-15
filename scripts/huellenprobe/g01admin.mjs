// **Der Admin: Rolle, Sicherung, Update** — G01
// (#2248),
// am nativen Fenster und **in einem Zug**.
//
// Drei Wege, je mit dem Fehler, der einem Menschen dort begegnet:
//
//   1. **Rolle** — einen Nutzer anlegen und ihm das eigene Profil zum Lesen
//      freigeben; vorher der Knopf ohne Namen,
//   2. **Sicherung** — mit Kennwort sichern, weiterarbeiten (ein Set kommt
//      über `mse` dazu) und zurückspielen; vorher das falsche Kennwort,
//   3. **Update** — eine neuere Fassung eines Pakets über den Plan einbinden;
//      vorher »Nichts ändern«.
//
// Das Update der **Anwendung** fährt er nicht: es braucht eine Freigabe auf
// GitHub und ein ausgeliefertes Bündel (I03, #2261). Und das Update eines
// **bezogenen** Baums belegt `g04bezuege` gegen GitHub selbst.
import { existsSync, mkdirSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

import { zaehler } from '../huellenzaehler.mjs'

const meldungen = "return [...document.querySelectorAll('[role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim()).filter(Boolean).join(' | ')"
const feldMit = (beschriftung) => `[...document.querySelectorAll('input')].find((e) => e.checkVisibility() && (e.labels?.[0]?.innerText ?? '').trim().startsWith(${JSON.stringify(beschriftung)}))`
const setzeWahl = (beschriftung, wert) => `const s = [...document.querySelectorAll('select')].find((w) => w.checkVisibility() && (w.labels?.[0]?.textContent ?? '').trim().startsWith(${JSON.stringify(beschriftung)})); if (!s) return false; Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(s, ${JSON.stringify(wert)}); s.dispatchEvent(new Event('change', { bubbles: true })); return true`
const kennwort = 'korrekt pferd batterie heftklammer'

function fassung(ort, version) {
  const paket = join(ort, `v${version}`, 'probe-fremd.mse-style')
  mkdirSync(paket, { recursive: true })
  writeFileSync(join(paket, 'style'), `mse version: 2.0.0\ngame: probe\nshort name: Fremd\nversion: ${version}\n`)

  return paket
}

async function abschnitt(p, z, tafel, { name, beginn, entscheidungen, fehler, ziel }, weg) {
  p.notiz(`${name} — Beginn: ${beginn}`)
  z.beginne()
  const erreicht = await weg()
  const { handlungen, dauer } = z.ende(name)
  p.pruefe(`${name}: Ziel erreicht — ${ziel}`, erreicht === true, typeof erreicht === 'string' ? erreicht : '')
  tafel.push({ name, beginn, entscheidungen, fehler, ziel, handlungen, dauer })
}

async function tippeIn(p, z, ausdruck, text, beschreibung) {
  await p.warte(`${beschreibung} steht`, () => p.seite(`return !!(${ausdruck})`), 20000)
  await z.klickAuf(ausdruck, beschreibung)
  await p.seite(`(${ausdruck}).focus(); return true`)
  await z.tippe(text)
}

export default async function g01admin(p) {
  const ort = p.ort({ zuhause: true })
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  const alt = fassung(ort.ort, '2026-01-01')
  const neu = fassung(ort.ort, '2026-02-02')
  p.mse(ort, 'paket-installieren', alt)
  p.mse(ort, 'set-neu', 'Planset', 'probe', 'probe-fremd')
  const sicherung = join(ort.ort, 'admin.msesicherung')
  const tafel = []

  await p.starte({ ort })
  const z = zaehler(p)

  await abschnitt(p, z, tafel, {
    name: 'Rolle',
    beginn: 'Nutzer und Profile (⌘\')',
    entscheidungen: 'wer; welche Rolle — Leser, Mitarbeiter oder Verwalter',
    fehler: 'ohne Namen — der Knopf ist gesperrt, solange das Feld leer ist',
    ziel: 'Mia liest das eigene Profil mit',
  }, async () => {
    await z.taste("'", 'befehl')
    await p.wartetext('Neuen Nutzer anlegen', 20000)
    await z.klick('Neuen Nutzer anlegen')
    await p.warte('das Namensfeld steht', () => p.seite(`return !!(${feldMit('Anzeigename des neuen Nutzers')})`), 20000)
    // **Ohne Namen ist der Knopf gesperrt** — der Fehler wird verhindert,
    // nicht gemeldet; der erste Lauf klickte ins Leere (2026-10-10).
    const gesperrt = await p.seite("return [...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Nutzer anlegen')?.disabled === true")
    p.zeige(`ohne Namen: ${gesperrt ? 'der Knopf ist gesperrt' : 'der Knopf ist frei'}`)
    p.pruefe('ohne Namen lässt sich kein Nutzer anlegen', gesperrt)
    await p.bild('g01a-ohne-namen')
    await tippeIn(p, z, feldMit('Anzeigename des neuen Nutzers'), 'Mia', 'das Namensfeld')
    await z.klick('Nutzer anlegen')
    await p.warte('Mia steht da', () => p.mse(ort, 'nutzer').includes('Mia'), 20000)
    await z.text('Profile')
    await p.wartetext('Wer an diesem Profil arbeiten darf', 20000)
    // **Eingeklappt seit UXR-20** (#1928): der Kasten will geöffnet werden.
    await z.text('Wer an diesem Profil arbeiten darf')
    await p.warte('die Wahl »Wem« steht', () => p.seite("return [...document.querySelectorAll('select')].some((w) => w.checkVisibility() && (w.labels?.[0]?.textContent ?? '').trim().startsWith('Wem'))"), 10000)
    const mia = await p.seite("const s = [...document.querySelectorAll('select')].find((w) => w.checkVisibility() && (w.labels?.[0]?.textContent ?? '').trim().startsWith('Wem')); return [...(s?.options ?? [])].find((o) => o.textContent.includes('Mia'))?.value ?? ''")
    p.zeige(`Mia hat die Kennung ${mia}`)
    await z.wahl(setzeWahl('Wem', mia))
    await z.klick('Freigeben')
    await p.bild('g01a-freigegeben')
    const freigaben = await p.warte('die Freigabe steht im Bestand', () => { const f = p.mse(ort, 'freigaben', '1'); return f.includes('Mia') ? f : false }, 20000)

    // **Eine Zeile je Freigabe: Kennung, Rolle, Name** (`mse freigaben`).
    return freigaben.split('\n').some((zeile) => zeile.includes('Mia') && /\bleser\b/.test(zeile)) ? true : freigaben
  })

  await abschnitt(p, z, tafel, {
    name: 'Sicherung',
    beginn: 'Einstellungen › Daten und Sicherung (⌘,)',
    entscheidungen: 'mit oder ohne Kennwort; wohin; ob zurückgespielt wird',
    fehler: 'beim Zurückspielen das falsche Kennwort',
    ziel: 'der Stand der Sicherung ist zurück — das Set, das danach kam, ist fort',
  }, async () => {
    await z.taste(',', 'befehl')
    await p.wartetext('Daten und Sicherung', 20000)
    await z.text('Daten und Sicherung')
    await p.ruhe()
    await tippeIn(p, z, feldMit('Kennwort (freiwillig)'), kennwort, 'das Kennwortfeld')
    await tippeIn(p, z, feldMit('Kennwort wiederholen'), kennwort, 'die Wiederholung')
    await z.dialog(sicherung)
    await z.klick('Bestand sichern …')
    await p.warte('die Sicherung liegt da', () => Promise.resolve(existsSync(sicherung)), 30000)
    // **Weitergearbeitet wird woanders** — ein Set kommt dazu, und das
    // Zurückspielen soll es wieder nehmen. Über `mse`, denn das ist nicht
    // der Weg, den hier jemand geht.
    p.mse(ort, 'set-neu', 'Nach der Sicherung', 'probe', 'probe-fremd')
    await z.dialog(sicherung)
    await z.klick('Sicherung zurückspielen …')
    await tippeIn(p, z, feldMit('Kennwort der Sicherung'), 'falsches kennwort hier', 'das Kennwort der Sicherung')
    await z.klick('Öffnen')
    const falsch = await p.warte('das falsche Kennwort wird abgewiesen', () => p.seite("return document.body.innerText.includes('Das Kennwort passt nicht') ? 'Das Kennwort passt nicht zu dieser Sicherung.' : false"), 30000)
    p.zeige(`falsches Kennwort: ${falsch}`)
    await p.bild('g01a-falsches-kennwort')
    await p.seite(`const f = ${feldMit('Kennwort der Sicherung')}; Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set.call(f, ''); f.dispatchEvent(new Event('input', { bubbles: true })); return true`)
    await tippeIn(p, z, feldMit('Kennwort der Sicherung'), kennwort, 'das Kennwort der Sicherung')
    await z.klick('Öffnen')
    await p.wartetext('Diese Sicherung zurückspielen?', 30000)
    await p.bild('g01a-vorschau')
    await z.klick('Zurückspielen')
    await p.wartetext('Zurückgespielt.', 30000)
    const sets = p.mse(ort, 'sets')

    return !sets.includes('Nach der Sicherung') && sets.includes('Planset') ? true : sets
  })

  await abschnitt(p, z, tafel, {
    name: 'Update',
    beginn: 'Ressourcen › Pakete (⌘6)',
    entscheidungen: 'den Plan lesen: Fassungen, betroffene Sets, Quelle',
    fehler: '»Nichts ändern« beim ersten Plan',
    ziel: 'probe-fremd in der neuen Fassung, die alte zum Zurückholen',
  }, async () => {
    const zeile = "document.querySelector('tr[aria-label=\"probe-fremd\"]')"
    await z.taste('6', 'befehl')
    await p.warte('das Paket steht', () => p.seite(`return !!${zeile}`), 20000)
    await z.dialog(neu)
    await z.klick('Paket einbinden …')
    await p.wartetext('probe-fremd.mse-style ersetzen', 15000)
    await z.klick('Nichts ändern')
    await p.warte('der Plan ist fort', () => p.seite("return !document.body.innerText.includes('probe-fremd.mse-style ersetzen')"), 10000)
    await z.dialog(neu)
    await z.klick('Paket einbinden …')
    await p.wartetext('probe-fremd.mse-style ersetzen', 15000)
    await p.bild('g01a-plan')
    await z.klick('Ersetzen')
    const fassungJetzt = () => JSON.parse(p.mse(ort, 'pakete', '--json')).pakete.find((eines) => eines.name === 'probe-fremd')?.version ?? ''
    await p.warte('die neue Fassung steht', () => fassungJetzt() === '2026-02-02', 15000)

    return fassungJetzt() === '2026-02-02'
  })

  p.notiz('Die Tafel')
  for (const zeile of tafel) {
    p.zeige(`| ${zeile.name} | ${zeile.beginn} | ${zeile.entscheidungen} | ${zeile.fehler} | ${zeile.ziel} | ${String(zeile.handlungen)} | ${(zeile.dauer / 1000).toFixed(1).replace('.', ',')} s |`)
  }
  p.pruefe('⌘Q beendet', await p.beende())
}
