import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Navigate, useLocation, useNavigate } from 'react-router-dom';
import { Lock, Mail, ShoppingCart } from 'lucide-react';
import { useState } from 'react';
import { Button, Input } from '@/components/ui';
import { applyServerErrors } from '@/lib/form';
import { useAuth } from '../hooks/useAuth';
import { defaultPathForRole } from '@/routes/menu';

const schema = z.object({
  email: z.string().min(1, 'Email wajib diisi').email('Format email tidak valid'),
  password: z.string().min(1, 'Password wajib diisi'),
});
type FormValues = z.infer<typeof schema>;

export default function LoginPage() {
  const { login, status, user } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [formError, setFormError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: { email: '', password: '' } });

  if (status === 'authenticated' && user) return <Navigate to={defaultPathForRole(user.role.code)} replace />;

  const onSubmit = handleSubmit(async (values) => {
    setFormError(null);
    try {
      const u = await login(values);
      const from = (location.state as { from?: string } | null)?.from;
      navigate(from && from !== '/login' ? from : defaultPathForRole(u.role.code), { replace: true });
    } catch (err) {
      setFormError(applyServerErrors(err, setError, ['email', 'password']));
    }
  });

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-4 py-10">
      <div className="w-full max-w-sm">
        <div className="mb-8 flex flex-col items-center text-center">
          <span className="mb-4 flex h-16 w-16 items-center justify-center rounded-xl bg-primary text-white shadow-card">
            <ShoppingCart className="h-8 w-8" />
          </span>
          <h1 className="text-large-title">POS UMKM</h1>
          <p className="mt-1 text-subheadline text-text-secondary">Masuk ke dashboard toko Anda</p>
        </div>

        <form onSubmit={onSubmit} className="space-y-4 rounded-xl bg-card p-6 shadow-card" noValidate>
          {formError && (
            <div className="rounded-md bg-danger/10 px-3 py-2.5 text-footnote text-danger" role="alert">
              {formError}
            </div>
          )}
          <Input
            label="Email"
            type="email"
            autoComplete="email"
            autoFocus
            placeholder="nama@toko.id"
            leftIcon={<Mail className="h-4 w-4" />}
            error={errors.email?.message}
            {...register('email')}
          />
          <Input
            label="Password"
            type="password"
            autoComplete="current-password"
            placeholder="••••••••"
            leftIcon={<Lock className="h-4 w-4" />}
            error={errors.password?.message}
            {...register('password')}
          />
          <Button type="submit" block size="lg" loading={isSubmitting}>
            Masuk
          </Button>
        </form>
        <p className="mt-6 text-center text-footnote text-text-secondary">
          Demo: owner@toko.id / owner123
        </p>
      </div>
    </div>
  );
}
