import { Radar } from "lucide-react"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"

export function LoginForm({
  className,
  onSubmit,
  error,
  loading,
  ...props
}) {
  return (
    <div className={cn("flex flex-col gap-6 text-white", className)} {...props}>
      <form onSubmit={onSubmit}>
        <FieldGroup>
          <div className="flex flex-col items-center gap-2 text-center">
            <a href="#" className="flex flex-col items-center gap-2 font-medium">
              <div className="flex size-8 items-center justify-center rounded-md">
                <Radar className="size-6" />
              </div>
              <span className="sr-only">Nodara</span>
            </a>
            <h1 className="text-xl font-bold">Welcome to Nodara</h1>
            <FieldDescription className="text-white/60">
              Enter your credentials to access the console.
            </FieldDescription>
          </div>
          <Field>
            <FieldLabel htmlFor="email">Email</FieldLabel>
            <Input className="border-white/20 text-white placeholder:text-white/40 focus-visible:border-white focus-visible:ring-white/30" id="identifier" name="identifier" autoComplete="username" placeholder="admin@nodara.dev" required />
          </Field>
          <Field>
            <div className="flex items-center">
              <FieldLabel htmlFor="password">Password</FieldLabel>
              <a
                href="/forgot-password"
                className="ml-auto text-sm text-white/70 underline-offset-4 hover:text-white hover:underline">
                Forgot your password?
              </a>
            </div>
            <Input className="border-white/20 text-white placeholder:text-white/40 focus-visible:border-white focus-visible:ring-white/30" id="password" name="password" autoComplete="current-password" placeholder="••••••••" required />
          </Field>
          {error && <p className="text-sm text-red-300">{error}</p>}
          <Field>
            <Button className="bg-white text-black hover:bg-white/90" type="submit" disabled={loading}>{loading ? 'Signing in…' : 'Login'}</Button>
          </Field>
        </FieldGroup>
      </form>
    </div>
  );
}
