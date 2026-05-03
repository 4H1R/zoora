import type { NavFn } from "@/lib/data-table"

import { useNavigate, useSearch } from "@tanstack/react-router"

import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"

export type StatusTab = {
  value: string
  label: string
  count?: number
}

export function StatusTabs({ tabs }: { tabs: StatusTab[] }) {
  const { status } = useSearch({ strict: false }) as { status?: string }
  const navigate = useNavigate() as unknown as NavFn

  return (
    <div className="hide-scrollbar overflow-x-auto border-b">
      <Tabs
        value={status ?? "all"}
        onValueChange={(v) =>
          navigate({ search: (prev) => ({ ...prev, status: v === "all" ? undefined : v, page: 1 }) })
        }
      >
        <TabsList variant="line">
          {tabs.map((tab) => (
            <TabsTrigger key={tab.value} value={tab.value}>
              {tab.label}
              {tab.count != null && (
                <span className="bg-muted text-muted-foreground ms-1.5 rounded px-1.5 py-0.5 text-[11px] tabular-nums">
                  {tab.count}
                </span>
              )}
            </TabsTrigger>
          ))}
        </TabsList>
      </Tabs>
    </div>
  )
}
