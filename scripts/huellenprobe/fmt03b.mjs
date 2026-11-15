// **Ein langer Lauf am nativen Fenster: Stand, Anhalten, Wiederholen** —
// FMT-03.AK-02 und AK-04
// (#1692):
//
// - ein Druckbogen über vierhundert Karten zeigt an seiner Seite, wie weit er
//   ist,
// - *Abbrechen* hält ihn an: kein halber Bogen liegt da, und die Seite sagt es
//   ohne Fehlermeldung, mit dem Stand bis dahin,
// - derselbe Knopf legt ihn danach ganz,
// - die Hintergrundaufgaben der Startseite nennen beide Läufe.
//
// **Ein eigenes Zuhause**: das große Set entsteht in der Probe.
import { existsSync, mkdirSync, mkdtempSync, statSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(name)})`
const balken = "document.querySelector('[role=progressbar][aria-valuenow]')"
const stand = (anfang) => `const e = [...document.querySelectorAll('[role=status]')].find((s) => s.checkVisibility() && s.innerText.startsWith(${JSON.stringify(anfang)})); return e ? e.innerText : false`

function grossesSet(ordner, zahl) {
  const pfad = join(ordner, 'gross.mse-set')
  mkdirSync(pfad)

  let text = 'mse version: 2.0.0\ngame: probe\nstylesheet: schlicht\nset info:\n\ttitle: Großes Probeset\n'
  for (let nummer = 1; nummer <= zahl; nummer++) text += `card:\n\tname: Karte ${String(nummer)}\n`
  writeFileSync(join(pfad, 'set'), text)

  return pfad
}

export default async function fmt03b(p) {
  const ort = p.ort({ zuhause: true })
  const ordner = mkdtempSync(join(tmpdir(), 'fmt03b-'))
  p.mse(ort, 'oeffnen', grossesSet(ordner, 400))
  const ziel = join(ordner, 'bogen.pdf')

  await p.starte({ ort })
  await p.taste('2', 'befehl')
  await p.wartetext('Großes Probeset', 20000)
  await p.klickeText('Großes Probeset')
  await p.wartetext('Karte anlegen', 20000)
  await p.klick('Ablage', 'Drucken …')
  await p.warte('der Druckbogen steht', () => p.seite(`return !!${knopf('Druckbogen legen …')}`), 20000)

  p.notiz('Ein Bogen über 400 Karten, angehalten')
  await p.dialogantwort(ziel)
  await p.klickeAuf(knopf('Druckbogen legen …'), 'Druckbogen legen …')
  const gezaehlt = await p.warte('der Balken zählt', () => p.seite(`const b = ${balken}; return b && Number(b.getAttribute('aria-valuenow')) > 0 ? b.getAttribute('aria-valuenow') + ' von ' + b.getAttribute('aria-valuemax') : false`), 30000)
  p.pruefe('an der Seite steht, wie weit er ist', true, gezaehlt)
  await p.bild('laeuft')

  await p.klickeAuf(knopf('Abbrechen'), 'Abbrechen')
  const satz = await p.warte('die Seite sagt, dass er angehalten ist', () => p.seite(stand('Abgebrochen.')), 30000)
  p.pruefe('mit dem Stand bis dahin und ohne Fehlermeldung',
    satz.includes('Bis dahin:') && await p.seite("return !document.querySelector('[role=alert]')"), satz)
  p.pruefe('es liegt kein halber Bogen da', !existsSync(ziel))
  await p.bild('abgebrochen')

  p.notiz('Derselbe Knopf legt ihn ganz')
  await p.dialogantwort(ziel)
  await p.warte('der Knopf ist wieder frei', () => p.seite(`return !${knopf('Druckbogen legen …')}.disabled`))
  await p.klickeAuf(knopf('Druckbogen legen …'), 'Druckbogen legen …')
  const fertig = await p.warte('der Bogen ist gelegt', () => p.seite(stand('400 Karten')), 240000)
  p.pruefe('der Bogen liegt da, ganz', existsSync(ziel) && statSync(ziel).size > 0, `${String(statSync(ziel).size)} Bytes`)
  // **Der Satz darüber, einmal gezählt** — hier stand „400 Karten Karten
  // auf 45 Seiten", gefunden im ersten Lauf dieser Probe.
  p.pruefe('der Satz nennt Karten und Seiten je einmal', /^400 Karten auf \d+ Seiten: /.test(fertig), fertig.split('\n')[0])
  await p.bild('gelegt')

  p.notiz('Die Hintergrundaufgaben nennen beide Läufe')
  await p.taste('1', 'befehl')
  const buch = await p.warte('die Übersicht nennt sie', () => p.seite("const b = document.querySelector('section[aria-label=\"Hintergrundaufgaben\"]'); return b && b.innerText.includes('abgebrochen') && b.innerText.includes('fertig') ? b.innerText : false"), 20000)
  p.pruefe('als »Druckbogen legen · bogen.pdf«, einmal abgebrochen, einmal fertig',
    buch.split('Druckbogen legen · bogen.pdf').length === 3, buch.replaceAll('\n', ' · '))
  // **Und der Bogen gilt als Ausgabe** — bis FMT-03 stand darunter „Noch
  // nichts ausgegeben". **Seit UXR-03 eine Reihe statt vier Listen** (#1911):
  // die Zeile des Sets sagt, was zuletzt mit ihm geschah.
  const ausgegeben = await p.seite("const b = document.querySelector('section[aria-label=\"Weiterarbeiten\"]'); const z = b ? [...b.querySelectorAll('li')].find((e) => e.innerText.includes('Großes Probeset')) : null; return z ? z.innerText : ''")
  p.pruefe('die Zeile des Sets nennt es ausgegeben', ausgegeben.includes('ausgegeben'), ausgegeben.replaceAll('\n', ' · '))
  await p.seite("document.querySelector('section[aria-label=\"Hintergrundaufgaben\"]').scrollIntoView({ block: 'center' }); return true")
  await p.bild('hintergrundaufgaben')

  await p.beende()
}
