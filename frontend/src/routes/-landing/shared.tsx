import type { ComponentProps, ReactNode } from "react"

import { motion, useReducedMotion } from "motion/react"

import { Eyebrow } from "@/components/eyebrow"
import { cn } from "@/lib/utils"

export const EASE_OUT = [0.22, 1, 0.36, 1] as const

export function scrollToSection(id: string) {
  document.getElementById(id)?.scrollIntoView({ behavior: "smooth", block: "start" })
}

interface RevealProps extends ComponentProps<typeof motion.div> {
  delay?: number
  y?: number
}

/** Scroll-triggered fade/rise/deblur — the landing page's one entrance gesture. */
export function Reveal({ delay = 0, y = 28, ...props }: RevealProps) {
  const reduced = useReducedMotion()
  return (
    <motion.div
      initial={reduced ? { opacity: 0 } : { opacity: 0, y, filter: "blur(6px)" }}
      whileInView={reduced ? { opacity: 1 } : { opacity: 1, y: 0, filter: "blur(0px)" }}
      viewport={{ once: true, margin: "-60px" }}
      transition={{ duration: 0.8, delay, ease: EASE_OUT }}
      {...props}
    />
  )
}

interface FloatProps {
  children: ReactNode
  className?: string
  delay?: number
  /** Vertical drift amplitude in px for the idle loop. */
  drift?: number
}

/** Pop-in once, then drift gently forever. For the hero's floating room cards. */
export function Float({ children, className, delay = 0, drift = 8 }: FloatProps) {
  const reduced = useReducedMotion()
  return (
    <motion.div
      className={className}
      initial={reduced ? { opacity: 0 } : { opacity: 0, y: 16, scale: 0.88 }}
      whileInView={reduced ? { opacity: 1 } : { opacity: 1, y: 0, scale: 1 }}
      viewport={{ once: true }}
      transition={{ delay, duration: 0.7, ease: EASE_OUT }}
    >
      <motion.div
        animate={reduced ? undefined : { y: [0, -drift, 0] }}
        transition={{ duration: 5.5, repeat: Infinity, ease: "easeInOut", delay }}
      >
        {children}
      </motion.div>
    </motion.div>
  )
}

interface SectionHeadingProps {
  eyebrow: string
  title: ReactNode
  subtitle?: string
  className?: string
}

export function SectionHeading({ eyebrow, title, subtitle, className }: SectionHeadingProps) {
  return (
    <Reveal className={cn("mx-auto flex max-w-2xl flex-col items-center text-center", className)}>
      <Eyebrow className="text-primary">{eyebrow}</Eyebrow>
      <h2 className="font-heading mt-4 text-3xl font-semibold tracking-tight text-balance sm:text-4xl lg:text-[2.75rem] lg:leading-[1.15]">
        {title}
      </h2>
      {subtitle ? (
        <p className="text-muted-foreground mt-4 max-w-xl text-base leading-relaxed text-pretty sm:text-lg">
          {subtitle}
        </p>
      ) : null}
    </Reveal>
  )
}
