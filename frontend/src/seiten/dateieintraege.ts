// **Die Handgriffe an einer Vorlagendatei** — für das Menü »Datei« über dem
// Editor **und** den Rechtsklick an der Datei im Baum (AP-5,
// #2362).
// Eine Liste an einer Stelle: keine Aktion soll es nur im Kontextmenü geben
// (U2, #1413), und keine nur im Menü.
import type { Menueeintrag } from '../bausteine/Menue'
import type { Dateiaktionen as Handgriffe } from '../haken/useDateiaktionen'
import type { Katalog } from '../sprache/kataloge'

export function dateieintraege(
  aktionen: Handgriffe,
  hauptdatei: boolean,
  gesperrt: boolean,
  texte: Katalog['seiten']['vorlagen'],
): Menueeintrag[] {
  const { setzeModus } = aktionen

  // **Die Hauptdatei bietet weder Umbenennen noch Löschen an**; Ungesichertes
  // lässt beide ruhen.
  //
  // **Ein Eintrag trägt eine Aktion nur, wenn sein Schlüssel eine nennt** —
  // `aktionskatalog.test.ts`; „Neue Datei" ist keine Katalogaktion.
  const andereDatei: Menueeintrag[] = [
    {
      schluessel: 'umbenennen',
      name: texte.dateiUmbenennen,
      aktion: 'bearbeiten',
      angehalten: gesperrt,
      aufWahl: () => {
        setzeModus('umbenennen')
      },
    },
    {
      schluessel: 'entfernen',
      name: texte.dateiLoeschen,
      aktion: 'loeschen',
      angehalten: gesperrt,
      aufWahl: () => {
        setzeModus('entfernen')
      },
    },
  ]
  return [
    {
      schluessel: 'neu',
      name: texte.dateiNeu,
      aufWahl: () => {
        setzeModus('anlegen')
      },
    },
    {
      schluessel: 'hinzufuegen',
      name: texte.dateiHinzufuegen,
      aktion: 'anlegen',
      angehalten: aktionen.laeuft,
      aufWahl: () => void aktionen.einfuegen(),
    },
    ...(hauptdatei ? [] : andereDatei),
  ]
}
