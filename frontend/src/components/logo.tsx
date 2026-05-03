import { useTranslation } from "react-i18next"

import { cn } from "@/lib/utils"

interface LogoProps {
  className?: string
}

export function Logo({ className }: LogoProps) {
  const { t } = useTranslation()
  return <p className={cn("font-bold", className)}>{t("common.brandName")}</p>
}
