// Proben für die Kartenseite — eine Zeile, eine ganze Karte, ein Feld.
//
// Sie stehen hier und nicht in jeder Testdatei, weil die Seite seit
// #891 **drei**
// Gestalten derselben Karten kennt: die Liste bekommt Zeilen ohne Feldwerte,
// der Editor holt die aufgeschlagene Karte einzeln, und die Massenänderung ein
// Feld über das ganze Set. Wer alle drei in jeder Datei von Hand nachbaut, hält
// sie nicht zusammen — und genau daraus entsteht eine Vorrichtung, die
// großzügiger ist als der Betrieb (§3).
import type { KarteDTO, KartenzeileDTO } from '../api'

export function probeZeile(
  teile: Partial<KartenzeileDTO> = {},
): KartenzeileDTO {
  return { id: '5', nummer: 0, name: 'Feuerball', felder: 1, ...teile }
}

export function probeKarte(teile: Partial<KarteDTO> = {}): KarteDTO {
  return {
    id: '5',
    nummer: 0,
    werte: [{ feld: 'name', wert: 'Feuerball' }],
    ...teile,
  }
}

// belegeKartenwege belegt alle drei Wege aus **einer** Liste von Zeilen.
//
// **Der Name der Zeile wird zum Feldwert der Karte.** Sonst zeigte die Liste
// „Feuerball" und der Editor daneben etwas anderes — ein Widerspruch, den ein
// Test nicht bemerkt, weil er nur eine Hälfte ansieht.
export function belegeKartenwege(zeilen: readonly KartenzeileDTO[]) {
  const werte = (zeile: KartenzeileDTO) =>
    zeile.felder === 0 ? [] : [{ feld: 'name', wert: zeile.name }]

  return {
    Kartenzeilen: () => Promise.resolve({ zeilen: [...zeilen] }),

    // **Eine Karte, die es nicht gibt, ist der leere Verbund** — dieselbe
    // Antwort wie auf der Go-Seite, wo `leereKarte()` mit leerer Werteliste
    // zurückkommt. Ein `undefined` wäre großzügiger als der Betrieb.
    Karte: (karteID: string) => {
      const zeile = zeilen.find((eine) => eine.id === karteID)

      return Promise.resolve({
        karte: zeile
          ? { id: zeile.id, nummer: zeile.nummer, werte: werte(zeile) }
          : { id: '', nummer: 0, werte: [] },
      })
    },

    // **Nur Karten, die das Feld haben.** Dieselbe Zusage wie die Abfrage:
    // eine Zeile mit leerem Wert sagte, das Feld sei leer.
    Kartenfeldwerte: (_setID: string, feld: string) =>
      Promise.resolve({
        werte: zeilen.flatMap((zeile) =>
          werte(zeile)
            .filter((wert) => wert.feld === feld)
            .map((wert) => ({ id: zeile.id, wert: wert.wert })),
        ),
      }),
  }
}

