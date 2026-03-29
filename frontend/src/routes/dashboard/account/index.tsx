import { useState, useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { createPortal } from 'react-dom'
import { cn } from '@/lib/utils'
import { CustomSelect } from '@/components/common/CustomSelect'
import { useAuthStore } from '@/stores/auth'
import {
  useProfileQuery,
  useUpdateProfileMutation,
  useChangeEmailMutation,
  useChangePhoneMutation,
  useChangePasswordMutation,
  useBillingDetailsQuery,
  useUpdateBillingDetailsMutation,
} from '@/features/account/queries'
import {
  ENTITY_TYPE_LABELS,
  ENTITY_FIELDS,
  type LegalEntityType,
  type BillingDetails,
} from '@/features/account/types'
import { User, Shield, FileText, LogOut, Check, X, Pencil } from 'lucide-react'

const inputClassName = cn(
  'w-full px-4 py-2.5 rounded border border-neutral-900',
  'text-sm placeholder:text-neutral-400 bg-white',
  'focus:outline-none focus:ring-2 focus:ring-neutral-900/10',
  'transition-colors',
  'disabled:cursor-not-allowed disabled:bg-neutral-100 disabled:border-neutral-300 disabled:text-neutral-500',
)

// ---------------------------------------------------------------------------
// Profile Section
// ---------------------------------------------------------------------------

function ProfileSection() {
  const { data: profile, isLoading } = useProfileQuery()
  const updateProfile = useUpdateProfileMutation()
  const changeEmail = useChangeEmailMutation()
  const changePhone = useChangePhoneMutation()
  const hasRendered = useRef(false)

  const [name, setName] = useState('')
  const [nameChanged, setNameChanged] = useState(false)
  const [nameSaved, setNameSaved] = useState(false)

  const [editingEmail, setEditingEmail] = useState(false)
  const [newEmail, setNewEmail] = useState('')
  const [emailSent, setEmailSent] = useState(false)

  const [editingPhone, setEditingPhone] = useState(false)
  const [newPhone, setNewPhone] = useState('')
  const [phoneSent, setPhoneSent] = useState(false)

  useEffect(() => {
    if (profile) {
      setName(profile.name)
    }
  }, [profile])

  useEffect(() => {
    if (profile) {
      setNameChanged(name !== profile.name)
    }
  }, [name, profile])

  useEffect(() => {
    if (nameSaved) {
      const t = setTimeout(() => setNameSaved(false), 3000)
      return () => clearTimeout(t)
    }
  }, [nameSaved])

  useEffect(() => {
    if (emailSent) {
      const t = setTimeout(() => setEmailSent(false), 5000)
      return () => clearTimeout(t)
    }
  }, [emailSent])

  useEffect(() => {
    if (phoneSent) {
      const t = setTimeout(() => setPhoneSent(false), 5000)
      return () => clearTimeout(t)
    }
  }, [phoneSent])

  const handleSaveName = async () => {
    await updateProfile.mutate({ name })
    setNameSaved(true)
    setNameChanged(false)
  }

  const handleChangeEmail = async () => {
    if (!newEmail.trim()) return
    await changeEmail.mutate({ new_email: newEmail.trim() })
    setEditingEmail(false)
    setNewEmail('')
    setEmailSent(true)
  }

  const handleChangePhone = async () => {
    if (!newPhone.trim()) return
    await changePhone.mutate({ new_phone: newPhone.trim() })
    setEditingPhone(false)
    setNewPhone('')
    setPhoneSent(true)
  }

  if (isLoading && !hasRendered.current) {
    return <div className="border border-neutral-900 rounded bg-white p-6 h-48 animate-pulse" />
  }

  hasRendered.current = true

  return (
    <div className="border border-neutral-900 rounded bg-white p-6 animate-in">
      <div className="flex items-center gap-2 mb-5">
        <User className="w-5 h-5 text-neutral-900" />
        <h2 className="text-lg font-semibold text-neutral-900">Профиль</h2>
      </div>

      <div className="space-y-5">
        {/* Name — directly editable */}
        <div>
          <label className="block text-sm font-medium text-neutral-700 mb-1.5">Имя</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className={inputClassName}
            placeholder="Ваше имя"
          />
        </div>

        {/* Email — change flow */}
        <div>
          <label className="block text-sm font-medium text-neutral-700 mb-1.5">Почта</label>
          <div className="flex items-center gap-3">
            <input
              type="email"
              value={profile?.email ?? ''}
              disabled
              className={inputClassName}
            />
            {!editingEmail && (
              <button
                type="button"
                onClick={() => setEditingEmail(true)}
                className="shrink-0 text-sm text-[#EF3219] hover:text-[#FF5C47] font-medium transition-colors flex items-center gap-1"
              >
                <Pencil className="w-3.5 h-3.5" />
                Сменить
              </button>
            )}
          </div>
          {editingEmail && (
            <div className="mt-2 flex items-center gap-2">
              <input
                type="email"
                value={newEmail}
                onChange={(e) => setNewEmail(e.target.value)}
                placeholder="Новая почта"
                className={inputClassName}
                autoFocus
              />
              <button
                type="button"
                onClick={handleChangeEmail}
                disabled={changeEmail.isPending || !newEmail.trim()}
                className="shrink-0 px-4 py-2.5 rounded bg-neutral-900 text-white text-sm font-medium hover:bg-neutral-700 active:bg-[#EF3219] transition-colors duration-300 focus:outline-none disabled:opacity-50"
              >
                {changeEmail.isPending ? '...' : 'Сохранить'}
              </button>
              <button
                type="button"
                onClick={() => { setEditingEmail(false); setNewEmail('') }}
                className="shrink-0 text-sm text-neutral-400 hover:text-neutral-600 transition-colors"
              >
                Отмена
              </button>
            </div>
          )}
          {emailSent && (
            <p className="mt-2 text-sm text-green-600">Ссылка для подтверждения отправлена на новую почту.</p>
          )}
        </div>

        {/* Phone — change flow */}
        <div>
          <label className="block text-sm font-medium text-neutral-700 mb-1.5">Телефон</label>
          <div className="flex items-center gap-3">
            <input
              type="tel"
              value={profile?.phone ?? '—'}
              disabled
              className={inputClassName}
            />
            {!editingPhone && (
              <button
                type="button"
                onClick={() => setEditingPhone(true)}
                className="shrink-0 text-sm text-[#EF3219] hover:text-[#FF5C47] font-medium transition-colors flex items-center gap-1"
              >
                <Pencil className="w-3.5 h-3.5" />
                Сменить
              </button>
            )}
          </div>
          {editingPhone && (
            <div className="mt-2 flex items-center gap-2">
              <input
                type="tel"
                value={newPhone}
                onChange={(e) => setNewPhone(e.target.value)}
                placeholder="Новый номер телефона"
                className={inputClassName}
                autoFocus
              />
              <button
                type="button"
                onClick={handleChangePhone}
                disabled={changePhone.isPending || !newPhone.trim()}
                className="shrink-0 px-4 py-2.5 rounded bg-neutral-900 text-white text-sm font-medium hover:bg-neutral-700 active:bg-[#EF3219] transition-colors duration-300 focus:outline-none disabled:opacity-50"
              >
                {changePhone.isPending ? '...' : 'Сохранить'}
              </button>
              <button
                type="button"
                onClick={() => { setEditingPhone(false); setNewPhone('') }}
                className="shrink-0 text-sm text-neutral-400 hover:text-neutral-600 transition-colors"
              >
                Отмена
              </button>
            </div>
          )}
          {phoneSent && (
            <p className="mt-2 text-sm text-green-600">Код подтверждения отправлен на новый номер.</p>
          )}
        </div>
      </div>

      {/* Save name */}
      <div className="mt-6 pt-5 border-t border-neutral-200 flex items-center gap-3">
        <button
          type="button"
          onClick={handleSaveName}
          disabled={!nameChanged || updateProfile.isPending}
          className={cn(
            'px-5 py-2.5 rounded text-sm font-semibold',
            'bg-neutral-900 text-white hover:bg-neutral-700 active:bg-[#EF3219] transition-colors duration-300 focus:outline-none',
            'disabled:bg-neutral-300 disabled:cursor-not-allowed disabled:hover:bg-neutral-300 disabled:active:bg-neutral-300',
          )}
        >
          {updateProfile.isPending ? 'Сохранение...' : 'Сохранить изменения'}
        </button>
        {nameSaved && (
          <div className="flex items-center gap-1.5 text-sm text-green-600 font-medium">
            <Check className="w-4 h-4" />
            Сохранено
          </div>
        )}
        {updateProfile.isError && (
          <p className="text-sm text-red-500">Не удалось сохранить.</p>
        )}
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Security Section
// ---------------------------------------------------------------------------

function SecuritySection() {
  const changePassword = useChangePasswordMutation()
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [success, setSuccess] = useState(false)

  const allFilled = currentPassword.length > 0 && newPassword.length > 0 && confirmPassword.length > 0
  const passwordsMatch = newPassword === confirmPassword

  useEffect(() => {
    if (success) {
      const t = setTimeout(() => setSuccess(false), 3000)
      return () => clearTimeout(t)
    }
  }, [success])

  const handleSubmit = async () => {
    if (!allFilled || !passwordsMatch) return
    await changePassword.mutate({
      current_password: currentPassword,
      new_password: newPassword,
    })
    setCurrentPassword('')
    setNewPassword('')
    setConfirmPassword('')
    setSuccess(true)
  }

  return (
    <div className="border border-neutral-900 rounded bg-white p-6 animate-in animate-in-delay-1">
      <div className="flex items-center gap-2 mb-5">
        <Shield className="w-5 h-5 text-neutral-900" />
        <h2 className="text-lg font-semibold text-neutral-900">Безопасность</h2>
      </div>

      <div className="space-y-4 max-w-md">
        <div>
          <label className="block text-sm font-medium text-neutral-700 mb-1.5">Текущий пароль</label>
          <input
            type="password"
            value={currentPassword}
            onChange={(e) => setCurrentPassword(e.target.value)}
            className={inputClassName}
            autoComplete="current-password"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-neutral-700 mb-1.5">Новый пароль</label>
          <input
            type="password"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            className={inputClassName}
            autoComplete="new-password"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-neutral-700 mb-1.5">Повтор нового пароля</label>
          <input
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            className={inputClassName}
            autoComplete="new-password"
          />
          {confirmPassword.length > 0 && !passwordsMatch && (
            <p className="mt-1 text-xs text-red-500">Пароли не совпадают</p>
          )}
        </div>
      </div>

      <div className="mt-6 pt-5 border-t border-neutral-200 flex items-center gap-3">
        <button
          type="button"
          onClick={handleSubmit}
          disabled={!allFilled || !passwordsMatch || changePassword.isPending}
          className={cn(
            'px-5 py-2.5 rounded text-sm font-semibold',
            'bg-neutral-900 text-white hover:bg-neutral-700 active:bg-[#EF3219] transition-colors duration-300 focus:outline-none',
            'disabled:bg-neutral-300 disabled:cursor-not-allowed disabled:hover:bg-neutral-300 disabled:active:bg-neutral-300',
          )}
        >
          {changePassword.isPending ? 'Сохранение...' : 'Сменить пароль'}
        </button>
        {success && (
          <div className="flex items-center gap-1.5 text-sm text-green-600 font-medium">
            <Check className="w-4 h-4" />
            Пароль изменён
          </div>
        )}
        {changePassword.isError && (
          <p className="text-sm text-red-500">Неверный текущий пароль или ошибка сервера.</p>
        )}
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Billing Details Section (Реквизиты)
// ---------------------------------------------------------------------------

function BillingDetailsSection() {
  const { data, isLoading } = useBillingDetailsQuery()
  const updateDetails = useUpdateBillingDetailsMutation()
  const hasRendered = useRef(false)

  const [entityType, setEntityType] = useState<LegalEntityType>('none')
  const [fields, setFields] = useState<Record<string, string>>({})
  const [success, setSuccess] = useState(false)

  useEffect(() => {
    if (data) {
      setEntityType(data.entity_type)
      setFields(data.fields ?? {})
    }
  }, [data])

  useEffect(() => {
    if (success) {
      const t = setTimeout(() => setSuccess(false), 3000)
      return () => clearTimeout(t)
    }
  }, [success])

  const handleSave = async () => {
    const payload: BillingDetails = { entity_type: entityType, fields }
    await updateDetails.mutate(payload)
    setSuccess(true)
  }

  const fieldDefs = entityType !== 'none' ? ENTITY_FIELDS[entityType] : []

  if (isLoading && !hasRendered.current) {
    return <div className="border border-neutral-900 rounded bg-white p-6 h-32 animate-pulse" />
  }

  hasRendered.current = true

  return (
    <div className="border border-neutral-900 rounded bg-white p-6 animate-in animate-in-delay-2">
      <div className="flex items-center gap-2 mb-5">
        <FileText className="w-5 h-5 text-neutral-900" />
        <h2 className="text-lg font-semibold text-neutral-900">Реквизиты</h2>
      </div>

      <div className="max-w-md">
        <label className="block text-sm font-medium text-neutral-700 mb-1.5">Тип лица</label>
        <CustomSelect
          value={entityType}
          onChange={(val) => setEntityType(val as LegalEntityType)}
          options={(Object.keys(ENTITY_TYPE_LABELS) as LegalEntityType[]).map((key) => ({
            value: key,
            label: ENTITY_TYPE_LABELS[key],
          }))}
        />
      </div>

      {fieldDefs.length > 0 && (
        <div className="mt-5 grid grid-cols-1 sm:grid-cols-2 gap-4">
          {fieldDefs.map((fd) => (
            <div key={fd.key}>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">{fd.label}</label>
              <input
                type="text"
                value={fields[fd.key] ?? ''}
                onChange={(e) => setFields((prev) => ({ ...prev, [fd.key]: e.target.value }))}
                className={inputClassName}
              />
            </div>
          ))}
        </div>
      )}

      {entityType !== 'none' && (
        <div className="mt-6 pt-5 border-t border-neutral-200 flex items-center gap-3">
          <button
            type="button"
            onClick={handleSave}
            disabled={updateDetails.isPending}
            className={cn(
              'px-5 py-2.5 rounded text-sm font-semibold',
              'bg-neutral-900 text-white hover:bg-neutral-700 active:bg-[#EF3219] transition-colors duration-300 focus:outline-none',
              'disabled:bg-neutral-300 disabled:cursor-not-allowed disabled:hover:bg-neutral-300 disabled:active:bg-neutral-300',
            )}
          >
            {updateDetails.isPending ? 'Сохранение...' : 'Сохранить'}
          </button>
          {success && (
            <div className="flex items-center gap-1.5 text-sm text-green-600 font-medium">
              <Check className="w-4 h-4" />
              Сохранено
            </div>
          )}
          {updateDetails.isError && (
            <p className="text-sm text-red-500">Не удалось сохранить.</p>
          )}
        </div>
      )}
    </div>
  )
}

// ---------------------------------------------------------------------------
// Page
// ---------------------------------------------------------------------------

export default function AccountPage() {
  const navigate = useNavigate()
  const logout = useAuthStore((s) => s.logout)
  const [showLogoutConfirm, setShowLogoutConfirm] = useState(false)

  const handleLogout = async () => {
    await logout()
    navigate('/auth/login')
  }

  return (
    <div className="max-w-2xl space-y-8">
      <div className="animate-in">
        <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">
          Настройки аккаунта
        </h1>
        <p className="font-mono text-xs text-neutral-300 uppercase tracking-wider mt-1">
          Управление профилем, безопасностью и реквизитами
        </p>
      </div>

      <ProfileSection />
      <SecuritySection />
      <BillingDetailsSection />

      <div className="pt-4 animate-in animate-in-delay-3">
        <button
          type="button"
          onClick={() => setShowLogoutConfirm(true)}
          className={cn(
            'flex items-center gap-2 px-5 py-2.5 rounded text-sm font-medium',
            'text-red-600 bg-red-50 hover:bg-red-100 transition-colors duration-300',
          )}
        >
          <LogOut className="w-4 h-4" />
          Выйти из аккаунта
        </button>
      </div>

      {/* Logout confirmation */}
      {showLogoutConfirm && createPortal(
        <div
          className="fixed inset-0 z-[9999] flex items-center justify-center bg-black/30"
          onClick={() => setShowLogoutConfirm(false)}
        >
          <div
            className="bg-white border border-neutral-900 rounded p-6 w-full max-w-sm mx-4 shadow-lg"
            onClick={(e) => e.stopPropagation()}
          >
            <h3 className="text-lg font-semibold text-neutral-900 mb-2">Выйти из аккаунта?</h3>
            <p className="text-sm text-neutral-500 mb-6">Вы уверены, что хотите выйти?</p>
            <div className="flex items-center gap-3">
              <button
                type="button"
                onClick={handleLogout}
                className="px-5 py-2.5 rounded text-sm font-semibold text-white bg-red-600 hover:bg-red-700 transition-colors"
              >
                Выйти
              </button>
              <button
                type="button"
                onClick={() => setShowLogoutConfirm(false)}
                className="px-5 py-2.5 rounded text-sm font-medium text-neutral-600 hover:text-neutral-900 transition-colors"
              >
                Отмена
              </button>
            </div>
          </div>
        </div>,
        document.body,
      )}
    </div>
  )
}
