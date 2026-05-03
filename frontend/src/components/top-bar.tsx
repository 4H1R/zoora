import { Bell, Search } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { SidebarTrigger } from "@/components/ui/sidebar"
import { cn } from "@/lib/utils"

interface BreadcrumbItem {
  label: string
  active?: boolean
}

interface TopBarProps {
  breadcrumbs?: BreadcrumbItem[]
  actions?: React.ReactNode
}

export function TopBar({ breadcrumbs, actions }: TopBarProps) {
  const { t } = useTranslation()

  return (
    <header className="bg-background sticky top-0 z-10 flex h-14 items-center gap-4 border-b px-6">
      <SidebarTrigger className="-ms-2" />
      <Separator orientation="vertical" className="h-4" />

      {breadcrumbs && (
        <nav className="flex items-center gap-1.5 text-[13px] font-medium">
          {breadcrumbs.map((item, i) => (
            <span key={i} className="flex items-center gap-1.5">
              {i > 0 && <span className="text-muted-foreground/50">/</span>}
              <span className={cn(item.active ? "text-foreground" : "text-muted-foreground")}>{item.label}</span>
            </span>
          ))}
        </nav>
      )}

      <div className="bg-muted/50 text-muted-foreground ms-auto flex w-[280px] items-center gap-2 rounded-lg border px-2.5 py-1.5 text-[13px] lg:w-[320px]">
        <Search className="size-3.5" strokeWidth={1.75} />
        <span className="flex-1">{t("common.search")}</span>
        <kbd className="bg-background text-muted-foreground rounded border px-1.5 py-0.5 text-[11px]">⌘K</kbd>
      </div>

      {actions}

      <div className="relative">
        <Button variant="ghost" size="icon" className="size-8 rounded-lg">
          <Bell className="size-4" strokeWidth={1.75} />
        </Button>
        <span className="bg-primary border-background absolute -end-0.5 -top-0.5 size-[7px] rounded-full border-2" />
      </div>
    </header>
  )
}
