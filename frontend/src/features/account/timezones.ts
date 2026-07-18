import type { SelectOption } from '@/components/common/CustomSelect'

// IANA-таймзоны РФ и СНГ. Значение уходит на backend как есть
// (валидируется через time.LoadLocation).
export const TIMEZONE_OPTIONS: SelectOption[] = [
  { value: 'Europe/Kaliningrad', label: 'Калининград (МСК−1)' },
  { value: 'Europe/Moscow', label: 'Москва (МСК)' },
  { value: 'Europe/Samara', label: 'Самара (МСК+1)' },
  { value: 'Asia/Yekaterinburg', label: 'Екатеринбург (МСК+2)' },
  { value: 'Asia/Omsk', label: 'Омск (МСК+3)' },
  { value: 'Asia/Krasnoyarsk', label: 'Красноярск (МСК+4)' },
  { value: 'Asia/Irkutsk', label: 'Иркутск (МСК+5)' },
  { value: 'Asia/Yakutsk', label: 'Якутск (МСК+6)' },
  { value: 'Asia/Vladivostok', label: 'Владивосток (МСК+7)' },
  { value: 'Asia/Magadan', label: 'Магадан (МСК+8)' },
  { value: 'Asia/Kamchatka', label: 'Камчатка (МСК+9)' },
  { value: 'Europe/Minsk', label: 'Минск (МСК)' },
  { value: 'Asia/Almaty', label: 'Алматы (МСК+2)' },
  { value: 'Asia/Tashkent', label: 'Ташкент (МСК+2)' },
  { value: 'Asia/Bishkek', label: 'Бишкек (МСК+3)' },
  { value: 'Asia/Yerevan', label: 'Ереван (МСК+1)' },
  { value: 'Asia/Baku', label: 'Баку (МСК+1)' },
  { value: 'Asia/Tbilisi', label: 'Тбилиси (МСК+1)' },
  { value: 'Asia/Dubai', label: 'Дубай (МСК+1)' },
]
