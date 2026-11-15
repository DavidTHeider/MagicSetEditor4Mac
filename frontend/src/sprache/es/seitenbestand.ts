// The texts for **database maintenance** — back up and restore
// (#940).
//
// Split out of `seiten.ts` the same way the card page was: one catalogue may
// live in several files as long as it arrives as **one** (§9).

export const bestandspflege = {
  titel: 'Base de datos',
  hinweis:
    'Todo lo que guarda esta aplicación: sets, cartas, valores de campo, el ' +
    'historial y las imágenes insertadas. Una copia de seguridad lleva ambas ' +
    'cosas en un solo archivo: la base de datos y las imágenes.',
  sichern: 'Hacer copia de seguridad de la base de datos …',
  waehlen: 'Restaurar una copia de seguridad …',
  rueckfrage: '¿Restaurar esta copia de seguridad?',
  datei: 'Archivo',
  angelegtAm: 'Base de datos creada',
  sets: 'Sets',
  karten: 'Cartas',
  unbekannt: '—',
  warnung:
    'La base de datos actual se sustituirá. No se pierde: queda al lado como ' +
    'bestand-vor-rueckspielung.db, hasta la próxima restauración.',
  bilderLage: 'Imágenes insertadas',
  bilderErsetzt:
    'Esta copia de seguridad incluye imágenes. Ocupan el lugar de las ' +
    'actuales; las anteriores quedan al lado como beiwerk-vorher.',
  bilderGeleert:
    'Esta copia de seguridad no incluye imágenes: procede de un estado en el ' +
    'que no había ninguna insertada. Las actuales se eliminan; quedan al lado ' +
    'como beiwerk-vorher.',
  bilderUnberuehrt:
    'Esta copia de seguridad es una única base de datos de la versión 0.1.0 o ' +
    '0.2.0 y no incluye imágenes. Las actuales se quedan sin tocar; en otro ' +
    'ordenador faltarían por completo.',
  bestaetigen: 'Restaurar',
  abbrechen: 'Cancelar',

  // **The password of a backup** (OPS-06, #1723).
  kennwort: 'Contraseña (opcional)',
  kennwortWiederholen: 'Repetir contraseña',
  kennwortHinweis:
    'Con contraseña, la copia de seguridad queda cifrada y también se puede abrir en otro ordenador. Una contraseña olvidada no se puede recuperar: la copia de seguridad se pierde.',
  kennwortVerschieden: 'Las dos contraseñas no coinciden.',
  klartext:
    'Hacer la copia sin contraseña: el archivo queda entonces en texto sin cifrar',
  klartextHinweis:
    'La biblioteca está cifrada. Sin contraseña, la copia de seguridad quedaría desprotegida en el soporte al que se guarda.',
  gesperrt: 'Esta copia de seguridad está cifrada.',
  gesperrtKennwort: 'Contraseña de la copia de seguridad',
  entsperren: 'Abrir',
  verschluesselt: 'Cifrada',
  ja: 'sí',
  nein: 'no',
  paketeEnthalten: 'Paquetes propios incluidos',
  paketeFehlen: 'Por volver a descargar',
  keine: 'ninguno',
  zurueckgespielt:
    'Restaurada. La base de datos anterior está ahora en {beiseite}.',
  rueckkehrUnveraendert: 'Tu base de datos anterior no ha cambiado.',
  rueckkehrUnvollstaendig:
    'También falló la vuelta atrás: partes de la base de datos anterior están al lado como bestand-vor-rueckspielung.db y beiwerk-vorher. Restaura una copia de seguridad o devuélvelas a mano.',
  wiederherstellungTitel: 'No se pudo abrir la base de datos',
  wiederherstellungHinweis:
    'La aplicación volverá a funcionar cuando haya un archivo de base de ' +
    'datos utilizable. Restaurar una copia de seguridad es el camino: ' +
    'sustituye el archivo que está fallando ahora.',
  wiederherstellungGrund: 'Lo que se ha comunicado',
  wiederherstellungHinweisHeil:
    'El archivo de la biblioteca probablemente está bien: el problema está en algo a su alrededor. Corrígelo como se describe abajo y vuelve a intentarlo; no hay nada que sustituir.',
  erneutVersuchen: 'Volver a intentarlo',
  weiterGescheitert:
    'Se volvió a intentar: la biblioteca sigue sin poder abrirse.',

  wegZurueckTitel: 'El camino de vuelta',
  wegZurueck:
    'Para volver atrás, usa «{datei}» en la carpeta de tu biblioteca: ' +
    '{sets} set(s), {karten} cartas, versión del esquema {stand}. Elígela abajo.',

  chiffreTitel: 'Cifrado de la biblioteca',
  chiffreAnHinweis:
    'La biblioteca se guarda cifrada en el disco. Sin la clave no se puede ' +
    'abrir, ni siquiera tú.',
  chiffreAusHinweis:
    'La biblioteca se guarda sin cifrar. Cualquiera que tenga el archivo ' +
    'puede leer todo lo que contiene: nombres, fechas de nacimiento, ' +
    'direcciones, fotos y métodos de inicio de sesión.',
  chiffreAnschalten: 'Activar el cifrado …',
  chiffreAusschalten: 'Desactivar el cifrado …',
  chiffreFrageAn: '¿Cifrar ahora la biblioteca?',
  chiffreFrageAus: '¿Desactivar de nuevo el cifrado?',
  chiffreWarnungAn:
    'La clave se crea ahora y va al llavero de este Mac. No se puede ' +
    'volver a calcular y no existe en ningún otro sitio: si se pierde, se ' +
    'pierden los datos. Haz antes una copia de seguridad.',
  chiffreWarnungAus:
    'Todo volverá a guardarse sin cifrar y la clave se eliminará del ' +
    'llavero.',
  chiffreGrenzen:
    'El cifrado **no** protege contra la manipulación. Quien compare ' +
    'varias copias del mismo archivo puede ver qué partes han cambiado. ' +
    '**Una copia de seguridad tiene su propia contraseña**: arriba, al hacer ' +
    'la copia; sin ella quedaría en texto sin cifrar, y eso exige entonces ' +
    'un «hacer la copia sin contraseña» explícito.',
  chiffreLaeuft: 'Reescribiendo la biblioteca …',

  platzUeberschrift: 'Espacio en disco',
  platzErklaerung:
    'Lo que guarda esta aplicación y cuánto ocupa: la carpeta entera, porque lo que no pertenece a ningún concepto se cuenta en **Otros**. Las **copias de seguridad previas a las migraciones** se hacen en cada inicio que reconstruye la biblioteca; lo que se conserva es el **camino de vuelta**: la copia de seguridad más reciente que esta versión puede abrir. Del **registro** se conserva el archivo en curso.',
  platzDateien: '{anzahl} archivo(s)',
  platzFreigeben: 'Borrar las anteriores',
  platzFreigebenFuer: '{posten}: borrar las anteriores',
  platzFreigebenFrage: '{posten}: ¿borrar de verdad las anteriores?',
  platzFreigebenErklaerung:
    'Los archivos anteriores se eliminan y no se pueden recuperar. Lo que se necesita como camino de vuelta se conserva.',
  platzLoeschenJa: 'Borrar',
  platzNamen: {
    bestand: 'Biblioteca',
    bilder: 'Imágenes insertadas',
    pakete: 'Paquetes descargados',
    vorsicherungen: 'Copias de seguridad previas a las migraciones',
    protokoll: 'Registro',
    sonstiges: 'Otros archivos de la carpeta de la aplicación',
  },
  alter: {
    ueberschrift: 'Última copia de seguridad',
    unbekannt: 'desconocida',
    unbekanntSatz:
      'Esta versión nunca ha guardado una copia de la biblioteca. Desconocida significa desconocida: aquí nunca pone «hoy» sin que se haya releído un archivo.',
    heute: 'hoy',
    gestern: 'hace un día',
    vorTagen: 'hace {tage} días',
    alt: '{alter} · más de 14 días',
    unstimmig: 'hora posterior al reloj',
    satz: 'Última copia el {zeitpunkt} desde {flaeche} en {datei}, releída y comprobada.',
    satzPlan:
      'Última copia el {zeitpunkt} según el plan en {datei}, releída y comprobada.',
    unstimmigSatz:
      'La última copia lleva la fecha {zeitpunkt}, posterior al reloj de este Mac. No se puede decir su antigüedad; haz una copia para asegurarte.',
    flaechen: {
      fenster: 'la ventana',
      cli: 'la línea de comandos',
      mcp: 'MCP',
      rest: 'el servicio',
    },
    planSchalter: 'Copia semanal en una carpeta, conservar siete',
    planHilfe:
      'Solo después de elegir una carpeta. Solo se borran sus propias copias, nunca otros archivos. Guarda mientras la ventana está abierta.',
    planOrdner: 'Carpeta: {ordner}',
    planZuletzt: 'Última copia según el plan el {zeitpunkt}.',
    planNie: 'El plan aún no ha hecho ninguna copia.',
    planFehler: 'El plan no pudo hacer la última copia: {grund}',
  },
} as const
