"use client";

import type { Control } from "react-hook-form";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import type { CheckoutFormValues } from "../checkout-client";

interface PaymentSectionProps {
  control: Control<CheckoutFormValues>;
}

export function PaymentSection({ control }: PaymentSectionProps) {
  return (
    <div className="space-y-4">
      <h2 className="font-semibold text-base">Payment</h2>
      <FormField
        control={control}
        name="card_number"
        render={({ field }) => (
          <FormItem>
            <FormLabel htmlFor="card_number">Card Number</FormLabel>
            <FormControl>
              <Input
                id="card_number"
                placeholder="4242 4242 4242 4242"
                maxLength={19}
                inputMode="numeric"
                {...field}
                onChange={(e) => {
                  // Allow only digits and spaces
                  const val = e.target.value.replace(/[^\d\s]/g, "");
                  field.onChange(val);
                }}
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />
      <div className="grid grid-cols-2 gap-4">
        <FormField
          control={control}
          name="exp"
          render={({ field }) => (
            <FormItem>
              <FormLabel htmlFor="exp">Expiry (MM/YY)</FormLabel>
              <FormControl>
                <Input id="exp" placeholder="12/26" maxLength={5} {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={control}
          name="cvc"
          render={({ field }) => (
            <FormItem>
              <FormLabel htmlFor="cvc">CVC</FormLabel>
              <FormControl>
                <Input
                  id="cvc"
                  placeholder="123"
                  maxLength={4}
                  inputMode="numeric"
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      </div>
    </div>
  );
}
