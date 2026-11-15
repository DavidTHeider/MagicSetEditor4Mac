// **A trade in four phases** — TRADE-02
// (#1716).
// Maschinell übersetzt.
export const handelsassistent = {
  schrittleiste: 'En qué punto está esta operación',
  phasen: {
    gegenueber: 'Con quién',
    angebot: 'Qué intercambiáis',
    zusagen: 'Reservas',
    verbuchen: 'Registro',
  },
  erledigt: 'hecho',
  jetzt: 'ahora',
  kommt: 'por venir',
  schritte: 'Lo que puedes hacer ahora',
  keinSchritt: 'De momento no tienes nada que hacer aquí.',
  neu: 'Un intercambio nuevo: reúne lo que va y viene y ábrelo.',
  abgleich: 'Lo que podríais intercambiar',
  unterwegs: 'Lo que está en camino',

  // deinGegenueber setzt sich in „Wartet auf {name}" ein — der Raum
  // gehört ohnehin einem Gegenüber (TRADE-02).
  deinGegenueber: 'la otra parte',

  // Das Protokoll eines Handels (TRADE-03.AK-03).
  verlauf: 'Historial',
  verlaufWechsel: 'Revisión {revision}: {von} → {nach} · {zeit}',
  verlaufNeu: 'Revisión {revision}: {nach} · {zeit}',
  verlaufLeer: 'Aún no hay ningún cambio.',
  verlaufOhneHandel:
    'Elige un intercambio — aquí verás cómo llegó hasta ese punto.',

  // Die zwei Spalten eines Tauschs mit Preisen (TRADE-02b, #1951).
  zeileIn: 'Añadir línea: {seite}',
  zeileKurz: '+ Línea',
  keinPreis: 'sin precio',
  differenzDir: '⇆ {betrag} a tu favor',
  differenzGegenueber: '⇆ {betrag} a favor de tu interlocutor',
  ausgeglichen: '⇆ Equilibrado',
  differenzTeil: '{zahl} sin precio — solo se cuenta lo que lo tiene.',
  preisHinweis:
    'Precios del {datum}, de {quelle}. Son orientativos, no vinculantes.',
  keinePreise: 'Todavía no hay precios para estas cartas.',
  zeileVerwerfen: 'Cancelar',
}
