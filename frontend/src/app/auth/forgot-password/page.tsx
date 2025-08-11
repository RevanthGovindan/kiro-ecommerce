import { Metadata } from 'next';
import { ForgotPasswordForm } from '@/components/auth/ForgotPasswordForm';

export const metadata: Metadata = {
  title: 'Forgot Password - Ecommerce Store',
  description: 'Reset your password to regain access to your account.',
};

export default function ForgotPasswordPage() {
  return (
    <div className="min-h-screen bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <ForgotPasswordForm />
    </div>
  );
}