// **Die Menüleiste am nativen Fenster** — #255
// (Issue (#255)) und
// ENG-02.AK-04, Kernweg 2.
//
// Was #255 als „nur am Fenster zu sehen" offen ließ, hier in der echten
// Leiste, so wie AppKit sie hält:
//
// 1. Beschriftungen, Kürzel und Rollen in der Sprache des Fensters,
// 2. _Dienste_ ist das Dienstemenü von macOS,
// 3. _Ausblenden_ und _Alle einblenden_ wirken, _Alle nach vorne bringen_ auch,
// 4. ⌘Z in einem Textfeld nimmt die Eingabe zurück,
// 5. die Rücktaste löscht im Textfeld (_Löschen_ trägt kein Kürzel),
// 6. ein Sprachwechsel benennt die Leiste um — und zurück.
//
// **Was macOS selbst anhängt** — Schreibtools, Automatisch ausfüllen,
// Diktat, Emoji —, folgt der Sprache des **Systems**, nicht der des
// Fensters: eine Anwendung gibt AppKit ihre Sprache nur beim Start. Die
// Zusage „kein englischer Titel in der deutschen Leiste" gilt deshalb bei
// einem deutschen System; welche Sprachen es führte, steht im Beleg.
import { execFileSync } from 'node:child_process'

const deutsch = {
  anwendung: ['Über MagicSetEditor4Mac', 'Einstellungen …', 'Dienste', 'MagicSetEditor4Mac ausblenden', 'Andere ausblenden', 'Alle einblenden', 'MagicSetEditor4Mac beenden'],
  bearbeiten: ['Widerrufen', 'Wiederholen', 'Ausschneiden', 'Kopieren', 'Einsetzen', 'Löschen', 'Alles auswählen', 'Sprachausgabe'],
  fenster: ['Im Dock ablegen', 'Zoomen', 'Alle nach vorne bringen'],
}
const englisch = {
  anwendung: ['About MagicSetEditor4Mac', 'Settings …', 'Services', 'Hide MagicSetEditor4Mac', 'Hide Others', 'Show All', 'Quit MagicSetEditor4Mac'],
  bearbeiten: ['Undo', 'Redo', 'Cut', 'Copy', 'Paste', 'Delete', 'Select All', 'Speech'],
}

// ordnung ist die Folge der Menüs oben — die eigenen zwischen Darstellung
// und Fenster, wie macOS sie führt (UXR-02, #1910).
const ordnung = {
  de: ['MagicSetEditor4Mac', 'Ablage', 'Bearbeiten', 'Darstellung', 'Set', 'Fenster', 'Hilfe'],
  en: ['MagicSetEditor4Mac', 'File', 'Edit', 'View', 'Set', 'Window', 'Help'],
}

// englisch sind die Titel, die eine englische Leiste von macOS trägt — keiner
// davon darf in einer deutschen stehen, auch nicht in dem, was macOS selbst
// anhängt.
const englischeTitel = new Set(['Services', 'Hide Others', 'Show All', 'Undo', 'Redo', 'Cut', 'Copy', 'Paste', 'Delete', 'Select All', 'Speech', 'Start Speaking', 'Stop Speaking', 'Minimize', 'Zoom', 'Bring All to Front', 'Writing Tools', 'AutoFill', 'Start Dictation…', 'Emoji & Symbols', 'Enter Full Screen', 'Show Tab Bar', 'Show All Tabs', 'Window', 'Edit', 'File', 'View', 'Help'])
const alleTitel = (punkte) => punkte.flatMap((p) => (p.trenner || p.verborgen ? [] : [p.titel, ...alleTitel(p.unter ?? [])]))

const sichtbar = (punkte) => punkte.filter((p) => !p.trenner && !p.verborgen)
// **Leerzeichen gleich gemacht**: der Katalog setzt vor „…" ein schmales
// geschütztes, und darauf kommt es hier nicht an.
const titel = (punkte) => sichtbar(punkte).map((p) => p.titel.replace(/\s/gu, ' '))
const oben = (leiste, name) => leiste.find((p) => p.titel === name)
const eintrag = (menue, name) => sichtbar(menue.unter).find((p) => p.titel === name)

