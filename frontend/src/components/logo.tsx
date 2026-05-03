import { cn } from "@/lib/utils"

interface LogoProps {
  className?: string
}

export function Logo({ className }: LogoProps) {
  return <p className={cn("font-bold", className)}>Edu Connect</p>
}
