import { Metadata } from 'next';
import { LoginForm } from '@/components/auth/LoginForm';

export const metadata: Metadata = {
  title: 'Login - Ecommerce Store',
  description: 'Login to your account to access your orders and profile.',
};

export default function LoginPage() {
  return (
    <div className="min-h-screen bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <LoginForm />
    </div>
  );
}