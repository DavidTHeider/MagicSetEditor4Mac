// Die Texte der **Werkzeugseiten** — Konsole und Schlüsselwörter.
//
// Aus `seiten.ts` herausgelöst, als die Datei über 400 Zeilen wuchs (§2).
// **Ein Katalog darf sich auf mehrere Dateien verteilen**, solange er als
// **einer** ankommt — an der Zahl der Kataloge hängt der Vertragstest gegen
// die Go-Seite, nicht an der Zahl der Dateien.

export const werkzeugseiten = {
  pack: {
    titel: 'Sobre aleatorio',
    erklaerung:
      'Un sobre con las cartas de este set, según las reglas del paquete de juego.',
  },
  symbol: {
    titel: 'Símbolo',
    erklaerung: 'El símbolo de este set: dibujado, no cargado.',
  },
  konsole: {
    titel: 'Consola',
    erklaerung:
      'Una expresión de MSE Script en el contexto de la carta abierta.',
  },
  schluesselwoerter: {
    titel: 'Palabras clave',
    erklaerung:
      'Las palabras clave de este set, y las que trae consigo su juego.',
  },
}
