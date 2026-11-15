// **Der Satz zu einem Zweck** — aus `Peerkontaktzeile.tsx` herausgelöst, als
// der Einladungstausch (NET-01, #1705) denselben brauchte: zwei Tafeln für
// dieselben vier Zwecke liefen beim fünften auseinander.

// zwecksatz gibt den Satz zu einem Zweck.
export function zwecksatz(
  eigene: {
    zweckPaarung: string
    zweckTeilen: string
    zweckHandel: string
    zweckPartie: string
  },
  zweck: string,
): string {
  const saetze: Record<string, string> = {
    paarung: eigene.zweckPaarung,
    teilen: eigene.zweckTeilen,
    handel: eigene.zweckHandel,
    partie: eigene.zweckPartie,
  }

  return saetze[zweck] ?? zweck
}
