// **Die neun Gruppen der Einstellungsseite** — UX-04.AK-01
// (#1676).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge. Ein Katalog darf sich auf mehrere Dateien
// verteilen, solange er als **einer** ankommt (§9).
export const einstellungsgruppennamen = {
  gruppen: {
    allgemein: 'General',
    bereiche: 'Áreas y funciones',
    erscheinungsbild: 'Apariencia',
    nutzer: 'Usuario y perfil',
    fach: 'Funciones específicas',
    ki: 'IA',
    verbindungen: 'Conexiones',
    dienste: 'Servicios',
    daten: 'Datos y copias de seguridad',
    diagnose: 'Diagnóstico',
  },

  untertitel: {
    allgemein: 'Idioma, inicio, atajos',
    bereiche: 'Espacios de trabajo, funciones, madurez',
    erscheinungsbild: 'Tema, tipografía, ventana',
    nutzer: 'Inicio de sesión, desbloqueo',
    fach: 'Cartas, colección, partidas',
    ki: 'Proveedores, modelos, costes',
    verbindungen: 'Red, interlocutores, CLI, MCP, REST',
    dienste: 'Servicios de cartas, precios, mercado',
    daten: 'Almacén, rutas, copias de seguridad',
    diagnose: 'Registros, medición, técnica',
  },

  abschnitte: {
    fenster: 'Ventana y barra lateral',
    spalten: 'Columnas y vistas',
    handel: 'Límites del comercio',
    protokoll: 'Registro',
    messung: 'Medición',
    speicher: 'Cachés e índice de búsqueda',
    adressen: 'Direcciones',
    ordner: 'Carpetas y nombres de archivo',
    eingriffe: 'Volver atrás, limpiar, restablecer',
    mitnehmen: 'Llevarse los ajustes',
  },
} as const
