// **Der Wechsel des Schlüsselworts unter den Regeln aus §7** — UXR-F3
// (#2009).
//
// # Warum diese Datei dasteht
//
// **Weil jetzt eines offen ist und nicht mehr alle.** Vorher stand jedes Wort
// mit seinen Feldern da, und ein Wechsel gab es nicht; mit einem Editor für
// das gewählte Wort gibt es ihn — und mit ihm die vier Regeln:
//
// 1. **Ein Wechsel räumt den offenen Kontext**: der Editor des vorigen
//    Wortes ist weg, nichts davon steht unter dem nächsten.
// 2. **Eine Antwort gehört zu dem Stand, aus dem sie stammt**: wer A
//    speichert und B wählt, bekommt A's Antwort nicht in B's Feldern.
// 3. **Geschrieben ist geschrieben**: eine offene Eingabe wird geschrieben,
//    bevor der Editor ihr Wort verlässt — auch wenn der Klick den Fokus nicht
//    bewegt (WebKit gibt einem Knopf beim Klick keinen).
// 4. **Ein Vorgang je Datensatz**: die Wahl ist keiner, sie darf also
//    mitten in einer Speicherung geschehen.
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, raeumeApi } from '../test/api'
import {
  belege,
  editor,
  fuenf,
  liste,
  texte,
  waehle,
  wort,
  zeige,
  zeile,
  type Probe,
} from '../test/schluesselwoerter'
import { Lascheninhalt } from './Lascheninhalt'

afterEach(() => {
  raeumeApi()
})

const konflikt = {
  fehler: {
    code: 'konflikt' as const,
    schluessel: 'probe',
    meldung: 'Dieses Set hat sich inzwischen geändert.',
  },
}

type Methode = (...werte: unknown[]) => Promise<unknown>

// haelt lässt eine Bindung der Probe warten, bis der Test sie entlässt.
//
// **Das Gegenstück zu `belegeApi`, das nur eine Makroaufgabe wartet**: hier
// bestimmt der Test, wann die Antwort da ist — mitten in einem Wechsel.
function haelt(
  probe: Probe,
  name: 'SchluesselwortSetzen' | 'SchluesselwortLoeschen',
  antwort?: unknown,
) {
  const echt = probe[name].getMockImplementation() as Methode
  const wartende = { entlassen: (): void => undefined }

  probe[name].mockImplementation(
    (...werte: unknown[]) =>
      new Promise((fertig) => {
        wartende.entlassen = () => {
          fertig(antwort ?? echt(...werte))
        }
      }),
  )

  return wartende
}

async function tippeInsMuster(name: string, text: string) {
  const feld = within(await editor(name)).getByRole('textbox', {
    name: texte.muster,
  })
  await userEvent.clear(feld)
  await userEvent.type(feld, text)

  return feld
}

const musterzelle = (name: string) =>
  within(screen.getByRole('row', { name })).getAllByRole('gridcell')[1]

