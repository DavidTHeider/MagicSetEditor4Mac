// **Die Menüleiste unter einem englischen System** — #255, die Handabnahme
// aus #1871.
//
// `leiste.mjs` belegt ein deutsches System. Was macOS selbst an _Bearbeiten_
// hängt — Diktat, Emoji, Schreibtools, Automatisch ausfüllen —, folgt der
// Sprache, die AppKit beim Start für **diesen Prozess** wählt. Die Probe
// setzt sie mit `-AppleLanguages (en)` und das Fenster auf Englisch — **ohne
// die Einstellung des Rechners anzufassen**: was eine Anwendung als
// Systemsprache sieht, ist genau diese Liste.
//
// Geprüft wird, dass die ganze Leiste englisch ist — auch das, was macOS
// anhängt — und dass kein deutscher Titel stehen bleibt.
import { execFileSync } from 'node:child_process'

const englisch = {
  anwendung: ['About MagicSetEditor4Mac', 'Settings …', 'Services', 'Hide MagicSetEditor4Mac', 'Hide Others', 'Show All', 'Quit MagicSetEditor4Mac'],
  bearbeiten: ['Undo', 'Redo', 'Cut', 'Copy', 'Paste', 'Delete', 'Select All', 'Speech'],
}

// deutscheTitel sind die, die eine deutsche Leiste trägt — keiner davon darf
// in der englischen stehen, auch nicht in dem, was macOS anhängt.
const deutscheTitel = new Set(['Bearbeiten', 'Fenster', 'Ablage', 'Darstellung', 'Hilfe', 'Widerrufen', 'Wiederholen', 'Ausschneiden', 'Kopieren', 'Einsetzen', 'Löschen', 'Alles auswählen', 'Sprachausgabe', 'Dienste', 'Andere ausblenden', 'Alle einblenden', 'Im Dock ablegen', 'Zoomen', 'Alle nach vorne bringen', 'Diktat starten …', 'Diktat starten…', 'Emoji & Symbole', 'Schreibtools', 'Automatisch ausfüllen', 'Vollbildmodus aktivieren'])

const sichtbar = (punkte) => punkte.filter((p) => !p.trenner && !p.verborgen)
const titel = (punkte) => sichtbar(punkte).map((p) => p.titel.replace(/\s/gu, ' '))
const alleTitel = (punkte) => punkte.flatMap((p) => (p.trenner || p.verborgen ? [] : [p.titel.replace(/\s/gu, ' '), ...alleTitel(p.unter ?? [])]))
const oben = (leiste, name) => leiste.find((p) => p.titel === name)

export default async function leisteenglisch(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'en')

  p.notiz('1 — gestartet mit -AppleLanguages (en), das Fenster auf Englisch')
  p.zeige(`Sprachen des Rechners (unverändert): ${execFileSync('defaults', ['read', '-g', 'AppleLanguages']).toString().replace(/\s+/g, ' ').trim()}`)
  await p.starte({ ort, sprachen: 'en' })

  const leiste = await p.menue()
  p.zeige(`oben: ${titel(leiste).join(' · ')}`)

  const anwendung = oben(leiste, 'MagicSetEditor4Mac')
  const bearbeiten = oben(leiste, 'Edit')
  const fenster = oben(leiste, 'Window')
  p.pruefe('Anwendungsmenü, Edit und Window stehen da', Boolean(anwendung && bearbeiten && fenster))
  // **Die eigenen Menüs zwischen View und Window** — UXR-02 (#1910).
  p.pruefe('die Menüs stehen in der Ordnung von macOS', titel(leiste).join(' · ') === 'MagicSetEditor4Mac · File · Edit · View · Set · Window · Help', titel(leiste).join(' · '))
  p.pruefe('das Anwendungsmenü ist englisch', JSON.stringify(titel(anwendung.unter)) === JSON.stringify(englisch.anwendung), titel(anwendung.unter).join(' · '))
  p.pruefe('Edit beginnt mit den eigenen Einträgen', JSON.stringify(titel(bearbeiten.unter).slice(0, englisch.bearbeiten.length)) === JSON.stringify(englisch.bearbeiten), titel(bearbeiten.unter).join(' · '))

  p.notiz('2 — was macOS an Edit hängt')
  // **Geöffnet, nicht nur gelesen** — wie die Dienste in `leiste.mjs`: was
  // macOS anhängt, kann erst beim Aufklappen dastehen.
  const offen = await p.frage({ was: 'aufklappen', pfad: ['Edit'] })
  const angehaengt = titel(bearbeiten.unter).slice(englisch.bearbeiten.length)
  p.zeige(`angehängt (gelesen): ${angehaengt.join(' · ') || '—'}`)
  p.zeige(`aufgeklappt: ${offen.join(' · ')}`)
  const offenDeutsch = offen.filter((t) => deutscheTitel.has(t.replace(/\s/gu, ' ')))
  p.pruefe('aufgeklappt steht kein deutscher Titel in Edit', offenDeutsch.length === 0, offenDeutsch.join(', '))
  const deutschGefunden = alleTitel(leiste).filter((t) => deutscheTitel.has(t))
  p.pruefe('kein deutscher Titel in der ganzen Leiste — auch nicht in dem, was macOS anhängt', deutschGefunden.length === 0, deutschGefunden.join(', '))
  await p.bild('leiste-englisch')
}
