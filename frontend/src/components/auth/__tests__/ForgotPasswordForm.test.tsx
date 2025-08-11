import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ForgotPasswordForm } from '../ForgotPasswordForm';
import { apiClient } from '../../../lib/api';

// Mock the API client
jest.mock('../../../lib/api', () => ({
  apiClient: {
    forgotPassword: jest.fn(),
  },
}));

beforeEach(() => {
  jest.clearAllMocks();
});

describe('ForgotPasswordForm', () => {
  it('renders forgot password form with all required elements', () => {
    render(<ForgotPasswordForm />);
    
    expect(screen.getByText('Reset Your Password')).toBeInTheDocument();
    expect(screen.getByText('Enter your email address and we\'ll send you a link to reset your password.')).toBeInTheDocument();
    expect(screen.getByLabelText('Email Address')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Send Reset Link' })).toBeInTheDocument();
    expect(screen.getByText('Back to Login')).toBeInTheDocument();
  });

  it('validates required email field', async () => {
    render(<ForgotPasswordForm />);
    
    const submitButton = screen.getByRole('button', { name: 'Send Reset Link' });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Email is required')).toBeInTheDocument();
    });
    
    expect(apiClient.forgotPassword).not.toHaveBeenCalled();
  });

  it('validates email format', async () => {
    render(<ForgotPasswordForm />);
    
    const emailInput = screen.getByLabelText('Email Address');
    const submitButton = screen.getByRole('button', { name: 'Send Reset Link' });
    
    fireEvent.change(emailInput, { target: { value: 'invalid-email' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Email is invalid')).toBeInTheDocument();
    });
    
    expect(apiClient.forgotPassword).not.toHaveBeenCalled();
  });

  it('submits form with valid email', async () => {
    (apiClient.forgotPassword as jest.Mock).mockResolvedValue(undefined);
    
    render(<ForgotPasswordForm />);
    
    const emailInput = screen.getByLabelText('Email Address');
    const submitButton = screen.getByRole('button', { name: 'Send Reset Link' });
    
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(apiClient.forgotPassword).toHaveBeenCalledWith({
        email: 'test@example.com',
      });
    });
  });

  it('shows success message after successful submission', async () => {
    (apiClient.forgotPassword as jest.Mock).mockResolvedValue(undefined);
    
    render(<ForgotPasswordForm />);
    
    const emailInput = screen.getByLabelText('Email Address');
    const submitButton = screen.getByRole('button', { name: 'Send Reset Link' });
    
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Check Your Email')).toBeInTheDocument();
      expect(screen.getByText(/We've sent a password reset link to/)).toBeInTheDocument();
      expect(screen.getByText('test@example.com')).toBeInTheDocument();
      expect(screen.getByText('Back to Login')).toBeInTheDocument();
      expect(screen.getByText('Try a different email')).toBeInTheDocument();
    });
  });

  it('handles forgot password error', async () => {
    const errorMessage = 'Email not found';
    (apiClient.forgotPassword as jest.Mock).mockRejectedValue(new Error(errorMessage));
    
    render(<ForgotPasswordForm />);
    
    const emailInput = screen.getByLabelText('Email Address');
    const submitButton = screen.getByRole('button', { name: 'Send Reset Link' });
    
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });
  });

  it('shows loading state during submission', async () => {
    (apiClient.forgotPassword as jest.Mock).mockImplementation(() => 
      new Promise(resolve => setTimeout(resolve, 100))
    );
    
    render(<ForgotPasswordForm />);
    
    const emailInput = screen.getByLabelText('Email Address');
    const submitButton = screen.getByRole('button', { name: 'Send Reset Link' });
    
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.click(submitButton);
    
    expect(screen.getByText('Sending...')).toBeInTheDocument();
    expect(submitButton).toBeDisabled();
    
    await waitFor(() => {
      expect(screen.getByText('Check Your Email')).toBeInTheDocument();
    });
  });

  it('clears errors when user starts typing', async () => {
    render(<ForgotPasswordForm />);
    
    const emailInput = screen.getByLabelText('Email Address');
    const submitButton = screen.getByRole('button', { name: 'Send Reset Link' });
    
    // Trigger validation error
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Email is required')).toBeInTheDocument();
    });
    
    // Start typing to clear error
    fireEvent.change(emailInput, { target: { value: 'test' } });
    
    expect(screen.queryByText('Email is required')).not.toBeInTheDocument();
  });

  it('allows trying different email after success', async () => {
    (apiClient.forgotPassword as jest.Mock).mockResolvedValue(undefined);
    
    render(<ForgotPasswordForm />);
    
    const emailInput = screen.getByLabelText('Email Address');
    const submitButton = screen.getByRole('button', { name: 'Send Reset Link' });
    
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Check Your Email')).toBeInTheDocument();
    });
    
    // Click try different email
    const tryDifferentButton = screen.getByText('Try a different email');
    fireEvent.click(tryDifferentButton);
    
    // Should return to form
    expect(screen.getByText('Reset Your Password')).toBeInTheDocument();
    expect(screen.getByLabelText('Email Address')).toHaveValue('');
  });
});