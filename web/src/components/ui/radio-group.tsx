"use client"

import * as React from "react"
import { cn } from "@/lib/utils"

interface RadioGroupProps extends React.ComponentProps<"div"> {
  value?: string
  onValueChange?: (value: string) => void
}

function RadioGroup({ className, value, onValueChange, children, ...props }: RadioGroupProps) {
  return (
    <div
      role="radiogroup"
      className={cn("grid gap-2", className)}
      data-value={value}
      {...props}
    >
      {React.Children.map(children, (child) => {
        if (React.isValidElement<RadioGroupItemProps>(child) && child.type === RadioGroupItem) {
          return React.cloneElement(child, { _groupValue: value, _onGroupChange: onValueChange })
        }
        return child
      })}
    </div>
  )
}

interface RadioGroupItemProps extends Omit<React.ComponentProps<"input">, "type"> {
  value: string
  _groupValue?: string
  _onGroupChange?: (value: string) => void
}

function RadioGroupItem({ className, value, _groupValue, _onGroupChange, ...props }: RadioGroupItemProps) {
  return (
    <input
      type="radio"
      className={cn(
        "aspect-square h-4 w-4 rounded-full border border-primary text-primary ring-offset-background focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
        className
      )}
      value={value}
      checked={_groupValue === value}
      onChange={() => _onGroupChange?.(value)}
      {...props}
    />
  )
}

export { RadioGroup, RadioGroupItem }
