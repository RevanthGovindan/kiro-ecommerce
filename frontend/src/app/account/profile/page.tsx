import { Metadata } from 'next';
import { ProfileForm } from '@/components/profile/ProfileForm';

export const metadata: Metadata = {
  title: 'Profile - Ecommerce Store',
  description: 'Manage your account information and preferences.',
};

export default function ProfilePage() {
  return (
    <div className="min-h-screen bg-gray-50 py-8 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Account Profile</h1>
          <p className="mt-2 text-gray-600">
            Manage your personal information and account settings.
          </p>
        </div>
        <ProfileForm />
      </div>
    </div>
  );
}