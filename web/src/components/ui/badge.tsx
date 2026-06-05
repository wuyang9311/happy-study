import * as React from "react"

import { cn } from "../../lib/utils"

function Badge({ className, ...props }: React.ComponentProps<"span">) {
  return (
    <span data-slot="badge" className={cn("focus-visible:border-ring focus-visible:ring-ring/50 inline-flex w-fit shrink-0 items-center justify-center gap-1 overflow-hidden rounded-full border px-2 py-0.5 text-xs font-medium whitespace-nowrap transition-[color,box-shadow] focus-visible:ring-[3px]", className)} {...props} />
  )
}

export { Badge }
