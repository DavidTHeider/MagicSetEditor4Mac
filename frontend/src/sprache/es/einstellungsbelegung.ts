// The mapping of ⌘0 to ⌘9 — UX-02.AK-03 (#1674), split out like the German file.
export const einstellungsbelegung = {
  belegungUeberschrift: 'Asignación de ⌘0 a ⌘9',
  belegungHinweis:
    'Qué página abre un dígito con ⌘. Por defecto son las áreas de la barra lateral; «Como antes de 0.6.0» recupera la asignación hasta la 0.5.0, cuya tabla figura en el registro de cambios.',
  belegungFuer: 'Página para {kuerzel}',
  belegungFrei: '—',
  belegungVor06: 'Como antes de 0.6.0',
  belegungVorgabe: 'Volver al valor por defecto',
  kuerzel: {
    palette: 'Abrir la paleta de comandos',
    suche: 'Buscar en el set',
    zurueck: 'Deshacer un cambio',
    vor: 'Rehacer un cambio',
    sichern: 'Guardar el set',
    sichernUnter: 'Guardar el set como …',
    oeffnen: 'Abrir un set …',
    neu: 'Crear un set nuevo',
    schliessen: 'Cerrar el set',
    drucken: 'Ir a la hoja de impresión',
    navigator: 'Mostrar u ocultar el navegador',
    inspector: 'Mostrar u ocultar el inspector',

    // Back and forward between pages — UX-03.AK-02 (#1675).
    seiteZurueck: 'Página anterior',
    seiteVor: 'Página siguiente',
  },
} as const
