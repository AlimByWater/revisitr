export interface UpdateProfileRequest {
  name: string
}

export interface ChangeEmailRequest {
  new_email: string
}

export interface ChangePhoneRequest {
  new_phone: string
}

export interface ChangePasswordRequest {
  current_password: string
  new_password: string
}

export type LegalEntityType = 'none' | 'self_employed' | 'ip' | 'ooo' | 'other'

export interface BillingDetails {
  entity_type: LegalEntityType
  fields: Record<string, string>
}

export const ENTITY_TYPE_LABELS: Record<LegalEntityType, string> = {
  none: 'Не указано',
  self_employed: 'Самозанятый',
  ip: 'ИП',
  ooo: 'ООО',
  other: 'Другое юр. лицо',
}

export interface EntityFieldDef {
  key: string
  label: string
}

export const ENTITY_FIELDS: Record<Exclude<LegalEntityType, 'none'>, EntityFieldDef[]> = {
  self_employed: [
    { key: 'full_name', label: 'ФИО' },
    { key: 'inn', label: 'ИНН' },
    { key: 'address', label: 'Адрес' },
    { key: 'email', label: 'Email' },
    { key: 'phone', label: 'Телефон' },
  ],
  ip: [
    { key: 'full_name', label: 'ФИО ИП' },
    { key: 'inn', label: 'ИНН' },
    { key: 'ogrnip', label: 'ОГРНИП' },
    { key: 'legal_address', label: 'Адрес регистрации' },
    { key: 'postal_address', label: 'Почтовый адрес' },
    { key: 'email', label: 'Email' },
    { key: 'phone', label: 'Телефон' },
  ],
  ooo: [
    { key: 'full_name', label: 'Полное наименование' },
    { key: 'short_name', label: 'Сокращённое наименование' },
    { key: 'inn', label: 'ИНН' },
    { key: 'kpp', label: 'КПП' },
    { key: 'ogrn', label: 'ОГРН' },
    { key: 'legal_address', label: 'Юридический адрес' },
    { key: 'postal_address', label: 'Почтовый адрес' },
    { key: 'director_name', label: 'ФИО руководителя' },
    { key: 'email', label: 'Email' },
    { key: 'phone', label: 'Телефон' },
  ],
  other: [
    { key: 'name', label: 'Наименование' },
    { key: 'inn', label: 'ИНН' },
    { key: 'kpp', label: 'КПП' },
    { key: 'ogrn', label: 'ОГРН' },
    { key: 'legal_address', label: 'Юридический адрес' },
    { key: 'postal_address', label: 'Почтовый адрес' },
    { key: 'contact_name', label: 'ФИО контактного лица' },
    { key: 'email', label: 'Email' },
    { key: 'phone', label: 'Телефон' },
  ],
}
