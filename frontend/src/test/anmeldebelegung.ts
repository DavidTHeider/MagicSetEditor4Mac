// **Was jede Belegung der Nutzerseite mitbringen muss** — Werkstück 4 von
// Phase 4.12
// (#1351).
//
// # Warum es diesen Helfer gibt
//
// **Die Nutzerseite holt sechs Antworten in einem Zug** (§7, Regel 2). Wer nur
// eine belegt, bekommt von `src/test/api.ts` einen Fehler mit dem Namen der
// fehlenden Methode — richtig so, aber in jedem Lauf derselbe. Dieser Helfer
// legt die fünf dazu, um die es dem einzelnen Lauf nicht geht.
import { vi } from 'vitest'

import type { Api, NutzerDTO } from '../api'

// einNutzer ist der Mensch, den ein Lauf ohne eigene Belegung meint.
//
// **Mit einem Profil**, denn ohne eines zeigt die Seite ihre Laschen gar
// nicht — und ein Wächter über eine Seite, die nicht gezeichnet wurde, prüft
// nichts (§10).
export function einNutzer(): NutzerDTO {
  return {
    id: '1',
    name: 'Ich',
    bild: '',
    email: '',
    telefon: '',
    anonym: false,
    anrede: '',
    namenszusatz: '',
    vorname: '',
    weiterenamen: '',
    nachname: '',
    slogan: '',
    kurzebio: '',
    langebio: '',
    fassung: '1',
    profile: [
      {
        id: '10',
        nutzerID: '1',
        name: 'Mein Profil',
        fassung: '1',
        subslogan: '',
        kurzebeschreibung: '',
        langebeschreibung: '',
        badge: '',
        technisch: false,
        ablauf: '',
      },
    ],
  }
}

export function belegeAnmeldung(): Partial<Api> {
  return {
    // **Der Mensch selbst und das Profil, in dem er arbeitet** — ohne diese
    // beiden zeigt die Seite „Die Methode Nutzer ist nicht belegt." und hat
    // **keine Laschen**.
    Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einNutzer()] })),
    GezeigtesProfil: vi.fn(() =>
      Promise.resolve({ profil: einNutzer().profile[0] ?? null }),
    ),

    // **Und wer arbeitet** — Werkstück 17 von Phase 4.12 (#1400). Die Seite
    // fragt im selben Zug; ohne Beleg stünde neben der Liste eine zweite
    // Meldung, und derselbe Wächter prüfte wieder eine Fehlermeldung.
    WerArbeitet: vi.fn(() => Promise.resolve({ nutzerID: einNutzer().id })),

    Anbieter: vi.fn(() =>
      Promise.resolve({ anbieter: ['github', 'google', 'microsoft', 'x'] }),
    ),
    Anbieterkonten: vi.fn(() => Promise.resolve({ konten: [] })),

    // **Und die Profilbilder** — Werkstück 8 von Phase 4.12 (#1351). Sie
    // kommen im selben Zug wie die Liste; **leer** ist der Fall, den jeder
    // Lauf ohne eigene Belegung meint.
    Nutzerbilder: vi.fn(() => Promise.resolve({ bilder: {}, badges: {} })),

    // **Und die zweiten Faktoren** — Werkstück 7 von Phase 4.12 (#1351). Die
    // Lasche *Konto* zeigt sie, seit es sie gibt; ohne diesen Beleg bekäme
    // jeder Lauf über diese Lasche eine zweite Meldung, die mit seinem Fall
    // nichts zu tun hat.
    //
    // **Ohne Biometrie**, weil ein Rechner ohne Sensor der Fall ist, den
    // jeder Lauf ohne eigene Belegung meint.
    Faktoren: vi.fn(() =>
      Promise.resolve({
        faktoren: [],
        biometrie: false,
        biometrieart: '',
        biometriegrund: 'in diesem Lauf gibt es keinen Sensor',
      }),
    ),

    // **Und die Zugriffstoken des gezeigten Profils** — Werkstück 10 von
    // Phase 4.12 (#1351). Die Lasche *Profile* lädt sie, seit es sie gibt;
    // ohne diesen Beleg bekäme jeder Lauf dort eine zweite Meldung, die mit
    // seinem Fall nichts zu tun hat — und das ist teuer, weil sie ein
    // `role="alert"` ist und `findByRole('alert')` dann zwei findet.
    //
    // **Leer ist der Fall, den jeder Lauf ohne eigene Belegung meint**;
    // geprüft werden sie in `bausteine/zugriffstoken.test.tsx`.
    Token: vi.fn(() => Promise.resolve({ token: [] })),

    // **Und die Vorgaben des gezeigten Profils** — Werkstück 15 desselben
    // Issues, aus demselben Grund: die Lasche *Profile* lädt sie.
    Profilvorgaben: vi.fn(() => Promise.resolve({ vorgaben: [] })),

    // **Und die Freigaben des gezeigten Profils** — Werkstück 11 desselben
    // Issues, aus demselben Grund.
    //
    // **`verwalter` ist der Fall, den ein Lauf ohne eigene Belegung meint**:
    // die Lasche zeigt dann das Formular, so wie sie es dem Eigner zeigt.
    Freigaben: vi.fn(() =>
      Promise.resolve({
        freigaben: [],
        sichtbarkeit: 'privat',
        eigeneRolle: 'verwalter',
      }),
    ),
  }
}
