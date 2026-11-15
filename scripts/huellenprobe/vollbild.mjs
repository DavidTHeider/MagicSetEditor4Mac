// **Vollbild hinein und heraus am nativen Fenster** — UX-14.AK-04
// (#1686)
// und #1994:
// der Eintrag, den macOS an *Darstellung* hängt, sagt im Vollbild, dass er es
// beendet — Wails schaltet die Frage danach ab (`vollbild_darwin.go`) —, und
// ⌃⌘F öffnet nicht die Suche der Seite.
//
// **Gegen den Titel des Systems und nicht gegen einen Wortlaut**: wie macOS
// den Eintrag nennt, hängt an der Sprache des Systems. Geprüft wird, dass er
// sich im Vollbild ändert und danach wieder heißt wie vorher.

const finde = (punkte, eltern = []) => {
  for (const punkt of punkte) {
    if (punkt.aktion === 'toggleFullScreen:') return { punkt, pfad: [...eltern, punkt.titel] }
    const unten = finde(punkt.unter ?? [], [...eltern, punkt.titel])
    if (unten) return unten
  }
  return undefined
}

const nah = (a, b) => Math.abs(a - b) <= 2

export default async function vollbild(p) {
  await p.starte()

  // **Erst aufgeklappt** — macOS hängt den Eintrag an *Darstellung*, wenn
  // das Menü zum ersten Mal aufgeht; gelesen ohne Aufklappen steht er nicht
  // da (erster Lauf am 2026-10-01).
  const darstellung = (await p.menue()).find((punkt) => ['Darstellung', 'View'].includes(punkt.titel))
  p.pruefe('Darstellung steht in der Leiste', darstellung !== undefined)
  if (darstellung === undefined) return
  p.zeige(`aufgeklappt: ${(await p.frage({ was: 'aufklappen', pfad: [darstellung.titel] })).join(' · ')}`)
  await p.taste('Escape')
  await new Promise((los) => setTimeout(los, 500))

  const vorher = finde(await p.menue())
  p.pruefe('macOS hängt einen Vollbild-Eintrag an', vorher !== undefined, vorher?.pfad.join(' › '))
  if (vorher === undefined) return

  const groesse = (await p.fenster()).rahmen

  p.notiz('Über die Leiste hinein')
  await p.klick(...vorher.pfad)
  await p.warte('das Fenster ist im Vollbild', async () => (await p.fenster()).vollbild, 10000)
  await new Promise((los) => setTimeout(los, 1500))
  const im = finde(await p.menue())
  p.pruefe('im Vollbild heißt der Eintrag anders', im !== undefined && im.punkt.titel !== vorher.punkt.titel,
    `»${vorher.punkt.titel}« → »${im?.punkt.titel ?? '—'}«`)
  await p.bild('vollbild-im')

  p.notiz('Über die Leiste heraus')
  await p.klick(...(im ?? vorher).pfad)
  await p.warte('das Fenster ist nicht mehr im Vollbild', async () => !(await p.fenster()).vollbild, 10000)
  await new Promise((los) => setTimeout(los, 1500))
  const nachher = finde(await p.menue())
  p.pruefe('danach heißt er wieder wie vorher', nachher?.punkt.titel === vorher.punkt.titel, nachher?.punkt.titel)
  const danach = (await p.fenster()).rahmen
  p.pruefe('die Größe ist die von vorher', nah(danach.B, groesse.B) && nah(danach.H, groesse.H),
    `${String(groesse.B)} × ${String(groesse.H)} → ${String(danach.B)} × ${String(danach.H)}`)

  // **⌃⌘F gehört dem System und nicht der Suche.** Seit macOS 15 trägt der
  // Eintrag *fn F*; ⌃⌘F schaltet auf macOS 26 auch Safari nicht mehr ins
  // Vollbild (gemessen am 2026-10-01 in der Abnahme-VM). Geprüft wird
  // deshalb, was #1994 war: die Seite schlug bei ⌃⌘F die Suche auf.
  p.notiz('⌃⌘F öffnet nicht die Suche')
  // **Gemessen an der Taste und nicht an einer Seite**: ein Horcher ganz
  // hinten am Fenster sieht, ob die Anwendung die Taste für sich nahm
  // (`preventDefault`). ⌘F ist der Gegenbeleg, dass jemand zuhört.
  await p.seite("window.__gesehen = []; addEventListener('keydown', (e) => { if (e.key === 'f') window.__gesehen.push({ ctrl: e.ctrlKey, meta: e.metaKey, genommen: e.defaultPrevented }) }); return true")
  await p.taste('f', 'ctrl', 'befehl')
  await new Promise((los) => setTimeout(los, 1000))
  if ((await p.fenster()).vollbild) {
    await p.taste('f', 'ctrl', 'befehl')
    await p.warte('wieder im Fenster', async () => !(await p.fenster()).vollbild, 10000)
  }
  await p.taste('f', 'befehl')
  await new Promise((los) => setTimeout(los, 1000))
  const gesehen = await p.seite('return window.__gesehen')
  p.zeige(`gesehen: ${JSON.stringify(gesehen)}`)
  const mitCtrl = gesehen.filter((g) => g.ctrl && g.meta)
  const ohneCtrl = gesehen.filter((g) => !g.ctrl && g.meta)
  p.pruefe('⌃⌘F nimmt die Seite nicht für sich', mitCtrl.every((g) => !g.genommen), JSON.stringify(mitCtrl))
  p.pruefe('⌘F nimmt sie — jemand hört zu', ohneCtrl.some((g) => g.genommen), JSON.stringify(ohneCtrl))

  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
}
