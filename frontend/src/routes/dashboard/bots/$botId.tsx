import { Link, useParams, useSearchParams } from "react-router-dom";
import { Suspense, lazy, memo, useState, useEffect } from "react";
import { cn } from "@/lib/utils";
import { useBotQuery } from "@/features/bots/queries";
import { botsApi } from "@/features/bots/api";
import { menusApi } from "@/features/menus/api";
import { Drawer } from "@/components/common/Drawer";
import { usePOSQuery } from "@/features/pos/queries";
import { useProgramsQuery } from "@/features/loyalty/queries";
import { ErrorState } from "@/components/common/ErrorState";
import { CardSkeleton } from "@/components/common/LoadingSkeleton";
import { CustomSelect } from "@/components/common/CustomSelect";
import {
  ArrowLeft,
  Plus,
  Trash2,
  Check,
  Settings,
  Puzzle,
  Store,
  GripVertical,
  ChevronDown,
  ChevronRight,
  Link2,
  Lock,
  SlidersHorizontal,
  AlertTriangle,
} from "lucide-react";
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from "@dnd-kit/core";
import {
  SortableContext,
  verticalListSortingStrategy,
  useSortable,
  arrayMove,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import type {
  Bot,
  BotSettings,
  BotButton,
  FormField,
} from "@/features/bots/types";
import type { POSLocation } from "@/features/pos/types";
import { campaignsApi } from "@/features/campaigns/api";
import type { MessageContent, MessagePlaceholder } from "@/features/telegram-preview";
import { renderTextWithEmoji } from "@/features/telegram-preview/components/renderEmoji";
import { EmojiPicker } from "@/features/emoji-packs";
import { ButtonStylePicker } from "@/features/telegram-preview/components/ButtonStylePicker";
import {
  canAddPreset,
  deriveBotRequirements,
  getEditableButtons,
  getSystemButtons,
  MODULE_DEFS,
  normalizeBotSettings,
  STANDARD_FIELD_PRESETS,
  STANDARD_FIELD_PRESET_LABELS,
  syncSystemButtons,
  isStandardField,
  withModuleDefaults,
} from "@/features/bots/settings";

const TelegramPreview = lazy(() =>
  import("@/features/telegram-preview").then((module) => ({
    default: module.TelegramPreview,
  })),
);
const MessageContentEditor = lazy(() =>
  import("@/features/telegram-preview").then((module) => ({
    default: module.MessageContentEditor,
  })),
);

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const TABS = [
  { id: "connection", label: "Подключение", icon: Link2 },
  { id: "general", label: "Основное", icon: Settings },
  { id: "modules", label: "Модули", icon: Puzzle },
] as const;

const FIELD_TYPE_OPTIONS = [
  { value: "text", label: "Текст" },
  { value: "email", label: "Email" },
  { value: "phone", label: "Телефон" },
  { value: "date", label: "Дата" },
  { value: "select", label: "Выбор" },
];

type VisibleTabId = (typeof TABS)[number]["id"];
type TabId = VisibleTabId;

const statusConfig = {
  active: { label: "Активен", className: "bg-emerald-500/10 text-emerald-700 border border-emerald-500/20" },
  inactive: {
    label: "Неактивен",
    className: "bg-neutral-100 text-neutral-500 border border-neutral-200",
  },
  pending: { label: "Ожидает", className: "bg-amber-500/10 text-amber-700 border border-amber-500/20" },
  error: { label: "Ошибка", className: "bg-red-500/10 text-red-700 border border-red-500/20" },
} as const;

const inputClassName = cn(
  "w-full px-4 py-2.5 rounded border border-neutral-200",
  "text-sm placeholder:text-neutral-400",
  "focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent",
  "transition-colors",
  "disabled:opacity-50 disabled:cursor-not-allowed",
);

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString("ru-RU", {
    day: "numeric",
    month: "short",
    year: "numeric",
  });
}

function buildMessagePlaceholders(fields: FormField[]): MessagePlaceholder[] {
  const seen = new Set<string>();
  const placeholders: MessagePlaceholder[] = [];

  for (const field of fields) {
    const name = field.name.trim();
    if (!name) continue;

    const tokenName = name === "birthday" ? "birth_date" : name;
    const token = `{${tokenName}}`;
    if (seen.has(token)) continue;

    placeholders.push({
      token,
      label: field.label.trim() || name,
      fieldType: field.type,
    });
    seen.add(token);
  }

  return placeholders;
}

function buttonContentFromValue(value: string): MessageContent {
  return {
    parts: [{ type: "text", text: value, parse_mode: "Markdown" }],
  };
}

function normalizeButton(button: BotButton): BotButton {
  return {
    ...button,
    type: "text",
    content:
      button.content && button.content.parts?.length > 0
        ? button.content
        : buttonContentFromValue(button.value ?? ""),
  };
}

function buttonValueFromContent(
  content?: MessageContent,
  fallback = "",
): string {
  if (!content || !content.parts?.length) return fallback;

  const firstTextPart = content.parts.find(
    (part) => part.type === "text" && part.text?.trim(),
  );
  if (firstTextPart?.text) return firstTextPart.text;

  const firstCaptionPart = content.parts.find((part) => part.text?.trim());
  if (firstCaptionPart?.text) return firstCaptionPart.text;

  return fallback;
}

function welcomeMessageFromContent(
  content?: MessageContent,
  fallback = "",
): string {
  if (!content || !content.parts?.length) return fallback;

  const firstTextLike = content.parts.find((part) => part.text?.trim());
  if (firstTextLike?.text) return firstTextLike.text;

  return fallback;
}

function buttonSummary(
  content?: MessageContent,
  fallback = "",
): React.ReactNode {
  const source =
    content && content.parts?.length > 0
      ? content
      : buttonContentFromValue(fallback);
  const rawText = buttonValueFromContent(source, fallback)
    .replace(/\s+/g, " ")
    .trim();
  const mediaCount = source.parts.filter((part) => part.type !== "text").length;
  const inlineButtons = (source.buttons || []).reduce(
    (sum, row) => sum + row.length,
    0,
  );

  const details = [
    source.parts.length > 1 ? `${source.parts.length} блока` : null,
    mediaCount > 0 ? `${mediaCount} медиа` : null,
    inlineButtons > 0 ? `${inlineButtons} inline-кнопок` : null,
  ].filter(Boolean);

  const detailStr = details.join(" · ");
  if (!rawText && !detailStr) return "Ответ пока не настроен";
  if (!rawText) return detailStr;

  // Truncate by visible length (excluding marker syntax), keeping markers intact
  const EMOJI_RE = /\{\{emoji:[^}]+\}\}/g;
  let visibleLen = 0;
  let cutIndex = rawText.length;
  let lastIndex = 0;
  for (const match of rawText.matchAll(EMOJI_RE)) {
    const textBefore = rawText.slice(lastIndex, match.index!);
    if (visibleLen + textBefore.length >= 80) {
      cutIndex = lastIndex + (80 - visibleLen);
      break;
    }
    visibleLen += textBefore.length + 1; // +1 for emoji as single char
    lastIndex = match.index! + match[0].length;
  }
  const remaining = rawText.slice(lastIndex);
  if (visibleLen + remaining.length >= 80 && cutIndex === rawText.length) {
    cutIndex = lastIndex + (80 - visibleLen);
  }
  const truncated =
    cutIndex < rawText.length ? rawText.slice(0, cutIndex) + "…" : rawText;
  // Ensure we don't break in the middle of a marker — find last complete marker
  const safeText = truncated.replace(/\{\{emoji:[^}]*$/, "…");
  const rendered = renderTextWithEmoji(safeText);
  return (
    <>
      {rendered}
      {detailStr ? <> · {detailStr}</> : null}
    </>
  );
}

function SectionBlock({
  eyebrow,
  title,
  description,
  actions,
  children,
}: {
  eyebrow?: string;
  title: string;
  description?: string;
  actions?: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <section className="rounded border border-neutral-900 bg-white p-4 sm:p-5 md:p-6">
      <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between mb-5">
        <div className="min-w-0">
          {eyebrow && (
            <div className="font-mono text-[10px] uppercase tracking-wider text-neutral-300 mb-2">
              {eyebrow}
            </div>
          )}
          <h3 className="text-sm font-semibold text-neutral-700">{title}</h3>
          {description && (
            <p className="text-sm text-neutral-500 mt-1 max-w-2xl">
              {description}
            </p>
          )}
        </div>
        {actions && <div className="shrink-0 w-full md:w-auto">{actions}</div>}
      </div>
      {children}
    </section>
  );
}

