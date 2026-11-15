// **Eine Setkarte: der Name führt, das Seltene steht im Menü** — UXR-04
// (#1912).
//
// # Was gemessen war
//
// Jede Karte trug neben ihrem Namen **fünf gleichrangige Knöpfe** — Sichern,
// Sichern unter …, Als Umschlag ablegen, Über eine Exportvorlage …, Löschen …
// —, dazu *+ Marke*: sieben Bedienelemente je Set. Die Liste zeigt sechzig auf
// einmal, das sind 420 (R6 aus `docs/UIREVIEW.md` erlaubt 200), und schon vier
// Karten im Bild reißen R2. Das Artboard *Sets* zeichnet je Set eine Zeile,
// die öffnet — und sonst nichts.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { SetDTO } from '../api'
import type { Gestalt } from '../bausteine/Listenansicht'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { belegeApi, raeumeApi } from '../test/api'
import { stillerSpeicherlauf } from '../test/speicherlauf'
import { stillerUmschlaglauf } from '../test/umschlaglauf'
import type { Setloeschung } from '../haken/useSetloeschung'
import type { Speicherlauf } from '../haken/useSetSpeichern'
import type { Umschlaglauf } from '../haken/useUmschlag'
import { Setkarte } from './Setkarte'

afterEach(() => {
  raeumeApi()
})

const worte = kataloge.de.seiten.sets
const zustaende = kataloge.de.bausteine.zustaende

const einSet: SetDTO = {
  id: '7',
  name: 'Mein Set',
  spiel: 'magic',
  stil: 'm15',
  karten: 3,
  herkunft: '/tmp/mein.mse-set',
  herkunftFehlt: false,
}

function ruhendeLoeschung(teile: Partial<Setloeschung> = {}): Setloeschung {
  return {
    loeschen: vi.fn(() => Promise.resolve()),
    laeuft: () => false,
    befunde: () => [],
    fehler: () => undefined,
    ...teile,
  }
}

function zeige({
  speichern = stillerSpeicherlauf(),
  loeschung = ruhendeLoeschung(),
  umschlag = stillerUmschlaglauf(),
  waehlen = vi.fn(),
  gestalt = 'kachel',
}: {
  gestalt?: Gestalt
  speichern?: Speicherlauf
  loeschung?: Setloeschung
  umschlag?: Umschlaglauf
  waehlen?: (setID: string) => void
} = {}) {
  return render(
    <Sprachraum sprache="de">
      <Setkarte
        gestalt={gestalt}
        satz={einSet}
        waehlen={waehlen}
        speichern={speichern}
        loeschung={loeschung}
        umschlag={umschlag}
      />
    </Sprachraum>,
  )
}

const menueName = einsetzen(worte.handlungenAn, { name: 'Mein Set' })

async function waehleImMenue(eintrag: string) {
  await userEvent.click(screen.getByRole('button', { name: menueName }))
  await userEvent.click(screen.getByRole('menuitem', { name: eintrag }))
}