describe('Der Wechsel während einer laufenden Speicherung', () => {
  it('zeigt das gewählte Wort, während das vorige noch schreibt', async () => {
    const probe = zeige(fuenf())
    const speichern = haelt(probe, 'SchluesselwortSetzen')

    await waehle('Kicker')
    await tippeInsMuster('Kicker', 'neu')
    await userEvent.tab()
    await waitFor(() => {
      expect(probe.SchluesselwortSetzen).toHaveBeenCalledWith(
        '3',
        '2',
        'muster',
        'neu',
        expect.any(Array),
      )
    })

    // **Mitten in der Speicherung gewählt** — die Wahl ist kein Vorgang.
    await waehle('Rausch')
    const rausch = await editor('Rausch')
    expect(
      within(rausch).getByRole('textbox', { name: texte.muster }),
    ).toHaveValue('Rausch ')

    speichern.entlassen()

    await waitFor(() => {
      expect(musterzelle('Kicker')).toHaveTextContent('neu')
    })
    // **Die Antwort für Kicker ändert nichts an Rausch**: gewählt, offen und
    // unverändert — und wieder bedienbar, denn der Vorgang ist zu Ende.
    expect(screen.getByRole('row', { name: 'Rausch' })).toHaveAttribute(
      'aria-selected',
      'true',
    )
    const muster = within(await editor('Rausch')).getByRole('textbox', {
      name: texte.muster,
    })
    expect(muster).toHaveValue('Rausch ')
    expect(muster).toBeEnabled()
  })

  // **Die Antwort gehört zu ihrem Wort** (Regel 2): scheitert A, während B
  // offen ist, steht die Meldung nicht anonym über B's Feldern, sondern nennt
  // A. Und ist A selbst offen, braucht sie den Namen nicht.
  it('nennt das Wort, dessen Speicherung scheiterte', async () => {
    const probe = zeige(fuenf())
    const speichern = haelt(probe, 'SchluesselwortSetzen', konflikt)

    await waehle('Kicker')
    await tippeInsMuster('Kicker', 'neu')
    await userEvent.tab()
    await waitFor(() => {
      expect(probe.SchluesselwortSetzen).toHaveBeenCalled()
    })
    await waehle('Rausch')

    speichern.entlassen()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Kicker — Dieses Set hat sich inzwischen geändert.',
    )

    await waehle('Kicker')
    expect(await screen.findByRole('alert')).toHaveTextContent(
      /^Dieses Set hat sich inzwischen geändert\.$/,
    )
  })

  // **Ein Wort ohne Namen nennt sich, wie die Liste es nennt** — frisch
  // angelegt ist es leer, und „ — Dieses Set …" ohne Anfang wäre ein Satz,
  // der mitten im Gedanken beginnt.
  it('nennt ein Wort ohne Namen beim Platzhalter', async () => {
    const probe = zeige([wort({ id: '1' }), wort({ id: '2', wort: 'Sog' })])
    const speichern = haelt(probe, 'SchluesselwortSetzen', konflikt)

    await waehle(texte.ohneName)
    await tippeInsMuster(texte.ohneName, 'neu')
    await userEvent.tab()
    await waitFor(() => {
      expect(probe.SchluesselwortSetzen).toHaveBeenCalled()
    })
    await waehle('Sog')

    speichern.entlassen()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      `${texte.ohneName} — Dieses Set hat sich inzwischen geändert.`,
    )
  })

  it('lässt eine Wahl stehen, die während des Löschens fiel', async () => {
    const probe = zeige(fuenf())
    const loeschen = haelt(probe, 'SchluesselwortLoeschen')

    await waehle('Kicker')
    const knopf = await screen.findByRole('button', {
      name: `${texte.loeschen} — Kicker`,
    })
    await waitFor(() => {
      expect(knopf).toBeEnabled()
    })
    await userEvent.click(knopf)
    await waitFor(() => {
      expect(probe.SchluesselwortLoeschen).toHaveBeenCalled()
    })
    await waehle('Sog')

    loeschen.entlassen()

    await waitFor(() => {
      expect(screen.queryByRole('row', { name: 'Kicker' })).toBeNull()
    })
    // **Gewählt bleibt Sog**, nicht das Wort, das auf Kicker folgte — und der
    // Fokus wird nicht von der Zeile gezogen, die der Mensch gar nicht
    // angesehen hat.
    expect(screen.getByRole('row', { name: 'Sog' })).toHaveAttribute(
      'aria-selected',
      'true',
    )
    await editor('Sog')
    expect(screen.getByRole('row', { name: 'Rausch' })).not.toHaveFocus()
  })
})

describe('Das Warten auf das neue Wort', () => {
  // **Zwischen dem Anlegen und dem Nachladen ist das neue Wort gewählt und
  // noch nicht da.** Der Editor darf in dem Augenblick weder das vorige Wort
  // zeigen noch „wähle eines" sagen — beides wäre falsch.
  it('sagt nicht, dass nichts gewählt sei, und zeigt nicht das vorige', async () => {
    const probe = zeige(fuenf())
    await waehle('Kicker')
    await editor('Kicker')

    let entlassen: () => void = () => undefined
    probe.vorLesen = () =>
      new Promise((fertig) => {
        entlassen = fertig
      })

    await userEvent.click(screen.getByRole('button', { name: texte.anlegen }))
    await waitFor(() => {
      expect(probe.SchluesselwortAnlegen).toHaveBeenCalled()
    })
    // **Mit Beleg, dass die Seite jetzt wartet**: die Liste hat das neue
    // Wort noch nicht, und die Antwort ist unterwegs.
    await antwortAbwarten()
    expect(screen.queryByRole('row', { name: texte.ohneName })).toBeNull()

    expect(screen.queryByText(texte.waehleEines)).toBeNull()
    expect(screen.queryByRole('group')).toBeNull()

    entlassen()

    await waitFor(() => {
      expect(screen.getByRole('textbox', { name: texte.wort })).toHaveFocus()
    })
  })
})

