// The bulk-change preview — work package U6 of phase 5.1 (#1417). See the
// German file for why it lives apart from `seitenkarten.ts`.
export const massenvorschauEn = {
  vorschau: 'Vista previa',
  vorschauErklaerung:
    'La vista previa muestra, carta por carta, qué cambiaría. No se escribe nada hasta pulsar «Aplicar», y solo si nadie ha escrito mientras tanto.',
  karte: 'Carta',
  alt: 'Antes',
  neu: 'Después',
  keinWertBisher: 'sin valor',
  bleibtGleich: 'sin cambios: ya está así',
  zusammenfassung:
    'Se modificarán {geaendert} carta(s); {gleich} se quedan como están.',
  nichtsZuTun:
    'Todas las cartas seleccionadas ya tienen este valor. No hay nada que cambiar.',
  konflikt:
    'Estas cartas han cambiado desde la vista previa. No se ha escrito ninguna: vuelve a obtener la vista previa:',
  inzwischenGeaendert: 'modificada desde entonces',
  angewendet: '{anzahl} carta(s) modificada(s).',
} as const
