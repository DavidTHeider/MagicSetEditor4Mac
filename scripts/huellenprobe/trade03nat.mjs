// **Genau eine Buchung, auch wenn jemand zweimal drückt und neu startet** —
// TRADE-03.NAT-01 (#1717).
//
// Ein Tausch steht nach beidseitiger Zustimmung, Vorbereitung und Übergabe in
// *Verbuchen*. Am **nativen Fenster** wird verbucht, gleich noch einmal, die
// Anwendung beendet und neu gestartet, und ein drittes Mal verbucht.
//
// **Seit #2060 steht der Knopf nur, solange `moegliche` ihn nennt** — nach der
// eigenen Buchung ist er fort (`Handelsbuchung.tsx`). Bis zum 2026-10-05
// drückte dieses Drehbuch ihn trotzdem ein zweites Mal und blieb daran
// hängen. Jetzt prüft es, **dass er fort ist**, und geht den zweiten und
// dritten Druck über den Weg, den der Knopf ginge: Freigabe, dann Buchung,
// über die Bindungen des Fensters. Danach
// muss der Bestand **eine** Bewegung tragen, der Stand „Bei dir verbucht"
// heißen, und als Nächstes steht der Abgleich an — **bei dir**, denn die
// Matrix lässt nach der eigenen Buchung den Abgleich melden; was von der
// Gegenseite fehlt, sagt die Buchung selbst („Nur du hast unterschrieben").
//
// **Das Gegenüber ist hier eine Kennung und kein zweites Programm** — seine
// Schritte geht die Kommandozeile mit `[wer]`, wie in `trade02`. Ob die
// Gegenstelle ihrerseits genau einmal bucht, prüft der Draht zweier
// Anwendungen in `internal/flaechen/handelsdraht_test.go`.
import { execFileSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import { kindumgebung } from '../huellenorte.mjs'

const wurzel = fileURLToPath(new URL('../..', import.meta.url))

// eroeffne legt einen Tausch an, in dem das Gegenüber eine Karte gibt.
function eroeffne(ort, gegenueber) {
  const positionen = JSON.stringify([
    { von: gegenueber, printing: 'bolt-m10', menge: 1, guete: 'nearMint', sprache: 'de', finish: 'nonfoil' },
  ])

  return execFileSync(`${wurzel}.go/bin/mse`, ['handel', 'eroeffnen', '', 'tausch', gegenueber], {
    cwd: wurzel,
    input: positionen,
    env: kindumgebung(ort),
  }).toString().split('\t')[0]
}

const zeile = "[...document.querySelectorAll('li button')].find((b) => b.innerText.startsWith('bnat'))"
const verbuchen = "document.querySelector('[data-handelsassistent] nav li:nth-child(4) button')"
const buchung = "[...document.querySelectorAll('section')].find((s) => s.querySelector('h3')?.innerText === 'Einen Handel verbuchen')"
const knopf = `[...((${buchung})?.querySelectorAll('button') ?? [])].find((b) => b.innerText.trim() === 'Verbuchen')`
const buchungstext = `return (${buchung})?.innerText ?? ''`

// oeffne wählt den Tausch und schlägt *Verbuchen* auf — **auf die Wirkung
// gewartet**, nicht auf den Klick (§3).
async function oeffne(p) {
  await p.klickeAuf(zeile, 'die Zeile des Tauschs')
  await p.warte('der Tausch ist gewählt', () => p.seite(`return (${zeile})?.getAttribute('aria-current') === 'true'`), 20000)
  await p.klickeAuf(verbuchen, '»Verbuchen« in der Schrittleiste')
}

// drueckeVerbuchen drückt den Knopf und wartet, bis der Stand dasteht.
async function drueckeVerbuchen(p, wann) {
  await p.klickeAuf(knopf, `»Verbuchen« (${wann})`)
  await p.warte(`verbucht (${wann})`, async () => (await p.seite(buchungstext)).includes('Bei dir verbucht'), 20000)
}

// nochmals geht den Weg des Knopfes ohne den Knopf — Freigabe, dann Buchung —
// und gibt zurück, wo er endete: an der Freigabe, an der Buchung oder gebucht.
async function nochmals(p, vorgang) {
  const v = JSON.stringify(vorgang)

  return p.seite(`const b = window.go.app.Bindungen; const f = await b.HandelFreigabeErteilen(${v}, 'verbuchen', ''); if (f.fehler) return 'Freigabe abgewiesen: ' + f.fehler.code; const a = await b.Handelsverbuchung(${v}, f.kennung); return a.fehler ? 'Buchung abgewiesen: ' + a.fehler.code : 'angenommen'`)
}

// bewegungen zählt, was das Journal zu diesem Vorgang trägt.
function bewegungen(p, ort, vorgang) {
  return JSON.parse(p.mse(ort, 'handelsjournal', vorgang, '--json')).length
}

export default async function trade03nat(p) {
  const ort = p.ort()

  p.notiz('1 — ein Tausch bis zur Übergabe')
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'peer-anlegen')
  const vorgang = eroeffne(ort, 'bnat')
  for (const schritt of [['anbieten'], ['zustimmen'], ['zustimmen', 'bnat'], ['vorbereiten'], ['uebergeben'], ['uebergeben', 'bnat']]) {
    p.mse(ort, 'handel', schritt[0], vorgang, ...schritt.slice(1))
  }
  p.pruefe('er steht in der Übergabe', JSON.parse(p.mse(ort, 'handel', 'stand', vorgang, '--json')).zustand === 'uebergabeOffen')

  await p.starte({ ort })
  await p.taste('5', 'befehl')

  p.notiz('2 — verbuchen, und gleich noch einmal')
  await oeffne(p)
  await drueckeVerbuchen(p, 'erstes Mal')
  p.pruefe('eine Bewegung', bewegungen(p, ort, vorgang) === 1, String(bewegungen(p, ort, vorgang)))
  await p.warte('der Knopf ist fort', () => p.seite(`return !(${knopf})`), 10000)
  p.pruefe('»Verbuchen« steht nach der Buchung nicht mehr da', true)
  const zweiter = await nochmals(p, vorgang)
  p.zeige(`der zweite Druck über die Bindungen: ${zweiter}`)
  const nachZweitem = await p.seite(buchungstext)
  p.zeige(nachZweitem.replace(/\s+/g, ' '))
  p.pruefe('keine Meldung eines Fehlers', !(await p.seite(`return [...((${buchung})?.querySelectorAll('[role=alert]') ?? [])].some((e) => e.innerText.trim() !== '')`)))
  p.pruefe('weiter eine Bewegung', bewegungen(p, ort, vorgang) === 1, String(bewegungen(p, ort, vorgang)))
  await p.seite(`(${buchung})?.scrollIntoView({ block: 'center' })`)
  await p.bild('zweimal')
  p.pruefe('⌘Q beendet', await p.beende())

  p.notiz('3 — nach dem Neustart')
  await p.starte({ ort })
  await p.taste('5', 'befehl')
  await oeffne(p)
  await p.warte('der Stand steht', async () => (await p.seite(buchungstext)).includes('Bei dir verbucht'), 20000)
  p.pruefe('auch nach dem Neustart kein »Verbuchen«', !(await p.seite(`return !!(${knopf})`)))
  p.zeige(`der dritte Druck über die Bindungen: ${await nochmals(p, vorgang)}`)
  p.pruefe('noch immer eine Bewegung', bewegungen(p, ort, vorgang) === 1, String(bewegungen(p, ort, vorgang)))

  const stand = JSON.parse(p.mse(ort, 'handel', 'stand', vorgang, '--json'))
  p.zeige(`Stand ${stand.zustand}, Lage ${stand.lage}, nächster Schritt ${stand.naechster || '—'}`)
  p.pruefe('bei dir verbucht', stand.zustand === 'lokalVerbucht', stand.zustand)
  p.pruefe('als Nächstes der Abgleich', stand.lage === 'dran' && stand.naechster === 'abgleichen', `${stand.lage} ${stand.naechster}`)
  p.pruefe('das Fenster sagt es', (await p.seite("return document.querySelector('[data-handelsassistent]')?.innerText ?? ''")).includes('Du bist dran: Abgleich melden'))
  p.pruefe('und was von der Gegenseite fehlt', (await p.seite(buchungstext)).includes('Nur du hast unterschrieben'))

  const protokoll = JSON.parse(p.mse(ort, 'handel', 'protokoll', vorgang, '--json')).zeilen
  p.zeige(protokoll.map((z) => `${z.von || '-'} → ${z.nach}`).join(' | '))
  p.pruefe('genau ein Wechsel nach »lokal verbucht«', protokoll.filter((z) => z.nach === 'lokalVerbucht').length === 1)
  await p.seite(`(${buchung})?.scrollIntoView({ block: 'center' })`)
  await p.bild('neustart')

  p.pruefe('⌘Q beendet', await p.beende())
}
