import { Logo } from "@/components/logo"
import { Spinner } from "@/components/ui/spinner"

export function SplashScreen() {
  return (
    <div className="flex h-screen w-full flex-col items-center justify-center gap-4">
      <Logo className="text-2xl" />
      <Spinner className="size-6 text-muted-foreground" />
    </div>
  )
}