async function feldMitText(p, text) {
  // **Die Befehlspalette hat ein Suchfeld mit Fokus** — ⌘K geht dabei selbst
  // den Weg einer Taste: Leiste, dann Webansicht.
  await p.seite("window.__tasten = []; addEventListener('keydown', (e) => window.__tasten.push(e.key + (e.metaKey ? '+meta' : '')), true); return true")
  await p.taste('k', 'befehl')
  try {
    await p.warte('⌘K öffnet die Palette mit Fokus im Suchfeld', () =>
      p.seite("return document.activeElement?.tagName === 'INPUT'"))
  } catch (grund) {
    p.zeige(`Seite: ${JSON.stringify(await p.seite("return { fokus: document.activeElement?.tagName + ' ' + (document.activeElement?.getAttribute('aria-label') ?? ''), tasten: window.__tasten, dialoge: [...document.querySelectorAll('[role=dialog]')].map((d) => d.getAttribute('aria-label') ?? d.id) }"))}, Fenster: ${JSON.stringify(await p.fenster())}`)
    throw grund
  }
  await p.tippe(text)

  return p.warte(`im Feld steht »${text}«`, async () =>
    (await p.seite('return document.activeElement.value')) === text)
}

export default async function leiste(p) {
  await p.starte()
  // **Nach vorn kommt sie nicht**: seit macOS 14 aktiviert sich eine
  // Anwendung nur, wenn die vordere ihr den Vortritt lässt, und ein Start
  // aus einem Hintergrundprozess tut das nicht. Tasten und Klicks erreichen
  // sie trotzdem — über ihre eigene Ereignisschlange; die Leiste am oberen
  // Bildschirmrand zeigt dann aber die der vorderen Anwendung, und deshalb
  // belegt diese Probe die Leiste mit dem, was AppKit hält, und nicht mit
  // einem Bild davon.
  p.zeige(`vorn: ${String((await p.fenster()).aktiv)}`)
  p.zeige(`Sprachen des Systems: ${execFileSync('defaults', ['read', '-g', 'AppleLanguages']).toString().replace(/\s+/g, ' ').trim()}`)

  p.notiz('1 — die Leiste auf Deutsch')
  const de = await p.menue()
  p.zeige(`oben: ${titel(de).join(' · ')}`)
  const anwendung = oben(de, 'MagicSetEditor4Mac')
  const bearbeiten = oben(de, 'Bearbeiten')
  const fenster = oben(de, 'Fenster')
  p.pruefe('Anwendungsmenü, Bearbeiten und Fenster stehen da', Boolean(anwendung && bearbeiten && fenster))
  // **Die eigenen Menüs zwischen Darstellung und Fenster** — UXR-02 (#1910),
  // nach Apples „The menu bar" (*App-specific menus*); bis dahin stand *Set*
  // vor Bearbeiten.
  p.pruefe('die Menüs stehen in der Ordnung von macOS', titel(de).join(' · ') === ordnung.de.join(' · '), titel(de).join(' · '))
  p.pruefe('das Anwendungsmenü ist deutsch', JSON.stringify(titel(anwendung.unter)) === JSON.stringify(deutsch.anwendung), titel(anwendung.unter).join(' · '))
  p.pruefe('Bearbeiten beginnt mit den eigenen Einträgen', JSON.stringify(titel(bearbeiten.unter).slice(0, deutsch.bearbeiten.length)) === JSON.stringify(deutsch.bearbeiten), titel(bearbeiten.unter).join(' · '))
  const englischGefunden = alleTitel(de).filter((t) => englischeTitel.has(t))
  p.pruefe('kein englischer Titel in der ganzen Leiste — auch nicht in dem, was macOS anhängt', englischGefunden.length === 0, englischGefunden.join(', '))
  p.pruefe('Fenster ist deutsch', deutsch.fenster.every((name) => titel(fenster.unter).includes(name)), titel(fenster.unter).join(' · '))

  const kuerzel = (menue, name) => eintrag(menue, name)?.kuerzel ?? ''
  const aktion = (menue, name) => eintrag(menue, name)?.aktion ?? ''
  for (const [menue, name, zeichen, wirkung] of [
    [anwendung, 'Einstellungen …', '⌘,', 'eigene'],
    [anwendung, 'MagicSetEditor4Mac ausblenden', '⌘H', 'hide:'],
    [anwendung, 'Andere ausblenden', '⌥⌘H', 'hideOtherApplications:'],
    [anwendung, 'Alle einblenden', '', 'unhideAllApplications:'],
    [anwendung, 'MagicSetEditor4Mac beenden', '⌘Q', 'terminate:'],
    [bearbeiten, 'Widerrufen', '⌘Z', 'undo:'],
    [bearbeiten, 'Wiederholen', '⇧⌘Z', 'redo:'],
    [bearbeiten, 'Ausschneiden', '⌘X', 'cut:'],
    [bearbeiten, 'Kopieren', '⌘C', 'copy:'],
    [bearbeiten, 'Einsetzen', '⌘V', 'paste:'],
    [bearbeiten, 'Löschen', '', 'delete:'],
    [bearbeiten, 'Alles auswählen', '⌘A', 'selectAll:'],
    [fenster, 'Im Dock ablegen', '⌘M', 'performMiniaturize:'],
    [fenster, 'Zoomen', '', 'performZoom:'],
    [fenster, 'Alle nach vorne bringen', '', 'arrangeInFront:'],
  ]) {
    p.pruefe(`»${name}« hat ${zeichen || 'kein Kürzel'} und ${wirkung}`,
      kuerzel(menue, name) === zeichen && aktion(menue, name) === wirkung,
      `${kuerzel(menue, name) || '—'}, ${aktion(menue, name) || '—'}`)
  }
  const sprache = eintrag(bearbeiten, 'Sprachausgabe')
  p.pruefe('Sprachausgabe startet und stoppt', JSON.stringify(sichtbar(sprache?.unter ?? []).map((e) => e.aktion)) === JSON.stringify(['startSpeaking:', 'stopSpeaking:']))

  p.notiz('4 und 5 — ⌘Z und die Rücktaste im Textfeld')
  await feldMitText(p, 'abc')
  await p.taste('Backspace')
  p.pruefe('die Rücktaste löscht ein Zeichen', await p.warte('»ab«', async () => (await p.seite('return document.activeElement.value')) === 'ab'))
  // **⌘Z erreicht die Seite, und die Seite lässt es durch** — dann nimmt
  // AppKit es als Kürzel von »Widerrufen«. Den letzten Schritt geht die
  // Probe über den Eintrag selbst (`imFenster`): eine Anwendung, die aus
  // einem Hintergrundprozess startet, kommt seit macOS 14 nicht nach vorn und
  // hat kein Schlüsselfenster, an dessen Ersten Antwortenden AppKit schicken
  // könnte. Ein Mensch hätte vorher ins Fenster geklickt.
  await p.seite("window.__z = null; addEventListener('keydown', (e) => { if (e.key === 'z' && e.metaKey) setTimeout(() => { window.__z = e.defaultPrevented ? 'verhindert' : 'durchgelassen' }) }, true); return true")
  await p.taste('z', 'befehl')
  const z = await p.warte('⌘Z kommt an der Seite an', () => p.seite('return window.__z'))
  p.pruefe('die Seite lässt ⌘Z im Textfeld durch', z === 'durchgelassen', z)
  await p.imFenster('Bearbeiten', 'Widerrufen')
  const nachWiderruf = await p.warte('»Widerrufen« nimmt zurück', async () => {
    const wert = await p.seite('return document.activeElement.value')
    return wert !== 'ab' ? { wert } : false
  })
  p.pruefe('»Widerrufen« nimmt die Eingabe im Feld zurück, nicht eine fachliche Änderung', ['abc', ''].includes(nachWiderruf.wert), `danach »${nachWiderruf.wert}«`)
  await p.bild('palette-nach-widerruf')
  await p.taste('Escape')

  p.notiz('2 — Dienste zeigt die Dienste von macOS')
  // **Geöffnet, nicht gelesen**: macOS trägt die Dienste erst ein, wenn das
  // Menü aufgeht. Die Leitung klappt es über dem Fenster auf, liest es offen
  // und schließt es wieder.
  const dienste = await p.frage({ was: 'aufklappen', pfad: ['MagicSetEditor4Mac', 'Dienste'] })
  p.pruefe('»Dienste« zeigt beim Öffnen, was macOS anbietet', dienste.length > 0,
    `${String(dienste.length)} Einträge, zuerst: ${dienste.slice(0, 4).join(' · ')}`)

  p.notiz('3 — Ausblenden, Alle einblenden, Alle nach vorne')
  await p.klick('MagicSetEditor4Mac', 'MagicSetEditor4Mac ausblenden')
  await p.warte('»Ausblenden« blendet die Anwendung aus', async () => (await p.fenster()).verborgen)
  p.pruefe('ausgeblendet', true)
  await p.klick('MagicSetEditor4Mac', 'Alle einblenden')
  await p.warte('»Alle einblenden« bringt sie zurück', async () => !(await p.fenster()).verborgen)
  p.pruefe('wieder eingeblendet', true)
  await p.klick('Fenster', 'Alle nach vorne bringen')
  const vorn = await p.fenster()
  p.pruefe('»Alle nach vorne bringen« lässt das Fenster sichtbar und vorn', vorn.sichtbar && !vorn.verborgen, JSON.stringify({ sichtbar: vorn.sichtbar, schluessel: vorn.schluessel }))

  p.notiz('6 — die Sprache wechseln, und die Leiste folgt')
  await p.taste(',', 'befehl')
  await p.warte('⌘, öffnet die Einstellungen', () => p.seite("return [...document.querySelectorAll('select')].some((w) => w.checkVisibility() && w.labels?.[0]?.textContent.includes('Sprache'))"))
  const waehle = (wert) => p.seite(`const s = [...document.querySelectorAll('select')].find((w) => w.checkVisibility() && w.labels?.[0]?.textContent.match(/Sprache|Language/)); s.value = '${wert}'; s.dispatchEvent(new Event('change', { bubbles: true })); return true`)
  await waehle('en')
  const en = await p.warte('die Leiste wird englisch', async () => {
    const leiste = await p.menue()
    return oben(leiste, 'Edit') ? leiste : false
  })
  p.zeige(`oben: ${titel(en).join(' · ')}`)
  p.pruefe('und auf Englisch in derselben Ordnung', titel(en).join(' · ') === ordnung.en.join(' · '), titel(en).join(' · '))
  p.pruefe('das Anwendungsmenü ist englisch', JSON.stringify(titel(oben(en, 'MagicSetEditor4Mac').unter)) === JSON.stringify(englisch.anwendung), titel(oben(en, 'MagicSetEditor4Mac').unter).join(' · '))
  p.pruefe('Edit beginnt mit den eigenen Einträgen auf Englisch', JSON.stringify(titel(oben(en, 'Edit').unter).slice(0, englisch.bearbeiten.length)) === JSON.stringify(englisch.bearbeiten), titel(oben(en, 'Edit').unter).join(' · '))
  // **Was macOS anhängt, folgt der Sprache des Systems** — nicht der des
  // Fensters. Eine Anwendung kann AppKit nur beim Start eine andere Sprache
  // geben; hier steht, was es bei einem deutschen System zeigt.
  p.zeige(`von macOS angehängt: ${titel(oben(en, 'Edit').unter).slice(englisch.bearbeiten.length).join(' · ')}`)
  p.pruefe('Window und File heißen englisch', Boolean(oben(en, 'Window') && oben(en, 'File')))
  const services = await p.frage({ was: 'aufklappen', pfad: ['MagicSetEditor4Mac', 'Services'] })
  p.pruefe('»Services« zeigt weiter die Dienste', services.length > 0, `${String(services.length)} Einträge`)
  await waehle('de')
  await p.warte('und wieder deutsch', async () => Boolean(oben(await p.menue(), 'Bearbeiten')))
  p.pruefe('zurück auf Deutsch', true)

  p.notiz('Beenden über ⌘Q')
  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
}
