import {
  BookOpenIcon,
  CalendarCheckIcon,
  ClipboardListIcon,
  GraduationCapIcon,
} from "lucide-react"
import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"

import type { DashboardTileSpec } from "./tile-grid"

export function useStudentTiles(orgId: string): DashboardTileSpec[] {
  const { can } = useAccess()
  const { t } = useTranslation()

  const tiles: DashboardTileSpec[] = []
  if (can("quizzes:take") || can("quizzes:view")) {
    tiles.push({
      key: "exams",
      label: t("org.portal.tiles.exams"),
      icon: <ClipboardListIcon />,
      to: "/org/$orgId/exams",
      params: { orgId },
    })
  }
  if (can("gradebook:view_own")) {
    tiles.push({
      key: "grades",
      label: t("org.portal.tiles.grades"),
      icon: <GraduationCapIcon />,
      to: "/org/$orgId/grades",
      params: { orgId },
    })
  }
  if (can("attendance:view_own")) {
    tiles.push({
      key: "attendance",
      label: t("org.portal.tiles.attendance"),
      icon: <CalendarCheckIcon />,
      to: "/org/$orgId/attendance",
      params: { orgId },
    })
  }
  if (can("classes:view") || can("classes:view_any")) {
    tiles.push({
      key: "courses",
      label: t("org.portal.tiles.courses"),
      icon: <BookOpenIcon />,
      to: "/org/$orgId/classes",
      params: { orgId },
    })
  }
  return tiles
}
