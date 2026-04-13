import { Link } from 'react-router-dom'

export default function ForgotPasswordPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-white px-6">
      <div className="w-full max-w-[420px] text-center">
        <h1 className="text-3xl font-bold tracking-tight text-neutral-900 mb-4">Восстановление пароля</h1>
        <p className="text-neutral-500 mb-8">
          Функция восстановления пароля пока в разработке. Напишите администратору или вернитесь ко входу.
        </p>
        <Link
          to="/auth/login"
          className="inline-flex items-center justify-center py-3 px-5 rounded bg-neutral-900 text-white text-sm font-medium hover:bg-neutral-700 transition-colors"
        >
          Вернуться ко входу
        </Link>
      </div>
    </div>
  )
}