// ausKarten leitet die drei Wege aus **einem** `Karten`-Beleg ab (#891).
//
// **Damit ein Test weiter Karten beschreibt und nicht drei Antworten.** Bis
// zum 2026-08-25 belegte jede Prüfung der Kartenseite `Karten` und bekam von
// dort Zeilen **und** Feldwerte; seitdem sind es drei Bindungen. Sie einzeln
// nachzubauen hieße, in jeder Datei dieselbe Projektion zu wiederholen — und
// die läuft auseinander, sobald eine Datei sie anders schreibt.
//
// **Der Beleg wird je Ladevorgang genau einmal gefragt.** `Karte` und
// `Kartenfeldwerte` greifen auf die Antwort zu, die `Kartenzeilen` schon
// geholt hat, statt neu zu fragen. Ohne das zählte ein
// `mockResolvedValueOnce` anders als im Betrieb: die zweite Antwort wäre bei
// der ersten Anzeige verbraucht, und der Test prüfte einen Stand, den es nie
// gab — genau die Vorrichtung, die großzügiger ist als der Betrieb (§3).
//
// **Der Setbezug kommt aus der zuletzt geholten Liste.** `Karte` bekommt nur
// eine Kartenkennung; ohne diesen Merker wüsste die Probe nicht, welches Set
// gemeint ist. Solange keine Liste geholt wurde, gilt `vorgabe`.
export function ausKarten(
  karten: (setID: string) => Promise<{ karten: KarteDTO[] }>,
  vorgabe = '1',
) {
  let zuletzt = vorgabe
  let offen: Promise<KarteDTO[]> | undefined

  const frisch = async (setID: string) => (await karten(setID)).karten

  // gehalten gibt die Liste des letzten Ladevorgangs — oder holt sie, wenn
  // noch keiner lief.
  const gehalten = (setID: string) => (offen ??= frisch(setID))

  return {
    Kartenzeilen: async (setID: string) => {
      zuletzt = setID
      offen = frisch(setID)

      return {
        zeilen: (await offen).map((karte) => ({
          id: karte.id,
          nummer: karte.nummer,
          name: karte.werte.find((wert) => wert.feld === 'name')?.wert ?? '',
          felder: karte.werte.length,
        })),
      }
    },

    Karte: async (karteID: string) => ({
      karte: (await gehalten(zuletzt)).find(
        (karte) => karte.id === karteID,
      ) ?? {
        id: '',
        nummer: 0,
        werte: [],
      },
    }),

    Kartenfeldwerte: async (setID: string, feld: string) => ({
      werte: (await gehalten(setID)).flatMap((karte) =>
        karte.werte
          .filter((wert) => wert.feld === feld)
          .map((wert) => ({ id: karte.id, wert: wert.wert })),
      ),
    }),

    // **Und der Verweis auf einen echten Druck** (#1167).
    //
    // # Warum er hierher gehört und nicht in jede Datei einzeln
    //
    // Weil die Kartenlasche ihn **beim Aufschlagen jeder Karte** holt — er
    // ist ein vierter Weg desselben Ladevorgangs. Ihn in jeder Prüfung
    // nachzubauen hieße, dieselbe Antwort an zwanzig Stellen zu wiederholen.
    //
    // **Keiner ist die Vorgabe, und das ist der Betrieb**: die allermeisten
    // Karten sind auf keinen Druck bezogen. Eine Vorrichtung, die hier einen
    // erfände, wäre großzügiger als der Betrieb (§3). Wer einen braucht,
    // belegt `KarteVerweis` daneben — der spätere Beleg gewinnt.
    KarteVerweis: () =>
      Promise.resolve({ verweis: null, befunde: [], fehler: null }),

    // **Und die Gestalt der Liste** (#1168) — aus demselben Grund: die
    // Kartenlasche holt sie bei jedem Aufbau.
    //
    // **Die Spalten kommen aus den Karten selbst**, damit die Vorrichtung
    // nicht großzügiger ist als der Betrieb: was hier als Spalte steht, steht
    // auch in den Feldwerten. Sichtbar ist nur `name` — der Vorschlag eines
    // Spielpakets, bevor jemand wählt.
    Kartenlistenaufbau: async (setID: string) => {
      const liste = await gehalten(setID)
      const felder = [
        ...new Set(
          liste.flatMap((karte) => karte.werte.map((wert) => wert.feld)),
        ),
      ]

      return {
        spalten: felder.map((feld) => ({
          feld,
          aufschrift: '',
          sichtbar: feld === 'name',
          ausrichtung: 'links',
        })),
        zeilen: liste.map((karte) => ({
          karteID: karte.id,
          nummer: karte.nummer + 1,
          werte: Object.fromEntries(
            karte.werte.map((wert) => [wert.feld, wert.wert]),
          ),

          // **Ohne Marken** — wer sie prüft, belegt `Marken` und
          // `KartenMarken` selbst
          // (#1208).
          marken: [],
        })),
        sortierfeld: '',
        aufsteigend: true,
        ohneVorlage: false,
        fehler: null,
      }
    },
  }
}
