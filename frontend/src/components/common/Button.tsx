import { cn } from '@/lib/utils'
import React from 'react'

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'dark'
  size?: 'sm' | 'md'
  asChild?: boolean
  leftIcon?: React.ReactNode
}

const base =
  'inline-flex items-center justify-center gap-2 rounded text-sm font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed focus:outline-none focus:ring-2 focus:ring-accent/20'

const variants: Record<NonNullable<ButtonProps['variant']>, string> = {
  primary: 'bg-accent text-white hover:bg-accent-hover active:bg-accent/80',
  secondary: 'bg-white border border-neutral-200 text-neutral-700 hover:bg-neutral-50',
  dark: 'bg-neutral-900 text-white hover:bg-neutral-700 active:bg-neutral-800',
}

const sizes: Record<NonNullable<ButtonProps['size']>, string> = {
  sm: 'py-1.5 px-3 text-xs',
  md: 'px-4 py-2.5',
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = 'primary', size = 'md', leftIcon, children, className, ...props }, ref) => {
    return (
      <button
        ref={ref}
        type="button"
        className={cn(base, variants[variant], sizes[size], className)}
        {...props}
      >
        {leftIcon}
        {children}
      </button>
    )
  },
)

Button.displayName = 'Button'
