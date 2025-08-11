'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { apiClient } from '@/lib/api';

export const ForgotPasswordForm: React.FC = () => {
  const [email, setEmail] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [isLoading, setIsLoading] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);

  const validateForm = () => {
    const newErrors: Record<string, string> = {};

    if (!email) {
      newErrors.email = 'Email is required';
    } else if (!/\S+@\S+\.\S+/.test(email)) {
      newErrors.email = 'Email is invalid';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    setIsLoading(true);
    setErrors({});

    try {
      await apiClient.forgotPassword({ email });
      setIsSuccess(true);
    } catch (error) {
      console.error('Forgot password failed:', error);
      setErrors({ 
        general: error instanceof Error ? error.message : 'Failed to send reset email. Please try again.' 
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(e.target.value);
    
    // Clear error when user starts typing
    if (errors.email) {
      setErrors(prev => ({ ...prev, email: '' }));
    }
  };

  if (isSuccess) {
    return (
      <div className="max-w-md mx-auto">
        <div className="bg-white p-8 rounded-lg shadow-md text-center">
          <div className="mb-4">
            <svg className="w-16 h-16 text-green-500 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <h2 className="text-2xl font-bold mb-4">Check Your Email</h2>
          <p className="text-gray-600 mb-6">
            We&apos;ve sent a password reset link to <strong>{email}</strong>. 
            Please check your email and follow the instructions to reset your password.
          </p>
          <div className="space-y-3">
            <Link href="/auth/login">
              <Button className="w-full">Back to Login</Button>
            </Link>
            <button
              onClick={() => {
                setIsSuccess(false);
                setEmail('');
              }}
              className="w-full text-sm text-blue-600 hover:text-blue-500"
            >
              Try a different email
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-md mx-auto">
      <div className="bg-white p-8 rounded-lg shadow-md">
        <h2 className="text-2xl font-bold text-center mb-6">Reset Your Password</h2>
        <p className="text-gray-600 text-center mb-6">
          Enter your email address and we&apos;ll send you a link to reset your password.
        </p>
        
        {errors.general && (
          <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
            {errors.general}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
              Email Address
            </label>
            <Input
              id="email"
              name="email"
              type="email"
              value={email}
              onChange={handleChange}
              placeholder="Enter your email"
              className={errors.email ? 'border-red-500' : ''}
              disabled={isLoading}
            />
            {errors.email && (
              <p className="mt-1 text-sm text-red-600">{errors.email}</p>
            )}
          </div>

          <Button
            type="submit"
            className="w-full"
            disabled={isLoading}
          >
            {isLoading ? 'Sending...' : 'Send Reset Link'}
          </Button>
        </form>

        <div className="mt-6 text-center">
          <Link
            href="/auth/login"
            className="text-sm text-blue-600 hover:text-blue-500"
          >
            Back to Login
          </Link>
        </div>
      </div>
    </div>
  );
};