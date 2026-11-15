// **Die Nutzerseite für einen Lauf** — UXR-20
// (#1928).
//
// **Dieselbe Form wie `test/paketseite.tsx`**: die Belegung, ohne die die
// Seite keine Laschen zeigt (`belegeAnmeldung`), und darüber, was der Lauf
// selbst meint.
import { render } from '@testing-library/react'
import { vi } from 'vitest'

import { type Api, type NutzerDTO } from '../api'
import { Nutzerseite } from '../seiten/Nutzerseite'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi } from './api'
import { belegeAnmeldung, einNutzer } from './anmeldebelegung'
import { probeprofil } from './profil'

export function belegeSeite(teile: Partial<Api>) {
  return belegeApi({ ...belegeAnmeldung(), ...teile })
}

export function zeigeNutzerseite() {
  render(
    <Sprachraum sprache="de">
      <Nutzerseite />
    </Sprachraum>,
  )
}

// mensch baut einen Nutzer mit seinen Profilen.
function mensch(
  id: string,
  name: string,
  profile: readonly (readonly [string, string])[],
): NutzerDTO {
  return {
    ...einNutzer(),
    id,
    name,
    profile: profile.map(([kennung, profil]) =>
      probeprofil({ id: kennung, nutzerID: id, name: profil }),
    ),
  }
}

// wahlen sind `anzahl` Einträge für eine Vorgabe.
const wahlen = (anzahl: number, wort: string) =>
  Array.from({ length: anzahl }, (_, stelle) => ({
    wert: String(stelle + 1),
    beschriftung: `${wort} ${String(stelle + 1)}`,
  }))

// **Ein Bestand, wie ihn ein Haushalt hat** — drei Menschen, sieben Profile,
// zwei Freigaben, drei Token, fünf Vorgaben mit zwölf Decks und sechs
// Lagerorten. **Der Probebestand der Hüllenprobe hat einen Menschen und ein
// Profil**, und an ihm fiel die Formularwand nicht auf (UXR-20).
export const vielBestand = [
  mensch('1', 'Ich', [
    ['10', 'Mein Profil'],
    ['11', 'Werkstatt'],
    ['12', 'Turnier'],
    ['13', 'Automat'],
  ]),
  mensch('2', 'Anna', [['20', 'Annas Sicht']]),
  mensch('3', 'Ben', [
    ['30', 'Bens Sammlung'],
    ['31', 'Draft'],
  ]),
]

export function belegeVielBestand(teile: Partial<Api> = {}) {
  return belegeSeite({
    Nutzer: vi.fn(() => Promise.resolve({ nutzer: vielBestand })),
    GezeigtesProfil: vi.fn(() =>
      Promise.resolve({ profil: vielBestand[0]?.profile[0] ?? null }),
    ),
    Anbieterkonten: vi.fn(() =>
      Promise.resolve({
        konten: [{ anbieter: 'github', anzeigename: 'ich' }],
      }),
    ),
    Faktoren: vi.fn(() =>
      Promise.resolve({
        faktoren: [{ art: 'totp', bezeichnung: 'TOTP' }],
        biometrie: true,
        biometrieart: 'Touch ID',
        biometriegrund: '',
      }),
    ),
    Freigaben: vi.fn(() =>
      Promise.resolve({
        freigaben: [
          {
            profilID: '10',
            nutzerID: '2',
            name: 'Anna',
            rolle: 'leser',
            angelegt: '',
          },
          {
            profilID: '10',
            nutzerID: '3',
            name: 'Ben',
            rolle: 'mitarbeiter',
            angelegt: '',
          },
        ],
        sichtbarkeit: 'privat',
        eigeneRolle: 'verwalter',
      }),
    ),
    Profilvorgaben: vi.fn(() =>
      Promise.resolve({
        vorgaben: [
          {
            art: 'format',
            mehrwertig: false,
            werte: [],
            moeglich: wahlen(8, 'Format'),
            verwaist: [],
          },
          {
            art: 'spiel',
            mehrwertig: false,
            werte: [],
            moeglich: wahlen(3, 'Spiel'),
            verwaist: [],
          },
          {
            art: 'vorlage',
            mehrwertig: false,
            werte: [],
            moeglich: wahlen(20, 'Vorlage'),
            verwaist: [],
          },
          {
            art: 'deck',
            mehrwertig: true,
            werte: [],
            moeglich: wahlen(12, 'Deck'),
            verwaist: [],
          },
          {
            art: 'lagerort',
            mehrwertig: true,
            werte: [],
            moeglich: wahlen(6, 'Ort'),
            verwaist: [],
          },
        ],
      }),
    ),
    Token: vi.fn(() =>
      Promise.resolve({
        token: ['Druckerei', 'Bot', 'Sicherung'].map((bezeichnung, stelle) => ({
          id: String(stelle + 1),
          profilID: '10',
          bezeichnung,
          principal: 'dienst',
          scope: 'lesen',
          ablauf: '',
          angelegt: '2026-09-01T00:00:00Z',
          abgelaufen: false,
          letzteNutzung: '',
          vorgaengerBis: '',
          partie: '',
          sitz: '',
          fassung: '1',
        })),
      }),
    ),
    Tokenablaufwahlen: vi.fn(() =>
      Promise.resolve({ wahlen: [{ name: 'nie', ablauf: '' }] }),
    ),
    ...teile,
  })
}

// felder zählt, was auf der Seite an Feldern gezeichnet ist — wie das
// Drehbuch `uirevue`: Eingaben, Auswahlen, Textfelder, Schalter. **Was
// verborgen ist, zählt nicht**: ein `Aufklappkasten` zeichnet erst beim
// Öffnen und verbirgt danach.
export function gezeichneteFelder(): number {
  return [
    ...document.body.querySelectorAll(
      'input:not([type=hidden]), select, textarea, [role=combobox], [role=switch]',
    ),
  ].filter((feld) => feld.closest('[hidden]') === null).length
}
