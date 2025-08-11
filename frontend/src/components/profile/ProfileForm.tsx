'use client';

import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { useAuth } from '@/contexts/AuthContext';
import { apiClient, UpdateProfileRequest, ChangePasswordRequest } from '@/lib/api';

export const ProfileForm: React.FC = () => {
  const { user, updateUser } = useAuth();
  const [activeTab, setActiveTab] = useState<'profile' | 'password'>('profile');
  
  // Profile form state
  const [profileData, setProfileData] = useState({
    firstName: '',
    lastName: '',
    phone: '',
  });
  const [profileErrors, setProfileErrors] = useState<Record<string, string>>({});
  const [isProfileLoading, setIsProfileLoading] = useState(false);
  const [profileSuccess, setProfileSuccess] = useState('');

  // Password form state
  const [passwordData, setPasswordData] = useState({
    currentPassword: '',
    newPassword: '',
    confirmPassword: '',
  });
  const [passwordErrors, setPasswordErrors] = useState<Record<string, string>>({});
  const [isPasswordLoading, setIsPasswordLoading] = useState(false);
  const [passwordSuccess, setPasswordSuccess] = useState('');

  useEffect(() => {
    if (user) {
      setProfileData({
        firstName: user.firstName || '',
        lastName: user.lastName || '',
        phone: user.phone || '',
      });
    }
  }, [user]);

  const validateProfileForm = () => {
    const newErrors: Record<string, string> = {};

    if (!profileData.firstName.trim()) {
      newErrors.firstName = 'First name is required';
    }

    if (!profileData.lastName.trim()) {
      newErrors.lastName = 'Last name is required';
    }

    if (profileData.phone && !/^\+?[\d\s\-\(\)]+$/.test(profileData.phone)) {
      newErrors.phone = 'Phone number is invalid';
    }

    setProfileErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const validatePasswordForm = () => {
    const newErrors: Record<string, string> = {};

    if (!passwordData.currentPassword) {
      newErrors.currentPassword = 'Current password is required';
    }

    if (!passwordData.newPassword) {
      newErrors.newPassword = 'New password is required';
    } else if (passwordData.newPassword.length < 6) {
      newErrors.newPassword = 'Password must be at least 6 characters';
    }

    if (!passwordData.confirmPassword) {
      newErrors.confirmPassword = 'Please confirm your new password';
    } else if (passwordData.newPassword !== passwordData.confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match';
    }

    setPasswordErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleProfileSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateProfileForm()) {
      return;
    }

    setIsProfileLoading(true);
    setProfileErrors({});
    setProfileSuccess('');

    try {
      const updateData: UpdateProfileRequest = {
        firstName: profileData.firstName,
        lastName: profileData.lastName,
        phone: profileData.phone || undefined,
      };
      
      const updatedUser = await apiClient.updateProfile(updateData);
      updateUser(updatedUser);
      setProfileSuccess('Profile updated successfully!');
    } catch (error) {
      console.error('Profile update failed:', error);
      setProfileErrors({ 
        general: error instanceof Error ? error.message : 'Profile update failed. Please try again.' 
      });
    } finally {
      setIsProfileLoading(false);
    }
  };

  const handlePasswordSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validatePasswordForm()) {
      return;
    }

    setIsPasswordLoading(true);
    setPasswordErrors({});
    setPasswordSuccess('');

    try {
      const changeData: ChangePasswordRequest = {
        currentPassword: passwordData.currentPassword,
        newPassword: passwordData.newPassword,
      };
      
      await apiClient.changePassword(changeData);
      setPasswordSuccess('Password changed successfully!');
      setPasswordData({
        currentPassword: '',
        newPassword: '',
        confirmPassword: '',
      });
    } catch (error) {
      console.error('Password change failed:', error);
      setPasswordErrors({ 
        general: error instanceof Error ? error.message : 'Password change failed. Please try again.' 
      });
    } finally {
      setIsPasswordLoading(false);
    }
  };

  const handleProfileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setProfileData(prev => ({ ...prev, [name]: value }));
    
    // Clear error when user starts typing
    if (profileErrors[name]) {
      setProfileErrors(prev => ({ ...prev, [name]: '' }));
    }
    
    // Clear success message
    if (profileSuccess) {
      setProfileSuccess('');
    }
  };

  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setPasswordData(prev => ({ ...prev, [name]: value }));
    
    // Clear error when user starts typing
    if (passwordErrors[name]) {
      setPasswordErrors(prev => ({ ...prev, [name]: '' }));
    }
    
    // Clear success message
    if (passwordSuccess) {
      setPasswordSuccess('');
    }
  };

  if (!user) {
    return (
      <div className="text-center py-8">
        <p className="text-gray-600">Please log in to view your profile.</p>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto">
      <div className="bg-white rounded-lg shadow-md">
        {/* Tab Navigation */}
        <div className="border-b border-gray-200">
          <nav className="flex">
            <button
              onClick={() => setActiveTab('profile')}
              className={`px-6 py-3 text-sm font-medium border-b-2 ${
                activeTab === 'profile'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              Profile Information
            </button>
            <button
              onClick={() => setActiveTab('password')}
              className={`px-6 py-3 text-sm font-medium border-b-2 ${
                activeTab === 'password'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              Change Password
            </button>
          </nav>
        </div>

        <div className="p-6">
          {activeTab === 'profile' && (
            <div>
              <h2 className="text-xl font-semibold mb-6">Profile Information</h2>
              
              {profileSuccess && (
                <div className="mb-4 p-3 bg-green-100 border border-green-400 text-green-700 rounded">
                  {profileSuccess}
                </div>
              )}
              
              {profileErrors.general && (
                <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
                  {profileErrors.general}
                </div>
              )}

              <form onSubmit={handleProfileSubmit} className="space-y-4">
                <div>
                  <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
                    Email Address
                  </label>
                  <Input
                    id="email"
                    type="email"
                    value={user.email}
                    disabled
                    className="bg-gray-50"
                  />
                  <p className="mt-1 text-sm text-gray-500">Email cannot be changed</p>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label htmlFor="firstName" className="block text-sm font-medium text-gray-700 mb-1">
                      First Name
                    </label>
                    <Input
                      id="firstName"
                      name="firstName"
                      type="text"
                      value={profileData.firstName}
                      onChange={handleProfileChange}
                      placeholder="First name"
                      className={profileErrors.firstName ? 'border-red-500' : ''}
                      disabled={isProfileLoading}
                    />
                    {profileErrors.firstName && (
                      <p className="mt-1 text-sm text-red-600">{profileErrors.firstName}</p>
                    )}
                  </div>

                  <div>
                    <label htmlFor="lastName" className="block text-sm font-medium text-gray-700 mb-1">
                      Last Name
                    </label>
                    <Input
                      id="lastName"
                      name="lastName"
                      type="text"
                      value={profileData.lastName}
                      onChange={handleProfileChange}
                      placeholder="Last name"
                      className={profileErrors.lastName ? 'border-red-500' : ''}
                      disabled={isProfileLoading}
                    />
                    {profileErrors.lastName && (
                      <p className="mt-1 text-sm text-red-600">{profileErrors.lastName}</p>
                    )}
                  </div>
                </div>

                <div>
                  <label htmlFor="phone" className="block text-sm font-medium text-gray-700 mb-1">
                    Phone Number
                  </label>
                  <Input
                    id="phone"
                    name="phone"
                    type="tel"
                    value={profileData.phone}
                    onChange={handleProfileChange}
                    placeholder="Enter your phone number"
                    className={profileErrors.phone ? 'border-red-500' : ''}
                    disabled={isProfileLoading}
                  />
                  {profileErrors.phone && (
                    <p className="mt-1 text-sm text-red-600">{profileErrors.phone}</p>
                  )}
                </div>

                <Button
                  type="submit"
                  disabled={isProfileLoading}
                  className="w-full sm:w-auto"
                >
                  {isProfileLoading ? 'Updating...' : 'Update Profile'}
                </Button>
              </form>
            </div>
          )}

          {activeTab === 'password' && (
            <div>
              <h2 className="text-xl font-semibold mb-6">Change Password</h2>
              
              {passwordSuccess && (
                <div className="mb-4 p-3 bg-green-100 border border-green-400 text-green-700 rounded">
                  {passwordSuccess}
                </div>
              )}
              
              {passwordErrors.general && (
                <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
                  {passwordErrors.general}
                </div>
              )}

              <form onSubmit={handlePasswordSubmit} className="space-y-4">
                <div>
                  <label htmlFor="currentPassword" className="block text-sm font-medium text-gray-700 mb-1">
                    Current Password
                  </label>
                  <Input
                    id="currentPassword"
                    name="currentPassword"
                    type="password"
                    value={passwordData.currentPassword}
                    onChange={handlePasswordChange}
                    placeholder="Enter your current password"
                    className={passwordErrors.currentPassword ? 'border-red-500' : ''}
                    disabled={isPasswordLoading}
                  />
                  {passwordErrors.currentPassword && (
                    <p className="mt-1 text-sm text-red-600">{passwordErrors.currentPassword}</p>
                  )}
                </div>

                <div>
                  <label htmlFor="newPassword" className="block text-sm font-medium text-gray-700 mb-1">
                    New Password
                  </label>
                  <Input
                    id="newPassword"
                    name="newPassword"
                    type="password"
                    value={passwordData.newPassword}
                    onChange={handlePasswordChange}
                    placeholder="Enter your new password"
                    className={passwordErrors.newPassword ? 'border-red-500' : ''}
                    disabled={isPasswordLoading}
                  />
                  {passwordErrors.newPassword && (
                    <p className="mt-1 text-sm text-red-600">{passwordErrors.newPassword}</p>
                  )}
                </div>

                <div>
                  <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700 mb-1">
                    Confirm New Password
                  </label>
                  <Input
                    id="confirmPassword"
                    name="confirmPassword"
                    type="password"
                    value={passwordData.confirmPassword}
                    onChange={handlePasswordChange}
                    placeholder="Confirm your new password"
                    className={passwordErrors.confirmPassword ? 'border-red-500' : ''}
                    disabled={isPasswordLoading}
                  />
                  {passwordErrors.confirmPassword && (
                    <p className="mt-1 text-sm text-red-600">{passwordErrors.confirmPassword}</p>
                  )}
                </div>

                <Button
                  type="submit"
                  disabled={isPasswordLoading}
                  className="w-full sm:w-auto"
                >
                  {isPasswordLoading ? 'Changing Password...' : 'Change Password'}
                </Button>
              </form>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};