describe('Die Setkarte', () => {
  // **Als Kachel und als Zeile dieselben** — die Gestalt ordnet an, sie
  // ändert keine Handlung (UX-11).
  it.each(['kachel', 'zeile'] as const)(
    'zeigt als %s den Namen und ein Menü — und keine Reihe gleichrangiger Knöpfe',
    (gestalt) => {
      zeige({ gestalt })

      expect(
        screen.getAllByRole('button').map((knopf) => knopf.textContent),
      ).toEqual(['Mein Set', worte.handlungenKurz])
    },
  )

  // **Der Dateiname steht da, der Pfad im Titel** — AP-11 (#2368), B-27:
  // ein Pfad aus `/var/folders/…` an der Karte ist Technik, gebraucht wird
  // sein Ende. Vorgelesen wird der ganze.
  it('nennt den Dateinamen und trägt den Pfad im Titel', () => {
    zeige()

    const name = screen.getByText('mein.mse-set')
    expect(name).toHaveAttribute('aria-hidden', 'true')
    expect(name.closest('p')).toHaveAttribute('title', '/tmp/mein.mse-set')
    expect(screen.getByText('/tmp/mein.mse-set')).toHaveClass('sr-only')
  })

  it('schlägt das Set über seinen Namen auf', async () => {
    const waehlen = vi.fn()
    zeige({ waehlen })

    await userEvent.click(
      screen.getByRole('button', { name: `Mein Set — ${worte.oeffnen}` }),
    )

    expect(waehlen).toHaveBeenCalledWith('7')
  })

  it('führt im Menü das Seltene, das Löschen zuletzt', async () => {
    zeige()

    await userEvent.click(screen.getByRole('button', { name: menueName }))

    const menue = screen.getByRole('menu', { name: menueName })
    expect(
      within(menue)
        .getAllByRole('menuitem')
        .map((eintrag) => eintrag.textContent),
    ).toEqual([
      `⤓${worte.speichern}`,
      `⤓${worte.speichernUnter}`,
      worte.umschlagAblegen,
      worte.vorlageAusfuhr,
      `✕${worte.loeschen}`,
    ])
  })

  it.each([
    [worte.speichern, false],
    [worte.speichernUnter, true],
  ])('sichert über »%s«', async (eintrag, unter) => {
    const speichern = stillerSpeicherlauf()
    zeige({ speichern })

    await waehleImMenue(eintrag)

    expect(speichern.speichern).toHaveBeenCalledWith('7', unter)
  })

  it('hält das Sichern an, solange es für dieses Set läuft', async () => {
    zeige({ speichern: stillerSpeicherlauf({ laeuft: (id) => id === '7' }) })

    await userEvent.click(screen.getByRole('button', { name: menueName }))

    expect(
      screen.getByRole('menuitem', { name: worte.speichern }),
    ).toHaveAttribute('aria-disabled', 'true')
  })
})

describe('Der Umschlag an der Setkarte', () => {
  it('legt das Set als Umschlag ab', async () => {
    const umschlag = stillerUmschlaglauf()
    zeige({ umschlag })

    await waehleImMenue(worte.umschlagAblegen)

    expect(umschlag.schnueren).toHaveBeenCalledWith('7')
  })

  it('ist angehalten, solange dieses Set hinausgeht', async () => {
    zeige({
      umschlag: stillerUmschlaglauf({
        laeuft: (was) => was === 'schnueren-7',
      }),
    })

    await userEvent.click(screen.getByRole('button', { name: menueName }))

    expect(
      screen.getByRole('menuitem', { name: worte.umschlagAblegen }),
    ).toHaveAttribute('aria-disabled', 'true')
  })

  // **Die Rückmeldung steht an der Karte, aus der es kam** (§6). Bis UXR-04
  // stand sie nur in der Annahme unter „Ein Set von woanders einlesen" — und
  // die war zugeklappt, also gar nicht gezeichnet.
  it('meldet das Ablegen an der Karte', () => {
    zeige({
      umschlag: stillerUmschlaglauf({
        fuer: '7',
        geschnuert: {
          objekte: [],
          urheber: '',
          fingerabdruck: '',
          pfad: '/tmp/mein.msepackage',
          abgebrochen: false,
        },
      }),
    })

    expect(screen.getByRole('status')).toHaveTextContent(
      worte.umschlagGeschnuert,
    )
  })

  it('meldet einen Fehler des Ablegens an der Karte', () => {
    zeige({
      umschlag: stillerUmschlaglauf({ fuer: '7', fehler: 'Kein Platz.' }),
    })

    expect(screen.getByRole('alert')).toHaveTextContent(
      `${zustaende.fehler} Kein Platz.`,
    )
  })

  it('schweigt, wenn ein anderes Set hinausging', () => {
    zeige({
      umschlag: stillerUmschlaglauf({ fuer: '8', fehler: 'Kein Platz.' }),
    })

    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })
})

