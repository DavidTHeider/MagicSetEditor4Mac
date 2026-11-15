// **The account kinds on the user page** — piece 6 of phase 4.12 (#1351).
export const konto = {
  eigenes: '你自己的账户',
  eigenesErklaerung:
    '这里的地址或电话号码是一种标识：它说明是谁在登录。邮箱或号码是否真实存在，程序不会检查——它不发送任何东西，也不需要任何第三方服务。',

  kennung: '地址或电话号码',
  kennungHinweis: '例如 name@example.org 或 +4915112345678',

  kennwort: '密码',
  kennwortHinweis: '至少十二个字符',
  kennwortAnmelden: '密码',

  anzeigename: '显示名称（可选）',
  anzeigenameHinweis: '留空时，将根据标识生成',

  eintragen: '创建账户',
  eingetragen: '已创建：{name}',

  anmelden: '登录',
  angemeldet: '已登录为 {name}',

  ohneAngaben: '不填写任何信息',
  ohneAngabenErklaerung:
    '匿名账户会得到一个随机标识作为名称，没有密码，也没有地址。如果你根本不想要账户，也不需要：首次启动时的用户就在那里，应用程序照常工作。',
  anonym: '创建匿名账户',
  anonymAngelegt: '已创建：{name}',

  kennwortNieZurueck:
    '密码以受保护的形式存储，不会再次显示。忘记密码的人应设置一个新密码，而不是去读取旧密码。',
} as const

// **The second factor** — piece 7 of phase 4.12 (#1351).
export const faktor = {
  titel: '第二重验证',
  erklaerung:
    '登录和切换档案时需要第二重验证。验证器应用中的验证码在任何地方都能用；Touch ID 只能在这台电脑上使用。',

  totp: '应用验证码',
  biometrie: '生物识别',

  einrichten: '设置',
  entfernen: '移除',

  totpEinrichten: '设置应用验证码',
  totpEntfernen: '移除应用验证码',
  biometrieEinrichten: '设置 Touch ID 或 Face ID',
  biometrieEntfernen: '移除 Touch ID 或 Face ID',

  totpEntfernenFrage: '移除应用验证码？',
  biometrieEntfernenFrage: '移除 Touch ID 或 Face ID？',
  entfernenErklaerung:
    '登录和切换档案将不再要求它。想重新启用的人需要重新设置——使用带新密钥的应用。',
  entfernenJa: '移除',

  einmal: '这些内容只显示这一次',
  einmalErklaerung:
    '请将此密钥输入你的验证器应用，并妥善保管恢复码。之后两者都无法再次查看——丢失后需要重新设置。',
  codesUeberschrift: '{zahl} 个恢复码，每个只能使用一次：',
  bestaetigen: '请输入应用中的一个验证码。此后第二重验证才会生效。',
} as const
