"use client"

import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import type { Address } from "@/lib/types"

const addressSchema = z.object({
  fullName: z.string().min(2, "Full name is required"),
  line1: z.string().min(1, "Address line 1 is required"),
  line2: z.string(),
  city: z.string().min(1, "City is required"),
  state: z.string().min(1, "State is required"),
  postalCode: z.string().min(1, "Postal code is required"),
  country: z.string().min(1, "Country is required"),
})

type AddressFormData = z.infer<typeof addressSchema>

interface AddressStepProps {
  defaultValues?: Partial<Address>
  onSubmit: (address: Address) => void
}

export function AddressStep({ defaultValues, onSubmit }: AddressStepProps) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<AddressFormData>({
    resolver: zodResolver(addressSchema),
    defaultValues: {
      fullName: defaultValues?.fullName ?? "",
      line1: defaultValues?.line1 ?? "",
      line2: defaultValues?.line2 ?? "",
      city: defaultValues?.city ?? "",
      state: defaultValues?.state ?? "",
      postalCode: defaultValues?.postalCode ?? "",
      country: defaultValues?.country ?? "US",
    },
  })

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" data-testid="address-form">
      <div className="space-y-1">
        <Label htmlFor="fullName">Full Name</Label>
        <Input id="fullName" {...register("fullName")} aria-invalid={!!errors.fullName} />
        {errors.fullName && <p className="text-sm text-destructive">{errors.fullName.message}</p>}
      </div>

      <div className="space-y-1">
        <Label htmlFor="line1">Address Line 1</Label>
        <Input id="line1" {...register("line1")} aria-invalid={!!errors.line1} />
        {errors.line1 && <p className="text-sm text-destructive">{errors.line1.message}</p>}
      </div>

      <div className="space-y-1">
        <Label htmlFor="line2">Address Line 2 (optional)</Label>
        <Input id="line2" {...register("line2")} />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-1">
          <Label htmlFor="city">City</Label>
          <Input id="city" {...register("city")} aria-invalid={!!errors.city} />
          {errors.city && <p className="text-sm text-destructive">{errors.city.message}</p>}
        </div>
        <div className="space-y-1">
          <Label htmlFor="state">State</Label>
          <Input id="state" {...register("state")} aria-invalid={!!errors.state} />
          {errors.state && <p className="text-sm text-destructive">{errors.state.message}</p>}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-1">
          <Label htmlFor="postalCode">Postal Code</Label>
          <Input id="postalCode" {...register("postalCode")} aria-invalid={!!errors.postalCode} />
          {errors.postalCode && <p className="text-sm text-destructive">{errors.postalCode.message}</p>}
        </div>
        <div className="space-y-1">
          <Label htmlFor="country">Country</Label>
          <Input id="country" {...register("country")} aria-invalid={!!errors.country} />
          {errors.country && <p className="text-sm text-destructive">{errors.country.message}</p>}
        </div>
      </div>

      <Button type="submit" size="lg" className="w-full">
        Continue to Payment
      </Button>
    </form>
  )
}