describe('Das Löschen an der Setkarte', () => {
  it('fragt erst, wenn es im Menü gewählt ist — und dann mit dem Fokus im Feld', async () => {
    zeige()

    expect(screen.queryByLabelText(worte.loeschenFeld)).not.toBeInTheDocument()

    await waehleImMenue(worte.loeschen)

    expect(screen.getByLabelText(worte.loeschenFeld)).toHaveFocus()
    expect(screen.getByRole('alert')).toHaveTextContent('Mein Set')
  })

  it('räumt die Frage samt getipptem Namen weg, wenn jemand abbricht', async () => {
    zeige()

    await waehleImMenue(worte.loeschen)
    await userEvent.type(screen.getByLabelText(worte.loeschenFeld), 'Mein')
    await userEvent.click(
      screen.getByRole('button', { name: worte.loeschenNein }),
    )

    expect(screen.queryByLabelText(worte.loeschenFeld)).not.toBeInTheDocument()

    // **Der getippte Name geht mit** — sonst wäre der nächste Klick auf
    // „Löschen …" schon halb bestätigt.
    await waehleImMenue(worte.loeschen)
    expect(screen.getByLabelText(worte.loeschenFeld)).toHaveValue('')
  })

  it('löscht mit dem getippten Namen', async () => {
    const loeschen = vi.fn(() => Promise.resolve())
    zeige({ loeschung: ruhendeLoeschung({ loeschen }) })

    await waehleImMenue(worte.loeschen)
    await userEvent.type(screen.getByLabelText(worte.loeschenFeld), 'Mein Set')
    await userEvent.click(
      screen.getByRole('button', { name: worte.loeschenJa }),
    )

    expect(loeschen).toHaveBeenCalledWith('7', 'Mein Set')
  })

  // **Die Reste eines geglückten Löschens stehen, auch ohne offene Frage.**
  it('zeigt die Befunde eines geglückten Löschens', () => {
    zeige({
      loeschung: ruhendeLoeschung({
        befunde: () => ['Die Bilder ließen sich nicht wegräumen.'],
      }),
    })

    expect(screen.getByRole('status')).toHaveTextContent(
      'Die Bilder ließen sich nicht wegräumen.',
    )
  })

  it('ist angehalten, solange das Set gelöscht wird', async () => {
    zeige({ loeschung: ruhendeLoeschung({ laeuft: (id) => id === '7' }) })

    await userEvent.click(screen.getByRole('button', { name: menueName }))

    expect(
      screen.getByRole('menuitem', { name: worte.loeschen }),
    ).toHaveAttribute('aria-disabled', 'true')
  })
})

describe('Die Exportvorlage an der Setkarte', () => {
  // **Geladen wird beim Öffnen, nicht beim Bau der Liste** — dreißig Sets
  // führten sonst dreißig Paketläufe. Und mit dem Menü genügt **ein** Schritt:
  // wer „Über eine Exportvorlage …" wählt, hat gesagt, was er will.
  it('lädt die Vorlagen erst, wenn sie im Menü gewählt ist — und schließt wieder', async () => {
    const Exportvorlagen = vi.fn(() =>
      Promise.resolve({
        vorlagen: [
          {
            paket: 'probe-ausfuhr.mse-export-template',
            name: 'Ausfuhr der Probe',
            dateityp: '*.txt|*.txt',
            optionen: [],
          },
        ],
        befunde: [],
      }),
    )
    belegeApi({
      Exportvorlagen,
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
    })
    zeige()

    expect(Exportvorlagen).not.toHaveBeenCalled()

    await waehleImMenue(worte.vorlageAusfuhr)

    expect(
      await screen.findByRole('combobox', { name: worte.vorlageWaehlen }),
    ).toBeInTheDocument()
    expect(Exportvorlagen).toHaveBeenCalledTimes(1)

    await userEvent.click(
      screen.getByRole('button', { name: worte.vorlageSchliessen }),
    )

    expect(
      screen.queryByRole('combobox', { name: worte.vorlageWaehlen }),
    ).not.toBeInTheDocument()
  })
})
