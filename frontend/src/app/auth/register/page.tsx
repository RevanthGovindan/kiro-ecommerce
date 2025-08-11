import { Metadata } from 'next';
import { RegisterForm } from '@/components/auth/RegisterForm';

export const metadata: Metadata = {
  title: 'Sign Up - Ecommerce Store',
  description: 'Create an account to start shopping and track your orders.',
};

export default function RegisterPage() {
  return (
    <div className="min-h-screen bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <RegisterForm />
    </div>
  );
}