"use client";

import { useState, useActionState } from "react";
import { Eye, EyeOff } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { login, getDemoCredentials } from "../lib/actions";
import z from "zod";

const loginSchema = z.object({
  email: z.string().email("Please enter a valid email"),
  password: z.string().min(1, "Password is required"),
});

type LoginFormData = z.infer<typeof loginSchema>;

export const Role = {
  ADMIN: "admin",
  MANAGER: "manager",
  TUTOR: "tutor",
  STUDENT: "student",
} as const;

export function LoginForm() {
  const [state, action, isPending] = useActionState(login, undefined);
  const [showPassword, setShowPassword] = useState(false);

  const handleDemoLogin = async (role: string) => {
    const credentials = await getDemoCredentials(role);
    if (!credentials) return;

    const emailInput = document.getElementById("email") as HTMLInputElement;
    const passwordInput = document.getElementById(
      "password",
    ) as HTMLInputElement;

    if (emailInput && passwordInput) {
      emailInput.value = credentials.email;
      passwordInput.value = credentials.password;

      // Trigger form submission
      const form = emailInput.closest("form");
      if (form) {
        form.requestSubmit();
      }
    }
  };

  return (
    <div className="space-y-6">
      <form action={action} className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="email">Email</Label>
          <Input
            id="email"
            name="email"
            type="email"
            placeholder="you@example.com"
            defaultValue=""
            required
          />
          {state?.errors?.email && (
            <p className="text-sm text-destructive">{state.errors.email[0]}</p>
          )}
        </div>

        <div className="space-y-2">
          <Label htmlFor="password">Password</Label>
          <div className="relative">
            <Input
              id="password"
              name="password"
              type={showPassword ? "text" : "password"}
              placeholder="••••••••"
              defaultValue=""
              required
            />
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
            >
              {showPassword ? (
                <EyeOff className="h-4 w-4" />
              ) : (
                <Eye className="h-4 w-4" />
              )}
            </button>
          </div>
          {state?.errors?.password && (
            <p className="text-sm text-destructive">
              {state.errors.password[0]}
            </p>
          )}
        </div>

        {state?.errors?.form && (
          <p className="text-sm text-destructive text-center">
            {state.errors.form[0]}
          </p>
        )}

        <Button type="submit" className="w-full" disabled={isPending}>
          {isPending ? "Signing in..." : "Sign In"}
        </Button>
      </form>

      {/* Demo credentials */}
      {process.env.NODE_ENV === "development" && (
        <div className="mt-6 pt-6 border-t">
          <p className="text-xs text-muted-foreground text-center mb-3">
            Demo Credentials
          </p>
          <div className="grid grid-cols-3 gap-3 text-xs">
            <button
              type="button"
              className="bg-muted rounded-lg p-3 text-left hover:bg-muted/80 transition-colors"
              onClick={() => handleDemoLogin(Role.ADMIN)}
            >
              <p className="font-medium text-red-500">Admin</p>
              <p className="text-muted-foreground mt-1">(Click to login)</p>
            </button>
            <button
              type="button"
              className="bg-muted rounded-lg p-3 text-left hover:bg-muted/80 transition-colors"
              onClick={() => handleDemoLogin(Role.MANAGER)}
            >
              <p className="font-medium text-emerald-500">Manager</p>
              <p className="text-muted-foreground mt-1">(Click to login)</p>
            </button>
            <button
              type="button"
              className="bg-muted rounded-lg p-3 text-left hover:bg-muted/80 transition-colors"
              onClick={() => handleDemoLogin(Role.TUTOR)}
            >
              <p className="font-medium text-primary">Tutor</p>
              <p className="text-muted-foreground mt-1">(Click to login)</p>
            </button>
            <button
              type="button"
              className="bg-muted rounded-lg p-3 text-left hover:bg-muted/80 transition-colors"
              onClick={() => handleDemoLogin(Role.STUDENT)}
            >
              <p className="font-medium text-primary">Student</p>
              <p className="text-muted-foreground mt-1">(Click to login)</p>
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
