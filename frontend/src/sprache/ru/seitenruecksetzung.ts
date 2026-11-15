// The **staged reset** — work item O4 of phase 5.2
// (#1428).

export const ruecksetzung = {
  titel: 'Сброс',
  hinweis:
    'Шесть видов — от самого лёгкого до самого тяжёлого. Что затрагивает ' +
    'каждый, показано заранее — с числами и названиями.',

  artWaehlen: 'Что нужно сбросить?',

  arten: {
    layout: 'Только внешний вид окна',
    profileinstellungen: 'Настройки этого профиля',
    nutzereinstellungen: 'Мои собственные настройки',
    profil: 'Этот профиль и всё, что ему принадлежит',
    nutzer: 'Я и все мои профили',
    anwendung: 'Всё приложение',
  } as Record<string, string>,

  nachsehen: 'Посмотреть, что будет затронуто',
  nachsehenLaeuft: 'Проверка …',

  nichtsBetroffen: 'Здесь нечего сбрасывать.',
  gruppe: '{was}: {anzahl}',

  gruppen: {
    einstellungen: 'Настройки',
    profile: 'Профили',
    sets: 'Наборы',
    karten: 'Карты',
    decks: 'Колоды',
    exemplare: 'Экземпляры в коллекции',
    token: 'Токены доступа',
    freigaben: 'Записи о разрешениях',
    zugaenge: 'Доступы к профилям',
    bilder: 'Изображения',
  } as Record<string, string>,

  sicherung:
    'Здесь данные будут потеряны. Сначала создай резервную копию базы данных — в разделе выше.',

  hindernisse: {
    offeneRueckkehr:
      'Количества в этом профиле не урегулированы после отката. Проверь их ' +
      'и подтверди откат, прежде чем удалять.',
    fremdesZiel: 'Это не твоё.',
    letzterNutzer:
      'Это последний человек в этой базе данных — если удалить его, ' +
      'не останется никого.',
    laufenderAuftrag: 'Что-то ещё выполняется. Дождись завершения.',
    nachweisFehlt:
      'Для сброса всего приложения нужно явное разрешение на этом ' +
      'компьютере.',
  } as Record<string, string>,

  nameTippen: 'Введи имя для подтверждения',
  freigabeSatz:
    'Для сброса всего приложения нужно твоё разрешение на этом компьютере. Оно действует пять минут и на один новый старт.',
  freigeben: 'Выдать разрешение',
  freigebenLaeuft: 'Выдаём разрешение …',
  freigegeben: 'Разрешено до {zeit}.',
  kennungFuerKommandozeile:
    'То же разрешение действует в командной строке (mse zuruecksetzen anwendung):',

  tun: 'Сбросить',
  tunLaeuft: 'Сброс …',

  bilanz: 'Возвращено настроек: {anzahl}.',
  beiseite: 'Прежняя база данных сохранена здесь: {ort}.',
}
