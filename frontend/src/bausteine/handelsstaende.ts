// **Die dreizehn Zustände eines Handels in Worten** — Werkstück T6 von Phase
// 5.6 (#1471),
// herausgezogen mit
// #1637.
//
// **Eine Stelle und nicht zwei.** Der Satz stand in `Handelsbuchung.tsx`, und
// seit es den Vorgangskasten daneben gibt, braucht ihn ein zweites Bauteil.
// Zwei Fassungen liefen beim vierzehnten Zustand auseinander — und die, die
// niemand ändert, ist die zweite.

// standsatz gibt den Satz zu einem Handelszustand.
//
// # Warum hier dreizehn stehen und nicht fünf
//
// **Weil die Idempotenz jeden hergibt.** Ein zweiter Ruf auf einen schon
// verbuchten Handel bringt den Zustand zurück, den er gerade hat — und das
// kann jeder der dreizehn aus der Matrix sein.
//
// **Der Rückfall ist der Zustand selbst**, dieselbe Überlegung wie bei
// `standsatzKontakt` in `Peerkontaktzeile.tsx`: ein vierzehnter, den dieses
// Fenster noch nicht kennt, ist immer noch eine Auskunft — und `lokalVerbucht`
// am Fenster zu zeigen war der Befund, für den es diesen Satz gibt
// (Sichtprüfung am 2026-09-18).
export function standsatz(eigene: Handelsstandsaetze, zustand: string): string {
  const saetze: Record<string, string> = {
    entwurf: eigene.buchungStandEntwurf,
    angeboten: eigene.buchungStandAngeboten,
    verhandlung: eigene.buchungStandVerhandlung,
    angenommen: eigene.buchungStandAngenommen,
    vorbereitet: eigene.buchungStandVorbereitet,
    uebergabeOffen: eigene.buchungStandUebergabeOffen,
    lokalVerbucht: eigene.buchungStandLokalVerbucht,
    abgleichOffen: eigene.buchungStandAbgleichOffen,
    beidseitigVerbucht: eigene.buchungStandBeidseitigVerbucht,
    abgelehnt: eigene.buchungStandAbgelehnt,
    abgebrochen: eigene.buchungStandAbgebrochen,
    abgelaufen: eigene.buchungStandAbgelaufen,
    klaerung: eigene.buchungStandKlaerung,
  }

  return saetze[zustand] ?? zustand
}

// Handelsstandsaetze sind die dreizehn Sätze aus dem Katalog.
interface Handelsstandsaetze {
  buchungStandEntwurf: string
  buchungStandAngeboten: string
  buchungStandVerhandlung: string
  buchungStandAngenommen: string
  buchungStandVorbereitet: string
  buchungStandUebergabeOffen: string
  buchungStandLokalVerbucht: string
  buchungStandAbgleichOffen: string
  buchungStandBeidseitigVerbucht: string
  buchungStandAbgelehnt: string
  buchungStandAbgebrochen: string
  buchungStandAbgelaufen: string
  buchungStandKlaerung: string
}

// artsatz gibt den Namen einer Handelsart (TRADE-03.AK-02, #1717). **Der
// Rückfall ist die Art selbst** — dieselbe Überlegung wie bei `standsatz`.
export function artsatz(
  eigene: { arten: Readonly<Record<string, string>> },
  art: string,
): string {
  return eigene.arten[art] ?? art
}
