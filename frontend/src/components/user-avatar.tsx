import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { cn } from "@/lib/utils"

const avatarColors = [
  { bg: "bg-[#dcfce7]", text: "text-[#166534]" },
  { bg: "bg-[#dbeafe]", text: "text-[#1e3a8a]" },
  { bg: "bg-[#fce7f3]", text: "text-[#9d174d]" },
  { bg: "bg-[#fef3c7]", text: "text-[#78350f]" },
  { bg: "bg-[#e4e4e7]", text: "text-[#3f3f46]" },
  { bg: "bg-[#ddd6fe]", text: "text-[#4c1d95]" },
  { bg: "bg-[#fee2e2]", text: "text-[#991b1b]" },
] as const

function getColorFromName(name: string) {
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }
  return avatarColors[Math.abs(hash) % avatarColors.length]
}

export function getInitials(name: string | undefined): string {
  return name?.at(0)?.toUpperCase() ?? "?"
}

interface UserAvatarProps {
  name: string
  src?: string
  size?: "sm" | "md" | "lg"
  online?: boolean
  className?: string
}

const sizeClasses = {
  sm: "size-6 text-[9px]",
  md: "size-7 text-[11px]",
  lg: "size-9 text-[13px]",
}

export function UserAvatar({ name, src, size = "md", online, className }: UserAvatarProps) {
  const color = getColorFromName(name)
  const initials = getInitials(name)

  return (
    <div className="relative">
      <Avatar className={cn(sizeClasses[size], className)}>
        {src && <AvatarImage src={src} alt={name} />}
        <AvatarFallback className={cn(color.bg, color.text, "font-semibold")}>{initials}</AvatarFallback>
      </Avatar>
      {online !== undefined && (
        <span
          className={cn(
            "border-background absolute -end-0.5 -bottom-0.5 size-2 rounded-full border-2",
            online ? "bg-[var(--green-500)]" : "bg-muted-foreground"
          )}
        />
      )}
    </div>
  )
}

interface AvatarStackProps {
  users: { name: string; src?: string }[]
  max?: number
  size?: "sm" | "md"
}

export function AvatarStack({ users, max = 3, size = "sm" }: AvatarStackProps) {
  const visible = users.slice(0, max)
  const remaining = users.length - max

  return (
    <div className="flex -space-x-1.5">
      {visible.map((user) => (
        <UserAvatar
          key={user.name}
          name={user.name}
          src={user.src}
          size={size}
          className="border-background border-2"
        />
      ))}
      {remaining > 0 && (
        <Avatar className={cn(sizeClasses[size], "border-background border-2")}>
          <AvatarFallback className="bg-muted text-muted-foreground text-[8px] font-medium">
            +{remaining}
          </AvatarFallback>
        </Avatar>
      )}
    </div>
  )
}
