import { Link } from 'react-router-dom'

export function Footer() {
  return (
    <footer className="bg-neutral-900 text-white mt-12" style={{ backgroundImage: "url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='40' height='40'%3E%3Crect x='3' y='5' width='1' height='1' fill='%23fff' opacity='0.08'/%3E%3Crect x='15' y='2' width='1' height='1' fill='%23fff' opacity='0.06'/%3E%3Crect x='29' y='7' width='1' height='1' fill='%23fff' opacity='0.09'/%3E%3Crect x='8' y='14' width='1' height='1' fill='%23fff' opacity='0.07'/%3E%3Crect x='22' y='11' width='1' height='1' fill='%23fff' opacity='0.08'/%3E%3Crect x='36' y='16' width='1' height='1' fill='%23fff' opacity='0.06'/%3E%3Crect x='1' y='23' width='1' height='1' fill='%23fff' opacity='0.09'/%3E%3Crect x='18' y='21' width='1' height='1' fill='%23fff' opacity='0.05'/%3E%3Crect x='27' y='25' width='1' height='1' fill='%23fff' opacity='0.1'/%3E%3Crect x='10' y='30' width='1' height='1' fill='%23fff' opacity='0.08'/%3E%3Crect x='24' y='32' width='1' height='1' fill='%23fff' opacity='0.06'/%3E%3Crect x='34' y='28' width='1' height='1' fill='%23fff' opacity='0.08'/%3E%3Crect x='5' y='37' width='1' height='1' fill='%23fff' opacity='0.07'/%3E%3Crect x='19' y='39' width='1' height='1' fill='%23fff' opacity='0.09'/%3E%3Crect x='31' y='36' width='1' height='1' fill='%23fff' opacity='0.06'/%3E%3Crect x='38' y='9' width='1' height='1' fill='%23fff' opacity='0.08'/%3E%3Crect x='12' y='18' width='1' height='1' fill='%23fff' opacity='0.09'/%3E%3Crect x='28' y='35' width='1' height='1' fill='%23fff' opacity='0.06'/%3E%3Crect x='37' y='23' width='1' height='1' fill='%23fff' opacity='0.08'/%3E%3Crect x='6' y='27' width='1' height='1' fill='%23fff' opacity='0.07'/%3E%3C/svg%3E\")", backgroundSize: "40px 40px" }}>
      <div className="px-4 sm:px-8 lg:px-16 py-10">
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-8">
          {/* Brand */}
          <div>
            <span className="text-xl font-bold tracking-tight">rev/sitr</span>
            <p className="text-sm text-neutral-400 mt-2 leading-relaxed">
              SaaS-платформа лояльности<br />для HoReCa на базе Telegram
            </p>
          </div>

          {/* Product */}
          <div>
            <h3 className="text-sm font-semibold uppercase tracking-wider text-neutral-400 mb-3">Продукт</h3>
            <ul className="space-y-2">
              <li><Link to="/dashboard" className="text-sm text-neutral-300 hover:text-white transition-colors">Панель управления</Link></li>
              <li><a href="#" className="text-sm text-neutral-300 hover:text-white transition-colors">Тарифы</a></li>
              <li><a href="#" className="text-sm text-neutral-300 hover:text-white transition-colors">Маркетинг «под ключ»</a></li>
            </ul>
          </div>

          {/* Support */}
          <div>
            <h3 className="text-sm font-semibold uppercase tracking-wider text-neutral-400 mb-3">Поддержка</h3>
            <ul className="space-y-2">
              <li><a href="#" className="text-sm text-neutral-300 hover:text-white transition-colors">Помощь</a></li>
              <li><a href="#" className="text-sm text-neutral-300 hover:text-white transition-colors">Контакты</a></li>
              <li><a href="#" className="text-sm text-neutral-300 hover:text-white transition-colors">Telegram</a></li>
            </ul>
          </div>

          {/* Legal */}
          <div>
            <h3 className="text-sm font-semibold uppercase tracking-wider text-neutral-400 mb-3">Документы</h3>
            <ul className="space-y-2">
              <li><a href="#" className="text-sm text-neutral-300 hover:text-white transition-colors">Оферта</a></li>
              <li><a href="#" className="text-sm text-neutral-300 hover:text-white transition-colors">Политика конфиденциальности</a></li>
              <li><a href="#" className="text-sm text-neutral-300 hover:text-white transition-colors">Обработка персональных данных</a></li>
            </ul>
          </div>
        </div>

        {/* Bottom line */}
        <div className="mt-10 pt-6 border-t border-neutral-800 flex flex-col sm:flex-row items-center justify-between gap-3">
          <p className="text-xs text-neutral-500">&copy; 2026 Revisitr. Все права защищены.</p>
          <p className="text-xs text-neutral-500">Сделано в России</p>
        </div>
      </div>
    </footer>
  )
}