describe('Eine offene Eingabe beim Wechsel', () => {
  // **WebKit gibt einem Knopf beim Klick keinen Fokus**: das Feld behält ihn,
  // verliert kein `blur`, und der Editor wird mit dem Text darin abgeräumt.
  // `fireEvent` bewegt den Fokus nicht — wie dort.
  it('wird geschrieben, bevor ein anderes Wort gewählt wird', async () => {
    const probe = zeige(fuenf())

    await waehle('Kicker')
    const feld = await tippeInsMuster('Kicker', 'neu')
    expect(feld).toHaveFocus()

    const knopf = within(await zeile('Rausch')).getByRole('button')
    fireEvent.pointerDown(knopf)
    fireEvent.click(knopf)

    await waitFor(() => {
      expect(probe.SchluesselwortSetzen).toHaveBeenCalledWith(
        '3',
        '2',
        'muster',
        'neu',
        expect.any(Array),
      )
    })
    await editor('Rausch')
  })

  it('wird geschrieben, bevor ein neues Wort den Editor ablöst', async () => {
    const probe = zeige(fuenf())

    await waehle('Kicker')
    await tippeInsMuster('Kicker', 'neu')

    const anlegen = screen.getByRole('button', { name: texte.anlegen })
    fireEvent.pointerDown(anlegen)
    fireEvent.click(anlegen)

    await waitFor(() => {
      expect(probe.SchluesselwortSetzen).toHaveBeenCalledWith(
        '3',
        '2',
        'muster',
        'neu',
        expect.any(Array),
      )
    })
  })

  // **Ohne Eingabe im Editor geschieht nichts dergleichen** — ein Fokus
  // außerhalb (hier: in der Liste) wird nicht angetastet.
  it('lässt einen Fokus außerhalb des Editors in Ruhe', async () => {
    const probe = zeige(fuenf())

    await waehle('Kicker')
    await editor('Kicker')
    const zeileKicker = await zeile('Kicker')
    zeileKicker.focus()

    await waehle('Sog')

    expect(probe.SchluesselwortSetzen).not.toHaveBeenCalled()
  })
})

describe('Die Wahl über einen neuen Stand', () => {
  // **Eine Rücknahme baut den Editor neu, aber nicht die Wahl ab** — die
  // Felder stehen auf `defaultValue` und brauchen den Neuaufbau; wer danach
  // wieder von vorn wählen müsste, hätte die Rücknahme bestraft.
  it('behält die Wahl, wenn der Verlauf einen Schritt zurückgeht', async () => {
    const probe = zeige(fuenf())

    await waehle('Kicker')
    await editor('Kicker')
    Object.assign(probe.bestand.find((eines) => eines.id === '2') ?? {}, {
      muster: 'zurück',
    })

    probe.zeichneNeu(1)

    await waitFor(() => {
      expect(
        within(screen.getByRole('group', { name: /Kicker/ })).getByRole(
          'textbox',
          { name: texte.muster },
        ),
      ).toHaveValue('zurück')
    })
    expect(screen.getByRole('row', { name: 'Kicker' })).toHaveAttribute(
      'aria-selected',
      'true',
    )
  })

  // **Ein Wort, das es nicht mehr gibt, ist nicht gewählt** — die Wahl hält
  // eine Kennung, kein Wort, und eine Kennung ohne Wort ist keine Wahl.
  it('wählt nichts mehr, wenn das Wort verschwunden ist', async () => {
    const probe = zeige(fuenf())

    await waehle('Kicker')
    await editor('Kicker')
    probe.bestand = probe.bestand.filter((eines) => eines.id !== '2')

    probe.zeichneNeu(1)

    expect(await screen.findByText(texte.waehleEines)).toBeInTheDocument()
    expect(screen.queryByRole('group')).toBeNull()
  })
})

describe('Der Wechsel des Sets', () => {
  // **Regel 1 für das Bezugsobjekt Set**: der `key` in `Lascheninhalt` räumt
  // Wahl, Editor und Meldung mit ab (`docs/SCHREIBWEGE.md`, Gestalt „Ein
  // `key`").
  it('räumt die Wahl mit ab', async () => {
    belege(fuenf())
    const lasche = (setID: string) => (
      <Sprachraum sprache="de">
        <Lascheninhalt
          seite="schluesselwoerter"
          setID={setID}
          karteID={undefined}
          wechseln={() => undefined}
          oeffneKarte={() => undefined}
          oeffneSymbol={() => undefined}
          bestandGeaendert={() => undefined}
          verlaufsstand={0}
          verlaufsfehler={undefined}
        />
      </Sprachraum>
    )
    const { rerender } = render(lasche('3'))

    await waehle('Kicker')
    await editor('Kicker')

    rerender(lasche('4'))

    expect(await screen.findByText(texte.waehleEines)).toBeInTheDocument()
    expect(
      within(await liste()).queryAllByRole('row', { selected: true }),
    ).toHaveLength(0)
  })
})
