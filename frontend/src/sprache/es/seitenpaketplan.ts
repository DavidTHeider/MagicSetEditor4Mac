// The texts of the **package inspector and the plan before a package
// change** — DOM-01 (#1696). Their German source is `de/seitenpaketplan.ts`.

export const paketplan = {
  arbeitsplatz: 'Espacio de trabajo de paquetes',
  navigator: 'Búsqueda y filtros',
  inspector: 'Paquete',
  zentrum: 'Paquetes',

  herkunft: 'Origen',
  herkunftAlle: 'Cualquier origen',
  herkunftsarten: {
    mitgeliefert: 'incluido',
    bezogen: 'descargado',
    uebernommen: 'importado',
    eigen: 'propio',
  },

  spalteStand: 'Estado',
  standEingebunden: 'instalado',
  standGeaendert: 'modificado localmente',

  inspectorLeer:
    'Elige un paquete de la lista: aquí aparecerán su origen, su versión, sus dependencias y lo que provocaría un cambio.',
  kopf: '{art} · versión {version}',
  kopfOhneFassung: '{art} · sin versión',
  eingebunden: 'Instalado',
  keinSet: 'Ningún set de este perfil lo usa.',
  einSet: 'Un set lo usa.',
  vieleSets: '{anzahl} sets lo usan.',
  setsUnbekannt: 'Qué sets lo necesitan no figura en ninguna cabecera de set.',
  geaendert:
    'Modificado localmente desde que se instaló. Al reemplazarlo se pregunta antes; el cambio se conserva como versión anterior.',
  woher: 'Procedencia',
  woherMitgeliefert: 'Incluido con la aplicación.',
  woherBezogen: 'Descargado de {herkunft}.',
  woherUebernommen: 'Importado de {herkunft}.',
  woherEigen: 'Creado aquí o colocado a mano en la carpeta de paquetes.',
  pruefsumme: 'Suma de comprobación',
  lizenz: 'Licencia',
  lizenzUnbekannt:
    'sin indicar: un paquete de MSE no lleva campo de licencia; los contenidos pertenecen a sus titulares de derechos',
  ort: 'Ubicación',
  haengtAb: 'Depende de',
  gebrauchtVon: 'Lo necesitan',
  nichts: 'nada',

  einbinden: 'Instalar paquete …',
  beziehen: 'Obtener un paquete …',
  beziehenTitel: 'Obtener un paquete',
  loeschen: 'Eliminar …',
  zurueck: 'Recuperar la versión anterior …',
  weitergeben: 'Compartir como instalador …',
  weitergegeben: 'Guardado como instalador: {datei}',

  titel: {
    loeschen: 'Eliminar {paket}',
    einbinden: 'Instalar {paket}',
    ersetzen: 'Reemplazar {paket}',
    zurueck: 'Restablecer {paket} a su versión anterior',
  },
  fassung: 'Versión {von} → {zu}',
  ohneFassung: 'sin versión',
  sets: 'Sets afectados',
  keineSets: 'Ningún set de este perfil resulta afectado.',
  konflikte: 'Dejarían de cumplirse después',
  keineKonflikte: 'No se rompe ninguna dependencia.',
  planGeaendert: 'El paquete se ha modificado localmente desde que se instaló.',
  ueberschreiben:
    'Reemplazar el cambio local: se conserva como versión anterior',
  nichtMoeglich: {
    loeschen:
      'Este paquete no está en la carpeta administrada. Aquí solo se elimina lo que instaló esta aplicación.',
    einbinden: 'Este paquete no se puede instalar aquí.',
    ersetzen: 'Este paquete no se puede reemplazar aquí.',
    zurueck:
      'No hay versión anterior. Se crea una cuando instalas una versión más reciente sobre este paquete.',
  },
  bestaetigen: {
    loeschen: 'Eliminar',
    einbinden: 'Instalar',
    ersetzen: 'Reemplazar',
    zurueck: 'Restablecer',
  },
  verwerfen: 'No cambiar nada',
  bedarf: {
    ueberschrift: 'Paquetes que le faltan a este set',
    erklaerung:
      'Sin ellos, la vista previa dibuja un plan de emergencia u omite lo que aporta el paquete. Instala el paquete desde una carpeta, o búscalo entre los paquetes para descargarlo o importarlo desde MSE2.',
    suchen: 'Buscar entre los paquetes',
  },
  quelle: 'De {quelle}',
  keineRechte:
    'Un paquete trae plantillas y MSE-Script. El script no tiene acceso a archivos, a la red ni a programas, y lo que el paquete necesita no se descarga junto con él: aparece arriba como conflicto.',
}
