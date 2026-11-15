// The **staged reset** — work item O4 of phase 5.2
// (#1428).

export const ruecksetzung = {
  titel: 'Restablecer',
  hinweis:
    'Seis tipos, del más ligero al más pesado. Qué toca cada uno se muestra ' +
    'antes, con cifras y nombres.',

  artWaehlen: '¿Qué quieres restablecer?',

  arten: {
    layout: 'Solo el aspecto de la ventana',
    profileinstellungen: 'Los ajustes de este perfil',
    nutzereinstellungen: 'Mis propios ajustes',
    profil: 'Este perfil y todo lo que le pertenece',
    nutzer: 'Yo y todos mis perfiles',
    anwendung: 'Toda la aplicación',
  } as Record<string, string>,

  nachsehen: 'Ver qué se vería afectado',
  nachsehenLaeuft: 'Comprobando …',

  nichtsBetroffen: 'Aquí no hay nada que se pueda restablecer.',
  gruppe: '{was}: {anzahl}',

  gruppen: {
    einstellungen: 'Ajustes',
    profile: 'Perfiles',
    sets: 'Sets',
    karten: 'Cartas',
    decks: 'Mazos',
    exemplare: 'Ejemplares de la colección',
    token: 'Tokens de acceso',
    freigaben: 'Registros de autorización',
    zugaenge: 'Accesos a perfiles',
    bilder: 'Imágenes',
  } as Record<string, string>,

  sicherung:
    'Aquí se pierden datos. Haz antes una copia de seguridad de la base de datos, en el área de arriba.',

  hindernisse: {
    offeneRueckkehr:
      'Las cantidades de este perfil están sin aclarar tras una reversión. ' +
      'Revísalas y da el visto bueno a la reversión antes de eliminar.',
    fremdesZiel: 'Eso no te pertenece.',
    letzterNutzer:
      'Es la última persona de esta base de datos: eliminarla no dejaría a ' +
      'nadie.',
    laufenderAuftrag: 'Todavía hay algo en marcha. Espera a que termine.',
    nachweisFehlt:
      'Restablecer toda la aplicación requiere una autorización explícita en ' +
      'este equipo.',
  } as Record<string, string>,

  nameTippen: 'Escribe el nombre para confirmar',
  freigabeSatz:
    'Restablecer toda la aplicación requiere tu autorización en este ordenador. Vale cinco minutos y para un solo nuevo comienzo.',
  freigeben: 'Conceder autorización',
  freigebenLaeuft: 'Concediendo autorización …',
  freigegeben: 'Autorizado hasta {zeit}.',
  kennungFuerKommandozeile:
    'La misma autorización vale en la línea de órdenes (mse zuruecksetzen anwendung):',

  tun: 'Restablecer',
  tunLaeuft: 'Restableciendo …',

  bilanz: '{anzahl} ajustes revertidos.',
  beiseite: 'La base de datos anterior está en {ort}.',
}