function PreviewFallback() {
  return (
    <div className="w-full max-w-[360px] overflow-hidden rounded-[2rem] border border-neutral-200 bg-white shadow-sm">
      <div className="h-[min(520px,65vh)] sm:h-[520px] animate-pulse bg-neutral-100" />
    </div>
  );
}

function EditorFallback() {
  return (
    <div className="space-y-3 rounded border border-neutral-200 bg-white p-4">
      <div className="h-11 rounded bg-neutral-100 animate-pulse" />
      <div className="h-32 rounded bg-neutral-100 animate-pulse" />
      <div className="h-11 rounded bg-neutral-100 animate-pulse" />
    </div>
  );
}

function WarningBanner({
  items,
  onSelectTab,
}: {
  items: ReturnType<typeof deriveBotRequirements>;
  onSelectTab: (tab: TabId) => void;
}) {
  if (items.length === 0) return null;

  return (
    <div className="mb-6 flex items-start gap-3 rounded border border-orange-400/50 bg-orange-50 p-4">
      <AlertTriangle className="w-4 h-4 shrink-0 mt-0.5 text-orange-500" />
      <div className="min-w-0 flex-1">
        <p className="text-sm font-semibold text-orange-900">
          Для запуска бота не хватает настроек
        </p>
        <p className="text-sm text-orange-800/90 mt-0.5">
          Заполните недостающие поля, чтобы бот стал доступен клиентам.
        </p>
        <div className="mt-2 flex flex-wrap gap-1.5">
          {items.map((item) => (
            <button
              key={item.id}
              type="button"
              onClick={() => onSelectTab(item.tab)}
              className="inline-flex items-center gap-1 px-2.5 py-1 rounded border border-orange-400/40 bg-white/70 text-orange-900 text-xs font-medium hover:bg-white hover:border-orange-500 transition-colors"
            >
              {item.label}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}

/** Per-tab save hook returning state + handler */
function useSaveAction(_id: number, updater: () => Promise<void>) {
  const [isSaving, setIsSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [saveSuccess, setSaveSuccess] = useState(false);

  useEffect(() => {
    if (!saveSuccess) return;
    const t = setTimeout(() => setSaveSuccess(false), 3000);
    return () => clearTimeout(t);
  }, [saveSuccess]);

  const save = async () => {
    setIsSaving(true);
    setSaveError(null);
    setSaveSuccess(false);
    try {
      await updater();
      setSaveSuccess(true);
    } catch {
      setSaveError("Не удалось сохранить изменения. Попробуйте ещё раз.");
    } finally {
      setIsSaving(false);
    }
  };

  return { isSaving, saveError, saveSuccess, save };
}

function SaveButton({
  isSaving,
  saveError,
  saveSuccess,
  onSave,
}: {
  isSaving: boolean;
  saveError: string | null;
  saveSuccess: boolean;
  onSave: () => void;
}) {
  return (
    <div className="flex flex-col gap-3 mt-6 pt-5 border-t border-neutral-200 sm:flex-row sm:items-center sm:justify-between">
      <div>
        {saveError && (
          <div className="flex items-center gap-2 px-4 py-2.5 rounded bg-red-50 border border-red-100">
            <span className="text-red-500 font-mono text-xs">&#x2715;</span>
            <p className="text-sm text-red-600">{saveError}</p>
          </div>
        )}
        {saveSuccess && (
          <div className="flex items-center gap-2 px-4 py-2.5 rounded bg-green-50 border border-green-100">
            <Check className="w-3.5 h-3.5 text-green-600" />
            <p className="text-sm text-green-600 font-medium">
              Изменения сохранены
            </p>
          </div>
        )}
      </div>
      <button
        type="button"
        onClick={onSave}
        disabled={isSaving}
        className={cn(
          "inline-flex w-full items-center justify-center rounded py-3 px-6 sm:w-auto",
          "bg-accent text-white text-sm font-semibold",
          "hover:bg-accent-hover active:bg-accent/80",
          "transition-all duration-150",
          "focus:outline-none focus:ring-2 focus:ring-accent/20",
          "shadow-sm shadow-accent/20",
          "disabled:opacity-50 disabled:cursor-not-allowed",
        )}
      >
        {isSaving ? "Сохраняем…" : "Сохранить изменения"}
      </button>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Sortable item wrapper
// ---------------------------------------------------------------------------

function SortableItem({
  id,
  children,
}: {
  id: string;
  children: (props: {
    listeners: ReturnType<typeof useSortable>["listeners"];
    attributes: ReturnType<typeof useSortable>["attributes"];
    setNodeRef: ReturnType<typeof useSortable>["setNodeRef"];
  }) => React.ReactNode;
}) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id,
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <div ref={setNodeRef} style={style}>
      {children({ listeners, attributes, setNodeRef })}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Page Component
// ---------------------------------------------------------------------------

export default function BotDetailPage() {
  const { botId } = useParams<{ botId: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const id = Number(botId);
  const {
    data: bot,
    isLoading,
    isError,
  } = useBotQuery(isNaN(id) || id <= 0 ? 0 : id);
  const [activeTab, setActiveTab] = useState<TabId>(() => {
    const tab = searchParams.get("tab");
    return TABS.some((item) => item.id === tab) ? (tab as TabId) : "connection";
  });
  const [visitedTabs, setVisitedTabs] = useState<Set<TabId>>(
    () => new Set([activeTab]),
  );
  const [settings, setSettings] = useState<BotSettings | null>(null);
  const [selectedProgramId, setSelectedProgramId] = useState<
    number | undefined
  >(bot?.program_id);
  const [boundPosIds, setBoundPosIds] = useState<number[]>([]);
  const [posBindingsLoaded, setPosBindingsLoaded] = useState(false);
  const { data: posLocations = [] } = usePOSQuery();

  useEffect(() => {
    const tab = searchParams.get("tab");
    if (tab && TABS.some((item) => item.id === tab) && tab !== activeTab) {
      setActiveTab(tab as TabId);
    }
  }, [activeTab, searchParams]);

  useEffect(() => {
    setVisitedTabs((prev) =>
      prev.has(activeTab) ? prev : new Set([...prev, activeTab]),
    );
  }, [activeTab]);

  useEffect(() => {
    if (bot?.settings) {
      const normalized = normalizeBotSettings(bot.settings);
      setSettings({
        ...normalized,
        buttons: normalized.buttons.map(normalizeButton),
      });
    }
  }, [bot]);

  useEffect(() => {
    setSelectedProgramId(bot?.program_id);
  }, [bot?.program_id]);

  useEffect(() => {
    if (!id || Number.isNaN(id) || id <= 0) return;

    let isMounted = true;

    menusApi
      .getBotPOSLocations(id)
      .then((response) => {
        if (!isMounted) return;
        setBoundPosIds(response.pos_ids ?? []);
        setPosBindingsLoaded(true);
      })
      .catch(() => {
        if (!isMounted) return;
        setBoundPosIds([]);
        setPosBindingsLoaded(true);
      });

    return () => {
      isMounted = false;
    };
  }, [id]);

  useEffect(() => {
    if (!settings || !posBindingsLoaded) return;

    if (
      bot?.settings.contacts_pos_ids === undefined &&
      (settings.contacts_pos_ids?.length ?? 0) === 0 &&
      boundPosIds.length > 0
    ) {
      setSettings((current) =>
        current ? { ...current, contacts_pos_ids: boundPosIds } : current,
      );
    }
  }, [
    bot?.settings.contacts_pos_ids,
    boundPosIds,
    posBindingsLoaded,
    settings,
  ]);

  if (isLoading) {
    return (
      <div className="max-w-4xl">
        <div className="mb-6 animate-in">
          <div className="h-4 w-32 shimmer rounded" />
        </div>
        <div className="space-y-6">
          {[0, 1, 2].map((i) => (
            <div
              key={i}
              className={cn("animate-in", `animate-in-delay-${i + 1}`)}
            >
              <CardSkeleton />
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (isNaN(id) || id <= 0 || isError || !bot) {
    return (
      <div className="max-w-4xl">
        <Link
          to="/dashboard/bots"
          className="inline-flex items-center gap-1.5 text-sm text-neutral-500 hover:text-neutral-700 transition-colors mb-6"
        >
          <ArrowLeft className="w-4 h-4" />
          Назад к списку
        </Link>
        <ErrorState
          title="Не удалось загрузить бота"
          message="Проверьте подключение к серверу и попробуйте снова."
        />
      </div>
    );
  }

  const status = statusConfig[bot.status as keyof typeof statusConfig] ?? {
    label: bot.status,
    className: "bg-neutral-100 text-neutral-500",
  };

  const requirements = settings
    ? deriveBotRequirements({
        bot: {
          name: bot.name,
          username: bot.username,
          program_id: selectedProgramId,
          created_by_telegram_id: bot.created_by_telegram_id,
        },
        settings,
        boundPosIds,
      })
    : [];

  const handleTabChange = (tab: TabId) => {
    setActiveTab(tab);
    const next = new URLSearchParams(searchParams);
    next.set("tab", tab);
    setSearchParams(next, { replace: true });
  };

  return (
    <div className="max-w-4xl">
      {/* Header */}
      <div className="animate-in mb-4">
        <div className="flex items-center gap-3">
          <Link
            to="/dashboard/bots"
            className="p-2 rounded text-neutral-400 hover:text-neutral-600 hover:bg-neutral-100 transition-colors"
          >
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <h1 className="font-display text-3xl font-bold text-neutral-900 tracking-tight truncate">
                {bot.name}
              </h1>
              <span
                className={cn(
                  "font-mono text-[10px] px-2 py-0.5 rounded uppercase tracking-wider shrink-0",
                  status.className,
                )}
              >
                {status.label}
              </span>
            </div>
            <p className="text-xs text-neutral-400 uppercase tracking-wider mt-1 truncate">
              {bot.username ? `@${bot.username}` : "Настройки бота"} · Создан {formatDate(bot.created_at)}
            </p>
          </div>
        </div>
      </div>

      <WarningBanner items={requirements} onSelectTab={handleTabChange} />

      {/* Tabs */}
      <div className="mb-6 border-b border-neutral-200 overflow-x-auto animate-in animate-in-delay-1">
        <div className="flex gap-1 min-w-max">
          {TABS.map((tab) => (
            <button
              key={tab.id}
              type="button"
              onClick={() => handleTabChange(tab.id)}
              className={cn(
                "inline-flex min-h-11 items-center gap-1.5 px-4 py-2.5 text-sm font-medium whitespace-nowrap",
                "border-b-2 transition-colors",
                activeTab === tab.id
                  ? "border-accent text-accent"
                  : "border-transparent text-neutral-500 hover:text-neutral-700",
              )}
            >
              <tab.icon className="w-4 h-4" />
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      {/* Tab content — keep-alive: each visited tab stays mounted, inactive ones hidden via CSS to avoid remount lag */}
      <div className="animate-in animate-in-delay-2">
        {settings && (
          <>
            {visitedTabs.has("connection") && (
              <div hidden={activeTab !== "connection"}>
                <ConnectionTab
                  bot={bot}
                  botId={id}
                  settings={settings}
                  setSettings={setSettings}
                  selectedProgramId={selectedProgramId}
                  setSelectedProgramId={setSelectedProgramId}
                  boundPosIds={boundPosIds}
                  setBoundPosIds={setBoundPosIds}
                  posLocations={posLocations}
                  posBindingsLoaded={posBindingsLoaded}
                />
              </div>
            )}
            {visitedTabs.has("general") && (
              <div hidden={activeTab !== "general"}>
                <GeneralTab
                  botId={id}
                  settings={settings}
                  setSettings={setSettings}
                  botName={bot.name}
                  boundPosIds={boundPosIds}
                  posLocations={posLocations}
                />
              </div>
            )}
            {visitedTabs.has("modules") && (
              <div hidden={activeTab !== "modules"}>
                <ModulesTab
                  botId={id}
                  selectedProgramId={selectedProgramId}
                  settings={settings}
                  setSettings={setSettings}
                  boundPosIds={boundPosIds}
                  posLocations={posLocations}
                />
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}

// ===========================================================================
// Tab: Подключение (Connection) — General info + POS
// ===========================================================================

const ConnectionTab = memo(function ConnectionTab({
  bot,
  botId,
  settings,
  setSettings,
  selectedProgramId,
  setSelectedProgramId,
  boundPosIds,
  setBoundPosIds,
  posLocations,
  posBindingsLoaded,
}: {
  bot: Bot;
  botId: number;
  settings: BotSettings;
  setSettings: React.Dispatch<React.SetStateAction<BotSettings | null>>;
  selectedProgramId?: number;
  setSelectedProgramId: React.Dispatch<
    React.SetStateAction<number | undefined>
  >;
  boundPosIds: number[];
  setBoundPosIds: React.Dispatch<React.SetStateAction<number[]>>;
  posLocations: POSLocation[];
  posBindingsLoaded: boolean;
}) {
  const [advanced, setAdvanced] = useState(false);
  const { data: programs = [] } = useProgramsQuery();
  const {
    isSaving: isSavingGeneral,
    saveError: generalSaveError,
    saveSuccess: generalSaveSuccess,
    save: saveGeneral,
  } = useSaveAction(botId, () =>
    botsApi
      .update(botId, {
        program_id: selectedProgramId,
      })
      .then(() => undefined),
  );

  return (
    <div className="space-y-6">
      <section className="bg-white rounded border border-neutral-900 p-4 sm:p-5 md:p-6">
        <div className="flex items-center justify-between mb-5">
          <div>
            <p className="font-mono text-[10px] uppercase tracking-wider text-neutral-300 mb-1">Информация</p>
            <h3 className="text-sm font-semibold text-neutral-700">Общие данные</h3>
          </div>
          <button
            type="button"
            onClick={() => setAdvanced(!advanced)}
            className={cn(
              "text-xs font-medium px-3 py-1.5 rounded transition-colors",
              advanced
                ? "bg-accent/10 text-accent"
                : "bg-neutral-100 text-neutral-500 hover:bg-neutral-200",
            )}
          >
            {advanced ? "Скрыть тех. детали" : "Показать тех. детали"}
          </button>
        </div>

        <div className="space-y-4">
        <InfoRow label="Название" value={bot.name} />
        <InfoRow
          label="Username"
          value={bot.username ? `@${bot.username}` : "---"}
        />
        <InfoRow label="Статус" value={bot.status} />
        <InfoRow label="Дата создания" value={formatDate(bot.created_at)} />
        <div className="rounded border border-neutral-200 bg-neutral-50/70 p-4">
          <label className="text-sm text-neutral-500" htmlFor="loyalty-program">
            Программа лояльности
          </label>
          <div className="mt-2">
            <CustomSelect
              value={selectedProgramId ? String(selectedProgramId) : ""}
              onChange={(v) => setSelectedProgramId(v ? Number(v) : undefined)}
              options={[
                { value: "", label: "Без программы" },
                ...programs.map((p) => ({ value: String(p.id), label: p.name })),
              ]}
              light
            />
          </div>
          <p className="mt-2 text-xs text-neutral-500">
            Сохраняется через основное обновление бота, без перехода в раздел
            лояльности.
          </p>
          <div className="mt-4">
            <SaveButton
              isSaving={isSavingGeneral}
              saveError={generalSaveError}
              saveSuccess={generalSaveSuccess}
              onSave={saveGeneral}
            />
          </div>
        </div>

        {advanced && (
          <>
            <p className="text-xs text-neutral-400 leading-relaxed">
              Здесь собраны технические данные бота. Они нужны в редких случаях
              — например, если поддержку попросят прислать ID бота или проверить
              токен.
            </p>
            {bot.token_masked && (
              <div>
                <span className="block text-sm text-neutral-500 mb-1">
                  Токен
                </span>
                <div
                  className={cn(
                    "px-4 py-2.5 rounded bg-neutral-50 border border-neutral-200",
                    "font-mono text-sm text-neutral-600 select-none",
                  )}
                >
                  {bot.token_masked}
                </div>
              </div>
            )}
            <div>
              <span className="block text-sm text-neutral-500 mb-1">
                ID бота
              </span>
              <div className="font-mono text-sm text-neutral-600">{bot.id}</div>
            </div>
            <div>
              <span className="block text-sm text-neutral-500 mb-1">
                ID организации
              </span>
              <div className="font-mono text-sm text-neutral-600">
                {bot.org_id}
              </div>
            </div>
          </>
        )}
        </div>
      </section>

      <section className="bg-white rounded border border-neutral-900 p-4 sm:p-5 md:p-6">
        <div className="mb-5">
          <p className="font-mono text-[10px] uppercase tracking-wider text-neutral-300 mb-1">Локации</p>
          <h3 className="text-sm font-semibold text-neutral-700">POS-точки бота</h3>
        </div>
        <POSSection
          botId={botId}
          settings={settings}
          setSettings={setSettings}
          posLocations={posLocations}
          boundPosIds={boundPosIds}
          setBoundPosIds={setBoundPosIds}
          isLoaded={posBindingsLoaded}
        />
      </section>
    </div>
  );
});

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between py-2 border-b border-neutral-200/50 last:border-0">
      <span className="text-sm text-neutral-500">{label}</span>
      <span className="text-sm font-medium text-neutral-900">{value}</span>
    </div>
  );
}

/** POS section embedded inside ConnectionTab */
function POSSection({
  botId,
  settings,
  setSettings,
  posLocations,
  boundPosIds,
  setBoundPosIds,
  isLoaded,
}: {
  botId: number;
  settings: BotSettings;
  setSettings: React.Dispatch<React.SetStateAction<BotSettings | null>>;
  posLocations: POSLocation[];
  boundPosIds: number[];
  setBoundPosIds: React.Dispatch<React.SetStateAction<number[]>>;
  isLoaded: boolean;
}) {
  const { isSaving, saveError, saveSuccess, save } = useSaveAction(botId, () =>
    Promise.all([
      menusApi.setBotPOSLocations(botId, boundPosIds),
      botsApi.updateSettings(botId, {
        pos_selector_enabled: settings.pos_selector_enabled,
      }),
    ]).then(() => undefined),
  );

  if (!isLoaded) {
    return <p className="text-sm text-neutral-500">Загружаем точки продаж…</p>;
  }

  const locations = posLocations ?? [];

  if (locations.length === 0) {
    return (
      <div className="text-center py-6">
        <Store className="w-10 h-10 text-neutral-300 mx-auto mb-3" />
        <p className="text-sm text-neutral-500 mb-4">Пока нет точек продаж</p>
        <Link
          to="/dashboard/pos"
          className={cn(
            "inline-flex items-center gap-1.5 py-2 px-4 rounded text-sm font-medium",
            "bg-accent text-white hover:bg-accent-hover transition-colors",
          )}
        >
          Управление POS
          <ChevronRight className="w-4 h-4" />
        </Link>
      </div>
    );
  }

  const togglePos = (posId: number) => {
    setBoundPosIds((prev) =>
      prev.includes(posId) ? prev.filter((x) => x !== posId) : [...prev, posId],
    );
  };

  return (
    <>
      <p className="text-sm text-neutral-400 mb-5">
        Выберите, в каких точках будет работать этот бот. Если точек несколько,
        гость сможет выбрать нужную после запуска бота.
      </p>

      <div className="space-y-2">
        {locations.map((loc: POSLocation) => {
          const isOn = boundPosIds.includes(loc.id)
          return (
            <button
              key={loc.id}
              type="button"
              role="switch"
              aria-checked={isOn}
              onClick={(e) => {
                e.preventDefault()
                togglePos(loc.id)
              }}
              disabled={isSaving}
              className={cn(
                "w-full flex items-center gap-3 p-3 rounded transition-colors text-left border",
                isOn ? "bg-accent/5 border-accent/30" : "bg-neutral-50 border-transparent hover:border-neutral-200",
                "disabled:opacity-50 disabled:cursor-not-allowed",
              )}
            >
              <span
                className={cn(
                  "flex h-4 w-4 shrink-0 items-center justify-center rounded border transition-colors",
                  isOn ? "bg-accent border-accent" : "bg-white border-neutral-300",
                )}
              >
                {isOn && <Check className="w-3 h-3 text-white" />}
              </span>
              <span className="min-w-0">
                <span className="block text-sm font-medium text-neutral-900">{loc.name}</span>
                {loc.address && (
                  <span className="block text-xs text-neutral-500 truncate">{loc.address}</span>
                )}
              </span>
            </button>
          )
        })}
      </div>

      <div className="mt-5 rounded border border-neutral-200 bg-neutral-50/70 p-4">
        <div className="flex items-center justify-between gap-4">
          <div>
            <p className="text-sm font-medium text-neutral-900">
              Добавить выбор заведения при запуске бота
            </p>
            <p className="text-xs text-neutral-500 mt-1">
              Если к боту привязано больше одной точки, перед приветствием
              покажем выбор заведения.
            </p>
          </div>
          <button
            type="button"
            role="switch"
            aria-checked={Boolean(settings.pos_selector_enabled)}
            onClick={() =>
              setSettings((current) =>
                current
                  ? {
                      ...current,
                      pos_selector_enabled: !current.pos_selector_enabled,
                    }
                  : current,
              )
            }
            className={cn(
              "relative h-6 w-10 shrink-0 rounded-full transition-colors",
              settings.pos_selector_enabled ? "bg-accent" : "bg-neutral-300",
            )}
          >
            <span
              className={cn(
                "absolute left-0.5 top-0.5 h-5 w-5 rounded-full bg-white shadow transition-transform",
                settings.pos_selector_enabled && "translate-x-4",
              )}
            />
          </button>
        </div>
      </div>

      <SaveButton
        isSaving={isSaving}
        saveError={saveError}
        saveSuccess={saveSuccess}
        onSave={save}
      />
    </>
  );
}

// ===========================================================================
// Tab: Основное (General) — Messages + Buttons + Form combined
// ===========================================================================

const GeneralTab = memo(function GeneralTab({
  botId,
  settings,
  setSettings,
  botName,
  boundPosIds,
  posLocations,
}: {
  botId: number;
  settings: BotSettings;
  setSettings: React.Dispatch<React.SetStateAction<BotSettings | null>>;
  botName: string;
  boundPosIds: number[];
  posLocations: POSLocation[];
}) {
  const [expandedButtonId, setExpandedButtonId] = useState<string | null>(null);
  const [configureDrawer, setConfigureDrawer] = useState<
    { index: number; module: "home" | "contacts" } | null
  >(null);
  const messagePlaceholders = buildMessagePlaceholders(
    settings.registration_form,
  );

  const { isSaving, saveError, saveSuccess, save } = useSaveAction(botId, () =>
    botsApi.updateSettings(botId, {
      welcome_message: settings.welcome_message,
      welcome_content: settings.welcome_content,
      buttons: settings.buttons.map((button) => ({
        ...button,
        type: "text",
        value: buttonValueFromContent(button.content, button.value),
        content: button.content,
      })),
      registration_form: settings.registration_form,
      contacts_pos_ids: settings.contacts_pos_ids,
    }),
  );

  const editableButtons = getEditableButtons(settings.buttons);
  const systemButtons = getSystemButtons(settings.modules);

  // --- Buttons dnd ---
  const buttonSensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor),
  );

  const buttonIds = settings.buttons.map((_, i) => `btn-${i}`);

  const handleButtonDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) return;
    const oldIndex = buttonIds.indexOf(String(active.id));
    const newIndex = buttonIds.indexOf(String(over.id));
    setSettings((current) => {
      if (!current) return current;
      return {
        ...current,
        buttons: arrayMove(current.buttons, oldIndex, newIndex),
      };
    });
  };

  const addButton = () => {
    if (editableButtons.length >= 10) return;
    setSettings((s) =>
      s
        ? {
            ...s,
            buttons: syncSystemButtons(
              [
                ...s.buttons,
                {
                  label: "",
                  type: "text",
                  value: "",
                  content: buttonContentFromValue(""),
                },
              ],
              s.modules,
            ),
          }
        : s,
    );
    setExpandedButtonId(`btn-${settings.buttons.length}`);
  };

  const updateButton = (
    index: number,
    field: keyof BotButton,
    value: string | boolean,
  ) => {
    setSettings((s) => {
      if (!s) return s;
      const updated = s.buttons.map((btn, i) =>
        i === index ? { ...btn, [field]: value } : btn,
      );
      return { ...s, buttons: updated };
    });
  };

  const removeButton = (index: number) => {
    setSettings((s) =>
      s ? { ...s, buttons: s.buttons.filter((_, i) => i !== index) } : s,
    );
  };

  const updateButtonContent = (index: number, content: MessageContent) => {
    setSettings((s) => {
      if (!s) return s;
      const updated = s.buttons.map((btn, i) =>
        i === index
          ? {
              ...btn,
              content,
              value: buttonValueFromContent(content, btn.value),
            }
          : btn,
      );
      return { ...s, buttons: updated };
    });
  };

  const toggleButtonExpanded = (id: string) => {
    setExpandedButtonId((current) => (current === id ? null : id));
  };

  // --- Form dnd ---
  const formSensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor),
  );

  const fieldIds = settings.registration_form.map((_, i) => `field-${i}`);

  const handleFieldDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) return;
    const oldIndex = fieldIds.indexOf(String(active.id));
    const newIndex = fieldIds.indexOf(String(over.id));
    setSettings((s) =>
      s
        ? {
            ...s,
            registration_form: arrayMove(
              s.registration_form,
              oldIndex,
              newIndex,
            ),
          }
        : s,
    );
  };

  const addField = () => {
    setSettings((s) =>
      s
        ? {
            ...s,
            registration_form: [
              ...s.registration_form,
              { name: "", label: "", type: "text", required: false },
            ],
          }
        : s,
    );
  };

  const addPreset = (preset: FormField) => {
    if (!canAddPreset(settings.registration_form, preset.name)) return;
    setSettings((s) =>
      s
        ? { ...s, registration_form: [...s.registration_form, { ...preset }] }
        : s,
    );
  };

  const updateField = (
    index: number,
    field: keyof FormField,
    value: string | boolean,
  ) => {
    setSettings((s) => {
      if (!s) return s;
      const updated = s.registration_form.map((f, i) =>
        i === index ? { ...f, [field]: value } : f,
      );
      return { ...s, registration_form: updated };
    });
  };

  const removeField = (index: number) => {
    setSettings((s) =>
      s
        ? {
            ...s,
            registration_form: s.registration_form.filter(
              (_, i) => i !== index,
            ),
          }
        : s,
    );
  };

  // Welcome content state (composite message)
  const welcomeContent: import("@/features/telegram-preview").MessageContent =
    settings.welcome_content && settings.welcome_content.parts?.length > 0
      ? settings.welcome_content
      : {
          parts: settings.welcome_message
            ? [
                {
                  type: "text" as const,
                  text: settings.welcome_message,
                  parse_mode: "Markdown" as const,
                },
              ]
            : [
                {
                  type: "text" as const,
                  text: "",
                  parse_mode: "Markdown" as const,
                },
              ],
        };

  return (
    <div className="space-y-6">
        <SectionBlock
          eyebrow="Старт"
          title="Приветственное сообщение"
          description="Это первое сообщение, которое увидит гость после /start. Здесь можно собрать текст, медиа и inline-кнопки."
        >
          <div className="grid grid-cols-1 xl:grid-cols-[minmax(0,1.1fr)_minmax(340px,0.9fr)] gap-6 xl:gap-8 items-start">
            <div className="space-y-4">
              <div className="rounded border border-neutral-200 bg-neutral-50/70 p-4">
                <Suspense fallback={<EditorFallback />}>
                  <MessageContentEditor
                    value={welcomeContent}
                    onChange={(content) => {
                      setSettings((s) =>
                        s
                          ? {
                              ...s,
                              welcome_content: content,
                              welcome_message: welcomeMessageFromContent(
                                content,
                                s.welcome_message,
                              ),
                            }
                          : s,
                      );
                    }}
                    onUpload={campaignsApi.uploadFile}
                    maxParts={5}
                    placeholders={messagePlaceholders}
                  />
                </Suspense>
              </div>
            </div>

            <div className="xl:sticky xl:top-6 flex justify-center">
              <Suspense fallback={<PreviewFallback />}>
                <TelegramPreview
                  botName={botName}
                  content={welcomeContent}
                  showFrame
                  className="w-full max-w-[360px]"
                />
              </Suspense>
            </div>
          </div>
        </SectionBlock>

        <SectionBlock
          eyebrow="Клавиатура"
          title="Кнопки меню бота"
          description="Системные кнопки можно переставлять драгом, но они не редактируются и не удаляются здесь. Их содержимое настраивается в соответствующих разделах."
          actions={
            <button
              type="button"
              onClick={addButton}
              disabled={isSaving || editableButtons.length >= 10}
              className={cn(
                "inline-flex min-h-11 w-full items-center justify-center gap-1.5 rounded px-3 text-sm font-medium text-accent sm:w-auto",
                "hover:text-accent/80 transition-colors",
                "disabled:opacity-50 disabled:cursor-not-allowed",
              )}
            >
              <Plus className="w-4 h-4" />
              Добавить кнопку
            </button>
          }
        >
          {settings.buttons.length === 0 ? (
            <p className="text-sm text-neutral-400 text-center py-6">
              Кнопки ещё не добавлены. Создайте первую кнопку для меню бота.
            </p>
          ) : (
            <DndContext
              sensors={buttonSensors}
              collisionDetection={closestCenter}
              onDragEnd={handleButtonDragEnd}
            >
              <SortableContext
                items={buttonIds}
                strategy={verticalListSortingStrategy}
              >
                <div className="space-y-3">
                  {settings.buttons.map((button, index) => (
                    <SortableItem key={buttonIds[index]} id={buttonIds[index]}>
                      {({ listeners, attributes }) => {
                        const isExpanded =
                          expandedButtonId === buttonIds[index];
                        const buttonContent =
                          button.content && button.content.parts?.length > 0
                            ? button.content
                            : buttonContentFromValue(button.value);
                        const isSystemButton = Boolean(
                          button.is_system || button.managed_by_module,
                        );
                        const buttonModule = systemButtons.find(
                          (item) =>
                            item.managed_by_module === button.managed_by_module,
                        );
                        const isHomeButton = button.managed_by_module === "home";
                        const isContactsButton = button.managed_by_module === "contacts";
                        const canExpand =
                          !isSystemButton ||
                          (isSystemButton && buttonModule?.isExpandable) ||
                          isHomeButton ||
                          isContactsButton;
                        const configureHref =
                          button.managed_by_module === "menu"
                            ? `/dashboard/bots/${botId}/menu`
                            : button.managed_by_module === "contacts"
                              ? `/dashboard/bots/${botId}?tab=general`
                              : button.managed_by_module === "home"
                                ? `/dashboard/bots/${botId}?tab=general`
                                : button.managed_by_module
                                  ? `/dashboard/bots/${botId}?tab=modules`
                                  : null;

                        return (
                          <div
                            className={cn(
                              "rounded border overflow-hidden",
                              isSystemButton
                                ? "border-amber-200 bg-amber-50/60"
                                : "border-neutral-200 bg-white",
                            )}
                          >
                            <div className="flex items-start gap-3 px-4 py-3">
                              <button
                                type="button"
                                className="mt-1 inline-flex min-h-11 min-w-11 items-center justify-center rounded text-neutral-400 hover:text-neutral-600 hover:bg-neutral-100 cursor-grab active:cursor-grabbing touch-none"
                                {...listeners}
                                {...attributes}
                                aria-label="Перетащить"
                              >
                                <GripVertical className="w-4 h-4" />
                              </button>

                              <div className="flex-1 min-w-0 self-center">
                                <div className="flex items-center gap-2 text-sm font-semibold text-neutral-900">
                                  {isSystemButton && (
                                    <Lock className="h-4 w-4 shrink-0 text-amber-700" />
                                  )}
                                  <span className="truncate">
                                    {button.label || `Кнопка ${index + 1}`}
                                  </span>
                                </div>
                                <div className="text-xs text-neutral-500 mt-0.5 break-words">
                                  {isSystemButton
                                    ? buttonModule?.isExpandable
                                      ? button.content?.parts?.[0]?.text
                                        ? `${button.content.parts[0].text.slice(0, 60)}${button.content.parts[0].text.length > 60 ? "…" : ""}`
                                        : "Нажмите для настройки приветствия"
                                      : `Системная кнопка${buttonModule ? ` · ${buttonModule.configureLabel}` : ""}`
                                    : buttonSummary(
                                        button.content,
                                        button.value,
                                      )}
                                </div>
                              </div>

                              <div className="flex items-center gap-1.5 self-center">
                                {/* Chevron — only for user-created buttons */}
                                {!isSystemButton && (
                                  <button
                                    type="button"
                                    onClick={() => toggleButtonExpanded(buttonIds[index])}
                                    disabled={!canExpand}
                                    aria-label={canExpand ? (isExpanded ? "Свернуть" : "Развернуть") : undefined}
                                    className={cn(
                                      "inline-flex h-8 w-8 shrink-0 items-center justify-center rounded text-neutral-400 transition-colors",
                                      canExpand ? "hover:text-neutral-700 hover:bg-neutral-100" : "opacity-0 pointer-events-none",
                                    )}
                                  >
                                    <ChevronDown
                                      className={cn(
                                        "h-4 w-4 transition-transform",
                                        isExpanded && "rotate-180",
                                      )}
                                    />
                                  </button>
                                )}

                                {/* Configure */}
                                {(isHomeButton || isContactsButton) ? (
                                  <button
                                    type="button"
                                    onClick={() => toggleButtonExpanded(buttonIds[index])}
                                    className="mr-2 inline-flex h-8 cursor-pointer items-center justify-center gap-1.5 rounded px-2.5 text-xs font-medium text-accent border border-accent/30 bg-accent/5 hover:bg-accent/10 transition-colors"
                                  >
                                    <SlidersHorizontal className="h-3.5 w-3.5" />
                                    Настроить
                                    <ChevronDown
                                      className={cn(
                                        "h-3.5 w-3.5 transition-transform",
                                        isExpanded && "rotate-180",
                                      )}
                                    />
                                  </button>
                                ) : isSystemButton && !buttonModule?.isExpandable && configureHref ? (
                                  <Link
                                    to={configureHref}
                                    className="mr-2 inline-flex h-8 items-center justify-center gap-1.5 rounded px-2.5 text-xs font-medium text-accent border border-accent/30 bg-accent/5 hover:bg-accent/10 transition-colors"
                                  >
                                    <SlidersHorizontal className="h-3.5 w-3.5" />
                                    Настроить
                                  </Link>
                                ) : null}

                                {/* Toggle */}
                                <button
                                  type="button"
                                  role="switch"
                                  aria-checked={button.is_enabled !== false}
                                  onClick={() =>
                                    updateButton(
                                      index,
                                      "is_enabled",
                                      button.is_enabled === false,
                                    )
                                  }
                                  disabled={isSaving}
                                  title={button.is_enabled === false ? "Включить" : "Выключить"}
                                  className={cn(
                                    "relative h-5 w-9 shrink-0 rounded-full transition-colors",
                                    "disabled:opacity-50 disabled:cursor-not-allowed",
                                    button.is_enabled === false ? "bg-neutral-300" : "bg-accent",
                                  )}
                                >
                                  <span
                                    className={cn(
                                      "absolute top-0.5 h-4 w-4 rounded-full bg-white shadow transition-transform",
                                      button.is_enabled === false ? "left-0.5" : "left-[18px]",
                                    )}
                                  />
                                </button>

                                {/* Delete — disabled for system buttons */}
                                {isSystemButton ? (
                                  <span
                                    title={
                                      button.managed_by_module === "home" ||
                                      button.managed_by_module === "contacts"
                                        ? "Дефолтную кнопку нельзя удалить"
                                        : `Чтобы удалить, выключите модуль «${buttonModule?.configureLabel ?? button.managed_by_module}»`
                                    }
                                    className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded text-neutral-300 cursor-not-allowed"
                                  >
                                    <Trash2 className="h-4 w-4" />
                                  </span>
                                ) : (
                                  <button
                                    type="button"
                                    onClick={() => removeButton(index)}
                                    disabled={isSaving}
                                    className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded text-neutral-400 hover:text-red-600 hover:bg-red-50 transition-colors disabled:opacity-50"
                                    aria-label={`Удалить кнопку ${index + 1}`}
                                  >
                                    <Trash2 className="h-4 w-4" />
                                  </button>
                                )}
                              </div>
                            </div>

                            {isExpanded && canExpand && isHomeButton && (
                              <div className="border-t border-neutral-200 bg-neutral-50/60 px-4 py-4">
                                <label
                                  htmlFor={`home-label-${index}`}
                                  className="block text-xs font-mono uppercase tracking-wider text-neutral-400 mb-1.5"
                                >
                                  Название кнопки в боте
                                </label>
                                <input
                                  id={`home-label-${index}`}
                                  type="text"
                                  value={button.label}
                                  onChange={(e) =>
                                    updateButton(index, "label", e.target.value)
                                  }
                                  maxLength={32}
                                  placeholder="Главная страница"
                                  className={inputClassName}
                                />
                                <p className="mt-2 text-xs text-neutral-500 leading-relaxed">
                                  Эта кнопка работает как /start — возвращает к приветственному сообщению. Меняется только подпись.
                                </p>
                              </div>
                            )}

                            {isExpanded && canExpand && isContactsButton && (
                              <div className="border-t border-neutral-200 bg-neutral-50/60 px-4 py-4 space-y-3">
                                <p className="text-xs text-neutral-500 leading-relaxed">
                                  Выберите, данные каких точек продаж бот пришлёт по нажатию на эту кнопку. Доступны только привязанные в «Подключении» POS.
                                </p>
                                {boundPosIds.length === 0 ? (
                                  <p className="text-sm text-neutral-500">
                                    В «Подключении» пока не выбрана ни одна точка продаж.
                                  </p>
                                ) : (
                                  <div className="space-y-2">
                                    {posLocations
                                      .filter((loc) => boundPosIds.includes(loc.id))
                                      .map((loc) => {
                                        const selected = settings.contacts_pos_ids?.includes(loc.id) ?? false
                                        return (
                                          <button
                                            key={loc.id}
                                            type="button"
                                            role="switch"
                                            aria-checked={selected}
                                            onClick={() =>
                                              setSettings((s) => {
                                                if (!s) return s
                                                const current = s.contacts_pos_ids ?? []
                                                const next = current.includes(loc.id)
                                                  ? current.filter((id) => id !== loc.id)
                                                  : [...current, loc.id]
                                                return { ...s, contacts_pos_ids: next }
                                              })
                                            }
                                            className={cn(
                                              "w-full flex items-center gap-3 rounded border bg-white px-3 py-2.5 text-left transition-colors",
                                              selected ? "border-accent/30 bg-accent/5" : "border-neutral-200 hover:border-neutral-300",
                                            )}
                                          >
                                            <span
                                              className={cn(
                                                "flex h-4 w-4 shrink-0 items-center justify-center rounded border transition-colors",
                                                selected ? "bg-accent border-accent" : "bg-white border-neutral-300",
                                              )}
                                            >
                                              {selected && <Check className="w-3 h-3 text-white" />}
                                            </span>
                                            <span className="min-w-0">
                                              <span className="block text-sm font-medium text-neutral-900">{loc.name}</span>
                                              {loc.address && (
                                                <span className="block text-xs text-neutral-500 truncate">{loc.address}</span>
                                              )}
                                            </span>
                                          </button>
                                        )
                                      })}
                                  </div>
                                )}
                              </div>
                            )}

                            {isExpanded && canExpand && !isHomeButton && !isContactsButton && (
                              <div className="border-t border-neutral-200 bg-neutral-50/60 px-4 py-4">
                                <div className="grid grid-cols-1 xl:grid-cols-[minmax(0,1.05fr)_minmax(340px,0.95fr)] gap-6 xl:gap-8 items-start">
                                  <div className="space-y-3">
                                    <div className="flex items-center gap-2">
                                      <EmojiPicker
                                        onSelect={(item) => {
                                          updateButton(
                                            index,
                                            "icon_custom_emoji_id",
                                            item.tg_custom_emoji_id || "",
                                          );
                                          updateButton(
                                            index,
                                            "icon_image_url",
                                            item.image_url,
                                          );
                                        }}
                                        triggerClassName="w-9 h-9 border border-neutral-200 bg-white hover:bg-neutral-50 shrink-0 rounded"
                                      >
                                        {button.icon_image_url ? (
                                          <img
                                            src={button.icon_image_url}
                                            alt=""
                                            className="w-5 h-5 rounded-sm object-cover"
                                          />
                                        ) : undefined}
                                      </EmojiPicker>
                                      <ButtonStylePicker
                                        value={button.style || ""}
                                        onChange={(style) =>
                                          updateButton(index, "style", style)
                                        }
                                      />
                                      {!isSystemButton ? (
                                        <input
                                          type="text"
                                          value={button.label}
                                          onChange={(e) =>
                                            updateButton(
                                              index,
                                              "label",
                                              e.target.value,
                                            )
                                          }
                                          placeholder="Текст кнопки в меню бота"
                                          disabled={isSaving}
                                          className={cn(
                                            inputClassName,
                                            "flex-1",
                                          )}
                                          aria-label={`Название кнопки ${index + 1}`}
                                        />
                                      ) : (
                                        <span className="text-sm font-medium text-neutral-700">
                                          {button.label}
                                        </span>
                                      )}
                                    </div>
                                    {isSystemButton && buttonModule && (
                                      <div className="text-xs text-amber-700 bg-amber-50 border border-amber-200 rounded px-3 py-2 mb-3">
                                        {buttonModule.managed_by_module ===
                                          "menu" &&
                                          "📋 Категории меню добавятся к сообщению автоматически"}
                                        {buttonModule.managed_by_module ===
                                          "booking" &&
                                          "📅 Выбор даты и времени добавится автоматически"}
                                        {buttonModule.managed_by_module ===
                                          "feedback" &&
                                          "💬 Форма обратной связи добавится автоматически"}
                                      </div>
                                    )}
                                    <Suspense fallback={<EditorFallback />}>
                                      <MessageContentEditor
                                        value={buttonContent}
                                        onChange={(content) =>
                                          updateButtonContent(index, content)
                                        }
                                        onUpload={campaignsApi.uploadFile}
                                        maxParts={5}
                                        placeholders={messagePlaceholders}
                                      />
                                    </Suspense>
                                  </div>

                                  <div className="xl:sticky xl:top-6 flex justify-center">
                                    <Suspense fallback={<PreviewFallback />}>
                                      <TelegramPreview
                                        botName={button.label || botName}
                                        content={buttonContent}
                                        showFrame
                                        className="w-full max-w-[360px]"
                                      />
                                    </Suspense>
                                  </div>
                                </div>
                              </div>
                            )}
                          </div>
                        );
                      }}
                    </SortableItem>
                  ))}
                </div>
              </SortableContext>
            </DndContext>
          )}

          {editableButtons.length >= 10 && (
            <p className="text-xs text-neutral-400 mt-3">Максимум 10 кнопок</p>
          )}
        </SectionBlock>

        <SectionBlock
          eyebrow="Контакты"
          title="Точки продаж для кнопки «Контакты»"
          description="Отметьте, какие точки продаж отправлять по кнопке «Контакты». Если ничего не выбрать, по умолчанию берём все привязанные к боту точки."
        >
          {boundPosIds.length === 0 ? (
            <p className="text-sm text-neutral-400">
              Сначала привяжите хотя бы одну точку продаж во вкладке
              «Подключение».
            </p>
          ) : (
            <div className="space-y-2">
              {posLocations
                .filter((location) => boundPosIds.includes(location.id))
                .map((location) => {
                  const isChecked = (settings.contacts_pos_ids ?? []).includes(
                    location.id,
                  );

                  return (
                    <label
                      key={location.id}
                      className={cn(
                        "flex items-center gap-3 rounded border p-3 transition-colors",
                        isChecked
                          ? "border-accent/25 bg-accent/5"
                          : "border-neutral-200 bg-neutral-50/70",
                      )}
                    >
                      <input
                        type="checkbox"
                        checked={isChecked}
                        onChange={() => {
                          setSettings((current) => {
                            if (!current) return current;
                            const currentIds = current.contacts_pos_ids ?? [];
                            return {
                              ...current,
                              contacts_pos_ids: currentIds.includes(location.id)
                                ? currentIds.filter((id) => id !== location.id)
                                : [...currentIds, location.id],
                            };
                          });
                        }}
                        className="h-4 w-4 rounded border-neutral-300 text-accent focus:ring-accent/20"
                      />
                      <div>
                        <div className="text-sm font-medium text-neutral-900">
                          {location.name}
                        </div>
                        <div className="text-xs text-neutral-500">
                          {location.address}
                        </div>
                      </div>
                    </label>
                  );
                })}
            </div>
          )}
        </SectionBlock>

        <SectionBlock
          eyebrow="Анкета"
          title="Поля регистрации"
          description="Эти поля бот попросит у гостя при первом запуске. Оставляйте только то, что действительно нужно для работы с клиентом."
          actions={
            <button
              type="button"
              onClick={addField}
              disabled={isSaving}
              className={cn(
                "inline-flex min-h-11 w-full items-center justify-center gap-1.5 rounded px-3 text-sm font-medium text-accent sm:w-auto",
                "hover:text-accent/80 transition-colors",
                "disabled:opacity-50 disabled:cursor-not-allowed",
              )}
            >
              <Plus className="w-4 h-4" />
              Добавить поле
            </button>
          }
        >
          <div className="flex flex-wrap gap-2 mb-4">
            {STANDARD_FIELD_PRESETS.map((preset) => (
              <button
                key={preset.name}
                type="button"
                onClick={() => addPreset(preset)}
                disabled={
                  isSaving ||
                  !canAddPreset(settings.registration_form, preset.name)
                }
                className={cn(
                  "inline-flex min-h-11 items-center rounded px-3 py-1.5 text-xs font-medium transition-colors",
                  "bg-neutral-100 text-neutral-600 hover:bg-neutral-200",
                  "disabled:opacity-50 disabled:cursor-not-allowed",
                )}
              >
                + {STANDARD_FIELD_PRESET_LABELS[preset.name] ?? preset.label}
              </button>
            ))}
          </div>

          {settings.registration_form.length === 0 ? (
            <p className="text-sm text-neutral-400 text-center py-6">
              Поля регистрации ещё не добавлены.
            </p>
          ) : (
            <DndContext
              sensors={formSensors}
              collisionDetection={closestCenter}
              onDragEnd={handleFieldDragEnd}
            >
              <SortableContext
                items={fieldIds}
                strategy={verticalListSortingStrategy}
              >
                <div className="space-y-3">
                  {settings.registration_form.map((field, index) => (
                    <SortableItem key={fieldIds[index]} id={fieldIds[index]}>
                      {({ listeners, attributes }) => (
                        <div className="flex items-start gap-3 p-4 rounded bg-white border border-neutral-200">
                          <button
                            type="button"
                            className="mt-2.5 inline-flex min-h-11 min-w-11 items-center justify-center rounded text-neutral-400 hover:text-neutral-600 hover:bg-neutral-100 cursor-grab active:cursor-grabbing touch-none"
                            {...listeners}
                            {...attributes}
                            aria-label="Перетащить"
                          >
                            <GripVertical className="w-4 h-4" />
                          </button>
                          <div className="flex-1 grid grid-cols-1 sm:grid-cols-2 gap-3">
                            {isStandardField(field) ? (
                              <div className="rounded border border-neutral-200 bg-neutral-50 px-4 py-2.5">
                                <div className="text-[11px] uppercase tracking-[0.18em] text-neutral-400">
                                  Внутреннее имя
                                </div>
                                <div className="mt-1 text-sm font-medium text-neutral-700">
                                  {field.name}
                                </div>
                              </div>
                            ) : (
                              <input
                                type="text"
                                value={field.name}
                                onChange={(e) =>
                                  updateField(index, "name", e.target.value)
                                }
                                placeholder="Внутреннее имя, например favorite_drink"
                                disabled={isSaving}
                                className={inputClassName}
                                aria-label={`Ключ поля ${index + 1}`}
                              />
                            )}
                            <input
                              type="text"
                              value={field.label}
                              onChange={(e) =>
                                updateField(index, "label", e.target.value)
                              }
                              placeholder="Как это поле увидит гость"
                              disabled={isSaving}
                              className={inputClassName}
                              aria-label={`Название поля ${index + 1}`}
                            />
                            <CustomSelect
                              value={field.type}
                              onChange={(v) => updateField(index, "type", v)}
                              options={FIELD_TYPE_OPTIONS}
                              disabled={isSaving}
                              light
                            />
                            <label className="flex items-center gap-2 px-4 py-2.5">
                              <input
                                type="checkbox"
                                checked={field.required}
                                onChange={(e) =>
                                  updateField(
                                    index,
                                    "required",
                                    e.target.checked,
                                  )
                                }
                                disabled={isSaving}
                                className="h-4 w-4 rounded border-neutral-300 text-accent focus:ring-accent/20"
                              />
                              <span className="text-sm text-neutral-700">
                                Обязательно
                              </span>
                            </label>
                            {field.type === "select" && (
                              <div className="col-span-2">
                                <input
                                  type="text"
                                  value={(field.options ?? []).join(", ")}
                                  onChange={(e) => {
                                    const opts = e.target.value
                                      .split(",")
                                      .map((s) => s.trim())
                                      .filter(Boolean);
                                    setSettings((s) => {
                                      if (!s) return s;
                                      const updated = s.registration_form.map(
                                        (f, i) =>
                                          i === index
                                            ? { ...f, options: opts }
                                            : f,
                                      );
                                      return {
                                        ...s,
                                        registration_form: updated,
                                      };
                                    });
                                  }}
                                  placeholder="Варианты через запятую, например: Москва, Санкт-Петербург"
                                  disabled={isSaving}
                                  className={inputClassName}
                                />
                                <p className="text-[11px] text-neutral-400 mt-1">
                                  Гость увидит эти варианты как кнопки в боте.
                                </p>
                              </div>
                            )}
                            {(field.type === "email" ||
                              field.type === "phone" ||
                              field.type === "date") && (
                              <p className="col-span-2 text-[11px] text-neutral-400">
                                {field.type === "email" &&
                                  "Бот проверит, что гость ввёл корректный email."}
                                {field.type === "phone" &&
                                  "Бот проверит, что номер телефона введён в понятном формате."}
                                {field.type === "date" &&
                                  "Бот подскажет нужный формат даты: ДД.ММ.ГГГГ."}
                              </p>
                            )}
                            {field.name === "gender" && (
                              <p className="col-span-2 text-[11px] text-neutral-400">
                                Пол добавлен как стандартное поле. Внутреннее
                                имя фиксировано и не редактируется.
                              </p>
                            )}
                          </div>
                          <button
                            type="button"
                            onClick={() => removeField(index)}
                            disabled={isSaving}
                            className="mt-2.5 inline-flex min-h-11 min-w-11 items-center justify-center rounded text-neutral-500 hover:text-red-600 hover:bg-red-50 transition-colors disabled:opacity-50"
                            aria-label={`Удалить поле ${index + 1}`}
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      )}
                    </SortableItem>
                  ))}
                </div>
              </SortableContext>
            </DndContext>
          )}
        </SectionBlock>

      <SaveButton
        isSaving={isSaving}
        saveError={saveError}
        saveSuccess={saveSuccess}
        onSave={save}
      />

    </div>
  );
});

// ===========================================================================
// Tab: Модули (Modules)
// ===========================================================================

const ModulesTab = memo(function ModulesTab({
  botId,
  selectedProgramId,
  settings,
  setSettings,
  boundPosIds,
  posLocations,
}: {
  botId: number;
  selectedProgramId?: number;
  settings: BotSettings;
  setSettings: React.Dispatch<React.SetStateAction<BotSettings | null>>;
  boundPosIds: number[];
  posLocations: POSLocation[];
}) {
  const { isSaving, saveError, saveSuccess, save } = useSaveAction(botId, () =>
    botsApi.updateSettings(botId, {
      modules: settings.modules,
      module_configs: settings.module_configs,
    }),
  );

  const toggleModule = (moduleKey: string) => {
    setSettings((s) => {
      if (!s) return s;
      const isEnabled = s.modules.includes(moduleKey);
      const modules = s.modules.includes(moduleKey)
        ? s.modules.filter((m) => m !== moduleKey)
        : [...s.modules, moduleKey];
      return {
        ...s,
        buttons: syncSystemButtons(s.buttons, modules),
        modules,
        module_configs: isEnabled
          ? s.module_configs
          : withModuleDefaults(
              moduleKey,
              s.module_configs,
              posLocations,
              boundPosIds,
            ),
      };
    });
  };

  return (
    <div className="space-y-6">
      <section className="bg-white rounded border border-neutral-900 p-4 sm:p-5 md:p-6">
        <div className="mb-5">
          <p className="font-mono text-[10px] uppercase tracking-wider text-neutral-300 mb-1">Функционал</p>
          <h3 className="text-sm font-semibold text-neutral-700">Модули бота</h3>
        </div>

        <div className="grid gap-3 sm:grid-cols-2">
        {MODULE_DEFS.map((mod) => {
          const isActive = settings.modules.includes(mod.key);
          const configHref = moduleConfigHref(
            mod.key,
            botId,
            selectedProgramId,
          );
          return (
            <div
              key={mod.key}
              className={cn(
                "p-4 rounded border transition-colors",
                isActive
                  ? "border-accent/30 bg-accent/5"
                  : "border-neutral-200 bg-neutral-50",
              )}
            >
              <div className="flex items-center justify-between">
                <div className="min-w-0 mr-3">
                  <p className="text-sm font-medium text-neutral-900">
                    {mod.label}
                  </p>
                  <p className="text-xs text-neutral-500 mt-0.5">
                    {mod.description}
                  </p>
                </div>
                <button
                  type="button"
                  role="switch"
                  aria-checked={isActive}
                  aria-label={`${mod.label} модуль`}
                  onClick={() => toggleModule(mod.key)}
                  disabled={isSaving}
                  className={cn(
                    "relative shrink-0 w-10 h-6 rounded-full transition-colors",
                    "focus:outline-none focus:ring-2 focus:ring-accent/20",
                    "disabled:opacity-50 disabled:cursor-not-allowed",
                    isActive ? "bg-accent" : "bg-neutral-300",
                  )}
                >
                  <span
                    className={cn(
                      "absolute top-0.5 left-0.5 w-5 h-5 rounded-full bg-white shadow transition-transform",
                      isActive && "translate-x-4",
                    )}
                  />
                </button>
              </div>
              <Link
                to={configHref}
                aria-label={`Настроить ${mod.label}`}
                className="inline-flex min-h-11 items-center gap-1.5 mt-3 rounded text-xs text-accent hover:text-accent/80 font-medium transition-colors"
              >
                <SlidersHorizontal className="w-3.5 h-3.5" />
                Настроить <ChevronRight className="w-3 h-3" />
              </Link>
            </div>
          );
        })}
        </div>
      </section>

      <SaveButton
        isSaving={isSaving}
        saveError={saveError}
        saveSuccess={saveSuccess}
        onSave={save}
      />
    </div>
  );
});

function moduleConfigHref(
  moduleKey: string,
  botId: number,
  selectedProgramId?: number,
): string {
  if (moduleKey === "menu") {
    return `/dashboard/bots/${botId}/menu`;
  }
  if (moduleKey === "loyalty") {
    return selectedProgramId
      ? `/dashboard/loyalty/${selectedProgramId}?botId=${botId}`
      : `/dashboard/loyalty?botId=${botId}`;
  }
  if (moduleKey === "booking") {
    return `/dashboard/bots/${botId}/booking`;
  }
  if (moduleKey === "feedback") {
    return `/dashboard/bots/${botId}/feedback`;
  }
  return `/dashboard/bots/${botId}?tab=modules`;
}
