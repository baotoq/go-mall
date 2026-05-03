"use client";

import type { Control } from "react-hook-form";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import type { CheckoutFormValues } from "../checkout-client";

interface ShippingSectionProps {
  control: Control<CheckoutFormValues>;
}

export function ShippingSection({ control }: ShippingSectionProps) {
  return (
    <div className="space-y-4">
      <h2 className="font-semibold text-base">Shipping Address</h2>
      <FormField
        control={control}
        name="line1"
        render={({ field }) => (
          <FormItem>
            <FormLabel htmlFor="line1">Address Line 1</FormLabel>
            <FormControl>
              <Input id="line1" placeholder="123 Main St" {...field} />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />
      <FormField
        control={control}
        name="line2"
        render={({ field }) => (
          <FormItem>
            <FormLabel htmlFor="line2">Address Line 2 (optional)</FormLabel>
            <FormControl>
              <Input id="line2" placeholder="Apt, suite, etc." {...field} />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />
      <div className="grid grid-cols-2 gap-4">
        <FormField
          control={control}
          name="city"
          render={({ field }) => (
            <FormItem>
              <FormLabel htmlFor="city">City</FormLabel>
              <FormControl>
                <Input id="city" placeholder="City" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={control}
          name="state"
          render={({ field }) => (
            <FormItem>
              <FormLabel htmlFor="state">State</FormLabel>
              <FormControl>
                <Input id="state" placeholder="IL" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      </div>
      <div className="grid grid-cols-2 gap-4">
        <FormField
          control={control}
          name="postal_code"
          render={({ field }) => (
            <FormItem>
              <FormLabel htmlFor="postal_code">Postal Code</FormLabel>
              <FormControl>
                <Input id="postal_code" placeholder="62701" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={control}
          name="country"
          render={({ field }) => (
            <FormItem>
              <FormLabel htmlFor="country">Country</FormLabel>
              <FormControl>
                <Input id="country" placeholder="US" maxLength={2} {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      </div>
    </div>
  );
}
