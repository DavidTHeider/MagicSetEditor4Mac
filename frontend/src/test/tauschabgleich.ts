// **Ein Abgleich, wie ihn der Markttisch beim Aufschlagen holt** — #2494:
// der Tisch gleicht die geholte Sicht ab, ohne dass jemand klickt. Wer einen
// Markttisch zeigt, belegt deshalb auch `Tauschabgleich`.
import type {
  Tauschabgleichantwort,
  Tauschkandidat,
} from '../apitauschabgleich'

const leereSeite = {
  zeilen: [],
  stueck: 0,
  gedeckt: 0,
  offen: 0,
  wert: 0,
  waehrung: '',
  bepreist: false,
}

// einAbgleich sagt, was du bekommen könntest — ohne Vorschlag, denn drüben
// sucht niemand etwas von dir.
export function einAbgleich(
  bekomme: Tauschkandidat[] = [],
): Tauschabgleichantwort {
  return {
    kennung: 'du-1',
    gebe: [],
    bekomme,
    vorschlag: { gebe: leereSeite, bekomme: leereSeite },
    hatVorschlag: false,
    befunde: [],
    fehler: null,
  }
}
