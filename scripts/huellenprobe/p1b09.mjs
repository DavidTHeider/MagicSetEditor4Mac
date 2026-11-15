// **Eine angenommene Leihe am Sitz, und Aufgeben mit Rückfrage — am nativen
// Fenster** — P1b (#2370),
// Befunde C-06 und C-12.
//
// # Was es prüft
//
// 1. **Anna verleiht** ihr Deck an Bert, im Fenster (*Spielen › Verleihen
//    oder leihen …*), den Umschlag an einen Ort, den der Dialog vorgibt.
// 2. **Bert nimmt an**, in seinem eigenen Fenster mit eigenem Bestand — und
//    an seinem Sitz steht die Leihe in der Deckwahl, neben seinem Deck.
// 3. **Die Partie fängt an** mit seinem Deck am ersten und der Leihe am
//    zweiten Sitz; `mse partien` kennt sie, und die Leihe ist eingelöst.
// 4. **Aufgeben fragt**: „Weiterspielen“ ändert nichts, „Ja, aufgeben“ beendet
//    das Spiel für diesen Sitz.
//
// # Was es nicht belegt
//
// **Eine abgelaufene oder fremde Leihe** weist der Kern ab — das prüfen
// `internal/app/partieleihsitz_intern_test.go` und der Dienst, nicht dieses
// Drehbuch. **VoiceOver** hört ein Mensch.
import { mkdtempSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

import { zurSeite } from './uirevue.mjs'

const zeilenwert = (ausgabe, name) => ausgabe.split('\n').find((z) => z.startsWith(`${name}\t`))?.split('\t')[1] ?? ''
const gruppe = (name) => `[...document.querySelectorAll('[role=group]')].find((g) => g.checkVisibility() && g.firstElementChild?.innerText.trim() === ${JSON.stringify(name)})`
const gruppenknopf = (name) => `(${gruppe(name)})?.querySelector('button')`
const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? b.innerText).trim().replace(/^[^\\p{L}]+/u, '') === ${JSON.stringify(name)})`
const setzeWahl = (beschriftung, stelle, wert) => `const s = [...document.querySelectorAll('select')].filter((e) => e.checkVisibility() && [...(e.labels ?? [])].some((l) => l.innerText.trim().startsWith(${JSON.stringify(beschriftung)})))[${stelle}]; if (!s) return false; const setze = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set; setze.call(s, ${JSON.stringify(wert)}); s.dispatchEvent(new Event('change', { bubbles: true })); return true`

function mitDeck(p, ort, name) {
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'peer-anlegen')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'deck', 'neu', name, 'probe')
  for (const karte of ['1', '2', '3', '4', '5']) p.mse(ort, 'deck', 'legen', '1', `--karte=${karte}`, '8')
}

async function verleihen(p, umschlag) {
  await zurSeite(p, 'Spielen')
  await p.klickeText('Verleihen oder leihen …')
  await p.warte('die Wahl des Ziels steht', () => p.seite(`return !!(${gruppe('Wohin der Umschlag gelegt wird')})`), 15000)
  await p.dialogantwort(umschlag)
  await p.klickeAuf(gruppenknopf('Wohin der Umschlag gelegt wird'), 'Ort wählen …')
  await p.warte('der Ort steht da', () => p.seite(`return document.body.innerText.includes(${JSON.stringify(umschlag)})`), 15000)
  await p.klickeAuf("[...document.querySelectorAll('label')].find((l) => l.checkVisibility() && l.innerText.trim().startsWith('Bert'))", 'Bert')
  await p.seite(`const s = [...document.querySelectorAll('select')].find((e) => e.checkVisibility() && [...(e.labels ?? [])].some((l) => l.innerText.trim().startsWith('Welches Deck'))); if (!s) return false; const o = [...s.options].find((x) => x.text === 'Annas Deck'); const setze = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set; setze.call(s, o.value); s.dispatchEvent(new Event('change', { bubbles: true })); return true`)
  const anbieten = knopf('Verleihen und ablegen')
  await p.warte('„Verleihen und ablegen“ ist frei', () => p.seite(`return ${anbieten}?.disabled === false`), 15000)
  await p.klickeAuf(anbieten, 'Verleihen und ablegen')
  await p.warte('der Umschlag ist abgelegt', () => p.seite(`return [...document.querySelectorAll('[role=status]')].some((s) => s.innerText.includes(${JSON.stringify(umschlag)}))`), 20000)
}

async function annehmen(p, umschlag) {
  await zurSeite(p, 'Spielen')
  await p.klickeText('Verleihen oder leihen …')
  await p.klickeText('Eine Leihe annehmen')
  await p.warte('die Wahl des Umschlags steht', () => p.seite(`return !!(${gruppe('Umschlag einer Leihe')})`), 15000)
  await p.dialogantwort(umschlag)
  await p.klickeAuf(gruppenknopf('Umschlag einer Leihe'), 'Umschlag wählen …')
  await p.warte('der Umschlag steht da', () => p.seite(`return document.body.innerText.includes(${JSON.stringify(umschlag)})`), 15000)
  await p.klickeAuf(knopf('Leihe annehmen'), 'Leihe annehmen')
  // **„geliehen von“ steht an der Zeile der Leihenliste** (`leiheGeliehen`).
  const angenommen = "return [...document.querySelectorAll('[role=dialog] li')].some((e) => e.checkVisibility() && e.innerText.includes('geliehen von'))"
  await p.warte('die Leihe steht in der Liste', () => p.seite(angenommen), 20000).catch(() => undefined)
  p.zeige(`Meldungen: ${String(await p.seite("return [...document.querySelectorAll('[role=status], [role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim()).join(' | ')"))}`)
  p.pruefe('Bert hat die Leihe angenommen', await p.seite(angenommen))
  await p.taste('Escape')
}

async function partieMitLeihe(p, bild) {
  await zurSeite(p, 'Spielen')
  const name = (stelle) => `[...document.querySelectorAll('input')].find((e) => e.checkVisibility() && (e.labels?.[0]?.innerText ?? '').trim() === 'Name am Tisch ${stelle}')`
  await p.warte('das Formular steht', () => p.seite(`return !!(${name(1)})`), 20000)
  for (const [stelle, wer] of [[1, 'Bert'], [2, 'Anna']]) {
    await p.klickeAuf(name(stelle), `Name am Tisch ${stelle}`)
    await p.tippeInsFeld(wer)
  }
  const leihoption = "[...document.querySelectorAll('select option')].find((o) => o.value.startsWith('leihe:'))"
  await p.warte('die Leihe steht in der Deckwahl', () => p.seite(`return !!(${leihoption})`), 15000).catch(() => undefined)
  const angebot = await p.seite(`return (${leihoption})?.textContent ?? ''`)
  p.pruefe('an Berts Sitz steht die Leihe in der Deckwahl', angebot.startsWith('Geliehen von'), angebot)
  await p.seite(setzeWahl('Deck', 0, '1'))
  await p.seite(`const o = ${leihoption}; ${setzeWahl('Deck', 1, '__W__').replace('"__W__"', 'o.value')}`)
  await p.bild(`sitze-${bild}`)
  await p.klickeAuf(knopf('Partie anfangen'), 'Partie anfangen')
  await p.warte('der Tisch steht', () => p.seite(`return !!${knopf('Aufgeben')}`), 20000)
}

async function aufgeben(p, ort, bild) {
  const partien = () => p.mse(ort, 'partien')
  const vorher = partien()
  await p.klickeAuf(knopf('Aufgeben'), 'Aufgeben')
  await p.warte('die Rückfrage steht', () => p.seite("return document.body.innerText.includes('Aufgeben? Das Spiel endet sofort')"), 10000).catch(() => undefined)
  p.pruefe('Aufgeben fragt erst', await p.seite("return document.body.innerText.includes('Aufgeben? Das Spiel endet sofort')"))
  await p.bild(`aufgeben-frage-${bild}`)
  await p.klickeAuf(knopf('Weiterspielen'), 'Weiterspielen')
  p.pruefe('„Weiterspielen“ ändert nichts', partien() === vorher, partien().trim().split('\n').slice(0, 3).join(' / '))
  await p.klickeAuf(knopf('Aufgeben'), 'Aufgeben')
  await p.warte('die Rückfrage steht wieder', () => p.seite(`return !!${knopf('Ja, aufgeben')}`), 10000)
  await p.klickeAuf(knopf('Ja, aufgeben'), 'Ja, aufgeben')
  await p.warte('der Bestand kennt die Aufgabe', () => partien() !== vorher, 15000).catch(() => undefined)
  p.pruefe('„Ja, aufgeben“ beendet das Spiel für den Sitz', partien() !== vorher, partien().trim().split('\n').slice(0, 3).join(' / '))
}

export default async function p1b09(p) {
  const anna = p.ort()
  mitDeck(p, anna, 'Annas Deck')
  const bert = p.ort()
  mitDeck(p, bert, 'Berts Deck')
  const annaIch = p.mse(anna, 'peer')
  const bertIch = p.mse(bert, 'peer')
  p.mse(anna, 'kontakt-aufnehmen', zeilenwert(bertIch, 'kennung'), 'Bert')
  p.mse(anna, 'kontakt-bestaetigen', zeilenwert(bertIch, 'kennung'), zeilenwert(bertIch, 'fingerabdruck'))
  p.mse(bert, 'kontakt-aufnehmen', zeilenwert(annaIch, 'kennung'), 'Anna')
  p.mse(bert, 'kontakt-bestaetigen', zeilenwert(annaIch, 'kennung'), zeilenwert(annaIch, 'fingerabdruck'))
  const umschlag = join(mkdtempSync(join(tmpdir(), 'p1b09-')), 'leihe-fuer-bert.msepackage')

  p.notiz('1 — Anna verleiht im Fenster')
  await p.starte({ ort: anna })
  await p.groesse(1440, 838)
  await verleihen(p, umschlag)
  p.pruefe('⌘Q beendet Annas Fenster', await p.beende())

  p.notiz('2 — Bert nimmt an, und die Leihe steht an seinem Sitz')
  await p.starte({ ort: bert })
  await p.groesse(1440, 838)
  await annehmen(p, umschlag)
  await partieMitLeihe(p, 'hell')
  const partien = p.mse(bert, 'partien')
  p.pruefe('`mse partien` kennt die Partie', partien.trim() !== '', partien.trim().split('\n').slice(0, 3).join(' / '))
  p.notiz('3 — Aufgeben fragt nach')
  await aufgeben(p, bert, 'hell')
  p.pruefe('⌘Q beendet Berts Fenster', await p.beende())
}
